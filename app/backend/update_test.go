package backend

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestVersionComparison(t *testing.T) {
	for _, tt := range []struct {
		candidate, current string
		newer              bool
	}{
		{"0.7.0", "0.6.0", true},
		{"v0.7.0", "0.6.0", true},
		{"0.6.1", "0.6.0", true},
		{"1.0.0", "0.9.9", true},
		{"0.6.0", "0.6.0", false},
		{"0.5.9", "0.6.0", false},
		{"dev", "0.6.0", false},
		{"0.7.0", "dev", false},
		{"0.7", "0.6.0", false},
		{"0.7.0-beta", "0.6.0", false},
	} {
		if got := versionNewer(tt.candidate, tt.current); got != tt.newer {
			t.Errorf("versionNewer(%q, %q) = %v，期望 %v", tt.candidate, tt.current, got, tt.newer)
		}
	}
}

// releaseServer 起一个只提供静态文件的"假发布源"，用来验证检查、下载与校验的完整链路。
type releaseServer struct {
	*httptest.Server
	files map[string][]byte
}

func newReleaseServer(t *testing.T) *releaseServer {
	t.Helper()
	server := &releaseServer{files: map[string][]byte{}}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		body, ok := server.files[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	return server
}

func digestOf(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

// withVersion 把包级版本号临时设成指定值：测试里的引擎也要有版本才谈得上"比较"。
func withVersion(t *testing.T, version string) {
	t.Helper()
	previous := Version
	Version = version
	t.Cleanup(func() { Version = previous })
}

func TestUpdateCheckDownloadAndVerify(t *testing.T) {
	server := newReleaseServer(t)
	asset := "GoStudy-linux-amd64"
	payload := bytes.Repeat([]byte("GoStudy 更新包"), 512)
	server.files[asset] = payload
	manifest := releaseManifest{
		Version:  "9.9.9",
		Released: "2026-10-16",
		Notes:    "测试版本",
		Files: map[string]updateFile{
			asset: {Size: int64(len(payload)), SHA256: digestOf(payload)},
		},
	}
	encoded, _ := json.Marshal(manifest)
	server.files[latestManifest] = encoded

	dir := t.TempDir()
	withVersion(t, "0.6.0")
	e := testEngine(t, dir)
	t.Setenv("GOSTUDY_UPDATE_BASES", server.URL)
	t.Setenv("GOSTUDY_CONTENT_BASES", server.URL)

	status := e.checkUpdate(false, "cn")
	if status.App == nil || status.App.Version != "9.9.9" || status.App.Asset != asset {
		t.Fatalf("应当发现新版本：%+v", status)
	}
	if status.App.Size != int64(len(payload)) {
		t.Fatal(status.App.Size)
	}
	if status.Checked == 0 {
		t.Fatal("检查成功后应记录时间")
	}

	// 同一个引擎再检查，时间戳应当更新而不是重来
	if again := e.checkUpdate(false, "cn"); again.Checked < status.Checked {
		t.Fatal("检查时间应单调递增")
	}

	state, err := e.downloadUpdate("cn")
	if err != nil {
		t.Fatal(err)
	}
	if state.State != "downloading" && state.State != "ready" {
		t.Fatal(state)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if got := e.downloads.snapshot(); got.State == "ready" || got.State == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	final := e.downloads.snapshot()
	if final.State != "ready" {
		t.Fatalf("下载应当成功：%+v", final)
	}
	if final.Received != int64(len(payload)) || final.Total != int64(len(payload)) {
		t.Fatalf("进度不对：%+v", final)
	}
	saved, err := os.ReadFile(filepath.Join(dir, "updates", asset))
	if err != nil || !bytes.Equal(saved, payload) {
		t.Fatal("下载的文件内容不对", err)
	}
	if final.Mode != "restart" {
		t.Fatalf("桌面端应当提示重启：%+v", final)
	}
}

// 先更新课程内容、再更新程序：两个下载用同一个状态机，但不能互相认错。
func TestContentThenAppDownload(t *testing.T) {
	server := newReleaseServer(t)
	asset := "GoStudy-linux-amd64"
	payload := []byte("新的程序")
	server.files[asset] = payload
	manifest := releaseManifest{
		Version: "9.9.9",
		Files:   map[string]updateFile{asset: {Size: int64(len(payload)), SHA256: digestOf(payload)}},
	}
	encoded, _ := json.Marshal(manifest)
	server.files[latestManifest] = encoded
	courses := t.TempDir()
	writeCourseDir(t, courses, "内容包版")
	publishContent(t, server, courses, "content-1")

	withVersion(t, "0.6.0")
	e := testEngine(t, t.TempDir())
	t.Setenv("GOSTUDY_UPDATE_BASES", server.URL)
	t.Setenv("GOSTUDY_CONTENT_BASES", server.URL)

	if _, err := e.applyContent("cn"); err != nil {
		t.Fatal(err)
	}
	// 内容包下载完会留下 ready 状态，程序更新不能被它挡住
	if state := e.downloads.snapshot(); state.Kind != kindContent {
		t.Fatalf("内容下载应标记为 content：%+v", state)
	}
	if _, err := e.downloadUpdate("cn"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if got := e.downloads.snapshot(); got.State == "ready" || got.State == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	final := e.downloads.snapshot()
	if final.State != "ready" || final.Kind != kindApp || final.Asset != asset {
		t.Fatalf("程序更新包应当真的被下载：%+v", final)
	}
	if _, err := os.Stat(filepath.Join(e.updatesDir(), asset)); err != nil {
		t.Fatal("程序更新包没有落盘", err)
	}
}

func TestUpdateDownloadRejectsBadChecksum(t *testing.T) {
	server := newReleaseServer(t)
	asset := "GoStudy-linux-amd64"
	payload := []byte("真正的更新包")
	server.files[asset] = payload
	manifest := releaseManifest{
		Version: "9.9.9",
		Files: map[string]updateFile{
			// 大小对得上、摘要对不上：专门验证校验这一关
			asset: {Size: int64(len(payload)), SHA256: digestOf([]byte("别的文件"))},
		},
	}
	encoded, _ := json.Marshal(manifest)
	server.files[latestManifest] = encoded

	dir := t.TempDir()
	withVersion(t, "0.6.0")
	e := testEngine(t, dir)
	t.Setenv("GOSTUDY_UPDATE_BASES", server.URL)

	if _, err := e.downloadUpdate("cn"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if got := e.downloads.snapshot(); got.State == "ready" || got.State == "failed" {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	final := e.downloads.snapshot()
	if final.State != "failed" || !strings.Contains(final.Error, "校验失败") {
		t.Fatalf("校验失败必须报错并丢弃：%+v", final)
	}
	if _, err := os.Stat(filepath.Join(dir, "updates", asset)); !os.IsNotExist(err) {
		t.Fatal("校验不通过的文件不应留在磁盘上")
	}
}

func TestUpdateAutoCheckThrottles(t *testing.T) {
	server := newReleaseServer(t)
	manifest := releaseManifest{Version: "9.9.9", Files: map[string]updateFile{}}
	encoded, _ := json.Marshal(manifest)
	server.files[latestManifest] = encoded

	e := testEngine(t, t.TempDir())
	t.Setenv("GOSTUDY_UPDATE_BASES", server.URL)
	t.Setenv("GOSTUDY_CONTENT_BASES", server.URL)

	if first := e.checkUpdate(false, "cn"); first.Skipped {
		t.Fatal("手动检查不该被节流")
	}
	if second := e.checkUpdate(true, "cn"); !second.Skipped {
		t.Fatal("24 小时内的自动检查应当跳过")
	}
	if manual := e.checkUpdate(false, "cn"); manual.Skipped {
		t.Fatal("手动检查始终可执行")
	}
}

// writeCourseDir 把内置课程里的一门课改写成指定标题，作为内容包素材。
func writeCourseDir(t *testing.T, dir, title string) {
	t.Helper()
	data, err := curriculumFS.ReadFile("courses/grammar.json")
	if err != nil {
		t.Fatal(err)
	}
	var file map[string]any
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatal(err)
	}
	course := file["course"].(map[string]any)
	course["title"] = title
	updated, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "grammar.json"), updated, 0600); err != nil {
		t.Fatal(err)
	}
}

// publishContent 用课程目录造一个内容包并挂到假发布源上。
func publishContent(t *testing.T, server *releaseServer, dir, version string) {
	t.Helper()
	pack, files, err := contentPackFromDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	info := contentInfo{
		Version: version,
		File:    contentPack,
		Count:   len(files),
		Pack:    updateFile{Size: int64(len(pack)), SHA256: digestOf(pack)},
	}
	encoded, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	server.files[contentManifest] = encoded
	server.files[contentPack] = pack
}

func TestContentUpdateAppliesAndSurvivesRestart(t *testing.T) {
	server := newReleaseServer(t)
	courses := t.TempDir()
	const newTitle = "Go 语言基础 · 语法（内容包版）"
	writeCourseDir(t, courses, newTitle)
	publishContent(t, server, courses, "2026-10-16-abc123")

	dir := t.TempDir()
	withVersion(t, "0.6.0")
	e := testEngine(t, dir)
	t.Setenv("GOSTUDY_UPDATE_BASES", server.URL)
	t.Setenv("GOSTUDY_CONTENT_BASES", server.URL)

	status := e.checkUpdate(false, "cn")
	if status.ContentUpdate == nil || status.ContentUpdate.Version != "2026-10-16-abc123" || status.ContentUpdate.Current != "" {
		t.Fatalf("应当发现内容更新：%+v", status)
	}

	version, err := e.applyContent("cn")
	if err != nil {
		t.Fatal(err)
	}
	if version != "2026-10-16-abc123" {
		t.Fatal(version)
	}
	// 内存里的课程已经换成新版
	var title string
	for _, course := range e.curriculumSnapshot().courses {
		if course.ID == "grammar" {
			title = course.Title
		}
	}
	if title != newTitle {
		t.Fatalf("课程标题应更新为 %q，实际 %q", newTitle, title)
	}
	// 落盘：重启后依然是新的
	e2 := testEngine(t, dir)
	for _, course := range e2.curriculumSnapshot().courses {
		if course.ID == "grammar" && course.Title != newTitle {
			t.Fatalf("重启后内容没有保留：%q", course.Title)
		}
	}
	if e2.contentVersion() != "2026-10-16-abc123" {
		t.Fatal(e2.contentVersion())
	}
	// 已经是这个版本：再检查就不再提示
	if again := e2.checkUpdate(false, "cn"); again.ContentUpdate != nil {
		t.Fatalf("同版本内容不该再提示：%+v", again.ContentUpdate)
	}
	// 幂等：重复应用不报错
	if _, err := e2.applyContent("cn"); err != nil {
		t.Fatal(err)
	}
}

func TestContentPackRejectsBadInput(t *testing.T) {
	e := testEngine(t, t.TempDir())
	dir := e.contentDir()

	// 坏 JSON：结构上是一个合法的包，但内容不合法，落盘时必须被拦下
	badJSON := writeZip(t, map[string][]byte{"grammar.json": []byte("{ 不是 JSON")})
	if _, err := readContentPack(badJSON); err != nil {
		t.Fatalf("结构校验应当先通过：%v", err)
	}
	if err := e.installContentPack(badJSON, contentInfo{Version: "broken"}); err == nil {
		t.Fatal("坏 JSON 不该被应用")
	}
	if _, err := os.Stat(filepath.Join(dir, "grammar.json")); !os.IsNotExist(err) {
		t.Fatal("坏内容包不应留下文件")
	}
	// 目录穿越与多余文件：一律拒绝
	for name, payload := range map[string][]byte{
		"../evil.json": []byte(`{}`),
		"a/b.json":     []byte(`{}`),
		"readme.txt":   []byte("hello"),
		"grammar.JSON": []byte(`{}`),
	} {
		if _, err := readContentPack(writeZip(t, map[string][]byte{name: payload})); err == nil {
			t.Fatalf("%s 不该被接受", name)
		}
	}
	// 结构合法但课程内容不合法（小节数量不足）：也不该应用
	bad := map[string][]byte{"grammar.json": []byte(`{"order":1,"course":{"id":"grammar","title":"x","description":"y","level":"基础","duration":"1 分钟"},"lessons":[]}`)}
	pack, err := readContentPack(writeZip(t, bad))
	if err != nil {
		t.Fatal(err)
	}
	if err := e.installContentPack(writeZip(t, map[string][]byte{"grammar.json": bad["grammar.json"]}), contentInfo{Version: "broken"}); err == nil {
		t.Fatal("课程内容不合法时不该落盘")
	}
	if _, err := os.Stat(filepath.Join(dir, "grammar.json")); !os.IsNotExist(err) {
		t.Fatal("失败的内容包不应留下任何文件")
	}
	if len(pack) != 1 {
		t.Fatal("结构校验应当先通过")
	}
}

// writeZip 把若干文件打成一个临时 zip，返回路径。
func writeZip(t *testing.T, files map[string][]byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "content.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, payload := range files {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(payload); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

// Android 上不能自己替换程序：安装这一步只是把包准备好，由外壳拉起系统安装器。
func TestAndroidInstallHandsOffToSystem(t *testing.T) {
	dir := t.TempDir()
	withVersion(t, "0.6.0")
	e, err := New(Options{DataDir: dir, RunTimeout: 100 * time.Millisecond, AITimeout: 200 * time.Millisecond, Platform: "android"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.aiClient.CloseIdleConnections)

	if got := assetName("android"); got != "GoStudy-android-release.apk" {
		t.Fatalf("Android 的产物名不对：%s", got)
	}
	// 还没下载就点安装：应当明确报错，而不是去动程序文件
	if _, err := e.installUpdate(); err == nil || !strings.Contains(err.Error(), "还没有下载好") {
		t.Fatalf("未下载时应当报错：%v", err)
	}
	// 把包放到约定位置，安装步骤应当放行并告诉界面去跳自定义地址
	if err := os.MkdirAll(e.updatesDir(), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(e.updatesDir(), assetName("android")), []byte("apk"), 0600); err != nil {
		t.Fatal(err)
	}
	e.downloads.set(func(s *downloadState) { *s = downloadState{State: "ready", Kind: kindApp, Mode: "android"} })
	state, err := e.installUpdate()
	if err != nil {
		t.Fatal(err)
	}
	if state.Mode != "android" {
		t.Fatalf("界面需要知道该跳 gostudy://install：%+v", state)
	}
	if installMode("android") != "android" {
		t.Fatal("installMode 应当区分 Android")
	}
}

func TestRelaunchArgs(t *testing.T) {
	got := relaunchArgs([]string{"-no-browser", "-addr", "127.0.0.1:7731", "-data-dir", "/tmp/x", "-diagnose"}, "/tmp/GoStudy.exe.old")
	want := []string{"-cleanup", "/tmp/GoStudy.exe.old", "-no-browser", "-data-dir", "/tmp/x"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("重启参数不对：%v", got)
	}
	// -addr=值 的写法也要去掉
	got = relaunchArgs([]string{"-addr=127.0.0.1:7731"}, "old")
	if strings.Join(got, " ") != "-cleanup old" {
		t.Fatalf("等号写法没处理：%v", got)
	}
}

func TestInstallStrategyFallsBackToManual(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "GoStudy")
	if err := os.WriteFile(exe, []byte("binary"), 0700); err != nil {
		t.Fatal(err)
	}
	if got := installStrategy(exe); got != "inplace" {
		t.Fatalf("可写目录里的免安装版应当原地替换：%s", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "unins000.exe"), []byte("uninstaller"), 0700); err != nil {
		t.Fatal(err)
	}
	if got := installStrategy(exe); got != "installer" {
		t.Fatalf("有卸载器的安装版应当走安装包：%s", got)
	}
}
