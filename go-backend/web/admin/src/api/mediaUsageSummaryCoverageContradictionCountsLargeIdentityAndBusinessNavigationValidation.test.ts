import { describe, expect, it } from 'vitest'
import { validateReadOnlyMediaUsageSummaryReport, type MediaUsageSummaryReport, type MediaUsageSummaryReference } from './mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'
import { mediaUsageReferenceAuthorizedBusinessListDestination } from '@/lib/mediaUsageReferencePermissionCheckedBusinessListNavigation'

function createValidRecognizedMediaUsageSummaryReport(): MediaUsageSummaryReport {
  return { coverage: 'recognized_business_sources_only', candidate_limit: 200, snapshot_started_at: '2026-10-19T01:00:00Z', completed_at: '2026-10-19T01:00:01Z',
    counts: { all: 1, no_recognized_references: 1, cross_business: 0, retained: 0, anomalous: 0, restricted: 0 },
    rows: [{ asset_id: '9007199254740993', title: 'recognized scope only', library_partition: 'shared', no_recognized_references: true,
      has_restricted_references: false, pending_business_draft: false, visible_reference_total: 0, cross_business: false, retained: false, anomalous: false,
      categories: [], references: [], reference_preview_limited: false }] }
}

it('preserves large decimal identities and explicit recognized-source coverage', () => {
  expect(validateReadOnlyMediaUsageSummaryReport(createValidRecognizedMediaUsageSummaryReport()).rows[0]!.asset_id).toBe('9007199254740993')
})

it.each(['coverage', 'counts', 'restricted-zero', 'missing-state', 'duplicate', 'negative-total', 'numeric-id', 'timestamps', 'preview', 'null-row'])(
  'rejects %s without converting invalid data into an unused conclusion', (cause) => {
    const report = createValidRecognizedMediaUsageSummaryReport()
    const row = report.rows[0]!
    if (cause === 'coverage') (report as unknown as Record<string, unknown>).coverage = 'all_sources'
    if (cause === 'counts') report.counts.no_recognized_references = 0
    if (cause === 'restricted-zero') row.has_restricted_references = true
    if (cause === 'missing-state') delete (row as unknown as Record<string, unknown>).no_recognized_references
    if (cause === 'duplicate') report.rows.push({ ...row })
    if (cause === 'negative-total') row.visible_reference_total = -1
    if (cause === 'numeric-id') (row as unknown as Record<string, unknown>).asset_id = 42
    if (cause === 'timestamps') report.completed_at = '2026-10-18T01:00:00Z'
    if (cause === 'preview') row.reference_preview_limited = true
    if (cause === 'null-row') report.rows = [null as never]
    expect(() => validateReadOnlyMediaUsageSummaryReport(report)).toThrow()
  },
)

it('accepts a restricted-only row without exposing its hidden reference total', () => {
  const report = createValidRecognizedMediaUsageSummaryReport()
  report.rows[0]!.no_recognized_references = false
  report.rows[0]!.has_restricted_references = true
  report.counts.no_recognized_references = 0
  report.counts.restricted = 1
  expect(validateReadOnlyMediaUsageSummaryReport(report).rows[0]!.visible_reference_total).toBe(0)
})

it('keeps an unbound business draft out of the unused filter and rejects contradictory draft states', () => {
  const report = createValidRecognizedMediaUsageSummaryReport()
  Object.assign(report.rows[0]!, { no_recognized_references: false, pending_business_draft: true, retained: true })
  report.counts.no_recognized_references = 0
  report.counts.retained = 1
  expect(validateReadOnlyMediaUsageSummaryReport(report).rows[0]!.pending_business_draft).toBe(true)
  report.rows[0]!.retained = false
  expect(() => validateReadOnlyMediaUsageSummaryReport(report)).toThrow()
})

describe('authorized business navigation', () => {
  const reference: MediaUsageSummaryReference = { category: 'catalog', resource_type: 'product_media', resource_id: '9007199254740993', parent_resource_id: '10', label: 'media', field: 'url', usage_state: 'active' }
  it('opens the existing business list only with its business view permission', () => {
    expect(mediaUsageReferenceAuthorizedBusinessListDestination(reference, (permission) => permission === 'product:view')).toEqual({ name: 'CatalogProducts' })
    expect(mediaUsageReferenceAuthorizedBusinessListDestination(reference, () => false)).toBeNull()
  })
  it.each(['missing_parent', 'retained_deleted_parent', 'retained_deleted_placement', 'immutable_history'])('does not invent a link for %s records', (usage_state) => {
    expect(mediaUsageReferenceAuthorizedBusinessListDestination({ ...reference, usage_state }, () => true)).toBeNull()
  })
  it.each(['ticket_message', 'media_asset_registration', 'unknown', 'https://untrusted.example.test/'])('does not invent a destination for %s', (resource_type) => {
    expect(mediaUsageReferenceAuthorizedBusinessListDestination({ ...reference, resource_type }, () => true)).toBeNull()
  })
})
