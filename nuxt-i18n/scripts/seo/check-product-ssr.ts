import locales from '../../app/i18n/locales.manifest'

const targetUrl = process.env.SEO_PRODUCT_URL

if (!targetUrl) {
  throw new Error('Set SEO_PRODUCT_URL to a rendered product URL before running this check.')
}

const expectJsonLd = process.env.SEO_EXPECT_JSON_LD !== 'false'

const response = await fetch(targetUrl)
if (!response.ok) {
  throw new Error(`Product URL returned HTTP ${response.status}: ${targetUrl}`)
}

const html = await response.text()

const decodeHtml = (value: string) => value
  .replace(/&amp;/g, '&')
  .replace(/&quot;/g, '"')
  .replace(/&#39;/g, "'")
  .replace(/&lt;/g, '<')
  .replace(/&gt;/g, '>')

const stripTags = (value: string) => decodeHtml(value.replace(/<[^>]*>/g, '').replace(/\s+/g, ' ').trim())

const readTagAttribute = (tag: string, attribute: string) => {
  const match = tag.match(new RegExp(`\\b${attribute}\\s*=\\s*["']([^"']*)["']`, 'i'))
  return match?.[1] ? decodeHtml(match[1]) : ''
}

const titleMatch = html.match(/<title[^>]*>([\s\S]*?)<\/title>/i)
const titleText = titleMatch ? stripTags(titleMatch[1]) : ''
if (!titleText) {
  throw new Error('Initial HTML must include a non-empty title. ')
}

const h1Matches = [...html.matchAll(/<h1\b[^>]*>([\s\S]*?)<\/h1>/gi)]
if (h1Matches.length !== 1) {
  throw new Error(`Expected exactly one H1 in initial HTML, found ${h1Matches.length}.`)
}
const h1Text = stripTags(h1Matches[0][1])

const linkTags = [...html.matchAll(/<link\b[^>]*>/gi)].map((match) => match[0])
const canonicalMatches = linkTags.filter((tag) => readTagAttribute(tag, 'rel').toLowerCase() === 'canonical')
if (canonicalMatches.length !== 1) {
  throw new Error(`Expected exactly one canonical link, found ${canonicalMatches.length}.`)
}
const canonicalUrl = readTagAttribute(canonicalMatches[0], 'href')
if (!canonicalUrl) {
  throw new Error('Canonical link must contain an href.')
}

const canonicalUrlObject = new URL(canonicalUrl)
if (canonicalUrlObject.search || canonicalUrlObject.hash) {
  throw new Error('Canonical link must not contain a query string or hash.')
}

const localeCodes = new Set(
  locales.map((locale) => String(locale.code || '').trim().toLowerCase()).filter(Boolean),
)

const stripLocalePrefix = (value: string) => {
  const parsed = new URL(value)
  const segments = parsed.pathname.split('/').filter(Boolean)
  if (segments[0] && localeCodes.has(segments[0].toLowerCase())) {
    segments.shift()
  }
  return segments.length ? `/${segments.join('/')}` : '/'
}

const canonicalPath = stripLocalePrefix(canonicalUrl)
if (!/^\/products\/[^/]+$/.test(canonicalPath)) {
  throw new Error(`Product canonical must use the flat /products/:slug route: ${canonicalUrl}`)
}

const alternateLinks = linkTags
  .filter((tag) => readTagAttribute(tag, 'rel').toLowerCase() === 'alternate')
  .map((tag) => ({
    hreflang: readTagAttribute(tag, 'hreflang').toLowerCase(),
    href: readTagAttribute(tag, 'href'),
  }))
  .filter((link) => link.hreflang && link.href)
if (!alternateLinks.some((link) => link.hreflang === 'x-default')) {
  throw new Error('Initial HTML must include an x-default hreflang link.')
}
if (!alternateLinks.some((link) => link.hreflang !== 'x-default')) {
  throw new Error('Initial HTML must include at least one locale hreflang link.')
}

const jsonLdMatches = [...html.matchAll(
  /<script[^>]+type="application\/ld\+json"[^>]*>([\s\S]*?)<\/script>/gi,
)]

if (!jsonLdMatches.length) {
  if (!expectJsonLd) {
    console.log(`Product SSR SEO negative checks passed: ${targetUrl}`)
    process.exit(0)
  }
  throw new Error('No JSON-LD script body was found in the initial HTML.')
}

if (!expectJsonLd) {
  throw new Error('Expected no JSON-LD script in the initial HTML, but one was found.')
}

const schemas = jsonLdMatches.map((match) => {
  try {
    return JSON.parse(match[1])
  } catch (error) {
    throw new Error(`Invalid JSON-LD in initial HTML: ${error instanceof Error ? error.message : String(error)}`)
  }
})

const productSchema = schemas.find((schema) => (
  schema?.['@type'] === 'Product' || schema?.['@type'] === 'ProductGroup'
))
if (!productSchema) {
  throw new Error('No Product or ProductGroup JSON-LD object was found in the initial HTML.')
}

if (!productSchema.name || !productSchema.url || productSchema.url !== canonicalUrl) {
  throw new Error('Product structured data must include a URL matching canonical.')
}
if (h1Text !== String(productSchema.name).trim()) {
  throw new Error(`Product structured data name must match H1: ${productSchema.name} !== ${h1Text}`)
}
if (!titleText.toLowerCase().includes(String(productSchema.name).trim().toLowerCase())) {
  throw new Error(`Page title must contain the visible product name: ${titleText}`)
}

const breadcrumbSchema = schemas.find((schema) => schema?.['@type'] === 'BreadcrumbList')
if (!breadcrumbSchema) {
  throw new Error('Product SSR HTML must include a BreadcrumbList JSON-LD object.')
}

const breadcrumbItems = Array.isArray(breadcrumbSchema.itemListElement)
  ? breadcrumbSchema.itemListElement
  : []
if (breadcrumbItems.length < 4) {
  throw new Error('BreadcrumbList must include Home, Shop, at least one category, and Product.')
}

const breadcrumbPaths = breadcrumbItems.map((item: any, index: number) => {
  if (!item || typeof item !== 'object') {
    throw new Error(`Breadcrumb item ${index + 1} is not an object.`)
  }
  if (Number(item.position) !== index + 1) {
    throw new Error(`Breadcrumb item ${index + 1} has an invalid position.`)
  }
  if (!String(item.name || '').trim()) {
    throw new Error(`Breadcrumb item ${index + 1} is missing a name.`)
  }
  const itemUrl = String(item.item || '').trim()
  if (!itemUrl) {
    throw new Error(`Breadcrumb item ${index + 1} is missing an item URL.`)
  }
  const parsed = new URL(itemUrl)
  if (parsed.origin !== canonicalUrlObject.origin || parsed.search || parsed.hash) {
    throw new Error(`Breadcrumb item ${index + 1} must be a clean same-site URL.`)
  }
  return stripLocalePrefix(itemUrl)
})

if (breadcrumbPaths[0] !== '/') {
  throw new Error(`Breadcrumb must start at Home: ${breadcrumbPaths[0]}`)
}
if (breadcrumbPaths[1] !== '/shop') {
  throw new Error(`Breadcrumb second item must be /shop: ${breadcrumbPaths[1]}`)
}
const breadcrumbCategoryPaths = breadcrumbPaths.slice(2, -1)
if (!breadcrumbCategoryPaths.length || breadcrumbCategoryPaths.some((path) => !path.startsWith('/shop/'))) {
  throw new Error('Breadcrumb must contain at least one real /shop/... category path.')
}
if (breadcrumbPaths[breadcrumbPaths.length - 1] !== canonicalPath) {
  throw new Error('Breadcrumb product item must match the flat product canonical path.')
}

const validateImages = (images: unknown, label: string) => {
  if (!Array.isArray(images) || !images.length) {
    throw new Error(`${label} must include at least one image.`)
  }
  for (const image of images) {
    if (!/^https?:\/\//i.test(String(image || ''))) {
      throw new Error(`${label} image must be an absolute public URL: ${String(image || '')}`)
    }
  }
}

validateImages(productSchema.image, productSchema['@type'])

const validateOffer = (offer: any, label: string) => {
  if (!offer) {
    throw new Error(`${label} must include an Offer.`)
  }
  if (!Number.isFinite(Number(offer.price)) || Number(offer.price) <= 0) {
    throw new Error(`${label} Offer price is not a positive number.`)
  }
  if (!/^[A-Z]{3}$/.test(String(offer.priceCurrency || ''))) {
    throw new Error(`${label} Offer currency is not a three-letter code.`)
  }
  if (!/^https:\/\/schema\.org\/(InStock|OutOfStock)$/.test(String(offer.availability || ''))) {
    throw new Error(`${label} Offer availability is invalid.`)
  }
  if (!offer.url || offer.url !== canonicalUrl && productSchema['@type'] === 'Product') {
    throw new Error(`${label} Offer URL must match canonical for a single product.`)
  }
}

if (productSchema['@type'] === 'ProductGroup') {
  if (!productSchema.productGroupID || !Array.isArray(productSchema.hasVariant) || productSchema.hasVariant.length < 2) {
    throw new Error('ProductGroup must include a stable productGroupID and at least two variants.')
  }
  const variantUrls = new Set<string>()
  for (const [index, variant] of productSchema.hasVariant.entries()) {
    if (variant?.['@type'] !== 'Product' || !variant.name || !variant.url) {
      throw new Error(`ProductGroup variant ${index + 1} is incomplete.`)
    }
    validateImages(variant.image, `ProductGroup variant ${index + 1}`)
    if (variantUrls.has(variant.url)) {
      throw new Error(`ProductGroup variant URL is duplicated: ${variant.url}`)
    }
    variantUrls.add(variant.url)
    validateOffer(variant.offers, `ProductGroup variant ${index + 1}`)
  }
} else {
  validateOffer(productSchema.offers, 'Product')
}

console.log(`Product SSR SEO checks passed: ${targetUrl}`)
