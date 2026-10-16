//go:build windows

package backend

import (
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"unsafe"
)

/*
关闭窗口后的询问框。

系统消息框没有"不再提示"勾选框，而这里一次要问清楚两件事：留在后台静默运行还是
直接退出、以后还要不要再问。所以自己建一个小窗口：一段说明 + 一个勾选框 + 两个按钮，
控件全部用 user32 现成的 STATIC / BUTTON 类。

两条硬约束：
  - 窗口消息循环必须固定在同一个系统线程上，所以这里 LockOSThread，窗口过程用
    syscall.NewCallback 注册，每次询问自己跑自己的消息循环。
  - 任何一步失败都返回 Answered=false，由调用方退回系统消息框或直接退出。
    "问不出来"绝不能变成"进程留在后台"。
*/
const (
	wmDestroy        = 0x0002
	wmClose          = 0x0010
	wmKeyDown        = 0x0100
	wmCommand        = 0x0111
	wmNCCreate       = 0x0081
	wmCtlColorStatic = 0x0138

	wsChild         = 0x40000000
	wsVisible       = 0x10000000
	wsOverlapped    = 0x00000000
	wsCaption       = 0x00C00000
	wsSysMenu       = 0x00080000
	wsExTopmost     = 0x00000008
	bsPushButton    = 0x00000000
	bsDefPushButton = 0x00000001
	bsAutoCheckBox  = 0x00000003
	ssLeft          = 0x00000000
	bmGetCheck      = 0x00F0
	bstChecked      = 0x0001
	wmSetFont       = 0x0030
	vkEscape        = 0x1B
	colorWindow     = 5
	smCxScreen      = 0
	smCyScreen      = 1
	idcArrow        = 32512
)

const (
	closePromptBackgroundID = 1
	closePromptQuitID       = 2
)

var (
	user32dll   = syscall.NewLazyDLL("user32.dll")
	gdi32dll    = syscall.NewLazyDLL("gdi32.dll")
	kernel32dll = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW             = user32dll.NewProc("RegisterClassExW")
	procCreateWindowExW              = user32dll.NewProc("CreateWindowExW")
	procDefWindowProcW               = user32dll.NewProc("DefWindowProcW")
	procDestroyWindow                = user32dll.NewProc("DestroyWindow")
	procShowWindow                   = user32dll.NewProc("ShowWindow")
	procGetMessageW                  = user32dll.NewProc("GetMessageW")
	procTranslateMessage             = user32dll.NewProc("TranslateMessage")
	procDispatchMessageW             = user32dll.NewProc("DispatchMessageW")
	procPostQuitMessage              = user32dll.NewProc("PostQuitMessage")
	procAdjustWindowRectEx           = user32dll.NewProc("AdjustWindowRectEx")
	procSendMessageW                 = user32dll.NewProc("SendMessageW")
	procSetForegroundWindow          = user32dll.NewProc("SetForegroundWindow")
	procGetSystemMetrics             = user32dll.NewProc("GetSystemMetrics")
	procGetSysColorBrush             = user32dll.NewProc("GetSysColorBrush")
	procSetThreadDpiAwarenessContext = user32dll.NewProc("SetThreadDpiAwarenessContext")
	procGetDpiForSystem              = user32dll.NewProc("GetDpiForSystem")
	procMessageBoxW                  = user32dll.NewProc("MessageBoxW")

	procCreateFontW  = gdi32dll.NewProc("CreateFontW")
	procDeleteObject = gdi32dll.NewProc("DeleteObject")

	// GetModuleHandleW 在 kernel32 里：写错 DLL 会让 LazyProc.Call 直接 panic。
	procGetModuleHandleW = kernel32dll.NewProc("GetModuleHandleW")

	// 窗口过程要一直可达，否则可能被当成垃圾回收掉。
	closePromptProc = syscall.NewCallback(closePromptWndProc)
	// promptGate 保证同一时刻只弹一个框：两个线程各跑一个消息循环会互相抢焦点。
	promptGate sync.Mutex
	// promptBox 是当前这个框的答案结构。
	// 同一时刻只有一个框（promptGate 保证），所以用包级变量就够；
	// 窗口过程由 Windows 在我们的消息循环里回调，用原子指针交接才不会有数据竞争。
	promptBox atomic.Pointer[closePromptBox]
)

