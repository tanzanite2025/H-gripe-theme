import { expect, it, vi } from 'vitest'
import { applyMissingGalleryParentRepairWithReviewedFingerprint, isExactPositiveMediaAssociationIdentity, validateMediaAssociationDiagnosisReport, validateMissingGalleryParentRepairPreview } from './mediaReferenceAssociationDiagnosisAndAuditedGalleryParentRepair'
import axios from '@/utils/axios'
vi.mock('@/utils/axios', () => ({ default: { post: vi.fn() } }))
const largeID = '9007199254740993'
function createValidMediaAssociationDiagnosis() {
  return { asset_id: largeID, coverage: 'recognized_business_sources_only', checked_at: '2026-10-19T00:00:00Z', has_restricted_references: true, entries: [
    { category: 'catalog', resource_type: 'gallery_image', resource_id: '3', parent_resource_id: '4', label: '历史图片', field: 'url', usage_state: 'missing_parent', diagnosis: 'missing_parent_requires_business_review' },
  ], entry_limit: 200, entries_limited: false }
}
it('preserves IDs beyond JavaScript integer precision and rejects aliases and overflow', () => {
  expect(isExactPositiveMediaAssociationIdentity(largeID)).toBe(true)
  for (const value of [9007199254740992, '01', '+1', '0', '1.0', '18446744073709551616']) expect(isExactPositiveMediaAssociationIdentity(value)).toBe(false)
  expect(validateMediaAssociationDiagnosisReport(createValidMediaAssociationDiagnosis(), largeID).asset_id).toBe(largeID)
})
it.each(['asset', 'coverage', 'restricted', 'truncated', 'state', 'registration', 'identity'])('rejects contradictory %s diagnosis responses', (scenario) => {
  const report = createValidMediaAssociationDiagnosis()
  if (scenario === 'asset') report.asset_id = '2'
  if (scenario === 'coverage') report.coverage = 'all'
  if (scenario === 'restricted') report.has_restricted_references = undefined as unknown as boolean
  if (scenario === 'truncated') report.entries_limited = true
  if (scenario === 'state') report.entries[0]!.usage_state = 'active'
  if (scenario === 'registration') report.entries[0]!.diagnosis = 'registration_conflict_requires_ownership_evidence'
  if (scenario === 'identity') report.entries[0]!.resource_id = '01'
  expect(() => validateMediaAssociationDiagnosisReport(report, largeID)).toThrow()
})
it('checks every preview identity and fingerprint before sending a mutation', async () => {
  const preview = { asset_id: largeID, image_id: '3', missing_gallery_id: '4', destination_gallery_id: '5', destination_gallery_name: '草稿', image_title: '原图', fingerprint: 'a'.repeat(64) }
  expect(validateMissingGalleryParentRepairPreview(preview, largeID, '3', '5')).toBe(preview)
  expect(() => validateMissingGalleryParentRepairPreview({ ...preview, destination_gallery_id: '6' }, largeID, '3', '5')).toThrow()
  vi.mocked(axios.post).mockResolvedValue({ data: { data: { repaired: true, review: preview } } })
  await applyMissingGalleryParentRepairWithReviewedFingerprint(preview, '审核归属', new AbortController().signal)
  expect(axios.post).toHaveBeenCalledWith(`/api/admin/media/assets/${largeID}/missing-gallery-parent-repairs`, { image_id: '3', destination_gallery_id: '5', expected_fingerprint: 'a'.repeat(64), reason: '审核归属' }, expect.any(Object))
})
