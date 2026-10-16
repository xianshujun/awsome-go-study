package backend

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

/*
对话存储的边界：state.json 有 4 MiB 上限，这里再给对话单独一份预算。
超出时先丢最旧的对话，再丢最旧的消息，保证写入始终能成功、文件不会无限长大。
*/
const (
	maxConversations   = 40
	maxMessagesPerChat = 200
	maxMessageBytes    = 32 << 10
	maxChatBytes       = 1 << 20
	maxChatTitleRunes  = 60
	chatTitleRunes     = 24
	chatPreviewRunes   = 60
	// 每条消息除了正文还有 role、时间戳和 JSON 结构，估算预算时一起算进去。
	messageOverhead = 96
)

type StoredMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	At      int64  `json:"at"`
}

type ConversationSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LessonID     string `json:"lessonId"`
	UpdatedAt    int64  `json:"updatedAt"`
	MessageCount int    `json:"messageCount"`
	Preview      string `json:"preview"`
}

type ConversationList struct {
	ActiveID string                `json:"activeId"`
	Items    []ConversationSummary `json:"items"`
}

type ConversationView struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	LessonID     string          `json:"lessonId"`
	UpdatedAt    int64           `json:"updatedAt"`
	MessageCount int             `json:"messageCount"`
	Messages     []StoredMessage `json:"messages"`
}

type conversationPayload struct {
	ConversationList
	Conversation *ConversationView `json:"conversation,omitempty"`
}

func newChatID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("c%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func chatSummary(c conversation) ConversationSummary {
	s := ConversationSummary{ID: c.ID, Title: c.Title, LessonID: c.LessonID, UpdatedAt: c.UpdatedAt, MessageCount: len(c.Messages)}
	if n := len(c.Messages); n > 0 {
		s.Preview = oneLine(c.Messages[n-1].Content, chatPreviewRunes)
	}
	return s
}

func chatView(c conversation) ConversationView {
	messages := c.Messages
	if messages == nil {
		messages = []StoredMessage{}
	}
	return ConversationView{ID: c.ID, Title: c.Title, LessonID: c.LessonID, UpdatedAt: c.UpdatedAt, MessageCount: len(messages), Messages: messages}
}

// chatList 只带摘要：对话正文由前端在打开时按需取，bootstrap 不必背着全部历史。
func chatList(s state, view *ConversationView) conversationPayload {
	items := make([]ConversationSummary, 0, len(s.Conversations))
	for _, c := range s.Conversations {
		items = append(items, chatSummary(c))
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].UpdatedAt > items[j].UpdatedAt })
	return conversationPayload{ConversationList: ConversationList{ActiveID: s.ActiveConversation, Items: items}, Conversation: view}
}

func findChat(s *state, id string) int {
	for i, c := range s.Conversations {
		if c.ID == id {
			return i
		}
	}
	return -1
}

// oneLine 把多行文本压成一行并按 rune 截断，用于标题与列表里的预览。
func oneLine(text string, limit int) string {
	text = strings.Join(strings.Fields(text), " ")
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return string(runes[:limit]) + "…"
}

func chatBytes(c conversation) int {
	total := messageOverhead + len(c.Title) + len(c.LessonID)
	for _, m := range c.Messages {
		total += len(m.Content) + messageOverhead
	}
	return total
}

// tailWithin 从最新的消息往回保留，直到用完预算。
// 连最新一条都放不下时返回空——宁可整段丢掉，也不让总预算被突破。
func tailWithin(messages []StoredMessage, budget int) []StoredMessage {
	used := 0
	start := len(messages)
	for i := len(messages) - 1; i >= 0; i-- {
		used += len(messages[i].Content) + messageOverhead
		if used > budget {
			if start == len(messages) {
				return nil
			}
			break
		}
		start = i
	}
	return append([]StoredMessage{}, messages[start:]...)
}

// trimConversations 按更新时间倒序整理，并压回各项上限；返回的切片可安全替换原值。
func trimConversations(list []conversation) []conversation {
	sorted := append([]conversation{}, list...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].UpdatedAt > sorted[j].UpdatedAt })
	if len(sorted) > maxConversations {
		sorted = sorted[:maxConversations]
	}
	budget := maxChatBytes
	kept := make([]conversation, 0, len(sorted))
	for _, c := range sorted {
		if len(c.Messages) > maxMessagesPerChat {
			c.Messages = append([]StoredMessage{}, c.Messages[len(c.Messages)-maxMessagesPerChat:]...)
		}
		if c.Messages == nil {
			c.Messages = []StoredMessage{}
		}
		cost := chatBytes(c)
		if cost <= budget {
			kept = append(kept, c)
			budget -= cost
			continue
		}
		c.Messages = tailWithin(c.Messages, budget)
		if len(c.Messages) > 0 {
			kept = append(kept, c)
		}
		break
	}
	return kept
}

