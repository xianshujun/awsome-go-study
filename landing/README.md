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

## 约定

- **零渐变**：与应用同一套设计令牌（近中性底色、单一强调色、代码区墨色表面）。
- **动效不能把内容藏住**：`rise()` / `reveal()` 在「减少动效」时返回空对象，元素首帧即最终状态。
- 下载入口统一指向 `releases/latest`，发新版本后页面自动生效。
