import {
  getSupportedWheelsetLacingCrossCounts,
  type WheelsetLacingTopologySelection,
} from './wheelsetLacingSelectionContract'

export const WHEELSET_LACING_DISPLAY_GEOMETRY_CONTRACT_VERSION = 'v1.12-backend-display-geometry'

export const WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT = Object.freeze({
  symmetric1To1: 'symmetric_1to1',
  uniform2To1: 'uniform_2to1',
  uniform18H2To1: 'uniform_18h_2to1',
  g3TwentyOneHoleTriplet2To1: 'g3_21h_triplet_2to1',
} as const)

export type WheelsetLacingDisplayGeometryLayout =
  typeof WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT[keyof typeof WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT]

export const WHEELSET_LACING_SPOKE_HEAD_STYLE = Object.freeze({
  jBend: 'j_bend',
  straightPull: 'straight_pull',
} as const)

export type WheelsetLacingSpokeHeadStyle =
  typeof WHEELSET_LACING_SPOKE_HEAD_STYLE[keyof typeof WHEELSET_LACING_SPOKE_HEAD_STYLE]

export interface WheelsetLacingDisplayGeometryTopologySelection {
  topologyId: string
  selection: string
  displayLayout: WheelsetLacingDisplayGeometryLayout
  spokeHeadStyle: WheelsetLacingSpokeHeadStyle
}

type UnknownRecord = Record<string, unknown>

const isUnknownRecord = (value: unknown): value is UnknownRecord => (
  typeof value === 'object' && value !== null && !Array.isArray(value)
)

const requireUnknownRecord = (value: unknown, label: string): UnknownRecord => {
  if (!isUnknownRecord(value)) {
    throw new Error(`${label} must be an object`)
  }
  return value
}

const requireUnknownArray = (value: unknown, label: string): unknown[] => {
  if (!Array.isArray(value)) {
    throw new Error(`${label} must be an array`)
  }
  return value
}

const requireFiniteNumber = (value: unknown, label: string): number => {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    throw new Error(`${label} must be a finite number`)
  }
  return value
}

const requireInteger = (value: unknown, label: string): number => {
  const number = requireFiniteNumber(value, label)
  if (!Number.isInteger(number)) {
    throw new Error(`${label} must be an integer`)
  }
  return number
}

const requireSide = (value: unknown, label: string): 'A' | 'B' => {
  if (value !== 'A' && value !== 'B') {
    throw new Error(`${label} must be A or B`)
  }
  return value
}

const requireSpokeType = (value: unknown, label: string): 'leading' | 'trailing' | 'nondrive' => {
  if (value !== 'leading' && value !== 'trailing' && value !== 'nondrive') {
    throw new Error(`${label} has an unsupported spoke type`)
  }
  return value
}

const requireBoolean = (value: unknown, label: string): boolean => {
  if (typeof value !== 'boolean') {
    throw new Error(`${label} must be a boolean`)
  }
  return value
}

