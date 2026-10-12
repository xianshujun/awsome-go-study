package backend

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type storedSettings struct {
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
	APIKey  string `json:"apiKey,omitempty"`
}
type state struct {
	Progress Progress       `json:"progress"`
	Settings storedSettings `json:"settings"`
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
	if info, err := f.Stat(); err != nil {
		return nil, err
	} else if info.Size() > 4<<20 {
		return nil, errors.New("本地学习状态过大")
	}
	dec := json.NewDecoder(io.LimitReader(f, 4<<20))
	if err = dec.Decode(&s.value); err != nil {
		return nil, fmt.Errorf("读取学习状态失败: %w", err)
	}
	if err = dec.Decode(&struct{}{}); err != io.EOF {
		return nil, errors.New("本地学习状态包含多余数据")
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
	if err = os.Chmod(s.path, 0600); err != nil {
		return nil, err
	}
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
	return Settings{BaseURL: s.BaseURL, Model: s.Model, HasKey: s.APIKey != ""}
}
