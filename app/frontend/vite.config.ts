import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: "../web/dist",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: {
          editor: ["@uiw/react-codemirror", "@codemirror/lang-go"],
          markdown: ["react-markdown"],
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
