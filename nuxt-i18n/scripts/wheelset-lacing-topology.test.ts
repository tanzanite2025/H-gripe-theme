import assert from 'node:assert/strict'
import {
  assertWheelsetLacingTopology,
  buildWheelsetLacingTopology,
  getSupportedWheelsetLacingCrossCounts,
} from '../app/utils/wheelsetLacingTopology.ts'
import { calculateWheelsetLacingGeometryProjectionMetrics } from '../app/utils/wheelsetLacingGeometryProjectionMetrics.ts'

const dimensions = {
  rimRadius: 232,
  flangeRadiusA: 66,
  flangeRadiusB: 54,
}

assert.deepEqual(getSupportedWheelsetLacingCrossCounts(16), [0, 1])
assert.deepEqual(getSupportedWheelsetLacingCrossCounts(21), [2])
assert.deepEqual(getSupportedWheelsetLacingCrossCounts('24_2to1'), [2])

const assertOneToOne = (values: number[], expectedCount: number) => {
  assert.equal(values.length, expectedCount)
  assert.equal(new Set(values).size, expectedCount)
}

for (const [holeCount, supportedCrossCounts] of [
  [16, [0, 1]],
  [20, [0, 1, 2]],
  [24, [0, 1, 2, 3]],
  [28, [0, 1, 2, 3]],
  [32, [0, 1, 2, 3, 4]],
  [36, [0, 1, 2, 3, 4]],
] as const) {
  for (const crossCount of supportedCrossCounts) {
    const topology = buildWheelsetLacingTopology(holeCount, crossCount, dimensions)
    assert.equal(topology.spokes.length, holeCount, `${holeCount}H ${crossCount}X spoke count`)
  }
}

const symmetric = buildWheelsetLacingTopology(24, 2, dimensions)
assert.equal(symmetric.distribution, 'symmetric_1to1')
assert.equal(symmetric.rimHoles.length, 24)
assert.equal(symmetric.hubHolesA.length, 12)
assert.equal(symmetric.hubHolesB.length, 12)
assert.equal(symmetric.spokes.length, 24)
assertOneToOne(symmetric.spokes.map(spoke => spoke.rim.id), 24)
assertOneToOne(symmetric.spokes.map(spoke => `${spoke.side}:${spoke.hub.id}` as unknown as number), 24)
assert.ok(Math.abs(calculateWheelsetLacingGeometryProjectionMetrics(symmetric).aggregateMeanAbsoluteProjectionAngleDegrees - 76.025) < 0.01)

const symmetricDriveSideTangentialComponents = symmetric.spokes
  .filter(spoke => spoke.side === 'A')
  .map((spoke) => {
    const hubRadius = Math.hypot(spoke.hub.x, spoke.hub.y)
    const spokeX = spoke.rim.x - spoke.hub.x
    const spokeY = spoke.rim.y - spoke.hub.y
    const spokeLength = Math.hypot(spokeX, spokeY)
    return ((-spoke.hub.y * spokeX) + (spoke.hub.x * spokeY)) / (hubRadius * spokeLength)
  })
assert.ok(Math.abs(symmetricDriveSideTangentialComponents.reduce((sum, value) => sum + value, 0)) < 1e-10)
assert.ok(symmetricDriveSideTangentialComponents.some(value => value < 0))
assert.ok(symmetricDriveSideTangentialComponents.some(value => value > 0))

