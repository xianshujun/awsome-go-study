# GoStudy 落地页

GitHub Pages 上的产品页，用和应用相同的技术栈搭建：**Vite + React 19 + Radix UI + Motion**。

```bash
npm install
npm run dev        # 本地预览（截图需要先按下面一步复制）
npm run build      # 产物在 dist/
```

## 截图来自应用仓库

页面引用的截图不在这里维护，而是部署时从 `app/docs/screenshots/` 复制进 `dist/screenshots/`：

```bash
npm run build
mkdir -p dist/screenshots && cp ../app/docs/screenshots/*.png dist/screenshots/
```

这样应用界面更新、截图重出之后，落地页自动跟着变，不会出现两处不一致。

## 字体

`src/fonts/` 里的两款拉丁子集（SIL OFL 1.1）与 `app/frontend/public/fonts/` 是同一份文件，复制过来是为了让 Vite 生成带哈希的资源并自动处理相对路径。更新方式：

```bash
cp ../app/frontend/public/fonts/*.woff2 ../app/frontend/public/fonts/LICENSE-*.txt src/fonts/
```

## 优雅版 UI 与动效

在原有 React、Radix UI 与 Motion 基础上，新增 **Radix Themes** 按钮与主题组件，以及 **GSAP ScrollTrigger** 滚动编排。视觉使用暖白、墨色和苔绿纯色；大标题、细线与编号组织内容。

- 标题新增自托管思源宋体子集 `src/fonts/noto-serif-sc-landing.woff2`，覆盖当前页面汉字与标点，约 76 KB，附 SIL OFL 授权；增加标题汉字时需重新生成子集或使用系统回退字体。
- 首屏标题逐行揭示，代码示例按顺序呈现，滚动时演示台轻微移动。
- 阅读进度、截图缩放、标签指示器和弹窗使用不同节奏；不接管鼠标，不劫持滚动。
- 移动端提供 Radix Dialog 导航，支持键盘关闭和焦点恢复。
- 系统开启「减少动态效果」时，关闭滚动视差、标题位移、CSS 过渡和滚动平滑，所有内容直接可见。
- 2026-10-13 起，Vite 开发服务器直接提供应用截图，生产构建自动复制截图到 `dist/screenshots/`。上面的手动复制命令仍可用于独立核对，正常预览不再需要额外复制。

## 约定

- **零渐变**：与应用同一套设计令牌（近中性底色、单一强调色、代码区墨色表面）。
- **动效不能把内容藏住**：`rise()` / `reveal()` 在「减少动效」时直接返回最终状态，元素首帧即最终状态；页面打开期间切换系统偏好也立即复位。
- 下载入口统一指向 `releases/latest/download/<资源名>`，发新版本后页面自动生效。

## 下载区怎么工作

`src/main.tsx` 里的 `downloads` 数组只存平台、Release 资源名和文案，链接由 `latestAssetUrl()` 拼成
`https://github.com/xianshujun/awsome-go-study/releases/latest/download/<资源名>`：这个地址会 302 到最新一次
Release 的对应文件，所以**发布新版本之后页面不用改**，点一下就开始下载，不再跳到 Releases 页面。

页面另外读一次 `releases/latest` 接口，把版本号、发布日期和文件大小填进卡片（`loadRelease()`）：

| 情形 | 页面表现 |
|---|---|
| 读到数据 | 卡片显示「6.5 MB · 直接下载」，标题下方出现「当前最新版本 vX.Y.Z · 日期」 |
| 读不到（限流 / 断网 / 还没有 Release） | 退回卡片上的静态文案，直链照旧可用 |
| 读到了、但该版本没有这个产物 | 这张卡片改跳 Releases 列表并说明原因，不会出现点进去 404 的按钮 |

结果按 30 分钟缓存在 `sessionStorage`，和 Star 数一样是为了省下未认证 GitHub API 的 60 次/小时配额。

新增平台或在 Release 里改了资源名，改 `downloads` 数组即可——资源名要和 `.github/workflows/release.yml`
上传的名字一致（`GoStudy-Setup.exe`、`GoStudy.exe`、`GoStudy-android-release.apk`、`GoStudy-linux-amd64`、
`GoStudy-linux-arm64`）。

### 下载源由页面自动选（国内走 Gitee）

`GITEE_REPO` 常量指向 Gitee 上的镜像仓库（默认 `pang-xianshu/gostudy-release`，与 workflow 里的
`vars.GITEE_REPO` 保持一致；留空就完全不走镜像）。**页面上不出现"国内镜像"这类字眼，也不给用户选择**，
`assetUrl()` 按下面的规则逐文件决定链接：

| 访客 | 镜像状态 | 走哪里 |
|---|---|---|
| 国内（时区 `Asia/Shanghai` 等，或 `zh*` + 亚洲时区） | 还没查完 | Gitee（先按镜像走，查完再纠正） |
| 国内 | 有、且 tag 与 GitHub 最新版本一致、该文件在 | Gitee |
| 国内 | 镜像落后一个版本 / 少这个文件 / 接口读不到 | GitHub |
| 海外 | 任何情况 | GitHub |

镜像数据来自 `loadMirror()`：读 Gitee 的发行版列表接口（匿名可用、允许跨域），按 30 分钟缓存在
`sessionStorage`。判断条件写成"宁可回 GitHub"——镜像不可信时绝不把用户送过去下到旧版本或 404。

截图核对用的假接口脚本在 `/tmp/shot/mock-both.py` + `gl-source-check.mjs`（本机临时文件，不在仓库里），
覆盖「国内+镜像正常 / 国内+镜像落后 / 国内+缺文件 / 海外+正常 / 国内+接口不通」五种情况。

