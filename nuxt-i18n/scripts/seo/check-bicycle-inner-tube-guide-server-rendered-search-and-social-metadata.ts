import assert from 'node:assert/strict'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import sharp from 'sharp'

type SearchMetadataHtmlNode = DefaultTreeAdapterMap['node']
type SearchMetadataHtmlElement = DefaultTreeAdapterMap['element']

const storefrontBaseUrl = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontBaseUrl) {
  throw new Error('Set SEO_STOREFRONT_BASE_URL to the running storefront origin before checking inner-tube guide metadata.')
}

function collectSearchMetadataHtmlElements(node: SearchMetadataHtmlNode): SearchMetadataHtmlElement[] {
  const elements: SearchMetadataHtmlElement[] = 'tagName' in node ? [node] : []
  if ('childNodes' in node) {
    for (const child of node.childNodes) elements.push(...collectSearchMetadataHtmlElements(child))
  }
  return elements
}

function readSearchMetadataElementAttribute(element: SearchMetadataHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readSearchMetadataElementText(node: SearchMetadataHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readSearchMetadataElementText).join('') : ''
}

function readUniqueSearchMetadataContent(elements: SearchMetadataHtmlElement[], identifier: string): string {
  const matchingElements = elements.filter(element => element.tagName === 'meta' && (
    readSearchMetadataElementAttribute(element, 'name') === identifier
    || readSearchMetadataElementAttribute(element, 'property') === identifier
  ))
  assert.equal(matchingElements.length, 1, `Expected one ${identifier} meta tag`)
  const content = readSearchMetadataElementAttribute(matchingElements[0]!, 'content')
  assert.ok(content.trim(), `${identifier} must not be empty`)
  return content
}

async function fetchServerRenderedGuideMetadata(routePath: string) {
  const response = await fetch(new URL(routePath, storefrontBaseUrl), { signal: AbortSignal.timeout(60000) })
  assert.equal(response.status, 200, `${routePath} must return HTTP 200`)
  const elements = collectSearchMetadataHtmlElements(parse(await response.text()))
  const titleElements = elements.filter(element => element.tagName === 'title')
  assert.equal(titleElements.length, 1, `${routePath} must have one title`)
  const canonicalElements = elements.filter(element => element.tagName === 'link'
    && readSearchMetadataElementAttribute(element, 'rel') === 'canonical')
  assert.equal(canonicalElements.length, 1, `${routePath} must have one canonical`)
  const canonicalUrl = new URL(readSearchMetadataElementAttribute(canonicalElements[0]!, 'href'))
  const requestedUrl = new URL(routePath, storefrontBaseUrl)
  const calculatorCanonicalQuery = requestedUrl.pathname.endsWith('/valve-length-fitment-calculator')
    && requestedUrl.searchParams.get('tab') === 'calculator' ? '?tab=calculator' : ''
  assert.equal(canonicalUrl.pathname, requestedUrl.pathname)
  assert.equal(canonicalUrl.search, calculatorCanonicalQuery, 'Only the explicit calculator tab may enter the canonical URL')
  assert.equal(canonicalUrl.hash, '')
  return { elements, title: readSearchMetadataElementText(titleElements[0]!), canonicalUrl }
}

const innerTubeGuideRouteChecks = [
  { routePath: '/tools/valve-length-fitment-calculator', title: 'Bicycle Inner Tube Sizes & Valve Calculator', locale: 'en_US' },
  { routePath: '/zh_cn/tools/valve-length-fitment-calculator', title: '自行车内胎选择指南：尺寸、气嘴与框高计算器', locale: 'zh_CN' },
  { routePath: '/tools/valve-length-fitment-calculator?rim_depth=60&fitment_mode=automatic', title: 'Bicycle Inner Tube Sizes & Valve Calculator', locale: 'en_US' },
  { routePath: '/zh_cn/tools/valve-length-fitment-calculator?rim_depth=80&fitment_mode=manual&valve_length=80', title: '自行车内胎选择指南：尺寸、气嘴与框高计算器', locale: 'zh_CN' },
  { routePath: '/tools/valve-length-fitment-calculator?tab=calculator', title: 'Bicycle Inner Tube Valve Length & Extender Calculator', locale: 'en_US' },
  { routePath: '/zh_cn/tools/valve-length-fitment-calculator?tab=calculator&rim_depth=80&utm_source=forum', title: '自行车内胎气嘴长度与延长嘴计算器', locale: 'zh_CN' },
  { routePath: '/tools/valve-length-fitment-calculator?tab=introduction&rim_depth=80', title: 'Bicycle Inner Tube Sizes & Valve Calculator', locale: 'en_US' },
]

