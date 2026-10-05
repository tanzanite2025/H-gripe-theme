import { computed, ref, watch, type Ref } from 'vue'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'
import type {
  TireRimReferenceRimSystem,
  TireRimWidthRange,
  TireRimWidthReferenceSuggestion,
} from '~/data/tireguides/tireRimWidthReferencePresentation'
import {
  parseTireWidthInputAsMillimeters,
  type TireWidthInputUnit,
} from '~/data/tireguides/tireRimWidthInputConversion'

export type RimType = TireRimReferenceRimSystem

export interface TireRimPhysicalMetrics {
  inflatedTireWidth: TireRimWidthRange
  aeroTargetOuterWidth: TireRimWidthRange
}

export interface TireRimSuggestion {
  isCalculated: boolean
  isCalculatedFromPossibleReference: boolean
  isPossibleReference: boolean
  resultKind: TireRimWidthReferenceSuggestion['result_kind']
  rimWidthRanges: TireRimWidthRange[]
  physical: TireRimPhysicalMetrics
  calculationRows: {
    lowerTireWidth: number
    upperTireWidth: number
  } | null
  modelVersion: string
  knowledgeAsOf: string
}

export const TIRE_WIDTH_INPUT_MIN = 18
export const TIRE_WIDTH_INPUT_MAX = 127

const tireRimWidthReferenceSolveEndpoint = '/engineering/tire-rim/solve'

interface TireRimWidthReferenceSolveResponse {
  data?: TireRimWidthReferenceSuggestion
}

const mapTireRimWidthReferenceSuggestion = (
  suggestion: TireRimWidthReferenceSuggestion,
): TireRimSuggestion => ({
  isCalculated: suggestion.result_kind === 'interpolated',
  isCalculatedFromPossibleReference: suggestion.result_kind === 'interpolated'
    && suggestion.source_rows.some(sourceRow => sourceRow.kind === 'possible_reference'),
  isPossibleReference: suggestion.result_kind === 'possible_reference',
  resultKind: suggestion.result_kind,
  rimWidthRanges: suggestion.rim_width_ranges,
  physical: {
    inflatedTireWidth: suggestion.derived_metrics.inflated_tire_width_mm,
    aeroTargetOuterWidth: suggestion.derived_metrics.aero_target_outer_width_mm,
  },
  calculationRows: suggestion.calculation
    ? {
        lowerTireWidth: suggestion.calculation.lower_tire_width_mm,
        upperTireWidth: suggestion.calculation.upper_tire_width_mm,
      }
    : null,
  modelVersion: suggestion.model_version,
  knowledgeAsOf: suggestion.knowledge_as_of,
})

export const useTireRimWidthReferenceRecommendation = (
  tireWidthInput: Ref<string>,
  rimType: Ref<RimType>,
  tireWidthInputUnit: Ref<TireWidthInputUnit> = ref<TireWidthInputUnit>('mm'),
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

  watch([parsedTireWidth, rimType], () => {
    void resolveCurrentTireRimWidthReference()
  }, { immediate: true })

  return {
    parsedTireWidth,
    tireRimSuggestion,
    recommendationPending,
    recommendationError,
    tireWidthOutOfRange,
    tireWidthHasNoPublishedBracket,
  }
}
