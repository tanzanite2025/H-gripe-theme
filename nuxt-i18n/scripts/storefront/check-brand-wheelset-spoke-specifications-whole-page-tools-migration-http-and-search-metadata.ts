import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { parse } from 'parse5'
import locales from '../../app/i18n/locales.manifest.js'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.js'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.js'

const storefrontOrigin = process.env.WHEELSET_SPOKE_SPECIFICATIONS_BASE_URL || 'http://localhost:9100'
const baselineDirectory = process.env.WHEELSET_SPOKE_SPECIFICATIONS_MIGRATION_BASELINE_DIRECTORY
const migratedToolPath = '/tools/brand-wheelset-spoke-specs'
const retiredGuidePath = '/guides/spokeguides/brand-wheelset-spoke-specs'
const localeCodes = locales.map(locale => locale.code)
const publishedLocaleCodes = new Set(['en', 'zh_cn'])
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200', localeCodes, defaultLocale: 'en',
})

type ParsedHtmlElement = {
  nodeName: string
  value?: string
  attrs?: { name: string; value: string }[]
  childNodes?: ParsedHtmlElement[]
}
type WholePageContentSnapshot = {
  prefix: string
  query: string
  rows: { id: string; text: string }[]
  introduction: string[]
  shimanoNote: string[]
}

function collectParsedHtmlElements(element: ParsedHtmlElement): ParsedHtmlElement[] {
  return [element, ...(element.childNodes || []).flatMap(collectParsedHtmlElements)]
}

function readParsedHtmlElementAttribute(element: ParsedHtmlElement, name: string): string {
  return element.attrs?.find(attribute => attribute.name === name)?.value || ''
}

function readNormalizedServerRenderedElementText(element: ParsedHtmlElement): string {
  return collectParsedHtmlElements(element)
    .filter(child => child.nodeName === '#text')
    .map(child => child.value || '').join(' ').replace(/\s+/g, ' ').trim()
}

function hasParsedHtmlElementClass(element: ParsedHtmlElement, className: string): boolean {
  return readParsedHtmlElementAttribute(element, 'class').split(/\s+/).includes(className)
}

function captureWholePageSpecificationContent(elements: ParsedHtmlElement[], prefix: string, query: string): WholePageContentSnapshot {
  return {
    prefix,
    query,
    rows: elements
      .filter(element => element.nodeName === 'tr' && readParsedHtmlElementAttribute(element, 'id').startsWith('wheelset-'))
      .map(element => ({ id: readParsedHtmlElementAttribute(element, 'id'), text: readNormalizedServerRenderedElementText(element) })),
    introduction: elements.filter(element => hasParsedHtmlElementClass(element, 'wheelset-spoke-lookup__intro'))
      .map(readNormalizedServerRenderedElementText),
    shimanoNote: elements.filter(element => hasParsedHtmlElementClass(element, 'brand-wheelset-spoke-repair-kit-products__note'))
      .map(readNormalizedServerRenderedElementText),
  }
}

const baselineSnapshots: WholePageContentSnapshot[] = baselineDirectory
  ? JSON.parse(readFileSync(join(baselineDirectory, 'wheelset_spoke_specifications_server_rendered_rows_before_migration.json'), 'utf8'))
  : []

