// https://nuxt.com/docs/api/configuration/nuxt-config
import { defineNuxtConfig } from 'nuxt/config'
import locales from './app/i18n/locales.manifest'
import { buildStorefrontRouteRules } from './config/storefront/route-rules'

const env = ((globalThis as unknown as { process?: { env?: Record<string, string | undefined> } }).process?.env) || {}
const trimTrailingSlash = (value: string) => value.replace(/\/$/, '')
const isTruthyEnv = (value: string | undefined) => ['1', 'true', 'yes', 'on'].includes(
  String(value || '').trim().toLowerCase()
)
const hostFromUrl = (value: string) => {
  if (!value) return ''

  try {
    return new URL(value).host
  } catch {
    return ''
  }
}
const publicApiBase = trimTrailingSlash(
  env.NUXT_PUBLIC_API_BASE || env.GO_API_BASE || env.API_BASE || ''
)
const publicSiteUrl = trimTrailingSlash(env.NUXT_SITE_URL || env.NUXT_PUBLIC_SITE_URL || '')
const internalApiOrigin = trimTrailingSlash(env.API_INTERNAL_ORIGIN || 'http://localhost:9200')
const imageProvider = env.NUXT_IMAGE_PROVIDER || 'ipx'
const imageInternalOrigin = trimTrailingSlash(env.NUXT_IMAGE_INTERNAL_ORIGIN || '')
const imageMaxInputPixels = Number(env.NUXT_IMAGE_MAX_INPUT_PIXELS || '24000000')
const imageOptimizeUploads = isTruthyEnv(
  env.NUXT_IMAGE_OPTIMIZE_UPLOADS || env.NUXT_PUBLIC_IMAGE_UPLOAD_OPTIMIZATION_ENABLED
)
const siteImageDomain = hostFromUrl(publicSiteUrl)
const apiImageDomain = hostFromUrl(publicApiBase)
const internalImageDomain = hostFromUrl(imageInternalOrigin)
const configuredImageDomains = (env.NUXT_IMAGE_DOMAINS || '')
  .split(',')
  .map(value => value.trim())
  .filter(Boolean)
  .map(value => hostFromUrl(value.includes('://') ? value : `https://${value}`) || value)
const imageDomains = [...new Set([
  siteImageDomain,
  apiImageDomain,
  internalImageDomain,
  ...configuredImageDomains,
].filter(Boolean))]
const htmlCacheDefault = env.NODE_ENV === 'production' ? 'true' : 'false'
const htmlCacheEnabled = String(env.NUXT_HTML_CACHE_ENABLED ?? htmlCacheDefault).toLowerCase() !== 'false'

const tabbedPageRoutes = [
  {
    basePath: '/guides/tireguides',
    tabs: ['size', 'match', 'tubeless', 'installation', 'choose', 'rims', 'tube'],
  },
  {
    basePath: '/guides/wheelset-buyers',
    tabs: ['overview', 'safety-instructions', 'sample-assembly', 'special-order', 'appearance-logo', 'choose-freehub', 'wheel-components', 'optional'],
  },
  {
    basePath: '/company/about',
    tabs: ['factory', 'appearance', 'hole-patterns', 'facility', 'manufacture', 'qualitycontrol'],
  },
  {
    basePath: '/support/warranty',
    tabs: ['change-cancel', 'damaged-lost', 'returns', 'warranty', 'accidental-damage', 'protection', 'submit-warranty'],
  },
  {
    basePath: '/support/test-report',
    tabs: ['rim-test-report', 'wheelset-test-report', 'tension', 'wheelset-assembly'],
  },
  {
    basePath: '/spoke-calculator',
    tabs: ['calculator', 'parameter'],
  },
  {
    basePath: '/membershipandpoints',
    tabs: ['myinfo', 'levers', 'exchange'],
  },
  {
    basePath: '/picture-warehouse',
    tabs: ['riders', 'brand'],
  },
]

const normalizePagePath = (path: string) => `/${path.replace(/^\/+/, '')}`.replace(/\/+$/, '') || '/'

const addTabbedPageRoutes = (pages: any[]) => {
  for (const entry of tabbedPageRoutes) {
    const basePath = normalizePagePath(entry.basePath)
    const basePage = pages.find(page => normalizePagePath(page.path || '') === basePath)
    if (!basePage?.file) continue

    const name = String(basePage.name || basePath.replace(/[^\w]+/g, '-').replace(/^-|-$/g, ''))
    const tabPattern = entry.tabs.join('|')
    const tabPath = `${basePath}/:tab(${tabPattern})`

    if (pages.some(page => normalizePagePath(page.path || '') === tabPath)) continue

    pages.push({
      ...basePage,
      name: `${name}-tab`,
      path: tabPath,
    })
  }
}

const storefrontI18nLocales = locales.map((locale) => {
  const localeFile = locale.file || `${locale.code}.json`

  return {
    ...locale,
    language: locale.language || locale.iso || locale.code,
    file: localeFile,
    files: locale.files || [localeFile],
  }
})

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

