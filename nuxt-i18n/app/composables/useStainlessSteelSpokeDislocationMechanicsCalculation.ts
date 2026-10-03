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

export interface StainlessSteelSpokeReferenceCalculationResult {
  model_id: string
  effective_area_mm2: number
  reference_calculation_stress_mpa: number
  reference_calculation_total_strain_percent: number
  reference_calculation_total_elongation_mm: number
  reference_calculation_elastic_elongation_mm: number
  reference_calculation_permanent_elongation_mm: number
  yield_reference_force_n: number
  curve_endpoint_force_n: number
  fracture_reference_total_elongation_mm: number
}

export interface StainlessSteelSpokeReferenceCalculation {
  effective_spoke_length_mm: number
  reference_calculation_tension_kgf: number
  reference_calculation_tension_n: number
  yield_strength_mpa: number
  curve_endpoint_stress_mpa: number
  curve_endpoint_total_strain_percent: number
  curve_endpoint_total_elongation_mm: number
  total_elongation_to_failure_percent: number
  fracture_reference_total_elongation_mm: number
  results: StainlessSteelSpokeReferenceCalculationResult[]
}

export interface StainlessSteelSpokeMaterialReferenceSource {
  title: string
  reference_type: string
  location?: string
  url?: string
  role: string
}

export interface StainlessSteelSpokeEngineeringStressStrainPoint {
  engineering_strain: number
  engineering_stress_mpa: number
}

export interface StainlessSteelSpokeMaterialReference {
  id: string
  material_name: string
  material_family: string
  manufacturing_condition: string
  elastic_modulus_mpa: number
  yield_strength_mpa: number
  ultimate_tensile_strength_mpa: number
  total_elongation_to_failure_percent: number
  curve_type: string
  stress_strain_curve: StainlessSteelSpokeEngineeringStressStrainPoint[]
  data_version: string
  data_status: string
  source_references: StainlessSteelSpokeMaterialReferenceSource[]
  applicability: string
  notes: string
}

