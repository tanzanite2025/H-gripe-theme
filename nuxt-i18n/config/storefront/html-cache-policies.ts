export interface StorefrontHtmlCachePolicy {
  name: string
  description: string
  paths: string[]
  maxAge: number
  staleMaxAge: number
}

const hour = 60 * 60
const day = 24 * hour
const week = 7 * day

export const storefrontHtmlCacheDurations = {
  contentPage: {
    maxAge: hour,
    staleMaxAge: day,
  },
  stablePolicyPage: {
    maxAge: day,
    staleMaxAge: week,
  },
} as const

export const storefrontHtmlCachePolicies: StorefrontHtmlCachePolicy[] = [
  {
    name: 'content',
    description: 'Editorial and guide pages that change less frequently than product data.',
    paths: ['/resources/blog/**', '/guides/**', '/resources/picture-warehouse/**', '/faq'],
    ...storefrontHtmlCacheDurations.contentPage,
  },
  {
    name: 'stable-policy',
    description: 'Company, policy, and support content with long-lived public HTML.',
    paths: [
      '/company/**',
      '/policies/**',
      '/support/faqs',
      '/support/payment',
      '/support/shipping',
      '/support/warranty',
    ],
    ...storefrontHtmlCacheDurations.stablePolicyPage,
  },
]

export const storefrontNoStorePagePaths = [
  // Home HTML carries build-specific modulepreload links. Keep localized home pages
  // out of stale HTML cache so deploys cannot reference retired _nuxt assets.
  '/',
  // Product HTML contains price, exchange-rate, configuration, and availability
  // snapshots. Transactional APIs must be consulted before a purchase.
  '/products/**',
  // Query/search/filter state lives in the URL, but the list is data-heavy and
  // changes frequently. Keep it uncached until a query-aware cache strategy is added.
  '/shop',
  '/resources/membershipandpoints/**',
  '/resources/spoke-calculator/**',
  '/resources/workbench',
  '/support/test-report',
  '/support/warranty-check',
]
