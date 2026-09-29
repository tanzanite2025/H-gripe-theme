import { reactive } from 'vue'
import {
  createSpokeWheelCatalogSelection,
  type SpokeWheelCatalogSelection,
} from '~/types/spokeCalculator'

/**
 * Owns the catalog/preset lookup selection shown below the manual calculator.
 * These IDs are deliberately separate from the measured calculation draft.
 */
export const useSpokeCalculatorCatalogSelection = () => {
  const front = reactive<SpokeWheelCatalogSelection>(createSpokeWheelCatalogSelection())
  const rear = reactive<SpokeWheelCatalogSelection>(createSpokeWheelCatalogSelection())

  return {
    front,
    rear,
  }
}
