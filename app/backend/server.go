package backend

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Options struct {
	DataDir    string
	Addr       string
	RunTimeout time.Duration
	AITimeout  time.Duration
}
type proof struct {
	hash       [32]byte
	valid      bool
	generation uint64
}
type Engine struct {
	token     string
	store     *store
	runner    *Runner
	courses   []Course
	lessons   []Lesson
	byID      map[string]Lesson
	mu        sync.Mutex
	proofs    map[string]proof
	nextRun   uint64
	aiGate    chan struct{}
	aiClient  *http.Client
	aiTimeout time.Duration
	host      string
	frontend  http.Handler
	quit      chan struct{}
	quitOnce  sync.Once
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
	courses, lessons, err := LoadCurriculum()
	if err != nil {
		return nil, err
	}
	e := &Engine{token: hex.EncodeToString(secret), store: s, runner: NewRunner(opts.RunTimeout), courses: courses, lessons: lessons, byID: map[string]Lesson{}, proofs: map[string]proof{}, aiGate: make(chan struct{}, 2), aiTimeout: opts.AITimeout, aiClient: newAIClient(opts.AITimeout), frontend: frontend, quit: make(chan struct{})}
	for _, l := range lessons {
		e.byID[l.ID] = l
	}
	return e, nil
}

// Token 仅供可信本地启动器生成启动 URL，不经 API 返回。
func (e *Engine) Token() string { return e.token }

// Quit 在用户主动退出时关闭，启动器据此结束进程。
func (e *Engine) Quit() <-chan struct{} { return e.quit }
func (e *Engine) requestQuit()           { e.quitOnce.Do(func() { close(e.quit) }) }
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
	supplied := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if len(supplied) != len(e.token) || subtle.ConstantTimeCompare([]byte(supplied), []byte(e.token)) != 1 || !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		writeError(w, http.StatusUnauthorized, "本地访问凭据无效，请使用启动器提供的地址")
		return
	}
	want := http.MethodPost
	if r.URL.Path == "/api/bootstrap" {
		want = http.MethodGet
	}
	if r.Method != want {
		w.Header().Set("Allow", want)
		writeError(w, 405, "不支持此请求方法")
		return
	}
	if want == http.MethodPost && r.URL.Path != "/api/reset" && r.URL.Path != "/api/quit" {
		media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || media != "application/json" {
			writeError(w, 415, "请使用 application/json 请求体")
			return
		}
	}
	switch r.URL.Path {
	case "/api/bootstrap":
		s := e.store.snapshot()
		writeJSON(w, 200, Bootstrap{e.courses, e.lessons, s.Progress, publicSettings(s.Settings), runtimeInfo()})
	case "/api/run":
		e.handleRun(w, r)
	case "/api/progress":
		e.handleProgress(w, r)
	case "/api/settings":
		e.handleSettings(w, r)
	case "/api/chat":
		e.handleChat(w, r)
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
	l, ok := e.byID[id]
	if !ok {
		writeError(w, 400, "课程不存在")
	}
	return l, ok
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
	result := e.runner.Run(r.Context(), req.Code, l.ExpectedOutput)
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
	reply, err := e.chat(r.Context(), req, l)
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"reply": reply})
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
