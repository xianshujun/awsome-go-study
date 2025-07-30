> **[↩ 返回到目录](../doc.md)**

---

### 编写 Web 应用程序

**目录**
*   介绍
*   开始
*   数据结构
*   介绍 net/http 包（一个插曲）
*   使用 net/http 提供 Wiki 页面
*   编辑页面
*   html/template 包
*   处理不存在的页面
*   保存页面
*   错误处理
*   模板缓存
*   验证
*   介绍函数字面量和闭包
*   动手试试！
*   其他任务

---

### 介绍

**本教程涵盖内容：**
*   创建一个带有加载和保存方法的数据结构
*   使用 `net/http` 包构建 Web 应用程序
*   使用 `html/template` 包处理 HTML 模板
*   使用 `regexp` 包验证用户输入
*   使用闭包

**前提知识：**
*   具备编程经验
*   了解基本的 Web 技术（HTTP, HTML）
*   具备一些 UNIX/DOS 命令行知识

---

### 开始

目前，您需要一台 FreeBSD、Linux、macOS 或 Windows 机器来运行 Go。本教程中，我们将使用 `$` 来表示命令提示符。

1.  安装 Go（参见 [安装说明](https://go.dev/doc/install)）。
2.  在您的 `GOPATH` 内为本教程创建一个新目录，并切换到该目录：
    ```
    $ mkdir gowiki
    $ cd gowiki
    ```
3.  创建一个名为 `wiki.go` 的文件，在您喜欢的编辑器中打开它，并添加以下代码行：
    ```go
    package main

    import (
        "fmt"
        "os"
    )
    ```
    我们从 Go 标准库中导入了 `fmt` 和 `os` 包。稍后，当我们实现更多功能时，会向这个导入声明中添加更多包。

---

### 数据结构

让我们从定义数据结构开始。一个 Wiki 由一系列相互链接的页面组成，每个页面都有一个标题和一个正文（页面内容）。在这里，我们将 `Page` 定义为一个结构体，它有两个字段，分别代表标题和正文。

```go
type Page struct {
    Title string
    Body  []byte
}
```
`[]byte` 类型表示“一个字节切片”。（有关切片的更多内容，请参见《[切片：用法和内部原理](https://blog.golang.org/slices-intro)》。）`Body` 元素是 `[]byte` 而不是 `string`，因为这是我们将在下面使用的 `io` 库所期望的类型。

`Page` 结构体描述了页面数据在内存中的存储方式。但持久化存储怎么办？我们可以通过为 `Page` 创建一个 `save` 方法来解决：

```go
func (p *Page) save() error {
    filename := p.Title + ".txt"
    return os.WriteFile(filename, p.Body, 0600)
}
```
这个方法的签名读作：“这是一个名为 `save` 的方法，它以 `p`（一个指向 `Page` 的指针）作为接收者。它不接受任何参数，并返回一个 `error` 类型的值。”

这个方法会将 `Page` 的 `Body` 保存到一个文本文件中。为了简单起见，我们将使用 `Title` 作为文件名。

`save` 方法返回一个 `error` 值，因为 `WriteFile`（一个将字节切片写入文件的标准库函数）的返回类型就是 `error`。`save` 方法返回这个 `error` 值，以便应用程序在写入文件时出现问题时能够处理它。如果一切顺利，`Page.save()` 将返回 `nil`（指针、接口和其他一些类型的零值）。

传递给 `WriteFile` 的八进制整数字面量 `0600` 表示文件应该以仅当前用户可读写的权限创建。（详情请参见 Unix 的 `open(2)` 手册页。）

除了保存页面，我们还需要加载页面：

```go
func loadPage(title string) *Page {
    filename := title + ".txt"
    body, _ := os.ReadFile(filename)
    return &Page{Title: title, Body: body}
}
```
`loadPage` 函数根据 `title` 参数构建文件名，将文件内容读取到一个新变量 `body` 中，并返回一个用正确的 `title` 和 `body` 值构造的 `Page` 字面量的指针。

函数可以返回多个值。标准库函数 `os.ReadFile` 返回 `[]byte` 和 `error`。在 `loadPage` 中，我们还没有处理 `error`；下划线 (`_`) 表示的“空白标识符”被用来丢弃 `error` 返回值（本质上是将该值赋给“无物”）。

但是，如果 `ReadFile` 遇到了错误怎么办？例如，文件可能不存在。我们不应该忽略这样的错误。让我们修改函数，使其返回 `*Page` 和 `error`。

```go
func loadPage(title string) (*Page, error) {
    filename := title + ".txt"
    body, err := os.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}
```
该函数的调用者现在可以检查第二个参数；如果它是 `nil`，那么就成功加载了一个 `Page`。如果不是，它将是一个错误，可以由调用者处理（详情请参见 [语言规范](https://golang.org/ref/spec)）。

至此，我们有了一个简单的数据结构以及保存和加载文件的能力。让我们编写一个 `main` 函数来测试我们所写的内容：

```go
func main() {
    p1 := &Page{Title: "TestPage", Body: []byte("This is a sample Page.")}
    p1.save()
    p2, _ := loadPage("TestPage")
    fmt.Println(string(p2.Body))
}
```
编译并执行此代码后，将创建一个名为 `TestPage.txt` 的文件，其中包含 `p1` 的内容。然后该文件将被读入结构体 `p2`，并将其 `Body` 元素打印到屏幕上。

您可以像这样编译和运行程序：
```
$ go build wiki.go
$ ./wiki
This is a sample Page.
```
（如果您使用的是 Windows，则必须输入 `wiki` 而不带 `./` 来运行程序。）

[点击此处查看我们到目前为止编写的代码](https://go.dev/play/p/381855941)

---

### 介绍 net/http 包（一个插曲）

这是一个简单 Web 服务器的完整工作示例：

```go
//go:build ignore
package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`main` 函数首先调用 `http.HandleFunc`，它告诉 `http` 包使用 `handler` 处理对 Web 根目录 (`"/"`) 的所有请求。

然后它调用 `http.ListenAndServe`，指定它应该在任何接口的 8080 端口上监听 (`":8080"`)。（现在不用担心它的第二个参数 `nil`。）这个函数会一直阻塞，直到程序终止。

`ListenAndServe` 总是返回一个错误，因为它只在发生意外错误时才会返回。为了记录该错误，我们用 `log.Fatal` 包装了函数调用。

`handler` 函数的类型是 `http.HandlerFunc`。它以 `http.ResponseWriter` 和 `http.Request` 作为其参数。

`http.ResponseWriter` 值用于组装 HTTP 服务器的响应；通过向它写入数据，我们将数据发送给 HTTP 客户端。

`http.Request` 是一个表示客户端 HTTP 请求的数据结构。`r.URL.Path` 是请求 URL 的路径组件。末尾的 `[1:]` 意味着“从第 1 个字符到末尾创建 Path 的一个子切片”。这会从路径名中去掉前导的 `/`。

如果您运行此程序并访问 URL：
```
http://localhost:8080/monkeys
```
程序将显示一个包含以下内容的页面：
```
Hi there, I love monkeys!
```

---

### 使用 net/http 提供 Wiki 页面

要使用 `net/http` 包，必须先导入它：
```go
import (
    "fmt"
    "os"
    "log"
    "net/http"
)
```

让我们创建一个处理器 `viewHandler`，它将允许用户查看一个 Wiki 页面。它将处理以 `/view/` 为前缀的 URL。

```go
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, _ := loadPage(title)
    fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.Title, p.Body)
}
```
同样，注意使用 `_` 来忽略 `loadPage` 的错误返回值。这里这样做是为了简化，通常被认为是不好的做法。我们稍后会处理这个问题。

首先，这个函数从 `r.URL.Path`（请求 URL 的路径组件）中提取页面标题。`Path` 通过 `[len("/view/"):]` 被重新切片，以去掉请求路径中前导的 `/view/` 组件。这是因为路径总是以 `/view/` 开头，而这不属于页面标题的一部分。

然后，该函数加载页面数据，用一个简单的 HTML 字符串格式化页面，并将其写入 `w`（`http.ResponseWriter`）。

为了使用这个处理器，我们重写 `main` 函数，使用 `viewHandler` 来处理 `/view/` 路径下的任何请求。

```go
func main() {
    http.HandleFunc("/view/", viewHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

[点击此处查看我们到目前为止编写的代码](https://go.dev/play/p/623208925)

让我们创建一些页面数据（作为 `test.txt`），编译我们的代码，并尝试提供一个 Wiki 页面。

在编辑器中打开 `test.txt` 文件，并保存字符串 `"Hello world"`（不带引号）。

```
$ go build wiki.go
$ ./wiki
```
（如果您使用的是 Windows，则必须输入 `wiki` 而不带 `./` 来运行程序。）

运行此 Web 服务器后，访问 `http://localhost:8080/view/test` 应该会显示一个标题为 "test" 的页面，其中包含 "Hello world" 这几个字。

---

### 编辑页面

一个没有编辑页面能力的 Wiki 就不是真正的 Wiki。让我们创建两个新的处理器：一个名为 `editHandler` 的处理器来显示一个“编辑页面”表单，另一个名为 `saveHandler` 的处理器来保存通过表单输入的数据。

首先，我们将它们添加到 `main()` 中：
```go
func main() {
    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`editHandler` 函数加载页面（如果不存在，则创建一个空的 `Page` 结构体），并显示一个 HTML 表单。

```go
func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    fmt.Fprintf(w, "<h1>Editing %s</h1>"+
        "<form action=\"/save/%s\" method=\"POST\">"+
        "<textarea name=\"body\">%s</textarea><br>"+
        "<input type=\"submit\" value=\"Save\">"+
        "</form>",
        p.Title, p.Title, p.Body)
}
```
这个函数可以正常工作，但所有这些硬编码的 HTML 很丑陋。当然，有更好的方法。

---

### html/template 包

`html/template` 包是 Go 标准库的一部分。我们可以使用 `html/template` 将 HTML 保存在单独的文件中，这样我们就可以在不修改底层 Go 代码的情况下更改编辑页面的布局。

首先，我们必须将 `html/template` 添加到导入列表中。我们也不再使用 `fmt` 了，所以必须将其移除。
```go
import (
    "html/template"
    "os"
    "net/http"
)
```

让我们创建一个包含 HTML 表单的模板文件。打开一个名为 `edit.html` 的新文件，并添加以下行：
```html
<h1>Editing {{.Title}}</h1>
<form action="/save/{{.Title}}" method="POST">
<div><textarea name="body" rows="20" cols="80">{{printf "%s" .Body}}</textarea></div>
<div><input type="submit" value="Save"></div>
</form>
```

修改 `editHandler` 以使用模板，而不是硬编码的 HTML：
```go
func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    t, _ := template.ParseFiles("edit.html")
    t.Execute(w, p)
}
```
`template.ParseFiles` 函数将读取 `edit.html` 的内容并返回一个 `*template.Template`。

`Execute` 方法执行模板，将生成的 HTML 写入 `http.ResponseWriter`。`.Title` 和 `.Body` 这些带点的标识符指的是 `p.Title` 和 `p.Body`。

模板指令用双大括号包围。`{{printf "%s" .Body}}` 指令是一个函数调用，它将 `.Body` 作为字符串而不是字节流输出，与调用 `fmt.Printf` 相同。`html/template` 包有助于保证模板操作生成的 HTML 是安全且正确的。例如，它会自动转义任何大于号 (`>`)，将其替换为 `&gt;`，以确保用户数据不会破坏表单的 HTML。

既然我们现在使用模板了，让我们为 `viewHandler` 创建一个名为 `view.html` 的模板：
```html
<h1>{{.Title}}</h1>
<p>[<a href="/edit/{{.Title}}">edit</a>]</p>
<div>{{printf "%s" .Body}}</div>
```

相应地修改 `viewHandler`：
```go
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, _ := loadPage(title)
    t, _ := template.ParseFiles("view.html")
    t.Execute(w, p)
}
```

请注意，我们在两个处理器中使用了几乎完全相同的模板代码。让我们通过将模板代码移动到自己的函数中来消除这种重复：
```go
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, _ := template.ParseFiles(tmpl + ".html")
    t.Execute(w, p)
}
```

然后修改处理器以使用该函数：
```go
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, _ := loadPage(title)
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/edit/"):]
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}
```

如果我们注释掉 `main` 中未实现的 `save` 处理器的注册，我们就可以再次编译和测试我们的程序了。

[点击此处查看我们到目前为止编写的代码](https://go.dev/play/p/584292007)

---

### 处理不存在的页面

如果你访问 `/view/APageThatDoesntExist` 会怎样？你会看到一个包含 HTML 的页面。这是因为它忽略了 `loadPage` 的错误返回值，并继续尝试用无数据填充模板。相反，如果请求的 `Page` 不存在，它应该将客户端重定向到编辑页面，以便可以创建内容：
```go
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/view/"):]
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}
```
`http.Redirect` 函数向 HTTP 响应添加一个 HTTP 状态码 `http.StatusFound` (302) 和一个 `Location` 头。

---

### 保存页面

`saveHandler` 函数将处理位于编辑页面上的表单提交。在取消注释 `main` 中的相关行后，让我们来实现这个处理器：
```go
func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    p.save()
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
```
页面标题（在 URL 中提供）和表单的唯一字段 `Body` 被存储在一个新的 `Page` 中。然后调用 `save()` 方法将数据写入文件，并将客户端重定向到 `/view/` 页面。

`FormValue` 返回的值是 `string` 类型。我们必须在它能放入 `Page` 结构体之前将其转换为 `[]byte`。我们使用 `[]byte(body)` 来执行转换。

---

### 错误处理

我们的程序中有几处忽略了错误。这是一种不好的做法，因为当错误发生时，程序会有意想不到的行为。更好的解决方案是处理这些错误并向用户返回错误消息。这样，如果出了问题，服务器会按照我们期望的方式运行，用户也会收到通知。

首先，让我们处理 `renderTemplate` 中的错误：
```go
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    t, err := template.ParseFiles(tmpl + ".html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    err = t.Execute(w, p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```
`http.Error` 函数发送一个指定的 HTTP 响应码（在这种情况下是“内部服务器错误”）和错误消息。将此操作放入一个单独的函数的决策现在开始显现出好处了。

现在让我们修复 `saveHandler`：
```go
func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[len("/save/"):]
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
```
`p.save()` 期间发生的任何错误都将报告给用户。

---

### 模板缓存

这段代码有一个低效之处：`renderTemplate` 每次渲染页面时都会调用 `ParseFiles`。更好的方法是在程序初始化时只调用一次 `ParseFiles`，将所有模板解析到一个单一的 `*Template` 中。然后我们可以使用 `ExecuteTemplate` 方法来渲染一个特定的模板。

首先，我们创建一个名为 `templates` 的全局变量，并用 `ParseFiles` 初始化它。
```go
var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
```
`template.Must` 函数是一个便利的包装器，当传入一个非 `nil` 的错误值时会引发 panic，否则原样返回 `*Template`。在这里引发 panic 是合适的；如果模板无法加载，唯一明智的做法就是退出程序。

`ParseFiles` 函数接受任意数量的字符串参数，这些参数标识我们的模板文件，并将这些文件解析成以基本文件名命名的模板。如果我们向程序中添加更多模板，我们将把它们的名字添加到 `ParseFiles` 调用的参数中。

然后我们修改 `renderTemplate` 函数，以 `templates.ExecuteTemplate` 方法调用适当的模板：
```go
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}
```
请注意，模板名称是模板文件名，因此我们必须将 `".html"` 附加到 `tmpl` 参数上。

---

### 验证

正如您可能已经观察到的，这个程序有一个严重的安全漏洞：用户可以提供服务器上任意路径进行读写。为了缓解这个问题，我们可以编写一个函数，使用正则表达式来验证标题。

首先，将 `"regexp"` 添加到导入列表中。然后我们可以创建一个全局变量来存储我们的验证表达式：
```go
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")
```
`regexp.MustCompile` 函数将解析并编译正则表达式，并返回一个 `regexp.Regexp`。`MustCompile` 与 `Compile` 的区别在于，如果表达式编译失败，`MustCompile` 会引发 panic，而 `Compile` 会将错误作为第二个参数返回。

现在，让我们编写一个函数，使用 `validPath` 表达式来验证路径并提取页面标题：
```go
func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil {
        http.NotFound(w, r)
        return "", errors.New("invalid Page Title")
    }
    return m[2], nil // 标题是第二个子表达式。
}
```
如果标题有效，它将与一个 `nil` 错误值一起返回。如果标题无效，该函数将向 HTTP 连接写入一个“404 未找到”错误，并向处理器返回一个错误。为了创建一个新错误，我们必须导入 `errors` 包。

让我们在每个处理器中调用 `getTitle`：
```go
func viewHandler(w http.ResponseWriter, r *http.Request) {
    title, err := getTitle(w, r)
    if err != nil {
        return
    }
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    title, err := getTitle(w, r)
    if err != nil {
        return
    }
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    title, err := getTitle(w, r)
    if err != nil {
        return
    }
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err = p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
```

---

### 介绍函数字面量和闭包

在每个处理器中捕获错误条件引入了大量的重复代码。如果我们能用一个函数来包装每个处理器，进行这种验证和错误检查，那会怎么样？Go 的函数字面量提供了一种强大的抽象功能，可以帮助我们实现这一点。

首先，我们重写每个处理器的函数定义，以接受一个字符串标题：
```go
func viewHandler(w http.ResponseWriter, r *http.Request, title string)
func editHandler(w http.ResponseWriter, r *http.Request, title string)
func saveHandler(w http.ResponseWriter, r *http.Request, title string)
```

现在，我们定义一个包装函数，它接受上述类型的函数，并返回一个 `http.HandlerFunc` 类型的函数（适合传递给 `http.HandleFunc` 函数）：
```go
func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 在这里，我们将从 Request 中提取页面标题，
        // 并调用提供的处理器 'fn'
    }
}
```
返回的函数被称为**闭包**，因为它封闭了其外部定义的值。在这种情况下，`makeHandler` 的单个参数 `fn` 被闭包所封闭。变量 `fn` 将是我们的 `save`、`edit` 或 `view` 处理器之一。

现在，我们可以从 `getTitle` 中取出代码，并在这里使用它（做了一些小的修改）：
```go
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        m := validPath.FindStringSubmatch(r.URL.Path)
        if m == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, m[2])
    }
}
```
`makeHandler` 返回的闭包是一个接受 `http.ResponseWriter` 和 `http.Request` 的函数（换句话说，就是一个 `http.HandlerFunc`）。闭包从请求路径中提取标题，并使用 `validPath` 正则表达式对其进行验证。如果标题无效，将使用 `http.NotFound` 函数向 `ResponseWriter` 写入一个错误。如果标题有效，封闭的处理器函数 `fn` 将使用 `ResponseWriter`、`Request` 和标题作为参数被调用。

现在，我们可以在 `main` 中，在将处理器注册到 `http` 包之前，用 `makeHandler` 包装它们：
```go
func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

最后，我们从处理器函数中移除对 `getTitle` 的调用，使它们变得简单得多：
```go
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    p, err := loadPage(title)
    if err != nil {
        p = &Page{Title: title}
    }
    renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    p := &Page{Title: title, Body: []byte(body)}
    err := p.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
```

---

### 动手试试！

[点击此处查看最终的代码列表](https://go.dev/play/p/584292007)

重新编译代码并运行应用程序：
```
$ go build wiki.go
$ ./wiki
```
访问 `http://localhost:8080/view/ANewPage` 应该会向您显示页面编辑表单。然后，您应该能够输入一些文本，单击“保存”，并被重定向到新创建的页面。

---

### 其他任务

以下是一些您可能想自己尝试的简单任务：
*   将模板存储在 `tmpl/` 目录中，将页面数据存储在 `data/` 目录中。
*   添加一个处理器，使 Web 根目录重定向到 `/view/FrontPage`。
*   通过使页面模板成为有效的 HTML 并添加一些 CSS 规则来美化页面。
*   通过将 `[PageName]` 的实例转换为 `<a href="/view/PageName">PageName</a>` 来实现页面间链接。（提示：您可以使用 `regexp.ReplaceAllFunc` 来实现这一点）

> **[↩ 返回到目录](../doc.md)**