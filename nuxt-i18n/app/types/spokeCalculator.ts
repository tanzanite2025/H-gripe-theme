import type { HubGeometry } from '~/data/spoke-calculator/database'

export type SpokeHeadType = 'j_bend' | 'straight_pull'
export type SpokeWheelSide = 'front' | 'rear'

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