type wndClassExW struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msgW struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      struct{ x, y int32 }
}

type rectW struct{ left, top, right, bottom int32 }

// closePromptBox 是一次询问框的答案；只由窗口过程写、弹框函数读。
type closePromptBox struct {
	mu         sync.Mutex
	background bool
	remember   bool
	done       bool
	check      uintptr
}

func (b *closePromptBox) choose(background, remember bool) {
	b.mu.Lock()
	b.background, b.remember, b.done = background, remember, true
	b.mu.Unlock()
}

func (b *closePromptBox) answer() (bool, bool, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.background, b.remember, b.done
}

func closePromptWndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	box := promptBox.Load()
	switch message {
	case wmCommand:
		if box == nil {
			break
		}
		switch uint16(wParam & 0xffff) {
		case closePromptBackgroundID:
			box.choose(true, checkboxChecked(box.check))
			procDestroyWindow.Call(hwnd)
		case closePromptQuitID:
			box.choose(false, checkboxChecked(box.check))
			procDestroyWindow.Call(hwnd)
		}
	case wmClose:
		// 关掉询问框本身：按"直接退出"处理，不留后台进程。
		if box != nil {
			box.choose(false, false)
		}
		procDestroyWindow.Call(hwnd)
	case wmKeyDown:
		if wParam == vkEscape {
			if box != nil {
				box.choose(false, false)
			}
			procDestroyWindow.Call(hwnd)
		}
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	case wmCtlColorStatic:
		// 说明文字要跟窗口底色一致，否则深色主题下会是一块突兀的方块。
		brush, _, _ := procGetSysColorBrush.Call(colorWindow)
		return brush
	}
	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return ret
}

func checkboxChecked(check uintptr) bool {
	if check == 0 {
		return false
	}
	state, _, _ := procSendMessageW.Call(check, bmGetCheck, 0, 0)
	return state == bstChecked
}

/*
askCloseBehavior 弹出自建询问框；建不起来时退回系统消息框。

窗口过程跑在 Windows 的回调里，任何意外都不能带崩整个进程：
这里兜住 panic，退回到系统消息框；再兜不住就返回"没问出来"，
由调用方按直接退出收尾——总之不会留下一个关不掉的进程。
*/
func askCloseBehavior() (answer CloseAnswer) {
	defer func() {
		if recover() != nil {
			answer = askCloseMessageBox()
		}
	}()
	if background, remember, ok := showClosePrompt(); ok {
		return CloseAnswer{Background: background, Remember: remember, Answered: true}
	}
	return askCloseMessageBox()
}

// procsReady 检查要用的系统函数都在（个别精简或注入过的系统上会缺），
// 缺一个就不冒险调用：LazyProc.Call 找不到导出时会直接 panic。
func procsReady(procs ...*syscall.LazyProc) bool {
	for _, proc := range procs {
		if err := proc.Find(); err != nil {
			return false
		}
	}
	return true
}

// closePromptAvailable 表示这个平台能在窗口关掉之后弹一次询问框。
func closePromptAvailable() bool { return true }

