import { parse } from 'parse5'
import locales from '../../app/i18n/locales.manifest.ts'

type SitemapNode = {
  tagName?: string
  value?: string
  childNodes?: SitemapNode[]
}

type DynamicSitemapEntry = {
  loc?: string
  source_type?: string
  canonical_path?: string
}

const rawBaseUrl = String(process.env.SEO_SITEMAP_BASE_URL || '').trim()
if (!rawBaseUrl) {
  throw new Error('Set SEO_SITEMAP_BASE_URL to the running storefront origin before checking the sitemap.')
}

const baseUrl = new URL(rawBaseUrl)
const baseOrigin = baseUrl.origin
const localeCodes = locales
  .map((locale) => String(locale.code || '').trim().toLowerCase())
  .filter(Boolean)
const localeCodeSet = new Set(localeCodes)
const defaultLocale = 'en'

const localeKey = (value: unknown) => String(value || '').trim().toLowerCase().replace(/-/g, '_')

const sitemapLocaleByKey = new Map<string, string>()
for (const locale of locales) {
  const code = String(locale.code || '').trim().toLowerCase()
  if (!code) continue
  sitemapLocaleByKey.set(localeKey(code), code)
  sitemapLocaleByKey.set(localeKey(locale.iso), code)
  sitemapLocaleByKey.set(localeKey(locale.language), code)
}

const fail = (message: string): never => {
  throw new Error(`[seo-sitemap] ${message}`)
}

const fetchText = async (url: string, label: string) => {
  const response = await fetch(url)
  if (!response.ok) {
    fail(`${label} returned HTTP ${response.status}: ${url}`)
  }
  return response.text()
}

const findNodes = (node: SitemapNode, tagName: string): SitemapNode[] => {
  const matches: SitemapNode[] = []
  if (String(node.tagName || '').toLowerCase() === tagName.toLowerCase()) {
    matches.push(node)
  }
  for (const child of node.childNodes || []) {
    matches.push(...findNodes(child, tagName))
  }
  return matches
}

const textContent = (node: SitemapNode): string => {
  const ownText = String(node.value || '')
  return ownText + (node.childNodes || []).map(textContent).join('')
}

const parseContainerLocs = (xml: string, containerTag: 'sitemap' | 'url', label: string) => {
  const document = parse(xml) as unknown as SitemapNode
  const locs = findNodes(document, containerTag)
    .map((container) => findNodes(container, 'loc')[0])
    .map((loc) => (loc ? textContent(loc).trim() : ''))
    .filter(Boolean)

  if (!locs.length) {
    fail(`${label} contains no <${containerTag}><loc> entries.`)
  }
  return locs
}

