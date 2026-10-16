//go:build windows

package backend

import (
	"os/exec"
	"strings"
	"syscall"
)

/*
Windows 用 GUI 子系统启动，没有控制台；应用窗口是 Edge/Chrome 的 --app 模式。
更新重启前要把这个窗口关掉，否则会留下一个指向已经退出引擎的空白页。
*/
const (
	detachedProcess = 0x00000008
	newProcessGroup = 0x00000200
	createNoWindow  = 0x08000000
)

// closeAppWindows 关掉指向本机地址的应用窗口；失败也不影响更新，只是会留个空白页。
func closeAppWindows(host string) {
	script := strings.Join([]string{
		"Get-CimInstance Win32_Process -Filter \"Name='msedge.exe' or Name='chrome.exe'\"",
		"Where-Object { $_.CommandLine -like '*--app=http://" + host + "*' }",
		"ForEach-Object { Stop-Process -Id $_.ProcessId -Force -ErrorAction SilentlyContinue }",
	}, " | ")
	command := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	_ = command.Run()
}

// startDetached 启动一个不随本进程退出的子进程（更新后重启自己时用）。
func startDetached(command *exec.Cmd) error {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: detachedProcess | newProcessGroup}
	if err := command.Start(); err != nil {
		return err
	}
	// 不等待子进程；它的存活与本进程无关。
	return command.Process.Release()
}