const validatePointListAgainstHoleList = (
  pointsValue: unknown,
  holesValue: unknown,
  label: string,
): Map<string, UnknownRecord> => {
  const points = requireUnknownArray(pointsValue, `${label} display points`)
  const holes = requireUnknownArray(holesValue, `${label} topology holes`)
  if (points.length !== holes.length) {
    throw new Error(`${label} display point count does not match topology hole count`)
  }

  const expectedHoleKeys = new Set<string>()
  for (const holeValue of holes) {
    const hole = requireUnknownRecord(holeValue, `${label} topology hole`)
    const id = requireInteger(hole.id, `${label} topology hole id`)
    const side = requireSide(hole.side, `${label} topology hole side`)
    const key = `${side}:${id}`
    if (expectedHoleKeys.has(key)) {
      throw new Error(`${label} topology holes contain a duplicate ${key}`)
    }
    expectedHoleKeys.add(key)
  }

  const pointsByKey = new Map<string, UnknownRecord>()
  for (const pointValue of points) {
    const point = requireUnknownRecord(pointValue, `${label} display point`)
    const id = requireInteger(point.id, `${label} display point id`)
    const side = requireSide(point.side, `${label} display point side`)
    requireFiniteNumber(point.angle, `${label} display point angle`)
    requireFiniteNumber(point.x, `${label} display point x`)
    requireFiniteNumber(point.y, `${label} display point y`)
    if (Math.hypot(point.x as number, point.y as number) <= 0) {
      throw new Error(`${label} display point ${side}:${id} has no radial position`)
    }
    const key = `${side}:${id}`
    if (!expectedHoleKeys.has(key)) {
      throw new Error(`${label} display point ${key} is absent from topology holes`)
    }
    if (pointsByKey.has(key)) {
      throw new Error(`${label} display points contain a duplicate ${key}`)
    }
    pointsByKey.set(key, point)
  }
  if (pointsByKey.size !== expectedHoleKeys.size) {
    throw new Error(`${label} display points do not cover every topology hole`)
  }
  return pointsByKey
}

const coordinatesMatch = (left: UnknownRecord, right: UnknownRecord): boolean => (
  left.id === right.id
  && left.side === right.side
  && left.x === right.x
  && left.y === right.y
)

const validateDisplayGeometrySpokeMappings = (
  geometrySpokesValue: unknown,
  topologySpokesValue: unknown,
  rimPointsByKey: Map<string, UnknownRecord>,
  hubPointsAByKey: Map<string, UnknownRecord>,
  hubPointsBByKey: Map<string, UnknownRecord>,
): void => {
  const geometrySpokes = requireUnknownArray(geometrySpokesValue, 'display geometry spokes')
  const topologySpokes = requireUnknownArray(topologySpokesValue, 'topology spokes')
  if (geometrySpokes.length !== topologySpokes.length) {
    throw new Error('display geometry spoke count does not match topology spoke count')
  }

  const topologySpokeById = new Map<number, UnknownRecord>()
  for (const topologySpokeValue of topologySpokes) {
    const topologySpoke = requireUnknownRecord(topologySpokeValue, 'topology spoke')
    const id = requireInteger(topologySpoke.id, 'topology spoke id')
    if (topologySpokeById.has(id)) {
      throw new Error(`topology spokes contain duplicate id ${id}`)
    }
    topologySpokeById.set(id, topologySpoke)
  }

  const geometrySpokeIds = new Set<number>()
  for (const geometrySpokeValue of geometrySpokes) {
    const geometrySpoke = requireUnknownRecord(geometrySpokeValue, 'display geometry spoke')
    const id = requireInteger(geometrySpoke.id, 'display geometry spoke id')
    if (geometrySpokeIds.has(id)) {
      throw new Error(`display geometry spokes contain duplicate id ${id}`)
    }
    geometrySpokeIds.add(id)

    const topologySpoke = topologySpokeById.get(id)
    if (!topologySpoke) {
      throw new Error(`display geometry spoke ${id} is absent from topology spokes`)
    }
    const side = requireSide(geometrySpoke.side, `display geometry spoke ${id} side`)
    const type = requireSpokeType(geometrySpoke.type, `display geometry spoke ${id}`)
    if (side !== topologySpoke.side || type !== topologySpoke.type) {
      throw new Error(`display geometry spoke ${id} metadata does not match topology`)
    }

    const hubPoint = requireUnknownRecord(geometrySpoke.hub, `display geometry spoke ${id} hub`)
    const rimPoint = requireUnknownRecord(geometrySpoke.rim, `display geometry spoke ${id} rim`)
    const hubId = requireInteger(hubPoint.id, `display geometry spoke ${id} hub id`)
    const rimId = requireInteger(rimPoint.id, `display geometry spoke ${id} rim id`)
    const hubSide = requireSide(hubPoint.side, `display geometry spoke ${id} hub side`)
    const rimSide = requireSide(rimPoint.side, `display geometry spoke ${id} rim side`)
    const expectedHubPoints = side === 'A' ? hubPointsAByKey : hubPointsBByKey
    const expectedHubPoint = expectedHubPoints.get(`${hubSide}:${hubId}`)
    const expectedRimPoint = rimPointsByKey.get(`${rimSide}:${rimId}`)
    if (!expectedHubPoint || !expectedRimPoint || !coordinatesMatch(hubPoint, expectedHubPoint) || !coordinatesMatch(rimPoint, expectedRimPoint)) {
      throw new Error(`display geometry spoke ${id} endpoints do not match topology coordinates`)
    }

    if (topologySpoke.hub_hole_id !== hubId || topologySpoke.rim_hole_id !== rimId || topologySpoke.side !== hubSide || topologySpoke.side !== rimSide) {
      throw new Error(`display geometry spoke ${id} endpoints do not match topology mapping`)
    }
  }

  if (geometrySpokeIds.size !== topologySpokeById.size) {
    throw new Error('display geometry spokes do not cover every topology spoke')
  }
}

