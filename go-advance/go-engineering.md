# Go语言工程实践

## 前言

本文档将全面介绍Go语言的工程实践，帮助读者更好地理解和使用Go语言进行工程开发。通过学习本文档，你将掌握Go语言工程实践的核心概念和常用技巧，包括项目结构、测试、性能优化、错误处理、日志、配置管理等。

## 第一章：项目结构

### 1.1 标准项目结构

Go语言有一个推荐的目录结构，遵循这个结构可以让你的项目更易于维护和理解。

**基本项目结构**

```
myproject/
├── cmd/
│   ├── app1/
│   │   └── main.go
│   └── app2/
│       └── main.go
├── internal/
│   ├── pkg1/
│   │   └── pkg1.go
│   └── pkg2/
│       └── pkg2.go
├── pkg/
│   ├── pkg1/
│   │   └── pkg1.go
│   └── pkg2/
│       └── pkg2.go
├── api/
│   ├── proto/
│   │   └── service.proto
│   └── openapi/
│       └── api.yaml
├── web/
│   ├── static/
│   └── templates/
├── configs/
│   ├── config.yaml
│   └── config.prod.yaml
├── scripts/
│   ├── build.sh
│   └── deploy.sh
├── build/
│   └── docker/
│       └── Dockerfile
├── deployments/
│   ├── k8s/
│   │   └── deployment.yaml
│   └── docker-compose.yml
├── test/
│   ├── testdata/
│   └── integration/
├── docs/
│   ├── architecture.md
│   └── api.md
├── tools/
│   └── tools.go
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── .gitignore
```

**目录说明**

- `cmd/`: 应用程序的入口点，每个子目录包含一个主程序
- `internal/`: 内部包，不能被外部项目导入
- `pkg/`: 外部可用的库代码
- `api/`: API定义文件，如protobuf、OpenAPI等
- `web/`: Web应用相关的静态资源和模板
- `configs/`: 配置文件
- `scripts/`: 构建和部署脚本
- `build/`: 构建配置和Dockerfile
- `deployments/`: 部署配置
- `test/`: 测试数据和集成测试
- `docs/`: 文档
- `tools/`: 工具和依赖管理
- `go.mod/go.sum`: Go模块依赖
- `Makefile`: 构建脚本
- `README.md`: 项目说明
- `.gitignore`: Git忽略文件

### 1.2 模块管理

Go模块是Go语言依赖管理的官方解决方案。

**初始化模块**

```bash
# 初始化新模块
go mod init github.com/username/myproject

# 查看模块信息
go list -m all

# 查看依赖图
go mod graph
```

**添加依赖**

```bash
# 添加依赖
go get github.com/gin-gonic/gin

# 添加指定版本的依赖
go get github.com/gin-gonic/gin@v1.9.1

# 添加测试依赖
go get github.com/stretchr/testify@latest
```

**更新依赖**

```bash
# 更新所有依赖
go get -u ./...

# 更新指定依赖
go get -u github.com/gin-gonic/gin

# 整理依赖
go mod tidy

# 验证依赖
go mod verify
```

**vendor模式**

```bash
# 创建vendor目录
go mod vendor

# 使用vendor模式构建
go build -mod=vendor
```

### 1.3 代码组织

**包的命名**

```go
// 包名应该简短、小写、有意义
package user

// 避免使用下划线或驼峰命名
package user_service  // 不推荐
package userService   // 不推荐

// 避免使用复数形式
package users  // 不推荐
package user   // 推荐
```

**文件组织**

```go
// user.go - 主要类型定义
package user

type User struct {
    ID       string
    Name     string
    Email    string
    Password string
}

// user_repository.go - 数据访问层
package user

type Repository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
    Update(user *User) error
    Delete(id string) error
}

// user_service.go - 业务逻辑层
package user

type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

func (s *Service) CreateUser(user *User) error {
    return s.repo.Create(user)
}
```

### 1.4 构建配置

**Makefile示例**

