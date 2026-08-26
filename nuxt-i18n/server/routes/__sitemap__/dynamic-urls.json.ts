import locales from '~/i18n/locales.manifest'

type ProductSitemapRecord = {
  slug?: string
  locale?: string
  updated_at?: string
}

type PostSitemapRecord = {
  slug?: string
  locale?: string
  tags?: string
  updated_at?: string
  published_at?: string | null
}

type FAQPageSitemapRecord = {
  route_path?: string
  locale?: string
  updated_at?: string
}

type ProductCategorySitemapRecord = {
  route_path?: string
  locale?: string
  updated_at?: string
  children?: ProductCategorySitemapRecord[]
}

type ProductListResponse = {
  data?: ProductSitemapRecord[]
  has_more?: boolean
}

type ProductCategoryListResponse = {
  data?: {
    tree?: ProductCategorySitemapRecord[]
    flat?: ProductCategorySitemapRecord[]
  }
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
  locale?: string
  source_type?: string
  canonical_path?: string
}

const maxPages = 100
const pageSize = 24
const defaultLocale = 'en'
const localeCodes = locales
  .map((locale) => String(locale.code || '').trim())
  .filter(Boolean)
const localeCodeSet = new Set(localeCodes)
const localeKey = (value: unknown) => String(value || '').trim().toLowerCase().replace(/-/g, '_')
const localeCodeByKey = new Map<string, string>()
for (const locale of locales) {
  const code = String(locale.code || '').trim().toLowerCase()
  if (!code) continue
  localeCodeByKey.set(localeKey(code), code)
  localeCodeByKey.set(localeKey(locale.iso), code)
  localeCodeByKey.set(localeKey(locale.language), code)
}

const normalizeLocale = (value: unknown, fallback = '') => {
  return localeCodeByKey.get(localeKey(value)) || fallback
}

