import { computed, ref, watch, type Ref } from 'vue'
import { useAsyncData } from '#imports'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'

export interface StainlessSteelSpokeModelReference {
  id: string
  display_name: string
  geometry_description: string
  effective_area_mm2: number
  material_reference_id: string
  data_version: string
  data_status: string
}

export interface StainlessSteelSpokeMaterialReferenceSource {
  title: string
  reference_type: string
  location?: string
  url?: string
  role: string
}

export interface StainlessSteelSpokeMaterialReference {
  id: string
  material_name: string
  material_family: string
  manufacturing_condition: string
  elastic_modulus_mpa: number
  macro_yield_reference_mpa: number
  ultimate_tensile_strength_reference_mpa: number
  data_version: string
  data_status: string
  source_references: StainlessSteelSpokeMaterialReferenceSource[]
  applicability: string
  notes: string
}

export interface StainlessSteelSpokeDislocationMechanicsPhysicalReferences {
  macro_yield_reference_mpa: number
  ultimate_tensile_strength_reference_mpa: number
  elastic_modulus_mpa: number
  carbon_rim_spoke_hole_warning_force_n: number
  high_stress_warning_force_n: number
  critical_yield_ratio_percent: number
  high_stress_yield_ratio_percent: number
}

export interface StainlessSteelSpokeDislocationMechanicsInputBounds {
  nominal_working_tension_min_exclusive_n: number
  nominal_working_tension_max_n: number
  effective_spoke_area_min_exclusive_mm2: number
  effective_spoke_area_max_mm2: number
  effective_spoke_length_min_exclusive_mm: number
  effective_spoke_length_max_mm: number
  overload_ratio_min_percent: number
  overload_ratio_max_percent: number
}

export interface StainlessSteelSpokeDislocationMechanicsMetadata {
  model_version: string
  model_catalog_id: string
  model_catalog_version: string
  material_catalog_id: string
  material_catalog_version: string
  material_reference: StainlessSteelSpokeMaterialReference
  models: StainlessSteelSpokeModelReference[]
  physical_references: StainlessSteelSpokeDislocationMechanicsPhysicalReferences
  input_bounds: StainlessSteelSpokeDislocationMechanicsInputBounds
}

export type StainlessSteelSpokeDislocationMechanicsSafetyLevel = 'reference' | 'warning' | 'critical'

export type StainlessSteelSpokeDislocationMechanicsCalculationErrorKind = 'invalid-input' | 'unavailable'

export interface StainlessSteelSpokeDislocationMechanicsCalculationResult {
  model_version: string
  model_id: string
  model_reference: StainlessSteelSpokeModelReference
  model_catalog_id: string
  model_catalog_version: string
  material_reference_id: string
  material_catalog_id: string
  material_catalog_version: string
  material_reference: StainlessSteelSpokeMaterialReference
  effective_area_mm2: number
  effective_spoke_length_mm: number
  elastic_modulus_mpa: number
  nominal_working_tension_n: number
  overload_ratio_percent: number
  work_stress_mpa: number
  work_yield_ratio_percent: number
  work_elastic_elongation_mm: number
  overload_force_n: number
  overload_delta_force_n: number
  overload_stress_mpa: number
  overload_yield_ratio_percent: number
  overload_elastic_elongation_mm: number
  overload_delta_elastic_elongation_mm: number
  yield_margin_percent: number
  safety_level: StainlessSteelSpokeDislocationMechanicsSafetyLevel
  rim_warning_margin_n: number
  rim_warning_margin_percent: number
  rim_warning_excess_n: number
}

interface StainlessSteelSpokeDislocationMechanicsMetadataResponse {
  data?: StainlessSteelSpokeDislocationMechanicsMetadata
}

interface StainlessSteelSpokeDislocationMechanicsCalculationResponse {
  data?: StainlessSteelSpokeDislocationMechanicsCalculationResult
}

// Keep the catalog revision in the URL so a previously cached metadata
// response cannot hide the material-reference contract or model presets.
const stainlessSteelSpokeDislocationMechanicsMetadataEndpoint = '/engineering/spoke-dislocation-mechanics/metadata?catalog_version=1.2.0&material_catalog_version=1.0.2'
const stainlessSteelSpokeDislocationMechanicsCalculationEndpoint = '/engineering/spoke-dislocation-mechanics/calculate'
const stainlessSteelSpokeDislocationMechanicsRequestTimeoutMilliseconds = 8000

