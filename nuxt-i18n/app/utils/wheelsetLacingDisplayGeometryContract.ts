import type { WheelsetLacingHoleSelection } from './wheelsetLacingSelectionContract'

export const WHEELSET_LACING_DISPLAY_GEOMETRY_CONTRACT_VERSION = 'v1.5-backend-display-geometry'

export const WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT = Object.freeze({
  symmetric1To1: 'symmetric_1to1',
  uniform2To1: 'uniform_2to1',
  uniform18H2To1: 'uniform_18h_2to1',
  g3Triplet2To1: 'g3_triplet_2to1',
} as const)

export type WheelsetLacingDisplayGeometryLayout =
  typeof WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT[keyof typeof WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT]

export interface WheelsetLacingDisplayGeometryTopologySelection {
  topologyId: string
  displayLayout: WheelsetLacingDisplayGeometryLayout
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
  if (displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1) {
    if (!enabled || requireInteger(spacing.group_count, 'display geometry G3 group spacing.group_count') !== 7) {
      throw new Error('G3 display geometry must expose seven enabled groups')
    }
    const groupPitch = requireFiniteNumber(spacing.group_pitch_degrees, 'display geometry G3 group spacing.group_pitch_degrees')
    const spacingAToB = requireFiniteNumber(spacing.spacing_a_to_b_degrees, 'display geometry G3 group spacing.spacing_a_to_b_degrees')
    const spacingBToA = requireFiniteNumber(spacing.spacing_b_to_a_degrees, 'display geometry G3 group spacing.spacing_b_to_a_degrees')
    if (groupPitch <= 0 || spacingAToB < 0 || spacingBToA < 0 || spacingAToB + spacingBToA >= groupPitch) {
      throw new Error('G3 display geometry group spacing leaves no positive inter-group gap')
    }
    return
  }
  if (enabled) {
    throw new Error('non-G3 display geometry must not expose G3 group spacing')
  }
}

const validateUniform18HTwoToOneTopologyShape = (
  topology: UnknownRecord,
  rimHoles: unknown[],
  hubHolesA: unknown[],
  hubHolesB: unknown[],
): void => {
  if (requireInteger(topology.cross, 'topology cross') !== 2) {
    throw new Error('uniform 18H 2:1 display geometry must use 2X crossing')
  }
  if (rimHoles.length !== 18 || hubHolesA.length !== 12 || hubHolesB.length !== 6) {
    throw new Error('uniform 18H 2:1 display geometry must contain 18 rim, 12 drive-side, and 6 non-drive-side holes')
  }
  rimHoles.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform 18H rim hole ${index}`)
    const expectedSide = index % 3 === 1 ? 'B' : 'A'
    if (requireSide(hole.side, `uniform 18H rim hole ${index} side`) !== expectedSide) {
      throw new Error('uniform 18H 2:1 rim holes must repeat the A-B-A sequence')
    }
  })
  hubHolesA.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform 18H hub A hole ${index}`)
    if (requireSide(hole.side, `uniform 18H hub A hole ${index} side`) !== 'A') {
      throw new Error('uniform 18H drive-side hub holes must all be side A')
    }
  })
  hubHolesB.forEach((holeValue, index) => {
    const hole = requireUnknownRecord(holeValue, `uniform 18H hub B hole ${index}`)
    if (requireSide(hole.side, `uniform 18H hub B hole ${index} side`) !== 'B') {
      throw new Error('uniform 18H non-drive-side hub holes must all be side B')
    }
  })
}

/**
 * Maps a UI topology selection to its canonical ID and exact display layout.
 * The layout is explicit metadata; it is never inferred from a hole count.
 */
export const resolveWheelsetLacingDisplayGeometryTopologySelection = (
  holeSelection: WheelsetLacingHoleSelection,
  cross: number,
): WheelsetLacingDisplayGeometryTopologySelection => {
  switch (holeSelection) {
    case 21:
      return {
        topologyId: '21h-g3-2to1',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1,
      }
    case '24_2to1':
      return {
        topologyId: '24h-uniform-2to1',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1,
      }
    case '18_2to1':
      return {
        topologyId: '18h-uniform-2to1',
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1,
      }
    case 16:
    case 20:
    case 24:
    case 28:
    case 32:
    case 36:
      return {
        topologyId: `${holeSelection}h-symmetric-1to1-${cross}x`,
        displayLayout: WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.symmetric1To1,
      }
    default:
      throw new Error(`unregistered wheelset lacing display topology ${String(holeSelection)}`)
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
  if (topology.display_layout !== geometry.display_layout) {
    throw new Error('topology display layout does not match the projection display layout')
  }
  if (!Object.values(WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT).includes(topology.display_layout as WheelsetLacingDisplayGeometryLayout)) {
    throw new Error(`unsupported display geometry layout ${String(topology.display_layout)}`)
  }
  const holeCount = requireInteger(topology.hole_count, 'topology hole_count')
  requireInteger(topology.cross, 'topology cross')
  const distributionByLayout: Record<WheelsetLacingDisplayGeometryLayout, string> = {
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.symmetric1To1]: 'symmetric_1to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform2To1]: 'uniform_2to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1]: 'uniform_2to1',
    [WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1]: 'g3_2to1',
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
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.g3Triplet2To1 && holeCount !== 21) {
    throw new Error('G3 display geometry must contain 21 rim holes')
  }

  const topologyRimHoles = requireUnknownArray(topology.rim_holes, 'topology rim holes')
  const topologyHubHolesA = requireUnknownArray(topology.hub_holes_a, 'topology hub A holes')
  const topologyHubHolesB = requireUnknownArray(topology.hub_holes_b, 'topology hub B holes')
  const topologySpokes = requireUnknownArray(topology.spokes, 'topology spokes')
  if (topologyRimHoles.length !== holeCount || topologySpokes.length !== holeCount) {
    throw new Error('topology rim-hole and spoke counts must equal topology hole_count')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1
    && (topologyHubHolesA.length !== 12 || topologyHubHolesB.length !== 6)) {
    throw new Error('uniform 18H 2:1 display geometry must contain 12 drive-side and 6 non-drive-side hub holes')
  }
  if (expectedSelection.displayLayout === WHEELSET_LACING_DISPLAY_GEOMETRY_LAYOUT.uniform18H2To1) {
    validateUniform18HTwoToOneTopologyShape(topology, topologyRimHoles, topologyHubHolesA, topologyHubHolesB)
  }
  const rimPointsByKey = validatePointListAgainstHoleList(geometry.rim_holes, topologyRimHoles, 'rim')
  const hubPointsAByKey = validatePointListAgainstHoleList(geometry.hub_holes_a, topologyHubHolesA, 'hub A')
  const hubPointsBByKey = validatePointListAgainstHoleList(geometry.hub_holes_b, topologyHubHolesB, 'hub B')
  validateDisplayGeometrySpokeMappings(geometry.spokes, topologySpokes, rimPointsByKey, hubPointsAByKey, hubPointsBByKey)
  validateDisplayGeometryMetrics(geometry.metrics)
  validateDisplayGeometryFlangeProfile(geometry.flange_profile)
  validateDisplayGeometryLayoutSpecificFields(expectedSelection.displayLayout, geometry.g3_group_spacing)
}