```makefile
# 变量定义
APP_NAME=myapp
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date +%FT%T%z)
GO_FLAGS=-ldflags="-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
.PHONY: all
all: build

# 构建
.PHONY: build
build:
    go build $(GO_FLAGS) -o bin/$(APP_NAME) ./cmd/app

# 运行
.PHONY: run
run:
    go run ./cmd/app

# 测试
.PHONY: test
test:
    go test -v ./...

# 测试覆盖率
.PHONY: coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# 格式化代码
.PHONY: fmt
fmt:
    go fmt ./...

# 代码检查
.PHONY: lint
lint:
    golangci-lint run

# 清理
.PHONY: clean
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Docker构建
.PHONY: docker-build
docker-build:
    docker build -t $(APP_NAME):$(VERSION) .

# Docker运行
.PHONY: docker-run
docker-run:
    docker run -p 8080:8080 $(APP_NAME):$(VERSION)
```

### 1.5 实例：创建一个新项目

让我们创建一个简单的Web项目。

**项目结构**

```
mywebapp/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── handler.go
│   ├── service/
│   │   └── service.go
│   └── repository/
│       └── repository.go
├── pkg/
│   └── util/
│       └── util.go
├── configs/
│   └── config.yaml
├── go.mod
└── README.md
```

**main.go**

```go
package main

import (
    "log"
    "mywebapp/internal/handler"
    "mywebapp/internal/service"
    "mywebapp/internal/repository"
)

var (
    Version   = "dev"
    BuildTime = "unknown"
)

func main() {
    log.Printf("Starting server, version: %s, build time: %s", Version, BuildTime)
    
    repo := repository.NewRepository()
    svc := service.NewService(repo)
    h := handler.NewHandler(svc)
    
    if err := h.Run(":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### 1.6 小结

本章介绍了Go语言的项目结构，包括标准项目结构、模块管理、代码组织、构建配置等。掌握好Go语言的项目结构对于编写可维护的项目非常重要。

下一章我们将学习Go语言的测试，包括单元测试、集成测试、基准测试、测试覆盖率等，以及它们的使用方法和注意事项。

## 第二章：测试

### 2.1 单元测试

Go语言内置了测试框架，使得编写单元测试变得非常简单。

**基本测试**

```go
package math

func Add(a, b int) int {
    return a + b
}

func Subtract(a, b int) int {
    return a - b
}
```

```go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5
    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}

func TestSubtract(t *testing.T) {
    result := Subtract(5, 3)
    expected := 2
    if result != expected {
        t.Errorf("Subtract(5, 3) = %d; want %d", result, expected)
    }
}
```

**表驱动测试**

```go
package math

import "testing"

func TestAddTableDriven(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"mixed", 2, -3, -1},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

**子测试**

```go
package math

import "testing"

func TestSubtractSubtests(t *testing.T) {
    t.Run("positive", func(t *testing.T) {
        result := Subtract(5, 3)
        if result != 2 {
            t.Errorf("expected 2, got %d", result)
        }
    })

    t.Run("negative", func(t *testing.T) {
        result := Subtract(-5, -3)
        if result != -2 {
            t.Errorf("expected -2, got %d", result)
        }
    })
}
```

### 2.2 测试辅助函数

**使用testing.T辅助函数**

```go
package math

import "testing"

func TestAddHelper(t *testing.T) {
    tests := []struct {
        a, b     int
        expected int
    }{
        {2, 3, 5},
        {-2, -3, -5},
        {0, 0, 0},
    }

    for _, tt := range tests {
        result := Add(tt.a, tt.b)
        if result != tt.expected {
            t.Fatalf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
        }
    }
}
```

**使用assert库**

```go
package math

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestAddAssert(t *testing.T) {
    result := Add(2, 3)
    assert.Equal(t, 5, result, "Add(2, 3) should equal 5")
}

func TestAddTableDrivenAssert(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2.3 Mock和Stub

**使用接口进行Mock**

```go
package user