class StainlessSteelSpokeDislocationMechanicsRequestTimeoutError extends Error {
  constructor(message: string) {
    super(message)
    this.name = 'StainlessSteelSpokeDislocationMechanicsRequestTimeoutError'
  }
}

const requestStainlessSteelSpokeDislocationMechanicsWithTimeout = async <T>(
  request: ReturnType<typeof useApiRequest>['request'],
  path: string,
  init: Parameters<ReturnType<typeof useApiRequest>['request']>[1],
  fallbackMessage: string,
  parentAbortSignal?: AbortSignal,
) => {
  const timeoutAbortController = new AbortController()
  let didTimeout = false
  const inheritedAbortSignal = parentAbortSignal || init?.signal || undefined
  const abortFromParent = () => timeoutAbortController.abort()
  inheritedAbortSignal?.addEventListener('abort', abortFromParent, { once: true })
  const timeoutHandle = setTimeout(() => {
    didTimeout = true
    timeoutAbortController.abort()
  }, stainlessSteelSpokeDislocationMechanicsRequestTimeoutMilliseconds)

  try {
    return await request<T>(path, {
      ...init,
      signal: timeoutAbortController.signal,
    }, fallbackMessage)
  } catch (error) {
    if (didTimeout) {
      throw new StainlessSteelSpokeDislocationMechanicsRequestTimeoutError(fallbackMessage)
    }
    throw error
  } finally {
    clearTimeout(timeoutHandle)
    inheritedAbortSignal?.removeEventListener('abort', abortFromParent)
  }
}

