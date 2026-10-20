package backend

/*
线上事故演练场。

学习课程练的是"写对一段代码"，这里练的是"在一堆互相矛盾的现象里找到真正的根因"。
演练环境完全由 `labs/labs.json` 描述：服务拓扑、指标、日志、链路、中间件命令输出、
诊断选项与评分标准都在 JSON 里，所以新增事故不需要改程序，发一个内容包就能到达用户。

每日轮换不靠服务端调度：按日期取模选场景，再用日期做随机种子渲染订单号、Pod 名、
热点 Key 等细节。同一个日期打开永远是同一场事故，换一天就是另一场。
*/

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

const (
	labPackFile        = "labs.json"
	labPackVersion     = 1
	maxLabPackBytes    = 2 << 20
	maxLabCommandBytes = 400
	maxLabOutputBytes  = 24 << 10
	maxLabSessionCmds  = 300
	maxLabResults      = 200
	labArchiveDays     = 6
)

//go:embed labs/labs.json
var labFS embed.FS

var labIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,40}$`)
var labPlaceholder = regexp.MustCompile(`\{\{([A-Za-z][A-Za-z0-9]*)\}\}`)

type labPack struct {
	Version   int           `json:"version"`
	Company   string        `json:"company"`
	Scenarios []labScenario `json:"scenarios"`
}

type labService struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	Replicas string `json:"replicas,omitempty"`
	CPU      string `json:"cpu,omitempty"`
	Memory   string `json:"memory,omitempty"`
	Note     string `json:"note,omitempty"`
}

type labMetric struct {
	Service string    `json:"service"`
	Name    string    `json:"name"`
	Unit    string    `json:"unit,omitempty"`
	Alert   float64   `json:"alert,omitempty"`
	Points  []float64 `json:"points"`
}

type labLog struct {
	At      int               `json:"at"`
	Service string            `json:"service"`
	Level   string            `json:"level"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type labSpan struct {
	Service string `json:"service"`
	Op      string `json:"op"`
	MS      int    `json:"ms"`
	Note    string `json:"note,omitempty"`
}

type labTrace struct {
	OrderID string    `json:"orderId"`
	TotalMS int       `json:"totalMs"`
	Spans   []labSpan `json:"spans"`
}

type labOption struct {
	Label  string `json:"label"`
	Detail string `json:"detail,omitempty"`
}

type labScenario struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle,omitempty"`
	Brief    string   `json:"brief"`
	Mission  string   `json:"mission"`
	Business string   `json:"business,omitempty"`
	Level    string   `json:"level"`
	Minutes  int      `json:"minutes"`
	Tags     []string `json:"tags"`

	Alerts   []string `json:"alerts"`
	Timeline []string `json:"timeline,omitempty"`

	Services []labService `json:"services"`
	Metrics  []labMetric  `json:"metrics"`
	Logs     []labLog     `json:"logs"`
	Traces   []labTrace   `json:"traces"`

	Commands map[string]string `json:"commands"`
	Presets  []string          `json:"presets"`

	RootCauses       []labOption `json:"rootCauses"`
	CorrectRootCause int         `json:"correctRootCause"`
	Fixes            []labOption `json:"fixes"`
	CorrectFix       int         `json:"correctFix"`
	Hints            []string    `json:"hints"`
	Evidence         []string    `json:"evidence"`
	Debrief          string      `json:"debrief"`
}

func (s labScenario) alert() string {
	if len(s.Alerts) == 0 {
		return ""
	}
	return s.Alerts[0]
}

// readEmbeddedLabPack 读编译进程序的那一份场景。
func readEmbeddedLabPack() (labPack, error) {
	data, err := labFS.ReadFile("labs/" + labPackFile)
	if err != nil {
		return labPack{}, err
	}
	return parseLabPack(data)
}

