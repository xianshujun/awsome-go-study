package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/traefik/yaegi/interp"
)

const MaxCodeBytes = 32 << 10
const MaxOutputBytes = 16 << 10

type RunResult struct {
	Output     string `json:"output"`
	Error      string `json:"error,omitempty"`
	Passed     bool   `json:"passed"`
	DurationMS int64  `json:"durationMs"`
}

type Runner struct {
	gate    chan struct{}
	timeout time.Duration
}

func NewRunner(timeout time.Duration) *Runner {
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	return &Runner{gate: make(chan struct{}, 1), timeout: timeout}
}

type boundedOutput struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	overflow bool
	closed   bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return len(p), nil
	}
	n := len(p)
	if remaining := MaxOutputBytes - b.buf.Len(); len(p) > remaining {
		p = p[:remaining]
		b.overflow = true
	}
	_, _ = b.buf.Write(p)
	return n, nil
}
func (b *boundedOutput) finish() (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.closed = true
	return b.buf.String(), b.overflow
}

// 只注册纯函数与内存输出，不引入 stdlib.Symbols，也不允许读取源码包。
func learnerSymbols(out *boundedOutput) interp.Exports {
	wrap := func(v any) reflect.Value { return reflect.ValueOf(v) }
	return interp.Exports{
		"fmt/fmt": {
			"Print":   wrap(func(a ...any) (int, error) { return fmt.Fprint(out, a...) }),
			"Println": wrap(func(a ...any) (int, error) { return fmt.Fprintln(out, a...) }),
			"Printf":  wrap(func(f string, a ...any) (int, error) { return fmt.Fprintf(out, f, a...) }),
			"Sprint":  wrap(fmt.Sprint), "Sprintln": wrap(fmt.Sprintln), "Sprintf": wrap(fmt.Sprintf), "Errorf": wrap(fmt.Errorf),
		},
		"strings/strings": {"Contains": wrap(strings.Contains), "HasPrefix": wrap(strings.HasPrefix), "HasSuffix": wrap(strings.HasSuffix), "Split": wrap(strings.Split), "Join": wrap(strings.Join), "TrimSpace": wrap(strings.TrimSpace), "ToUpper": wrap(strings.ToUpper), "ToLower": wrap(strings.ToLower), "Count": wrap(strings.Count), "ReplaceAll": wrap(strings.ReplaceAll)},
		"strconv/strconv": {"Atoi": wrap(strconv.Atoi), "Itoa": wrap(strconv.Itoa), "ParseBool": wrap(strconv.ParseBool), "FormatBool": wrap(strconv.FormatBool)},
		"math/math":       {"Abs": wrap(math.Abs), "Ceil": wrap(math.Ceil), "Floor": wrap(math.Floor), "Sqrt": wrap(math.Sqrt), "Min": wrap(math.Min), "Max": wrap(math.Max), "Round": wrap(math.Round), "Pi": wrap(math.Pi)},
	}
}
func validateCode(code string) error {
	if len(code) > MaxCodeBytes {
		return errors.New("代码不能超过 32 KiB")
	}
	f, err := parser.ParseFile(token.NewFileSet(), "lesson.go", code, 0)
	if err != nil {
		return err
	}
	if f.Name.Name != "main" {
		return errors.New("请使用 package main 和 func main")
	}
	for _, imp := range f.Imports {
		name, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return err
		}
		if name != "fmt" && name != "strings" && name != "strconv" && name != "math" {
			return fmt.Errorf("练习环境不支持导入 %q，仅支持 fmt、strings、strconv、math", name)
		}
	}
	var validation error
	ast.Inspect(f, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && strings.HasPrefix(id.Name, "__lesson") {
			validation = errors.New("标识符 __lesson 前缀由练习环境保留")
		}
		// 禁止绕过输出上限的内建打印。
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && (id.Name == "print" || id.Name == "println") {
				validation = errors.New("请使用 fmt.Print 系列输出")
			}
		}
		return true
	})
	return validation
}
func (r *Runner) Run(ctx context.Context, code, expected string) (result RunResult) {
	start := time.Now()
	defer func() {
		result.DurationMS = time.Since(start).Milliseconds()
		if recover() != nil {
			result.Error = "代码解释失败，请检查语法与类型"
			result.Passed = false
		}
	}()
	if err := validateCode(code); err != nil {
		result.Error = err.Error()
		return
	}
	select {
	case r.gate <- struct{}{}:
		defer func() { <-r.gate }()
	case <-ctx.Done():
		result.Error = "运行已取消"
		return
	default:
		result.Error = "已有练习正在运行，请稍后重试"
		return
	}
	evalCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	out := &boundedOutput{}
	i := interp.New(interp.Options{Stdout: out, Stderr: out})
	symbols := learnerSymbols(out)
	var finished atomic.Bool
	var childFailed atomic.Bool
	var childCount atomic.Int32
	symbols["lessonengine/lessonengine"] = map[string]reflect.Value{
		"Done":   reflect.ValueOf(func() { finished.Store(true); cancel() }),
		"Active": reflect.ValueOf(func() bool { return evalCtx.Err() == nil }),
		"Go": reflect.ValueOf(func(fn func()) {
			if childCount.Add(1) > 64 {
				childFailed.Store(true)
				cancel()
				return
			}
			go func() {
				defer func() {
					if recover() != nil {
						childFailed.Store(true)
						cancel()
					}
				}()
				if evalCtx.Err() == nil {
					fn()
				}
			}()
		}),
	}
	if err := i.Use(symbols); err != nil {
		result.Error = "练习环境初始化失败"
		return
	}
	// 包装 main：只有正常返回才成功，并在同一次 EvalWithContext 内取消所有执行帧。
	file, _ := parser.ParseFile(token.NewFileSet(), "lesson.go", code, 0)
	// 在循环体中加入协作式检查，补足子 goroutine 延迟启动与解释器取消的竞态。
	check := func() ast.Stmt {
		return &ast.IfStmt{Cond: &ast.UnaryExpr{Op: token.NOT, X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("__lessonEngine"), Sel: ast.NewIdent("Active")}}}, Body: &ast.BlockStmt{List: []ast.Stmt{&ast.BranchStmt{Tok: token.BREAK}}}}
	}
	ast.Inspect(file, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			call := goStmt.Call
			// 先在父任务中求值函数与参数，再交给受限启动器，保持 Go 的参数求值时机。
			fset := token.NewFileSet()
			var fn bytes.Buffer
			_ = format.Node(&fn, fset, call.Fun)
			var generated strings.Builder
			generated.WriteString("func() { __lessonFn := " + fn.String() + "; ")
			for index, arg := range call.Args {
				var text bytes.Buffer
				_ = format.Node(&text, fset, arg)
				generated.WriteString(fmt.Sprintf("__lessonArg%d := %s; ", index, text.String()))
			}
			generated.WriteString("__lessonEngine.Go(func(){ __lessonFn(")
			for index := range call.Args {
				if index > 0 {
					generated.WriteString(",")
				}
				generated.WriteString(fmt.Sprintf("__lessonArg%d", index))
				if index == len(call.Args)-1 && call.Ellipsis.IsValid() {
					generated.WriteString("...")
				}
			}
			generated.WriteString(") }) }()")
			wrapped, e := parser.ParseExpr(generated.String())
			if e == nil {
				goStmt.Call = wrapped.(*ast.CallExpr)
			}
		}
		var body *ast.BlockStmt
		switch x := n.(type) {
		case *ast.ForStmt:
			body = x.Body
		case *ast.RangeStmt:
			body = x.Body
		}
		if body != nil {
			body.List = append([]ast.Stmt{check()}, body.List...)
		}
		return true
	})
	ast.Inspect(file, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			for index, stmt := range block.List {
				if g, ok := stmt.(*ast.GoStmt); ok {
					block.List[index] = &ast.ExprStmt{X: g.Call}
				}
			}
		}
		if label, ok := n.(*ast.LabeledStmt); ok {
			if g, ok := label.Stmt.(*ast.GoStmt); ok {
				label.Stmt = &ast.ExprStmt{X: g.Call}
			}
		}
		return true
	})
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "main" {
			fn.Name.Name = "__lessonMain"
		}
	}
	file.Decls = append([]ast.Decl{&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{&ast.ImportSpec{Name: ast.NewIdent("__lessonEngine"), Path: &ast.BasicLit{Kind: token.STRING, Value: `"lessonengine"`}}}}}, file.Decls...)
	var source bytes.Buffer
	if e := format.Node(&source, token.NewFileSet(), file); e != nil {
		result.Error = "无法解析代码"
		return
	}
	source.WriteString("\nfunc main() { __lessonMain(); __lessonEngine.Done(); <-make(chan struct{}) }\n")
	_, err := i.EvalWithContext(evalCtx, source.String())
	if finished.Load() && errors.Is(err, context.Canceled) {
		err = nil
	}
	result.Output, _ = out.finish()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			result.Error = "运行超时，请检查无限循环或阻塞的 channel"
		} else if errors.Is(err, context.Canceled) {
			result.Error = "运行已取消"
		} else {
			result.Error = err.Error()
		}
	}
	if childFailed.Load() {
		result.Error = "并发任务异常或超过每次运行 64 个任务的上限"
	}
	if err == nil && !finished.Load() {
		result.Error = "程序未正常结束"
	}
	_, overflow := out.finish()
	if overflow {
		result.Error = "输出超过 16 KiB，请减少打印"
	}
	result.Passed = result.Error == "" && strings.TrimSpace(result.Output) == strings.TrimSpace(expected)
	return
}
