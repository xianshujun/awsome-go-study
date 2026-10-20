package backend

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func labTestNow() time.Time {
	return time.Date(2026, 10, 20, 11, 30, 0, 0, time.Local)
}

func TestLabPackLoadsAndRotates(t *testing.T) {
	engine, err := newLabEngine("")
	if err != nil {
		t.Fatal(err)
	}
	pack, _, err := loadLabPack("")
	if err != nil {
		t.Fatal(err)
	}
	if len(pack.Scenarios) < 40 {
		t.Fatalf("内置事故场景至少应有 40 个，实际 %d", len(pack.Scenarios))
	}
	boot := engine.bootstrap(LabProgress{}, true, labTestNow())
	if boot.Company == "" || boot.Today.ID == "" || len(boot.Catalog) != len(pack.Scenarios) {
		t.Fatalf("演练首页数据不完整：%+v", boot)
	}
	if len(boot.Archive) != labArchiveDays {
		t.Fatalf("历史存档应为 %d 天，实际 %d", labArchiveDays, len(boot.Archive))
	}
	// 同一天反复打开必须是同一场事故，换一天则轮换。
	first, err := engine.openCase(boot.Today.Scenario, boot.Today.Date, LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	again, err := engine.openCase(boot.Today.Scenario, boot.Today.Date, LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != again.ID || first.Title != again.Title {
		t.Fatalf("同一天的事故应当稳定：%s / %s", first.ID, again.ID)
	}
	yesterday, err := engine.openCase(boot.Archive[0].Scenario, boot.Archive[0].Date, LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	if yesterday.ID == first.ID {
		t.Fatal("不同日期应当得到不同的事故实例")
	}
	if len(first.Services) < 3 || len(first.Metrics) < 2 || len(first.Presets) == 0 {
		t.Fatalf("事故现场不完整：%+v", first)
	}
	// 占位符必须被替换，不能把 {{orderId}} 直接交给用户。
	all := first.Brief + first.Mission + strings.Join(first.Alerts, " ")
	if strings.Contains(all, "{{") {
		t.Fatalf("事故内容里还有未替换的占位符：%s", all)
	}
}

func TestLabCommandsEvidenceAndGrading(t *testing.T) {
	engine, err := newLabEngine("")
	if err != nil {
		t.Fatal(err)
	}
	now := labTestNow()
	view, err := engine.openCase("redis-hotkey", "2026-10-20", LabProgress{}, now)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{
		"redis info",
		"metrics order-api",
		"logs order-api --level=ERROR --since=30m",
	} {
		result, err := engine.run(view.ID, command)
		if err != nil {
			t.Fatalf("%s: %v", command, err)
		}
		if strings.TrimSpace(result.Output) == "" {
			t.Fatalf("%s 没有输出", command)
		}
	}
	// 中间件命令必须来自场景包，不能由程序编造。
	if output, err := engine.run(view.ID, "redis slowlog get 10"); err != nil || !strings.Contains(output.Output, "HGETALL") {
		t.Fatalf("redis slowlog 输出异常：%v / %s", err, output.Output)
	}
	if output, err := engine.run(view.ID, "kubectl delete pod x"); err != nil || !strings.Contains(output.Output, "只读") {
		t.Fatalf("只读演练环境不应执行删除：%v / %s", err, output.Output)
	}

	hint, err := engine.hint(view.ID)
	if err != nil || hint.Hint == "" || hint.HintsUsed != 1 {
		t.Fatalf("提示异常：%+v / %v", hint, err)
	}
	result, progress, err := engine.submit(view.ID, 1, 1, LabProgress{})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed || !result.RootCauseCorrect || !result.FixCorrect {
		t.Fatalf("正确答案应当通过：%+v", result)
	}
	if result.EvidenceHit < 3 || result.EvidenceHit > result.EvidenceTotal {
		t.Fatalf("证据计数异常：%+v", result)
	}
	if result.Score <= 0 || result.Score > 100 || result.Hints != 1 {
		t.Fatalf("评分异常：%+v", result)
	}
	if progress.Solved != 1 || progress.Results[view.ID].Attempts != 1 {
		t.Fatalf("进度没有正确累计：%+v", progress)
	}

	// 错误答案不能通过，但证据分仍然保留，方便学习者看到"查了但判断错"。
	wrongView, err := engine.openCase("redis-hotkey", "2026-10-19", LabProgress{}, now)
	if err != nil {
		t.Fatal(err)
	}
	wrong, _, err := engine.submit(wrongView.ID, 0, 0, progress)
	if err != nil {
		t.Fatal(err)
	}
	if wrong.Passed || wrong.RootCauseCorrect || wrong.FixCorrect {
		t.Fatalf("错误答案不应通过：%+v", wrong)
	}
}

func TestLabAIToolDispatch(t *testing.T) {
	engine, err := newLabEngine("")
	if err != nil {
		t.Fatal(err)
	}
	view, err := engine.openCase("redis-hotkey", "2026-10-20", LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	session, ok := engine.session(view.ID)
	if !ok {
		t.Fatal("会话不存在")
	}
	if output := labTool(session.c, "query_logs", `{"service":"order-api","level":"ERROR","sinceMinutes":30,"limit":5}`); !strings.Contains(output, "ERROR") {
		t.Fatalf("query_logs 没有返回错误日志：%s", output)
	}
	if output := labTool(session.c, "inspect_component", `{"component":"redis","command":"info"}`); !strings.Contains(output, "used_memory") {
		t.Fatalf("inspect_component 没有返回 redis info：%s", output)
	}
	if output := labTool(session.c, "get_pods", `{}`); !strings.Contains(output, "CrashLoopBackOff") && !strings.Contains(output, "Running") {
		t.Fatalf("get_pods 输出异常：%s", output)
	}
	if output := labTool(session.c, "get_trace", `{"orderId":"--slow"}`); !strings.Contains(output, "ms") {
		t.Fatalf("get_trace 输出异常：%s", output)
	}
}

// 内容包版本必须和发布脚本算出来的一致，否则用户会被反复提示"有新内容"。
func TestLabContentVersionMatchesReleaseScript(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("没有 bash")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("没有 python3")
	}
	out := t.TempDir()
	command := exec.Command("bash", "../scripts/make-update-manifest.sh", "--content-only", "--out-dir", out)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("发布脚本失败：%v\n%s", err, output)
	}
	data, err := os.ReadFile(filepath.Join(out, "content.json"))
	if err != nil {
		t.Fatal(err)
	}
	var info contentInfo
	if err := json.Unmarshal(data, &info); err != nil {
		t.Fatal(err)
	}
	engine := testEngine(t, t.TempDir())
	if engine.contentVersion() != info.Version {
		t.Fatalf("内容版本不一致：程序 %s，脚本 %s", engine.contentVersion(), info.Version)
	}
	if info.MinAppVersion != "0.10.0" {
		t.Fatalf("带事故场景的内容包必须要求 0.10.0，实际 %q", info.MinAppVersion)
	}
	archive, err := zip.OpenReader(filepath.Join(out, "content.zip"))
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	found := false
	for _, entry := range archive.File {
		if entry.Name == labPackFile {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("内容包里没有 labs.json")
	}
}

// 模型发起工具调用时，后端要把 tool 结果回填，再让模型给出最终回复。
func TestLabAIToolLoop(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Tools    []json.RawMessage `json:"tools"`
			Messages []struct {
				Role       string `json:"role"`
				Content    string `json:"content"`
				ToolCallID string `json:"tool_call_id"`
			} `json:"messages"`
		}
		_ = json.Unmarshal(body, &payload)
		call := calls.Add(1)
		if call == 1 {
			if len(payload.Tools) == 0 {
				t.Error("第一轮应当带上工具定义")
			}
			_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"先查日志。","tool_calls":[{"id":"call_1","type":"function","function":{"name":"query_logs","arguments":"{\"service\":\"order-api\",\"level\":\"ERROR\",\"limit\":5}"}}]}}]}`))
			return
		}
		foundToolResult := false
		for _, message := range payload.Messages {
			if message.Role == "tool" && strings.Contains(message.Content, "ERROR") {
				foundToolResult = true
			}
		}
		if !foundToolResult {
			t.Error("第二轮应当收到工具查询结果")
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"先看订单服务的 ERROR 日志，重点确认 Kafka 副本是否满足 acks=all。"}}]}`))
	}))
	defer server.Close()

	e := testEngine(t, t.TempDir())
	view, err := e.labs.openCase("redis-hotkey", "2026-10-20", LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	session, ok := e.labs.session(view.ID)
	if !ok {
		t.Fatal("会话不存在")
	}
	result, err := e.labChat(context.Background(), server.URL+"/v1", storedSettings{Model: "test"}, session, labChatRequest{
		CaseID:  view.ID,
		Message: "帮我看看订单服务为什么超时",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result.Reply, "acks=all") {
		t.Fatalf("最终回复不对：%s", result.Reply)
	}
	if len(result.Steps) != 1 || result.Steps[0].Tool != "query_logs" {
		t.Fatalf("工具步骤没有记录：%+v", result.Steps)
	}
	if calls.Load() != 2 {
		t.Fatalf("应当调用两轮模型，实际 %d", calls.Load())
	}
}

// 兼容服务不支持 tools 时，退化成自动附带证据的普通对话，而不是把功能弄坏。
func TestLabAIFallbackWithoutTools(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		calls.Add(1)
		if strings.Contains(string(body), `"tools"`) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"tools not supported"}}`))
			return
		}
		if !strings.Contains(string(body), "自动检索") {
			t.Error("退化请求应当带上自动检索的证据包")
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"从已附带的证据看，先核对日志时间线与 Redis 连接池。"}}]}`))
	}))
	defer server.Close()

	e := testEngine(t, t.TempDir())
	view, err := e.labs.openCase("redis-hotkey", "2026-10-20", LabProgress{}, labTestNow())
	if err != nil {
		t.Fatal(err)
	}
	session, ok := e.labs.session(view.ID)
	if !ok {
		t.Fatal("会话不存在")
	}
	result, err := e.labChat(context.Background(), server.URL+"/v1", storedSettings{Model: "test"}, session, labChatRequest{
		CaseID:  view.ID,
		Message: "先看什么",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Steps[0].Tool != "auto_evidence" {
		t.Fatalf("应当标记为自动证据：%+v", result.Steps)
	}
	if calls.Load() != 2 {
		t.Fatalf("应当先失败一次再退化一次，实际 %d", calls.Load())
	}
}

// 事故场景也能跟着内容包热更新：应用后立即生效，重启后仍然保留。
func TestLabPackHotUpdate(t *testing.T) {
	e := testEngine(t, t.TempDir())
	pack, err := readEmbeddedLabPack()
	if err != nil {
		t.Fatal(err)
	}
	pack.Company = "测试商城"
	pack.Scenarios[0].Title = "内容包里的新事故标题"
	data, err := json.Marshal(pack)
	if err != nil {
		t.Fatal(err)
	}
	path := writeZip(t, map[string][]byte{labPackFile: data})
	if err := e.installContentPack(path, contentInfo{Version: "lab-content-1"}); err != nil {
		t.Fatal(err)
	}
	if company := e.labs.packSnapshot().Company; company != "测试商城" {
		t.Fatalf("热更新后公司名没变：%s", company)
	}
	scn, ok := e.labs.findScenario(pack.Scenarios[0].ID)
	if !ok || scn.Title != "内容包里的新事故标题" {
		t.Fatalf("热更新后场景标题没变：%+v", scn)
	}
	if got := e.contentVersion(); got != "lab-content-1" {
		t.Fatalf("内容版本没有落盘：%s", got)
	}
	// 重启后从磁盘上的 content/labs.json 读回来。
	e2 := testEngine(t, e.dataDir)
	if company := e2.labs.packSnapshot().Company; company != "测试商城" {
		t.Fatalf("重启后场景包丢失：%s", company)
	}

	// 坏的场景包必须整包拒绝，不能覆盖掉正在用的那一份。
	bad := writeZip(t, map[string][]byte{labPackFile: []byte(`{"version":1,"company":"坏包","scenarios":[]}`)})
	if err := e2.installContentPack(bad, contentInfo{Version: "lab-content-2"}); err == nil {
		t.Fatal("坏场景包不该被应用")
	}
	if company := e2.labs.packSnapshot().Company; company != "测试商城" {
		t.Fatalf("坏包不应覆盖现有场景：%s", company)
	}
}

// 48 个场景里的每条预设、证据和场景命令都必须真的能在终端里执行成功；
// 同时用正确选项跑一遍提交，确保每个场景的答案下标都能通过。
func TestLabEveryScenarioIsPlayable(t *testing.T) {
	engine, err := newLabEngine("")
	if err != nil {
		t.Fatal(err)
	}
	pack, _, err := loadLabPack("")
	if err != nil {
		t.Fatal(err)
	}
	now := labTestNow()
	progress := LabProgress{}
	for _, scn := range pack.Scenarios {
		view, err := engine.openCase(scn.ID, "2026-10-20", progress, now)
		if err != nil {
			t.Fatalf("%s: %v", scn.ID, err)
		}
		session, ok := engine.session(view.ID)
		if !ok {
			t.Fatalf("%s 会话不存在", scn.ID)
		}
		for _, command := range append(append([]string{}, scn.Presets...), scn.Evidence...) {
			result, err := engine.run(view.ID, session.c.render(command))
			if err != nil {
				t.Errorf("%s 命令 %q 执行失败：%v", scn.ID, command, err)
				continue
			}
			if strings.TrimSpace(result.Output) == "" {
				t.Errorf("%s 命令 %q 没有输出", scn.ID, command)
			}
		}
		for key := range scn.Commands {
			if _, err := engine.run(view.ID, session.c.render(key)); err != nil {
				t.Errorf("%s 场景命令 %q 执行失败：%v", scn.ID, key, err)
			}
		}
		result, updated, err := engine.submit(view.ID, scn.CorrectRootCause, scn.CorrectFix, progress)
		if err != nil {
			t.Fatalf("%s 提交失败：%v", scn.ID, err)
		}
		if !result.Passed {
			t.Errorf("%s 正确答案没有通过：%+v", scn.ID, result)
		}
		progress = updated
	}
	if progress.Solved != len(pack.Scenarios) {
		t.Fatalf("应累计 %d 个已复盘，实际 %d", len(pack.Scenarios), progress.Solved)
	}
}