func parseLabPack(data []byte) (labPack, error) {
	if len(data) > maxLabPackBytes {
		return labPack{}, errors.New("事故场景包超过大小上限")
	}
	var pack labPack
	dec := json.NewDecoder(bytes.NewReader(data))
	// 新字段必须跟着新版本程序一起发布，旧程序看到不认识的字段要明确拒绝，
	// 而不是带着半懂的状态继续跑。
	dec.DisallowUnknownFields()
	if err := dec.Decode(&pack); err != nil {
		return labPack{}, fmt.Errorf("事故场景 JSON 无法解析：%w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return labPack{}, errors.New("事故场景 JSON 只允许一个对象")
	}
	if err := validateLabPack(pack); err != nil {
		return labPack{}, err
	}
	return pack, nil
}

func validateLabPack(pack labPack) error {
	if pack.Version < 1 {
		return errors.New("事故场景包缺少版本号")
	}
	if strings.TrimSpace(pack.Company) == "" {
		return errors.New("事故场景包缺少公司名称")
	}
	if len(pack.Scenarios) == 0 || len(pack.Scenarios) > 200 {
		return fmt.Errorf("事故场景数量需在 1 到 200 之间，当前 %d", len(pack.Scenarios))
	}
	seen := map[string]bool{}
	for index := range pack.Scenarios {
		scn := &pack.Scenarios[index]
		if err := validateLabScenario(scn); err != nil {
			return fmt.Errorf("第 %d 个事故场景 %q：%w", index+1, scn.ID, err)
		}
		if seen[scn.ID] {
			return fmt.Errorf("事故场景 ID 重复：%s", scn.ID)
		}
		seen[scn.ID] = true
	}
	return nil
}

func validateLabScenario(scn *labScenario) error {
	if !labIDPattern.MatchString(scn.ID) {
		return errors.New("id 需为小写字母、数字或连字符")
	}
	for name, value := range map[string]string{
		"title": scn.Title, "brief": scn.Brief, "mission": scn.Mission,
		"level": scn.Level, "debrief": scn.Debrief,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s 不能为空", name)
		}
	}
	if scn.Minutes < 5 || scn.Minutes > 120 {
		return errors.New("minutes 需在 5 到 120 之间")
	}
	if len(scn.Tags) == 0 {
		return errors.New("至少需要一个标签")
	}
	if len(scn.Alerts) == 0 || len(scn.Alerts) > 12 {
		return errors.New("alerts 需要 1 到 12 条")
	}
	if len(scn.Services) < 3 || len(scn.Services) > 24 {
		return errors.New("services 需要 3 到 24 个")
	}
	for _, svc := range scn.Services {
		if strings.TrimSpace(svc.Name) == "" || strings.TrimSpace(svc.Role) == "" {
			return errors.New("服务名与角色不能为空")
		}
		switch svc.Status {
		case "healthy", "degraded", "critical":
		default:
			return fmt.Errorf("服务 %s 的 status 只能是 healthy、degraded 或 critical", svc.Name)
		}
	}
	if len(scn.Metrics) < 2 || len(scn.Metrics) > 24 {
		return errors.New("metrics 需要 2 到 24 组")
	}
	for _, metric := range scn.Metrics {
		if strings.TrimSpace(metric.Service) == "" || strings.TrimSpace(metric.Name) == "" {
			return errors.New("指标的服务与名称不能为空")
		}
		if len(metric.Points) < 4 || len(metric.Points) > 120 {
			return errors.New("指标采样点需要 4 到 120 个")
		}
		for _, point := range metric.Points {
			if math.IsNaN(point) || math.IsInf(point, 0) || point < 0 {
				return fmt.Errorf("指标 %s 含有非法数值", metric.Name)
			}
		}
	}
	if len(scn.Logs) < 5 || len(scn.Logs) > 400 {
		return errors.New("logs 需要 5 到 400 条")
	}
	for _, entry := range scn.Logs {
		if entry.At < 0 || entry.At > 24*60 {
			return errors.New("日志的 at 需在 0 到 1440 分钟之间")
		}
		if strings.TrimSpace(entry.Service) == "" || strings.TrimSpace(entry.Message) == "" {
			return errors.New("日志的服务与内容不能为空")
		}
		switch entry.Level {
		case "DEBUG", "INFO", "WARN", "ERROR":
		default:
			return fmt.Errorf("日志级别只能是 DEBUG、INFO、WARN、ERROR，当前 %q", entry.Level)
		}
	}
	if len(scn.Traces) > 12 {
		return errors.New("traces 最多 12 条")
	}
	for _, trace := range scn.Traces {
		if strings.TrimSpace(trace.OrderID) == "" || len(trace.Spans) < 2 || len(trace.Spans) > 32 {
			return errors.New("链路需要订单号与 2 到 32 个 span")
		}
	}
	if len(scn.Commands) < 3 || len(scn.Commands) > 60 {
		return errors.New("commands 需要 3 到 60 条")
	}
	normalized := map[string]bool{}
	for key, value := range scn.Commands {
		name := normalizeLabCommand(key)
		if name == "" || strings.TrimSpace(value) == "" {
			return errors.New("commands 的键与输出不能为空")
		}
		if len(value) > 8<<10 {
			return fmt.Errorf("命令 %q 的输出超过 8 KiB", key)
		}
		if normalized[name] {
			return fmt.Errorf("命令规范化后重复：%s", key)
		}
		normalized[name] = true
	}
	if len(scn.Presets) == 0 || len(scn.Presets) > 12 {
		return errors.New("presets 需要 1 到 12 条")
	}
	if len(scn.RootCauses) < 3 || len(scn.RootCauses) > 6 {
		return errors.New("rootCauses 需要 3 到 6 个选项")
	}
	if scn.CorrectRootCause < 0 || scn.CorrectRootCause >= len(scn.RootCauses) {
		return errors.New("correctRootCause 下标越界")
	}
	if len(scn.Fixes) < 3 || len(scn.Fixes) > 6 {
		return errors.New("fixes 需要 3 到 6 个选项")
	}
	if scn.CorrectFix < 0 || scn.CorrectFix >= len(scn.Fixes) {
		return errors.New("correctFix 下标越界")
	}
	if len(scn.Hints) < 2 || len(scn.Hints) > 6 {
		return errors.New("hints 需要 2 到 6 条")
	}
	if len(scn.Evidence) == 0 || len(scn.Evidence) > 10 {
		return errors.New("evidence 需要 1 到 10 条")
	}
	for _, option := range append(append([]labOption{}, scn.RootCauses...), scn.Fixes...) {
		if strings.TrimSpace(option.Label) == "" {
			return errors.New("诊断选项不能为空")
		}
	}
	return nil
}

// mergeLabPacks 用内容包里的场景覆盖内置场景；没覆盖到的内置场景继续保留。
func mergeLabPacks(base, override labPack) labPack {
	merged := base
	merged.Version = override.Version
	if strings.TrimSpace(override.Company) != "" {
		merged.Company = override.Company
	}
	index := make(map[string]int, len(merged.Scenarios))
	for i, scn := range merged.Scenarios {
		index[scn.ID] = i
	}
	for _, scn := range override.Scenarios {
		if i, ok := index[scn.ID]; ok {
			merged.Scenarios[i] = scn
			continue
		}
		merged.Scenarios = append(merged.Scenarios, scn)
		index[scn.ID] = len(merged.Scenarios) - 1
	}
	return merged
}

// loadLabPack 读内置场景，再用数据目录 content/labs.json 覆盖。
func loadLabPack(overrideDir string) (labPack, []string, error) {
	pack, err := readEmbeddedLabPack()
	if err != nil {
		return labPack{}, nil, err
	}
	warnings := []string{}
	if overrideDir == "" {
		return pack, warnings, nil
	}
	data, err := os.ReadFile(filepath.Join(overrideDir, labPackFile))
	if errors.Is(err, os.ErrNotExist) {
		return pack, warnings, nil
	}
	if err != nil {
		return pack, []string{"事故场景包读取失败，继续使用内置场景"}, nil
	}
	override, err := parseLabPack(data)
	if err != nil {
		return pack, []string{"事故场景包校验失败，继续使用内置场景：" + err.Error()}, nil
	}
	return mergeLabPacks(pack, override), warnings, nil
}

/*
labCase 是一场被"实例化"的事故：同一个场景配上具体日期与随机种子，
订单号、Pod 名、热点 Key、指标读数都会跟着变，避免每天看到一模一样的数字。
*/
type labCase struct {
	ID       string
	Scenario labScenario
	Date     time.Time
	End      time.Time
	Seed     uint64
	Values   map[string]string
	Commands map[string]string
}

func labCaseID(scenarioID string, date time.Time) string {
	return scenarioID + "-" + date.Format("20060102")
}

func newLabCase(scn labScenario, date, now time.Time) *labCase {
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	end := time.Date(date.Year(), date.Month(), date.Day(), 23, 30, 0, 0, date.Location())
	if sameDate(date, now) {
		// 今天的演练时钟贴近用户打开应用的时刻，日志时间看起来才像"刚刚发生"。
		end = now.Truncate(time.Minute)
	}
	seed := labSeed(date, scn.ID)
	c := &labCase{
		ID:       labCaseID(scn.ID, date),
		Scenario: scn,
		Date:     date,
		End:      end,
		Seed:     seed,
		Commands: map[string]string{},
	}
	c.Values = labValues(scn, date, end, seed)
	for key, value := range scn.Commands {
		// 场景作者可以写 kubectl exec {{pod}} ...；这里先把占位符换成这一天的真实值，
		// 用户从快捷命令或证据里点出来的才是可执行的那一条。
		c.Commands[normalizeLabCommand(c.render(key))] = value
	}
	return c
}

func sameDate(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

func labSeed(date time.Time, id string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(date.Format("2006-01-02") + ":" + id))
	return h.Sum64()
}

