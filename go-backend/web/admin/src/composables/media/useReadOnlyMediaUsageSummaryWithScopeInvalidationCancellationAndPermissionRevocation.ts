import { computed, onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue'
import { isAxiosError } from 'axios'
import { requestReadOnlyMediaUsageSummary, type MediaUsageSummaryParams, type MediaUsageSummaryReport } from '@/api/mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'

export function useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation(
  allowed: Readonly<Ref<boolean>>, params: Readonly<Ref<MediaUsageSummaryParams>>, authorizationIdentity: Readonly<Ref<string>>, invalidationRevision: Readonly<Ref<number>>,
) {
  const open = ref(false)
  const loading = ref(false)
  const error = ref('')
  const report = shallowRef<MediaUsageSummaryReport | null>(null)
  const scopeIdentity = computed(() => JSON.stringify([params.value, authorizationIdentity.value, invalidationRevision.value]))
  let generation = 0
  let controller: AbortController | null = null
  let disposed = false

  function clearAndCancelMediaUsageSummary() {
    generation++
    controller?.abort()
    controller = null
    loading.value = false
    report.value = null
    error.value = ''
  }
  async function rebuildMediaUsageSummary() {
    if (disposed || !allowed.value || !open.value || loading.value) return
    clearAndCancelMediaUsageSummary()
    const currentGeneration = generation
    const currentScope = scopeIdentity.value
    const currentController = new AbortController()
    controller = currentController
    loading.value = true
    try {
      const result = await requestReadOnlyMediaUsageSummary({ ...params.value }, currentController.signal)
      if (!disposed && allowed.value && currentGeneration === generation && currentScope === scopeIdentity.value) report.value = result
    } catch (failure) {
      if (disposed || !allowed.value || currentGeneration !== generation) return
      const code = isAxiosError(failure) ? failure.response?.data?.code : ''
      error.value = code === 'media_usage_summary_limit_exceeded' ? '超过单次核对上限，请关闭后缩小筛选范围（最多 200 个资源、20,000 条引用）' :
        code === 'media_usage_summary_running' ? '已有使用汇总正在进行，请稍后重试' : '使用汇总未完成，未生成结论，请稍后重试'
    } finally {
      if (currentGeneration === generation) { loading.value = false; controller = null }
    }
  }
  function setMediaUsageSummaryDialogOpen(value: boolean) {
    if (disposed || (value && !allowed.value) || value === open.value) return
    clearAndCancelMediaUsageSummary()
    open.value = value
    if (value) void rebuildMediaUsageSummary()
  }
  watch([allowed, scopeIdentity], () => { open.value = false; clearAndCancelMediaUsageSummary() }, { flush: 'sync' })
  onScopeDispose(() => { disposed = true; open.value = false; clearAndCancelMediaUsageSummary() })
  return { open, loading, error, report, setMediaUsageSummaryDialogOpen, rebuildMediaUsageSummary }
}
