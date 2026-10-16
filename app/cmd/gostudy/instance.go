package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

/*
单实例。

同一份学习数据只让一个 GoStudy 进程用：状态文件没有跨进程锁，两个进程同时写会互相
覆盖；用户也会在任务管理器里看到好几个 GoStudy.exe。数据目录里放一个 instance.json，
启动时先看它：

  - 地址还能应答 → 已经有实例在跑。窗口开着就提示一句；窗口已经关了（用户在关闭提示里
    选了后台静默）就替它把窗口重新打开，然后本进程退出。
  - 还在跑但版本不一样（刚装完新版）→ 先请旧实例退出，再由新版本接管。
  - 不应答（上次异常退出留下的文件）→ 当成没有实例，直接覆盖。

任何一步出意外都按"没有实例"继续启动：宁可偶尔多起一个进程，也不能让用户打不开应用。
*/
const instanceFileName = "instance.json"

// instanceInfo 是 instance.json 的内容，够另一个进程找到这个实例、也够它认出自己那一份。
type instanceInfo struct {
	// ID 是本次运行的随机标识：退出时只删自己写的文件，
	// 不然更新重启时老进程会把新进程刚写好的那份删掉。
	ID      string `json:"id"`
	PID     int    `json:"pid"`
	URL     string `json:"url"`
	Version string `json:"version,omitempty"`
	Started int64  `json:"started"`
}

// instanceFile 返回数据目录里的实例文件路径；数据目录规则与引擎、日志保持一致。
func instanceFile(dataDir string) string {
	if dataDir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return ""
		}
		dataDir = filepath.Join(base, "go-study")
	}
	return filepath.Join(dataDir, instanceFileName)
}

func newInstanceID() string {
	seed := make([]byte, 8)
	if _, err := rand.Read(seed); err != nil {
		return ""
	}
	return hex.EncodeToString(seed)
}

func readInstance(path string) (instanceInfo, bool) {
	file, err := os.Open(path)
	if err != nil {
		return instanceInfo{}, false
	}
	defer file.Close()
	var info instanceInfo
	if err := json.NewDecoder(io.LimitReader(file, 8<<10)).Decode(&info); err != nil {
		return instanceInfo{}, false
	}
	if info.URL == "" {
		return instanceInfo{}, false
	}
	return info, true
}

// writeInstance 用临时文件加改名写入，避免另一个进程读到半份 JSON。
func writeInstance(path string, info instanceInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".instance-*")
	if err != nil {
		return err
	}
	name := temp.Name()
	defer os.Remove(name)
	if err = temp.Chmod(0600); err != nil {
		temp.Close()
		return err
	}
	if _, err = temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err = temp.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// removeInstance 只删还是自己的那一份。
func removeInstance(path, id string) {
	if path == "" || id == "" {
		return
	}
	current, ok := readInstance(path)
	if !ok || current.ID != id {
		return
	}
	_ = os.Remove(path)
}

// probeInstance 问一句 /api/status：地址还活着吗、它的应用窗口还开着吗。
func probeInstance(base, token string, timeout time.Duration) (alive, windowOpen bool) {
	request, err := http.NewRequest(http.MethodGet, base+"/api/status", nil)
	if err != nil {
		return false, false
	}
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return false, false
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return false, false
	}
	var status struct {
		WindowOpen bool `json:"windowOpen"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<10)).Decode(&status); err != nil {
		return false, false
	}
	return true, status.WindowOpen
}

// shutdownInstance 请旧实例退出，并等它把地址让出来，好让新版本接管。
func shutdownInstance(base, token string, wait time.Duration) bool {
	request, err := http.NewRequest(http.MethodPost, base+"/api/quit", nil)
	if err != nil {
		return false
	}
	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{Timeout: 2 * time.Second}
	if response, err := client.Do(request); err == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1<<10))
		response.Body.Close()
	}
	deadline := time.Now().Add(wait)
	for {
		// 先等一小会儿再探：进程要停止引擎、退出还要一点时间。
		time.Sleep(250 * time.Millisecond)
		if alive, _ := probeInstance(base, token, 400*time.Millisecond); !alive {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
	}
}

/*
handleExistingInstance 处理"已经有实例在跑"的情况，返回 true 表示本进程该退出了。

gui 为真时才会去开窗口或弹提示（对应 -no-browser 关掉的情况，那时只往终端打印地址）。
*/
func handleExistingInstance(path string, gui bool, buildVersion string, log *startupLog) bool {
	existing, ok := readInstance(path)
	if !ok {
		return false
	}
	base, token, err := splitLaunchURL(existing.URL)
	if err != nil {
		log.write("实例文件内容不对，按没有实例处理：%v", err)
		_ = os.Remove(path)
		return false
	}
	alive, windowOpen := probeInstance(base, token, 900*time.Millisecond)
	if !alive {
		log.write("实例文件里的进程（PID %d）已经不在了，继续启动", existing.PID)
		_ = os.Remove(path)
		return false
	}
	// 版本换了多半是刚装完新版：让旧实例先退，新版接管，别让用户一直跑着老版本。
	if existing.Version != "" && existing.Version != buildVersion && buildVersion != "dev" && buildVersion != "0.0.0" {
		log.write("发现 %s 的实例在运行，请它退出后由 %s 接管", existing.Version, buildVersion)
		if shutdownInstance(base, token, 6*time.Second) {
			return false
		}
		log.write("旧实例没有退出，仍然启动新实例")
		return false
	}
	if windowOpen {
		log.write("已经有 GoStudy 在运行（PID %d），不再启动第二个进程", existing.PID)
		if gui {
			showErrorDialog("GoStudy 已经在运行", "已经有一个 GoStudy 在运行，不用再打开一个。\n\n"+
				"如果找不到它的窗口，可以在任务栏里找找，或者先退出它再重新打开。\n\n"+
				"当前地址：\n"+existing.URL)
		} else {
			log.write("已有实例地址：%s", existing.URL)
			fmt.Println("GoStudy 已经在运行：" + existing.URL)
		}
		return true
	}
	// 窗口关掉但进程还在后台静默运行：替用户把窗口开回来。
	log.write("已有实例在后台静默运行（PID %d），重新打开它的窗口", existing.PID)
	if !gui {
		fmt.Println("GoStudy 在后台运行：" + existing.URL)
		return true
	}
	if name, err := openAppWindow(existing.URL, log); err != nil {
		log.write("没能重新打开窗口：%v", err)
		showErrorDialog("GoStudy 在后台运行", "GoStudy 的进程还在后台静默运行，但没能自动打开窗口。\n\n"+
			"可以手动把下面的地址复制到浏览器打开：\n\n"+existing.URL)
	} else {
		log.write("已通过 %s 重新打开窗口", name)
	}
	return true
}