for (const { code, iso } of locales) {
  const prefix = code === 'en' ? '' : `/${code}`
  const toolPath = `${prefix}${migratedToolPath}`
  assert.equal(findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id, 'tools')
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')
  assert.ok(routeRules[`${prefix}/tools`]?.cache, 'Tool hub retains its separate cache policy')
  assert.equal(routeRules[`${prefix}${retiredGuidePath}`]?.redirect, undefined)
  for (const query of ['', '?page=2', '?model=erc-1100-dicut-35']) {
    const response = await fetch(`${storefrontOrigin}${prefix}${retiredGuidePath}${query}`, { redirect: 'manual' })
    assert.equal(response.status, 404, `${code}${query}: retired guide`)
    assert.equal(response.headers.get('location'), null, 'Retired routes have no compatibility redirect')
  }

  const seoStates = publishedLocaleCodes.has(code)
    ? [
        ['', '', false], ['?page=1', '', false], ['?page=2', '?page=2', false],
        ['?model=erc-1100-dicut-35', '', true], ['?page=invalid', '', true],
        ['?page=999', '?page=7', false], ['?search=ERC', '', true],
      ] as const
    : [['', '', true]] as const

  for (const [query, canonicalQuery, shouldNoIndex] of seoStates) {
    const response = await fetch(`${storefrontOrigin}${toolPath}${query}`)
    assert.equal(response.status, 200, `${code}${query}: migrated tool`)
    const cacheControl = response.headers.get('cache-control') || ''
    if (!(process.env.STOREFRONT_ALLOW_DEVELOPMENT_CACHE_HEADERS === 'true' && cacheControl === 'no-transform')) {
      assert.match(cacheControl, /no-store/)
    }
    const elements = collectParsedHtmlElements(parse(await response.text()) as unknown as ParsedHtmlElement)
    const canonicalLinks = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1)
    const canonical = new URL(readParsedHtmlElementAttribute(canonicalLinks[0]!, 'href'))
    // Unpublished locales remain noindex; every canonical uses the new route.
    assert.ok(canonical.pathname.endsWith(migratedToolPath))
    if (publishedLocaleCodes.has(code)) assert.equal(canonical.pathname, toolPath)
    assert.equal(canonical.search, canonicalQuery)
    const alternates = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'alternate' && readParsedHtmlElementAttribute(element, 'hreflang'))
    assert.deepEqual(alternates.map(element => readParsedHtmlElementAttribute(element, 'hreflang')).sort(), ['en-US', 'x-default', 'zh-CN'].sort())
    for (const alternate of alternates) {
      const url = new URL(readParsedHtmlElementAttribute(alternate, 'href'))
      assert.ok(url.pathname.endsWith(migratedToolPath))
      assert.equal(url.search, canonicalQuery)
    }
    const robots = elements.filter(element => element.nodeName === 'meta' && readParsedHtmlElementAttribute(element, 'name') === 'robots')
    assert.equal(robots.some(element => readParsedHtmlElementAttribute(element, 'content').includes('noindex')), shouldNoIndex, `${code}${query}: robots`)
    const content = captureWholePageSpecificationContent(elements, prefix, query)
    assert.ok(content.rows.length > 0)
    assert.equal(content.introduction.length, 1)
    assert.equal(content.shimanoNote.length, 1)
    assert.ok(elements.some(element => readParsedHtmlElementAttribute(element, 'id') === 'brand-wheelset-spoke-repair-kit-products'))
    if (query.startsWith('?model=')) assert.ok(content.rows.some(row => row.id === 'wheelset-dt-swiss-erc-1100-dicut-35'))
    const baseline = baselineSnapshots.find(snapshot => snapshot.prefix === prefix && snapshot.query === query)
    if (baseline) assert.deepEqual(content, baseline, `${code}${query}: complete specs, introduction, and repair notes unchanged`)
    const schemaScripts = elements.filter(element => element.nodeName === 'script' && readParsedHtmlElementAttribute(element, 'type') === 'application/ld+json')
    const schemas = schemaScripts.map(element => JSON.parse((element.childNodes || []).map(child => child.value || '').join('')))
    const collection = schemas.find(schema => schema['@type'] === 'CollectionPage')
    assert.ok(collection, 'Existing CollectionPage and ItemList remain present')
    assert.equal(collection.url, canonical.href)
    assert.equal(collection.mainEntity.numberOfItems, content.rows.length)
    for (const item of collection.mainEntity.itemListElement) {
      assert.ok(new URL(item.url).pathname.endsWith(migratedToolPath))
      assert.ok(item.url.includes('#wheelset-'))
      assert.ok(!item.url.includes(retiredGuidePath))
    }
  }

  const xmlResponse = await fetch(`${storefrontOrigin}/__sitemap__/${iso}.xml`)
  assert.equal(xmlResponse.status, 200)
  const xml = await xmlResponse.text()
  assert.ok(!xml.includes(retiredGuidePath), `${code}: no retired routes in XML or alternate links`)
  const toolBlocks = (xml.match(/<url>[^]*?<\/url>/g) || []).filter(block => block.includes(migratedToolPath))
  assert.equal(toolBlocks.length, publishedLocaleCodes.has(code) ? 1 : 0, `${code}: published sitemap locales only`)
  for (const block of toolBlocks) {
    assert.ok(block.includes(`>${storefrontOrigin}${toolPath}</loc>`))
    const alternateLanguages = [...block.matchAll(/hreflang="([^"]+)"/g)].map(match => match[1]).sort()
    assert.deepEqual(alternateLanguages, ['en-US', 'x-default', 'zh-CN'].sort())
  }
}

if (baselineDirectory) {
  const response = await fetch(`${storefrontOrigin}/api/v1/wheelset-spoke-specs/models`)
  assert.equal(response.status, 200)
  const payload = await response.json() as { data?: { models?: unknown[] }; models?: unknown[] }
  const models = payload.data?.models || payload.models
  const beforeMigrationModels = JSON.parse(readFileSync(join(baselineDirectory, 'wheelset_spoke_specifications_api_models_before_migration.json'), 'utf8'))
  assert.deepEqual(models, beforeMigrationModels, 'All public API spoke and nipple values remain unchanged')
}

const sitemapResponse = await fetch(`${storefrontOrigin}/__sitemap__/dynamic-urls.json`)
assert.equal(sitemapResponse.status, 200)
const sitemapEntries = await sitemapResponse.json() as { loc: string }[]
assert.deepEqual(sitemapEntries.filter(entry => entry.loc.endsWith(migratedToolPath)).map(entry => entry.loc).sort(), [migratedToolPath, `/zh_cn${migratedToolPath}`].sort())
assert.ok(sitemapEntries.every(entry => !entry.loc.includes(retiredGuidePath)))
for (const code of publishedLocaleCodes) {
  const prefix = code === 'en' ? '' : `/${code}`
  const response = await fetch(`${storefrontOrigin}${prefix}/tools`)
  assert.equal(response.status, 200)
  const html = await response.text()
  assert.ok(html.includes(`href="${prefix}${migratedToolPath}"`))
}
console.log('Brand wheelset whole-page migration passed: 60 retired URLs, published/fallback locale SEO, complete specifications, model links, cache policy, and 20 XML sitemaps.')
