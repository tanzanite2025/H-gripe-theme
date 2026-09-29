import { computed } from 'vue'
import type { HubModel, RimModel } from '~/data/spoke-calculator/database'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'
import type {
  SpokeCalculatorCatalogOptions,
  SpokeWheelCatalogSelection,
} from '~/types/spokeCalculator'

/**
 * Adapts the shared spoke catalog to the two wheel configurations used by the
 * recorded-result lookup. Selecting a catalog item never writes dimensions
 * into the manual calculator draft.
 */
export const useSpokeCalculatorWheelCatalog = (
  frontSelection: SpokeWheelCatalogSelection,
  rearSelection: SpokeWheelCatalogSelection,
) => {
  const { rims, hubs } = useSpokeCalculatorCatalog()

  const rimBrandOptions = computed(() => rims.value.map(brand => ({
    label: brand.name,
    value: brand.id,
  })))

  const hubBrandOptions = computed(() => hubs.value.map(brand => ({
    label: brand.name,
    value: brand.id,
  })))

  const rimModelsFor = (selection: SpokeWheelCatalogSelection) => computed<RimModel[]>(() => {
    if (!selection.rimBrandId) return []
    const brand = rims.value.find(item => item.id === selection.rimBrandId)
    return brand ? brand.items : []
  })

  const hubModelsFor = (selection: SpokeWheelCatalogSelection) => computed<HubModel[]>(() => {
    if (!selection.hubBrandId) return []
    const brand = hubs.value.find(item => item.id === selection.hubBrandId)
    return brand ? brand.items : []
  })

  const frontRimModels = rimModelsFor(frontSelection)
  const frontHubModels = hubModelsFor(frontSelection)
  const rearRimModels = rimModelsFor(rearSelection)
  const rearHubModels = hubModelsFor(rearSelection)

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

  const frontOptions = computed<SpokeCalculatorCatalogOptions>(() => ({
    rimBrandOptions: rimBrandOptions.value,
    rimModelOptions: frontRimModelOptions.value,
    hubBrandOptions: hubBrandOptions.value,
    hubModelOptions: frontHubModelOptions.value,
  }))

  const rearOptions = computed<SpokeCalculatorCatalogOptions>(() => ({
    rimBrandOptions: rimBrandOptions.value,
    rimModelOptions: rearRimModelOptions.value,
    hubBrandOptions: hubBrandOptions.value,
    hubModelOptions: rearHubModelOptions.value,
  }))

  return {
    frontOptions,
    rearOptions,
  }
}
