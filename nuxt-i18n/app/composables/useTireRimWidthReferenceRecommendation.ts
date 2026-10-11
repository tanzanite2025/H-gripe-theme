import { computed, onMounted, onScopeDispose, ref, watch, type Ref } from 'vue'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'
import type {
  TireRimReferenceRimSystem,
  TireRimEngineeringCalculation,
  TireRimEngineeringCalculationResponse,
  TireRimWidthRange,
  TireRimWidthReferenceSuggestion,
  TireRimReverseSuggestion as TireRimReverseSuggestionResponse,
  TireRimApplicationProfile,
  TireRimBestFitment,
  TireRimStandardReference,
  TireRimFitmentRuleReference,
  TireRimNormalFitmentTableReference,
} from '~/data/tireguides/tireRimWidthReferencePresentation'
import { mapTireRimBackendBestFitmentForPresentation } from '~/data/tireguides/tireRimWidthReferencePresentation'
import {
  parseTireWidthInputAsMillimeters,
  type TireWidthInputUnit,
} from '~/data/tireguides/tireRimWidthInputConversion'

export type RimType = TireRimReferenceRimSystem

export interface TireRimPhysicalMetrics {
  tireWidth: TireRimWidthRange
  aeroTargetOuterWidth: TireRimWidthRange | null
  rimOuterWidthEstimates: Array<{
    rimInnerWidthMm: number
    wallThicknessPerSideEstimateMm: number
    rimOuterWidthEstimateMm: number
  }>
}

export interface TireRimSuggestion {
  normalFitmentReference?: TireRimNormalFitmentTableReference
  isEngineeringRecommended: boolean
  isEngineeringReference: boolean
  aeroRuleApplies: boolean
  rimWidthRanges: TireRimWidthRange[]
  recommendedRimWidthRanges: TireRimWidthRange[]
  normalFitmentRimWidthRanges: TireRimWidthRange[]
  bestFitment: TireRimBestFitment | null
  applicationProfile: TireRimApplicationProfile
  applicationProfileSelection: string
  standardReference: TireRimStandardReference
  physical: TireRimPhysicalMetrics
  modelVersion: string
  knowledgeAsOf: string
  engineering: TireRimEngineeringCalculation | null
}

export interface TireRimReverseTireWidthMatch {
  tireWidthMm: number
  targetTireWidthDifferenceMm: number
}

export interface TireRimReverseSuggestion {
  normalFitmentReference?: TireRimNormalFitmentTableReference
  rimInnerWidthMm: number
  rimSystem: RimType
  bestFitment: TireRimBestFitment | null
  applicationProfile: TireRimApplicationProfile
  resolvedApplicationProfile: TireRimApplicationProfile
  applicationProfileSelection: string
  standardReference: TireRimStandardReference
  fitmentRuleReference: TireRimFitmentRuleReference
  aeroRuleApplies: boolean
  aeroTargetTireWidthMm: number | null
  bestMatch: TireRimReverseTireWidthMatch | null
  normalFitmentTireWidthRange: TireRimWidthRange
  recommendedFitmentTireWidthRange: TireRimWidthRange | null
  rimOuterWidthEstimate: {
    rimInnerWidthMm: number
    wallThicknessPerSideEstimateMm: number
    rimOuterWidthEstimateMm: number
  }
  modelVersion: string
  knowledgeAsOf: string
}

export const TIRE_WIDTH_INPUT_MIN = 18
export const TIRE_WIDTH_INPUT_MAX = 132

const tireRimWidthReferenceSolveEndpoint = '/engineering/tire-rim/solve'

interface TireRimWidthReferenceSolveResponse {
  data?: TireRimWidthReferenceSuggestion
}

const mapTireRimEngineeringCalculation = (
  calculation?: TireRimEngineeringCalculationResponse,
): TireRimEngineeringCalculation | null => {
  if (!calculation) return null
  return {
    normalFitmentReference: calculation.normal_fitment_reference,
    modelVersion: calculation.model_version,
    verdict: calculation.verdict,
    verdictReason: calculation.verdict_reason,
    rimSystem: calculation.rim_system,
    rimInnerWidthMm: calculation.rim_inner_width_mm,
    tireWidthMm: calculation.tire_width_mm,
    applicationProfile: calculation.application_profile,
    applicationProfileSelection: calculation.application_profile_selection,
    standardReference: calculation.standard_reference,
    fitmentRuleReference: calculation.fitment_rule_reference,
    allowedTireWidthRange: calculation.allowed_tire_width_range,
    recommendedTireWidthRange: calculation.recommended_tire_width_range,
    aeroRuleApplies: calculation.aero_rule_applies,
    aeroTargetOuterWidthMm: calculation.aero_target_outer_width_mm ?? null,
    aeroTargetTireWidthMm: calculation.aero_target_tire_width_mm ?? null,
    maximumPressure: calculation.maximum_pressure,
    tubeRecommendation: {
      band: calculation.tube_recommendation.band,
      widthMinMm: calculation.tube_recommendation.width_min_mm,
      widthMaxMm: calculation.tube_recommendation.width_max_mm,
    },
  }
}

