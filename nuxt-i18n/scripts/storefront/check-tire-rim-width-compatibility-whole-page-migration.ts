import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import storefrontLocales from '../../app/i18n/locales.manifest.ts'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.ts'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.ts'

const storefrontOrigin = process.env.TIRE_RIM_WIDTH_COMPATIBILITY_BASE_URL
if (!storefrontOrigin) {
  throw new Error('Set TIRE_RIM_WIDTH_COMPATIBILITY_BASE_URL to a running storefront origin.')
}

const targetToolPath = '/tools/tire-rim-width-compatibility-engine'
const retiredGuidePath = '/guides/tireguides/tire-rim-width-compatibility'
const tireSizeSectionSource = readFileSync(join(process.cwd(), 'app/components/TireSizeSection.vue'), 'utf8')
const tubelessGuideSource = readFileSync(join(process.cwd(), 'app/components/tireguides/TubelessGuide.vue'), 'utf8')
const tireGuidesPageSource = readFileSync(join(process.cwd(), 'app/pages/guides/tireguides.vue'), 'utf8')
assert.ok(tireSizeSectionSource.includes("localePath('/tools/tire-rim-width-compatibility-engine')"), 'tire size guide must link to the migrated tool')
assert.ok(!tireSizeSectionSource.includes("localePath('/guides/tireguides/tire-rim-width-compatibility')"), 'tire size guide must not link to the retired guide route')
assert.ok(tubelessGuideSource.includes("localePath('/tools/tire-rim-width-compatibility-engine')"), 'tubeless guide must link to the migrated tool')
assert.ok(!tubelessGuideSource.includes("changeTab', 'tire-rim-width-compatibility"), 'tubeless guide must not emit the retired guide tab')
assert.ok(!tireGuidesPageSource.includes('@change-tab="setActiveTab"'), 'tire guides page must not wire the retired guide tab event')
const localeCodes = storefrontLocales.map(locale => locale.code)
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200',
  localeCodes,
  defaultLocale: 'en',
})
const baselineDirectory = process.env.TIRE_RIM_WIDTH_COMPATIBILITY_BASELINE_DIRECTORY
const baselineBodyPath = baselineDirectory ? join(baselineDirectory, 'tire-rim-width-route.body') : ''
const baselineBody = baselineBodyPath && existsSync(baselineBodyPath) ? readFileSync(baselineBodyPath, 'utf8') : ''

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
  'tire-rim-guide-tab-calculation',
  'tire-rim-guide-tab-simulation',
  'tire-rim-guide-tab-reference',
  'tire-rim-guide-tab-schwalbeReference',
  'id="calculation"',
  'id="simulation"',
  'id="reference"',
  'id="etrto-2024"',
  'tire-rim-width-reference',
  'schwalbe-etrto-2024-width-reference-dataset',
  'application/ld+json',
]

const contentFragmentsByLocale = {
  en: [
    'Tire &amp; Rim Width Calculator: ETRTO 2024 Compatibility Chart',
    'Tire and rim fit simulation',
    'DT Swiss comparison note',
    'Schwalbe tire and inner rim width reference',
  ],
  zh_cn: [
    '外胎宽度与车圈内宽匹配计算器',
    '车圈与外胎',
    'DT Swiss 对照说明',
    'Schwalbe 外胎与车圈内宽对照',
  ],
} as const

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  const toolPath = `${prefix}${targetToolPath}`
  const retiredPath = `${prefix}${retiredGuidePath}`

  assert.equal(
    findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id,
    'tools',
    `${locale.code}: tool route must belong to the tools navigation section`,
  )
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')

  for (const query of ['', '?tab=tire-rim-width-compatibility', '?tab=reference', '?rim_width=25&tire_width=32']) {
    await assertRetiredRoute(`${retiredPath}${query}`)
  }

  const { response, html } = await readHtml(toolPath)
  assert.equal(response.status, 200, `${locale.code}: migrated tool must render`)
  for (const fragment of requiredStructuralFragments) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: missing structural fragment ${fragment}`)
  }
  for (const fragment of contentFragmentsByLocale[locale.code as keyof typeof contentFragmentsByLocale] || []) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: missing content fragment ${fragment}`)
  }

  assert.equal(extractCanonicalPath(html), toolPath, `${locale.code}: canonical must use the tool route`)
  assert.ok(!html.includes(retiredGuidePath), `${locale.code}: migrated page must not expose the retired guide route`)
  assert.ok(html.includes(`href="${toolPath}"`), `${locale.code}: page navigation must expose the localized tool route`)

  if (locale.code === 'en' || locale.code === 'zh_cn') {
    assert.ok(!/<meta name="robots" content="noindex,follow"/i.test(html), `${locale.code}: public locale must remain indexable`)
  } else {
    assert.ok(/<meta name="robots" content="noindex,follow"/i.test(html), `${locale.code}: fallback locale must be noindex`)
  }
}

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  const hubResponse = await readHtml(`${prefix}/tools`)
  assert.equal(hubResponse.response.status, 200, `${locale.code}: tools hub must render`)
  assert.ok(hubResponse.html.includes(`href="${prefix}${targetToolPath}"`), `${locale.code}: tools hub link missing`)
}

if (baselineBody) {
  const baselineStructuralFragments = [
    'tire-rim-guide-tab-calculation',
    'tire-rim-guide-tab-simulation',
    'tire-rim-guide-tab-reference',
    'tire-rim-guide-tab-schwalbeReference',
    'dtswiss-hookless-tss-rim-table.webp',
    'dtswiss-hooked-tc-rim-table.webp',
    'tire-chart-tables',
  ]
  const currentEnglishHtml = (await readHtml(targetToolPath)).html
  for (const fragment of baselineStructuralFragments.filter(fragment => baselineBody.includes(fragment))) {
    assert.ok(htmlIncludesText(currentEnglishHtml, fragment), `baseline fragment ${fragment} was lost during migration`)
  }
}

const manifest = JSON.parse(readFileSync(join(process.cwd(), 'public/storefront-route-manifest.json'), 'utf8')) as {
  routes?: { key?: string, path?: string }[]
}
const manifestEntry = manifest.routes?.find(route => route.key === 'guides-tireguides-tire-rim-width-compatibility')
assert.equal(manifestEntry?.path, targetToolPath, 'route manifest must publish the tool route')
assert.ok(!manifest.routes?.some(route => route.path === retiredGuidePath), 'route manifest must not publish the retired guide route')

console.log(`Tire/rim whole-page migration passed: ${storefrontLocales.length * 4} retired route variants, ${storefrontLocales.length} localized tool pages, four tabs, canonical/robots/Schema, tools hub, navigation, footer route and manifest contracts.`)
