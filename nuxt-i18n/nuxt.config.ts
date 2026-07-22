// https://nuxt.com/docs/api/configuration/nuxt-config
import { defineNuxtConfig } from 'nuxt/config'
import locales from './app/i18n/locales.manifest.js'

const env = ((globalThis as unknown as { process?: { env?: Record<string, string | undefined> } }).process?.env) || {}
const trimTrailingSlash = (value: string) => value.replace(/\/$/, '')
const publicApiBase = trimTrailingSlash(
  env.NUXT_PUBLIC_API_BASE || env.GO_API_BASE || env.API_BASE || ''
)
const internalApiOrigin = trimTrailingSlash(env.API_INTERNAL_ORIGIN || 'http://localhost:9000')

type RollupBuildWarning = {
  code?: string
  id?: string
  message?: string
  plugin?: string
}

const normalizeModuleId = (id: string) => id.replace(/\\/g, '/')

const getManualChunkName = (id: string) => {
  const moduleId = normalizeModuleId(id)
  if (!moduleId.includes('/node_modules/')) return undefined

  if (
    moduleId.includes('/vue/') ||
    moduleId.includes('/@vue/') ||
    moduleId.includes('/vue-router/') ||
    moduleId.includes('/pinia/') ||
    moduleId.includes('/@pinia/')
  ) {
    return 'vendor-vue'
  }

  if (
    moduleId.includes('/@intlify/') ||
    moduleId.includes('/vue-i18n/') ||
    moduleId.includes('/@nuxtjs/i18n/')
  ) {
    return 'vendor-i18n'
  }

  if (
    moduleId.includes('/@iconify/') ||
    moduleId.includes('/@nuxt/icon/')
  ) {
    return 'vendor-icons'
  }

  if (
    moduleId.includes('/@vueuse/') ||
    moduleId.includes('/focus-trap/') ||
    moduleId.includes('/zod/')
  ) {
    return 'vendor-utils'
  }

  return undefined
}

const shouldIgnoreBuildWarning = (warning: RollupBuildWarning) => {
  const message = warning.message || ''
  const id = normalizeModuleId(warning.id || '')

  return (
    (
      warning.plugin === 'nuxt:module-preload-polyfill' &&
      message.includes('Sourcemap is likely to be incorrect')
    ) ||
    (
      warning.code === 'INVALID_ANNOTATION' &&
      message.includes('#__PURE__') &&
      (id.includes('/node_modules/@vueuse/core/dist/index.js') || message.includes('@vueuse/core'))
    )
  )
}

const shouldIgnoreNitroBuildWarning = (warning: RollupBuildWarning) => {
  const message = warning.message || ''

  return (
    shouldIgnoreBuildWarning(warning) ||
    ['CIRCULAR_DEPENDENCY', 'EVAL'].includes(warning.code || '') ||
    message.includes('Unsupported source map comment')
  )
}

export default defineNuxtConfig({
  extends: ['./layers/admin', './layers/shop'],
  compatibilityDate: '2025-07-15',
  // 使用 app 作为源码目录，启用 app/pages 与 app/components
  srcDir: 'app',

  // Long cache for local Twemoji flags
  routeRules: {
    '/api/**': {
      proxy: `${internalApiOrigin}/api/**`
    },
    '/twemoji/svg/**': {
      headers: {
        'cache-control': 'public, max-age=31536000, immutable'
      }
    },
    '/company/about': {
      redirect: '/company/ourstory'
    },
    '/guides/technical': {
      redirect: '/guides'
    },
  },

  // @ts-expect-error: site config supported by Nuxt runtime, not in TS defs
  site: {
    url: 'https://tanzanite.site',
  },

  modules: ['@nuxtjs/i18n', '@nuxtjs/sitemap', '@nuxt/image', '@pinia/nuxt', '@nuxt/icon', '@nuxt/fonts'],

  icon: {
    localApiEndpoint: '/_nuxt_icon',
  },

  i18n: {
    restructureDir: 'app',
    locales,
    lazy: true,
    langDir: 'i18n/locales',
    defaultLocale: 'en',
    fallbackLocale: 'en',
    strategy: 'prefix_except_default',
    detectBrowserLanguage: {
      useCookie: true,
      cookieKey: 'i18n_redirected',
      redirectOn: 'root',
      alwaysRedirect: false,
      fallbackLocale: 'en'
    },
    bundle: {
      optimizeTranslationDirective: false
    },
    baseUrl: 'https://tanzanite.site',
  },

  css: [
    '~/assets/css/tailwind.css',
    '~/assets/css/guide-sections.css',
    '~/assets/css/components/nav.css',
    '~/assets/css/components/whatsapp-mobile-drawer.css',
  ],

  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {},
    },
  },

  vite: {
    build: {
      rollupOptions: {
        output: {
          manualChunks: getManualChunkName,
        },
        onwarn(warning, warn) {
          if (shouldIgnoreBuildWarning(warning as RollupBuildWarning)) return
          warn(warning)
        },
        onLog(level, log, handler) {
          if (level === 'warn' && shouldIgnoreBuildWarning(log as RollupBuildWarning)) return
          handler(level, log)
        },
      },
    },
  },

  app: {
    baseURL: '/',
    buildAssetsDir: '_nuxt/',
    cdnURL: process.env.CDN_URL || undefined,
    head: {
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' }
      ],
    }
  },

  // 启用默认的 SSR + 预渲染，以便生成完整静态 HTML
  ssr: true,

  nitro: {
    preset: env.NITRO_PRESET || 'node-server',
    rollupConfig: {
      onwarn(warning, warn) {
        if (shouldIgnoreNitroBuildWarning(warning as RollupBuildWarning)) return
        warn(warning)
      },
      onLog(level, log, handler) {
        if (level === 'warn' && shouldIgnoreNitroBuildWarning(log as RollupBuildWarning)) return
        handler(level, log)
      },
    },
  },

  runtimeConfig: {
    public: {
      apiBase: publicApiBase,
      blogApiMode: env.NUXT_PUBLIC_BLOG_API_MODE || env.BLOG_API_MODE || 'auto',
      siteTitle: env.NUXT_SITE_TITLE || 'Tanzanite',
      siteUrl: env.NUXT_SITE_URL || 'https://tanzanite.site',
      googleClientId: env.NUXT_PUBLIC_GOOGLE_CLIENT_ID || env.GOOGLE_CLIENT_ID || '',
      socialLinks: env.NUXT_SOCIAL_LINKS
        ? JSON.parse(env.NUXT_SOCIAL_LINKS)
        : []
    }
  },

  devtools: false
})
