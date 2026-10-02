<template>
  <div class="simulator">
    <header class="topbar">
      <div>
        <h2>{{ t('guidesTirePressure.dashboard.title') }}</h2>
      </div>
      <p>{{ t('guidesTirePressure.dashboard.subtitle') }}</p>
    </header>

    <aside class="notice" role="note">
      <span class="notice-mark" aria-hidden="true">!</span>
      <div class="notice-copy">
        <strong>{{ t('guidesTirePressure.dashboard.controlBaselineTitle') }}</strong>
        <p>{{ t('guidesTirePressure.dashboard.controlBaselineBody') }}</p>
      </div>
    </aside>
    <p class="model-badge">{{ t('guidesTirePressure.dashboard.baseline') }} · {{ t('guidesTirePressure.dashboard.scopeTitle') }} · {{ t('guidesTirePressure.dashboard.modelVersionLabel', { value: modelVersionLabel }) }}</p>
    <p class="model-provenance">{{ t('guidesTirePressure.dashboard.provenanceLabel', { load: loadSourceLabel, width: tireWidthSourceLabel, body: tireBodyNormalizationLabel }) }}</p>

    <section class="hero-grid">
      <article class="hero-card front">
        <div><b>{{ t('guidesTirePressure.dashboard.frontWheel') }}</b><span class="load-tag">{{ formatRatio(frontRatio) }} {{ t('guidesTirePressure.dashboard.load') }}</span><strong>{{ formatPressure(frontPsi) }} <small>PSI</small></strong><em>{{ formatBar(frontBar) }}</em><p>{{ t('guidesTirePressure.dashboard.loadValue', { value: formatEngineeringNumber(frontLoad, 1) }) }} · {{ t('guidesTirePressure.dashboard.contactValue', { value: formatArea(frontArea) }) }}</p></div>
        <span class="status" :class="statusClass(frontMargin)">{{ marginLabel(frontMargin) }}</span>
      </article>
      <article class="hero-card rear">
        <div><b>{{ t('guidesTirePressure.dashboard.rearWheel') }}</b><span class="load-tag">{{ formatRatio(rearRatio) }} {{ t('guidesTirePressure.dashboard.load') }}</span><strong>{{ formatPressure(rearPsi) }} <small>PSI</small></strong><em>{{ formatBar(rearBar) }}</em><p>{{ t('guidesTirePressure.dashboard.loadValue', { value: formatEngineeringNumber(rearLoad, 1) }) }} · {{ t('guidesTirePressure.dashboard.contactValue', { value: formatArea(rearArea) }) }}</p></div>
        <span class="status" :class="statusClass(rearMargin)">{{ marginLabel(rearMargin) }}</span>
      </article>
    </section>

    <section class="stations">
      <TirePressureControls />

      <article class="station force-data">
        <header><b>{{ t('guidesTirePressure.dashboard.dynamicDataTitle') }}</b><span class="ok">● {{ t('guidesTirePressure.dashboard.dataViewStatus') }}</span></header>
        <div class="force-data-grid">
          <article class="force-wheel-card force-wheel-card--front">
            <header><b>{{ t('guidesTirePressure.dashboard.frontWheel') }}</b><span>{{ formatRatio(frontRatio) }} {{ t('guidesTirePressure.dashboard.load') }}</span></header>
            <div class="metric-grid">
              <div class="metric metric--primary"><span>{{ t('guidesTirePressure.dashboard.resultantForceMetric') }}</span><strong>{{ formatForce(backendFront?.resultant_contact_force_n ?? null) }} N</strong><small>{{ formatResultantMass(backendFront?.resultant_contact_force_n) }} · {{ formatResultantG(backendFront?.resultant_contact_force_n, backendFront?.vertical_load_n) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.verticalLoadMetric') }}</span><strong>{{ formatForce(backendFront?.vertical_load_n ?? null) }} N</strong><small>{{ formatEngineeringNumber(frontLoad, 1) }} kg</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.lateralDemandMetric') }}</span><strong>{{ formatForce(frontDemand) }} N</strong><small>{{ t('guidesTirePressure.dashboard.leanAngle') }} {{ formatEngineeringNumber(currentBackendDynamics?.dynamics.lean_angle_deg, 1) }}°</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.gripLimitMetric') }}</span><strong>{{ formatForce(frontMax) }} N</strong><small>μ {{ formatEngineeringNumber(backendFront?.mu_nominal, 2) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.gripMarginMetric') }}</span><strong>{{ marginLabel(frontMargin) }}</strong><small>{{ t('guidesTirePressure.dashboard.fixedMuLabel', { value: formatEngineeringNumber(backendFront?.mu_nominal, 2) }) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.contactAreaMetric') }}</span><strong>{{ formatArea(frontArea) }} cm²</strong><small>{{ formatPatchDimensions(backendFront?.estimated_contact_patch_width_mm, backendFront?.estimated_contact_patch_length_mm) }}</small></div>
            </div>
          </article>
          <article class="force-wheel-card force-wheel-card--rear">
            <header><b>{{ t('guidesTirePressure.dashboard.rearWheel') }}</b><span>{{ formatRatio(rearRatio) }} {{ t('guidesTirePressure.dashboard.load') }}</span></header>
            <div class="metric-grid">
              <div class="metric metric--primary"><span>{{ t('guidesTirePressure.dashboard.resultantForceMetric') }}</span><strong>{{ formatForce(backendRear?.resultant_contact_force_n ?? null) }} N</strong><small>{{ formatResultantMass(backendRear?.resultant_contact_force_n) }} · {{ formatResultantG(backendRear?.resultant_contact_force_n, backendRear?.vertical_load_n) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.verticalLoadMetric') }}</span><strong>{{ formatForce(backendRear?.vertical_load_n ?? null) }} N</strong><small>{{ formatEngineeringNumber(rearLoad, 1) }} kg</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.lateralDemandMetric') }}</span><strong>{{ formatForce(rearDemand) }} N</strong><small>{{ t('guidesTirePressure.dashboard.leanAngle') }} {{ formatEngineeringNumber(currentBackendDynamics?.dynamics.lean_angle_deg, 1) }}°</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.gripLimitMetric') }}</span><strong>{{ formatForce(rearMax) }} N</strong><small>μ {{ formatEngineeringNumber(backendRear?.mu_nominal, 2) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.gripMarginMetric') }}</span><strong>{{ marginLabel(rearMargin) }}</strong><small>{{ t('guidesTirePressure.dashboard.fixedMuLabel', { value: formatEngineeringNumber(backendRear?.mu_nominal, 2) }) }}</small></div>
              <div class="metric"><span>{{ t('guidesTirePressure.dashboard.contactAreaMetric') }}</span><strong>{{ formatArea(rearArea) }} cm²</strong><small>{{ formatPatchDimensions(backendRear?.estimated_contact_patch_width_mm, backendRear?.estimated_contact_patch_length_mm) }}</small></div>
            </div>
          </article>
        </div>
        <div class="grip-grid"><div class="meter"><div><b>{{ t('guidesTirePressure.dashboard.frontMargin') }}</b><strong>{{ marginLabel(frontMargin) }}</strong></div><div class="bar"><i :style="{ width: `${Math.max(5, Math.min(100, frontMargin ?? 0))}%`, background: '#0284c7' }"></i></div><small>{{ t('guidesTirePressure.dashboard.demandValue', { value: formatForce(frontDemand) }) }} · {{ t('guidesTirePressure.dashboard.maxGripValue', { value: formatForce(frontMax) }) }}</small></div><div class="meter"><div><b>{{ t('guidesTirePressure.dashboard.rearMargin') }}</b><strong>{{ marginLabel(rearMargin) }}</strong></div><div class="bar"><i :style="{ width: `${Math.max(5, Math.min(100, rearMargin ?? 0))}%`, background: '#d97706' }"></i></div><small>{{ t('guidesTirePressure.dashboard.demandValue', { value: formatForce(rearDemand) }) }} · {{ t('guidesTirePressure.dashboard.maxGripValue', { value: formatForce(rearMax) }) }}</small></div></div>
        <div class="diagnostic"><b>{{ t('guidesTirePressure.dashboard.diagnosticTitle') }}</b><span :class="statusClass(frontMargin === null || rearMargin === null ? null : Math.min(frontMargin, rearMargin))">{{ forceDemoStatus }}</span><p>{{ diagnosticText }}</p></div>
      </article>
    </section>

    <section class="science"><header><b>{{ t('guidesTirePressure.dashboard.scienceTitle') }}</b><span class="tag">{{ t('guidesTirePressure.dashboard.baseline') }}</span></header><div class="science-grid"><article><b>{{ t('guidesTirePressure.dashboard.mechanicalTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.mechanicalBody') }}</p></article><article class="danger"><b>{{ t('guidesTirePressure.dashboard.riskTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.riskBody') }}</p></article><article class="safe"><b>{{ t('guidesTirePressure.dashboard.scopeTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.scopeBody') }}</p></article></div></section>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, provide, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { ApiRequestError, useApiRequest } from '~/composables/useApiRequest'
