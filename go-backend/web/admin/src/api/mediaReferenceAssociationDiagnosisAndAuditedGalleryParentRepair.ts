import axios from '@/utils/axios'
import { requireApiObject, requireApiObjectField, unwrapApiPayload } from '@/utils/apiResponse'
import type { MediaUsageSummaryReference } from './mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'

export interface MediaAssociationDiagnosisEntry extends MediaUsageSummaryReference {
  diagnosis: 'registration_conflict_requires_ownership_evidence' | 'missing_parent_requires_business_review' | 'deleted_history_must_remain_retained'
}
export interface MediaAssociationDiagnosisReport {
  asset_id: string
  coverage: 'recognized_business_sources_only'
  checked_at: string
  has_restricted_references: boolean
  entries: MediaAssociationDiagnosisEntry[]
  entry_limit: 200
  entries_limited: boolean
}
export interface MissingGalleryParentRepairPreview {
  asset_id: string
  image_id: string
  missing_gallery_id: string
  destination_gallery_id: string
  destination_gallery_name: string
  image_title: string
  fingerprint: string
}
export function isExactPositiveMediaAssociationIdentity(value: unknown): value is string {
  return typeof value === 'string' && /^[1-9]\d{0,19}$/.test(value) && BigInt(value) <= 18446744073709551615n
}
export function validateMediaAssociationDiagnosisReport(value: unknown, assetID: string): MediaAssociationDiagnosisReport {
  const report = requireApiObject(value, 'media association diagnosis') as unknown as MediaAssociationDiagnosisReport
  if (!isExactPositiveMediaAssociationIdentity(assetID) || report.asset_id !== assetID || report.coverage !== 'recognized_business_sources_only' ||
    typeof report.checked_at !== 'string' || !Number.isFinite(Date.parse(report.checked_at)) || typeof report.has_restricted_references !== 'boolean' ||
    report.entry_limit !== 200 || typeof report.entries_limited !== 'boolean' || !Array.isArray(report.entries) || report.entries.length > 200 ||
    (report.entries_limited && report.entries.length !== 200)) throw new Error('Invalid media association diagnosis')
  for (const entry of report.entries) {
    if (!entry || ['resource_type', 'category', 'label', 'field', 'usage_state'].some((key) => typeof entry[key as keyof MediaAssociationDiagnosisEntry] !== 'string') ||
      ![entry.resource_id, entry.parent_resource_id].every((id) => id === '0' || isExactPositiveMediaAssociationIdentity(id))) throw new Error('Invalid association identity')
    if (entry.diagnosis === 'registration_conflict_requires_ownership_evidence') {
      if (entry.category !== 'registration' || entry.resource_id !== '0' || entry.parent_resource_id !== '0') throw new Error('Invalid registration redaction')
    } else if (entry.diagnosis === 'missing_parent_requires_business_review') {
      if (entry.usage_state !== 'missing_parent' || entry.resource_id === '0') throw new Error('Invalid missing parent diagnosis')
    } else if (entry.diagnosis === 'deleted_history_must_remain_retained') {
      if (!['retained_deleted_parent', 'retained_deleted_placement'].includes(entry.usage_state)) throw new Error('Invalid retained history diagnosis')
    } else throw new Error('Unknown media association diagnosis')
  }
  return report
}
export function validateMissingGalleryParentRepairPreview(value: unknown, assetID: string, imageID: string, destinationID: string): MissingGalleryParentRepairPreview {
  const preview = requireApiObject(value, 'gallery repair preview') as unknown as MissingGalleryParentRepairPreview
  if (![assetID, imageID, destinationID].every(isExactPositiveMediaAssociationIdentity) || preview.asset_id !== assetID || preview.image_id !== imageID ||
    preview.destination_gallery_id !== destinationID || !(preview.missing_gallery_id === '0' || isExactPositiveMediaAssociationIdentity(preview.missing_gallery_id)) ||
    preview.missing_gallery_id === destinationID || typeof preview.image_title !== 'string' || typeof preview.destination_gallery_name !== 'string' ||
    typeof preview.fingerprint !== 'string' || !/^[a-f0-9]{64}$/.test(preview.fingerprint)) throw new Error('Invalid gallery repair preview')
  return preview
}
async function requestMediaAssociationOperationPayload(path: string, request: Promise<unknown>) {
  return requireApiObject(unwrapApiPayload(await request, path), path)
}
export async function requestMediaReferenceAssociationDiagnosis(assetID: string, signal: AbortSignal) {
  if (!isExactPositiveMediaAssociationIdentity(assetID)) throw new Error('Invalid asset identity')
  const path = `/api/admin/media/assets/${assetID}/association-diagnosis`
  const payload = await requestMediaAssociationOperationPayload(path, axios.get(path, { signal, suppressGlobalErrorToast: true }))
  return validateMediaAssociationDiagnosisReport(requireApiObjectField(payload, 'diagnosis', path), assetID)
}
export async function requestMissingGalleryParentRepairPreview(assetID: string, imageID: string, destinationID: string, signal: AbortSignal) {
  if (![assetID, imageID, destinationID].every(isExactPositiveMediaAssociationIdentity)) throw new Error('Invalid repair identities')
  const path = `/api/admin/media/assets/${assetID}/missing-gallery-parent-repair-previews`
  const payload = await requestMediaAssociationOperationPayload(path, axios.post(path, { image_id: imageID, destination_gallery_id: destinationID }, { signal, suppressGlobalErrorToast: true }))
  return validateMissingGalleryParentRepairPreview(requireApiObjectField(payload, 'preview', path), assetID, imageID, destinationID)
}
export async function applyMissingGalleryParentRepairWithReviewedFingerprint(preview: MissingGalleryParentRepairPreview, reason: string, signal: AbortSignal) {
  validateMissingGalleryParentRepairPreview(preview, preview.asset_id, preview.image_id, preview.destination_gallery_id)
  const path = `/api/admin/media/assets/${preview.asset_id}/missing-gallery-parent-repairs`
  const payload = await requestMediaAssociationOperationPayload(path, axios.post(path, { image_id: preview.image_id, destination_gallery_id: preview.destination_gallery_id, expected_fingerprint: preview.fingerprint, reason }, { signal, suppressGlobalErrorToast: true }))
  if (payload.repaired !== true) throw new Error('Repair result was not confirmed')
  validateMissingGalleryParentRepairPreview(requireApiObjectField(payload, 'review', path), preview.asset_id, preview.image_id, preview.destination_gallery_id)
}
