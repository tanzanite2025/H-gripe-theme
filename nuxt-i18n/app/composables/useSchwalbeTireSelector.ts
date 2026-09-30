import { computed, ref, watch } from 'vue'
import { useAsyncData, useRoute, useRouter } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  mergeSchwalbeTireCatalogFilterQuery,
  parseSchwalbeTireCatalogFilterQuery,
  type SchwalbeCatalogSort,
  type SchwalbeTireCatalogFilterQueryState,
} from '~/data/tireguides/schwalbeTireCatalogFilterQuery'
import {
  fetchSchwalbeTireCatalogSelectorPage,
  type SchwalbeTireCatalogSelectorPage,
  type SchwalbeTireCatalogSelectorRequest,
} from '~/data/tireguides/schwalbeCatalog'

export type { SchwalbeCatalogSort } from '~/data/tireguides/schwalbeTireCatalogFilterQuery'

/**
 * Facets edited inside the filter drawer.  The drawer keeps a local draft and
 * applies this complete patch only after the user chooses “Show results”.
 */
export type SchwalbeTireCatalogFacetFilterState = Pick<
  SchwalbeTireCatalogFilterQueryState,
  | 'innerRimWidthMm'
  | 'nominalTireWidthMinMm'
  | 'nominalTireWidthMaxMm'
  | 'nominalTireWidthsMm'
  | 'wheelSizeKeys'
  | 'beadSeatDiametersMm'
  | 'casingConstructions'
  | 'radialOnly'
  | 'beads'
  | 'seals'
  | 'eBikeRatings'
>

export const SCHWALBE_CATALOG_PAGE_SIZE = 20

type PaginationToken = number | 'ellipsis'

const parsePage = (value: unknown): number | null => {
  const candidate = Array.isArray(value) ? value[0] : value
  if (typeof candidate !== 'string' || !/^[1-9]\d*$/.test(candidate.trim())) return null
  const parsed = Number(candidate)
  return Number.isSafeInteger(parsed) ? parsed : null
}

const readRouteSearch = (value: unknown): string => {
  const candidate = Array.isArray(value) ? value[0] : value
  return typeof candidate === 'string' ? candidate.trim() : ''
}

const parseSelectorRouteFilterState = (query: Record<string, unknown>): SchwalbeTireCatalogFilterQueryState => ({
  ...parseSchwalbeTireCatalogFilterQuery(query),
  // This selector no longer exposes minimum single-tire load as a useful
  // filter. Keep the shared query/API contract for other consumers, but never
  // let this retired URL field silently constrain selector results.
  minimumLoadKg: null,
  // Color and compound remain catalog dimensions and shared query fields, but
  // are intentionally not selector facets for the beginner-facing drawer.
  // Ignore legacy URL values so a hidden condition cannot silently constrain
  // the visible result set.
  colors: [],
  compounds: [],
})

const emptySelectorPage = (): SchwalbeTireCatalogSelectorPage => ({
  items: [],
  page: 1,
  page_size: SCHWALBE_CATALOG_PAGE_SIZE,
  total: 0,
  total_pages: 1,
  filter_options: {
    modelNames: [],
    nominalTireWidthsMm: [],
    wheelSizes: [],
    beadSeatDiametersMm: [],
    versionLabels: [],
    casingConstructions: [],
    compounds: [],
    colors: [],
    beads: [],
    seals: [],
    eBikeRatings: [],
  },
  rim_width_context: null,
})