const uniformTwoToOne = buildWheelsetLacingTopology('24_2to1', 2, dimensions)
assert.equal(uniformTwoToOne.distribution, 'uniform_2to1')
assert.equal(uniformTwoToOne.rimHoles.length, 24)
assert.equal(uniformTwoToOne.hubHolesA.length, 16)
assert.equal(uniformTwoToOne.hubHolesB.length, 8)
assert.equal(uniformTwoToOne.spokes.length, 24)
assert.equal(uniformTwoToOne.spokes.filter(spoke => spoke.side === 'A').length, 16)
assert.equal(uniformTwoToOne.spokes.filter(spoke => spoke.side === 'B').length, 8)
assertOneToOne(uniformTwoToOne.spokes.map(spoke => spoke.rim.id), 24)
const uniformGeometryProjectionMetrics = calculateWheelsetLacingGeometryProjectionMetrics(uniformTwoToOne)
assert.ok(Math.abs(uniformGeometryProjectionMetrics.aggregateMeanAbsoluteProjectionAngleDegrees - 54.228) < 0.01)
assert.ok(uniformGeometryProjectionMetrics.minimumAbsoluteDriveSideProjectionAngleDegrees < uniformGeometryProjectionMetrics.maximumAbsoluteDriveSideProjectionAngleDegrees)
assert.ok(Math.abs(
  uniformGeometryProjectionMetrics.aggregateMeanAbsoluteProjectionAngleDegrees
  - Math.atan2(
    uniformGeometryProjectionMetrics.meanAbsoluteTangentialProjectionPercent,
    uniformGeometryProjectionMetrics.meanAbsoluteRadialProjectionPercent,
  ) * (180 / Math.PI),
) < 1e-10)
assert.ok(uniformGeometryProjectionMetrics.meanAbsoluteTangentialProjectionPercent >= 0)
assert.ok(uniformGeometryProjectionMetrics.meanAbsoluteRadialProjectionPercent >= 0)

// The 24H 2:1 rim remains evenly drilled. The A-A-B side assignment is the
// distribution rule; it must not be represented as G3-style grouped holes.
const rimAngles = uniformTwoToOne.rimHoles.map(hole => hole.angle)
const rimStep = (Math.PI * 2) / 24
for (let index = 1; index < rimAngles.length; index += 1) {
  assert.ok(Math.abs((rimAngles[index] - rimAngles[index - 1]) - rimStep) < 1e-10)
}

const g3 = buildWheelsetLacingTopology(21, 2, dimensions)
assert.equal(g3.distribution, 'g3_2to1')
assert.equal(g3.rimHoles.length, 21)
assert.equal(g3.hubHolesA.length, 14)
assert.equal(g3.hubHolesB.length, 7)
assert.equal(g3.spokes.length, 21)
assert.equal(g3.spokes.filter(spoke => spoke.side === 'A').length, 14)
assert.equal(g3.spokes.filter(spoke => spoke.side === 'B').length, 7)
assert.ok(Math.abs(calculateWheelsetLacingGeometryProjectionMetrics(g3).aggregateMeanAbsoluteProjectionAngleDegrees - 45.385) < 0.01)

for (let group = 0; group < 7; group += 1) {
  const holes = g3.rimHoles.slice(group * 3, group * 3 + 3)
  assert.deepEqual(holes.map(hole => hole.side), ['A', 'B', 'A'])
}

assert.throws(
  () => buildWheelsetLacingTopology(21, 1, dimensions),
  /requires 2X/,
)
assert.throws(
  () => buildWheelsetLacingTopology('24_2to1', 0, dimensions),
  /requires 2X/,
)

const invalidCoordinateTopology = structuredClone(symmetric)
invalidCoordinateTopology.rimHoles[0].x = Number.NaN
assert.throws(() => assertWheelsetLacingTopology(invalidCoordinateTopology), /finite angle and coordinates/)

const duplicateSpokeIdTopology = structuredClone(symmetric)
duplicateSpokeIdTopology.spokes[1].id = duplicateSpokeIdTopology.spokes[0].id
assert.throws(() => assertWheelsetLacingTopology(duplicateSpokeIdTopology), /invalid or duplicate ID/)

const foreignRimReferenceTopology = structuredClone(symmetric)
foreignRimReferenceTopology.spokes[0].rim = { ...foreignRimReferenceTopology.spokes[0].rim, x: 999 }
assert.throws(() => assertWheelsetLacingTopology(foreignRimReferenceTopology), /references a rim hole outside/)

assert.throws(
  () => buildWheelsetLacingTopology(24, 2, { ...dimensions, flangeRadiusA: Number.POSITIVE_INFINITY }),
  /finite positive number/,
)

console.log('wheelset lacing topology tests passed')
