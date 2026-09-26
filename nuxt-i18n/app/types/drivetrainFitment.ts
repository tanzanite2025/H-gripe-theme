export type FreehubStandard =
  | 'HG-11'
  | 'HG'
  | 'HG-L2'
  | 'XDR'
  | 'XD'
  | 'MICRO_SPLINE'
  | 'N3W'
  | 'CAMPY_CLASSIC'

export interface DrivetrainSpacerPart {
  thickness_mm: number
  part_code?: string
  position?: string
  description: string
}

export interface DrivetrainSpacerRequirement {
  required: boolean
  thickness_mm: number
  part_code?: string
  position?: string
  description: string
  parts?: DrivetrainSpacerPart[]
}

export interface DrivetrainFitmentOption {
  standard: FreehubStandard
  display_name: string
  spacer: DrivetrainSpacerRequirement
  image_src: string
  notes?: string
}

export interface DrivetrainCassetteRule {
  rule_id: string
  brand: string
  cassette_spec: string
  display_name: string
  hint_groupsets: string
  speed: number
  min_cog_teeth: number
  max_cog_teeth: number
  recommended_freehub: FreehubStandard
  fitment_options: DrivetrainFitmentOption[]
  image_src: string
  mechanical_notes: string
  mechanical_fact?: string
  rule_version: string
}

export interface DrivetrainMatrixResponse {
  rules: DrivetrainCassetteRule[]
  rule_version: string
  knowledge_as_of: string
}

export interface DrivetrainApiEnvelope<T> {
  code: number | string
  data: T
  message?: string
}

export interface DrivetrainCalculateRequest {
  brand: string
  cassette_spec: string
  freehub_standard?: FreehubStandard
}

export interface DrivetrainCalculationResponse extends DrivetrainCassetteRule {
  selected_freehub?: DrivetrainFitmentOption
  spacer: DrivetrainSpacerRequirement
}