import TirePressureControls from './TirePressureControls.vue'

const { t } = useI18n()
const { request } = useApiRequest()
const riderWeight = ref(70)
const bikeWeight = ref(8.5)
const leanAngle = ref(0)
const fixedFrontPsi = ref(50)
const fixedRearPsi = ref(55)
const speedKmh = ref(30)
const postureId = ref('race')
const tireWidth = ref(28)
const dynamicsState = ref<'PENDING' | 'BACKEND DEMO' | 'UNAVAILABLE'>('PENDING')
const backendDynamics = ref<BackendDynamics | null>(null)
const leanPresets = [0, 15, 28, 40]
const postureConfigs = [
  { id: 'race', key: 'race' },
  { id: 'endurance', key: 'endurance' },
  { id: 'tt', key: 'tt' },
  { id: 'gravel', key: 'gravel' },
]
const postures = computed(() => postureConfigs.map(item => ({ ...item, label: t(`guidesTirePressure.dashboard.${item.key}`) })))
const apiPosition = computed(() => ({ race: 'AGGRESSIVE_RACE', endurance: 'ENDURANCE', tt: 'TT_AERO', gravel: 'GRAVEL_TRAIL' }[postureId.value] || 'ENDURANCE'))
const dynamicsInputKey = computed(() => JSON.stringify({
  rider: riderWeight.value,
  bike: bikeWeight.value,
  tire: tireWidth.value,
  speed: speedKmh.value,
  position: apiPosition.value,
  lean: leanAngle.value,
  frontPsi: fixedFrontPsi.value,
  rearPsi: fixedRearPsi.value,
}))
const committedDynamicsInputKey = ref<string | null>(null)
const currentBackendDynamics = computed(() => committedDynamicsInputKey.value === dynamicsInputKey.value ? backendDynamics.value : null)
const totalWeight = computed(() => riderWeight.value + bikeWeight.value)
const backendFront = computed(() => currentBackendDynamics.value?.dynamics.front)
const backendRear = computed(() => currentBackendDynamics.value?.dynamics.rear)
const frontLoad = computed(() => currentBackendDynamics.value?.front_load_kg ?? null)
const rearLoad = computed(() => currentBackendDynamics.value?.rear_load_kg ?? null)
const frontRatio = computed(() => frontLoad.value === null ? null : Math.round(frontLoad.value / totalWeight.value * 100))
const rearRatio = computed(() => rearLoad.value === null ? null : Math.round(rearLoad.value / totalWeight.value * 100))
const frontPsi = computed(() => fixedFrontPsi.value)
const rearPsi = computed(() => fixedRearPsi.value)
const frontBar = computed(() => frontPsi.value === null ? null : frontPsi.value / 14.5037738)
const rearBar = computed(() => rearPsi.value === null ? null : rearPsi.value / 14.5037738)
const frontArea = computed(() => backendFront.value?.estimated_static_contact_area_cm2 ?? null)
const rearArea = computed(() => backendRear.value?.estimated_static_contact_area_cm2 ?? null)
const frontDemand = computed(() => backendFront.value?.lateral_demand_n ?? null)
const rearDemand = computed(() => backendRear.value?.lateral_demand_n ?? null)
const frontMax = computed(() => backendFront.value?.idealized_grip_limit_n ?? null)
const rearMax = computed(() => backendRear.value?.idealized_grip_limit_n ?? null)
const frontMargin = computed(() => backendFront.value?.grip_margin_pct ?? null)
const rearMargin = computed(() => backendRear.value?.grip_margin_pct ?? null)
const equivalentTurnRadiusM = computed(() => currentBackendDynamics.value?.dynamics.equivalent_turn_radius_m ?? null)
const lateralAccelerationG = computed(() => currentBackendDynamics.value?.dynamics.lateral_acceleration_g ?? null)
const demonstrationFrictionCoefficientNominal = computed(() => backendFront.value?.mu_nominal ?? null)
const tireBodyNormalizationFactor = computed(() => currentBackendDynamics.value?.dynamics.tire_body_normalization_factor ?? null)
const tireBodyNormalizationLabel = computed(() => tireBodyNormalizationFactor.value === null
  ? t('guidesTirePressure.dashboard.backendPending')
  : t('guidesTirePressure.dashboard.tireBodyBaseline', { value: tireBodyNormalizationFactor.value.toFixed(2) }))
