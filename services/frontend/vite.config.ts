import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";


export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // dev: относительные /api/v1 проксируем на gateway, чтобы фронт был same-origin.
    proxy: {
      "/api": "http://localhost:8080",
    },
  },
});
