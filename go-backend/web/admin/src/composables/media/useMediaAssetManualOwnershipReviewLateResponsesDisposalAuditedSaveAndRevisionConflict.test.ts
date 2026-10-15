import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import mediaApi, { type MediaAssetLibraryOwnershipReview } from '@/api/media'
import { useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave } from './useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave'

vi.mock('@/api/media', () => ({ default: { reviewAssetLibraryOwnership: vi.fn(), reclassifyAssetLibraryOwnership: vi.fn() } }))
const scopes: EffectScope[] = []

function createDeferredManualMediaOwnershipReviewResponse() {
  let resolve!: (value: MediaAssetLibraryOwnershipReview) => void
  const promise = new Promise<MediaAssetLibraryOwnershipReview>((resolvePromise) => { resolve = resolvePromise })
  return { promise, resolve }
}

function createManualMediaOwnershipReviewResponse(assetID = 1, revision = 0): MediaAssetLibraryOwnershipReview {
  return { current: { media_asset_id: assetID, library_partition: 'unclassified', upload_origin: 'legacy_unknown', classification_method: 'historical_unknown', customer_service_media_purpose: '', classification_revision: revision },
    can_reclassify: true, allowed_choices: [{ library_partition: 'catalog', customer_service_media_purpose: '' }, { library_partition: 'customer_service', customer_service_media_purpose: 'reusable_reply' }],
    suggested_ownership: { library_partition: 'catalog', customer_service_media_purpose: '' }, suggestion_basis: 'single_recognized_business_category_requires_manual_confirmation',
    reference_coverage: 'recognized_business_sources_only', recognized_reference_total: 1, reference_category_counts: { catalog: 1 }, retained_deleted_parent_references: 0, missing_parent_references: 0 }
}

function createDisposableManualMediaOwnershipReview() {
  const scope = effectScope()
  scopes.push(scope)
  const onSaved = vi.fn()
  const state = scope.run(() => useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave(onSaved))!
  return { scope, onSaved, state }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockResolvedValue(createManualMediaOwnershipReviewResponse())
  vi.mocked(mediaApi.reclassifyAssetLibraryOwnership).mockResolvedValue({ id: 1, classification_revision: 1 })
})
afterEach(() => scopes.splice(0).forEach((scope) => scope.stop()))

