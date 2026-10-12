//go:build !windows

package main

// 非 Windows 平台始终有终端可看错误输出。
func hasConsole() bool { return true }

func showErrorDialog(string, string) {}
