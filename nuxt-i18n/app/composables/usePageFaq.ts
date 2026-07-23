import { computed, ref } from 'vue'
import { useAsyncData } from '#imports'
import { fetchFaqData, getFaqData } from '~/data/faq'
import type { FaqCategory, PageFaqProps } from '~/data/faq/types'

export async function usePageFaq(props: PageFaqProps) {
  const { data: asyncFaqData } = await useAsyncData(
    `faq-${props.pageId}`,
    () => props.data ? Promise.resolve(props.data) : fetchFaqData(props.pageId),
  )
  const faqData = computed(() => props.data || asyncFaqData.value || getFaqData(props.pageId))
  const displayTitle = computed(() => props.title || faqData.value?.title || 'Frequently Asked Questions')
  const expandedItems = ref<Set<string>>(new Set())

  const toggleItem = (itemId: string) => {
    if (expandedItems.value.has(itemId)) {
      expandedItems.value.delete(itemId)
    } else {
      expandedItems.value.add(itemId)
    }
    expandedItems.value = new Set(expandedItems.value)
  }

  const displayCategories = computed(() => {
    if (!faqData.value?.categories) return []

    if (!props.maxItems) {
      return faqData.value.categories
    }

    let remainingItems = props.maxItems
    const limitedCategories: FaqCategory[] = []

    for (const category of faqData.value.categories) {
      if (remainingItems <= 0) break

      const itemsToTake = Math.min(category.items.length, remainingItems)
      limitedCategories.push({
        ...category,
        items: category.items.slice(0, itemsToTake),
      })
      remainingItems -= itemsToTake
    }

    return limitedCategories
  })

  const hasMoreItems = computed(() => {
    if (!props.maxItems || !faqData.value?.categories) return false

    const totalItems = faqData.value.categories.reduce(
      (sum, category) => sum + category.items.length,
      0
    )
    return totalItems > props.maxItems
  })

  return {
    faqData,
    displayTitle,
    displayCategories,
    expandedItems,
    toggleItem,
    hasMoreItems
  }
}
