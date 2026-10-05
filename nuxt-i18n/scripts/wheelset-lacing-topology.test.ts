import assert from 'node:assert/strict'
import {
  getSupportedWheelsetLacingCrossCounts,
  type WheelsetLacingHoleSelection,
} from '../app/utils/wheelsetLacingSelectionContract.ts'
import {
  resolveWheelsetLacingDisplayGeometryTopologySelection,
  validateWheelsetLacingDisplayGeometryResponse,
  WHEELSET_LACING_DISPLAY_GEOMETRY_CONTRACT_VERSION,
  WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT,
  type WheelsetLacingDisplayGeometryLayout,
} from '../app/utils/wheelsetLacingDisplayGeometryContract.ts'

const expectedCrossCountsByHoleSelection: ReadonlyArray<readonly [WheelsetLacingHoleSelection, readonly number[]]> = [
  [16, [0, 1]],
  [20, [0, 1, 2]],
  [21, [2]],
  [24, [0, 1, 2, 3]],
  [28, [0, 1, 2, 3]],
  [32, [0, 1, 2, 3, 4]],
  [36, [0, 1, 2, 3, 4]],
  ['18_2to1', [2]],
  ['24_2to1', [2]],
]

for (const [holeSelection, expectedCrossCounts] of expectedCrossCountsByHoleSelection) {
  assert.deepEqual(
    getSupportedWheelsetLacingCrossCounts(holeSelection),
    expectedCrossCounts,
    `supported cross counts for ${String(holeSelection)}H`,
  )
}

assert.deepEqual(
  resolveWheelsetLacingDisplayGeometryTopologySelection(21, 2),
  { topologyId: '21h-g3-2to1', displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1 },
)
assert.deepEqual(
  resolveWheelsetLacingDisplayGeometryTopologySelection('24_2to1', 2),
  { topologyId: '24h-uniform-2to1', displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1 },
)
assert.deepEqual(
  resolveWheelsetLacingDisplayGeometryTopologySelection('18_2to1', 2),
  { topologyId: '18h-uniform-2to1', displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1 },
)
assert.deepEqual(
  resolveWheelsetLacingDisplayGeometryTopologySelection(24, 2),
  { topologyId: '24h-symmetric-1to1-2x', displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.symmetric1To1 },
)
assert.throws(
  () => resolveWheelsetLacingDisplayGeometryTopologySelection(16, 2),
  /unsupported 16 wheelset lacing cross count 2/,
)
assert.throws(
  () => resolveWheelsetLacingDisplayGeometryTopologySelection('18_2to1', 1),
  /unsupported 18_2to1 wheelset lacing cross count 1/,
)
assert.throws(
  () => resolveWheelsetLacingDisplayGeometryTopologySelection(19 as WheelsetLacingHoleSelection, 2),
  /unregistered wheelset lacing display topology/,
)
assert.throws(
  () => validateWheelsetLacingDisplayGeometryResponse(
    {
      contract_version: WHEELSET_LACING_DISPLAY_GEOMETRY_CONTRACT_VERSION,
      display_layout: 'future_18h_2to1',
      topology: {
        topology_id: '18h-2to1',
        display_layout: 'future_18h_2to1',
      },
    },
    {
      topologyId: '18h-2to1',
      displayLayout: 'future_18h_2to1' as WheelsetLacingDisplayGeometryLayout,
    },
  ),
  /unsupported display geometry layout/,
)

console.log('wheelset lacing selection contract tests passed')
