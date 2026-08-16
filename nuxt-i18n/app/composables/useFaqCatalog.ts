import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { fetchAllFaqData, resolvePageFaqDataList } from '~/data/faq'

export async function useFaqCatalog() {
  const { locale } = useI18n()
  const {
    data: asyncAllPages,
    pending,
    error,
    refresh: refreshFaqCatalog,
  } = await useAsyncData(
    () => `faq-catalog-${locale.value}`,
    () => fetchAllFaqData(),
    { watch: [locale] },
  )

  const allPages = computed(() => resolvePageFaqDataList(asyncAllPages.value || []))

  return {
    allPages,
    pending,
    error,
    refreshFaqCatalog,
  }
}