type Repository interface {
    FindByID(id string) (*User, error)
    Create(user *User) error
}

type Service struct {
    repo Repository
}

func (s *Service) GetUser(id string) (*User, error) {
    return s.repo.FindByID(id)
}
```

```go
package user

import "testing"

type MockRepository struct {
    users map[string]*User
}

func NewMockRepository() *MockRepository {
    return &MockRepository{
        users: make(map[string]*User),
    }
}

func (m *MockRepository) FindByID(id string) (*User, error) {
    return m.users[id], nil
}

func (m *MockRepository) Create(user *User) error {
    m.users[user.ID] = user
    return nil
}

func TestServiceGetUser(t *testing.T) {
    mockRepo := NewMockRepository()
    mockRepo.users["1"] = &User{ID: "1", Name: "Alice"}

    service := &Service{repo: mockRepo}
    user, err := service.GetUser("1")

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if user.Name != "Alice" {
        t.Errorf("expected name Alice, got %s", user.Name)
    }
}
```

### 2.4 集成测试

集成测试用于测试多个组件之间的交互。

**HTTP服务测试**

```go
package handler

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestHandler(t *testing.T) {
    handler := NewHandler()

    req, err := http.NewRequest("GET", "/health", nil)
    if err != nil {
        t.Fatal(err)
    }

    rr := httptest.NewRecorder()
    handler.ServeHTTP(rr, req)

    if status := rr.Code; status != http.StatusOK {
        t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
    }

    expected := `{"status":"ok"}`
    if rr.Body.String() != expected {
        t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
    }
}
```

**数据库集成测试**

```go
package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    repo := NewRepository(db)

    user := &User{
        ID:    "1",
        Name:  "Alice",
        Email: "alice@example.com",
    }

    err := repo.Create(user)
    assert.NoError(t, err)

    found, err := repo.FindByID("1")
    assert.NoError(t, err)
    assert.Equal(t, "Alice", found.Name)
}
```

### 2.5 基准测试

基准测试用于测量代码的性能。

**基本基准测试**

```go
package math

import "testing"

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}

func BenchmarkSubtract(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Subtract(5, 3)
    }
}
```

**带重置的基准测试**

```go
package math

import "testing"

func BenchmarkAddWithReset(b *testing.B) {
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
```

**并行基准测试**

```go
package math

import "testing"

func BenchmarkAddParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            Add(2, 3)
        }
    })
}
```

### 2.6 测试覆盖率

测试覆盖率用于衡量测试的完整性。

**生成覆盖率报告**

```bash
# 生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成HTML报告
go tool cover -html=coverage.out -o coverage.html
```

**设置覆盖率阈值**

```bash
# 检查覆盖率是否达到阈值
go test -coverprofile=coverage.out -covermode=atomic ./...
```

**在CI中使用覆盖率**

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      - name: Test
        run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
      - name: Coverage
        run: go tool cover -func=coverage.out
```

### 2.7 实例：完整的测试套件

让我们为一个用户服务创建完整的测试套件。

```go
package user

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(id string) (*User, error) {
    args := m.Called(id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

func (m *MockRepository) Create(user *User) error {
    args := m.Called(user)
    return args.Error(0)
}

func TestServiceCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        user    *User
        setup   func(*MockRepository)
        wantErr bool
    }{
        {
            name: "success",
            user: &User{ID: "1", Name: "Alice", Email: "alice@example.com"},
            setup: func(m *MockRepository) {
                m.On("Create", mock.AnythingOfType("*user.User")).Return(nil)
            },
            wantErr: false,
        },
        {
            name: "invalid email",
            user: &User{ID: "1", Name: "Alice", Email: "invalid"},
            setup: func(m *MockRepository) {
                m.On("Create", mock.AnythingOfType("*user.User")).Return(assert.AnError)
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockRepository)
            tt.setup(mockRepo)

            service := &Service{repo: mockRepo}
            err := service.CreateUser(tt.user)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

### 2.8 小结

本章介绍了Go语言的测试，包括单元测试、集成测试、基准测试、测试覆盖率等。我们学习了Go语言如何编写测试、使用测试辅助函数、进行Mock和Stub、测量性能等。掌握好Go语言的测试对于编写高质量的代码非常重要。

下一章我们将学习Go语言的性能优化，包括性能分析、内存优化、并发优化等，以及它们的使用方法和注意事项。

## 第三章：性能优化

### 3.1 性能分析

Go语言提供了强大的性能分析工具，帮助我们找出性能瓶颈。

**使用pprof进行CPU分析**

```go
package main

