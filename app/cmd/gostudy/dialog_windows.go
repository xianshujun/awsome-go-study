//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
	procConsole    = user32.NewProc("GetConsoleWindow")
)

// hasConsole 判断进程是否连接了控制台窗口；以 windowsgui 子系统构建时为 false，
// 此时启动错误改用系统对话框提示，双击应用的用户才能看到原因。
func hasConsole() bool {
	handle, _, _ := procConsole.Call()
	return handle != 0
}

// showErrorDialog 显示一个模态提示框，文本可选中复制，方便用户反馈地址或错误。
func showErrorDialog(title, text string) {
	caption, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return
	}
	message, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return
	}
	const iconWarning, setForeground = 0x30, 0x10000
	_, _, _ = procMessageBox.Call(0, uintptr(unsafe.Pointer(message)), uintptr(unsafe.Pointer(caption)), iconWarning|setForeground)
}
