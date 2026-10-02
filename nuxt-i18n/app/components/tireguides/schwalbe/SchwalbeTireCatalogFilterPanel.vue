<template>
  <section class="schwalbe-filter-panel" :aria-label="label">
    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ eBikeRatingLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label class="schwalbe-filter-panel__option schwalbe-filter-panel__option--all">
          <input
            :checked="selectedEBikeRatings.length === 0"
            type="checkbox"
            @change="selectedEBikeRatings = []"
          >
          <span>{{ eBikeRatingAllLabel }}</span>
        </label>
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
        <label class="schwalbe-filter-panel__option schwalbe-filter-panel__option--all">
          <input
            :checked="selectedBeads.length === 0"
            type="checkbox"
            @change="selectedBeads = []"
          >
          <span>{{ beadAllLabel }}</span>
        </label>
        <label v-for="option in beadOptions" :key="option.value" class="schwalbe-filter-panel__option">
          <input v-model="selectedBeads" type="checkbox" :value="option.value">
          <span>{{ option.value }}</span>
        </label>
      </div>
    </fieldset>

    <fieldset class="schwalbe-filter-panel__inline-facet">
      <legend>{{ sealLabel }}</legend>
      <div class="schwalbe-filter-panel__options">
        <label class="schwalbe-filter-panel__option schwalbe-filter-panel__option--all">
          <input
            :checked="selectedSeals.length === 0"
            type="checkbox"
            @change="selectedSeals = []"
          >
          <span>{{ sealAllLabel }}</span>
        </label>
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
} from '~/data/tireguides/schwalbeTireCatalogFilterModel'

const selectedRadialOnly = defineModel<boolean>('selectedRadialOnly', { required: true })
const selectedBeads = defineModel<string[]>('selectedBeads', { required: true })
const selectedSeals = defineModel<string[]>('selectedSeals', { required: true })
const selectedEBikeRatings = defineModel<(string | null)[]>('selectedEBikeRatings', { required: true })
const emit = defineEmits<{
  reset: []
}>()

defineProps<{
  label: string
  radialGroupLabel: string
  radialAllLabel: string
  radialLabel: string
  beadLabel: string
  beadAllLabel: string
  sealLabel: string
  sealAllLabel: string
  eBikeRatingLabel: string
  eBikeRatingAllLabel: string
  eBikeUnratedLabel: string
  resetLabel: string
  beadOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  sealOptions: readonly SchwalbeTireCatalogFilterOption<string>[]
  eBikeRatingOptions: readonly SchwalbeTireCatalogFilterOption<string | null>[]
}>()

const hasSelection = computed(() => (
  selectedRadialOnly.value
  || selectedBeads.value.length > 0
  || selectedSeals.value.length > 0
  || selectedEBikeRatings.value.length > 0
))

</script>

<style scoped>
.schwalbe-filter-panel {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.65rem;
}

.schwalbe-filter-panel__options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.4rem;
  padding: 0.1rem 0.1rem 0.15rem;
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

.schwalbe-filter-panel__option--all input {
  appearance: none;
  flex: 0 0 0.9rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 50%;
  background: var(--tz-card-surface);
  cursor: pointer;
}

.schwalbe-filter-panel__option--all input:checked {
  border-color: var(--tz-text-primary);
  background: radial-gradient(circle, var(--tz-text-primary) 0 34%, transparent 39%);
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
