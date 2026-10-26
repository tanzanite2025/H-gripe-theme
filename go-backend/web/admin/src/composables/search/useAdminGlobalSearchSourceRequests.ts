import { computed, onScopeDispose, readonly, ref, shallowRef, toValue, watch, type MaybeRefOrGetter } from 'vue'
import type { AdminGlobalSearchResultGroup, AdminGlobalSearchSource } from '@/modules/search/adminGlobalSearchResultAndSourceContracts'

interface AdminGlobalSearchSourceRequestsOptions {
  query: MaybeRefOrGetter<string>
  sources: MaybeRefOrGetter<AdminGlobalSearchSource[]>
}

interface ActiveAdminGlobalSearchSourceRequest {
  controller: AbortController
  generation: number
  cancelDebounce?: () => void
}

export function useAdminGlobalSearchSourceRequests(options: AdminGlobalSearchSourceRequestsOptions) {
  const resultGroups = shallowRef<AdminGlobalSearchResultGroup[]>([])
  const searchRequestGeneration = ref(0)
  const loading = computed(() => resultGroups.value.some((group) => group.loading))
  const failedSourceLabels = computed(() => resultGroups.value.filter((group) => group.failed).map((group) => group.label))
  const activeRequests = new Map<number, ActiveAdminGlobalSearchSourceRequest>()
  let currentSources: AdminGlobalSearchSource[] = []
  let currentQuery = ''
  let disposed = false

  function cancelAllAdminGlobalSearchSourceRequests(): void {
    for (const request of activeRequests.values()) {
      request.controller.abort()
      request.cancelDebounce?.()
    }
    activeRequests.clear()
  }

  function isCurrentAdminGlobalSearchSourceRequest(index: number, request: ActiveAdminGlobalSearchSourceRequest): boolean {
    return !disposed && !request.controller.signal.aborted
      && request.generation === searchRequestGeneration.value && activeRequests.get(index) === request
  }

  function updateAdminGlobalSearchResultGroup(index: number, update: Partial<AdminGlobalSearchResultGroup>): void {
    resultGroups.value = resultGroups.value.map((group, groupIndex) => groupIndex === index ? { ...group, ...update } : group)
  }

  async function requestAdminGlobalSearchSourceResults(source: AdminGlobalSearchSource, index: number, debounceQuery: boolean): Promise<void> {
    const request: ActiveAdminGlobalSearchSourceRequest = { controller: new AbortController(), generation: searchRequestGeneration.value }
    const query = currentQuery
    activeRequests.set(index, request)
    updateAdminGlobalSearchResultGroup(index, { loading: true, failed: false, results: [] })
    try {
      const configuredDelay = source.queryDebounceMs ?? 0
      const delay = debounceQuery && query && Number.isFinite(configuredDelay) ? Math.max(0, configuredDelay) : 0
      if (delay > 0) {
        await new Promise<void>((resolve) => {
          const timer = setTimeout(resolve, delay)
          request.cancelDebounce = () => { clearTimeout(timer); resolve() }
        })
        request.cancelDebounce = undefined
      }
      if (!isCurrentAdminGlobalSearchSourceRequest(index, request)) return
      const context = { signal: request.controller.signal }
      const sourceResults = await (query ? source.search(query, context) : source.getRecommendations(context))
      if (!isCurrentAdminGlobalSearchSourceRequest(index, request)) return
      updateAdminGlobalSearchResultGroup(index, { results: sourceResults })
    } catch {
      if (!isCurrentAdminGlobalSearchSourceRequest(index, request)) return
      updateAdminGlobalSearchResultGroup(index, { failed: true })
    } finally {
      if (isCurrentAdminGlobalSearchSourceRequest(index, request)) {
        activeRequests.delete(index)
        updateAdminGlobalSearchResultGroup(index, { loading: false })
      }
    }
  }

  async function refreshAdminGlobalSearchResults(): Promise<void> {
    if (disposed) return
    cancelAllAdminGlobalSearchSourceRequests()
    searchRequestGeneration.value++
    currentSources = [...toValue(options.sources)]
    currentQuery = toValue(options.query).normalize('NFKC').trim()
    resultGroups.value = currentSources.map((source) => ({
      sourceId: source.id,
      label: currentQuery ? source.label : source.recommendationsLabel,
      results: [],
      failed: false,
      loading: true,
    }))
    await Promise.all(currentSources.map((source, index) => requestAdminGlobalSearchSourceResults(source, index, true)))
  }

  async function retryFailedAdminGlobalSearchSources(): Promise<void> {
    if (disposed) return
    const failedSourceIndices = resultGroups.value.flatMap((group, index) => group.failed ? [index] : [])
    await Promise.all(failedSourceIndices.map((index) => requestAdminGlobalSearchSourceResults(currentSources[index], index, false)))
  }

  watch([() => toValue(options.query), () => toValue(options.sources)], () => { void refreshAdminGlobalSearchResults() }, { immediate: true, flush: 'sync' })

  onScopeDispose(() => {
    disposed = true
    searchRequestGeneration.value++
    cancelAllAdminGlobalSearchSourceRequests()
    currentSources = []
    resultGroups.value = []
  })

  return {
    resultGroups,
    loading,
    failedSourceLabels,
    searchRequestGeneration: readonly(searchRequestGeneration),
    refreshAdminGlobalSearchResults,
    retryFailedAdminGlobalSearchSources,
  }
}