func labValues(scn labScenario, date, end time.Time, seed uint64) map[string]string {
	pod := func(service string, index int) string {
		h := fnv.New32a()
		_, _ = h.Write([]byte(fmt.Sprintf("%d:%s:%d", seed, service, index)))
		suffix := []string{"x2k4p", "m8q1z", "p7r3d"}[index%3]
		return fmt.Sprintf("%s-%05x-%s", service, h.Sum32()&0xfffff, suffix)
	}
	values := map[string]string{
		"date":        date.Format("2006-01-02"),
		"dateCompact": date.Format("20060102"),
		"clock":       end.Format("15:04"),
		"timestamp":   end.Format("2006-01-02 15:04:05"),
		"orderId":     fmt.Sprintf("ORD%s%04d", date.Format("20060102"), seed%10000),
		"traceId":     fmt.Sprintf("%016x", seed),
		"pod":         pod("order-api", 0),
		"gatewayPod":  pod("gateway", 0),
		"orderPod":    pod("order-api", 0),
		"workerPod":   pod("order-worker", 0),
		"redisPod":    pod("redis-cache", 0),
		"kafkaPod":    pod("kafka", 0),
		"mysqlPod":    pod("mysql", 0),
		"paymentPod":  pod("payment-api", 0),
		"node":        fmt.Sprintf("node-prod-a%d", seed%3+1),
		"hotkey":      fmt.Sprintf("sku:%06d", seed%900000+100000),
		"sku":         fmt.Sprintf("SKU-%06d", seed%900000+100000),
		"amount":      fmt.Sprintf("%d.%02d", seed%400+59, seed%100),
		"lag":         strconv.FormatInt(int64(seed%90000+8000), 10),
		"p99":         strconv.FormatInt(int64(seed%5000+3200), 10),
		"qps":         strconv.FormatInt(int64(seed%12000+8000), 10),
		"errorRate":   fmt.Sprintf("%.1f", float64(seed%180)/10+1),
		"customerId":  fmt.Sprintf("C%07d", seed%9000000+1000000),
	}
	for index, svc := range scn.Services {
		name := strings.ReplaceAll(svc.Name, "-", "")
		if _, ok := values[name+"Pod"]; !ok {
			values[name+"Pod"] = pod(svc.Name, index)
		}
	}
	return values
}

func (c *labCase) render(value string) string {
	if !strings.Contains(value, "{{") {
		return value
	}
	return labPlaceholder.ReplaceAllStringFunc(value, func(match string) string {
		key := match[2 : len(match)-2]
		if replacement, ok := c.Values[key]; ok {
			return replacement
		}
		return match
	})
}

func normalizeLabCommand(raw string) string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(raw)))
	if len(fields) == 0 {
		return ""
	}
	switch fields[0] {
	case "redis-cli":
		fields[0] = "redis"
	case "kafka-consumer-groups", "kafka-topics":
		fields[0] = "kafka"
	case "mysqladmin", "mysqlslap":
		fields[0] = "mysql"
	case "k":
		fields[0] = "kubectl"
	}
	return strings.Join(fields, " ")
}

func podNameFor(c *labCase, service string, index int) string {
	name := strings.ReplaceAll(service, "-", "")
	if value, ok := c.Values[name+"Pod"]; ok && index == 0 {
		return value
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s:%d", c.Seed, service, index)))
	suffix := []string{"x2k4p", "m8q1z", "p7r3d"}[index%3]
	return fmt.Sprintf("%s-%05x-%s", service, h.Sum32()&0xfffff, suffix)
}

/* ---------------- 引擎与进度 ---------------- */

type labSession struct {
	caseID   string
	c        *labCase
	commands []string
	hints    int
}

type labEngine struct {
	mu       sync.Mutex
	pack     labPack
	warnings []string
	sessions map[string]*labSession
}

func newLabEngine(overrideDir string) (*labEngine, error) {
	pack, warnings, err := loadLabPack(overrideDir)
	if err != nil {
		return nil, err
	}
	return &labEngine{pack: pack, warnings: warnings, sessions: map[string]*labSession{}}, nil
}

func (l *labEngine) reload(overrideDir string) error {
	pack, warnings, err := loadLabPack(overrideDir)
	if err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.pack = pack
	l.warnings = warnings
	// 正在进行的演练持有旧场景快照，继续用没有问题；清空会话只是让新内容立刻生效。
	l.sessions = map[string]*labSession{}
	return nil
}

func (l *labEngine) warningsCopy() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]string{}, l.warnings...)
}

type LabProgress struct {
	Results  map[string]LabResult `json:"results,omitempty"`
	LastCase string               `json:"lastCase,omitempty"`
	Solved   int                  `json:"solved"`
	Attempts int                  `json:"attempts"`
}

type LabResult struct {
	CaseID    string `json:"caseId"`
	Scenario  string `json:"scenario"`
	Date      string `json:"date"`
	Title     string `json:"title"`
	Passed    bool   `json:"passed"`
	Best      int    `json:"best"`
	Attempts  int    `json:"attempts"`
	Hints     int    `json:"hints"`
	UpdatedAt int64  `json:"updatedAt"`
}

func normalizeLabProgress(p LabProgress) LabProgress {
	if p.Results == nil {
		p.Results = map[string]LabResult{}
	}
	if len(p.Results) > maxLabResults {
		type pair struct {
			key    string
			result LabResult
		}
		items := make([]pair, 0, len(p.Results))
		for key, result := range p.Results {
			items = append(items, pair{key, result})
		}
		sort.Slice(items, func(i, j int) bool { return items[i].result.UpdatedAt > items[j].result.UpdatedAt })
		p.Results = make(map[string]LabResult, maxLabResults)
		for _, item := range items[:maxLabResults] {
			p.Results[item.key] = item.result
		}
	}
	p.Solved = 0
	p.Attempts = 0
	for _, result := range p.Results {
		if result.Passed {
			p.Solved++
		}
		p.Attempts += result.Attempts
	}
	return p
}

func cloneLabProgress(p LabProgress) LabProgress {
	out := p
	out.Results = make(map[string]LabResult, len(p.Results))
	for key, result := range p.Results {
		out.Results[key] = result
	}
	return out
}

