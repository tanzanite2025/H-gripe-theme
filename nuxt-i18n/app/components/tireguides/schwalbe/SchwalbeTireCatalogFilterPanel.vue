<template>
  <section class="schwalbe-filter-panel" :aria-label="label">
    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ tireWidthLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in tireWidthOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input
            v-model="selectedTireWidthsMm"
            type="checkbox"
            :value="option.value"
          >
          <span>{{ option.value }} mm</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ beadSeatDiameterLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in beadSeatDiameterOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input
            v-model="selectedBeadSeatDiametersMm"
            type="checkbox"
            :value="option.value"
          >
          <span>{{ option.value }} mm</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ beadLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in beadOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedBeads" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ sealLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in sealOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedSeals" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__group">
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

    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ colorLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in colorOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedColors" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__group">
      <legend>{{ compoundLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label v-for="option in compoundOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedCompounds" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <button
      v-if="hasSelection"
      type="button"
      class="schwalbe-filter-panel__reset"
      @click="resetFilters"
    >
      {{ resetLabel }}
    </button>

    <p class="schwalbe-filter-panel__hint">{{ scrollHint }}</p>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SchwalbeTireCatalogFilterOption } from '~/data/tireguides/schwalbeTireCatalogFilterModel'

const selectedTireWidthsMm = defineModel<number[]>('selectedTireWidthsMm', { required: true })
const selectedBeadSeatDiametersMm = defineModel<number[]>('selectedBeadSeatDiametersMm', { required: true })
const selectedBeads = defineModel<string[]>('selectedBeads', { required: true })
const selectedSeals = defineModel<string[]>('selectedSeals', { required: true })
const selectedEBikeRatings = defineModel<(string | null)[]>('selectedEBikeRatings', { required: true })
const selectedColors = defineModel<string[]>('selectedColors', { required: true })
const selectedCompounds = defineModel<string[]>('selectedCompounds', { required: true })

defineProps<{
  label: string
  tireWidthLabel: string
  beadSeatDiameterLabel: string
  beadLabel: string
  sealLabel: string
  eBikeRatingLabel: string
  eBikeUnratedLabel: string
  colorLabel: string
  compoundLabel: string
  scrollHint: string
  resetLabel: string
  tireWidthOptions: readonly SchwalbeTireCatalogFilterOption<number>[]
  beadSeatDiameterOptions: readonly SchwalbeTireCatalogFilterOption<number>[]
  beadOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  sealOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  eBikeRatingOptions: readonly SchwalbeTireCatalogFilterOption<string | null>[]
  colorOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  compoundOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
}>()

const hasSelection = computed(() => (
  selectedTireWidthsMm.value.length > 0
  || selectedBeadSeatDiametersMm.value.length > 0
  || selectedBeads.value.length > 0
  || selectedSeals.value.length > 0
  || selectedEBikeRatings.value.length > 0
  || selectedColors.value.length > 0
  || selectedCompounds.value.length > 0
))

const resetFilters = () => {
  selectedTireWidthsMm.value = []
  selectedBeadSeatDiametersMm.value = []
  selectedBeads.value = []
  selectedSeals.value = []
  selectedEBikeRatings.value = []
  selectedColors.value = []
  selectedCompounds.value = []
}
</script>

<style scoped>
.schwalbe-filter-panel {
  display: grid;
  grid-template-columns: repeat(7, minmax(0, 1fr)) auto;
  gap: 0.75rem;
  align-items: end;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.85rem;
  background: var(--tz-surface-subtle);
  padding: 0.85rem;
}

.schwalbe-filter-panel__group {
  display: grid;
  gap: 0.5rem;
  align-self: start;
  min-width: 0;
  margin: 0;
  border: 0;
  padding: 0;
}

.schwalbe-filter-panel__group legend {
  margin-bottom: 0.15rem;
  color: var(--tz-text-primary);
  font-size: 0.76rem;
  font-weight: 700;
}

.schwalbe-filter-panel__options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  max-height: 5.2rem;
  overflow: auto;
  scrollbar-color: var(--tz-border-strong) transparent;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  padding: 0.1rem 0.1rem 0.15rem;
}

.schwalbe-filter-panel__options::-webkit-scrollbar {
  width: 0.4rem;
}

.schwalbe-filter-panel__options::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: var(--tz-border-strong);
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

.schwalbe-filter-panel__reset {
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

.schwalbe-filter-panel__hint {
  grid-column: 1 / -1;
  margin: -0.2rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  line-height: 1.35;
}

@media (min-width: 761px) {
  .schwalbe-filter-panel__options {
    max-height: 8.5rem;
  }
}

@media (min-width: 761px) and (max-width: 1100px) {
  .schwalbe-filter-panel {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1101px) and (max-width: 1440px) {
  .schwalbe-filter-panel {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}

@media (max-width: 760px) {
  .schwalbe-filter-panel {
    grid-template-columns: minmax(0, 1fr);
  }

  .schwalbe-filter-panel__options {
    max-height: 7rem;
  }
}
</style>
