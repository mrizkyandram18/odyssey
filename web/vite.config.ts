import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const buildTime = Date.now().toString()

export default defineConfig({
  plugins: [
    react(),
    {
      name: 'generate-version-json',
      generateBundle() {
        this.emitFile({
          type: 'asset',
          fileName: 'version.json',
          source: JSON.stringify({
            build_time: buildTime,
            timestamp: new Date().toISOString(),
          }),
        })
      },
    },
  ],
  define: {
    'import.meta.env.VITE_BUILD_TIME': JSON.stringify(buildTime),
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: undefined
      }
    }
  }
})
