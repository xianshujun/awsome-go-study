> **[↩ 返回到目录](doc.md)**
---

### 如何编写 Go 代码

**目录**
*   介绍
*   代码组织
*   您的第一个程序
    *   从您的模块导入包
    *   从远程模块导入包
*   测试
*   接下来做什么
*   获取帮助

---

### 介绍

本文档演示了如何在模块内开发一个简单的 Go 包，并介绍了 `go` 命令工具——这是获取、构建和安装 Go 模块、包和命令的标准方式。

---

### 代码组织

Go 程序被组织成**包**（packages）。一个包是同一个目录下一起编译的源文件的集合。在一个源文件中定义的函数、类型、变量和常量，对同一包内的所有其他源文件都是可见的。

一个**代码仓库**（repository）包含一个或多个**模块**（modules）。一个模块是相关 Go 包的集合，它们会一起发布。一个 Go 代码仓库通常只包含一个模块，位于仓库的根目录。该目录下的 `go.mod` 文件声明了**模块路径**（module path）：这是该模块内所有包的导入路径前缀。该模块包含其 `go.mod` 文件所在目录及其子目录中的包，直到遇到下一个包含另一个 `go.mod` 文件的子目录（如果有的话）为止。

> **注意**：在构建代码之前，您不需要将代码发布到远程仓库。模块可以在本地定义，而不必属于某个仓库。但养成像将来要发布代码那样来组织代码的习惯是很好的。

每个模块的路径不仅作为其包的导入路径前缀，还指明了 `go` 命令应该从哪里下载它。例如，为了下载 `golang.org/x/tools` 模块，`go` 命令会查询由 `https://golang.org/x/tools` 指示的仓库（详情见[此处](https://go.dev/ref/mod#vcs-find)）。

**导入路径**（import path）是用于导入包的字符串。一个包的导入路径是其模块路径加上其在模块内的子目录路径。例如，`github.com/google/go-cmp` 模块包含一个位于 `cmp/` 目录下的包。该包的导入路径就是 `github.com/google/go-cmp/cmp`。标准库中的包没有模块路径前缀。

---

### 您的第一个程序

要编译和运行一个简单的程序，首先选择一个模块路径（我们使用 `example/user/hello`），并创建一个声明该路径的 `go.mod` 文件：

```
$ mkdir hello # 或者，如果它已存在于版本控制系统中，则克隆它。
$ cd hello
$ go mod init example/user/hello
go: creating new go.mod: module example/user/hello
$ cat go.mod
module example/user/hello
go 1.16
$
```

Go 源文件中的第一行语句必须是 `package name`。可执行命令必须始终使用 `package main`。

接下来，在该目录中创建一个名为 `hello.go` 的文件，内容如下：

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, world.")
}
```

现在，您可以使用 `go` 工具来构建和安装该程序：

```
$ go install example/user/hello
$
```

此命令构建了 `hello` 命令，生成一个可执行二进制文件。然后，它将该二进制文件安装到 `$HOME/go/bin/hello`（在 Windows 上则是 `%USERPROFILE%\go\bin\hello.exe`）。

安装目录由 `GOPATH` 和 `GOBIN` 环境变量控制。如果设置了 `GOBIN`，二进制文件将被安装到该目录。如果设置了 `GOPATH`，二进制文件将被安装到 `GOPATH` 列表中第一个目录的 `bin` 子目录下。否则，二进制文件将被安装到默认 `GOPATH` 的 `bin` 子目录下（`$HOME/go` 或 `%USERPROFILE%\go`）。

您可以使用 `go env` 命令为未来的 `go` 命令设置环境变量的默认值：

```
$ go env -w GOBIN=/somewhere/else/bin
$
```

要取消之前用 `go env -w` 设置的变量，请使用 `go env -u`：

```
$ go env -u GOBIN
$
```

像 `go install` 这样的 `go` 命令是在包含当前工作目录的模块上下文中执行的。如果工作目录不在 `example/user/hello` 模块内，`go install` 可能会失败。

为了方便起见，`go` 命令接受相对于工作目录的路径，如果没有提供其他路径，则默认使用当前工作目录中的包。因此，在我们的工作目录中，以下命令都是等价的：

```
$ go install example/user/hello
$ go install .
$ go install
```

接下来，让我们运行程序以确保它能正常工作。为了更方便，我们将安装目录添加到我们的 `PATH` 中，以便轻松运行二进制文件：

```bash
# Windows 用户请参阅 /wiki/SettingGOPATH 以设置 %PATH%。
$ export PATH=$PATH:$(dirname $(go list -f '{{.Target}}' .))
$ hello
Hello, world.
$
```

如果您在使用源代码控制系统，现在是初始化仓库、添加文件并提交第一次更改的好时机。同样，这一步是可选的：您不需要使用源代码控制来编写 Go 代码。

```
$ git init
Initialized empty Git repository in /home/user/hello/.git/
$ git add go.mod hello.go
$ git commit -m "initial commit"
[master (root-commit) 0b4507d] initial commit
 1 file changed, 7 insertion(+)
 create mode 100644 go.mod hello.go
