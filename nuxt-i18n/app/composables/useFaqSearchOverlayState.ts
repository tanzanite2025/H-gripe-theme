import { watch, type Ref } from 'vue'
import type { GlobalAllFaqFlatItem } from '~/data/faq'

interface UseFaqSearchOverlayStateOptions {
  enabled: boolean
  searchQuery: Ref<string>
  searchResults: Readonly<Ref<GlobalAllFaqFlatItem[]>>
  resetTopic: () => void
  resetExpandedItems: () => void
  expandItem: (itemId: string) => void
}

export function useFaqSearchOverlayState(
  options: UseFaqSearchOverlayStateOptions,
) {
  const resetSearchOverlayState = () => {
    if (!options.enabled) return

    options.searchQuery.value = ''
    options.resetTopic()
    options.resetExpandedItems()
  }

  if (options.enabled) {
    watch(
      [options.searchQuery, options.searchResults],
      ([query, results]) => {
        if (!query.trim()) {
          options.resetExpandedItems()
          return
        }

        const firstResult = results[0]
        if (firstResult) {
          options.expandItem(firstResult.id)
        } else {
          options.resetExpandedItems()
        }
      },
      { immediate: true },
    )
  }

  return {
    resetSearchOverlayState,
  }
}
