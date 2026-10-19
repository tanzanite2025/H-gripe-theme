import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import { requestReadOnlyMediaPhysicalStorageCapacityAudit, type MediaPhysicalStorageCapacityAuditReport } from '@/api/mediaPhysicalStorageCapacityReadOnlyAudit'
import { useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation } from './useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation'

vi.mock('@/api/mediaPhysicalStorageCapacityReadOnlyAudit', () => ({ requestReadOnlyMediaPhysicalStorageCapacityAudit: vi.fn() }))
const scopes: EffectScope[] = []
function createDeferredPhysicalCapacityAuditResponse() {
  let resolve!: (value: MediaPhysicalStorageCapacityAuditReport) => void
  let reject!: (value: unknown) => void
  const promise = new Promise<MediaPhysicalStorageCapacityAuditReport>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}
function createDisposablePhysicalCapacityAudit() {
  const allowed = ref(true)
  const scope = effectScope(); scopes.push(scope)
  const state = scope.run(() => useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation(allowed))!
  return { state, allowed, scope }
}
const report = { physical_scan_complete: true } as MediaPhysicalStorageCapacityAuditReport
beforeEach(() => vi.resetAllMocks())
afterEach(() => scopes.splice(0).forEach((scope) => scope.stop()))

it('starts only after an explicit open and prevents duplicate scans', async () => {
  const response = createDeferredPhysicalCapacityAuditResponse()
  vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mockReturnValue(response.promise)
  const { state } = createDisposablePhysicalCapacityAudit()
  expect(requestReadOnlyMediaPhysicalStorageCapacityAudit).not.toHaveBeenCalled()
  state.setPhysicalStorageCapacityDialogOpen(true)
  await state.runPhysicalStorageCapacityAudit()
  expect(requestReadOnlyMediaPhysicalStorageCapacityAudit).toHaveBeenCalledTimes(1)
  response.resolve(report)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
  expect(state.loading.value).toBe(false)
})

it('aborts on close and ignores an old completion after reopening', async () => {
  const first = createDeferredPhysicalCapacityAuditResponse(), second = createDeferredPhysicalCapacityAuditResponse()
  vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)
  const { state } = createDisposablePhysicalCapacityAudit()
  state.setPhysicalStorageCapacityDialogOpen(true)
  const oldSignal = vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mock.calls[0]![0]
  state.setPhysicalStorageCapacityDialogOpen(false)
  expect(oldSignal.aborted).toBe(true)
  state.setPhysicalStorageCapacityDialogOpen(true)
  first.resolve(report)
  await Promise.resolve()
  expect(state.report.value).toBeNull()
  expect(state.loading.value).toBe(true)
  second.resolve(report)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
})

it('clears successful reports when permission is revoked and prevents new requests', async () => {
  vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mockResolvedValue(report)
  const { state, allowed } = createDisposablePhysicalCapacityAudit()
  state.setPhysicalStorageCapacityDialogOpen(true)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
  allowed.value = false
  expect(state.open.value).toBe(false)
  expect(state.report.value).toBeNull()
  state.setPhysicalStorageCapacityDialogOpen(true)
  await state.runPhysicalStorageCapacityAudit()
  expect(requestReadOnlyMediaPhysicalStorageCapacityAudit).toHaveBeenCalledTimes(1)
})

it.each(['permission', 'dispose'])('aborts an in-flight request on %s and ignores late errors', async (cause) => {
  const response = createDeferredPhysicalCapacityAuditResponse()
  vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mockReturnValue(response.promise)
  const { state, allowed, scope } = createDisposablePhysicalCapacityAudit()
  state.setPhysicalStorageCapacityDialogOpen(true)
  const signal = vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mock.calls[0]![0]
  if (cause === 'permission') allowed.value = false
  else scope.stop()
  expect(signal.aborted).toBe(true)
  response.reject(new Error('late error'))
  await Promise.resolve()
  expect(state.error.value).toBe('')
  expect(state.report.value).toBeNull()
  expect(state.loading.value).toBe(false)
})

it('preserves the incomplete status and replaces an old report with a retry error', async () => {
  const incomplete = { ...report, physical_scan_complete: false }
  vi.mocked(requestReadOnlyMediaPhysicalStorageCapacityAudit).mockResolvedValueOnce(incomplete).mockRejectedValueOnce(new Error('offline'))
  const { state } = createDisposablePhysicalCapacityAudit()
  state.setPhysicalStorageCapacityDialogOpen(true)
  await vi.waitFor(() => expect(state.report.value?.physical_scan_complete).toBe(false))
  await state.runPhysicalStorageCapacityAudit()
  expect(state.report.value).toBeNull()
  expect(state.error.value).toContain('失败')
})
