package backend

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

/*
应用内卸载。

安装版由 Inno Setup 打包，程序目录里会放一个 unins*.exe；点"卸载"就把它拉起来，
然后本进程退出——不退出的话，卸载程序删不掉正在运行的 GoStudy.exe。

学习进度、笔记与对话都在用户配置目录里，不在程序目录，所以卸载不会碰它们；
这里是免安装版（目录里没有卸载程序）时给用户一句能照做的话。
*/
const portableUninstallHint = "没有找到卸载程序：当前是免安装版，直接删掉 GoStudy.exe 所在的文件夹即可。学习进度、笔记与对话保存在用户配置目录里，删程序不会影响它们。"

// findUninstaller 只在 Windows 上找 Inno 卸载程序；免安装版返回空。
func findUninstaller(platform string) string {
	if platform != "windows" {
		return ""
	}
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	return findUninstallerIn(filepath.Dir(executable))
}

// findUninstallerIn 在给定目录里找 unins*.exe：Inno 装多个版本时会叫 unins001.exe 之类。
func findUninstallerIn(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		name := strings.ToLower(entry.Name())
		if entry.IsDir() || !strings.HasPrefix(name, "unins") || !strings.HasSuffix(name, ".exe") {
			continue
		}
		return filepath.Join(dir, entry.Name())
	}
	return ""
}

/*
handleUninstall 拉起卸载程序，然后收尾退出。

先在响应之前把进程拉起来：起不来就直接告诉用户（多半是权限或文件被删了），
而不是先关掉界面再说"没法卸载"。起来了就退回界面，等一会儿再关窗口、退出进程。
*/
func (e *Engine) handleUninstall(w http.ResponseWriter, r *http.Request) {
	if e.uninstaller == "" {
		writeError(w, 409, portableUninstallHint)
		return
	}
	// /SILENT：应用里已经确认过一次，卸载程序只显示进度窗口，不再多问一遍。
	command := exec.Command(e.uninstaller, "/SILENT", "/NORESTART")
	if err := startDetached(command); err != nil {
		e.log("卸载程序没能启动：%v", err)
		writeError(w, 500, "没能启动卸载程序，请到「应用和功能」里卸载 GoStudy。")
		return
	}
	e.log("已启动卸载程序：%s", e.uninstaller)
	writeJSON(w, 200, map[string]string{"status": "uninstalling"})
	go func() {
		// 留给响应与页面上最后那句话的时间，然后关掉窗口、退出进程，
		// 让卸载程序能把 GoStudy.exe 删干净。
		time.Sleep(1200 * time.Millisecond)
		if e.platform == "windows" {
			closeAppWindows(e.host)
		}
		time.Sleep(300 * time.Millisecond)
		e.requestQuit()
	}()
}
