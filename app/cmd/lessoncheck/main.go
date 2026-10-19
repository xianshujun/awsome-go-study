// lessoncheck 校验课程内容文件：结构、可执行性与「starter 必须不通过」。
//
//	go run ./cmd/lessoncheck -file backend/courses/grammar.json
//	go run ./cmd/lessoncheck -embedded            # 校验内置的整套课程
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/xianshujun/awsome-go-study/app/backend"
)

func main() {
	file := flag.String("file", "", "单个课程 JSON 文件路径")
	embedded := flag.Bool("embedded", false, "校验后端内置的全部课程")
	flag.Parse()
	if (*file == "") == !*embedded {
		fmt.Fprintln(os.Stderr, "用法：lessoncheck -file backend/courses/xxx.json 或 lessoncheck -embedded")
		os.Exit(2)
	}
	var groups []group
	if *embedded {
		courses, lessons, err := backend.LoadCurriculum()
		if err != nil {
			fail("%v", err)
		}
		for _, course := range courses {
			var items []backend.Lesson
			for _, lesson := range lessons {
				if lesson.CourseID == course.ID {
					items = append(items, lesson)
				}
			}
			groups = append(groups, group{course, items})
		}
	} else {
		data, err := os.ReadFile(*file)
		if err != nil {
			fail("%v", err)
		}
		course, lessons, err := backend.LoadCourseFile(data)
		if err != nil {
			fail("%s: %v", filepath.Base(*file), err)
		}
		groups = append(groups, group{course, lessons})
	}

	runner := backend.NewRunner(2 * time.Second)
	failed := 0
	for _, g := range groups {
		fmt.Printf("课程 %s · %s（%d 节）\n", g.course.ID, g.course.Title, len(g.lessons))
		for _, lesson := range g.lessons {
			solution := runner.Run(context.Background(), lesson.Example, lesson.ExpectedOutput, lesson.Checks...)
			if !solution.Passed {
				failed++
				detail, _ := json.Marshal(solution)
				fmt.Printf("  ✗ %-24s 参考解答未通过：%s\n", lesson.ID, detail)
				continue
			}
			starter := runner.Run(context.Background(), lesson.Starter, lesson.ExpectedOutput, lesson.Checks...)
			if starter.Passed {
				failed++
				fmt.Printf("  ✗ %-24s starter 与期望输出一致，任务缺少修改空间\n", lesson.ID)
				continue
			}
			if len(lesson.Checks) > 0 {
				allPassed := true
				for _, check := range starter.Checks {
					if !check.Passed {
						allPassed = false
					}
				}
				if allPassed {
					failed++
					fmt.Printf("  ✗ %-24s starter 没通过输出，但所有结构检查都通过了：检查太松\n", lesson.ID)
					continue
				}
			}
			fmt.Printf("  ✓ %-24s %s\n", lesson.ID, lesson.Title)
		}
	}
	if failed > 0 {
		fail("共 %d 个问题", failed)
	}
	fmt.Println("全部小节校验通过")
}

type group struct {
	course  backend.Course
	lessons []backend.Lesson
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
