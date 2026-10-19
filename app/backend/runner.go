package backend

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
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

// 单次 make 的容量上限。它不是语言限制，而是给本地学习应用兜底：
// 没有它，一句 make([]byte, 1<<30) 就能把整个进程拖进 OOM，而 Go 的 OOM 无法 recover。
const (
	maxMakeElements = 1 << 20
	maxChanBuffer   = 1 << 16
)

type CheckResult struct {
	Kind   string `json:"kind"`
	Name   string `json:"name,omitempty"`
	Label  string `json:"label"`
	Hint   string `json:"hint,omitempty"`
	Passed bool   `json:"passed"`
}

type RunResult struct {
	Output     string        `json:"output"`
	Error      string        `json:"error,omitempty"`
	Passed     bool          `json:"passed"`
	DurationMS int64         `json:"durationMs"`
	Checks     []CheckResult `json:"checks,omitempty"`
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

// parseLessonCode 做语法与安全边界检查，并把 AST 交给后续的结构检查与改写。
func parseLessonCode(code string) (*ast.File, error) {
	if len(code) > MaxCodeBytes {
		return nil, errors.New("代码不能超过 32 KiB")
	}
	f, err := parser.ParseFile(token.NewFileSet(), "lesson.go", code, 0)
	if err != nil {
		return nil, err
	}
	if f.Name.Name != "main" {
		return nil, errors.New("请使用 package main 和 func main")
	}
	for _, imp := range f.Imports {
		name, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			return nil, err
		}
		if name != "fmt" && name != "strings" && name != "strconv" && name != "math" {
			return nil, fmt.Errorf("练习环境不支持导入 %q，仅支持 fmt、strings、strconv、math", name)
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
	if validation != nil {
		return nil, validation
	}
	if err := guardAllocations(f); err != nil {
		return nil, err
	}
	return f, nil
}

func validateCode(code string) error {
	_, err := parseLessonCode(code)
	return err
}

// checkKinds 是课程文件里允许的结构检查类型。
var checkKinds = map[string]bool{
	"call":            true,
	"defer":           true,
	"twoValueReceive": true,
	"twoValueCall":    true,
	"pointerReceiver": true,
	"errorWrap":       true,
	"appendAssign":    true,
	"mapCommaOk":      true,
	"typeSwitch":      true,
	"addressOf":       true,
	"sliceExpr":       true,
	"conversion":      true,
	"stringLiteral":   true,
	"binaryOp":        true,
	"rangeOver":       true,
	"funcDecl":        true,
	"ident":           true,
}

func checkNeedsName(kind string) bool {
	switch kind {
	case "call", "defer", "pointerReceiver", "appendAssign", "addressOf", "sliceExpr",
		"conversion", "stringLiteral", "binaryOp", "rangeOver", "funcDecl", "ident":
		return true
	}
	return false
}

func checkLabel(check LessonCheck) string {
	name := check.Name
	switch check.Kind {
	case "call":
		if check.Arg != "" {
			return "调用 " + name + "(" + check.Arg + ")"
		}
		return "调用 " + name
	case "defer":
		return "用 defer 调用 " + name
	case "twoValueReceive":
		if name != "" {
			return "用 v, ok := <-" + name + " 接收"
		}
		return "用两值接收检查通道是否关闭"
	case "twoValueCall":
		return "处理 " + name + " 的第二个返回值"
	case "pointerReceiver":
		return "把 " + name + " 的方法接收者改成指针"
	case "errorWrap":
		return "用 %w 包装下层错误"
	case "appendAssign":
		return "接住 append 的返回值"
	case "mapCommaOk":
		return "用逗号 ok 判断键是否存在"
	case "typeSwitch":
		return "用 type switch 按动态类型分支"
	case "addressOf":
		return "传入 &" + name
	case "sliceExpr":
		return "用 " + name + " 的切片表达式共享底层数组"
	case "conversion":
		return "显式转换为 " + name
	case "stringLiteral":
		return "代码里包含格式 " + name
	case "binaryOp":
		return "使用运算符 " + name
	case "rangeOver":
		return "用 range 遍历 " + name
	case "funcDecl":
		return "定义函数 " + name
	case "ident":
		return "使用 " + name
	}
	return check.Kind
}

func checkHint(check LessonCheck) string {
	if strings.TrimSpace(check.Hint) != "" {
		return check.Hint
	}
	return "这一课要求" + checkLabel(check) + "；只把输出打印出来不算过关。"
}

// sourceFacts 是一次遍历收集到的代码事实；结构检查都在这上面做匹配。
type sourceFacts struct {
	calls              map[string]bool
	callCount          map[string]int
	callArgs           map[string]map[string]bool
	callArgCount       map[string]map[string]int
	defers             map[string]bool
	deferCount         map[string]int
	deferArgs          map[string]map[string]bool
	deferArgCount      map[string]map[string]int
	twoValueReceive    map[string]bool
	twoValueReceiveAny bool
	twoValueCall       map[string]bool
	twoValueCallAny    bool
	pointerReceiver    map[string]bool
	errorWrap          bool
	appendAssign       map[string]bool
	mapCommaOk         map[string]bool
	mapCommaOkAny      bool
	typeSwitch         bool
	addressOf          map[string]bool
	sliceExpr          map[string]bool
	literals           []string
	binaryOps          map[string]bool
	rangeOver          map[string]bool
	funcs              map[string]bool
	idents             map[string]bool
}

// exprName 把表达式压成一个用于匹配的名字：Ident、pkg.Func、xs[i] 里的 xs 等。
func exprName(expr ast.Expr) string {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if base := exprName(x.X); base != "" {
			return base + "." + x.Sel.Name
		}
		return x.Sel.Name
	case *ast.IndexExpr:
		return exprName(x.X)
	case *ast.IndexListExpr:
		return exprName(x.X)
	case *ast.ParenExpr:
		return exprName(x.X)
	case *ast.StarExpr:
		return exprName(x.X)
	}
	return ""
}

// shortName 取 selector 的最后一段，让 a.processed[key] 与 processed[key] 都能匹配。
func shortName(expr ast.Expr) string {
	name := exprName(expr)
	if index := strings.LastIndex(name, "."); index >= 0 {
		return name[index+1:]
	}
	return name
}

func addCall(facts *sourceFacts, name string, args *ast.CallExpr) {
	if name == "" {
		return
	}
	names := []string{name}
	if index := strings.LastIndex(name, "."); index >= 0 {
		names = append(names, name[index+1:])
	}
	for _, item := range names {
		facts.calls[item] = true
		facts.callCount[item]++
		if facts.callArgs[item] == nil {
			facts.callArgs[item] = map[string]bool{}
		}
		if facts.callArgCount[item] == nil {
			facts.callArgCount[item] = map[string]int{}
		}
		for _, arg := range args.Args {
			if id, ok := arg.(*ast.Ident); ok {
				facts.callArgs[item][id.Name] = true
				facts.callArgCount[item][id.Name]++
			}
		}
	}
}

func collectFacts(file *ast.File) *sourceFacts {
	facts := &sourceFacts{
		calls:           map[string]bool{},
		callCount:       map[string]int{},
		callArgs:        map[string]map[string]bool{},
		callArgCount:    map[string]map[string]int{},
		defers:          map[string]bool{},
		deferCount:      map[string]int{},
		deferArgs:       map[string]map[string]bool{},
		deferArgCount:   map[string]map[string]int{},
		twoValueReceive: map[string]bool{},
		twoValueCall:    map[string]bool{},
		pointerReceiver: map[string]bool{},
		appendAssign:    map[string]bool{},
		mapCommaOk:      map[string]bool{},
		addressOf:       map[string]bool{},
		sliceExpr:       map[string]bool{},
		binaryOps:       map[string]bool{},
		rangeOver:       map[string]bool{},
		funcs:           map[string]bool{},
		idents:          map[string]bool{},
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			addCall(facts, exprName(x.Fun), x)
			if exprName(x.Fun) == "fmt.Errorf" {
				for _, arg := range x.Args {
					if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						if text, err := strconv.Unquote(lit.Value); err == nil && strings.Contains(text, "%w") {
							facts.errorWrap = true
						}
					}
				}
			}
		case *ast.DeferStmt:
			if x.Call != nil {
				name := exprName(x.Call.Fun)
				if name != "" {
					names := []string{name}
					if index := strings.LastIndex(name, "."); index >= 0 {
						names = append(names, name[index+1:])
					}
					for _, item := range names {
						facts.defers[item] = true
						facts.deferCount[item]++
						if facts.deferArgs[item] == nil {
							facts.deferArgs[item] = map[string]bool{}
						}
						if facts.deferArgCount[item] == nil {
							facts.deferArgCount[item] = map[string]int{}
						}
						for _, arg := range x.Call.Args {
							if id, ok := arg.(*ast.Ident); ok {
								facts.deferArgs[item][id.Name] = true
								facts.deferArgCount[item][id.Name]++
							}
						}
					}
				}
			}
		case *ast.AssignStmt:
			if len(x.Lhs) == 1 && len(x.Rhs) == 1 {
				if call, ok := x.Rhs[0].(*ast.CallExpr); ok && exprName(call.Fun) == "append" {
					facts.appendAssign[shortName(x.Lhs[0])] = true
				}
			}
			if len(x.Lhs) == 2 && len(x.Rhs) == 1 {
				second, isIdent := x.Lhs[1].(*ast.Ident)
				if isIdent && second.Name == "_" {
					break
				}
				switch rhs := x.Rhs[0].(type) {
				case *ast.UnaryExpr:
					if rhs.Op == token.ARROW {
						facts.twoValueReceive[shortName(rhs.X)] = true
						facts.twoValueReceiveAny = true
					}
				case *ast.IndexExpr:
					facts.mapCommaOk[shortName(rhs.X)] = true
					facts.mapCommaOkAny = true
				case *ast.CallExpr:
					name := exprName(rhs.Fun)
					facts.twoValueCall[name] = true
					if index := strings.LastIndex(name, "."); index >= 0 {
						facts.twoValueCall[name[index+1:]] = true
					}
					facts.twoValueCallAny = true
				}
			}
		case *ast.FuncDecl:
			if x.Name != nil {
				facts.funcs[x.Name.Name] = true
			}
			if x.Recv != nil && len(x.Recv.List) == 1 {
				if star, ok := x.Recv.List[0].Type.(*ast.StarExpr); ok {
					facts.pointerReceiver[exprName(star.X)] = true
				}
			}
		case *ast.TypeSwitchStmt:
			facts.typeSwitch = true
		case *ast.UnaryExpr:
			if x.Op == token.AND {
				facts.addressOf[exprName(x.X)] = true
			}
		case *ast.SliceExpr:
			facts.sliceExpr[shortName(x.X)] = true
		case *ast.BasicLit:
			if x.Kind == token.STRING {
				if text, err := strconv.Unquote(x.Value); err == nil {
					facts.literals = append(facts.literals, text)
				}
			}
		case *ast.BinaryExpr:
			facts.binaryOps[x.Op.String()] = true
		case *ast.RangeStmt:
			if x.X != nil {
				facts.rangeOver[shortName(x.X)] = true
			}
		case *ast.Ident:
			facts.idents[x.Name] = true
		}
		return true
	})
	return facts
}

