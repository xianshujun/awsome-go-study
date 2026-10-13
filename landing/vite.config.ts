import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  // 部署在 /awsome-go-study/ 子路径下，资源一律用相对路径
  base: "./",
  build: { outDir: "dist", emptyOutDir: true },
});