const mapTireRimWidthReferenceSuggestion = (
  suggestion: TireRimWidthReferenceSuggestion,
): TireRimSuggestion => {
  const isEngineeringRecommended = suggestion.result_kind === 'engineering_recommended'
  const isEngineeringReference = suggestion.result_kind === 'engineering_reference'

  return {
    normalFitmentReference: suggestion.normal_fitment_reference,
    isEngineeringRecommended,
    isEngineeringReference,
    aeroRuleApplies: suggestion.aero_rule_applies,
    rimWidthRanges: suggestion.rim_width_ranges,
    recommendedRimWidthRanges: suggestion.recommended_rim_width_ranges ?? [],
    normalFitmentRimWidthRanges: suggestion.normal_fitment_rim_width_ranges,
    bestFitment: mapTireRimBackendBestFitmentForPresentation(suggestion.best_fitment),
    applicationProfile: suggestion.application_profile,
    applicationProfileSelection: suggestion.application_profile_selection,
    standardReference: suggestion.standard_reference,
    physical: {
      tireWidth: suggestion.derived_metrics.tire_width_mm,
      aeroTargetOuterWidth: suggestion.derived_metrics.aero_target_outer_width_mm ?? null,
      rimOuterWidthEstimates: (suggestion.rim_outer_width_estimates ?? []).map(estimate => ({
        rimInnerWidthMm: estimate.rim_inner_width_mm,
        wallThicknessPerSideEstimateMm: estimate.wall_thickness_per_side_estimate_mm,
        rimOuterWidthEstimateMm: estimate.rim_outer_width_estimate_mm,
      })),
    },
    modelVersion: suggestion.model_version,
    knowledgeAsOf: suggestion.knowledge_as_of,
    engineering: mapTireRimEngineeringCalculation(suggestion.engineering),
  }
}

const mapTireRimReverseSuggestion = (
  suggestion: TireRimReverseSuggestionResponse,
): TireRimReverseSuggestion => ({
  normalFitmentReference: suggestion.normal_fitment_reference,
  rimInnerWidthMm: suggestion.rim_inner_width_mm,
  rimSystem: suggestion.rim_system,
  bestFitment: mapTireRimBackendBestFitmentForPresentation(suggestion.best_fitment),
  applicationProfile: suggestion.application_profile,
  resolvedApplicationProfile: suggestion.resolved_application_profile,
  applicationProfileSelection: suggestion.application_profile_selection,
  standardReference: suggestion.standard_reference,
  fitmentRuleReference: suggestion.fitment_rule_reference,
  aeroRuleApplies: suggestion.aero_rule_applies,
  aeroTargetTireWidthMm: suggestion.aero_target_tire_width_mm ?? null,
  bestMatch: suggestion.best_match
    ? {
        tireWidthMm: suggestion.best_match.tire_width_mm,
        targetTireWidthDifferenceMm: suggestion.best_match.target_tire_width_difference_mm,
      }
    : null,
  normalFitmentTireWidthRange: suggestion.normal_fitment_tire_width_range,
  recommendedFitmentTireWidthRange: suggestion.recommended_fitment_tire_width_range,
  rimOuterWidthEstimate: {
    rimInnerWidthMm: suggestion.rim_outer_width_estimate.rim_inner_width_mm,
    wallThicknessPerSideEstimateMm: suggestion.rim_outer_width_estimate.wall_thickness_per_side_estimate_mm,
    rimOuterWidthEstimateMm: suggestion.rim_outer_width_estimate.rim_outer_width_estimate_mm,
  },
  modelVersion: suggestion.model_version,
  knowledgeAsOf: suggestion.knowledge_as_of,
})

