//go:build !windows

package backend

/*
非 Windows 平台没有随程序走的系统对话框：Linux 上应用窗口只是浏览器里的一个标签页，
窗口关掉之后没法再向用户问一句话。这里的做法是直接按"直接退出"收尾——
用户关掉窗口之后进程还留着，正是之前那个"关不掉的进程"。
*/
func askCloseBehavior() CloseAnswer { return CloseAnswer{Answered: false} }

// closePromptAvailable 表示这个平台能在窗口关掉之后弹一次询问框。
func closePromptAvailable() bool { return false }