const validateStraightPullProjection = (
  projectionValue: unknown,
  topologySpokes: unknown[],
  topologyHubHolesAValue: unknown,
  topologyHubHolesBValue: unknown,
  rimPointsByKey: Map<string, UnknownRecord>,
  baseHubPointsAByKey: Map<string, UnknownRecord>,
  baseHubPointsBByKey: Map<string, UnknownRecord>,
): void => {
  const projection = requireUnknownRecord(projectionValue, 'straight-pull display projection')
  const flangeCenterA = requireUnknownRecord(projection.flange_center_a, 'straight-pull flange A center')
  const flangeCenterB = requireUnknownRecord(projection.flange_center_b, 'straight-pull flange B center')
  const topologyHubHolesA = requireUnknownArray(topologyHubHolesAValue, 'topology hub A holes')
  const topologyHubHolesB = requireUnknownArray(topologyHubHolesBValue, 'topology hub B holes')
  const projectedHubPointsAByKey = validatePointListAgainstHoleList(
    projection.hub_holes_a,
    topologyHubHolesA,
    'straight-pull projected hub A',
  )
  const projectedHubPointsBByKey = validatePointListAgainstHoleList(
    projection.hub_holes_b,
    topologyHubHolesB,
    'straight-pull projected hub B',
  )

  for (const [projectedPointsByKey, basePointsByKey, center, sideLabel] of [
    [projectedHubPointsAByKey, baseHubPointsAByKey, flangeCenterA, 'A'],
    [projectedHubPointsBByKey, baseHubPointsBByKey, flangeCenterB, 'B'],
  ] as const) {
    const centerX = requireFiniteNumber(center.x, `straight-pull flange ${sideLabel} center x`)
    const centerY = requireFiniteNumber(center.y, `straight-pull flange ${sideLabel} center y`)
    for (const [key, point] of projectedPointsByKey) {
      const basePoint = basePointsByKey.get(key)
      if (!basePoint
        || point.angle !== basePoint.angle
        || Math.abs(Number(point.x) - (Number(basePoint.x) + centerX)) > 0.02
        || Math.abs(Number(point.y) - (Number(basePoint.y) + centerY)) > 0.02) {
        throw new Error(`straight-pull projected hub ${sideLabel} point ${key} does not match its flange plane`)
      }
    }
  }

  validateDisplayGeometrySpokeMappings(
    projection.spokes,
    topologySpokes,
    rimPointsByKey,
    projectedHubPointsAByKey,
    projectedHubPointsBByKey,
  )
}

