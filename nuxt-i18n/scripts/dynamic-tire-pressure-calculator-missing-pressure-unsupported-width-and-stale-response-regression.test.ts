import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { setTimeout as waitForTimer } from 'node:timers/promises'
import ts from 'typescript'
import * as vue from 'vue'
import type {
  TirePressureCalculatorDynamicsRequestInputs,
  TirePressureCalculatorDynamicsState,
} from '../app/composables/useTirePressureCalculatorDynamicsRequest.js'

// Exercise the real client transport with Vue reactivity. Nuxt's API context
// and component disposal hook are the only dependencies replaced by this harness.
const dynamicsComposableSource = await readFile(new URL(
  '../app/composables/useTirePressureCalculatorDynamicsRequest.ts', import.meta.url,
), 'utf8')
const compiledClientDynamicsComposable = ts.transpileModule(
  dynamicsComposableSource.replaceAll('import.meta.client', 'true'),
  { compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022 } },
).outputText

const wheelDynamicsFixture = {
  load_kg: 35,
  vertical_load_n: 343.23,
  lateral_demand_n: 0,
  idealized_grip_limit_n: 240.26,
  grip_margin_pct: 100,
  resultant_contact_force_n: 343.23,
  mu_nominal: 0.7,
}
const backendDynamicsEnvelope = {
  code: 0,
  data: {
    model_version: 'cornering-demo-v7-pressure-friction-status',
    load_source: 'POSTURE_BASELINE',
    front_load_kg: 35,
    rear_load_kg: 43.5,
    dynamics: {
      model_version: 'cornering-demo-v7-pressure-friction-status',
      model_status: 'DEMONSTRATION',
      force_frame: 'GROUND_FRAME',
      lean_angle_deg: 0,
      speed_kmh: 30,
      lateral_acceleration_mps2: 0,
      lateral_acceleration_g: 0,
      surface_condition: 'FLAT_ROAD',
      tire_body_normalization_factor: 1,
      tire_body_normalization_source: 'NORMALIZED_BASELINE',
      tire_width_mm: 28,
      tire_width_source: 'NOMINAL_UNCORRECTED',
      front: wheelDynamicsFixture,
      rear: wheelDynamicsFixture,
    },
    warnings: [],
  },
}

const submittedDynamicsBodies: Record<string, unknown>[] = []
const componentDisposalCallbacks: Array<() => void> = []
let failNextDynamicsRequest = false
let pauseNextDynamicsRequest = false
const pausedDynamicsRequest = { resolve: null as ((response: unknown) => void) | null }

async function requestTirePressureDynamicsForRegression(path: string, init: { body: string }) {
  assert.equal(path, '/engineering/tire-pressure/dynamics')
  submittedDynamicsBodies.push(JSON.parse(init.body))
  if (failNextDynamicsRequest) {
    failNextDynamicsRequest = false
    throw new Error('Simulated backend failure')
  }
  if (pauseNextDynamicsRequest) {
    pauseNextDynamicsRequest = false
    return new Promise(resolve => { pausedDynamicsRequest.resolve = resolve })
  }
  return backendDynamicsEnvelope
}

const composableModuleExports = {} as typeof import('../app/composables/useTirePressureCalculatorDynamicsRequest.js')
new Function('require', 'exports', compiledClientDynamicsComposable)((moduleName: string) => {
  if (moduleName === 'vue') return {
    ...vue,
    onBeforeUnmount: (callback: () => void) => componentDisposalCallbacks.push(callback),
  }
  if (moduleName === '~/composables/useApiRequest') return {
    ApiRequestError: class extends Error {},
    useApiRequest: () => ({ request: requestTirePressureDynamicsForRegression }),
  }
  throw new Error(`Unexpected transport dependency: ${moduleName}`)
}, composableModuleExports)

const selectedArticleNumber = vue.ref<string | null>(null)
const selectedWidthSupported = vue.ref(true)
const manualOperatingPressureInputAllowed = vue.ref(false)
const referencePressurePsi = vue.ref<number | null>(null)
const pressureBoundsPsi = vue.ref<{ minimumPsi: number; maximumPsi: number } | null>(null)
const inputs: TirePressureCalculatorDynamicsRequestInputs = {
  riderWeight: vue.ref(70),
  bikeWeight: vue.ref(8.5),
  tireWidth: vue.ref(28),
  speedKmh: vue.ref(30),
  apiPosition: vue.computed(() => 'AGGRESSIVE_RACE'),
  leanAngle: vue.ref(0),
  selectedTireArticleNo: vue.computed(() => selectedArticleNumber.value),
  selectedTireWidthSupported: vue.computed(() => selectedWidthSupported.value),
  manualOperatingPressureInputAllowed: vue.computed(() => manualOperatingPressureInputAllowed.value),
  selectedTireReferencePressurePsi: vue.computed(() => referencePressurePsi.value),
  selectedTirePressureBoundsPsi: vue.computed(() => pressureBoundsPsi.value),
  pressureComparisonEnabled: vue.ref(false),
  wetPressureScenarioEnabled: vue.ref(false),
  pressureComparisonAvailable: vue.computed(() => pressureBoundsPsi.value !== null),
  frontComparisonPressurePsi: vue.computed(() => inputs.pressureComparisonEnabled.value ? pressureBoundsPsi.value?.minimumPsi ?? null : null),
  rearComparisonPressurePsi: vue.computed(() => inputs.pressureComparisonEnabled.value ? pressureBoundsPsi.value?.minimumPsi ?? null : null),
  wetPressureWaterFilmDepthMm: 1,
}
const regressionEffectScope = vue.effectScope()
const dynamicsTransport = regressionEffectScope.run(() => (
  composableModuleExports.useTirePressureCalculatorDynamicsRequest(inputs)
))!

