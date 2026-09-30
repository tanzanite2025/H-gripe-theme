<template>
  <section class="schwalbe-filter-panel" :aria-label="label">
    <details name="schwalbe-filter-accordion" class="schwalbe-filter-panel__accordion" open @toggle="closeOtherAccordions">
      <summary class="schwalbe-filter-panel__accordion-title">
        <span>{{ tireWidthLabel }}</span>
        <span v-if="selectedTireWidthMinMm !== null || selectedTireWidthMaxMm !== null" class="schwalbe-filter-panel__selection-count">
          <span aria-hidden="true">1</span>
          <span class="schwalbe-filter-panel__visually-hidden">{{ selectedCountLabel(1) }}</span>
        </span>
      </summary>
      <div class="schwalbe-filter-panel__fields">
        <fieldset class="schwalbe-filter-panel__group">
          <legend class="schwalbe-filter-panel__visually-hidden">{{ tireWidthLabel }}</legend>
          <div class="schwalbe-filter-panel__range-inputs">
            <label class="schwalbe-filter-panel__range-input">
              <span class="schwalbe-filter-panel__range-label">{{ tireWidthMinLabel }}</span>
              <span class="schwalbe-filter-panel__range-control">
                <input
                  type="number"
                  step="1"
                  inputmode="numeric"
                  :min="tireWidthMinimum || 1"
                  :max="tireWidthMaximum || undefined"
                  :aria-label="`${tireWidthLabel} ${tireWidthMinLabel}`"
                  :value="selectedTireWidthMinMm ?? ''"
                  @change="updateTireWidthMinFromInput"
                >
                <span>mm</span>
              </span>
            </label>
            <label class="schwalbe-filter-panel__range-input">
              <span class="schwalbe-filter-panel__range-label">{{ tireWidthMaxLabel }}</span>
              <span class="schwalbe-filter-panel__range-control">
                <input
                  type="number"
                  step="1"
                  inputmode="numeric"
                  :min="tireWidthMinimum || 1"
                  :max="tireWidthMaximum || undefined"
                  :aria-label="`${tireWidthLabel} ${tireWidthMaxLabel}`"
                  :value="selectedTireWidthMaxMm ?? ''"
                  @change="updateTireWidthMaxFromInput"
                >
                <span>mm</span>
              </span>
            </label>
          </div>
          <div class="schwalbe-filter-panel__range-slider" :class="{ 'schwalbe-filter-panel__range-slider--empty': tireWidthValues.length === 0 }">
            <span class="schwalbe-filter-panel__range-track" aria-hidden="true" />
            <input
              class="schwalbe-filter-panel__range-slider-input schwalbe-filter-panel__range-slider-input--min"
              type="range"
              min="0"
              :max="tireWidthSliderMaximum"
              step="1"
              :value="tireWidthMinIndex"
              :disabled="tireWidthValues.length === 0"
              :aria-label="`${tireWidthLabel} ${tireWidthMinLabel}`"
              :aria-valuetext="`${selectedTireWidthMinMm ?? tireWidthMinimum} mm`"
              @input="updateTireWidthMinFromSlider"
            >
            <input
              class="schwalbe-filter-panel__range-slider-input schwalbe-filter-panel__range-slider-input--max"
              type="range"
              min="0"
              :max="tireWidthSliderMaximum"
              step="1"
              :value="tireWidthMaxIndex"
              :disabled="tireWidthValues.length === 0"
              :aria-label="`${tireWidthLabel} ${tireWidthMaxLabel}`"
              :aria-valuetext="`${selectedTireWidthMaxMm ?? tireWidthMaximum} mm`"
              @input="updateTireWidthMaxFromSlider"
            >
          </div>
          <p v-if="tireWidthValues.length > 0" class="schwalbe-filter-panel__hint">
            {{ tireWidthMinimum }}–{{ tireWidthMaximum }} mm
          </p>
          <button
            v-if="hasTireWidthSelection"
            type="button"
            class="schwalbe-filter-panel__clear-field"
            @click="clearTireWidth"
          >
            {{ tireWidthClearLabel }}
          </button>
        </fieldset>
      </div>
    </details>

    <details name="schwalbe-filter-accordion" class="schwalbe-filter-panel__accordion" @toggle="closeOtherAccordions">
      <summary class="schwalbe-filter-panel__accordion-title">
        <span>{{ wheelSizeLabel }}</span>
        <span v-if="selectedWheelSizeKeys.length > 0" class="schwalbe-filter-panel__selection-count">
          <span aria-hidden="true">{{ selectedWheelSizeKeys.length }}</span>
          <span class="schwalbe-filter-panel__visually-hidden">{{ selectedCountLabel(selectedWheelSizeKeys.length) }}</span>
        </span>
      </summary>
      <div class="schwalbe-filter-panel__fields">
        <fieldset class="schwalbe-filter-panel__group">
          <legend class="schwalbe-filter-panel__visually-hidden">{{ wheelSizeLabel }}</legend>
          <div
            class="schwalbe-filter-panel__options"
            :class="{ 'schwalbe-filter-panel__options--many': hasManyOptions(wheelSizeOptions) }"
          >
            <label v-for="option in wheelSizeOptions" :key="option.value" class="schwalbe-filter-panel__option">
              <input v-model="selectedWheelSizeKeys" type="checkbox" :value="option.value">
              <span>{{ wheelSizeOptionLabel(option) }}</span>
            </label>
          </div>
        </fieldset>

      </div>
    </details>

    <div class="schwalbe-filter-panel__rim-match" :title="rimWidthHint">
      <label class="schwalbe-filter-panel__rim-match-field">
        <span class="schwalbe-filter-panel__range-label">{{ rimWidthInnerWidthLabel }}</span>
        <span class="schwalbe-filter-panel__visually-hidden">{{ rimWidthHint }}</span>
        <span class="schwalbe-filter-panel__numeric-control schwalbe-filter-panel__rim-width-control">
          <input
            :value="innerRimWidthInput ?? ''"
            type="number"
            min="1"
            step="0.1"
            inputmode="decimal"
            :aria-label="rimWidthInnerWidthLabel"
            :aria-invalid="rimWidthInputInvalid || rimWidthWheelSizeInvalid"
            @input="updateInnerRimWidthInput"
          >
          <span>mm</span>
        </span>
      </label>
      <button
        v-if="hasInnerRimWidthInput"
        type="button"
        class="schwalbe-filter-panel__clear-field"
        @click="clearInnerRimWidth"
      >
        {{ rimWidthClearLabel }}
      </button>
      <p v-if="rimWidthInputInvalid" class="schwalbe-filter-panel__error" role="alert">
        {{ rimWidthInvalidLabel }}
      </p>
      <p v-else-if="rimWidthWheelSizeInvalid" class="schwalbe-filter-panel__error" role="alert">
        {{ rimWidthChooseWheelSizeLabel }}
      </p>
    </div>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ eBikeRatingLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label
          v-for="(option, index) in eBikeRatingOptions"
          :key="option.value ?? `unrated-${index}`"
          class="schwalbe-filter-panel__option"
        >
          <input v-model="selectedEBikeRatings" type="checkbox" :value="option.value">
          <span>{{ option.value ?? eBikeUnratedLabel }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ casingConstructionLabel }}</legend>
      <div
        class="schwalbe-filter-panel__options"
        :class="{ 'schwalbe-filter-panel__options--many': hasManyOptions(casingConstructionOptions) }"
      >
        <label v-for="option in casingConstructionOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedCasingConstructions" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ radialGroupLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label class="schwalbe-filter-panel__option">
          <input
            v-model="selectedRadialOnly"
            type="radio"
            name="schwalbe-radial-filter"
            :value="false"
          >
          <span>{{ radialAllLabel }}</span>
        </label>
        <label class="schwalbe-filter-panel__option">
          <input
            v-model="selectedRadialOnly"
            type="radio"
            name="schwalbe-radial-filter"
            :value="true"
          >
          <span>{{ radialLabel }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ beadLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in beadOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedBeads" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ sealLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in sealOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedSeals" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <button v-if="hasSelection" type="button" class="schwalbe-filter-panel__reset" @click="emit('reset')">
      {{ resetLabel }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type {
  SchwalbeTireCatalogFilterOption,
  SchwalbeTireCatalogWheelSizeOption,
} from '~/data/tireguides/schwalbeTireCatalogFilterModel'

const selectedTireWidthsMm = defineModel<number[]>('selectedTireWidthsMm', { required: true })
const selectedTireWidthMinMm = defineModel<number | null>('selectedTireWidthMinMm', { required: true })
const selectedTireWidthMaxMm = defineModel<number | null>('selectedTireWidthMaxMm', { required: true })
const selectedWheelSizeKeys = defineModel<string[]>('selectedWheelSizeKeys', { required: true })
const innerRimWidthInput = defineModel<string | number | null>('innerRimWidthInput', { required: true })
const selectedBeadSeatDiametersMm = defineModel<number[]>('selectedBeadSeatDiametersMm', { required: true })
const selectedCasingConstructions = defineModel<string[]>('selectedCasingConstructions', { required: true })
const selectedRadialOnly = defineModel<boolean>('selectedRadialOnly', { required: true })
const selectedBeads = defineModel<string[]>('selectedBeads', { required: true })
const selectedSeals = defineModel<string[]>('selectedSeals', { required: true })
const selectedEBikeRatings = defineModel<(string | null)[]>('selectedEBikeRatings', { required: true })
const emit = defineEmits<{
  reset: []
}>()

const props = defineProps<{
  label: string
  selectedCountTemplate: string
  tireWidthLabel: string
  tireWidthMinLabel: string
  tireWidthMaxLabel: string
  tireWidthClearLabel: string
  wheelSizeLabel: string
  wheelSizeOptionTemplate: string
  rimWidthHint: string
  rimWidthInnerWidthLabel: string
  rimWidthInvalidLabel: string
  rimWidthChooseWheelSizeLabel: string
  rimWidthClearLabel: string
  casingConstructionLabel: string
  radialGroupLabel: string
  radialAllLabel: string
  radialLabel: string
  beadLabel: string
  sealLabel: string
  eBikeRatingLabel: string
  eBikeUnratedLabel: string
  resetLabel: string
  tireWidthOptions: readonly SchwalbeTireCatalogFilterOption<number>[]
  wheelSizeOptions: readonly SchwalbeTireCatalogWheelSizeOption[]
  casingConstructionOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  beadOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  sealOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  eBikeRatingOptions: readonly SchwalbeTireCatalogFilterOption<string | null>[]
}>()

const hasSelection = computed(() => (
  selectedTireWidthMinMm.value !== null
  || selectedTireWidthMaxMm.value !== null
  || selectedTireWidthsMm.value.length > 0
  || selectedWheelSizeKeys.value.length > 0
  || hasInnerRimWidthInput.value
  || selectedBeadSeatDiametersMm.value.length > 0
  || selectedCasingConstructions.value.length > 0
  || selectedRadialOnly.value
  || selectedBeads.value.length > 0
  || selectedSeals.value.length > 0
  || selectedEBikeRatings.value.length > 0
))

const hasTireWidthSelection = computed(() => (
  selectedTireWidthMinMm.value !== null
  || selectedTireWidthMaxMm.value !== null
  || selectedTireWidthsMm.value.length > 0
))

const normalizedInnerRimWidthInput = computed(() => {
  const value = innerRimWidthInput.value
  const normalized = value === null || value === undefined ? '' : String(value).trim()
  if (!normalized) return null
  const parsed = Number(normalized)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
})

const hasInnerRimWidthInput = computed(() => {
  const value = innerRimWidthInput.value
  return Boolean(value !== null && value !== undefined && String(value).trim())
})

const rimWidthInputInvalid = computed(() => hasInnerRimWidthInput.value && normalizedInnerRimWidthInput.value === null)
const rimWidthWheelSizeInvalid = computed(() => hasInnerRimWidthInput.value && selectedWheelSizeKeys.value.length !== 1)

const clearInnerRimWidth = () => {
  innerRimWidthInput.value = ''
}

const wheelSizeOptionLabel = (option: SchwalbeTireCatalogWheelSizeOption) => props.wheelSizeOptionTemplate
  .replace('{diameter}', option.wheelDiameterIn)
  .replace('{bsd}', String(option.beadSeatDiameterMm))

const manyOptionThreshold = 5
const hasManyOptions = (options: readonly unknown[]) => options.length >= manyOptionThreshold

const tireWidthValues = computed(() => (
  [...new Set(props.tireWidthOptions
    .map(option => option.value)
    .filter(value => Number.isSafeInteger(value) && value > 0))]
    .sort((left, right) => left - right)
))

const tireWidthMinimum = computed(() => tireWidthValues.value[0] ?? 0)
const tireWidthMaximum = computed(() => tireWidthValues.value[tireWidthValues.value.length - 1] ?? 0)
const tireWidthSliderMaximum = computed(() => Math.max(0, tireWidthValues.value.length - 1))

const indexForTireWidth = (value: number | null, side: 'min' | 'max'): number => {
  const values = tireWidthValues.value
  if (values.length === 0 || value === null) return side === 'min' ? 0 : values.length - 1

  if (side === 'min') {
    const index = values.findIndex(candidate => candidate >= value)
    return index >= 0 ? index : values.length - 1
  }

  for (let index = values.length - 1; index >= 0; index -= 1) {
    const candidate = values[index]
    if (candidate !== undefined && candidate <= value) return index
  }
  return 0
}

const tireWidthMinIndex = computed(() => indexForTireWidth(selectedTireWidthMinMm.value, 'min'))
const tireWidthMaxIndex = computed(() => indexForTireWidth(selectedTireWidthMaxMm.value, 'max'))

const snapTireWidthValue = (value: number | null, side: 'min' | 'max'): number | null => {
  if (value === null || tireWidthValues.value.length === 0) return null
  return tireWidthValues.value[indexForTireWidth(value, side)] ?? null
}

const setTireWidthMinimum = (value: number | null) => {
  const nextValue = snapTireWidthValue(value, 'min')
  const currentMaximum = selectedTireWidthMaxMm.value
  innerRimWidthInput.value = ''
  selectedTireWidthsMm.value = []
  selectedTireWidthMinMm.value = nextValue
  if (currentMaximum !== null) {
    selectedTireWidthMaxMm.value = nextValue !== null && nextValue > currentMaximum
      ? nextValue
      : currentMaximum
  }
}

const setTireWidthMaximum = (value: number | null) => {
  const nextValue = snapTireWidthValue(value, 'max')
  const currentMinimum = selectedTireWidthMinMm.value
  innerRimWidthInput.value = ''
  selectedTireWidthsMm.value = []
  selectedTireWidthMaxMm.value = nextValue
  if (currentMinimum !== null) {
    selectedTireWidthMinMm.value = nextValue !== null && nextValue < currentMinimum
      ? nextValue
      : currentMinimum
  }
}

const readTireWidthInput = (event: Event): number | null => {
  const inputValue = (event.target as HTMLInputElement).value.trim()
  if (!inputValue) return null
  const value = Number(inputValue)
  return Number.isSafeInteger(value) && value > 0 ? value : null
}

const updateTireWidthMinFromInput = (event: Event) => setTireWidthMinimum(readTireWidthInput(event))
const updateTireWidthMaxFromInput = (event: Event) => setTireWidthMaximum(readTireWidthInput(event))

const clearTireWidth = () => {
  selectedTireWidthsMm.value = []
  selectedTireWidthMinMm.value = null
  selectedTireWidthMaxMm.value = null
}

const updateInnerRimWidthInput = (event: Event) => {
  const value = (event.target as HTMLInputElement).value
  innerRimWidthInput.value = value
  if (value.trim()) {
    selectedTireWidthsMm.value = []
    selectedTireWidthMinMm.value = null
    selectedTireWidthMaxMm.value = null
  }
}

const readTireWidthSliderIndex = (event: Event): number => {
  const value = Number((event.target as HTMLInputElement).value)
  return Number.isSafeInteger(value)
    ? Math.min(Math.max(value, 0), tireWidthSliderMaximum.value)
    : 0
}

const updateTireWidthMinFromSlider = (event: Event) => {
  setTireWidthMinimum(tireWidthValues.value[readTireWidthSliderIndex(event)] ?? null)
}

const updateTireWidthMaxFromSlider = (event: Event) => {
  setTireWidthMaximum(tireWidthValues.value[readTireWidthSliderIndex(event)] ?? null)
}

const closeOtherAccordions = (event: Event) => {
  const current = event.currentTarget
  if (!(current instanceof HTMLDetailsElement) || !current.open) return

  const panel = current.closest('.schwalbe-filter-panel')
  panel?.querySelectorAll<HTMLDetailsElement>('.schwalbe-filter-panel__accordion[open]').forEach((accordion) => {
    if (accordion !== current) accordion.open = false
  })
}

const selectedCountLabel = (count: number) => props.selectedCountTemplate.replace('{count}', String(count))

</script>

<style scoped>
.schwalbe-filter-panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.65rem;
}

.schwalbe-filter-panel__accordion {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.schwalbe-filter-panel__accordion-title {
  display: flex;
  min-height: 2.85rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.65rem 0.8rem;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 700;
  cursor: pointer;
  list-style: none;
}

.schwalbe-filter-panel__accordion-title::-webkit-details-marker {
  display: none;
}

.schwalbe-filter-panel__accordion-title::after {
  width: 0.45rem;
  height: 0.45rem;
  flex: 0 0 auto;
  border-right: 1.5px solid currentColor;
  border-bottom: 1.5px solid currentColor;
  content: '';
  transform: translateY(-0.12rem) rotate(45deg);
  transition: transform 0.16s ease;
}

.schwalbe-filter-panel__accordion[open] > .schwalbe-filter-panel__accordion-title {
  border-bottom: 1px solid var(--tz-border-subtle);
}

.schwalbe-filter-panel__accordion[open] > .schwalbe-filter-panel__accordion-title::after {
  transform: translateY(0.12rem) rotate(225deg);
}

.schwalbe-filter-panel__selection-count {
  display: inline-grid;
  width: 1.4rem;
  height: 1.4rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background: color-mix(in srgb, var(--tz-action-primary) 12%, var(--tz-card-surface));
  color: var(--tz-action-primary);
  font-size: 0.7rem;
  font-weight: 800;
}

.schwalbe-filter-panel__visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  clip-path: inset(50%);
}

.schwalbe-filter-panel__fields {
  display: flex;
  flex-direction: column;
  gap: 0.9rem;
  padding: 0.8rem;
}

.schwalbe-filter-panel__range-inputs {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.6rem;
}

.schwalbe-filter-panel__range-input {
  display: grid;
  min-width: 0;
  gap: 0.25rem;
}

.schwalbe-filter-panel__range-label {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  font-weight: 700;
}

.schwalbe-filter-panel__range-control {
  display: flex;
  min-height: 2.35rem;
  align-items: center;
  gap: 0.45rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.65rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  padding: 0.25rem 0.55rem;
}

.schwalbe-filter-panel__range-control input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 0.82rem;
}

