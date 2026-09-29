import type { SpokeTensionRatio } from '~~/types/spoke'

export type SpokeHeadType = 'j_bend' | 'straight_pull'
export type SpokeWheelSide = 'front' | 'rear'
export type SpokeNippleType = 'standard' | 'hidden'
export type SpokeProfile = 'round_2_0' | 'round_1_8' | 'bladed_0_9x2_2'
export type SpokeInterlacing = 'off' | 'on'

export const SPOKE_WIZARD_STEPS = [
  { id: 'head_type', number: 1 },
  { id: 'erd', number: 2 },
  { id: 'hub_geometry', number: 3 },
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

  rimBrandId: string | null
  rimModelId: string | null
  hubBrandId: string | null
  hubModelId: string | null

  erd: number | null
  rimOffsetMm: number
  leftFlange: number | null
  rightFlange: number | null
  leftFlangePcd: number | null
  rightFlangePcd: number | null
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
  rimBrandId: null,
  rimModelId: null,
  hubBrandId: null,
  hubModelId: null,
  erd: null,
  rimOffsetMm: 0,
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

export const createSpokeCalculatorWizardDraft = (): SpokeCalculatorWizardDraft => ({
  front: createSpokeWheelWizardDraft(),
  rear: createSpokeWheelWizardDraft(),
})