export interface StainlessSteelSpokeDislocationMechanicsPhysicalReferences {
  yield_strength_mpa: number
  ultimate_tensile_strength_mpa: number
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
  reference_calculation: StainlessSteelSpokeReferenceCalculation
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
  work_total_strain_percent: number
  work_elastic_strain_percent: number
  work_plastic_strain_percent: number
  work_total_elongation_mm: number
  work_elastic_elongation_mm: number
  work_permanent_elongation_mm: number
  overload_force_n: number
  overload_delta_force_n: number
  overload_stress_mpa: number
  overload_yield_ratio_percent: number
  overload_total_strain_percent: number
  overload_elastic_strain_percent: number
  overload_plastic_strain_percent: number
  overload_total_elongation_mm: number
  overload_elastic_elongation_mm: number
  overload_permanent_elongation_mm: number
  overload_delta_elastic_elongation_mm: number
  overload_delta_total_elongation_mm: number
  overload_delta_permanent_elongation_mm: number
  fracture_reference_total_elongation_mm: number
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

// Keep the catalog revision and reference contract in the URL so a previously
// cached metadata response cannot hide the material-boundary table contract.
const stainlessSteelSpokeDislocationMechanicsMetadataEndpoint = '/engineering/spoke-dislocation-mechanics/metadata?catalog_version=1.3.0&material_catalog_version=2.1.0&reference_calculation=calculation-tension-180kgf-v1'
const stainlessSteelSpokeDislocationMechanicsCalculationEndpoint = '/engineering/spoke-dislocation-mechanics/calculate'
const stainlessSteelSpokeDislocationMechanicsRequestTimeoutMilliseconds = 8000
const stainlessSteelSpokeDislocationMechanicsExpectedModelCatalogVersion = '1.3.0'
const stainlessSteelSpokeDislocationMechanicsExpectedMaterialCatalogVersion = '2.1.0'

const isFiniteStainlessSteelSpokeDislocationMechanicsNumber = (value: unknown): value is number => (
  typeof value === 'number' && Number.isFinite(value)
)

const hasCurrentStainlessSteelSpokeMaterialReferenceContract = (
  value: unknown,
): value is StainlessSteelSpokeMaterialReference => {
  if (!value || typeof value !== 'object') return false
  const material = value as Partial<StainlessSteelSpokeMaterialReference>
  return typeof material.id === 'string'
    && typeof material.material_name === 'string'
    && typeof material.material_family === 'string'
    && typeof material.manufacturing_condition === 'string'
    && isFiniteStainlessSteelSpokeDislocationMechanicsNumber(material.elastic_modulus_mpa)
    && isFiniteStainlessSteelSpokeDislocationMechanicsNumber(material.yield_strength_mpa)
    && isFiniteStainlessSteelSpokeDislocationMechanicsNumber(material.ultimate_tensile_strength_mpa)
     && isFiniteStainlessSteelSpokeDislocationMechanicsNumber(material.total_elongation_to_failure_percent)
    && typeof material.curve_type === 'string'
    && Array.isArray(material.stress_strain_curve)
    && material.stress_strain_curve.length > 1
}

const hasCurrentStainlessSteelSpokeReferenceCalculationContract = (
  value: unknown,
  models: unknown,
): value is StainlessSteelSpokeReferenceCalculation => {
  if (!value || typeof value !== 'object' || !Array.isArray(models)) return false
  const referenceCalculation = value as Partial<StainlessSteelSpokeReferenceCalculation>
  if (!isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.effective_spoke_length_mm)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.reference_calculation_tension_kgf)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.reference_calculation_tension_n)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.yield_strength_mpa)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.curve_endpoint_stress_mpa)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.curve_endpoint_total_strain_percent)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.curve_endpoint_total_elongation_mm)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.total_elongation_to_failure_percent)
    || !isFiniteStainlessSteelSpokeDislocationMechanicsNumber(referenceCalculation.fracture_reference_total_elongation_mm)
    || !Array.isArray(referenceCalculation.results)) {
    return false
  }
  const modelIDs = new Set(models
    .filter((model): model is { id: string } => Boolean(model && typeof model === 'object' && typeof (model as { id?: unknown }).id === 'string'))
    .map(model => model.id))
  if (modelIDs.size === 0 || referenceCalculation.results.length !== modelIDs.size) return false
  const resultIDs = new Set<string>()
  return referenceCalculation.results.every(result => {
    if (!result || typeof result !== 'object') return false
    const typedResult = result as Partial<StainlessSteelSpokeReferenceCalculationResult>
    if (typeof typedResult.model_id !== 'string' || !modelIDs.has(typedResult.model_id) || resultIDs.has(typedResult.model_id)) {
      return false
    }
    const numericFields: Array<keyof StainlessSteelSpokeReferenceCalculationResult> = [
      'effective_area_mm2',
      'reference_calculation_stress_mpa',
      'reference_calculation_total_strain_percent',
      'reference_calculation_total_elongation_mm',
      'reference_calculation_elastic_elongation_mm',
      'reference_calculation_permanent_elongation_mm',
      'yield_reference_force_n',
      'curve_endpoint_force_n',
      'fracture_reference_total_elongation_mm',
    ]
    if (!numericFields.every(field => isFiniteStainlessSteelSpokeDislocationMechanicsNumber(typedResult[field]))) {
      return false
    }
    resultIDs.add(typedResult.model_id)
    return true
  })
}

const isCurrentStainlessSteelSpokeDislocationMechanicsMetadataResponse = (
  value: unknown,
): value is StainlessSteelSpokeDislocationMechanicsMetadataResponse => {
  if (!value || typeof value !== 'object') return false
  const data = (value as { data?: unknown }).data
  if (!data || typeof data !== 'object') return false
  const metadata = data as Partial<StainlessSteelSpokeDislocationMechanicsMetadata>
  return metadata.model_catalog_version === stainlessSteelSpokeDislocationMechanicsExpectedModelCatalogVersion
    && metadata.material_catalog_version === stainlessSteelSpokeDislocationMechanicsExpectedMaterialCatalogVersion
    && hasCurrentStainlessSteelSpokeMaterialReferenceContract(metadata.material_reference)
    && Array.isArray(metadata.models)
    && hasCurrentStainlessSteelSpokeReferenceCalculationContract(metadata.reference_calculation, metadata.models)
}