import (
    "log"
    "os"
    "runtime/pprof"
)

func fibonacci(n int) int {
    if n <= 1 {
        return n
    }
    return fibonacci(n-1) + fibonacci(n-2)
}

func main() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil {
        log.Fatal(err)
    }
    defer pprof.StopCPUProfile()

    for i := 0; i < 10; i++ {
        fibonacci(30)
    }
}
```

**分析CPU性能**

```bash
# 运行程序生成profile文件
go run main.go

# 分析profile文件
go tool pprof cpu.prof

# 在pprof交互界面中
top 10
list fibonacci
web
```

**使用pprof进行内存分析**

```go
package main

import (
    "log"
    "os"
    "runtime/pprof"
)

func main() {
    f, err := os.Create("mem.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    data := make([]byte, 1024*1024*100)
    _ = data

    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Fatal(err)
    }
}
```

**HTTP pprof**

```go
package main

import (
    "log"
    "net/http"
    _ "net/http/pprof"
)

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()

    select {}
}
```

访问 `http://localhost:6060/debug/pprof/` 查看性能分析页面。

### 3.2 内存优化

内存优化是提高程序性能的重要手段。

**减少内存分配**

```go
// 不好的做法：频繁分配内存
func badConcatenate(strings []string) string {
    var result string
    for _, s := range strings {
        result += s
    }
    return result
}

// 好的做法：预分配内存
func goodConcatenate(strings []string) string {
    var builder strings.Builder
    builder.Grow(1024)
    for _, s := range strings {
        builder.WriteString(s)
    }
    return builder.String()
}
```

**使用对象池**

```go
package main

import (
    "sync"
)

type Buffer struct {
    data []byte
}

var bufferPool = sync.Pool{
    New: func() interface{} {
        return &Buffer{
            data: make([]byte, 1024),
        }
    },
}

func processData() {
    buf := bufferPool.Get().(*Buffer)
    defer bufferPool.Put(buf)

    buf.data[0] = 1
}

func main() {
    for i := 0; i < 1000; i++ {
        processData()
    }
}
```

**避免内存泄漏**

```go
// 不好的做法：goroutine泄漏
func badLeak() {
    ch := make(chan int)
    go func() {
        for {
            select {
            case <-ch:
            }
        }
    }()
}

// 好的做法：使用context取消
func goodLeak(ctx context.Context) {
    ch := make(chan int)
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case <-ch:
            }
        }
    }()
}
```

### 3.3 并发优化

并发是Go语言的优势，正确使用并发可以大幅提高性能。

**使用Worker Pool**

```go
package main

import (
    "sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 100)
    results := make(chan int, 100)

    var wg sync.WaitGroup

    for i := 1; i <= 5; i++ {
        wg.Add(1)
        go worker(i, jobs, results, &wg)
    }

    for j := 1; j <= 100; j++ {
        jobs <- j
    }
    close(jobs)

    wg.Wait()
    close(results)

    for result := range results {
        _ = result
    }
}
```

**使用sync.Map**

```go
package main

import (
    "sync"
)

var cache sync.Map

func getFromCache(key string) (interface{}, bool) {
    return cache.Load(key)
}

func setCache(key string, value interface{}) {
    cache.Store(key, value)
}

func main() {
    setCache("key1", "value1")
    if value, ok := getFromCache("key1"); ok {
        _ = value
    }
}
```

