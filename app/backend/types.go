package backend

import "runtime"

type Course struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Level       string   `json:"level"`
	Duration    string   `json:"duration"`
	LessonIDs   []string `json:"lessonIds"`
}
type Quiz struct {
	Question    string   `json:"question"`
	Options     []string `json:"options"`
	Answer      int      `json:"answer"`
	Explanation string   `json:"explanation"`
}
type Lesson struct {
	ID             string   `json:"id"`
	CourseID       string   `json:"courseId"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Duration       int      `json:"duration"`
	Concept        string   `json:"concept"`
	Example        string   `json:"example"`
	Starter        string   `json:"starter"`
	Task           string   `json:"task"`
	ExpectedOutput string   `json:"expectedOutput"`
	Hints          []string `json:"hints"`
	Quiz           Quiz     `json:"quiz"`
	Takeaways      []string `json:"takeaways"`
	Source         string   `json:"source"`
}
type Progress struct {
	Completed  []string          `json:"completed"`
	Drafts     map[string]string `json:"drafts"`
	Notes      map[string]string `json:"notes"`
	LastLesson string            `json:"lastLesson"`
	// LastUpdateCheck 是上一次检查更新的时间戳（秒）；自动检查按它做 24 小时节流。
	LastUpdateCheck int64 `json:"lastUpdateCheck,omitempty"`
}
type Settings struct {
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
	HasKey  bool   `json:"hasKey"`
}
type Runtime struct {
	Platform string `json:"platform"`
	Engine   string `json:"engine"`
	// Version 是当前构建的版本号，Content 是已应用的课程内容包版本（没有则为空）。
	Version string `json:"version"`
	Content string `json:"content,omitempty"`
}
type Bootstrap struct {
	Courses  []Course `json:"courses"`
	Lessons  []Lesson `json:"lessons"`
	Progress Progress `json:"progress"`
	Settings Settings `json:"settings"`
	Runtime  Runtime  `json:"runtime"`
	// 只带对话摘要；正文在打开某一段时单独取，首页不必背全部历史。
	Conversations ConversationList `json:"conversations"`
}

func runtimeInfo() Runtime    { return Runtime{Platform: runtime.GOOS, Engine: "yaegi"} }
func runtimePlatform() string { return runtime.GOOS }

func emptyProgress() Progress {
	return Progress{Completed: []string{}, Drafts: map[string]string{}, Notes: map[string]string{}}
}
