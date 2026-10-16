package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestInstanceFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, instanceFileName)
	info := instanceInfo{ID: newInstanceID(), PID: os.Getpid(), URL: "http://127.0.0.1:9/#token=abc", Version: "0.8.0", Started: 1}
	if info.ID == "" {
		t.Fatal("拿不到实例标识")
	}
	if err := writeInstance(path, info); err != nil {
		t.Fatal(err)
	}
	got, ok := readInstance(path)
	if !ok || got.ID != info.ID || got.URL != info.URL || got.Version != info.Version {
		t.Fatalf("读回来的实例信息不对：%+v ok=%v", got, ok)
	}
	// 更新重启时新进程会先写好自己的那一份，老进程退出时不能顺手删掉它。
	removeInstance(path, "别人的标识")
	if _, ok := readInstance(path); !ok {
		t.Fatal("不该删掉不是自己写的实例文件")
	}
	removeInstance(path, info.ID)
	if _, ok := readInstance(path); ok {
		t.Fatal("自己写的实例文件应当在退出时删掉")
	}
	// 坏文件与空文件都按"没有实例"处理。
	if err := os.WriteFile(path, []byte("{不是 JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, ok := readInstance(path); ok {
		t.Fatal("坏掉的实例文件不该被认成实例")
	}
}

func TestProbeInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprint(w, `{"status":"ok","windowOpen":true}`)
	}))
	defer server.Close()
	if alive, windowOpen := probeInstance(server.URL, "secret", time.Second); !alive || !windowOpen {
		t.Fatalf("应当探到活着的实例且窗口开着：alive=%v windowOpen=%v", alive, windowOpen)
	}
	if alive, _ := probeInstance(server.URL, "错误凭据", time.Second); alive {
		t.Fatal("凭据不对不算活着，免得把别人的服务当成自己的实例")
	}
	// 没人监听的地址：上次异常退出留下的文件就走这条判断。
	if alive, _ := probeInstance("http://127.0.0.1:1", "secret", 300*time.Millisecond); alive {
		t.Fatal("没人应答的地址不该算活着")
	}
}

// fakeInstance 模拟一个已经在跑的实例：/api/status 报告窗口状态，/api/quit 触发回调。
func fakeInstance(t *testing.T, windowOpen bool, onQuit func()) (*httptest.Server, string) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = fmt.Fprintf(w, `{"status":"ok","windowOpen":%t}`, windowOpen)
	})
	mux.HandleFunc("/api/quit", func(w http.ResponseWriter, r *http.Request) {
		if onQuit != nil {
			onQuit()
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server, server.URL + "/#token=secret"
}

func TestHandleExistingInstance(t *testing.T) {
	log := &startupLog{}
	cases := []struct {
		name       string
		windowOpen bool
		version    string
		build      string
		want       bool
	}{
		{"窗口开着就别再起一个", true, "0.8.0", "0.8.0", true},
		{"窗口关了就把它的窗口开回来", false, "0.8.0", "0.8.0", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, instanceFileName)
			_, launchURL := fakeInstance(t, tt.windowOpen, nil)
			if err := writeInstance(path, instanceInfo{ID: newInstanceID(), PID: 4242, URL: launchURL, Version: tt.version}); err != nil {
				t.Fatal(err)
			}
			// gui=false：测试里不去开浏览器窗口，只看判断结果。
			if got := handleExistingInstance(path, false, tt.build, log); got != tt.want {
				t.Fatalf("handleExistingInstance = %v，想要 %v", got, tt.want)
			}
		})
	}

	t.Run("版本不一样时请旧实例退出", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, instanceFileName)
		quitRequested := make(chan struct{}, 1)
		server, launchURL := fakeInstance(t, false, func() { quitRequested <- struct{}{} })
		// 收到退出请求之后，旧实例过一会儿才真的把端口让出来。
		go func() {
			<-quitRequested
			server.Close()
		}()
		if err := writeInstance(path, instanceInfo{ID: newInstanceID(), PID: 4242, URL: launchURL, Version: "0.7.0"}); err != nil {
			t.Fatal(err)
		}
		// 旧实例退出后要由新版本继续启动（handled=false），而不是让用户一直跑着旧版本。
		if handleExistingInstance(path, false, "0.8.0", log) {
			t.Fatal("旧实例退出后应当由新版本继续启动")
		}
	})

	t.Run("地址不应答就正常启动", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, instanceFileName)
		if err := writeInstance(path, instanceInfo{ID: newInstanceID(), PID: 4242, URL: "http://127.0.0.1:1/#token=secret", Version: "0.8.0"}); err != nil {
			t.Fatal(err)
		}
		if handleExistingInstance(path, false, "0.8.0", log) {
			t.Fatal("地址不应答时应当继续启动")
		}
		if _, ok := readInstance(path); ok {
			t.Fatal("过期的实例文件应当清掉")
		}
	})

	t.Run("文件格式不对就正常启动", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, instanceFileName)
		if err := writeInstance(path, instanceInfo{ID: newInstanceID(), URL: "http://127.0.0.1:1234/没有凭据"}); err != nil {
			t.Fatal(err)
		}
		if handleExistingInstance(path, false, "0.8.0", log) {
			t.Fatal("地址格式不对时应当继续启动")
		}
	})

	t.Run("没有文件就是头一份", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), instanceFileName)
		if handleExistingInstance(path, false, "0.8.0", log) {
			t.Fatal("没有实例文件时应当继续启动")
		}
	})
}

func TestShutdownInstance(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/quit", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	server := httptest.NewServer(mux)
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"status":"ok","windowOpen":true}`)
	})
	if shutdownInstance(server.URL, "secret", time.Second) {
		t.Fatal("实例还在应答时不该报告已退出")
	}
	server.Close()
	// 已经关掉的实例：一次探活就够，等的时间应当远小于上限。
	started := time.Now()
	if !shutdownInstance(server.URL+"/#token=secret", "secret", 5*time.Second) {
		t.Fatal("实例已经停了却报告没退出")
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("探到地址已经不应答就该返回，实际等了 %v", elapsed)
	}
}
