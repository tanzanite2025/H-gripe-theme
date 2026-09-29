import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import type { HubGeometry, HubModel, RimModel } from '~/data/spoke-calculator/database'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'
import type {
  SpokeCalculatorWheelOptions,
  SpokeWheelBuildConfig,
} from '~/types/spokeCalculator'

/**
 * Adapts the shared spoke catalog to the two wheel configurations used by the
 * calculator. Catalog loading stays in useSpokeCalculatorCatalog; this hook
 * owns only translated select options, dependent model lists, and the
 * geometry/ERD values populated after a model is selected.
 */
export const useSpokeCalculatorWheelCatalog = (
  frontConfig: SpokeWheelBuildConfig,
  rearConfig: SpokeWheelBuildConfig,
) => {
  const { t } = useI18n()
  const { rims, hubs, options: catalogOptions } = useSpokeCalculatorCatalog()

  const spokeCountOptions = computed(() => catalogOptions.value.spokeCounts)
  const crossingTranslationKeys: Record<number, string> = {
    0: 'radial',
    1: 'one',
    2: 'two',
    3: 'three',
    4: 'four',
  }
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

  const rimBrandOptions = computed(() => rims.value.map(brand => ({
    label: brand.name,
    value: brand.id,
  })))

  const hubBrandOptions = computed(() => hubs.value.map(brand => ({
    label: brand.name,
    value: brand.id,
  })))

  const rimModelsFor = (config: SpokeWheelBuildConfig) => computed<RimModel[]>(() => {
    if (!config.rimBrandId) return []
    const brand = rims.value.find(item => item.id === config.rimBrandId)
    return brand ? brand.items : []
  })

  const hubModelsFor = (config: SpokeWheelBuildConfig) => computed<HubModel[]>(() => {
    if (!config.hubBrandId) return []
    const brand = hubs.value.find(item => item.id === config.hubBrandId)
    return brand ? brand.items : []
  })

  const frontRimModels = rimModelsFor(frontConfig)
  const frontHubModels = hubModelsFor(frontConfig)
  const rearRimModels = rimModelsFor(rearConfig)
  const rearHubModels = hubModelsFor(rearConfig)

  const frontRimModelOptions = computed(() => frontRimModels.value.map(rim => ({
    label: rim.name,
    value: rim.id,
  })))
  const frontHubModelOptions = computed(() => frontHubModels.value.map(hub => ({
    label: hub.name,
    value: hub.id,
  })))
  const rearRimModelOptions = computed(() => rearRimModels.value.map(rim => ({
    label: rim.name,
    value: rim.id,
  })))
  const rearHubModelOptions = computed(() => rearHubModels.value.map(hub => ({
    label: hub.name,
    value: hub.id,
  })))

  const applyHubGeometry = (config: SpokeWheelBuildConfig, geometry?: HubGeometry | null) => {
    if (!geometry) return
    config.leftFlange = geometry.leftFlange ?? null
    config.rightFlange = geometry.rightFlange ?? null
    config.leftFlangePcd = geometry.leftFlangePcd ?? null
    config.rightFlangePcd = geometry.rightFlangePcd ?? null
    config.spokeHoleDiameterMm = geometry.spokeHoleDiameter ?? config.spokeHoleDiameterMm
  }

  const syncRimGeometry = (
    config: SpokeWheelBuildConfig,
    rimModelId: string | null,
    models: RimModel[],
  ) => {
    if (!rimModelId) {
      config.erd = null
      return
    }
    const model = models.find(item => item.id === rimModelId)
    if (model && model.erd != null) {
      config.erd = model.erd
    }
  }

  const syncHubGeometry = (
    config: SpokeWheelBuildConfig,
    hubModelId: string | null,
    models: HubModel[],
    wheel: 'front' | 'rear',
  ) => {
    if (!hubModelId) {
      config.leftFlange = null
      config.rightFlange = null
      config.leftFlangePcd = null
      config.rightFlangePcd = null
      return
    }
    const model = models.find(item => item.id === hubModelId)
    applyHubGeometry(config, wheel === 'front' ? model?.front : model?.rear)
  }

  watch(
    () => frontConfig.rimModelId,
    rimModelId => syncRimGeometry(frontConfig, rimModelId, frontRimModels.value),
  )
  watch(
    () => frontConfig.hubModelId,
    hubModelId => syncHubGeometry(frontConfig, hubModelId, frontHubModels.value, 'front'),
  )
  watch(
    () => rearConfig.rimModelId,
    rimModelId => syncRimGeometry(rearConfig, rimModelId, rearRimModels.value),
  )
  watch(
    () => rearConfig.hubModelId,
    hubModelId => syncHubGeometry(rearConfig, hubModelId, rearHubModels.value, 'rear'),
  )

  const frontOptions = computed<SpokeCalculatorWheelOptions>(() => ({
    spokeCountOptions: spokeCountOptions.value,
    lacingOptions: lacingOptions.value,
    nippleTypeOptions: nippleTypeOptions.value,
    rimBrandOptions: rimBrandOptions.value,
    rimModelOptions: frontRimModelOptions.value,
    hubBrandOptions: hubBrandOptions.value,
    hubModelOptions: frontHubModelOptions.value,
    spokeHeadTypeOptions: spokeHeadTypeOptions.value,
    spokeProfileOptions: spokeProfileOptions.value,
    interlacingOptions: interlacingOptions.value,
  }))

  const rearOptions = computed<SpokeCalculatorWheelOptions>(() => ({
    spokeCountOptions: spokeCountOptions.value,
    lacingOptions: lacingOptions.value,
    nippleTypeOptions: nippleTypeOptions.value,
    rimBrandOptions: rimBrandOptions.value,
    rimModelOptions: rearRimModelOptions.value,
    hubBrandOptions: hubBrandOptions.value,
    hubModelOptions: rearHubModelOptions.value,
    spokeHeadTypeOptions: spokeHeadTypeOptions.value,
    spokeProfileOptions: spokeProfileOptions.value,
    interlacingOptions: interlacingOptions.value,
  }))

  return {
    frontOptions,
    rearOptions,
  }
}
