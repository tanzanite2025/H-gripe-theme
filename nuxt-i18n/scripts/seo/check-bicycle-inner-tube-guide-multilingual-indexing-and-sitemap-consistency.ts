import assert from 'node:assert/strict'
import { parse, type DefaultTreeAdapterMap } from 'parse5'
import storefrontLocales from '../../app/i18n/locales.manifest.ts'

type InnerTubeIndexingHtmlNode = DefaultTreeAdapterMap['node']
type InnerTubeIndexingHtmlElement = DefaultTreeAdapterMap['element']

const storefrontBaseUrl = process.env.SEO_STOREFRONT_BASE_URL
if (!storefrontBaseUrl) throw new Error('Set SEO_STOREFRONT_BASE_URL to a running storefront origin.')

const guidePath = '/tools/valve-length-fitment-calculator'
const expectedAlternatePaths = new Map([
  ['en-US', guidePath],
  ['zh-CN', `/zh_cn${guidePath}`],
  ['x-default', guidePath],
])
let canonicalSiteOrigin = ''

function collectInnerTubeIndexingHtmlElements(node: InnerTubeIndexingHtmlNode): InnerTubeIndexingHtmlElement[] {
  const elements: InnerTubeIndexingHtmlElement[] = 'tagName' in node ? [node] : []
  if ('childNodes' in node) {
    for (const child of node.childNodes) elements.push(...collectInnerTubeIndexingHtmlElements(child))
  }
  return elements
}

function readInnerTubeIndexingHtmlAttribute(element: InnerTubeIndexingHtmlElement, name: string): string {
  return element.attrs.find(attribute => attribute.name === name)?.value || ''
}

function readInnerTubeIndexingHtmlText(node: InnerTubeIndexingHtmlNode): string {
  if (node.nodeName === '#text' && 'value' in node) return node.value
  return 'childNodes' in node ? node.childNodes.map(readInnerTubeIndexingHtmlText).join('') : ''
}

async function fetchInnerTubeIndexingVerificationResponse(routePath: string): Promise<Response> {
  const response = await fetch(new URL(routePath, storefrontBaseUrl), { signal: AbortSignal.timeout(60_000) })
  assert.equal(response.status, 200, `${routePath} must return HTTP 200`)
  return response
}

function assertInnerTubeIndexingUrlMatchesPublishedPath(href: string, expectedPath: string): void {
  const url = new URL(href)
  assert.equal(url.origin, canonicalSiteOrigin)
  assert.equal(url.pathname, expectedPath)
  assert.equal(url.search, '', 'Tracking and calculator parameters must not enter indexing URLs')
  assert.equal(url.hash, '')
}

function assertInnerTubeGuidePublishedAlternateLinks(elements: InnerTubeIndexingHtmlElement[]): void {
  const alternates = elements.filter(element => /^(?:xhtml:)?link$/.test(element.tagName)
    && readInnerTubeIndexingHtmlAttribute(element, 'rel') === 'alternate')
  assert.equal(alternates.length, 3, 'Only published English, Chinese and x-default may be declared')
  assert.deepEqual(alternates.map(element => readInnerTubeIndexingHtmlAttribute(element, 'hreflang')).sort(), [...expectedAlternatePaths.keys()].sort())
  for (const alternate of alternates) {
    const language = readInnerTubeIndexingHtmlAttribute(alternate, 'hreflang')
    assertInnerTubeIndexingUrlMatchesPublishedPath(readInnerTubeIndexingHtmlAttribute(alternate, 'href'), expectedAlternatePaths.get(language)!)
  }
}

function collectInnerTubeGuideFitmentJsonLdGraph(elements: InnerTubeIndexingHtmlElement[]): Record<string, unknown>[] {
  return elements.filter(element => element.tagName === 'script'
    && readInnerTubeIndexingHtmlAttribute(element, 'type') === 'application/ld+json')
    .flatMap(element => {
      const schema = JSON.parse(readInnerTubeIndexingHtmlText(element))
      return Array.isArray(schema['@graph']) ? schema['@graph'] : [schema]
    })
    .filter(entity => String(entity['@id'] || '').endsWith('#inner-tube-guide')
      || String(entity['@id'] || '').includes('#inner-tube-fitment-'))
}