const normalizePath = (value: unknown, label: string) => {
  const raw = String(value || '').trim()
  if (!raw || /[?#]/.test(raw)) {
    fail(`${label} must be a path without query/hash: ${raw || '<empty>'}`)
  }

  const parsed = (() => {
    try {
      return new URL(raw, baseUrl)
    } catch {
      return fail(`${label} is not a valid URL/path: ${raw}`)
    }
  })()

  if (parsed.origin !== baseOrigin) {
    fail(`${label} points to another origin: ${parsed.origin}`)
  }

  const normalized = `/${parsed.pathname.replace(/^\/+|\/+$/g, '')}`
  return normalized === '//' ? '/' : normalized
}

const stripLocalePrefix = (path: string) => {
  const segments = path.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''
  if (firstSegment && localeCodeSet.has(firstSegment)) {
    const remainder = segments.slice(1)
    return remainder.length ? `/${remainder.join('/')}` : '/'
  }
  return path
}

const localeForPath = (path: string) => {
  const firstSegment = path.split('/').filter(Boolean)[0] || ''
  return firstSegment && localeCodeSet.has(firstSegment) ? firstSegment : defaultLocale
}

const isProductPath = (path: string) => {
  const segments = stripLocalePrefix(path).split('/').filter(Boolean)
  return segments.length === 2 && segments[0] === 'products' && Boolean(segments[1])
}

const isShopCategoryPath = (path: string) => {
  const barePath = stripLocalePrefix(path)
  return barePath.startsWith('/shop/') && barePath !== '/shop/'
}

const legacyShopPathForProduct = (path: string) => {
  const localePrefix = localeForPath(path) === defaultLocale
    ? ''
    : `/${localeForPath(path)}`
  const slug = stripLocalePrefix(path).split('/').filter(Boolean)[1]
  return `${localePrefix}/shop/${slug}`
}

const indexUrl = new URL('/sitemap.xml', baseUrl).toString()
const indexXml = await fetchText(indexUrl, 'sitemap index')
const sitemapLocs = parseContainerLocs(indexXml, 'sitemap', 'sitemap index')
const sitemapPaths = new Set(sitemapLocs.map((loc) => normalizePath(loc, 'sitemap index loc')))

if (sitemapPaths.size !== sitemapLocs.length) {
  fail('sitemap index contains duplicate child sitemap URLs.')
}

const shardEntries = new Map<string, {
  sitemapUrl: string
  locale: string
  paths: string[]
  pathSet: Set<string>
}>()
const allXmlPaths: string[] = []

for (const sitemapLoc of sitemapLocs) {
  const sitemapUrl = new URL(sitemapLoc, baseUrl)
  const sitemapName = sitemapUrl.pathname.split('/').filter(Boolean).pop()?.replace(/\.xml$/i, '') || ''
  const shardLocale = sitemapLocaleByKey.get(localeKey(sitemapName))
    || fail(`cannot map sitemap shard "${sitemapName}" to the shared locale registry.`)

  const xml = await fetchText(sitemapUrl.toString(), `sitemap shard ${sitemapName}`)
  const paths = parseContainerLocs(xml, 'url', `sitemap shard ${sitemapName}`)
    .map((loc) => normalizePath(loc, `sitemap shard ${sitemapName} loc`))

  for (const path of paths) {
    const pathLocale = localeForPath(path)
    if (pathLocale !== shardLocale) {
      fail(`sitemap shard ${sitemapName} contains ${path}, which belongs to locale ${pathLocale}.`)
    }
  }

  shardEntries.set(shardLocale, {
    sitemapUrl: sitemapUrl.toString(),
    locale: shardLocale,
    paths,
    pathSet: new Set(paths),
  })
  allXmlPaths.push(...paths)
}

const xmlPathCounts = new Map<string, number>()
for (const path of allXmlPaths) {
  xmlPathCounts.set(path, (xmlPathCounts.get(path) || 0) + 1)
}

const duplicateXmlPaths = [...xmlPathCounts.entries()]
  .filter(([, count]) => count > 1)
  .map(([path]) => path)
if (duplicateXmlPaths.length) {
  fail(`sitemap XML contains duplicate URLs: ${duplicateXmlPaths.slice(0, 5).join(', ')}`)
}

const dynamicUrl = new URL('/__sitemap__/dynamic-urls.json', baseUrl).toString()
const dynamicResponse = await fetch(dynamicUrl)
if (!dynamicResponse.ok) {
  fail(`dynamic sitemap source returned HTTP ${dynamicResponse.status}: ${dynamicUrl}`)
}

const dynamicPayload = await dynamicResponse.json() as unknown
if (!Array.isArray(dynamicPayload)) {
  fail('dynamic sitemap source must return an array.')
}

const dynamicEntries = dynamicPayload as DynamicSitemapEntry[]
const dynamicProductPaths = new Set<string>()
const dynamicCategoryPaths = new Set<string>()
for (const [index, entry] of dynamicEntries.entries()) {
  const path = normalizePath(entry.loc, `dynamic sitemap entry ${index + 1}`)
  const isDeclaredProduct = String(entry.source_type || '').trim() === 'product'
  const isDeclaredCategory = String(entry.source_type || '').trim() === 'category'

  if (isDeclaredCategory || isShopCategoryPath(path)) {
    if (isDeclaredProduct || !isShopCategoryPath(path)) {
      fail(`dynamic category entry ${index + 1} is not a /shop/... path: ${path}`)
    }

    const canonicalPath = normalizePath(
      entry.canonical_path || '',
      `dynamic category entry ${index + 1} canonical_path`,
    )
    if (canonicalPath !== path) {
      fail(`dynamic category entry ${index + 1} canonical_path does not match loc: ${canonicalPath} !== ${path}`)
    }
    dynamicCategoryPaths.add(path)
    continue
  }

  if (!isDeclaredProduct && !isProductPath(path)) continue

  if (!isProductPath(path)) {
    fail(`dynamic product entry ${index + 1} is not a /products/:slug path: ${path}`)
  }

  const canonicalPath = normalizePath(
    entry.canonical_path || '',
    `dynamic product entry ${index + 1} canonical_path`,
  )
  if (canonicalPath !== path) {
    fail(`dynamic product entry ${index + 1} canonical_path does not match loc: ${canonicalPath} !== ${path}`)
  }
  dynamicProductPaths.add(path)
}

const xmlProductPaths = new Set(allXmlPaths.filter(isProductPath))
for (const productPath of dynamicProductPaths) {
  if (!xmlPathCounts.has(productPath)) {
    fail(`product URL from dynamic JSON is missing from all XML sitemap shards: ${productPath}`)
  }

  const shard = shardEntries.get(localeForPath(productPath))
  if (!shard?.pathSet.has(productPath)) {
    fail(`product URL is not present in its locale XML shard: ${productPath}`)
  }

  const legacyShopPath = legacyShopPathForProduct(productPath)
  if (xmlPathCounts.has(legacyShopPath)) {
    fail(`legacy /shop product URL must not be emitted in XML sitemap: ${legacyShopPath}`)
  }
}

for (const productPath of xmlProductPaths) {
  if (!dynamicProductPaths.has(productPath)) {
    fail(`XML sitemap contains a product URL absent from the dynamic product source: ${productPath}`)
  }
}

const xmlCategoryPaths = new Set(allXmlPaths.filter(isShopCategoryPath))
for (const categoryPath of dynamicCategoryPaths) {
  if (!xmlPathCounts.has(categoryPath)) {
    fail(`category URL from dynamic JSON is missing from all XML sitemap shards: ${categoryPath}`)
  }

  const shard = shardEntries.get(localeForPath(categoryPath))
  if (!shard?.pathSet.has(categoryPath)) {
    fail(`category URL is not present in its locale XML shard: ${categoryPath}`)
  }
}

for (const categoryPath of xmlCategoryPaths) {
  if (!dynamicCategoryPaths.has(categoryPath)) {
    fail(`XML sitemap contains a category URL absent from the dynamic category source: ${categoryPath}`)
  }
}

console.log(
  `[seo-sitemap] passed: ${shardEntries.size} locale shards, `
  + `${xmlPathCounts.size} XML URLs, ${dynamicCategoryPaths.size} category URLs, `
  + `${dynamicProductPaths.size} product URLs.`,
)
