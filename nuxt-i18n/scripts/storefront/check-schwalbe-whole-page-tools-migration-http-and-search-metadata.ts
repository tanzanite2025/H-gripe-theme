import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { parse } from 'parse5'
import locales from '../../app/i18n/locales.manifest.js'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.js'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.js'

const storefrontOrigin = process.env.SCHWALBE_SELECTOR_BASE_URL || 'http://localhost:9100'
const migratedSelectorPath = '/tools/schwalbe-tire-selector'
const previousSelectorPath = '/guides/tireguides/schwalbe-tire-selector'
const localeCodes = locales.map(locale => locale.code)
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200', localeCodes, defaultLocale: 'en',
})

type ParsedHtmlElement = {
  nodeName: string
  attrs?: { name: string; value: string }[]
  childNodes?: ParsedHtmlElement[]
}

function collectParsedHtmlElements(element: ParsedHtmlElement): ParsedHtmlElement[] {
  return [element, ...(element.childNodes || []).flatMap(collectParsedHtmlElements)]
}

function readParsedHtmlElementAttribute(element: ParsedHtmlElement, name: string): string {
  return element.attrs?.find(attribute => attribute.name === name)?.value || ''
}

for (const { code } of locales) {
  const prefix = code === 'en' ? '' : `/${code}`
  const expectedSelectorPath = `${prefix}${migratedSelectorPath}`
  assert.equal(findPrimaryMegaNavSectionByPath(expectedSelectorPath, primaryMegaNavSections, localeCodes)?.id, 'tools')
  assert.equal(routeRules[expectedSelectorPath]?.headers?.['cache-control'], 'no-store, max-age=0')
  assert.ok(routeRules[`${prefix}/tools`]?.cache)
  assert.equal(routeRules[`${prefix}/tools/future-calculator`], undefined, 'Future calculators must choose their own cache policy')

  for (const query of ['', '?page=2', '?search=Kojak&wheel_size=26-559']) {
    const response = await fetch(`${storefrontOrigin}${prefix}${previousSelectorPath}${query}`, { redirect: 'manual' })
    assert.equal(response.status, 301, `${code}${query}: server redirect`)
    const destination = new URL(response.headers.get('location') || '', storefrontOrigin)
    assert.equal(destination.pathname, expectedSelectorPath)
    assert.equal(destination.search, query)
  }
}

for (const code of ['en', 'zh_cn']) {
  const prefix = code === 'en' ? '' : `/${code}`
  for (const [query, canonicalQuery, shouldNoIndex] of [
    ['', '', false], ['?page=1', '', false], ['?page=2', '?page=2', false],
    ['?search=Kojak&wheel_size=26-559', '', true], ['?page=invalid', '', true],
  ] as const) {
    const response = await fetch(`${storefrontOrigin}${prefix}${migratedSelectorPath}${query}`)
    assert.equal(response.status, 200)
    const cacheControl = response.headers.get('cache-control') || ''
    // Nuxt's Vite development response replaces route headers with no-transform.
    // Production checks stay strict unless the local verifier opts in explicitly.
    if (!(process.env.STOREFRONT_ALLOW_DEVELOPMENT_CACHE_HEADERS === 'true' && cacheControl === 'no-transform')) {
      assert.match(cacheControl, /no-store/)
    }
    const html = await response.text()
    const elements = collectParsedHtmlElements(parse(html) as unknown as ParsedHtmlElement)
    const canonicalLinks = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'canonical')
    assert.equal(canonicalLinks.length, 1)
    const canonical = new URL(readParsedHtmlElementAttribute(canonicalLinks[0]!, 'href'))
    assert.equal(canonical.pathname, `${prefix}${migratedSelectorPath}`)
    assert.equal(canonical.search, canonicalQuery)
    const alternateLinks = elements.filter(element => element.nodeName === 'link' && readParsedHtmlElementAttribute(element, 'rel') === 'alternate' && readParsedHtmlElementAttribute(element, 'hreflang'))
    assert.equal(alternateLinks.length, localeCodes.length + 1)
    for (const alternate of alternateLinks) {
      const url = new URL(readParsedHtmlElementAttribute(alternate, 'href'))
      assert.ok(url.pathname.endsWith(migratedSelectorPath))
      assert.equal(url.search, canonicalQuery)
    }
    const robotElements = elements.filter(element => element.nodeName === 'meta' && readParsedHtmlElementAttribute(element, 'name') === 'robots')
    assert.equal(robotElements.some(element => readParsedHtmlElementAttribute(element, 'content').includes('noindex')), shouldNoIndex)
    assert.ok(elements.some(element => readParsedHtmlElementAttribute(element, 'id') === 'schwalbe-selector-title'))
    assert.ok(html.includes('schwalbe-selector__section-panel--intro'), 'Whole-page introduction must remain server rendered')
    assert.ok(html.includes('schwalbe-selector__section-panel--search'), 'Whole-page catalog must remain server rendered')
  }
  const hubResponse = await fetch(`${storefrontOrigin}${prefix}/tools`)
  assert.equal(hubResponse.status, 200)
  const hub = await hubResponse.text()
  const messages = JSON.parse(readFileSync(new URL(`../../app/i18n/messages/${code}/toolsNavigation.json`, import.meta.url), 'utf8'))
  assert.ok(hub.includes(messages.description), `${code}: localized hub description`)
  assert.ok(hub.includes(`href="${prefix}${migratedSelectorPath}"`))
}

const sitemapResponse = await fetch(`${storefrontOrigin}/__sitemap__/dynamic-urls.json`)
assert.equal(sitemapResponse.status, 200)
const sitemapEntries = await sitemapResponse.json() as { loc: string }[]
assert.ok(sitemapEntries.some(entry => entry.loc === migratedSelectorPath), 'Backend manifest sync must advertise the migrated selector')
assert.ok(sitemapEntries.some(entry => entry.loc === '/tools'))
assert.ok(sitemapEntries.every(entry => !entry.loc.includes(previousSelectorPath)), 'Sitemap must omit old routes, including stale FAQ paths')
console.log('Schwalbe whole-page migration passed: 60 localized redirects, SSR content and SEO states, navigation ownership, cache policy, and sitemap.')
