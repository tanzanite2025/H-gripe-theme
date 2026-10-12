import assert from 'node:assert/strict'
import { readFileSync, readdirSync } from 'node:fs'
import storefrontLocales from '../app/i18n/locales.manifest.ts'
import {
  getInnerTubeGuideActualContentLanguageTag,
  innerTubeGuideCanonicalRoutePath,
  innerTubeGuidePublishedLocaleCodes,
  innerTubeGuidePublishedLocaleSeoRoutes,
  innerTubeGuideUnpublishedLocaleSitemapExcludes,
  isInnerTubeGuidePublishedLocale,
  isInnerTubeGuideSitemapRouteAllowed,
  readInnerTubeGuideRouteLocaleCode,
} from '../app/utils/seo/innerTubeGuidePublishedLocaleIndexingPolicy.ts'

const routeManifest = JSON.parse(readFileSync(new URL('../public/storefront-route-manifest.json', import.meta.url), 'utf8'))
const innerTubeGuideManifestEntry = routeManifest.routes.find((entry: { path: string }) => entry.path === innerTubeGuideCanonicalRoutePath)
assert.ok(innerTubeGuideManifestEntry, 'The published guide must exist in the backend route manifest')
assert.deepEqual([...innerTubeGuideManifestEntry.sitemap_locales].sort(), [...innerTubeGuidePublishedLocaleCodes].sort())

for (const messageNamespace of ['guidesTireInnerTube', 'guidesTireguides']) {
  const translatedLocaleCodes = readdirSync(new URL(`../app/i18n/page-messages/${messageNamespace}/`, import.meta.url))
    .filter(filename => filename.endsWith('.json'))
    .map(filename => filename.slice(0, -5))
  assert.deepEqual(translatedLocaleCodes.sort(), [...innerTubeGuidePublishedLocaleCodes].sort(),
    'Update indexing policy and route manifest together when publishing a complete translation')
}

assert.deepEqual(innerTubeGuidePublishedLocaleSeoRoutes, [
  { code: 'en', path: '/tools/valve-length-fitment-calculator', hreflang: 'en-US' },
  { code: 'zh_cn', path: '/zh_cn/tools/valve-length-fitment-calculator', hreflang: 'zh-CN' },
])

for (const locale of storefrontLocales) {
  const localizedGuidePath = `${locale.code === 'en' ? '' : `/${locale.code}`}${innerTubeGuideCanonicalRoutePath}`
  const hasPublishedTranslation = locale.code === 'en' || locale.code === 'zh_cn'
  assert.equal(isInnerTubeGuidePublishedLocale(locale.code), hasPublishedTranslation)
  assert.equal(getInnerTubeGuideActualContentLanguageTag(locale.code), locale.code === 'zh_cn' ? 'zh-CN' : 'en-US')
  for (const routePathOrUrl of [localizedGuidePath, `${localizedGuidePath}/`, `${localizedGuidePath}?rim_depth=60#calculator`, `https://store.example${localizedGuidePath}`]) {
    assert.equal(readInnerTubeGuideRouteLocaleCode(routePathOrUrl), locale.code)
    assert.equal(isInnerTubeGuideSitemapRouteAllowed(routePathOrUrl), hasPublishedTranslation)
  }
  assert.equal(innerTubeGuideUnpublishedLocaleSitemapExcludes.includes(localizedGuidePath), !hasPublishedTranslation)
}

assert.equal(innerTubeGuideUnpublishedLocaleSitemapExcludes.length, storefrontLocales.length - 2)
assert.equal(isInnerTubeGuideSitemapRouteAllowed(`/en${innerTubeGuideCanonicalRoutePath}`), false, 'Do not publish a duplicate prefixed English URL')
for (const unrelatedPath of ['/guides/tireguides', '/fr/guides/tireguides/tire-frame-clearance', '/fr/tools/valve-length-fitment-calculator-more']) {
  assert.equal(readInnerTubeGuideRouteLocaleCode(unrelatedPath), null)
  assert.equal(isInnerTubeGuideSitemapRouteAllowed(unrelatedPath), true, 'The indexing policy must stay scoped to this guide')
}

console.log(`Inner-tube guide locale policy passed: route manifest, both translation namespaces, ${storefrontLocales.length} locales, unrelated-route isolation`)
