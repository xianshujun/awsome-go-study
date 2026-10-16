package backend

import (
	"errors"
	"net/http"
	"strings"
	"time"
)

/*
窗口关闭后的处理。

应用窗口是 Edge/Chrome 的应用模式窗口，用户点右上角的叉不会通知引擎，所以窗口关掉
之后进程会一直留着——任务管理器里就是那个"关不掉的 GoStudy.exe"。这里的做法是：
页面在关闭或刷新时用 sendBeacon 发一个 /api/window-closed 过来，引擎先等一小会儿
（刷新同样会触发 pagehide，新页面一发请求就撤销这次关闭），确认页面真的走了以后，
再按设置决定直接退出、留在后台静默运行，还是弹一次系统询问框。

判断"窗口还在不在"只认页面自己报的信，不做心跳轮询：窗口最小化、浏览器给后台页
降频定时器都会让心跳变慢，误判成"窗口已关闭"会把正在用的人踢出去。
*/
const (
	closeActionAsk        = "ask"
	closeActionBackground = "background"
	closeActionQuit       = "quit"

	// defaultCloseGrace 是页面关掉之后等待"页面是不是又回来了"的时间。
	defaultCloseGrace = 3 * time.Second
)

// normalizeCloseAction 把状态文件里读到的值收敛成三种已知行为，非法值按"询问"处理。
func normalizeCloseAction(raw string) string {
	switch raw {
	case closeActionBackground:
		return closeActionBackground
	case closeActionQuit:
		return closeActionQuit
	default:
		return closeActionAsk
	}
}

/*
CloseAnswer 是一次"关闭窗口后怎么办"询问的结果。
*/
type CloseAnswer struct {
	// Background 为真表示用户选了留在后台静默运行，否则直接退出。
	Background bool
	// Remember 为真表示用户勾了"不再提示"，这次选择写回设置。
	Remember bool
	// Answered 为假表示没能问出结果（对话框不可用），调用方按直接退出收尾。
	Answered bool
}

// pageActivity 记录页面还活着：任何带凭据的请求都算，并撤销待处理的关闭。
func (e *Engine) pageActivity() {
	e.pageMu.Lock()
	e.pageOpen = true
	timer := e.closeTimer
	e.closeTimer = nil
	e.pageMu.Unlock()
	if timer != nil {
		timer.Stop()
	}
}

// WindowOpen 报告应用窗口是否还开着，供单实例逻辑决定是复用已有进程还是重新开窗。
func (e *Engine) WindowOpen() bool {
	e.pageMu.Lock()
	defer e.pageMu.Unlock()
	return e.pageOpen
}

// windowClosed 处理页面发来的"窗口要关了"。
func (e *Engine) windowClosed() {
	// Android 外壳自己管生命周期：页面被销毁不等于用户要退出应用。
	if e.platform == "android" {
		return
	}
	e.pageMu.Lock()
	e.pageOpen = false
	if e.closeTimer != nil {
		e.closeTimer.Stop()
	}
	e.closeTimer = time.AfterFunc(e.closeGrace, e.resolveWindowClosed)
	e.pageMu.Unlock()
}

// resolveWindowClosed 在宽限期之后收尾：页面没回来就按设置处理。
func (e *Engine) resolveWindowClosed() {
	e.pageMu.Lock()
	open := e.pageOpen
	e.closeTimer = nil
	e.pageMu.Unlock()
	if open || e.quitting() {
		// 页面又回来了（刷新、浏览器恢复会话），或者更新/退出流程正在收尾。
		return
	}
	switch normalizeCloseAction(e.store.snapshot().Settings.CloseAction) {
	case closeActionBackground:
		e.log("窗口已关闭，按设置继续在后台静默运行")
	case closeActionQuit:
		e.log("窗口已关闭，按设置直接退出")
		e.requestQuit()
	default:
		e.askAndClose()
	}
}

// askAndClose 弹出询问框；问不出来（对话框不可用）时按直接退出处理，绝不留后台进程。
func (e *Engine) askAndClose() {
	answer := e.askClose()
	if !answer.Answered {
		e.log("没能弹出关闭方式询问框，直接退出")
		e.requestQuit()
		return
	}
	if answer.Remember {
		action := closeActionQuit
		if answer.Background {
			action = closeActionBackground
		}
		if _, err := e.store.update(func(s *state) error { s.Settings.CloseAction = action; return nil }); err != nil {
			e.log("保存关闭方式失败：%v", err)
		}
	}
	if answer.Background {
		e.log("窗口已关闭，窗口提示里选择继续在后台静默运行")
		return
	}
	e.log("窗口已关闭，窗口提示里选择直接退出")
	e.requestQuit()
}

// quitting 报告是否已经在退出流程里。
func (e *Engine) quitting() bool {
	select {
	case <-e.quit:
		return true
	default:
		return false
	}
}

// log 把引擎里的关键动作写进启动日志；没接日志就什么都不做。
func (e *Engine) log(format string, args ...any) {
	if e.logf == nil {
		return
	}
	e.logf(format, args...)
}

// setCloseAction 保存设置里的关闭行为，只接受三种已知取值。
func (e *Engine) setCloseAction(raw string) (storedSettings, error) {
	switch raw {
	case closeActionAsk, closeActionBackground, closeActionQuit:
	default:
		return storedSettings{}, errCloseAction
	}
	s, err := e.store.update(func(s *state) error { s.Settings.CloseAction = raw; return nil })
	if err != nil {
		return storedSettings{}, err
	}
	return s.Settings, nil
}

/*
handleCloseBehavior 保存"关闭窗口后怎么办"。

设置页里改的、询问框里勾"不再提示"记下的，是同一项设置：都写在这份状态文件里，
窗口已经关掉之后引擎还能读到它。
*/
func (e *Engine) handleCloseBehavior(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Action string `json:"action"`
	}
	if !decode(w, r, &req) {
		return
	}
	settings, err := e.setCloseAction(strings.TrimSpace(req.Action))
	if err != nil {
		if errors.Is(err, errCloseAction) {
			writeError(w, 400, err.Error())
			return
		}
		writeError(w, 500, "保存关闭方式失败")
		return
	}
	writeJSON(w, 200, publicSettings(settings))
}

var errCloseAction = errors.New("关闭方式只能是 ask、background 或 quit")

// closeState 只为测试与诊断准备的一份快照：页面是否在线、有没有在等宽限期结束。
type closeState struct {
	pageOpen bool
	pending  bool
}

func (e *Engine) closeSnapshot() closeState {
	e.pageMu.Lock()
	defer e.pageMu.Unlock()
	return closeState{pageOpen: e.pageOpen, pending: e.closeTimer != nil}
}
