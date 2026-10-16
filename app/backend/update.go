package backend

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

/*
更新走两条频道，都是 GitHub Release 上的静态文件，同一批内容也镜像到 Gitee：

	latest  频道：update.json + 各平台产物（GoStudy.exe / Setup / linux / apk）
	content 频道：content.json + content.zip（课程内容包，可以单独发，不必发版）

检查更新就是下载这两个小 JSON，不调 GitHub API，因此不受接口限流影响。
*/
const (
	latestManifest  = "update.json"
	contentManifest = "content.json"
	contentPack     = "content.zip"
	// 清单和产物各自的大小上限，防止把别的什么东西当更新下载进来。
	maxManifestBytes = 128 << 10
	maxAssetBytes    = 512 << 20
	// 下载卡死判定：90 秒没有任何新数据就放弃，而不是干等到超时。
	stallLimit = 90 * time.Second
	// 单个请求的最长总时长（大文件走 CDN，给足）。
	downloadTimeout = 30 * time.Minute
	manifestTimeout = 20 * time.Second
)

func latestBases() []string {
	if custom := splitList(os.Getenv("GOSTUDY_UPDATE_BASES")); len(custom) > 0 {
		return custom
	}
	return []string{
		"https://gitee.com/pang-xianshu/gostudy-release/releases/download/latest",
		"https://github.com/xianshujun/awsome-go-study/releases/latest/download",
	}
}

// contentBases 与程序清单同源：内容包和产物在同一次发布里，
// 这样 Gitee 的 latest 别名（指向最新一次发布）对两者都成立。
func contentBases() []string {
	if custom := splitList(os.Getenv("GOSTUDY_CONTENT_BASES")); len(custom) > 0 {
		return custom
	}
	return latestBases()
}

func splitList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, strings.TrimRight(trimmed, "/"))
		}
	}
	return out
}

// preferred 把国内访客更快的 Gitee 排到前面；海外访客反过来试。
func preferred(bases []string, prefer string) []string {
	if strings.EqualFold(prefer, "cn") || len(bases) < 2 {
		return bases
	}
	out := make([]string, 0, len(bases))
	for i := len(bases) - 1; i >= 0; i-- {
		out = append(out, bases[i])
	}
	return out
}

// newUpdateClient 与 AI 客户端不同：发布资源在 CDN 上，必然有跳转，所以允许有限次重定向，
// 但只允许跳到 HTTPS（本机回环除外，方便测试与自建镜像）。总超时不设，改由每次请求自己控制。
func newUpdateClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.MaxResponseHeaderBytes = 32 << 10
	transport.ResponseHeaderTimeout = manifestTimeout
	transport.IdleConnTimeout = 30 * time.Second
	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("更新地址跳转过多")
			}
			host := req.URL.Hostname()
			if req.URL.Scheme != "https" && !(req.URL.Scheme == "http" && (host == "localhost" || isLoopbackHost(host))) {
				return errors.New("更新地址必须使用 HTTPS")
			}
			return nil
		},
	}
}