func applyLabAttempt(progress LabProgress, attempt LabResult) LabProgress {
	progress = cloneLabProgress(progress)
	if progress.Results == nil {
		progress.Results = map[string]LabResult{}
	}
	if previous, ok := progress.Results[attempt.CaseID]; ok {
		attempt.Passed = attempt.Passed || previous.Passed
		attempt.Best = maxInt(attempt.Best, previous.Best)
		attempt.Attempts += previous.Attempts
		attempt.Hints = maxInt(attempt.Hints, previous.Hints)
	}
	progress.Results[attempt.CaseID] = attempt
	progress.LastCase = attempt.CaseID
	return normalizeLabProgress(progress)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

/* ---------------- 每日轮换与视图 ---------------- */

type LabCaseSummary struct {
	ID         string   `json:"id"`
	Scenario   string   `json:"scenario"`
	Date       string   `json:"date"`
	Title      string   `json:"title"`
	Subtitle   string   `json:"subtitle,omitempty"`
	Level      string   `json:"level"`
	Minutes    int      `json:"minutes"`
	Tags       []string `json:"tags"`
	Alert      string   `json:"alert,omitempty"`
	Mode       string   `json:"mode"`
	Solved     bool     `json:"solved"`
	Best       int      `json:"best,omitempty"`
	Attempts   int      `json:"attempts,omitempty"`
	Hints      int      `json:"hints,omitempty"`
	IsToday    bool     `json:"isToday,omitempty"`
	ScenarioID string   `json:"scenarioId"`
}

type LabBootstrap struct {
	Company       string           `json:"company"`
	Tagline       string           `json:"tagline"`
	Today         LabCaseSummary   `json:"today"`
	Archive       []LabCaseSummary `json:"archive"`
	Catalog       []LabCaseSummary `json:"catalog"`
	Progress      LabProgress      `json:"progress"`
	ModelReady    bool             `json:"modelReady"`
	ScenarioCount int              `json:"scenarioCount"`
	PackVersion   int              `json:"packVersion"`
	Warnings      []string         `json:"warnings,omitempty"`
}

type LabCaseView struct {
	ID         string       `json:"id"`
	Scenario   string       `json:"scenario"`
	Company    string       `json:"company"`
	Date       string       `json:"date"`
	Clock      string       `json:"clock"`
	Title      string       `json:"title"`
	Subtitle   string       `json:"subtitle,omitempty"`
	Brief      string       `json:"brief"`
	Mission    string       `json:"mission"`
	Business   string       `json:"business,omitempty"`
	Level      string       `json:"level"`
	Minutes    int          `json:"minutes"`
	Tags       []string     `json:"tags"`
	Alerts     []string     `json:"alerts"`
	Timeline   []string     `json:"timeline,omitempty"`
	Services   []labService `json:"services"`
	Metrics    []labMetric  `json:"metrics"`
	Presets    []string     `json:"presets"`
	RootCauses []labOption  `json:"rootCauses"`
	Fixes      []labOption  `json:"fixes"`
	Hints      []string     `json:"hints"`
	HintCount  int          `json:"hintCount"`
	Result     *LabResult   `json:"result,omitempty"`
	History    []labHistory `json:"history,omitempty"`
}

type labHistory struct {
	Command string `json:"command"`
}

type LabRunResult struct {
	Command string `json:"command"`
	Output  string `json:"output"`
	At      int64  `json:"at"`
}

type LabSubmitResult struct {
	Passed           bool        `json:"passed"`
	Score            int         `json:"score"`
	Best             int         `json:"best"`
	RootCauseCorrect bool        `json:"rootCauseCorrect"`
	FixCorrect       bool        `json:"fixCorrect"`
	CorrectRootCause int         `json:"correctRootCause"`
	CorrectFix       int         `json:"correctFix"`
	EvidenceHit      int         `json:"evidenceHit"`
	EvidenceTotal    int         `json:"evidenceTotal"`
	Hints            int         `json:"hints"`
	Message          string      `json:"message"`
	Debrief          string      `json:"debrief"`
	Result           LabResult   `json:"result"`
	Progress         LabProgress `json:"progress"`
}

type LabHintResult struct {
	Hint      string `json:"hint"`
	Index     int    `json:"index"`
	HintsUsed int    `json:"hintsUsed"`
	Total     int    `json:"total"`
}

func (l *labEngine) packSnapshot() labPack {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pack
}

func scenarioIndexFor(date time.Time, count int) int {
	if count <= 0 {
		return 0
	}
	days := date.Unix() / 86400
	index := int(days % int64(count))
	if index < 0 {
		index += count
	}
	return index
}

func (l *labEngine) bootstrap(progress LabProgress, modelReady bool, now time.Time) LabBootstrap {
	pack := l.packSnapshot()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	index := scenarioIndexFor(today, len(pack.Scenarios))
	boot := LabBootstrap{
		Company:       pack.Company,
		Tagline:       "真实链路不会按教材出故障。这里每天一场事故，从告警、日志、指标和链路里把根因挖出来。",
		Progress:      normalizeLabProgress(progress),
		ModelReady:    modelReady,
		ScenarioCount: len(pack.Scenarios),
		PackVersion:   pack.Version,
		Warnings:      l.warningsCopy(),
	}
	boot.Today = l.summary(pack.Scenarios[index], today, "今日事故", boot.Progress, now, true)
	for offset := 1; offset <= labArchiveDays; offset++ {
		date := today.AddDate(0, 0, -offset)
		scn := pack.Scenarios[scenarioIndexFor(date, len(pack.Scenarios))]
		boot.Archive = append(boot.Archive, l.summary(scn, date, "历史存档", boot.Progress, now, false))
	}
	for _, scn := range pack.Scenarios {
		boot.Catalog = append(boot.Catalog, l.summary(scn, today, "场景库", boot.Progress, now, scn.ID == boot.Today.Scenario))
	}
	return boot
}

func (l *labEngine) summary(scn labScenario, date time.Time, mode string, progress LabProgress, now time.Time, isToday bool) LabCaseSummary {
	c := newLabCase(scn, date, now)
	id := c.ID
	summary := LabCaseSummary{
		ID:         id,
		Scenario:   scn.ID,
		ScenarioID: scn.ID,
		Date:       c.Date.Format("2006-01-02"),
		Title:      c.render(scn.Title),
		Subtitle:   c.render(scn.Subtitle),
		Level:      scn.Level,
		Minutes:    scn.Minutes,
		Tags:       append([]string{}, scn.Tags...),
		Alert:      c.render(scn.alert()),
		Mode:       mode,
		IsToday:    isToday,
	}
	if result, ok := progress.Results[id]; ok {
		summary.Solved = result.Passed
		summary.Best = result.Best
		summary.Attempts = result.Attempts
		summary.Hints = result.Hints
	}
	return summary
}

func (l *labEngine) findScenario(id string) (labScenario, bool) {
	pack := l.packSnapshot()
	for _, scn := range pack.Scenarios {
		if scn.ID == id {
			return scn, true
		}
	}
	return labScenario{}, false
}

func (l *labEngine) openCase(scenarioID, dateRaw string, progress LabProgress, now time.Time) (*LabCaseView, error) {
	scn, ok := l.findScenario(scenarioID)
	if !ok {
		return nil, errors.New("事故场景不存在")
	}
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if strings.TrimSpace(dateRaw) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(dateRaw), now.Location())
		if err != nil {
			return nil, errors.New("日期格式应为 2006-01-02")
		}
		date = parsed
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	if date.After(today) {
		date = today
	}
	c := newLabCase(scn, date, now)
	l.mu.Lock()
	session, ok := l.sessions[c.ID]
	if !ok {
		session = &labSession{caseID: c.ID, c: c}
		l.sessions[c.ID] = session
	}
	if result, ok := progress.Results[c.ID]; ok && result.Hints > session.hints {
		session.hints = minInt(result.Hints, len(scn.Hints))
	}
	hints := append([]string{}, scn.Hints[:session.hints]...)
	history := make([]labHistory, 0, len(session.commands))
	for _, command := range session.commands {
		history = append(history, labHistory{Command: command})
	}
	l.mu.Unlock()
	pack := l.packSnapshot()
	view := &LabCaseView{
		ID:         c.ID,
		Scenario:   scn.ID,
		Company:    pack.Company,
		Date:       c.Date.Format("2006-01-02"),
		Clock:      c.End.Format("2006-01-02 15:04"),
		Title:      c.render(scn.Title),
		Subtitle:   c.render(scn.Subtitle),
		Brief:      c.render(scn.Brief),
		Mission:    c.render(scn.Mission),
		Business:   c.render(scn.Business),
		Level:      scn.Level,
		Minutes:    scn.Minutes,
		Tags:       append([]string{}, scn.Tags...),
		Presets:    make([]string, 0, len(scn.Presets)),
		RootCauses: make([]labOption, 0, len(scn.RootCauses)),
		Fixes:      make([]labOption, 0, len(scn.Fixes)),
		Hints:      hints,
		HintCount:  len(scn.Hints),
		History:    history,
	}
	for _, alert := range scn.Alerts {
		view.Alerts = append(view.Alerts, c.render(alert))
	}
	for _, item := range scn.Timeline {
		view.Timeline = append(view.Timeline, c.render(item))
	}
	for _, svc := range scn.Services {
		svc.Note = c.render(svc.Note)
		view.Services = append(view.Services, svc)
	}
	for _, metric := range scn.Metrics {
		view.Metrics = append(view.Metrics, c.varyMetric(metric))
	}
	for _, preset := range scn.Presets {
		view.Presets = append(view.Presets, c.render(preset))
	}
	for _, option := range scn.RootCauses {
		view.RootCauses = append(view.RootCauses, labOption{Label: c.render(option.Label), Detail: c.render(option.Detail)})
	}
	for _, option := range scn.Fixes {
		view.Fixes = append(view.Fixes, labOption{Label: c.render(option.Label), Detail: c.render(option.Detail)})
	}
	if result, ok := progress.Results[c.ID]; ok {
		copied := result
		view.Result = &copied
	}
	return view, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

/*
varyMetric 给同一场景的每日实例一点确定性的扰动：日期不同，曲线形状略有差异，
但最后一点保持原值，告警是否触发不会因为随机数而变。整数指标仍然取整，
避免出现“47.3 个 under-replicated 分区”这种不真实的读数。
*/
func (c *labCase) varyMetric(metric labMetric) labMetric {
	if len(metric.Points) < 2 {
		return metric
	}
	integer := true
	for _, point := range metric.Points {
		if math.Trunc(point) != point {
			integer = false
			break
		}
	}
	factor := 0.94 + float64(c.Seed%11)/100
	out := metric
	out.Points = append([]float64{}, metric.Points...)
	for index := 0; index < len(out.Points)-1; index++ {
		value := metric.Points[index] * factor
		if (c.Seed>>uint(index%16))&1 == 1 {
			value = metric.Points[index] / factor
		}
		if integer {
			value = math.Round(value)
		}
		out.Points[index] = value
	}
	return out
}

func (l *labEngine) session(caseID string) (*labSession, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	session, ok := l.sessions[caseID]
	return session, ok
}

func (l *labEngine) run(caseID, command string) (LabRunResult, error) {
	if strings.TrimSpace(command) == "" {
		return LabRunResult{}, errors.New("请输入命令")
	}
	if len(command) > maxLabCommandBytes {
		return LabRunResult{}, errors.New("命令过长")
	}
	l.mu.Lock()
	session, ok := l.sessions[caseID]
	if !ok {
		l.mu.Unlock()
		return LabRunResult{}, errors.New("演练会话已失效，请重新打开这场事故")
	}
	l.mu.Unlock()
	output, err := session.c.run(command)
	if err != nil {
		return LabRunResult{}, err
	}
	normalized := normalizeLabCommand(command)
	l.mu.Lock()
	session.commands = append(session.commands, normalized)
	if len(session.commands) > maxLabSessionCmds {
		session.commands = session.commands[len(session.commands)-maxLabSessionCmds:]
	}
	l.mu.Unlock()
	return LabRunResult{Command: command, Output: output, At: time.Now().Unix()}, nil
}

func (l *labEngine) inspect(caseID, command string) string {
	l.mu.Lock()
	session, ok := l.sessions[caseID]
	l.mu.Unlock()
	if !ok {
		return "演练会话已失效。"
	}
	output, err := session.c.run(command)
	if err != nil {
		return err.Error()
	}
	return output
}

func (l *labEngine) transcript(caseID string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if session, ok := l.sessions[caseID]; ok {
		return append([]string{}, session.commands...)
	}
	return nil
}

func (l *labEngine) hint(caseID string) (LabHintResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	session, ok := l.sessions[caseID]
	if !ok {
		return LabHintResult{}, errors.New("演练会话已失效，请重新打开这场事故")
	}
	total := len(session.c.Scenario.Hints)
	if session.hints >= total {
		return LabHintResult{}, errors.New("提示已经全部用完，试着把日志时间线和指标变化对齐")
	}
	hint := session.c.render(session.c.Scenario.Hints[session.hints])
	session.hints++
	return LabHintResult{Hint: hint, Index: session.hints - 1, HintsUsed: session.hints, Total: total}, nil
}

func (l *labEngine) reset(caseID string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	session, ok := l.sessions[caseID]
	if !ok {
		return errors.New("演练会话已失效，请重新打开这场事故")
	}
	session.commands = nil
	session.hints = 0
	return nil
}

func (l *labEngine) submit(caseID string, root, fix int, progress LabProgress) (LabSubmitResult, LabProgress, error) {
	l.mu.Lock()
	session, ok := l.sessions[caseID]
	if !ok {
		l.mu.Unlock()
		return LabSubmitResult{}, progress, errors.New("演练会话已失效，请重新打开这场事故")
	}
	commands := append([]string{}, session.commands...)
	hints := session.hints
	l.mu.Unlock()
	scn := session.c.Scenario
	if root < 0 || root >= len(scn.RootCauses) || fix < 0 || fix >= len(scn.Fixes) {
		return LabSubmitResult{}, progress, errors.New("请选择根因与处置方案")
	}
	hit := 0
	for _, evidence := range scn.Evidence {
		needle := normalizeLabCommand(session.c.render(evidence))
		for _, command := range commands {
			if strings.Contains(command, needle) {
				hit++
				break
			}
		}
	}
	rootCorrect := root == scn.CorrectRootCause
	fixCorrect := fix == scn.CorrectFix
	base := 0
	if rootCorrect {
		base += 50
	}
	if fixCorrect {
		base += 20
	}
	if total := len(scn.Evidence); total > 0 {
		base += int(math.Round(30 * float64(hit) / float64(total)))
	}
	penalty := minInt(15, hints*5)
	score := base - penalty
	if score < 0 {
		score = 0
	}
	passed := rootCorrect && fixCorrect
	message := "还差得比较远：先回到告警时间线，确认哪个依赖最先异常，再看指标是容量、延迟还是错误。"
	switch {
	case passed && score >= 90:
		message = "根因和处置都对，关键证据也查得扎实。这一次可以写进复盘了。"
	case passed:
		message = "方向对了，处置也正确。关键证据还可以再补几条，下次会更快。"
	case rootCorrect:
		message = "根因判断正确，但处置方案会留下隐患。想想怎么在不扩大影响面的前提下止血。"
	case fixCorrect:
		message = "处置能暂时止血，但根因判断偏了；同样的问题还会再来。"
	}
	result := LabResult{
		CaseID:    caseID,
		Scenario:  scn.ID,
		Date:      session.c.Date.Format("2006-01-02"),
		Title:     session.c.render(scn.Title),
		Passed:    passed,
		Best:      score,
		Attempts:  1,
		Hints:     hints,
		UpdatedAt: time.Now().Unix(),
	}
	updated := applyLabAttempt(progress, result)
	if previous, ok := updated.Results[caseID]; ok {
		result = previous
	}
	return LabSubmitResult{
		Passed:           passed,
		Score:            score,
		Best:             result.Best,
		RootCauseCorrect: rootCorrect,
		FixCorrect:       fixCorrect,
		CorrectRootCause: scn.CorrectRootCause,
		CorrectFix:       scn.CorrectFix,
		EvidenceHit:      hit,
		EvidenceTotal:    len(scn.Evidence),
		Hints:            hints,
		Message:          message,
		Debrief:          session.c.render(scn.Debrief),
		Result:           result,
		Progress:         updated,
	}, updated, nil
}

/* ---------------- 命令执行 ---------------- */

func (c *labCase) run(raw string) (string, error) {
	command := normalizeLabCommand(raw)
	if command == "" {
		return "", errors.New("请输入命令")
	}
	switch command {
	case "help", "?":
		return c.help(), nil
	case "submit":
		return "诊断请用右侧的「提交诊断」表单：先选根因，再选处置方案。终端里的证据会被计入评分。", nil
	case "date", "clock", "now":
		return "演练时钟：" + c.End.Format("2006-01-02 15:04:05") + "\n（日志时间是相对这一刻往前推的）", nil
	}
	// 场景作者显式提供的命令优先：其中也可能有 kubectl exec 这类只读排查方式，
	// 不能被下面的“变更拦截”误伤。
	if output, ok := c.Commands[command]; ok {
		return capLabOutput(c.render(output)), nil
	}
	if labMutating(command) {
		return "演练环境是只读的：这里不真正执行变更。请把处置方案写进右侧诊断，真实环境里变更前也要先评估影响面。", nil
	}
	fields := strings.Fields(command)
	switch fields[0] {
	case "logs":
		return c.logsCommand(fields[1:])
	case "metrics", "metric":
		return c.metricsCommand(fields[1:])
	case "trace", "traces":
		return c.traceCommand(strings.Join(fields[1:], " ")), nil
	case "services", "status", "top":
		return c.servicesOutput(), nil
	case "deploys", "deploy":
		if len(fields) == 1 || fields[1] == "history" {
			if output, ok := c.Commands["deploy history"]; ok {
				return capLabOutput(c.render(output)), nil
			}
			return c.timelineOutput(), nil
		}
	case "kubectl":
		return c.kubectlCommand(fields[1:])
	}
	suggestions := c.suggestions(fields[0])
	if suggestions == "" {
		return "", errors.New("无法识别这条命令。输入 help 查看可用命令")
	}
	return "", fmt.Errorf("无法识别这条命令。你可能想输入：\n%s", suggestions)
}

func capLabOutput(output string) string {
	if len(output) <= maxLabOutputBytes {
		return output
	}
	return truncate(output, maxLabOutputBytes) + "\n…（输出已截断，请缩小查询范围）"
}

// labMutating 只拦真正有副作用的动词；不能用子串匹配，
// 否则 seckill-api 会被当成 kill，kubectl exec（只读排查）也会被误伤。
func labMutating(command string) bool {
	for _, prefix := range []string{
		"kubectl delete ", "kubectl apply ", "kubectl patch ", "kubectl edit ",
		"kubectl scale ", "kubectl drain ", "kubectl cordon ", "kubectl uncordon ",
		"kubectl taint ", "kubectl rollout restart ", "kubectl rollout undo ",
		"kubectl rollout pause ", "kubectl rollout resume ",
		"redis shutdown", "redis flushall", "redis flushdb", "redis del ",
		"redis config set ", "redis expire ", "redis rename ",
		"mysql drop ", "mysql truncate ", "mysql delete ", "mysql update ",
		"mysql insert ", "mysql alter ", "mysql create ",
		"kafka topics --delete", "kafka consumer-groups --delete",
		"delete ", "restart ", "scale ", "kill ", "drain ", "rollback ",
	} {
		if strings.HasPrefix(command, prefix) {
			return true
		}
	}
	return false
}

func (c *labCase) help() string {
	var b strings.Builder
	b.WriteString("可用命令\n")
	b.WriteString("  help                                查看这份帮助\n")
	b.WriteString("  services                            服务与实例状态\n")
	b.WriteString("  logs <服务> [--level=ERROR] [--grep=关键词] [--since=30m] [--limit=50]\n")
	b.WriteString("  metrics [服务] [指标名]             看指标曲线与统计值\n")
	b.WriteString("  trace --slow | trace <订单号>       看慢链路或指定订单的 span\n")
	b.WriteString("  kubectl get pods | describe pod <名> | logs <名> | top pods\n")
	b.WriteString("  deploy history                      发布与变更记录\n")
	b.WriteString("  date                                演练时钟\n")
	b.WriteString("\n场景专用命令\n")
	keys := make([]string, 0, len(c.Commands))
	for key := range c.Commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteString("  " + key + "\n")
	}
	if len(c.Scenario.Presets) > 0 {
		b.WriteString("\n建议从这些命令开始\n")
		for _, preset := range c.Scenario.Presets {
			b.WriteString("  " + c.render(preset) + "\n")
		}
	}
	return b.String()
}

