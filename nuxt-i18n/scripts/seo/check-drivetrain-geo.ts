import locales from '../../app/i18n/locales.manifest.ts'
import type { DrivetrainCassetteRule, DrivetrainMatrixResponse } from '../../app/types/drivetrainFitment.ts'
import {
  DRIVETRAIN_BRAND_ALIASES,
  DRIVETRAIN_PAGE_LAST_MODIFIED,
  drivetrainRuleAnchor,
} from '../../app/utils/drivetrainFitmentGEO.ts'

const rawBaseUrl = String(process.env.DRIVETRAIN_GEO_BASE_URL || '').trim()
if (!rawBaseUrl) {
  throw new Error('Set DRIVETRAIN_GEO_BASE_URL to the running storefront origin before running this check.')
}

const baseUrl = new URL(rawBaseUrl)
const guidePath = '/guides/wheelset-buyers/choose-freehub'

const fail = (message: string): never => {
  throw new Error(`[drivetrain-geo] ${message}`)
}

const fetchJson = async <T>(path: string) => {
  const response = await fetch(new URL(path, baseUrl), { headers: { accept: 'application/json' } })
  if (!response.ok) fail(`${path} returned HTTP ${response.status}`)
  return response.json() as Promise<{ data: T }>
}

const matrixEnvelope = await fetchJson<DrivetrainMatrixResponse>('/api/v1/fitment/drivetrain/matrix')
const matrix = matrixEnvelope.data
if (!matrix.knowledge_as_of || !/^\d{4}-\d{2}-\d{2}$/.test(matrix.knowledge_as_of)) {
  fail(`matrix knowledge_as_of must be an ISO date, got ${matrix.knowledge_as_of || '<empty>'}`)
}
if (!matrix.rule_version) fail('matrix rule_version is empty')

const rules = matrix.rules as DrivetrainCassetteRule[]
const expectedAnchors = rules.flatMap(rule => rule.fitment_options.map(option => drivetrainRuleAnchor(rule, option)))
const expectedBrands = DRIVETRAIN_BRAND_ALIASES.flatMap(brand => [brand.name, ...brand.alternateNames])
if (new Set(expectedAnchors).size !== expectedAnchors.length) {
  fail('rule anchors are not unique')
}

for (const locale of locales) {
  const localePrefix = locale.code === 'en' ? '' : `/${locale.code}`
  const url = new URL(`${localePrefix}${guidePath}`, baseUrl)
  const response = await fetch(url, { headers: { accept: 'text/html', cookie: '' } })
  if (!response.ok) fail(`${locale.code} guide returned HTTP ${response.status}`)
  const html = await response.text()

  for (const anchor of expectedAnchors) {
    if (!new RegExp(`(?:id|href)=["']#?${anchor}["']`, 'i').test(html)) {
      fail(`${locale.code} guide is missing stable rule anchor ${anchor}`)
    }
  }
  for (const brand of expectedBrands) {
    if (!html.includes(brand)) fail(`${locale.code} guide is missing brand alias/entity ${brand}`)
  }
  if (!html.includes(`dateModified":"${DRIVETRAIN_PAGE_LAST_MODIFIED}`)) {
    fail(`${locale.code} JSON-LD dateModified does not match page metadata ${DRIVETRAIN_PAGE_LAST_MODIFIED}`)
  }
  if (!html.includes(`knowledge_as_of`) || !html.includes(matrix.knowledge_as_of)) {
    fail(`${locale.code} HTML does not expose the matrix knowledge_as_of ${matrix.knowledge_as_of}`)
  }
  if (!html.includes('Direct compatibility answers')) {
    fail(`${locale.code} guide is missing direct answer section`)
  }
}

console.log(`[drivetrain-geo] verified ${locales.length} locales, ${expectedAnchors.length} rule anchors, rule version ${matrix.rule_version}, knowledge cutoff ${matrix.knowledge_as_of}.`)
