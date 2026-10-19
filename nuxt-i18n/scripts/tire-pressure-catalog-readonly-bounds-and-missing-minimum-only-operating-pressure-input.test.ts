import assert from 'node:assert/strict'
import { effectScope, ref } from 'vue'
import {
  useTirePressureCatalogBoundsAndMissingMinimumOperatingPressureInput,
} from '../app/composables/useTirePressureCatalogBoundsAndMissingMinimumOperatingPressureInput.js'
import type { SchwalbeTirePressureReferenceCatalogItem } from '../app/data/tireguides/schwalbeCatalog.js'

function createTirePressureReferenceCatalogFixture(
  pressureFields: Partial<SchwalbeTirePressureReferenceCatalogItem>,
): SchwalbeTirePressureReferenceCatalogItem {
  return { article_no: 'test-catalog-model', model_name: 'Catalog tire', etrto: '28-622', source_checked_at: '2026-09-28', ...pressureFields }
}

const selectedTireModel = ref<SchwalbeTirePressureReferenceCatalogItem | null>(null)
const pressureInputEffectScope = effectScope()
const pressureInputs = pressureInputEffectScope.run(() => (
  useTirePressureCatalogBoundsAndMissingMinimumOperatingPressureInput(selectedTireModel)
))!

try {
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false)
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null)

  selectedTireModel.value = createTirePressureReferenceCatalogFixture({ min_pressure_psi: 40, max_pressure_psi: 80 })
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false)
  assert.deepEqual(pressureInputs.selectedTirePressureBoundsPsi.value, { minimumPsi: 40, maximumPsi: 80 })
  pressureInputs.manualOperatingPressurePsi.value = 999
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, 60, 'Complete catalog data must ignore manual input')

  selectedTireModel.value = createTirePressureReferenceCatalogFixture({ min_pressure_bar: 2, max_pressure_bar: 4 })
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false, 'A bar-only minimum is still catalog data')
  assert.equal(pressureInputs.manualOperatingPressurePsi.value, '', 'Switching models must clear entered pressure')
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, 43.5)

  for (const incompleteOrInvalidCatalog of [
    {},
    { min_pressure_psi: 40 },
    { min_pressure_bar: 2 },
    { min_pressure_psi: 0, max_pressure_psi: 80 },
    { min_pressure_psi: 90, max_pressure_psi: 80 },
    { max_pressure_psi: 0 },
    { max_pressure_psi: Number.NaN },
  ]) {
    selectedTireModel.value = createTirePressureReferenceCatalogFixture(incompleteOrInvalidCatalog)
    assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false,
      'Manual input is allowed only for an omitted minimum and a usable catalog maximum')
    pressureInputs.manualOperatingPressurePsi.value = 10
    assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null)
  }

  selectedTireModel.value = createTirePressureReferenceCatalogFixture({
    article_no: '11654264', model_name: 'Al Mighty', etrto: '120-559', max_pressure_bar: 2, max_pressure_psi: 30,
  })
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, true)
  assert.equal(pressureInputs.manualOperatingPressureValidationState.value, 'REQUIRED')
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null, 'Do not invent a default pressure')
  for (const invalidPressure of [0, -1, Number.NaN, Number.POSITIVE_INFINITY, 'invalid']) {
    pressureInputs.manualOperatingPressurePsi.value = invalidPressure
    assert.equal(pressureInputs.manualOperatingPressureValidationState.value, 'INVALID')
    assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null)
  }
  pressureInputs.manualOperatingPressurePsi.value = 30.1
  assert.equal(pressureInputs.manualOperatingPressureValidationState.value, 'ABOVE_CATALOG_MAXIMUM')
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null)
  for (const validPressure of [10, 15.5, 30]) {
    pressureInputs.manualOperatingPressurePsi.value = validPressure
    assert.equal(pressureInputs.manualOperatingPressureValidationState.value, null)
    assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, validPressure)
    assert.equal(pressureInputs.selectedTirePressureBoundsPsi.value, null, 'Actual pressure must not become a catalog minimum')
  }

  selectedTireModel.value = createTirePressureReferenceCatalogFixture({ max_pressure_bar: 2 })
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, true)
  pressureInputs.manualOperatingPressurePsi.value = 29.1
  assert.equal(pressureInputs.manualOperatingPressureValidationState.value, 'ABOVE_CATALOG_MAXIMUM')
  pressureInputs.manualOperatingPressurePsi.value = 10
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, 10)

  selectedTireModel.value = createTirePressureReferenceCatalogFixture({ min_pressure_psi: 5, max_pressure_psi: 30, etrto: '120-559' })
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false, 'Wide tires with a catalog minimum remain read only')
  assert.equal(pressureInputs.manualOperatingPressurePsi.value, '')
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, 17.5)
  selectedTireModel.value = null
  assert.equal(pressureInputs.manualOperatingPressureInputAllowed.value, false)
  assert.equal(pressureInputs.selectedTireReferencePressurePsi.value, null)
  console.log('Catalog read-only bounds, missing-minimum-only input, maximum validation, and model-switch reset checks passed.')
} finally {
  pressureInputEffectScope.stop()
}
