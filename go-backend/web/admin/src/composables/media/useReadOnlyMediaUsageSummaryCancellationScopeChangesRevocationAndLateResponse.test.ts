import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import { requestReadOnlyMediaUsageSummary, type MediaUsageSummaryParams, type MediaUsageSummaryReport } from '@/api/mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'
import { useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation } from './useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation'

vi.mock('@/api/mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates', () => ({ requestReadOnlyMediaUsageSummary: vi.fn() }))
const scopes: EffectScope[] = []
const report = { coverage: 'recognized_business_sources_only', rows: [] } as unknown as MediaUsageSummaryReport
function createDeferredMediaUsageSummaryResponse() {
  let resolve!: (value: MediaUsageSummaryReport) => void
  let reject!: (value: unknown) => void
  const promise = new Promise<MediaUsageSummaryReport>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}
function createDisposablePermissionScopedMediaUsageSummary() {
  const allowed = ref(true), params = ref<MediaUsageSummaryParams>({ library_partition: 'shared' }), identity = ref('actor:1,media:view'), revision = ref(0)
  const scope = effectScope(); scopes.push(scope)
  const state = scope.run(() => useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation(allowed, params, identity, revision))!
  return { allowed, params, identity, revision, scope, state }
}
beforeEach(() => vi.resetAllMocks())
afterEach(() => scopes.splice(0).forEach((scope) => scope.stop()))

it('waits for explicit open and scans the complete selected scope only once', async () => {
  const response = createDeferredMediaUsageSummaryResponse()
  vi.mocked(requestReadOnlyMediaUsageSummary).mockReturnValue(response.promise)
  const { state } = createDisposablePermissionScopedMediaUsageSummary()
  expect(requestReadOnlyMediaUsageSummary).not.toHaveBeenCalled()
  state.setMediaUsageSummaryDialogOpen(true)
  await state.rebuildMediaUsageSummary()
  expect(requestReadOnlyMediaUsageSummary).toHaveBeenCalledTimes(1)
  expect(vi.mocked(requestReadOnlyMediaUsageSummary).mock.calls[0]![0]).toEqual({ library_partition: 'shared' })
  response.resolve(report)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
})

it.each(['close', 'filter', 'identity', 'permission', 'refresh', 'dispose'])('clears and aborts on %s, ignoring late success and preventing automatic rescan', async (cause) => {
  const response = createDeferredMediaUsageSummaryResponse()
  vi.mocked(requestReadOnlyMediaUsageSummary).mockReturnValue(response.promise)
  const { state, params, identity, allowed, revision, scope } = createDisposablePermissionScopedMediaUsageSummary()
  state.setMediaUsageSummaryDialogOpen(true)
  const signal = vi.mocked(requestReadOnlyMediaUsageSummary).mock.calls[0]![1]!
  if (cause === 'close') state.setMediaUsageSummaryDialogOpen(false)
  if (cause === 'filter') params.value.library_partition = 'catalog'
  if (cause === 'identity') identity.value = 'actor:2,media:view'
  if (cause === 'permission') allowed.value = false
  if (cause === 'refresh') revision.value++
  if (cause === 'dispose') scope.stop()
  expect(signal.aborted).toBe(true)
  response.resolve(report)
  await Promise.resolve()
  expect(state.report.value).toBeNull()
  expect(state.open.value).toBe(false)
  expect(state.loading.value).toBe(false)
  expect(requestReadOnlyMediaUsageSummary).toHaveBeenCalledTimes(1)
})

it('ignores an old error after reopening and clears an old report before a failed rebuild', async () => {
  const first = createDeferredMediaUsageSummaryResponse(), second = createDeferredMediaUsageSummaryResponse()
  vi.mocked(requestReadOnlyMediaUsageSummary).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise).mockRejectedValueOnce(new Error('offline'))
  const { state } = createDisposablePermissionScopedMediaUsageSummary()
  state.setMediaUsageSummaryDialogOpen(true)
  state.setMediaUsageSummaryDialogOpen(false)
  state.setMediaUsageSummaryDialogOpen(true)
  first.reject(new Error('old error'))
  await Promise.resolve()
  expect(state.error.value).toBe('')
  second.resolve(report)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
  await state.rebuildMediaUsageSummary()
  expect(state.report.value).toBeNull()
  expect(state.error.value).toContain('未生成结论')
})

it('clears a completed report on business permission changes even when media:view remains', async () => {
  vi.mocked(requestReadOnlyMediaUsageSummary).mockResolvedValue(report)
  const { state, identity } = createDisposablePermissionScopedMediaUsageSummary()
  state.setMediaUsageSummaryDialogOpen(true)
  await vi.waitFor(() => expect(state.report.value).toEqual(report))
  identity.value = 'actor:1,media:view,product:view'
  expect(state.report.value).toBeNull()
  expect(state.open.value).toBe(false)
})