.schwalbe-filter-panel__range-control input:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-filter-panel__range-control > span {
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.schwalbe-filter-panel__range-slider {
  position: relative;
  display: grid;
  min-height: 2.45rem;
  align-items: center;
  padding: 0 0.35rem;
}

.schwalbe-filter-panel__range-track {
  position: absolute;
  right: 0.35rem;
  left: 0.35rem;
  height: 0.3rem;
  border-radius: 999px;
  background: var(--tz-border-subtle);
}

.schwalbe-filter-panel__range-slider-input {
  position: relative;
  z-index: 2;
  grid-area: 1 / 1;
  width: 100%;
  height: 2.45rem;
  margin: 0;
  appearance: none;
  pointer-events: none;
  background: transparent;
}

.schwalbe-filter-panel__range-slider-input--max {
  z-index: 3;
}

.schwalbe-filter-panel__range-slider-input::-webkit-slider-runnable-track {
  height: 0.3rem;
  background: transparent;
}

.schwalbe-filter-panel__range-slider-input::-moz-range-track {
  height: 0.3rem;
  background: transparent;
}

.schwalbe-filter-panel__range-slider-input::-webkit-slider-thumb {
  width: 1.15rem;
  height: 1.15rem;
  margin-top: -0.425rem;
  appearance: none;
  border: 2px solid var(--tz-card-surface);
  border-radius: 999px;
  background: var(--tz-action-primary);
  box-shadow: 0 1px 4px rgb(15 23 42 / 0.28);
  cursor: pointer;
  pointer-events: auto;
}

.schwalbe-filter-panel__range-slider-input::-moz-range-thumb {
  width: 0.9rem;
  height: 0.9rem;
  border: 2px solid var(--tz-card-surface);
  border-radius: 999px;
  background: var(--tz-action-primary);
  box-shadow: 0 1px 4px rgb(15 23 42 / 0.28);
  cursor: pointer;
  pointer-events: auto;
}

.schwalbe-filter-panel__range-slider-input:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 0.2rem;
}

