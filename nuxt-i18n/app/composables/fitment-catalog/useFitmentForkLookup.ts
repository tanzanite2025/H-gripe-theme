import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import { useFitmentCatalogApi } from '~/composables/fitment-catalog/useFitmentCatalogApi'
import { toUserFacingApiError } from '~/utils/storefrontApiFailures'
import type {
  FitmentCatalogPagination,
  FitmentForkEntry,
} from '~/types/fitmentCatalog'

const initialPagination = (): FitmentCatalogPagination => ({
  page: 1,
  page_size: 10,
  total: 0,
  total_pages: 0,
})

export function useFitmentForkLookup() {
  const { t } = useI18n()
  const { fetchForkEntries } = useFitmentCatalogApi()
  const searchQuery = ref('')
  const year = ref<number | null>(null)
  const entries = ref<FitmentForkEntry[]>([])
  const pagination = ref<FitmentCatalogPagination>(initialPagination())
  const isSearching = ref(false)
  const error = ref('')
  const hasSearched = ref(false)
  let requestSequence = 0

  const currentPage = computed(() => pagination.value.page)
  const canGoPrevious = computed(() => currentPage.value > 1 && !isSearching.value)
  const canGoNext = computed(() => (
    currentPage.value < pagination.value.total_pages && !isSearching.value
  ))

  const load = async (page = 1) => {
    const requestId = ++requestSequence
    isSearching.value = true
    error.value = ''
    hasSearched.value = true

    try {
      const response = await fetchForkEntries({
        search: searchQuery.value.trim() || undefined,
        year: year.value && Number.isFinite(year.value) ? year.value : undefined,
        page,
        page_size: pagination.value.page_size,
      })
      if (requestId !== requestSequence) return

      entries.value = response.data?.fork_entries || []
      pagination.value = response.data?.pagination || initialPagination()
    } catch (cause) {
      if (requestId !== requestSequence) return
      entries.value = []
      pagination.value = initialPagination()
      error.value = toUserFacingApiError(
        cause,
        t(
          'fitmentCatalog.states.unavailable',
          'Fitment data is temporarily unavailable. Please try again.',
        ),
      )
    } finally {
      if (requestId === requestSequence) {
        isSearching.value = false
      }
    }
  }

  const submit = () => load(1)

  const clear = () => {
    requestSequence += 1
    searchQuery.value = ''
    year.value = null
    entries.value = []
    pagination.value = initialPagination()
    isSearching.value = false
    error.value = ''
    hasSearched.value = false
  }

  const goToPage = (page: number) => {
    if (page < 1 || page > pagination.value.total_pages || isSearching.value) return
    return load(page)
  }

  return {
    searchQuery,
    year,
    entries,
    pagination,
    currentPage,
    isSearching,
    error,
    hasSearched,
    canGoPrevious,
    canGoNext,
    submit,
    clear,
    goToPage,
  }
}
