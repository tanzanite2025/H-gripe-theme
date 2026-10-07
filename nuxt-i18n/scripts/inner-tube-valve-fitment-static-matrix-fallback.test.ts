import assert from 'node:assert/strict'
import { createServerRenderedInnerTubeValveFitmentMatrixFallback } from '../app/data/tireguides/innerTubeValveFitmentServerRenderedFallback.ts'

const fallbackMatrix = createServerRenderedInnerTubeValveFitmentMatrixFallback()
const fiftyMillimetreRimRow = fallbackMatrix.rows.find(row => row.rim_depth_mm === 50)
const oneHundredMillimetreRimRow = fallbackMatrix.rows.find(row => row.rim_depth_mm === 100)

assert.ok(fiftyMillimetreRimRow, 'the SSR matrix should include the common 50 mm rim depth')
assert.equal(fiftyMillimetreRimRow.passage_depth_mm, 43.5)
assert.equal(fiftyMillimetreRimRow.minimum_required_length_mm, 65)
assert.deepEqual(
  fiftyMillimetreRimRow.clearances.map(cell => [cell.valve_length_mm, cell.effective_exposure_mm, cell.status]),
  [
    [40, -3.5, 'unsafe'],
    [48, 4.5, 'unsafe'],
    [60, 16.5, 'marginal'],
    [80, 36.5, 'optimal'],
  ],
)
assert.deepEqual(
  [
    fiftyMillimetreRimRow.recommended_result.recommendation.valve_length_mm,
    fiftyMillimetreRimRow.recommended_result.recommendation.extender_length_mm,
  ],
  [80, 0],
  'a 50 mm rim should recommend an 80 mm native valve with no extender',
)

assert.ok(oneHundredMillimetreRimRow, 'the SSR matrix should include the supported 100 mm rim depth')
assert.deepEqual(
  [
    oneHundredMillimetreRimRow.recommended_result.recommendation.valve_length_mm,
    oneHundredMillimetreRimRow.recommended_result.recommendation.extender_length_mm,
  ],
  [80, 40],
)
assert.equal(fallbackMatrix.rows.length, fallbackMatrix.preset_rim_depths_mm.length)

console.log('Inner-tube valve fitment static matrix fallback tests passed.')
