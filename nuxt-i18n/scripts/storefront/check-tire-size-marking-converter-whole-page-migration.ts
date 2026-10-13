import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import storefrontLocales from '../../app/i18n/locales.manifest.ts'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.ts'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.ts'

const storefrontOrigin = process.env.TIRE_SIZE_MARKING_CONVERTER_BASE_URL
if (!storefrontOrigin) {
  throw new Error('Set TIRE_SIZE_MARKING_CONVERTER_BASE_URL to a running storefront origin.')
}

const targetToolPath = '/tools/tire-size-marking-converter'
const retiredGuidePath = '/guides/tireguides/tire-size-markings'
const guidePageSource = readFileSync(join(process.cwd(), 'app/pages/guides/tireguides.vue'), 'utf8')
const guideTabsSource = readFileSync(join(process.cwd(), 'app/utils/pageSubNavigationData.ts'), 'utf8')
const innerTubeGuideSource = readFileSync(join(process.cwd(), 'app/components/tireguides/InnerTubeGuide.vue'), 'utf8')
const standaloneToolPageSource = readFileSync(join(process.cwd(), 'app/pages/tools/tire-size-marking-converter.vue'), 'utf8')
assert.ok(standaloneToolPageSource.includes('<TireSizeGuide />'), 'standalone tool must render the complete tire-size page')
assert.ok(!guideTabsSource.includes("id: 'tire-size-markings'"), 'retired tire-size page must not remain a virtual guide tab')
assert.ok(guidePageSource.includes("retiredGuideTabIds = new Set(['choose-inner-tube', 'tire-rim-width-compatibility', 'tire-size-markings'])"), 'guide page must 404 the retired tire-size tab')
assert.ok(innerTubeGuideSource.includes("localePath('/tools/tire-size-marking-converter')"), 'inner-tube guide must link to the migrated tire-size tool')
assert.ok(!innerTubeGuideSource.includes("localePath('/guides/tireguides/tire-size-markings')"), 'inner-tube guide must not link to the retired tire-size route')

const localeCodes = storefrontLocales.map(locale => locale.code)
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200',
  localeCodes,
  defaultLocale: 'en',
})

const readHtml = async (routePath: string) => {
  const response = await fetch(new URL(routePath, storefrontOrigin), {
    redirect: 'manual',
    signal: AbortSignal.timeout(60_000),
  })
  return { response, html: await response.text() }
}

const htmlIncludesText = (html: string, fragment: string) => (
  html.includes(fragment) || html.includes(fragment.replaceAll('&', '&amp;'))
)

const assertRetiredRoute = async (routePath: string) => {
  const { response } = await readHtml(routePath)
  assert.equal(response.status, 404, `${routePath} must return a real 404`)
  assert.equal(response.headers.get('location'), null, `${routePath} must not redirect`)
}

const extractCanonicalPath = (html: string) => {
  const canonicalMatch = html.match(/<link[^>]+rel="canonical"[^>]+href="([^"]+)"/i)
  assert.ok(canonicalMatch, 'canonical link missing')
  return new URL(canonicalMatch![1]!).pathname
}

const requiredStructuralFragments = [
  'tire-size-marking-mapping-calculator',
  'tire-size-availability-table',
  'tire-size-guide-actions',
  'schwalbe-tiresize.webp',
  '622',
]

const contentFragmentsByLocale = {
  en: [
    'How to Read Tire Size Markings',
    'Size marking mapping calculator',
    'Why 28 inch and 29 inch can both be 622',
  ],
  zh_cn: [
    '轮胎尺寸标注解读',
    '尺寸标注映射计算器',
    '为什么 28 英寸和 29 英寸都可能是 622',
  ],
} as const

const baselineDirectory = process.env.TIRE_SIZE_MARKING_CONVERTER_BASELINE_DIRECTORY
const baselineFragments = baselineDirectory
  ? ['tire-size-marking-mapping-calculator', 'tire-size-availability-table', 'tire-size-guide-actions', 'schwalbe-tiresize.webp', '622']
    .filter(fragment => existsSync(join(baselineDirectory, 'tire-size-markings-en.html'))
      && readFileSync(join(baselineDirectory, 'tire-size-markings-en.html'), 'utf8').includes(fragment))
  : []

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  const toolPath = `${prefix}${targetToolPath}`
  const retiredPath = `${prefix}${retiredGuidePath}`

  assert.equal(
    findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id,
    'tools',
    `${locale.code}: tire-size tool must belong to the tools navigation section`,
  )
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')

  for (const query of ['', '?tab=tire-size-markings', '?size=37-622&utm_source=forum']) {
    await assertRetiredRoute(`${retiredPath}${query}`)
  }
  await assertRetiredRoute(`${prefix}/guides/tireguides?tab=tire-size-markings`)

  const { response, html } = await readHtml(toolPath)
  assert.equal(response.status, 200, `${locale.code}: migrated tool must render`)
  for (const fragment of requiredStructuralFragments) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: missing structural fragment ${fragment}`)
  }
  for (const fragment of contentFragmentsByLocale[locale.code as keyof typeof contentFragmentsByLocale] || []) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: missing content fragment ${fragment}`)
  }
  for (const fragment of baselineFragments) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: baseline fragment ${fragment} was lost during migration`)
  }

  assert.equal(extractCanonicalPath(html), toolPath, `${locale.code}: canonical must use the tool route`)
  assert.ok(!html.includes(retiredGuidePath), `${locale.code}: migrated page must not expose the retired guide route`)
  assert.ok(html.includes(`href="${toolPath}"`), `${locale.code}: page navigation must expose the localized tool route`)
}

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  const hubResponse = await readHtml(`${prefix}/tools`)
  assert.equal(hubResponse.response.status, 200, `${locale.code}: tools hub must render`)
  assert.ok(hubResponse.html.includes(`href="${prefix}${targetToolPath}"`), `${locale.code}: tools hub link missing`)
}

const manifest = JSON.parse(readFileSync(join(process.cwd(), 'public/storefront-route-manifest.json'), 'utf8')) as {
  routes?: { key?: string, path?: string }[]
}
const manifestEntry = manifest.routes?.find(route => route.key === 'guides-tireguides-tire-size-markings')
assert.equal(manifestEntry?.path, targetToolPath, 'route manifest must publish the tire-size tool route')
assert.ok(!manifest.routes?.some(route => route.path === retiredGuidePath), 'route manifest must not publish the retired guide route')

console.log(`Tire-size whole-page migration passed: ${storefrontLocales.length * 4} retired route variants, ${storefrontLocales.length} localized tool pages, complete mapping/reference content, canonical/SEO, tools hub and manifest contracts.`)
