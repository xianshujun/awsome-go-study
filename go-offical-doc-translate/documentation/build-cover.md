> **[↩ 返回到目录](doc.md)**


# 集成测试的覆盖率分析支持

## 目录
- 概述
- 构建用于覆盖率分析的二进制文件
- 运行经过覆盖率插桩的二进制文件
- 处理覆盖率数据文件
- 常见问题
- 参考资源
- 术语表


从 Go 1.20 开始，Go 支持从应用程序和集成测试（即针对 Go 程序的更大、更复杂的测试）中收集覆盖率分析数据。

## 概述
Go 通过 `go test -coverprofile=... <pkg_target>` 命令，为在包级单元测试中收集覆盖率分析数据提供了易于使用的支持。从 Go 1.20 开始，用户现在可以为更大规模的集成测试收集覆盖率分析数据：这些测试更重量级、更复杂，会对特定的应用程序二进制文件执行多次运行。

对于单元测试，收集覆盖率分析数据并生成报告需要两个步骤：首先运行 `go test -coverprofile=...`，然后调用 `go tool cover {-func,-html}` 生成报告。

对于集成测试，则需要三个步骤：构建步骤、运行步骤（可能包括多次调用构建步骤生成的二进制文件），以及最后的报告生成步骤，如下所述。

## 构建用于覆盖率分析的二进制文件
要构建用于收集覆盖率分析数据的应用程序，在对应用程序二进制目标调用 `go build` 时传递 `-cover` 标志。请参见下面的 `go build -cover` 调用示例。然后，可以通过设置环境变量来运行生成的二进制文件，以捕获覆盖率分析数据（参见下一节关于运行的内容）。

### 如何选择要插桩的包
在特定的 `go build -cover` 调用期间，Go 命令会选择主模块中的包进行覆盖率分析；默认情况下，其他参与构建的包（在 go.mod 中列出的依赖项，或属于 Go 标准库的包）不会被包含在内。

例如，这里有一个包含主包、本地主模块包 greetings 以及一组从模块外部导入的包的简单程序，其中包括 rsc.io/quote 和 fmt 等（完整程序链接）。

```bash
$ cat go.mod
module mydomain.com

go 1.20

require rsc.io/quote v1.5.2

require (
    golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c // indirect
    rsc.io/sampler v1.3.0 // indirect
)

$ cat myprogram.go
package main

import (
    "fmt"
    "mydomain.com/greetings"
    "rsc.io/quote"
)

func main() {
    fmt.Printf("I say %q and %q\n", quote.Hello(), greetings.Goodbye())
}
$ cat greetings/greetings.go
package greetings

func Goodbye() string {
    return "see ya"
}
$ go build -cover -o myprogram.exe .
$
```

如果使用 `-cover` 命令行标志构建此程序并运行，分析数据中将恰好包含两个包：main 和 mydomain.com/greetings；其他依赖包将被排除。

希望更精确地控制哪些包被包含在覆盖率分析中的用户，可以使用 `-coverpkg` 标志进行构建。例如：

```bash
$ go build -cover -o myprogramMorePkgs.exe -coverpkg=io,mydomain.com,rsc.io/quote .
$
```

在上面的构建中，来自 mydomain.com 的主包以及 rsc.io/quote 和 io 包被选中进行分析；由于 mydomain.com/greetings 没有被特别列出，即使它位于主模块中，也会被排除在分析之外。

## 运行经过覆盖率插桩的二进制文件
使用 `-cover` 构建的二进制文件在执行结束时，会将分析数据文件写入到通过环境变量 GOCOVERDIR 指定的目录中。例如：

```bash
$ go build -cover -o myprogram.exe myprogram.go
$ mkdir somedata
$ GOCOVERDIR=somedata ./myprogram.exe
I say "Hello, world." and "see ya"
$ ls somedata
covcounters.c6de772f99010ef5925877a7b05db4cc.2424989.1670252383678349347
covmeta.c6de772f99010ef5925877a7b05db4cc
$
```

注意写入到 somedata 目录的两个文件：这些（二进制）文件包含覆盖率结果。有关如何从这些数据文件生成人类可读结果的更多信息，请参见下面的报告生成部分。

如果未设置 GOCOVERDIR 环境变量，经过覆盖率插桩的二进制文件仍会正确执行，但会发出警告。例如：

```bash
$ ./myprogram.exe
warning: GOCOVERDIR not set, no coverage data emitted
I say "Hello, world." and "see ya"
$
```

