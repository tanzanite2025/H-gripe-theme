export type TireRimReferenceRimSystem = 'hookless' | 'hooked'

export interface TireRimWidthRange {
  min: number
  max: number
}

export interface TireRimWidthReferenceCalculation {
  method: string
  lower_tire_width_mm: number
  upper_tire_width_mm: number
  interpolation_ratio: number
}

export interface TireRimWidthReferenceSourceRow {
  tire_width_mm: number
  kind: 'recommended' | 'possible_reference' | string
}

export interface TireRimWidthReferenceSuggestion {
  tire_width_mm: number
  rim_system: TireRimReferenceRimSystem
  result_kind:
    | 'exact_recommended'
    | 'possible_reference'
    | 'interpolated'
    | 'engineering_recommended'
    | 'engineering_reference'
    | string
  rim_width_ranges: TireRimWidthRange[]
  source_rows: TireRimWidthReferenceSourceRow[]
  calculation?: TireRimWidthReferenceCalculation
  derived_metrics: {
    status: string
    inflated_tire_width_mm: TireRimWidthRange
    aero_target_outer_width_mm: TireRimWidthRange
  }
  model_version: string
  knowledge_as_of: string
  limitations: string[]
}

export interface TireRimWidthReferenceMatrixRow {
  rim_system: TireRimReferenceRimSystem
  tire_width_mm: number
  inch: string
  recommended: TireRimWidthRange[]
  possible: TireRimWidthRange[]
}

export interface TireRimWidthReferenceMetadata {
  model_version: string
  knowledge_as_of: string
  source: {
    name: string
    provenance: string
  }
  rows: TireRimWidthReferenceMatrixRow[]
  methodology: {
    exact_rows: string
    interpolated_rows: string
    possible_rows: string
    derived_metrics: string
  }
  limitations: string[]
}

export const formatTireRimWidthReferenceRange = ({ min, max }: TireRimWidthRange): string => (
  min === max ? String(min) : String(min) + '-' + String(max)
)

export const formatTireRimWidthReferenceRanges = (values: TireRimWidthRange[]): string => (
  values.map(formatTireRimWidthReferenceRange).join(', ')
)
