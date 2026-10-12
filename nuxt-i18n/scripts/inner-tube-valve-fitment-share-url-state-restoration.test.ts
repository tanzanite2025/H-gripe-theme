import assert from 'node:assert/strict'
import { parseQuery } from 'vue-router'
import { createServerRenderedInnerTubeValveFitmentMatrixFallback } from '../app/data/tireguides/innerTubeValveFitmentServerRenderedFallback.ts'
import {
  createInnerTubeValveFitmentShareQueryParameters,
  hasInnerTubeValveFitmentShareQueryParameters,
  readInnerTubeValveFitmentShareQueryParameterSignature,
  restoreInnerTubeValveFitmentStateFromShareQuery,
  type InnerTubeValveFitmentShareState,
} from '../app/utils/innerTubeValveFitmentShareUrlState.ts'

const fitmentMetadata = createServerRenderedInnerTubeValveFitmentMatrixFallback()
const defaultState: InnerTubeValveFitmentShareState = {
  rimDepth: 50, fitmentMode: 'automatic', manualValveLength: 60, manualExtenderLength: 0,
  pumpHeadGripDepth: 15, rimDepthUncertainty: 2,
}
const manualState: InnerTubeValveFitmentShareState = {
  rimDepth: 80, fitmentMode: 'manual', manualValveLength: 80, manualExtenderLength: 30,
  pumpHeadGripDepth: 22, rimDepthUncertainty: 4,
}

for (const state of [defaultState, manualState, { ...manualState, fitmentMode: 'automatic' as const }]) {
  const query = parseQuery(new URLSearchParams(createInnerTubeValveFitmentShareQueryParameters(state)).toString())
  assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(query, fitmentMetadata), state, 'Shared calculations must round-trip every input')
  assert.ok(hasInnerTubeValveFitmentShareQueryParameters(query))
}

assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery({}, fitmentMetadata), defaultState)
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(parseQuery('rim_depth=60'), fitmentMetadata), {
  ...defaultState, rimDepth: 60,
}, 'Partial or historical URLs must not inherit the previous manual calculation')
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(parseQuery(
  'rim_depth=&fitment_mode=invalid&valve_length=77&extender_length=Infinity&pump_head_grip_depth=%20&rim_depth_uncertainty=NaN',
), fitmentMetadata), defaultState, 'Empty and unsupported values must use defaults, not produce invalid backend inputs')
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(parseQuery(
  'rim_depth=1000&pump_head_grip_depth=-50&rim_depth_uncertainty=99',
), fitmentMetadata), { ...defaultState, rimDepth: 100, pumpHeadGripDepth: 10, rimDepthUncertainty: 5 })
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(parseQuery(
  'rim_depth=61.6&rim_depth=80&pump_head_grip_depth=19.4&rim_depth_uncertainty=0',
), fitmentMetadata), { ...defaultState, rimDepth: 62, pumpHeadGripDepth: 19, rimDepthUncertainty: 0 })
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery(parseQuery('rim_depth&valve_length'), fitmentMetadata), defaultState)

assert.equal(hasInnerTubeValveFitmentShareQueryParameters({ utm_source: 'forum' }), false)
assert.equal(hasInnerTubeValveFitmentShareQueryParameters({ rim_depth: null }), true)
assert.notEqual(readInnerTubeValveFitmentShareQueryParameterSignature({}), readInnerTubeValveFitmentShareQueryParameterSignature({ rim_depth: null }))
assert.equal(
  readInnerTubeValveFitmentShareQueryParameterSignature({ rim_depth: '60', utm_source: 'forum' }),
  readInnerTubeValveFitmentShareQueryParameterSignature({ rim_depth: '60', utm_source: 'email' }),
  'Tracking parameters must not trigger calculator state restoration',
)

const customizedDefaultsMetadata = { ...fitmentMetadata, default_pump_head_grip_depth_mm: 20, default_rim_depth_uncertainty_mm: 3 }
assert.deepEqual(restoreInnerTubeValveFitmentStateFromShareQuery({}, customizedDefaultsMetadata), {
  ...defaultState, pumpHeadGripDepth: 20, rimDepthUncertainty: 3,
}, 'Defaults must follow the backend model metadata')

console.log('Inner-tube valve fitment share URL state restoration tests passed.')