### 涉及多次运行的测试
在许多情况下，集成测试可能涉及多次程序运行；当程序使用 `-cover` 构建时，每次运行都会生成一个新的数据文件。例如：

```bash
$ mkdir somedata2
$ GOCOVERDIR=somedata2 ./myprogram.exe          // 第一次运行
I say "Hello, world." and "see ya"
$ GOCOVERDIR=somedata2 ./myprogram.exe -flag    // 第二次运行
I say "Hello, world." and "see ya"
$ ls somedata2
covcounters.890814fca98ac3a4d41b9bd2a7ec9f7f.2456041.1670259309405583534
covcounters.890814fca98ac3a4d41b9bd2a7ec9f7f.2456047.1670259309410891043
covmeta.890814fca98ac3a4d41b9bd2a7ec9f7f
$
```

覆盖率数据输出文件有两种类型：元数据文件（包含在多次运行中不变的项，如源文件名和函数名）和计数器数据文件（记录程序执行的部分）。

在上面的示例中，第一次运行生成了两个文件（计数器和元数据），而第二次运行只生成了一个计数器数据文件：由于元数据在多次运行中不会改变，因此只需要写入一次。

## 处理覆盖率数据文件
Go 1.20 引入了一个新工具 `covdata`，可用于从 GOCOVERDIR 目录读取和处理覆盖率数据文件。

Go 的 covdata 工具以多种模式运行。covdata 工具调用的一般形式为：

```bash
$ go tool covdata <mode> -i=<dir1,dir2,...> ...flags...
```

其中，`-i` 标志提供了要读取的目录列表，每个目录都来自经过覆盖率插桩的二进制文件的一次执行（通过 GOCOVERDIR）。

### 创建覆盖率分析报告
本节讨论如何使用 `go tool covdata` 从覆盖率数据文件生成人类可读的报告。

#### 报告已覆盖语句的百分比
要为每个经过插桩的包报告“已覆盖语句百分比”指标，请使用命令 `go tool covdata percent -i=<directory>`。使用上面运行部分中的示例：

```bash
$ ls somedata
covcounters.c6de772f99010ef5925877a7b05db4cc.2424989.1670252383678349347
covmeta.c6de772f99010ef5925877a7b05db4cc
$ go tool covdata percent -i=somedata
    main    coverage: 100.0% of statements
    mydomain.com/greetings  coverage: 100.0% of statements
$
```

这里的“已覆盖语句”百分比与 `go test -cover` 报告的百分比直接对应。

#### 转换为传统文本格式
可以使用 covdata 的 textfmt 选项，将二进制覆盖率数据文件转换为 `go test -coverprofile=<outfile>` 生成的传统文本格式。然后，生成的文本文件可用于 `go tool cover -func` 或 `go tool cover -html` 以创建其他报告。例如：

```bash
$ ls somedata
covcounters.c6de772f99010ef5925877a7b05db4cc.2424989.1670252383678349347
covmeta.c6de772f99010ef5925877a7b05db4cc
$ go tool covdata textfmt -i=somedata -o profile.txt
$ cat profile.txt
mode: set
mydomain.com/myprogram.go:10.13,12.2 1 1
mydomain.com/greetings/greetings.go:3.23,5.2 1 1
$ go tool cover -func=profile.txt
mydomain.com/greetings/greetings.go:3:  Goodbye     100.0%
mydomain.com/myprogram.go:10:       main        100.0%
total:                  (statements)    100.0%
$
```

#### 合并
`go tool covdata` 的 merge 子命令可用于合并来自多个数据目录的分析数据。

例如，考虑一个在 macOS 和 Windows 上都运行的程序。该程序的作者可能希望将在每个操作系统上的单独运行得到的覆盖率分析数据合并到一个分析数据集合中，以便生成跨平台的覆盖率摘要。例如：

```bash
$ ls windows_datadir
covcounters.f3833f80c91d8229544b25a855285890.1025623.1667481441036838252
covcounters.f3833f80c91d8229544b25a855285890.1025628.1667481441042785007
covmeta.f3833f80c91d8229544b25a855285890
$ ls macos_datadir
covcounters.b245ad845b5068d116a4e25033b429fb.1025358.1667481440551734165
covcounters.b245ad845b5068d116a4e25033b429fb.1025364.1667481440557770197
covmeta.b245ad845b5068d116a4e25033b429fb
$ ls macos_datadir
$ mkdir merged
$ go tool covdata merge -i=windows_datadir,macos_datadir -o merged
$
```

