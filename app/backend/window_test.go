package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func closeEngine(t *testing.T, platform string, grace time.Duration, ask func() CloseAnswer) *Engine {
	t.Helper()
	e, err := New(Options{
		DataDir:    t.TempDir(),
		Platform:   platform,
		CloseGrace: grace,
		AskClose:   ask,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(e.aiClient.CloseIdleConnections)
	return e
}

func waitQuit(e *Engine, wait time.Duration) bool {
	select {
	case <-e.Quit():
		return true
	case <-time.After(wait):
		return false
	}
}

// 选"直接退出"：窗口关掉之后进程必须结束，这是"关不掉的进程"的正面用例。
func TestWindowCloseAsksAndQuits(t *testing.T) {
	asked := 0
	e := closeEngine(t, "", 30*time.Millisecond, func() CloseAnswer {
		asked++
		return CloseAnswer{Answered: true}
	})
	e.windowClosed()
	if !waitQuit(e, 2*time.Second) {
		t.Fatal("窗口关闭后没有退出")
	}
	if asked != 1 {
		t.Fatalf("应该问一次，实际 %d 次", asked)
	}
}

// 选"后台静默运行"并勾了不再提示：进程留着，选择写回设置，页面状态标记为已关闭。
func TestWindowCloseBackgroundRemembers(t *testing.T) {
	e := closeEngine(t, "", 30*time.Millisecond, func() CloseAnswer {
		return CloseAnswer{Answered: true, Background: true, Remember: true}
	})
	e.windowClosed()
	if waitQuit(e, 300*time.Millisecond) {
		t.Fatal("选了后台静默运行，进程不该退出")
	}
	if got := e.store.snapshot().Settings.CloseAction; got != closeActionBackground {
		t.Fatalf("记住的关闭方式 = %q", got)
	}
	if state := e.closeSnapshot(); state.pageOpen || state.pending {
		t.Fatalf("窗口已经关掉，状态应为已关闭且没有待处理计时器：%+v", state)
	}
}

// 设置里已经定好了就直接照做，不再打扰用户。
func TestWindowCloseHonoursSetting(t *testing.T) {
	quit := closeEngine(t, "", 20*time.Millisecond, func() CloseAnswer {
		t.Error("设置成直接退出时不该再弹询问框")
		return CloseAnswer{}
	})
	if _, err := quit.setCloseAction(closeActionQuit); err != nil {
		t.Fatal(err)
	}
	quit.windowClosed()
	if !waitQuit(quit, time.Second) {
		t.Fatal("设置成直接退出却没有退出")
	}

	background := closeEngine(t, "", 20*time.Millisecond, func() CloseAnswer {
		t.Error("设置成后台静默时不该再弹询问框")
		return CloseAnswer{}
	})
	if _, err := background.setCloseAction(closeActionBackground); err != nil {
		t.Fatal(err)
	}
	background.windowClosed()
	if waitQuit(background, 300*time.Millisecond) {
		t.Fatal("设置成后台静默却退出了")
	}
}

// 询问框弹不出来时必须按直接退出收尾：问不出来绝不能变成"进程留在后台"。
func TestWindowCloseWithoutPromptQuits(t *testing.T) {
	e := closeEngine(t, "", 20*time.Millisecond, func() CloseAnswer { return CloseAnswer{} })
	e.windowClosed()
	if !waitQuit(e, time.Second) {
		t.Fatal("没能询问时应当直接退出")
	}
}

// 刷新页面也会触发 pagehide，新页面一发请求就要撤销这次关闭。
func TestPageActivityCancelsPendingClose(t *testing.T) {
	e := closeEngine(t, "", 100*time.Millisecond, func() CloseAnswer {
		t.Error("刷新页面不该触发关闭询问")
		return CloseAnswer{}
	})
	e.windowClosed()
	time.Sleep(20 * time.Millisecond)
	e.pageActivity()
	if waitQuit(e, 400*time.Millisecond) {
		t.Fatal("页面又回来了，不该退出")
	}
}

// Android 外壳自己管生命周期：页面被销毁不等于用户要退出应用。
func TestAndroidIgnoresWindowClose(t *testing.T) {
	e := closeEngine(t, "android", 20*time.Millisecond, func() CloseAnswer {
		t.Error("Android 上不该弹 Windows 的询问框")
		return CloseAnswer{}
	})
	e.windowClosed()
	if waitQuit(e, 300*time.Millisecond) {
		t.Fatal("Android 上不该因为页面消失而退出")
	}
}

// 用户从设置里退出时，正在等宽限期的窗口关闭流程不该再插一手。
func TestWindowCloseAfterQuitRequest(t *testing.T) {
	e := closeEngine(t, "", 20*time.Millisecond, func() CloseAnswer {
		t.Error("已经在退出流程里，不该再弹询问框")
		return CloseAnswer{}
	})
	e.requestQuit()
	e.windowClosed()
	time.Sleep(150 * time.Millisecond)
}

func postRaw(t *testing.T, e *Engine, method, target, contentType, token string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, "http://127.0.0.1:12345"+target, nil)
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	record := httptest.NewRecorder()
	e.ServeHTTP(record, request)
	return record
}

// 页面关闭时用的是 sendBeacon：带不了请求头，只能把凭据放在查询串里；
// 这个口子只开给这一个端点。
func TestWindowClosedEndpointAcceptsBeacon(t *testing.T) {
	e := closeEngine(t, "", 30*time.Millisecond, func() CloseAnswer { return CloseAnswer{Answered: true} })
	// 不带 Content-Type、不带 Authorization，只有查询串凭据。
	if w := postRaw(t, e, http.MethodPost, "/api/window-closed?token="+e.Token(), "", ""); w.Code != 200 {
		t.Fatalf("sendBeacon 形式的报信应当被接受：%d %s", w.Code, w.Body.String())
	}
	// 凭据不对、或者换成别的接口，查询串里的凭据一律不认。
	if w := postRaw(t, e, http.MethodPost, "/api/window-closed?token=wrong", "", ""); w.Code != 401 {
		t.Fatalf("凭据不对应当 401，实际 %d", w.Code)
	}
	if w := postRaw(t, e, http.MethodPost, "/api/quit?token="+e.Token(), "application/json", ""); w.Code != 401 {
		t.Fatalf("查询串凭据不该对别的接口生效，实际 %d", w.Code)
	}
	// 只有 POST 这一种形式认查询串凭据；GET 连凭据都对不上，更不会被执行。
	if w := postRaw(t, e, http.MethodGet, "/api/window-closed?token="+e.Token(), "", ""); w.Code != 401 {
		t.Fatalf("查询串凭据只对 POST 生效，实际 %d", w.Code)
	}
	if w := postRaw(t, e, http.MethodGet, "/api/window-closed", "", e.Token()); w.Code != 405 {
		t.Fatalf("报信只接受 POST，实际 %d", w.Code)
	}
	if !waitQuit(e, 2*time.Second) {
		t.Fatal("收到报信之后应当按设置退出")
	}
}

// 页面回来（刷新、内容更新后重载）时，收到报信之后紧接着的 API 请求要能救回来。
func TestAPIRequestCancelsPendingClose(t *testing.T) {
	e := closeEngine(t, "", 60*time.Millisecond, func() CloseAnswer {
		t.Error("页面又发请求了，不该弹出关闭询问")
		return CloseAnswer{}
	})
	if w := postRaw(t, e, http.MethodPost, "/api/window-closed?token="+e.Token(), "", ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if w := callAPI(t, e, "/api/bootstrap", nil); w.Code != 200 {
		t.Fatal(w.Code)
	}
	if waitQuit(e, 300*time.Millisecond) {
		t.Fatal("页面已经回来了，不该退出")
	}
}

func TestCloseBehaviorEndpoint(t *testing.T) {
	e := testEngine(t, t.TempDir())
	if w := callAPI(t, e, "/api/close-behavior", map[string]string{"action": closeActionQuit}); w.Code != 200 {
		t.Fatalf("保存关闭方式失败：%d %s", w.Code, w.Body.String())
	}
	var settings Settings
	if err := json.Unmarshal(callAPI(t, e, "/api/bootstrap", nil).Body.Bytes(), &struct {
		Settings *Settings `json:"settings"`
	}{Settings: &settings}); err != nil {
		t.Fatal(err)
	}
	if settings.CloseAction != closeActionQuit {
		t.Fatalf("bootstrap 里的关闭方式 = %q", settings.CloseAction)
	}
	if w := callAPI(t, e, "/api/close-behavior", map[string]string{"action": "nonsense"}); w.Code != 400 {
		t.Fatalf("非法取值应当 400，实际 %d %s", w.Code, w.Body.String())
	}
	if w := callAPI(t, e, "/api/close-behavior", map[string]string{"action": closeActionAsk}); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
}

// /api/status 是启动器探活用的小接口：要能报告窗口还开着没有。
func TestStatusReportsWindow(t *testing.T) {
	e := testEngine(t, t.TempDir())
	var status struct {
		Status     string `json:"status"`
		WindowOpen bool   `json:"windowOpen"`
	}
	if err := json.Unmarshal(callAPI(t, e, "/api/status", nil).Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	// 启动器紧接着就会打开应用窗口，所以在收到"页面要走了"之前都算窗口开着：
	// 这样用户连点两下快捷键时，第二个进程不会误以为窗口关着又开一个。
	if status.Status != "ok" || !status.WindowOpen {
		t.Fatalf("启动时应当按窗口开着记账：%+v", status)
	}
	e.windowClosed()
	if err := json.Unmarshal(callAPI(t, e, "/api/status", nil).Body.Bytes(), &status); err != nil {
		t.Fatal(err)
	}
	if status.WindowOpen {
		t.Fatal("页面已经报过信，应当报告窗口已关闭")
	}
}

// 免安装版没有卸载程序：接口要给出能照做的提示，而不是假装卸载。
func TestUninstallWithoutInstaller(t *testing.T) {
	e := testEngine(t, t.TempDir())
	w := callAPI(t, e, "/api/uninstall", map[string]any{})
	if w.Code != 409 {
		t.Fatalf("免安装版应当明确拒绝，实际 %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "免安装版") {
		t.Fatalf("提示里应当说明是免安装版：%s", w.Body.String())
	}
}

func TestFindUninstallerIn(t *testing.T) {
	dir := t.TempDir()
	if got := findUninstallerIn(dir); got != "" {
		t.Fatalf("空目录不该找到卸载程序：%q", got)
	}
	if err := os.WriteFile(filepath.Join(dir, "unins001.exe"), []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	if got := findUninstallerIn(dir); got != filepath.Join(dir, "unins001.exe") {
		t.Fatalf("应当找到 unins001.exe，实际 %q", got)
	}
	if got := findUninstaller("linux"); got != "" {
		t.Fatalf("非 Windows 平台不该找卸载程序：%q", got)
	}
	// 其它平台的构建里不会有卸载入口。
	engine := testEngine(t, t.TempDir())
	if engine.uninstaller != "" {
		t.Fatalf("测试环境里不该有卸载程序：%q", engine.uninstaller)
	}
}
