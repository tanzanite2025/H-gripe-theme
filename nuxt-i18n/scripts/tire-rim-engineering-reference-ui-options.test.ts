import assert from 'node:assert/strict'
import {
  TIRE_RIM_ENGINEERING_MAX_RIM_INNER_WIDTH_MM,
  TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM,
  TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM,
  TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM,
} from '../app/data/tireguides/tireRimEngineeringReferenceUiOptions.js'

assert.deepEqual(
  [...TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM],
  Array.from(
    { length: TIRE_RIM_ENGINEERING_MAX_RIM_INNER_WIDTH_MM - TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM + 1 },
    (_, index) => TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM + index,
  ),
)
assert.equal(TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM[0], 13)
assert.equal(
  TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM.at(-1),
  100,
)
assert.deepEqual(
  [...TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM],
  [18, 20, 23, 25, 28, 30, 32, 35, 38, 40, 45, 50, 55, 60, 70, 80, 90, 100, 110, 115, 120, 127, 132],
)

console.log('Tire/rim engineering workbench UI options are stable.')
