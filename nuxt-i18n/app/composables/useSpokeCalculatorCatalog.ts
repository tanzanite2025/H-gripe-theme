import { computed } from 'vue'
import { useState } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import {
  DEFAULT_SPOKE_CATALOG,
  SPOKE_CALCULATOR_OPTIONS,
  type SpokeCatalog,
} from '~/data/spoke-calculator/database'
import { normalizeSpokeCatalogPayload } from '../utils/spokeCatalogNormalizer'

type SpokeCatalogSource = 'empty' | 'api' | 'dev-fallback' | 'error'

interface SpokeCatalogState {
  catalog: SpokeCatalog
  loading: boolean
  error: string | null
  source: SpokeCatalogSource
}

const emptyCatalog = (): SpokeCatalog => ({
  options: SPOKE_CALCULATOR_OPTIONS,
  rims: [],
  hubs: [],
  presets: [],
})

const createEmptyState = (): SpokeCatalogState => ({
  catalog: emptyCatalog(),
  loading: false,
  error: null,
  source: 'empty',
})

const inFlightLoads: Record<string, Promise<SpokeCatalog> | undefined> = {}

export const useSpokeCalculatorCatalog = () => {
  const { baseURL, request: apiRequest } = useApiRequest()
  const state = useState<SpokeCatalogState>(`spoke-calculator-catalog:${baseURL}`, createEmptyState)

  const applyDevFallback = (reason: string): SpokeCatalog => {
    if (!import.meta.dev) {
      state.value.catalog = emptyCatalog()
      state.value.source = 'error'
      state.value.error = reason
      return state.value.catalog
    }

    // eslint-disable-next-line no-console
    console.warn(`[spoke catalog] using development fallback: ${reason}`)
    state.value.catalog = DEFAULT_SPOKE_CATALOG
    state.value.source = 'dev-fallback'
    state.value.error = null
    return state.value.catalog
  }

  const loadCatalog = async (): Promise<SpokeCatalog> => {
    if (state.value.loading && inFlightLoads[baseURL]) {
      return inFlightLoads[baseURL]!
    }

    state.value.loading = true
    state.value.error = null

    const loadRequest = apiRequest<unknown>('/spoke/export', {}, 'Failed to load spoke calculator data')
      .then((payload) => {
        const catalog = normalizeSpokeCatalogPayload(payload)
        state.value.catalog = catalog
        state.value.source = 'api'
        return catalog
      })
      .catch((error: any) => {
        const message = error?.data?.message || error?.message || 'Failed to load spoke calculator data.'
        return applyDevFallback(message)
      })
      .finally(() => {
        state.value.loading = false
        inFlightLoads[baseURL] = undefined
      })

    inFlightLoads[baseURL] = loadRequest
    return loadRequest
  }

  if (!state.value.loading && state.value.source === 'empty') {
    void loadCatalog()
  }

  return {
    catalog: computed(() => state.value.catalog),
    options: computed(() => state.value.catalog.options),
    rims: computed(() => state.value.catalog.rims),
    hubs: computed(() => state.value.catalog.hubs),
    presets: computed(() => state.value.catalog.presets),
    loading: computed(() => state.value.loading),
    error: computed(() => state.value.error),
    source: computed(() => state.value.source),
    loadCatalog,
  }
}
