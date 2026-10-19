import { computed, ref, watch, type Ref } from 'vue'
import type { SchwalbeTirePressureReferenceCatalogItem } from '~/data/tireguides/schwalbeCatalog'

export const minimumSupportedTirePressureNominalWidthMm = 20
export const maximumSupportedTirePressureNominalWidthMm = 120

/** Keeps catalog pressure data authoritative and permits a single operating
 * pressure input only when the selected record omits its minimum pressure. */
export function useTirePressureCatalogBoundsAndMissingMinimumOperatingPressureInput(
  selectedTireModel: Ref<SchwalbeTirePressureReferenceCatalogItem | null>,
) {
  const manualOperatingPressurePsi = ref<number | string>('')
  const selectedTireMinimumPressurePsi = computed(() => {
    const tire = selectedTireModel.value
    return tire?.min_pressure_psi ?? (tire?.min_pressure_bar === undefined ? null : tire.min_pressure_bar * 14.5037738)
  })
  const selectedTireMaximumPressurePsi = computed(() => {
    const tire = selectedTireModel.value
    return tire?.max_pressure_psi ?? (tire?.max_pressure_bar === undefined ? null : tire.max_pressure_bar * 14.5037738)
  })
  const selectedTirePressureBoundsPsi = computed(() => {
    const minimumPsi = selectedTireMinimumPressurePsi.value
    const maximumPsi = selectedTireMaximumPressurePsi.value
    if (minimumPsi === null || maximumPsi === null
      || !Number.isFinite(minimumPsi) || !Number.isFinite(maximumPsi)
      || minimumPsi <= 0 || maximumPsi <= 0 || minimumPsi > maximumPsi) return null
    return { minimumPsi, maximumPsi }
  })
  const manualOperatingPressureInputAllowed = computed(() => {
    const tire = selectedTireModel.value
    const maximumPsi = selectedTireMaximumPressurePsi.value
    return tire !== null
      && tire.min_pressure_psi === undefined && tire.min_pressure_bar === undefined
      && maximumPsi !== null && Number.isFinite(maximumPsi) && maximumPsi > 0
  })
  const manualOperatingPressureValidationState = computed<'REQUIRED' | 'INVALID' | 'ABOVE_CATALOG_MAXIMUM' | null>(() => {
    if (!manualOperatingPressureInputAllowed.value) return null
    const value = manualOperatingPressurePsi.value
    if (value === '') return 'REQUIRED'
    if (typeof value !== 'number' || !Number.isFinite(value) || value <= 0) return 'INVALID'
    if (value > selectedTireMaximumPressurePsi.value!) return 'ABOVE_CATALOG_MAXIMUM'
    return null
  })
  const selectedTireReferencePressurePsi = computed(() => {
    const bounds = selectedTirePressureBoundsPsi.value
    if (bounds) return Number(((bounds.minimumPsi + bounds.maximumPsi) / 2).toFixed(1))
    if (manualOperatingPressureInputAllowed.value && manualOperatingPressureValidationState.value === null) {
      return manualOperatingPressurePsi.value as number
    }
    return null
  })

  watch(selectedTireModel, () => { manualOperatingPressurePsi.value = '' }, { flush: 'sync' })

  return {
    selectedTireMinimumPressurePsi,
    selectedTireMaximumPressurePsi,
    selectedTirePressureBoundsPsi,
    selectedTireReferencePressurePsi,
    manualOperatingPressureInputAllowed,
    manualOperatingPressurePsi,
    manualOperatingPressureValidationState,
  }
}
