import type { SpokeTensionRatio } from '~~/types/spoke'

export type SpokeHeadType = 'j_bend' | 'straight_pull'
export type SpokeWheelSide = 'front' | 'rear'
export type SpokeNippleType = 'standard' | 'hidden'
export type SpokeProfile = 'round_2_0' | 'round_1_8' | 'bladed_0_9x2_2'
export type SpokeInterlacing = 'off' | 'on'

export const SPOKE_WIZARD_STEPS = [
  { id: 'head_type', number: 1 },
  { id: 'erd', number: 2 },
  { id: 'alternating_drilling', number: 3 },
  { id: 'hub_geometry', number: 4 },
] as const

export type SpokeWizardStep = typeof SPOKE_WIZARD_STEPS[number]['number']
export const SPOKE_WIZARD_STEP_COUNT = SPOKE_WIZARD_STEPS.length

/** Complete input configuration for one wheel. */
export interface SpokeWheelBuildConfig {
  spokeCount: number
  crossing: number
  nippleType: SpokeNippleType
  nippleLength: number | null
  spokeHeadType: SpokeHeadType
  spokeHoleDiameterMm: number | null
  straightPullTangentOffsetMm: number
  spokeProfile: SpokeProfile
  targetTensionN: number
  alternatingDrillingOffsetMm: number
  interlacing: SpokeInterlacing
  interlaceCompensationMm: number

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
  spokeCountOptions: SpokeCalculatorSelectOption[]
  lacingOptions: SpokeCalculatorSelectOption[]
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
  leftSource: SpokeResultSource | null
  rightSource: SpokeResultSource | null
}

/** The wizard and calculator share one complete configuration per wheel. */
export type SpokeWheelWizardDraft = SpokeWheelBuildConfig

export interface SpokeCalculatorWizardDraft {
  front: SpokeWheelWizardDraft
  rear: SpokeWheelWizardDraft
}

export const createSpokeWheelWizardDraft = (): SpokeWheelWizardDraft => ({
  spokeCount: 32,
  crossing: 3,
  nippleType: 'standard',
  nippleLength: 12,
  spokeHeadType: 'j_bend',
  spokeHoleDiameterMm: 2.5,
  straightPullTangentOffsetMm: 0.8,
  spokeProfile: 'round_2_0',
  targetTensionN: 0,
  alternatingDrillingOffsetMm: 0,
  interlacing: 'off',
  interlaceCompensationMm: 0.45,
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
