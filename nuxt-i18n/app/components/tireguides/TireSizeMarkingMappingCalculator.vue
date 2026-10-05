<template>
  <section
    class="tire-size-marking-mapping-calculator rounded-2xl bg-[var(--tz-card-surface)] p-4 shadow-md md:p-5"
    :aria-label="labels.title"
  >
    <div class="tire-size-marking-mapping-calculator__intro">
      <h3 class="text-lg font-bold tz-text-primary">
        {{ labels.title }}
      </h3>
      <p class="tz-text-secondary">
        {{ labels.description }}
      </p>
    </div>

    <div
      class="tire-size-marking-mapping-calculator__controls"
      :aria-busy="completeTireSizeMarkingMappingRows.length === 0"
    >
      <label class="tire-size-marking-mapping-calculator__field">
        <span class="tire-size-marking-mapping-calculator__label">
          {{ labels.etrto }}
        </span>
        <SpokeCalculatorSelect
          id="tire-size-marking-etrto-select"
          :model-value="selectedTireSizeMarkingMappingIdentifier"
          :options="etrtoTireSizeMarkingMappingOptions"
          :aria-label="labels.etrto"
          :keep-menu-within-viewport="true"
          :disabled="completeTireSizeMarkingMappingRows.length === 0"
          @update:model-value="updateSelectedTireSizeMarkingIdentifier"
        />
      </label>

      <label class="tire-size-marking-mapping-calculator__field">
        <span class="tire-size-marking-mapping-calculator__label">
          {{ labels.inch }}
        </span>
        <SpokeCalculatorSelect
          id="tire-size-marking-inch-select"
          :model-value="selectedTireSizeMarkingMappingIdentifier"
          :options="inchTireSizeMarkingMappingOptions"
          :aria-label="labels.inch"
          :keep-menu-within-viewport="true"
          :disabled="completeTireSizeMarkingMappingRows.length === 0"
          @update:model-value="updateSelectedTireSizeMarkingIdentifier"
        />
      </label>

      <label class="tire-size-marking-mapping-calculator__field">
        <span class="tire-size-marking-mapping-calculator__label">
          {{ labels.french }}
        </span>
        <SpokeCalculatorSelect
          id="tire-size-marking-french-select"
          :model-value="selectedTireSizeMarkingMappingIdentifier"
          :options="frenchTireSizeMarkingMappingOptions"
          :aria-label="labels.french"
          :keep-menu-within-viewport="true"
          :disabled="completeTireSizeMarkingMappingRows.length === 0"
          @update:model-value="updateSelectedTireSizeMarkingIdentifier"
        />
      </label>
    </div>

    <div
      v-if="selectedTireSizeMarkingMappingRow"
      class="tire-size-marking-mapping-calculator__current"
    >
      <span class="tire-size-marking-mapping-calculator__current-label">
        {{ labels.resultLabel }}
      </span>
      <div class="tire-size-marking-mapping-calculator__current-values">
        <strong>{{ selectedTireSizeMarkingMappingRow.etrto }}</strong>
        <span aria-hidden="true">·</span>
        <span>{{ selectedTireSizeMarkingMappingRow.inch }}</span>
        <span aria-hidden="true">·</span>
        <span>{{ selectedTireSizeMarkingMappingRow.french }}</span>
      </div>
    </div>

    <p class="tire-size-marking-mapping-calculator__note tz-text-muted">
      {{ labels.note }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import {
  tireSizeMarkingMappingRows,
  type TireSizeMarkingMappingRow,
} from '~/data/tireguides/tireSizeMarkingMappingData'

interface TireSizeMarkingMappingCalculatorLabels {
  title: string
  description: string
  etrto: string
  inch: string
  french: string
  resultLabel: string
  note: string
}

interface TireSizeMarkingMappingCalculatorProperties {
  labels: TireSizeMarkingMappingCalculatorLabels
  mappingRows?: readonly TireSizeMarkingMappingRow[]
  initialMappingIdentifier?: string
  initialEtrto?: string
  initialInch?: string
  initialFrench?: string
}

interface CompleteTireSizeMarkingMappingRow extends TireSizeMarkingMappingRow {
  mappingIdentifier: string
}

interface TireSizeMarkingMappingCalculatorOption {
  value: string
  label: string
}

type TireSizeMarkingMappingOptionField = 'etrto' | 'inch' | 'french'

const props = withDefaults(defineProps<TireSizeMarkingMappingCalculatorProperties>(), {
  mappingRows: () => tireSizeMarkingMappingRows,
})

const completeTireSizeMarkingMappingRows = computed<CompleteTireSizeMarkingMappingRow[]>(() => (
  props.mappingRows
    .filter(row => Boolean(row.etrto && row.inch && row.french))
    .map((row, rowIndex) => ({
      ...row,
      mappingIdentifier: `tire-size-marking-mapping-${rowIndex + 1}`,
    }))
))

const selectedTireSizeMarkingMappingIdentifier = ref('')

const findInitialTireSizeMarkingMappingIdentifier = (): string => {
  const availableRows = completeTireSizeMarkingMappingRows.value
  const requestedMappingIdentifier = props.initialMappingIdentifier?.trim()
  if (requestedMappingIdentifier && availableRows.some(row => row.mappingIdentifier === requestedMappingIdentifier)) {
    return requestedMappingIdentifier
  }

  const requestedFieldValues: Array<[TireSizeMarkingMappingOptionField, string | undefined]> = [
    ['etrto', props.initialEtrto],
    ['inch', props.initialInch],
    ['french', props.initialFrench],
  ]

  for (const [field, requestedValue] of requestedFieldValues) {
    const normalizedRequestedValue = requestedValue?.trim()
    if (!normalizedRequestedValue) continue

    const matchingRow = availableRows.find(row => row[field] === normalizedRequestedValue)
    if (matchingRow) return matchingRow.mappingIdentifier
  }

  return availableRows[0]?.mappingIdentifier ?? ''
}

watch(
  completeTireSizeMarkingMappingRows,
  (availableRows) => {
    const selectedMappingStillExists = availableRows.some(
      row => row.mappingIdentifier === selectedTireSizeMarkingMappingIdentifier.value,
    )
    if (!selectedMappingStillExists) {
      selectedTireSizeMarkingMappingIdentifier.value = findInitialTireSizeMarkingMappingIdentifier()
    }
  },
  { immediate: true },
)

watch(
  () => [
    props.initialMappingIdentifier,
    props.initialEtrto,
    props.initialInch,
    props.initialFrench,
  ],
  () => {
    selectedTireSizeMarkingMappingIdentifier.value = findInitialTireSizeMarkingMappingIdentifier()
  },
)

const selectedTireSizeMarkingMappingRow = computed(() => (
  completeTireSizeMarkingMappingRows.value.find(
    row => row.mappingIdentifier === selectedTireSizeMarkingMappingIdentifier.value,
  ) ?? null
))

const tireSizeMarkingMappingFieldValueCounts = computed(() => {
  const counts: Record<TireSizeMarkingMappingOptionField, Map<string, number>> = {
    etrto: new Map(),
    inch: new Map(),
    french: new Map(),
  }

  for (const row of completeTireSizeMarkingMappingRows.value) {
    for (const field of Object.keys(counts) as TireSizeMarkingMappingOptionField[]) {
      const fieldValue = row[field]
      counts[field].set(fieldValue, (counts[field].get(fieldValue) ?? 0) + 1)
    }
  }

  return counts
})

const createTireSizeMarkingMappingCalculatorOptions = (
  field: TireSizeMarkingMappingOptionField,
): TireSizeMarkingMappingCalculatorOption[] => (
  completeTireSizeMarkingMappingRows.value.map((row) => {
    const fieldValue = row[field]
    const hasDuplicateFieldValue = (tireSizeMarkingMappingFieldValueCounts.value[field].get(fieldValue) ?? 0) > 1

    return {
      value: row.mappingIdentifier,
      label: hasDuplicateFieldValue ? `${fieldValue} · ${row.etrto}` : fieldValue,
    }
  })
)

const updateSelectedTireSizeMarkingIdentifier = (value: string | number | null) => {
  if (typeof value === 'string') {
    selectedTireSizeMarkingMappingIdentifier.value = value
  }
}

const etrtoTireSizeMarkingMappingOptions = computed(() => (
  createTireSizeMarkingMappingCalculatorOptions('etrto')
))

const inchTireSizeMarkingMappingOptions = computed(() => (
  createTireSizeMarkingMappingCalculatorOptions('inch')
))

const frenchTireSizeMarkingMappingOptions = computed(() => (
  createTireSizeMarkingMappingCalculatorOptions('french')
))
</script>

<style scoped>
.tire-size-marking-mapping-calculator {
  display: grid;
  gap: 0.9rem;
  text-align: left;
}

.tire-size-marking-mapping-calculator__intro {
  display: grid;
  gap: 0.3rem;
  max-width: 52rem;
}

.tire-size-marking-mapping-calculator__intro h3,
.tire-size-marking-mapping-calculator__intro p {
  margin: 0;
}

.tire-size-marking-mapping-calculator__intro p {
  font-size: 0.86rem;
  line-height: 1.55;
}

.tire-size-marking-mapping-calculator__controls {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.65rem;
}

.tire-size-marking-mapping-calculator__field {
  display: grid;
  min-width: 0;
  gap: 0.4rem;
}

.tire-size-marking-mapping-calculator__label,
.tire-size-marking-mapping-calculator__current-label {
  color: var(--tz-text-primary);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.tire-size-marking-mapping-calculator__current {
  display: grid;
  gap: 0.4rem;
  border: 1px solid color-mix(in srgb, var(--tz-site-accent) 28%, var(--tz-border-subtle));
  border-radius: 0.8rem;
  background: var(--tz-surface-muted);
  padding: 0.6rem 0.75rem;
}

.tire-size-marking-mapping-calculator__current-values {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 0.45rem;
  color: var(--tz-text-secondary);
  font-family: var(--tz-font-ui);
  font-size: 0.9rem;
}

.tire-size-marking-mapping-calculator__current-values strong {
  color: var(--tz-site-accent);
}

.tire-size-marking-mapping-calculator__note {
  margin: 0;
  font-size: 0.75rem;
  line-height: 1.5;
}

@media (max-width: 767px) {
  .tire-size-marking-mapping-calculator__controls {
    grid-template-columns: 1fr;
  }
}
</style>
