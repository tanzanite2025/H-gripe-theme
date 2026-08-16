import { computed, ref, type Ref } from 'vue'
import {
  groupGlobalAllFaqItemsByPage,
  type GlobalAllFaqFlatItem,
  type GlobalAllFaqsDisplayGroup,
} from '~/data/faq'

const DEFAULT_VISIBLE_GROUPS = 3
const GROUPS_PER_LOAD = 3

export function useFaqGroupedResults(
  filteredItems: Readonly<Ref<GlobalAllFaqFlatItem[]>>,
) {
  const visibleGroupsCount = ref(DEFAULT_VISIBLE_GROUPS)
  const groupedItems = computed<GlobalAllFaqsDisplayGroup[]>(() => (
    groupGlobalAllFaqItemsByPage(filteredItems.value)
  ))
  const displayedGroups = computed(() => (
    groupedItems.value.slice(0, visibleGroupsCount.value)
  ))
  const hasMoreGroups = computed(() => (
    groupedItems.value.length > visibleGroupsCount.value
  ))

  const loadMoreGroups = () => {
    visibleGroupsCount.value = Math.min(
      groupedItems.value.length,
      visibleGroupsCount.value + GROUPS_PER_LOAD,
    )
  }

  const showAllGroups = () => {
    visibleGroupsCount.value = groupedItems.value.length
  }

  const resetGroups = () => {
    visibleGroupsCount.value = DEFAULT_VISIBLE_GROUPS
  }

  const showGroupsThrough = (groupIndex: number) => {
    visibleGroupsCount.value = Math.max(3, groupIndex + 1)
  }

  return {
    displayedGroups,
    hasMoreGroups,
    loadMoreGroups,
    showAllGroups,
    resetGroups,
    showGroupsThrough,
  }
}
