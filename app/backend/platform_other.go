//go:build !windows

package backend

import (
	"os/exec"
	"syscall"
)

// closeAppWindows 只在 Windows 上有意义：那里的应用窗口是 Edge/Chrome 的应用模式。
// Linux 上窗口由 xdg-open 交给默认浏览器，交给用户自己关。
func closeAppWindows(string) {}

// startDetached 让子进程脱离当前会话，父进程退出不会把它一起带走。
func startDetached(command *exec.Cmd) error {
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