const checkedSocialPreviewImageUrls = new Set<string>()
for (const guideCheck of innerTubeGuideRouteChecks) {
  const { elements, title, canonicalUrl } = await fetchServerRenderedGuideMetadata(guideCheck.routePath)
  assert.ok(title === guideCheck.title || title.startsWith(`${guideCheck.title} | `), 'Expected localized guide title with an optional configured brand')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:title'), title)
  assert.equal(readUniqueSearchMetadataContent(elements, 'twitter:title'), title)
  const description = readUniqueSearchMetadataContent(elements, 'description')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:description'), description)
  assert.equal(readUniqueSearchMetadataContent(elements, 'twitter:description'), description)
  if (!canonicalUrl.search) {
    for (const valveType of ['Presta', 'Schrader', 'Dunlop']) assert.ok(description.includes(valveType))
  } else {
    assert.match(description, /rim depth|轮圈框高/)
    const alternates = elements.filter(element => element.tagName === 'link'
      && readSearchMetadataElementAttribute(element, 'rel') === 'alternate')
    assert.equal(alternates.length, 3)
    for (const alternate of alternates) {
      assert.equal(new URL(readSearchMetadataElementAttribute(alternate, 'href')).search, '?tab=calculator')
    }
  }
  assert.ok(elements.some(element => element.tagName === 'a'
    && /[?&]tab=calculator(?:&|#|$)/.test(readSearchMetadataElementAttribute(element, 'href'))),
  'The calculator URL must be discoverable through a server-rendered link')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:type'), 'article')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:locale'), guideCheck.locale)
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:url'), canonicalUrl.toString())
  assert.equal(readUniqueSearchMetadataContent(elements, 'twitter:card'), 'summary_large_image')
  const imageUrl = new URL(readUniqueSearchMetadataContent(elements, 'og:image'))
  assert.equal(imageUrl.origin, canonicalUrl.origin)
  assert.ok(imageUrl.pathname.endsWith('/bicycle-inner-tube-size-valve-and-rim-depth-guide-social-preview.png'))
  assert.equal(readUniqueSearchMetadataContent(elements, 'twitter:image'), imageUrl.toString())
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:image:type'), 'image/png')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:image:width'), '1200')
  assert.equal(readUniqueSearchMetadataContent(elements, 'og:image:height'), '630')
  const imageAlt = readUniqueSearchMetadataContent(elements, 'og:image:alt')
  assert.equal(readUniqueSearchMetadataContent(elements, 'twitter:image:alt'), imageAlt)
  if (guideCheck.locale === 'zh_CN') assert.match(description + imageAlt, /[\u4e00-\u9fff]/)
  assert.doesNotMatch(title + description + imageAlt, /guidesTireguides\./, 'Translation keys must not leak into HTML')
  checkedSocialPreviewImageUrls.add(imageUrl.toString())
  console.log(`Inner-tube guide SSR metadata passed: ${guideCheck.routePath}`)
}

for (const imageUrl of checkedSocialPreviewImageUrls) {
  // Fetch through the verification storefront even when its canonical points
  // at a production hostname; do not accidentally check an older live asset.
  const imageResponse: Response = await fetch(new URL(new URL(imageUrl).pathname, storefrontBaseUrl))
  assert.equal(imageResponse.status, 200, 'The configured social preview must be served without authentication')
  assert.match(imageResponse.headers.get('content-type') || '', /^image\/png\b/)
  const imageBytes = Buffer.from(await imageResponse.arrayBuffer())
  assert.ok(imageBytes.length <= 30_000, 'Keep the dedicated share image within its 30 KB transfer budget')
  const imageMetadata = await sharp(imageBytes).metadata()
  assert.equal(imageMetadata.format, 'png')
  assert.equal(imageMetadata.width, 1200)
  assert.equal(imageMetadata.height, 630)
  console.log(`Inner-tube guide social preview passed: PNG, 1200 × 630, HTTP 200, ${imageBytes.length} bytes`)
}

for (const routePath of ['/guides/tireguides', '/guides/tireguides/tire-frame-clearance', '/zh_cn/guides/tireguides/tire-frame-clearance', '/tools/tire-rim-width-compatibility-engine']) {
  const { elements, title } = await fetchServerRenderedGuideMetadata(routePath)
  assert.doesNotMatch(title, /Bicycle Inner Tube Sizes|自行车内胎选择指南/)
  const metadataContents = elements.filter(element => element.tagName === 'meta')
    .map(element => readSearchMetadataElementAttribute(element, 'content')).join(' ')
  assert.doesNotMatch(metadataContents, /bicycle-inner-tube-size-valve-and-rim-depth-guide-social-preview/)
  assert.doesNotMatch(metadataContents, /Compare Presta, Schrader and Dunlop valves|通过框高与气筒夹头参数/)
  console.log(`Other guide SSR metadata isolation passed: ${routePath}`)
}
