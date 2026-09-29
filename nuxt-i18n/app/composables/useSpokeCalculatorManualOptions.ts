import { computed } from 'vue'
import { useI18n } from '#imports'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'
import type { SpokeCalculatorManualOptions } from '~/types/spokeCalculator'

/**
 * Builds the translated options used by the manual calculator only.
 *
 * This composable deliberately exposes no rim/hub catalog options. Those
 * belong to the independent recorded-result lookup system.
 */
export const useSpokeCalculatorManualOptions = () => {
  const { t } = useI18n()
  const { options: catalogOptions } = useSpokeCalculatorCatalog()

  const crossingTranslationKeys: Record<number, string> = {
    0: 'radial',
    1: 'one',
    2: 'two',
    3: 'three',
    4: 'four',
  }

  const spokeCountOptions = computed(() => catalogOptions.value.spokeCounts)
  const lacingOptions = computed(() => catalogOptions.value.crossings.map(option => ({
    ...option,
    label: t(
      `resourcesSpokeCalculator.calculator.options.crossing.${crossingTranslationKeys[option.value] || option.value}`,
      option.label,
    ),
  })))
  const nippleTypeOptions = computed(() => catalogOptions.value.nippleTypes.map(option => ({
    ...option,
    label: t(
      `resourcesSpokeCalculator.calculator.options.nippleType.${option.value}`,
      option.label,
    ),
  })))

  const spokeHeadTypeOptions = computed(() => [
    {
      value: 'j_bend',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.jBend'),
    },
    {
      value: 'straight_pull',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPull'),
    },
  ])

  const spokeProfileOptions = computed(() => [
    {
      value: 'round_2_0',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.round20'),
    },
    {
      value: 'round_1_8',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.round18'),
    },
    {
      value: 'bladed_0_9x2_2',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.bladed0922'),
    },
  ])

  const interlacingOptions = computed(() => [
    {
      value: 'off',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.off'),
    },
    {
      value: 'on',
      label: t('resourcesSpokeCalculator.calculator.physicalCorrections.on'),
    },
  ])

  const options = computed<SpokeCalculatorManualOptions>(() => ({
    spokeCountOptions: spokeCountOptions.value,
    lacingOptions: lacingOptions.value,
    nippleTypeOptions: nippleTypeOptions.value,
    spokeHeadTypeOptions: spokeHeadTypeOptions.value,
    spokeProfileOptions: spokeProfileOptions.value,
    interlacingOptions: interlacingOptions.value,
  }))

  return { options }
}
