// https://nuxt.com/docs/api/configuration/nuxt-config
import { defineNuxtConfig } from 'nuxt/config'
import locales from './app/i18n/locales.manifest.js'

const env = ((globalThis as unknown as { process?: { env?: Record<string, string | undefined> } }).process?.env) || {}
const trimTrailingSlash = (value: string) => value.replace(/\/$/, '')
const publicApiBase = trimTrailingSlash(
  env.NUXT_PUBLIC_API_BASE || env.GO_API_BASE || env.API_BASE || ''
)
const wpApiBase = trimTrailingSlash(
  env.NUXT_PUBLIC_WP_API_BASE || env.WP_API_BASE || (publicApiBase ? `${publicApiBase}/wp-json` : '/wp-json')
)

export default defineNuxtConfig({
  extends: ['./layers/admin', './layers/shop'],
  compatibilityDate: '2025-07-15',
  // 使用 app 作为源码目录，启用 app/pages 与 app/components
  srcDir: 'app',

  // Long cache for local Twemoji flags
  routeRules: {
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

  app: {
    baseURL: '/',
    buildAssetsDir: '_nuxt/',
    cdnURL: process.env.CDN_URL || 'https://cdn.tanzanite.site/',
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
    preset: 'cloudflare-pages'
  },

  // 配置 WordPress API 端点
  runtimeConfig: {
    public: {
      apiBase: publicApiBase,
      wpApiBase,
      blogApiMode: env.NUXT_PUBLIC_BLOG_API_MODE || env.BLOG_API_MODE || 'auto',
      siteTitle: env.NUXT_SITE_TITLE || 'Tanzanite',
      siteUrl: env.NUXT_SITE_URL || 'https://tanzanite.site',
      googleClientId: env.GOOGLE_CLIENT_ID || '',
      socialLinks: env.NUXT_SOCIAL_LINKS
        ? JSON.parse(env.NUXT_SOCIAL_LINKS)
        : []
    }
  },

  devtools: false
})
