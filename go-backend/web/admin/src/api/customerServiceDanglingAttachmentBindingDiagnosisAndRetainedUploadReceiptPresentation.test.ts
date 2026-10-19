import { expect, it } from 'vitest'
import { validateMediaAssociationDiagnosisReport } from './mediaReferenceAssociationDiagnosisAndAuditedGalleryParentRepair'
import { validateReadOnlyMediaUsageSummaryReport, type MediaUsageSummaryReport } from './mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'
import { mediaReferenceTypeLabel } from '@/lib/mediaReferencePresentation'
import { mediaUsageReferenceAuthorizedBusinessListDestination } from '@/lib/mediaUsageReferencePermissionCheckedBusinessListNavigation'

it('preserves the orphan binding identity without inventing a customer conversation or repair destination', () => {
  const entry = { category: 'support', resource_type: 'ticket_message_attachment_binding', resource_id: '18446744073709551615', parent_resource_id: '0', label: '附件绑定的消息记录不存在', field: 'media_asset_id', usage_state: 'missing_parent', diagnosis: 'missing_parent_requires_business_review' }
  const report = validateMediaAssociationDiagnosisReport({ asset_id: '1', coverage: 'recognized_business_sources_only', checked_at: '2026-10-19T00:00:00Z', has_restricted_references: false, entries: [entry], entry_limit: 200, entries_limited: false }, '1')
  expect(report.entries[0]!.resource_id).toBe('18446744073709551615')
  expect(report.entries[0]!.parent_resource_id).toBe('0')
  expect(mediaReferenceTypeLabel(entry.resource_type)).toBe('工单附件绑定')
  expect(mediaUsageReferenceAuthorizedBusinessListDestination(entry, () => true)).toBeNull()
})

it.each(['missing_parent', 'retained_deleted_parent', 'retained_upload_receipt', 'pending_upload'])('retains authoritative receipt state %s instead of interpreting every receipt as pending', (usage_state) => {
  const reference = { category: 'support', resource_type: 'media_upload_session', resource_id: '1', parent_resource_id: '23', label: '上传收据保留', field: 'upload_receipt', usage_state }
  const anomalous = usage_state === 'missing_parent'
  const report: MediaUsageSummaryReport = {
    coverage: 'recognized_business_sources_only', snapshot_started_at: '2026-10-19T00:00:00Z', completed_at: '2026-10-19T00:00:01Z', candidate_limit: 200,
    counts: { all: 1, no_recognized_references: 0, cross_business: 0, retained: 1, anomalous: Number(anomalous), restricted: 0 },
    rows: [{ asset_id: '1', title: 'Attachment', library_partition: 'customer_service', no_recognized_references: false, has_restricted_references: false,
      pending_business_draft: false, visible_reference_total: 1, cross_business: false, retained: true, anomalous, categories: ['support'], references: [reference], reference_preview_limited: false }],
  }
  expect(validateReadOnlyMediaUsageSummaryReport(report).rows[0]!.references[0]!.usage_state).toBe(usage_state)
  expect(mediaReferenceTypeLabel('media_upload_session')).toBe('附件上传收据')
  expect(mediaUsageReferenceAuthorizedBusinessListDestination(reference, () => true)).toBeNull()
})
