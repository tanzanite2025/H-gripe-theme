type InnerTubeValveFitmentFallbackRecommendationKey =
  | 'shortValve'
  | 'standardValve'
  | 'mediumValve'
  | 'highRimValve'
  | 'mediumValveWithExtender'
  | 'highRimValveWithExtender'
type InnerTubeValveFitmentFallbackRecommendationReasonKey =
  | 'shortestNativeValveMeetsMinimum'
  | 'shortestExtenderAssemblyReachesMinimum'
type InnerTubeValveFitmentFallbackStatus = 'optimal' | 'marginal' | 'unsafe'

interface InnerTubeValveFitmentFallbackRecommendationData {
  valve_length_mm: number
  extender_length_mm: number
  total_assembly_length_mm: number
  minimum_required_length_mm: number
  preferred_minimum_length_mm: number
  recommendation_key: InnerTubeValveFitmentFallbackRecommendationKey
  recommendation_reason_key: InnerTubeValveFitmentFallbackRecommendationReasonKey
}

interface InnerTubeValveFitmentFallbackMatrixRecommendationSummary {
  recommendation_key: InnerTubeValveFitmentFallbackRecommendationKey
  recommendation: InnerTubeValveFitmentFallbackRecommendationData
}

interface InnerTubeValveFitmentFallbackMatrixRow {
  rim_depth_mm: number
  passage_depth_mm: number
  minimum_required_length_mm: number
  clearances: Array<{
    valve_length_mm: number
    effective_exposure_mm: number
    status: InnerTubeValveFitmentFallbackStatus
  }>
  recommended_result: InnerTubeValveFitmentFallbackMatrixRecommendationSummary
}

interface InnerTubeValveFitmentServerRenderedMatrixFallbackMetadata {
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
  rows: InnerTubeValveFitmentFallbackMatrixRow[]
  calculation_method: string
  source: { name: string; provenance: string }
  limitations: string[]
}

// Keep these static fallback assumptions aligned with go-backend/internal/domain/innertubefitment/inner_tube_valve_fitment_engine.go.
const innerTubeValveFitmentFallbackRimDepthPresetsMillimetres = [28, 35, 40, 45, 50, 55, 60, 70, 75, 80, 85, 90, 95, 100]
const innerTubeValveFitmentFallbackBaseValveLengthsMillimetres = [40, 48, 60, 80]
const innerTubeValveFitmentFallbackExtenderLengthsMillimetres = [0, 20, 30, 40, 60]
const innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres = 15
const innerTubeValveFitmentFallbackOuterLipOffsetMillimetres = 6.5

interface InnerTubeValveFitmentFallbackAssemblyRecommendation {
  valveLengthMillimetres: number
  extenderLengthMillimetres: number
  recommendationKey: InnerTubeValveFitmentFallbackRecommendationKey
  recommendationReasonKey: InnerTubeValveFitmentFallbackRecommendationReasonKey
}

const getServerRenderedInnerTubeValveFitmentRecommendationKey = (
  valveLengthMillimetres: number,
  extenderLengthMillimetres: number,
): InnerTubeValveFitmentFallbackRecommendationKey => {
  if (extenderLengthMillimetres === 0 && valveLengthMillimetres <= 40) return 'shortValve'
  if (extenderLengthMillimetres === 0 && valveLengthMillimetres <= 48) return 'standardValve'
  if (extenderLengthMillimetres === 0 && valveLengthMillimetres <= 60) return 'mediumValve'
  if (extenderLengthMillimetres === 0) return 'highRimValve'
  if (valveLengthMillimetres <= 60) return 'mediumValveWithExtender'
  return 'highRimValveWithExtender'
}

const getServerRenderedInnerTubeValveFitmentRecommendation = (
  rimDepthMillimetres: number,
): InnerTubeValveFitmentFallbackAssemblyRecommendation => {
  const minimumRequiredLengthMillimetres = rimDepthMillimetres + innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres
  const preferredMinimumLengthMillimetres = minimumRequiredLengthMillimetres + 5

  const shortestNativeValveLength = innerTubeValveFitmentFallbackBaseValveLengthsMillimetres
    .find(valveLength => valveLength >= minimumRequiredLengthMillimetres)
  if (shortestNativeValveLength !== undefined) {
    return {
      valveLengthMillimetres: shortestNativeValveLength,
      extenderLengthMillimetres: 0,
      recommendationKey: getServerRenderedInnerTubeValveFitmentRecommendationKey(shortestNativeValveLength, 0),
      recommendationReasonKey: 'shortestNativeValveMeetsMinimum',
    }
  }

  const supportedAssemblies = innerTubeValveFitmentFallbackBaseValveLengthsMillimetres
    .flatMap(valveLength => innerTubeValveFitmentFallbackExtenderLengthsMillimetres.map((extenderLength) => ({
      valveLength,
      extenderLength,
      totalLength: valveLength + extenderLength,
    })))
    .filter(assembly => assembly.totalLength >= minimumRequiredLengthMillimetres)
    .sort((left, right) => left.totalLength - right.totalLength
      || left.extenderLength - right.extenderLength
      || right.valveLength - left.valveLength)
  const shortestSupportedAssembly = supportedAssemblies[0]
  if (!shortestSupportedAssembly) {
    throw new Error(`No server-rendered valve and extender fallback reaches ${minimumRequiredLengthMillimetres} mm`)
  }

  return {
    valveLengthMillimetres: shortestSupportedAssembly.valveLength,
    extenderLengthMillimetres: shortestSupportedAssembly.extenderLength,
    recommendationKey: getServerRenderedInnerTubeValveFitmentRecommendationKey(
      shortestSupportedAssembly.valveLength,
      shortestSupportedAssembly.extenderLength,
    ),
    recommendationReasonKey: 'shortestExtenderAssemblyReachesMinimum',
  }
}

