import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: './',
  build: { outDir: '../nacl_backend/static' },
  server: {
    proxy: {
      '/api/': {
        target: 'http://localhost:3333',
        changeOrigin: true
      }
    },
    allowedHosts: true
  }
})
