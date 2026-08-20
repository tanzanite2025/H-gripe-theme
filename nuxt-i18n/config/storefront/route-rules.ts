import {
  storefrontHtmlCachePolicies,
  storefrontNoStorePagePaths,
  type StorefrontHtmlCachePolicy,
} from './html-cache-policies'
import { htmlRouteCacheStorageBase } from './html-cache-runtime'

type StorefrontRouteRule = {
  proxy?: string
  redirect?: string | {
    to: string
    statusCode: 301 | 308
  }
  headers?: Record<string, string>
  cache?: {
    base: string
    maxAge: number
    staleMaxAge: number
    swr: true
  }
}

type StorefrontRouteRules = Record<string, StorefrontRouteRule>

interface BuildStorefrontRouteRulesOptions {
  internalApiOrigin: string
  localeCodes: string[]
  defaultLocale: string
  htmlCacheEnabled?: boolean
}

const immutableAssetHeaders = {
  'cache-control': 'public, max-age=31536000, immutable',
}

const immutableFontHeaders = {
  ...immutableAssetHeaders,
  'access-control-allow-origin': '*',
}

const revalidatingFontStylesheetHeaders = {
  'cache-control': 'no-cache',
  'access-control-allow-origin': '*',
}

const noStoreHeaders = {
  'cache-control': 'no-store, max-age=0',
}

const immutableAssetRoutePatterns = [
  '/_nuxt/**',
  '/company/aboutus/**',
  '/company/globalpartners/**',
  '/company/ourstory/factory/**',
  '/company/ourstory/holepatterns/**',
  '/company/ourstory/ourstory/**',
  '/icons/**',
  '/images/**',
  '/testreport/**',
  '/twemoji/svg/**',
  '/whatsappmodel/**',
]

const storefrontRedirects = [
  {
    from: '/guides/technical',
    to: '/guides',
    statusCode: 301,
  },
  {
    from: '/faq',
    to: '/support/faqs',
    statusCode: 301,
  },
] as const

const normalizePath = (path: string) => {
  if (path === '/') return '/'
  return `/${path.replace(/^\/+/, '')}`.replace(/\/+$/, '')
}

const localizedPaths = (path: string, localeCodes: string[], defaultLocale: string) => {
  const normalized = normalizePath(path)
  const paths = normalized === '/' ? [] : [normalized]

  for (const locale of localeCodes) {
    if (!locale || locale === defaultLocale) continue
    paths.push(`/${locale}${normalized === '/' ? '' : normalized}`)
  }

  return paths
}

const cacheRule = (policy: Pick<StorefrontHtmlCachePolicy, 'maxAge' | 'staleMaxAge'>) => ({
  cache: {
    base: htmlRouteCacheStorageBase,
    maxAge: policy.maxAge,
    staleMaxAge: policy.staleMaxAge,
    swr: true,
  },
} satisfies StorefrontRouteRule)

const addLocalizedRule = (
  routeRules: StorefrontRouteRules,
  paths: string[],
  localeCodes: string[],
  defaultLocale: string,
  rule: StorefrontRouteRule,
) => {
  for (const path of paths) {
    for (const localizedPath of localizedPaths(path, localeCodes, defaultLocale)) {
      routeRules[localizedPath] = rule
    }
  }
}

const addLocalizedRedirect = (
  routeRules: StorefrontRouteRules,
  from: string,
  to: string,
  statusCode: 301 | 308,
  localeCodes: string[],
  defaultLocale: string,
) => {
  routeRules[normalizePath(from)] = {
    redirect: {
      to: normalizePath(to),
      statusCode,
    },
  }

  for (const locale of localeCodes) {
    if (!locale || locale === defaultLocale) continue
    routeRules[`/${locale}${normalizePath(from)}`] = {
      redirect: {
        to: `/${locale}${normalizePath(to)}`,
        statusCode,
      },
    }
  }
}

export const buildStorefrontRouteRules = ({
  internalApiOrigin,
  localeCodes,
  defaultLocale,
  htmlCacheEnabled = true,
}: BuildStorefrontRouteRulesOptions) => {
  const routeRules: StorefrontRouteRules = {
    '/api/**': {
      proxy: `${internalApiOrigin.replace(/\/$/, '')}/api/**`,
      headers: noStoreHeaders,
    },
    '/': {
      // "/" can redirect based on the i18n cookie. Do not cache that decision.
      headers: noStoreHeaders,
    },
    '/_internal/**': {
      headers: noStoreHeaders,
    },
  }

  for (const pattern of immutableAssetRoutePatterns) {
    routeRules[pattern] = {
      headers: immutableAssetHeaders,
    }
  }

  routeRules['/fonts/**'] = {
    headers: immutableFontHeaders,
  }

  // Stripe loads this stable path inside an iframe; its referenced font files are content-addressed.
  routeRules['/fonts/maple-ui.css'] = {
    headers: revalidatingFontStylesheetHeaders,
  }

  if (htmlCacheEnabled) {
    for (const policy of storefrontHtmlCachePolicies) {
      addLocalizedRule(routeRules, policy.paths, localeCodes, defaultLocale, cacheRule(policy))
    }
  }

  addLocalizedRule(routeRules, storefrontNoStorePagePaths, localeCodes, defaultLocale, {
    headers: noStoreHeaders,
  })

  for (const redirect of storefrontRedirects) {
    addLocalizedRedirect(routeRules, redirect.from, redirect.to, redirect.statusCode, localeCodes, defaultLocale)
  }

  return routeRules
}
