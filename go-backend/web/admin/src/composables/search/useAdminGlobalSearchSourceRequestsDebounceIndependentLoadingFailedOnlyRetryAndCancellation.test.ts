import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref, shallowRef, type EffectScope } from 'vue'
import type { AdminGlobalSearchResult, AdminGlobalSearchSource } from '@/modules/search/adminGlobalSearchResultAndSourceContracts'
import { useAdminGlobalSearchSourceRequests } from './useAdminGlobalSearchSourceRequests'

const activeSourceRequestTestScopes: EffectScope[] = []

beforeEach(() => vi.useFakeTimers())
afterEach(() => {
  activeSourceRequestTestScopes.splice(0).forEach((scope) => scope.stop())
  vi.useRealTimers()
})

function createAdminGlobalSearchSourceRequestTestResult(id: string, sourceId = 'pages'): AdminGlobalSearchResult {
  return { id, sourceId, title: id, description: id, categoryLabel: 'Group', linkLabel: `/${id}`, routeTarget: `/${id}` }
}

function createAdminGlobalSearchSourceRequestTestSource(overrides: Partial<AdminGlobalSearchSource> = {}): AdminGlobalSearchSource {
  return {
    id: 'pages', label: 'Pages', recommendationsLabel: 'Common pages',
    getRecommendations: () => [createAdminGlobalSearchSourceRequestTestResult('common')],
    search: (query) => [createAdminGlobalSearchSourceRequestTestResult(query)],
    ...overrides,
  }
}

function createDeferredAdminGlobalSearchSourceRequestTestResults() {
  let resolve!: (results: AdminGlobalSearchResult[]) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<AdminGlobalSearchResult[]>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}

function createDisposableAdminGlobalSearchSourceRequests(initialSources: AdminGlobalSearchSource[], initialQuery = '') {
  const query = ref(initialQuery)
  const sources = shallowRef(initialSources)
  const scope = effectScope()
  activeSourceRequestTestScopes.push(scope)
  const state = scope.run(() => useAdminGlobalSearchSourceRequests({ query, sources }))!
  return { scope, query, sources, state }
}