const validateDisplayGeometryMetrics = (metricsValue: unknown): void => {
  const metrics = requireUnknownRecord(metricsValue, 'display geometry metrics')
  for (const field of [
    'aggregate_mean_absolute_projection_angle_degrees',
    'minimum_absolute_drive_side_projection_angle_degrees',
    'maximum_absolute_drive_side_projection_angle_degrees',
    'mean_absolute_tangential_projection_percent',
    'mean_absolute_radial_projection_percent',
  ]) {
    requireFiniteNumber(metrics[field], `display geometry metrics.${field}`)
  }
  if (requireInteger(metrics.drive_side_spoke_count, 'display geometry metrics.drive_side_spoke_count') <= 0) {
    throw new Error('display geometry metrics.drive_side_spoke_count must be positive')
  }
}

const validateDisplayGeometryFlangeProfile = (profileValue: unknown): void => {
  const profile = requireUnknownRecord(profileValue, 'display geometry flange profile')
  for (const field of [
    'centerline_x',
    'flange_a_x',
    'flange_b_x',
    'axle_left_x',
    'axle_right_x',
    'flange_offset_a_mm',
    'flange_offset_b_mm',
    'total_flange_span_mm',
  ]) {
    requireFiniteNumber(profile[field], `display geometry flange profile.${field}`)
  }
}

const validateDisplayGeometryLayoutSpecificFields = (
  displayLayout: WheelsetLacingDisplayGeometryLayout,
  spacingValue: unknown,
): void => {
  const spacing = requireUnknownRecord(spacingValue, 'display geometry G3 group spacing')
  const enabled = requireBoolean(spacing.enabled, 'display geometry G3 group spacing.enabled')
  if (displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3TwentyOneHoleTriplet2To1) {
    if (!enabled || requireInteger(spacing.group_count, 'display geometry G3 group spacing.group_count') !== 7) {
      throw new Error('G3 display geometry must expose seven enabled groups')
    }
    const groupPitch = requireFiniteNumber(spacing.group_pitch_degrees, 'display geometry G3 group spacing.group_pitch_degrees')
    const spacingAToB = requireFiniteNumber(spacing.spacing_a_to_b_degrees, 'display geometry G3 group spacing.spacing_a_to_b_degrees')
    const spacingBToA = requireFiniteNumber(spacing.spacing_b_to_a_degrees, 'display geometry G3 group spacing.spacing_b_to_a_degrees')
    const spacingAToNextGroupA = requireFiniteNumber(spacing.spacing_a_to_next_group_a_degrees, 'display geometry G3 group spacing.spacing_a_to_next_group_a_degrees')
    const parallelHoleSpacing = requireFiniteNumber(spacing.parallel_hole_spacing_mm, 'display geometry G3 group spacing.parallel_hole_spacing_mm')
    const sideAFlangeHoleCircleRadius = requireFiniteNumber(spacing.side_a_flange_hole_circle_radius_mm, 'display geometry G3 group spacing.side_a_flange_hole_circle_radius_mm')
    const sideAFlangePCD = requireFiniteNumber(spacing.side_a_flange_pcd_mm, 'display geometry G3 group spacing.side_a_flange_pcd_mm')
    const flangePCDRoundingToleranceMM = 0.011
    const spacingTotal = spacingAToB + spacingBToA + spacingAToNextGroupA
    if (groupPitch <= 0 || spacingAToB <= 0 || spacingBToA <= 0 || spacingAToNextGroupA <= 0 || Math.abs(spacingTotal - groupPitch) > 0.01) {
      throw new Error('G3 display geometry group spacing must close all three rim-hole gaps to one group pitch')
    }
    if (parallelHoleSpacing <= 0 || sideAFlangeHoleCircleRadius <= 0 || Math.abs(sideAFlangePCD - 2 * sideAFlangeHoleCircleRadius) > flangePCDRoundingToleranceMM) {
      throw new Error('G3 parallel-hole spacing and A-side flange dimensions must be positive and consistent')
    }
    return
  }
  if (enabled) {
    throw new Error('non-G3 display geometry must not expose G3 group spacing')
  }
}

