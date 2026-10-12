//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"
)

// 注意：这里只声明懒加载句柄，不会在包初始化时解析导出，避免启动即崩溃。
var (
	user32         = syscall.NewLazyDLL("user32.dll")
	procMessageBox = user32.NewProc("MessageBoxW")
)

// hasConsole 通过标准输出判断是否附着控制台。
// 不再调用 user32：部分机器上（安全软件注入、非标准系统模块）GetProcAddress 解析
// GetConsoleWindow 会失败，LazyProc.Call 随即 panic；GUI 程序没有控制台，panic
// 就等于进程静默退出——这正是"双击没反应"的根因。
func hasConsole() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// showErrorDialog 尽力弹出系统对话框；任何失败都退化为写错误文件，绝不 panic。
func showErrorDialog(title, text string) {
	if !tryMessageBox(title, text) {
		writeErrorNotice(title, text)
	}
}

func tryMessageBox(title, text string) (ok bool) {
	// 极端环境下 DLL 可能被替换或缺少导出，这里把 panic 转成失败返回。
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	if err := procMessageBox.Find(); err != nil {
		return false
	}
	caption, err := syscall.UTF16PtrFromString(title)
	if err != nil {
		return false
	}
	message, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		return false
	}
	const iconWarning, setForeground = 0x30, 0x10000
	procMessageBox.Call(0, uintptr(unsafe.Pointer(message)), uintptr(unsafe.Pointer(caption)), iconWarning|setForeground)
	return true
}
