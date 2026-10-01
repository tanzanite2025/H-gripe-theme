export type LacingHoleSelection =
  | 16
  | 20
  | 21
  | 24
  | 28
  | 32
  | 36
  | '24_2to1'

export type LacingSide = 'A' | 'B'
export type LacingSpokeType = 'leading' | 'trailing' | 'nondrive'
export type LacingDistribution = 'symmetric_1to1' | 'uniform_2to1' | 'g3_2to1'

export interface LacingTopologyDimensions {
  rimRadius: number
  flangeRadiusA: number
  flangeRadiusB: number
}

export interface LacingPoint {
  id: number
  side: LacingSide
  role?: LacingSpokeType | 'radial'
  angle: number
  x: number
  y: number
}

export interface LacingSpoke {
  id: number
  side: LacingSide
  type: LacingSpokeType
  hub: LacingPoint
  rim: LacingPoint
}

export interface WheelsetLacingTopology {
  selection: LacingHoleSelection
  cross: number
  distribution: LacingDistribution
  rimHoles: LacingPoint[]
  hubHolesA: LacingPoint[]
  hubHolesB: LacingPoint[]
  spokes: LacingSpoke[]
}

const STANDARD_CROSS_RULES: Record<number, readonly number[]> = {
  16: [0, 1],
  20: [0, 1, 2],
  24: [0, 1, 2, 3],
  28: [0, 1, 2, 3],
  32: [0, 1, 2, 3, 4],
  36: [0, 1, 2, 3, 4],
}

/**
 * Returns the supported preview cross counts for one topology selection.
 * The page uses this function for its controls so the UI cannot drift from
 * the same selection contract enforced by buildWheelsetLacingTopology().
 */
export const getSupportedWheelsetLacingCrossCounts = (
  holes: LacingHoleSelection,
): readonly number[] => {
  if (holes === 21 || holes === '24_2to1') return [2]
  return STANDARD_CROSS_RULES[holes] || []
}

const assertFinitePositive = (value: number, label: string) => {
  if (!Number.isFinite(value) || value <= 0) {
    throw new Error(`[CRITICAL] ${label} must be a finite positive number. Received: ${value}`)
  }
}

const assertSelection = (holes: LacingHoleSelection, cross: number) => {
  if (!Number.isInteger(cross) || cross < 0 || cross > 4) {
    throw new Error(`[CRITICAL] Invalid lacing cross count: ${cross}. Expected an integer from 0 to 4.`)
  }

  const allowedCross = getSupportedWheelsetLacingCrossCounts(holes)
  if (!allowedCross || !allowedCross.includes(cross)) {
    if (holes === 21 || holes === '24_2to1') {
      throw new Error(`[CRITICAL] ${holes === 21 ? '21H G3' : '24H uniform 2:1'} requires 2X on the crossed side.`)
    }
    throw new Error(`[CRITICAL] Unsupported lacing combination: ${String(holes)}H ${cross}X.`)
  }
}

const pointOnCircle = (
  id: number,
  side: LacingSide,
  role: LacingPoint['role'],
  angle: number,
  radius: number,
): LacingPoint => ({
  id,
  side,
  role,
  angle,
  x: radius * Math.cos(angle),
  y: radius * Math.sin(angle),
})

const requiredPointAt = (
  points: readonly LacingPoint[],
  index: number,
  label: string,
): LacingPoint => {
  const point = points[index]
  if (point === undefined) {
    throw new Error(`[CRITICAL] ${label} point ${index} is missing.`)
  }
  return point
}

