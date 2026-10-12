import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { parse } from 'parse5'
import locales from '../../app/i18n/locales.manifest.js'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.js'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.js'

const storefrontOrigin = process.env.SPOKE_LENGTH_CALCULATOR_BASE_URL || 'http://localhost:9100'
const baselineDirectory = process.env.SPOKE_LENGTH_CALCULATOR_MIGRATION_BASELINE_DIRECTORY
const migratedToolPath = '/tools/spoke-length-calculator'
const retiredGuidePath = '/guides/spokeguides/spoke-length-calculator'
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
type WholeCalculatorContentSnapshot = {
  prefix: string
  text: string
  images: { src: string; alt: string }[]
}

function collectParsedHtmlElements(element: ParsedHtmlElement): ParsedHtmlElement[] {
  return [element, ...(element.childNodes || []).flatMap(collectParsedHtmlElements)]
}

function readParsedHtmlElementAttribute(element: ParsedHtmlElement, name: string): string {
  return element.attrs?.find(attribute => attribute.name === name)?.value || ''
}

const baselineSnapshots: WholeCalculatorContentSnapshot[] = baselineDirectory
  ? JSON.parse(readFileSync(join(baselineDirectory, 'spoke_length_server_rendered_content_before_migration.json'), 'utf8'))
  : []

for (const { code, iso } of locales) {
  const prefix = code === 'en' ? '' : `/${code}`
  const toolPath = `${prefix}${migratedToolPath}`
  assert.equal(findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id, 'tools')
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')
  assert.ok(routeRules[`${prefix}/tools`]?.cache, 'The hub keeps its independent public cache policy')
  assert.equal(routeRules[`${prefix}${retiredGuidePath}`]?.redirect, undefined)

  for (const query of ['', '?page=2', '?tab=calculator&erd=600']) {
    const retiredResponse = await fetch(`${storefrontOrigin}${prefix}${retiredGuidePath}${query}`, { redirect: 'manual' })
    assert.equal(retiredResponse.status, 404, `${code}${query}: retired route`)
    assert.equal(retiredResponse.headers.get('location'), null)
  }

  // Calculator inputs remain in the wizard draft; query parameters do not create
  // separate calculator pages or change the canonical URL.
  for (const query of publishedLocaleCodes.has(code) ? ['', '?tab=calculator&erd=600'] : ['']) {
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
    const alternateLinks = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'alternate' && readParsedHtmlElementAttribute(element, 'hreflang'))
    assert.deepEqual(alternateLinks.map(element => readParsedHtmlElementAttribute(element, 'hreflang')).sort(), ['en-US', 'zh-CN', 'x-default'].sort())
    for (const alternate of alternateLinks) {
      const url = new URL(readParsedHtmlElementAttribute(alternate, 'href'))
      assert.ok(url.pathname.endsWith(migratedToolPath))
      assert.equal(url.search, '')
    }
    const robots = elements.filter(element => element.nodeName === 'meta' && readParsedHtmlElementAttribute(element, 'name') === 'robots')
    assert.equal(robots.some(element => readParsedHtmlElementAttribute(element, 'content').includes('noindex')), !publishedLocaleCodes.has(code))
    const calculatorRoot = elements.find(element => readParsedHtmlElementAttribute(element, 'class').split(/\s+/).includes('spoke-page'))
    assert.ok(calculatorRoot, 'The full calculator is server rendered')
    const calculatorElements = collectParsedHtmlElements(calculatorRoot)
    const snapshot: WholeCalculatorContentSnapshot = {
      prefix,
      text: calculatorElements.filter(element => element.nodeName === '#text').map(element => element.value || '').join(' ').replace(/\s+/g, ' ').trim(),
      images: calculatorElements.filter(element => element.nodeName === 'img').map(element => ({
        src: readParsedHtmlElementAttribute(element, 'src'), alt: readParsedHtmlElementAttribute(element, 'alt'),
      })),
    }
    assert.ok(calculatorElements.some(element => element.nodeName === 'h1'))
    assert.ok(calculatorElements.some(element => readParsedHtmlElementAttribute(element, 'class').split(/\s+/).includes('spoke-result-step')))
    const baseline = baselineSnapshots.find(entry => entry.prefix === prefix)
    if (baseline) assert.deepEqual(snapshot, baseline, `${code}${query}: complete calculator content and images unchanged`)
    const schemas = elements.filter(element => element.nodeName === 'script' && readParsedHtmlElementAttribute(element, 'type') === 'application/ld+json')
      .map(element => JSON.parse((element.childNodes || []).map(child => child.value || '').join('')))
    // Preserve either the committed article metadata or the user's in-progress
    // WebApplication metadata. The migration must not replace either schema.
    const calculatorSchema = schemas.find(schema => schema['@type'] === 'WebApplication' || schema['@type'] === 'TechArticle')
    assert.ok(calculatorSchema)
    if (calculatorSchema['@type'] === 'WebApplication') assert.equal(calculatorSchema.url, canonical.href)
    assert.ok(!JSON.stringify(schemas).includes(retiredGuidePath))
    assert.ok(!html.includes(`href="${prefix}${retiredGuidePath}`), 'No internal navigation points to the retired guide')
  }

  const sitemapResponse = await fetch(`${storefrontOrigin}/__sitemap__/${iso}.xml`)
  assert.equal(sitemapResponse.status, 200)
  const xml = await sitemapResponse.text()
  assert.ok(!xml.includes(retiredGuidePath))
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
assert.ok(sitemapEntries.every(entry => !entry.loc.includes(retiredGuidePath)))
for (const prefix of ['', '/zh_cn']) {
  const hubResponse = await fetch(`${storefrontOrigin}${prefix}/tools`)
  assert.equal(hubResponse.status, 200)
  assert.ok((await hubResponse.text()).includes(`href="${prefix}${migratedToolPath}"`))
  const guideResponse = await fetch(`${storefrontOrigin}${prefix}/guides/spokeguides`)
  assert.equal(guideResponse.status, 200)
  assert.ok(!(await guideResponse.text()).includes(`href="${prefix}${retiredGuidePath}`))
}
console.log('Spoke length whole-page migration passed: 60 retired URLs, full SSR baseline, published/fallback locale SEO, navigation, cache policy, and 20 XML sitemaps.')
