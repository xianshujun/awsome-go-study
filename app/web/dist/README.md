# 这个目录是前端构建输出

`app/web/dist` 由 `npm --prefix app/frontend run build` 生成，产物通过 `//go:embed dist`
打进 Go 二进制，正常构建（`app/scripts/*.sh`）会自动重建它。

本说明文件是**唯一被提交进仓库的文件**，作用是让 `//go:embed dist` 在还没有构建产物时
也能编译通过——否则全新克隆或用 Go 工具链直接编译会报 `pattern dist: no matching files found`。

> 只提交这一个说明文件；构建产生的 html / js / css 都在 `.gitignore` 里，不要手动加入版本库。
