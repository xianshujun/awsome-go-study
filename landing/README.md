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
- 下载入口统一指向 `releases/latest`，发新版本后页面自动生效。