async function waitForTirePressureDynamicsState(expectedState: TirePressureCalculatorDynamicsState) {
  const deadline = Date.now() + 4000
  while (dynamicsTransport.dynamicsState.value !== expectedState && Date.now() < deadline) {
    await waitForTimer(20)
  }
  assert.equal(dynamicsTransport.dynamicsState.value, expectedState)
}

try {
  assert.equal(dynamicsTransport.dynamicsState.value, 'MODEL_REQUIRED')
  selectedArticleNumber.value = 'supported-width-without-minimum-pressure'
  await vue.nextTick()
  assert.equal(dynamicsTransport.dynamicsState.value, 'PRESSURE_REFERENCE_UNAVAILABLE')
  await waitForTimer(500)
  assert.equal(submittedDynamicsBodies.length, 0)

  selectedArticleNumber.value = '11654264'
  selectedWidthSupported.value = false
  await vue.nextTick()
  assert.equal(dynamicsTransport.dynamicsState.value, 'TIRE_WIDTH_UNSUPPORTED')
  referencePressurePsi.value = 20
  pressureBoundsPsi.value = { minimumPsi: 10, maximumPsi: 30 }
  await vue.nextTick()
  await waitForTimer(500)
  assert.equal(submittedDynamicsBodies.length, 0, 'Unsupported width must not reuse the previous 28 mm input')

  selectedArticleNumber.value = 'supported-width-with-complete-pressure-range'
  selectedWidthSupported.value = true
  referencePressurePsi.value = 60
  pressureBoundsPsi.value = { minimumPsi: 40, maximumPsi: 80 }
  await waitForTirePressureDynamicsState('BACKEND DEMO')
  assert.equal(submittedDynamicsBodies.length, 1)
  assert.equal(submittedDynamicsBodies[0]?.front_operating_pressure_psi, 60)
  assert.equal(submittedDynamicsBodies[0]?.rear_operating_pressure_psi, 60)
  assert.equal(dynamicsTransport.currentBackendDynamics.value?.front_load_kg, 35)

  inputs.pressureComparisonEnabled.value = true
  await waitForTirePressureDynamicsState('PENDING')
  await waitForTirePressureDynamicsState('BACKEND DEMO')
  assert.equal(submittedDynamicsBodies.at(-1)?.front_comparison_pressure_psi, 40)
  assert.equal(submittedDynamicsBodies.at(-1)?.rear_comparison_pressure_psi, 40)

  pauseNextDynamicsRequest = true
  inputs.riderWeight.value = 71
  await waitForTirePressureDynamicsState('PENDING')
  await waitForTimer(500)
  assert.ok(pausedDynamicsRequest.resolve, 'The request should be in flight before clearing the model')
  selectedArticleNumber.value = null
  referencePressurePsi.value = null
  pressureBoundsPsi.value = null
  await vue.nextTick()
  assert.equal(dynamicsTransport.dynamicsState.value, 'MODEL_REQUIRED')
  assert.equal(dynamicsTransport.currentBackendDynamics.value, null)
  pausedDynamicsRequest.resolve(backendDynamicsEnvelope)
  await waitForTimer(500)
  assert.equal(dynamicsTransport.dynamicsState.value, 'MODEL_REQUIRED', 'Late results must not restore a cleared model')
  assert.equal(dynamicsTransport.currentBackendDynamics.value, null)

  failNextDynamicsRequest = true
  inputs.riderWeight.value = 72
  inputs.pressureComparisonEnabled.value = false
  selectedArticleNumber.value = 'supported-width-with-complete-pressure-range'
  referencePressurePsi.value = 60
  pressureBoundsPsi.value = { minimumPsi: 40, maximumPsi: 80 }
  await waitForTirePressureDynamicsState('UNAVAILABLE')
  dynamicsTransport.retryTirePressureDynamics()
  await waitForTirePressureDynamicsState('BACKEND DEMO')

  selectedArticleNumber.value = 'supported-width-without-pressure-range'
  referencePressurePsi.value = null
  pressureBoundsPsi.value = null
  await vue.nextTick()
  assert.equal(dynamicsTransport.dynamicsState.value, 'PRESSURE_REFERENCE_UNAVAILABLE')
  assert.equal(dynamicsTransport.currentBackendDynamics.value, null, 'A blocked selection must clear prior results')
  manualOperatingPressureInputAllowed.value = true
  inputs.tireWidth.value = 120
  await vue.nextTick()
  assert.equal(dynamicsTransport.dynamicsState.value, 'OPERATING_PRESSURE_REQUIRED')
  referencePressurePsi.value = 10
  await waitForTirePressureDynamicsState('BACKEND DEMO')
  assert.equal(submittedDynamicsBodies.at(-1)?.nominal_tire_width_mm, 120)
  assert.equal(submittedDynamicsBodies.at(-1)?.front_operating_pressure_psi, 10)
  assert.equal(submittedDynamicsBodies.at(-1)?.rear_operating_pressure_psi, 10)
  assert.ok(!('front_minimum_pressure_psi' in submittedDynamicsBodies.at(-1)!))
  assert.ok(!('front_comparison_pressure_psi' in submittedDynamicsBodies.at(-1)!))
  console.log('Tire-pressure missing-reference, unsupported-width, calculation, comparison, stale-response, and retry regressions passed.')
} finally {
  for (const disposeComponent of componentDisposalCallbacks) disposeComponent()
  regressionEffectScope.stop()
}
