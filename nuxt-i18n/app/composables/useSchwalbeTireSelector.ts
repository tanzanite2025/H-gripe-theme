import { computed, ref, watch } from 'vue'
import { useAsyncData, useRoute, useRouter } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  fetchSchwalbeTireCatalog,
  type SchwalbeTireCatalogItem,
} from '~/data/tireguides/schwalbeCatalog'

export type SchwalbeCatalogSort = 'model' | 'etrto' | 'article'

export const useSchwalbeTireSelector = async () => {
  const route = useRoute()
  const router = useRouter()
  const { request } = useApiRequest()

  const initialSearch = typeof route.query.search === 'string' ? route.query.search.trim() : ''
  const searchInput = ref(initialSearch)
  const submittedSearch = ref(initialSearch)
  const selectedModel = ref('ALL')
  const sortBy = ref<SchwalbeCatalogSort>('model')

  const requestKey = computed(() => `schwalbe-tire-catalog:${submittedSearch.value}`)
  const { data, pending, error, refresh } = await useAsyncData<SchwalbeTireCatalogItem[]>(
    requestKey,
    () => fetchSchwalbeTireCatalog(request, submittedSearch.value),
    { default: () => [] },
  )

  const items = computed(() => data.value || [])
  const modelOptions = computed(() => [
    'ALL',
    ...[...new Set(items.value.map(item => item.model_name.trim()).filter(Boolean))]
      .sort((left, right) => left.localeCompare(right)),
  ])

  const visibleItems = computed(() => {
    const filtered = selectedModel.value === 'ALL'
      ? items.value
      : items.value.filter(item => item.model_name === selectedModel.value)
    return [...filtered].sort((left, right) => {
      if (sortBy.value === 'etrto') {
        return left.etrto.localeCompare(right.etrto, undefined, { numeric: true }) || left.article_no.localeCompare(right.article_no)
      }
      if (sortBy.value === 'article') {
        return left.article_no.localeCompare(right.article_no, undefined, { numeric: true })
      }
      return left.model_name.localeCompare(right.model_name) || left.etrto.localeCompare(right.etrto, undefined, { numeric: true }) || left.article_no.localeCompare(right.article_no)
    })
  })

  watch(modelOptions, (options) => {
    if (!options.includes(selectedModel.value)) selectedModel.value = 'ALL'
  })

  watch(() => route.query.search, (value) => {
    const nextSearch = typeof value === 'string' ? value.trim() : ''
    if (nextSearch === submittedSearch.value) return
    searchInput.value = nextSearch
    submittedSearch.value = nextSearch
  })

  const submitSearch = async () => {
    const nextSearch = searchInput.value.trim()
    submittedSearch.value = nextSearch
    const query = { ...route.query }
    if (nextSearch) query.search = nextSearch
    else delete query.search
    await router.replace({ query })
  }

  const clearSearch = async () => {
    searchInput.value = ''
    await submitSearch()
  }

  return {
    searchInput,
    submittedSearch,
    selectedModel,
    sortBy,
    items,
    visibleItems,
    modelOptions,
    pending,
    error,
    refresh,
    submitSearch,
    clearSearch,
  }
}