const modelVersionLabel = computed(() => currentBackendDynamics.value?.dynamics.model_version ?? t('guidesTirePressure.dashboard.backendPending'))
const loadSourceLabel = computed(() => currentBackendDynamics.value?.load_source ?? t('guidesTirePressure.dashboard.backendPending'))
const tireWidthSourceLabel = computed(() => {
  const source = currentBackendDynamics.value?.dynamics.tire_width_source
  return source === 'MEASURED'
    ? t('guidesTirePressure.dashboard.measuredWidthSource')
    : source === 'NOMINAL_UNCORRECTED'
      ? t('guidesTirePressure.dashboard.nominalWidthSource')
      : t('guidesTirePressure.dashboard.backendPending')
})
const leanDesc = computed(() => t(`guidesTirePressure.dashboard.${leanAngle.value < 10 ? 'straight' : leanAngle.value < 20 ? 'gentle' : leanAngle.value < 32 ? 'fastCorner' : 'maximumLean'}`))
const postureLabel = computed(() => postures.value.find(item => item.id === postureId.value)?.label || '')
const leanPresetLabel = (preset: number) => preset === 0 ? t('guidesTirePressure.dashboard.straight') : preset === 15 ? t('guidesTirePressure.dashboard.gentle') : preset === 28 ? t('guidesTirePressure.dashboard.fastCorner') : t('guidesTirePressure.dashboard.maximumLean')
const forceDemoStatus = computed(() => dynamicsState.value === 'BACKEND DEMO' ? t('guidesTirePressure.dashboard.forceDemoStatus') : dynamicsState.value === 'UNAVAILABLE' ? t('guidesTirePressure.dashboard.backendUnavailable') : t('guidesTirePressure.dashboard.backendPending'))
const diagnosticText = computed(() => {
  if (frontDemand.value === null || rearDemand.value === null) return t('guidesTirePressure.dashboard.backendPending')
  return t('guidesTirePressure.dashboard.diagnosticBody', {
    speed: formatEngineeringNumber(currentBackendDynamics.value?.dynamics.speed_kmh, 1),
    lean: formatEngineeringNumber(currentBackendDynamics.value?.dynamics.lean_angle_deg, 1),
    radius: equivalentTurnRadiusM.value === null ? '∞' : formatEngineeringNumber(equivalentTurnRadiusM.value, 2),
    front: formatEngineeringNumber(frontDemand.value, 1),
    rear: formatEngineeringNumber(rearDemand.value, 1),
  })
})

