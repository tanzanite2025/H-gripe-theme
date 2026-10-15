import { computed, onScopeDispose, ref } from 'vue'
import { isAxiosError } from 'axios'
import mediaApi, { type MediaAsset, type MediaAssetLibraryOwnershipReview, type MediaLibraryPartition } from '@/api/media'

export function useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave(onOwnershipSaved: () => void | Promise<void>) {
  const open = ref(false)
  const asset = ref<MediaAsset | null>(null)
  const review = ref<MediaAssetLibraryOwnershipReview | null>(null)
  const selectedPartition = ref<MediaLibraryPartition | ''>('')
  const reason = ref('')
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const revisionConflict = ref(false)
  let requestGeneration = 0
  let requestController: AbortController | null = null
  let disposed = false

  const canSave = computed(() => Boolean(review.value?.can_reclassify && !loading.value && !saving.value && !revisionConflict.value
    && review.value.allowed_choices.some((choice) => choice.library_partition === selectedPartition.value)
    && (review.value.current.classification_method !== 'manual_reviewed' || review.value.current.library_partition !== selectedPartition.value)
    && [...reason.value.trim()].length >= 3 && [...reason.value.trim()].length <= 1000))

  function cancelPreviousMediaAssetOwnershipReviewRequest(): number {
    requestController?.abort()
    requestController = null
    return ++requestGeneration
  }

  async function loadCurrentMediaAssetOwnershipReview(): Promise<void> {
    if (disposed || !open.value || !asset.value || saving.value) return
    const currentAssetID = asset.value.id
    const generation = cancelPreviousMediaAssetOwnershipReviewRequest()
    const controller = new AbortController()
    requestController = controller
    review.value = null
    selectedPartition.value = ''
    reason.value = ''
    error.value = ''
    revisionConflict.value = false
    loading.value = true
    try {
      const result = await mediaApi.reviewAssetLibraryOwnership(currentAssetID, controller.signal)
      if (disposed || generation !== requestGeneration || !open.value) return
      review.value = result
      if (result.allowed_choices.some((choice) => choice.library_partition === result.current.library_partition)) selectedPartition.value = result.current.library_partition
    } catch {
      if (!disposed && generation === requestGeneration && !controller.signal.aborted) error.value = '无法加载归属审核，请重试。'
    } finally {
      if (!disposed && generation === requestGeneration) { loading.value = false; requestController = null }
    }
  }

  function openMediaAssetManualOwnershipReview(selectedAsset: MediaAsset): void {
    if (disposed || saving.value) return
    asset.value = selectedAsset
    open.value = true
    void loadCurrentMediaAssetOwnershipReview()
  }

  function setMediaAssetOwnershipReviewDialogOpen(value: boolean): void {
    if (saving.value) return
    open.value = value
    if (!value) {
      cancelPreviousMediaAssetOwnershipReviewRequest()
      asset.value = null
      review.value = null
      error.value = ''
      loading.value = false
    }
  }

  async function saveReviewedMediaAssetLibraryOwnership(): Promise<void> {
    if (disposed || !canSave.value || !review.value || !asset.value) return
    const choice = review.value.allowed_choices.find((item) => item.library_partition === selectedPartition.value)
    if (!choice) return
    const generation = cancelPreviousMediaAssetOwnershipReviewRequest()
    const controller = new AbortController()
    requestController = controller
    saving.value = true
    error.value = ''
    try {
      await mediaApi.reclassifyAssetLibraryOwnership(asset.value.id, { ...choice, expected_revision: review.value.current.classification_revision, reason: reason.value.trim() }, controller.signal)
      if (disposed || generation !== requestGeneration) return
      saving.value = false
      setMediaAssetOwnershipReviewDialogOpen(false)
      await onOwnershipSaved()
    } catch (failure) {
      if (disposed || generation !== requestGeneration || controller.signal.aborted) return
      if (isAxiosError(failure) && failure.response?.status === 409) {
        revisionConflict.value = true
        error.value = '归属已被其他人修改，请重新加载并审核后保存。'
      } else {
        error.value = '未能确认保存结果，请重新加载核对归属后再操作。'
        revisionConflict.value = true
      }
    } finally {
      if (!disposed && generation === requestGeneration) { saving.value = false; requestController = null }
    }
  }

  onScopeDispose(() => { disposed = true; cancelPreviousMediaAssetOwnershipReviewRequest() })
  return { open, asset, review, selectedPartition, reason, loading, saving, error, revisionConflict, canSave,
    openMediaAssetManualOwnershipReview, setMediaAssetOwnershipReviewDialogOpen,
    loadCurrentMediaAssetOwnershipReview, saveReviewedMediaAssetLibraryOwnership }
}