const getServerRenderedInnerTubeValveFitmentStatus = (exposureMillimetres: number): InnerTubeValveFitmentFallbackStatus => {
  if (exposureMillimetres >= innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres + 3) return 'optimal'
  if (exposureMillimetres >= innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres) return 'marginal'
  return 'unsafe'
}

const createServerRenderedInnerTubeValveFitmentMatrixRecommendationSummary = (
  rimDepthMillimetres: number,
): InnerTubeValveFitmentFallbackMatrixRecommendationSummary => {
  const minimumRequiredLengthMillimetres = rimDepthMillimetres + innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres
  const recommendation = getServerRenderedInnerTubeValveFitmentRecommendation(rimDepthMillimetres)
  const totalAssemblyLengthMillimetres = recommendation.valveLengthMillimetres + recommendation.extenderLengthMillimetres
  const recommendationKey = recommendation.recommendationKey

  return {
    recommendation_key: recommendationKey,
    recommendation: {
      valve_length_mm: recommendation.valveLengthMillimetres,
      extender_length_mm: recommendation.extenderLengthMillimetres,
      total_assembly_length_mm: totalAssemblyLengthMillimetres,
      minimum_required_length_mm: minimumRequiredLengthMillimetres,
      preferred_minimum_length_mm: minimumRequiredLengthMillimetres + 5,
      recommendation_key: recommendationKey,
      recommendation_reason_key: recommendation.recommendationReasonKey,
    },
  }
}

export const createServerRenderedInnerTubeValveFitmentMatrixFallback = (): InnerTubeValveFitmentServerRenderedMatrixFallbackMetadata => ({
  model_version: 'inner-tube-valve-fitment-v1',
  knowledge_as_of: '2026-10-05',
  rim_depth_min_mm: 20,
  rim_depth_max_mm: 100,
  outer_lip_offset_mm: innerTubeValveFitmentFallbackOuterLipOffsetMillimetres,
  minimum_safe_exposure_mm: innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres,
  preferred_exposure_mm: innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres + 3,
  pump_head_grip_depth_min_mm: 10,
  pump_head_grip_depth_max_mm: 30,
  default_pump_head_grip_depth_mm: innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres,
  rim_depth_uncertainty_max_mm: 5,
  default_rim_depth_uncertainty_mm: 2,
  base_valve_length_options_mm: innerTubeValveFitmentFallbackBaseValveLengthsMillimetres,
  extender_length_options_mm: innerTubeValveFitmentFallbackExtenderLengthsMillimetres,
  preset_rim_depths_mm: innerTubeValveFitmentFallbackRimDepthPresetsMillimetres,
  rows: innerTubeValveFitmentFallbackRimDepthPresetsMillimetres.map((rimDepthMillimetres) => {
    const passageDepthMillimetres = rimDepthMillimetres - innerTubeValveFitmentFallbackOuterLipOffsetMillimetres
    const minimumRequiredLengthMillimetres = rimDepthMillimetres + innerTubeValveFitmentFallbackPumpHeadGripDepthMillimetres

    return {
      rim_depth_mm: rimDepthMillimetres,
      passage_depth_mm: passageDepthMillimetres,
      minimum_required_length_mm: minimumRequiredLengthMillimetres,
      clearances: innerTubeValveFitmentFallbackBaseValveLengthsMillimetres.map(valveLengthMillimetres => {
        const effectiveExposureMillimetres = valveLengthMillimetres - passageDepthMillimetres
        return {
          valve_length_mm: valveLengthMillimetres,
          effective_exposure_mm: effectiveExposureMillimetres,
          status: getServerRenderedInnerTubeValveFitmentStatus(effectiveExposureMillimetres),
        }
      }),
      recommended_result: createServerRenderedInnerTubeValveFitmentMatrixRecommendationSummary(rimDepthMillimetres),
    }
  }),
  calculation_method: 'Static server-rendered fallback mirrors the Go fitment reference model. passage_depth = rim_depth - 6.5 mm; effective_exposure = total_assembly_length - passage_depth; minimum_required_length = rim_depth + pump_head_grip_depth; status is optimal at grip depth + 3 mm, marginal at grip depth, and unsafe below grip depth.',
  source: {
    name: 'Tanzanite inner-tube valve fitment reference model',
    provenance: 'Provided board formula and engineering assumptions; not a manufacturer certification',
  },
  limitations: [
    'This is a reference model for valve fitment, not a certification for a specific rim, tube, pump, or wheelset.',
    'Rim depth must be checked against the exact rim cross-section and valve-hole geometry before installation.',
    'The 6.5 mm outer-lip offset and the default 15 mm pump-head grip depth with 18 mm preferred exposure are model parameters; manufacturer instructions take precedence.',
    'The calculation does not certify tire pressure, hookless compatibility, or the safety of a tire and rim combination.',
  ],
})
