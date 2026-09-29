import { reactive } from 'vue'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import {
  createSpokeCalculatorWizardDraft,
  type SpokeCalculatorWizardDraft,
  type SpokeHeadType,
  type SpokeWheelSide,
} from '~/types/spokeCalculator'

/**
 * Owns the values collected by the calculator wizard.
 *
 * Step components should emit changes through these methods. The page can
 * still expose computed v-model bridges while the older calculator panel is
 * migrated in small, reviewable steps.
 */
export const useSpokeCalculatorWizard = () => {
  const draft = reactive<SpokeCalculatorWizardDraft>(createSpokeCalculatorWizardDraft())

  const setHeadType = (headType: SpokeHeadType) => {
    draft.front.headType = headType
    draft.rear.headType = headType
  }

  const setWheelHeadType = (side: SpokeWheelSide, headType: SpokeHeadType) => {
    draft[side].headType = headType
  }

  const setErd = (side: SpokeWheelSide, erdMm: number | null) => {
    draft[side].erdMm = erdMm
  }

  const setHubGeometry = (side: SpokeWheelSide, geometry: HubGeometry) => {
    draft[side].hubGeometry = { ...geometry }
  }

  return {
    draft,
    setHeadType,
    setWheelHeadType,
    setErd,
    setHubGeometry,
  }
}
