> [点此返回上级页面](../README.md)

# Qwiklabs：Go 程序的引导式实践

Qwiklabs（现名 **Google Cloud Skills Boost**，平台入口已迁移至 `skills.google`）提供一系列动手实验（labs），带领你在真实的 Google Cloud 环境中完成与 Go 相关的实践任务。本节整理 go.dev/learn 页面推荐的三项实验。

> **说明**：实验正文托管在 Google Cloud Skills Boost 平台上，为英文交互式实验，需注册登录后按步骤操作。本页翻译并整理了每项实验的官方简介与任务清单；实验本体请在平台上完成。

---

<a id="lab-1"></a>
## 1. 在 Google Cloud 无服务器平台上部署 Go 应用

- **实验编号**：GSP702
- **时长**：1 小时 10 分钟
- **学分**：5 积分
- **官方简介（译）**：Go 是 Google 推出的开源编程语言，便于大规模构建快速、可靠、高效的软件。在本实验中，你将通过把一个简单的 Go 应用部署到 Cloud Run Functions 和 App Engine 来探索 Go 的基础用法，随后使用该 Go 应用访问 BigQuery 和 Firestore 中的数据。
- **实验内容**：
  - 设置 Firestore 数据库并导入数据；
  - 初步了解 Cloud Build 的能力；
  - 探索 BigQuery 与 Firestore 中的数据；
  - 将 Go 应用部署到 App Engine 和 Cloud Run Functions；
  - 查看 Go 应用代码；
  - 在各个平台上测试该应用。
- **实验入口**：<https://google.qwiklabs.com/focuses/10532?parent=catalog>

---

<a id="lab-2"></a>
## 2. 使用 Go 代码操作 Google Cloud 数据源

- **实验编号**：GSP701
- **时长**：1 小时 10 分钟
- **学分**：5 积分
- **官方简介（译）**：Go 是 Google 支持的开源编程语言，便于大规模构建快速、可靠、高效的软件。在本实验中，你将在 Cloud Shell 中交互式地编译、测试一个 Go 应用，然后把它部署到 App Engine，并使用该 Go 应用访问 BigQuery 和 Firestore 中的数据。
- **实验内容**：
  - 对 Google 开发的某个开源工具进行代码走查；
  - 探索从 Google Cloud 提取数据的代码模式；
  - 在浏览器中测试你的 Go 应用；
  - 将 Go 应用部署到 App Engine；
  - 通过浏览器或 curl 命令测试已部署的 Go 应用。
- **实验入口**：<https://google.qwiklabs.com/focuses/10531?parent=catalog>

---

<a id="lab-3"></a>
## 3. 在 App Engine 上开始使用 Go

- **实验编号**：GSP070
- **官方实验名**：App Engine: Qwik Start - Go
- **时长**：20 分钟
- **学分**：1 积分
- **官方简介（译）**：App Engine 让开发者可以专注于最擅长的事情——写代码。App Engine **标准环境**基于运行在 Google 基础设施上的容器实例，容器预装了若干可用运行时之一，每个运行时还包含支持 App Engine 标准 API 的库；对许多应用而言，标准环境的运行时和库已经足够。App Engine 为 Go 提供了**标准环境**和**灵活环境**两种选择：两者都采用以代码为中心的开发流程，都能快速、高效地扩容以应对增长的需求，也都基于 Google 成熟的托管技术，让你以最小的运维开销构建 Web、移动和 IoT 应用；但两者在一些重要方面有所不同。标准环境让应用能在高负载、大数据量下可靠运行，你的应用运行在独立的安全环境中，与服务器硬件、操作系统和物理位置无关，该环境支持 **Go 1.11**；灵活环境基于 Compute Engine，可在负载均衡的同时自动伸缩应用。
- **实验内容**：向 App Engine 部署一个 Go 应用。
- **实验入口**：<https://google.qwiklabs.com/focuses/2754?parent=catalog>

---

> [点此返回上级页面](../README.md)
