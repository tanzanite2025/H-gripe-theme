import { computed } from 'vue'
import { useState } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import type { SpokeRecordedResult } from '~/data/spoke-calculator/database'
import { normalizeSpokeRecordedResultsPayload } from '~/utils/spokeCatalogNormalizer'

type RecordedResultsSource = 'empty' | 'api' | 'error'

interface RecordedResultsState {
  presets: SpokeRecordedResult[]
  loading: boolean
  error: string | null
  source: RecordedResultsSource
}

const createEmptyState = (): RecordedResultsState => ({
  presets: [],
  loading: false,
  error: null,
  source: 'empty',
})

const inFlightLoads: Record<string, Promise<SpokeRecordedResult[]> | undefined> = {}

/**
 * Reads backend-entered spoke lengths for the lower recorded-result search.
 * This state is intentionally separate from the manual calculator catalog and
 * never writes to the wizard draft or its geometry inputs.
 */
export const useSpokeCalculatorRecordedResults = () => {
  const { baseURL, request: apiRequest } = useApiRequest()
  const state = useState<RecordedResultsState>(
    `spoke-calculator-recorded-results:${baseURL}`,
    createEmptyState,
  )

  const applyError = (reason: string): SpokeRecordedResult[] => {
    state.value.presets = []
    state.value.source = 'error'
    state.value.error = reason
    return state.value.presets
  }

  const loadResults = async (): Promise<SpokeRecordedResult[]> => {
    if (state.value.loading && inFlightLoads[baseURL]) {
      return inFlightLoads[baseURL]!
    }

    state.value.loading = true
    state.value.error = null

    const loadRequest = apiRequest<unknown>(
      '/spoke/catalog/results',
      {},
      'Failed to load recorded spoke results',
    )
      .then((payload) => {
        const results = normalizeSpokeRecordedResultsPayload(payload).presets
        state.value.presets = results
        state.value.source = 'api'
        return results
      })
      .catch((error: any) => {
        const message = error?.data?.message || error?.message || 'Failed to load recorded spoke results.'
        return applyError(message)
      })
      .finally(() => {
        state.value.loading = false
        inFlightLoads[baseURL] = undefined
      })

    inFlightLoads[baseURL] = loadRequest
    return loadRequest
  }

  if (!state.value.loading && state.value.source === 'empty') {
    void loadResults()
  }

  return {
    presets: computed(() => state.value.presets),
    loading: computed(() => state.value.loading),
    error: computed(() => state.value.error),
    source: computed(() => state.value.source),
    loadResults,
  }
}