type BackendWheelDynamics = {
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
  mu_nominal: number
}
type BackendDynamics = {
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
    front: BackendWheelDynamics
    rear: BackendWheelDynamics
  }
  warnings: string[]
}

const tirePressureDynamicsResponseCacheLimit = 32
const dynamicsResponseCache = new Map<string, BackendDynamics>()
let dynamicsSequence = 0
let dynamicsRequestInFlight = false
let dynamicsRefreshQueued = false
let dynamicsRetryNotBefore = 0
let isTirePressureCalculatorUnmounted = false

function readCachedTirePressureDynamics(inputKey: string) {
  const cachedDynamics = dynamicsResponseCache.get(inputKey)
  if (!cachedDynamics) return null

  // Promote recently reused inputs so the bounded cache behaves as an LRU.
  dynamicsResponseCache.delete(inputKey)
  dynamicsResponseCache.set(inputKey, cachedDynamics)
  return cachedDynamics
}

function cacheTirePressureDynamics(inputKey: string, dynamics: BackendDynamics) {
  dynamicsResponseCache.delete(inputKey)
  dynamicsResponseCache.set(inputKey, dynamics)
  while (dynamicsResponseCache.size > tirePressureDynamicsResponseCacheLimit) {
    const oldestInputKey = dynamicsResponseCache.keys().next().value
    if (oldestInputKey === undefined) break
    dynamicsResponseCache.delete(oldestInputKey)
  }
}

