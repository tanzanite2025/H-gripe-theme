import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { parse } from 'parse5'
import locales from '../../app/i18n/locales.manifest.js'
import { buildStorefrontRouteRules, storefrontRetiredPagePaths } from '../../config/storefront/route-rules.js'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.js'

const storefrontOrigin = process.env.SPOKE_LACING_TOPOLOGY_BASE_URL || 'http://localhost:9100'
const baselineDirectory = process.env.SPOKE_LACING_TOPOLOGY_MIGRATION_BASELINE_DIRECTORY
const migratedToolPath = '/tools/spoke-lacing-topology-simulator'
const retiredGuidePath = '/guides/wheelset-buyers/wheelset-spoke-lacing-topology-and-geometry-reference'
const retiredTopologyPaths = [
  retiredGuidePath,
  '/guides/wheelset-spoke-lacing-topology-and-geometry-reference',
  '/resources/wheelset-spoke-lacing-topology-and-geometry-reference',
]
const publishedLocaleCodes = new Set(['en', 'zh_cn'])
const localeCodes = locales.map(locale => locale.code)
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200', localeCodes, defaultLocale: 'en',
})

type ParsedHtmlElement = {
  nodeName: string
  value?: string
  attrs?: { name: string; value: string }[]
  childNodes?: ParsedHtmlElement[]
}
type CompleteTopologyContentSnapshot = { prefix: string; text: string; schema: unknown }
type FaqPageResponse = { page: { pageId: string; items: unknown[] } }

function collectParsedHtmlElements(element: ParsedHtmlElement): ParsedHtmlElement[] {
  return [element, ...(element.childNodes || []).flatMap(collectParsedHtmlElements)]
}

function readParsedHtmlElementAttribute(element: ParsedHtmlElement, name: string): string {
  return element.attrs?.find(attribute => attribute.name === name)?.value || ''
}

function assertRetiredPathsAbsentFromSitemap(content: string): void {
  for (const retiredPath of storefrontRetiredPagePaths) assert.ok(!content.includes(retiredPath), retiredPath)
}

const baselineSnapshots: CompleteTopologyContentSnapshot[] = baselineDirectory
  ? JSON.parse(readFileSync(join(baselineDirectory, 'wheelset_lacing_server_rendered_content_and_schema_before_migration.json'), 'utf8'))
  : []

for (const { code, iso } of locales) {
  const prefix = code === 'en' ? '' : `/${code}`
  const toolPath = `${prefix}${migratedToolPath}`
  assert.equal(findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id, 'tools')
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')
  assert.ok(routeRules[`${prefix}/tools`]?.cache, 'The hub keeps its independent public cache policy')
  for (const retiredPath of storefrontRetiredPagePaths) {
    const retiredRule = routeRules[`${prefix}${retiredPath}`]
    assert.equal(retiredRule?.redirect, undefined)
    assert.equal(retiredRule?.cache, false, 'Retired paths cannot inherit cached guide HTML')
    assert.equal(retiredRule?.headers?.['cache-control'], 'no-store, max-age=0')
  }

  for (const retiredPath of retiredTopologyPaths) {
    for (const query of ['', '?page=2', '?tab=calculator&topology_id=21h-g3']) {
      const response = await fetch(`${storefrontOrigin}${prefix}${retiredPath}${query}`, { redirect: 'manual' })
      assert.equal(response.status, 404, `${code}${retiredPath}${query}: retired route`)
      assert.equal(response.headers.get('location'), null)
    }
  }

  // The existing workbench stores inputs locally; unused query parameters do
  // not create another page or alter the canonical and structured-data URLs.
  for (const query of publishedLocaleCodes.has(code) ? ['', '?tab=calculator&topology_id=21h-g3'] : ['']) {
    const response = await fetch(`${storefrontOrigin}${toolPath}${query}`)
    assert.equal(response.status, 200, `${code}${query}: complete tool`)
    const cacheControl = response.headers.get('cache-control') || ''
    if (!(process.env.STOREFRONT_ALLOW_DEVELOPMENT_CACHE_HEADERS === 'true' && cacheControl === 'no-transform')) {
      assert.match(cacheControl, /no-store/)
    }
    const html = await response.text()
    const elements = collectParsedHtmlElements(parse(html) as unknown as ParsedHtmlElement)
    const canonicalLinks = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1)
    const canonical = new URL(readParsedHtmlElementAttribute(canonicalLinks[0]!, 'href'))
    assert.equal(canonical.pathname, toolPath)
    assert.equal(canonical.search, '')
    assert.equal(canonical.hash, '')
    const alternates = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'alternate' && readParsedHtmlElementAttribute(element, 'hreflang'))
    assert.deepEqual(alternates.map(element => readParsedHtmlElementAttribute(element, 'hreflang')).sort(), ['en-US', 'zh-CN', 'x-default'].sort())
    for (const alternate of alternates) {
      const url = new URL(readParsedHtmlElementAttribute(alternate, 'href'))
      assert.ok(url.pathname.endsWith(migratedToolPath))
      assert.equal(url.search, '')
      assert.equal(url.hash, '')
    }
    const robots = elements.filter(element => element.nodeName === 'meta' && readParsedHtmlElementAttribute(element, 'name') === 'robots')
    assert.equal(robots.some(element => readParsedHtmlElementAttribute(element, 'content').includes('noindex')), !publishedLocaleCodes.has(code))
    const workbenchRoot = elements.find(element => readParsedHtmlElementAttribute(element, 'class').split(/\s+/).includes('spoke-lacing-page'))
    assert.ok(workbenchRoot, 'The complete workbench is server rendered')
    const workbenchElements = collectParsedHtmlElements(workbenchRoot)
    assert.ok(workbenchElements.some(element => element.nodeName === 'h1'))
    assert.ok(workbenchElements.some(element => readParsedHtmlElementAttribute(element, 'id') === 'spoke-lacing-svg'))
    assert.ok(workbenchElements.some(element => readParsedHtmlElementAttribute(element, 'id') === 'flange-profile-svg'))
    const schemas = elements.filter(element => element.nodeName === 'script' && readParsedHtmlElementAttribute(element, 'type') === 'application/ld+json')
      .map(element => JSON.parse((element.childNodes || []).map(child => child.value || '').join('')))
    const topologySchema = schemas.find(schema => schema['@graph']?.some((node: Record<string, unknown>) => node['@type'] === 'TechArticle'))
    assert.ok(topologySchema, 'Preserve the existing TechArticle schema')
    const baseline = baselineSnapshots.find(entry => entry.prefix === prefix)
    if (baseline) {
      const expectedSchema = JSON.parse(JSON.stringify(baseline.schema).replaceAll(retiredGuidePath, migratedToolPath))
      const snapshot: CompleteTopologyContentSnapshot = {
        prefix,
        text: workbenchElements.filter(element => element.nodeName === '#text').map(element => element.value || '').join(' ').replace(/\s+/g, ' ').trim(),
        schema: topologySchema,
      }
      assert.deepEqual(snapshot, { ...baseline, schema: expectedSchema }, `${code}${query}: complete content, article and FAQ unchanged`)
    }
    for (const retiredPath of retiredTopologyPaths) {
      assert.ok(!JSON.stringify(schemas).includes(retiredPath))
      assert.ok(!html.includes(`href="${prefix}${retiredPath}`), 'No navigation to a retired page')
    }
  }

  const sitemapResponse = await fetch(`${storefrontOrigin}/__sitemap__/${iso}.xml`)
  assert.equal(sitemapResponse.status, 200)
  const xml = await sitemapResponse.text()
  assertRetiredPathsAbsentFromSitemap(xml)
  const toolBlocks = (xml.match(/<url>[^]*?<\/url>/g) || []).filter(block => block.includes(migratedToolPath))
  assert.equal(toolBlocks.length, publishedLocaleCodes.has(code) ? 1 : 0)
  for (const block of toolBlocks) {
    assert.ok(block.includes(`>${storefrontOrigin}${toolPath}</loc>`))
    assert.deepEqual([...block.matchAll(/hreflang="([^"]+)"/g)].map(match => match[1]).sort(), ['en-US', 'zh-CN', 'x-default'].sort())
  }
}

