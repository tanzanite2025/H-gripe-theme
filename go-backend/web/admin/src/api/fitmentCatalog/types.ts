export type FitmentYearMode = 'single' | 'range' | 'all' | 'unknown'
export type HubSpecificationPosition = 'front' | 'rear'
export type HubSpecificationAxleType = 'quick_release' | 'thru_axle' | 'bolt_on' | 'other'

export interface HubSpecification {
  id: number
  spec_code: string
  display_name: string
  position: HubSpecificationPosition
  axle_type: HubSpecificationAxleType
  axle_spacing_mm: number
  wr_mm: number | null
  wl_mm: number | null
  pcdr_mm: number | null
  pcdl_mm: number | null
  notes: string
  is_enabled: boolean
  sort_order: number
  frame_reference_count: number
  fork_reference_count: number
  created_at?: string
  updated_at?: string
}

export interface FrameFitmentEntry {
  id: number
  brand_name: string
  model_name: string
  series_name: string
  generation_name: string
  year_mode: FitmentYearMode
  year_from: number | null
  year_to: number | null
  market_code: string
  notes: string
  is_enabled: boolean
  sort_order: number
  hub_specifications: HubSpecification[]
  hub_specification_count: number
  created_at?: string
  updated_at?: string
}

export interface FrameFitmentEntryPayload {
  brand_name: string
  model_name: string
  series_name: string
  generation_name: string
  year_mode: FitmentYearMode
  year_from: number | null
  year_to: number | null
  market_code: string
  notes: string
  is_enabled: boolean
  sort_order: number
  hub_specification_ids: number[]
}

export interface ForkFitmentEntry {
  id: number
  brand_name: string
  model_name: string
  series_name: string
  generation_name: string
  year_mode: FitmentYearMode
  year_from: number | null
  year_to: number | null
  market_code: string
  notes: string
  is_enabled: boolean
  sort_order: number
  hub_specifications: HubSpecification[]
  hub_specification_count: number
  created_at?: string
  updated_at?: string
}

export interface ForkFitmentEntryPayload {
  brand_name: string
  model_name: string
  series_name: string
  generation_name: string
  year_mode: FitmentYearMode
  year_from: number | null
  year_to: number | null
  market_code: string
  notes: string
  is_enabled: boolean
  sort_order: number
  hub_specification_ids: number[]
}

export interface HubSpecificationPayload {
  spec_code: string
  display_name: string
  position: HubSpecificationPosition
  axle_type: HubSpecificationAxleType
  axle_spacing_mm: number
  wr_mm: number | null
  wl_mm: number | null
  pcdr_mm: number | null
  pcdl_mm: number | null
  notes: string
  is_enabled: boolean
  sort_order: number
}

export interface FitmentCatalogPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface FrameFitmentEntryListPayload {
  entries: FrameFitmentEntry[]
  pagination: FitmentCatalogPagination
}

export interface ForkFitmentEntryListPayload {
  entries: ForkFitmentEntry[]
  pagination: FitmentCatalogPagination
}

export interface HubSpecificationListPayload {
  specifications: HubSpecification[]
  pagination: FitmentCatalogPagination
}
