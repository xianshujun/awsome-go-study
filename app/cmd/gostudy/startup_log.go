package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// startupLog 记录每次启动的关键步骤，双击运行失败时可以直接看这个文件。
// 优先放在程序所在目录（用户下载的位置好找），不可写时退回用户配置目录。
type startupLog struct {
	file *os.File
	path string
}

func newStartupLog() *startupLog {
	candidates := make([]string, 0, 2)
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Join(filepath.Dir(exe), "GoStudy-startup.log"))
	}
	if dir, err := os.UserConfigDir(); err == nil {
		candidates = append(candidates, filepath.Join(dir, "go-study", "startup.log"))
	}
	for _, path := range candidates {
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			continue
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			continue
		}
		log := &startupLog{file: file, path: path}
		log.write("========== 启动 %s ==========", time.Now().Format("2006-01-02 15:04:05"))
		return log
	}
	return &startupLog{}
}

func (l *startupLog) write(format string, args ...any) {
	if l == nil || l.file == nil {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		return
	}
	fmt.Fprintf(l.file, format+"\n", args...)
	_ = l.file.Sync()
}

func (l *startupLog) close() {
	if l != nil && l.file != nil {
		_ = l.file.Close()
	}
}

// writeErrorNotice 在程序目录（不可写时退到用户配置目录）写下错误文件，
// 作为对话框不可用时的最后一道可见性保障。
func writeErrorNotice(title, text string) {
	body := title + "\n\n" + text + "\n"
	for _, path := range noticePaths() {
		if err := os.WriteFile(path, []byte(body), 0600); err == nil {
			return
		}
	}
}

func noticePaths() []string {
	paths := make([]string, 0, 2)
	if exe, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(exe), "GoStudy-error.txt"))
	}
	if dir, err := os.UserConfigDir(); err == nil {
		paths = append(paths, filepath.Join(dir, "go-study", "error.txt"))
	}
	return paths
}
