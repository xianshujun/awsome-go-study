package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/xianshujun/awsome-go-study/app/backend"
	"github.com/xianshujun/awsome-go-study/app/web"
)

// version 由构建脚本注入：-X .../app/backend.Version=0.11.1。
// 直接 go run 的源码构建是 dev，更新检查会跳过程序本身的比较。
var version = backend.Version

func main() {
	noBrowser := flag.Bool("no-browser", false, "不自动打开应用窗口")
	cleanup := flag.String("cleanup", "", "内部参数：启动后删除上一次更新留下的旧程序文件")
	addr := flag.String("addr", "127.0.0.1:0", "本机监听地址，默认随机端口")
	dataDir := flag.String("data-dir", "", "学习状态目录，默认使用系统用户配置目录")
	showVersion := flag.Bool("version", false, "显示版本信息")
	diagnose := flag.Bool("diagnose", false, "启动后用对话框显示地址和日志位置")
	selfTest := flag.Bool("selftest", false, "逐步自检启动环境并输出结果，不打开窗口")
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
	// GUI 程序没有控制台，panic 会静默退出；这里统一记录并弹窗，避免"双击没反应"。
	defer func() {
		if r := recover(); r != nil {
			log.write("发生 panic：%v\n%s", r, debug.Stack())
			reportProblem("GoStudy 启动失败（内部错误）", fmt.Sprintf("%v\n\n详细日志：%s", r, log.path))
		}
		log.close()
	}()
	// 自我更新后的第一件事：把上一版留下的 GoStudy.exe.old 删掉。
	// 旧进程刚退出时文件可能还被占用，所以重试几次；删不掉也只是多占一份空间。
	if *cleanup != "" {
		go removeOldBinary(*cleanup, log)
	}

	log.write("版本 %s，平台 %s/%s，可执行文件：%s", version, runtime.GOOS, runtime.GOARCH, executablePath())
	fmt.Printf("GoStudy %s\n", version)

	if *selfTest {
		runSelfTest(*dataDir, log)
		return
	}

	/*
		单实例：同一份学习数据只让一个进程用。
		更新重启（带 -cleanup）时跳过这步——那一刻旧进程还在退出，探活一定会探到它，
		跳过才能让新版本顺利接管。
	*/
	instancePath := instanceFile(*dataDir)
	if *cleanup == "" && instancePath != "" && handleExistingInstance(instancePath, !*noBrowser, version, log) {
		log.write("已有实例在运行，本次启动结束")
		return
	}

	// 逐步记录，任何一步失败或被杀掉都能从日志看出停在哪里。
	log.write("步骤 1/3：准备前端资源")
	handler := web.Handler()
	log.write("步骤 1/3：完成")
	log.write("步骤 2/3：启动本地引擎（监听 %s）", *addr)
	server, err := backend.Start(backend.Options{Addr: *addr, DataDir: *dataDir, Logf: log.write}, handler)
	if err != nil {
		log.write("步骤 2/3：失败：%v", err)
		reportProblem("GoStudy 启动失败", "本地学习引擎无法启动：\n"+err.Error()+"\n\n详细日志："+log.path)
		pauseIfConsole()
		os.Exit(1)
	}
	log.write("步骤 2/3：完成，监听 %s", server.URL())
	fmt.Println("GoStudy 已启动。关闭本窗口或按 Ctrl+C 可退出。")
	fmt.Println(server.URL())

	// 记下这次的地址与凭据：下次双击 GoStudy 时先认这个文件，不再另起一个进程。
	record := instanceInfo{ID: newInstanceID(), PID: os.Getpid(), URL: server.URL(), Version: version, Started: time.Now().Unix()}
	if instancePath != "" {
		if err := writeInstance(instancePath, record); err != nil {
			log.write("写入实例文件失败（不影响使用）：%v", err)
		} else {
			defer removeInstance(instancePath, record.ID)
		}
	}

	if !*noBrowser {
		log.write("步骤 3/3：打开应用窗口")
		command, err := openAppWindow(server.URL(), log)
		if err != nil {
			log.write("步骤 3/3：失败：%v", err)
			reportProblem("GoStudy 无法打开界面", "没能自动打开应用窗口。\n\n可以手动复制下面的地址到浏览器打开：\n\n"+server.URL()+"\n\n详细日志："+log.path)
			if !hasConsole() {
				// 没有窗口就无法复制地址，退一步打开默认浏览器。
				_ = openDefaultBrowser(server.URL(), log)
			}
		} else {
			log.write("步骤 3/3：完成，已通过 %s 打开", command)
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

// removeOldBinary 清理旧版本可执行文件；失败不报错，只记一行日志。
func removeOldBinary(path string, log *startupLog) {
	// 稍等一下再动手：退出中的旧进程可能还按着这个文件。
	time.Sleep(2 * time.Second)
	for attempt := 0; attempt < 10; attempt++ {
		err := os.Remove(path)
		if err == nil || os.IsNotExist(err) {
			log.write("已清理旧版本文件：%s", path)
			return
		}
		time.Sleep(time.Second)
	}
	log.write("旧版本文件暂时删不掉，可以手动删除：%s", path)
}

// runSelfTest 逐项检查启动依赖，把结果写进日志并用对话框展示，
// 用来定位"双击没反应"到底是哪一步失败。
func runSelfTest(dataDir string, log *startupLog) {
	type result struct {
		name string
		err  error
	}
	var results []result
	step := func(name string, run func() error) {
		log.write("自检：%s …", name)
		err := run()
		if err != nil {
			log.write("自检：%s 失败：%v", name, err)
		} else {
			log.write("自检：%s 通过", name)
		}
		results = append(results, result{name, err})
	}

	step("用户配置目录可写", func() error {
		dir := dataDir
		if dir == "" {
			base, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			dir = filepath.Join(base, "go-study")
		}
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		probe := filepath.Join(dir, "write-test.tmp")
		if err := os.WriteFile(probe, []byte("ok"), 0600); err != nil {
			return err
		}
		return os.Remove(probe)
	})
	step("本机端口监听", func() error {
		listener, err := net.Listen("tcp4", "127.0.0.1:0")
		if err != nil {
			return err
		}
		return listener.Close()
	})
	step("前端资源加载", func() error {
		web.Handler()
		return nil
	})
	step("课程内容加载", func() error {
		courses, lessons, err := backend.LoadCurriculum()
		if err != nil {
			return err
		}
		if len(courses) == 0 || len(lessons) == 0 {
			return fmt.Errorf("课程为空")
		}
		return nil
	})
	step("浏览器可用性", func() error {
		name, available := findBrowser()
		if !available {
			return fmt.Errorf("未找到 Edge/Chrome，将退回默认浏览器")
		}
		log.write("自检：可用浏览器 %s", name)
		return nil
	})
	step("完整引擎启动与接口响应", func() error {
		server, err := backend.Start(backend.Options{Addr: "127.0.0.1:0", DataDir: dataDir}, web.Handler())
		if err != nil {
			return err
		}
		defer func() {
			shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = server.Stop(shutdown)
		}()
		base, token, err := splitLaunchURL(server.URL())
		if err != nil {
			return err
		}
		request, err := http.NewRequest(http.MethodGet, base+"/api/bootstrap", nil)
		if err != nil {
			return err
		}
		request.Header.Set("Authorization", "Bearer "+token)
		client := &http.Client{Timeout: 5 * time.Second}
		response, err := client.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return fmt.Errorf("接口返回 %d", response.StatusCode)
		}
		return nil
	})

	var lines []string
	failed := 0
	for _, item := range results {
		if item.err != nil {
			failed++
			lines = append(lines, "✗ "+item.name+"："+item.err.Error())
		} else {
			lines = append(lines, "✓ "+item.name)
		}
	}
	summary := strings.Join(lines, "\n")
	log.write("自检完成：\n%s", summary)
	fmt.Println(summary)
	if !hasConsole() {
		title := "GoStudy 自检通过"
		if failed > 0 {
			title = "GoStudy 自检发现问题"
		}
		showErrorDialog(title, summary+"\n\n日志文件：\n"+log.path)
	}
}

// splitLaunchURL 把启动地址拆成基础地址与访问凭证，供自检请求使用。
func splitLaunchURL(launchURL string) (string, string, error) {
	index := strings.Index(launchURL, "/#token=")
	if index < 0 {
		return "", "", fmt.Errorf("启动地址格式不符合预期")
	}
	base := launchURL[:index]
	token := launchURL[index+len("/#token="):]
	if base == "" || token == "" {
		return "", "", fmt.Errorf("启动地址缺少地址或凭证")
	}
	return base, token, nil
}

// findBrowser 返回第一个可用的浏览器应用窗口目标。
func findBrowser() (string, bool) {
	if runtime.GOOS != "windows" {
		return "系统默认浏览器", true
	}
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
	for _, item := range candidates {
		if _, err := os.Stat(item.path); err == nil {
			return item.name, true
		}
	}
	return "", false
}

func pauseIfConsole() {
	if !hasConsole() {
		return
	}
	fmt.Print("\n按回车键退出…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
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
