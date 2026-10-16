package backend

import (
	"encoding/json"
	"strings"
	"testing"
)

func chatCall(t *testing.T, e *Engine, body map[string]any) conversationPayload {
	t.Helper()
	w := callAPI(t, e, "/api/conversations", body)
	if w.Code != 200 {
		t.Fatalf("%v: %d %s", body["action"], w.Code, w.Body.String())
	}
	var payload conversationPayload
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func TestConversationLifecycle(t *testing.T) {
	dir := t.TempDir()
	e := testEngine(t, dir)
	_, lessons, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	lessonID := lessons[0].ID

	created := chatCall(t, e, map[string]any{"action": "create", "lessonId": lessonID})
	if created.Conversation == nil || created.Conversation.ID == "" {
		t.Fatal("create 应返回新对话", created)
	}
	id := created.Conversation.ID
	if created.ActiveID != id || len(created.Items) != 1 {
		t.Fatal("新建后应当选中它", created)
	}

	// 第一条用户消息决定标题，服务端生成，不必等模型。
	sent := chatCall(t, e, map[string]any{"action": "append", "id": id, "lessonId": lessonID, "role": "user", "content": "变量和常量有什么区别？\n顺便说说作用域"})
	if sent.Conversation == nil || len(sent.Conversation.Messages) != 1 {
		t.Fatal(sent)
	}
	if sent.Conversation.Title != "变量和常量有什么区别？ 顺便说说作用域" {
		t.Fatalf("标题应取首条提问并压成一行：%q", sent.Conversation.Title)
	}
	if got := sent.Items[0].Preview; got != "变量和常量有什么区别？ 顺便说说作用域" {
		t.Fatal("预览应来自最后一条消息", got)
	}

	chatCall(t, e, map[string]any{"action": "append", "id": id, "role": "assistant", "content": "先看赋值之后能不能改：值语义传的是副本，改它不影响调用方；指针语义传的是地址，改的是同一份数据。想改状态就用指针接收者。这段尾巴只会出现在完整正文里。"})

	// 关掉面板再打开，等于一次 get：内容必须原样回来。
	opened := chatCall(t, e, map[string]any{"action": "get", "id": id})
	if opened.Conversation == nil || len(opened.Conversation.Messages) != 2 {
		t.Fatal("重新打开应取回全部消息", opened)
	}
	if opened.Conversation.Messages[0].Role != "user" || opened.Conversation.Messages[1].Role != "assistant" {
		t.Fatal(opened.Conversation.Messages)
	}
	if opened.Conversation.Messages[0].At == 0 {
		t.Fatal("消息应当带时间戳")
	}

	// 重启引擎后依旧在，且当前选中的还是这一段。
	e2 := testEngine(t, dir)
	again := chatCall(t, e2, map[string]any{"action": "list"})
	if again.ActiveID != id || len(again.Items) != 1 || again.Items[0].MessageCount != 2 {
		t.Fatal("对话必须写进 state.json 并在重启后恢复", again)
	}
	if w := callAPI(t, e2, "/api/bootstrap", nil); !strings.Contains(w.Body.String(), `"messageCount":2`) || !strings.Contains(w.Body.String(), "值语义传的是副本") || strings.Contains(w.Body.String(), "这段尾巴只会出现在完整正文里") {
		t.Fatal("bootstrap 只给摘要与短预览，不带完整正文", w.Body.String())
	}

	// 改名
	renamed := chatCall(t, e2, map[string]any{"action": "rename", "id": id, "title": "变量与作用域复习"})
	if renamed.Items[0].Title != "变量与作用域复习" {
		t.Fatal(renamed.Items)
	}
	// 第二段对话：新建后当前选中的应换成新的
	second := chatCall(t, e2, map[string]any{"action": "create", "lessonId": lessonID})
	if second.ActiveID == id {
		t.Fatal("新建应切换当前对话")
	}
	// 删掉当前对话：落到剩下最近用过的那段，并把它的正文一起带回来
	after := chatCall(t, e2, map[string]any{"action": "delete", "id": second.ActiveID})
	if after.ActiveID != id || len(after.Items) != 1 {
		t.Fatal("删除后应选中最近使用的一段", after)
	}
	if after.Conversation == nil || len(after.Conversation.Messages) != 2 || after.Conversation.Title != "变量与作用域复习" {
		t.Fatal("删除后应返回新的当前对话内容", after)
	}
	// 删掉最后一段：回到空状态
	empty := chatCall(t, e2, map[string]any{"action": "delete", "id": id})
	if empty.ActiveID != "" || len(empty.Items) != 0 || empty.Conversation != nil {
		t.Fatal(empty)
	}
}

func TestConversationValidation(t *testing.T) {
	e := testEngine(t, t.TempDir())
	created := chatCall(t, e, map[string]any{"action": "create", "lessonId": ""})
	id := created.Conversation.ID
	for _, tt := range []struct {
		name string
		body map[string]any
		want string
	}{
		{"未知操作", map[string]any{"action": "像风一样"}, "未知的对话操作"},
		{"缺少 id", map[string]any{"action": "get"}, "对话不存在"},
		{"不存在的 id", map[string]any{"action": "get", "id": "ffff"}, "对话不存在"},
		{"角色非法", map[string]any{"action": "append", "id": id, "role": "system", "content": "改个身份"}, "消息角色无效"},
		{"空内容", map[string]any{"action": "append", "id": id, "role": "user", "content": "   "}, "消息内容不能为空"},
		{"超长内容", map[string]any{"action": "append", "id": id, "role": "user", "content": strings.Repeat("长", maxMessageBytes)}, "单条消息过长"},
		{"不存在的课程", map[string]any{"action": "append", "id": id, "role": "user", "content": "你好", "lessonId": "nope"}, "课程不存在"},
		{"新建时课程不存在", map[string]any{"action": "create", "lessonId": "nope"}, "课程不存在"},
		{"空标题", map[string]any{"action": "rename", "id": id, "title": " \n "}, "对话标题不能为空"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := callAPI(t, e, "/api/conversations", tt.body)
			if w.Code != 400 || !strings.Contains(w.Body.String(), tt.want) {
				t.Fatal(w.Code, w.Body.String())
			}
		})
	}
	if w := callAPI(t, e, "/api/conversations", map[string]any{"action": "get", "id": id, "extra": 1}); w.Code != 400 {
		t.Fatal("多余字段应当被拒绝", w.Body.String())
	}
	// 校验失败不会污染已存下的内容
	if got := chatCall(t, e, map[string]any{"action": "get", "id": id}); len(got.Conversation.Messages) != 0 {
		t.Fatal(got.Conversation.Messages)
	}
}