.schwalbe-filter-panel__range-slider-input:disabled {
  opacity: 0.5;
}

.schwalbe-filter-panel__range-slider--empty {
  min-height: 1rem;
}

.schwalbe-filter-panel__group {
  display: grid;
  align-content: start;
  gap: 0.45rem;
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.schwalbe-filter-panel__group legend {
  margin-bottom: 0.1rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
}

.schwalbe-filter-panel__options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  padding: 0.1rem 0.1rem 0.15rem;
}

.schwalbe-filter-panel__options--many {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: stretch;
}

.schwalbe-filter-panel__option {
  display: inline-flex;
  min-height: 2rem;
  align-items: center;
  gap: 0.35rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  padding: 0.25rem 0.6rem;
  font-size: 0.76rem;
  cursor: pointer;
  white-space: nowrap;
}

.schwalbe-filter-panel__options--many .schwalbe-filter-panel__option {
  width: 100%;
  min-width: 0;
  justify-content: flex-start;
  white-space: normal;
}

.schwalbe-filter-panel__option:has(input:checked) {
  border-color: var(--tz-action-primary);
  background: color-mix(in srgb, var(--tz-action-primary) 9%, var(--tz-card-surface));
}

.schwalbe-filter-panel__option input {
  width: 0.9rem;
  height: 0.9rem;
  margin: 0;
  accent-color: var(--tz-action-primary);
}