**使用singleflight**

```go
package main

import (
    "sync"
    "time"
)

var (
    sf    singleflight.Group
    cache = make(map[string]string)
)

func getData(key string) (string, error) {
    if value, ok := cache[key]; ok {
        return value, nil
    }

    result, err, _ := sf.Do(key, func() (interface{}, error) {
        time.Sleep(time.Second)
        value := "data for " + key
        cache[key] = value
        return value, nil
    })

    if err != nil {
        return "", err
    }

    return result.(string), nil
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            data, _ := getData("key1")
            _ = data
        }()
    }
    wg.Wait()
}
```

### 3.4 I/O优化

I/O操作通常是性能瓶颈，优化I/O可以显著提高性能。

**使用缓冲**

```go
package main

import (
    "bufio"
    "os"
)

func writeBuffered(filename string, data []string) error {
    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := bufio.NewWriter(file)
    defer writer.Flush()

    for _, line := range data {
        _, err := writer.WriteString(line + "\n")
        if err != nil {
            return err
        }
    }

    return nil
}

func main() {
    data := []string{"line1", "line2", "line3"}
    writeBuffered("output.txt", data)
}
```

**批量操作**

```go
package main

import (
    "database/sql"
    "fmt"
)

func batchInsert(db *sql.DB, users []User) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, user := range users {
        _, err := stmt.Exec(user.Name, user.Email)
        if err != nil {
            return err
        }
    }

    return tx.Commit()
}

type User struct {
    Name  string
    Email string
}

func main() {
    users := []User{
        {Name: "Alice", Email: "alice@example.com"},
        {Name: "Bob", Email: "bob@example.com"},
    }
    batchInsert(nil, users)
    fmt.Println("Batch insert completed")
}
```

### 3.5 实例：性能优化案例

让我们优化一个实际的性能问题。

**优化前的代码**

```go
package main

import (
    "fmt"
    "strings"
)

func processText(text string) string {
    var result string
    words := strings.Fields(text)
    
    for _, word := range words {
        result += strings.ToUpper(word) + " "
    }
    
    return result
}

func main() {
    text := "hello world this is a test"
    result := processText(text)
    fmt.Println(result)
}
```

**优化后的代码**

```go
package main

import (
    "fmt"
    "strings"
    "sync"
)

func processTextOptimized(text string) string {
    words := strings.Fields(text)
    var builder strings.Builder
    builder.Grow(len(text) * 2)
    
    for _, word := range words {
        builder.WriteString(strings.ToUpper(word))
        builder.WriteByte(' ')
    }
    
    return builder.String()
}

func processTextConcurrent(text string) string {
    words := strings.Fields(text)
    var wg sync.WaitGroup
    results := make([]string, len(words))
    
    for i, word := range words {
        wg.Add(1)
        go func(idx int, w string) {
            defer wg.Done()
            results[idx] = strings.ToUpper(w)
        }(i, word)
    }
    
    wg.Wait()
    
    var builder strings.Builder
    builder.Grow(len(text) * 2)
    for _, result := range results {
        builder.WriteString(result)
        builder.WriteByte(' ')
    }
    
    return builder.String()
}

func main() {
    text := "hello world this is a test"
    
    result1 := processTextOptimized(text)
    fmt.Println("Optimized:", result1)
    
    result2 := processTextConcurrent(text)
    fmt.Println("Concurrent:", result2)
}
```

### 3.6 小结

本章介绍了Go语言的性能优化，包括性能分析、内存优化、并发优化、I/O优化等。我们学习了Go语言如何进行性能分析、优化内存使用、提高并发性能、优化I/O操作等。掌握好Go语言的性能优化对于编写高性能的程序非常重要。

下一章我们将学习Go语言的错误处理，包括错误定义、错误处理模式、错误包装等，以及它们的使用方法和注意事项。

## 第四章：错误处理

### 4.1 错误定义

Go语言的错误处理机制简单而强大，通过error接口实现。