func (c *labCase) suggestions(first string) string {
	lines := []string{}
	for key := range c.Commands {
		if strings.HasPrefix(key, first) {
			lines = append(lines, "  "+key)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n")
}

func (c *labCase) servicesOutput() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "服务\t角色\t状态\t副本\tCPU\t内存\t说明")
	for _, svc := range c.Scenario.Services {
		status := map[string]string{"healthy": "正常", "degraded": "降级", "critical": "严重"}[svc.Status]
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			svc.Name, c.render(svc.Role), status, svc.Replicas, svc.CPU, svc.Memory, c.render(svc.Note))
	}
	_ = w.Flush()
	b.WriteString("\n提示：状态是入口视角的聚合，不代表依赖没问题；继续用 logs、metrics 和中间件命令确认。\n")
	return b.String()
}

func (c *labCase) timelineOutput() string {
	if len(c.Scenario.Timeline) == 0 {
		return "没有额外的变更记录。"
	}
	var b strings.Builder
	b.WriteString("变更与告警时间线\n")
	for _, item := range c.Scenario.Timeline {
		b.WriteString("  " + c.render(item) + "\n")
	}
	return b.String()
}

func (c *labCase) parseFlags(args []string) (positional []string, flags map[string]string) {
	flags = map[string]string{}
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			pair := strings.TrimPrefix(arg, "--")
			if index := strings.Index(pair, "="); index >= 0 {
				flags[pair[:index]] = pair[index+1:]
			} else {
				flags[pair] = "true"
			}
			continue
		}
		positional = append(positional, arg)
	}
	return positional, flags
}