describe('manual media ownership review keeps provenance separate and saves the reviewed revision', () => {
  it('does not claim to save another reason when already-reviewed ownership has no change', async () => {
    const reviewed = createManualMediaOwnershipReviewResponse()
    reviewed.current = { ...reviewed.current, library_partition: 'catalog', classification_method: 'manual_reviewed', classification_revision: 1 }
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockResolvedValueOnce(reviewed)
    const { state } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    state.reason.value = '再次核对相同的归属'
    expect(state.canSave.value).toBe(false)
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(mediaApi.reclassifyAssetLibraryOwnership).not.toHaveBeenCalled()
    state.selectedPartition.value = 'customer_service'
    expect(state.canSave.value).toBe(true)
  })
  it('requires a deliberate allowed choice and valid reason before saving and sends only mutable ownership fields', async () => {
    const { state, onSaved } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1, upload_origin: 'legacy_unknown' })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.selectedPartition.value).toBe('')
    expect(state.canSave.value).toBe(false)
    state.selectedPartition.value = 'catalog'
    state.reason.value = ' 短 '
    expect(state.canSave.value).toBe(false)
    state.reason.value = ' 已核对商品说明图 '
    expect(state.canSave.value).toBe(true)
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(mediaApi.reclassifyAssetLibraryOwnership).toHaveBeenCalledWith(1, { library_partition: 'catalog', customer_service_media_purpose: '', expected_revision: 0, reason: '已核对商品说明图' }, expect.any(AbortSignal))
    expect(onSaved).toHaveBeenCalledTimes(1)
    expect(state.open.value).toBe(false)
    expect(state.saving.value).toBe(false)
  })

  it('does not let a late response from another asset replace the current review', async () => {
    const old = createDeferredManualMediaOwnershipReviewResponse()
    const current = createDeferredManualMediaOwnershipReviewResponse()
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockReturnValueOnce(old.promise).mockReturnValueOnce(current.promise)
    const { state } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    const oldSignal = vi.mocked(mediaApi.reviewAssetLibraryOwnership).mock.calls[0]![1]!
    state.openMediaAssetManualOwnershipReview({ id: 2 })
    expect(oldSignal.aborted).toBe(true)
    old.resolve(createManualMediaOwnershipReviewResponse(1))
    await Promise.resolve()
    expect(state.review.value).toBeNull()
    expect(state.loading.value).toBe(true)
    current.resolve(createManualMediaOwnershipReviewResponse(2))
    await vi.waitFor(() => expect(state.review.value?.current.media_asset_id).toBe(2))
  })

  it('clears the review on closing and ignores completion after scope disposal', async () => {
    const old = createDeferredManualMediaOwnershipReviewResponse()
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockReturnValueOnce(old.promise)
    const { state, scope } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    const signal = vi.mocked(mediaApi.reviewAssetLibraryOwnership).mock.calls[0]![1]!
    state.setMediaAssetOwnershipReviewDialogOpen(false)
    expect(signal.aborted).toBe(true)
    old.resolve(createManualMediaOwnershipReviewResponse())
    await Promise.resolve()
    expect(state.review.value).toBeNull()
    const pending = createDeferredManualMediaOwnershipReviewResponse()
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockReturnValueOnce(pending.promise)
    state.openMediaAssetManualOwnershipReview({ id: 2 })
    scope.stop()
    pending.resolve(createManualMediaOwnershipReviewResponse(2))
    await Promise.resolve()
    expect(state.review.value).toBeNull()
  })

  it('requires another review after a conflict instead of silently retrying with a new version', async () => {
    vi.mocked(mediaApi.reclassifyAssetLibraryOwnership).mockRejectedValueOnce({ isAxiosError: true, response: { status: 409 } })
    const { state, onSaved } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    state.selectedPartition.value = 'catalog'; state.reason.value = '核对商品素材'
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(state.revisionConflict.value).toBe(true)
    expect(state.canSave.value).toBe(false)
    expect(state.error.value).toContain('其他人')
    expect(onSaved).not.toHaveBeenCalled()
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(mediaApi.reclassifyAssetLibraryOwnership).toHaveBeenCalledTimes(1)
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockResolvedValueOnce(createManualMediaOwnershipReviewResponse(1, 2))
    await state.loadCurrentMediaAssetOwnershipReview()
    expect(state.reason.value).toBe('')
    expect(state.review.value?.current.classification_revision).toBe(2)
    state.selectedPartition.value = 'customer_service'; state.reason.value = '重新核对客服常用素材'
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(mediaApi.reclassifyAssetLibraryOwnership).toHaveBeenLastCalledWith(1, expect.objectContaining({ expected_revision: 2, customer_service_media_purpose: 'reusable_reply' }), expect.any(AbortSignal))
  })

  it('prevents duplicate saves or switching the asset while its classification is being committed', async () => {
    let resolveSave!: (value: { id: number }) => void
    vi.mocked(mediaApi.reclassifyAssetLibraryOwnership).mockReturnValueOnce(new Promise((resolve) => { resolveSave = resolve }))
    const { state } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    state.selectedPartition.value = 'catalog'; state.reason.value = '核对商品素材'
    const saving = state.saveReviewedMediaAssetLibraryOwnership()
    state.setMediaAssetOwnershipReviewDialogOpen(false)
    state.openMediaAssetManualOwnershipReview({ id: 2 })
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(state.open.value).toBe(true); expect(state.asset.value?.id).toBe(1)
    expect(mediaApi.reclassifyAssetLibraryOwnership).toHaveBeenCalledTimes(1)
    resolveSave({ id: 1 }); await saving
    expect(state.open.value).toBe(false)
  })

  it('does not expose ownership choices or permit saving restricted assets', async () => {
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockResolvedValueOnce({ ...createManualMediaOwnershipReviewResponse(), can_reclassify: false, restriction_reason: 'conversation_attachments_require_conversation_lifecycle', allowed_choices: [] })
    const { state } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    state.selectedPartition.value = 'catalog'; state.reason.value = '不能整理会话附件'
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(mediaApi.reclassifyAssetLibraryOwnership).not.toHaveBeenCalled()
  })

  it('blocks automatic resubmission after an ambiguous network failure and clears failed reads for retry', async () => {
    vi.mocked(mediaApi.reviewAssetLibraryOwnership).mockRejectedValueOnce(new Error('network unavailable'))
    const { state } = createDisposableManualMediaOwnershipReview()
    state.openMediaAssetManualOwnershipReview({ id: 1 })
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.review.value).toBeNull(); expect(state.error.value).toContain('重试')
    await state.loadCurrentMediaAssetOwnershipReview()
    state.selectedPartition.value = 'catalog'; state.reason.value = '核对商品素材'
    vi.mocked(mediaApi.reclassifyAssetLibraryOwnership).mockRejectedValueOnce(new Error('response lost'))
    await state.saveReviewedMediaAssetLibraryOwnership()
    expect(state.canSave.value).toBe(false); expect(state.error.value).toContain('核对归属')
  })
})
