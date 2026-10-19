package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Options struct {
	DataDir    string
	Addr       string
	RunTimeout time.Duration
	AITimeout  time.Duration
	// Platform 供更新逻辑挑选产物；留空按 runtime.GOOS，Android 外壳显式传 "android"。
	Platform string
	// CloseGrace 是页面关掉之后等"页面是否又回来"的时间；留空用 defaultCloseGrace。
	CloseGrace time.Duration
	// AskClose 是关闭窗口后问用户的方式；留空弹系统询问框，测试里换成固定答案。
	AskClose func() CloseAnswer
	// Logf 把引擎里的关键动作写进启动日志；留空就不记。
	Logf func(format string, args ...any)
}
type proof struct {
	hash       [32]byte
	valid      bool
	generation uint64
}

// curriculum 是课程的一份不可变快照；内容热更新时整体换掉，读的地方不用加锁。
type curriculum struct {
	courses  []Course
	lessons  []Lesson
	byID     map[string]Lesson
	warnings []string
}

type Engine struct {
	token        string
	store        *store
	runner       *Runner
	cur          atomic.Pointer[curriculum]
	mu           sync.Mutex
	proofs       map[string]proof
	nextRun      uint64
	aiGate       chan struct{}
	aiClient     *http.Client
	aiTimeout    time.Duration
	dataDir      string
	platform     string
	updateClient *http.Client
	downloads    *downloader
	host         string
	frontend     http.Handler
	quit         chan struct{}
	quitOnce     sync.Once
	// 窗口关闭相关：页面在线状态、宽限计时器、询问方式。
	pageMu     sync.Mutex
	pageOpen   bool
	closeTimer *time.Timer
	closeGrace time.Duration
	askClose   func() CloseAnswer
	// logf 由启动器接上，用来把"窗口关了、按设置退出"这类动作写进启动日志。
	logf func(format string, args ...any)
	// uninstaller 是程序目录里的 Inno 卸载程序路径，空表示当前是免安装版。
	uninstaller string
}

func New(opts Options, frontend http.Handler) (*Engine, error) {
	s, err := newStore(opts.DataDir)
	if err != nil {
		return nil, err
	}
	secret := make([]byte, 32)
	if _, err = rand.Read(secret); err != nil {
		return nil, err
	}
	if opts.AITimeout <= 0 {
		opts.AITimeout = 30 * time.Second
	}
	// 状态文件所在目录就是数据目录：内容包和更新包都放它下面。
	dataDir := filepath.Dir(s.path)
	courses, lessons, warnings, err := loadCurriculum(filepath.Join(dataDir, contentDirName))
	if err != nil {
		return nil, err
	}
	platform := opts.Platform
	if platform == "" {
		platform = runtime.GOOS
	}
	grace := opts.CloseGrace
	if grace <= 0 {
		grace = defaultCloseGrace
	}
	askClose := opts.AskClose
	if askClose == nil {
		askClose = askCloseBehavior
	}
	e := &Engine{
		token:        hex.EncodeToString(secret),
		store:        s,
		runner:       NewRunner(opts.RunTimeout),
		mu:           sync.Mutex{},
		proofs:       map[string]proof{},
		aiGate:       make(chan struct{}, 2),
		aiTimeout:    opts.AITimeout,
		aiClient:     newAIClient(opts.AITimeout),
		dataDir:      dataDir,
		platform:     platform,
		updateClient: newUpdateClient(),
		downloads:    &downloader{},
		frontend:     frontend,
		quit:         make(chan struct{}),
		closeGrace:   grace,
		askClose:     askClose,
		logf:         opts.Logf,
		uninstaller:  findUninstaller(platform),
		// 启动器紧接着就会打开应用窗口，所以先按"窗口开着"记账：页面被关掉时会
		// 明确报信，那时才改成已关闭；单实例判断也读它。
		pageOpen: true,
	}
	e.cur.Store(newCurriculum(courses, lessons, warnings))
	return e, nil
}

func newCurriculum(courses []Course, lessons []Lesson, warnings []string) *curriculum {
	byID := make(map[string]Lesson, len(lessons))
	for _, lesson := range lessons {
		byID[lesson.ID] = lesson
	}
	return &curriculum{courses: courses, lessons: lessons, byID: byID, warnings: warnings}
}