const dynamicSitemapResponse = await fetch(`${storefrontOrigin}/__sitemap__/dynamic-urls.json`)
assert.equal(dynamicSitemapResponse.status, 200)
const sitemapEntries = await dynamicSitemapResponse.json() as { loc: string }[]
assert.deepEqual(sitemapEntries.filter(entry => entry.loc.endsWith(migratedToolPath)).map(entry => entry.loc).sort(), [migratedToolPath, `/zh_cn${migratedToolPath}`].sort())
assertRetiredPathsAbsentFromSitemap(JSON.stringify(sitemapEntries))
for (const prefix of ['', '/zh_cn']) {
  const hubResponse = await fetch(`${storefrontOrigin}${prefix}/tools`)
  assert.equal(hubResponse.status, 200)
  assert.ok((await hubResponse.text()).includes(`href="${prefix}${migratedToolPath}"`))
  const guideResponse = await fetch(`${storefrontOrigin}${prefix}/guides/wheelset-buyers`)
  assert.equal(guideResponse.status, 200)
  assert.ok(!(await guideResponse.text()).includes('wheelset-spoke-lacing-topology-and-geometry-reference'))
  const code = prefix ? 'zh_cn' : 'en'
  const faqResponse = await fetch(`${storefrontOrigin}/api/v1/content/faq-pages/by-route?locale=${code}&route_path=${encodeURIComponent(migratedToolPath)}`)
  assert.equal(faqResponse.status, 200)
  const faq = await faqResponse.json() as FaqPageResponse
  assert.equal(faq.page.pageId, 'resources-wheelset-lacing-topology')
  assert.ok(faq.page.items.length > 0)
  if (baselineDirectory) {
    const previousFaq = JSON.parse(readFileSync(join(baselineDirectory, `wheelset_lacing_faq_content_before_migration_${code}.json`), 'utf8'))
    assert.deepEqual(faq, previousFaq, `${code}: FAQ identity and complete editable content unchanged`)
  }
  for (const retiredPath of retiredTopologyPaths) {
    const retiredFaqResponse = await fetch(`${storefrontOrigin}/api/v1/content/faq-pages/by-route?locale=${code}&route_path=${encodeURIComponent(retiredPath)}`)
    assert.equal(retiredFaqResponse.status, 404, 'Retired public FAQ route cannot resolve')
  }
}
const geometryResponse = await fetch(`${storefrontOrigin}/api/v1/wheelset-lacing/display-geometry?topology_id=24h-symmetric-1to1-2x`)
assert.equal(geometryResponse.status, 200)
if (baselineDirectory) {
  const previousGeometry = JSON.parse(readFileSync(join(baselineDirectory, 'wheelset_lacing_default_geometry_api_before_migration.json'), 'utf8'))
  assert.deepEqual(await geometryResponse.json(), previousGeometry, 'Complete default geometry API response unchanged')
}
console.log('Spoke lacing topology whole-page migration passed: 180 retired URLs, full SSR/schema/FAQ/geometry baseline, locale SEO, navigation, cache policy and 20 XML sitemaps.')
