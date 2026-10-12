package backend

import "runtime"

type Course struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Level       string   `json:"level"`
	Duration    string   `json:"duration"`
	LessonIDs   []string `json:"lessonIds"`
	Accent      string   `json:"accent"`
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
}
type Settings struct {
	BaseURL string `json:"baseUrl"`
	Model   string `json:"model"`
	HasKey  bool   `json:"hasKey"`
}
type Runtime struct {
	Platform string `json:"platform"`
	Engine   string `json:"engine"`
}
type Bootstrap struct {
	Courses  []Course `json:"courses"`
	Lessons  []Lesson `json:"lessons"`
	Progress Progress `json:"progress"`
	Settings Settings `json:"settings"`
	Runtime  Runtime  `json:"runtime"`
}

func runtimeInfo() Runtime { return Runtime{Platform: runtime.GOOS, Engine: "yaegi"} }
func emptyProgress() Progress {
	return Progress{Completed: []string{}, Drafts: map[string]string{}, Notes: map[string]string{}}
}
