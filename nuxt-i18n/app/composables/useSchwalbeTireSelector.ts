import { computed, ref, watch } from 'vue'
import { useAsyncData, useRoute, useRouter } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  fetchSchwalbeTireCatalog,
  type SchwalbeTireCatalogItem,
} from '~/data/tireguides/schwalbeCatalog'

export type SchwalbeCatalogSort = 'model' | 'etrto' | 'article'

export const SCHWALBE_CATALOG_PAGE_SIZE = 20

type PaginationToken = number | 'ellipsis'

const parsePage = (value: unknown): number | null => {
  if (typeof value !== 'string' || !/^[1-9]\d*$/.test(value.trim())) return null
  const parsed = Number(value)
  return Number.isSafeInteger(parsed) ? parsed : null
}

export const useSchwalbeTireSelector = async () => {
  const route = useRoute()
  const router = useRouter()
  const { request } = useApiRequest()

  const initialSearch = typeof route.query.search === 'string' ? route.query.search.trim() : ''
  const searchInput = ref(initialSearch)
  const submittedSearch = ref(initialSearch)
  const selectedModel = ref('ALL')
  const sortBy = ref<SchwalbeCatalogSort>('model')
  const requestedPage = computed(() => parsePage(route.query.page) || 1)

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

  const filteredItems = computed(() => {
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

  const totalItems = computed(() => filteredItems.value.length)
  const totalPages = computed(() => Math.max(1, Math.ceil(totalItems.value / SCHWALBE_CATALOG_PAGE_SIZE)))
  const currentPage = computed(() => Math.min(requestedPage.value, totalPages.value))
  const visibleItems = computed(() => {
    const start = (currentPage.value - 1) * SCHWALBE_CATALOG_PAGE_SIZE
    return filteredItems.value.slice(start, start + SCHWALBE_CATALOG_PAGE_SIZE)
  })

  const paginationPages = computed<PaginationToken[]>(() => {
    if (totalPages.value <= 7) {
      return Array.from({ length: totalPages.value }, (_, index) => index + 1)
    }

    const candidates = new Set([1, totalPages.value, currentPage.value])
    if (currentPage.value > 1) candidates.add(currentPage.value - 1)
    if (currentPage.value < totalPages.value) candidates.add(currentPage.value + 1)
    const sorted = [...candidates].sort((left, right) => left - right)
    const result: PaginationToken[] = []
    sorted.forEach((page, index) => {
      const previous = sorted[index - 1]
      if (previous !== undefined && page - previous > 1) result.push('ellipsis')
      result.push(page)
    })
    return result
  })

  const pageQuery = (page: number) => {
    const query = { ...route.query }
    if (page <= 1) delete query.page
    else query.page = String(page)
    return { query }
  }

  watch(modelOptions, (options) => {
    if (!options.includes(selectedModel.value)) selectedModel.value = 'ALL'
  })

  watch(() => route.query.search, (value) => {
    const nextSearch = typeof value === 'string' ? value.trim() : ''
    if (nextSearch === submittedSearch.value) return
    searchInput.value = nextSearch
    submittedSearch.value = nextSearch
  })

  watch([selectedModel, sortBy], ([nextModel, nextSort], [previousModel, previousSort]) => {
    if (nextModel === previousModel && nextSort === previousSort) return
    if (!route.query.page) return
    const query = { ...route.query }
    delete query.page
    void router.replace({ query })
  })

  watch([pending, totalPages], ([isPending, pages]) => {
    if (!import.meta.client || isPending || !route.query.page) return

    const parsedPage = parsePage(route.query.page)
    const normalizedPage = Math.min(parsedPage || 1, pages)
    const query = { ...route.query }
    if (normalizedPage <= 1) delete query.page
    else query.page = String(normalizedPage)

    const currentPageQuery = typeof route.query.page === 'string' ? route.query.page : ''
    if (query.page === currentPageQuery) return
    void router.replace({ query })
  }, { immediate: true })

  const submitSearch = async () => {
    const nextSearch = searchInput.value.trim()
    submittedSearch.value = nextSearch
    const query = { ...route.query }
    if (nextSearch) query.search = nextSearch
    else delete query.search
    delete query.page
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
    filteredItems,
    totalItems,
    visibleItems,
    modelOptions,
    pageSize: SCHWALBE_CATALOG_PAGE_SIZE,
    currentPage,
    totalPages,
    paginationPages,
    pageQuery,
    pending,
    error,
    refresh,
    submitSearch,
    clearSearch,
  }
}
