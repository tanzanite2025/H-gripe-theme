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
    <p class="model-badge">{{ t('guidesTirePressure.dashboard.baseline') }} · {{ t('guidesTirePressure.dashboard.scopeTitle') }}</p>

    <section class="hero-grid">
      <article class="hero-card front">
        <div><b>{{ t('guidesTirePressure.dashboard.frontWheel') }}</b><span class="load-tag">{{ frontRatio }}% {{ t('guidesTirePressure.dashboard.load') }}</span><strong>{{ frontPsi.toFixed(1) }} <small>PSI</small></strong><em>{{ frontBar.toFixed(2) }} bar</em><p>{{ t('guidesTirePressure.dashboard.loadValue', { value: frontLoad.toFixed(1) }) }} · {{ t('guidesTirePressure.dashboard.contactValue', { value: frontAreaDisplay.toFixed(2) }) }}</p></div>
        <span class="status" :class="statusClass(frontMargin)">{{ marginLabel(frontMargin) }}</span>
      </article>
      <article class="hero-card rear">
        <div><b>{{ t('guidesTirePressure.dashboard.rearWheel') }}</b><span class="load-tag">{{ rearRatio }}% {{ t('guidesTirePressure.dashboard.load') }}</span><strong>{{ rearPsi.toFixed(1) }} <small>PSI</small></strong><em>{{ rearBar.toFixed(2) }} bar</em><p>{{ t('guidesTirePressure.dashboard.loadValue', { value: rearLoad.toFixed(1) }) }} · {{ t('guidesTirePressure.dashboard.contactValue', { value: rearAreaDisplay.toFixed(2) }) }}</p></div>
        <span class="status" :class="statusClass(rearMargin)">{{ marginLabel(rearMargin) }}</span>
      </article>
    </section>

    <section class="stations">
      <TirePressureControls />

      <article class="station cad">
        <p class="dynamics-state">{{ dynamicsState === 'BACKEND DEMO' ? t('guidesTirePressure.dashboard.scopeTitle') : t('guidesTirePressure.dashboard.baseline') }}</p>
        <header><b>{{ t('guidesTirePressure.dashboard.cadTitle') }}</b><span class="ok">● {{ t('guidesTirePressure.dashboard.viewStatus') }}</span></header>
        <TirePressureCad />
        <div class="contact-explainer">
          <b>{{ t('guidesTirePressure.dashboard.contactAreaLabel') }}</b>
          <p>{{ t('guidesTirePressure.dashboard.contactAreaDescription') }}</p>
          <div class="contact-formulas">
            <code>{{ t('guidesTirePressure.dashboard.contactFormula', { wheel: t('guidesTirePressure.dashboard.frontWheel'), load: (frontLoad * 9.80665).toFixed(0), pressure: (frontPsi * 6.894757293).toFixed(0), area: frontArea.toFixed(2) }) }}</code>
            <code>{{ t('guidesTirePressure.dashboard.contactFormula', { wheel: t('guidesTirePressure.dashboard.rearWheel'), load: (rearLoad * 9.80665).toFixed(0), pressure: (rearPsi * 6.894757293).toFixed(0), area: rearArea.toFixed(2) }) }}</code>
          </div>
        </div>
        <div class="grip-grid"><div class="meter"><div><b>{{ t('guidesTirePressure.dashboard.frontMargin') }}</b><strong>{{ marginLabel(frontMargin) }}</strong></div><div class="bar"><i :style="{ width: Math.max(5, Math.min(100, frontMargin)) + '%', background: '#0284c7' }"></i></div><small>{{ t('guidesTirePressure.dashboard.demandValue', { value: Math.round(frontDemandDisplay) }) }} · {{ t('guidesTirePressure.dashboard.maxGripValue', { value: Math.round(frontMax) }) }}</small></div><div class="meter"><div><b>{{ t('guidesTirePressure.dashboard.rearMargin') }}</b><strong>{{ marginLabel(rearMargin) }}</strong></div><div class="bar"><i :style="{ width: Math.max(5, Math.min(100, rearMargin)) + '%', background: '#d97706' }"></i></div><small>{{ t('guidesTirePressure.dashboard.demandValue', { value: Math.round(rearDemandDisplay) }) }} · {{ t('guidesTirePressure.dashboard.maxGripValue', { value: Math.round(rearMax) }) }}</small></div></div>
        <div class="diagnostic"><b>{{ t('guidesTirePressure.dashboard.diagnosticTitle') }}</b><span :class="statusClass(Math.min(frontMargin, rearMargin))">{{ impedanceStatus }}</span><p>{{ diagnosticText }}</p></div>
      </article>
    </section>

    <section class="science"><header><b>{{ t('guidesTirePressure.dashboard.scienceTitle') }}</b><span class="tag">{{ t('guidesTirePressure.dashboard.baseline') }}</span></header><div class="science-grid"><article><b>{{ t('guidesTirePressure.dashboard.mechanicalTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.mechanicalBody') }}</p></article><article class="danger"><b>{{ t('guidesTirePressure.dashboard.riskTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.riskBody') }}</p></article><article class="safe"><b>{{ t('guidesTirePressure.dashboard.scopeTitle') }}</b><p>{{ t('guidesTirePressure.dashboard.scopeBody') }}</p></article></div></section>
  </div>
