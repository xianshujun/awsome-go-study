//go:build !windows

package main

import "os"

// 非 Windows 平台同样按标准输出判断是否附着终端。
func hasConsole() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// 没有系统对话框时，把错误写成文件，保证信息不会彻底丢失。
func showErrorDialog(title, text string) { writeErrorNotice(title, text) }
