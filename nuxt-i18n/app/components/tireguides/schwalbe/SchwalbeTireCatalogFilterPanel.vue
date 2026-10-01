<template>
  <section class="schwalbe-filter-panel" :aria-label="label">
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

const selectedWheelSizeKeys = defineModel<string[]>('selectedWheelSizeKeys', { required: true })
const selectedBeadSeatDiametersMm = defineModel<number[]>('selectedBeadSeatDiametersMm', { required: true })
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
  wheelSizeLabel: string
  wheelSizeOptionTemplate: string
  radialGroupLabel: string
  radialAllLabel: string
  radialLabel: string
  beadLabel: string
  sealLabel: string
  eBikeRatingLabel: string
  eBikeUnratedLabel: string
  resetLabel: string
  wheelSizeOptions: readonly SchwalbeTireCatalogWheelSizeOption[]
  beadOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  sealOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  eBikeRatingOptions: readonly SchwalbeTireCatalogFilterOption<string | null>[]
}>()

const hasSelection = computed(() => (
  selectedWheelSizeKeys.value.length > 0
  || selectedBeadSeatDiametersMm.value.length > 0
  || selectedRadialOnly.value
  || selectedBeads.value.length > 0
  || selectedSeals.value.length > 0
  || selectedEBikeRatings.value.length > 0
))

const wheelSizeOptionLabel = (option: SchwalbeTireCatalogWheelSizeOption) => props.wheelSizeOptionTemplate
  .replace('{diameter}', option.wheelDiameterIn)
  .replace('{bsd}', String(option.beadSeatDiameterMm))

const manyOptionThreshold = 5
const hasManyOptions = (options: readonly unknown[]) => options.length >= manyOptionThreshold

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