export const useTireRimWidthReferenceRecommendation = (
  tireWidthInput: Ref<string>,
  rimType: Ref<RimType>,
  tireWidthInputUnit: Ref<TireWidthInputUnit> = ref<TireWidthInputUnit>('mm'),
  rimInnerWidthMM?: Ref<number | null>,
  enabled: Ref<boolean> = ref(true),
  applicationProfile: Ref<TireRimApplicationProfile> = ref('automatic_by_width'),
) => {
  const { request } = useApiRequest()
  const parsedTireWidth = computed(() => {
    return parseTireWidthInputAsMillimeters(
      String(tireWidthInput.value),
      tireWidthInputUnit.value,
    )
  })
  const tireRimSuggestion = ref<TireRimSuggestion | null>(null)
  const recommendationPending = ref(false)
  const recommendationError = ref<unknown>(null)
  let requestSequence = 0
  let requestAbortController: AbortController | null = null

  const tireWidthOutOfRange = computed(() => {
    const raw = parsedTireWidth.value
    return raw !== null && (
      !Number.isInteger(raw)
      || raw < TIRE_WIDTH_INPUT_MIN
      || raw > TIRE_WIDTH_INPUT_MAX
    )
  })
  const tireWidthHasNoPublishedBracket = computed(() => {
    const raw = parsedTireWidth.value
    return raw !== null
      && !tireWidthOutOfRange.value
      && !recommendationPending.value
      && !tireRimSuggestion.value
      && recommendationError.value instanceof ApiRequestError
      && recommendationError.value.code === 'NO_PUBLISHED_BRACKET'
  })

  const resolveCurrentTireRimWidthReference = async () => {
    const width = parsedTireWidth.value
    const currentRimType = rimType.value
    const currentRequestSequence = ++requestSequence
    requestAbortController?.abort()
    requestAbortController = null
    tireRimSuggestion.value = null
    recommendationError.value = null

    if (
      !import.meta.client
      || !enabled.value
      || width === null
      || tireWidthOutOfRange.value
      || !Number.isInteger(width)
    ) {
      recommendationPending.value = false
      return
    }

    requestAbortController = new AbortController()
    recommendationPending.value = true
    try {
      const response = await request<TireRimWidthReferenceSolveResponse>(
        tireRimWidthReferenceSolveEndpoint,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: requestAbortController.signal,
          body: JSON.stringify({
            tire_width_mm: width,
            rim_system: currentRimType,
            application_profile: applicationProfile.value,
            ...(rimInnerWidthMM?.value !== null && rimInnerWidthMM
              ? { rim_inner_width_mm: rimInnerWidthMM.value }
              : {}),
          }),
        },
        'Tire/rim reference is temporarily unavailable',
      )
      if (currentRequestSequence !== requestSequence || response.data === undefined) return
      tireRimSuggestion.value = mapTireRimWidthReferenceSuggestion(response.data)
    } catch (error) {
      if (currentRequestSequence !== requestSequence || (error instanceof DOMException && error.name === 'AbortError')) return
      recommendationError.value = error
    } finally {
      if (currentRequestSequence === requestSequence) {
        recommendationPending.value = false
      }
    }
  }

  watch([parsedTireWidth, rimType, enabled, applicationProfile, ...(rimInnerWidthMM ? [rimInnerWidthMM] : [])], () => {
    void resolveCurrentTireRimWidthReference()
  })

  onMounted(() => { void resolveCurrentTireRimWidthReference() })

  onScopeDispose(() => {
    requestSequence++
    requestAbortController?.abort()
  })

  return {
    parsedTireWidth,
    tireRimSuggestion,
    recommendationPending,
    recommendationError,
    tireWidthOutOfRange,
    tireWidthHasNoPublishedBracket,
  }
}

export const useTireRimWidthReferenceRecommendationForKnownRimInnerWidth = (
  rimInnerWidthMM: Ref<number | null>,
  rimType: Ref<RimType>,
  enabled: Ref<boolean> = ref(true),
  applicationProfile: Ref<TireRimApplicationProfile> = ref('automatic_by_width'),
) => {
  const { request } = useApiRequest()
  const tireRimReverseSuggestion = ref<TireRimReverseSuggestion | null>(null)
  const recommendationPending = ref(false)
  const recommendationError = ref<unknown>(null)
  let requestSequence = 0
  let requestAbortController: AbortController | null = null

  const resolveCurrentTireRimReverseReference = async () => {
    const currentInnerWidth = rimInnerWidthMM.value
    const currentRimType = rimType.value
    const currentRequestSequence = ++requestSequence
    requestAbortController?.abort()
    requestAbortController = null
    tireRimReverseSuggestion.value = null
    recommendationError.value = null

    if (!import.meta.client || !enabled.value || currentInnerWidth === null) {
      recommendationPending.value = false
      return
    }

    requestAbortController = new AbortController()
    recommendationPending.value = true
    try {
      const response = await request<{ data?: TireRimReverseSuggestionResponse }>(
        tireRimWidthReferenceSolveEndpoint,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: requestAbortController.signal,
          body: JSON.stringify({
            rim_inner_width_mm: currentInnerWidth,
            rim_system: currentRimType,
            application_profile: applicationProfile.value,
          }),
        },
        'Tire/rim reverse reference is temporarily unavailable',
      )
      if (currentRequestSequence !== requestSequence || response.data === undefined) return
      tireRimReverseSuggestion.value = mapTireRimReverseSuggestion(response.data)
    } catch (error) {
      if (currentRequestSequence !== requestSequence || (error instanceof DOMException && error.name === 'AbortError')) return
      recommendationError.value = error
    } finally {
      if (currentRequestSequence === requestSequence) {
        recommendationPending.value = false
      }
    }
  }

  watch([rimInnerWidthMM, rimType, enabled, applicationProfile], () => {
    void resolveCurrentTireRimReverseReference()
  })

  onMounted(() => { void resolveCurrentTireRimReverseReference() })

  onScopeDispose(() => {
    requestSequence++
    requestAbortController?.abort()
  })

  return {
    tireRimReverseSuggestion,
    recommendationPending,
    recommendationError,
  }
}
