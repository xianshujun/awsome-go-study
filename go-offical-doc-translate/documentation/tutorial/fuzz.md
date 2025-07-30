### 教程：开始使用模糊测试（Fuzzing）


> 💡 **小贴士**：如果觉得本节内容较为抽象，可以参考这份通俗易懂的讲解：[模糊测试通俗讲解](fuzz-easydescribe.md)


> **[↩ 返回到目录](../doc.md)**






本教程介绍了 Go 语言中模糊测试的基础知识。模糊测试会向您的测试中输入随机数据，以尝试发现漏洞或导致程序崩溃的输入。模糊测试可以发现的一些漏洞示例包括：SQL 注入、缓冲区溢出、拒绝服务和跨站脚本攻击。

在本教程中，您将为一个简单函数编写模糊测试，运行 `go` 命令，并调试和修复代码中的问题。

如需了解本教程中使用的术语，请参见 Go 模糊测试术语表。

您将依次完成以下章节：
- 创建代码文件夹
- 添加待测试代码
- 添加单元测试
- 添加模糊测试
- 修复两个 bug
- 探索更多资源

> **注意**：有关其他教程，请参见《教程》部分。
>
> **注意**：Go 的模糊测试目前支持部分内置类型，具体列表请参见 [Go 模糊测试文档](https://go.dev/doc/fuzz/)，未来将支持更多内置类型。

---

### 前提条件

- 安装 Go 1.18 或更高版本。安装说明请参见 [安装 Go](https://go.dev/doc/install)。
- 一个用于编辑代码的工具。任何文本编辑器都可以。
- 一个命令行终端。Go 在 Linux 和 Mac 上的任意终端中运行良好，在 Windows 上可使用 PowerShell 或 cmd。
- 一个支持模糊测试的环境。目前，Go 的带覆盖率检测的模糊测试仅支持 AMD64 和 ARM64 架构。

---

### 创建代码文件夹

首先，创建一个文件夹来存放您将编写的代码。

1. 打开命令行终端，并切换到您的主目录。
    - 在 Linux 或 Mac 上：
      ```
      $ cd
      ```
    - 在 Windows 上：
      ```
      C:\> cd %HOMEPATH%
      ```
   教程其余部分将使用 `$` 作为提示符。您使用的命令在 Windows 上也同样适用。

2. 从命令行创建一个名为 `fuzz` 的目录来存放代码：
   ```
   $ mkdir fuzz
   $ cd fuzz
   ```

3. 创建一个模块来管理您的代码。
   运行 `go mod init` 命令，并提供模块路径：
   ```
   $ go mod init example/fuzz
   go: creating new go.mod: module example/fuzz
   ```

> **注意**：在生产代码中，您应指定更符合自身需求的模块路径。详情请参见《管理依赖项》。

接下来，您将添加一个用于反转字符串的简单函数，稍后我们将对其进行模糊测试。

---

### 添加待测试代码

在这一步中，您将添加一个反转字符串的函数。

#### 编写代码

1. 使用文本编辑器，在 `fuzz` 目录中创建一个名为 `main.go` 的文件。

2. 在 `main.go` 文件的顶部，粘贴以下包声明：
   ```go
   package main
   ```
   独立运行的程序（与库相对）总是在 `package main` 中。

3. 在包声明下方，粘贴以下函数声明：
   ```go
   func Reverse(s string) string {
       b := []byte(s)
       for i, j := 0, len(b)-1; i < len(b)/2; i, j = i+1, j-1 {
           b[i], b[j] = b[j], b[i]
       }
       return string(b)
   }
   ```
   此函数接收一个字符串，逐字节循环处理，并在最后返回反转后的字符串。

> **注意**：此代码基于 `golang.org/x/example` 中的 `stringutil.Reverse` 函数。

4. 在 `main.go` 文件顶部、包声明下方，粘贴以下 `main` 函数，用于初始化一个字符串，将其反转，打印输出，然后再次反转：
   ```go
   func main() {
       input := "The quick brown fox jumped over the lazy dog"
       rev := Reverse(input)
       doubleRev := Reverse(rev)
       fmt.Printf("original: %q\n", input)
       fmt.Printf("reversed: %q\n", rev)
       fmt.Printf("reversed again: %q\n", doubleRev)
   }
   ```
   此函数将执行几次 `Reverse` 操作，然后将输出打印到命令行。这有助于观察代码运行效果，也可能对调试有帮助。

5. `main` 函数使用了 `fmt` 包，因此您需要导入它。
   `main.go` 文件的前几行代码应如下所示：
   ```go
   package main

   import "fmt"
   ```

#### 运行代码

在包含 `main.go` 的目录中，从命令行运行代码：
```
$ go run .
original: "The quick brown fox jumped over the lazy dog"
reversed: "god yzal eht revo depmuj xof nworb kciuq ehT"
reversed again: "The quick brown fox jumped over the lazy dog"
```

您可以看到原始字符串、其反转结果，以及再次反转的结果（应与原始字符串相同）。

现在代码已经运行，是时候进行测试了。

---

### 添加单元测试

在这一步中，您将为 `Reverse` 函数编写一个基本的单元测试。

#### 编写代码

1. 使用文本编辑器，在 `fuzz` 目录中创建一个名为 `reverse_test.go` 的文件。

2. 将以下代码粘贴到 `reverse_test.go` 中：
   ```go
   package main

   import (
       "testing"
   )

   func TestReverse(t *testing.T) {
       testcases := []struct {
           in, want string
       }{
           {"Hello, world", "dlrow ,olleH"},
           {" ", " "},
           {"!12345", "54321!"},
       }
       for _, tc := range testcases {
           rev := Reverse(tc.in)
           if rev != tc.want {
               t.Errorf("Reverse: %q, want %q", rev, tc.want)
           }
       }
   }
   ```
   这个简单的测试将断言列出的输入字符串能被正确反转。

#### 运行代码

使用 `go test` 运行单元测试：
```
$ go test
PASS
ok      example/fuzz  0.013s
```

接下来，您将把这个单元测试转换为模糊测试。

---

### 添加模糊测试

单元测试有一些局限性，例如每个输入都必须由开发者手动添加。模糊测试的一个优势是它能自动生成代码的输入，可能发现您编写的测试用例未能覆盖的边界情况。

在本节中，您将把单元测试转换为模糊测试，从而以更少的工作生成更多输入！

> **注意**：您可以在同一个 `*_test.go` 文件中保留单元测试、基准测试和模糊测试，但在此示例中，您将把单元测试转换为模糊测试。

#### 编写代码

在您的文本编辑器中，将 `reverse_test.go` 中的单元测试替换为以下模糊测试代码：
```go
func FuzzReverse(f *testing.F) {
    testcases := []string{"Hello, world", " ", "!12345"}
    for _, tc := range testcases {
        f.Add(tc)  // 使用 f.Add 提供种子语料库（seed corpus）
    }
    f.Fuzz(func(t *testing.T, orig string) {
        rev := Reverse(orig)
        doubleRev := Reverse(rev)
        if orig != doubleRev {
            t.Errorf("Before: %q, after: %q", orig, doubleRev)
        }
        if utf8.ValidString(orig) && !utf8.ValidString(rev) {
            t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
        }
    })
}
```

模糊测试也有一些局限性。在单元测试中，您可以预测 `Reverse` 函数的预期输出，并验证实际输出是否符合预期。

例如，在测试用例 `Reverse("Hello, world")` 中，单元测试明确指定了返回值为 `"dlrow ,olleH"`。

但在模糊测试中，您无法预测预期输出，因为您无法控制输入。

不过，`Reverse` 函数有一些可以验证的属性。本模糊测试检查了以下两个属性：
1. 将字符串反转两次应能恢复原始值。
2. 反转后的字符串应保持有效的 UTF-8 状态。

请注意单元测试与模糊测试在语法上的区别：
- 函数名以 `FuzzXxx` 开头，而不是 `TestXxx`，参数类型为 `*testing.F` 而不是 `*testing.T`。
- 不再使用 `t.Run`，而是使用 `f.Fuzz`，它接收一个模糊目标函数，该函数的参数是 `*testing.T` 和需要模糊的类型。您在单元测试中的输入通过 `f.Add` 作为种子语料库输入提供。

确保已导入新的包 `unicode/utf8`：
```go
package main

import (
    "testing"
    "unicode/utf8"
)
```

现在单元测试已转换为模糊测试，是时候再次运行测试了。

#### 运行代码

1. 首先不进行模糊测试，仅运行模糊测试以确保种子输入通过：
   ```
   $ go test
   PASS
   ok      example/fuzz  0.013s
   ```
   您也可以使用 `go test -run=FuzzReverse` 仅运行模糊测试（如果文件中有其他测试）。

2. 使用 `-fuzz` 标志运行 `FuzzReverse` 进行模糊测试，查看是否能生成导致失败的随机字符串输入。命令如下：
   ```
   $ go test -fuzz=Fuzz
   ```
   另一个有用的标志是 `-fuzztime`，它可以限制模糊测试的时间。例如，`-fuzztime 10s` 表示测试将在 10 秒后退出（前提是未提前发现失败）。详见 [cmd/go 文档](https://pkg.go.dev/cmd/go) 中的测试标志说明。

   现在运行以下命令：
   ```
   $ go test -fuzz=Fuzz
   fuzz: elapsed: 0s, gathering baseline coverage: 0/3 completed
   fuzz: elapsed: 0s, gathering baseline coverage: 3/3 completed, now fuzzing with 8 workers
   fuzz: minimizing 38-byte failing input file...
   --- FAIL: FuzzReverse (0.01s)
       --- FAIL: FuzzReverse (0.00s)
           reverse_test.go:20: Reverse produced invalid UTF-8 string "\x9c\xdd"
       Failing input written to testdata/fuzz/FuzzReverse/af69258a12129d6cbba438df5d5f25ba0ec050461c116f777e77ea7c9a0d217a
       To re-run:
       go test -run=FuzzReverse/af69258a12129d6cbba438df5d5f25ba0ec050461c116f777e77ea7c9a0d217a
   FAIL
   exit status 1
   FAIL    example/fuzz  0.030s
   ```

   模糊测试过程中发现了失败，导致问题的输入已被写入种子语料库文件中。下次运行 `go test` 时，即使没有 `-fuzz` 标志，也会自动运行此失败输入。要查看导致失败的输入，可在文本编辑器中打开 `testdata/fuzz/FuzzReverse` 目录下的语料库文件。您的文件内容可能不同，但格式一致：
   ```
   go test fuzz v1
   string("泃")
   ```
   文件第一行表示编码版本，其后每行代表语料库条目中每个类型的值。由于模糊目标只有一个输入，因此版本后只有一个值。

3. 再次运行 `go test`（不带 `-fuzz` 标志），新的失败语料库条目将被使用：
   ```
   $ go test
   --- FAIL: FuzzReverse (0.00s)
       --- FAIL: FuzzReverse/af69258a12129d6cbba438df5d5f25ba0ec050461c116f777e77ea7c9a0d217a (0.00s)
           reverse_test.go:20: Reverse produced invalid string
   FAIL
   exit status 1
   FAIL    example/fuzz  0.016s
   ```

由于测试失败，现在是时候进行调试了。

---

### 修复无效字符串错误

在本节中，您将调试失败并修复 bug。

您可以先花些时间思考并尝试自己解决问题，然后再继续。

#### 诊断错误

有多种方法可以调试此错误。如果您使用 VS Code 作为编辑器，可以设置调试器进行排查。

在本教程中，我们将通过向终端输出有用的调试信息来进行分析。

首先，查看 `utf8.ValidString` 的文档：
> `ValidString` 报告字符串 `s` 是否完全由有效的 UTF-8 编码的符文（runes）组成。

当前的 `Reverse` 函数是逐字节反转字符串，这正是问题所在。为了保留原始字符串中 UTF-8 编码的符文，我们必须改为逐符文反转。

要检查输入（例如中文字符“泃”）为何在反转后产生无效字符串，可以检查反转字符串中的符文数量。

#### 编写代码

在您的文本编辑器中，将 `FuzzReverse` 中的模糊目标函数替换为以下代码：
```go
f.Fuzz(func(t *testing.T, orig string) {
    rev := Reverse(orig)
    doubleRev := Reverse(rev)
    t.Logf("Number of runes: orig=%d, rev=%d, doubleRev=%d", utf8.RuneCountInString(orig), utf8.RuneCountInString(rev), utf8.RuneCountInString(doubleRev))
    if orig != doubleRev {
        t.Errorf("Before: %q, after: %q", orig, doubleRev)
    }
    if utf8.ValidString(orig) && !utf8.ValidString(rev) {
        t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
    }
})
```
此 `t.Logf` 行将在发生错误或使用 `-v` 标志运行测试时输出到命令行，有助于调试此问题。

#### 运行代码

使用 `go test` 运行测试：
```
$ go test
--- FAIL: FuzzReverse (0.00s)
    --- FAIL: FuzzReverse/28f36ef487f23e6c7a81ebdaa9feffe2f2b02b4cddaa6252e87f69863046a5e0 (0.00s)
        reverse_test.go:16: Number of runes: orig=1, rev=3, doubleRev=1
        reverse_test.go:21: Reverse produced invalid UTF-8 string "\x83\xb3\xe6"
FAIL
exit status 1
FAIL    example/fuzz    0.598s
```

整个种子语料库使用的字符串中每个字符都是单字节。但像“泃”这样的字符可能需要多个字节。因此，逐字节反转会破坏多字节字符的有效性。

> **注意**：如果您对 Go 如何处理字符串感兴趣，请阅读博客文章《[Go 中的字符串、字节、符文和字符](https://blog.golang.org/strings)》以深入了解。

在更好地理解 bug 后，现在可以修正 `Reverse` 函数中的错误。

#### 修复错误

为了修正 `Reverse` 函数，让我们改为按符文遍历字符串，而不是按字节。

##### 编写代码

在您的文本编辑器中，将现有的 `Reverse()` 函数替换为以下代码：
```go
func Reverse(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}
```
关键区别在于，`Reverse` 现在是逐符文迭代字符串，而不是逐字节。注意，这只是一个示例，并未正确处理组合字符。

##### 运行代码

1. 使用 `go test` 运行测试：
   ```
   $ go test
   PASS
   ok      example/fuzz  0.016s
   ```
   测试现在通过了！

2. 再次使用 `go test -fuzz` 进行模糊测试，查看是否还有新 bug：
   ```
   $ go test -fuzz=Fuzz
   fuzz: elapsed: 0s, gathering baseline coverage: 0/37 completed
   fuzz: minimizing 506-byte failing input file...
   fuzz: elapsed: 0s, gathering baseline coverage: 5/37 completed
   --- FAIL: FuzzReverse (0.02s)
       --- FAIL: FuzzReverse (0.00s)
           reverse_test.go:33: Before: "\x91", after: ""
       Failing input written to testdata/fuzz/FuzzReverse/1ffc28f7538e29d79fce69fef20ce5ea72648529a9ca10bea392bcff28cd015c
       To re-run:
       go test -run=FuzzReverse/1ffc28f7538e29d79fce69fef20ce5ea72648529a9ca10bea392bcff28cd015c
   FAIL
   exit status 1
   FAIL    example/fuzz  0.032s
   ```

我们发现字符串在反转两次后与原始值不同。这次输入本身是无效的 Unicode。但如果我们在对字符串进行模糊测试，怎么会这样？

让我们再次调试。

---

### 修复双重反转错误

在本节中，您将调试双重反转失败并修复 bug。

您可以先花些时间思考并尝试自己解决问题，然后再继续。

#### 诊断错误

与之前类似，有多种方法可以调试此失败。在这种情况下，使用调试器是个好方法。

在本教程中，我们将在 `Reverse` 函数中输出有用的调试信息。

仔细观察反转后的字符串以发现错误。在 Go 中，字符串是只读的字节切片，可以包含非有效 UTF-8 的字节。原始字符串是一个包含一个字节 `'\x91'` 的字节切片。当输入字符串被转换为 `[]rune` 时，Go 会将该字节编码为 UTF-8，并将其替换为 UTF-8 字符 ``。当我们比较替换后的 UTF-8 字符与输入字节切片时，它们显然不相等。

#### 编写代码

1. 在您的文本编辑器中，将 `Reverse` 函数替换为以下代码：
   ```go
   func Reverse(s string) string {
       fmt.Printf("input: %q\n", s)
       r := []rune(s)
       fmt.Printf("runes: %q\n", r)
       for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
           r[i], r[j] = r[j], r[i]
       }
       return string(r)
   }
   ```
   这将帮助我们理解将字符串转换为符文切片时发生了什么错误。

#### 运行代码

这次我们只想运行失败的测试以检查日志。为此，使用 `go test -run`。

要运行 `FuzzXxx/testdata` 中的特定语料库条目，可以将 `{FuzzTestName}/{filename}` 作为 `-run` 的参数。这在调试时很有用。复制终端中的唯一哈希值（您的哈希值可能不同）：
```
$ go test -run=FuzzReverse/28f36ef487f23e6c7a81ebdaa9feffe2f2b02b4cddaa6252e87f69863046a5e0
input: "\x91"
runes: ['']
input: ""
runes: ['']
--- FAIL: FuzzReverse (0.00s)
    --- FAIL: FuzzReverse/28f36ef487f23e6c7a81ebdaa9feffe2f2b02b4cddaa6252e87f69863046a5e0 (0.00s)
        reverse_test.go:16: Number of runes: orig=1, rev=1, doubleRev=1
        reverse_test.go:18: Before: "\x91", after: ""
FAIL
exit status 1
FAIL    example/fuzz    0.145s
```

现在我们知道输入是无效的 Unicode，让我们修复 `Reverse` 函数中的错误。

#### 修复错误

为了解决此问题，如果 `Reverse` 的输入不是有效的 UTF-8，则返回一个错误。

##### 编写代码

1. 在您的文本编辑器中，将现有的 `Reverse` 函数替换为以下代码：
   ```go
   func Reverse(s string) (string, error) {
       if !utf8.ValidString(s) {
           return s, errors.New("input is not valid UTF-8")
       }
       r := []rune(s)
       for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
           r[i], r[j] = r[j], r[i]
       }
       return string(r), nil
   }
   ```
   此更改将在输入字符串包含无效 UTF-8 字符时返回错误。

2. 由于 `Reverse` 函数现在返回错误，修改 `main` 函数以忽略额外的错误值。将现有 `main` 函数替换为：
   ```go
   func main() {
       input := "The quick brown fox jumped over the lazy dog"
       rev, revErr := Reverse(input)
       doubleRev, doubleRevErr := Reverse(rev)
       fmt.Printf("original: %q\n", input)
       fmt.Printf("reversed: %q, err: %v\n", rev, revErr)
       fmt.Printf("reversed again: %q, err: %v\n", doubleRev, doubleRevErr)
   }
   ```
   这些 `Reverse` 调用应返回 `nil` 错误，因为输入字符串是有效的 UTF-8。

3. 您需要导入 `errors` 和 `unicode/utf8` 包。`main.go` 中的导入语句应如下：
   ```go
   import (
       "errors"
       "fmt"
       "unicode/utf8"
   )
   ```

4. 修改 `reverse_test.go` 文件，以检查错误并在生成错误时跳过测试：
   ```go
   func FuzzReverse(f *testing.F) {
       testcases := []string{"Hello, world", " ", "!12345"}
       for _, tc := range testcases {
           f.Add(tc)  // 使用 f.Add 提供种子语料库
       }
       f.Fuzz(func(t *testing.T, orig string) {
           rev, err1 := Reverse(orig)
           if err1 != nil {
               return
           }
           doubleRev, err2 := Reverse(rev)
           if err2 != nil {
               return
           }
           if orig != doubleRev {
               t.Errorf("Before: %q, after: %q", orig, doubleRev)
           }
           if utf8.ValidString(orig) && !utf8.ValidString(rev) {
               t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
           }
       })
   }
   ```
   除了 `return`，您也可以调用 `t.Skip()` 来停止该模糊输入的执行。

#### 运行代码

1. 使用 `go test` 运行测试：
   ```
   $ go test
   PASS
   ok      example/fuzz  0.019s
   ```

2. 使用 `go test -fuzz=Fuzz` 进行模糊测试，几秒后按 `Ctrl-C` 停止。如果没有失败，模糊测试将无限运行，可通过 `Ctrl-C` 中断：
   ```
   $ go test -fuzz=Fuzz
   fuzz: elapsed: 0s, gathering baseline coverage: 0/38 completed
   fuzz: elapsed: 0s, gathering baseline coverage: 38/38 completed, now fuzzing with 4 workers
   fuzz: elapsed: 3s, execs: 86342 (28778/sec), new interesting: 2 (total: 35)
   fuzz: elapsed: 6s, execs: 193490 (35714/sec), new interesting: 4 (total: 37)
   ...
   ^Cfuzz: elapsed: 3m48s, execs: 7335316 (31648/sec), new interesting: 8 (total: 41)
   PASS
   ok      example/fuzz  228.000s
   ```

3. 使用 `go test -fuzz=Fuzz -fuzztime 30s` 进行 30 秒模糊测试（若无失败则退出）：
   ```
   $ go test -fuzz=Fuzz -fuzztime 30s
   fuzz: elapsed: 0s, gathering baseline coverage: 0/5 completed
   fuzz: elapsed: 0s, gathering baseline coverage: 5/5 completed, now fuzzing with 4 workers
   ...
   fuzz: elapsed: 30s, execs: 1172817 (30281/sec), new interesting: 17 (total: 17)
   PASS
   ok      example/fuzz  31.025s
   ```
   模糊测试通过了！

除了 `-fuzz` 标志外，`go test` 还新增了多个标志，详见文档。

有关模糊测试输出中术语的更多信息，请参见 [Go 模糊测试文档](https://go.dev/security/fuzz/)。例如，“new interesting” 指扩展了现有模糊测试语料库代码覆盖率的输入。模糊测试开始时，“new interesting” 数量会迅速增加，随着新代码路径的发现而出现几次峰值，然后逐渐减少。

---

### 结论

干得漂亮！您刚刚学习了 Go 语言中的模糊测试。

下一步是选择您代码中的一个函数进行模糊测试并尝试一下！如果模糊测试发现了 bug，可以考虑将其添加到“战利品展示柜”中。

如果遇到问题或有功能建议，请提交 issue。

如需讨论或对功能提供一般反馈，您也可以参与 Gophers Slack 中的 `#fuzzing` 频道。

请访问 [go.dev/security/fuzz](https://go.dev/security/fuzz) 查阅更多文档。

---

### 完整代码

**— main.go —**
```go
package main

import (
    "errors"
    "fmt"
    "unicode/utf8"
)

func main() {
    input := "The quick brown fox jumped over the lazy dog"
    rev, revErr := Reverse(input)
    doubleRev, doubleRevErr := Reverse(rev)
    fmt.Printf("original: %q\n", input)
    fmt.Printf("reversed: %q, err: %v\n", rev, revErr)
    fmt.Printf("reversed again: %q, err: %v\n", doubleRev, doubleRevErr)
}

func Reverse(s string) (string, error) {
    if !utf8.ValidString(s) {
        return s, errors.New("input is not valid UTF-8")
    }
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r), nil
}
```

**— reverse_test.go —**
```go
package main

import (
    "testing"
    "unicode/utf8"
)

func FuzzReverse(f *testing.F) {
    testcases := []string{"Hello, world", " ", "!12345"}
    for _, tc := range testcases {
        f.Add(tc) // 使用 f.Add 提供种子语料库
    }
    f.Fuzz(func(t *testing.T, orig string) {
        rev, err1 := Reverse(orig)
        if err1 != nil {
            return
        }
        doubleRev, err2 := Reverse(rev)
        if err2 != nil {
            return
        }
        if orig != doubleRev {
            t.Errorf("Before: %q, after: %q", orig, doubleRev)
        }
        if utf8.ValidString(orig) && !utf8.ValidString(rev) {
            t.Errorf("Reverse produced invalid UTF-8 string %q", rev)
        }
    })
}
```

---
> **[↩ 返回到目录](../doc.md)**