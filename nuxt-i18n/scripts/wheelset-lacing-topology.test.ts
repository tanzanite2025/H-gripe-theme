import assert from 'node:assert/strict'
import {
  getSupportedWheelsetLacingCrossCounts,
  type WheelsetLacingHoleSelection,
} from '../app/utils/wheelsetLacingSelectionContract.ts'

const expectedCrossCountsByHoleSelection: ReadonlyArray<readonly [WheelsetLacingHoleSelection, readonly number[]]> = [
  [16, [0, 1]],
  [20, [0, 1, 2]],
  [21, [2]],
  [24, [0, 1, 2, 3]],
  [28, [0, 1, 2, 3]],
  [32, [0, 1, 2, 3, 4]],
  [36, [0, 1, 2, 3, 4]],
  ['24_2to1', [2]],
]

for (const [holeSelection, expectedCrossCounts] of expectedCrossCountsByHoleSelection) {
  assert.deepEqual(
    getSupportedWheelsetLacingCrossCounts(holeSelection),
    expectedCrossCounts,
    `supported cross counts for ${String(holeSelection)}H`,
  )
}

console.log('wheelset lacing selection contract tests passed')