func isLoopbackHost(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

type updateFile struct {
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// releaseManifest 是 latest 频道里的 update.json。
type releaseManifest struct {
	Version  string                `json:"version"`
	Released string                `json:"released"`
	Notes    string                `json:"notes"`
	Files    map[string]updateFile `json:"files"`
}

// contentInfo 是 content 频道里的 content.json。
type contentInfo struct {
	Version string     `json:"version"`
	File    string     `json:"file"`
	Pack    updateFile `json:"pack"`
	Count   int        `json:"count"`
	Notes   string     `json:"notes"`
}

type UpdateOffer struct {
	Version  string `json:"version"`
	Released string `json:"released,omitempty"`
	Notes    string `json:"notes,omitempty"`
	Size     int64  `json:"size"`
	Asset    string `json:"asset"`
	Source   string `json:"source"`
}

type ContentOffer struct {
	Version string `json:"version"`
	Size    int64  `json:"size"`
	Notes   string `json:"notes,omitempty"`
	Source  string `json:"source"`
	Current string `json:"current,omitempty"`
}

type UpdateStatus struct {
	Current       string        `json:"current"`
	Content       string        `json:"content,omitempty"`
	App           *UpdateOffer  `json:"app,omitempty"`
	ContentUpdate *ContentOffer `json:"contentUpdate,omitempty"`
	Checked       int64         `json:"checked"`
	Skipped       bool          `json:"skipped,omitempty"`
	Error         string        `json:"error,omitempty"`
	// 内容包里有坏文件时的提示：好的部分照用，坏的那几个说出来。
	Warning string `json:"warning,omitempty"`
}

// assetName 是当前平台在发布目录里对应的文件名。
func assetName(platform string) string {
	switch platform {
	case "android":
		return "GoStudy-android-release.apk"
	case "windows":
		return "GoStudy.exe"
	case "darwin":
		return ""
	default:
		if runtime.GOARCH == "arm64" {
			return "GoStudy-linux-arm64"
		}
		return "GoStudy-linux-amd64"
	}
}

const setupAssetName = "GoStudy-Setup.exe"

// downloadState 是一次下载任务的状态，前端按 400ms 轮询它。
type downloadState struct {
	State    string `json:"state"`
	Received int64  `json:"received"`
	Total    int64  `json:"total"`
	Asset    string `json:"asset,omitempty"`
	Error    string `json:"error,omitempty"`
	// ready 之后告诉界面下一步做什么：android 跳自定义 scheme，桌面直接重启。
	Mode string `json:"mode,omitempty"`
	// Kind 说明这次下载的是什么："app" 程序、"installer" 安装包、"content" 课程内容。
	// 有了它才不会把"内容包已就绪"当成"程序更新包已就绪"。
	Kind string `json:"kind,omitempty"`
}

const (
	kindApp       = "app"
	kindInstaller = "installer"
	kindContent   = "content"
)

type downloader struct {
	mu    sync.Mutex
	state downloadState
}

func (d *downloader) snapshot() downloadState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}

func (d *downloader) set(change func(*downloadState)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	change(&d.state)
}

func (d *downloader) busy() bool {
	state := d.snapshot().State
	return state == "downloading" || state == "verifying"
}

func fetchManifest(client *http.Client, bases []string, name string, target any) (string, error) {
	var lastErr error
	for _, base := range bases {
		body, err := fetchSmall(client, base+"/"+name)
		if err != nil {
			lastErr = err
			continue
		}
		if err := json.Unmarshal(body, target); err != nil {
			lastErr = fmt.Errorf("%s 内容无法解析", name)
			continue
		}
		return base, nil
	}
	if lastErr == nil {
		lastErr = errors.New("没有可用的更新源")
	}
	return "", lastErr
}

func fetchSmall(client *http.Client, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), manifestTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, errors.New("网络不可用，请稍后再试")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("更新清单不可用：HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxManifestBytes+1))
	if err != nil {
		return nil, errors.New("读取更新清单失败")
	}
	if len(body) > maxManifestBytes {
		return nil, errors.New("更新清单过大")
	}
	return body, nil
}

/*
fetchTo 下载一个文件并按清单校验：边下边算 sha256，大小与摘要都对得上才落到 target。
90 秒没有新数据就判定卡住并中止，避免界面上的进度条一直不动。
*/
func fetchTo(client *http.Client, url, target string, file updateFile, onProgress func(int64)) error {
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return errors.New("更新地址无效")
	}
	var stall struct {
		mu   sync.Mutex
		last time.Time
	}
	stall.last = time.Now()
	stalled := make(chan struct{})
	defer close(stalled)
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stalled:
				return
			case <-ticker.C:
				stall.mu.Lock()
				idle := time.Since(stall.last)
				stall.mu.Unlock()
				if idle > stallLimit {
					cancel()
					return
				}
			}
		}
	}()

	response, err := client.Do(request)
	if err != nil {
		return errors.New("下载失败，请检查网络后重试")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("下载失败：HTTP %d", response.StatusCode)
	}
	if response.ContentLength > maxAssetBytes {
		return errors.New("文件超过大小上限")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return errors.New("无法创建更新目录")
	}
	part := target + ".part"
	out, err := os.Create(part)
	if err != nil {
		return errors.New("无法写入更新文件")
	}
	defer func() { _ = os.Remove(part) }()

	digest := sha256.New()
	buffer := make([]byte, 256<<10)
	var written int64
	for {
		read, readErr := response.Body.Read(buffer)
		if read > 0 {
			if written+int64(read) > maxAssetBytes {
				out.Close()
				return errors.New("文件超过大小上限")
			}
			if _, err := out.Write(buffer[:read]); err != nil {
				out.Close()
				return errors.New("写入更新文件失败")
			}
			digest.Write(buffer[:read])
			written += int64(read)
			stall.mu.Lock()
			stall.last = time.Now()
			stall.mu.Unlock()
			if onProgress != nil {
				onProgress(written)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			out.Close()
			if ctx.Err() != nil {
				return errors.New("下载停滞或超时，已中止；请重试")
			}
			return errors.New("下载中断，请重试")
		}
	}
	if err := out.Close(); err != nil {
		return errors.New("写入更新文件失败")
	}
	if file.Size > 0 && written != file.Size {
		return errors.New("文件大小与清单不一致，已丢弃")
	}
	if file.SHA256 != "" {
		if got := hex.EncodeToString(digest.Sum(nil)); !strings.EqualFold(got, file.SHA256) {
			return errors.New("校验失败，已丢弃；请稍后重试")
		}
	}
	if err := os.Rename(part, target); err != nil {
		return errors.New("保存更新文件失败")
	}
	return nil
}

