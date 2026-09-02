export const MAX_STOREFRONT_PRECONNECT_ORIGINS = 4

export const STOREFRONT_LOADING_TIERS = {
  criticalSsr: [
    'SiteHeader',
    'HomeHero',
    'GradientDockMenuShell',
  ],
  visibleHydration: [
    'SiteHeader',
    'HomeHero',
  ],
  nearFoldHydration: [
    'HomeRideCategoryStrip',
    'HomeMainProductCategories',
    'HomeStorePicksGuide',
    'HomePurchasePath',
    'HomeFeaturedProducts',
    'HomeShopWithConfidence',
    'HomeFeaturesTabs',
    'HomeFaqPreview',
    'HomeFinalCta',
    'AppFooter',
  ],
  intentOnly: [
    'SidePanel',
    'CartDrawer',
    'CheckoutModal',
    'ShopSearchSheet',
    'GlobalPagesSearchOverlay',
    'GlobalProductDetailBottomSheet',
    'WhatsAppChatModal',
    'GradientDockQuickBuy',
  ],
  idleBackground: [
    'GradientDockMenu',
    'CookieConsent',
    'BehaviorAttributionBootstrap',
    'analytics.client',
    'device-fingerprint.client',
  ],
} as const

// Storefront loading contract:
// 1. SSR immediately: hero, header shell, dock shell.
// 2. Near-fold hydration: sections that are likely to become visible next.
// 3. Intent-only: drawers, overlays, chat, cart, quick-buy.
// 4. Idle/background: analytics, fingerprinting, attribution, consent shell.
export const STOREFRONT_NEAR_FOLD_HYDRATION_OPTIONS = {
  rootMargin: '0px 0px 96px 0px',
} as const

export const STOREFRONT_FOOTER_HYDRATION_OPTIONS = {
  rootMargin: '0px 0px 256px 0px',
} as const

export const STOREFRONT_IDLE_CLIENT_WORK = {
  delayMs: 15000,
  idleTimeoutMs: 5000,
} as const

export const STOREFRONT_READ_COUNT_WARMUP = {
  delayMs: 5000,
  idleTimeoutMs: 2000,
} as const

export const STOREFRONT_SESSION_WARMUP = {
  delayMs: 6500,
  idleTimeoutMs: 3000,
} as const

export const STOREFRONT_CATEGORY_PREFETCH_WARMUP = {
  delayMs: 7500,
  idleTimeoutMs: 5000,
} as const

export const STOREFRONT_HERO_DESKTOP_MEDIA_QUERY = '(min-width: 1024px)'

export const STOREFRONT_HERO_MOBILE_MEDIA_QUERY = '(max-width: 1023px)'

export const STOREFRONT_ANALYTICS_WARMUP = {
  delayMs: 7000,
  idleTimeoutMs: 4000,
} as const

export const STOREFRONT_DEVICE_FINGERPRINT_WARMUP = {
  delayMs: 7000,
  idleTimeoutMs: 3000,
} as const

export const HOME_HERO_SHOWCASE_SSR_TIMEOUT_MS = 900

export const isStorefrontMobileUserAgent = (userAgent: unknown): boolean => {
  const value = String(userAgent || '').toLowerCase()
  return /\b(android|iphone|ipod|ipad|iemobile|opera mini|mobile)\b/.test(value)
}

export type StorefrontPreconnectLink = {
  key: string
  rel: 'preconnect'
  href: string
  crossorigin?: 'anonymous'
}

const resolveOrigin = (value: unknown, baseOrigin: string): string => {
  const candidate = String(value || '').trim()
  if (!candidate) return ''

  try {
    const url = candidate.includes('://')
      ? new URL(candidate)
      : new URL(candidate.startsWith('/') ? candidate : `https://${candidate}`, baseOrigin)
    return ['http:', 'https:'].includes(url.protocol) ? url.origin : ''
  } catch {
    return ''
  }
}

export const splitStorefrontConfiguredOrigins = (value: unknown): string[] => (
  String(value || '')
    .split(',')
    .map(item => item.trim())
    .filter(Boolean)
)

const normalizeOrigin = (origin: string) => origin.replace(/\/$/, '').toLowerCase()

const isInternalPreconnectHostname = (hostname: string): boolean => {
  const host = hostname.trim().replace(/\.$/, '').toLowerCase()
  if (!host) return true

  if (
    host === 'localhost' ||
    host === 'host.docker.internal' ||
    host.endsWith('.localhost') ||
    host.endsWith('.local') ||
    host.endsWith('.internal')
  ) {
    return true
  }

  if (!host.includes('.') && !host.includes(':')) return true
  if (/^(127|10)\./.test(host)) return true
  if (/^192\.168\./.test(host)) return true
  if (/^169\.254\./.test(host)) return true

  const private172Match = host.match(/^172\.(\d{1,2})\./)
  if (private172Match) {
    const block = Number(private172Match[1])
    if (block >= 16 && block <= 31) return true
  }

  return host === '::1' || host.startsWith('fc') || host.startsWith('fd') || host.startsWith('fe80:')
}

const isUsablePreconnectOrigin = (origin: string, excludedOrigins: Set<string>) => {
  try {
    const url = new URL(origin)
    if (!['http:', 'https:'].includes(url.protocol)) return false
    if (excludedOrigins.has(normalizeOrigin(url.origin))) return false
    return !isInternalPreconnectHostname(url.hostname)
  } catch {
    return false
  }
}

export const createStorefrontPreconnectLinks = (
  candidates: unknown[],
  currentOrigin: string,
  options: {
    siteOrigin?: unknown
    maxOrigins?: number
  } = {},
): StorefrontPreconnectLink[] => {
  const publicSiteOrigin = resolveOrigin(options.siteOrigin, currentOrigin)
  const excludedOrigins = new Set(
    [currentOrigin, publicSiteOrigin]
      .map(origin => normalizeOrigin(origin || ''))
      .filter(Boolean),
  )

  return Array.from(new Set(
    candidates
      .map(candidate => resolveOrigin(candidate, currentOrigin))
      .filter(origin => isUsablePreconnectOrigin(origin, excludedOrigins)),
  ))
    .slice(0, options.maxOrigins ?? MAX_STOREFRONT_PRECONNECT_ORIGINS)
    .map(origin => ({
      key: `storefront-preconnect:${origin}`,
      rel: 'preconnect',
      href: origin,
      crossorigin: 'anonymous',
    }))
}
