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

const version = "0.2.0"

func main() {
	noBrowser := flag.Bool("no-browser", false, "不自动打开应用窗口")
	addr := flag.String("addr", "127.0.0.1:0", "本机监听地址，默认随机端口")
	dataDir := flag.String("data-dir", "", "学习状态目录，默认使用系统用户配置目录")
	showVersion := flag.Bool("version", false, "显示版本信息")
	diagnose := flag.Bool("diagnose", false, "启动后用对话框显示地址和日志位置")
	flag.Parse()

	if *showVersion {
		info := fmt.Sprintf("GoStudy %s\n运行平台：%s/%s\nGo 版本：%s", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		fmt.Println(info)
		if !hasConsole() {
			showErrorDialog("GoStudy 版本", info)
		}
		return
	}

	log := newStartupLog()
	defer log.close()
	log.write("版本 %s，平台 %s/%s，可执行文件：%s", version, runtime.GOOS, runtime.GOARCH, executablePath())
	if hasConsole() {
		fmt.Printf("GoStudy %s\n", version)
	}

	server, err := backend.Start(backend.Options{Addr: *addr, DataDir: *dataDir}, web.Handler())
	if err != nil {
		log.write("本地引擎启动失败：%v", err)
		reportProblem("GoStudy 启动失败", "本地学习引擎无法启动：\n"+err.Error()+"\n\n详细日志："+log.path)
		os.Exit(1)
	}
	log.write("本地引擎已启动，监听 %s", server.URL())
	if hasConsole() {
		fmt.Println("GoStudy 已启动。关闭本窗口或按 Ctrl+C 可退出。")
		fmt.Println(server.URL())
	}

	if !*noBrowser {
		command, err := openAppWindow(server.URL(), log)
		if err != nil {
			log.write("自动打开窗口失败：%v", err)
			reportProblem("GoStudy 无法打开界面", "没能自动打开应用窗口。\n\n可以手动复制下面的地址到浏览器打开：\n\n"+server.URL()+"\n\n详细日志："+log.path)
			if !hasConsole() {
				// 没有窗口就无法复制地址，退一步打开默认浏览器。
				_ = openDefaultBrowser(server.URL(), log)
			}
		} else {
			log.write("已通过 %s 打开界面", command)
		}
	}
	if *diagnose {
		showErrorDialog("GoStudy 运行中", "应用已在运行：\n\n"+server.URL()+"\n\n日志文件：\n"+log.path+"\n\n在应用内点击「偏好与 AI 设置 → 退出 GoStudy」可结束程序。")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	// 用户既可以用 Ctrl+C 结束进程，也可以在应用内点击「退出 GoStudy」。
	select {
	case <-ctx.Done():
	case <-server.Done():
	}
	log.write("已收到退出请求，正在停止")
	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Stop(shutdown)
	log.write("已退出")
}

func executablePath() string {
	exe, err := os.Executable()
	if err != nil {
		return "未知"
	}
	return exe
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

// openAppWindow 依次尝试 Edge、Chrome 的应用窗口模式，失败再退回默认浏览器。
// 应用窗口模式没有地址栏，看起来接近独立桌面程序。
func openAppWindow(address string, log *startupLog) (string, error) {
	switch runtime.GOOS {
	case "windows":
		type candidate struct {
			name string
			path string
		}
		var candidates []candidate
		for _, base := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles"), os.Getenv("LOCALAPPDATA")} {
			if base == "" {
				continue
			}
			candidates = append(candidates,
				candidate{"Edge", filepath.Join(base, "Microsoft", "Edge", "Application", "msedge.exe")},
				candidate{"Chrome", filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe")},
			)
		}
		candidates = append(candidates, candidate{"Edge", "msedge.exe"}, candidate{"Chrome", "chrome.exe"})
		for _, item := range candidates {
			if err := launch(item.path, "--app="+address, "--window-size=1320,900"); err == nil {
				return item.name, nil
			} else {
				log.write("启动 %s 失败（%s）：%v", item.name, item.path, err)
			}
		}
		log.write("未找到 Edge/Chrome，退回默认浏览器")
		return "默认浏览器", launch("rundll32.exe", "url.dll,FileProtocolHandler", address)
	case "darwin":
		return "open", launch("open", address)
	default:
		return "xdg-open", launch("xdg-open", address)
	}
}

func openDefaultBrowser(address string, log *startupLog) error {
	log.write("尝试打开默认浏览器")
	switch runtime.GOOS {
	case "windows":
		return launch("rundll32.exe", "url.dll,FileProtocolHandler", address)
	case "darwin":
		return launch("open", address)
	default:
		return launch("xdg-open", address)
	}
}