$
```

`go` 命令通过请求一个对应的 HTTPS URL 并读取嵌入在 HTML 响应中的元数据来定位包含给定模块路径的仓库（参见 `go help importpath`）。许多托管服务已经为包含 Go 代码的仓库提供了该元数据，因此，让您的模块路径与仓库的 URL 匹配，通常是让其他人可以使用您的模块的最简单方法。

---

### 从您的模块导入包

让我们编写一个 `morestrings` 包，并从 `hello` 程序中使用它。首先，为该包创建一个目录 `$HOME/hello/morestrings`，然后在该目录中创建一个名为 `reverse.go` 的文件，内容如下：

```go
// Package morestrings implements additional functions to manipulate UTF-8
// encoded strings, beyond what is provided in the standard "strings" package.
package morestrings

// ReverseRunes returns its argument string reversed rune-wise left to right.
func ReverseRunes(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}
```

由于我们的 `ReverseRunes` 函数以大写字母开头，因此它是**导出的**（exported），可以在导入了我们 `morestrings` 包的其他包中使用。

让我们用 `go build` 测试该包是否能编译：

```
$ cd $HOME/hello/morestrings
$ go build
$
```

这不会产生输出文件。相反，它会将编译好的包保存在本地构建缓存中。

在确认 `morestrings` 包能构建后，让我们从 `hello` 程序中使用它。为此，修改您原来的 `$HOME/hello/hello.go` 文件以使用 `morestrings` 包：

```go
package main

import (
    "fmt"
    "example/user/hello/morestrings"
)

func main() {
    fmt.Println(morestrings.ReverseRunes("!oG ,olleH"))
}
```

安装 `hello` 程序：

```
$ go install example/user/hello
```

运行程序的新版本，您应该会看到一条新的、被反转的消息：

```
$ hello
Hello, Go!
```

---

### 从远程模块导入包

一个导入路径可以描述如何使用 Git 或 Mercurial 等版本控制系统来获取包的源代码。`go` 工具利用这一特性自动从远程仓库获取包。例如，要在您的程序中使用 `github.com/google/go-cmp/cmp`：

```go
package main

import (
    "fmt"
    "example/user/hello/morestrings"
    "github.com/google/go-cmp/cmp"
)

func main() {
    fmt.Println(morestrings.ReverseRunes("!oG ,olleH"))
    fmt.Println(cmp.Diff("Hello World", "Hello Go"))
}
```

现在您有了一个对外部模块的依赖，您需要下载该模块并将其版本记录在 `go.mod` 文件中。`go mod tidy` 命令会为导入的包添加缺失的模块依赖，并移除不再使用的模块依赖。

```
$ go mod tidy
go: finding module for package github.com/google/go-cmp/cmp
go: found github.com/google/go-cmp/cmp in github.com/google/go-cmp v0.5.4
$ go install example/user/hello
$ hello
Hello, Go!
  string(
-     "Hello World",
+     "Hello Go",
  )
$ cat go.mod
module example/user/hello
go 1.16
require github.com/google/go-cmp v0.5.4
$
```

模块依赖项会自动下载到由 `GOPATH` 环境变量指示的目录的 `pkg/mod` 子目录中。给定版本的模块的下载内容会在所有需要该版本的其他模块之间共享，因此 `go` 命令会将这些文件和目录标记为只读。要删除所有已下载的模块，可以向 `go clean` 命令传递 `-modcache` 标志：

```
$ go clean -modcache
$
```

---

### 测试

Go 拥有一个轻量级的测试框架，由 `go test` 命令和 `testing` 包组成。

您通过创建一个文件名以 `_test.go` 结尾的文件来编写测试，该文件包含名为 `TestXXX` 且签名为 `func (t *testing.T)` 的函数。测试框架会运行每个这样的函数；如果该函数调用了 `t.Error` 或 `t.Fail` 等失败函数，该测试就被认为是失败的。

通过创建文件 `$HOME/hello/morestrings/reverse_test.go` 并包含以下 Go 代码，为 `morestrings` 包添加一个测试。

```go
package morestrings

import "testing"

func TestReverseRunes(t *testing.T) {
    cases := []struct {
        in, want string
    }{
        {"Hello, world", "dlrow ,olleH"},
        {"Hello, 世界", "界世 ,olleH"},
        {"", ""},
    }
    for _, c := range cases {
        got := ReverseRunes(c.in)
        if got != c.want {
            t.Errorf("ReverseRunes(%q) == %q, want %q", c.in, got, c.want)
        }
    }
}
```

然后使用 `go test` 运行测试：

```
$ cd $HOME/hello/morestrings
$ go test
PASS
ok  	example/user/hello/morestrings 0.165s
$
```

运行 `go help test` 并查看 `testing` 包文档以获取更多详细信息。

---

### 接下来做什么

*   订阅 [golang-announce 邮件列表](https://groups.google.com/g/golang-announce)，以便在新的稳定版 Go 发布时收到通知。
*   阅读《[Effective Go](https://go.dev/doc/effective_go)》，了解编写清晰、地道的 Go 代码的技巧。
*   参加《[Go 语言之旅](https://go.dev/tour/)》，学习语言本身。
*   访问[文档页面](https://go.dev/doc/)，获取一系列关于 Go 语言及其库和工具的深度文章。

---

### 获取帮助

*   **实时帮助**：在社区运营的 [Gophers Slack 服务器](https://invite.slack.golangbridge.org/) 中向乐于助人的 Gophers 们提问（[点击此处获取邀请](https://invite.slack.golangbridge.org/)）。
*   **官方邮件列表**：用于讨论 Go 语言的官方邮件列表是 [Go Nuts](https://groups.google.com/g/golang-nuts)。
*   **报告 Bug**：使用 [Go 问题追踪器](https://github.com/golang/go/issues) 来报告 Bug。

> **[↩ 返回到目录](doc.md)**