上面的合并操作将合并来自指定输入目录的数据，并将一组新的合并数据文件写入到 “merged” 目录。

#### 包选择
大多数 `go tool covdata` 命令支持使用 `-pkg` 标志在操作中执行包选择；`-pkg` 的参数形式与 Go 命令的 `-coverpkg` 标志所使用的形式相同。例如：

```bash
$ ls somedata
covcounters.c6de772f99010ef5925877a7b05db4cc.2424989.1670252383678349347
covmeta.c6de772f99010ef5925877a7b05db4cc
$ go tool covdata percent -i=somedata -pkg=mydomain.com/greetings
    mydomain.com/greetings  coverage: 100.0% of statements
$ go tool covdata percent -i=somedata -pkg=nonexistentpackage
$
```

`-pkg` 标志可用于为特定报告选择感兴趣的特定包子集。

## 常见问题

### 如何请求对 go.mod 文件中提到的所有导入包进行覆盖率插桩？
默认情况下，`go build -cover` 会对所有主模块包进行覆盖率插桩，但不会对主模块外部的导入（例如标准库包或 go.mod 中列出的导入）进行插桩。请求对所有非标准库依赖项进行插桩的一种方法是，将 `go list` 的输出提供给 `-coverpkg`。下面再次使用上面引用的示例程序进行说明：

```bash
$ go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' -deps . | paste -sd "," > pkgs.txt
$ go build -o myprogram.exe -coverpkg=`cat pkgs.txt` .
$ mkdir somedata
$ GOCOVERDIR=somedata ./myprogram.exe
$ go tool covdata percent -i=somedata
    golang.org/x/text/internal/tag  coverage: 78.4% of statements
    golang.org/x/text/language  coverage: 35.5% of statements
    mydomain.com    coverage: 100.0% of statements
    mydomain.com/greetings  coverage: 100.0% of statements
    rsc.io/quote    coverage: 25.0% of statements
    rsc.io/sampler  coverage: 86.7% of statements
$
```

### 可以在 GOPATH/GO111MODULE=off 模式下使用 `go build -cover` 吗？
可以，`go build -cover` 在 GO111MODULE=off 模式下确实可以工作。在 GO111MODULE=off 模式下构建程序时，只有在命令行上特别指定为目标的包会被进行覆盖率插桩。使用 `-coverpkg` 标志可以在分析中包含其他包。

### 如果我的程序发生 panic，覆盖率数据会被写入吗？
使用 `go build -cover` 构建的程序只有在程序调用 `os.Exit()` 或从 main.main 正常返回时，才会在执行结束时写出完整的分析数据。如果程序因未恢复的 panic 而终止，或者程序遇到致命异常（例如段错误、除零等），在运行期间执行的语句的分析数据将会丢失。

### `-coverpkg=main` 会选择我的主包进行分析吗？
`-coverpkg` 标志接受导入路径列表，而不是包名列表。如果希望选择主包进行覆盖率插桩，请通过导入路径而不是名称来指定它。例如（使用此示例程序）：

```bash
$ go list -m
mydomain.com
$ go build -coverpkg=main -o oops.exe .
warning: no packages being built depend on matches for pattern main
$ go build -coverpkg=mydomain.com -o myprogram.exe .
$ mkdir somedata
$ GOCOVERDIR=somedata ./myprogram.exe
I say "Hello, world." and "see ya"
$ go tool covdata percent -i=somedata
    mydomain.com    coverage: 100.0% of statements
$
```

## 参考资源
- 介绍 Go 1.2 中单元测试覆盖率的博客文章：
  单元测试覆盖率分析是作为 Go 1.2 版本的一部分引入的；有关详细信息，请参见此博客文章。
- 文档：
  cmd/go 包文档描述了与覆盖率相关的构建和测试标志。
- 技术细节：
  设计草案
  提案

## 术语表
- 单元测试：与特定 Go 包相关联的、在 *_test.go 文件中、利用 Go 的 testing 包的测试。
- 集成测试：针对特定应用程序或二进制文件的更全面、更重量级的测试。集成测试通常包括构建一个程序或一组程序，然后在可能基于或可能不基于 Go 的 testing 包的测试工具的控制下，使用多个输入和场景对这些程序执行一系列运行。







> **[↩ 返回到目录](doc.md)**