// reloadCurriculum 重新读一遍课程（内置 + 数据目录里的内容包），整体换掉快照。
func (e *Engine) reloadCurriculum() error {
	courses, lessons, warnings, err := loadCurriculum(e.contentDir())
	if err != nil {
		return fmt.Errorf("课程内容无法载入：%w", err)
	}
	e.cur.Store(newCurriculum(courses, lessons, warnings))
	return nil
}

// curriculumSnapshot 给处理器用的一次性快照。
func (e *Engine) curriculumSnapshot() *curriculum { return e.cur.Load() }

// Token 仅供可信本地启动器生成启动 URL，不经 API 返回。
func (e *Engine) Token() string { return e.token }

// Quit 在用户主动退出时关闭，启动器据此结束进程。
func (e *Engine) Quit() <-chan struct{} { return e.quit }
func (e *Engine) requestQuit()          { e.quitOnce.Do(func() { close(e.quit) }) }
func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")
	// 校验 Host 防止 DNS 重绑定；启动后必须精确匹配绑定地址。
	host, _, err := net.SplitHostPort(r.Host)
	if err != nil || host != "127.0.0.1" || e.host != "" && r.Host != e.host {
		writeError(w, http.StatusForbidden, "只允许本机回环访问")
		return
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host {
		writeError(w, http.StatusForbidden, "不允许跨来源访问")
		return
	}
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		writeError(w, http.StatusForbidden, "不允许跨站访问")
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/api" {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			writeError(w, 405, "不支持此请求方法")
			return
		}
		if e.frontend == nil {
			http.NotFound(w, r)
			return
		}
		e.frontend.ServeHTTP(w, r)
		return
	}
	supplied, ok := e.credential(r)
	if !ok || len(supplied) != len(e.token) || subtle.ConstantTimeCompare([]byte(supplied), []byte(e.token)) != 1 {
		writeError(w, http.StatusUnauthorized, "本地访问凭据无效，请使用启动器提供的地址")
		return
	}
	// 只有页面自己发的请求才算"页面还在"：页面关闭时报的信、以及启动器用来探活的
	// /api/status 都不算，否则探活会把已经关掉的窗口重新标成在线。
	switch r.URL.Path {
	case "/api/window-closed", "/api/status":
	default:
		e.pageActivity()
	}
	want := http.MethodPost
	switch r.URL.Path {
	case "/api/bootstrap", "/api/status":
		want = http.MethodGet
	}
	if r.Method != want {
		w.Header().Set("Allow", want)
		writeError(w, 405, "不支持此请求方法")
		return
	}
	switch r.URL.Path {
	case "/api/reset", "/api/quit", "/api/window-closed":
		// 这几个接口不带请求体，不做 JSON 校验。
	default:
		if want == http.MethodPost {
			media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if err != nil || media != "application/json" {
				writeError(w, 415, "请使用 application/json 请求体")
				return
			}
		}
	}
	switch r.URL.Path {
	case "/api/bootstrap":
		s := e.store.snapshot()
		cur := e.curriculumSnapshot()
		writeJSON(w, 200, Bootstrap{
			Courses:  cur.courses,
			Lessons:  cur.lessons,
			Progress: s.Progress,
			Settings: publicSettings(s.Settings),
			Runtime: Runtime{
				Platform:     runtimePlatform(),
				Engine:       "yaegi",
				Version:      versionInfo(),
				Content:      e.contentVersion(),
				ClosePrompt:  closePromptAvailable(),
				CanUninstall: e.uninstaller != "",
			},
			Conversations: chatList(s, nil).ConversationList,
		})
	case "/api/status":
		// 给启动器探活用的最小响应：有没有实例在跑、它的窗口还开着吗。
		writeJSON(w, 200, map[string]any{
			"status":     "ok",
			"version":    versionInfo(),
			"windowOpen": e.WindowOpen(),
		})
	case "/api/window-closed":
		writeJSON(w, 200, map[string]string{"status": "noted"})
		e.windowClosed()
	case "/api/close-behavior":
		e.handleCloseBehavior(w, r)
	case "/api/uninstall":
		e.handleUninstall(w, r)
	case "/api/run":
		e.handleRun(w, r)
	case "/api/progress":
		e.handleProgress(w, r)
	case "/api/settings":
		e.handleSettings(w, r)
	case "/api/chat":
		e.handleChat(w, r)
	case "/api/conversations":
		e.handleConversations(w, r)
	case "/api/update":
		e.handleUpdate(w, r)
	case "/api/quit":
		writeJSON(w, 200, map[string]string{"status": "stopping"})
		e.requestQuit()
	case "/api/reset":
		e.mu.Lock()
		defer e.mu.Unlock()
		s, err := e.store.update(func(s *state) error { s.Progress = emptyProgress(); return nil })
		if err != nil {
			writeError(w, 500, "保存学习进度失败")
			return
		}
		e.proofs = map[string]proof{}
		writeJSON(w, 200, s.Progress)
	default:
		writeError(w, 404, "接口不存在")
	}
}

