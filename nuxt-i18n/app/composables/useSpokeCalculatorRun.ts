import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useSpokeCalculator } from '~/composables/useSpokeCalculator'
import type {
  SpokeWheelBuildConfig,
  SpokeWheelResult,
  SpokeWheelSide,
} from '~/types/spokeCalculator'

/**
 * Owns one explicit calculation run for the front and rear configurations.
 * Input editing stays in the wizard draft; API calls and result lifecycle stay
 * here so the visual calculator component does not become another state store.
 */
export const useSpokeCalculatorRun = (
  frontConfig: SpokeWheelBuildConfig,
  rearConfig: SpokeWheelBuildConfig,
) => {
  const { t } = useI18n()
  const { calculateWheel } = useSpokeCalculator()
  const { track: trackBehaviorEvent } = useBehaviorEvents()

  const loading = ref(false)
  const error = ref<string | null>(null)
  const frontResult = ref<SpokeWheelResult | null>(null)
  const rearResult = ref<SpokeWheelResult | null>(null)
  const lastTrackedCalculation = ref('')

  const buildWheelResult = async (
    config: SpokeWheelBuildConfig,
    wheel: SpokeWheelSide,
  ): Promise<SpokeWheelResult | null> => {
    let calculated: Awaited<ReturnType<typeof calculateWheel>>
    try {
      calculated = await calculateWheel(
        config,
        wheel,
        t('resourcesSpokeCalculator.calculator.action.calculationFailed'),
      )
    } catch (requestError: unknown) {
      error.value = requestError instanceof Error
        ? requestError.message
        : t('resourcesSpokeCalculator.calculator.action.calculationFailed')
      return null
    }

    const leftLengthMm = calculated?.leftLengthMm ?? null
    const rightLengthMm = calculated?.rightLengthMm ?? null

    if (leftLengthMm == null && rightLengthMm == null) return null

    return {
      leftLengthMm,
      rightLengthMm,
      tensionRatio: calculated?.tensionRatio ?? null,
      leftSource: calculated?.leftLengthMm != null ? 'calculated' : null,
      rightSource: calculated?.rightLengthMm != null ? 'calculated' : null,
    }
  }

  const updateResults = async () => {
    const [front, rear] = await Promise.all([
      buildWheelResult(frontConfig, 'front'),
      buildWheelResult(rearConfig, 'rear'),
    ])
    frontResult.value = front
    rearResult.value = rear

    return [frontResult.value, rearResult.value].filter(result => (
      result && (result.leftLengthMm != null || result.rightLengthMm != null)
    )).length
  }

  const onCalculate = async () => {
    error.value = null
    loading.value = true

    try {
      const completedWheelCount = await updateResults()

      if (completedWheelCount > 0) {
        const fingerprint = JSON.stringify({
          front: frontConfig,
          rear: rearConfig,
        })

        if (fingerprint !== lastTrackedCalculation.value) {
          lastTrackedCalculation.value = fingerprint
          trackBehaviorEvent({
            eventType: 'calculator_use',
            metadata: {
              source: 'spoke_calculator',
              wheel_count: completedWheelCount,
              front_spoke_count: frontConfig.spokeCount,
              rear_spoke_count: rearConfig.spokeCount,
              front_crossing: frontConfig.crossing,
              rear_crossing: rearConfig.crossing,
              front_rim_offset_mm: frontConfig.rimOffsetMm,
              rear_rim_offset_mm: rearConfig.rimOffsetMm,
              front_rim_selected: Boolean(frontConfig.rimModelId),
              rear_rim_selected: Boolean(rearConfig.rimModelId),
              front_hub_selected: Boolean(frontConfig.hubModelId),
              rear_hub_selected: Boolean(rearConfig.hubModelId),
            },
          })
        }
      }
    } catch (requestError: unknown) {
      error.value = requestError instanceof Error
        ? requestError.message
        : t('resourcesSpokeCalculator.calculator.action.calculationFailed')
    } finally {
      loading.value = false
    }
  }

  watch(
    () => ({
      front: { ...frontConfig },
      rear: { ...rearConfig },
    }),
    () => {
      // Results are generated explicitly by the Calculate action. Avoid firing
      // an API request for every slider/input keystroke (and wasting the quota).
      frontResult.value = null
      rearResult.value = null
    },
    { deep: true },
  )

  return {
    loading,
    error,
    frontResult,
    rearResult,
    onCalculate,
  }
}