func matchCheck(check LessonCheck, facts *sourceFacts) bool {
	minimum := check.Min
	if minimum < 1 {
		minimum = 1
	}
	switch check.Kind {
	case "call":
		if check.Arg != "" {
			return facts.callArgCount[check.Name][check.Arg] >= minimum
		}
		return facts.callCount[check.Name] >= minimum
	case "defer":
		if check.Arg != "" {
			return facts.deferArgCount[check.Name][check.Arg] >= minimum
		}
		return facts.deferCount[check.Name] >= minimum
	case "twoValueReceive":
		if check.Name == "" {
			return facts.twoValueReceiveAny
		}
		return facts.twoValueReceive[check.Name]
	case "twoValueCall":
		if check.Name == "" {
			return facts.twoValueCallAny
		}
		return facts.twoValueCall[check.Name]
	case "pointerReceiver":
		return facts.pointerReceiver[check.Name]
	case "errorWrap":
		return facts.errorWrap
	case "appendAssign":
		if check.Name == "" {
			return len(facts.appendAssign) > 0
		}
		return facts.appendAssign[check.Name]
	case "mapCommaOk":
		if check.Name == "" {
			return facts.mapCommaOkAny
		}
		return facts.mapCommaOk[check.Name]
	case "typeSwitch":
		return facts.typeSwitch
	case "addressOf":
		return facts.addressOf[check.Name]
	case "sliceExpr":
		return facts.sliceExpr[check.Name]
	case "conversion":
		return facts.calls[check.Name]
	case "stringLiteral":
		for _, text := range facts.literals {
			if strings.Contains(text, check.Name) {
				return true
			}
		}
		return false
	case "binaryOp":
		return facts.binaryOps[check.Name]
	case "rangeOver":
		return facts.rangeOver[check.Name]
	case "funcDecl":
		return facts.funcs[check.Name]
	case "ident":
		return facts.idents[check.Name]
	}
	return false
}