</template>

<script setup lang="ts">
import { computed, provide, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import TirePressureControls from './TirePressureControls.vue'
import TirePressureCad from './TirePressureCad.vue'

const { t } = useI18n()
const { request } = useApiRequest()
const riderWeight = ref(70)
const bikeWeight = ref(8.5)
const leanAngle = ref(0)
const pressureMode = ref<'auto' | 'fixed'>('auto')
const fixedFrontPsi = ref(50)
const fixedRearPsi = ref(55)
const postureId = ref('race')
const surfaceId = ref('good')
const weather = ref<'dry' | 'wet'>('dry')
const rimSystem = ref<'hookless' | 'hooked'>('hookless')
const tireWidth = ref(28)
const rimWidth = ref(25)
const dynamicsState = ref('LOCAL DEMO')
const backendDynamics = ref<BackendDynamics | null>(null)
const leanPresets = [0, 15, 28, 40]
const postureConfigs = [{ id: 'race', key: 'race', front: .47 }, { id: 'endurance', key: 'endurance', front: .45 }, { id: 'tt', key: 'tt', front: .48 }, { id: 'gravel', key: 'gravel', front: .44 }]
const surfaceConfigs = [{ id: 'track', key: 'track', mu: .88, muLow: .70, muHigh: 1 }, { id: 'good', key: 'goodAsphalt', mu: .76, muLow: .58, muHigh: .90 }, { id: 'coarse', key: 'chipSeal', mu: .76, muLow: .58, muHigh: .90 }, { id: 'cobbles', key: 'cobbles', mu: .64, muLow: .42, muHigh: .82 }, { id: 'gravel', key: 'gravelRoad', mu: .52, muLow: .30, muHigh: .70 }]
const postures = computed(() => postureConfigs.map(item => ({ ...item, label: `${t(`guidesTirePressure.dashboard.${item.key}`)} ${Math.round(item.front * 100)}:${100 - Math.round(item.front * 100)}` })))
const surfaces = computed(() => surfaceConfigs.map(item => ({ ...item, label: t(`guidesTirePressure.dashboard.${item.key}`) })))
const posture = computed(() => postures.value.find(item => item.id === postureId.value) || postures.value[0]!)
const surface = computed(() => surfaces.value.find(item => item.id === surfaceId.value) || surfaces.value[1]!)
const totalWeight = computed(() => riderWeight.value + bikeWeight.value)
const frontRatio = computed(() => Math.round(posture.value.front * 100))
const rearRatio = computed(() => 100 - frontRatio.value)
const frontLoad = computed(() => totalWeight.value * posture.value.front)
const rearLoad = computed(() => totalWeight.value - frontLoad.value)
const measuredWidth = computed(() => tireWidth.value + .4 * (rimWidth.value - 19))
const frontPsi = computed(() => pressureMode.value === 'fixed' ? fixedFrontPsi.value : calcPsi(frontLoad.value))
const rearPsi = computed(() => pressureMode.value === 'fixed' ? fixedRearPsi.value : calcPsi(rearLoad.value))
const frontBar = computed(() => frontPsi.value / 14.5038)
const rearBar = computed(() => rearPsi.value / 14.5038)
const frontArea = computed(() => contact(frontLoad.value, frontPsi.value))
const rearArea = computed(() => contact(rearLoad.value, rearPsi.value))
const frontDemand = computed(() => frontLoad.value * 9.80665 * Math.tan(leanAngle.value * Math.PI / 180))
const rearDemand = computed(() => rearLoad.value * 9.80665 * Math.tan(leanAngle.value * Math.PI / 180))
const frontMu = computed(() => demoMu(surface.value, weather.value, frontPsi.value))
const rearMu = computed(() => demoMu(surface.value, weather.value, rearPsi.value))
const localFrontMax = computed(() => frontLoad.value * 9.80665 * frontMu.value.nominal)
const localRearMax = computed(() => rearLoad.value * 9.80665 * rearMu.value.nominal)
const localFrontMargin = computed(() => margin(localFrontMax.value, frontDemand.value))
const localRearMargin = computed(() => margin(localRearMax.value, rearDemand.value))
const backendFront = computed(() => backendDynamics.value?.dynamics.front)
const backendRear = computed(() => backendDynamics.value?.dynamics.rear)
// Keep the displayed contact area synchronized with the current controls.
// Backend dynamics responses may arrive after the inputs have already changed.
const frontAreaDisplay = computed(() => frontArea.value)
const rearAreaDisplay = computed(() => rearArea.value)
const frontDemandDisplay = computed(() => backendFront.value?.lateral_demand_n ?? frontDemand.value)
const rearDemandDisplay = computed(() => backendRear.value?.lateral_demand_n ?? rearDemand.value)
const frontMax = computed(() => backendFront.value?.idealized_grip_limit_n ?? localFrontMax.value)
const rearMax = computed(() => backendRear.value?.idealized_grip_limit_n ?? localRearMax.value)
const frontMargin = computed(() => backendFront.value?.grip_margin_pct ?? localFrontMargin.value)
const rearMargin = computed(() => backendRear.value?.grip_margin_pct ?? localRearMargin.value)
const leanDesc = computed(() => t(`guidesTirePressure.dashboard.${leanAngle.value < 10 ? 'straight' : leanAngle.value < 20 ? 'gentle' : leanAngle.value < 32 ? 'fastCorner' : 'maximumLean'}`))
const postureLabel = computed(() => posture.value.label)
const surfaceLabel = computed(() => surface.value.label)
const leanPresetLabel = (preset: number) => preset === 0 ? t('guidesTirePressure.dashboard.straight') : preset === 15 ? t('guidesTirePressure.dashboard.gentle') : preset === 28 ? t('guidesTirePressure.dashboard.fastCorner') : t('guidesTirePressure.dashboard.maximumLean')
const impedanceStatus = computed(() => dynamicsState.value === 'BACKEND DEMO' ? t('guidesTirePressure.dashboard.scopeTitle') : t('guidesTirePressure.dashboard.baseline'))
const diagnosticText = computed(() => t('guidesTirePressure.dashboard.diagnosticBody', { lean: leanAngle.value, front: Math.round(frontDemand.value), rear: Math.round(rearDemand.value) }))
const wheels = computed(() => {
  const radians = leanAngle.value * Math.PI / 180
  const cosine = Math.cos(radians)
  const entries = [
    { id: 'front', x: 145, color: '#0284c7', load: frontLoad.value, psi: frontPsi.value, area: frontAreaDisplay.value, demand: frontDemandDisplay.value },
    { id: 'rear', x: 395, color: '#d97706', load: rearLoad.value, psi: rearPsi.value, area: rearAreaDisplay.value, demand: rearDemandDisplay.value },
  ]

  return entries.map(wheel => {
    const gRatio = 1 / Math.max(.5, cosine)
    const tireHalfWidth = clamp(measuredWidth.value * .72, 18, 26)
    const airHalfWidth = tireHalfWidth * .68
    const rimHalfWidth = tireHalfWidth * .48
    const rimRadius = 51
    const demandLength = Math.min(68, wheel.demand * .12)

    return {
      ...wheel,
      geometry: {
        centerY: 150,
        tireRadius: 70,
        tireHalfWidth,
        airHalfWidth,
        rimHalfWidth,
        rimRadius,
        axleHalfLength: tireHalfWidth + 14,
        contactX: wheel.x,
        contactShadow: clamp(measuredWidth.value * .42, 11, 18) / Math.max(.7, cosine),
      },
      force: {
        resultantX: wheel.x,
        resultantKg: (wheel.id === 'front' ? backendFront.value?.resultant_contact_force_n : backendRear.value?.resultant_contact_force_n)
          ? ((wheel.id === 'front' ? backendFront.value?.resultant_contact_force_n : backendRear.value?.resultant_contact_force_n) || 0) / 9.80665
          : wheel.load * gRatio,
        gRatio,
        demand: wheel.demand,
        demandLength,
        demandStartX: wheel.x - demandLength / 2,
        demandEndX: wheel.x + demandLength / 2,
        demandLabelX: wheel.x,
        demandLabelY: 288,
        axisY: 298,
        maxGrip: wheel.id === 'front' ? frontMax.value : rearMax.value,
        gripLength: Math.min(68, (wheel.id === 'front' ? frontMax.value : rearMax.value) * .12),
      },
      patch: patchGeometry(wheel),
    }
  })
})

type BackendWheelDynamics = {
  lateral_demand_n: number
  idealized_grip_limit_n: number
  grip_margin_pct: number
  resultant_contact_force_n: number
  estimated_static_contact_area_cm2?: number
}
type BackendDynamics = { dynamics: { front: BackendWheelDynamics; rear: BackendWheelDynamics } }

const apiPosition = computed(() => ({ race: 'AGGRESSIVE_RACE', endurance: 'ENDURANCE', tt: 'TT_AERO', gravel: 'GRAVEL_TRAIL' }[postureId.value] || 'ENDURANCE'))
const apiSurface = computed(() => ({ track: 'SMOOTH_TRACK', good: 'ROUGH_CHIP', coarse: 'ROUGH_CHIP', cobbles: 'COBBLES', gravel: 'UNPAVED_GRAVEL' }[surfaceId.value] || 'ROUGH_CHIP'))
const dynamicsInputKey = computed(() => JSON.stringify({
  rider: riderWeight.value,
  bike: bikeWeight.value,
  tire: tireWidth.value,
  rim: rimWidth.value,
  rimSystem: rimSystem.value,
  position: apiPosition.value,
  surface: apiSurface.value,
  weather: weather.value,
  lean: leanAngle.value,
  frontPsi: frontPsi.value,
  rearPsi: rearPsi.value,
}))

let dynamicsSequence = 0
let dynamicsController: AbortController | null = null
async function refreshDynamics() {
  if (!import.meta.client) return
  const requestId = ++dynamicsSequence
  const inputKey = dynamicsInputKey.value
  dynamicsController?.abort()
  dynamicsController = new AbortController()
  dynamicsState.value = 'BACKEND DEMO'
  try {
    const response = await request<{ data: BackendDynamics }>('/engineering/tire-pressure/dynamics', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      signal: dynamicsController.signal,
      body: JSON.stringify({
        rider_weight_kg: riderWeight.value,
        bike_weight_kg: bikeWeight.value,
        nominal_tire_width_mm: tireWidth.value,
        inner_rim_width_mm: rimWidth.value,
        rim_system: rimSystem.value === 'hookless' ? 'HOOKLESS' : 'HOOKED',
        riding_position: apiPosition.value,
        surface_condition: apiSurface.value,
        casing_type: 'TUBELESS',
        weather_condition: weather.value === 'wet' ? 'WET' : 'DRY',
        lean_angle_deg: leanAngle.value,
        front_operating_pressure_psi: frontPsi.value,
        rear_operating_pressure_psi: rearPsi.value,
      }),
    })
    if (requestId !== dynamicsSequence || inputKey !== dynamicsInputKey.value) return
    backendDynamics.value = response.data
  } catch (error) {
    if (requestId !== dynamicsSequence || inputKey !== dynamicsInputKey.value || (error instanceof DOMException && error.name === 'AbortError')) return
    backendDynamics.value = null
    dynamicsState.value = 'LOCAL DEMO'
  }
}

