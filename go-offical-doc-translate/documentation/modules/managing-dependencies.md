> [返回到目录](../doc.md)
---

### 包管理

#### 目录
*   介绍
*   模块
*   依赖关系
*   发布
*   代理
*   工具

#### 介绍
Go 的包管理是模块化的，基于模块（modules）的概念。模块是一组相关的 Go 包，共享一个公共的导入路径前缀和一个共同的版本。本文档解释了如何创建和管理模块，以及如何使用 Go 工具链来处理依赖关系。

#### 模块
一个模块由一个 `go.mod` 文件定义，该文件位于模块根目录中。`go.mod` 文件指定模块的路径（也称为模块名）、Go 版本和依赖项。

要创建一个新模块，请在包含源代码的目录中运行 `go mod init` 命令：
```bash
go mod init example.com/mymodule
```
这将创建一个名为 `go.mod` 的文件，内容如下：
```go
module example.com/mymodule

go 1.21
```

模块路径（`example.com/mymodule`）是你的模块的唯一标识符。它通常与模块源代码所在的版本控制仓库的 URL 相对应。

#### 依赖关系
当你在代码中导入一个不在你模块中的包时，Go 工具会自动将其添加为依赖项。例如，如果你的代码包含 `import "golang.org/x/text"`，运行 `go build` 或 `go mod tidy` 会自动将 `golang.org/x/text` 添加到 `go.mod` 文件中，并在 `go.sum` 文件中记录其校验和。

你也可以使用 `go get` 命令显式地添加、升级或降级依赖项：
```bash
# 添加或升级依赖项到最新版本
go get golang.org/x/text

# 升级到特定版本
go get golang.org/x/text@v0.3.5

# 降级到特定版本
go get golang.org/x/text@v0.3.4

# 移除依赖项
go get golang.org/x/text@none
```

`go.sum` 文件包含每个依赖模块的加密校验和。这可以防止依赖项在不知不觉中被修改。

#### 发布
为了让你的模块可以被他人使用，你需要将其发布到一个公共的版本控制仓库（如 GitHub、GitLab）。

发布一个新版本时，你需要创建一个版本标签（tag）。Go 模块遵循语义化版本控制（Semantic Versioning）规范。版本号的格式为 `vMAJOR.MINOR.PATCH`。

*   **MAJOR**：当你进行不兼容的 API 变更时递增。
*   **MINOR**：当你以向后兼容的方式添加新功能时递增。
*   **PATCH**：当你进行向后兼容的 bug 修复时递增。

例如，要发布 `v1.0.0` 版本，你可以在命令行中运行：
```bash
git tag v1.0.0
git push origin v1.0.0
```

一旦你发布了版本标签，其他开发者就可以通过 `go get` 命令来使用你的模块。

#### 代理
Go 模块代理（Go module proxy）是一个服务，它缓存公开可用的模块，以提高下载速度和可靠性。

默认情况下，`go` 命令会使用 Google 的公共代理 `proxy.golang.org`。你也可以配置使用其他代理，或者完全禁用代理。

你可以通过设置 `GOPROXY` 环境变量来配置代理：
```bash
# 使用 Google 的公共代理（默认）
export GOPROXY=https://proxy.golang.org,direct

# 使用其他代理
export GOPROXY=https://goproxy.cn,direct

# 完全禁用代理，直接从源下载
export GOPROXY=direct
```

`direct` 关键字表示如果代理无法提供模块，则直接从源（如 GitHub）下载。

#### 工具
Go 工具链提供了一系列用于管理模块的命令：

*   `go mod init`：初始化一个新的模块。
*   `go mod tidy`：清理 `go.mod` 文件，添加缺失的依赖项，移除未使用的依赖项。
*   `go mod download`：下载并缓存所有依赖项。
*   `go mod verify`：验证依赖项的校验和是否与 `go.sum` 文件匹配。
*   `go list -m all`：列出当前模块及其所有依赖项。
*   `go list -m -versions <module>`：列出指定模块的所有可用版本。

---

> [返回到目录](../doc.md)
