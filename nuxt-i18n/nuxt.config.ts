// https://nuxt.com/docs/api/configuration/nuxt-config
const env = ((globalThis as unknown as { process?: { env?: Record<string, string | undefined> } }).process?.env) || {}

export default defineNuxtConfig({
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
  },

  site: {
    url: 'https://tanzanite.site',
  },

  modules: ['@nuxtjs/i18n', '@nuxtjs/sitemap'],

  i18n: {
    locales: [
      {
        code: 'en',
        iso: 'en-US',
        name: 'English',
        file: 'en.json'
      },
      {
        code: 'fr',
        iso: 'fr-FR',
        name: 'Français',
        file: 'fr.json'
      },
      {
        code: 'de',
        iso: 'de-DE',
        name: 'Deutsch',
        file: 'de.json'
      },
      {
        code: 'es',
        iso: 'es-ES',
        name: 'Español',
        file: 'es.json'
      },
      {
        code: 'ja',
        iso: 'ja-JP',
        name: '日本語',
        file: 'ja.json'
      },
      {
        code: 'ko',
        iso: 'ko-KR',
        name: '한국어',
        file: 'ko.json'
      },
      {
        code: 'it',
        iso: 'it-IT',
        name: 'Italiano',
        file: 'it.json'
      },
      {
        code: 'pt',
        iso: 'pt-PT',
        name: 'Português',
        file: 'pt.json'
      },
      {
        code: 'ru',
        iso: 'ru-RU',
        name: 'Русский',
        file: 'ru.json'
      },
      {
        code: 'ar',
        iso: 'ar-SA',
        name: 'العربية',
        file: 'ar.json',
        dir: 'rtl'
      },
      {
        code: 'fi',
        iso: 'fi-FI',
        name: 'Suomi',
        file: 'fi.json'
      },
      {
        code: 'da',
        iso: 'da-DK',
        name: 'Dansk',
        file: 'da.json'
      },
      {
        code: 'th',
        iso: 'th-TH',
        name: 'ไทย',
        file: 'th.json'
      },
      {
        code: 'sv',
        iso: 'sv-SE',
        name: 'Svenska',
        file: 'sv.json'
      },
      {
        code: 'id',
        iso: 'id-ID',
        name: 'Bahasa Indonesia',
        file: 'id.json'
      },
      {
        code: 'ms',
        iso: 'ms-MY',
        name: 'Bahasa Melayu',
        file: 'ms.json'
      },
      {
        code: 'be',
        iso: 'be-BY',
        name: 'Беларуская',
        file: 'be.json'
      },
      {
        code: 'tr',
        iso: 'tr-TR',
        name: 'Türkçe',
        file: 'tr.json'
      },
      {
        code: 'bn',
        iso: 'bn-BD',
        name: 'বাংলা',
        file: 'bn.json'
      },
      {
        code: 'fa',
        iso: 'fa-IR',
        name: 'فارسی',
        file: 'fa.json',
        dir: 'rtl'
      },
      {
        code: 'nl',
        iso: 'nl-NL',
        name: 'Nederlands',
        file: 'nl.json'
      },
      {
        code: 'hi',
        iso: 'hi-IN',
        name: 'हिन्दी',
        file: 'hi.json'
      },
      {
        code: 'ur',
        iso: 'ur-PK',
        name: 'اردو',
        file: 'ur.json',
        dir: 'rtl'
      },
      {
        code: 'mr',
        iso: 'mr-IN',
        name: 'मराठी',
        file: 'mr.json'
      },
      {
        code: 'pcm',
        iso: 'pcm-NG',
        name: 'Nigerian Pidgin',
        file: 'pcm.json'
      },
      {
        code: 'fil',
        iso: 'fil-PH',
        name: 'Filipino',
        file: 'fil.json'
      },
      {
        code: 'te',
        iso: 'te-IN',
        name: 'తెలుగు',
        file: 'te.json'
      },
      {
        code: 'ha',
        iso: 'ha-NG',
        name: 'Hausa',
        file: 'ha.json'
      },
      {
        code: 'ps',
        iso: 'ps-AF',
        name: 'پښتو',
        file: 'ps.json',
        dir: 'rtl'
      },
      {
        code: 'sw',
        iso: 'sw-KE',
        name: 'Kiswahili',
        file: 'sw.json'
      },
      {
        code: 'tl',
        iso: 'tl-PH',
        name: 'Tagalog',
        file: 'tl.json'
      },
      {
        code: 'ta',
        iso: 'ta-IN',
        name: 'தமிழ்',
        file: 'ta.json'
      },
      {
        code: 'jv',
        iso: 'jv-ID',
        name: 'Basa Jawa',
        file: 'jv.json'
      },
      {
        code: 'zh_cn',
        iso: 'zh-CN',
        name: '简体中文',
        file: 'zh_cn.json'
      }
    ],
    lazy: true,
    langDir: '../i18n/locales',
    defaultLocale: 'en',
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
    preset: 'static'
  },

  // 配置 WordPress API 端点
  runtimeConfig: {
    public: {
      wpApiBase: env.WP_API_BASE || 'https://tanzanite.site/wp-json',
      siteTitle: env.NUXT_SITE_TITLE || 'Tanzanite',
      siteUrl: env.NUXT_SITE_URL || 'https://tanzanite.site',
      socialLinks: env.NUXT_SOCIAL_LINKS
        ? JSON.parse(env.NUXT_SOCIAL_LINKS)
        : []
    }
  },

  devtools: false
})