// checkUpdate 拉两个频道的清单并给出结论。auto 为 true 时按 24 小时节流。
func (e *Engine) checkUpdate(auto bool, prefer string) UpdateStatus {
	s := e.store.snapshot()
	status := UpdateStatus{Current: versionInfo(), Content: e.contentVersion(), Checked: s.Progress.LastUpdateCheck}
	if auto && time.Since(time.Unix(s.Progress.LastUpdateCheck, 0)) < autoCheckInterval {
		status.Skipped = true
		return status
	}

	var manifest releaseManifest
	_, err := fetchManifest(e.updateClient, preferred(latestBases(), prefer), latestManifest, &manifest)
	appOK := err == nil
	if appOK {
		asset := assetName(e.platform)
		if file, ok := manifest.Files[asset]; ok && asset != "" && versionNewer(manifest.Version, Version) {
			status.App = &UpdateOffer{Version: manifest.Version, Released: manifest.Released, Notes: manifest.Notes, Size: file.Size, Asset: asset}
		}
	} else {
		status.Error = err.Error()
	}

	var content contentInfo
	_, contentErr := fetchManifest(e.updateClient, preferred(contentBases(), prefer), contentManifest, &content)
	if contentErr == nil {
		if applied := e.contentVersion(); content.Version != "" && content.Version != applied {
			status.ContentUpdate = &ContentOffer{Version: content.Version, Size: content.Pack.Size, Notes: content.Notes, Current: applied}
		}
	} else if status.Error == "" {
		// 还没发过内容包时读不到不算故障，只留一句提示。
		status.Warning = "暂时读不到课程内容更新"
	}

	if appOK || contentErr == nil {
		now := time.Now().Unix()
		if _, updateErr := e.store.update(func(s *state) error { s.Progress.LastUpdateCheck = now; return nil }); updateErr == nil {
			status.Checked = now
		}
	}
	return status
}

// downloadUpdate 把当前平台的产物下到数据目录（后台进行，前端轮询进度）。
func (e *Engine) downloadUpdate(prefer string) (downloadState, error) {
	if current := e.downloads.snapshot(); e.downloads.busy() {
		return current, nil
	}
	// 已经下好同一个包就直接复用，别让用户重复下载十几兆。
	if current := e.downloads.snapshot(); current.State == "ready" && current.Kind == kindApp {
		return current, nil
	}
	var manifest releaseManifest
	base, err := fetchManifest(e.updateClient, preferred(latestBases(), prefer), latestManifest, &manifest)
	if err != nil {
		return downloadState{}, err
	}
	asset := assetName(e.platform)
	if asset == "" {
		return downloadState{}, errors.New("当前平台没有可下载的程序产物")
	}
	file, ok := manifest.Files[asset]
	if !ok {
		return downloadState{}, fmt.Errorf("发布清单里没有 %s", asset)
	}
	if !versionNewer(manifest.Version, Version) {
		return downloadState{}, errors.New("当前已经是最新版本")
	}
	target := filepath.Join(e.updatesDir(), asset)
	e.downloads.set(func(s *downloadState) {
		*s = downloadState{State: "downloading", Total: file.Size, Asset: asset, Kind: kindApp}
	})
	go func() {
		err := fetchTo(e.updateClient, base+"/"+asset, target, file, func(received int64) {
			e.downloads.set(func(s *downloadState) { s.Received = received })
		})
		if err != nil {
			_ = os.Remove(target)
			e.downloads.set(func(s *downloadState) { s.State = "failed"; s.Error = err.Error() })
			return
		}
		e.downloads.set(func(s *downloadState) {
			s.State = "ready"
			s.Received = file.Size
			s.Total = file.Size
			s.Mode = installMode(e.platform)
		})
	}()
	return e.downloads.snapshot(), nil
}

func installMode(platform string) string {
	if platform == "android" {
		return "android"
	}
	return "restart"
}

