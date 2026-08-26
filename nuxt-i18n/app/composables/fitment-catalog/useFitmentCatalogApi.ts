import { useApiRequest } from '~/composables/useApiRequest'
import type {
  FitmentCatalogEnvelope,
  FitmentCatalogListQuery,
  FitmentFrameEntriesResponse,
  FitmentFrameEntry,
  FitmentForkEntriesResponse,
  FitmentForkEntry,
  FitmentHubSpecification,
  FitmentHubSpecificationQuery,
  FitmentHubSpecificationsResponse,
} from '~/types/fitmentCatalog'

const fitmentCatalogPath = '/fitment-catalog'

type QueryValue = string | number | undefined

const appendQuery = (path: string, values: Record<string, QueryValue>) => {
  const params = new URLSearchParams()

  Object.entries(values).forEach(([key, value]) => {
    if (value === undefined || value === '') return
    params.set(key, String(value))
  })

  const query = params.toString()
  return query ? `${path}?${query}` : path
}

const listQueryValues = (query: FitmentCatalogListQuery) => ({
  search: query.search?.trim(),
  year: query.year === undefined || query.year === '' ? undefined : query.year,
  page: query.page,
  page_size: query.page_size,
})

const hubSpecificationQueryValues = (query: FitmentHubSpecificationQuery) => ({
  search: query.search?.trim(),
  position: query.position,
  axle_type: query.axle_type,
  page: query.page,
  page_size: query.page_size,
})

export function useFitmentCatalogApi() {
  const { request } = useApiRequest()

  const fetchFrameEntries = (query: FitmentCatalogListQuery = {}) => {
    return request<FitmentCatalogEnvelope<FitmentFrameEntriesResponse>>(
      appendQuery(`${fitmentCatalogPath}/frame-entries`, listQueryValues(query)),
      {},
      'Failed to load frame fitment records',
    )
  }

  const fetchFrameEntry = (id: number) => {
    return request<FitmentCatalogEnvelope<{ entry: FitmentFrameEntry }>>(
      `${fitmentCatalogPath}/frame-entries/${encodeURIComponent(String(id))}`,
      {},
      'Failed to load frame fitment record',
    )
  }

  const fetchForkEntries = (query: FitmentCatalogListQuery = {}) => {
    return request<FitmentCatalogEnvelope<FitmentForkEntriesResponse>>(
      appendQuery(`${fitmentCatalogPath}/fork-entries`, listQueryValues(query)),
      {},
      'Failed to load fork fitment records',
    )
  }

  const fetchForkEntry = (id: number) => {
    return request<FitmentCatalogEnvelope<{ entry: FitmentForkEntry }>>(
      `${fitmentCatalogPath}/fork-entries/${encodeURIComponent(String(id))}`,
      {},
      'Failed to load fork fitment record',
    )
  }

  const fetchHubSpecifications = (query: FitmentHubSpecificationQuery = {}) => {
    return request<FitmentCatalogEnvelope<FitmentHubSpecificationsResponse>>(
      appendQuery(`${fitmentCatalogPath}/hub-specifications`, hubSpecificationQueryValues(query)),
      {},
      'Failed to load hub specifications',
    )
  }

  const fetchHubSpecification = (id: number) => {
    return request<FitmentCatalogEnvelope<{ hub_specification: FitmentHubSpecification }>>(
      `${fitmentCatalogPath}/hub-specifications/${encodeURIComponent(String(id))}`,
      {},
      'Failed to load hub specification',
    )
  }

  return {
    fetchFrameEntries,
    fetchFrameEntry,
    fetchForkEntries,
    fetchForkEntry,
    fetchHubSpecifications,
    fetchHubSpecification,
  }
}
