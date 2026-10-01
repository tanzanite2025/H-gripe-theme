import type { WheelsetLacingTopology } from './wheelsetLacingTopology.js'

const DEGREES_PER_RADIAN = 180 / Math.PI
const NORMALIZED_COMPONENT_NUMERIC_TOLERANCE = 1e-12

/**
 * Geometry-only direction metrics derived from the generated display line
 * segments. These values are not wheel efficiency, stiffness, tension, or
 * strength measurements.
 */
export interface WheelsetLacingGeometryProjectionMetrics {
  /**
   * Angle formed by the mean absolute tangential and radial components of the
   * drive-side display vectors, in degrees. This is not a signed vector angle.
   */
  aggregateMeanAbsoluteProjectionAngleDegrees: number
  /** Smallest per-spoke absolute drive-side projection angle, in degrees. */
  minimumAbsoluteDriveSideProjectionAngleDegrees: number
  /** Largest per-spoke absolute drive-side projection angle, in degrees. */
  maximumAbsoluteDriveSideProjectionAngleDegrees: number
  /** Mean absolute tangential component of drive-side unit spoke vectors. */
  meanAbsoluteTangentialProjectionPercent: number
  /** Mean absolute radial component of drive-side unit spoke vectors. */
  meanAbsoluteRadialProjectionPercent: number
  /** Number of drive-side spoke vectors included in the aggregate. */
  driveSideSpokeCount: number
}

const assertFiniteNormalizedComponent = (value: number, label: string): number => {
  if (!Number.isFinite(value)) {
    throw new Error(`[CRITICAL] ${label} must be finite. Received: ${value}`)
  }
  if (value < -1 - NORMALIZED_COMPONENT_NUMERIC_TOLERANCE || value > 1 + NORMALIZED_COMPONENT_NUMERIC_TOLERANCE) {
    throw new Error(`[CRITICAL] ${label} must be within [-1, 1]. Received: ${value}`)
  }
  return Math.min(1, Math.max(-1, value))
}

const calculateAbsoluteTangentialUnitVectorComponent = (
  hubX: number,
  hubY: number,
  spokeVectorX: number,
  spokeVectorY: number,
  hubRadius: number,
  spokeLength: number,
): number => {
  const component = ((-hubY * spokeVectorX) + (hubX * spokeVectorY)) / (hubRadius * spokeLength)
  return Math.abs(assertFiniteNormalizedComponent(component, 'Tangential spoke projection'))
}

const calculateAbsoluteRadialUnitVectorComponent = (
  hubX: number,
  hubY: number,
  spokeVectorX: number,
  spokeVectorY: number,
  hubRadius: number,
  spokeLength: number,
): number => {
  const component = ((hubX * spokeVectorX) + (hubY * spokeVectorY)) / (hubRadius * spokeLength)
  return Math.abs(assertFiniteNormalizedComponent(component, 'Radial spoke projection'))
}

const mean = (values: number[]): number => {
  if (values.length === 0) {
    throw new Error('[CRITICAL] Cannot calculate a mean from an empty geometry sample set.')
  }
  return values.reduce((sum, value) => sum + value, 0) / values.length
}

/**
 * Calculates the existing page's geometry projection cards from the exact
 * generated spoke vectors. The current display contract intentionally uses
 * absolute component magnitudes: the leading and trailing spokes have
 * opposite signed tangential directions, so averaging signed vectors would
 * cancel the topology's visual direction. The aggregate angle therefore uses
 * atan2(mean(abs(tangential)), mean(abs(radial))). It is a display-geometry
 * summary, not a signed force direction or a mechanical performance value.
 */
export const calculateWheelsetLacingGeometryProjectionMetrics = (
  topology: WheelsetLacingTopology,
): WheelsetLacingGeometryProjectionMetrics => {
  const driveSideSpokes = topology.spokes.filter(spoke => spoke.side === 'A')
  if (driveSideSpokes.length === 0) {
    throw new Error('[CRITICAL] Cannot calculate geometry projections without drive-side spokes.')
  }

  const perSpokeSamples = driveSideSpokes.map(spoke => {
    const hubRadius = Math.hypot(spoke.hub.x, spoke.hub.y)
    const spokeVectorX = spoke.rim.x - spoke.hub.x
    const spokeVectorY = spoke.rim.y - spoke.hub.y
    const spokeLength = Math.hypot(spokeVectorX, spokeVectorY)
    if (!Number.isFinite(hubRadius) || hubRadius <= 0 || !Number.isFinite(spokeLength) || spokeLength <= 0) {
      throw new Error(`[CRITICAL] Invalid display geometry for spoke ${spoke.id}.`)
    }

    const tangentialProjection = calculateAbsoluteTangentialUnitVectorComponent(
      spoke.hub.x,
      spoke.hub.y,
      spokeVectorX,
      spokeVectorY,
      hubRadius,
      spokeLength,
    )
    const radialProjection = calculateAbsoluteRadialUnitVectorComponent(
      spoke.hub.x,
      spoke.hub.y,
      spokeVectorX,
      spokeVectorY,
      hubRadius,
      spokeLength,
    )

    return {
      tangentialProjection,
      radialProjection,
      projectionAngleDegrees: Math.atan2(tangentialProjection, radialProjection) * DEGREES_PER_RADIAN,
    }
  })

  const meanTangentialProjection = mean(perSpokeSamples.map(sample => sample.tangentialProjection))
  const meanRadialProjection = mean(perSpokeSamples.map(sample => sample.radialProjection))

  return {
    aggregateMeanAbsoluteProjectionAngleDegrees: Math.atan2(
      meanTangentialProjection,
      meanRadialProjection,
    ) * DEGREES_PER_RADIAN,
    minimumAbsoluteDriveSideProjectionAngleDegrees: Math.min(
      ...perSpokeSamples.map(sample => sample.projectionAngleDegrees),
    ),
    maximumAbsoluteDriveSideProjectionAngleDegrees: Math.max(
      ...perSpokeSamples.map(sample => sample.projectionAngleDegrees),
    ),
    meanAbsoluteTangentialProjectionPercent: meanTangentialProjection * 100,
    meanAbsoluteRadialProjectionPercent: meanRadialProjection * 100,
    driveSideSpokeCount: driveSideSpokes.length,
  }
}
