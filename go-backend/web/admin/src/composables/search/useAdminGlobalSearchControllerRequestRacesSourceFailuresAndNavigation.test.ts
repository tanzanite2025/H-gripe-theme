import { afterEach, describe, expect, it, vi } from 'vitest'
import { effectScope, shallowRef, type EffectScope } from 'vue'
import type { AdminGlobalSearchResult, AdminGlobalSearchSource } from '@/modules/search/adminGlobalSearchResultAndSourceContracts'
import { useAdminGlobalSearchController } from './useAdminGlobalSearchController'

const activeSearchTestScopes: EffectScope[] = []
afterEach(() => activeSearchTestScopes.splice(0).forEach((scope) => scope.stop()))

function createAdminGlobalSearchResultForTest(id: string, sourceId = 'pages'): AdminGlobalSearchResult {
  return { id, sourceId, title: id, description: id, categoryLabel: 'Group', linkLabel: `/${id}`, routeTarget: `/${id}` }
}

function createAdminGlobalSearchSourceForTest(overrides: Partial<AdminGlobalSearchSource> = {}): AdminGlobalSearchSource {
  return {
    id: 'pages', label: 'Pages', recommendationsLabel: 'Common pages',
    getRecommendations: () => [createAdminGlobalSearchResultForTest('common')],
    search: (query) => [createAdminGlobalSearchResultForTest(query)],
    ...overrides,
  }
}

function createDeferredAdminGlobalSearchResults() {
  let resolve!: (results: AdminGlobalSearchResult[]) => void
  let reject!: (error: unknown) => void
  const promise = new Promise<AdminGlobalSearchResult[]>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}

function createDisposableAdminGlobalSearchController(initialSources = [createAdminGlobalSearchSourceForTest()]) {
  const sources = shallowRef(initialSources)
  const navigateToResult = vi.fn(async () => true)
  const closeSearchDialog = vi.fn()
  const scope = effectScope()
  activeSearchTestScopes.push(scope)
  const state = scope.run(() => useAdminGlobalSearchController({ sources, navigateToResult, closeSearchDialog }))!
  return { scope, sources, state, navigateToResult, closeSearchDialog }
}

function createAdminGlobalSearchKeyboardEventForTest(overrides: Partial<KeyboardEvent> = {}): KeyboardEvent {
  return { isComposing: false, preventDefault: vi.fn(), ...overrides } as unknown as KeyboardEvent
}

