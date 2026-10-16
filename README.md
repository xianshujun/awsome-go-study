# 🚀 Go 学习之路 - go-study

<p align="center">
  <a href="https://github.com/xianshujun/awsome-go-study/stargazers"><img src="https://img.shields.io/github/stars/xianshujun/awsome-go-study?color=0f5666&label=Stars" alt="GitHub Stars"></a>
  <a href="https://xianshujun.github.io/awsome-go-study/"><img src="https://img.shields.io/badge/落地页-在线预览-0f5666.svg" alt="落地页"></a>
  <img src="https://img.shields.io/badge/语言-Go-blue.svg" alt="Language">
  <img src="https://img.shields.io/badge/类型-学习资源-brightgreen.svg" alt="Type">
  <img src="https://img.shields.io/badge/难度-从入门到进阶-orange.svg" alt="Difficulty">
</p>

## 📚 项目介绍

> 此项目主要是为了帮助初学 Go 的同学，快速掌握面试常见内容，同时对于有经验的同学，也可以作为一个整理好的参考资料。

> 除此之外，我们还将提供一份完整的**Go 官方文档的翻译**，以方便大家查阅。

### 💡 tips：Go 官方文档翻译（AI 翻译 + 人工校正）

为了帮助更多人更好地学习 Go 语言，我们将 Go 官方文档翻译成中文，并进行人工校正。让大家更容易理解官方文档内容。
点击下方链接，查看翻译后的文档：

👉 [go 官方文档的翻译](go-offical-doc-translate/README.md)

## 🎯 学习目标

- ✅ 快速掌握 Go 语言核心概念
- ✅ 深入理解并发编程
- ✅ 熟悉标准库和工程实践
- ✅ 掌握实际开发技能

## 🛣️ Go 学习路线

### 🔵 1. Go 语言基础
- 📘 [go语言语法](go-basic/go-grammar.md) - 打好基础
- 📘 [go并发编程](go-basic/go-concurrency.md) - 掌握核心特性

### 🟡 2. Go 语言进阶
- 📒 [go语言深挖](go-advance/go-deepdive.md) - 深入理解语言特性
- 📒 [go语言标准库](go-advance/go-standardlib.md) - 熟悉常用库
- 📒 [工程实践](go-advance/go-engineering.md) - 实际项目应用与工程规范

### 🔴 3. Go 语言开发实战
- 📕 [go语言网络开发](go-practice/go-webdev.md) - Web 开发技能
- 📕 [中间件与工具集成](go-practice/go-middleware.md) - 系统集成能力
- 📕 [go语言微服务](go-practice/go-microservices.md) - 现代架构实践

## 🖥️ GoStudy 交互式学习应用

[GoStudy](app/README.md) 提供 React 界面与 Go 本地学习引擎，包含引导课程、代码运行、理解题、学习笔记和可配置的 AI 导师。课程与练习可以离线使用，不需要部署服务端。

课程内容依据本仓库资料整理：`go-basic`、`go-advance`、`go-practice` 的每篇文档对应一门课程，每节练习都标注来源章节，方便对照原文继续阅读。

**核心特点**

- **引导式学习闭环**：先读懂概念，再改代码、真运行、答理解题，最后记笔记——每一步都有反馈。
- **8 门课程 · 43 个练习**：Go 语言基础（语法 / 并发）、Go 语言进阶（语言深挖 / 标准库 / 工程实践）、Go 开发实战（网络开发 / 中间件 / 微服务）。
- **真实运行代码**：练习由内置 Go 解释器在本机执行，限制导入符号、运行时间与输出大小，不开放文件与网络访问。
- **可配置的 AI 导师**：支持 OpenAI Chat Completions 兼容接口，也可以接本机模型；结合当前课程、代码和笔记做引导式提问，只有你主动提问时才发送上下文。面板停在右侧跟练习并排，打开它不影响改代码和运行；对话记在本机，可以新建、切换、重命名、删除，收起或重启后都能接着问。
- **应用内更新**：设置页里就能检查、下载并更新——Windows 免安装版原地替换自身、安装版静默升级并重启，Android 下载后交给系统安装器；课程内容还能单独更新，改讲义不用等新版本。只拉版本清单和发布文件，不采集任何使用数据。
- **本地优先**：没有账号，学习进度、笔记和草稿都保存在本机，断网也能继续学。

**在线落地页**：<https://xianshujun.github.io/awsome-go-study/> —— 一页看完功能、界面截图与三平台下载入口，源码在 [`landing/`](landing/)。

**界面预览**

| 学习空间 | 探索课程 |
|---|---|
| ![学习空间](app/docs/screenshots/01-home.png) | ![探索课程](app/docs/screenshots/02-courses.png) |
| **课程与练习**（标注来源章节） | **AI 导师**（可配置模型） |
| ![课程与练习](app/docs/screenshots/03-lesson.png) | ![AI 导师](app/docs/screenshots/04-assistant.png) |

| 偏好与 AI 设置 | 手机端 |
|---|---|
| ![偏好与 AI 设置](app/docs/screenshots/05-settings.png) | ![手机端](app/docs/screenshots/06-mobile.png) |

