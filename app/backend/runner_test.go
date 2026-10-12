package backend

import (
	"context"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCurriculumSolutions(t *testing.T) {
	courses, lessons, err := LoadCurriculum()
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) < 3 || len(lessons) < 12 {
		t.Fatalf("课程内容过少：%d 个课程、%d 个小节", len(courses), len(lessons))
	}
	runner := NewRunner(time.Second)
	seen := map[string]bool{}
	for _, l := range lessons {
		seen[l.ID] = true
		t.Run(l.ID, func(t *testing.T) {
			// 参考解答必须真实通过，否则学习者永远完不成这一课。
			result := runner.Run(context.Background(), l.Example, l.ExpectedOutput)
			if !result.Passed {
				t.Fatalf("solution failed: %+v", result)
			}
			// starter 必须不能通过，否则这一课没有可动手的内容。
			if starter := runner.Run(context.Background(), l.Starter, l.ExpectedOutput); starter.Passed {
				t.Fatal("starter passes without edits")
			}
		})
	}
	for _, c := range courses {
		if len(c.LessonIDs) < 3 {
			t.Fatalf("课程 %s 的小节过少", c.ID)
		}
		for _, id := range c.LessonIDs {
			if !seen[id] {
				t.Fatalf("课程 %s 引用了不存在的小节 %s", c.ID, id)
			}
		}
	}
}
func TestRunnerRestrictionsAndRecovery(t *testing.T) {
	r := NewRunner(100 * time.Millisecond)
	tests := []struct{ name, code, fragment string }{
		{"syntax", "package main\nfunc main( {", "expected"},
		{"os", `package main; import "os"; func main(){os.Exit(0)}`, "不支持导入"},
		{"net", `package main; import "net/http"; func main(){}`, "不支持导入"},
		{"unsafe", `package main; import "unsafe"; func main(){}`, "不支持导入"},
		{"print", `package main; func main(){println("escape")}`, "fmt.Print"},
		{"reserved", `package main; func main(){ __lessonMain() }`, "保留"},
		{"panic", `package main; func main(){panic("bad")}`, "bad"},
		{"loop", `package main; func main(){for{}}`, "超时"},
		{"blocked", `package main; func main(){<-make(chan int)}`, "超时"},
		{"child", `package main; func main(){go func(){for{}}(); <-make(chan int)}`, "超时"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := r.Run(context.Background(), tt.code, "")
			if res.Passed || !strings.Contains(res.Error, tt.fragment) {
				t.Fatalf("unexpected: %+v", res)
			}
		})
	}
	out := NewRunner(time.Second).Run(context.Background(), `package main; import "fmt"; func main(){fmt.Printf("%20000s", "x")}`, "")
	if len(out.Output) > MaxOutputBytes || !strings.Contains(out.Error, "输出超过") {
		t.Fatalf("output bounds: %+v", out)
	}
	if res := r.Run(context.Background(), `package main; import "fmt"; func main(){fmt.Println("ok")}`, "ok"); !res.Passed {
		t.Fatalf("runner did not recover: %+v", res)
	}
}
func TestExecutionCapacity(t *testing.T) {
	r := NewRunner(time.Second)
	r.gate <- struct{}{}
	result := r.Run(context.Background(), `package main;func main(){}`, "")
	<-r.gate
	if !strings.Contains(result.Error, "已有练习") {
		t.Fatal(result)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result = r.Run(ctx, `package main;func main(){for{}}`, "")
	if result.Passed || !strings.Contains(result.Error, "取消") {
		t.Fatal(result)
	}
}
func TestChildProtection(t *testing.T) {
	r := NewRunner(time.Second)
	for _, code := range []string{
		`package main;func main(){go func(){panic("child") }();<-make(chan int)}`,
		`package main;func main(){for i:=0;i<70;i++{go func(){<-make(chan int)}()};<-make(chan int)}`,
	} {
		res := r.Run(context.Background(), code, "")
		if !strings.Contains(res.Error, "并发任务") {
			t.Fatalf("child protection: %+v", res)
		}
	}
	res := r.Run(context.Background(), `package main;import "fmt";func main(){c:=make(chan int);n:=4;go func(x int){c<-x}(n);n=9;fmt.Println(<-c)}`, "4")
	if !res.Passed {
		t.Fatalf("go arguments changed semantics: %+v", res)
	}
}

func TestNoRunawayGoroutines(t *testing.T) {
	r := NewRunner(150 * time.Millisecond)
	before := runtime.NumGoroutine()
	for n := 0; n < 8; n++ {
		res := r.Run(context.Background(), `package main; import "fmt"; func main(){go func(){for{}}();fmt.Println("ok")}`, "ok")
		if !res.Passed {
			t.Fatalf("normal main with child: %+v", res)
		}
		res = r.Run(context.Background(), `package main;func main(){go func(){for{}}();for{}}`, "")
		if !strings.Contains(res.Error, "超时") {
			t.Fatal(res)
		}
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before+2 && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if after := runtime.NumGoroutine(); after > before+2 {
		t.Fatalf("runaway goroutines: before=%d after=%d", before, after)
	}
}