const buildG3Topology = (
  dimensions: LacingTopologyDimensions,
): Pick<WheelsetLacingTopology, 'rimHoles' | 'hubHolesA' | 'hubHolesB' | 'spokes'> => {
  const groups = 7
  // Display-only half-spacing for the two A-side holes in each G3 triplet.
  // This is a canvas proportion, not a measured rim drilling dimension.
  const G3_RIM_GROUP_HALF_SPACING_RADIANS = 0.085
  const rimHoles: LacingPoint[] = []
  const hubHolesA: LacingPoint[] = []
  const hubHolesB: LacingPoint[] = []
  const spokes: LacingSpoke[] = []

  for (let group = 0; group < groups; group += 1) {
    const centerAngle = (group * 2 * Math.PI) / groups - Math.PI / 2
    rimHoles.push(
      pointOnCircle(group * 3, 'A', 'trailing', centerAngle - G3_RIM_GROUP_HALF_SPACING_RADIANS, dimensions.rimRadius),
      pointOnCircle(group * 3 + 1, 'B', 'radial', centerAngle, dimensions.rimRadius),
      pointOnCircle(group * 3 + 2, 'A', 'leading', centerAngle + G3_RIM_GROUP_HALF_SPACING_RADIANS, dimensions.rimRadius),
    )
  }

  for (let group = 0; group < groups; group += 1) {
    const centerAngle = (group * 2 * Math.PI) / groups - Math.PI / 2
    const midAngle = centerAngle + Math.PI / groups
    hubHolesA.push(
      pointOnCircle(group * 2, 'A', undefined, midAngle - Math.PI / 14, dimensions.flangeRadiusA),
      pointOnCircle(group * 2 + 1, 'A', undefined, midAngle + Math.PI / 14, dimensions.flangeRadiusA),
    )
  }

  for (let group = 0; group < groups; group += 1) {
    const angle = (group * 2 * Math.PI) / groups - Math.PI / 2
    hubHolesB.push(pointOnCircle(group, 'B', undefined, angle, dimensions.flangeRadiusB))
  }

  for (let group = 0; group < groups; group += 1) {
    spokes.push({
      id: spokes.length,
      side: 'B',
      type: 'nondrive',
      hub: requiredPointAt(hubHolesB, group, 'G3 flange B'),
      rim: requiredPointAt(rimHoles, group * 3 + 1, 'G3 rim'),
    })
  }

  for (let group = 0; group < groups; group += 1) {
    const nextGroup = (group + 1) % groups
    spokes.push({
      id: spokes.length,
      side: 'A',
      type: 'trailing',
      hub: requiredPointAt(hubHolesA, group * 2, 'G3 flange A'),
      rim: requiredPointAt(rimHoles, nextGroup * 3, 'G3 rim'),
    })
    spokes.push({
      id: spokes.length,
      side: 'A',
      type: 'leading',
      hub: requiredPointAt(hubHolesA, group * 2 + 1, 'G3 flange A'),
      rim: requiredPointAt(rimHoles, group * 3 + 2, 'G3 rim'),
    })
  }

  return { rimHoles, hubHolesA, hubHolesB, spokes }
}

const buildUniformTwoToOneTopology = (
  dimensions: LacingTopologyDimensions,
): Pick<WheelsetLacingTopology, 'rimHoles' | 'hubHolesA' | 'hubHolesB' | 'spokes'> => {
  const total = 24
  const rimHoles: LacingPoint[] = []
  const hubHolesA: LacingPoint[] = []
  const hubHolesB: LacingPoint[] = []
  const spokes: LacingSpoke[] = []

  // The rim drilling is evenly spaced. The 2:1 distribution is represented by
  // the repeating A-A-B assignment around the rim, not by grouped G3 holes.
  for (let index = 0; index < total; index += 1) {
    const angle = (index * 2 * Math.PI) / total - Math.PI / 2 + Math.PI / total
    const side: LacingSide = index % 3 === 1 ? 'B' : 'A'
    rimHoles.push(pointOnCircle(index, side, undefined, angle, dimensions.rimRadius))
  }

  const bRimHoles = rimHoles.filter(hole => hole.side === 'B')
  const aRimHoles = rimHoles.filter(hole => hole.side === 'A')

  for (let index = 0; index < 16; index += 1) {
    const angle = (index * 2 * Math.PI) / 16 - Math.PI / 2
    hubHolesA.push(pointOnCircle(index, 'A', undefined, angle, dimensions.flangeRadiusA))
  }

  for (let index = 0; index < 8; index += 1) {
    const rimHole = requiredPointAt(bRimHoles, index, '24H 2:1 rim B')
    hubHolesB.push(pointOnCircle(index, 'B', undefined, rimHole.angle, dimensions.flangeRadiusB))
  }

  for (let index = 0; index < 8; index += 1) {
    spokes.push({
      id: spokes.length,
      side: 'B',
      type: 'nondrive',
      hub: requiredPointAt(hubHolesB, index, '24H 2:1 flange B'),
      rim: requiredPointAt(bRimHoles, index, '24H 2:1 rim B'),
    })
  }

  for (let index = 0; index < 16; index += 1) {
    const hub = requiredPointAt(hubHolesA, index, '24H 2:1 flange A')
    const targetIndex = index % 2 === 0
      ? (index + 2) % 16
      : (index - 2 + 16) % 16
    spokes.push({
      id: spokes.length,
      side: 'A',
      type: index % 2 === 0 ? 'leading' : 'trailing',
      hub,
      rim: requiredPointAt(aRimHoles, targetIndex, '24H 2:1 rim A'),
    })
  }

  return { rimHoles, hubHolesA, hubHolesB, spokes }
}

