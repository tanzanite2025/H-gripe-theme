import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'

export type InnerTubeValveFitmentMode = 'automatic' | 'manual'
export type InnerTubeValveFitmentStatus = 'optimal' | 'marginal' | 'unsafe'
export type InnerTubeValveStructure = 'removableCore' | 'fixedCore' | 'rvcExtender' | 'rootProtection'
export type InnerTubeRecommendationReasonKey =
  | 'shortestNativeValveMeetsMinimum'
  | 'shortestExtenderAssemblyReachesMinimum'
  | 'manualCombination'
export type InnerTubeRecommendationKey =
  | 'shortValve'
  | 'standardValve'
  | 'mediumValve'
  | 'highRimValve'
  | 'mediumValveWithExtender'
  | 'highRimValveWithExtender'

export interface InnerTubeValveFitmentModelSource {
  name: string
  provenance: string
}

export interface InnerTubeValveFitmentRecommendation {
  valve_length_mm: number
  extender_length_mm: number
  total_assembly_length_mm: number
  minimum_required_length_mm: number
  preferred_minimum_length_mm: number
  recommendation_key: InnerTubeRecommendationKey
  recommendation_reason_key: InnerTubeRecommendationReasonKey
}

export interface InnerTubeValveFitmentResult {
  model_version: string
  knowledge_as_of: string
  rim_depth_mm: number
  outer_lip_offset_mm: number
  passage_depth_mm: number
  minimum_required_length_mm: number
  preferred_minimum_length_mm: number
  minimum_length_margin_mm: number
  preferred_length_margin_mm: number
  valve_length_mm: number
  extender_length_mm: number
  total_assembly_length_mm: number
  effective_exposure_mm: number
  status: InnerTubeValveFitmentStatus
  recommendation_key: InnerTubeRecommendationKey
  recommendation_reason_key: InnerTubeRecommendationReasonKey
  recommendation_reason: string
  recommendation: InnerTubeValveFitmentRecommendation
  alternatives: InnerTubeValveFitmentAlternative[]
  pump_head_grip_depth_mm: number
  preferred_exposure_mm: number
  rim_depth_uncertainty_mm: number
  uncertainty_review: InnerTubeValveFitmentUncertaintyReview | null
  calculation_method: string
  source: InnerTubeValveFitmentModelSource
  limitations: string[]
}

export interface InnerTubeValveFitmentUncertaintyBoundary {
  rim_depth_mm: number
  passage_depth_mm: number
  effective_exposure_mm: number
  minimum_required_length_mm: number
  minimum_length_margin_mm: number
  status: InnerTubeValveFitmentStatus
}

export interface InnerTubeValveFitmentUncertaintyReview {
  rim_depth_uncertainty_mm: number
  lower_bound: InnerTubeValveFitmentUncertaintyBoundary
  upper_bound: InnerTubeValveFitmentUncertaintyBoundary
  is_status_stable: boolean
  worst_case_status: InnerTubeValveFitmentStatus
}

export interface InnerTubeValveFitmentClearanceCell {
  valve_length_mm: number
  effective_exposure_mm: number
  status: InnerTubeValveFitmentStatus
}

export interface InnerTubeValveFitmentAlternative {
  valve_length_mm: number
  extender_length_mm: number
  total_assembly_length_mm: number
  effective_exposure_mm: number
  minimum_length_margin_mm: number
  preferred_length_margin_mm: number
  status: InnerTubeValveFitmentStatus
  is_automatic_recommendation: boolean
}

export interface InnerTubeValveFitmentMatrixRow {
  rim_depth_mm: number
  passage_depth_mm: number
  minimum_required_length_mm: number
  clearances: InnerTubeValveFitmentClearanceCell[]
  recommended_result: InnerTubeValveFitmentResult
}

export interface InnerTubeValveFitmentMatrixMetadata {
  model_version: string
  knowledge_as_of: string
  rim_depth_min_mm: number
  rim_depth_max_mm: number
  outer_lip_offset_mm: number
  minimum_safe_exposure_mm: number
  preferred_exposure_mm: number
  pump_head_grip_depth_min_mm: number
  pump_head_grip_depth_max_mm: number
  default_pump_head_grip_depth_mm: number
  rim_depth_uncertainty_max_mm: number
  default_rim_depth_uncertainty_mm: number
  base_valve_length_options_mm: number[]
  extender_length_options_mm: number[]
  preset_rim_depths_mm: number[]
  rows: InnerTubeValveFitmentMatrixRow[]
  calculation_method: string
  source: InnerTubeValveFitmentModelSource
  limitations: string[]
}

export interface InnerTubeValveFitmentSolveInput {
  rim_depth_mm: number
  mode: InnerTubeValveFitmentMode
  valve_length_mm?: number
  extender_length_mm?: number
  pump_head_grip_depth_mm?: number
  rim_depth_uncertainty_mm?: number
}

interface InnerTubeValveFitmentAPIEnvelope<T> {
  data?: T
}

const innerTubeValveFitmentMatrixEndpoint = '/engineering/inner-tube-fitment/matrix'
const innerTubeValveFitmentSolveEndpoint = '/engineering/inner-tube-fitment/solve'

export const useInnerTubeValveFitmentCalculator = async () => {
  const { request } = useApiRequest()
  const { data: matrixEnvelope, pending: matrixPending, error: matrixError, refresh: refreshMatrix } = await useAsyncData<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentMatrixMetadata>>(
    'inner-tube-valve-fitment-engineering-matrix-v1',
    async () => request<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentMatrixMetadata>>(
      innerTubeValveFitmentMatrixEndpoint,
      {},
      'Inner-tube valve fitment matrix is temporarily unavailable',
    ),
    { default: () => ({}) },
  )

  const metadata = computed(() => matrixEnvelope.value?.data ?? null)
  const defaultResult = computed(() => (
    metadata.value?.rows.find(row => row.rim_depth_mm === 50)?.recommended_result ?? null
  ))

  const solveInnerTubeValveFitmentWithBackend = async (input: InnerTubeValveFitmentSolveInput) => {
    const response = await request<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentResult>>(
      innerTubeValveFitmentSolveEndpoint,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      },
      'Inner-tube valve fitment calculation is temporarily unavailable',
    )
    if (!response.data) {
      throw new Error('Inner-tube valve fitment response is missing data')
    }
    return response.data
  }

  return {
    metadata,
    defaultResult,
    matrixPending,
    matrixError,
    refreshMatrix,
    solveInnerTubeValveFitmentWithBackend,
  }
}
