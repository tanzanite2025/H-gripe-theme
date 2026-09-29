import assert from 'node:assert/strict'
import type { ApiRequestFunction } from '../app/composables/useApiRequest'
import {
  fetchSchwalbeTireRimWidthCombinationRules,
  schwalbeTireRimWidthCombinationRulesEndpoint,
} from '../app/data/tireguides/schwalbeTireRimWidthCombinationRules'

const request: ApiRequestFunction = async <T>() => ({
  code: 0,
  data: [
    {
      tire_width_min_mm: 20,
      tire_width_max_mm: 21,
      inner_rim_width_min_mm: 15,
      inner_rim_width_max_mm: 17,
      source_basis: 'ETRTO Standard 2024; Schwalbe possible combinations guidance',
      source_version: '05/2024',
      source_url: 'https://example.test/rim-width.pdf',
      source_checked_at: '2026-09-28',
    },
  ],
} as T)

assert.equal(
  schwalbeTireRimWidthCombinationRulesEndpoint,
  '/products/schwalbe-tire-rim-width-combination-rules',
)

const rules = await fetchSchwalbeTireRimWidthCombinationRules(request)
assert.deepEqual(rules, [
  {
    tireWidthMinMm: 20,
    tireWidthMaxMm: 21,
    innerRimWidthMinMm: 15,
    innerRimWidthMaxMm: 17,
    sourceBasis: 'ETRTO Standard 2024; Schwalbe possible combinations guidance',
    sourceVersion: '05/2024',
    sourceUrl: 'https://example.test/rim-width.pdf',
    sourceCheckedAt: '2026-09-28',
  },
])

const invalidRequest: ApiRequestFunction = async <T>() => ({
  code: 0,
  data: [{
    tire_width_min_mm: 0,
    tire_width_max_mm: 21,
    inner_rim_width_min_mm: 15,
    inner_rim_width_max_mm: 17,
    source_basis: 'ETRTO Standard 2024',
    source_version: '05/2024',
    source_url: 'https://example.test/rim-width.pdf',
    source_checked_at: '2026-09-28',
  }],
} as T)

await assert.rejects(
  () => fetchSchwalbeTireRimWidthCombinationRules(invalidRequest),
  /invalid tire_width_min_mm/,
)

console.log('Schwalbe rim-width rule adapter checks passed.')