export const useSchwalbeTireSelector = async () => {
  const route = useRoute()
  const router = useRouter()
  const { request } = useApiRequest()

  const initialSearch = readRouteSearch(route.query.search)
  const initialFilterState = parseSelectorRouteFilterState(
    route.query as Record<string, unknown>,
  )
  const filterState = ref<SchwalbeTireCatalogFilterQueryState>(initialFilterState)
  const searchInput = ref(initialSearch)
  const submittedSearch = ref(initialSearch)
  const updateFilterState = (patch: Partial<SchwalbeTireCatalogFilterQueryState>) => {
    filterState.value = { ...filterState.value, ...patch }
  }
  const applyFacetFilterState = (facetState: SchwalbeTireCatalogFacetFilterState) => {
    // Assign one complete state object so the route watcher performs a single
    // navigation and the catalog request is not restarted for every control.
    const wheelSizeKeys = [...facetState.wheelSizeKeys]
    const hasExplicitTireWidth = facetState.nominalTireWidthMinMm !== null
      || facetState.nominalTireWidthMaxMm !== null
      || facetState.nominalTireWidthsMm.length > 0
    const requestedInnerRimWidthMm = facetState.innerRimWidthMm
      !== null
      && Number.isFinite(facetState.innerRimWidthMm)
      && facetState.innerRimWidthMm > 0
      ? facetState.innerRimWidthMm
      : null

    // A concrete tire-width range and a rim-width match are two different
    // ways of choosing the tire's size. Keeping both would create a hidden AND
    // condition that is easy to trigger accidentally and often yields zero
    // rows. Choosing a width in the drawer therefore switches to width mode.
    // The same rule also makes the drawer's wheel-size and inner-width fields
    // one atomic choice: a rim-width match can only use exactly one wheel-size
    // pair, while a normal wheel-size facet may contain several pairs.
    const nextInnerRimWidthMm = hasExplicitTireWidth || wheelSizeKeys.length !== 1
      ? null
      : requestedInnerRimWidthMm

    updateFilterState({
      ...facetState,
      // Minimum single-tire load remains a legacy API field for other
      // consumers, but it is no longer exposed in this selector. Applying the
      // current drawer therefore retires any legacy URL condition.
      minimumLoadKg: null,
      nominalTireWidthsMm: [...facetState.nominalTireWidthsMm],
      wheelSizeKeys,
      innerRimWidthMm: nextInnerRimWidthMm,
      // A newly selected wheel-size pair is more precise than the legacy BSD
      // facet. Drop the hidden legacy constraint so an old shared link cannot
      // make a visibly selected wheel size return zero rows.
      beadSeatDiametersMm: wheelSizeKeys.length > 0 ? [] : [...facetState.beadSeatDiametersMm],
      casingConstructions: [...facetState.casingConstructions],
      beads: [...facetState.beads],
      seals: [...facetState.seals],
      eBikeRatings: [...facetState.eBikeRatings],
      colors: [],
      compounds: [],
    })
  }

  const getFacetFilterState = (): SchwalbeTireCatalogFacetFilterState => ({
    innerRimWidthMm: filterState.value.innerRimWidthMm,
    nominalTireWidthMinMm: filterState.value.nominalTireWidthMinMm,
    nominalTireWidthMaxMm: filterState.value.nominalTireWidthMaxMm,
    nominalTireWidthsMm: [...filterState.value.nominalTireWidthsMm],
    wheelSizeKeys: [...filterState.value.wheelSizeKeys],
    beadSeatDiametersMm: [...filterState.value.beadSeatDiametersMm],
    casingConstructions: [...filterState.value.casingConstructions],
    radialOnly: filterState.value.radialOnly,
    beads: [...filterState.value.beads],
    seals: [...filterState.value.seals],
    eBikeRatings: [...filterState.value.eBikeRatings],
  })
  const selectedModel = computed({
    get: () => filterState.value.modelName ?? 'ALL',
    set: (modelName: string) => updateFilterState({
      modelName: modelName === 'ALL' ? null : modelName,
    }),
  })
  const selectedInnerRimWidthMm = computed<number | null>({
    get: () => filterState.value.innerRimWidthMm,
    set: (innerRimWidthMm) => updateFilterState({
      innerRimWidthMm: innerRimWidthMm !== null
        && Number.isFinite(innerRimWidthMm)
        && innerRimWidthMm > 0
        ? innerRimWidthMm
        : null,
    }),
  })
  const selectedTireWidthsMm = computed({
    get: () => filterState.value.nominalTireWidthsMm,
    set: (nominalTireWidthsMm: number[]) => updateFilterState({ nominalTireWidthsMm }),
  })
  const selectedTireWidthMinMm = computed<number | null>({
    // A single legacy exact-width value can be represented losslessly as a
    // closed range, so old shared links remain visible in the new control.
    get: () => filterState.value.nominalTireWidthMinMm
      ?? (filterState.value.nominalTireWidthsMm.length === 1
        ? filterState.value.nominalTireWidthsMm[0] ?? null
        : null),
    set: (nominalTireWidthMinMm) => updateFilterState({
      nominalTireWidthMinMm: nominalTireWidthMinMm !== null
        && Number.isSafeInteger(nominalTireWidthMinMm)
        && nominalTireWidthMinMm > 0
        ? nominalTireWidthMinMm
        : null,
      nominalTireWidthsMm: [],
    }),
  })
  const selectedTireWidthMaxMm = computed<number | null>({
    get: () => filterState.value.nominalTireWidthMaxMm
      ?? (filterState.value.nominalTireWidthsMm.length === 1
        ? filterState.value.nominalTireWidthsMm[0] ?? null
        : null),
    set: (nominalTireWidthMaxMm) => updateFilterState({
      nominalTireWidthMaxMm: nominalTireWidthMaxMm !== null
        && Number.isSafeInteger(nominalTireWidthMaxMm)
        && nominalTireWidthMaxMm > 0
        ? nominalTireWidthMaxMm
        : null,
      nominalTireWidthsMm: [],
    }),
  })
  const selectedBeadSeatDiametersMm = computed({
    get: () => filterState.value.beadSeatDiametersMm,
    set: (beadSeatDiametersMm: number[]) => updateFilterState({ beadSeatDiametersMm }),
  })
  const selectedWheelSizeKeys = computed({
    get: () => filterState.value.wheelSizeKeys,
    set: (wheelSizeKeys: string[]) => updateFilterState({
      wheelSizeKeys,
      beadSeatDiametersMm: wheelSizeKeys.length > 0 ? [] : filterState.value.beadSeatDiametersMm,
    }),
  })
  const selectedCasingConstructions = computed({
    get: () => filterState.value.casingConstructions,
    set: (casingConstructions: string[]) => updateFilterState({ casingConstructions }),
  })
  const selectedRadialOnly = computed({
    get: () => filterState.value.radialOnly,
    set: (radialOnly: boolean) => updateFilterState({ radialOnly }),
  })
  const selectedBeads = computed({
    get: () => filterState.value.beads,
    set: (beads: string[]) => updateFilterState({ beads }),
  })
  const selectedSeals = computed({
    get: () => filterState.value.seals,
    set: (seals: string[]) => updateFilterState({ seals }),
  })
  const selectedEBikeRatings = computed({
    get: () => filterState.value.eBikeRatings,
    set: (eBikeRatings: (string | null)[]) => updateFilterState({ eBikeRatings }),
  })
  const sortBy = computed<SchwalbeCatalogSort>({
    get: () => filterState.value.sortBy,
    set: (sortBy) => updateFilterState({ sortBy }),
  })
  const requestedPage = computed(() => parsePage(route.query.page) || 1)
  const routeFilterState = computed(() => parseSelectorRouteFilterState(
    route.query as Record<string, unknown>,
  ))
  const selectorRequest = computed<SchwalbeTireCatalogSelectorRequest>(() => ({
    search: readRouteSearch(route.query.search),
    page: requestedPage.value,
    ...routeFilterState.value,
  }))

  const requestKey = computed(() => `schwalbe-tire-catalog-selector:${JSON.stringify(selectorRequest.value)}`)
  const { data, pending, error, refresh } = await useAsyncData<SchwalbeTireCatalogSelectorPage>(
    requestKey,
    () => fetchSchwalbeTireCatalogSelectorPage(request, selectorRequest.value),
    { default: emptySelectorPage },
  )

  const items = computed(() => data.value?.items || [])
  const filterOptions = computed(() => data.value?.filter_options || emptySelectorPage().filter_options)
  const modelOptions = computed(() => [
    'ALL',
    ...[...new Set([
      ...filterOptions.value.modelNames.map(option => option.value),
      ...(filterState.value.modelName === null ? [] : [filterState.value.modelName]),
    ])].sort((left, right) => left.localeCompare(right)),
  ])
  const totalItems = computed(() => data.value?.total || 0)
  const totalPages = computed(() => Math.max(1, data.value?.total_pages || 1))
  const rimWidthContext = computed(() => data.value?.rim_width_context ?? null)
  const currentPage = computed(() => Math.min(requestedPage.value, totalPages.value))
  const visibleItems = computed(() => items.value)

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
    filterState.value.modelName !== null
    || filterState.value.nominalTireWidthsMm.length > 0
    || filterState.value.nominalTireWidthMinMm !== null
    || filterState.value.nominalTireWidthMaxMm !== null
    || filterState.value.innerRimWidthMm !== null
    || filterState.value.wheelSizeKeys.length > 0
    || filterState.value.beadSeatDiametersMm.length > 0
    || filterState.value.minimumLoadKg !== null
    || filterState.value.casingConstructions.length > 0
    || filterState.value.radialOnly
    || filterState.value.beads.length > 0
    || filterState.value.seals.length > 0
    || filterState.value.eBikeRatings.length > 0
  ))

  let applyingRouteFilterState = false
  let routeFilterStateNavigationVersion = 0
  let isRouteFilterStateNavigationPending = false

  const updateRouteFilterState = () => {
    if (!import.meta.client || applyingRouteFilterState) return
    const query = mergeSchwalbeTireCatalogFilterQuery(
      route.query as Record<string, unknown>,
      filterState.value,
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
    () => JSON.stringify(filterState.value),
    updateRouteFilterState,
    { flush: 'sync' },
  )

  watch(
    () => JSON.stringify(routeFilterState.value),
    () => {
      applyingRouteFilterState = true
      filterState.value = routeFilterState.value
      applyingRouteFilterState = false
    },
    { flush: 'sync' },
  )

  const clearFacetFilters = () => {
    updateFilterState({
      innerRimWidthMm: null,
      nominalTireWidthMinMm: null,
      nominalTireWidthMaxMm: null,
      nominalTireWidthsMm: [],
      wheelSizeKeys: [],
      beadSeatDiametersMm: [],
      minimumLoadKg: null,
      casingConstructions: [],
      radialOnly: false,
      beads: [],
      seals: [],
      eBikeRatings: [],
      colors: [],
      compounds: [],
    })
  }

  const clearRimWidthSecondaryFilters = () => {
    const wheelSizeKeys = filterState.value.wheelSizeKeys.slice(0, 1)
    const preserveRimWidthMatch = filterState.value.innerRimWidthMm !== null
      && wheelSizeKeys.length === 1

    updateFilterState({
      innerRimWidthMm: preserveRimWidthMatch ? filterState.value.innerRimWidthMm : null,
      nominalTireWidthMinMm: null,
      nominalTireWidthMaxMm: null,
      nominalTireWidthsMm: [],
      wheelSizeKeys: preserveRimWidthMatch ? wheelSizeKeys : [],
      beadSeatDiametersMm: [],
      minimumLoadKg: null,
      casingConstructions: [],
      radialOnly: false,
      beads: [],
      seals: [],
      eBikeRatings: [],
      colors: [],
      compounds: [],
    })
  }

  watch(() => route.query.search, (value) => {
    const nextSearch = readRouteSearch(value)
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
    selectedInnerRimWidthMm,
    selectedTireWidthsMm,
    selectedTireWidthMinMm,
    selectedTireWidthMaxMm,
    selectedWheelSizeKeys,
    selectedBeadSeatDiametersMm,
    selectedCasingConstructions,
    selectedRadialOnly,
    selectedBeads,
    selectedSeals,
    selectedEBikeRatings,
    getFacetFilterState,
    applyFacetFilterState,
    sortBy,
    items,
    totalItems,
    rimWidthContext,
    visibleItems,
    modelOptions,
    tireWidthOptions: computed(() => filterOptions.value.nominalTireWidthsMm),
    wheelSizeOptions: computed(() => filterOptions.value.wheelSizes),
    casingConstructionOptions: computed(() => filterOptions.value.casingConstructions),
    beadOptions: computed(() => filterOptions.value.beads),
    sealOptions: computed(() => filterOptions.value.seals),
    eBikeRatingOptions: computed(() => filterOptions.value.eBikeRatings),
    hasActiveFilters,
    clearFacetFilters,
    clearRimWidthSecondaryFilters,
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
