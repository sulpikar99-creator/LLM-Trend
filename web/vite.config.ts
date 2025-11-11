import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: parseInt(process.env.NOFX_FRONTEND_PORT || '3000'),
    proxy: {
      '/api': {
        target: `http://localhost:${process.env.NOFX_BACKEND_PORT || '8080'}`,
        changeOrigin: true,
      },
      '/ws': {
        target: `ws://localhost:${process.env.NOFX_BACKEND_PORT || '8080'}`,
        ws: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: true,
  },
})
