import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { fetchFaqData } from '~/data/faq'
import type { PageFaqProps } from '~/data/faq/types'
import { useFaqAccordionState } from '~/composables/useFaqAccordionState'

export async function usePageFaq(props: PageFaqProps) {
  const { locale, t } = useI18n()
  const { data: asyncFaqData } = await useAsyncData(
    () => `faq-${props.pageId}-${locale.value}`,
    () => props.data ? Promise.resolve(props.data) : fetchFaqData(props.pageId),
    { watch: [locale] }
  )
  const faqData = computed(() => props.data || asyncFaqData.value || null)
  const displayTitle = computed(() => props.title || faqData.value?.title || t('faq.title'))
  const {
    expandedItems,
    toggleItem,
    resetExpandedItems,
  } = useFaqAccordionState()

  const displayItems = computed(() => (
    props.maxItems
      ? faqData.value?.items.slice(0, props.maxItems) || []
      : faqData.value?.items || []
  ))

  const hasMoreItems = computed(() => {
    return Boolean(
      props.maxItems
      && faqData.value
      && faqData.value.items.length > props.maxItems,
    )
  })

  return {
    faqData,
    displayTitle,
    displayItems,
    expandedItems,
    toggleItem,
    resetExpandedItems,
    hasMoreItems
  }
}
