import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const frontendPort = parseInt(process.env.FRONTEND_PORT || '3000', 10)
const backendPort = process.env.PORT || '8080'

export default defineConfig({
  plugins: [react()],
  server: {
    port: frontendPort,
    proxy: {
      '/api': {
        target: `https://127.0.0.1:${backendPort}`,
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
