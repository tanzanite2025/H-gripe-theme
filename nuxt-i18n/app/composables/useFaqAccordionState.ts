import { ref } from 'vue'

export function useFaqAccordionState() {
  const expandedItems = ref<Set<string>>(new Set())

  const toggleItem = (itemId: string) => {
    expandedItems.value = expandedItems.value.has(itemId)
      ? new Set<string>()
      : new Set([itemId])
  }

  const resetExpandedItems = () => {
    expandedItems.value = new Set<string>()
  }

  const expandItem = (itemId: string) => {
    expandedItems.value = new Set([itemId])
  }

  return {
    expandedItems,
    toggleItem,
    resetExpandedItems,
    expandItem,
  }
}
