import { computed, ref, watch } from 'vue'
import { useAsyncData, useRoute, useRouter } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  buildSchwalbeTireCatalogFilterOptions,
  filterSchwalbeTireCatalogItems,
} from '~/data/tireguides/schwalbeTireCatalogFilterModel'
import {
  mergeSchwalbeTireCatalogFilterQuery,
  parseSchwalbeTireCatalogFilterQuery,
  type SchwalbeCatalogSort,
} from '~/data/tireguides/schwalbeTireCatalogFilterQuery'
import {
  fetchSchwalbeTireCatalog,
  type SchwalbeTireCatalogItem,
} from '~/data/tireguides/schwalbeCatalog'

export type { SchwalbeCatalogSort } from '~/data/tireguides/schwalbeTireCatalogFilterQuery'

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
  const initialFilterState = parseSchwalbeTireCatalogFilterQuery(
    route.query as Record<string, unknown>,
  )
  const searchInput = ref(initialSearch)
  const submittedSearch = ref(initialSearch)
  const selectedModel = ref(initialFilterState.modelName || 'ALL')
  const selectedTireWidthsMm = ref(initialFilterState.nominalTireWidthsMm)
  const selectedBeadSeatDiametersMm = ref(initialFilterState.beadSeatDiametersMm)
  const selectedBeads = ref(initialFilterState.beads)
  const selectedSeals = ref(initialFilterState.seals)
  const selectedEBikeRatings = ref(initialFilterState.eBikeRatings)
  const sortBy = ref<SchwalbeCatalogSort>(initialFilterState.sortBy)
  const requestedPage = computed(() => parsePage(route.query.page) || 1)

  const requestKey = computed(() => `schwalbe-tire-catalog:${submittedSearch.value}`)
  const { data, pending, error, refresh } = await useAsyncData<SchwalbeTireCatalogItem[]>(
    requestKey,
    () => fetchSchwalbeTireCatalog(request, submittedSearch.value),
    { default: () => [] },
  )

  const items = computed(() => data.value || [])
  const filterOptions = computed(() => buildSchwalbeTireCatalogFilterOptions(items.value))
  const modelOptions = computed(() => [
    'ALL',
    ...[...new Set([
      ...items.value.map(item => item.model_name.trim()).filter(Boolean),
      ...(selectedModel.value === 'ALL' ? [] : [selectedModel.value]),
    ])].sort((left, right) => left.localeCompare(right)),
  ])

  const filteredItems = computed(() => {
    const filtered = filterSchwalbeTireCatalogItems(items.value, {
      modelNames: selectedModel.value === 'ALL' ? [] : [selectedModel.value],
      nominalTireWidthMm: selectedTireWidthsMm.value,
      beadSeatDiameterMm: selectedBeadSeatDiametersMm.value,
      beads: selectedBeads.value,
      seals: selectedSeals.value,
      eBikeRatings: selectedEBikeRatings.value,
    })
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

  const hasActiveFilters = computed(() => (
    selectedModel.value !== 'ALL'
    || selectedTireWidthsMm.value.length > 0
    || selectedBeadSeatDiametersMm.value.length > 0
    || selectedBeads.value.length > 0
    || selectedSeals.value.length > 0
    || selectedEBikeRatings.value.length > 0
  ))

  let applyingRouteFilterState = false
  let routeFilterStateNavigationVersion = 0
  let isRouteFilterStateNavigationPending = false

  const updateRouteFilterState = () => {
    if (!import.meta.client || applyingRouteFilterState) return
    const query = mergeSchwalbeTireCatalogFilterQuery(
      route.query as Record<string, unknown>,
      {
        modelName: selectedModel.value === 'ALL' ? null : selectedModel.value,
        nominalTireWidthsMm: selectedTireWidthsMm.value,
        beadSeatDiametersMm: selectedBeadSeatDiametersMm.value,
        beads: selectedBeads.value,
        seals: selectedSeals.value,
        eBikeRatings: selectedEBikeRatings.value,
        sortBy: sortBy.value,
      },
    )
    delete query.page

    const targetFullPath = router.resolve({
      path: route.path,
      query: query as typeof route.query,
      hash: route.hash,
    }).fullPath
    if (targetFullPath === route.fullPath) return
    const navigationVersion = ++routeFilterStateNavigationVersion
    isRouteFilterStateNavigationPending = true
    void router.replace({ query: query as typeof route.query }).then(
      () => {
        if (navigationVersion === routeFilterStateNavigationVersion) {
          isRouteFilterStateNavigationPending = false
        }
      },
      () => {
        if (navigationVersion === routeFilterStateNavigationVersion) {
          isRouteFilterStateNavigationPending = false
        }
      },
    )
  }

  watch(
    () => JSON.stringify([
      selectedModel.value,
      selectedTireWidthsMm.value,
      selectedBeadSeatDiametersMm.value,
      selectedBeads.value,
      selectedSeals.value,
      selectedEBikeRatings.value,
      sortBy.value,
    ]),
    updateRouteFilterState,
    { flush: 'sync' },
  )

  watch(
    () => [
      route.query.model,
      route.query.tire_width_mm,
      route.query.bead_seat_diameter_mm,
      route.query.bead,
      route.query.seal,
      route.query.e_bike_rating,
      route.query.sort,
    ],
    () => {
      const nextFilterState = parseSchwalbeTireCatalogFilterQuery(
        route.query as Record<string, unknown>,
      )
      applyingRouteFilterState = true
      selectedModel.value = nextFilterState.modelName || 'ALL'
      selectedTireWidthsMm.value = nextFilterState.nominalTireWidthsMm
      selectedBeadSeatDiametersMm.value = nextFilterState.beadSeatDiametersMm
      selectedBeads.value = nextFilterState.beads
      selectedSeals.value = nextFilterState.seals
      selectedEBikeRatings.value = nextFilterState.eBikeRatings
      sortBy.value = nextFilterState.sortBy
      applyingRouteFilterState = false
    },
    { flush: 'sync' },
  )

  watch(() => route.query.search, (value) => {
    const nextSearch = typeof value === 'string' ? value.trim() : ''
    if (nextSearch === submittedSearch.value) return
    searchInput.value = nextSearch
    submittedSearch.value = nextSearch
  })

  watch([pending, totalPages, requestedPage], ([isPending, pages, requested]) => {
    if (!import.meta.client || isPending || isRouteFilterStateNavigationPending || !route.query.page) return

    const normalizedPage = Math.min(requested, pages)
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
    selectedTireWidthsMm,
    selectedBeadSeatDiametersMm,
    selectedBeads,
    selectedSeals,
    selectedEBikeRatings,
    sortBy,
    items,
    filteredItems,
    totalItems,
    visibleItems,
    modelOptions,
    tireWidthOptions: computed(() => filterOptions.value.nominalTireWidthsMm),
    beadSeatDiameterOptions: computed(() => filterOptions.value.beadSeatDiametersMm),
    beadOptions: computed(() => filterOptions.value.beads),
    sealOptions: computed(() => filterOptions.value.seals),
    eBikeRatingOptions: computed(() => filterOptions.value.eBikeRatings),
    hasActiveFilters,
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
