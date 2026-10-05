import { computed, onBeforeUnmount, ref, watch, type ComputedRef, type Ref } from 'vue'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'

export type TirePressureCalculatorDynamicsState = 'PENDING' | 'BACKEND DEMO' | 'UNAVAILABLE'

export type TirePressureCalculatorVerticalDeformationEstimate = {
  data_status: string
  dataset_version: string
  reference_pressure_bar: number
  operating_pressure_psi: number
  operating_pressure_bar: number
  vertical_load_n: number
  reference_deflection_mm: number
  estimated_deflection_mm: number
  relative_deformation_index: number
  reference_vertical_stiffness_n_per_mm: number
  estimated_vertical_stiffness_n_per_mm: number
  calculation_method: string
  pressure_scaling_assumption: string
}

export type TirePressureCalculatorPressureContactAreaComparison = {
  reference_pressure_psi: number
  comparison_pressure_psi: number
  reference_contact_area_cm2: number
  comparison_contact_area_cm2: number
  area_change_pct: number
  reference_contact_patch_width_mm?: number
  reference_contact_patch_length_mm?: number
  comparison_contact_patch_width_mm?: number
  comparison_contact_patch_length_mm?: number
}

export type TirePressureCalculatorPressureFrictionCoefficientEstimate = {
  data_status: string
  operating_pressure_psi: number
  nominal_coefficient: number
  estimated_coefficient: number
  relative_coefficient_index: number
  pressure_effect_applied: boolean
  calculation_method: string
}

export type TirePressureCalculatorWetPressureCompensation = {
  water_film_depth_mm: number
  speed_kmh: number
  lateral_demand_ratio: number
  friction_retention_ratio: number
  reference_pressure_psi: number
  equivalent_pressure_psi: number
  pressure_reduction_psi: number
  pressure_reduction_pct: number
  reference_contact_area_cm2: number
  equivalent_contact_area_cm2: number
  contact_area_change_pct: number
  wet_grip_limit_at_reference_pressure_n: number
  wet_grip_margin_pct: number
  pressure_clamped_to_minimum: boolean
  minimum_pressure_psi?: number
  surface_texture_baseline: string
  rubber_baseline: string
}

export type TirePressureCalculatorBackendWheelDynamics = {
  load_kg: number
  vertical_load_n: number
  lateral_demand_n: number
  idealized_grip_limit_n: number
  grip_margin_pct: number
  resultant_contact_force_n: number
  estimated_static_contact_area_cm2?: number
  estimated_equivalent_circular_contact_diameter_mm?: number
  estimated_contact_patch_width_mm?: number
  estimated_contact_patch_length_mm?: number
  vertical_deformation?: TirePressureCalculatorVerticalDeformationEstimate
  pressure_contact_area_comparison?: TirePressureCalculatorPressureContactAreaComparison
  pressure_friction_coefficient?: TirePressureCalculatorPressureFrictionCoefficientEstimate
  wet_pressure_compensation?: TirePressureCalculatorWetPressureCompensation
  mu_nominal: number
}

export type TirePressureCalculatorBackendDynamics = {
  model_version: string
  load_source: string
  front_load_kg: number
  rear_load_kg: number
  dynamics: {
    model_version: string
    model_status: string
    force_frame: string
    lean_angle_deg: number
    speed_kmh: number
    equivalent_turn_radius_m?: number
    lateral_acceleration_mps2: number
    lateral_acceleration_g: number
    surface_condition: string
    tire_body_normalization_factor: number
    tire_body_normalization_source: string
    tire_width_mm: number
    tire_width_source: string
    front: TirePressureCalculatorBackendWheelDynamics
    rear: TirePressureCalculatorBackendWheelDynamics
  }
  warnings: string[]
}

export type TirePressureCalculatorPressureBounds = {
  minimumPsi: number
  maximumPsi: number
}

export type TirePressureCalculatorDynamicsRequestInputs = {
  riderWeight: Ref<number>
  bikeWeight: Ref<number>
  tireWidth: Ref<number>
  speedKmh: Ref<number>
  apiPosition: ComputedRef<string>
  leanAngle: Ref<number>
  selectedTireArticleNo: ComputedRef<string | null>
  selectedTireReferencePressurePsi: ComputedRef<number | null>
  selectedTirePressureBoundsPsi: ComputedRef<TirePressureCalculatorPressureBounds | null>
  pressureComparisonEnabled: Ref<boolean>
  wetPressureScenarioEnabled: Ref<boolean>
  pressureComparisonAvailable: ComputedRef<boolean>
  frontComparisonPressurePsi: ComputedRef<number | null>
  rearComparisonPressurePsi: ComputedRef<number | null>
  wetPressureWaterFilmDepthMm: number
}