// installUpdate 用掉下好的包。桌面端要替换自身，所以先把响应发回去再动手。
func (e *Engine) installUpdate() (downloadState, error) {
	current := e.downloads.snapshot()
	if current.State != "ready" || current.Kind != kindApp {
		return current, errors.New("还没有下载好可用的更新包")
	}
	if e.platform == "android" {
		// Java 侧按固定约定去 <filesDir>/updates/GoStudy.apk 取包，页面只负责跳自定义地址。
		if _, err := os.Stat(filepath.Join(e.updatesDir(), assetName("android"))); err != nil {
			return current, errors.New("更新包不在了，请重新下载")
		}
		return current, nil
	}
	if e.platform != "windows" && e.platform != "linux" {
		return current, errors.New("当前平台请在发布页手动下载")
	}
	go func() {
		// 留出时间把响应发出去，再替换文件并退出。
		time.Sleep(400 * time.Millisecond)
		if err := e.applyUpdate(); err != nil {
			e.downloads.set(func(s *downloadState) { s.State = "failed"; s.Error = err.Error() })
			return
		}
		e.requestQuit()
	}()
	return current, nil
}

/*
applyUpdate 在 Windows 与 Linux 上"改名自己"来完成原地替换：
系统允许重命名正在运行的可执行文件，先让自己的名字空出来，再把新文件放上去，
最后启动新进程、由它删掉旧的。整个过程只有两次同目录改名，中间几乎没有空窗。
*/
func (e *Engine) applyUpdate() error {
	asset := assetName(e.platform)
	downloaded := filepath.Join(e.updatesDir(), asset)
	if _, err := os.Stat(downloaded); err != nil {
		return errors.New("更新包不在了，请重新下载")
	}
	executable, err := os.Executable()
	if err != nil {
		return errors.New("无法定位当前程序文件")
	}
	switch installStrategy(executable) {
	case "installer":
		// 安装版交给安装包升级，快捷方式、卸载信息与显示版本都跟着走。
		if e.platform == "windows" {
			closeAppWindows(e.host)
		}
		return e.runInstaller()
	case "manual":
		return errors.New("没有权限替换程序文件，请到发布页手动下载新版本")
	}
	dir, name := filepath.Dir(executable), filepath.Base(executable)
	staged := filepath.Join(dir, name+".new")
	old := filepath.Join(dir, name+".old")
	if err := copyFile(downloaded, staged); err != nil {
		return errors.New("无法写入程序目录，请到发布页手动下载新版本")
	}
	if err := os.Remove(old); err != nil && !os.IsNotExist(err) {
		// 上次留下的旧文件删不掉不影响这次更新，新进程启动时还会再删一次。
		_ = err
	}
	if err := os.Rename(executable, old); err != nil {
		_ = os.Remove(staged)
		return errors.New("无法替换正在运行的程序文件，请手动更新")
	}
	if err := os.Rename(staged, executable); err != nil {
		_ = os.Rename(old, executable)
		_ = os.Remove(staged)
		return errors.New("替换程序文件失败，已还原")
	}
	if e.platform == "windows" {
		// 先把旧的应用窗口关掉，否则会留下一个指向已经退出引擎的空白页。
		closeAppWindows(e.host)
	}
	command := exec.Command(executable, relaunchArgs(os.Args[1:], old)...)
	command.Dir = dir
	startDetached(command)
	return nil
}

// installStrategy 决定这次更新走哪条路：
//
//	installer —— 程序目录里有 Inno 的卸载器，说明是安装版，交给安装包升级
//	inplace   —— 免安装版，直接用改名替换
//	manual    —— 目录不可写，两者都做不了，让用户自己下载
func installStrategy(executable string) string {
	dir := filepath.Dir(executable)
	if _, err := os.Stat(filepath.Join(dir, "unins000.exe")); err == nil {
		return "installer"
	}
	if writableDir(dir) {
		return "inplace"
	}
	return "manual"
}