const normalizePath = (value: unknown) => {
  const raw = String(value || '').trim()
  if (!raw || !raw.startsWith('/') || /[?#]/.test(raw)) return ''
  if (raw === '/') return '/'
  return `/${raw.replace(/^\/+|\/+$/g, '')}`
}

const localizedPath = (locale: string, path: string) => {
  const normalizedPath = normalizePath(path)
  const normalizedLocale = normalizeLocale(locale)
  if (!normalizedPath || !normalizedLocale) return ''
  if (normalizedLocale === defaultLocale) return normalizedPath
  return `/${normalizedLocale}${normalizedPath}`
}

const pathMatchesLocale = (path: string, locale: string) => {
  const normalizedPath = normalizePath(path)
  const normalizedLocale = normalizeLocale(locale)
  if (!normalizedPath || !normalizedLocale) return false

  const firstSegment = normalizedPath.split('/').filter(Boolean)[0] || ''
  if (normalizedLocale === defaultLocale) {
    return !localeCodeSet.has(firstSegment)
  }
  return firstSegment === normalizedLocale
}

const stripLocalePrefix = (path: string) => {
  const normalizedPath = normalizePath(path)
  if (!normalizedPath || normalizedPath === '/') return normalizedPath

  const segments = normalizedPath.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''
  if (firstSegment && localeCodeSet.has(firstSegment)) {
    const remainder = segments.slice(1)
    return remainder.length ? `/${remainder.join('/')}` : '/'
  }
  return normalizedPath
}

const isProductPath = (path: string) => {
  const segments = stripLocalePrefix(path).split('/').filter(Boolean)
  return segments.length === 2 && segments[0] === 'products'
}

const isShopCategoryPath = (path: string) => {
  const barePath = stripLocalePrefix(path)
  return barePath.startsWith('/shop/') && barePath !== '/shop/'
}

const getLastmod = (...values: unknown[]) => {
  for (const value of values) {
    const date = String(value || '').trim()
    if (date) return date
  }
  return undefined
}

const addEntry = (entries: SitemapEntry[], seen: Set<string>, entry: SitemapEntry) => {
  const loc = normalizePath(entry.loc)
  if (!loc || seen.has(loc)) return
  seen.add(loc)
  entries.push({
    ...entry,
    loc,
  })
}

const collectByLocale = async <T>(
  label: string,
  loader: (locale: string) => Promise<T[]>,
) => {
  const batches = await Promise.all(localeCodes.map(async (locale) => {
    try {
      return await loader(locale)
    } catch (error) {
      throw new Error(
        `[sitemap] failed to load ${label} for ${locale}: ${error instanceof Error ? error.message : String(error)}`,
        { cause: error },
      )
    }
  }))
  return batches.flat()
}

const fetchCatalogSitemapRoutes = async (apiOrigin: string) => {
  try {
    const response = await $fetch<{ items?: SitemapEntry[] }>(`${apiOrigin}/api/v1/storefront/sitemap-routes`, {
      params: {
        limit: 50000,
      },
    })
    if (!Array.isArray(response?.items)) return []

    const routes: SitemapEntry[] = []
    for (const route of response.items) {
      const loc = normalizePath(route?.loc)
      if (!loc) continue
      routes.push({
        ...route,
        loc,
        canonical_path: normalizePath(route.canonical_path) || loc,
      })
    }
    return routes
  } catch (error) {
    throw new Error(
      `[sitemap] failed to load storefront route catalog sitemap routes: ${error instanceof Error ? error.message : String(error)}`,
      { cause: error },
    )
  }
}

const fetchAllProducts = async (apiOrigin: string, locale: string) => {
  const products: ProductSitemapRecord[] = []

  for (let page = 1; page <= maxPages; page += 1) {
    const response = await $fetch<ProductListResponse>(`${apiOrigin}/api/v1/products`, {
      params: {
        locale,
        status: 'active',
        page,
        page_size: pageSize,
      },
      headers: {
        'Accept-Language': locale,
      },
    })
    const items = Array.isArray(response?.data) ? response.data : []
    products.push(...items.map((item) => ({
      ...item,
      locale: normalizeLocale(item.locale, locale),
    })))

    if (items.length === 0 || response?.has_more !== true || items.length < pageSize) break
  }

  return products
}

const flattenCategories = (tree: ProductCategorySitemapRecord[]) => {
  const categories: ProductCategorySitemapRecord[] = []
  const visit = (items: ProductCategorySitemapRecord[]) => {
    for (const item of items) {
      categories.push(item)
      if (Array.isArray(item.children) && item.children.length > 0) {
        visit(item.children)
      }
    }
  }
  visit(tree)
  return categories
}

const fetchAllCategories = async (apiOrigin: string, locale: string) => {
  const response = await $fetch<ProductCategoryListResponse>(`${apiOrigin}/api/v1/products/categories`, {
    headers: {
      'Accept-Language': locale,
    },
  })
  const data = response?.data
  const items = Array.isArray(data?.flat) && data.flat.length > 0
    ? data.flat
    : Array.isArray(data?.tree)
      ? flattenCategories(data.tree)
      : []
  return items.map((item) => ({
    ...item,
    locale: normalizeLocale(item.locale, locale),
  }))
}

const fetchAllFAQPages = async (apiOrigin: string, locale: string) => {
  const response = await $fetch<FAQPageListResponse>(`${apiOrigin}/api/v1/content/faq-pages`, {
    params: { locale },
    headers: {
      'Accept-Language': locale,
    },
  })
  return (Array.isArray(response?.pages) ? response.pages : []).map((item) => ({
    ...item,
    locale: normalizeLocale(item.locale, locale),
  }))
}

const fetchAllPosts = async (apiOrigin: string, locale: string) => {
  const posts: PostSitemapRecord[] = []

  for (let page = 1; page <= maxPages; page += 1) {
    const response = await $fetch<PostListResponse>(`${apiOrigin}/api/v1/content/posts`, {
      params: {
        locale,
        status: 'published',
        page,
        page_size: pageSize,
      },
      headers: {
        'Accept-Language': locale,
      },
    })
    const items = Array.isArray(response?.data) ? response.data : []
    posts.push(...items.map((item) => ({
      ...item,
      locale: normalizeLocale(item.locale, locale),
    })))

    const totalPages = Number(response?.total_pages || 0)
    if (items.length === 0 || (totalPages > 0 && page >= totalPages) || items.length < pageSize) break
  }

  return posts
}

const buildPostPath = (post: PostSitemapRecord) => {
  const slug = String(post.slug || '').trim()
  const locale = normalizeLocale(post.locale)
  if (!slug || !locale) return ''

  const tags = String(post.tags || '')
    .split(',')
    .map((tag) => tag.trim().toLowerCase())
  const category = tags.includes('news')
    ? '/news'
    : tags.includes('wheelsbuild')
      ? '/wheelsbuild'
      : ''

  return localizedPath(
    locale,
    `/resources/blog${category}/${encodeURIComponent(slug)}`,
  )
}

const catalogRouteAllowed = (
  route: SitemapEntry,
  categoryPaths: Set<string>,
  productPaths: Set<string>,
) => {
  const loc = normalizePath(route.loc)
  const canonicalPath = normalizePath(route.canonical_path) || loc
  if (!loc || canonicalPath !== loc) return false

  if (route.source_type === 'product' || isProductPath(loc)) {
    return productPaths.has(loc)
  }
  if (isShopCategoryPath(loc)) {
    return categoryPaths.has(loc)
  }
  return true
}

export default defineSitemapEventHandler(async () => {
  const config = useRuntimeConfig()
  const apiOrigin = String(config.apiInternalOrigin || '').replace(/\/$/, '')
  if (!apiOrigin) {
    throw new Error('[sitemap] API internal origin is not configured.')
  }

  const entries: SitemapEntry[] = []
  const seen = new Set<string>()

  const [catalogRoutes, categories, products, faqPages, posts] = await Promise.all([
    fetchCatalogSitemapRoutes(apiOrigin),
    collectByLocale('product categories', (locale) => fetchAllCategories(apiOrigin, locale)),
    collectByLocale('products', (locale) => fetchAllProducts(apiOrigin, locale)),
    collectByLocale('FAQ pages', (locale) => fetchAllFAQPages(apiOrigin, locale)),
    collectByLocale('blog posts', (locale) => fetchAllPosts(apiOrigin, locale)),
  ])

  const categoryPaths = new Set<string>()
  for (const category of categories) {
    const routePath = normalizePath(category.route_path)
    const locale = normalizeLocale(category.locale)
    if (!routePath || !locale || !pathMatchesLocale(routePath, locale)) continue
    categoryPaths.add(routePath)
    addEntry(entries, seen, {
      loc: routePath,
      lastmod: getLastmod(category.updated_at),
      changefreq: 'monthly',
      priority: 0.7,
      locale,
      source_type: 'category',
      canonical_path: routePath,
    })
  }

  const productPaths = new Set<string>()
  for (const product of products) {
    const slug = String(product.slug || '').trim()
    const locale = normalizeLocale(product.locale)
    const routePath = slug && locale
      ? localizedPath(locale, `/products/${encodeURIComponent(slug)}`)
      : ''
    if (!routePath) continue
    productPaths.add(routePath)
    addEntry(entries, seen, {
      loc: routePath,
      lastmod: getLastmod(product.updated_at),
      changefreq: 'weekly',
      priority: 0.8,
      locale,
      source_type: 'product',
      canonical_path: routePath,
    })
  }

  for (const page of faqPages) {
    const routePath = normalizePath(page.route_path)
    const locale = normalizeLocale(page.locale)
    if (!routePath || !locale || !pathMatchesLocale(routePath, locale)) continue

    addEntry(entries, seen, {
      loc: routePath,
      lastmod: getLastmod(page.updated_at),
      changefreq: 'monthly',
      priority: 0.6,
      locale,
    })
  }

  for (const post of posts) {
    const routePath = buildPostPath(post)
    if (!routePath || !pathMatchesLocale(routePath, normalizeLocale(post.locale))) continue

    addEntry(entries, seen, {
      loc: routePath,
      lastmod: getLastmod(post.updated_at, post.published_at),
      changefreq: 'monthly',
      priority: 0.6,
      locale: normalizeLocale(post.locale),
    })
  }

  for (const route of catalogRoutes) {
    if (!catalogRouteAllowed(route, categoryPaths, productPaths)) continue
    addEntry(entries, seen, route)
  }

  return entries
})