const buildSymmetricTopology = (
  holes: Exclude<LacingHoleSelection, 21 | '24_2to1'>,
  cross: number,
  dimensions: LacingTopologyDimensions,
): Pick<WheelsetLacingTopology, 'rimHoles' | 'hubHolesA' | 'hubHolesB' | 'spokes'> => {
  const total = holes
  const flangeCount = total / 2
  const rimHoles: LacingPoint[] = []
  const hubHolesA: LacingPoint[] = []
  const hubHolesB: LacingPoint[] = []
  const spokes: LacingSpoke[] = []

  for (let index = 0; index < total; index += 1) {
    const angle = (index * 2 * Math.PI) / total - Math.PI / 2 + Math.PI / total
    const side: LacingSide = index % 2 === 0 ? 'A' : 'B'
    rimHoles.push(pointOnCircle(index, side, undefined, angle, dimensions.rimRadius))
  }

  const rimA = rimHoles.filter(hole => hole.side === 'A')
  const rimB = rimHoles.filter(hole => hole.side === 'B')

  for (let index = 0; index < flangeCount; index += 1) {
    const rimHoleA = requiredPointAt(rimA, index, 'Symmetric rim A')
    const rimHoleB = requiredPointAt(rimB, index, 'Symmetric rim B')
    hubHolesA.push(pointOnCircle(index, 'A', undefined, rimHoleA.angle, dimensions.flangeRadiusA))
    hubHolesB.push(pointOnCircle(index, 'B', undefined, rimHoleB.angle, dimensions.flangeRadiusB))
  }

  if (cross === 0) {
    for (let index = 0; index < flangeCount; index += 1) {
      spokes.push(
        {
          id: spokes.length,
          side: 'A',
          type: 'leading',
          hub: requiredPointAt(hubHolesA, index, 'Symmetric flange A'),
          rim: requiredPointAt(rimA, index, 'Symmetric rim A'),
        },
        {
          id: spokes.length + 1,
          side: 'B',
          type: 'nondrive',
          hub: requiredPointAt(hubHolesB, index, 'Symmetric flange B'),
          rim: requiredPointAt(rimB, index, 'Symmetric rim B'),
        },
      )
    }
    return { rimHoles, hubHolesA, hubHolesB, spokes }
  }

  for (let index = 0; index < flangeCount; index += 1) {
    const isLeading = index % 2 === 0
    const targetIndex = isLeading
      ? (index + cross) % flangeCount
      : (index - cross + flangeCount) % flangeCount
    spokes.push({
      id: spokes.length,
      side: 'A',
      type: isLeading ? 'leading' : 'trailing',
      hub: requiredPointAt(hubHolesA, index, 'Symmetric flange A'),
      rim: requiredPointAt(rimA, targetIndex, 'Symmetric rim A'),
    })
  }

  for (let index = 0; index < flangeCount; index += 1) {
    const isLeading = index % 2 === 0
    const targetIndex = isLeading
      ? (index + cross) % flangeCount
      : (index - cross + flangeCount) % flangeCount
    spokes.push({
      id: spokes.length,
      side: 'B',
      type: 'nondrive',
      hub: requiredPointAt(hubHolesB, index, 'Symmetric flange B'),
      rim: requiredPointAt(rimB, targetIndex, 'Symmetric rim B'),
    })
  }

  return { rimHoles, hubHolesA, hubHolesB, spokes }
}

