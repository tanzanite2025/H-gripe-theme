import type { SpokeCalcInput } from '~~/types/spoke'
import type { SpokeWheelBuildConfig, SpokeWheelSide } from '~/types/spokeCalculator'

/**
 * Returns whether this wheel has enough manual information for a calculation.
 * Catalog selection belongs to the separate preset/search system and cannot
 * satisfy or override the calculator's measured geometry.
 */
export const hasSpokeCalculationGeometry = (config: SpokeWheelBuildConfig) => {
  return Boolean(
    config.erd
    && config.leftFlangePcd
    && config.rightFlangePcd
    && config.leftFlange != null
    && config.rightFlange != null,
  )
}

/**
 * Converts the UI's wheel configuration into the backend request contract.
 * Backend-specific field names stay in this adapter instead of spreading
 * through the template and calculation orchestration.
 */
export const toSpokeCalcInput = (
  config: SpokeWheelBuildConfig,
  wheel: SpokeWheelSide,
): SpokeCalcInput => ({
  // Manual calculator input is intentionally independent from the catalog
  // selection used by the preset/search system.
  rimId: '',
  hubId: '',
  wheelPosition: wheel,
  spokeCount: config.spokeCount,
  crossing: config.crossing,
  nippleType: config.nippleType,
  nippleLengthMm: config.nippleLength,
  spokeHeadType: config.spokeHeadType,
  spokeHoleDiameterMm: config.spokeHoleDiameterMm,
  straightPullTangentOffsetMm: config.straightPullTangentOffsetMm,
  spokeProfile: config.spokeProfile,
  targetTensionN: config.targetTensionN,
  alternatingDrillingOffsetMm: config.alternatingDrillingOffsetMm,
  interlacing: config.interlacing === 'on',
  interlaceCompensationMm: config.interlaceCompensationMm,
  rimOffsetMm: config.rimOffsetMm,
  erdMm: config.erd,
  leftFlangeMm: config.leftFlange,
  rightFlangeMm: config.rightFlange,
  leftFlangePcdMm: config.leftFlangePcd,
  rightFlangePcdMm: config.rightFlangePcd,
})
