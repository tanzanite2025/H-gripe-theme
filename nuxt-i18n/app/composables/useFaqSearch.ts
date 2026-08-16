import { computed, ref, type Ref } from 'vue'
import {
  filterGlobalAllFaqItems,
  type GlobalAllFaqFlatItem,
} from '~/data/faq'

export function useFaqSearch(
  allItems: Readonly<Ref<GlobalAllFaqFlatItem[]>>,
  activePageId: Readonly<Ref<string>>,
) {
  const searchQuery = ref('')
  const filteredItems = computed(() => (
    filterGlobalAllFaqItems(
      allItems.value,
      searchQuery.value,
      activePageId.value,
    )
  ))

  const searchResultLimit = 6
  const searchResults = computed(() => filteredItems.value.slice(0, searchResultLimit))
  const searchResultCount = computed(() => filteredItems.value.length)

  return {
    searchQuery,
    filteredItems,
    searchResults,
    searchResultCount,
  }
}
