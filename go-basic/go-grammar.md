# Go语言语法 - 打好基础

> 面向对象：刚接触 Go、想系统打牢语法基础，或正在准备 Go 相关面试的开发者。
> 阅读方式：每节先讲"是什么、为什么"（配生活化类比与图示），再给**可直接运行的最小示例**，最后补"常见坑"与本页面试题。
> 环境要求：Go 1.22+（正文基础示例兼容更早版本，`go mod` 相关用法以 1.22+ 为准）；Linux/macOS/Windows 均可。

## 目录

- [第一章：Go语言简介与环境搭建](#第一章go语言简介与环境搭建)
- [第二章：变量与常量](#第二章变量与常量)
- [第三章：基本数据类型](#第三章基本数据类型)
- [第四章：运算符](#第四章运算符)
- [第五章：控制流](#第五章控制流)
- [第六章：函数](#第六章函数)
- [第七章：数组、切片和映射](#第七章数组切片和映射)
- [第八章：结构体和方法](#第八章结构体和方法)
- [第九章：错误处理和测试](#第九章错误处理和测试)
- [第十章：包管理和模块](#第十章包管理和模块)
- [总结](#总结)
- [附录：高频面试题汇总](#附录高频面试题汇总)

---

## 第一章：Go语言简介与环境搭建

> 本章总览：从装好环境到跑通第一个程序，只需五步。

```mermaid
flowchart TD
    A[下载并安装 Go] --> B[配置 PATH 与 GOPATH]
    B --> C[执行 go version 验证]
    C --> D[编写 hello.go]
    D --> E[go run hello.go]
    E --> F[看到 Hello World 输出]
    C -.失败.-> B
```

### 1.1 Go语言是什么

Go语言（又称Golang）是由Google在2007年开始开发，2009年正式开源的一种编程语言。它的设计目标是让编程变得更加简单、高效和可靠。Go语言的创造者包括Ken Thompson（Unix之父）、Rob Pike（UTF-8设计者）和Robert Griesemer，这些计算机科学界的巨匠为Go语言注入了深厚的理论基础和工程实践。

Go语言的主要特点可以概括为以下几个方面：

**简单易学**：Go语言的语法设计非常简洁，只有25个关键字，比大多数编程语言都要少。这意味着初学者可以在很短的时间内掌握基本语法，快速上手编写程序。Go语言的设计哲学是"少即是多"，通过减少语言特性的数量，让开发者能够专注于解决问题本身，而不是纠结于复杂的语法细节。

**高性能**：Go语言是一种编译型语言，代码会被编译成机器码直接运行，不需要像解释型语言那样在运行时进行解释。这使得Go程序的运行速度非常接近C和C++等传统高性能语言。同时，Go语言的垃圾回收机制经过精心设计，能够在保证内存安全的同时，将性能开销降到最低。

**并发支持**：这是Go语言最显著的特点之一。Go语言内置了goroutine（轻量级线程）和channel（通道）等并发原语，让并发编程变得简单而安全。开发者可以轻松地创建成千上万个goroutine，而不用担心线程创建和切换的开销。这使得Go语言特别适合编写高并发、高吞吐量的网络服务。

**静态类型**：Go语言是静态类型语言，这意味着变量的类型在编译时就能确定，编译器可以在编译阶段发现很多类型错误。但Go语言的类型系统又非常灵活，支持类型推断等特性，既保证了类型安全，又不会让代码显得过于冗长。

**标准库丰富**：Go语言自带了功能强大的标准库，涵盖了网络编程、文件操作、加密解密、JSON处理、HTTP服务等各个方面。开发者可以直接使用这些标准库，而不需要依赖大量的第三方库，这大大提高了开发效率和代码的可靠性。

**跨平台**：Go语言支持多种操作系统和硬件架构，包括Windows、Linux、macOS等主流操作系统，以及x86、ARM等硬件架构。开发者可以在一个平台上编写代码，然后轻松地编译到其他平台上运行，这极大地简化了跨平台开发的流程。

**工具链完善**：Go语言自带了一套完整的开发工具，包括编译器、格式化工具、测试工具、文档生成工具等。特别是`go fmt`命令，可以自动格式化代码，保证了团队中不同开发者编写的代码风格一致，这对于大型项目的协作开发非常重要。

### 1.2 为什么选择Go语言

在选择编程语言时，我们需要考虑多个因素，包括学习曲线、开发效率、运行性能、生态系统等。Go语言在这些方面都有不错的表现，特别适合以下场景：

**网络服务开发**：Go语言天生适合编写网络服务。它的标准库中包含了完善的HTTP服务器和客户端实现，开发者可以用很少的代码就构建出高性能的Web服务。同时，Go语言的并发特性使得处理大量并发请求变得轻而易举，这对于现代互联网应用来说非常重要。

**微服务架构**：微服务架构要求每个服务都是独立部署、独立运行的，Go语言编译后的单个可执行文件不依赖外部库，部署非常方便。同时，Go语言的启动速度快、内存占用低，非常适合作为微服务的实现语言。

**云原生应用**：Docker和Kubernetes等云原生技术都是用Go语言编写的，这说明Go语言在云原生领域的地位。Go语言的静态类型和编译特性使得它非常适合构建容器化的应用。

**命令行工具**：Go语言可以编译成单个可执行文件，不需要安装运行时环境，这使得它非常适合开发命令行工具。很多流行的开发工具都是用Go语言编写的，比如Hugo（静态网站生成器）、gRPC（RPC框架）等。

**系统编程**：虽然Go语言不能像C语言那样直接操作硬件，但它提供了足够的底层访问能力，可以用于系统编程。同时，Go语言的内存安全特性避免了C语言中常见的内存泄漏、空指针等问题。

### 1.3 安装Go语言环境

在开始编写Go程序之前，我们需要先安装Go语言环境。Go语言支持多种操作系统，安装过程都比较简单。

**在Linux上安装Go语言**

对于Ubuntu/Debian系统，可以使用包管理器安装：

```bash
sudo apt-get update
sudo apt-get install golang-go
```

但是包管理器中的Go版本可能不是最新的，推荐从官方网站下载最新版本：

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
```

然后需要配置环境变量，编辑`~/.bashrc`或`~/.zshrc`文件，添加以下内容：

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

保存文件后，执行`source ~/.bashrc`或`source ~/.zshrc`使配置生效。

**在macOS上安装Go语言**

macOS用户可以使用Homebrew安装：

```bash
brew install go
```

或者从官方网站下载安装包，双击安装即可。

**在Windows上安装Go语言**

Windows用户可以从Go语言官网下载安装包，双击运行安装程序，按照提示完成安装。安装程序会自动配置环境变量。

**验证安装**

安装完成后，打开终端或命令提示符，输入以下命令验证安装是否成功：

```bash
go version
```

如果看到类似`go version go1.21.5 linux/amd64`的输出，说明Go语言已经成功安装。

### 1.4 Go语言工作空间

Go语言使用工作空间（workspace）的概念来组织代码。默认情况下，Go语言的工作空间是用户主目录下的`go`文件夹，这个文件夹通常被称为`GOPATH`。

工作空间包含三个子目录：

- `src`：存放Go源代码
- `pkg`：存放编译后的包文件
- `bin`：存放编译后的可执行文件

从Go 1.11版本开始，Go语言引入了Go Modules（模块）机制，这使得开发者可以在任意目录下创建Go项目，而不需要强制使用`GOPATH`。Go Modules已经成为Go语言推荐的依赖管理方式，我们将在后续章节中详细介绍。

### 1.5 第一个Go程序

让我们编写第一个Go程序，按照传统，我们从"Hello, World!"开始。

> 💡 通俗理解：Go 程序就像一份"说明书"：`package` 声明它属于哪个"部门"，`import` 说明要用哪些"外部工具"，`func main` 是唯一的总入口——程序永远从 `main` 函数的第一行开始往下走。所以无论代码写得多花哨，找到 `main` 就等于找到了起点。

> ⚠️ **常见坑**：入口文件必须有 `package main`，且函数名必须是全小写的 `main`。少了 `package main` 会报 `expected 'package', found ...`；把 `main` 写成 `Main` 或改成别的名字，会报 `function main is undeclared in the main package`。文件名本身随意，入口函数名不能变。

创建一个文件`hello.go`，内容如下：

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

让我们逐行解释这个程序：

- `package main`：声明这个文件属于`main`包。在Go语言中，每个文件都必须属于一个包，`main`包是一个特殊的包，它定义了一个可独立执行的程序。
- `import "fmt"`：导入`fmt`包，这个包提供了格式化输入输出的功能。
- `func main()`：定义`main`函数，这是程序的入口点。当运行一个Go程序时，会先执行`main`包中的`main`函数。
- `fmt.Println("Hello, World!")`：调用`fmt`包的`Println`函数，输出字符串"Hello, World!"并换行。

运行这个程序：

```bash
go run hello.go
```

你会看到输出：

```
Hello, World!
```

也可以先编译再运行：

```bash
go build hello.go
./hello
```

或者直接使用`go run`命令，它会自动编译并运行程序。

### 1.6 Go语言开发工具

虽然使用文本编辑器和命令行就可以开发Go程序，但使用合适的开发工具可以大大提高开发效率。

**Visual Studio Code**

VS Code是目前最流行的Go语言开发工具之一，它有丰富的Go语言插件支持。安装Go插件后，VS Code可以提供代码补全、语法高亮、错误提示、调试等功能。

安装步骤：
1. 下载并安装VS Code
2. 在扩展市场中搜索"Go"
3. 安装Google官方的Go插件
4. 安装完成后，VS Code会提示安装Go语言相关的工具，按照提示安装即可

**GoLand**

GoLand是JetBrains公司推出的Go语言IDE，功能非常强大，包括智能代码补全、代码重构、调试、版本控制等。GoLand是付费软件，但对学生和开源项目开发者免费。

**其他工具**

除了IDE，还有一些命令行工具可以提高开发效率：

- `gofmt`：自动格式化Go代码
- `go vet`：检查代码中的常见错误
- `golint`：检查代码风格是否符合Go语言规范
- `go test`：运行测试
- `go doc`：查看文档

这些工具都可以通过命令行直接使用，也可以集成到IDE中。

### 1.7 小结

本章介绍了Go语言的基本概念、特点、安装方法和开发环境搭建。Go语言以其简洁的语法、强大的并发支持和完善的工具链，成为了现代软件开发的重要选择。在接下来的章节中，我们将深入学习Go语言的语法特性，从基础到进阶，逐步掌握这门强大的编程语言。

下一章我们将开始学习Go语言的基础语法，包括变量、常量、数据类型、运算符等内容。这些是编写Go程序的基础，掌握好这些内容对于后续的学习至关重要。

**面试题**

1. Go 语言有哪些主要特点？为什么它特别适合云原生与微服务？
2. Go 是编译型还是解释型语言？`go run` 与 `go build` 有什么区别？
3. `GOPATH` 与 Go Modules 是什么关系？为什么现在不推荐用 `GOPATH` 组织项目？
4. `package main` 和 `main` 函数有什么特殊之处？入口文件能否取别的名字？
5. `gofmt`、`go vet`、`go test` 这些工具分别解决什么问题？

## 第二章：变量与常量

> 本章总览：变量在哪一层声明，决定了它能被哪一层看见。

```mermaid
flowchart TD
    P[全局作用域 包级变量] --> L[函数作用域 局部变量]
    L --> B[块作用域 if for 内部]
    B --> S[内层同名变量 遮蔽外层]
    P -.首字母大写.-> E[可被其它包访问 导出]
    P -.首字母小写.-> N[仅包内可见 不导出]
```

### 2.1 变量的声明与使用

变量是程序中用于存储数据的容器。在Go语言中，变量的声明和使用非常灵活，有多种方式可以声明变量。

**基本变量声明**

Go语言使用`var`关键字来声明变量，基本语法如下：

```go
var 变量名 类型
```

例如：

```go
var name string
var age int
var isStudent bool
```

这里声明了三个变量：`name`是字符串类型，`age`是整数类型，`isStudent`是布尔类型。声明变量后，我们可以给它们赋值：

```go
name = "张三"
age = 20
isStudent = true
```

**变量初始化**

在声明变量的同时可以给变量赋初始值：

```go
var name string = "李四"
var age int = 25
var isStudent bool = false
```

**类型推断**

Go语言支持类型推断，当我们给变量赋初始值时，可以省略类型声明，编译器会根据初始值的类型自动推断变量的类型：

```go
var name = "王五"
var age = 30
var isStudent = true
```

这里`name`会被推断为`string`类型，`age`会被推断为`int`类型，`isStudent`会被推断为`bool`类型。

**短变量声明**

在函数内部，可以使用更简洁的短变量声明方式，使用`:=`操作符：

```go
name := "赵六"
age := 35
isStudent := false
```

短变量声明只能在函数内部使用，不能在函数外部使用。这是Go语言的一个限制，目的是保持代码的清晰性。

> ⚠️ **常见坑**：`:=` 左侧至少要有一个**新变量**，否则会报 `no new variables on left side of :=`。更隐蔽的是"变量遮蔽"：在 `if`、`for`、函数内部本想给外层变量赋值，却写成了 `:=`，结果创建出一个只在当前块可见的新变量，外层那个其实一点没变，排查起来很费劲。

**多变量声明**

Go语言支持同时声明多个变量：

```go
var name, age, isStudent string, int, bool
```

或者：

```go
var (
    name string
    age int
    isStudent bool
)
```

同时赋值：

```go
var name, age, isStudent = "钱七", 40, true
```

短变量声明也支持多变量：

```go
name, age, isStudent := "孙八", 45, false
```

**变量交换**

Go语言支持直接交换两个变量的值，不需要使用临时变量：

```go
a := 10
b := 20
a, b = b, a
fmt.Println(a, b) // 输出: 20 10
```

这个特性非常方便，特别是在排序算法中经常需要交换变量的值。

### 2.2 变量的作用域

变量的作用域指的是变量在程序中可以被访问的范围。Go语言中变量的作用域分为三种：局部变量、全局变量和形式参数。

> 💡 通俗理解：作用域就像公司里的"工牌权限"。全局变量是大堂公告，全公司都能看到；局部变量是你所在部门的内部邮件，出了部门门就作废；形式参数是会议室里的临时铭牌，会议一散就回收。内层同名变量遮蔽外层，相当于"本部门临时规定"盖过了"公司公告"，可一走出这个范围就恢复原样。

**局部变量**

在函数内部声明的变量称为局部变量，它们只能在函数内部访问：

```go
func main() {
    var x int = 10
    fmt.Println(x) // 可以访问
}

func another() {
    fmt.Println(x) // 编译错误：x未定义
}
```

局部变量的生命周期从声明开始，到函数结束时结束。函数结束后，局部变量会被销毁，释放占用的内存。

**全局变量**

在函数外部声明的变量称为全局变量，它们可以在整个包中被访问：

```go
var globalVar int = 100

func main() {
    fmt.Println(globalVar) // 可以访问
}

func another() {
    fmt.Println(globalVar) // 也可以访问
}
```

全局变量的生命周期从程序开始到程序结束。如果全局变量的首字母大写，它还可以被其他包访问。

**形式参数**

函数的参数也是局部变量的一种，它们的作用域仅限于函数内部：

```go
func add(a int, b int) int {
    return a + b
}

func main() {
    fmt.Println(a) // 编译错误：a未定义
}
```

**变量遮蔽**

在内部作用域中声明的变量可以遮蔽外部作用域中同名变量：

```go
var x int = 10

func main() {
    x := 20 // 遮蔽了全局变量x
    fmt.Println(x) // 输出: 20
    {
        x := 30 // 遮蔽了外层变量x
        fmt.Println(x) // 输出: 30
    }
    fmt.Println(x) // 输出: 20
}
```

虽然变量遮蔽是合法的，但不推荐使用，因为它会让代码变得难以理解。

### 2.3 常量的声明与使用

常量是指在程序运行期间其值不能改变的量。Go语言使用`const`关键字来声明常量。

**基本常量声明**

```go
const pi = 3.14159
const maxUsers = 1000
const appName = "MyApp"
```

常量声明时必须赋初始值，因为常量的值在声明后就不能再改变。

**类型推断**

与变量一样，常量也支持类型推断：

```go
const pi = 3.14159        // 被推断为float64
const maxUsers = 1000     // 被推断为int
const appName = "MyApp"   // 被推断为string
```

**指定类型**

也可以显式指定常量的类型：

```go
const pi float64 = 3.14159
const maxUsers int = 1000
const appName string = "MyApp"
```

**多常量声明**

可以同时声明多个常量：

```go
const (
    pi = 3.14159
    e = 2.71828
    goldenRatio = 1.618
)
```

**iota常量生成器**

`iota`是Go语言中的一个特殊常量，它可以被认为是一个可以被编译器修改的常量。`iota`在`const`关键字出现时被重置为0，在下一个`const`出现之前，每出现一行，`iota`所代表的数字会自动加1。

```go
const (
    Monday = iota    // 0
    Tuesday          // 1
    Wednesday        // 2
    Thursday         // 3
    Friday           // 4
    Saturday         // 5
    Sunday           // 6
)
```

`iota`也可以用于更复杂的表达式：

```go
const (
    _ = iota             // 0
    KB = 1 << (10 * iota) // 1024
    MB                    // 1048576
    GB                    // 1073741824
    TB                    // 1099511627776
)
```

这里使用了位运算，`1 << (10 * iota)`表示1左移10*iota位，即2的(10*iota)次方。

### 2.4 变量与常量的命名规范

良好的命名规范可以让代码更易读、更易维护。Go语言有一套推荐的命名规范。

**驼峰命名法**

Go语言推荐使用驼峰命名法（CamelCase），即每个单词的首字母大写（大驼峰）或除第一个单词外的每个单词首字母大写（小驼峰）。

```go
var userName string        // 小驼峰
var GetUserAge func() int // 大驼峰
```

**导出与不导出**

Go语言通过首字母的大小写来决定变量或函数是否可以被其他包访问。首字母大写的变量或函数可以被其他包访问（导出），首字母小写的变量或函数只能在当前包内访问（不导出）。

```go
var ExportedVar int     // 可以被其他包访问
var unexportedVar int   // 只能在当前包内访问
```

**有意义的命名**

变量名应该能够清楚地表达变量的用途，避免使用单个字母或无意义的缩写：

```go
// 好的命名
var userName string
var userAge int
var isLoggedIn bool

// 不好的命名
var n string
var a int
var b bool
```

**常量命名**

常量通常使用全大写字母，单词之间用下划线分隔：

```go
const MAX_CONNECTIONS = 100
const DEFAULT_TIMEOUT = 30
const API_VERSION = "v1"
```

### 2.5 变量的零值

在Go语言中，变量在声明时如果没有显式初始化，会被赋予一个默认值，称为零值。不同类型的零值不同：

- 数值类型（int、float等）：0
- 布尔类型：false
- 字符串类型：""（空字符串）
- 指针、函数、接口、切片、映射、通道：nil

```go
var i int
var f float64
var b bool
var s string
var p *int

fmt.Println(i) // 0
fmt.Println(f) // 0
fmt.Println(b) // false
fmt.Println(s) // ""（空字符串）
fmt.Println(p) // <nil>
```

### 2.6 类型转换

Go语言是强类型语言，不同类型的变量之间不能直接赋值，需要进行类型转换。

**基本类型转换**

```go
var i int = 42
var f float64 = float64(i)
var s string = string(i) // 注意：这会将数字转换为对应的ASCII字符，而不是数字的字符串表示
```

**字符串与数字转换**

将字符串转换为数字需要使用`strconv`包：

```go
import "strconv"

s := "123"
i, err := strconv.Atoi(s) // 字符串转整数
if err != nil {
    fmt.Println("转换失败")
}

f, err := strconv.ParseFloat(s, 64) // 字符串转浮点数
if err != nil {
    fmt.Println("转换失败")
}

num := 456
str := strconv.Itoa(num) // 整数转字符串
str = strconv.FormatFloat(3.14, 'f', 2, 64) // 浮点数转字符串
```

### 2.7 实例：计算圆的面积和周长

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个程序，计算圆的面积和周长。

```go
package main

import (
    "fmt"
    "math"
)

func main() {
    const pi = 3.14159
    
    var radius float64
    fmt.Print("请输入圆的半径: ")
    fmt.Scanln(&radius)
    
    area := pi * radius * radius
    circumference := 2 * pi * radius
    
    fmt.Printf("圆的面积: %.2f\n", area)
    fmt.Printf("圆的周长: %.2f\n", circumference)
}
```

这个程序：
1. 定义了一个常量`pi`表示圆周率
2. 声明了一个变量`radius`来存储用户输入的半径
3. 使用短变量声明计算面积和周长
4. 使用`fmt.Printf`格式化输出结果

运行这个程序，输入半径5，输出如下：

```
请输入圆的半径: 5
圆的面积: 78.54
圆的周长: 31.42
```

### 2.8 小结

本章介绍了Go语言中变量和常量的声明、使用和作用域。变量是程序中用于存储数据的容器，常量是程序运行期间不能改变的量。我们学习了多种声明变量的方式，包括基本声明、初始化、类型推断和短变量声明。我们还了解了变量的作用域、常量的使用以及`iota`常量生成器。

下一章我们将学习Go语言的基本数据类型，包括数值类型、字符串类型、布尔类型等，以及它们的使用方法和注意事项。

**面试题**

1. `var`、`:=`、`const` 三者的使用场景与限制分别是什么？
2. 短变量声明 `:=` 与赋值 `=` 有什么区别？为什么 `:=` 不能用于函数外？
3. Go 变量的零值是什么？不同类型各自的零值分别是什么？
4. 什么是变量遮蔽？它可能带来什么隐患？
5. `iota` 的规则是什么？如何用它定义一组枚举？

## 第三章：基本数据类型

> 本章总览：Go 的基本类型分成四大类，选型时先看范围再谈精度。

```mermaid
flowchart TD
    T[Go 基本数据类型] --> N[数值类型]
    T --> B[布尔类型 bool]
    T --> S[字符串类型 string]
    T --> C[字符类型 byte 与 rune]
    N --> I[整数 int 系列]
    N --> F[浮点 float32 float64]
    N --> X[复数 complex64 complex128]
    S --> SB[按字节索引 按字符遍历]
```

### 3.1 数值类型

Go语言提供了丰富的数值类型，可以分为整数类型、浮点类型和复数类型。

**整数类型**

Go语言提供了多种整数类型，每种类型的大小和取值范围都不同：

| 类型 | 大小 | 取值范围 |
|------|------|----------|
| int8 | 8位 | -128 到 127 |
| int16 | 16位 | -32768 到 32767 |
| int32 | 32位 | -2147483648 到 2147483647 |
| int64 | 64位 | -9223372036854775808 到 9223372036854775807 |
| uint8 | 8位 | 0 到 255 |
| uint16 | 16位 | 0 到 65535 |
| uint32 | 32位 | 0 到 4294967295 |
| uint64 | 64位 | 0 到 18446744073709551615 |

除了这些明确大小的整数类型，Go语言还提供了`int`和`uint`类型，它们的大小与平台相关，在32位系统上是32位，在64位系统上是64位。

```go
var a int8 = 100
var b int16 = 1000
var c int32 = 100000
var d int64 = 1000000000

var e uint8 = 200
var f uint16 = 5000
var g uint32 = 1000000
var h uint64 = 10000000000
```

**字节类型**

`byte`类型是`uint8`的别名，通常用于存储ASCII字符：

```go
var ch byte = 'A' // 65
fmt.Println(ch)   // 输出: 65
```

**浮点类型**

Go语言提供了两种浮点类型：

| 类型 | 大小 | 精度 |
|------|------|------|
| float32 | 32位 | 约7位小数 |
| float64 | 64位 | 约15位小数 |

`float64`是Go语言的默认浮点类型，它提供了更高的精度。

```go
var f1 float32 = 3.14
var f2 float64 = 3.141592653589793
```

**复数类型**

Go语言还支持复数类型，这在科学计算中非常有用：

| 类型 | 大小 | 说明 |
|------|------|------|
| complex64 | 64位 | 实部和虚部都是float32 |
| complex128 | 128位 | 实部和虚部都是float64 |

```go
var c1 complex64 = 1 + 2i
var c2 complex128 = 3 + 4i

fmt.Println(c1) // 输出: (1+2i)
fmt.Println(c2) // 输出: (3+4i)

fmt.Println(real(c1)) // 输出实部: 1
fmt.Println(imag(c1)) // 输出虚部: 2
```

**数值类型的选择**

选择合适的数值类型很重要，需要考虑以下因素：

1. **取值范围**：确保类型能够存储所有可能的值
2. **内存占用**：在满足需求的前提下，选择占用内存更小的类型
3. **性能**：通常使用平台相关的`int`类型性能最好
4. **精度**：对于需要高精度的计算，选择`float64`

```go
// 计算年龄，使用int8就足够了
var age int8 = 25

// 计算人口，需要更大的范围
var population int64 = 1400000000

// 计算金额，需要高精度
var amount float64 = 1234.56
```

### 3.2 布尔类型

布尔类型只有两个值：`true`和`false`。布尔类型通常用于条件判断和逻辑运算。

**布尔变量的声明**

```go
var isTrue bool = true
var isFalse bool = false

var isValid bool // 默认值为false
```

**布尔运算**

Go语言支持三种布尔运算符：

- 逻辑与（&&）：两个操作数都为true时结果为true
- 逻辑或（||）：两个操作数中有一个为true时结果为true
- 逻辑非（!）：操作数为true时结果为false，反之亦然

```go
a := true
b := false

fmt.Println(a && b) // 输出: false
fmt.Println(a || b) // 输出: true
fmt.Println(!a)     // 输出: false
```

**布尔类型的使用**

布尔类型主要用于条件判断：

```go
age := 18
isAdult := age >= 18

if isAdult {
    fmt.Println("你是成年人")
} else {
    fmt.Println("你还未成年")
}
```

### 3.3 字符串类型

字符串是Go语言中最常用的数据类型之一，用于存储文本数据。

> 💡 通俗理解：Go 的字符串像一个"只读的字节数组"，`len` 数的是字节（byte）个数，而一个汉字在 UTF-8 下通常占 3 个字节，所以 `len("你好")` 是 6 而不是 2。想按"字"来数或切，要改用 `rune`——这正是第 3.4 节字符类型要讲的事。

**字符串的声明**

```go
var s1 string = "Hello, World!"
s2 := "Go语言"
s3 := `这是一个
多行字符串`
```

注意：Go语言中字符串使用双引号（""）或反引号（``）来表示。双引号字符串支持转义字符，反引号字符串是原始字符串，不支持转义字符，但可以跨行。

**字符串的长度**

使用`len()`函数可以获取字符串的字节长度，使用`utf8.RuneCountInString()`函数可以获取字符串的字符数：

```go
import "unicode/utf8"

s := "你好，世界"
fmt.Println(len(s))                    // 输出: 15（字节长度）
fmt.Println(utf8.RuneCountInString(s)) // 输出: 5（字符数）
```

**字符串的索引和切片**

Go语言中字符串的索引和切片是基于字节的：

```go
s := "Hello, World!"
fmt.Println(s[0])       // 输出: 72（'H'的ASCII码）
fmt.Println(s[0:5])     // 输出: Hello
fmt.Println(s[7:])      // 输出: World!
fmt.Println(s[:5])      // 输出: Hello
```

> ⚠️ **常见坑**：按字节索引字符串时，`s[i]` 得到的是 `byte`（ASCII 码）而不是字符；对含中文的字符串做 `s[i]` 或 `s[i:j]`，一不小心就会切在半个汉字上，产生乱码。要按字符处理，请用 `for range` 遍历，或先转成 `[]rune`。

**字符串的拼接**

使用`+`运算符可以拼接字符串：

```go
s1 := "Hello"
s2 := "World"
s3 := s1 + ", " + s2 + "!"
fmt.Println(s3) // 输出: Hello, World!
```

对于大量字符串的拼接，建议使用`strings.Builder`，它更高效：

```go
import "strings"

var builder strings.Builder
builder.WriteString("Hello")
builder.WriteString(", ")
builder.WriteString("World")
builder.WriteString("!")
fmt.Println(builder.String()) // 输出: Hello, World!
```

**字符串的常用操作**

Go语言的`strings`包提供了丰富的字符串操作函数：

```go
import "strings"

s := "Hello, World!"

// 判断是否包含子串
fmt.Println(strings.Contains(s, "World")) // 输出: true

// 判断是否以某个前缀开头
fmt.Println(strings.HasPrefix(s, "Hello")) // 输出: true

// 判断是否以某个后缀结尾
fmt.Println(strings.HasSuffix(s, "!")) // 输出: true

// 查找子串的位置
fmt.Println(strings.Index(s, "World")) // 输出: 7

// 替换子串
fmt.Println(strings.Replace(s, "World", "Go", 1)) // 输出: Hello, Go!

// 转换为大写
fmt.Println(strings.ToUpper(s)) // 输出: HELLO, WORLD!

// 转换为小写
fmt.Println(strings.ToLower(s)) // 输出: hello, world!

// 去除首尾空格
s2 := "  Hello, World!  "
fmt.Println(strings.TrimSpace(s2)) // 输出: Hello, World!

// 分割字符串
parts := strings.Split(s, ",")
fmt.Println(parts) // 输出: [Hello  World!]

// 连接字符串
arr := []string{"Hello", "World"}
fmt.Println(strings.Join(arr, ", ")) // 输出: Hello, World
```

**字符串与数字的转换**

我们已经在第二章中介绍过字符串与数字的转换，这里再简单回顾一下：

```go
import "strconv"

// 字符串转整数
i, err := strconv.Atoi("123")
if err != nil {
    fmt.Println("转换失败")
}

// 整数转字符串
s := strconv.Itoa(456)

// 字符串转浮点数
f, err := strconv.ParseFloat("3.14", 64)
if err != nil {
    fmt.Println("转换失败")
}

// 浮点数转字符串
s = strconv.FormatFloat(3.14, 'f', 2, 64)
```

### 3.4 字符类型

Go语言中有两种字符类型：`byte`和`rune`。

**byte类型**

`byte`类型是`uint8`的别名，用于存储ASCII字符：

```go
var ch byte = 'A'
fmt.Println(ch) // 输出: 65
```

**rune类型**

`rune`类型是`int32`的别名，用于存储Unicode字符，可以表示任何Unicode字符：

```go
var ch rune = '中'
fmt.Println(ch) // 输出: 20013
```

**字符串的遍历**

遍历字符串时，如果使用`range`，会得到字符的索引和字符本身（rune类型）：

```go
s := "你好，世界"
for index, char := range s {
    fmt.Printf("索引: %d, 字符: %c, Unicode: %d\n", index, char, char)
}
```

输出：

```
索引: 0, 字符: 你, Unicode: 20320
索引: 3, 字符: 好, Unicode: 22909
索引: 6, 字符: ，, Unicode: 65292
索引: 9, 字符: 世, Unicode: 19990
索引: 12, 字符: 界, Unicode: 30028
```

### 3.5 类型别名

Go语言支持类型别名，类型别名和类型定义不同，类型别名只是给现有类型起了一个新名字，它们是完全相同的类型。

```go
type MyInt = int // 类型别名
type YourInt int // 类型定义

var a MyInt = 10
var b YourInt = 20

var c int = a // 可以，因为MyInt和int是相同的类型
// var d int = b // 不可以，因为YourInt和int是不同的类型
```

类型别名主要用于代码重构和兼容性处理。

### 3.6 实例：学生成绩管理系统

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个简单的学生成绩管理系统。

```go
package main

import (
    "fmt"
    "strconv"
    "strings"
)

type Student struct {
    name    string
    age     int
    score   float64
    isPass  bool
}

func main() {
    var studentCount int
    fmt.Print("请输入学生人数: ")
    fmt.Scanln(&studentCount)

    students := make([]Student, studentCount)

    for i := 0; i < studentCount; i++ {
        fmt.Printf("\n请输入第%d个学生的信息:\n", i+1)

        var name string
        fmt.Print("姓名: ")
        fmt.Scanln(&name)

        var age int
        fmt.Print("年龄: ")
        fmt.Scanln(&age)

        var scoreStr string
        fmt.Print("成绩: ")
        fmt.Scanln(&scoreStr)
        score, _ := strconv.ParseFloat(scoreStr, 64)

        isPass := score >= 60

        students[i] = Student{
            name:   name,
            age:    age,
            score:  score,
            isPass: isPass,
        }
    }

    fmt.Println("\n学生成绩汇总:")
    fmt.Println("姓名\t年龄\t成绩\t是否及格")
    fmt.Println(strings.Repeat("-", 40))

    var totalScore float64
    var passCount int

    _, student := range students {
        fmt.Printf("%s\t%d\t%.2f\t%v\n", student.name, student.age, student.score, student.isPass)
        totalScore += student.score
        if student.isPass {
            passCount++
        }
    }

    avgScore := totalScore / float64(studentCount)
    passRate := float64(passCount) / float64(studentCount) * 100

    fmt.Println(strings.Repeat("-", 40))
    fmt.Printf("平均成绩: %.2f\n", avgScore)
    fmt.Printf("及格率: %.2f%%\n", passRate)
}
```

这个程序：
1. 定义了一个`Student`结构体，包含姓名、年龄、成绩和是否及格等字段
2. 使用各种基本数据类型：string、int、float64、bool
3. 使用字符串操作函数：`strings.Repeat`
4. 进行数值计算和类型转换
5. 格式化输出结果

运行这个程序，输入3个学生的信息：

```
请输入学生人数: 3

请输入第1个学生的信息:
姓名: 张三
年龄: 18
成绩: 85

请输入第2个学生的信息:
姓名: 李四
年龄: 19
成绩: 72

请输入第3个学生的信息:
姓名: 王五
年龄: 18
成绩: 58

学生成绩汇总:
姓名    年龄    成绩    是否及格
----------------------------------------
张三    18      85.00   true
李四    19      72.00   true
王五    18      58.00   false
----------------------------------------
平均成绩: 71.67
及格率: 66.67%
```

### 3.7 小结

本章介绍了Go语言的基本数据类型，包括数值类型（整数、浮点数、复数）、布尔类型、字符串类型和字符类型。我们学习了每种类型的特点、使用方法和注意事项。数值类型用于存储数字，布尔类型用于条件判断，字符串类型用于存储文本，字符类型用于存储单个字符。我们还学习了字符串的常用操作和类型别名的概念。

下一章我们将学习Go语言的运算符，包括算术运算符、关系运算符、逻辑运算符、位运算符等，以及它们的使用方法和优先级。

**面试题**

1. `int` 和 `int32`、`uint` 有什么区别？该如何选择？
2. `len` 和 `cap` 的区别？字符串的 `len` 返回的到底是什么？
3. `byte` 和 `rune` 分别是什么的别名？它们如何影响字符串遍历？
4. `float32` 与 `float64` 的精度差异？为什么金额计算要小心浮点数？
5. 类型别名 `type A = int` 与类型定义 `type A int` 有什么不同？

## 第四章：运算符

> 本章总览：运算符按功能分六大类，优先级决定谁先算。

```mermaid
flowchart TD
    O[Go 运算符] --> A[算术运算符]
    O --> R[关系运算符]
    O --> L[逻辑运算符]
    O --> B[位运算符]
    O --> S[赋值与复合赋值]
    O --> M[其它 取地址 取长度 类型断言]
    A --> P[优先级 括号最先]
    R --> P
    L --> P
    B --> P
```

### 4.1 算术运算符

算术运算符用于执行基本的数学运算，包括加、减、乘、除、取余等。

**基本算术运算符**

| 运算符 | 描述 | 示例 |
|--------|------|------|
| + | 加法 | a + b |
| - | 减法 | a - b |
| * | 乘法 | a * b |
| / | 除法 | a / b |
| % | 取余 | a % b |

```go
a := 10
b := 3

fmt.Println(a + b)  // 输出: 13
fmt.Println(a - b)  // 输出: 7
fmt.Println(a * b)  // 输出: 30
fmt.Println(a / b)  // 输出: 3（整数除法）
fmt.Println(a % b)  // 输出: 1
```

**浮点数除法**

对于浮点数，除法运算会得到小数结果：

```go
a := 10.0
b := 3.0

fmt.Println(a / b)  // 输出: 3.3333333333333335
```

**自增和自减运算符**

Go语言支持自增（++）和自减（--）运算符，但它们只能作为语句使用，不能作为表达式：

```go
a := 10
a++  // 相当于 a = a + 1
fmt.Println(a)  // 输出: 11

a--  // 相当于 a = a - 1
fmt.Println(a)  // 输出: 10
```

注意：Go语言中的自增和自减运算符只能放在变量后面，不能放在变量前面，也不能作为表达式使用：

```go
a := 10
// b := a++  // 编译错误：不能作为表达式使用
// ++a      // 编译错误：只能放在变量后面
```

**字符串拼接**

`+`运算符也可以用于字符串拼接：

```go
s1 := "Hello"
s2 := "World"
s3 := s1 + ", " + s2 + "!"
fmt.Println(s3)  // 输出: Hello, World!
```

### 4.2 关系运算符

关系运算符用于比较两个值，返回布尔类型的结果。

**关系运算符列表**

| 运算符 | 描述 | 示例 |
|--------|------|------|
| == | 等于 | a == b |
| != | 不等于 | a != b |
| > | 大于 | a > b |
| < | 小于 | a < b |
| >= | 大于等于 | a >= b |
| <= | 小于等于 | a <= b |

```go
a := 10
b := 20

fmt.Println(a == b)  // 输出: false
fmt.Println(a != b)  // 输出: true
fmt.Println(a > b)   // 输出: false
fmt.Println(a < b)   // 输出: true
fmt.Println(a >= b)  // 输出: false
fmt.Println(a <= b)  // 输出: true
```

**字符串比较**

字符串也可以使用关系运算符进行比较，比较的是字符串的字典序：

```go
s1 := "apple"
s2 := "banana"

fmt.Println(s1 == s2)  // 输出: false
fmt.Println(s1 < s2)   // 输出: true（'a' < 'b'）
```

**浮点数比较**

由于浮点数的精度问题，直接比较浮点数是否相等可能会得到意外的结果：

```go
a := 0.1 + 0.2
b := 0.3

fmt.Println(a == b)  // 输出: false（由于精度问题）
```

比较浮点数时，应该判断它们的差值是否在一个很小的范围内：

```go
a := 0.1 + 0.2
b := 0.3
epsilon := 0.0001

fmt.Println(math.Abs(a-b) < epsilon)  // 输出: true
```

> ⚠️ **常见坑**：浮点数不能用 `==` 直接比较。`0.1 + 0.2 == 0.3` 在 Go 里会得到 `false`，因为二进制无法精确表示这些小数。正确做法是判断两者差值是否小于一个极小的阈值 `epsilon`，例如 `math.Abs(a-b) < 1e-9`。

### 4.3 逻辑运算符

逻辑运算符用于组合多个条件，返回布尔类型的结果。

**逻辑运算符列表**

| 运算符 | 描述 | 示例 |
|--------|------|------|
| && | 逻辑与 | a && b |
| \|\| | 逻辑或 | a \|\| b |
| ! | 逻辑非 | !a |

```go
a := true
b := false

fmt.Println(a && b)  // 输出: false（两个都为true才返回true）
fmt.Println(a || b)  // 输出: true（有一个为true就返回true）
fmt.Println(!a)      // 输出: false（取反）
```

**短路求值**

逻辑与和逻辑或运算符支持短路求值，即如果第一个操作数就能确定结果，就不会计算第二个操作数：

```go
a := false
b := true

result := a && (10/0 > 0)  // 不会执行10/0，因为a为false，结果已经确定为false
fmt.Println(result)        // 输出: false

a = true
result = a || (10/0 > 0)   // 不会执行10/0，因为a为true，结果已经确定为true
fmt.Println(result)        // 输出: true
```

短路求值可以提高程序的性能，也可以避免一些错误的发生。

### 4.4 位运算符

位运算符用于对整数的二进制位进行操作。

> 💡 通俗理解：位运算就是直接摆弄数字的"二进制开关"。`&` 是"两个开关都开才开"，`|` 是"有一个开就开"，`^` 是"两个不一样才开"，`<< n` 相当于乘以 2 的 n 次方，`>> n` 相当于除以 2 的 n 次方。它比乘除更快，常用于权限位、标志位和哈希计算。

**位运算符列表**

| 运算符 | 描述 | 示例 |
|--------|------|------|
| & | 按位与 | a & b |
| \| | 按位或 | a \| b |
| ^ | 按位异或 | a ^ b |
| << | 左移 | a << b |
| >> | 右移 | a >> b |
| &^ | 位清除 | a &^ b |

**按位与（&）**

按位与运算符对两个数的每一位进行与运算，只有两个对应的位都为1时，结果的位才为1：

```go
a := 5   // 二进制: 101
b := 3   // 二进制: 011

fmt.Println(a & b)  // 输出: 1（二进制: 001）
```

**按位或（|）**

按位或运算符对两个数的每一位进行或运算，只要两个对应的位有一个为1，结果的位就为1：

```go
a := 5   // 二进制: 101
b := 3   // 二进制: 011

fmt.Println(a | b)  // 输出: 7（二进制: 111）
```

**按位异或（^）**

按位异或运算符对两个数的每一位进行异或运算，当两个对应的位不同时，结果的位为1：

```go
a := 5   // 二进制: 101
b := 3   // 二进制: 011

fmt.Println(a ^ b)  // 输出: 6（二进制: 110）
```

**左移（<<）**

左移运算符将数的二进制位向左移动指定的位数，右边补0：

```go
a := 5   // 二进制: 101

fmt.Println(a << 1)  // 输出: 10（二进制: 1010）
fmt.Println(a << 2)  // 输出: 20（二进制: 10100）
```

左移一位相当于乘以2，左移n位相当于乘以2的n次方。

**右移（>>）**

右移运算符将数的二进制位向右移动指定的位数，左边补符号位：

```go
a := 20  // 二进制: 10100

fmt.Println(a >> 1)  // 输出: 10（二进制: 1010）
fmt.Println(a >> 2)  // 输出: 5（二进制: 101）
```

右移一位相当于除以2，右移n位相当于除以2的n次方。

**位清除（&^）**

位清除运算符将第一个操作数中与第二个操作数为1的位对应的位清零：

```go
a := 5   // 二进制: 101
b := 3   // 二进制: 011

fmt.Println(a &^ b)  // 输出: 4（二进制: 100）
```

### 4.5 赋值运算符

赋值运算符用于给变量赋值。

**基本赋值运算符**

```go
a := 10
a = 20  // 基本赋值
```

**复合赋值运算符**

Go语言支持复合赋值运算符，将运算和赋值合并为一个操作：

| 运算符 | 描述 | 等价于 |
|--------|------|--------|
| += | 加法赋值 | a = a + b |
| -= | 减法赋值 | a = a - b |
| *= | 乘法赋值 | a = a * b |
| /= | 除法赋值 | a = a / b |
| %= | 取余赋值 | a = a % b |
| &= | 按位与赋值 | a = a & b |
| \|= | 按位或赋值 | a = a \| b |
| ^= | 按位异或赋值 | a = a ^ b |
| <<= | 左移赋值 | a = a << b |
| >>= | 右移赋值 | a = a >> b |
| &^= | 位清除赋值 | a = a &^ b |

```go
a := 10
a += 5   // 相当于 a = a + 5
fmt.Println(a)  // 输出: 15

a -= 3   // 相当于 a = a - 3
fmt.Println(a)  // 输出: 12

a *= 2   // 相当于 a = a * 2
fmt.Println(a)  // 输出: 24

a /= 4   // 相当于 a = a / 4
fmt.Println(a)  // 输出: 6

a %= 4   // 相当于 a = a % 4
fmt.Println(a)  // 输出: 2
```

### 4.6 其他运算符

**地址运算符（&）**

地址运算符用于获取变量的内存地址：

```go
a := 10
fmt.Println(&a)  // 输出变量的内存地址，例如: 0xc0000a2008
```

**指针运算符（*）**

指针运算符用于获取指针指向的值：

```go
a := 10
p := &a  // p是指向a的指针
fmt.Println(*p)  // 输出: 10（获取p指向的值）
```

**取长度运算符**

`len()`函数用于获取字符串、数组、切片、映射等的长度：

```go
s := "Hello"
fmt.Println(len(s))  // 输出: 5

arr := [3]int{1, 2, 3}
fmt.Println(len(arr))  // 输出: 3
```

**取容量运算符**

`cap()`函数用于获取切片、通道等的容量：

```go
slice := make([]int, 3, 5)
fmt.Println(len(slice))  // 输出: 3（长度）
fmt.Println(cap(slice))  // 输出: 5（容量）
```

**类型断言运算符**

类型断言运算符用于将接口类型转换为具体类型：

```go
var i interface{} = "Hello"
s, ok := i.(string)
if ok {
    fmt.Println(s)  // 输出: Hello
}
```

### 4.7 运算符优先级

运算符优先级决定了表达式中运算的执行顺序。Go语言中运算符的优先级从高到低如下：

| 优先级 | 运算符 |
|--------|--------|
| 1 | * / % << >> & &^ |
| 2 | + - \| ^ |
| 3 | == != < <= > >= |
| 4 | && |
| 5 | \|\| |

当表达式中有多个运算符时，优先级高的先执行。如果优先级相同，则从左到右执行。

```go
a := 10 + 5 * 2  // 先执行乘法，再执行加法
fmt.Println(a)    // 输出: 20

a = (10 + 5) * 2  // 使用括号改变优先级
fmt.Println(a)    // 输出: 30
```

建议在复杂的表达式中使用括号来明确运算顺序，这样代码更易读。

### 4.8 实例：计算器程序

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个简单的计算器程序。

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    fmt.Println("简单计算器")
    fmt.Println("支持的操作: +, -, *, /")

    var num1Str, num2Str, op string

    fmt.Print("请输入第一个数字: ")
    fmt.Scanln(&num1Str)

    fmt.Print("请输入操作符: ")
    fmt.Scanln(&op)

    fmt.Print("请输入第二个数字: ")
    fmt.Scanln(&num2Str)

    num1, err1 := strconv.ParseFloat(num1Str, 64)
    num2, err2 := strconv.ParseFloat(num2Str, 64)

    if err1 != nil || err2 != nil {
        fmt.Println("输入的数字无效")
        return
    }

    var result float64
    var validOp bool

    switch op {
    case "+":
        result = num1 + num2
        validOp = true
    case "-":
        result = num1 - num2
        validOp = true
    case "*":
        result = num1 * num2
        validOp = true
    case "/":
        if num2 == 0 {
            fmt.Println("错误：除数不能为零")
            return
        }
        result = num1 / num2
        validOp = true
    default:
        validOp = false
    }

    if validOp {
        fmt.Printf("结果: %.2f %s %.2f = %.2f\n", num1, op, num2, result)
    } else {
        fmt.Println("错误：不支持的操作符")
    }
}
```

这个程序：
1. 使用字符串输入和数字转换
2. 使用switch语句处理不同的操作符
3. 使用算术运算符进行计算
4. 使用关系运算符进行条件判断
5. 格式化输出结果

运行这个程序：

```
简单计算器
支持的操作: +, -, *, /
请输入第一个数字: 10
请输入操作符: *
请输入第二个数字: 5
结果: 10.00 * 5.00 = 50.00
```

### 4.9 小结

本章介绍了Go语言的各种运算符，包括算术运算符、关系运算符、逻辑运算符、位运算符、赋值运算符等。我们学习了每种运算符的使用方法和注意事项。运算符是编程语言的基础，掌握好运算符的使用对于编写高效的程序非常重要。我们还学习了运算符的优先级，了解了如何在表达式中使用括号来明确运算顺序。

下一章我们将学习Go语言的控制流，包括条件语句、循环语句、跳转语句等，以及它们的使用方法和注意事项。

**面试题**

1. Go 的 `++`/`--` 为什么不能作为表达式使用？和 C 语言有何不同？
2. `&&` 和 `||` 的短路求值是什么？有什么实际用途？
3. `&^`（位清除）的运算规则是什么？举一个使用场景。
4. 为什么浮点数用 `==` 比较不可靠？应该怎么比较？
5. Go 运算符优先级分几级？如何避免优先级带来的歧义？

## 第五章：控制流

> 本章总览：`defer` 像出门前的待办清单，函数返回时倒序执行。

```mermaid
flowchart TD
    F[进入函数] --> D1[第一个 defer 入栈]
    D1 --> D2[第二个 defer 入栈]
    D2 --> W[执行函数主体]
    W --> P[正常 return 或 panic]
    P --> E2[后进先出 先执行第二个 defer]
    E2 --> E1[再执行第一个 defer]
    E1 --> END[函数真正结束]
```

### 5.1 if语句

if语句是Go语言中最基本的条件语句，用于根据条件执行不同的代码块。

**基本if语句**

```go
age := 18
if age >= 18 {
    fmt.Println("你是成年人")
}
```

**if-else语句**

```go
age := 16
if age >= 18 {
    fmt.Println("你是成年人")
} else {
    fmt.Println("你还未成年")
}
```

**if-else if-else语句**

```go
score := 85
if score >= 90 {
    fmt.Println("优秀")
} else if score >= 80 {
    fmt.Println("良好")
} else if score >= 60 {
    fmt.Println("及格")
} else {
    fmt.Println("不及格")
}
```

**if语句的初始化**

Go语言的if语句支持在条件前初始化变量，这个变量的作用域仅限于if语句块：

```go
if num := 10; num > 5 {
    fmt.Println("num大于5")
}
// 这里不能访问num变量
```

这个特性在处理错误时非常有用：

```go
if err := someFunction(); err != nil {
    fmt.Println("发生错误:", err)
    return
}
```

**if语句的注意事项**

1. Go语言的if语句不需要括号，但推荐使用大括号
2. 左大括号必须与if在同一行
3. 条件表达式必须是布尔类型，不能是其他类型

```go
// 不推荐
if (age >= 18) {
    fmt.Println("你是成年人")
}

// 推荐
if age >= 18 {
    fmt.Println("你是成年人")
}

// 编译错误：条件必须是布尔类型
if num := 10 {
    fmt.Println("num是10")
}
```

### 5.2 switch语句

switch语句是一种多分支选择结构，比多个if-else语句更清晰。

**基本switch语句**

```go
day := 3
switch day {
case 1:
    fmt.Println("星期一")
case 2:
    fmt.Println("星期二")
case 3:
    fmt.Println("星期三")
default:
    fmt.Println("其他")
}
```

**switch语句的注意事项**

1. Go语言的switch语句不需要break，每个case执行完后会自动跳出
2. 如果需要继续执行下一个case，可以使用fallthrough关键字
3. case后面的表达式可以是任意类型，不限于整数

```go
day := 1
switch day {
case 1:
    fmt.Println("星期一")
    fallthrough
case 2:
    fmt.Println("工作日")
default:
    fmt.Println("其他")
}
// 输出：
// 星期一
// 工作日
```

**switch语句的初始化**

与if语句一样，switch语句也支持初始化：

```go
switch num := 10; num {
case 10:
    fmt.Println("num是10")
case 20:
    fmt.Println("num是20")
default:
    fmt.Println("num是其他值")
}
```

**无表达式的switch语句**

switch语句可以没有表达式，这时相当于多个if-else语句：

```go
score := 85
switch {
case score >= 90:
    fmt.Println("优秀")
case score >= 80:
    fmt.Println("良好")
case score >= 60:
    fmt.Println("及格")
default:
    fmt.Println("不及格")
}
```

**多值case**

一个case可以匹配多个值：

```go
day := 6
switch day {
case 1, 2, 3, 4, 5:
    fmt.Println("工作日")
case 6, 7:
    fmt.Println("周末")
default:
    fmt.Println("无效的日期")
}
```

**类型switch**

switch语句还可以用于类型判断：

```go
var i interface{} = "Hello"

switch v := i.(type) {
case string:
    fmt.Println("字符串:", v)
case int:
    fmt.Println("整数:", v)
case bool:
    fmt.Println("布尔值:", v)
default:
    fmt.Println("未知类型")
}
```

### 5.3 for循环

for循环是Go语言中唯一的循环语句，它可以实现各种循环模式。

**基本for循环**

```go
for i := 0; i < 5; i++ {
    fmt.Println(i)
}
```

**while循环**

Go语言没有while关键字，但可以使用for循环实现while循环：

```go
i := 0
for i < 5 {
    fmt.Println(i)
    i++
}
```

**无限循环**

```go
for {
    fmt.Println("无限循环")
    // 需要使用break跳出循环
}
```

**for-range循环**

for-range循环用于遍历数组、切片、映射、字符串等：

```go
// 遍历切片
numbers := []int{1, 2, 3, 4, 5}
for index, value := range numbers {
    fmt.Printf("索引: %d, 值: %d\n", index, value)
}

// 遍历映射
ages := map[string]int{
    "张三": 18,
    "李四": 20,
}
for name, age := range ages {
    fmt.Printf("姓名: %s, 年龄: %d\n", name, age)
}

// 遍历字符串
s := "Hello"
for index, char := range s {
    fmt.Printf("索引: %d, 字符: %c\n", index, char)
}
```

**只使用索引或值**

如果只需要索引或值，可以使用下划线忽略不需要的值：

```go
numbers := []int{1, 2, 3, 4, 5}

// 只需要值
for _, value := range numbers {
    fmt.Println(value)
}

// 只需要索引
for index := range numbers {
    fmt.Println(index)
}
```

**for循环的注意事项**

1. for循环的三个部分都是可选的
2. 左大括号必须与for在同一行
3. 循环变量的作用域仅限于循环体

```go
// 正确
for i := 0; i < 5; i++ {
    fmt.Println(i)
}

// 错误：左大括号不能另起一行
for i := 0; i < 5; i++
{
    fmt.Println(i)
}
```

### 5.4 break和continue语句

break和continue语句用于控制循环的执行流程。

**break语句**

break语句用于跳出循环：

```go
for i := 0; i < 10; i++ {
    if i == 5 {
        break
    }
    fmt.Println(i)
}
// 输出: 0 1 2 3 4
```

**continue语句**

continue语句用于跳过当前循环的剩余部分，进入下一次循环：

```go
for i := 0; i < 10; i++ {
    if i%2 == 0 {
        continue
    }
    fmt.Println(i)
}
// 输出: 1 3 5 7 9
```

**标签跳转**

Go语言支持标签跳转，可以跳出多层循环：

```go
outer:
for i := 0; i < 3; i++ {
    for j := 0; j < 3; j++ {
        if i == 1 && j == 1 {
            break outer
        }
        fmt.Printf("i=%d, j=%d\n", i, j)
    }
}
// 输出:
// i=0, j=0
// i=0, j=1
// i=0, j=2
// i=1, j=0
```

### 5.5 goto语句

goto语句可以无条件跳转到标签处，但不推荐使用，因为它会让代码变得难以理解。

```go
i := 0
start:
if i < 5 {
    fmt.Println(i)
    i++
    goto start
}
// 输出: 0 1 2 3 4
```

**goto语句的注意事项**

1. goto语句只能在同一个函数内跳转
2. 不能跳过变量声明
3. 不能跳入到另一个代码块

```go
// 错误：跳过了变量声明
goto skip
var x int
skip:
fmt.Println(x)

// 错误：跳入到for循环
goto inside
for i := 0; i < 10; i++ {
inside:
    fmt.Println(i)
}
```

### 5.6 defer语句

defer语句用于延迟函数的执行，直到包含它的函数返回。

> 💡 通俗理解：`defer` 像"出门前的待办清单"——你先把要做的事贴到墙上，等真正走到门口（函数返回）时，它会**倒着**一件件执行。所以"打开文件后立刻 `defer file.Close()`"是最保险的写法，无论后面从哪个分支返回，关门这件事都不会被忘掉。

**基本defer语句**

```go
func main() {
    defer fmt.Println("最后执行")
    fmt.Println("首先执行")
}
// 输出:
// 首先执行
// 最后执行
```

**多个defer语句**

多个defer语句按照后进先出（LIFO）的顺序执行：

```go
func main() {
    defer fmt.Println("1")
    defer fmt.Println("2")
    defer fmt.Println("3")
}
// 输出:
// 3
// 2
// 1
```

**defer语句的参数**

defer语句的参数在defer时就已经确定：

```go
func main() {
    i := 0
    defer fmt.Println(i)
    i = 10
}
// 输出: 0
```

> ⚠️ **常见坑**：`defer` 的参数（以及接收者）在 **defer 语句执行的那一刻** 就被求值并保存，而不是等函数返回时才取。所以 `defer fmt.Println(i)` 打印的是当时的 `i`；想打印最终值要用闭包 `defer func(){ fmt.Println(i) }()`。另外多个 defer 是 LIFO（后进先出）顺序。

**defer语句的常见用途**

1. 关闭文件
2. 关闭数据库连接
3. 释放锁
4. 恢复panic

```go
// 关闭文件
func readFile(filename string) error {
    file, err := os.Open(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    // 读取文件内容
    return nil
}

// 恢复panic
func safeFunction() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("恢复panic:", r)
        }
    }()

    panic("发生错误")
}
```

### 5.7 实例：猜数字游戏

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个猜数字游戏。

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

func main() {
    rand.Seed(time.Now().UnixNano())
    target := rand.Intn(100) + 1
    var guess int
    var attempts int

    fmt.Println("猜数字游戏")
    fmt.Println("我已经想好了一个1到100之间的数字，请你猜猜看")

    for {
        attempts++
        fmt.Printf("第%d次猜测，请输入你的猜测: ", attempts)
        fmt.Scanln(&guess)

        if guess < 1 || guess > 100 {
            fmt.Println("请输入1到100之间的数字")
            continue
        }

        if guess == target {
            fmt.Printf("恭喜你！你猜对了！数字是%d，你用了%d次猜测\n", target, attempts)
            break
        } else if guess < target {
            fmt.Println("太小了，再试一次")
        } else {
            fmt.Println("太大了，再试一次")
        }

        if attempts >= 10 {
            fmt.Printf("很遗憾，你已经猜了10次了。正确答案是%d\n", target)
            break
        }
    }
}
```

这个程序：
1. 使用for循环实现游戏的主循环
2. 使用if-else语句判断猜测结果
3. 使用continue语句跳过无效输入
4. 使用break语句退出游戏
5. 使用随机数生成目标数字

运行这个程序：

```
猜数字游戏
我已经想好了一个1到100之间的数字，请你猜猜看
第1次猜测，请输入你的猜测: 50
太大了，再试一次
第2次猜测，请输入你的猜测: 25
太小了，再试一次
第3次猜测，请输入你的猜测: 37
太小了，再试一次
第4次猜测，请输入你的猜测: 43
太小了，再试一次
第5次猜测，请输入你的猜测: 46
恭喜你！你猜对了！数字是46，你用了5次猜测
```

### 5.8 小结

本章介绍了Go语言的控制流，包括if语句、switch语句、for循环、break和continue语句、goto语句和defer语句。我们学习了每种控制流的使用方法和注意事项。控制流是编程语言的基础，掌握好控制流的使用对于编写逻辑清晰的程序非常重要。我们还学习了defer语句的延迟执行机制，了解了它在资源管理中的重要作用。

下一章我们将学习Go语言的函数，包括函数的定义、参数传递、返回值、匿名函数、闭包等，以及它们的使用方法和注意事项。

**面试题**

1. Go 的 `for` 有哪几种写法？如何模拟 while 和"无限循环"？
2. `switch` 为什么不需要 `break`？`fallthrough` 的作用是什么？
3. 无表达式的 `switch` 和类型 `switch` 分别怎么用？
4. `defer` 的执行顺序与参数求值时机是怎样的？
5. `break`、`continue`、`goto` 与标签跳转的区别和使用禁忌？

## 第六章：函数

> 本章总览：一次函数调用，从实参复制到返回值写入的完整时序。

```mermaid
sequenceDiagram
    participant M as main 调用者
    participant F as 被调用函数
    participant D as defer 栈
    M->>F: 复制实参并传入
    F->>F: 执行函数体
    F->>D: 注册 defer
    F->>M: 写入返回值
    D-->>F: 函数返回前倒序执行
    M->>M: 接收返回值继续运行
```

### 6.1 函数的定义

函数是Go语言中的一等公民，是代码组织和重用的基本单位。

**基本函数定义**

```go
func add(a int, b int) int {
    return a + b
}
```

函数定义包含以下几个部分：
- `func`关键字：声明这是一个函数
- 函数名：`add`
- 参数列表：`(a int, b int)`，指定函数的参数和类型
- 返回值类型：`int`，指定函数返回值的类型
- 函数体：`{ return a + b }`，函数的具体实现

**调用函数**

```go
result := add(3, 5)
fmt.Println(result)  // 输出: 8
```

**多个参数相同类型**

如果多个参数的类型相同，可以简写：

```go
func add(a, b int) int {
    return a + b
}
```

**多个返回值**

Go语言支持函数返回多个值：

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("除数不能为零")
    }
    return a / b, nil
}
```

调用多返回值函数：

```go
result, err := divide(10, 2)
if err != nil {
    fmt.Println("发生错误:", err)
} else {
    fmt.Println("结果:", result)
}
```

**命名返回值**

可以为返回值命名，这样在函数体内可以直接使用：

```go
func calculate(a, b int) (sum, difference int) {
    sum = a + b
    difference = a - b
    return  // 相当于 return sum, difference
}
```

**可变参数函数**

Go语言支持可变参数函数，使用`...`表示：

```go
func sum(numbers ...int) int {
    total := 0
    for _, num := range numbers {
        total += num
    }
    return total
}
```

调用可变参数函数：

```go
result := sum(1, 2, 3, 4, 5)
fmt.Println(result)  // 输出: 15

numbers := []int{1, 2, 3, 4, 5}
result = sum(numbers...)
fmt.Println(result)  // 输出: 15
```

### 6.2 参数传递

Go语言中的参数传递有两种方式：值传递和引用传递。

> 💡 通俗理解：严格说 Go 只有**值传递**——传参时总是把实参复制一份给函数。所谓"引用传递"，其实是复制了"门牌号"（指针、切片头、映射引用），函数按门牌号找到同一间房子去改，所以原值看起来变了。理解这一点，就能明白为什么改 `slice` 元素会影响原切片、而给形参重新赋值 `s = []int{...}` 却不会。

**值传递**

值传递是将参数的值复制一份传递给函数，函数内部对参数的修改不会影响原始值：

```go
func modifyValue(x int) {
    x = 10
}

func main() {
    a := 5
    modifyValue(a)
    fmt.Println(a)  // 输出: 5（a的值没有改变）
}
```

**引用传递**

引用传递是将参数的地址传递给函数，函数内部对参数的修改会影响原始值：

```go
func modifyPointer(x *int) {
    *x = 10
}

func main() {
    a := 5
    modifyPointer(&a)
    fmt.Println(a)  // 输出: 10（a的值被修改了）
}
```

**切片和映射的传递**

切片和映射在Go语言中是引用类型，传递给函数时传递的是引用：

```go
func modifySlice(s []int) {
    s[0] = 100
}

func main() {
    numbers := []int{1, 2, 3}
    modifySlice(numbers)
    fmt.Println(numbers)  // 输出: [100 2 3]
}
```

但是，如果在函数内部对切片进行重新赋值，不会影响原始切片：

```go
func reassignSlice(s []int) {
    s = []int{4, 5, 6}
}

func main() {
    numbers := []int{1, 2, 3}
    reassignSlice(numbers)
    fmt.Println(numbers)  // 输出: [1 2 3]
}
```

> ⚠️ **常见坑**：把切片传给函数后，函数内 `s[0] = x` 能改到原底层数组，但 `s = append(s, y)` 或 `s = []int{...}` 只改了形参这个"切片头"，调用方看不到变化。若想让追加结果对调用方可见，要么把新切片返回出去，要么传 `*[]int`。

### 6.3 函数作为值

在Go语言中，函数是一等公民，可以像其他值一样被传递和赋值。

**函数类型**

函数类型由参数类型和返回值类型决定：

```go
type Operation func(int, int) int
```

**函数作为参数**

```go
func applyOperation(a, b int, op func(int, int) int) int {
    return op(a, b)
}

func add(a, b int) int {
    return a + b
}

func multiply(a, b int) int {
    return a * b
}

func main() {
    result1 := applyOperation(3, 5, add)
    fmt.Println(result1)  // 输出: 8

    result2 := applyOperation(3, 5, multiply)
    fmt.Println(result2)  // 输出: 15
}
```

**函数作为返回值**

```go
func getOperation(op string) func(int, int) int {
    switch op {
    case "add":
        return func(a, b int) int {
            return a + b
        }
    case "multiply":
        return func(a, b int) int {
            return a * b
        }
    default:
        return func(a, b int) int {
            return 0
        }
    }
}

func main() {
    addOp := getOperation("add")
    result := addOp(3, 5)
    fmt.Println(result)  // 输出: 8
}
```

### 6.4 匿名函数

匿名函数是没有名字的函数，可以在函数内部定义和使用。

**基本匿名函数**

```go
func main() {
    add := func(a, b int) int {
        return a + b
    }
    result := add(3, 5)
    fmt.Println(result)  // 输出: 8
}
```

**立即执行匿名函数**

```go
func main() {
    result := func(a, b int) int {
        return a + b
    }(3, 5)
    fmt.Println(result)  // 输出: 8
}
```

**匿名函数作为参数**

```go
func applyOperation(a, b int, op func(int, int) int) int {
    return op(a, b)
}

func main() {
    result := applyOperation(3, 5, func(a, b int) int {
        return a + b
    })
    fmt.Println(result)  // 输出: 8
}
```

### 6.5 闭包

闭包是指一个函数值，它引用了函数体之外的变量。闭包可以访问和修改外部变量。

**基本闭包**

```go
func adder() func(int) int {
    sum := 0
    return func(x int) int {
        sum += x
        return sum
    }
}

func main() {
    add := adder()
    fmt.Println(add(1))  // 输出: 1
    fmt.Println(add(2))  // 输出: 3
    fmt.Println(add(3))  // 输出: 6
}
```

**闭包的陷阱**

由于闭包会捕获外部变量，所以在循环中使用闭包时需要特别注意：

```go
func main() {
    funcs := []func(){}
    for i := 0; i < 3; i++ {
        funcs = append(funcs, func() {
            fmt.Println(i)
        })
    }
    for _, f := range funcs {
        f()
    }
    // 输出: 3 3 3（所有闭包都引用了同一个i）
}
```

解决方法是在循环内部创建一个新变量：

```go
func main() {
    funcs := []func(){}
    for i := 0; i < 3; i++ {
        i := i  // 创建新变量
        funcs = append(funcs, func() {
            fmt.Println(i)
        })
    }
    for _, f := range funcs {
        f()
    }
    // 输出: 0 1 2
}
```

### 6.6 defer与函数

defer语句与函数结合使用时，有一些特殊的行为需要注意。

**defer与返回值**

```go
func f1() (result int) {
    defer func() {
        result++
    }()
    return 0
}

func main() {
    fmt.Println(f1())  // 输出: 1
}
```

**defer与参数**

```go
func f2() {
    i := 0
    defer fmt.Println(i)
    i = 10
}

func main() {
    f2()  // 输出: 0（defer的参数在defer时就已经确定）
}
```

### 6.7 实例：函数式编程工具

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一些函数式编程工具。

```go
package main

import (
    "fmt"
)

func main() {
    numbers := []int{1, 2, 3, 4, 5}

    // Map：将函数应用到切片的每个元素
    doubled := Map(numbers, func(x int) int {
        return x * 2
    })
    fmt.Println("翻倍:", doubled)

    // Filter：过滤切片中的元素
    evens := Filter(numbers, func(x int) bool {
        return x%2 == 0
    })
    fmt.Println("偶数:", evens)

    // Reduce：将切片的元素归约为一个值
    sum := Reduce(numbers, 0, func(acc, x int) int {
        return acc + x
    })
    fmt.Println("求和:", sum)

    // Chain：链式调用
    result := Chain(
        numbers,
        FilterFunc(func(x int) bool { return x > 2 }),
        MapFunc(func(x int) int { return x * x }),
    )
    fmt.Println("链式调用:", result)
}

func Map(numbers []int, f func(int) int) []int {
    result := make([]int, len(numbers))
    for i, num := range numbers {
        result[i] = f(num)
    }
    return result
}

func Filter(numbers []int, f func(int) bool) []int {
    result := []int{}
    for _, num := range numbers {
        if f(num) {
            result = append(result, num)
        }
    }
    return result
}

func Reduce(numbers []int, initial int, f func(int, int) int) int {
    result := initial
    for _, num := range numbers {
        result = f(result, num)
    }
    return result
}

func Chain(numbers []int, operations ...func([]int) []int) []int {
    result := numbers
    for _, op := range operations {
        result = op(result)
    }
    return result
}

func MapFunc(f func(int) int) func([]int) []int {
    return func(numbers []int) []int {
        return Map(numbers, f)
    }
}

func FilterFunc(f func(int) bool) func([]int) []int {
    return func(numbers []int) []int {
        return Filter(numbers, f)
    }
}
```

这个程序：
1. 定义了Map、Filter、Reduce等函数式编程工具
2. 使用函数作为参数和返回值
3. 使用匿名函数和闭包
4. 实现了链式调用

运行这个程序：

```
翻倍: [2 4 6 8 10]
偶数: [2 4]
求和: 15
链式调用: [9 16 25]
```

### 6.8 小结

本章介绍了Go语言的函数，包括函数的定义、参数传递、函数作为值、匿名函数、闭包等。我们学习了每种函数特性的使用方法和注意事项。函数是Go语言中的一等公民，掌握好函数的使用对于编写模块化、可重用的代码非常重要。我们还学习了闭包的概念和陷阱，了解了如何在循环中正确使用闭包。

下一章我们将学习Go语言的数组、切片和映射，包括它们的定义、操作、遍历等，以及它们的使用方法和注意事项。

**面试题**

1. Go 的函数为什么被称为"一等公民"？函数类型如何声明？
2. 值传递与"引用传递"的区别？切片、映射传参到底传了什么？
3. 什么是闭包？闭包为什么会捕获并持有外部变量？
4. 在循环里创建闭包会遇到什么陷阱？如何解决？
5. `defer` 为什么能修改命名返回值？请举例说明。

## 第七章：数组、切片和映射

> 本章总览：切片是"窗口"，`append` 时容量够就原地写，不够才换新数组。

```mermaid
flowchart TD
    S[切片 s 长度与容量] --> A[底层数组]
    A --> AP[append 追加元素]
    AP --> C{容量是否够用}
    C -->|够用| R[原位写入 长度加一]
    C -->|不够| G[申请更大的底层数组]
    G --> CP[复制旧元素再写入]
    CP --> NP[新旧切片底层数组不再相同]
    R --> N[得到更新后的切片头]
    NP --> N
```

### 7.1 数组

数组是具有相同数据类型的固定长度的序列。数组在Go语言中是值类型，这意味着数组赋值时会复制整个数组。

**数组的声明**

```go
var arr1 [5]int  // 声明一个长度为5的整数数组
arr2 := [3]string{"a", "b", "c"}  // 声明并初始化
arr3 := [...]int{1, 2, 3, 4, 5}  // 让编译器推断长度
```

**数组的访问和修改**

```go
arr := [5]int{1, 2, 3, 4, 5}
fmt.Println(arr[0])  // 输出: 1
arr[0] = 10
fmt.Println(arr[0])  // 输出: 10
```

**数组的长度**

```go
arr := [5]int{1, 2, 3, 4, 5}
fmt.Println(len(arr))  // 输出: 5
```

**数组的遍历**

```go
arr := [5]int{1, 2, 3, 4, 5}

// 使用索引遍历
for i := 0; i < len(arr); i++ {
    fmt.Println(arr[i])
}

// 使用range遍历
for index, value := range arr {
    fmt.Printf("索引: %d, 值: %d\n", index, value)
}
```

**数组的比较**

相同类型的数组可以使用`==`和`!=`进行比较：

```go
arr1 := [3]int{1, 2, 3}
arr2 := [3]int{1, 2, 3}
arr3 := [3]int{1, 2, 4}

fmt.Println(arr1 == arr2)  // 输出: true
fmt.Println(arr1 == arr3)  // 输出: false
```

**数组的传递**

由于数组是值类型，传递给函数时会复制整个数组：

```go
func modifyArray(arr [3]int) {
    arr[0] = 100
}

func main() {
    arr := [3]int{1, 2, 3}
    modifyArray(arr)
    fmt.Println(arr)  // 输出: [1 2 3]（原数组没有被修改）
}
```

如果需要修改原数组，应该传递数组的指针：

```go
func modifyArray(arr *[3]int) {
    arr[0] = 100
}

func main() {
    arr := [3]int{1, 2, 3}
    modifyArray(&arr)
    fmt.Println(arr)  // 输出: [100 2 3]
}
```

### 7.2 切片

切片是对数组的抽象，提供了更灵活、更强大的序列操作。切片是引用类型，传递给函数时不会复制整个序列。

> 💡 通俗理解：切片不是"能自动变长的数组"，而是一个"窗口"（切片头），里面记着三样东西：指向底层数组的指针、当前长度 `len`、还能往后伸多远的容量 `cap`。`append` 时如果 `cap` 够用，就在原数组上往后写；不够用才换一块更大的数组并搬过去。因此两个切片可能共享同一段底层内存，改一个会"牵连"另一个，这正是许多诡异 bug 的根源。

**切片的声明**

```go
var s1 []int  // 声明一个切片，值为nil
s2 := []int{1, 2, 3, 4, 5}  // 声明并初始化
s3 := make([]int, 5)  // 创建一个长度为5的切片
s4 := make([]int, 5, 10)  // 创建一个长度为5、容量为10的切片
```

**切片的长度和容量**

切片有两个重要的属性：长度和容量。长度是切片中元素的数量，容量是从切片的第一个元素到底层数组的最后一个元素的数量。

```go
s := make([]int, 5, 10)
fmt.Println(len(s))  // 输出: 5（长度）
fmt.Println(cap(s))  // 输出: 10（容量）
```

**切片的访问和修改**

```go
s := []int{1, 2, 3, 4, 5}
fmt.Println(s[0])  // 输出: 1
s[0] = 10
fmt.Println(s[0])  // 输出: 10
```

**切片的追加**

使用`append`函数向切片追加元素：

```go
s := []int{1, 2, 3}
s = append(s, 4)
fmt.Println(s)  // 输出: [1 2 3 4]

s = append(s, 5, 6, 7)
fmt.Println(s)  // 输出: [1 2 3 4 5 6 7]
```

**切片的拼接**

```go
s1 := []int{1, 2, 3}
s2 := []int{4, 5, 6}
s1 = append(s1, s2...)
fmt.Println(s1)  // 输出: [1 2 3 4 5 6]
```

**切片的删除**

Go语言没有内置的删除函数，但可以使用切片操作实现：

```go
s := []int{1, 2, 3, 4, 5}

// 删除索引为2的元素
index := 2
s = append(s[:index], s[index+1:]...)
fmt.Println(s)  // 输出: [1 2 4 5]
```

**切片的复制**

使用`copy`函数复制切片：

```go
s1 := []int{1, 2, 3}
s2 := make([]int, len(s1))
copy(s2, s1)
fmt.Println(s2)  // 输出: [1 2 3]
```

**切片的截取**

```go
s := []int{1, 2, 3, 4, 5}
fmt.Println(s[1:3])  // 输出: [2 3]
fmt.Println(s[:3])   // 输出: [1 2 3]
fmt.Println(s[2:])   // 输出: [3 4 5]
fmt.Println(s[:])    // 输出: [1 2 3 4 5]
```

> ⚠️ **常见坑**：切片截取 `s[a:b]` 得到的新切片与原切片**共享底层数组**。若之后对其中一个 `append` 且未触发扩容，可能悄悄覆盖另一个的数据；需要隔离时用 `copy` 复制一份。另外，只声明未初始化的 `var s []int` 长度为 0，直接 `s[0] = 1` 会 panic（`index out of range`）。

**切片的遍历**

```go
s := []int{1, 2, 3, 4, 5}

// 使用索引遍历
for i := 0; i < len(s); i++ {
    fmt.Println(s[i])
}

// 使用range遍历
for index, value := range s {
    fmt.Printf("索引: %d, 值: %d\n", index, value)
}
```

**切片的传递**

由于切片是引用类型，传递给函数时不会复制整个序列：

```go
func modifySlice(s []int) {
    s[0] = 100
}

func main() {
    s := []int{1, 2, 3}
    modifySlice(s)
    fmt.Println(s)  // 输出: [100 2 3]（原切片被修改了）
}
```

### 7.3 映射

映射（Map）是Go语言中的一种无序的键值对集合，类似于其他语言中的字典或哈希表。

**映射的声明**

```go
var m1 map[string]int  // 声明一个映射，值为nil
m2 := map[string]int{"a": 1, "b": 2, "c": 3}  // 声明并初始化
m3 := make(map[string]int)  // 创建一个空映射
```

**映射的访问和修改**

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
fmt.Println(m["a"])  // 输出: 1
m["a"] = 10
fmt.Println(m["a"])  // 输出: 10
```

**映射的添加**

```go
m := make(map[string]int)
m["a"] = 1
m["b"] = 2
m["c"] = 3
fmt.Println(m)  // 输出: map[a:1 b:2 c:3]
```

**映射的删除**

使用`delete`函数删除映射中的元素：

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
delete(m, "b")
fmt.Println(m)  // 输出: map[a:1 c:3]
```

**映射的查询**

查询映射中的元素时，可以返回两个值：元素的值和一个布尔值，表示元素是否存在：

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
value, exists := m["a"]
if exists {
    fmt.Println("存在，值为:", value)
} else {
    fmt.Println("不存在")
}

value, exists = m["d"]
if exists {
    fmt.Println("存在，值为:", value)
} else {
    fmt.Println("不存在")
}
```

> ⚠️ **常见坑**：只声明未初始化的 map 是 `nil`，对 nil map 写入会 panic（`assignment to entry in nil map`），使用前必须先 `make`。另外直接 `m["x"]` 取不存在的键会返回**零值**而不报错，想区分"不存在"必须用两值形式 `v, ok := m["x"]`。

**映射的遍历**

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}

for key, value := range m {
    fmt.Printf("键: %s, 值: %d\n", key, value)
}
```

**映射的长度**

```go
m := map[string]int{"a": 1, "b": 2, "c": 3}
fmt.Println(len(m))  // 输出: 3
```

**映射的传递**

映射是引用类型，传递给函数时不会复制整个映射：

```go
func modifyMap(m map[string]int) {
    m["a"] = 100
}

func main() {
    m := map[string]int{"a": 1, "b": 2, "c": 3}
    modifyMap(m)
    fmt.Println(m)  // 输出: map[a:100 b:2 c:3]（原映射被修改了）
}
```

### 7.4 实例：学生成绩管理系统

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个学生成绩管理系统。

```go
package main

import (
    "fmt"
    "sort"
)

type Student struct {
    Name  string
    Age   int
    Score float64
}

func main() {
    students := []Student{
        {"张三", 18, 85.5},
        {"李四", 19, 72.0},
        {"王五", 18, 90.0},
        {"赵六", 20, 68.5},
    }

    fmt.Println("原始学生列表:")
    printStudents(students)

    fmt.Println("\n按成绩排序:")
    sortedStudents := make([]Student, len(students))
    copy(sortedStudents, students)
    sort.Slice(sortedStudents, func(i, j int) bool {
        return sortedStudents[i].Score > sortedStudents[j].Score
    })
    printStudents(sortedStudents)

    fmt.Println("\n成绩统计:")
    stats := calculateStats(students)
    fmt.Printf("平均成绩: %.2f\n", stats.Average)
    fmt.Printf("最高分: %.2f\n", stats.Max)
    fmt.Printf("最低分: %.2f\n", stats.Min)
    fmt.Printf("及格人数: %d\n", stats.PassCount)
    fmt.Printf("不及格人数: %d\n", stats.FailCount)

    fmt.Println("\n成绩分布:")
    distribution := calculateDistribution(students)
    for grade, count := range distribution {
        fmt.Printf("%s: %d人\n", grade, count)
    }
}

func printStudents(students []Student) {
    fmt.Println("姓名\t年龄\t成绩")
    fmt.Println("--------------------")
    for _, student := range students {
        fmt.Printf("%s\t%d\t%.2f\n", student.Name, student.Age, student.Score)
    }
}

type Stats struct {
    Average  float64
    Max      float64
    Min      float64
    PassCount int
    FailCount int
}

func calculateStats(students []Student) Stats {
    var total float64
    max := students[0].Score
    min := students[0].Score
    passCount := 0
    failCount := 0

    for _, student := range students {
        total += student.Score
        if student.Score > max {
            max = student.Score
        }
        if student.Score < min {
            min = student.Score
        }
        if student.Score >= 60 {
            passCount++
        } else {
            failCount++
        }
    }

    average := total / float64(len(students))

    return Stats{
        Average:  average,
        Max:      max,
        Min:      min,
        PassCount: passCount,
        FailCount: failCount,
    }
}

func calculateDistribution(students []Student) map[string]int {
    distribution := make(map[string]int)

    for _, student := range students {
        var grade string
        if student.Score >= 90 {
            grade = "优秀"
        } else if student.Score >= 80 {
            grade = "良好"
        } else if student.Score >= 60 {
            grade = "及格"
        } else {
            grade = "不及格"
        }
        distribution[grade]++
    }

    return distribution
}
```

这个程序：
1. 使用切片存储学生信息
2. 使用映射存储成绩分布
3. 使用切片的复制和排序
4. 使用映射的查询和统计
5. 使用结构体组织数据

运行这个程序：

```
原始学生列表:
姓名    年龄    成绩
--------------------
张三    18      85.50
李四    19      72.00
王五    18      90.00
赵六    20      68.50

按成绩排序:
姓名    年龄    成绩
--------------------
王五    18      90.00
张三    18      85.50
李四    19      72.00
赵六    20      68.50

成绩统计:
平均成绩: 79.00
最高分: 90.00
最低分: 68.50
及格人数: 4
不及格人数: 0

成绩分布:
优秀: 1人
良好: 1人
及格: 2人
不及格: 0人
```

### 7.5 小结

本章介绍了Go语言的数组、切片和映射，包括它们的定义、操作、遍历等。我们学习了每种数据结构的特点和使用方法。数组是固定长度的序列，切片是动态长度的序列，映射是键值对的集合。掌握好这些数据结构的使用对于编写高效的程序非常重要。我们还学习了切片和映射的引用特性，了解了它们在函数传递时的行为。

下一章我们将学习Go语言的结构体和方法，包括结构体的定义、方法的定义、接口等，以及它们的使用方法和注意事项。

**面试题**

1. 数组和切片有什么本质区别？"数组是值类型"意味着什么？
2. 切片的 `len` 和 `cap` 分别是什么？`append` 触发扩容的策略是怎样的？
3. 两个切片共享底层数组时会有什么风险？如何切断关联？
4. map 的零值与 `nil map` 有什么区别？写入 nil map 会怎样？
5. `delete` 和两值查询 `v, ok := m[k]` 的用途分别是什么？

## 第八章：结构体和方法

> 本章总览：接口只认方法不认类型，同一条调用语句在运行期落到不同实现。

```mermaid
flowchart TD
    I[接口 Shape] --> R[Rectangle 实现 Area 与 Perimeter]
    I --> C[Circle 实现 Area 与 Perimeter]
    I --> T[Triangle 实现 Area 与 Perimeter]
    R --> V[同一函数 printShapeInfo]
    C --> V
    T --> V
    V --> O[运行时按动态类型调用对应方法 多态]
```

### 8.1 结构体

结构体是Go语言中用于组织相关数据的自定义类型，类似于其他语言中的类或对象。

**结构体的定义**

```go
type Person struct {
    Name string
    Age  int
    Address string
}
```

**结构体的实例化**

```go
var p1 Person  // 声明一个Person类型的变量
p2 := Person{}  // 创建一个空的Person
p3 := Person{Name: "张三", Age: 18}  // 指定字段初始化
p4 := Person{"李四", 19, "北京"}  // 按顺序初始化
```

**结构体字段的访问和修改**

```go
p := Person{Name: "张三", Age: 18}
fmt.Println(p.Name)  // 输出: 张三
p.Age = 19
fmt.Println(p.Age)  // 输出: 19
```

**结构体的指针**

```go
p := &Person{Name: "张三", Age: 18}
fmt.Println(p.Name)  // 输出: 张三（Go语言会自动解引用）
p.Age = 19
fmt.Println(p.Age)  // 输出: 19
```

**匿名结构体**

```go
p := struct {
    Name string
    Age  int
}{
    Name: "张三",
    Age:  18,
}
fmt.Println(p.Name)  // 输出: 张三
```

**嵌套结构体**

```go
type Address struct {
    City    string
    Country string
}

type Person struct {
    Name    string
    Age     int
    Address Address
}

p := Person{
    Name: "张三",
    Age:  18,
    Address: Address{
        City:    "北京",
        Country: "中国",
    },
}
fmt.Println(p.Address.City)  // 输出: 北京
```

**匿名嵌套结构体**

```go
type Address struct {
    City    string
    Country string
}

type Person struct {
    Name string
    Age  int
    Address
}

p := Person{
    Name: "张三",
    Age:  18,
    Address: Address{
        City:    "北京",
        Country: "中国",
    },
}
fmt.Println(p.City)  // 输出: 北京（可以直接访问嵌套结构体的字段）
```

### 8.2 方法

方法是绑定到特定类型的函数，类似于其他语言中的成员函数。

**方法的定义**

```go
type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}
    area := rect.Area()
    fmt.Println(area)  // 输出: 50
}
```

**值接收者和指针接收者**

方法可以使用值接收者或指针接收者：

```go
type Counter struct {
    count int
}

func (c Counter) Value() int {
    return c.count
}

func (c *Counter) Increment() {
    c.count++
}

func main() {
    counter := Counter{count: 0}
    fmt.Println(counter.Value())  // 输出: 0
    counter.Increment()
    fmt.Println(counter.Value())  // 输出: 1
}
```

值接收者和指针接收者的区别：
- 值接收者：方法内部对接收者的修改不会影响原始值
- 指针接收者：方法内部对接收者的修改会影响原始值

```go
type Person struct {
    Name string
    Age  int
}

func (p Person) ModifyValue() {
    p.Name = "修改后的名字"
}

func (p *Person) ModifyPointer() {
    p.Name = "修改后的名字"
}

func main() {
    p1 := Person{Name: "张三", Age: 18}
    p1.ModifyValue()
    fmt.Println(p1.Name)  // 输出: 张三（没有被修改）

    p2 := Person{Name: "李四", Age: 19}
    p2.ModifyPointer()
    fmt.Println(p2.Name)  // 输出: 修改后的名字（被修改了）
}
```

**何时使用值接收者，何时使用指针接收者**

- 使用值接收者的情况：
  - 接收者是不可变的
  - 接收者是一个小的数据结构
  - 接收者是一个基本类型或切片

- 使用指针接收者的情况：
  - 接收者需要被修改
  - 接收者是一个大的数据结构
  - 接收者是一个需要保持一致性的数据结构

> ⚠️ **常见坑**：接口里存的是"值 + 类型"。如果把一个**值类型**赋给接口，之后调用需要指针接收者的方法会因方法集不匹配而编译失败（提示 `does not implement`）。此外，值接收者的方法修改的是副本，不会影响原对象；需要修改状态时务必用指针接收者，并让接口持有指针。

### 8.3 接口

接口定义了一组方法的集合，任何类型只要实现了接口中定义的所有方法，就实现了该接口。

> 💡 通俗理解：接口像一份"职位说明书"，它不关心你是谁，只关心你会不会干这几件事。只要一个类型实现了说明书上列出的所有方法，它就"自动"满足该接口——Go 里这叫**隐式实现**，不需要像 Java 那样写 `implements`。于是我们能把不同具体类型装进同一个接口切片，用同一段代码统一处理，这就是多态。

**接口的定义**

```go
type Shape interface {
    Area() float64
    Perimeter() float64
}
```

**接口的实现**

```go
type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}
```

**接口的使用**

```go
func printShapeInfo(s Shape) {
    fmt.Printf("面积: %.2f, 周长: %.2f\n", s.Area(), s.Perimeter())
}

func main() {
    rect := Rectangle{Width: 10, Height: 5}
    circle := Circle{Radius: 5}

    printShapeInfo(rect)    // 输出: 面积: 50.00, 周长: 30.00
    printShapeInfo(circle)  // 输出: 面积: 78.54, 周长: 31.42
}
```

**空接口**

空接口不包含任何方法，所有类型都实现了空接口：

```go
var i interface{}
i = 10
fmt.Println(i)  // 输出: 10

i = "Hello"
fmt.Println(i)  // 输出: Hello

i = []int{1, 2, 3}
fmt.Println(i)  // 输出: [1 2 3]
```

**类型断言**

类型断言用于将接口类型转换为具体类型：

```go
var i interface{} = "Hello"

s, ok := i.(string)
if ok {
    fmt.Println("字符串:", s)  // 输出: 字符串: Hello
}

n, ok := i.(int)
if ok {
    fmt.Println("整数:", n)
} else {
    fmt.Println("不是整数")  // 输出: 不是整数
}
```

**类型switch**

类型switch用于判断接口的具体类型：

```go
func printType(i interface{}) {
    switch v := i.(type) {
    case int:
        fmt.Println("整数:", v)
    case string:
        fmt.Println("字符串:", v)
    case bool:
        fmt.Println("布尔值:", v)
    default:
        fmt.Println("未知类型")
    }
}

func main() {
    printType(10)        // 输出: 整数: 10
    printType("Hello")   // 输出: 字符串: Hello
    printType(true)      // 输出: 布尔值: true
    printType(3.14)      // 输出: 未知类型
}
```

### 8.4 实例：图形计算器

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个图形计算器。

```go
package main

import (
    "fmt"
    "math"
)

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Rectangle struct {
    Width  float64
    Height float64
}

func (r Rectangle) Area() float64 {
    return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
    return 2 * (r.Width + r.Height)
}

func (r Rectangle) String() string {
    return fmt.Sprintf("矩形(宽=%.2f, 高=%.2f)", r.Width, r.Height)
}

type Circle struct {
    Radius float64
}

func (c Circle) Area() float64 {
    return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.Radius
}

func (c Circle) String() string {
    return fmt.Sprintf("圆形(半径=%.2f)", c.Radius)
}

type Triangle struct {
    Base   float64
    Height float64
    SideA  float64
    SideB  float64
    SideC  float64
}

func (t Triangle) Area() float64 {
    return 0.5 * t.Base * t.Height
}

func (t Triangle) Perimeter() float64 {
    return t.SideA + t.SideB + t.SideC
}

func (t Triangle) String() string {
    return fmt.Sprintf("三角形(底=%.2f, 高=%.2f)", t.Base, t.Height)
}

func printShapeInfo(s Shape) {
    fmt.Printf("%s\n", s)
    fmt.Printf("面积: %.2f\n", s.Area())
    fmt.Printf("周长: %.2f\n", s.Perimeter())
    fmt.Println("--------------------")
}

func main() {
    shapes := []Shape{
        Rectangle{Width: 10, Height: 5},
        Circle{Radius: 5},
        Triangle{Base: 6, Height: 4, SideA: 3, SideB: 4, SideC: 5},
    }

    fmt.Println("图形计算器")
    fmt.Println("====================")

    for _, shape := range shapes {
        printShapeInfo(shape)
    }

    fmt.Println("总面积:")
    var totalArea float64
    for _, shape := range shapes {
        totalArea += shape.Area()
    }
    fmt.Printf("%.2f\n", totalArea)
}
```

这个程序：
1. 定义了Shape接口
2. 实现了Rectangle、Circle、Triangle三种图形
3. 使用接口的多态特性
4. 使用String方法自定义输出
5. 使用切片存储不同类型的图形

运行这个程序：

```
图形计算器
====================
矩形(宽=10.00, 高=5.00)
面积: 50.00
周长: 30.00
--------------------
圆形(半径=5.00)
面积: 78.54
周长: 31.42
--------------------
三角形(底=6.00, 高=4.00)
面积: 12.00
周长: 12.00
--------------------
总面积:
140.54
```

### 8.5 小结

本章介绍了Go语言的结构体和方法，包括结构体的定义、嵌套结构体、方法的定义、值接收者和指针接收者、接口的定义和实现等。我们学习了每种特性的使用方法和注意事项。结构体是Go语言中组织数据的基本方式，方法是绑定到类型的函数，接口定义了一组方法的集合。掌握好这些特性对于编写面向对象的Go程序非常重要。

下一章我们将学习Go语言的错误处理和测试，包括错误的定义、错误的处理、测试的编写等，以及它们的使用方法和注意事项。

**面试题**

1. 值接收者和指针接收者的区别？方法集规则如何影响接口实现？
2. Go 的接口是"隐式实现"的，这带来什么好处与坑？
3. 空接口 `interface{}`（`any`）能存什么？如何安全地取出具体类型？
4. 类型断言 `v, ok := i.(T)` 和类型 `switch` 分别怎么用？
5. 结构体嵌入与继承有什么区别？如何访问被嵌入类型的方法？

## 第九章：错误处理和测试

> 本章总览：错误是返回值，显式判断后决定包装、识别还是记录。

```mermaid
flowchart TD
    A[调用可能出错的函数] --> B{err 是否为 nil}
    B -->|是| C[使用正常返回值]
    B -->|否| D[用 fmt.Errorf 加 %w 包装]
    D --> E[向上返回 让调用者决策]
    E --> F{需要识别原始错误}
    F -->|是| G[errors.Is 或 errors.As]
    F -->|否| H[直接记录日志]
    C --> Z[流程继续]
    G --> Z
    H --> Z
```

### 9.1 错误处理

Go语言采用显式的错误处理机制，通过返回值来传递错误信息。这种机制让错误处理更加明确和可控。

> 💡 通俗理解：Go 把错误当成**普通返回值**，而不是异常。就像点外卖时骑手会先告诉你"送到了 / 路上出了状况"，你必须显式判断 `if err != nil` 再决定怎么办。好处是错误路径一眼可见、不会被悄悄吞掉；代价是要多写几行检查。`panic` 则是"着火了"级别的紧急情况，只用于不可恢复的错误，别拿它当 `return err` 用。

**错误的基本概念**

在Go语言中，错误是一个实现了`error`接口的值：

```go
type error interface {
    Error() string
}
```

**创建错误**

使用`errors.New()`函数创建错误：

```go
import "errors"

err := errors.New("发生错误")
fmt.Println(err)  // 输出: 发生错误
```

使用`fmt.Errorf()`函数创建格式化的错误：

```go
err := fmt.Errorf("错误代码: %d, 错误信息: %s", 404, "Not Found")
fmt.Println(err)  // 输出: 错误代码: 404, 错误信息: Not Found
```

**返回错误**

函数通常将错误作为最后一个返回值：

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("除数不能为零")
    }
    return a / b, nil
}
```

**处理错误**

调用函数时检查错误：

```go
result, err := divide(10, 0)
if err != nil {
    fmt.Println("发生错误:", err)
    return
}
fmt.Println("结果:", result)
```

**自定义错误类型**

可以定义自己的错误类型：

```go
type MyError struct {
    Code    int
    Message string
}

func (e *MyError) Error() string {
    return fmt.Sprintf("错误代码: %d, 错误信息: %s", e.Code, e.Message)
}

func doSomething() error {
    return &MyError{Code: 404, Message: "Not Found"}
}
```

**错误包装**

Go 1.13引入了错误包装功能，可以使用`fmt.Errorf`和`%w`动词包装错误：

```go
func readFile(filename string) error {
    _, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("读取文件失败: %w", err)
    }
    return nil
}
```

使用`errors.Is()`和`errors.As()`判断错误：

```go
err := readFile("test.txt")
if errors.Is(err, os.ErrNotExist) {
    fmt.Println("文件不存在")
}

var pathError *os.PathError
if errors.As(err, &pathError) {
    fmt.Println("路径错误:", pathError.Path)
}
```

> ⚠️ **常见坑**：只有用 `%w` 包装才能保留错误链，之后 `errors.Is`/`errors.As` 才能识别；用 `%v` 会把原始错误"拍平"成字符串，丢失类型信息。另一个高频坑是 `recover` 只有在 `defer` 里调用才有效，而且它只对**同一个 goroutine** 的 panic 生效，在别的 goroutine 里 panic 是接不住的。

**panic和recover**

`panic`用于处理不可恢复的错误，`recover`用于从panic中恢复：

```go
func safeFunction() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("从panic中恢复:", r)
        }
    }()

    panic("发生panic")
}

func main() {
    safeFunction()
    fmt.Println("程序继续执行")
}
```

**何时使用panic**

- 发生不可恢复的错误，如数组越界
- 程序无法继续执行
- 初始化失败

**何时使用error**

- 可预见的错误，如文件不存在
- 可以恢复的错误
- 需要调用者处理的错误

### 9.2 测试

Go语言内置了测试框架，使用`testing`包编写测试。

**基本测试**

测试文件以`_test.go`结尾，测试函数以`Test`开头：

```go
package main

import "testing"

func TestAdd(t *testing.T) {
    result := add(2, 3)
    expected := 5
    if result != expected {
        t.Errorf("add(2, 3) = %d; expected %d", result, expected)
    }
}
```

**运行测试**

```bash
go test
```

**表驱动测试**

使用表驱动测试可以测试多个用例：

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        a, b     int
        expected int
    }{
        {1, 2, 3},
        {2, 3, 5},
        {0, 0, 0},
        {-1, 1, 0},
    }

    for _, tt := range tests {
        result := add(tt.a, tt.b)
        if result != tt.expected {
            t.Errorf("add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
        }
    }
}
```

**子测试**

使用子测试可以组织测试用例：

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"正数相加", 2, 3, 5},
        {"零相加", 0, 0, 0},
        {"负数相加", -1, 1, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("add(%d, %d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

**基准测试**

基准测试用于测试函数的性能：

```go
func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        add(2, 3)
    }
}
```

运行基准测试：

```bash
go test -bench=.
```

**示例测试**

示例测试用于展示函数的用法：

```go
func ExampleAdd() {
    result := add(2, 3)
    fmt.Println(result)
    // Output: 5
}
```

**测试覆盖率**

查看测试覆盖率：

```bash
go test -cover
```

生成覆盖率报告：

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 9.3 实例：文件处理工具

让我们用一个实际的例子来巩固本章学到的知识。我们将编写一个文件处理工具。

```go
package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type FileProcessor struct {
    filename string
}

func NewFileProcessor(filename string) *FileProcessor {
    return &FileProcessor{filename: filename}
}

func (fp *FileProcessor) ReadLines() ([]string, error) {
    file, err := os.Open(fp.filename)
    if err != nil {
        return nil, fmt.Errorf("打开文件失败: %w", err)
    }
    defer file.Close()

    var lines []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        lines = append(lines, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, fmt.Errorf("读取文件失败: %w", err)
    }

    return lines, nil
}

func (fp *FileProcessor) CountWords() (int, error) {
    lines, err := fp.ReadLines()
    if err != nil {
        return 0, err
    }

    wordCount := 0
    for _, line := range lines {
        words := strings.Fields(line)
        wordCount += len(words)
    }

    return wordCount, nil
}

func (fp *FileProcessor) CountNumbers() (int, error) {
    lines, err := fp.ReadLines()
    if err != nil {
        return 0, err
    }

    numberCount := 0
    for _, line := range lines {
        words := strings.Fields(line)
        for _, word := range words {
            if _, err := strconv.Atoi(word); err == nil {
                numberCount++
            }
        }
    }

    return numberCount, nil
}

func (fp *FileProcessor) Process() (wordCount, numberCount int, err error) {
    wordCount, err = fp.CountWords()
    if err != nil {
        return 0, 0, err
    }

    numberCount, err = fp.CountNumbers()
    if err != nil {
        return 0, 0, err
    }

    return wordCount, numberCount, nil
}

func main() {
    if len(os.Args) < 2 {
        fmt.Println("请提供文件名")
        os.Exit(1)
    }

    filename := os.Args[1]
    processor := NewFileProcessor(filename)

    wordCount, numberCount, err := processor.Process()
    if err != nil {
        fmt.Println("处理文件失败:", err)
        os.Exit(1)
    }

    fmt.Printf("文件: %s\n", filename)
    fmt.Printf("单词数: %d\n", wordCount)
    fmt.Printf("数字数: %d\n", numberCount)
}
```

测试文件：

```go
package main

import (
    "os"
    "strings"
    "testing"
)

func TestFileProcessor_ReadLines(t *testing.T) {
    content := "Hello World\n123 456\nGo Language"
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    processor := NewFileProcessor(tmpfile.Name())
    lines, err := processor.ReadLines()
    if err != nil {
        t.Errorf("ReadLines() error = %v", err)
        return
    }

    expected := []string{"Hello World", "123 456", "Go Language"}
    if len(lines) != len(expected) {
        t.Errorf("ReadLines() = %v, want %v", lines, expected)
    }

    for i, line := range lines {
        if line != expected[i] {
            t.Errorf("ReadLines()[%d] = %v, want %v", i, line, expected[i])
        }
    }
}

func TestFileProcessor_CountWords(t *testing.T) {
    content := "Hello World\n123 456\nGo Language"
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    processor := NewFileProcessor(tmpfile.Name())
    wordCount, err := processor.CountWords()
    if err != nil {
        t.Errorf("CountWords() error = %v", err)
        return
    }

    expected := 6
    if wordCount != expected {
        t.Errorf("CountWords() = %d, want %d", wordCount, expected)
    }
}

func TestFileProcessor_CountNumbers(t *testing.T) {
    content := "Hello World\n123 456\nGo Language"
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    processor := NewFileProcessor(tmpfile.Name())
    numberCount, err := processor.CountNumbers()
    if err != nil {
        t.Errorf("CountNumbers() error = %v", err)
        return
    }

    expected := 2
    if numberCount != expected {
        t.Errorf("CountNumbers() = %d, want %d", numberCount, expected)
    }
}

func TestFileProcessor_Process(t *testing.T) {
    content := "Hello World\n123 456\nGo Language"
    tmpfile, err := os.CreateTemp("", "test")
    if err != nil {
        t.Fatal(err)
    }
    defer os.Remove(tmpfile.Name())

    if _, err := tmpfile.WriteString(content); err != nil {
        t.Fatal(err)
    }
    if err := tmpfile.Close(); err != nil {
        t.Fatal(err)
    }

    processor := NewFileProcessor(tmpfile.Name())
    wordCount, numberCount, err := processor.Process()
    if err != nil {
        t.Errorf("Process() error = %v", err)
        return
    }

    if wordCount != 6 {
        t.Errorf("Process() wordCount = %d, want %d", wordCount, 6)
    }

    if numberCount != 2 {
        t.Errorf("Process() numberCount = %d, want %d", numberCount, 2)
    }
}
```

这个程序：
1. 定义了FileProcessor结构体
2. 实现了文件读取、单词计数、数字计数等功能
3. 使用错误处理机制处理各种错误
4. 编写了完整的测试用例
5. 使用临时文件进行测试

运行测试：

```bash
go test -v
```

运行程序：

```bash
go run main.go test.txt
```

### 9.4 小结

本章介绍了Go语言的错误处理和测试，包括错误的定义、错误的处理、测试的编写等。我们学习了每种特性的使用方法和注意事项。Go语言采用显式的错误处理机制，让错误处理更加明确和可控。测试是保证代码质量的重要手段，Go语言内置的测试框架使得编写测试变得简单。掌握好错误处理和测试对于编写可靠的Go程序非常重要。

下一章我们将学习Go语言的包管理和模块，包括包的定义、模块的使用、依赖管理等，以及它们的使用方法和注意事项。

**面试题**

1. Go 的 `error` 接口是什么？为什么习惯把 error 放在最后一个返回值？
2. `panic` 与 `error` 的使用场景如何区分？什么时候该用 `recover`？
3. `errors.Is` 和 `errors.As` 的区别？`%w` 与 `%v` 包装错误有何不同？
4. 如何自定义错误类型并让它支持错误链判断？
5. Go 的表驱动测试和子测试怎么写？`go test -bench` 是做什么的？

## 第十章：包管理和模块

> 本章总览：一个 Go 项目从无模块到可构建，会经历的状态迁移。

```mermaid
stateDiagram-v2
    [*] --> 无模块
    无模块 --> 已初始化: go mod init
    已初始化 --> 已添加依赖: go get
    已添加依赖 --> 已整理: go mod tidy
    已整理 --> 已构建: go build 或 go run
    已构建 --> 已整理: 继续改代码
    已整理 --> [*]
```

### 10.1 包

包是Go语言中组织代码的基本单位，每个Go文件都必须属于一个包。

**包的声明**

每个Go文件的第一行都必须声明包名：

```go
package main
```

**main包**

`main`包是一个特殊的包，它定义了一个可独立执行的程序：

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
```

**包的命名**

包名通常使用小写字母，并且应该简短、有意义：

```go
package utils
package math
package strings
```

**包的导入**

使用`import`语句导入包：

```go
import "fmt"
import "math"
```

可以同时导入多个包：

```go
import (
    "fmt"
    "math"
    "strings"
)
```

**包的别名**

可以为导入的包指定别名：

```go
import (
    f "fmt"
    m "math"
)

func main() {
    f.Println(m.Pi)
}
```

**点导入**

点导入可以直接使用包中的标识符，不需要包名前缀：

```go
import . "fmt"

func main() {
    Println("Hello, World!")
}
```

**下划线导入**

下划线导入只执行包的初始化函数，不使用包中的任何标识符：

```go
import _ "database/sql"
```

### 10.2 模块

Go Modules是Go语言的依赖管理系统，从Go 1.11版本开始引入。

> 💡 通俗理解：**包**是代码的"文件夹"和命名空间，靠首字母大小写决定谁能被外部引用；**模块**是项目的"身份证 + 购物清单"，由 `go.mod` 定义模块路径、Go 版本与依赖，`go.sum` 则像"防伪码"记录每个依赖的校验和。有了模块，项目可以放在磁盘任意位置，不再被 `GOPATH` 绑死。

**初始化模块**

在项目根目录下执行：

```bash
go mod init example.com/myproject
```

这会创建一个`go.mod`文件：

```go
module example.com/myproject

go 1.21
```

**添加依赖**

使用`go get`命令添加依赖：

```bash
go get github.com/gin-gonic/gin
```

这会更新`go.mod`文件，并创建`go.sum`文件。

> ⚠️ **常见坑**：`go.mod` 和 `go.sum` 都要提交到版本库，但**不要手动编辑** `go.sum`；增删代码后跑 `go mod tidy` 让它自动对齐即可。另一个坑是 `module` 路径必须与 import 路径一致，否则会报无法解析的导入。使用私有模块时要设置 `GOPRIVATE`，避免被公开代理抓取。

**更新依赖**

更新所有依赖：

```bash
go get -u
```

更新特定依赖：

```bash
go get -u github.com/gin-gonic/gin
```

**移除依赖**

移除未使用的依赖：

```bash
go mod tidy
```

**查看依赖**

查看模块依赖：

```bash
go list -m all
```

**下载依赖**

下载依赖到本地缓存：

```bash
go mod download
```

**验证依赖**

验证依赖的完整性：

```bash
go mod verify
```

### 10.3 go.mod文件

`go.mod`文件定义了模块的基本信息和依赖关系。

**基本结构**

```go
module example.com/myproject

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/go-sql-driver/mysql v1.7.1
)

require (
    github.com/bytedance/sonic v1.9.1
    github.com/chenzhuoyu/base64x v0.0.0-20221115062448-fe3a3abad311
)
```

**module**

指定模块的路径：

```go
module example.com/myproject
```

**go**

指定Go语言的版本：

```go
go 1.21
```

**require**

指定模块的依赖：

```go
require github.com/gin-gonic/gin v1.9.1
```

**indirect**

间接依赖：

```go
require (
    github.com/bytedance/sonic v1.9.1 // indirect
)
```

**replace**

替换依赖：

```go
replace github.com/gin-gonic/gin => github.com/gin-gonic/gin v1.9.0
```

**exclude**

排除特定版本：

```go
exclude github.com/gin-gonic/gin v1.8.0
```

### 10.4 go.sum文件

`go.sum`文件记录了依赖的校验和，用于确保依赖的完整性。

**基本结构**

```
github.com/gin-gonic/gin v1.9.1 h1:...
github.com/gin-gonic/gin v1.9.1/go.mod h1:...
github.com/go-sql-driver/mysql v1.7.1 h1:...
github.com/go-sql-driver/mysql v1.7.1/go.mod h1:...
```

**作用**

- 确保依赖的完整性
- 防止依赖被篡改
- 支持可重现的构建

**注意事项**

- 不要手动编辑`go.sum`文件
- 将`go.sum`文件提交到版本控制系统
- 使用`go mod verify`验证依赖的完整性

### 10.5 私有模块

如果要使用私有模块，需要配置Git凭证。

**配置Git凭证**

```bash
git config --global url."https://username:token@github.com/".insteadOf "https://github.com/"
```

**配置GOPRIVATE**

```bash
go env -w GOPRIVATE=github.com/mycompany/*
```

**配置Go代理**

```bash
go env -w GOPROXY=https://goproxy.cn,direct
```

### 10.6 实例：创建一个Web应用

让我们用一个实际的例子来巩固本章学到的知识。我们将创建一个简单的Web应用。

**初始化项目**

```bash
mkdir mywebapp
cd mywebapp
go mod init example.com/mywebapp
```

**添加依赖**

```bash
go get github.com/gin-gonic/gin
```

**编写代码**

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

var users []User

func main() {
    r := gin.Default()

    r.GET("/users", getUsers)
    r.GET("/users/:id", getUser)
    r.POST("/users", createUser)
    r.PUT("/users/:id", updateUser)
    r.DELETE("/users/:id", deleteUser)

    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
    id := c.Param("id")
    for _, user := range users {
        if user.ID == parseID(id) {
            c.JSON(http.StatusOK, user)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func createUser(c *gin.Context) {
    var user User
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    user.ID = len(users) + 1
    users = append(users, user)
    c.JSON(http.StatusCreated, user)
}

func updateUser(c *gin.Context) {
    id := c.Param("id")
    var updatedUser User
    if err := c.ShouldBindJSON(&updatedUser); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    for i, user := range users {
        if user.ID == parseID(id) {
            updatedUser.ID = user.ID
            users[i] = updatedUser
            c.JSON(http.StatusOK, updatedUser)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func deleteUser(c *gin.Context) {
    id := c.Param("id")
    for i, user := range users {
        if user.ID == parseID(id) {
            users = append(users[:i], users[i+1:]...)
            c.JSON(http.StatusNoContent, nil)
            return
        }
    }
    c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
}

func parseID(id string) int {
    var result int
    fmt.Sscanf(id, "%d", &result)
    return result
}
```

**运行项目**

```bash
go run main.go
```

**测试API**

```bash
# 获取所有用户
curl http://localhost:8080/users

# 创建用户
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"张三","email":"zhangsan@example.com"}'

# 获取特定用户
curl http://localhost:8080/users/1

# 更新用户
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"李四","email":"lisi@example.com"}'

# 删除用户
curl -X DELETE http://localhost:8080/users/1
```

### 10.7 小结

本章介绍了Go语言的包管理和模块，包括包的定义、模块的使用、依赖管理等。我们学习了每种特性的使用方法和注意事项。包是Go语言中组织代码的基本单位，模块是Go语言的依赖管理系统。掌握好包管理和模块对于编写可维护的Go程序非常重要。

**面试题**

1. 包（package）和模块（module）有什么区别？
2. `go.mod` 里的 `module`、`go`、`require`、`replace`、`indirect` 各是什么含义？
3. `go.sum` 的作用是什么？为什么不要手动修改它？
4. `go get`、`go mod tidy`、`go mod download`、`go mod verify` 分别做什么？
5. 导入包的别名、点导入、下划线导入分别有什么用途和风险？

## 总结

本书系统地介绍了Go语言的基础语法，从Go语言的简介和环境搭建开始，逐步深入到变量、常量、数据类型、运算符、控制流、函数、数组、切片、映射、结构体、方法、接口、错误处理、测试、包管理和模块等各个方面。

通过学习本书，读者应该能够：

1. 理解Go语言的基本概念和特点
2. 掌握Go语言的基本语法
3. 能够编写简单的Go程序
4. 理解Go语言的面向对象特性
5. 掌握Go语言的错误处理和测试
6. 了解Go语言的包管理和模块

Go语言以其简洁的语法、强大的并发支持和完善的工具链，成为了现代软件开发的重要选择。希望本书能够帮助读者打下坚实的Go语言基础，为后续的学习和实践做好准备。

在掌握了Go语言的基础语法后，读者可以继续学习Go语言的并发编程、网络编程、微服务架构等高级主题，进一步提升自己的Go语言编程能力。

## 附录：高频面试题汇总

以下题目覆盖全书十章，建议先自测再看正文对应章节。

**语言与环境**

1. Go 语言有哪些主要特点？为什么它特别适合云原生与微服务？
2. Go 是编译型语言吗？`go run`、`go build`、`gofmt`、`go vet` 各自做什么？

**变量、常量与类型**

3. `var`、`:=`、`const` 的使用场景与限制分别是什么？
4. Go 变量和常量的零值分别是什么？不同类型各是什么？
5. 什么是变量遮蔽？为什么它容易引发隐蔽 bug？
6. `iota` 的规则是什么？如何用它定义一组枚举？
7. `len` 和 `cap` 的区别？字符串的 `len` 返回的是什么？
8. `byte` 和 `rune` 是什么的别名？遍历含中文的字符串要注意什么？
9. 类型别名 `type A = int` 与类型定义 `type A int` 有何不同？
10. 为什么浮点数不能用 `==` 直接比较？金额计算该如何处理？

**运算符与控制流**

11. `&&`、`||` 的短路求值是什么？有什么实际用途？
12. `&^`（位清除）的规则与典型用途是什么？
13. Go 的 `switch` 为什么不需要 `break`？`fallthrough` 有何作用？
14. 无表达式的 `switch` 和类型 `switch` 分别怎么用？
15. `defer` 的执行顺序与参数求值时机是怎样的？它如何修改命名返回值？
16. `break`、`continue`、`goto` 与标签跳转的区别与禁忌？

**函数**

17. 值传递与所谓"引用传递"的区别？切片、映射传参到底传了什么？
18. 什么是闭包？循环里创建闭包会踩什么坑，如何解决？
19. 函数为什么是 Go 的"一等公民"？如何把函数当参数和返回值？

**复合类型**

20. 数组与切片的本质区别？"数组是值类型"意味着什么？
21. 切片的 `len` 与 `cap` 如何配合？`append` 扩容时发生了什么？
22. 两个切片共享底层数组有什么风险？如何切断关联？
23. map 的零值与 `nil map` 有何区别？写入 nil map 会怎样？

**面向对象与错误处理**

24. 值接收者与指针接收者有何区别？方法集如何影响接口实现？
25. Go 的接口为什么是"隐式实现"的？空接口能存什么、如何安全取出？
26. `error`、`panic`、`recover` 三者的定位与使用场景？`%w` 与 `%v` 有何不同？
27. 如何自定义错误类型并支持 `errors.Is`/`errors.As` 判断？

**工程与依赖**

28. 包和模块有什么区别？`go.mod` 与 `go.sum` 各承担什么职责？
29. `go get`、`go mod tidy`、`go mod download`、`go mod verify` 分别做什么？
30. Go 的表驱动测试、子测试、基准测试分别怎么写与运行？