**基本错误**

```go
package main

import (
    "errors"
    "fmt"
)

func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func main() {
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println("Result:", result)
}
```

**自定义错误类型**

```go
package main

import "fmt"

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on field %s: %s", e.Field, e.Message)
}

func validateUser(name, email string) error {
    if name == "" {
        return &ValidationError{Field: "name", Message: "name cannot be empty"}
    }
    if email == "" {
        return &ValidationError{Field: "email", Message: "email cannot be empty"}
    }
    return nil
}

func main() {
    err := validateUser("", "")
    if err != nil {
        fmt.Println("Validation error:", err)
        return
    }
}
```

**错误断言**

```go
package main

import (
    "errors"
    "fmt"
)

type MyError struct {
    Code    int
    Message string
}

func (e *MyError) Error() string {
    return fmt.Sprintf("error %d: %s", e.Code, e.Message)
}

func doSomething() error {
    return &MyError{Code: 404, Message: "not found"}
}

func main() {
    err := doSomething()
    if err != nil {
        var myErr *MyError
        if errors.As(err, &myErr) {
            fmt.Printf("MyError: Code=%d, Message=%s\n", myErr.Code, myErr.Message)
        } else {
            fmt.Println("Unknown error:", err)
        }
    }
}
```

### 4.2 错误处理模式

Go语言有多种错误处理模式，可以根据场景选择合适的模式。

**Sentinel错误**

```go
package main

import (
    "errors"
    "fmt"
)

var (
    ErrUserNotFound = errors.New("user not found")
    ErrInvalidInput = errors.New("invalid input")
)

func getUser(id int) (string, error) {
    if id <= 0 {
        return "", ErrInvalidInput
    }
    if id == 999 {
        return "", ErrUserNotFound
    }
    return "user", nil
}

func main() {
    _, err := getUser(999)
    if errors.Is(err, ErrUserNotFound) {
        fmt.Println("User not found")
        return
    }
    if errors.Is(err, ErrInvalidInput) {
        fmt.Println("Invalid input")
        return
    }
}
```

**错误包装**

```go
package main

import (
    "errors"
    "fmt"
)

func readConfig() error {
    return errors.New("config file not found")
}

func loadConfig() error {
    err := readConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    return nil
}

func main() {
    err := loadConfig()
    if err != nil {
        fmt.Println("Error:", err)
        
        var originalErr error
        if errors.As(err, &originalErr) {
            fmt.Println("Original error:", originalErr)
        }
    }
}
```

**错误链**

```go
package main

import (
    "errors"
    "fmt"
)

func step1() error {
    return errors.New("step1 failed")
}

func step2() error {
    err := step1()
    if err != nil {
        return fmt.Errorf("step2: %w", err)
    }
    return nil
}

func step3() error {
    err := step2()
    if err != nil {
        return fmt.Errorf("step3: %w", err)
    }
    return nil
}

func main() {
    err := step3()
    if err != nil {
        fmt.Println("Error chain:")
        fmt.Println(err)
        
        for err != nil {
            fmt.Println("  -", err)
            err = errors.Unwrap(err)
        }
    }
}
```

### 4.3 错误恢复

Go语言的panic和recover机制用于处理不可恢复的错误。

**Panic和Recover**

```go
package main

import "fmt"

func riskyOperation() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("Recovered from panic:", r)
        }
    }()
    
    fmt.Println("Starting risky operation")
    panic("something went wrong")
    fmt.Println("This won't be printed")
}

func main() {
    riskyOperation()
    fmt.Println("Program continues")
}
```

**在HTTP处理中使用Recover**

```go
package main

import (
    "fmt"
    "net/http"
)

func recoverMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                fmt.Println("Recovered from panic:", r)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

func handler(w http.ResponseWriter, r *http.Request) {
    panic("unexpected error")
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", handler)
    
    wrapped := recoverMiddleware(mux)
    http.ListenAndServe(":8080", wrapped)
}
```

### 4.4 错误日志