const assertLacingPointCollection = (
  points: LacingPoint[],
  label: string,
  expectedSide?: LacingSide,
) => {
  const seen = new Set<number>()
  points.forEach((point, index) => {
    if (!Number.isInteger(point.id) || point.id < 0) {
      throw new Error(`[CRITICAL] ${label} point ${index} must have a non-negative integer ID.`)
    }
    if (seen.has(point.id)) {
      throw new Error(`[CRITICAL] ${label} contains duplicate hole ID: ${point.id}`)
    }
    seen.add(point.id)
    if (point.side !== 'A' && point.side !== 'B') {
      throw new Error(`[CRITICAL] ${label} point ${point.id} has an unsupported side: ${point.side}`)
    }
    if (expectedSide && point.side !== expectedSide) {
      throw new Error(`[CRITICAL] ${label} point ${point.id} must belong to side ${expectedSide}.`)
    }
    if (!Number.isFinite(point.angle) || !Number.isFinite(point.x) || !Number.isFinite(point.y)) {
      throw new Error(`[CRITICAL] ${label} point ${point.id} must have finite angle and coordinates.`)
    }
  })
}

export const assertWheelsetLacingTopology = (topology: WheelsetLacingTopology) => {
  if (!topology || typeof topology !== 'object') {
    throw new Error('[CRITICAL] Wheelset lacing topology must be an object.')
  }
  assertSelection(topology.selection, topology.cross)

  const expectedTotal = topology.selection === '24_2to1'
    ? 24
    : topology.selection === 21
      ? 21
      : topology.selection

  const expectedDistribution: LacingDistribution = topology.selection === 21
    ? 'g3_2to1'
    : topology.selection === '24_2to1'
      ? 'uniform_2to1'
      : 'symmetric_1to1'
  if (topology.distribution !== expectedDistribution) {
    throw new Error(`[CRITICAL] Topology distribution does not match selection ${String(topology.selection)}.`)
  }

  const expectedA = topology.distribution === 'uniform_2to1'
    ? 16
    : topology.distribution === 'g3_2to1'
      ? 14
      : expectedTotal / 2
  const expectedB = topology.distribution === 'uniform_2to1'
    ? 8
    : topology.distribution === 'g3_2to1'
      ? 7
      : expectedTotal / 2

  if (topology.rimHoles.length !== expectedTotal) {
    throw new Error(`[CRITICAL] Rim hole count mismatch: expected ${expectedTotal}, received ${topology.rimHoles.length}`)
  }
  if (topology.hubHolesA.length !== expectedA || topology.hubHolesB.length !== expectedB) {
    throw new Error(`[CRITICAL] Flange hole count mismatch: expected A/B ${expectedA}/${expectedB}, received ${topology.hubHolesA.length}/${topology.hubHolesB.length}`)
  }
  if (topology.spokes.length !== expectedTotal) {
    throw new Error(`[CRITICAL] Spoke count mismatch: expected ${expectedTotal}, received ${topology.spokes.length}`)
  }

  assertLacingPointCollection(topology.rimHoles, 'Rim holes')
  assertLacingPointCollection(topology.hubHolesA, 'Flange A holes', 'A')
  assertLacingPointCollection(topology.hubHolesB, 'Flange B holes', 'B')

  const rimSideCounts = topology.rimHoles.reduce((counts, point) => {
    counts[point.side] += 1
    return counts
  }, { A: 0, B: 0 })
  if (rimSideCounts.A !== expectedA || rimSideCounts.B !== expectedB) {
    throw new Error(`[CRITICAL] Rim side count mismatch: expected A/B ${expectedA}/${expectedB}, received ${rimSideCounts.A}/${rimSideCounts.B}`)
  }

  const rimById = new Map(topology.rimHoles.map(point => [point.id, point]))
  const hubByKey = new Map([
    ...topology.hubHolesA.map(point => [`A:${point.id}`, point] as const),
    ...topology.hubHolesB.map(point => [`B:${point.id}`, point] as const),
  ])

  const rimUsage = new Map<number, number>()
  const hubUsage = new Map<string, number>()
  const spokeIds = new Set<number>()
  for (const spoke of topology.spokes) {
    if (!Number.isInteger(spoke.id) || spoke.id < 0 || spokeIds.has(spoke.id)) {
      throw new Error(`[CRITICAL] Spokes contain an invalid or duplicate ID: ${spoke.id}`)
    }
    spokeIds.add(spoke.id)
    if (spoke.side !== 'A' && spoke.side !== 'B') {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} has an unsupported side: ${spoke.side}`)
    }
    if (spoke.type !== 'leading' && spoke.type !== 'trailing' && spoke.type !== 'nondrive') {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} has an unsupported type: ${spoke.type}`)
    }
    const canonicalRim = rimById.get(spoke.rim.id)
    const canonicalHub = hubByKey.get(`${spoke.side}:${spoke.hub.id}`)
    if (!canonicalRim || canonicalRim.side !== spoke.side ||
      canonicalRim.x !== spoke.rim.x || canonicalRim.y !== spoke.rim.y || canonicalRim.angle !== spoke.rim.angle) {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} references a rim hole outside the topology contract.`)
    }
    if (!canonicalHub || canonicalHub.side !== spoke.side ||
      canonicalHub.x !== spoke.hub.x || canonicalHub.y !== spoke.hub.y || canonicalHub.angle !== spoke.hub.angle) {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} references a flange hole outside the topology contract.`)
    }
    rimUsage.set(spoke.rim.id, (rimUsage.get(spoke.rim.id) || 0) + 1)
    const hubKey = `${spoke.side}:${spoke.hub.id}`
    hubUsage.set(hubKey, (hubUsage.get(hubKey) || 0) + 1)
  }

  const missingRim = topology.rimHoles.filter(point => rimUsage.get(point.id) !== 1).map(point => point.id)
  if (missingRim.length > 0) {
    throw new Error(`[CRITICAL] Rim holes must map one-to-one to spokes. Invalid IDs: ${missingRim.join(', ')}`)
  }

  const missingHubs = [...topology.hubHolesA.map(point => `A:${point.id}`), ...topology.hubHolesB.map(point => `B:${point.id}`)]
    .filter(key => hubUsage.get(key) !== 1)
  if (missingHubs.length > 0) {
    throw new Error(`[CRITICAL] Flange holes must map one-to-one to spokes. Invalid IDs: ${missingHubs.join(', ')}`)
  }

  for (const spoke of topology.spokes) {
    if (spoke.hub.side !== spoke.side || spoke.rim.side !== spoke.side) {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} crosses a side assignment boundary.`)
    }
    if (spoke.side === 'B' && spoke.type !== 'nondrive') {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} on side B must be marked nondrive.`)
    }
    if (spoke.side === 'A' && spoke.type === 'nondrive') {
      throw new Error(`[CRITICAL] Spoke ${spoke.id} on side A cannot be marked nondrive.`)
    }
  }

  const sideCounts = topology.spokes.reduce((counts, spoke) => {
    counts[spoke.side] += 1
    return counts
  }, { A: 0, B: 0 })
  if (sideCounts.A !== expectedA || sideCounts.B !== expectedB) {
    throw new Error(`[CRITICAL] Spoke side count mismatch: expected A/B ${expectedA}/${expectedB}, received ${sideCounts.A}/${sideCounts.B}`)
  }
}