const runtimeSharpTraceDeps = (() => {
  if (process.platform === 'linux') {
    const report = process.report?.getReport?.() as {
      header?: {
        glibcVersionRuntime?: string
      }
    } | undefined
    const glibcVersionRuntime = report?.header?.glibcVersionRuntime
    if (glibcVersionRuntime) {
      return [
        'sharp',
        '@img/sharp-linux-x64*',
        '@img/sharp-libvips-linux-x64*',
      ]
    }

    return [
      'sharp',
      '@img/sharp-linuxmusl-x64*',
      '@img/sharp-libvips-linuxmusl-x64*',
    ]
  }

  if (process.platform === 'win32') {
    return [
      'sharp',
      '@img/sharp-win32-x64*',
    ]
  }

  return ['sharp']
})()

export default defineNuxtConfig({
  extends: ['./layers/admin', './layers/shop'],
  compatibilityDate: '2025-07-15',
  // 使用 app 作为源码目录，启用 app/pages 与 app/components
  srcDir: 'app',

  // Centralized route policy: API proxy, immutable assets, public HTML cache, and no-store pages.
  routeRules: {
    ...buildStorefrontRouteRules({
      internalApiOrigin,
      localeCodes: storefrontI18nLocales.map(locale => locale.code),
      defaultLocale: 'en',
      htmlCacheEnabled,
    })
  },

  site: {
    url: publicSiteUrl,
  },

  sitemap: {
    sources: ['/__sitemap__/dynamic-urls.json'],
  },

  image: {
    // IPX generates only the display-sized derivative. Original uploaded files
    // remain untouched, and remote transforms are restricted to known domains.
    provider: imageProvider,
    domains: imageDomains,
    alias: imageInternalOrigin
      ? {
          '/uploads/': `${imageInternalOrigin}/uploads/`,
        }
      : {},
    densities: [1, 2],
    ipx: {
      fs: {
        maxAge: 60 * 60,
      },
      http: {
        maxAge: 60 * 60,
      },
      sharpOptions: {
        failOn: 'error',
        limitInputPixels: Number.isFinite(imageMaxInputPixels) && imageMaxInputPixels > 0
          ? Math.floor(imageMaxInputPixels)
          : 24_000_000,
      },
    },
  },

  modules: ['@nuxtjs/i18n', '@nuxtjs/sitemap', '@nuxt/image', '@pinia/nuxt', '@nuxt/icon'],

  hooks: {
    'pages:extend'(pages) {
      addTabbedPageRoutes(pages)
    },
  },

  icon: {
    localApiEndpoint: '/_nuxt_icon',
    // CSS mode inserts style elements at runtime. SVG keeps icon rendering
    // compatible with the hash-based CSP used by cached SSR HTML.
    mode: 'svg',
  },

  i18n: {
    restructureDir: 'app',
    locales: storefrontI18nLocales as any,
    lazy: true,
    langDir: 'i18n/locales',
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
    baseUrl: publicSiteUrl,
  },

  css: [
    '~/assets/css/tailwind.css',
    '~/assets/css/components/nav.css',
  ],

  postcss: {
    plugins: {
      tailwindcss: {},
      autoprefixer: {},
    },
  },

  vite: {
    server: {
      allowedHosts: ['host.docker.internal'],
    },
    build: {
      sourcemap: false,
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
    cdnURL: env.CDN_URL || undefined,
    head: {
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' }
      ],
      script: [
        {
          src: '/zod-global-config.js',
        },
      ],
    }
  },

  // 启用默认的 SSR + 预渲染，以便生成完整静态 HTML
  ssr: true,

  // Source maps are useful in development but should not ship with the
  // storefront runtime image.
  sourcemap: {
    client: false,
    server: false,
  },

  nitro: {
    preset: env.NITRO_PRESET || 'node-server',
    sourceMap: false,
    externals: {
      traceInclude: runtimeSharpTraceDeps,
    },
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
    apiInternalOrigin: internalApiOrigin,
    imageInternalOrigin,
    public: {
      apiBase: publicApiBase,
      blogApiMode: env.NUXT_PUBLIC_BLOG_API_MODE || env.BLOG_API_MODE || 'auto',
      siteTitle: env.NUXT_SITE_TITLE || '',
      siteUrl: publicSiteUrl,
      imageDomains: imageDomains.join(','),
      imageUploadOptimizationEnabled: Boolean(imageInternalOrigin && imageOptimizeUploads),
      imageUploadAliasEnabled: Boolean(imageInternalOrigin && imageOptimizeUploads),
      googleClientId: env.NUXT_PUBLIC_GOOGLE_CLIENT_ID || env.GOOGLE_CLIENT_ID || '',
      requestSigningKey: env.NUXT_PUBLIC_REQUEST_SIGNING_KEY || '',
      turnstileSiteKey: env.NUXT_PUBLIC_TURNSTILE_SITE_KEY || '',
      socialLinks: env.NUXT_SOCIAL_LINKS
        ? JSON.parse(env.NUXT_SOCIAL_LINKS)
        : []
    }
  },

  devtools: false
})