describe('independent global admin search source requests', () => {
  it('runs local queries immediately while debouncing only the last remote query', async () => {
    const localSearch = vi.fn((query: string) => [createAdminGlobalSearchSourceRequestTestResult(query)])
    const remoteSearch = vi.fn((query: string) => [createAdminGlobalSearchSourceRequestTestResult(query, 'orders')])
    const { query, state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ search: localSearch }),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'orders', label: 'Orders', queryDebounceMs: 250, search: remoteSearch }),
    ])
    await vi.advanceTimersByTimeAsync(0)
    query.value = 'o'
    await vi.advanceTimersByTimeAsync(100)
    query.value = 'or'
    await vi.advanceTimersByTimeAsync(100)
    query.value = ' ＯＲＤＥＲ '
    await vi.advanceTimersByTimeAsync(249)

    expect(localSearch.mock.calls.map(([value]) => value)).toEqual(['o', 'or', 'ORDER'])
    expect(remoteSearch).not.toHaveBeenCalled()
    expect(state.resultGroups.value.map((group) => group.loading)).toEqual([false, true])
    expect(state.resultGroups.value[0].results[0].id).toBe('ORDER')
    await vi.advanceTimersByTimeAsync(1)
    expect(remoteSearch).toHaveBeenCalledTimes(1)
    expect(remoteSearch.mock.calls[0][0]).toBe('ORDER')
    expect(state.resultGroups.value[1].results[0].id).toBe('ORDER')
    expect(state.loading.value).toBe(false)
  })

  it('loads recommendations immediately and cancels queued queries when the input is cleared', async () => {
    const getRecommendations = vi.fn(() => [createAdminGlobalSearchSourceRequestTestResult('recommended')])
    const search = vi.fn(() => [])
    const { query, state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ queryDebounceMs: 250, getRecommendations, search }),
    ])
    expect(getRecommendations).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(0)
    expect(state.loading.value).toBe(false)
    query.value = 'queued'
    await vi.advanceTimersByTimeAsync(100)
    query.value = '\n　 '
    expect(getRecommendations).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(500)
    expect(search).not.toHaveBeenCalled()
    expect(state.resultGroups.value[0]).toMatchObject({ label: 'Common pages', loading: false, failed: false })
    expect(state.resultGroups.value[0].results[0].id).toBe('recommended')
  })

  it('cancels queued requests on source replacement and immediately removes revoked sources', async () => {
    const revokedSearch = vi.fn(() => [])
    const allowedSearch = vi.fn(() => [createAdminGlobalSearchSourceRequestTestResult('allowed')])
    const { sources, state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ queryDebounceMs: 250, search: revokedSearch }),
    ], 'protected')
    await vi.advanceTimersByTimeAsync(100)
    sources.value = [createAdminGlobalSearchSourceRequestTestSource({ id: 'allowed', label: 'Allowed pages', search: allowedSearch })]
    expect(allowedSearch).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(500)
    expect(revokedSearch).not.toHaveBeenCalled()
    expect(state.resultGroups.value[0].sourceId).toBe('allowed')
    sources.value = [createAdminGlobalSearchSourceRequestTestSource({ queryDebounceMs: 250, search: revokedSearch })]
    sources.value = []
    expect(state.resultGroups.value).toEqual([])
    expect(state.loading.value).toBe(false)
    await vi.advanceTimersByTimeAsync(500)
    expect(revokedSearch).not.toHaveBeenCalled()
  })

  it('aborts active requests when a new query arrives and ignores late success and rejection', async () => {
    const oldSuccess = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const oldFailure = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const search = vi.fn()
      .mockReturnValueOnce(oldSuccess.promise)
      .mockReturnValueOnce(oldFailure.promise)
      .mockReturnValue([createAdminGlobalSearchSourceRequestTestResult('latest')])
    const { query, state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ search }),
    ], 'first')
    query.value = 'second'
    expect(search.mock.calls[0][1].signal.aborted).toBe(true)
    query.value = 'latest'
    expect(search.mock.calls[1][1].signal.aborted).toBe(true)
    oldSuccess.resolve([createAdminGlobalSearchSourceRequestTestResult('stale')])
    oldFailure.reject(new Error('stale failure'))
    await vi.advanceTimersByTimeAsync(0)
    expect(state.resultGroups.value[0].results[0].id).toBe('latest')
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.loading.value).toBe(false)
  })

  it('finishes each source independently and isolates synchronous failures', async () => {
    const pending = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const { state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource(),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'orders', label: 'Orders', search: () => { throw new Error('offline') } }),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'products', label: 'Products', search: () => pending.promise }),
    ], 'current')
    await vi.advanceTimersByTimeAsync(0)
    expect(state.resultGroups.value.map((group) => group.loading)).toEqual([false, false, true])
    expect(state.failedSourceLabels.value).toEqual(['Orders'])
    expect(state.resultGroups.value[0].results[0].id).toBe('current')
    expect(state.loading.value).toBe(true)
    pending.resolve([])
    await vi.advanceTimersByTimeAsync(0)
    expect(state.loading.value).toBe(false)
  })

  it('retries failed sources immediately without resetting completed results or other active requests', async () => {
    const pendingOtherSource = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const pendingRetry = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const fastSearch = vi.fn(() => [createAdminGlobalSearchSourceRequestTestResult('retained')])
    const failedSearch = vi.fn().mockRejectedValueOnce(new Error('offline')).mockReturnValue(pendingRetry.promise)
    const otherSearch = vi.fn<AdminGlobalSearchSource['search']>(() => pendingOtherSource.promise)
    const { state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ search: fastSearch }),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'orders', label: 'Orders', queryDebounceMs: 250, search: failedSearch }),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'products', label: 'Products', search: otherSearch }),
    ], 'current')
    await vi.advanceTimersByTimeAsync(250)
    expect(state.failedSourceLabels.value).toEqual(['Orders'])
    const completedGroup = state.resultGroups.value[0]
    const otherSignal = otherSearch.mock.calls[0][1].signal
    const generation = state.searchRequestGeneration.value
    const retry = state.retryFailedAdminGlobalSearchSources()
    expect(failedSearch).toHaveBeenCalledTimes(2)
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.resultGroups.value.map((group) => group.loading)).toEqual([false, true, true])
    expect(state.resultGroups.value[0]).toBe(completedGroup)
    await state.retryFailedAdminGlobalSearchSources()
    expect(failedSearch).toHaveBeenCalledTimes(2)
    expect(fastSearch).toHaveBeenCalledTimes(1)
    expect(otherSearch).toHaveBeenCalledTimes(1)
    expect(otherSignal.aborted).toBe(false)
    expect(state.searchRequestGeneration.value).toBe(generation)

    pendingOtherSource.resolve([createAdminGlobalSearchSourceRequestTestResult('other-completed', 'products')])
    await vi.advanceTimersByTimeAsync(0)
    expect(state.resultGroups.value.map((group) => group.loading)).toEqual([false, true, false])
    pendingRetry.resolve([createAdminGlobalSearchSourceRequestTestResult('retried', 'orders')])
    await retry
    expect(state.resultGroups.value.flatMap((group) => group.results.map((result) => result.id))).toEqual(['retained', 'retried', 'other-completed'])
    expect(state.loading.value).toBe(false)
  })

  it.each([
    { change: 'query', completion: 'success' },
    { change: 'query', completion: 'failure' },
    { change: 'permissions', completion: 'success' },
    { change: 'permissions', completion: 'failure' },
  ])('ignores a stale retry $completion after $change change', async ({ change, completion }) => {
    const staleRetry = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const search = vi.fn()
      .mockRejectedValueOnce(new Error('offline'))
      .mockReturnValueOnce(staleRetry.promise)
      .mockReturnValue([createAdminGlobalSearchSourceRequestTestResult('latest')])
    const { query, sources, state } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ search }),
    ], 'original')
    await vi.advanceTimersByTimeAsync(0)
    const retry = state.retryFailedAdminGlobalSearchSources()
    const retrySignal = search.mock.calls[1][1].signal
    if (change === 'query') query.value = 'latest'
    else sources.value = [createAdminGlobalSearchSourceRequestTestSource({ id: 'allowed', search: () => [createAdminGlobalSearchSourceRequestTestResult('latest')] })]
    expect(retrySignal.aborted).toBe(true)
    if (completion === 'success') staleRetry.resolve([createAdminGlobalSearchSourceRequestTestResult('revoked')])
    else staleRetry.reject(new Error('late failure'))
    await retry
    await vi.advanceTimersByTimeAsync(0)
    expect(state.resultGroups.value[0].results[0].id).toBe('latest')
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.loading.value).toBe(false)
  })

  it('cancels timers and active requests on disposal and keeps closed search empty', async () => {
    const pending = createDeferredAdminGlobalSearchSourceRequestTestResults()
    const activeSearch = vi.fn<AdminGlobalSearchSource['search']>(() => pending.promise)
    const delayedSearch = vi.fn(() => [])
    const { query, sources, state, scope } = createDisposableAdminGlobalSearchSourceRequests([
      createAdminGlobalSearchSourceRequestTestSource({ search: activeSearch }),
      createAdminGlobalSearchSourceRequestTestSource({ id: 'orders', queryDebounceMs: 250, search: delayedSearch }),
    ], 'current')
    scope.stop()
    expect(activeSearch.mock.calls[0][1].signal.aborted).toBe(true)
    pending.reject(new Error('late failure'))
    query.value = 'after-close'
    sources.value = [createAdminGlobalSearchSourceRequestTestSource({ search: delayedSearch })]
    await state.refreshAdminGlobalSearchResults()
    await state.retryFailedAdminGlobalSearchSources()
    await vi.advanceTimersByTimeAsync(500)
    expect(delayedSearch).not.toHaveBeenCalled()
    expect(activeSearch).toHaveBeenCalledTimes(1)
    expect(state.resultGroups.value).toEqual([])
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.loading.value).toBe(false)
    expect(vi.getTimerCount()).toBe(0)
  })

  it('keeps loading false when no sources are available', async () => {
    const { query, state } = createDisposableAdminGlobalSearchSourceRequests([])
    expect(state.loading.value).toBe(false)
    query.value = 'query'
    await state.refreshAdminGlobalSearchResults()
    await state.retryFailedAdminGlobalSearchSources()
    expect(state.resultGroups.value).toEqual([])
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.loading.value).toBe(false)
  })
})