func parseLabMinutes(raw string, fallback int) int {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return fallback
	}
	multiplier := 1
	if strings.HasSuffix(raw, "h") {
		multiplier = 60
		raw = strings.TrimSuffix(raw, "h")
	} else if strings.HasSuffix(raw, "m") {
		raw = strings.TrimSuffix(raw, "m")
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value * multiplier
}

func (c *labCase) logsCommand(args []string) (string, error) {
	positional, flags := c.parseFlags(args)
	service := strings.TrimSpace(flags["service"])
	if service == "" && len(positional) > 0 {
		service = positional[0]
	}
	level := strings.ToUpper(strings.TrimSpace(flags["level"]))
	grep := strings.TrimSpace(flags["grep"])
	since := parseLabMinutes(flags["since"], 24*60)
	limit := 50
	if raw, ok := flags["limit"]; ok {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			limit = minInt(value, 200)
		}
	}
	if service != "" && service != "all" {
		found := false
		for _, svc := range c.Scenario.Services {
			if svc.Name == service {
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("没有这个服务：%s\n可用服务：%s", service, c.serviceNames())
		}
	}
	return c.logsOutput(service, level, grep, since, limit), nil
}

func (c *labCase) serviceNames() string {
	names := make([]string, 0, len(c.Scenario.Services))
	for _, svc := range c.Scenario.Services {
		names = append(names, svc.Name)
	}
	return strings.Join(names, ", ")
}

