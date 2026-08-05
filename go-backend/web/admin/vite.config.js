import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    host: '127.0.0.1',
    port: 9300,
    strictPort: true,
    proxy: {
      '/api': {
        target: 'http://localhost:9200',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
    rollupOptions: {
      output: {
        manualChunks(id) {
          const moduleId = id.replace(/\\/g, '/')
          if (!moduleId.includes('/node_modules/')) return undefined

          if (moduleId.includes('/zrender/')) {
            return 'vendor-zrender'
          }

          if (
            moduleId.includes('/echarts/core') ||
            moduleId.includes('/echarts/lib/export/') ||
            moduleId.includes('/echarts/lib/echarts') ||
            moduleId.includes('/echarts/lib/core/') ||
            moduleId.includes('/echarts/lib/model/') ||
            moduleId.includes('/echarts/lib/view/')
          ) {
            return 'vendor-echarts-core'
          }

          if (
            moduleId.includes('/echarts/charts') ||
            moduleId.includes('/echarts/lib/chart/')
          ) {
            return 'vendor-echarts-charts'
          }

          if (
            moduleId.includes('/echarts/components') ||
            moduleId.includes('/echarts/lib/component/') ||
            moduleId.includes('/echarts/lib/coord/')
          ) {
            return 'vendor-echarts-components'
          }

          if (moduleId.includes('/echarts/renderers')) {
            return 'vendor-echarts-renderers'
          }

          if (
            moduleId.includes('/echarts/lib/data/') ||
            moduleId.includes('/echarts/lib/scale/') ||
            moduleId.includes('/echarts/lib/visual/') ||
            moduleId.includes('/echarts/lib/processor/')
          ) {
            return 'vendor-echarts-data'
          }

          if (
            moduleId.includes('/echarts/lib/util/') ||
            moduleId.includes('/echarts/lib/label/') ||
            moduleId.includes('/echarts/lib/layout/')
          ) {
            return 'vendor-echarts-utils'
          }

          if (moduleId.includes('/echarts/')) {
            return 'vendor-echarts-runtime'
          }

          if (
            moduleId.includes('/vue/') ||
            moduleId.includes('/@vue/') ||
            moduleId.includes('/vue-router/') ||
            moduleId.includes('/pinia/')
          ) {
            return 'vendor-vue'
          }

          if (moduleId.includes('/@lucide/')) {
            return 'vendor-icons'
          }

          if (
            moduleId.includes('/reka-ui/') ||
            moduleId.includes('/@floating-ui/') ||
            moduleId.includes('/@tanstack/')
          ) {
            return 'vendor-ui'
          }

          if (moduleId.includes('/axios/')) {
            return 'vendor-http'
          }

          return undefined
        }
      }
    }
  }
})
