package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/xianshujun/awsome-go-study/app/backend"
	"github.com/xianshujun/awsome-go-study/app/web"
)

func main() {
	noBrowser := flag.Bool("no-browser", false, "不自动打开应用窗口")
	addr := flag.String("addr", "127.0.0.1:0", "本机监听地址，默认随机端口")
	dataDir := flag.String("data-dir", "", "学习状态目录，默认使用系统用户配置目录")
	flag.Parse()
	server, err := backend.Start(backend.Options{Addr: *addr, DataDir: *dataDir}, web.Handler())
	if err != nil {
		reportProblem("GoStudy 启动失败", "本地学习引擎无法启动：\n"+err.Error())
		os.Exit(1)
	}
	if hasConsole() {
		fmt.Println("GoStudy 已启动。关闭本窗口或按 Ctrl+C 可退出。")
		fmt.Println(server.URL())
	}
	if !*noBrowser {
		if err := openAppWindow(server.URL()); err != nil {
			reportProblem("GoStudy 无法打开界面", "请手动复制下面的地址，在浏览器中打开：\n\n"+server.URL())
			if !hasConsole() {
				// 没有窗口就无法粘贴地址，只能辅助打开默认浏览器。
				_ = openDefaultBrowser(server.URL())
			}
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// 用户既可以用 Ctrl+C 结束进程，也可以在应用内点击「退出 GoStudy」。
	select {
	case <-ctx.Done():
	case <-server.Done():
	}
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Stop(shutdown)
}

func reportProblem(title, message string) {
	fmt.Fprintln(os.Stderr, message)
	if !hasConsole() {
		showErrorDialog(title, message)
	}
}

func launch(command string, args ...string) error {
	c := exec.Command(command, args...)
	if err := c.Start(); err != nil {
		return err
	}
	go func() { _ = c.Wait() }()
	return nil
}

// openAppWindow 优先使用 Edge 的应用窗口模式，让界面看起来像桌面应用而不是浏览器标签页。
func openAppWindow(address string) error {
	switch runtime.GOOS {
	case "windows":
		candidates := []string{"msedge.exe"}
		for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
			if base != "" {
				candidates = append(candidates, filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe"))
			}
		}
		for _, edge := range candidates {
			if err := launch(edge, "--app="+address, "--window-size=1320,900"); err == nil {
				return nil
			}
		}
		// 没有 Edge 时退回默认浏览器，至少功能可用。
		return launch("rundll32.exe", "url.dll,FileProtocolHandler", address)
	case "darwin":
		return launch("open", address)
	default:
		return launch("xdg-open", address)
	}
}

func openDefaultBrowser(address string) error {
	switch runtime.GOOS {
	case "windows":
		return launch("rundll32.exe", "url.dll,FileProtocolHandler", address)
	case "darwin":
		return launch("open", address)
	default:
		return launch("xdg-open", address)
	}
}
