import { computed, onScopeDispose, ref, watch, type MaybeRefOrGetter } from 'vue'
import type { RouteLocationRaw } from 'vue-router'
import type { AdminGlobalSearchSource } from '@/modules/search/adminGlobalSearchResultAndSourceContracts'
import { useAdminGlobalSearchSourceRequests } from './useAdminGlobalSearchSourceRequests'

interface AdminGlobalSearchControllerOptions {
  sources: MaybeRefOrGetter<AdminGlobalSearchSource[]>
  navigateToResult: (target: RouteLocationRaw) => Promise<boolean>
  closeSearchDialog: () => void
}

export function useAdminGlobalSearchController(options: AdminGlobalSearchControllerOptions) {
  const query = ref('')
  const {
    resultGroups,
    loading,
    failedSourceLabels,
    searchRequestGeneration,
    refreshAdminGlobalSearchResults,
    retryFailedAdminGlobalSearchSources,
  } = useAdminGlobalSearchSourceRequests({ query, sources: options.sources })
  const navigating = ref(false)
  const navigationFailed = ref(false)
  const results = computed(() => resultGroups.value.flatMap((group) => group.results))
  let disposed = false

  function closeAdminGlobalSearchAfterUnmodifiedLinkActivation(event: MouseEvent): void {
    if (disposed || event.button !== 0 || event.ctrlKey || event.metaKey || event.shiftKey || event.altKey) return
    options.closeSearchDialog()
  }

  async function openFirstAdminGlobalSearchResultFromKeyboard(event: KeyboardEvent): Promise<void> {
    if (disposed || event.isComposing || event.keyCode === 229) return
    event.preventDefault()
    const firstResult = results.value[0]
    if (!firstResult || navigating.value) return
    const generation = searchRequestGeneration.value
    navigating.value = true
    navigationFailed.value = false
    try {
      const navigated = await options.navigateToResult(firstResult.routeTarget)
      if (disposed || generation !== searchRequestGeneration.value) return
      if (navigated) options.closeSearchDialog()
      else navigationFailed.value = true
    } catch {
      if (!disposed && generation === searchRequestGeneration.value) navigationFailed.value = true
    } finally {
      if (!disposed) navigating.value = false
    }
  }

  watch(searchRequestGeneration, () => { navigationFailed.value = false }, { flush: 'sync' })

  onScopeDispose(() => {
    disposed = true
    navigating.value = false
  })

  return {
    query,
    resultGroups,
    results,
    loading,
    failedSourceLabels,
    navigationFailed,
    refreshAdminGlobalSearchResults,
    retryFailedAdminGlobalSearchSources,
    closeAdminGlobalSearchAfterUnmodifiedLinkActivation,
    openFirstAdminGlobalSearchResultFromKeyboard,
  }
}