/*
showClosePrompt 建窗口并跑消息循环，返回用户的选择。
ok 为假表示窗口没能建起来（DLL 缺失、控件创建失败等），调用方另想办法。
*/
func showClosePrompt() (background, remember, ok bool) {
	// 同一时刻只弹一个；真撞上了就交给系统消息框兜底。
	if !promptGate.TryLock() {
		return false, false, false
	}
	defer promptGate.Unlock()

	if !procsReady(procRegisterClassExW, procCreateWindowExW, procDefWindowProcW, procDestroyWindow,
		procShowWindow, procGetMessageW, procTranslateMessage, procDispatchMessageW, procPostQuitMessage,
		procAdjustWindowRectEx, procSendMessageW, procSetForegroundWindow, procGetSystemMetrics,
		procGetSysColorBrush, procGetModuleHandleW, procCreateFontW, procDeleteObject) {
		return false, false, false
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	keep := make([]*uint16, 0, 8)
	text := func(value string) *uint16 {
		pointer, err := syscall.UTF16PtrFromString(value)
		if err != nil {
			return nil
		}
		keep = append(keep, pointer)
		return pointer
	}
	// 字符串只在 syscall 调用期间需要活着：调用完 Windows 已经复制走了。
	// 但 GC 看不见传给 syscall 的裸指针，所以这些指针要留在切片里。
	defer func() { runtime.KeepAlive(keep) }()
	ptr := func(value *uint16) uintptr { return uintptr(unsafe.Pointer(value)) }

	// 每个线程自己声明 DPI 感知，按真实缩放算尺寸：字不糊，框也不会小得看不清。
	if err := procSetThreadDpiAwarenessContext.Find(); err == nil {
		// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = -4
		procSetThreadDpiAwarenessContext.Call(^uintptr(3))
	}
	scale := 1.0
	if err := procGetDpiForSystem.Find(); err == nil {
		if dpi, _, _ := procGetDpiForSystem.Call(); dpi >= 96 {
			scale = math.Min(2.5, float64(dpi)/96)
		}
	}
	px := func(value int) int32 { return int32(math.Round(float64(value) * scale)) }

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := text("GoStudyClosePromptW")
	if className == nil {
		return false, false, false
	}
	class := wndClassExW{
		cbSize:        uint32(unsafe.Sizeof(wndClassExW{})),
		lpfnWndProc:   closePromptProc,
		hInstance:     hInstance,
		hCursor:       idcArrow,
		hbrBackground: colorWindow + 1,
		lpszClassName: className,
	}
	// 类只注册一次；已经注册过会返回 0，这里不需要处理。
	_, _, _ = procRegisterClassExW.Call(uintptr(unsafe.Pointer(&class)))

	clientWidth, clientHeight := px(470), px(204)
	frame := rectW{0, 0, clientWidth, clientHeight}
	style := uintptr(wsOverlapped | wsCaption | wsSysMenu)
	procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(&frame)), style, 0, 0)
	width, height := frame.right-frame.left, frame.bottom-frame.top
	screenWidth, _, _ := procGetSystemMetrics.Call(smCxScreen)
	screenHeight, _, _ := procGetSystemMetrics.Call(smCyScreen)
	box := &closePromptBox{}
	promptBox.Store(box)
	defer promptBox.Store(nil)
	hwnd, _, _ := procCreateWindowExW.Call(
		wsExTopmost,
		ptr(className),
		ptr(text("GoStudy · 关闭方式")),
		style,
		uintptr((int32(screenWidth)-width)/2), uintptr((int32(screenHeight)-height)/2),
		uintptr(width), uintptr(height),
		0, 0, hInstance, 0)
	if hwnd == 0 {
		return false, false, false
	}

	font := createUIFont(px(15), 400)
	titleFont := createUIFont(px(16), 700)
	defer func() {
		if font != 0 {
			procDeleteObject.Call(font)
		}
		if titleFont != 0 {
			procDeleteObject.Call(titleFont)
		}
	}()
	child := func(class, label string, id uintptr, style uint32, x, y, w, h int32) uintptr {
		handle, _, _ := procCreateWindowExW.Call(0,
			ptr(text(class)), ptr(text(label)), uintptr(style|wsChild|wsVisible),
			uintptr(x), uintptr(y), uintptr(w), uintptr(h), hwnd, id, hInstance, 0)
		return handle
	}
	title := child("STATIC", "应用窗口已经关闭", 0, ssLeft, px(20), px(16), px(430), px(24))
	body := child("STATIC",
		"要让 GoStudy 继续在后台静默运行，还是现在退出？\r\n"+
			"后台期间不会再开新窗口，下次打开会连回它。\r\n"+
			"学习进度、笔记与对话两种方式都保留。",
		0, ssLeft, px(20), px(46), px(430), px(64))
	check := child("BUTTON", "不再提示，以后按这次的选择处理", 3, bsAutoCheckBox, px(20), px(114), px(430), px(24))
	backgroundButton := child("BUTTON", "后台静默运行", closePromptBackgroundID, bsPushButton, px(192), px(158), px(130), px(32))
	quitButton := child("BUTTON", "直接退出程序", closePromptQuitID, bsDefPushButton, px(330), px(158), px(120), px(32))
	if title == 0 || body == 0 || check == 0 || backgroundButton == 0 || quitButton == 0 {
		procDestroyWindow.Call(hwnd)
		return false, false, false
	}
	box.mu.Lock()
	box.check = check
	box.mu.Unlock()
	for _, control := range []struct {
		handle uintptr
		font   uintptr
	}{{title, titleFont}, {body, font}, {check, font}, {backgroundButton, font}, {quitButton, font}} {
		if control.font != 0 {
			procSendMessageW.Call(control.handle, wmSetFont, control.font, 1)
		}
	}

	procShowWindow.Call(hwnd, 5) // SW_SHOW
	procSetForegroundWindow.Call(hwnd)

	var message msgW
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&message)), 0, 0, 0)
		if int32(ret) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&message)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&message)))
	}
	chosenBackground, chosenRemember, done := box.answer()
	runtime.KeepAlive(box)
	return chosenBackground, chosenRemember, done
}