**获取方式**

- **Windows（安装版）**：下载 `GoStudy-Setup.exe` 双击安装——可以自己选装在哪个目录，装完自动创建桌面与开始菜单快捷方式，卸载走系统「应用和功能」。
- **Windows（免安装）**：下载 `GoStudy.exe`，双击即用（也可从源码运行 `bash app/scripts/build-windows.sh` 自行构建）。
- **Android**：下载 `GoStudy-android-release.apk`（Android 8.0+，正式签名，可直接安装），或运行 `bash app/scripts/build-android.sh release` 自行构建并签名。
- **Linux**：下载 `GoStudy-linux-amd64` / `GoStudy-linux-arm64`，`chmod +x` 后直接运行，或执行 `bash app/scripts/build-linux.sh`。
- **从源码运行**：[应用说明](app/README.md) 的「2.1 本机启动」给出完整命令。

以上产物都在 [Releases](https://github.com/xianshujun/awsome-go-study/releases) 里；[落地页](https://xianshujun.github.io/awsome-go-study/)的下载按钮直接指向最新版本的产物文件，点一下就开始下载，不必再进 Release 页面挑文件。各平台产物的细节与构建方式见[应用说明](app/README.md)的「2.2 Windows 应用」与「2.3 Android 应用」。

**国内镜像（Gitee）**

GitHub 的产物文件走 `objects.githubusercontent.com`，国内经常只有几十 KB/s，所以每次发布都会把同一批文件镜像一份到 Gitee。**落地页不让你选**：它自己看访客的位置决定下载源——国内访客默认走 Gitee，海外访客走 GitHub，镜像版本对不上、少了文件或读不到时自动回落 GitHub；卡片上只有「直接下载」，不会出现两个源让你挑。

- 镜像地址与 GitHub 同形，把域名换掉即可：`https://gitee.com/pang-xianshu/gostudy-release/releases/download/latest/GoStudy-Setup.exe`（`latest` 也可以换成具体 tag）。
- **推荐做法：发版时直接在本机传。** 本机到 Gitee 实测 1.6 MB/s，5 个文件合计 110 MB 一分多钟就传完；命令见下面的「发布时在本机镜像」。内容是**从 GitHub Release 下载下来的同一批文件**，两边必然一致；上传是幂等的（同名文件先传新的、成功后再删旧的），随时可以重跑。
- `.github/workflows/release.yml` 里还有一个 `gitee-mirror` 任务作为兜底（没在本机传时它会自动跑）。
- ⚠️ 兜底任务走的 GitHub 运行器出口到 Gitee 会被限速（几十 KB/s，实测 25 分钟才传完 3 个文件；67 MB 的 APK 更久）。脚本会在「低于 10 KB/s 持续 60 秒」时直接报错退出（任务本身封顶 90 分钟），不会干挂——所以别等它，本机传更快。镜像的上传顺序是先传新的、成功后再删旧的，中途失败也不会把已镜像的版本改坏。
- **没配置令牌时这一步自动跳过**，流水线照常是绿的。

一次性开通步骤：

1. 注册 Gitee 账号；如提示需要实名认证，按提示完成（Gitee 的部分功能对实名有要求）。
2. 新建一个公开仓库专门放产物，例如 `gostudy-release`，并初始化 README——创建发行版要指定分支，所以仓库里得先有分支。
3. 生成私人令牌：Gitee → 头像 → 设置 → 私人令牌 → 生成新令牌，勾选 `projects` 权限。
4. 在 GitHub 仓库 → Settings → Secrets and variables → Actions 里加 secret `GITEE_TOKEN`（值为上一步的令牌）。想用别的仓库名，再在同一页的 Variables 里加 `GITEE_REPO`（默认 `pang-xianshu/gostudy-release`），同时改 `landing/src/main.tsx` 里的 `GITEE_REPO` 常量。
5. 补发已发布的版本：Actions → Release → Run workflow，`gitee_tag` 填 `v0.4.0`；之后每次推 `v*` tag 都会自动镜像。

发布时在本机镜像（推荐，需要令牌）：

```bash
GITEE_TOKEN=你的令牌 bash app/scripts/publish-gitee-release.sh \
    --repo pang-xianshu/gostudy-release --tag v0.6.0 --dir app/release
```

> ⚠️ 我们的包最大是 APK 约 65 MB，在 Gitee 免费仓库的附件限制内；如果以后有更大的产物被拒，日志里会写明原因，也可以只镜像桌面端两个文件。

Android 说明：APK 为**正式签名**（APK Signature Scheme v2 + v3，自签名证书，非应用商店分发），内含 arm64-v8a / armeabi-v7a / x86 / x86_64 四种 ABI，约 65 MB；安装时需要允许「未知来源」。签名密钥由维护者保管，一旦丢失将导致后续版本无法覆盖安装，请勿提交到仓库。
开始使用、AI 配置和构建方法见 [应用说明](app/README.md)。

---




###  贡献

欢迎提交 Issue 和 Pull Request 来帮助改进这个项目！

###  许可证

MIT License - 允许任何人自由使用、修改和分发代码，但需要保留原作者信息。
