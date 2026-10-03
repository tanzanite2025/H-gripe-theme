import { reactive, ref } from 'vue'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import {
  createSpokeCalculatorWizardDraft,
  SPOKE_WIZARD_STEPS,
  SPOKE_WIZARD_STEP_COUNT,
  type SpokeCalculatorWizardDraft,
  type SpokeHeadType,
  type SpokeInterlacing,
  type SpokeNippleType,
  type SpokeWheelSide,
  type SpokeWizardStep,
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
  const activeStep = ref<SpokeWizardStep>(SPOKE_WIZARD_STEPS[0].number)

  const goToStep = (step: number) => {
    const definition = SPOKE_WIZARD_STEPS.find(item => item.number === step)
    if (!definition) return false
    activeStep.value = definition.number
    return true
  }

  const nextStep = () => goToStep(activeStep.value + 1)
  const previousStep = () => goToStep(activeStep.value - 1)

  const setHeadType = (headType: SpokeHeadType) => {
    draft.front.spokeHeadType = headType
    draft.rear.spokeHeadType = headType
  }

  const setWheelHeadType = (side: SpokeWheelSide, headType: SpokeHeadType) => {
    draft[side].spokeHeadType = headType
  }

  const setErd = (side: SpokeWheelSide, erdMm: number | null) => {
    draft[side].erd = erdMm
  }

  const setHubGeometry = (side: SpokeWheelSide, geometry: HubGeometry) => {
    draft[side].leftFlange = geometry.leftFlange
    draft[side].rightFlange = geometry.rightFlange
    draft[side].leftFlangePcd = geometry.leftFlangePcd
    draft[side].rightFlangePcd = geometry.rightFlangePcd
  }

  const setAlternatingDrillingOffset = (side: SpokeWheelSide, offsetMm: number | null) => {
    draft[side].alternatingDrillingOffsetMm = offsetMm ?? 0
  }

  const setRimOffset = (side: SpokeWheelSide, offsetMm: number | null) => {
    draft[side].rimOffsetMm = offsetMm ?? 0
  }

  const setSpokeHoleDiameter = (side: SpokeWheelSide, diameterMm: number | null) => {
    draft[side].spokeHoleDiameterMm = diameterMm
  }

  const setInterlacing = (side: SpokeWheelSide, interlacing: SpokeInterlacing) => {
    draft[side].interlacing = interlacing
  }

  const setInterlaceCompensation = (side: SpokeWheelSide, compensationMm: number | null) => {
    draft[side].interlaceCompensationMm = compensationMm
  }

  const setSpokeElongationCompensation = (side: SpokeWheelSide, compensationMm: number | null) => {
    draft[side].spokeElongationCompensationMm = compensationMm
  }

  const setNippleType = (side: SpokeWheelSide, nippleType: SpokeNippleType) => {
    draft[side].nippleType = nippleType
  }

  const setNippleLength = (side: SpokeWheelSide, nippleLengthMm: number | null) => {
    draft[side].nippleLength = nippleLengthMm
  }

  return {
    draft,
    activeStep,
    stepCount: SPOKE_WIZARD_STEP_COUNT,
    availableStep: SPOKE_WIZARD_STEP_COUNT,
    goToStep,
    nextStep,
    previousStep,
    setHeadType,
    setWheelHeadType,
    setErd,
    setHubGeometry,
    setAlternatingDrillingOffset,
    setRimOffset,
    setSpokeHoleDiameter,
    setInterlacing,
    setInterlaceCompensation,
    setSpokeElongationCompensation,
    setNippleType,
    setNippleLength,
  }
}