记录错误对于调试和监控非常重要。

**使用log包**

```go
package main

import (
    "log"
    "os"
)

func main() {
    logFile, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    defer logFile.Close()

    log.SetOutput(logFile)
    log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

    log.Println("Application started")
    log.Printf("Processing item %d", 1)
    log.Println("Application finished")
}
```

**结构化日志**

```go
package main

import (
    "encoding/json"
    "log"
    "os"
    "time"
)

type Logger struct {
    logger *log.Logger
}

func NewLogger(filename string) *Logger {
    file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }
    return &Logger{
        logger: log.New(file, "", 0),
    }
}

func (l *Logger) Log(level, message string, fields map[string]interface{}) {
    entry := map[string]interface{}{
        "time":    time.Now().Format(time.RFC3339),
        "level":   level,
        "message": message,
    }
    
    for k, v := range fields {
        entry[k] = v
    }
    
    data, _ := json.Marshal(entry)
    l.logger.Println(string(data))
}

func main() {
    logger := NewLogger("app.log")
    
    logger.Log("INFO", "Application started", map[string]interface{}{
        "version": "1.0.0",
    })
    
    logger.Log("ERROR", "Failed to process request", map[string]interface{}{
        "error": "invalid input",
        "path":  "/api/users",
    })
}
```

### 4.5 实例：完整的错误处理

让我们为一个用户服务实现完整的错误处理。

```go
package user

import (
    "errors"
    "fmt"
)

var (
    ErrUserNotFound    = errors.New("user not found")
    ErrInvalidEmail    = errors.New("invalid email")
    ErrUserExists      = errors.New("user already exists")
    ErrInvalidPassword = errors.New("invalid password")
)

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s", e.Message)
}

type User struct {
    ID       string
    Name     string
    Email    string
    Password string
}

type Service struct {
    repo Repository
}

func (s *Service) CreateUser(user *User) error {
    if err := s.validateUser(user); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    exists, err := s.repo.ExistsByEmail(user.Email)
    if err != nil {
        return fmt.Errorf("check user existence: %w", err)
    }
    
    if exists {
        return ErrUserExists
    }
    
    if err := s.repo.Create(user); err != nil {
        return fmt.Errorf("create user: %w", err)
    }
    
    return nil
}

func (s *Service) validateUser(user *User) error {
    if user.Name == "" {
        return &ValidationError{Field: "name", Message: "name cannot be empty"}
    }
    
    if user.Email == "" {
        return &ValidationError{Field: "email", Message: "email cannot be empty"}
    }
    
    if !isValidEmail(user.Email) {
        return ErrInvalidEmail
    }
    
    if user.Password == "" {
        return &ValidationError{Field: "password", Message: "password cannot be empty"}
    }
    
    if len(user.Password) < 8 {
        return ErrInvalidPassword
    }
    
    return nil
}

func isValidEmail(email string) bool {
    return len(email) > 5 && email[len(email)-4:] == ".com"
}

type Repository interface {
    ExistsByEmail(email string) (bool, error)
    Create(user *User) error
}
```

### 4.6 小结

本章介绍了Go语言的错误处理，包括错误定义、错误处理模式、错误包装、错误恢复、错误日志等。我们学习了Go语言如何定义错误、处理错误、包装错误、恢复panic、记录错误等。掌握好Go语言的错误处理对于编写健壮的程序非常重要。

## 总结

本文档全面介绍了Go语言的工程实践，包括：

1. **项目结构**：标准项目结构、模块管理、代码组织、构建配置
2. **测试**：单元测试、集成测试、基准测试、测试覆盖率
3. **性能优化**：性能分析、内存优化、并发优化、I/O优化
4. **错误处理**：错误定义、错误处理模式、错误包装、错误恢复、错误日志

通过学习本文档，你应该对Go语言的工程实践有了全面的了解，能够更好地使用Go语言进行工程开发。希望本文档能够帮助你在Go语言的学习之路上更进一步！