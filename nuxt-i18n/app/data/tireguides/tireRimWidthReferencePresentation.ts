export type TireRimReferenceRimSystem = 'hookless' | 'hooked'

export type TireRimApplicationProfile =
  | 'automatic_by_width'
  | 'road_aero'
  | 'gravel_all_road'
  | 'mtb_xc_trail'
  | 'mtb_plus'
  | 'fat_snow'

export interface TireRimStandardReference {
  iso_5775_part_1: string
  iso_5775_part_2: string
  etrto_edition: string
  dt_swiss_used_by_calculation: boolean
  dt_swiss_reference_role: string
}

export interface TireRimFitmentRuleReference {
  normal_range_basis: string
  recommended_range_basis: string
  manufacturer_source_urls: string[]
  interpolated: boolean
}

export interface TireRimNormalFitmentTableReference {
  standard: string
  clause: string
  table: string
  source_kind: string
  scope: string
}

export interface TireRimBestFitmentResponse {
  rim_inner_width_mm: number
  tire_width_mm: number
  application_profile: TireRimApplicationProfile
  selection_method: string
  reason_code: string
  model_status: string
  target_tire_width_mm: number
  target_tire_width_difference_mm: number
  fitment_rule_reference: TireRimFitmentRuleReference
}

export interface TireRimBestFitment {
  rimInnerWidthMm: number
  tireWidthMm: number
  applicationProfile: TireRimApplicationProfile
  selectionMethod: string
  reasonCode: string
  modelStatus: string
  targetTireWidthMm: number
  targetTireWidthDifferenceMm: number
  fitmentRuleReference: TireRimFitmentRuleReference
}

/** Preserve an absent backend recommendation; never derive one from ranges. */
export const mapTireRimBackendBestFitmentForPresentation = (
  bestFitment: TireRimBestFitmentResponse | null,
): TireRimBestFitment | null => bestFitment === null ? null : ({
  rimInnerWidthMm: bestFitment.rim_inner_width_mm,
  tireWidthMm: bestFitment.tire_width_mm,
  applicationProfile: bestFitment.application_profile,
  selectionMethod: bestFitment.selection_method,
  reasonCode: bestFitment.reason_code,
  modelStatus: bestFitment.model_status,
  targetTireWidthMm: bestFitment.target_tire_width_mm,
  targetTireWidthDifferenceMm: bestFitment.target_tire_width_difference_mm,
  fitmentRuleReference: bestFitment.fitment_rule_reference,
})

export interface TireRimWidthRange {
  min: number
  max: number
}

export interface TireRimOuterWidthEstimate {
  rim_inner_width_mm: number
  wall_thickness_per_side_estimate_mm: number
  rim_outer_width_estimate_mm: number
}

export interface TireRimEngineeringCalculationResponse {
  normal_fitment_reference?: TireRimNormalFitmentTableReference
  model_version: string
  verdict: 'critical' | 'recommended' | 'reference'
  verdict_reason: string
  rim_system: TireRimReferenceRimSystem
  rim_inner_width_mm: number
  tire_width_mm: number
  application_profile: TireRimApplicationProfile
  application_profile_selection: string
  standard_reference: TireRimStandardReference
  fitment_rule_reference: TireRimFitmentRuleReference
  allowed_tire_width_range: TireRimWidthRange
  recommended_tire_width_range: TireRimWidthRange | null
  aero_rule_applies: boolean
  aero_target_outer_width_mm?: number
  aero_target_tire_width_mm?: number
  maximum_pressure: {
    bar: number
    psi: number
  } | null
  tube_recommendation: {
    band: string
    width_min_mm: number
    width_max_mm: number
  }
}

/** Engineering result fields used by the Vue workbench after API mapping. */
export interface TireRimEngineeringCalculation {
  normalFitmentReference?: TireRimNormalFitmentTableReference
  modelVersion: string
  verdict: 'critical' | 'recommended' | 'reference'
  verdictReason: string
  rimSystem: TireRimReferenceRimSystem
  rimInnerWidthMm: number
  tireWidthMm: number
  applicationProfile: TireRimApplicationProfile
  applicationProfileSelection: string
  standardReference: TireRimStandardReference
  fitmentRuleReference: TireRimFitmentRuleReference
  allowedTireWidthRange: TireRimWidthRange
  recommendedTireWidthRange: TireRimWidthRange | null
  aeroRuleApplies: boolean
  aeroTargetOuterWidthMm: number | null
  aeroTargetTireWidthMm: number | null
  maximumPressure: {
    bar: number
    psi: number
  } | null
  tubeRecommendation: {
    band: string
    widthMinMm: number
    widthMaxMm: number
  }
}

export interface TireRimReverseTireWidthMatch {
  tire_width_mm: number
  target_tire_width_difference_mm: number
}

export interface TireRimReverseSuggestion {
  normal_fitment_reference?: TireRimNormalFitmentTableReference
  rim_inner_width_mm: number
  rim_system: TireRimReferenceRimSystem
  application_profile: TireRimApplicationProfile
  resolved_application_profile: TireRimApplicationProfile
  application_profile_selection: string
  best_fitment: TireRimBestFitmentResponse | null
  standard_reference: TireRimStandardReference
  fitment_rule_reference: TireRimFitmentRuleReference
  aero_rule_applies: boolean
  aero_target_tire_width_mm?: number
  best_match?: TireRimReverseTireWidthMatch
  normal_fitment_tire_width_range: TireRimWidthRange
  recommended_fitment_tire_width_range: TireRimWidthRange | null
  rim_outer_width_estimate: TireRimOuterWidthEstimate
  model_version: string
  knowledge_as_of: string
  limitations: string[]
}

export interface TireRimWidthReferenceSuggestion {
  normal_fitment_reference?: TireRimNormalFitmentTableReference
  tire_width_mm: number
  rim_system: TireRimReferenceRimSystem
  application_profile: TireRimApplicationProfile
  application_profile_selection: string
  best_fitment: TireRimBestFitmentResponse | null
  normal_fitment_rim_width_ranges: TireRimWidthRange[]
  standard_reference: TireRimStandardReference
  result_kind:
    | 'engineering_recommended'
    | 'engineering_reference'
    | 'engineering_normal_fitment'
    | 'engineering_exact'
    | string
  aero_rule_applies: boolean
  rim_width_ranges: TireRimWidthRange[]
  recommended_rim_width_ranges?: TireRimWidthRange[]
  rim_outer_width_estimates?: TireRimOuterWidthEstimate[]
  source_rows?: Array<{ tire_width_mm: number; kind: string }>
  calculation?: {
    method: string
    lower_tire_width_mm: number
    upper_tire_width_mm: number
    interpolation_ratio: number
  }
  derived_metrics: {
    status: string
    tire_width_mm: TireRimWidthRange
    aero_rule_applies: boolean
    aero_target_outer_width_mm?: TireRimWidthRange
  }
  model_version: string
  knowledge_as_of: string
  limitations: string[]
  engineering?: TireRimEngineeringCalculationResponse
}

export const formatTireRimWidthReferenceRange = ({ min, max }: TireRimWidthRange): string => (
  min === max ? String(min) : String(min) + '-' + String(max)
)

export const formatTireRimWidthReferenceRanges = (values: TireRimWidthRange[]): string => (
  values.map(formatTireRimWidthReferenceRange).join(', ')
)
