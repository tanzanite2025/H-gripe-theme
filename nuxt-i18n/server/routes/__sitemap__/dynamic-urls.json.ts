type ProductSitemapRecord = {
  slug?: string
  locale?: string
  updated_at?: string
}

type PostSitemapRecord = {
  slug?: string
  locale?: string
  updated_at?: string
  published_at?: string | null
}

type FAQPageSitemapRecord = {
  route_path?: string
  locale?: string
  updated_at?: string
}

type ProductListResponse = {
  data?: ProductSitemapRecord[]
  has_more?: boolean
}

type FAQPageListResponse = {
  pages?: FAQPageSitemapRecord[]
}

type PostListResponse = {
  data?: PostSitemapRecord[]
  total_pages?: number
}

type SitemapPriority = 0 | 0.1 | 0.2 | 0.3 | 0.4 | 0.5 | 0.6 | 0.7 | 0.8 | 0.9 | 1

type SitemapEntry = {
  loc: string
  lastmod?: string
  changefreq: 'weekly' | 'monthly'
  priority: SitemapPriority
  _i18nTransform: true
}

const maxPages = 100
const pageSize = 24

const normalizePath = (value: unknown) => {
  const path = String(value || '').trim()
  if (!path || !path.startsWith('/')) return ''
  return `/${path.replace(/^\/+/, '').replace(/\/+$/, '')}`
}

const getLastmod = (...values: unknown[]) => {
  for (const value of values) {
    const date = String(value || '').trim()
    if (date) return date
  }
  return undefined
}

const addEntry = (entries: SitemapEntry[], seen: Set<string>, entry: SitemapEntry) => {
  if (!entry.loc || seen.has(entry.loc)) return
  seen.add(entry.loc)
  entries.push(entry)
}

const fetchAllProducts = async (apiOrigin: string) => {
  const products: ProductSitemapRecord[] = []

  for (let page = 1; page <= maxPages; page += 1) {
    const response = await $fetch<ProductListResponse>(`${apiOrigin}/api/v1/products`, {
      params: {
        locale: 'en',
        status: 'active',
        page,
        page_size: pageSize,
      },
    })
    const items = Array.isArray(response?.data) ? response.data : []
    products.push(...items)

    if (items.length === 0 || response?.has_more !== true || items.length < pageSize) break
  }

  return products
}

const fetchAllPosts = async (apiOrigin: string) => {
  const posts: PostSitemapRecord[] = []

  for (let page = 1; page <= maxPages; page += 1) {
    const response = await $fetch<PostListResponse>(`${apiOrigin}/api/v1/content/posts`, {
      params: {
        locale: 'en',
        status: 'published',
        page,
        page_size: pageSize,
      },
    })
    const items = Array.isArray(response?.data) ? response.data : []
    posts.push(...items)

    const totalPages = Number(response?.total_pages || 0)
    if (items.length === 0 || (totalPages > 0 && page >= totalPages) || items.length < pageSize) break
  }

  return posts
}

export default defineSitemapEventHandler(async () => {
  const config = useRuntimeConfig()
  const apiOrigin = String(config.apiInternalOrigin || '').replace(/\/$/, '')
  if (!apiOrigin) return []

  const entries: SitemapEntry[] = []
  const seen = new Set<string>()

  try {
    const [products, faqResponse, posts] = await Promise.all([
      fetchAllProducts(apiOrigin),
      $fetch<FAQPageListResponse>(`${apiOrigin}/api/v1/content/faq-pages`, {
        params: { locale: 'en' },
      }),
      fetchAllPosts(apiOrigin),
    ])

    for (const product of products) {
      const slug = String(product.slug || '').trim()
      if (!slug || product.locale && product.locale !== 'en') continue

      addEntry(entries, seen, {
        loc: `/shop/${encodeURIComponent(slug)}`,
        lastmod: getLastmod(product.updated_at),
        changefreq: 'weekly',
        priority: 0.8,
        _i18nTransform: true,
      })
    }

    for (const page of Array.isArray(faqResponse?.pages) ? faqResponse.pages : []) {
      const routePath = normalizePath(page.route_path)
      if (!routePath || page.locale && page.locale !== 'en') continue

      addEntry(entries, seen, {
        loc: routePath,
        lastmod: getLastmod(page.updated_at),
        changefreq: 'monthly',
        priority: 0.6,
        _i18nTransform: true,
      })
    }

    for (const post of posts) {
      const slug = String(post.slug || '').trim()
      if (!slug || post.locale && post.locale !== 'en') continue

      addEntry(entries, seen, {
        loc: `/blog/${encodeURIComponent(slug)}`,
        lastmod: getLastmod(post.updated_at, post.published_at),
        changefreq: 'monthly',
        priority: 0.6,
        _i18nTransform: true,
      })
    }
  } catch (error) {
    console.error('[sitemap] failed to load backend URLs', error)
  }

  return entries
})
