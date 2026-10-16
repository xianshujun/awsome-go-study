package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type storedSettings struct {
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey,omitempty"`
	// CloseAction 决定关闭应用窗口之后怎么处理：ask 询问、background 后台静默、quit 直接退出。
	// 页面已经关掉的时候引擎还要读它，所以只能存在这份状态里，不能放浏览器 localStorage。
	CloseAction string `json:"closeAction,omitempty"`
}
type conversation struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	LessonID  string          `json:"lessonId"`
	UpdatedAt int64           `json:"updatedAt"`
	Messages  []StoredMessage `json:"messages"`
}
type state struct {
	Progress Progress       `json:"progress"`
	Settings storedSettings `json:"settings"`
	// 对话与学习进度并列存放：它们同样是本机数据，但体量和大小的算法不同，
	// 单独一份预算（见 conversations.go）比塞进 Progress 更好控制。
	Conversations      []conversation `json:"conversations,omitempty"`
	ActiveConversation string         `json:"activeConversation,omitempty"`
}
type store struct {
	mu    sync.Mutex
	path  string
	value state
}

func newStore(dir string) (*store, error) {
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(base, "go-study")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	s := &store{path: filepath.Join(dir, "state.json"), value: state{Progress: emptyProgress()}}
	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(io.LimitReader(f, 4<<20))
	// 文件损坏、过大或含多余数据时保留备份后重新开始：
	// 否则用户会遇到「每次都启动失败、又不知道删哪个文件」的死局。
	if info.Size() > 4<<20 {
		return quarantine(s)
	}
	if err = dec.Decode(&s.value); err != nil {
		return quarantine(s)
	}
	if err = dec.Decode(&struct{}{}); err != io.EOF {
		return quarantine(s)
	}
	if s.value.Progress.Completed == nil {
		s.value.Progress.Completed = []string{}
	}
	if s.value.Progress.Drafts == nil {
		s.value.Progress.Drafts = map[string]string{}
	}
	if s.value.Progress.Notes == nil {
		s.value.Progress.Notes = map[string]string{}
	}
	// 用户手工编辑过、或被旧版本写大的对话记录，装载时先压回预算之内。
	s.value.Conversations = trimConversations(s.value.Conversations)
	if findChat(&s.value, s.value.ActiveConversation) < 0 {
		s.value.ActiveConversation = ""
	}
	if err = os.Chmod(s.path, 0600); err != nil {
		return nil, err
	}
	return s, nil
}

// quarantine 把无法解析的状态文件改名备份，并返回一个全新的空状态。
func quarantine(s *store) (*store, error) {
	backup := fmt.Sprintf("%s.corrupt-%s", s.path, time.Now().Format("20060102-150405"))
	if err := os.Rename(s.path, backup); err != nil {
		return nil, fmt.Errorf("本地学习状态无法读取，备份也失败: %w", err)
	}
	s.value = state{Progress: emptyProgress()}
	return s, nil
}

func cloneState(v state) state {
	p := v.Progress
	p.Completed = append([]string{}, p.Completed...)
	p.Drafts = map[string]string{}
	p.Notes = map[string]string{}
	for k, x := range v.Progress.Drafts {
		p.Drafts[k] = x
	}
	for k, x := range v.Progress.Notes {
		p.Notes[k] = x
	}
	v.Progress = p
	chats := make([]conversation, 0, len(v.Conversations))
	for _, c := range v.Conversations {
		c.Messages = append([]StoredMessage{}, c.Messages...)
		chats = append(chats, c)
	}
	v.Conversations = chats
	return v
}
func (s *store) snapshot() state { s.mu.Lock(); defer s.mu.Unlock(); return cloneState(s.value) }
func (s *store) update(change func(*state) error) (state, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := cloneState(s.value)
	if err := change(&next); err != nil {
		return state{}, err
	}
	data, err := json.Marshal(next)
	if err != nil {
		return state{}, err
	}
	if len(data) > 4<<20 {
		return state{}, errors.New("学习状态超过存储上限")
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".state-*")
	if err != nil {
		return state{}, err
	}
	name := f.Name()
	defer os.Remove(name)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return state{}, err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return state{}, err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return state{}, err
	}
	if err = f.Close(); err != nil {
		return state{}, err
	}
	// 同一目录的 rename 避免写入中断留下半份 JSON。
	if err = os.Rename(name, s.path); err != nil {
		return state{}, err
	}
	s.value = next
	return cloneState(next), nil
}
func publicSettings(s storedSettings) Settings {
	return Settings{
		BaseURL:     s.BaseURL,
		Model:       s.Model,
		HasKey:      s.APIKey != "",
		CloseAction: normalizeCloseAction(s.CloseAction),
	}
}
