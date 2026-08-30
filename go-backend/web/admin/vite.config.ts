import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const projectRoot = fileURLToPath(new URL('.', import.meta.url))

const normalizeAdminFontFallbacks = (source: string): string => source
  .replace(
    /var\(\s*--default-font-family\s*,\s*[^)]*\)/gi,
    'var(--tz-font-admin-ui)',
  )
  .replace(
    /var\(\s*--default-mono-font-family\s*,\s*[^)]*\)/gi,
    'var(--tz-font-admin-ui)',
  )
  .replace(
    /font-family\s*:\s*ui-sans-serif,\s*system-ui,\s*-apple-system,\s*BlinkMacSystemFont,\s*Segoe UI,\s*Roboto,\s*Helvetica Neue,\s*Arial,\s*Noto Sans,\s*sans-serif,\s*Apple Color Emoji,\s*Segoe UI Emoji,\s*Segoe UI Symbol,\s*Noto Color Emoji\s*;/gi,
    'font-family: var(--tz-font-admin-ui);',
  )

const adminFontAuthorityPlugin = (): Plugin => ({
  name: 'admin-font-authority',
  enforce: 'post',
  transform(code, id) {
    if (!id.includes('.css')) return null

    const normalized = normalizeAdminFontFallbacks(code)
    return normalized === code ? null : { code: normalized, map: null }
  },
  generateBundle(_options, bundle) {
    for (const output of Object.values(bundle)) {
      if (output.type !== 'asset' || !output.fileName.endsWith('.css')) continue
      output.source = normalizeAdminFontFallbacks(String(output.source))
    }
  },
})

export default defineConfig({
  plugins: [vue(), tailwindcss(), adminFontAuthorityPlugin()],
  resolve: {
    alias: {
      '@': resolve(projectRoot, 'src')
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
      },
      '/uploads': {
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