const tirePressureDynamicsResponseCacheLimit = 32
const tirePressureDynamicsMaximumAutomaticRetries = 3
const dynamicsResponseCache = new Map<string, TirePressureCalculatorBackendDynamics>()

const isRecord = (value: unknown): value is Record<string, unknown> => (
  typeof value === 'object' && value !== null && !Array.isArray(value)
)

const hasFiniteNumberFields = (record: Record<string, unknown>, fields: string[]) => fields.every((field) => {
  const value = record[field]
  return typeof value === 'number' && Number.isFinite(value)
})

const hasStringFields = (record: Record<string, unknown>, fields: string[]) => fields.every((field) => (
  typeof record[field] === 'string'
))

/**
 * Keeps malformed or partially compatible backend payloads out of the LRU.
 * The transport type is not a runtime guarantee, especially during rolling
 * deployments where an older API may still answer this request.
 */
function parseTirePressureCalculatorBackendDynamics(payload: unknown): TirePressureCalculatorBackendDynamics | null {
  if (!isRecord(payload)) return null
  const dynamics = payload.dynamics
  if (!isRecord(dynamics) || !Array.isArray(payload.warnings)) {
    return null
  }

  const front = dynamics.front
  const rear = dynamics.rear
  if (!isRecord(front) || !isRecord(rear)) return null
  if (
    !hasStringFields(payload, ['model_version', 'load_source'])
    || !hasFiniteNumberFields(payload, ['front_load_kg', 'rear_load_kg'])
    || !hasStringFields(dynamics, ['model_version', 'model_status', 'force_frame', 'surface_condition', 'tire_body_normalization_source', 'tire_width_source'])
    || !hasFiniteNumberFields(dynamics, ['lean_angle_deg', 'speed_kmh', 'lateral_acceleration_mps2', 'lateral_acceleration_g', 'tire_body_normalization_factor', 'tire_width_mm'])
    || !hasFiniteNumberFields(front, ['load_kg', 'vertical_load_n', 'lateral_demand_n', 'idealized_grip_limit_n', 'grip_margin_pct', 'resultant_contact_force_n', 'mu_nominal'])
    || !hasFiniteNumberFields(rear, ['load_kg', 'vertical_load_n', 'lateral_demand_n', 'idealized_grip_limit_n', 'grip_margin_pct', 'resultant_contact_force_n', 'mu_nominal'])
    || payload.warnings.some(warning => typeof warning !== 'string')
  ) {
    return null
  }

  for (const record of [dynamics, front, rear]) {
    for (const field of ['equivalent_turn_radius_m', 'estimated_static_contact_area_cm2', 'estimated_equivalent_circular_contact_diameter_mm', 'estimated_contact_patch_width_mm', 'estimated_contact_patch_length_mm']) {
      if (field in record && (typeof record[field] !== 'number' || !Number.isFinite(record[field]))) return null
    }
  }

  return payload as unknown as TirePressureCalculatorBackendDynamics
}

function readCachedTirePressureDynamics(inputKey: string) {
  const cachedDynamics = dynamicsResponseCache.get(inputKey)
  if (!cachedDynamics) return null

  dynamicsResponseCache.delete(inputKey)
  dynamicsResponseCache.set(inputKey, cachedDynamics)
  return cachedDynamics
}

function cacheTirePressureDynamics(inputKey: string, dynamics: TirePressureCalculatorBackendDynamics) {
  dynamicsResponseCache.delete(inputKey)
  dynamicsResponseCache.set(inputKey, dynamics)
  while (dynamicsResponseCache.size > tirePressureDynamicsResponseCacheLimit) {
    const oldestInputKey = dynamicsResponseCache.keys().next().value
    if (oldestInputKey === undefined) break
    dynamicsResponseCache.delete(oldestInputKey)
  }
}

