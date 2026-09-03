import { parse } from 'parse5'

interface HtmlNode {
  nodeName?: string
  tagName?: string
  value?: string
  attrs?: Array<{ name: string; value: string }>
  childNodes?: HtmlNode[]
}

const targetUrl = process.env.HOME_SSR_URL

if (!targetUrl) {
  throw new Error('Set HOME_SSR_URL to a rendered homepage URL before running this check.')
}

const response = await fetch(targetUrl, {
  headers: {
    accept: 'text/html',
  },
})

if (!response.ok) {
  throw new Error(`Homepage URL returned HTTP ${response.status}: ${targetUrl}`)
}

const html = await response.text()
const document = parse(html) as HtmlNode

const walk = (node: HtmlNode, visit: (current: HtmlNode) => void): void => {
  visit(node)
  for (const child of node.childNodes || []) {
    walk(child, visit)
  }
}

const getAttribute = (node: HtmlNode, name: string): string => {
  return node.attrs?.find((attribute) => attribute.name.toLowerCase() === name.toLowerCase())?.value || ''
}

const textContent = (node: HtmlNode): string => {
  if (node.nodeName === '#text') return node.value || ''
  return (node.childNodes || []).map(textContent).join(' ')
}

const deferredWrappers: HtmlNode[] = []
walk(document, (node) => {
  if (node.tagName !== 'div') return
  const classes = getAttribute(node, 'class').split(/\s+/).filter(Boolean)
  if (classes.includes('home-deferred-section')) {
    deferredWrappers.push(node)
  }
})

if (deferredWrappers.length !== 9) {
  throw new Error(`Expected 9 non-blank deferred wrappers in homepage SSR HTML, found ${deferredWrappers.length}.`)
}

for (const [index, wrapper] of deferredWrappers.entries()) {
  const hasElementChild = (wrapper.childNodes || []).some((child) => Boolean(child.tagName))
  const content = textContent(wrapper).replace(/\s+/g, ' ').trim()

  if (!hasElementChild || !content) {
    throw new Error(`Deferred wrapper ${index + 1} is blank in homepage SSR HTML.`)
  }
}

if (html.includes('ClientOnly')) {
  throw new Error('Homepage SSR HTML must not contain ClientOnly.')
}

const requiredIds = [
  'home-ride-category-strip',
  'home-main-product-categories',
  'home-store-picks-guide',
  'home-buying-path',
  'featured-products',
  'home-shop-with-confidence',
  'home-why-choose-us',
]

for (const id of requiredIds) {
  if (!new RegExp(`<[^>]+\\bid=["']${id}["']`, 'i').test(html)) {
    throw new Error(`Homepage SSR HTML is missing required section #${id}.`)
  }
}

if (!/\bclass=["'][^"']*\bhome-faq\b[^"']*["']/i.test(html) || !html.includes('Frequently Asked Questions')) {
  throw new Error('Homepage SSR HTML is missing the FAQ preview content.')
}

const linkTags = [...html.matchAll(/<link\b[^>]*>/gi)].map((match) => match[0])
const preloadPaths = new Set(
  linkTags
    .filter((tag) => /\brel\s*=\s*["'](?:modulepreload|prefetch)["']/i.test(tag))
    .map((tag) => getAttribute({
      attrs: [...tag.matchAll(/\b([a-z:-]+)\s*=\s*["']([^"']*)["']/gi)]
        .map((match) => ({ name: match[1], value: match[2] })),
    }, 'href'))
    .filter(Boolean)
    .map((href) => new URL(href, targetUrl).pathname)
)

const scriptPaths = [...html.matchAll(/<script\b[^>]*\bsrc=["']([^"']+)["'][^>]*>/gi)]
  .map((match) => new URL(match[1], targetUrl).pathname)
  .filter((pathname) => pathname.endsWith('.js'))

const routeChunkSources = await Promise.all(
  [...new Set([...scriptPaths, ...preloadPaths].filter((pathname) => pathname.endsWith('.js')))]
    .map(async (pathname) => {
      const url = new URL(pathname, targetUrl)
      const chunkResponse = await fetch(url)
      if (!chunkResponse.ok) return null

      const source = await chunkResponse.text()
      if (!source.includes('HomeRideCategoryStrip') || !source.includes('__vite__mapDeps')) {
        return null
      }

      return { url, source }
    }),
)

const routeChunk = routeChunkSources.find(Boolean)
if (!routeChunk) {
  throw new Error('Unable to locate the built homepage route chunk in the initial HTML resources.')
}

const lazyImportPaths = [
  ...routeChunk.source.matchAll(/import\(\s*["'](\.\/[^"']+\.js)["']\s*\)/g),
].map((match) => match[1])

if (lazyImportPaths.length !== 9) {
  throw new Error(`Expected 9 homepage deferred JS imports, found ${lazyImportPaths.length}.`)
}

const deferredAssetNames = [
  'HomeRideCategoryStrip',
  'HomeMainProductCategories',
  'HomeStorePicksGuide',
  'HomePurchasePath',
  'HomeFeaturedProducts',
  'HomeShopWithConfidence',
  'HomeFeaturesTabs',
  'HomeFaqPreview',
  'HomeFinalCta',
]

const deferredStylePaths = [
  ...routeChunk.source.matchAll(/["'](\.\/[^"']+\.css)["']/g),
]
  .map((match) => match[1])
  .filter((path) => deferredAssetNames.some((name) => path.includes(`/${name}.`)))

const deferredAssetPaths = [...new Set([
  ...lazyImportPaths,
  ...deferredStylePaths,
])].map((path) => new URL(path, routeChunk.url).pathname)

const eagerlyReferencedDeferredAssets = deferredAssetPaths.filter((pathname) => preloadPaths.has(pathname))
if (eagerlyReferencedDeferredAssets.length > 0) {
  throw new Error(
    `Homepage deferred assets are still in modulepreload/prefetch: ${eagerlyReferencedDeferredAssets.join(', ')}`,
  )
}

console.log(
  `Homepage SSR checks passed: ${deferredWrappers.length} populated wrappers, `
  + `${deferredAssetPaths.length} deferred assets excluded from modulepreload/prefetch.`,
)