let dynamicsTimer: ReturnType<typeof setTimeout> | undefined
watch([riderWeight, bikeWeight, leanAngle, postureId, surfaceId, weather, rimSystem, tireWidth, rimWidth, pressureMode, fixedFrontPsi, fixedRearPsi], () => {
  if (dynamicsTimer) clearTimeout(dynamicsTimer)
  dynamicsTimer = setTimeout(() => void refreshDynamics(), 120)
}, { immediate: true })

function calcPsi(load: number) {
  // The empirical constant is not calibrated for production; this is a visible demo estimate.
  return 135.2 * Math.pow(load, 1.1) / Math.pow(Math.max(1, measuredWidth.value), 1.45)
}
function contact(load: number, psi: number) {
  const pressurePa = Math.max(1, psi * 6894.757293)
  return load * 9.80665 / pressurePa * 10000
}
function demoMu(surfaceItem: { mu: number; muLow: number; muHigh: number }, weatherValue: 'dry' | 'wet', _psi: number) {
  if (weatherValue === 'wet') return { nominal: surfaceItem.mu * .65, low: surfaceItem.muLow * .45, high: surfaceItem.muHigh * .8 }
  return { nominal: surfaceItem.mu, low: surfaceItem.muLow, high: surfaceItem.muHigh }
}
function margin(max: number, demand: number) { return max <= 0 ? 0 : ((max - demand) / max) * 100 }
function clamp(value: number, min: number, max: number) { return Math.min(max, Math.max(min, value)) }
  function marginLabel(value: number) {
    const state = value >= 25 ? 'safe' : value >= 8 ? 'warning' : 'critical'
    const prefix = value >= 8 ? '+' : ''
    return `${prefix}${value.toFixed(0)}% ${t(`guidesTirePressure.dashboard.${state}`)}`
  }
  function statusClass(value: number) { return value >= 25 ? 'safe' : value >= 8 ? 'warning' : 'danger' }