for (const locale of storefrontLocales) {
  const localizedPath = `${locale.code === 'en' ? '' : `/${locale.code}`}${guidePath}`
  const hasPublishedTranslation = locale.code === 'en' || locale.code === 'zh_cn'
  const actualContentLanguage = locale.code === 'zh_cn' ? 'zh-CN' : 'en-US'
  for (const suffix of ['', '?rim_depth=80&fitment_mode=manual&valve_length=80&extender_length=30&pump_head_grip_depth=22&rim_depth_uncertainty=4&utm_source=forum']) {
    const response = await fetchInnerTubeIndexingVerificationResponse(localizedPath + suffix)
    const elements = collectInnerTubeIndexingHtmlElements(parse(await response.text()))
    const canonicalElements = elements.filter(element => element.tagName === 'link'
      && readInnerTubeIndexingHtmlAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalElements.length, 1)
    const canonicalUrl = readInnerTubeIndexingHtmlAttribute(canonicalElements[0]!, 'href')
    canonicalSiteOrigin ||= new URL(canonicalUrl).origin
    assertInnerTubeIndexingUrlMatchesPublishedPath(canonicalUrl, localizedPath)
    assertInnerTubeGuidePublishedAlternateLinks(elements)

    const robots = elements.filter(element => element.tagName === 'meta'
      && readInnerTubeIndexingHtmlAttribute(element, 'name') === 'robots')
    assert.ok(robots.length <= 1, 'No duplicate robots tags')
    const robotsContent = robots.map(element => readInnerTubeIndexingHtmlAttribute(element, 'content')).join(',').toLowerCase()
    const headerRobots = (response.headers.get('x-robots-tag') || '').toLowerCase()
    if (hasPublishedTranslation) {
      assert.doesNotMatch(robotsContent + headerRobots, /noindex|none/)
    } else {
      assert.equal(robotsContent, 'noindex,follow')
      assert.doesNotMatch(headerRobots, /nofollow/)
    }

    const guideElements = elements.filter(element => readInnerTubeIndexingHtmlAttribute(element, 'class').split(/\s+/).includes('inner-tube-guide'))
    assert.equal(guideElements.length, 1)
    assert.equal(readInnerTubeIndexingHtmlAttribute(guideElements[0]!, 'lang'), actualContentLanguage)
    const fitmentGraph = collectInnerTubeGuideFitmentJsonLdGraph(elements)
    assert.equal(fitmentGraph.length, 4, 'The whole-guide Article and three fitment entities must use the actual content language')
    for (const entity of fitmentGraph) assert.equal(entity.inLanguage, actualContentLanguage)
  }
  console.log(`Inner-tube guide indexing passed: ${locale.code}, ${hasPublishedTranslation ? 'indexable' : 'noindex,follow'}, plain and shared URLs`)
}

const dynamicEntries: { loc: string }[] = await (await fetchInnerTubeIndexingVerificationResponse('/__sitemap__/dynamic-urls.json')).json()
const dynamicGuidePaths = dynamicEntries.filter(entry => new URL(entry.loc, storefrontBaseUrl).pathname.endsWith(guidePath))
  .map(entry => new URL(entry.loc, storefrontBaseUrl).pathname)
assert.deepEqual(dynamicGuidePaths.sort(), [guidePath, `/zh_cn${guidePath}`].sort(), 'Old catalog or FAQ records must not reintroduce fallback locales')

