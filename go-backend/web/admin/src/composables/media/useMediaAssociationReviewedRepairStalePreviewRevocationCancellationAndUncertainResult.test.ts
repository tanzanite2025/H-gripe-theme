import { afterEach, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import { applyMissingGalleryParentRepairWithReviewedFingerprint, requestMediaReferenceAssociationDiagnosis, requestMissingGalleryParentRepairPreview, type MediaAssociationDiagnosisReport, type MissingGalleryParentRepairPreview } from '@/api/mediaReferenceAssociationDiagnosisAndAuditedGalleryParentRepair'
import { useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation } from './useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation'
vi.mock('@/api/mediaReferenceAssociationDiagnosisAndAuditedGalleryParentRepair', async (original) => ({ ...await original<object>(), requestMediaReferenceAssociationDiagnosis: vi.fn(), requestMissingGalleryParentRepairPreview: vi.fn(), applyMissingGalleryParentRepairWithReviewedFingerprint: vi.fn() }))
const scopes: EffectScope[] = []
const entry = { category: 'catalog', resource_type: 'gallery_image', resource_id: '3', parent_resource_id: '4', label: '图片', field: 'url', usage_state: 'missing_parent', diagnosis: 'missing_parent_requires_business_review' } as const
const report = { asset_id: '1', has_restricted_references: false, entries: [entry] } as unknown as MediaAssociationDiagnosisReport
const preview = { asset_id: '1', image_id: '3', missing_gallery_id: '4', destination_gallery_id: '5', fingerprint: 'a'.repeat(64), image_title: '图片', destination_gallery_name: '草稿' } satisfies MissingGalleryParentRepairPreview
function createDeferredAssociationResponse<T>() { let resolve!: (value: T) => void; let reject!: (value: unknown) => void; const promise = new Promise<T>((yes, no) => { resolve = yes; reject = no }); return { promise, resolve, reject } }
function createDisposableAssociationReview() {
  const scope = effectScope(); scopes.push(scope)
  const allowed = ref(true), canRepair = ref(true), identity = ref('actor1:media:view'), repaired = vi.fn()
  const review = scope.run(() => useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation(allowed, canRepair, identity, repaired))!
  return { scope, allowed, canRepair, identity, repaired, review }
}
afterEach(() => { scopes.splice(0).forEach((scope) => scope.stop()); vi.resetAllMocks() })
it.each(['close', 'identity', 'permission', 'repair permission', 'dispose'])('cancels diagnostics and ignores late results after %s', async (cause) => {
  const response = createDeferredAssociationResponse<MediaAssociationDiagnosisReport>()
  vi.mocked(requestMediaReferenceAssociationDiagnosis).mockReturnValue(response.promise)
  const state = createDisposableAssociationReview(); state.review.openMediaAssociationDiagnosis('1')
  const signal = vi.mocked(requestMediaReferenceAssociationDiagnosis).mock.calls[0]![1]
  if (cause === 'close') state.review.closeMediaAssociationDiagnosis()
  if (cause === 'identity') state.identity.value = 'actor2'
  if (cause === 'permission') state.allowed.value = false
  if (cause === 'repair permission') state.canRepair.value = false
  if (cause === 'dispose') state.scope.stop()
  expect(signal.aborted).toBe(true); response.resolve(report); await Promise.resolve()
  expect(state.review.report.value).toBeNull(); expect(state.review.open.value).toBe(false)
})
async function createSelectedGalleryAssociationReview() {
  vi.mocked(requestMediaReferenceAssociationDiagnosis).mockResolvedValue(report)
  const state = createDisposableAssociationReview(); state.review.openMediaAssociationDiagnosis('1')
  await vi.waitFor(() => expect(state.review.report.value).toBe(report))
  state.review.selectMissingGalleryImageForReviewedRepair(entry); state.review.destinationGalleryID.value = '5'
  return state
}
it('invalidates and cancels a preview when the destination changes', async () => {
  const state = await createSelectedGalleryAssociationReview(), response = createDeferredAssociationResponse<MissingGalleryParentRepairPreview>()
  vi.mocked(requestMissingGalleryParentRepairPreview).mockReturnValue(response.promise)
  const pending = state.review.previewMissingGalleryParentRepair()
  const signal = vi.mocked(requestMissingGalleryParentRepairPreview).mock.calls[0]![3]
  state.review.destinationGalleryID.value = '6'; expect(signal.aborted).toBe(true)
  response.resolve(preview); await pending; expect(state.review.preview.value).toBeNull()
})
it('blocks duplicate mutations and discards the preview after an uncertain response', async () => {
  const state = await createSelectedGalleryAssociationReview()
  vi.mocked(requestMissingGalleryParentRepairPreview).mockResolvedValue(preview)
  await state.review.previewMissingGalleryParentRepair(); state.review.reason.value = '确认图库归属'
  const response = createDeferredAssociationResponse<void>(); vi.mocked(applyMissingGalleryParentRepairWithReviewedFingerprint).mockReturnValue(response.promise)
  const pending = state.review.applyReviewedMissingGalleryParentRepair(); await state.review.applyReviewedMissingGalleryParentRepair()
  expect(applyMissingGalleryParentRepairWithReviewedFingerprint).toHaveBeenCalledTimes(1)
  response.reject(new Error('connection lost')); await pending
  expect(state.review.preview.value).toBeNull(); expect(state.review.error.value).toContain('结果未确认'); expect(state.repaired).not.toHaveBeenCalled()
})
it('refreshes the global list after confirmed repair and reinspects authoritative references', async () => {
  const state = await createSelectedGalleryAssociationReview()
  vi.mocked(requestMissingGalleryParentRepairPreview).mockResolvedValue(preview)
  vi.mocked(applyMissingGalleryParentRepairWithReviewedFingerprint).mockResolvedValue()
  await state.review.previewMissingGalleryParentRepair(); state.review.reason.value = '确认图库归属'
  await state.review.applyReviewedMissingGalleryParentRepair()
  expect(state.repaired).toHaveBeenCalledOnce(); expect(requestMediaReferenceAssociationDiagnosis).toHaveBeenCalledTimes(2); expect(state.review.preview.value).toBeNull()
})
