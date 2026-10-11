> [点此返回上级页面](../README.md)

# 学习资源

本页整理 go.dev/learn 页面在「精选教程」「指导式学习路径」中推荐的站外学习资源，并给出其中文说明与官方入口。

> **说明**：这些资源分别由 Go 官方或第三方维护，正文无法整体搬运到本项目翻译，故在此提供中文导读与官方链接，方便按需访问。

---

<a id="tour"></a>
## Go 语言之旅（A Tour of Go）

官方出品的交互式 Go 入门教程，共分四个部分，每部分结尾都附有若干练习，帮助巩固所学内容：

1. **基础**：基本语法与数据结构（变量、类型、流程控制、切片、映射、闭包等）；
2. **方法与接口**；
3. **泛型**；
4. **并发原语**（goroutine、channel 等）。

可以**在线体验**，也可以本地安装：

```
$ go install golang.org/x/website/tour@latest
```

安装后 `tour` 可执行文件会出现在你 GOPATH 的 `bin` 目录中。

- 官方入口：<https://go.dev/tour/welcome/1>

---

<a id="go-by-example"></a>
## Go by Example

由 Mark McGranaghan 与 Eli Bendersky 维护的**带注释示例程序**合集，是动手学习 Go 的绝佳起点。每个示例都配有可运行代码与逐行注释。除非另有说明，示例以**最新的 Go 主版本**为准，可能会用到较新的语言特性；若遇到无法运行的情况，请尝试升级到最新版本。

示例目录（按主题归类）：

**基础语法**
Hello World · Values · Variables · Constants · For · If/Else · Switch · Arrays · Slices · Maps · Functions · Multiple Return Values · Variadic Functions · Closures · Recursion · Range over Built-in Types · Pointers · Strings and Runes · Structs · Methods · Interfaces · Enums · Struct Embedding · Generics · Range over Iterators · Errors · Custom Errors

**并发**
Goroutines · Channels · Channel Buffering · Channel Synchronization · Channel Directions · Select · Timeouts · Non-Blocking Channel Operations · Closing Channels · Range over Channels · Timers · Tickers · Worker Pools · WaitGroups · Rate Limiting · Atomic Counters · Mutexes · Stateful Goroutines

**排序与错误处理**
Sorting · Sorting by Functions · Panic · Defer · Recover

**字符串、文本与数据格式**
String Functions · String Formatting · Text Templates · Regular Expressions · JSON · XML

**时间、数字与编码**
Time · Epoch · Time Formatting / Parsing · Random Numbers · Number Parsing · URL Parsing · SHA256 Hashes · Base64 Encoding

**文件**
Reading Files · Writing Files · Line Filters · File Paths · Directories · Temporary Files and Directories · Embed Directive

**测试、命令行与系统**
Testing and Benchmarking · Command-Line Arguments · Command-Line Flags · Command-Line Subcommands · Environment Variables · Logging · HTTP Client · HTTP Server · TCP Server · Context · Spawning Processes · Exec'ing Processes · Signals · Exit

- 官方入口：<https://gobyexample.com/>

---

<a id="go-web-examples"></a>
## Go Web Examples（Web 开发）

提供易于理解的代码片段，讲解如何用 Go 进行 Web 开发。它受 Go by Example 启发，目标是给出干净、详尽、可读性强的示例，覆盖从路由、模板到中间件、WebSocket 等 Web 开发基础。

示例目录：

Hello World · HTTP Server · Routing（使用 gorilla/mux）· MySQL Database · Templates · Assets and Files · Forms · Middleware（基础）· Middleware（进阶）· Sessions · JSON · Websockets · Password Hashing

- 官方入口：<https://gowebexamples.com/>

---

<a id="cli"></a>
## 命令行工具（CLIs）工作坊

**Building an Awesome CLI App in Go**，由 Steve Francia 与 Ashley McNamara 在 OSCON 2017 上编写并讲授的工作坊。它梳理了创建**用户友好的命令行界面与命令套件**所需的技术、原则和库，并带你一步步构建自己的 CLI 应用。内容涵盖：从如何设计与构建命令，到处理与解析标志（flags）、配置文件与远程配置系统，再到环境变量与 12-factor 应用。工作坊结束时，你将掌握实用的 Go 知识，并拥有一个可用的 CLI 应用。

**涵盖主题：**

- CLI 应用设计
- Go 编程语言入门
- Cobra CLI 框架介绍（被 Kubernetes、Docker、Hugo 等项目采用）
- 使用 Cobra 构建一个 Go 待办事项（Todo）应用

参与者只需具备通用编程知识，无需任何 Go 语言经验。

- 官方入口：<https://spf13.com/presentation/building-an-awesome-cli-app-in-go-oscon/>

---

<a id="beginner"></a>
## 刚接触编程？

入门课程 **Go Programming Language Tutorial | Golang Tutorial For Beginners**（由 Edureka 提供），面向编程新手，讲解编程基本原理与 Go 基础知识。

- 官方入口：<https://www.youtube.com/watch?v=Q0sKAMal4WQ>

---

> [点此返回上级页面](../README.md)
