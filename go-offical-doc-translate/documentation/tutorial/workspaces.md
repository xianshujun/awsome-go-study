# 教程：开始使用多模块工作区

## 目录

- 先决条件
- 为你的代码创建一个模块
- 创建工作区
- 下载并修改 `golang.org/x/example/hello` 模块
- 了解更多关于工作区的知识

tips: 这节内容较为抽象，如果不太理解，可以查看[详细解释](workspace-easydescribe.md)。

---

本教程将带你了解 Go 中**多模块工作区**（multi-module workspaces）的基础知识。通过多模块工作区，你可以告诉 Go 命令：你正在同时开发多个模块，并希望在这些模块之间轻松地构建和运行代码。

在本教程中，你将在一个共享的多模块工作区中创建两个模块，跨模块进行修改，并立即看到这些更改在构建中的效果。

> **提示**：更多教程请参见 [Go 官方教程列表](https://go.dev/doc/tutorial/)。

---

### 先决条件

- 安装了 **Go 1.18 或更高版本**；
- 一个用于编写代码的编辑器。任何文本编辑器都可以；
- 一个命令行终端。Linux 和 macOS 用户可使用任意终端，Windows 用户可使用 PowerShell 或 cmd。

本教程要求使用 Go 1.18 或更新版本。请确保你已从 [go.dev/dl](https://go.dev/dl) 下载并安装了符合版本要求的 Go。

---

### 为你的代码创建一个模块

我们先从创建一个模块开始，用于存放你即将编写的代码。

1. 打开终端，进入你的主目录。

   在 Linux 或 Mac 上：

   ```bash
   $ cd
   ```

   在 Windows 上：

   ```cmd
   C:\> cd %HOMEPATH%
   ```

   后续教程中的命令提示符统一用 `$` 表示，但这些命令在 Windows 上同样适用。

2. 创建一个名为 `workspace` 的目录，作为你的工作区根目录：

   ```bash
   $ mkdir workspace
   $ cd workspace
   ```

3. 创建并初始化 `hello` 模块

   本示例将创建一个名为 `hello` 的新模块，它会依赖 `golang.org/x/example` 模块。

   ```bash
   $ mkdir hello
   $ cd hello
   $ go mod init example.com/hello
   go: creating new go.mod: module example.com/hello
   ```

4. 添加对 `golang.org/x/example/hello/reverse` 包的依赖：

   ```bash
   $ go get golang.org/x/example/hello/reverse
   ```

5. 在 `hello` 目录下创建 `hello.go` 文件，内容如下：

   ```go
   package main

   import (
       "fmt"
       "golang.org/x/example/hello/reverse"
   )

   func main() {
       fmt.Println(reverse.String("Hello"))
   }
   ```

6. 运行程序：

   ```bash
   $ go run .
   olleH
   ```

   程序成功将 `"Hello"` 反转为 `"olleH"`，说明我们已经正确使用了外部模块中的函数。

---

### 创建工作区

接下来，我们将创建一个 `go.work` 文件，定义一个包含多个模块的工作区。

#### 初始化工作区

在 `workspace` 目录下运行：

```bash
$ go work init ./hello
```

`go work init` 命令会为包含 `./hello` 模块的工作区创建一个 `go.work` 文件。

Go 工具会生成如下内容的 `go.work` 文件：

```go
go 1.18

use ./hello
```

`go.work` 文件的语法与 `go.mod` 类似：

- `go` 指令指明了解析该文件所需的 Go 版本，类似于 `go.mod` 中的同名指令；
- `use` 指令告诉 Go，在构建时应将 `hello` 目录中的模块视为“主模块”。

这意味着，在 `workspace` 的任意子目录中，该工作区的配置都会生效。

#### 在工作区目录中运行程序

在 `workspace` 根目录下执行：

```bash
$ go run ./hello
olleH
```

Go 命令会自动将工作区中所有模块都视为主模块，因此即使你在模块外部运行命令，也能正确解析依赖。如果不使用工作区，这样的调用会失败，因为 Go 无法确定该使用哪个模块。

---

### 下载并修改 `golang.org/x/example/hello` 模块

接下来，我们将把 `golang.org/x/example/hello` 模块的源码下载到本地，加入工作区，并为其 `reverse` 包添加一个新功能 —— 反转整数。然后从我们的 `hello` 程序中调用这个新函数。

#### 1. 克隆仓库

在 `workspace` 目录下运行：

```bash
$ git clone https://go.googlesource.com/example
```

输出类似：

```
Cloning into 'example'...
remote: Total 165 (delta 27), reused 165 (delta 27)
Receiving objects: 100% (165/165), 434.18 KiB | 1022.00 KiB/s, done.
Resolving deltas: 100% (27/27), done.
```

克隆完成后，源码将位于 `./example` 目录中，而 `golang.org/x/example/hello` 模块的实际代码位于 `./example/hello`。

#### 2. 将模块加入工作区

运行以下命令，将本地模块添加到工作区：

```bash
$ go work use ./example/hello
```

该命令会在 `go.work` 文件中添加一条新的 `use` 指令。现在文件内容变为：

```go
go 1.18

use (
    ./hello
    ./example/hello
)
```

现在，工作区同时包含了：

- 我们自己的 `example.com/hello` 模块；
- 来自 `golang.org/x/example/hello` 的模块，它提供了 `reverse` 包。

**关键好处**：由于我们将本地副本加入了工作区，Go 会优先使用我们本地的 `reverse` 包，而不是之前通过 `go get` 下载到模块缓存中的版本。这使得我们可以自由修改并立即测试。

#### 3. 添加新函数

我们将在 `reverse` 包中添加一个用于反转整数的新函数。

创建文件 `workspace/example/hello/reverse/int.go`，内容如下：

```go
package reverse

import "strconv"

// Int 返回整数 i 的十进制数字反转结果
func Int(i int) int {
    i, _ = strconv.Atoi(String(strconv.Itoa(i)))
    return i
}
```

这个函数通过将整数转为字符串、调用已有的 `String` 函数反转，再转回整数，实现数字反转。

#### 4. 修改主程序以使用新函数

修改 `workspace/hello/hello.go` 文件，更新 `main` 函数：

```go
package main

import (
    "fmt"
    "golang.org/x/example/hello/reverse"
)

func main() {
    fmt.Println(reverse.String("Hello"), reverse.Int(24601))
}
```

我们同时调用了 `reverse.String` 和新的 `reverse.Int` 函数。

---

### 在工作区中运行代码

在 `workspace` 根目录下运行：

```bash
$ go run ./hello
olleH 10642
```

输出表明：

- `"Hello"` 被反转为 `"olleH"`；
- `24601` 被反转为 `10642`。

Go 命令通过 `go.work` 文件正确解析了两个模块的路径：

- 找到了 `hello` 模块；
- 使用本地的 `golang.org/x/example/hello` 模块来解析 `reverse` 包。

> **提示**：使用 `go.work` 可以避免在 `go.mod` 中手动添加 `replace` 指令，特别适合在多个模块间协同开发。

由于两个模块处于同一工作区，你可以在一个模块中修改代码，并立即在另一个模块中使用，开发效率大大提高。

---

### 后续步骤

当你完成开发并希望正式发布这些模块时，你需要为 `golang.org/x/example/hello` 模块发布一个正式版本，例如 `v0.1.0`。这通常通过在版本控制系统中为某个提交打上标签（tag）来实现。

发布完成后，你可以更新 `hello` 模块对它的依赖：

```bash
cd hello
go get golang.org/x/example/hello@v0.1.0
```

这样，即使脱离工作区，Go 命令也能正确解析并下载这个版本的模块。

---

### 了解更多关于工作区的知识

除了本教程中使用的 `go work init`，Go 还提供了其他几个用于管理工作区的子命令：

- `go work use [-r] [dir]`  
  将指定目录添加到 `go.work` 的 `use` 列表中；如果目录不存在，则将其从列表中移除。`-r` 参数表示递归检查子目录。

- `go work edit`  
  类似于 `go mod edit`，可直接编辑 `go.work` 文件，比如添加或删除模块。

- `go work sync`  
  将工作区构建列表中的依赖项同步到各个模块的 `go.mod` 文件中，确保依赖一致性。

更多细节，请参考 [Go 模块参考文档中的“工作区”章节](https://go.dev/ref/mod#workspaces)。

---

🎉 恭喜！你已经掌握了 Go 多模块工作区的基本用法。这种机制非常适合在微服务、组件库或多团队协作项目中进行高效开发。继续探索，让 Go 的模块系统为你的项目赋能！