func (c *labCase) logsOutput(service, level, grep string, since, limit int) string {
	type matched struct {
		at      int
		service string
		level   string
		message string
		fields  map[string]string
	}
	items := []matched{}
	for _, entry := range c.Scenario.Logs {
		if entry.At > since {
			continue
		}
		if service != "" && service != "all" && entry.Service != service {
			continue
		}
		if level != "" && entry.Level != level {
			continue
		}
		if grep != "" {
			haystack := c.render(entry.Message)
			for key, value := range entry.Fields {
				haystack += " " + key + "=" + value
			}
			if !strings.Contains(strings.ToLower(haystack), strings.ToLower(grep)) {
				continue
			}
		}
		items = append(items, matched{entry.At, entry.Service, entry.Level, entry.Message, entry.Fields})
	}
	if len(items) == 0 {
		return "没有匹配的日志。可以放宽 --level / --grep，或换一个服务；也可以先跑 services 看哪个依赖最先异常。"
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].at != items[j].at {
			return items[i].at > items[j].at
		}
		return items[i].service < items[j].service
	})
	if len(items) > limit {
		items = items[:limit]
	}
	sort.Slice(items, func(i, j int) bool { return items[i].at > items[j].at })
	var b strings.Builder
	scope := service
	if scope == "" || scope == "all" {
		scope = "全部服务"
	}
	fmt.Fprintf(&b, "# %s · 最近 %d 分钟 · %s%s · %d 条\n",
		scope, since, levelOrAll(level), grepPart(grep), len(items))
	// 时间从早到晚，读起来像一条真实的事件流。
	for i := len(items) - 1; i >= 0; i-- {
		item := items[i]
		when := c.End.Add(-time.Duration(item.at) * time.Minute)
		fmt.Fprintf(&b, "%s %-5s [%s] %s", when.Format("2006-01-02 15:04:05"), item.level, item.service, c.render(item.message))
		if len(item.fields) > 0 {
			keys := make([]string, 0, len(item.fields))
			for key := range item.fields {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(&b, " %s=%s", key, c.render(item.fields[key]))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func levelOrAll(level string) string {
	if level == "" {
		return "全部级别"
	}
	return level
}

func grepPart(grep string) string {
	if grep == "" {
		return ""
	}
	return " · grep=" + grep
}

func (c *labCase) metricsCommand(args []string) (string, error) {
	positional, _ := c.parseFlags(args)
	service := ""
	metric := ""
	if len(positional) > 0 {
		service = positional[0]
	}
	if len(positional) > 1 {
		metric = strings.Join(positional[1:], " ")
	}
	return c.metricsOutput(service, metric), nil
}

func (c *labCase) metricsOutput(service, metricName string) string {
	items := []labMetric{}
	for _, metric := range c.Scenario.Metrics {
		if service != "" && service != "all" && metric.Service != service {
			continue
		}
		if metricName != "" && !strings.Contains(strings.ToLower(metric.Name), strings.ToLower(metricName)) {
			continue
		}
		items = append(items, metric)
	}
	if len(items) == 0 {
		names := make([]string, 0, len(c.Scenario.Metrics))
		for _, metric := range c.Scenario.Metrics {
			names = append(names, metric.Service+" "+metric.Name)
		}
		return "没有匹配的指标。可用指标：\n  " + strings.Join(names, "\n  ")
	}
	var b strings.Builder
	for _, metric := range items {
		minimum, maximum, sum := metric.Points[0], metric.Points[0], 0.0
		for _, point := range metric.Points {
			minimum = math.Min(minimum, point)
			maximum = math.Max(maximum, point)
			sum += point
		}
		last := metric.Points[len(metric.Points)-1]
		average := sum / float64(len(metric.Points))
		fmt.Fprintf(&b, "%s · %s%s   最近 %d 分钟\n", metric.Service, metric.Name, unitSuffix(metric.Unit), len(metric.Points)*5)
		fmt.Fprintf(&b, "  min %s   avg %s   max %s   last %s\n",
			formatMetric(minimum), formatMetric(average), formatMetric(maximum), formatMetric(last))
		if metric.Alert > 0 {
			fmt.Fprintf(&b, "  告警阈值 > %s%s\n", formatMetric(metric.Alert), unitSuffix(metric.Unit))
		}
		fmt.Fprintf(&b, "  %s\n\n", sparkline(metric.Points))
	}
	return b.String()
}

func unitSuffix(unit string) string {
	if unit == "" {
		return ""
	}
	return " (" + unit + ")"
}

func formatMetric(value float64) string {
	if math.Abs(value) >= 100 {
		return fmt.Sprintf("%.0f", value)
	}
	if math.Abs(value) >= 10 {
		return fmt.Sprintf("%.1f", value)
	}
	return fmt.Sprintf("%.2f", value)
}

func sparkline(points []float64) string {
	runes := []rune("▁▂▃▄▅▆▇█")
	minimum, maximum := points[0], points[0]
	for _, point := range points {
		minimum = math.Min(minimum, point)
		maximum = math.Max(maximum, point)
	}
	var b strings.Builder
	for _, point := range points {
		index := 0
		if maximum > minimum {
			index = int(math.Round((point - minimum) / (maximum - minimum) * float64(len(runes)-1)))
		}
		b.WriteRune(runes[index])
	}
	return b.String()
}

func (c *labCase) traceCommand(arg string) string {
	arg = strings.TrimSpace(arg)
	if arg == "" || arg == "--slow" || arg == "-slow" || arg == "slow" {
		traces := append([]labTrace{}, c.Scenario.Traces...)
		sort.Slice(traces, func(i, j int) bool { return traces[i].TotalMS > traces[j].TotalMS })
		var b strings.Builder
		b.WriteString("最慢的几条链路\n")
		for index, trace := range traces {
			if index >= 5 {
				break
			}
			fmt.Fprintf(&b, "  %s  %d ms\n", c.render(trace.OrderID), trace.TotalMS)
		}
		b.WriteString("\n用 trace <订单号> 看具体 span。")
		return b.String()
	}
	for _, trace := range c.Scenario.Traces {
		orderID := c.render(trace.OrderID)
		if orderID == arg || trace.OrderID == arg {
			var b strings.Builder
			fmt.Fprintf(&b, "Trace %s  总耗时 %d ms\n", orderID, trace.TotalMS)
			for index, span := range trace.Spans {
				prefix := "  ├─"
				if index == len(trace.Spans)-1 {
					prefix = "  └─"
				}
				fmt.Fprintf(&b, "%s %-16s %-28s %6d ms", prefix, c.render(span.Service), c.render(span.Op), span.MS)
				if span.Note != "" {
					fmt.Fprintf(&b, "  %s", c.render(span.Note))
				}
				b.WriteString("\n")
			}
			return b.String()
		}
	}
	return "轨迹库里没有这个订单。用 trace --slow 看看最慢的几笔，或从日志里找 orderId 再查。"
}

func (c *labCase) kubectlCommand(args []string) (string, error) {
	if len(args) == 0 {
		return "用法：kubectl get pods | kubectl describe pod <名字> | kubectl logs <名字> | kubectl top pods", nil
	}
	switch args[0] {
	case "get":
		if len(args) >= 2 && (args[1] == "pods" || args[1] == "pod" || args[1] == "po") {
			return c.podsOutput(), nil
		}
	case "describe":
		if len(args) >= 2 {
			name := args[len(args)-1]
			if name == "pod" || name == "pods" {
				return "用法：kubectl describe pod <名字>", nil
			}
			return c.describePod(name), nil
		}
	case "logs":
		if len(args) >= 2 {
			name := args[len(args)-1]
			service := c.serviceForPod(name)
			if service == "" {
				return "", fmt.Errorf("找不到 Pod：%s", name)
			}
			return c.logsOutput(service, "", "", 24*60, 50), nil
		}
	case "top":
		if len(args) >= 2 && (args[1] == "pods" || args[1] == "pod") {
			return c.topOutput(), nil
		}
	}
	return "支持：kubectl get pods、kubectl describe pod <名字>、kubectl logs <名字>、kubectl top pods", nil
}

func (c *labCase) podsOutput() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE\tNODE")
	for _, svc := range c.Scenario.Services {
		total := 1
		if parts := strings.Split(svc.Replicas, "/"); len(parts) == 2 {
			if value, err := strconv.Atoi(parts[1]); err == nil && value > 0 {
				total = value
			}
		}
		for index := 0; index < total && index < 3; index++ {
			status := "Running"
			restarts := 0
			ready := "1/1"
			if svc.Status == "critical" && index == 0 {
				status = "CrashLoopBackOff"
				restarts = 17
				ready = "0/1"
			} else if svc.Status == "degraded" && index == 0 {
				restarts = 3
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t2d4h\t%s\n",
				podNameFor(c, svc.Name, index), ready, status, restarts, c.Values["node"])
		}
	}
	_ = w.Flush()
	return b.String()
}

func (c *labCase) serviceForPod(name string) string {
	for _, svc := range c.Scenario.Services {
		if strings.HasPrefix(name, svc.Name+"-") {
			return svc.Name
		}
	}
	return ""
}

func (c *labCase) describePod(name string) string {
	service := c.serviceForPod(name)
	if service == "" {
		return "找不到 Pod：" + name
	}
	var target labService
	for _, svc := range c.Scenario.Services {
		if svc.Name == service {
			target = svc
			break
		}
	}
	status := "Running"
	if target.Status == "critical" {
		status = "CrashLoopBackOff"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Name:         %s\nNamespace:    prod\nNode:         %s\nStatus:       %s\n",
		name, c.Values["node"], status)
	fmt.Fprintf(&b, "Containers:\n  %s:\n    Image:  registry.orange-mall.local/%s:2026.10.%s\n",
		target.Name, target.Name, c.Values["dateCompact"][6:])
	fmt.Fprintf(&b, "    CPU:    %s\n    Memory: %s\n", target.CPU, target.Memory)
	if target.Note != "" {
		fmt.Fprintf(&b, "    Note:   %s\n", c.render(target.Note))
	}
	b.WriteString("Events:\n")
	if target.Status == "healthy" {
		b.WriteString("  Normal  Started  kubelet  Started container\n")
	} else {
		fmt.Fprintf(&b, "  Warning Unhealthy %s  Readiness probe failed\n", c.End.Add(-17*time.Minute).Format("15:04:05"))
		fmt.Fprintf(&b, "  Warning BackOff   %s  Back-off restarting failed container\n", c.End.Add(-16*time.Minute).Format("15:04:05"))
	}
	return b.String()
}

func (c *labCase) topOutput() string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tCPU(cores)\tMEMORY(bytes)")
	index := 0
	for _, svc := range c.Scenario.Services {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", podNameFor(c, svc.Name, 0), valueOr(svc.CPU, "0m"), valueOr(svc.Memory, "0Mi"))
		index++
		if index >= 8 {
			break
		}
	}
	_ = w.Flush()
	return b.String()
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
