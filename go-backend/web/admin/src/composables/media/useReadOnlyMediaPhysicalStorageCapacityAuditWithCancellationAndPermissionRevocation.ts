import { onScopeDispose, ref, shallowRef, watch, type Ref } from 'vue'
import { isAxiosError } from 'axios'
import { requestReadOnlyMediaPhysicalStorageCapacityAudit, type MediaPhysicalStorageCapacityAuditReport } from '@/api/mediaPhysicalStorageCapacityReadOnlyAudit'

export function useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation(allowed: Readonly<Ref<boolean>>) {
  const open = ref(false)
  const loading = ref(false)
  const error = ref('')
  const report = shallowRef<MediaPhysicalStorageCapacityAuditReport | null>(null)
  let generation = 0
  let controller: AbortController | null = null
  let disposed = false

  function clearAndCancelPhysicalStorageCapacityAudit() {
    generation++
    controller?.abort()
    controller = null
    loading.value = false
    report.value = null
    error.value = ''
  }
  async function runPhysicalStorageCapacityAudit() {
    if (disposed || !allowed.value || !open.value || loading.value) return
    clearAndCancelPhysicalStorageCapacityAudit()
    const currentGeneration = generation
    controller = new AbortController()
    loading.value = true
    try {
      const result = await requestReadOnlyMediaPhysicalStorageCapacityAudit(controller.signal)
      if (currentGeneration === generation && allowed.value && !disposed) report.value = result
    } catch (failure) {
      if (currentGeneration !== generation || disposed || !allowed.value) return
      const code = isAxiosError(failure) ? failure.response?.data?.code : ''
      error.value = code === 'media_capacity_audit_running' ? '已有容量核对正在进行，请稍后重试' :
        code === 'media_capacity_snapshot_limit_exceeded' ? '登记记录超过本次核对上限，未生成容量结论' : '容量核对失败，请稍后重试'
    } finally {
      if (currentGeneration === generation) { loading.value = false; controller = null }
    }
  }
  function setPhysicalStorageCapacityDialogOpen(value: boolean) {
    if (disposed || (value && !allowed.value)) return
    if (open.value === value) return
    clearAndCancelPhysicalStorageCapacityAudit()
    open.value = value
    if (value) void runPhysicalStorageCapacityAudit()
  }
  watch(allowed, (value) => { if (!value) { open.value = false; clearAndCancelPhysicalStorageCapacityAudit() } }, { flush: 'sync' })
  onScopeDispose(() => { disposed = true; open.value = false; clearAndCancelPhysicalStorageCapacityAudit() })
  return { open, loading, error, report, setPhysicalStorageCapacityDialogOpen, runPhysicalStorageCapacityAudit }
}
