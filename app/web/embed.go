package web

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Dist 是前端构建输出；正常构建由脚本生成。
//
//go:embed dist
var Dist embed.FS

// 还没有构建前端时给出的提示页：直接跑 go run 也能看到可操作的说明，而不是 404。
const placeholderPage = `<!doctype html><html lang="zh-CN"><head><meta charset="utf-8">
<title>GoStudy · 前端资源未构建</title>
<style>
body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;
background:#fbfcfe;color:#3f3a5a;font-family:-apple-system,"PingFang SC","Microsoft YaHei",sans-serif}
main{max-width:620px;padding:40px}
h1{color:#130b43;font-size:22px;margin:0 0 14px}
code{background:#f0f3f9;border:1px solid #e1e5f0;border-radius:3px;padding:2px 6px;color:#130b43}
pre{background:#f0f3f9;border:1px solid #e1e5f0;border-radius:8px;padding:14px 16px;overflow:auto;line-height:1.8}
a{color:#6a01ff}
</style></head><body><main>
<h1>前端资源还没有构建</h1>
<p>Go 引擎已经在运行，但内嵌的前端文件为空。先构建界面再刷新本页：</p>
<pre>npm --prefix app/frontend ci --no-audit --no-fund
npm --prefix app/frontend run build</pre>
<p>或者直接用一键脚本（会自动完成上面的步骤）：<code>bash app/scripts/build-windows.sh</code>、
<code>bash app/scripts/build-linux.sh</code>。</p>
<p>说明见 <a href="https://github.com/xianshujun/awsome-go-study/blob/master/app/README.md">应用文档</a>。</p>
</main></body></html>`

func Handler() http.Handler {
	root, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic(err)
	}
	if _, err := fs.Stat(root, "index.html"); err != nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = io.WriteString(w, placeholderPage)
		})
	}
	files := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			name = "index.html"
		}
		if _, err := fs.Stat(root, name); err != nil {
			// 页面路由回退到入口，缺失的静态资源保留 404。
			if path.Ext(name) != "" {
				http.NotFound(w, r)
				return
			}
			clone := r.Clone(r.Context())
			copyURL := *r.URL
			copyURL.Path = "/"
			clone.URL = &copyURL
			files.ServeHTTP(w, clone)
			return
		}
		files.ServeHTTP(w, r)
	})
}