describe('global admin search controller', () => {
  it('uses recommendations for whitespace, then passes a normalized query to independent sources', async () => {
    const search = vi.fn((query: string) => [createAdminGlobalSearchResultForTest(query)])
    const { state } = createDisposableAdminGlobalSearchController([createAdminGlobalSearchSourceForTest({ search })])
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('common'))
    state.query.value = ' ＡＰＩ '
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('API'))
    expect(search.mock.calls[0][0]).toBe('API')
    expect(state.resultGroups.value[0].label).toBe('Pages')
    state.query.value = '\n  '
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('common'))
    expect(state.resultGroups.value[0].label).toBe('Common pages')
  })

  it('aborts previous queries and ignores both late success and late rejection', async () => {
    const first = createDeferredAdminGlobalSearchResults()
    const second = createDeferredAdminGlobalSearchResults()
    const signals: AbortSignal[] = []
    const source = createAdminGlobalSearchSourceForTest({ search: (query, { signal }) => {
      signals.push(signal)
      return query === 'first' ? first.promise : second.promise
    } })
    const { state } = createDisposableAdminGlobalSearchController([source])
    state.query.value = 'first'
    state.query.value = 'second'
    expect(signals[0].aborted).toBe(true)
    first.resolve([createAdminGlobalSearchResultForTest('stale')])
    await Promise.resolve()
    expect(state.results.value).toEqual([])
    expect(state.loading.value).toBe(true)
    second.resolve([createAdminGlobalSearchResultForTest('latest')])
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('latest'))

    const oldFailure = createDeferredAdminGlobalSearchResults()
    source.search = () => oldFailure.promise
    state.query.value = 'third'
    state.query.value = ''
    oldFailure.reject(new Error('late failure'))
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('common'))
    expect(state.failedSourceLabels.value).toEqual([])
  })

  it('clears results immediately when permissions or locale replace sources and ignores the old response', async () => {
    const pending = createDeferredAdminGlobalSearchResults()
    let originalSignal!: AbortSignal
    const { state, sources } = createDisposableAdminGlobalSearchController([createAdminGlobalSearchSourceForTest({ search: (_, { signal }) => {
      originalSignal = signal
      return pending.promise
    } })])
    state.query.value = 'protected'
    sources.value = [createAdminGlobalSearchSourceForTest({ label: 'Allowed pages', search: () => [] })]
    expect(originalSignal.aborted).toBe(true)
    expect(state.results.value).toEqual([])
    pending.resolve([createAdminGlobalSearchResultForTest('revoked-page')])
    await vi.waitFor(() => expect(state.loading.value).toBe(false))
    expect(state.results.value).toEqual([])
    expect(state.resultGroups.value[0].label).toBe('Allowed pages')
  })

  it('publishes fast sources while others are pending and isolates failures with retry support', async () => {
    const slowResults = createDeferredAdminGlobalSearchResults()
    const fastSearch = vi.fn((query: string) => [createAdminGlobalSearchResultForTest(query)])
    const slowSearch = vi.fn().mockReturnValueOnce(slowResults.promise).mockReturnValue([])
    const { state } = createDisposableAdminGlobalSearchController([
      createAdminGlobalSearchSourceForTest({ search: fastSearch }),
      createAdminGlobalSearchSourceForTest({ id: 'orders', label: 'Orders', search: slowSearch }),
    ])
    state.query.value = 'current'
    await vi.waitFor(() => expect(state.results.value[0]?.id).toBe('current'))
    expect(state.loading.value).toBe(true)
    slowResults.reject(new Error('offline'))
    await vi.waitFor(() => expect(state.failedSourceLabels.value).toEqual(['Orders']))
    expect(state.loading.value).toBe(false)
    expect(state.results.value[0].id).toBe('current')
    await state.retryFailedAdminGlobalSearchSources()
    expect(state.failedSourceLabels.value).toEqual([])
    expect(state.results.value[0].id).toBe('current')
    expect(fastSearch).toHaveBeenCalledTimes(1)
    expect(slowSearch).toHaveBeenCalledTimes(2)
  })

  it('aborts on disposal and prevents late responses and keyboard actions from changing closed search', async () => {
    const pending = createDeferredAdminGlobalSearchResults()
    let requestSignal!: AbortSignal
    const { state, scope, navigateToResult } = createDisposableAdminGlobalSearchController([createAdminGlobalSearchSourceForTest({ search: (_, { signal }) => {
      requestSignal = signal
      return pending.promise
    } })])
    state.query.value = 'pending'
    scope.stop()
    expect(requestSignal.aborted).toBe(true)
    pending.resolve([createAdminGlobalSearchResultForTest('late')])
    await Promise.resolve()
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    expect(state.results.value).toEqual([])
    expect(state.loading.value).toBe(false)
    expect(navigateToResult).not.toHaveBeenCalled()
  })

  it('preserves native modified links and closes only for an unmodified primary click', () => {
    const { state, closeSearchDialog } = createDisposableAdminGlobalSearchController()
    const primaryClick = { button: 0, ctrlKey: false, metaKey: false, shiftKey: false, altKey: false }
    for (const override of [{ button: 1 }, { ctrlKey: true }, { metaKey: true }, { shiftKey: true }, { altKey: true }]) {
      state.closeAdminGlobalSearchAfterUnmodifiedLinkActivation({ ...primaryClick, ...override } as MouseEvent)
    }
    expect(closeSearchDialog).not.toHaveBeenCalled()
    state.closeAdminGlobalSearchAfterUnmodifiedLinkActivation(primaryClick as MouseEvent)
    expect(closeSearchDialog).toHaveBeenCalledTimes(1)
  })

  it('ignores IME Enter, opens the first result and keeps the dialog open on failed navigation', async () => {
    const { state, navigateToResult, closeSearchDialog } = createDisposableAdminGlobalSearchController()
    await vi.waitFor(() => expect(state.results.value).toHaveLength(1))
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest({ isComposing: true }))
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest({ keyCode: 229 }))
    expect(navigateToResult).not.toHaveBeenCalled()
    navigateToResult.mockResolvedValueOnce(false)
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    expect(closeSearchDialog).not.toHaveBeenCalled()
    expect(state.navigationFailed.value).toBe(true)
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    expect(navigateToResult).toHaveBeenLastCalledWith('/common')
    expect(closeSearchDialog).toHaveBeenCalledTimes(1)
  })

  it('prevents a pending navigation from closing a disposed dialog', async () => {
    let completeNavigation!: (navigated: boolean) => void
    const pendingNavigation = new Promise<boolean>((resolve) => { completeNavigation = resolve })
    const { state, scope, navigateToResult, closeSearchDialog } = createDisposableAdminGlobalSearchController()
    await vi.waitFor(() => expect(state.results.value).toHaveLength(1))
    navigateToResult.mockReturnValue(pendingNavigation)
    const navigation = state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    scope.stop()
    completeNavigation(true)
    await navigation
    expect(closeSearchDialog).not.toHaveBeenCalled()
  })

  it.each(['query', 'permissions'])('prevents an old navigation from closing search after a %s change', async (change) => {
    let completeNavigation!: (navigated: boolean) => void
    const pendingNavigation = new Promise<boolean>((resolve) => { completeNavigation = resolve })
    const { state, sources, navigateToResult, closeSearchDialog } = createDisposableAdminGlobalSearchController()
    await vi.waitFor(() => expect(state.results.value).toHaveLength(1))
    navigateToResult.mockReturnValueOnce(pendingNavigation)
    const navigation = state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    if (change === 'query') state.query.value = 'updated'
    else sources.value = [createAdminGlobalSearchSourceForTest({ label: 'Allowed pages' })]
    completeNavigation(true)
    await navigation
    expect(closeSearchDialog).not.toHaveBeenCalled()
    expect(state.navigationFailed.value).toBe(false)
  })

  it('clears a navigation failure when a new search starts', async () => {
    const { state, navigateToResult } = createDisposableAdminGlobalSearchController()
    await vi.waitFor(() => expect(state.results.value).toHaveLength(1))
    navigateToResult.mockResolvedValueOnce(false)
    await state.openFirstAdminGlobalSearchResultFromKeyboard(createAdminGlobalSearchKeyboardEventForTest())
    expect(state.navigationFailed.value).toBe(true)
    state.query.value = 'updated'
    expect(state.navigationFailed.value).toBe(false)
  })
})