function patchGeometry(wheel: { id: string; x: number; area: number }) {
    const leanRadians = leanAngle.value * Math.PI / 180
    const halfWidthMm = measuredWidth.value * (wheel.id === 'rear' ? .38 : .36) * Math.max(.45, Math.cos(leanRadians * .75))
    const halfLengthMm = wheel.area * 100 / (Math.PI * halfWidthMm)
    return {
      cx: wheel.x + Math.sin(leanRadians) * 24,
      // The SVG ellipse is dimensioned from the same area equation shown to users:
      // A = pi * semi-length * semi-width, with mm² converted to cm².
      // Keep the drawing readable while retaining the physical dimensions in the label.
      rx: halfLengthMm * .65,
      ry: halfWidthMm * .65,
      rotate: leanAngle.value * .45,
      lengthCm: halfLengthMm * 2 / 10,
      widthCm: halfWidthMm * 2 / 10,
      state: leanAngle.value > 0
        ? t('guidesTirePressure.dashboard.shoulderLeanState', { angle: leanAngle.value })
        : t('guidesTirePressure.dashboard.straightCenterState'),
  }
}

provide('tirePressureModel', {
  t, totalWeight, leanAngle, leanDesc, leanPresets, leanPresetLabel, riderWeight, bikeWeight,
  postureLabel, postures, postureId, surfaceLabel, surfaces, surfaceId, weather, rimSystem, tireWidth, rimWidth,
  wheels, pressureMode, fixedFrontPsi, fixedRearPsi,
})
</script>

<style scoped>
.contact-explainer{margin-top:.8rem;padding:.75rem;border:1px solid var(--line);border-radius:.7rem;background:#f8fafc;font-size:.78rem}.contact-explainer p{margin:.35rem 0;color:var(--muted);line-height:1.5}.contact-formulas{display:grid;gap:.3rem}.contact-formulas code{display:block;overflow-wrap:anywhere;color:#334155;font-size:.7rem}
 .control-box > b{display:block}.control-box > span{display:block;margin-top:.2rem}.model-badge,.dynamics-state{margin:.45rem 0;color:var(--muted);font-size:.68rem;letter-spacing:.04em;text-transform:uppercase}.dynamics-state{margin:0 0 .45rem;color:#047857;font-weight:700}
</style>
