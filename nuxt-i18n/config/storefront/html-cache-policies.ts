export interface StorefrontHtmlCachePolicy {
  name: string
  description: string
  paths: string[]
  maxAge: number
  staleMaxAge: number
}

const minute = 60
const hour = 60 * minute
const day = 24 * hour
const week = 7 * day

export const storefrontHtmlCacheDurations = {
  productDetail: {
    // Product HTML may contain price/stock snapshots. Keep the fresh TTL short;
    // cart and checkout APIs remain the source of truth for transactional data.
    maxAge: 5 * minute,
    staleMaxAge: hour,
  },
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
    name: 'product-detail',
    description: 'Individual product detail pages with short-lived SSR HTML.',
    paths: ['/products/**'],
    ...storefrontHtmlCacheDurations.productDetail,
  },
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
  // Query/search/filter state lives in the URL, but the list is data-heavy and
  // changes frequently. Keep it uncached until a query-aware cache strategy is added.
  '/shop',
  '/resources/membershipandpoints/**',
  '/resources/spoke-calculator/**',
  '/support/test-report',
  '/support/warranty-check',
]