/*
credential 取出请求里的本地凭据。

页面在关闭那一瞬间用 sendBeacon 报信，而 sendBeacon 带不了自定义请求头，
所以只有这一个接口允许把凭据放在查询串里；其余接口一律只认 Authorization。
*/
func (e *Engine) credential(r *http.Request) (string, bool) {
	if header := r.Header.Get("Authorization"); strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer "), true
	}
	if r.URL.Path == "/api/window-closed" && r.Method == http.MethodPost {
		return r.URL.Query().Get("token"), true
	}
	return "", false
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		writeError(w, 400, "请求 JSON 无效或超过 128 KiB")
		return false
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		writeError(w, 400, "只允许一个 JSON 对象")
		return false
	}
	return true
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
func (e *Engine) lesson(w http.ResponseWriter, id string) (Lesson, bool) {
	l, ok := e.curriculumSnapshot().byID[id]
	if !ok {
		writeError(w, 400, "课程不存在")
	}
	return l, ok
}

// lessonExists 供对话与内容包校验使用：只判断课程是否还在，不写响应。
func (e *Engine) lessonExists(id string) bool {
	_, ok := e.curriculumSnapshot().byID[id]
	return ok
}
func (e *Engine) handleRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LessonID string `json:"lessonId"`
		Code     string `json:"code"`
	}
	if !decode(w, r, &req) {
		return
	}
	l, ok := e.lesson(w, req.LessonID)
	if !ok {
		return
	}
	e.mu.Lock()
	p := e.proofs[l.ID]
	e.nextRun++
	p.generation = e.nextRun
	p.valid = false
	e.proofs[l.ID] = p
	e.mu.Unlock()
	result := e.runner.Run(r.Context(), req.Code, l.ExpectedOutput, l.Checks...)
	e.mu.Lock()
	current := e.proofs[l.ID]
	if current.generation == p.generation {
		current.hash = sha256.Sum256([]byte(req.Code))
		current.valid = result.Passed
		e.proofs[l.ID] = current
	}
	e.mu.Unlock()
	writeJSON(w, 200, result)
}
func (e *Engine) handleProgress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LessonID   string  `json:"lessonId"`
		Code       *string `json:"code"`
		Note       *string `json:"note"`
		Complete   *bool   `json:"complete"`
		QuizAnswer *int    `json:"quizAnswer"`
	}
	if !decode(w, r, &req) {
		return
	}
	l, ok := e.lesson(w, req.LessonID)
	if !ok {
		return
	}
	if req.Code != nil && len(*req.Code) > MaxCodeBytes || req.Note != nil && len(*req.Note) > 8<<10 {
		writeError(w, 400, "草稿最多 32 KiB，笔记最多 8 KiB")
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if req.Complete != nil && *req.Complete {
		p := e.proofs[l.ID]
		code := e.store.snapshot().Progress.Drafts[l.ID]
		if req.Code != nil {
			code = *req.Code
		}
		if !p.valid || sha256.Sum256([]byte(code)) != p.hash {
			writeError(w, 409, "请先成功运行当前代码，再完成课程")
			return
		}
		if req.QuizAnswer == nil || *req.QuizAnswer != l.Quiz.Answer {
			writeError(w, 400, "请先正确回答本课测验")
			return
		}
	}
	s, err := e.store.update(func(s *state) error {
		p := &s.Progress
		p.LastLesson = l.ID
		if req.Code != nil {
			p.Drafts[l.ID] = *req.Code
		}
		if req.Note != nil {
			p.Notes[l.ID] = *req.Note
		}
		if req.Complete != nil {
			found := false
			completed := []string{}
			for _, id := range p.Completed {
				if id == l.ID {
					found = true
					if !*req.Complete {
						continue
					}
				}
				completed = append(completed, id)
			}
			if *req.Complete && !found {
				completed = append(completed, l.ID)
			}
			p.Completed = completed
		}
		return nil
	})
	if err != nil {
		writeError(w, 500, "保存学习进度失败")
		return
	}
	writeJSON(w, 200, s.Progress)
}
func (e *Engine) handleSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BaseURL  string  `json:"baseUrl"`
		Model    string  `json:"model"`
		APIKey   *string `json:"apiKey"`
		ClearKey bool    `json:"clearKey"`
	}
	if !decode(w, r, &req) {
		return
	}
	base, err := normalizeBaseURL(req.BaseURL)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	model := strings.TrimSpace(req.Model)
	if len(model) > 256 || strings.ContainsAny(model, "\r\n") {
		writeError(w, 400, "模型名称无效")
		return
	}
	if req.APIKey != nil && (len(*req.APIKey) > 4096 || strings.ContainsAny(*req.APIKey, "\r\n")) {
		writeError(w, 400, "API 密钥无效")
		return
	}
	if req.ClearKey && req.APIKey != nil && *req.APIKey != "" {
		writeError(w, 400, "清除密钥和设置密钥不能同时使用")
		return
	}
	s, err := e.store.update(func(s *state) error {
		oldBase := s.Settings.BaseURL
		s.Settings.BaseURL = base
		s.Settings.Model = model
		// 留空代表保留已有密钥；跨服务地址修改时清除旧密钥，避免无意发送给新服务。
		if base != oldBase {
			s.Settings.APIKey = ""
		}
		if req.APIKey != nil && *req.APIKey != "" {
			s.Settings.APIKey = *req.APIKey
		}
		if req.ClearKey {
			s.Settings.APIKey = ""
		}
		return nil
	})
	if err != nil {
		writeError(w, 500, "保存 AI 设置失败")
		return
	}
	writeJSON(w, 200, publicSettings(s.Settings))
}
func (e *Engine) handleChat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if !decode(w, r, &req) {
		return
	}
	l := Lesson{Title: "Go 学习与连接确认", Concept: "按学习者的显式问题提供帮助"}
	if req.LessonID != "" {
		var ok bool
		l, ok = e.lesson(w, req.LessonID)
		if !ok {
			return
		}
	}
	if err := validateChat(req); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	select {
	case e.aiGate <- struct{}{}:
		defer func() { <-e.aiGate }()
	default:
		writeError(w, 429, "AI 正在处理其他问题，请稍后重试")
		return
	}
	if req.Stream {
		e.handleChatStream(w, r, req, l)
		return
	}
	reply, err := e.chat(r.Context(), req, l)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"reply": reply})
}

