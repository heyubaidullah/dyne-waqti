import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// Builds directly into the Go package that embeds it (internal/api/adminui),
// so `go build` picks up whatever was last built here with no extra copy
// step. See internal/api/adminui/embed.go.
export default defineConfig({
  plugins: [react()],
  base: '/admin/',
  build: {
    outDir: '../../internal/api/adminui/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:3000',
      '/uploads': 'http://localhost:3000',
    },
  },
})