type conversationRequest struct {
	Action   string `json:"action"`
	ID       string `json:"id"`
	LessonID string `json:"lessonId"`
	Title    string `json:"title"`
	Role     string `json:"role"`
	Content  string `json:"content"`
}

func (e *Engine) handleConversations(w http.ResponseWriter, r *http.Request) {
	var req conversationRequest
	if !decode(w, r, &req) {
		return
	}
	if req.Action == "list" {
		s := e.store.snapshot()
		writeJSON(w, 200, chatList(s, nil))
		return
	}
	var payload conversationPayload
	_, err := e.store.update(func(s *state) error {
		view, err := e.applyConversation(s, req)
		if err != nil {
			return err
		}
		payload = chatList(*s, view)
		return nil
	})
	if err != nil {
		writeError(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, payload)
}

// applyConversation 在状态副本上执行一次对话操作。返回的视图为"操作后应该展示的对话"。
func (e *Engine) applyConversation(s *state, req conversationRequest) (*ConversationView, error) {
	now := time.Now().Unix()
	switch req.Action {
	case "create":
		if req.LessonID != "" {
			if _, ok := e.byID[req.LessonID]; !ok {
				return nil, errors.New("课程不存在")
			}
		}
		c := conversation{ID: newChatID(), LessonID: req.LessonID, UpdatedAt: now, Messages: []StoredMessage{}}
		s.Conversations = trimConversations(append(s.Conversations, c))
		s.ActiveConversation = c.ID
		view := chatView(c)
		return &view, nil
	case "get":
		i := findChat(s, req.ID)
		if i < 0 {
			return nil, errors.New("对话不存在")
		}
		s.ActiveConversation = req.ID
		view := chatView(s.Conversations[i])
		return &view, nil
	case "append":
		role := strings.TrimSpace(req.Role)
		if role != "user" && role != "assistant" {
			return nil, errors.New("消息角色无效")
		}
		content := strings.TrimSpace(req.Content)
		if content == "" {
			return nil, errors.New("消息内容不能为空")
		}
		if len(content) > maxMessageBytes {
			return nil, errors.New("单条消息过长")
		}
		i := findChat(s, req.ID)
		if i < 0 {
			return nil, errors.New("对话不存在")
		}
		if req.LessonID != "" {
			if _, ok := e.byID[req.LessonID]; !ok {
				return nil, errors.New("课程不存在")
			}
			// 记录最近一次讨论的课程，列表里就能看出这段对话在问哪一课。
			s.Conversations[i].LessonID = req.LessonID
		}
		c := s.Conversations[i]
		if c.Title == "" && role == "user" {
			c.Title = oneLine(content, chatTitleRunes)
		}
		c.Messages = append(append([]StoredMessage{}, c.Messages...), StoredMessage{Role: role, Content: content, At: now})
		if len(c.Messages) > maxMessagesPerChat {
			c.Messages = c.Messages[len(c.Messages)-maxMessagesPerChat:]
		}
		c.UpdatedAt = now
		s.Conversations[i] = c
		s.ActiveConversation = c.ID
		return viewAfterTrim(s, c.ID), nil
	case "rename":
		title := oneLine(req.Title, maxChatTitleRunes)
		if title == "" {
			return nil, errors.New("对话标题不能为空")
		}
		i := findChat(s, req.ID)
		if i < 0 {
			return nil, errors.New("对话不存在")
		}
		s.Conversations[i].Title = title
		view := chatView(s.Conversations[i])
		return &view, nil
	case "delete":
		i := findChat(s, req.ID)
		if i < 0 {
			return nil, errors.New("对话不存在")
		}
		removed := s.Conversations[i]
		s.Conversations = append(s.Conversations[:i:i], s.Conversations[i+1:]...)
		if s.ActiveConversation == removed.ID {
			s.ActiveConversation = ""
			if len(s.Conversations) > 0 {
				// 删掉当前对话后落到最近用过的那一段，而不是空白页。
				s.Conversations = trimConversations(s.Conversations)
				s.ActiveConversation = s.Conversations[0].ID
			}
		}
		if j := findChat(s, s.ActiveConversation); j >= 0 {
			view := chatView(s.Conversations[j])
			return &view, nil
		}
		return nil, nil
	default:
		return nil, errors.New("未知的对话操作")
	}
}

// viewAfterTrim 在压缩之后重新定位这条对话，确保返回的内容与真正存下来的完全一致。
func viewAfterTrim(s *state, id string) *ConversationView {
	s.Conversations = trimConversations(s.Conversations)
	if i := findChat(s, id); i >= 0 {
		view := chatView(s.Conversations[i])
		return &view
	}
	return nil
}

