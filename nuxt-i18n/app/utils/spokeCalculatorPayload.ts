import type { SpokeCalcInput } from '~~/types/spoke'
import type { SpokeWheelBuildConfig, SpokeWheelSide } from '~/types/spokeCalculator'

/**
 * Returns whether this wheel has enough information for a calculation.
 * Catalog selection and manual geometry follow the same rules as the
 * existing calculator panel during this extraction step.
 */
export const hasSpokeCalculationGeometry = (config: SpokeWheelBuildConfig) => {
  const hasCatalogSelection = Boolean(config.rimModelId && config.hubModelId)
  const hasManualGeometry = Boolean(
    config.erd
    && config.leftFlangePcd
    && config.rightFlangePcd
    && config.leftFlange != null
    && config.rightFlange != null,
  )

  return hasCatalogSelection || hasManualGeometry
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
  rimId: config.rimModelId || '',
  hubId: config.hubModelId || '',
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