function scheduleTirePressureDynamicsRefresh(delayMilliseconds = 450) {
  if (dynamicsTimer) clearTimeout(dynamicsTimer)
  dynamicsState.value = 'PENDING'
  dynamicsTimer = setTimeout(() => void refreshTirePressureGroundFrameDynamicsFromBackend(), delayMilliseconds)
}

async function refreshTirePressureGroundFrameDynamicsFromBackend() {
  if (!import.meta.client || isTirePressureCalculatorUnmounted) return
  const inputKey = dynamicsInputKey.value

  const cachedDynamics = readCachedTirePressureDynamics(inputKey)
  if (cachedDynamics) {
    backendDynamics.value = cachedDynamics
    committedDynamicsInputKey.value = inputKey
    dynamicsState.value = 'BACKEND DEMO'
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
    rider_weight_kg: riderWeight.value,
    bike_weight_kg: bikeWeight.value,
    nominal_tire_width_mm: tireWidth.value,
    speed_kmh: speedKmh.value,
    riding_position: apiPosition.value,
    surface_condition: 'FLAT_ROAD',
    lean_angle_deg: leanAngle.value,
  }
  body.front_operating_pressure_psi = fixedFrontPsi.value
  body.rear_operating_pressure_psi = fixedRearPsi.value
  try {
    const response = await request<{ data: BackendDynamics }>('/engineering/tire-pressure/dynamics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (isTirePressureCalculatorUnmounted) return
    cacheTirePressureDynamics(inputKey, response.data)
    if (requestId === dynamicsSequence && inputKey === dynamicsInputKey.value) {
      backendDynamics.value = response.data
      committedDynamicsInputKey.value = inputKey
      dynamicsState.value = 'BACKEND DEMO'
    }
    dynamicsRetryNotBefore = 0
  } catch (error) {
    if (isTirePressureCalculatorUnmounted || requestId !== dynamicsSequence || inputKey !== dynamicsInputKey.value) return
    if (error instanceof ApiRequestError && error.status === 429) {
      dynamicsRetryNotBefore = Date.now() + 2000
      dynamicsRefreshQueued = true
    }
    // Retain the response object for cache reuse, but the committed input key
    // prevents an old response from being displayed for new parameters.
    dynamicsState.value = 'UNAVAILABLE'
  } finally {
    dynamicsRequestInFlight = false
    if (isTirePressureCalculatorUnmounted) return
    if (dynamicsRefreshQueued || inputKey !== dynamicsInputKey.value) {
      dynamicsRefreshQueued = false
      scheduleTirePressureDynamicsRefresh()
    }
  }
}

let dynamicsTimer: ReturnType<typeof setTimeout> | undefined
watch([riderWeight, bikeWeight, leanAngle, speedKmh, postureId, tireWidth, fixedFrontPsi, fixedRearPsi], () => {
  scheduleTirePressureDynamicsRefresh()
}, { immediate: true })

onBeforeUnmount(() => {
  isTirePressureCalculatorUnmounted = true
  if (dynamicsTimer) {
    clearTimeout(dynamicsTimer)
    dynamicsTimer = undefined
  }
  dynamicsRefreshQueued = false
})

function formatEngineeringNumber(value: number | null | undefined, decimals = 1) {
  return value === null || value === undefined || !Number.isFinite(value) ? '—' : value.toFixed(decimals)
}
function formatPressure(value: number | null) { return formatEngineeringNumber(value, 1) }
function formatBar(value: number | null) { return value === null ? '—' : `${formatEngineeringNumber(value, 2)} bar` }
function formatTurnRadius(value: number | null) { return value === null ? '∞' : `${formatEngineeringNumber(value, 2)} m` }
function formatArea(value: number | null) { return formatEngineeringNumber(value, 2) }
function formatForce(value: number | null) { return formatEngineeringNumber(value, 0) }
function formatRatio(value: number | null) { return value === null ? '—' : `${value}%` }
function formatResultantMass(value: number | undefined) { return value === undefined || !Number.isFinite(value) ? '—' : `${(value / 9.80665).toFixed(1)} kg` }
function formatResultantG(resultantForce: number | undefined, verticalLoad: number | undefined) {
  return resultantForce === undefined || verticalLoad === undefined || verticalLoad <= 0 || !Number.isFinite(resultantForce) || !Number.isFinite(verticalLoad)
    ? '—'
    : `${(resultantForce / verticalLoad).toFixed(2)} G`
}
function formatPatchDimensions(widthMm: number | undefined, lengthMm: number | undefined) {
  return widthMm === undefined || lengthMm === undefined || !Number.isFinite(widthMm) || !Number.isFinite(lengthMm)
    ? '—'
    : `${widthMm.toFixed(1)} × ${lengthMm.toFixed(1)} mm`
}
function marginLabel(value: number | null) {
  if (value === null) return t('guidesTirePressure.dashboard.backendPending')
  const state = value >= 25 ? 'safe' : value >= 8 ? 'warning' : 'critical'
  const prefix = value >= 8 ? '+' : ''
  return `${prefix}${value.toFixed(0)}% ${t(`guidesTirePressure.dashboard.${state}`)}`
}
function statusClass(value: number | null) {
  if (value === null) return 'pending'
  return value >= 25 ? 'safe' : value >= 8 ? 'warning' : 'danger'
}
provide('tirePressureModel', {
  totalWeight, leanAngle, leanDesc, leanPresets, leanPresetLabel, riderWeight, bikeWeight,
  frontLoad, rearLoad, frontRatio, rearRatio,
  postureLabel, postures, postureId, tireWidth,
  fixedFrontPsi, fixedRearPsi, speedKmh, equivalentTurnRadiusM, lateralAccelerationG, formatTurnRadius,
  demonstrationFrictionCoefficientNominal,
})
</script>

<style scoped>
.simulator {
  --simulator-border: var(--tz-border-subtle, #dbe3ec);
  --simulator-muted: var(--tz-text-secondary, #64748b);
  --simulator-surface: var(--tz-card-surface, #ffffff);
  --simulator-soft: var(--tz-surface-muted, #f8fafc);
  width: 100%;
  max-width: none;
  margin: 0;
  color: var(--tz-text-primary, #0f172a);
  text-align: left;
}

.topbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.25rem 0 0.65rem;
  border-bottom: 1px solid var(--simulator-border);
}

.topbar h2 {
  margin: 0;
  color: var(--tz-text-primary, #0f172a);
  font-size: clamp(1rem, 1.2vw, 1.28rem);
  line-height: 1.2;
  letter-spacing: -0.01em;
}

.topbar p {
  max-width: 34rem;
  margin: 0;
  color: var(--simulator-muted);
  font-size: 0.74rem;
  line-height: 1.45;
  text-align: right;
}

.notice {
  display: flex;
  align-items: flex-start;
  gap: 0.65rem;
  margin-top: 0.75rem;
  padding: 0.65rem 0.75rem;
  border: 1px solid #f5d28a;
  border-radius: 0.7rem;
  background: #fffaf0;
  color: #713f12;
}

.notice-mark {
  display: inline-flex;
  flex: 0 0 auto;
  width: 1.35rem;
  height: 1.35rem;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: #f59e0b;
  color: #ffffff;
  font-size: 0.75rem;
  font-weight: 800;
}

.notice-copy {
  min-width: 0;
}

.notice-copy strong {
  display: block;
  font-size: 0.76rem;
  line-height: 1.25;
}

.notice-copy p {
  margin: 0.2rem 0 0;
  color: #92400e;
  font-size: 0.7rem;
  line-height: 1.45;
}

.model-badge,
.model-provenance {
  display: inline-block;
  margin: 0.55rem 0 0;
  color: var(--simulator-muted);
  font-size: 0.64rem;
  letter-spacing: 0.04em;
  line-height: 1.3;
  text-transform: uppercase;
}

.model-provenance {
  margin-left: 0.65rem;
  letter-spacing: 0;
  text-transform: none;
}

.hero-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
  margin-top: 0.75rem;
}

.hero-card {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
  min-width: 0;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.75rem;
  background: var(--simulator-surface);
  box-shadow: 0 0.25rem 1rem rgb(15 23 42 / 0.045);
}

.hero-card > div {
  display: grid;
  grid-template-columns: auto 1fr;
  align-items: baseline;
  column-gap: 0.5rem;
  min-width: 0;
}

.hero-card b {
  overflow: hidden;
  color: var(--tz-text-primary, #0f172a);
  font-size: 0.78rem;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.hero-card .load-tag {
  justify-self: end;
  color: var(--simulator-muted);
  font-size: 0.64rem;
  white-space: nowrap;
}

.hero-card strong {
  margin-top: 0.22rem;
  color: var(--tz-text-primary, #0f172a);
  font-size: 1.25rem;
  line-height: 1;
}

.hero-card strong small {
  font-size: 0.58rem;
  font-weight: 700;
}

.hero-card em {
  align-self: end;
  margin: 0 0 0 0.35rem;
  color: var(--simulator-muted);
  font-size: 0.65rem;
  font-style: normal;
  white-space: nowrap;
}

.hero-card p {
  grid-column: 1 / -1;
  margin: 0.35rem 0 0;
  color: var(--simulator-muted);
  font-size: 0.64rem;
  line-height: 1.35;
}

.status {
  flex: 0 0 auto;
  border-radius: 999px;
  padding: 0.25rem 0.45rem;
  font-size: 0.61rem;
  font-weight: 700;
  line-height: 1.2;
  white-space: nowrap;
}

.status.safe {
  background: #dcfce7;
  color: #166534;
}

.status.warning {
  background: #fef3c7;
  color: #92400e;
}

.status.danger,
.status.pending {
  background: #fee2e2;
  color: #991b1b;
}

.stations {
  display: grid;
  grid-template-columns: minmax(20rem, 0.82fr) minmax(0, 1.18fr);
  gap: 0.8rem;
  align-items: start;
  margin-top: 0.8rem;
}

.station {
  min-width: 0;
  padding: 0.8rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.8rem;
  background: var(--simulator-surface);
  box-shadow: 0 0.3rem 1.2rem rgb(15 23 42 / 0.045);
}

.station.force-data {
  grid-column: 2;
  grid-row: 1;
}

.station.force-data > header,
.science > header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  color: var(--tz-text-primary, #0f172a);
  font-size: 0.8rem;
}

.station.force-data > header .ok {
  color: #047857;
  font-size: 0.65rem;
  font-weight: 700;
}

.force-data-grid {
  display: grid;
  gap: 0.65rem;
  margin-top: 0.65rem;
}

.force-wheel-card {
  min-width: 0;
  padding: 0.65rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.65rem;
  background: var(--simulator-soft);
}

.force-wheel-card--front {
  border-top: 3px solid #0284c7;
}

.force-wheel-card--rear {
  border-top: 3px solid #d97706;
}

.force-wheel-card > header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  color: var(--tz-text-primary, #0f172a);
  font-size: 0.75rem;
}

.force-wheel-card > header span {
  color: var(--simulator-muted);
  font-size: 0.64rem;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.45rem;
  margin-top: 0.5rem;
}

.metric {
  min-width: 0;
  padding: 0.48rem 0.5rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.5rem;
  background: var(--simulator-surface);
}

.metric span,
.metric strong,
.metric small {
  display: block;
}

.metric span {
  overflow: hidden;
  color: var(--simulator-muted);
  font-size: 0.61rem;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.metric strong {
  margin-top: 0.18rem;
  color: var(--tz-text-primary, #0f172a);
  font-size: 0.86rem;
  line-height: 1.15;
  overflow-wrap: anywhere;
}

.metric--primary strong {
  font-size: 1rem;
}

.metric small {
  margin-top: 0.2rem;
  color: var(--simulator-muted);
  font-size: 0.58rem;
  line-height: 1.25;
}

.grip-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.55rem;
  margin-top: 0.65rem;
}

.meter {
  min-width: 0;
  padding: 0.55rem 0.6rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.6rem;
  background: var(--simulator-soft);
}

.meter > div:first-child {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  font-size: 0.68rem;
}

.meter strong {
  color: var(--tz-text-primary, #0f172a);
  font-size: 0.64rem;
  white-space: nowrap;
}

.bar {
  height: 0.35rem;
  margin-top: 0.38rem;
  overflow: hidden;
  border-radius: 999px;
  background: #e2e8f0;
}

.bar i {
  display: block;
  height: 100%;
  border-radius: inherit;
}

.meter small {
  display: block;
  margin-top: 0.28rem;
  color: var(--simulator-muted);
  font-size: 0.59rem;
  line-height: 1.35;
}

.diagnostic {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 0.35rem 0.6rem;
  margin-top: 0.65rem;
  padding: 0.6rem;
  border-left: 3px solid #94a3b8;
  background: var(--simulator-soft);
}

.diagnostic b {
  font-size: 0.7rem;
}

.diagnostic span {
  align-self: start;
  font-size: 0.62rem;
  font-weight: 700;
  white-space: nowrap;
}

.diagnostic span.safe {
  color: #047857;
}

.diagnostic span.warning {
  color: #b45309;
}

.diagnostic span.danger,
.diagnostic span.pending {
  color: #b91c1c;
}

.diagnostic p {
  grid-column: 1 / -1;
  margin: 0;
  color: var(--simulator-muted);
  font-size: 0.67rem;
  line-height: 1.45;
}

.science {
  margin-top: 0.8rem;
  padding: 0.8rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.8rem;
  background: var(--simulator-surface);
}

.science .tag {
  border-radius: 999px;
  background: #e2e8f0;
  color: #475569;
  padding: 0.2rem 0.45rem;
  font-size: 0.6rem;
  font-weight: 700;
}

.science-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.55rem;
  margin-top: 0.6rem;
}

.science-grid article {
  padding: 0.6rem;
  border: 1px solid var(--simulator-border);
  border-radius: 0.6rem;
  background: var(--simulator-soft);
}

.science-grid article.danger {
  border-color: #fecaca;
  background: #fff7f7;
}

.science-grid article.safe {
  border-color: #bbf7d0;
  background: #f4fff7;
}

.science-grid b {
  display: block;
  font-size: 0.7rem;
}

.science-grid p {
  margin: 0.25rem 0 0;
  color: var(--simulator-muted);
  font-size: 0.65rem;
  line-height: 1.45;
}

@media (max-width: 960px) {
  .stations {
    grid-template-columns: 1fr;
  }

  .station.force-data {
    grid-column: auto;
    grid-row: auto;
  }
}

@media (max-width: 640px) {
  .topbar {
    display: block;
  }

  .topbar p {
    max-width: none;
    margin-top: 0.35rem;
    text-align: left;
  }

  .model-badge,
  .model-provenance {
    display: block;
    margin-left: 0;
  }

  .hero-grid,
  .grip-grid,
  .science-grid {
    grid-template-columns: 1fr;
  }

  .hero-card {
    padding: 0.65rem;
  }

  .station {
    padding: 0.65rem;
  }
}
</style>
