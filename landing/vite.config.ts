import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { cpSync } from "node:fs";
import { fileURLToPath } from "node:url";

export default defineConfig({
  plugins: [
    react(),
    {
      name: "application-screenshots",
      closeBundle() {
        cpSync(
          fileURLToPath(new URL("../app/docs/screenshots", import.meta.url)),
          fileURLToPath(new URL("./dist/screenshots", import.meta.url)),
          { recursive: true },
        );
      },
      configureServer(server) {
        server.middlewares.use(
          "/screenshots",
          async (request, response, next) => {
            const { createReadStream, existsSync } = await import("node:fs");
            const file = request.url?.split("?")[0];
            if (!file || !/^\/[\w-]+\.png$/.test(file)) return next();
            const path = fileURLToPath(
              new URL(`../app/docs/screenshots${file}`, import.meta.url),
            );
            if (!existsSync(path)) return next();
            response.setHeader("Content-Type", "image/png");
            createReadStream(path).pipe(response);
          },
        );
      },
    },
  ],
  // 部署在 /awsome-go-study/ 子路径下，资源一律用相对路径
  base: "./",
  build: {
    outDir: "dist",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes("node_modules")) return;
          if (id.includes("gsap")) return "scroll-motion";
          if (id.includes("motion") || id.includes("framer")) return "motion";
          if (id.includes("radix") || id.includes("react-remove")) return "ui";
          return "vendor";
        },
      },
    },
  },
});
