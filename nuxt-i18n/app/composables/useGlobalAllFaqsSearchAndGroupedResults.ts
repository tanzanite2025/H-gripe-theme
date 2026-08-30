import { computed, ref, watch } from 'vue'
import { useRoute } from '#imports'
import {
  flattenGlobalAllFaqItems,
  type GlobalAllFaqFlatItem,
} from '~/data/faq'
import { useFaqAccordionState } from '~/composables/useFaqAccordionState'
import { useFaqCatalog } from '~/composables/useFaqCatalog'
import { useFaqDeepLink } from '~/composables/useFaqDeepLink'
import { useFaqGroupedResults } from '~/composables/useFaqGroupedResults'
import { useFaqSearch } from '~/composables/useFaqSearch'
import { useFaqSearchOverlayState } from '~/composables/useFaqSearchOverlayState'
import { useFaqTopics } from '~/composables/useFaqTopics'

export type GlobalAllFaqsViewMode = 'page' | 'search-overlay'

export interface GlobalAllFaqsSearchAndGroupedResultsOptions {
  mode?: GlobalAllFaqsViewMode
  syncDeepLink?: boolean
}

export type {
  GlobalAllFaqFlatItem,
  GlobalAllFaqSearchTopic,
  GlobalAllFaqsDisplayGroup,
} from '~/data/faq'

export async function useGlobalAllFaqsSearchAndGroupedResults(
  options: GlobalAllFaqsSearchAndGroupedResultsOptions = {},
) {
  const mode = options.mode ?? 'page'
  const isSearchOverlay = mode === 'search-overlay'
  const syncDeepLink = options.syncDeepLink ?? !isSearchOverlay
  const route = useRoute()

  const faqCatalogPromise = useFaqCatalog()

  const {
    allPages,
    pending: faqPending,
    refreshFaqCatalog,
  } = await faqCatalogPromise
  const activePageId = ref('all')
  const {
    expandedItems,
    toggleItem,
    resetExpandedItems,
    expandItem,
  } = useFaqAccordionState()

  const allItems = computed<GlobalAllFaqFlatItem[]>(() => {
    return flattenGlobalAllFaqItems(allPages.value)
  })

  const {
    activeTopicId,
    featuredItems,
    featuredTopics,
    activeTopic,
    topicItems,
    selectTopic: selectTopicState,
    resetTopic,
  } = useFaqTopics(allItems)
  const {
    searchQuery,
    filteredItems,
    searchResults,
    searchResultCount,
  } = useFaqSearch(allItems, activePageId)
  const {
    displayedGroups,
    hasMoreGroups,
    loadMoreGroups,
    showAllGroups,
    resetGroups,
    showGroupsThrough,
  } = useFaqGroupedResults(filteredItems)

  const {
    applyingDeepLink,
  } = useFaqDeepLink({
    enabled: syncDeepLink,
    route,
    allPages,
    allItems,
    activePageId,
    showAllGroups,
    showGroupsThrough,
    expandItem,
    resetExpandedItems,
  })

  const {
    resetSearchOverlayState,
  } = useFaqSearchOverlayState({
    enabled: isSearchOverlay,
    searchQuery,
    searchResults,
    resetTopic,
    resetExpandedItems,
    expandItem,
  })

  watch(activePageId, () => {
    if (applyingDeepLink.value) return

    resetExpandedItems()
    if (activePageId.value === 'all') {
      resetGroups()
    } else {
      showAllGroups()
    }
  })

  watch(searchQuery, () => {
    if (applyingDeepLink.value || isSearchOverlay) return

    if (searchQuery.value.trim()) {
      showAllGroups()
      resetExpandedItems()
      return
    }

    resetExpandedItems()
    if (activePageId.value === 'all') {
      resetGroups()
    }
  })

  const selectTopic = (topicId: string) => {
    selectTopicState(topicId)
    resetExpandedItems()
  }

  const pending = computed(() => faqPending.value)

  return {
    allPages,
    pending,
    refreshAllFaqData: refreshFaqCatalog,
    featuredItems,
    featuredTopics,
    activeTopicId,
    activeTopic,
    topicItems,
    searchQuery,
    activePageId,
    expandedItems,
    filteredItems,
    searchResults,
    searchResultCount,
    displayedGroups,
    hasMoreGroups,
    toggleItem,
    selectTopic,
    resetSearchOverlayState,
    loadMoreGroups,
  }
}
