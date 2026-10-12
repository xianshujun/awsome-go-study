package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// Dist 是前端构建输出；发布时由构建脚本生成 dist 内容。
//
//go:embed dist
var Dist embed.FS

func Handler() http.Handler {
	root, err := fs.Sub(Dist, "dist")
	if err != nil {
		panic(err)
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
