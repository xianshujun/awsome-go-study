package backend

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// 课程内容按仓库资料分文件维护：backend/courses/<course>.json。
// 每个文件对应 awsome-go-study 中的一篇文档，便于与原文对照更新。
//
//go:embed courses/*.json
var curriculumFS embed.FS

// courseFile 是单个课程文件的格式；lessonIds 由文件内顺序推导。
type courseFile struct {
	Order   int      `json:"order"`
	Course  Course   `json:"course"`
	Lessons []Lesson `json:"lessons"`

	// source 是文件名（如 grammar.json）。磁盘上的内容包按这个名字覆盖同名课程。
	source string
}

var lessonIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,40}$`)

// LoadCurriculum 加载内置课程；启动时校验，避免把坏内容发给学习者。
func LoadCurriculum() ([]Course, []Lesson, error) {
	files, err := readEmbeddedCourses()
	if err != nil {
		return nil, nil, err
	}
	return assembleCurriculum(files)
}

// readEmbeddedCourses 读出编译进程序的那一份课程。
func readEmbeddedCourses() ([]courseFile, error) {
	names, err := fs.Glob(curriculumFS, "courses/*.json")
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("内置课程为空")
	}
	sort.Strings(names)
	files := make([]courseFile, 0, len(names))
	for _, name := range names {
		data, err := curriculumFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		file, err := parseCourseFile(data)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path.Base(name), err)
		}
		file.source = path.Base(name)
		files = append(files, file)
	}
	return files, nil
}

/*
loadCurriculum 以内置课程为底，用 overrideDir 里的同名 JSON 覆盖对应课程，
并允许新增文件带来的新课程。内容热更新走的就是这条路：把新的课程文件放进
数据目录，不必重新发布程序。

坏文件只跳过并记一条警告——学习内容可以少一课，但不能因为一个坏文件就开不了应用。
*/
func loadCurriculum(overrideDir string) ([]Course, []Lesson, []string, error) {
	files, err := readEmbeddedCourses()
	if err != nil {
		return nil, nil, nil, err
	}
	warnings := []string{}
	if overrideDir != "" {
		names, err := filepath.Glob(filepath.Join(overrideDir, "*.json"))
		if err != nil {
			return nil, nil, nil, err
		}
		sort.Strings(names)
		index := make(map[string]int, len(files))
		for i, file := range files {
			index[file.source] = i
		}
		for _, full := range names {
			base := filepath.Base(full)
			data, err := os.ReadFile(full)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s 读取失败", base))
				continue
			}
			if len(data) > maxContentFileBytes {
				warnings = append(warnings, fmt.Sprintf("%s 超过单文件上限，已忽略", base))
				continue
			}
			file, err := parseCourseFile(data)
			if err != nil {
				warnings = append(warnings, fmt.Sprintf("%s 解析失败，已忽略", base))
				continue
			}
			file.source = base
			if i, ok := index[base]; ok {
				files[i] = file
			} else {
				files = append(files, file)
			}
		}
	}
	courses, lessons, err := assembleCurriculum(files)
	if err != nil {
		return nil, nil, warnings, err
	}
	return courses, lessons, warnings, nil
}

// LoadCourseFile 校验单个课程文件，供 lessoncheck 等维护工具使用。
func LoadCourseFile(data []byte) (Course, []Lesson, error) {
	file, err := parseCourseFile(data)
	if err != nil {
		return Course{}, nil, err
	}
	if err := validateCourseFile(file); err != nil {
		return Course{}, nil, err
	}
	file.Course.LessonIDs = make([]string, 0, len(file.Lessons))
	for _, lesson := range file.Lessons {
		file.Course.LessonIDs = append(file.Course.LessonIDs, lesson.ID)
	}
	return file.Course, file.Lessons, nil
}

func parseCourseFile(data []byte) (courseFile, error) {
	var file courseFile
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&file); err != nil {
		return courseFile{}, fmt.Errorf("解析课程 JSON 失败: %w", err)
	}
	return file, nil
}

// assembleCurriculum 按 order 排序，并做跨文件校验（课程 ID、小节 ID 全局唯一）。
func assembleCurriculum(files []courseFile) ([]Course, []Lesson, error) {
	sort.SliceStable(files, func(i, j int) bool { return files[i].Order < files[j].Order })
	courses := make([]Course, 0, len(files))
	lessons := make([]Lesson, 0, len(files)*6)
	courseIDs := map[string]bool{}
	lessonIDs := map[string]string{}
	for _, file := range files {
		if err := validateCourseFile(file); err != nil {
			return nil, nil, fmt.Errorf("课程 %q: %w", file.Course.ID, err)
		}
		if courseIDs[file.Course.ID] {
			return nil, nil, fmt.Errorf("课程 ID 重复: %s", file.Course.ID)
		}
		courseIDs[file.Course.ID] = true
		file.Course.LessonIDs = make([]string, 0, len(file.Lessons))
		for _, lesson := range file.Lessons {
			if owner, ok := lessonIDs[lesson.ID]; ok {
				return nil, nil, fmt.Errorf("小节 ID 重复: %s（%s 与 %s）", lesson.ID, owner, file.Course.ID)
			}
			lessonIDs[lesson.ID] = file.Course.ID
			lesson.CourseID = file.Course.ID
			file.Course.LessonIDs = append(file.Course.LessonIDs, lesson.ID)
			lessons = append(lessons, lesson)
		}
		courses = append(courses, file.Course)
	}
	return courses, lessons, nil
}

func validateCourseFile(file courseFile) error {
	if file.Order <= 0 {
		return fmt.Errorf("order 必须为正整数")
	}
	c := file.Course
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.Title) == "" || strings.TrimSpace(c.Description) == "" {
		return fmt.Errorf("课程 id、title、description 不能为空")
	}
	if strings.TrimSpace(c.Level) == "" || strings.TrimSpace(c.Duration) == "" {
		return fmt.Errorf("课程 level、duration 不能为空")
	}
	if len(file.Lessons) < 3 {
		return fmt.Errorf("每个课程至少 3 个小节，当前 %d 个", len(file.Lessons))
	}
	seen := map[string]bool{}
	for index, lesson := range file.Lessons {
		if err := validateLesson(lesson); err != nil {
			return fmt.Errorf("第 %d 个小节 %q: %w", index+1, lesson.ID, err)
		}
		if seen[lesson.ID] {
			return fmt.Errorf("小节 ID 在同一课程内重复: %s", lesson.ID)
		}
		seen[lesson.ID] = true
	}
	return nil
}

func validateLesson(l Lesson) error {
	if !lessonIDPattern.MatchString(l.ID) {
		return fmt.Errorf("id 需为小写字母、数字或连字符")
	}
	for name, value := range map[string]string{
		"title": l.Title, "description": l.Description, "concept": l.Concept, "example": l.Example,
		"starter": l.Starter, "task": l.Task, "expectedOutput": l.ExpectedOutput, "source": l.Source,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s 不能为空", name)
		}
	}
	if l.Duration <= 0 || l.Duration > 120 {
		return fmt.Errorf("duration 需在 1 到 120 分钟之间")
	}
	if l.Example == l.Starter {
		return fmt.Errorf("starter 必须与 example 不同，学习者才有修改空间")
	}
	for name, code := range map[string]string{"example": l.Example, "starter": l.Starter} {
		if len(code) > MaxCodeBytes {
			return fmt.Errorf("%s 超过 %d 字节", name, MaxCodeBytes)
		}
	}
	if len(l.ExpectedOutput) > 2048 {
		return fmt.Errorf("expectedOutput 过长")
	}
	if len(l.Hints) < 2 {
		return fmt.Errorf("至少需要 2 条提示")
	}
	for _, hint := range l.Hints {
		if strings.TrimSpace(hint) == "" {
			return fmt.Errorf("提示不能为空")
		}
	}
	if len(l.Takeaways) < 3 {
		return fmt.Errorf("至少需要 3 条要点回顾")
	}
	q := l.Quiz
	if strings.TrimSpace(q.Question) == "" || strings.TrimSpace(q.Explanation) == "" {
		return fmt.Errorf("理解题的问题与解析不能为空")
	}
	if len(q.Options) < 3 || len(q.Options) > 5 {
		return fmt.Errorf("理解题需要 3 到 5 个选项")
	}
	if q.Answer < 0 || q.Answer >= len(q.Options) {
		return fmt.Errorf("理解题答案下标越界")
	}
	for _, option := range q.Options {
		if strings.TrimSpace(option) == "" {
			return fmt.Errorf("理解题选项不能为空")
		}
	}
	return nil
}