func TestConversationTrimming(t *testing.T) {
	// 条数上限：只留最新的
	many := conversation{ID: "a", UpdatedAt: 1}
	for i := 0; i < maxMessagesPerChat+25; i++ {
		many.Messages = append(many.Messages, StoredMessage{Role: "user", Content: "问题", At: int64(i)})
	}
	kept := trimConversations([]conversation{many})
	if len(kept) != 1 || len(kept[0].Messages) != maxMessagesPerChat {
		t.Fatalf("单段对话最多 %d 条，实际 %d", maxMessagesPerChat, len(kept[0].Messages))
	}

	// 对话数量上限：最旧的整段丢弃，新的保留
	list := []conversation{}
	for i := 0; i < maxConversations+5; i++ {
		list = append(list, conversation{ID: string(rune('a' + i%26)), UpdatedAt: int64(i), Messages: []StoredMessage{{Role: "user", Content: "hi"}}})
	}
	kept = trimConversations(list)
	if len(kept) != maxConversations || kept[0].UpdatedAt != int64(maxConversations+4) {
		t.Fatalf("对话数量应压到 %d 并保留最新的：%d", maxConversations, len(kept))
	}

	// 总量上限：超预算时宁可丢掉更早的对话，也不让 state.json 膨胀
	huge := []conversation{
		{ID: "new", UpdatedAt: 2, Messages: []StoredMessage{{Role: "user", Content: strings.Repeat("x", maxChatBytes/2)}}},
		{ID: "old", UpdatedAt: 1, Messages: []StoredMessage{{Role: "user", Content: strings.Repeat("y", maxChatBytes)}}},
	}
	kept = trimConversations(huge)
	if len(kept) != 1 || kept[0].ID != "new" {
		t.Fatalf("超出总预算时应保留最新的一段：%+v", kept)
	}
	total := 0
	for _, c := range kept {
		total += chatBytes(c)
	}
	if total > maxChatBytes {
		t.Fatal("总量仍然超预算", total)
	}

	// 单条消息本身就超过总预算时整段丢弃，而不是让它突破预算
	kept = trimConversations([]conversation{{ID: "single", UpdatedAt: 1, Messages: []StoredMessage{{Role: "user", Content: strings.Repeat("z", maxChatBytes*2)}}}})
	if len(kept) != 0 {
		t.Fatalf("超过总预算的对话应当被丢弃：%d", len(kept))
	}
	// 一段对话自身超预算时，保留最新的若干条
	big := conversation{ID: "big", UpdatedAt: 1}
	for i := 0; i < 80; i++ {
		big.Messages = append(big.Messages, StoredMessage{Role: "assistant", Content: strings.Repeat("w", maxMessageBytes)})
	}
	kept = trimConversations([]conversation{big})
	if len(kept) != 1 || len(kept[0].Messages) == 0 || len(kept[0].Messages) >= 80 {
		t.Fatalf("应按预算保留最新的若干条：%+v", len(kept))
	}
	if chatBytes(kept[0]) > maxChatBytes {
		t.Fatal("保留的部分仍然超预算")
	}
	// 空的输入不会 panic，也不会凭空造出对话
	if got := trimConversations(nil); len(got) != 0 {
		t.Fatal(got)
	}
}