/**
 * Owns the calculator's dynamics transport lifecycle. The component only
 * supplies reactive inputs and renders the committed response, so request
 * caching and stale-response handling cannot become a second UI concern.
 */
export const useTirePressureCalculatorDynamicsRequest = (
  inputs: TirePressureCalculatorDynamicsRequestInputs,
) => {
  const { request } = useApiRequest()
  const dynamicsState = ref<TirePressureCalculatorDynamicsState>('PENDING')
  const backendDynamics = ref<TirePressureCalculatorBackendDynamics | null>(null)
  const committedDynamicsInputKey = ref<string | null>(null)
  const pressureComparisonActive = computed(() => inputs.pressureComparisonEnabled.value || inputs.wetPressureScenarioEnabled.value)
  const dynamicsInputKey = computed(() => JSON.stringify({
    rider: inputs.riderWeight.value,
    bike: inputs.bikeWeight.value,
    tire: inputs.tireWidth.value,
    speed: inputs.speedKmh.value,
    position: inputs.apiPosition.value,
    lean: inputs.leanAngle.value,
    tireArticleNo: inputs.selectedTireArticleNo.value,
    referencePressurePsi: inputs.selectedTireReferencePressurePsi.value,
    minimumPressurePsi: inputs.selectedTirePressureBoundsPsi.value?.minimumPsi ?? null,
    pressureComparisonEnabled: inputs.pressureComparisonEnabled.value,
    wetPressureScenarioEnabled: inputs.wetPressureScenarioEnabled.value,
    wetPressureWaterFilmDepthMm: inputs.wetPressureWaterFilmDepthMm,
    frontComparisonPressurePsi: inputs.frontComparisonPressurePsi.value,
    rearComparisonPressurePsi: inputs.rearComparisonPressurePsi.value,
  }))
  const currentBackendDynamics = computed(() => (
    committedDynamicsInputKey.value === dynamicsInputKey.value ? backendDynamics.value : null
  ))
  const comparisonRequestPending = computed(() => (
    inputs.pressureComparisonAvailable.value
    && pressureComparisonActive.value
    && dynamicsState.value === 'PENDING'
  ))
  const comparisonRequestFailed = computed(() => (
    inputs.pressureComparisonAvailable.value
    && pressureComparisonActive.value
    && dynamicsState.value === 'UNAVAILABLE'
  ))

  let dynamicsSequence = 0
  let dynamicsRequestInFlight = false
  let dynamicsRefreshQueued = false
  let dynamicsRetryNotBefore = 0
  const dynamicsRetryAttemptsByInputKey = new Map<string, number>()
  let dynamicsTimer: ReturnType<typeof setTimeout> | undefined
  let isCalculatorUnmounted = false

  function scheduleTirePressureDynamicsRefresh(delayMilliseconds = 450) {
    if (dynamicsTimer) clearTimeout(dynamicsTimer)
    dynamicsState.value = 'PENDING'
    dynamicsTimer = setTimeout(() => void refreshTirePressureGroundFrameDynamicsFromBackend(), delayMilliseconds)
  }

  function retryTirePressureDynamics() {
    dynamicsRetryAttemptsByInputKey.delete(dynamicsInputKey.value)
    dynamicsRetryNotBefore = 0
    dynamicsRefreshQueued = false
    scheduleTirePressureDynamicsRefresh(0)
  }

  async function refreshTirePressureGroundFrameDynamicsFromBackend() {
    if (!import.meta.client || isCalculatorUnmounted) return
    const inputKey = dynamicsInputKey.value

    if (inputs.selectedTireReferencePressurePsi.value === null) {
      backendDynamics.value = null
      committedDynamicsInputKey.value = null
      dynamicsState.value = 'PENDING'
      return
    }

    const cachedDynamics = readCachedTirePressureDynamics(inputKey)
    if (cachedDynamics) {
      backendDynamics.value = cachedDynamics
      committedDynamicsInputKey.value = inputKey
      dynamicsState.value = 'BACKEND DEMO'
      dynamicsRetryAttemptsByInputKey.delete(inputKey)
      return
    }

    if (Date.now() < dynamicsRetryNotBefore) {
      scheduleTirePressureDynamicsRefresh(Math.max(450, dynamicsRetryNotBefore - Date.now()))
      return
    }

    if (dynamicsRequestInFlight) {
      dynamicsRefreshQueued = true
      return
    }

    const requestId = ++dynamicsSequence
    dynamicsRequestInFlight = true
    dynamicsState.value = 'PENDING'
    const body: Record<string, unknown> = {
      rider_weight_kg: inputs.riderWeight.value,
      bike_weight_kg: inputs.bikeWeight.value,
      nominal_tire_width_mm: inputs.tireWidth.value,
      speed_kmh: inputs.speedKmh.value,
      riding_position: inputs.apiPosition.value,
      surface_condition: 'FLAT_ROAD',
      lean_angle_deg: inputs.leanAngle.value,
      front_operating_pressure_psi: inputs.selectedTireReferencePressurePsi.value,
      rear_operating_pressure_psi: inputs.selectedTireReferencePressurePsi.value,
    }
    if (inputs.wetPressureScenarioEnabled.value) {
      body.wet_pressure_demonstration_enabled = true
      body.water_film_depth_mm = inputs.wetPressureWaterFilmDepthMm
      const minimumPressurePsi = inputs.selectedTirePressureBoundsPsi.value?.minimumPsi
      if (minimumPressurePsi !== undefined) {
        body.front_minimum_pressure_psi = minimumPressurePsi
        body.rear_minimum_pressure_psi = minimumPressurePsi
      }
    } else if (inputs.frontComparisonPressurePsi.value !== null && inputs.rearComparisonPressurePsi.value !== null) {
      body.front_comparison_pressure_psi = inputs.frontComparisonPressurePsi.value
      body.rear_comparison_pressure_psi = inputs.rearComparisonPressurePsi.value
    }

    try {
      const response = await request<unknown>(
        '/engineering/tire-pressure/dynamics',
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        },
      )
      if (isCalculatorUnmounted) return
      const parsedDynamics = isRecord(response) ? parseTirePressureCalculatorBackendDynamics(response.data) : null
      if (!parsedDynamics) throw new Error('Invalid tire-pressure dynamics response')
      cacheTirePressureDynamics(inputKey, parsedDynamics)
      if (requestId === dynamicsSequence && inputKey === dynamicsInputKey.value) {
        backendDynamics.value = parsedDynamics
        committedDynamicsInputKey.value = inputKey
        dynamicsState.value = 'BACKEND DEMO'
      }
      dynamicsRetryAttemptsByInputKey.delete(inputKey)
      dynamicsRetryNotBefore = 0
    } catch (error) {
      if (isCalculatorUnmounted || requestId !== dynamicsSequence || inputKey !== dynamicsInputKey.value) return
      if (error instanceof ApiRequestError && error.status === 429) {
        const attempts = dynamicsRetryAttemptsByInputKey.get(inputKey) ?? 0
        if (attempts < tirePressureDynamicsMaximumAutomaticRetries) {
          dynamicsRetryAttemptsByInputKey.set(inputKey, attempts + 1)
          dynamicsRetryNotBefore = Date.now() + 2000
          dynamicsRefreshQueued = true
        }
      }
      dynamicsState.value = 'UNAVAILABLE'
    } finally {
      dynamicsRequestInFlight = false
      if (isCalculatorUnmounted) return
      if (dynamicsRefreshQueued || inputKey !== dynamicsInputKey.value) {
        dynamicsRefreshQueued = false
        scheduleTirePressureDynamicsRefresh()
      }
    }
  }

  watch([
    inputs.riderWeight,
    inputs.bikeWeight,
    inputs.leanAngle,
    inputs.speedKmh,
    inputs.apiPosition,
    inputs.tireWidth,
    inputs.selectedTireArticleNo,
    inputs.selectedTireReferencePressurePsi,
    inputs.selectedTirePressureBoundsPsi,
    inputs.pressureComparisonEnabled,
    inputs.wetPressureScenarioEnabled,
  ], () => {
    scheduleTirePressureDynamicsRefresh()
  }, { immediate: true })

  onBeforeUnmount(() => {
    isCalculatorUnmounted = true
    if (dynamicsTimer) {
      clearTimeout(dynamicsTimer)
      dynamicsTimer = undefined
    }
    dynamicsRefreshQueued = false
  })

  return {
    dynamicsState,
    backendDynamics,
    currentBackendDynamics,
    dynamicsInputKey,
    comparisonRequestPending,
    comparisonRequestFailed,
    retryTirePressureDynamics,
  }
}
