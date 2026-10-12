import { computed } from 'vue'
import { useStorefrontFaqAsyncDataWithLastSuccessfulResponse } from '~/composables/useStorefrontFaqAsyncDataWithLastSuccessfulResponse'
import { fetchAllFaqData, resolvePageFaqDataList } from '~/data/faq'

export async function useFaqCatalog() {
  const { locale } = useI18n()
  const {
    data: asyncAllPages,
    pending,
    error,
    refresh: refreshFaqCatalog,
  } = await useStorefrontFaqAsyncDataWithLastSuccessfulResponse(
    () => `faq-catalog-${locale.value}`,
    () => fetchAllFaqData(),
  )

  const allPages = computed(() => resolvePageFaqDataList(asyncAllPages.value || []))

  return {
    allPages,
    pending,
    error,
    refreshFaqCatalog,
  }
}