export const buildWheelsetLacingTopology = (
  holes: LacingHoleSelection,
  cross: number,
  dimensions: LacingTopologyDimensions,
): WheelsetLacingTopology => {
  assertSelection(holes, cross)
  assertFinitePositive(dimensions.rimRadius, 'Rim radius')
  assertFinitePositive(dimensions.flangeRadiusA, 'Flange A radius')
  assertFinitePositive(dimensions.flangeRadiusB, 'Flange B radius')

  let distribution: LacingDistribution
  let geometry: Pick<WheelsetLacingTopology, 'rimHoles' | 'hubHolesA' | 'hubHolesB' | 'spokes'>

  if (holes === 21) {
    distribution = 'g3_2to1'
    geometry = buildG3Topology(dimensions)
  } else if (holes === '24_2to1') {
    distribution = 'uniform_2to1'
    geometry = buildUniformTwoToOneTopology(dimensions)
  } else {
    distribution = 'symmetric_1to1'
    geometry = buildSymmetricTopology(holes, cross, dimensions)
  }

  const topology: WheelsetLacingTopology = {
    selection: holes,
    cross,
    distribution,
    ...geometry,
  }
  topology.spokes.forEach((spoke, index) => {
    spoke.id = index
  })
  assertWheelsetLacingTopology(topology)
  return topology
}