func createUIFont(height int32, weight int32) uintptr {
	if !procsReady(procCreateFontW) {
		return 0
	}
	face, err := syscall.UTF16PtrFromString("Microsoft YaHei UI")
	if err != nil {
		return 0
	}
	height = -height // 负数按字符高度算，正数会按字符外框算，字会偏大
	font, _, _ := procCreateFontW.Call(
		uintptr(height), 0, 0, 0, uintptr(weight),
		0, 0, 0, 1, // DEFAULT_CHARSET，中英文都交给系统替换字体
		0, 0, 0, 0,
		uintptr(unsafe.Pointer(face)))
	runtime.KeepAlive(face)
	return font
}

/*
askCloseMessageBox 是兜底：自建窗口失败时用系统消息框问一次。
消息框没有"不再提示"，所以答案里 Remember 永远为假，也不会改设置。
*/
func askCloseMessageBox() (answer CloseAnswer) {
	defer func() {
		if recover() != nil {
			answer = CloseAnswer{}
		}
	}()
	if !procsReady(procMessageBoxW) {
		return CloseAnswer{}
	}
	title, err := syscall.UTF16PtrFromString("GoStudy · 关闭方式")
	if err != nil {
		return CloseAnswer{}
	}
	text, err := syscall.UTF16PtrFromString("要让 GoStudy 继续在后台静默运行吗？\n\n" +
		"选择「否」会立刻结束进程；选择「是」进程留在后台，\n" +
		"下次打开 GoStudy 会直接连回它。")
	if err != nil {
		return CloseAnswer{}
	}
	// MB_YESNO | MB_ICONQUESTION | MB_DEFBUTTON2 | MB_TOPMOST | MB_SETFOREGROUND
	const flags = 0x00000004 | 0x00000020 | 0x00000200 | 0x00040000 | 0x00010000
	result, _, _ := procMessageBoxW.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), flags)
	runtime.KeepAlive(title)
	runtime.KeepAlive(text)
	switch result {
	case 6: // IDYES
		return CloseAnswer{Background: true, Answered: true}
	case 7: // IDNO
		return CloseAnswer{Answered: true}
	default:
		return CloseAnswer{}
	}
}