const validateUniformTwoToOneTopologyShape = (
  topology: UnknownRecord,
  rimHoles: unknown[],
  hubHolesA: unknown[],
  hubHolesB: unknown[],
  expectedRimHoleCount: number,
): void => {
  const holeCountLabel = `${expectedRimHoleCount}H`
  if (requireInteger(topology.cross, 'topology cross') !== 2) {
    throw new Error(`uniform ${holeCountLabel} 2:1 display geometry must use 2X crossing`)
  }
  const groupCount = expectedRimHoleCount / 3
  const straightPull = topology.spoke_head_style === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
  const expectedHubHolePairSize = straightPull ? 2 : 1
  const expectedDriveSideHubHoleCount = (groupCount * 2) / expectedHubHolePairSize
  const expectedNonDriveSideHubHoleCount = groupCount / expectedHubHolePairSize
  if (rimHoles.length !== expectedRimHoleCount
    || hubHolesA.length !== expectedDriveSideHubHoleCount
    || hubHolesB.length !== expectedNonDriveSideHubHoleCount) {
    throw new Error(`uniform ${holeCountLabel} 2:1 display geometry has incorrect rim or flange hole counts for its spoke head style`)
  }
  rimHoles.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform ${holeCountLabel} rim hole ${index}`)
    const expectedSide = index % 3 === 1 ? 'B' : 'A'
    if (requireSide(hole.side, `uniform ${holeCountLabel} rim hole ${index} side`) !== expectedSide) {
      throw new Error(`uniform ${holeCountLabel} 2:1 rim holes must repeat the A-B-A sequence`)
    }
    requireInteger(hole.id, `uniform ${holeCountLabel} rim hole ${index} id`)
  })
  hubHolesA.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform ${holeCountLabel} hub A hole ${index}`)
    if (requireSide(hole.side, `uniform ${holeCountLabel} hub A hole ${index} side`) !== 'A') {
      throw new Error(`uniform ${holeCountLabel} drive-side hub holes must all be side A`)
    }
    requireInteger(hole.id, `uniform ${holeCountLabel} hub A hole ${index} id`)
  })
  hubHolesB.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform ${holeCountLabel} hub B hole ${index}`)
    if (requireSide(hole.side, `uniform ${holeCountLabel} hub B hole ${index} side`) !== 'B') {
      throw new Error(`uniform ${holeCountLabel} non-drive-side hub holes must all be side B`)
    }
    requireInteger(hole.id, `uniform ${holeCountLabel} hub B hole ${index} id`)
  })
}

const validateTwentyOneHoleG3TopologyShape = (
  topology: UnknownRecord,
  rimHoles: unknown[],
  hubHolesA: unknown[],
  hubHolesB: unknown[],
  topologySpokes: unknown[],
): void => {
  const driveSideSpokeCount = 14
  if (requireInteger(topology.cross, 'G3 topology cross') !== 2
    || topology.spoke_head_style !== WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
    || rimHoles.length !== 21
    || hubHolesA.length !== driveSideSpokeCount
    || hubHolesB.length !== 7) {
    throw new Error('21-hole G3 topology must use its registered 2:1 metadata, straight-pull heads, and 21/14/7 hole counts')
  }

  rimHoles.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `G3 rim hole ${index}`)
    const expectedSide = index % 3 === 1 ? 'B' : 'A'
    if (requireSide(hole.side, `G3 rim hole ${index} side`) !== expectedSide) {
      throw new Error('21-hole G3 rim holes must repeat the A-B-A group sequence')
    }
    requireInteger(hole.id, `G3 rim hole ${index} id`)
  })

  const driveSpokesByHubHoleId = new Map<number, UnknownRecord>()
  const nonDriveSpokesByHubHoleId = new Map<number, UnknownRecord>()
  for (const spokeValue of topologySpokes) {
    const spoke = requireUnknownRecord(spokeValue, 'G3 topology spoke')
    const side = requireSide(spoke.side, 'G3 topology spoke side')
    const hubHoleId = requireInteger(spoke.hub_hole_id, 'G3 topology spoke hub hole id')
    if (side === 'A') {
      driveSpokesByHubHoleId.set(hubHoleId, spoke)
    } else {
      nonDriveSpokesByHubHoleId.set(hubHoleId, spoke)
    }
  }

  for (let hubHoleId = 0; hubHoleId < driveSideSpokeCount; hubHoleId++) {
    const spoke = driveSpokesByHubHoleId.get(hubHoleId)
    const group = Math.floor(hubHoleId / 2)
    const isTrailing = hubHoleId % 2 === 0
    const expectedType = isTrailing ? 'trailing' : 'leading'
    const expectedRimHoleId = isTrailing
      ? group * 3 + 2
      : ((group + 1) % 7) * 3
    if (!spoke
      || requireSpokeType(spoke.type, `G3 drive spoke ${hubHoleId}`) !== expectedType
      || requireInteger(spoke.rim_hole_id, `G3 drive spoke ${hubHoleId} rim hole id`) !== expectedRimHoleId) {
      throw new Error(`G3 drive spoke ${hubHoleId} must follow the grouped straight-pull parallel mapping`)
    }
  }

  for (let hubHoleId = 0; hubHoleId < 7; hubHoleId++) {
    const spoke = nonDriveSpokesByHubHoleId.get(hubHoleId)
    if (!spoke
      || requireSpokeType(spoke.type, `G3 non-drive spoke ${hubHoleId}`) !== 'nondrive'
      || requireInteger(spoke.rim_hole_id, `G3 non-drive spoke ${hubHoleId} rim hole id`) !== hubHoleId * 3 + 1) {
      throw new Error(`G3 non-drive spoke ${hubHoleId} must remain radial to its grouped rim hole`)
    }
  }
}

const validatePairedStraightPullTopologyShape = (
  hubHolesA: unknown[],
  hubHolesB: unknown[],
  topologySpokes: unknown[],
): void => {
  const spokeCountByHubHoleKey = new Map<string, number>()
  for (const spokeValue of topologySpokes) {
    const spoke = requireUnknownRecord(spokeValue, 'straight-pull topology spoke')
    const side = requireSide(spoke.side, 'straight-pull topology spoke side')
    const hubHoleId = requireInteger(spoke.hub_hole_id, 'straight-pull topology spoke hub hole id')
    const key = `${side}:${hubHoleId}`
    spokeCountByHubHoleKey.set(key, (spokeCountByHubHoleKey.get(key) ?? 0) + 1)
  }

  for (const [holes, expectedSide] of [[hubHolesA, 'A'], [hubHolesB, 'B']] as const) {
    for (const holeValue of holes) {
      const hole = requireUnknownRecord(holeValue, `straight-pull hub ${expectedSide} hole`)
      const id = requireInteger(hole.id, `straight-pull hub ${expectedSide} hole id`)
      const side = requireSide(hole.side, `straight-pull hub ${expectedSide} hole side`)
      if (side !== expectedSide || spokeCountByHubHoleKey.get(`${side}:${id}`) !== 2) {
        throw new Error(`straight-pull hub hole ${expectedSide}:${id} must be shared by exactly two spokes`)
      }
      spokeCountByHubHoleKey.delete(`${side}:${id}`)
    }
  }
  if (spokeCountByHubHoleKey.size !== 0) {
    throw new Error('straight-pull topology spokes must map only to registered shared flange holes')
  }
}

/**
 * Maps an explicit UI topology selection to its canonical ID and display
 * layout. The 21-hole G3 selector is independent from its hole count, and its
 * display layout cannot be reused by a future G3 hole-count variant.
 */
export const resolveWheelsetLacingDisplayGeometryTopologySelection = (
  topologySelection: WheelsetLacingTopologySelection,
  cross: number,
  requestedSpokeHeadStyle?: WheelsetLacingSpokeHeadStyle,
): WheelsetLacingDisplayGeometryTopologySelection => {
  const supportedCrossCounts = getSupportedWheelsetLacingCrossCounts(topologySelection)
  if (supportedCrossCounts.length === 0) {
    throw new Error(`unregistered wheelset lacing display topology ${String(topologySelection)}`)
  }
  if (!supportedCrossCounts.includes(cross)) {
    throw new Error(`unsupported ${String(topologySelection)} wheelset lacing cross count ${String(cross)}`)
  }
  const spokeHeadStyle = requestedSpokeHeadStyle
    ?? (topologySelection === '21_g3'
      ? WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
      : WHEELSET_LACING_SPOKE_HEAD_STYLE.jBend)
  if (!Object.values(WHEELSET_LACING_SPOKE_HEAD_STYLE).includes(spokeHeadStyle)) {
    throw new Error(`unsupported wheelset spoke head style ${String(spokeHeadStyle)}`)
  }
  if (topologySelection === '21_g3' && spokeHeadStyle !== WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull) {
    throw new Error('21-hole G3 is available only in straight-pull mode')
  }
  switch (topologySelection) {
    case '21_g3':
      return {
        topologyId: '21h-g3-2to1',
        selection: '21_g3',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3TwentyOneHoleTriplet2To1,
        spokeHeadStyle: WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull,
      }
    case '24_2to1':
      return {
        topologyId: spokeHeadStyle === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
          ? '24h-uniform-2to1-straight-pull'
          : '24h-uniform-2to1',
        selection: '24_2to1',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1,
        spokeHeadStyle,
      }
    case '18_2to1':
      return {
        topologyId: spokeHeadStyle === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
          ? '18h-uniform-2to1-straight-pull'
          : '18h-uniform-2to1',
        selection: '18_2to1',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1,
        spokeHeadStyle,
      }
    case 16:
    case 20:
    case 24:
    case 28:
    case 32:
    case 36:
      return {
        topologyId: spokeHeadStyle === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull
          ? `${topologySelection}h-symmetric-1to1-${cross}x-straight-pull`
          : `${topologySelection}h-symmetric-1to1-${cross}x`,
        selection: String(topologySelection),
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.symmetric1To1,
        spokeHeadStyle,
      }
    default:
    throw new Error(`unregistered wheelset lacing display topology ${String(topologySelection)}`)
  }
}

/**
 * Validates a backend projection before the page creates any SVG element from
 * it. This intentionally rejects unknown layouts and endpoint mismatches so
 * every registered 18H 2:1 projection is drawn only from its own layout.
 */
export const validateWheelsetLacingDisplayGeometryResponse = (
  value: unknown,
  expectedSelection: WheelsetLacingDisplayGeometryTopologySelection,
): void => {
  const geometry = requireUnknownRecord(value, 'wheelset lacing display geometry response')
  if (geometry.contract_version !== WHEELSET_LACING_DISPLAY_GEOMETRY_CONTRACT_VERSION) {
    throw new Error(`unsupported display geometry contract ${String(geometry.contract_version)}`)
  }
  if (geometry.display_layout !== expectedSelection.displayLayout) {
    throw new Error(`display geometry layout ${String(geometry.display_layout)} does not match the selected topology layout`)
  }

  const topology = requireUnknownRecord(geometry.topology, 'display geometry topology')
  if (topology.topology_id !== expectedSelection.topologyId) {
    throw new Error(`display geometry topology ${String(topology.topology_id)} does not match the selected topology`)
  }
  if (topology.selection !== expectedSelection.selection) {
    throw new Error(`display geometry selection ${String(topology.selection)} does not match the selected topology selection`)
  }
  if (topology.display_layout !== geometry.display_layout) {
    throw new Error('topology display layout does not match the projection display layout')
  }
  if (!Object.values(WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT).includes(topology.display_layout as WheelsetLacingDisplayGeometryLayout)) {
    throw new Error(`unsupported display geometry layout ${String(topology.display_layout)}`)
  }
  if (topology.spoke_head_style !== expectedSelection.spokeHeadStyle) {
    throw new Error('topology spoke head style does not match the selected spoke head tab')
  }
  const holeCount = requireInteger(topology.hole_count, 'topology hole_count')
  requireInteger(topology.cross, 'topology cross')
  const distributionByLayout: Record<WheelsetLacingDisplayGeometryLayout, string> = {
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.symmetric1To1]: 'symmetric_1to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1]: 'uniform_2to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1]: 'uniform_2to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3TwentyOneHoleTriplet2To1]: 'g3_2to1',
  }
  if (topology.distribution !== distributionByLayout[expectedSelection.displayLayout]) {
    throw new Error('topology distribution does not match the selected display layout')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1 && holeCount !== 24) {
    throw new Error('uniform 2:1 display geometry must contain 24 rim holes')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1 && holeCount !== 18) {
    throw new Error('uniform 18H 2:1 display geometry must contain 18 rim holes')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3TwentyOneHoleTriplet2To1 && holeCount !== 21) {
    throw new Error('21-hole G3 display geometry must contain 21 rim holes')
  }

  const topologyRimHoles = requireUnknownArray(topology.rim_holes, 'topology rim holes')
  const topologyHubHolesA = requireUnknownArray(topology.hub_holes_a, 'topology hub A holes')
  const topologyHubHolesB = requireUnknownArray(topology.hub_holes_b, 'topology hub B holes')
  const topologySpokes = requireUnknownArray(topology.spokes, 'topology spokes')
  if (topologyRimHoles.length !== holeCount || topologySpokes.length !== holeCount) {
    throw new Error('topology rim-hole and spoke counts must equal topology hole_count')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1) {
    validateUniformTwoToOneTopologyShape(topology, topologyRimHoles, topologyHubHolesA, topologyHubHolesB, 24)
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1) {
    validateUniformTwoToOneTopologyShape(topology, topologyRimHoles, topologyHubHolesA, topologyHubHolesB, 18)
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3TwentyOneHoleTriplet2To1) {
    validateTwentyOneHoleG3TopologyShape(topology, topologyRimHoles, topologyHubHolesA, topologyHubHolesB, topologySpokes)
  } else if (expectedSelection.spokeHeadStyle === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull) {
    validatePairedStraightPullTopologyShape(topologyHubHolesA, topologyHubHolesB, topologySpokes)
  }
  const rimPointsByKey = validatePointListAgainstHoleList(geometry.rim_holes, topologyRimHoles, 'rim')
  const hubPointsAByKey = validatePointListAgainstHoleList(geometry.hub_holes_a, topologyHubHolesA, 'hub A')
  const hubPointsBByKey = validatePointListAgainstHoleList(geometry.hub_holes_b, topologyHubHolesB, 'hub B')
  validateDisplayGeometrySpokeMappings(geometry.spokes, topologySpokes, rimPointsByKey, hubPointsAByKey, hubPointsBByKey)
  if (expectedSelection.spokeHeadStyle === WHEELSET_LACING_SPOKE_HEAD_STYLE.straightPull) {
    validateStraightPullProjection(
      geometry.straight_pull_projection,
      topologySpokes,
      topologyHubHolesA,
      topologyHubHolesB,
      rimPointsByKey,
      hubPointsAByKey,
      hubPointsBByKey,
    )
  } else if (geometry.straight_pull_projection !== undefined) {
    throw new Error('J-bend geometry must not expose straight-pull flange projections')
  }
  validateDisplayGeometryMetrics(geometry.metrics)
  validateDisplayGeometryFlangeProfile(geometry.flange_profile)
  validateDisplayGeometryLayoutSpecificFields(expectedSelection.displayLayout, geometry.g3_group_spacing)
}