const isCurrentStainlessSteelSpokeDislocationMechanicsCalculationResponse = (
  value: unknown,
): value is StainlessSteelSpokeDislocationMechanicsCalculationResponse => {
  if (!value || typeof value !== 'object') return false
  const data = (value as { data?: unknown }).data
  if (!data || typeof data !== 'object') return false
  const result = data as Partial<StainlessSteelSpokeDislocationMechanicsCalculationResult>
  const requiredNumericFields: Array<keyof StainlessSteelSpokeDislocationMechanicsCalculationResult> = [
    'effective_area_mm2',
    'effective_spoke_length_mm',
    'elastic_modulus_mpa',
    'work_total_elongation_mm',
    'work_elastic_elongation_mm',
    'work_permanent_elongation_mm',
    'overload_total_elongation_mm',
    'overload_elastic_elongation_mm',
    'overload_permanent_elongation_mm',
    'overload_delta_total_elongation_mm',
    'overload_delta_permanent_elongation_mm',
    'fracture_reference_total_elongation_mm',
  ]
  return result.model_catalog_version === stainlessSteelSpokeDislocationMechanicsExpectedModelCatalogVersion
    && result.material_catalog_version === stainlessSteelSpokeDislocationMechanicsExpectedMaterialCatalogVersion
    && hasCurrentStainlessSteelSpokeMaterialReferenceContract(result.material_reference)
    && requiredNumericFields.every(field => isFiniteStainlessSteelSpokeDislocationMechanicsNumber(result[field]))
}

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
    'stainless-steel-spoke-dislocation-mechanics-metadata-model-1-3-0-material-2-1-0-reference-180-kgf-270-mm-curve-boundaries-fracture-v3',
    async () => {
      try {
        const response = await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsMetadataResponse>(
          request,
          stainlessSteelSpokeDislocationMechanicsMetadataEndpoint,
          { method: 'GET' },
          'Spoke dislocation mechanics metadata is temporarily unavailable',
        )
        return isCurrentStainlessSteelSpokeDislocationMechanicsMetadataResponse(response) ? response : null
      } catch {
        return null
      }
    },
    { default: () => null },
  )

  const serverCalculationRequest = useAsyncData<StainlessSteelSpokeDislocationMechanicsCalculationResponse | null>(
    'stainless-steel-spoke-dislocation-mechanics-default-calculation-v8-aisi-304-curve-fracture-reference',
    async () => {
      try {
        const response = await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsCalculationResponse>(
          request,
          stainlessSteelSpokeDislocationMechanicsCalculationEndpoint,
          {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(buildStainlessSteelSpokeDislocationMechanicsCalculationRequestBody()),
          },
          'Spoke dislocation mechanics calculation is temporarily unavailable',
        )
        return isCurrentStainlessSteelSpokeDislocationMechanicsCalculationResponse(response) ? response : null
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
      const response = await requestStainlessSteelSpokeDislocationMechanicsWithTimeout<StainlessSteelSpokeDislocationMechanicsMetadataResponse>(
        request,
        stainlessSteelSpokeDislocationMechanicsMetadataEndpoint,
        { method: 'GET' },
        'Spoke dislocation mechanics metadata is temporarily unavailable',
      )
      if (!isCurrentStainlessSteelSpokeDislocationMechanicsMetadataResponse(response)) {
        throw new Error('Spoke dislocation mechanics metadata contract is outdated')
      }
      metadataResponse.value = response
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
      if (!isCurrentStainlessSteelSpokeDislocationMechanicsCalculationResponse(response)) {
        throw new Error('Spoke dislocation mechanics calculation contract is outdated')
      }
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
