import assert from 'node:assert/strict'
import { existsSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import storefrontLocales from '../../app/i18n/locales.manifest.ts'
import { buildStorefrontRouteRules } from '../../config/storefront/route-rules.ts'
import { findPrimaryMegaNavSectionByPath, primaryMegaNavSections } from '../../app/utils/primaryMegaNav.ts'

const storefrontOrigin = process.env.FREEHUB_CASSETTE_FITMENT_BASE_URL
if (!storefrontOrigin) throw new Error('Set FREEHUB_CASSETTE_FITMENT_BASE_URL to a running storefront origin.')

const baselineDirectory = process.env.FREEHUB_CASSETTE_FITMENT_MIGRATION_BASELINE_DIRECTORY
const migratedToolPath = '/tools/freehub-cassette-fitment-engine'
const retiredGuidePath = '/guides/wheelset-buyers/choose-freehub'
const localeCodes = storefrontLocales.map(locale => locale.code)
const routeRules = buildStorefrontRouteRules({
  internalApiOrigin: 'http://localhost:9200',
  localeCodes,
  defaultLocale: 'en',
})
const allowMissingMatrixApi = process.env.FREEHUB_CASSETTE_FITMENT_ALLOW_MISSING_MATRIX_API === 'true'

const baselineBodyForLocale = (localeCode: string) => {
  if (!baselineDirectory) return ''
  const fileName = localeCode === 'en' ? 'en-old-route.body' : `${localeCode}-old-route.body`
  const filePath = join(baselineDirectory, 'http-baseline', fileName)
  return existsSync(filePath) ? readFileSync(filePath, 'utf8') : ''
}

const readHtml = async (path: string, init?: RequestInit) => {
  const response: Response = await fetch(new URL(path, storefrontOrigin), {
    ...init,
    signal: AbortSignal.timeout(60_000),
  })
  return { response, html: await response.text() }
}

const htmlIncludesText = (html: string, fragment: string) => (
  html.includes(fragment) || html.includes(fragment.replaceAll('&', '&amp;'))
)

const assertRetiredRoute = async (path: string) => {
  const { response } = await readHtml(path, { redirect: 'manual' })
  assert.equal(response.status, 404, `${path} must return a real 404`)
  assert.equal(response.headers.get('location'), null, `${path} must not redirect`)
}

const matrixResponse: Response = await fetch(new URL('/api/v1/fitment/drivetrain/matrix', storefrontOrigin), {
  signal: AbortSignal.timeout(60_000),
})
let matrixAnchorCount = 0
if (matrixResponse.status === 200) {
  const matrixPayload = await matrixResponse.json() as {
    data?: { rules?: { fitment_options?: { standard?: string }[] }[] }
  }
  const matrixRules = matrixPayload.data?.rules || []
  matrixAnchorCount = matrixRules.reduce((count, rule) => count + (rule.fitment_options?.length || 0), 0)
  assert.ok(matrixAnchorCount > 0, 'Drivetrain matrix must contain compatibility options')
} else if (!allowMissingMatrixApi) {
  assert.fail(`Drivetrain matrix API returned HTTP ${matrixResponse.status}`)
} else {
  console.warn(`Drivetrain matrix API returned HTTP ${matrixResponse.status}; continuing with SSR/static migration checks.`)
}

const requiredStructuralFragments = [
  'drivetrain-dataset',
  'installation-howto',
  'freehub-groupset-helper',
  'shimano-micro-spline-11-12-speed-mountain-freehub.webp',
  'sram-xd-11-12-speed-mountain-freehub.webp',
  'Campagnolo-8-9-10-11-N3W-freehub.webp',
]
const englishContentFragments = [
  'How to choose the right freehub',
  'Identify the drivetrain brand',
  'Choose the cassette specification',
  'Read the result',
  'Where to find the cassette information',
  'Compatibility Cheatsheet',
  'Technical Reference: Freehub Standards & Groupsets',
  'Direct compatibility answers',
]
const simplifiedChineseContentFragments = [
  '如何选择正确的塔基',
  '先确认传动品牌',
  '再选择飞轮规格',
  '最后查看结果',
  '在哪里查看飞轮信息',
  '兼容性速查',
  '技术参考：塔基标准与套件',
  'Direct compatibility answers',
]

for (const locale of storefrontLocales) {
  const prefix = locale.code === 'en' ? '' : `/${locale.code}`
  const retiredRouteRule = routeRules[`${prefix}${retiredGuidePath}`]
  assert.equal(retiredRouteRule?.redirect, undefined, `${locale.code}: retired route must not have a redirect rule`)

  for (const suffix of ['', '?tab=choose-freehub', '?brand=Shimano&cassette_spec=shimano_12s_road_11_34t']) {
    await assertRetiredRoute(`${prefix}${retiredGuidePath}${suffix}`)
  }

  const toolPath = `${prefix}${migratedToolPath}`
  assert.equal(findPrimaryMegaNavSectionByPath(toolPath, primaryMegaNavSections, localeCodes)?.id, 'tools')
  assert.equal(routeRules[toolPath]?.headers?.['cache-control'], 'no-store, max-age=0')

  const { response, html } = await readHtml(toolPath)
  assert.equal(response.status, 200, `${locale.code}: migrated tool must render`)
  const localizedContentFragments = locale.code === 'zh_cn'
    ? simplifiedChineseContentFragments
    : englishContentFragments
  for (const fragment of [...localizedContentFragments, ...requiredStructuralFragments]) {
    assert.ok(htmlIncludesText(html, fragment), `${locale.code}: missing ${fragment}`)
  }
  assert.ok(!html.includes(retiredGuidePath), `${locale.code}: migrated page must not expose retired route`)
  if (matrixAnchorCount > 0) {
    assert.ok((html.match(/id="[^"]+"/g) || []).length >= matrixAnchorCount, `${locale.code}: compatibility anchors disappeared`)
  }

  const canonicalMatch = html.match(/<link[^>]+rel="canonical"[^>]+href="([^"]+)"/i)
  assert.ok(canonicalMatch, `${locale.code}: canonical link missing`)
  assert.equal(new URL(canonicalMatch![1]!).pathname, toolPath, `${locale.code}: canonical must use the tool route`)

  const baselineBody = baselineBodyForLocale(locale.code)
  if (baselineBody) {
    for (const fragment of [...localizedContentFragments, ...requiredStructuralFragments].filter(fragment => baselineBody.includes(fragment))) {
      assert.ok(htmlIncludesText(html, fragment), `${locale.code}: baseline fragment ${fragment} was lost`)
    }
  }
}

for (const prefix of ['', '/zh_cn']) {
  const toolsResponse: Response = await fetch(new URL(`${prefix}/tools`, storefrontOrigin), {
    signal: AbortSignal.timeout(60_000),
  })
  assert.equal(toolsResponse.status, 200)
  assert.ok((await toolsResponse.text()).includes(`href="${prefix}${migratedToolPath}"`), `${prefix}: tools hub link missing`)
}

console.log(`Freehub whole-page migration passed: ${storefrontLocales.length * 3} retired route variants, ${storefrontLocales.length} tool pages, ${matrixAnchorCount} matrix options, complete content and cache/navigation contracts.`)