const sitemapIndexElements = collectInnerTubeIndexingHtmlElements(parse(await (await fetchInnerTubeIndexingVerificationResponse('/sitemap.xml')).text()))
const shardUrls = sitemapIndexElements.filter(element => element.tagName === 'loc').map(readInnerTubeIndexingHtmlText)
assert.equal(new Set(shardUrls).size, shardUrls.length)
assert.equal(shardUrls.length, storefrontLocales.length, 'Verify every configured locale shard')
const publishedGuideSitemapUrls: string[] = []
for (const shardUrl of shardUrls) {
  const shardXml = await (await fetchInnerTubeIndexingVerificationResponse(new URL(shardUrl).pathname)).text()
  // Isolate each XML <url> before using the existing HTML parser: self-closing
  // xhtml:link elements are not HTML void elements and otherwise nest entries.
  for (const urlBlock of shardXml.match(/<url\b[^>]*>[\s\S]*?<\/url>/g) || []) {
    const elements = collectInnerTubeIndexingHtmlElements(parse(urlBlock))
    const locElements = elements.filter(element => element.tagName === 'loc')
    assert.equal(locElements.length, 1)
    const loc = readInnerTubeIndexingHtmlText(locElements[0]!).trim()
    if (!new URL(loc).pathname.endsWith(guidePath)) continue
    const path = new URL(loc).pathname
    assert.ok(path === guidePath || path === `/zh_cn${guidePath}`, `Fallback locale must not enter the sitemap: ${loc}`)
    const expectedShardLanguage = path === guidePath ? 'en-US' : 'zh-CN'
    assert.equal(new URL(shardUrl).pathname, `/__sitemap__/${expectedShardLanguage}.xml`, 'Published URLs must belong to their own locale shard')
    assertInnerTubeIndexingUrlMatchesPublishedPath(loc, path)
    assertInnerTubeGuidePublishedAlternateLinks(elements)
    publishedGuideSitemapUrls.push(path)
  }
}
assert.deepEqual(publishedGuideSitemapUrls.sort(), [guidePath, `/zh_cn${guidePath}`].sort())
console.log(`Inner-tube guide sitemap passed: dynamic source, ${shardUrls.length} XML shards, exactly two URLs, reciprocal published alternatives`)

const robotsText = await (await fetchInnerTubeIndexingVerificationResponse('/robots.txt')).text()
const robotsGroups = robotsText.split(/\r?\n\s*\r?\n/)
for (const crawler of ['AhrefsBot', 'SemrushBot', 'SimilarwebBot', 'BuiltWith']) {
  const crawlerGroup = robotsGroups.find(group => group.includes(`User-agent: ${crawler}`))
  assert.ok(crawlerGroup, `Preserve the existing ${crawler} exclusion`)
  assert.match(crawlerGroup, /^disallow:\s*\/\s*$/im)
}
const googleGroups = robotsGroups.filter(group => /^user-agent:\s*(?:\*|googlebot)\s*$/im.test(group))
assert.ok(googleGroups.length > 0)
for (const group of googleGroups) {
  assert.doesNotMatch(group, /^disallow:\s*\S+/im, 'Google must be able to crawl fallback pages and read noindex')
}
assert.match(robotsText, new RegExp(`^Sitemap: ${canonicalSiteOrigin.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}/sitemap\\.xml\\s*$`, 'im'))
console.log('Inner-tube guide robots.txt passed: Google crawling allowed, canonical sitemap URL advertised')

for (const path of ['/guides/tireguides/tire-frame-clearance', '/fr/guides/tireguides/tire-frame-clearance']) {
  const elements = collectInnerTubeIndexingHtmlElements(parse(await (await fetchInnerTubeIndexingVerificationResponse(path)).text()))
  const alternates = elements.filter(element => element.tagName === 'link' && readInnerTubeIndexingHtmlAttribute(element, 'rel') === 'alternate')
  assert.equal(alternates.length, storefrontLocales.length + 1, 'The two-locale policy must not leak into adjacent guides')
  assert.equal(elements.some(element => element.tagName === 'meta'
    && readInnerTubeIndexingHtmlAttribute(element, 'name') === 'robots'
    && /noindex/.test(readInnerTubeIndexingHtmlAttribute(element, 'content'))), false)
}
console.log('Inner-tube guide indexing isolation passed: adjacent guides retain their own policy')
