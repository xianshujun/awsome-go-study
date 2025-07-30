# 教程：开始使用泛型

> **[↩ 返回到目录](../doc.md)**
## 目录

- 先决条件
- 为代码创建文件夹
- 添加非泛型函数
- 添加一个能处理多种类型的泛型函数
- 调用泛型函数时省略类型参数
- 声明一个类型约束
- 结语
- 完整代码

---

本教程将介绍 Go 语言中泛型的基础知识。通过泛型，你可以声明并使用那些能够处理调用方提供的任意一组类型的函数或类型。

在本教程中，你将先声明两个简单的非泛型函数，然后用一个泛型函数来统一实现它们的逻辑。

你将依次完成以下步骤：

- 为代码创建文件夹
- 添加非泛型函数
- 添加一个能处理多种类型的泛型函数
- 调用泛型函数时省略类型参数
- 声明一个类型约束

> **提示**：更多教程请参见 [Go 官方教程列表](https://go.dev/doc/tutorial/)。

> **提示**：你也可以使用 Go Playground 的“Go dev branch”模式来编辑和运行程序。

---

### 先决条件

- 安装了 Go 1.18 或更高版本。安装方法请参考 [安装 Go](https://go.dev/doc/install)。
- 一个代码编辑器。任何文本编辑器都可以。
- 一个命令行终端。Linux 和 Mac 用户可使用任意终端，Windows 用户可使用 PowerShell 或 cmd。

---

### 为代码创建文件夹

首先，创建一个文件夹来存放你的代码。

1. 打开终端，进入你的主目录。

   在 Linux 或 Mac 上：

   ```bash
   $ cd
   ```

   在 Windows 上：

   ```cmd
   C:\> cd %HOMEPATH%
   ```

   后续教程中的命令提示符统一使用 `$`，但这些命令在 Windows 上同样适用。

2. 创建一个名为 `generics` 的目录：

   ```bash
   $ mkdir generics
   $ cd generics
   ```

3. 创建一个模块来管理你的代码依赖：

   ```bash
   $ go mod init example/generics
   go: creating new go.mod: module example/generics
   ```

   > **提示**：在生产环境中，你应该使用更具体的模块路径。详情请参考 [管理依赖](https://go.dev/ref/mod#module-path)。

接下来，我们将添加一些简单的代码来操作映射（map）。

---

### 添加非泛型函数

在这一步，你将添加两个函数，它们分别用于计算一个映射中所有值的总和。

之所以要声明两个函数而不是一个，是因为你要处理两种不同类型的映射：一种存储 `int64` 值，另一种存储 `float64` 值。

#### 编写代码

1. 在 `generics` 目录下，用编辑器创建一个名为 `main.go` 的文件。

2. 在文件顶部添加包声明：

   ```go
   package main
   ```

   所有独立运行的 Go 程序（非库）都属于 `main` 包。

3. 在包声明下方，添加以下两个函数：

   ```go
   // SumInts 计算 map 中所有 int64 值的总和
   func SumInts(m map[string]int64) int64 {
       var s int64
       for _, v := range m {
           s += v
       }
       return s
   }

   // SumFloats 计算 map 中所有 float64 值的总和
   func SumFloats(m map[string]float64) float64 {
       var s float64
       for _, v := range m {
           s += v
       }
       return s
   }
   ```

   说明：

    - 两个函数功能相同，只是处理的数据类型不同；
    - `SumInts` 接收 `map[string]int64` 类型的参数；
    - `SumFloats` 接收 `map[string]float64` 类型的参数。

4. 在 `main.go` 文件中，添加 `main` 函数来初始化数据并调用上述函数：

   ```go
   func main() {
       // 初始化一个存储 int64 值的映射
       ints := map[string]int64{
           "first":  34,
           "second": 12,
       }
       // 初始化一个存储 float64 值的映射
       floats := map[string]float64{
           "first":  35.98,
           "second": 26.99,
       }

       fmt.Printf("非泛型求和: %v 和 %v\n",
           SumInts(ints),
           SumFloats(floats))
   }
   ```

5. 在包声明下方，导入所需的包：

   ```go
   import "fmt"
   ```

6. 保存 `main.go`。

#### 运行代码

在 `main.go` 所在目录下运行：

```bash
$ go run .
```

输出应为：

```
非泛型求和: 46 和 62.97
```

使用泛型，你可以用一个函数代替这两个函数。接下来，我们将创建一个能同时处理整数和浮点数的泛型函数。

---

### 添加一个能处理多种类型的泛型函数

在这一步，你将创建一个**泛型函数**，它可以接收包含整数或浮点数的映射，从而替代之前写的两个函数。

为了支持多种类型，这个函数需要一种方式来声明它支持哪些类型。同时，调用它的代码也需要一种方式来指明是用整数还是浮点数映射。

为此，你将编写一个带有**类型参数**（type parameters）的函数。这些类型参数使得函数具有“泛型”特性，能够处理不同类型的参数。

你将通过**类型参数**和**普通函数参数**来调用这个函数。

每个类型参数都有一个**类型约束**（type constraint），它相当于类型参数的“元类型”。约束规定了调用代码可以使用的具体类型。

虽然类型约束通常代表一组类型，但在编译时，类型参数代表的是**单一类型**——即调用代码传入的类型。如果传入的类型不被约束允许，代码将无法编译。

记住：类型参数必须支持泛型代码中对它执行的所有操作。例如，如果你的函数试图对一个数值类型执行字符串操作（如索引），代码将无法通过编译。

在接下来的代码中，你将使用一个允许 `int64` 或 `float64` 类型的约束。

#### 编写代码

1. 在之前两个函数下方，添加以下泛型函数：

   ```go
   // SumIntsOrFloats 计算 map 中所有值的总和，支持 int64 和 float64 类型
   func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
       var s V
       for _, v := range m {
           s += v
       }
       return s
   }
   ```

   说明：

    - 函数有两个类型参数：`K` 和 `V`，写在方括号 `[]` 内；
    - `K` 的约束是 `comparable`，这是 Go 内置的预定义约束，表示该类型可以用于 `==` 和 `!=` 比较。由于映射的键必须是可比较的，所以 `K` 必须是 `comparable`；
    - `V` 的约束是 `int64 | float64`，表示 `V` 可以是 `int64` 或 `float64` 中的任意一种（`|` 表示类型联合）；
    - 函数参数 `m` 的类型是 `map[K]V`，即键为 `K`、值为 `V` 的映射；
    - 返回值类型也是 `V`。

2. 在 `main` 函数中，添加以下代码来调用泛型函数：

   ```go
   fmt.Printf("泛型求和: %v 和 %v\n",
       SumIntsOrFloats[string, int64](ints),
       SumIntsOrFloats[string, float64](floats))
   ```

   说明：

    - 调用泛型函数时，需要在方括号中指定**类型参数**（type arguments），告诉编译器 `K` 和 `V` 应该是什么类型；
    - 例如，`[string, int64]` 表示 `K` 是 `string`，`V` 是 `int64`；
    - 如下节所述，很多时候你可以省略这些类型参数，Go 能自动推断。

#### 运行代码

```bash
$ go run .
```

输出应为：

```
非泛型求和: 46 和 62.97
泛型求和: 46 和 62.97
```

Go 编译器在每次调用时，会用具体的类型替换泛型函数中的类型参数。

---

### 调用泛型函数时省略类型参数

在这一步，你将简化调用方式，**省略类型参数**。在很多情况下，Go 编译器可以根据函数参数的类型自动推断出类型参数。

> **注意**：这并不总是可行。例如，如果泛型函数没有参数，你就必须显式写出类型参数。

#### 编写代码

在 `main` 函数中添加：

```go
fmt.Printf("泛型求和，类型参数已推断: %v 和 %v\n",
    SumIntsOrFloats(ints),
    SumIntsOrFloats(floats))
```

这里我们省略了 `[string, int64]` 和 `[string, float64]`，因为 Go 可以从 `ints` 和 `floats` 的类型推断出这些信息。

#### 运行代码

```bash
$ go run .
```

输出应为：

```
非泛型求和: 46 和 62.97
泛型求和: 46 和 62.97
泛型求和，类型参数已推断: 46 和 62.97
```

---

### 声明一个类型约束

在最后一部分，你将把之前定义的类型联合（`int64 | float64`）提取到一个**接口**中，作为一个可复用的类型约束。这样可以让代码更简洁，尤其是当约束变得复杂时。

你可以将类型约束声明为一个接口。该约束允许任何实现该接口的类型。接口中也可以直接包含具体类型，就像下面这样。

#### 编写代码

1. 在 `import` 语句之后、`main` 函数之前，添加以下接口定义：

   ```go
   // Number 是一个类型约束，表示 int64 或 float64
   type Number interface {
       int64 | float64
   }
   ```

   这相当于把 `int64 | float64` 封装成了一个名为 `Number` 的“类型集合”。

2. 在已有函数下方，添加新的泛型函数：

   ```go
   // SumNumbers 计算 map 中所有值的总和，支持整数和浮点数
   func SumNumbers[K comparable, V Number](m map[K]V) V {
       var s V
       for _, v := range m {
           s += v
       }
       return s
   }
   ```

   这个函数逻辑与之前相同，但使用了 `Number` 接口作为 `V` 的约束。

3. 在 `main` 函数中调用新函数：

   ```go
   fmt.Printf("带约束的泛型求和: %v 和 %v\n",
       SumNumbers(ints),
       SumNumbers(floats))
   ```

   再次省略了类型参数，由 Go 自动推断。

#### 运行代码

```bash
$ go run .
```

输出应为：

```
非泛型求和: 46 和 62.97
泛型求和: 46 和 62.97
泛型求和，类型参数已推断: 46 和 62.97
带约束的泛型求和: 46 和 62.97
```

---

### 结语

🎉 干得漂亮！你已经成功入门了 Go 的泛型编程！

#### 建议后续学习：

- [Go 语言之旅](https://go.dev/tour/) 是学习 Go 基础的绝佳互动教程。
- 《[Effective Go](https://go.dev/doc/effective_go)》和《[如何编写 Go 代码](https://go.dev/doc/code)》中包含了大量实用的 Go 最佳实践。

---

### 完整代码

以下是本教程的完整代码：

```go
package main

import "fmt"

// Number 是一个类型约束，表示 int64 或 float64
type Number interface {
    int64 | float64
}

func main() {
    // 初始化一个存储 int64 值的映射
    ints := map[string]int64{
        "first":  34,
        "second": 12,
    }
    // 初始化一个存储 float64 值的映射
    floats := map[string]float64{
        "first":  35.98,
        "second": 26.99,
    }

    fmt.Printf("非泛型求和: %v 和 %v\n",
        SumInts(ints),
        SumFloats(floats))

    fmt.Printf("泛型求和: %v 和 %v\n",
        SumIntsOrFloats[string, int64](ints),
        SumIntsOrFloats[string, float64](floats))

    fmt.Printf("泛型求和，类型参数已推断: %v 和 %v\n",
        SumIntsOrFloats(ints),
        SumIntsOrFloats(floats))

    fmt.Printf("带约束的泛型求和: %v 和 %v\n",
        SumNumbers(ints),
        SumNumbers(floats))
}

// SumInts 计算 map 中所有 int64 值的总和
func SumInts(m map[string]int64) int64 {
    var s int64
    for _, v := range m {
        s += v
    }
    return s
}

// SumFloats 计算 map 中所有 float64 值的总和
func SumFloats(m map[string]float64) float64 {
    var s float64
    for _, v := range m {
        s += v
    }
    return s
}

// SumIntsOrFloats 计算 map 中所有值的总和，支持 int64 和 float64 类型
func SumIntsOrFloats[K comparable, V int64 | float64](m map[K]V) V {
    var s V
    for _, v := range m {
        s += v
    }
    return s
}

// SumNumbers 计算 map 中所有值的总和，支持整数和浮点数
func SumNumbers[K comparable, V Number](m map[K]V) V {
    var s V
    for _, v := range m {
        s += v
    }
    return s
}
```
> **[↩ 返回到目录](../doc.md)**