// checkLessonCode 在输出匹配之外检查学习者是否真的用上了这一课的语言构造。
func checkLessonCode(file *ast.File, checks []LessonCheck) []CheckResult {
	if len(checks) == 0 {
		return nil
	}
	facts := collectFacts(file)
	results := make([]CheckResult, 0, len(checks))
	for _, check := range checks {
		results = append(results, CheckResult{
			Kind:   check.Kind,
			Name:   check.Name,
			Label:  checkLabel(check),
			Hint:   checkHint(check),
			Passed: matchCheck(check, facts),
		})
	}
	return results
}

// allocGuard 把 make 的容量参数包一层，运行时先检查再分配。
func allocGuard(expr ast.Expr, limit int) ast.Expr {
	return &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: ast.NewIdent("__lessonEngine"), Sel: ast.NewIdent("CheckAlloc")},
		Args: []ast.Expr{expr, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(limit)}},
	}
}

func constValue(expr ast.Expr) (constant.Value, bool) {
	switch x := expr.(type) {
	case *ast.BasicLit:
		if x.Kind != token.INT {
			return nil, false
		}
		value := constant.MakeFromLiteral(x.Value, token.INT, 0)
		if value.Kind() != constant.Int {
			return nil, false
		}
		return value, true
	case *ast.ParenExpr:
		return constValue(x.X)
	case *ast.UnaryExpr:
		value, ok := constValue(x.X)
		if !ok {
			return nil, false
		}
		switch x.Op {
		case token.ADD:
			return value, true
		case token.SUB, token.XOR:
			return constant.UnaryOp(x.Op, value, 0), true
		}
	case *ast.BinaryExpr:
		left, ok := constValue(x.X)
		if !ok {
			return nil, false
		}
		right, ok := constValue(x.Y)
		if !ok {
			return nil, false
		}
		switch x.Op {
		case token.SHL, token.SHR:
			shift, exact := constant.Uint64Val(right)
			if !exact {
				return nil, false
			}
			// 只为判断“是否超过上限”，没必要真的算 1<<1000000000 这种巨型常量。
			if shift > 62 {
				return constant.MakeInt64(maxMakeElements + 1), true
			}
			return constant.Shift(left, x.Op, uint(shift)), true
		case token.ADD, token.SUB, token.MUL, token.QUO, token.REM, token.AND, token.OR, token.XOR, token.AND_NOT:
			if (x.Op == token.QUO || x.Op == token.REM) && constant.Sign(right) == 0 {
				return nil, false
			}
			return constant.BinaryOp(left, x.Op, right), true
		}
	}
	return nil, false
}