export const useStainlessSteelSpokeDislocationMechanicsCalculation = async (
  selectedModelId: Ref<string>,
  customEffectiveAreaMM2: Ref<number | null>,
  effectiveSpokeLengthMM: Ref<number | null>,
  nominalWorkingTensionN: Ref<number | null>,
  overloadRatioPercent: Ref<number | null>,
) => {
  const { request } = useApiRequest()
  const buildStainlessSteelSpokeDislocationMechanicsCalculationRequestBody = () => ({
    model_id: selectedModelId.value,
    custom_effective_area_mm2: selectedModelId.value === 'custom' ? customEffectiveAreaMM2.value : undefined,
    effective_spoke_length_mm: effectiveSpokeLengthMM.value,
    nominal_working_tension_n: nominalWorkingTensionN.value,
    overload_ratio_percent: overloadRatioPercent.value,
  })

  const metadataRequest = useAsyncData<StainlessSteelSpokeDislocationMechanicsMetadataResponse | null>(
    'stainless-steel-spoke-dislocation-mechanics-metadata-model-1-2-0-material-1-0-1',
    async () => {
      try {
        return await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsMetadataResponse>(
          request,
          stainlessSteelSpokeDislocationMechanicsMetadataEndpoint,
          { method: 'GET' },
          'Spoke dislocation mechanics metadata is temporarily unavailable',
        )
      } catch {
        return null
      }
    },
    { default: () => null },
  )

  const serverCalculationRequest = useAsyncData<StainlessSteelSpokeDislocationMechanicsCalculationResponse | null>(
    'stainless-steel-spoke-dislocation-mechanics-default-calculation-v4',
    async () => {
      try {
        return await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsCalculationResponse>(
          request,
          stainlessSteelSpokeDislocationMechanicsCalculationEndpoint,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(buildStainlessSteelSpokeDislocationMechanicsCalculationRequestBody()),
          },
          'Spoke dislocation mechanics calculation is temporarily unavailable',
        )
      } catch {
        return null
      }
    },
    { default: () => null },
  )
  const [
    { data: metadataResponse },
    { data: serverRenderedCalculationResponse },
  ] = await Promise.all([metadataRequest, serverCalculationRequest])

  const calculationResult = ref<StainlessSteelSpokeDislocationMechanicsCalculationResult | null>(
    serverRenderedCalculationResponse.value?.data ?? null,
  )
  // A missing SSR result still needs an authoritative client request. Keep the
  // UI in a loading state until that request completes instead of labelling a
  // temporarily unavailable backend as invalid user input.
  const calculationPending = ref(!calculationResult.value)
  const calculationError = ref<unknown>(null)
  const calculationErrorKind = ref<StainlessSteelSpokeDislocationMechanicsCalculationErrorKind | null>(null)
  const metadataLoadingError = ref<unknown>(null)
  let requestSequence = 0
  let requestAbortController: AbortController | null = null

  const metadata = computed(() => metadataResponse.value?.data ?? null)

  const requestStainlessSteelSpokeDislocationMechanicsMetadata = async () => {
    try {
      metadataLoadingError.value = null
      metadataResponse.value = await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsMetadataResponse>(
        request,
        stainlessSteelSpokeDislocationMechanicsMetadataEndpoint,
        { method: 'GET' },
        'Spoke dislocation mechanics metadata is temporarily unavailable',
      )
    } catch (error) {
      metadataLoadingError.value = error
    }
  }

  const resolveCurrentStainlessSteelSpokeDislocationMechanicsCalculation = async () => {
    if (!import.meta.client) {
      return
    }
    const currentRequestSequence = ++requestSequence
    requestAbortController?.abort()
    requestAbortController = null
    calculationResult.value = null
    calculationError.value = null
    calculationErrorKind.value = null

    requestAbortController = new AbortController()
    calculationPending.value = true
    try {
      const response = await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsCalculationResponse>(
        request,
        stainlessSteelSpokeDislocationMechanicsCalculationEndpoint,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          signal: requestAbortController.signal,
          body: JSON.stringify(buildStainlessSteelSpokeDislocationMechanicsCalculationRequestBody()),
        },
        'Spoke dislocation mechanics calculation is temporarily unavailable',
        requestAbortController.signal,
      )
      if (currentRequestSequence !== requestSequence || response.data === undefined) return
      calculationResult.value = response.data
    } catch (error) {
      if (currentRequestSequence !== requestSequence || (error instanceof DOMException && error.name === 'AbortError')) return
      calculationError.value = error
      calculationErrorKind.value = error instanceof ApiRequestError && [400, 422].includes(error.status)
        ? 'invalid-input'
        : 'unavailable'
    } finally {
      if (currentRequestSequence === requestSequence) {
        calculationPending.value = false
      }
    }
  }

  let stainlessSteelSpokeDislocationMechanicsCalculationResolutionScheduled = false
  const scheduleStainlessSteelSpokeDislocationMechanicsCalculationResolution = () => {
    if (stainlessSteelSpokeDislocationMechanicsCalculationResolutionScheduled) return
    stainlessSteelSpokeDislocationMechanicsCalculationResolutionScheduled = true
    void Promise.resolve().then(() => {
      stainlessSteelSpokeDislocationMechanicsCalculationResolutionScheduled = false
      void resolveCurrentStainlessSteelSpokeDislocationMechanicsCalculation()
    })
  }

  watch(
    [selectedModelId, effectiveSpokeLengthMM, nominalWorkingTensionN, overloadRatioPercent],
    () => {
      scheduleStainlessSteelSpokeDislocationMechanicsCalculationResolution()
    },
  )

  watch(customEffectiveAreaMM2, () => {
    if (selectedModelId.value === 'custom') {
      scheduleStainlessSteelSpokeDislocationMechanicsCalculationResolution()
    }
  })

  if (import.meta.client) {
    const metadataPromise = metadata.value
      ? Promise.resolve()
      : requestStainlessSteelSpokeDislocationMechanicsMetadata()
    const calculationPromise = calculationResult.value
      ? Promise.resolve()
      : resolveCurrentStainlessSteelSpokeDislocationMechanicsCalculation()
    void Promise.allSettled([metadataPromise, calculationPromise])
  }

  return {
    metadata,
    metadataLoadingError,
    calculationResult,
    calculationPending,
    calculationError,
    calculationErrorKind,
    resolveCurrentStainlessSteelSpokeDislocationMechanicsCalculation,
    requestStainlessSteelSpokeDislocationMechanicsMetadata,
  }
}
