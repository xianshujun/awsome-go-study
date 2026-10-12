import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../web/dist",
    // 不清空 dist：该目录有一个被版本库跟踪的 README.md，用于让 //go:embed 在没有
    // 构建产物时也能编译。构建脚本会在构建前清理旧产物、只保留这个说明文件。
    emptyOutDir: false,
    rollupOptions: {
      output: {
        manualChunks: {
          editor: ["@uiw/react-codemirror", "@codemirror/lang-go"],
          markdown: ["react-markdown"],
          // 动效与无障碍组件各自成块：都是本地文件，拆开只为让体积变化看得清。
          motion: ["motion"],
          radix: ["radix-ui"],
        },
      },
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://127.0.0.1:7731",
        changeOrigin: true,
        configure(proxy) {
          proxy.on("proxyReq", (request) => {
            request.setHeader("Origin", "http://127.0.0.1:7731");
          });
        },
      },
    },
  },
});