func intLiteral(expr ast.Expr) (int64, bool) {
	value, ok := constValue(expr)
	if !ok || value.Kind() != constant.Int {
		return 0, false
	}
	number, exact := constant.Int64Val(value)
	if !exact {
		// 超出 int64 的常量一定大于任何合理的内存上限，按“超限”处理。
		return maxMakeElements + 1, true
	}
	return number, exact
}

// guardAllocations 在解释执行前加上内存兜底：
// make 的 len/cap 在运行时检查，超大数组类型在解析阶段直接拒绝。
func guardAllocations(file *ast.File) error {
	var failure error
	ast.Inspect(file, func(n ast.Node) bool {
		if failure != nil {
			return false
		}
		if array, ok := n.(*ast.ArrayType); ok && array.Len != nil {
			if size, exact := intLiteral(array.Len); exact && size > maxMakeElements {
				failure = fmt.Errorf("数组长度不能超过 %d", maxMakeElements)
				return false
			}
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		id, ok := call.Fun.(*ast.Ident)
		if !ok || id.Name != "make" || len(call.Args) < 2 {
			return true
		}
		limit := 0
		isSlice := false
		switch call.Args[0].(type) {
		case *ast.ArrayType:
			limit = maxMakeElements
			isSlice = true
		case *ast.MapType:
			limit = maxMakeElements
		case *ast.ChanType:
			limit = maxChanBuffer
		default:
			return true
		}
		call.Args[1] = allocGuard(call.Args[1], limit)
		if isSlice && len(call.Args) >= 3 {
			call.Args[2] = allocGuard(call.Args[2], limit)
		}
		return true
	})
	return failure
}
func (r *Runner) Run(ctx context.Context, code, expected string, lessonChecks ...LessonCheck) (result RunResult) {
	start := time.Now()
	var memExceeded atomic.Bool
	defer func() {
		result.DurationMS = time.Since(start).Milliseconds()
		if recovered := recover(); recovered != nil {
			if memExceeded.Load() {
				result.Error = "单次分配过大，请缩小 make 的容量"
			} else {
				result.Error = "代码解释失败，请检查语法与类型"
			}
			result.Passed = false
		}
	}()
	file, err := parseLessonCode(code)
	if err != nil {
		result.Error = err.Error()
		return
	}
	result.Checks = checkLessonCode(file, lessonChecks)
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
		"CheckAlloc": reflect.ValueOf(func(size, limit int) int {
			if size < 0 || size > limit {
				memExceeded.Store(true)
				cancel()
				panic("练习分配过大")
			}
			return size
		}),
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
	_, err = i.EvalWithContext(evalCtx, source.String())
	if finished.Load() && errors.Is(err, context.Canceled) {
		err = nil
	}
	result.Output, _ = out.finish()
	if memExceeded.Load() {
		result.Error = "单次分配过大，请缩小 make 的容量"
	} else if err != nil {
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
	checksPassed := true
	for _, check := range result.Checks {
		if !check.Passed {
			checksPassed = false
		}
	}
	result.Passed = result.Error == "" && checksPassed && strings.TrimSpace(result.Output) == strings.TrimSpace(expected)
	return
}
