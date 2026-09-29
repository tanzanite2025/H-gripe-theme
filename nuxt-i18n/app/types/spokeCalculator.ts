import type { HubGeometry } from '~/data/spoke-calculator/database'
import type { SpokeTensionRatio } from '~~/types/spoke'

export type SpokeHeadType = 'j_bend' | 'straight_pull'
export type SpokeWheelSide = 'front' | 'rear'
export type SpokeNippleType = 'standard' | 'hidden'
export type SpokeProfile = 'round_2_0' | 'round_1_8' | 'bladed_0_9x2_2'
export type SpokeInterlacing = 'off' | 'on'

/**
 * The configuration currently edited by the legacy calculator panel.
 * Keeping this contract outside the component lets the request adapter and
 * the panel share one field definition while the panel is being split.
 */
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

/**
 * The small, page-owned draft used by the spoke calculator wizard.
 *
 * The calculator still has a larger build configuration in
 * SpokeCalculatorBlueprint.vue. Keeping this draft limited to the wizard's
 * first three steps lets us move those fields one group at a time without
 * creating a second source of truth for the values already collected here.
 */
export interface SpokeWheelWizardDraft {
  headType: SpokeHeadType
  erdMm: number | null
  hubGeometry: HubGeometry
}

export interface SpokeCalculatorWizardDraft {
  front: SpokeWheelWizardDraft
  rear: SpokeWheelWizardDraft
}

export const createEmptySpokeHubGeometry = (): HubGeometry => ({
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

export const createSpokeWheelWizardDraft = (): SpokeWheelWizardDraft => ({
  headType: 'j_bend',
  erdMm: null,
  hubGeometry: createEmptySpokeHubGeometry(),
})

export const createSpokeCalculatorWizardDraft = (): SpokeCalculatorWizardDraft => ({
  front: createSpokeWheelWizardDraft(),
  rear: createSpokeWheelWizardDraft(),
})
