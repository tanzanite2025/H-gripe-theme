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

defineProps<{
  label: string
  tireWidthLabel: string
  beadSeatDiameterLabel: string
  scrollHint: string
  resetLabel: string
  tireWidthOptions: readonly SchwalbeTireCatalogFilterOption<number>[]
  beadSeatDiameterOptions: readonly SchwalbeTireCatalogFilterOption<number>[]
}>()

const hasSelection = computed(() => (
  selectedTireWidthsMm.value.length > 0 || selectedBeadSeatDiametersMm.value.length > 0
))

const resetFilters = () => {
  selectedTireWidthsMm.value = []
  selectedBeadSeatDiametersMm.value = []
}
</script>

<style scoped>
.schwalbe-filter-panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
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

@media (max-width: 760px) {
  .schwalbe-filter-panel {
    grid-template-columns: minmax(0, 1fr);
  }

  .schwalbe-filter-panel__options {
    max-height: 7rem;
  }
}
</style>
