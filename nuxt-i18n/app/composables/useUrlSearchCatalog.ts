import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { fetchAllUrlSearchData } from '~/data/url-search/backend'

export async function useUrlSearchCatalog() {
  const { locale } = useI18n()
  const {
    data: asyncUrlSearchProfiles,
    pending,
    error,
    refresh: refreshUrlSearchCatalog,
  } = await useAsyncData(
    () => `url-search-catalog-${locale.value}`,
    () => fetchAllUrlSearchData(),
    { watch: [locale] },
  )

  const urlSearchProfiles = computed(() => asyncUrlSearchProfiles.value || [])

  return {
    urlSearchProfiles,
    pending,
    error,
    refreshUrlSearchCatalog,
  }
}
