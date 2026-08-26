export type FitmentCatalogYearMode = 'single' | 'range' | 'all' | 'unknown'

export type FitmentHubPosition = 'front' | 'rear'

export type FitmentHubAxleType =
  | 'quick_release'
  | 'thru_axle'
  | 'bolt_on'
  | 'other'

export interface FitmentCatalogPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface FitmentHubSpecification {
  id: number
  spec_code: string
  display_name: string
  position: FitmentHubPosition
  axle_type: FitmentHubAxleType
  axle_spacing_mm: number
  notes?: string
}

export interface FitmentFrameEntry {
  id: number
  brand_name: string
  model_name: string
  series_name?: string
  generation_name?: string
  year_mode: FitmentCatalogYearMode
  year_from?: number | null
  year_to?: number | null
  market_code?: string
  notes?: string
  hub_specifications: FitmentHubSpecification[]
  hub_specification_count: number
}

export interface FitmentForkEntry {
  id: number
  brand_name: string
  model_name: string
  series_name?: string
  generation_name?: string
  year_mode: FitmentCatalogYearMode
  year_from?: number | null
  year_to?: number | null
  market_code?: string
  notes?: string
  hub_specifications: FitmentHubSpecification[]
  hub_specification_count: number
}

export interface FitmentFrameEntriesResponse {
  frame_entries: FitmentFrameEntry[]
  pagination: FitmentCatalogPagination
}

export interface FitmentForkEntriesResponse {
  fork_entries: FitmentForkEntry[]
  pagination: FitmentCatalogPagination
}

export interface FitmentHubSpecificationsResponse {
  hub_specifications: FitmentHubSpecification[]
  pagination: FitmentCatalogPagination
}

export interface FitmentCatalogEnvelope<T> {
  code: number
  data: T
}

export interface FitmentCatalogListQuery {
  search?: string
  year?: number | string
  page?: number
  page_size?: number
}

export interface FitmentHubSpecificationQuery {
  search?: string
  position?: FitmentHubPosition
  axle_type?: FitmentHubAxleType
  page?: number
  page_size?: number
}

export type FitmentCatalogResourceType = 'frame' | 'fork'
