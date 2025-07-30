# 教程：开始使用 Go 语言

> **[↩ 返回到目录](../doc.md)**
## 目录

- 先决条件
- 安装 Go
- 编写一些代码
- 调用外部包中的代码
- 编写更多代码

---

在本教程中，你将快速入门 Go 语言编程。过程中你会完成以下内容：

- 安装 Go（如果尚未安装）；
- 编写一段简单的“Hello, world”程序；
- 使用 `go` 命令运行你的代码；
- 使用 Go 的包发现工具查找可在项目中使用的第三方包；
- 调用外部模块中的函数。

> **提示**：更多教程请参见 [教程页面](https://go.dev/doc/tutorial/)。

---

### 先决条件

- 具备一定的编程经验。虽然这里的代码很简单，但了解函数的基本概念会有所帮助。
- 一个用于编辑代码的工具。任何文本编辑器都可以胜任。大多数主流编辑器对 Go 都有良好支持，其中最受欢迎的是 [VSCode](https://code.visualstudio.com/)（免费）、[GoLand](https://www.jetbrains.com/go/)（付费）和 [Vim](https://www.vim.org/)（免费）。
- 一个命令行终端。Linux 和 macOS 用户可以使用任意终端，Windows 用户可使用 PowerShell 或 cmd。

---

### 安装 Go

请直接按照 [下载与安装步骤](https://go.dev/dl/) 进行操作即可。

---

### 编写一些代码

从经典的 “Hello, World” 开始你的 Go 之旅。

1. 打开命令行终端，并进入你的主目录。

   在 Linux 或 Mac 上：

   ```bash
   cd
   ```

   在 Windows 上：

   ```cmd
   cd %HOMEPATH%
   ```

2. 创建一个名为 `hello` 的目录，用来存放你的第一个 Go 程序。

   执行以下命令：

   ```bash
   mkdir hello
   cd hello
   ```

3. 启用依赖管理。

   当你的代码需要导入其他模块中的包时，你需要通过一个 `go.mod` 文件来管理这些依赖关系。这个文件定义了你当前项目的模块，并记录所有依赖的外部模块。

   运行 `go mod init` 命令来初始化模块，并指定模块名称（即模块路径）。

   在实际开发中，模块路径通常是你代码仓库的地址，比如 `github.com/yourname/mymodule`。如果你打算将模块发布供他人使用，这个路径必须是 Go 工具能够访问并下载的位置。

   对于本教程，我们使用一个示例路径：

   ```bash
   go mod init example/hello
   ```

   输出结果如下：

   ```
   go: creating new go.mod: module example/hello
   ```

4. 在文本编辑器中创建一个名为 `hello.go` 的文件，用于编写代码。

   将以下代码粘贴进去并保存：

   ```go
   package main

   import "fmt"

   func main() {
       fmt.Println("Hello, World!")
   }
   ```

   这段代码的含义是：

    - 使用 `package main` 声明主包（包是函数的集合，同一目录下的所有文件属于同一个包）；
    - 导入标准库中的 `fmt` 包，它提供了格式化输出的功能，比如向控制台打印信息；
    - 定义一个 `main` 函数，这是程序的入口点，运行主包时会自动执行该函数。

5. 运行你的程序，查看输出结果：

   ```bash
   go run .
   ```

   输出：

   ```
   Hello, World!
   ```

6. 探索更多 `go` 命令：

   ```bash
   go help
   ```

   这会列出所有可用的 Go 命令，帮助你更高效地开发。

---

### 调用外部包中的代码

当你需要实现某个功能，而别人可能已经写好了现成的代码时，就可以寻找合适的第三方包来复用。

1. 让我们的输出变得更有趣一点，借助一个外部模块中的函数来实现。

    - 打开 [pkg.go.dev](https://pkg.go.dev) 网站；
    - 搜索 “quote” 包；
    - 在搜索结果中找到 `rsc.io/quote` 模块的 v1 版本（应显示在 “Other major versions” 下的 `rsc.io/quote/v4` 旁边）；
    - 点击进入该包的页面；
    - 在文档的 “Index” 部分，你会看到一系列可调用的函数，我们将使用其中的 `Go()` 函数；
    - 注意顶部提示：`quote` 包属于 `rsc.io/quote` 模块。

   > **小知识**：你可以通过 pkg.go.dev 查找公开发布的模块，它们包含了各种实用的包。这些模块会不断更新迭代，你也可以随时升级项目以使用新版本。

2. 修改你的 Go 代码，导入 `rsc.io/quote` 包并调用其 `Go()` 函数。

   修改后的 `hello.go` 文件内容如下：

   ```go
   package main

   import "fmt"
   import "rsc.io/quote"

   func main() {
       fmt.Println(quote.Go())
   }
   ```

3. 更新依赖项。

   运行以下命令，让 Go 自动下载所需的模块并更新校验文件：

   ```bash
   go mod tidy
   ```

   输出示例：

   ```
   go: finding module for package rsc.io/quote
   go: found rsc.io/quote in rsc.io/quote v1.5.2
   ```

   此命令会：
    - 自动添加 `rsc.io/quote` 模块作为依赖；
    - 创建 `go.sum` 文件，用于验证模块完整性，防止被篡改。

4. 再次运行程序：

   ```bash
   go run .
   ```

   输出：

   ```
   Don't communicate by sharing memory, share memory by communicating.
   ```

   看到了吗？这句经典语录正是来自 `quote.Go()` 函数！

   > 这句话体现了 Go 语言并发设计的核心哲学：不要通过共享内存来通信，而应通过通信来共享内存。

   `go mod tidy` 自动下载了 `rsc.io/quote` 模块（默认最新版本 v1.5.2），这样你的程序才能成功调用其中的函数。

---

### 编写更多代码

通过这个简短的入门教程，你已经完成了 Go 的安装，学会了基础语法，并体验了如何使用外部模块。现在你已经有了一个良好的开端！

如果你想继续深入学习，不妨尝试下一个教程：[创建一个 Go 模块](https://go.dev/doc/tutorial/create-module)。

> **[↩ 返回到目录](../doc.md)**