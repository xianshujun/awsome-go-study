# 教程：使用 Go 和 Gin 开发一个 RESTful API

> **[↩ 返回到目录](../doc.md)**
## 目录

- 先决条件
- 设计 API 接口
- 为代码创建文件夹
- 创建数据
- 编写处理函数以返回所有条目
- 编写处理函数以添加新条目
- 编写处理函数以返回特定条目
- 结语
- 完整代码

---

本教程将带你了解如何使用 Go 语言和 Gin Web 框架（简称 Gin）来编写一个 RESTful Web 服务 API 的基础知识。

如果你对 Go 语言及其工具链有一定基础，那么你将能从本教程中收获最多。如果你是第一次接触 Go，建议先阅读 [《开始使用 Go》](https://go.dev/doc/tutorial/getting-started) 教程，快速入门。

Gin 简化了构建 Web 应用（包括 Web 服务）中的许多编码任务。在本教程中，你将使用 Gin 来实现请求路由、获取请求信息，并将数据以 JSON 格式返回。

你将构建一个提供两个接口的 RESTful API 服务器。示例项目是一个关于经典爵士唱片的数据仓库。

教程包含以下部分：

- 设计 API 接口
- 为代码创建文件夹
- 创建数据
- 编写处理函数以返回所有条目
- 编写处理函数以添加新条目
- 编写处理函数以返回特定条目

> **提示**：更多教程请参见 [Go 官方教程列表](https://go.dev/doc/tutorial/)。

如果你想在 Google Cloud Shell 中体验交互式教程，可点击下方按钮：

[在 Cloud Shell 中打开](https://console.cloud.google.com/cloudshell/open?git_repo=https://github.com/golang/tour&page=editor&open_in_editor=web-service-gin/main.go)

---

### 先决条件

- 安装 Go 1.16 或更高版本。安装方法请参考 [安装 Go](https://go.dev/doc/install)。
- 一个代码编辑器。任何文本编辑器都可以。
- 一个命令行终端。Linux 和 Mac 用户可使用任意终端，Windows 用户可使用 PowerShell 或 cmd。
- `curl` 工具。Linux 和 Mac 系统通常已自带。Windows 10 Insider Build 17063 及以上版本也已包含。更早版本的 Windows 可能需要手动安装。详情请参考 [Tar 和 Curl 来到 Windows](https://docs.microsoft.com/en-us/virtualization/community/team-blog/2017/20171219-tar-and-curl-come-to-windows)。

> 本教程要求 Go 1.16 或更高版本。

---

### 设计 API 接口

你将构建一个 API，用于访问一个售卖经典黑胶唱片的商店。因此，你需要提供一些接口，让客户端可以获取和添加专辑信息。

开发 API 时，通常第一步是**设计接口**。如果你的接口设计清晰易懂，使用者会更容易上手。

本教程中，你将创建以下接口：

#### `/albums`
- **GET**：获取所有专辑列表，以 JSON 格式返回。
- **POST**：接收客户端发送的 JSON 数据，添加一张新专辑。

#### `/albums/:id`
- **GET**：根据专辑的 ID 获取特定专辑，返回该专辑的 JSON 数据。

接下来，我们将为代码创建一个文件夹。

---

### 为代码创建文件夹

首先，创建一个项目目录来存放你的代码。

1. 打开终端，进入你的主目录。

   Linux 或 Mac：

   ```bash
   $ cd
   ```

   Windows：

   ```cmd
   C:\> cd %HOMEPATH%
   ```

   后续教程中的命令提示符统一使用 `$`，但这些命令在 Windows 上同样适用。

2. 创建一个名为 `web-service-gin` 的目录：

   ```bash
   $ mkdir web-service-gin
   $ cd web-service-gin
   ```

3. 初始化模块，用于管理依赖：

   ```bash
   $ go mod init example/web-service-gin
   go: creating new go.mod: module example/web-service-gin
   ```

   这条命令会创建一个 `go.mod` 文件，用于记录你项目所依赖的模块。关于模块路径的命名规则，详见 [管理依赖](https://go.dev/ref/mod#module-path)。

接下来，我们将设计数据结构来处理数据。

---

### 创建数据

为了简化本教程，数据将存储在内存中。在实际项目中，API 通常会与数据库交互。

> **注意**：内存存储意味着每次停止服务器时数据都会丢失，重启后会重新初始化。

#### 编写代码

1. 在 `web-service-gin` 目录下，用编辑器创建一个名为 `main.go` 的文件。

2. 在文件顶部添加包声明：

   ```go
   package main
   ```

   所有独立运行的 Go 程序（非库）都属于 `main` 包。

3. 在包声明下方，添加专辑结构体定义：

   ```go
   // album 表示一张唱片专辑的信息
   type album struct {
       ID     string  `json:"id"`
       Title  string  `json:"title"`
       Artist string  `json:"artist"`
       Price  float64 `json:"price"`
   }
   ```

   结构体标签（如 `json:"artist"`）用于指定结构体字段在序列化为 JSON 时的字段名。如果没有这些标签，JSON 会使用大写的字段名（如 `Artist`），这在 JSON 中并不常见。

4. 在结构体下方，添加一个包含初始数据的切片：

   ```go
   // albums 切片用于存放初始的专辑数据
   var albums = []album{
       {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
       {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
       {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
   }
   ```

接下来，我们将编写第一个接口的处理逻辑。

---

### 编写处理函数以返回所有条目

当客户端访问 `GET /albums` 时，你希望返回所有专辑的 JSON 数据。

为此，你需要：

- 编写返回数据的逻辑；
- 将请求路径映射到该逻辑。

> 注意：代码书写顺序与运行顺序相反，我们先写依赖，再写调用它的代码。

#### 编写代码

1. 在结构体代码下方，添加 `getAlbums` 函数：

   ```go
   // getAlbums 返回所有专辑的 JSON 列表
   func getAlbums(c *gin.Context) {
       c.IndentedJSON(http.StatusOK, albums)
   }
   ```

   说明：

    - 函数接收一个 `*gin.Context` 参数。函数名可以任意，Gin 不强制命名规则。
    - `gin.Context` 是 Gin 的核心，它封装了请求信息、JSON 序列化、响应写入等功能（注意：它与 Go 内置的 `context` 包不同）。
    - `c.IndentedJSON` 将 `albums` 切片序列化为格式化的 JSON 并写入响应。
    - 第一个参数是 HTTP 状态码，`http.StatusOK` 表示 200 OK。
    - 你也可以使用 `c.JSON` 生成更紧凑的 JSON，但 `IndentedJSON` 更适合调试。

2. 在 `main.go` 文件中，`albums` 变量声明下方，添加 `main` 函数：

   ```go
   func main() {
       router := gin.Default()
       router.GET("/albums", getAlbums)
       router.Run("localhost:8080")
   }
   ```

   说明：

    - `gin.Default()` 创建一个默认的路由引擎；
    - `router.GET` 将 `GET /albums` 请求映射到 `getAlbums` 函数；
    - `router.Run` 启动服务器，监听 `localhost:8080`。

3. 在文件顶部，`package main` 下方，导入所需包：

   ```go
   import (
       "net/http"
       "github.com/gin-gonic/gin"
   )
   ```

4. 保存 `main.go`。

#### 运行代码

1. 添加 Gin 模块依赖：

   ```bash
   $ go get .
   go get: added github.com/gin-gonic/gin v1.7.2
   ```

2. 运行程序：

   ```bash
   $ go run .
   ```

3. 在另一个终端窗口，使用 `curl` 发起请求：

   ```bash
   $ curl http://localhost:8080/albums
   ```

   输出应为：

   ```json
   [
       {
           "id": "1",
           "title": "Blue Train",
           "artist": "John Coltrane",
           "price": 56.99
       },
       {
           "id": "2",
           "title": "Jeru",
           "artist": "Gerry Mulligan",
           "price": 17.99
       },
       {
           "id": "3",
           "title": "Sarah Vaughan and Clifford Brown",
           "artist": "Sarah Vaughan",
           "price": 39.99
       }
   ]
   ```

🎉 恭喜！你已经启动了一个 API！接下来，我们将添加一个用于创建新专辑的接口。

---

### 编写处理函数以添加新条目

当客户端向 `/albums` 发起 `POST` 请求时，你希望将请求体中的 JSON 数据解析为一张新专辑，并添加到现有数据中。

#### 编写代码

1. 添加 `postAlbums` 函数（可放在文件末尾）：

   ```go
   // postAlbums 从请求体的 JSON 中解析并添加一张新专辑
   func postAlbums(c *gin.Context) {
       var newAlbum album
       // 使用 BindJSON 将请求体绑定到 newAlbum
       if err := c.BindJSON(&newAlbum); err != nil {
           return
       }
       // 将新专辑添加到切片
       albums = append(albums, newAlbum)
       c.IndentedJSON(http.StatusCreated, newAlbum)
   }
   ```

   说明：

    - `c.BindJSON(&newAlbum)` 自动将 JSON 数据解析并填充到 `newAlbum` 结构体；
    - `append` 将新专辑加入切片；
    - 返回 `201 Created` 状态码和新专辑的 JSON。

2. 修改 `main` 函数，添加 POST 路由：

   ```go
   func main() {
       router := gin.Default()
       router.GET("/albums", getAlbums)
       router.POST("/albums", postAlbums)
       router.Run("localhost:8080")
   }
   ```

#### 运行代码

1. 停止之前的服务器（按 `Ctrl+C`）。
2. 重新运行：

   ```bash
   $ go run .
   ```

3. 使用 `curl` 发起 POST 请求：

   ```bash
   $ curl http://localhost:8080/albums \
       --include \
       --header "Content-Type: application/json" \
       --request "POST" \
       --data '{"id": "4","title": "The Modern Sound of Betty Carter","artist": "Betty Carter","price": 49.99}'
   ```

   输出应包含：

   ```json
   HTTP/1.1 201 Created
   {
       "id": "4",
       "title": "The Modern Sound of Betty Carter",
       "artist": "Betty Carter",
       "price": 49.99
   }
   ```

4. 再次获取所有专辑，确认新专辑已添加：

   ```bash
   $ curl http://localhost:8080/albums
   ```

   列表中应包含 ID 为 `"4"` 的新专辑。

---

### 编写处理函数以返回特定条目

当客户端访问 `GET /albums/[id]` 时，你希望返回对应 ID 的专辑。

#### 编写代码

1. 添加 `getAlbumByID` 函数：

   ```go
   // getAlbumByID 根据客户端传入的 id 查找专辑并返回
   func getAlbumByID(c *gin.Context) {
       id := c.Param("id")
       // 遍历专辑列表，查找匹配 ID 的专辑
       for _, a := range albums {
           if a.ID == id {
               c.IndentedJSON(http.StatusOK, a)
               return
           }
       }
       // 未找到时返回 404
       c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
   }
   ```

   说明：

    - `c.Param("id")` 获取 URL 路径中的 `id` 参数；
    - 遍历 `albums` 切片查找匹配项；
    - 找到则返回 200 OK 和专辑数据；
    - 未找到则返回 404 和错误信息。

2. 修改 `main` 函数，添加带参数的 GET 路由：

   ```go
   func main() {
       router := gin.Default()
       router.GET("/albums", getAlbums)
       router.GET("/albums/:id", getAlbumByID)
       router.POST("/albums", postAlbums)
       router.Run("localhost:8080")
   }
   ```

    - `:id` 是 Gin 中的路径参数占位符。

#### 运行代码

1. 停止并重新运行服务器。
2. 使用 `curl` 请求特定专辑：

   ```bash
   $ curl http://localhost:8080/albums/2
   ```

   输出应为：

   ```json
   {
       "id": "2",
       "title": "Jeru",
       "artist": "Gerry Mulligan",
       "price": 17.99
   }
   ```

   如果请求一个不存在的 ID，会返回：

   ```json
   {"message": "album not found"}
   ```

---

### 结语

🎉 恭喜！你已经使用 Go 和 Gin 成功构建了一个简单的 RESTful Web 服务！

#### 建议后续学习：

- 如果你是 Go 新手，推荐阅读 [《Effective Go》](https://go.dev/doc/effective_go) 和 [《如何编写 Go 代码》](https://go.dev/doc/code)。
- [Go 语言之旅](https://go.dev/tour/) 是学习 Go 基础的绝佳互动教程。
- 想深入了解 Gin？查看 [Gin Web Framework 文档](https://gin-gonic.com/docs/)。

---

### 完整代码

以下是本教程最终的完整代码：

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// album 表示一张唱片专辑的信息
type album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}

// albums 切片用于存放初始的专辑数据
var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
    router := gin.Default()
    router.GET("/albums", getAlbums)
    router.GET("/albums/:id", getAlbumByID)
    router.POST("/albums", postAlbums)
    router.Run("localhost:8080")
}

// getAlbums 返回所有专辑的 JSON 列表
func getAlbums(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums 从请求体的 JSON 中解析并添加一张新专辑
func postAlbums(c *gin.Context) {
    var newAlbum album
    if err := c.BindJSON(&newAlbum); err != nil {
        return
    }
    albums = append(albums, newAlbum)
    c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID 根据客户端传入的 id 查找专辑并返回
func getAlbumByID(c *gin.Context) {
    id := c.Param("id")
    for _, a := range albums {
        if a.ID == id {
            c.IndentedJSON(http.StatusOK, a)
            return
        }
    }
    c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
```
> **[↩ 返回到目录](../doc.md)**