func writableDir(dir string) bool {
	probe, err := os.CreateTemp(dir, ".gostudy-write-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	probe.Close()
	_ = os.Remove(name)
	return true
}

// runInstaller 下载并用发布清单里的安装包静默升级；需要管理员权限时弹一次 UAC。
func (e *Engine) runInstaller() error {
	setup := filepath.Join(e.updatesDir(), setupAssetName)
	if _, err := os.Stat(setup); err != nil {
		if err := e.fetchInstaller(); err != nil {
			return err
		}
	}
	arguments := []string{"/VERYSILENT", "/SUPPRESSMSGBOXES", "/NORESTART", "/CLOSEAPPLICATIONS"}
	var command *exec.Cmd
	if e.platform != "windows" {
		return errors.New("当前平台请在发布页手动下载")
	}
	if writableDir(filepath.Dir(setup)) {
		command = exec.Command(setup, arguments...)
	} else {
		return errors.New("安装包无法运行，请到发布页手动下载")
	}
	if err := startDetached(command); err != nil {
		return errors.New("无法启动安装包，请到发布页手动下载")
	}
	return nil
}

// fetchInstaller 单独再下一个安装包（安装版升级用），同样按清单校验。
func (e *Engine) fetchInstaller() error {
	var manifest releaseManifest
	base, err := fetchManifest(e.updateClient, latestBases(), latestManifest, &manifest)
	if err != nil {
		return errors.New("拿不到发布清单，请检查网络后重试")
	}
	file, ok := manifest.Files[setupAssetName]
	if !ok {
		return errors.New("发布清单里没有安装包")
	}
	e.downloads.set(func(s *downloadState) {
		*s = downloadState{State: "downloading", Total: file.Size, Asset: setupAssetName}
	})
	err = fetchTo(e.updateClient, base+"/"+setupAssetName, filepath.Join(e.updatesDir(), setupAssetName), file, func(received int64) {
		e.downloads.set(func(s *downloadState) { s.Received = received })
	})
	if err != nil {
		e.downloads.set(func(s *downloadState) { s.State = "failed"; s.Error = err.Error() })
		return err
	}
	e.downloads.set(func(s *downloadState) { s.State = "ready"; s.Mode = "installer" })
	return nil
}

func copyFile(from, to string) error {
	source, err := os.Open(from)
	if err != nil {
		return err
	}
	defer source.Close()
	if err := os.MkdirAll(filepath.Dir(to), 0700); err != nil {
		return err
	}
	target, err := os.OpenFile(to, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	if _, err := io.Copy(target, source); err != nil {
		target.Close()
		return err
	}
	return target.Close()
}

/*
relaunchArgs 保留原来的参数，去掉会冲突或一次性的那几个：
新进程自己选端口，-diagnose / -selftest / -version 只在本次启动有意义，
另外把要删掉的旧文件路径通过 -cleanup 交给新进程。
*/
func relaunchArgs(args []string, oldPath string) []string {
	skip := map[string]bool{"-addr": true, "-cleanup": true, "-diagnose": true, "-selftest": true, "-version": true}
	out := []string{"-cleanup", oldPath}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name := arg
		if index := strings.IndexByte(arg, '='); index >= 0 {
			name = arg[:index]
			if skip[name] {
				continue
			}
			out = append(out, arg)
			continue
		}
		if skip[name] {
			// 形如 -addr 127.0.0.1:1 的参数，值也要一起跳过。
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
			}
			continue
		}
		out = append(out, arg)
	}
	return out
}

type updateRequest struct {
	Action string `json:"action"`
	// Prefer 是访客位置的提示（"cn" 优先 Gitee，其他优先 GitHub），由界面按语言与时区给出。
	Prefer string `json:"prefer"`
	// Auto 表示这是后台的定期检查，服务端按 24 小时节流。
	Auto bool `json:"auto"`
}

// handleUpdate 是版本与课程内容更新的唯一入口。
func (e *Engine) handleUpdate(w http.ResponseWriter, r *http.Request) {
	var req updateRequest
	if !decode(w, r, &req) {
		return
	}
	switch req.Action {
	case "status":
		writeJSON(w, 200, map[string]any{
			"current":  versionInfo(),
			"content":  e.contentVersion(),
			"download": e.downloads.snapshot(),
			"platform": e.platform,
			"channel":  assetName(e.platform),
		})
	case "check":
		writeJSON(w, 200, e.checkUpdate(req.Auto, req.Prefer))
	case "download":
		state, err := e.downloadUpdate(req.Prefer)
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, state)
	case "progress":
		writeJSON(w, 200, e.downloads.snapshot())
	case "install":
		state, err := e.installUpdate()
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, state)
	case "content":
		version, err := e.applyContent(req.Prefer)
		if err != nil {
			writeError(w, 400, err.Error())
			return
		}
		warning := ""
		if warnings := e.curriculumSnapshot().warnings; len(warnings) > 0 {
			warning = strings.Join(warnings, "；")
		}
		writeJSON(w, 200, map[string]any{"version": version, "content": e.contentVersion(), "warning": warning})
	default:
		writeError(w, 400, "未知的更新操作")
	}
}
