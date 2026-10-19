import { onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue'
import { applyMissingGalleryParentRepairWithReviewedFingerprint, isExactPositiveMediaAssociationIdentity, requestMediaReferenceAssociationDiagnosis, requestMissingGalleryParentRepairPreview,
  type MediaAssociationDiagnosisEntry, type MediaAssociationDiagnosisReport, type MissingGalleryParentRepairPreview } from '@/api/mediaReferenceAssociationDiagnosisAndAuditedGalleryParentRepair'

export function useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation(
  allowed: Readonly<Ref<boolean>>, canRepair: Readonly<Ref<boolean>>, authorizationAndListScope: Readonly<Ref<string>>, onRepaired: () => void,
) {
  const open = ref(false), assetID = ref(''), loading = ref(false), saving = ref(false), error = ref(''), notice = ref('')
  const report = shallowRef<MediaAssociationDiagnosisReport | null>(null), preview = shallowRef<MissingGalleryParentRepairPreview | null>(null)
  const selectedImageID = ref(''), destinationGalleryID = ref(''), reason = ref('')
  let generation = 0, disposed = false, controller: AbortController | null = null
  function cancelAndClearMediaAssociationReview() {
    generation++; controller?.abort(); controller = null
    loading.value = false; saving.value = false; report.value = null; preview.value = null; error.value = ''; notice.value = ''
    selectedImageID.value = ''; destinationGalleryID.value = ''; reason.value = ''
  }
  function closeMediaAssociationDiagnosis() { open.value = false; cancelAndClearMediaAssociationReview(); assetID.value = '' }
  async function loadMediaAssociationDiagnosis() {
    if (disposed || !open.value || !allowed.value || loading.value || saving.value) return
    generation++; controller?.abort(); controller = new AbortController()
    const current = generation, signal = controller.signal
    report.value = null; preview.value = null; selectedImageID.value = ''; error.value = ''; loading.value = true
    try {
      const result = await requestMediaReferenceAssociationDiagnosis(assetID.value, signal)
      if (current === generation && !disposed) report.value = result
    } catch { if (current === generation && !disposed) error.value = '诊断未完成，请重新核查' }
    finally { if (current === generation) { loading.value = false; controller = null } }
  }
  function openMediaAssociationDiagnosis(id: string) {
    if (disposed || !allowed.value || !isExactPositiveMediaAssociationIdentity(id)) return
    cancelAndClearMediaAssociationReview(); assetID.value = id; open.value = true; void loadMediaAssociationDiagnosis()
  }
  function selectMissingGalleryImageForReviewedRepair(entry: MediaAssociationDiagnosisEntry) {
    if (!canRepair.value || loading.value || saving.value || report.value?.has_restricted_references || !report.value?.entries.includes(entry) ||
      entry.resource_type !== 'gallery_image' || entry.usage_state !== 'missing_parent') return
    selectedImageID.value = entry.resource_id; destinationGalleryID.value = ''; preview.value = null; reason.value = ''; error.value = ''; notice.value = ''
  }
  async function previewMissingGalleryParentRepair() {
    if (!canRepair.value || !open.value || loading.value || saving.value || !selectedImageID.value || !isExactPositiveMediaAssociationIdentity(destinationGalleryID.value)) return
    generation++; controller?.abort(); controller = new AbortController()
    const current = generation
    loading.value = true; preview.value = null; error.value = ''
    try {
      const result = await requestMissingGalleryParentRepairPreview(assetID.value, selectedImageID.value, destinationGalleryID.value, controller.signal)
      if (current === generation && !disposed) preview.value = result
    } catch { if (current === generation && !disposed) error.value = '无法生成修复预览。请核对草稿图库、原图绑定及保留状态后重新诊断' }
    finally { if (current === generation) { loading.value = false; controller = null } }
  }
  async function applyReviewedMissingGalleryParentRepair() {
    const reviewed = preview.value, explanation = reason.value.trim()
    if (!canRepair.value || !open.value || loading.value || saving.value || !reviewed || [...explanation].length < 3 || [...explanation].length > 1000) return
    generation++; controller?.abort(); controller = new AbortController()
    const current = generation
    saving.value = true; error.value = ''; notice.value = ''; preview.value = null
    try {
      await applyMissingGalleryParentRepairWithReviewedFingerprint(reviewed, explanation, controller.signal)
      // Refresh global lists even if a close/filter change hid the response.
      if (!disposed) onRepaired()
      if (current === generation && !disposed) {
        saving.value = false
        notice.value = '已迁入草稿图库并记录审计；请在图库中继续审核'
        await loadMediaAssociationDiagnosis()
      }
    } catch { if (current === generation && !disposed) error.value = '修复结果未确认，请重新诊断并核对审计记录，勿直接重复提交' }
    finally { if (current === generation) { saving.value = false; controller = null } }
  }
  watch(destinationGalleryID, () => {
    preview.value = null
    if (loading.value && selectedImageID.value) { generation++; controller?.abort(); controller = null; loading.value = false }
  }, { flush: 'sync' })
  watch([allowed, canRepair, authorizationAndListScope], closeMediaAssociationDiagnosis, { flush: 'sync' })
  onScopeDispose(() => { disposed = true; closeMediaAssociationDiagnosis() })
  return { open, assetID, loading, saving, error, notice, report, preview, selectedImageID, destinationGalleryID, reason,
    closeMediaAssociationDiagnosis, openMediaAssociationDiagnosis, loadMediaAssociationDiagnosis, selectMissingGalleryImageForReviewedRepair, previewMissingGalleryParentRepair, applyReviewedMissingGalleryParentRepair }
}
