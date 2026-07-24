import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Phase1 WP-V: Vite 6+ · Vue3 · same-origin dev proxy to Go :3000
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': { target: 'http://127.0.0.1:3000', changeOrigin: true },
      '/healthz': { target: 'http://127.0.0.1:3000', changeOrigin: true },
      '/livez': { target: 'http://127.0.0.1:3000', changeOrigin: true },
      '/readyz': { target: 'http://127.0.0.1:3000', changeOrigin: true },
      '/v1': { target: 'http://127.0.0.1:3000', changeOrigin: true },
    },
  },
  build: {
    // Sanitize chunk file names: Go's //go:embed excludes files starting with
    // '_' or '.' from directory patterns, so strip the leading underscore that
    // @vitejs/plugin-vue puts on its internal _plugin-vue_export-helper chunk.
    rollupOptions: {
      output: {
        chunkFileNames: (chunkInfo) => {
          const name = (chunkInfo.name || 'chunk').replace(/^_+/, '')
          return `assets/${name}-[hash].js`
        },
      },
    },
  },
})