/*
handleChatStream 以 SSE 把模型输出逐段转发给页面。

先做一遍 planChat 校验配置，错误仍然用普通 JSON 返回；校验通过后才切到
text/event-stream，因此前端可以按 Content-Type 决定走流式还是普通 JSON。
*/
func (e *Engine) handleChatStream(w http.ResponseWriter, r *http.Request, req chatRequest, l Lesson) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "当前环境不支持流式响应")
		return
	}
	if _, err := e.planChat(req, l); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	send := func(value any) error {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}
	if err := send(map[string]any{"start": true}); err != nil {
		return
	}
	reply, err := e.chatStream(r.Context(), req, l, func(delta string) error {
		return send(map[string]string{"delta": delta})
	})
	if err != nil {
		_ = send(map[string]string{"error": err.Error()})
		return
	}
	_ = send(map[string]any{"done": true, "reply": reply})
}

type Server struct {
	engine   *Engine
	server   *http.Server
	listener net.Listener
	url      string
}

func Start(opts Options, frontend http.Handler) (*Server, error) {
	addr := opts.Addr
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil || host != "127.0.0.1" || port == "" {
		return nil, errors.New("监听地址必须为 127.0.0.1:端口")
	}
	e, err := New(opts, frontend)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}
	e.host = listener.Addr().String()
	server := &http.Server{Handler: e, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: opts.AITimeout + 15*time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 32 << 10}
	if server.WriteTimeout < 45*time.Second {
		server.WriteTimeout = 45 * time.Second
	}
	s := &Server{e, server, listener, "http://" + e.host + "/#token=" + url.QueryEscape(e.token)}
	go func() { _ = server.Serve(listener) }()
	return s, nil
}
func (s *Server) URL() string { return s.url }

// Done 表示用户已通过界面主动退出应用。
func (s *Server) Done() <-chan struct{} { return s.engine.Quit() }
func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		_ = s.server.Close()
	}
	s.engine.aiClient.CloseIdleConnections()
	return err
}