.schwalbe-filter-panel__option input:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-filter-panel__numeric-control {
  display: flex;
  width: min(100%, 10rem);
  min-height: 2rem;
  align-items: center;
  gap: 0.5rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.65rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  padding: 0.25rem 0.55rem;
}

.schwalbe-filter-panel__numeric-control input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: inherit;
  font: inherit;
  font-size: 0.8rem;
}

.schwalbe-filter-panel__numeric-control input:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-filter-panel__rim-match {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  padding: 0.45rem 0.7rem;
}

.schwalbe-filter-panel__rim-match-field {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  align-items: center;
  gap: 0.55rem;
}

.schwalbe-filter-panel__rim-match-field > .schwalbe-filter-panel__range-label {
  flex: 0 0 auto;
  white-space: nowrap;
}

.schwalbe-filter-panel__rim-width-control {
  flex: 0 1 12rem;
  width: 12rem;
}

.schwalbe-filter-panel__inline-facet {
  display: flex;
  min-width: 0;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.45rem 0.65rem;
  margin: 0;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  padding: 0.35rem 0.7rem;
}

.schwalbe-filter-panel__inline-facet > legend {
  flex: 0 0 auto;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  font-weight: 700;
  white-space: nowrap;
}

.schwalbe-filter-panel__inline-facet > .schwalbe-filter-panel__options {
  flex: 1 1 auto;
  min-width: 0;
  padding: 0;
}

.schwalbe-filter-panel__error {
  flex: 1 0 100%;
  margin: 0;
  color: var(--tz-status-danger-text);
  font-size: 0.68rem;
  line-height: 1.4;
}

.schwalbe-filter-panel__hint {
  max-width: 24rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.4;
}

.schwalbe-filter-panel__clear-field {
  justify-self: start;
  min-height: 1.9rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.55rem;
  background: transparent;
  color: var(--tz-text-primary);
  padding: 0.3rem 0.55rem;
  font: inherit;
  font-size: 0.7rem;
  font-weight: 700;
  cursor: pointer;
}

.schwalbe-filter-panel__clear-field:hover {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary);
}

.schwalbe-filter-panel__clear-field:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 2px;
}

.schwalbe-filter-panel__reset {
  grid-column: 1 / -1;
  justify-self: end;
  min-height: 2rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.6rem;
  background: transparent;
  color: var(--tz-text-primary);
  padding: 0.4rem 0.65rem;
  font: inherit;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
}

@media (max-width: 760.5px) {
  .schwalbe-filter-panel__option {
    min-height: 2.5rem;
  }
}
</style>
