import type { SpokeTensionRatio } from '~~/types/spoke'

export type SpokeHeadType = 'j_bend' | 'straight_pull'
export type SpokeWheelSide = 'front' | 'rear'
export type SpokeNippleType = 'standard' | 'hidden'
export type SpokeProfile = 'round_2_0' | 'round_1_8' | 'bladed_0_9x2_2'
export type SpokeInterlacing = 'off' | 'on'

export type SpokeCalculatorTopologyDistribution = 'symmetric_1to1' | 'uniform_2to1' | 'g3_2to1'

/** Discrete topology facts returned by the shared wheelset-lacing endpoint. */
export interface SpokeCalculatorTopologySummary {
  topologyId: string
  selection: string
  holeCount: number
  cross: number
  distribution: SpokeCalculatorTopologyDistribution
  displayLayout: string
  sideACount: number
  sideBCount: number
}

export const SPOKE_WIZARD_STEPS = [
  { id: 'head_type', number: 1 },
  { id: 'erd', number: 2 },
  { id: 'alternating_drilling', number: 3 },
  { id: 'hub_geometry', number: 4 },
  { id: 'physical_corrections', number: 5 },
  { id: 'nipple_type', number: 6 },
] as const

export type SpokeWizardStep = typeof SPOKE_WIZARD_STEPS[number]['number']
export const SPOKE_WIZARD_STEP_COUNT = SPOKE_WIZARD_STEPS.length

/** Complete input configuration for one wheel. */
export interface SpokeWheelBuildConfig {
  topologyId: string
  spokeCount: number
  crossing: number
  g3RimHoleSpacingAToBDegrees: number
  g3RimHoleSpacingBToADegrees: number
  g3RimHoleSpacingAToNextGroupADegrees: number
  nippleType: SpokeNippleType
  nippleLength: number | null
  spokeHeadType: SpokeHeadType
  spokeHoleDiameterMm: number | null
  straightPullTangentOffsetMm: number
  spokeProfile: SpokeProfile
  targetTensionN: number
  alternatingDrillingOffsetMm: number
  interlacing: SpokeInterlacing
  interlaceCompensationMm: number | null
  spokeElongationCompensationMm: number | null

  erd: number | null
  rimOffsetMm: number
  leftFlange: number | null
  rightFlange: number | null
  leftFlangePcd: number | null
  rightFlangePcd: number | null
}

/** Catalog/preset selection state kept outside the manual calculator input. */
export interface SpokeWheelCatalogSelection {
  rimBrandId: string | null
  rimModelId: string | null
  hubBrandId: string | null
  hubModelId: string | null
}

export interface SpokeCalculatorSelectOption {
  label: string
  value: string | number | null
}

/** Options owned by the manual calculator. */
export interface SpokeCalculatorManualOptions {
  topologyOptions: SpokeCalculatorSelectOption[]
  topologySummaries: SpokeCalculatorTopologySummary[]
  nippleTypeOptions: SpokeCalculatorSelectOption[]
  spokeHeadTypeOptions: SpokeCalculatorSelectOption[]
  spokeProfileOptions: SpokeCalculatorSelectOption[]
  interlacingOptions: SpokeCalculatorSelectOption[]
}

/** Options owned by the recorded catalog/search system. */
export interface SpokeCalculatorCatalogOptions {
  rimBrandOptions: SpokeCalculatorSelectOption[]
  rimModelOptions: SpokeCalculatorSelectOption[]
  hubBrandOptions: SpokeCalculatorSelectOption[]
  hubModelOptions: SpokeCalculatorSelectOption[]
}

export type SpokeResultSource = 'calculated'

export interface SpokeWheelResult {
  leftLengthMm: number | null
  rightLengthMm: number | null
  tensionRatio: SpokeTensionRatio | null
  topologyId: string | null
  distribution: SpokeCalculatorTopologyDistribution | null
  spokeLengths: SpokeLengthResult[]
  leftSource: SpokeResultSource | null
  rightSource: SpokeResultSource | null
}

export interface SpokeLengthResult {
  id: number
  side: 'A' | 'B'
  physicalSide: 'left' | 'right'
  type: 'leading' | 'trailing' | 'nondrive'
  hubHoleId: number
  rimHoleId: number
  lengthMm: number
}

/** The wizard and calculator share one complete configuration per wheel. */
export type SpokeWheelWizardDraft = SpokeWheelBuildConfig

export interface SpokeCalculatorWizardDraft {
  front: SpokeWheelWizardDraft
  rear: SpokeWheelWizardDraft
}

export const createSpokeWheelWizardDraft = (): SpokeWheelWizardDraft => ({
  topologyId: '32h-symmetric-1to1-3x',
  spokeCount: 32,
  crossing: 3,
  g3RimHoleSpacingAToBDegrees: 0,
  g3RimHoleSpacingBToADegrees: 0,
  g3RimHoleSpacingAToNextGroupADegrees: 0,
  nippleType: 'standard',
  nippleLength: 12,
  spokeHeadType: 'j_bend',
  spokeHoleDiameterMm: 2.5,
  straightPullTangentOffsetMm: 0.8,
  spokeProfile: 'round_2_0',
  targetTensionN: 0,
  alternatingDrillingOffsetMm: 0,
  interlacing: 'off',
  interlaceCompensationMm: null,
  spokeElongationCompensationMm: null,
  erd: null,
  rimOffsetMm: 0,
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

export const createSpokeWheelCatalogSelection = (): SpokeWheelCatalogSelection => ({
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
})

export const createSpokeCalculatorWizardDraft = (): SpokeCalculatorWizardDraft => ({
  front: createSpokeWheelWizardDraft(),
  rear: createSpokeWheelWizardDraft(),
})
