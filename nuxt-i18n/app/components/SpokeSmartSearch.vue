<template>
  <div class="spoke-smart-search w-full max-w-2xl mx-auto">
    <form class="spoke-smart-search__header relative mb-6" role="search" @submit.prevent="submitRecordedBuildSearch">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <svg class="h-5 w-5 tz-text-muted" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
        </svg>
      </div>
      <input
        v-model="query"
        type="search"
        enterkeyhint="search"
        :aria-label="t('resourcesSpokeCalculator.search.title')"
        :aria-busy="loading"
        :placeholder="t('resourcesSpokeCalculator.search.placeholder')"
        class="spoke-smart-search__input block w-full pl-10 pr-4 py-3"
      />
    </form>

    <div class="mb-4" role="status" aria-live="polite" aria-atomic="true">
      <p v-if="loading" class="tz-caption tz-text-muted">
        {{ t('resourcesSpokeCalculator.search.loading') }}
      </p>
      <div v-else-if="error" class="spoke-smart-search__empty p-4">
        <p class="tz-caption tz-text-secondary">{{ t('resourcesSpokeCalculator.search.loadError') }}</p>
        <button type="button" class="spoke-smart-search__retry mt-2" @click="loadResults">
          {{ t('resourcesSpokeCalculator.search.retry') }}
        </button>
      </div>
      <p v-else-if="query.trim().length === 1" class="tz-caption tz-text-muted">
        {{ t('resourcesSpokeCalculator.search.minimumQueryLength') }}
      </p>
      <p v-else-if="hasSearchCriteria && matchingConfigs.length" class="tz-caption tz-text-secondary">
        {{ t('resourcesSpokeCalculator.search.buildsFound', { count: matchingConfigs.length }) }}
      </p>
      <p v-else-if="hasSearchCriteria" class="spoke-smart-search__empty p-4 tz-caption tz-text-muted">
        {{ query.trim()
          ? t('resourcesSpokeCalculator.search.empty', { query: query.trim() })
          : t('resourcesSpokeCalculator.search.emptySelection') }}
      </p>
    </div>

    <TransitionGroup 
      name="list-results" 
      tag="div" 
      class="space-y-4"
    >
      <div 
        v-for="config in visibleRecordedBuilds"
        :key="config.id"
        class="spoke-smart-search__card group relative overflow-hidden transition-all duration-300"
      >
        <div class="relative p-5 grid gap-4 md:grid-cols-[1fr,auto]">
          <div>
            <h3 class="text-sm font-semibold tz-text-primary mb-1 transition-colors group-hover:text-[var(--tz-site-accent)]">
              {{ config.name }}
            </h3>
            <p v-if="config.description" class="tz-caption tz-text-secondary mb-3 line-clamp-1">
              {{ config.description }}
            </p>
            
            <div class="flex flex-wrap gap-2">
              <span class="spoke-smart-search__chip inline-flex items-center px-2 py-0.5 rounded tz-micro-label font-medium">
                {{ config.spokeCount }}H
              </span>
              <span class="spoke-smart-search__chip inline-flex items-center px-2 py-0.5 rounded tz-micro-label font-medium">
                {{ config.crossing }}X
              </span>
              <span class="spoke-smart-search__chip inline-flex items-center px-2 py-0.5 rounded tz-micro-label font-medium">
                {{ nippleTypeLabel(config.nippleType) }}
              </span>
            </div>
          </div>

          <div class="spoke-smart-search__result-stack min-w-[180px] border-t pt-4 md:border-t-0 md:border-l md:pl-6 md:pt-0">
            <div class="tz-compact-label tz-text-muted mb-2 text-center">
              {{ t('resourcesSpokeCalculator.search.verified') }}
            </div>
            <div class="grid grid-cols-2 gap-x-4 gap-y-2">
              <div v-for="cell in resultCells(config)" :key="cell.label" class="text-center">
                <div class="tz-compact-label tz-text-muted mb-0.5">{{ cell.label }}</div>
                <div class="text-lg font-mono font-bold text-[var(--tz-site-accent)]">
                  {{ cell.value }}<span class="tz-micro-label tz-text-muted ml-0.5">
                    {{ t('resourcesSpokeCalculator.search.unit') }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from '#imports'
import type { SpokeRecordedResult } from '~/data/spoke-calculator/database'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'
import { useSpokeCalculatorRecordedResults } from '~/composables/useSpokeCalculatorRecordedResults'
import type { SpokeWheelCatalogSelection } from '~/types/spokeCalculator'
import { filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections } from '~/utils/filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections'

const props = defineProps<{
  frontSelection?: SpokeWheelCatalogSelection | null
  rearSelection?: SpokeWheelCatalogSelection | null
}>()

const query = ref('')
const { t } = useI18n()
const { catalog, options: catalogOptions, source: catalogSource, loadCatalog } = useSpokeCalculatorCatalog()
const { presets, loading, error, source, loadResults } = useSpokeCalculatorRecordedResults()

onMounted(() => {
  if (catalogSource.value !== 'api') void loadCatalog()
})

interface LengthCell {
  label: string
  value: string
}

const nippleTypeLabels = computed(() => new Map(
  catalogOptions.value.nippleTypes.map(option => [option.value, option.label])
))

const activeCatalogSelections = computed(() => (
  [props.frontSelection, props.rearSelection]
    .filter((selection): selection is SpokeWheelCatalogSelection => Boolean(
      selection && (
        selection.rimBrandId
        || selection.rimModelId
        || selection.hubBrandId
        || selection.hubModelId
      ),
    ))
))

const hasSearchCriteria = computed(() => (
  query.value.trim().length >= 2 || (query.value.trim().length === 0 && activeCatalogSelections.value.length > 0)
))

const matchingConfigs = computed(() => filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
  presets.value,
  query.value,
  { front: props.frontSelection, rear: props.rearSelection },
  catalog.value,
))

const visibleRecordedBuilds = computed(() => loading.value || error.value ? [] : matchingConfigs.value)

async function submitRecordedBuildSearch(): Promise<void> {
  if (query.value.trim().length < 2) return
  if (source.value !== 'api') await loadResults()
}

function resultCells(preset: SpokeRecordedResult): LengthCell[] {
  return actualResultCells(preset)
}

function actualResultCells(preset: SpokeRecordedResult): LengthCell[] {
  const actual = preset.actualLengths
  if (!actual) return []

  return [
    { label: t('resourcesSpokeCalculator.search.frontLeft'), value: actual.frontLeft },
    { label: t('resourcesSpokeCalculator.search.frontRight'), value: actual.frontRight },
    { label: t('resourcesSpokeCalculator.search.rearLeft'), value: actual.rearLeft },
    { label: t('resourcesSpokeCalculator.search.rearRight'), value: actual.rearRight },
  ]
    .filter((cell): cell is { label: string; value: number } => cell.value != null)
    .map(cell => ({
      label: cell.label,
      value: formatLength(cell.value),
    }))
}

function formatLength(value: number) {
  return Number.isInteger(value) ? String(value) : value.toFixed(1)
}

function nippleTypeLabel(value: SpokeRecordedResult['nippleType']) {
  return t(
    `resourcesSpokeCalculator.calculator.options.nippleType.${value}`,
    nippleTypeLabels.value.get(value) || value,
  )
}
</script>

<style scoped>
.spoke-smart-search {
  color: var(--tz-text-primary);
}

.spoke-smart-search__input {
  border: 1px solid var(--tz-form-control-border) !important;
  border-radius: 0.5rem;
  background-color: var(--tz-form-control-surface) !important;
  background-image: none !important;
  color: var(--tz-text-primary) !important;
  box-shadow: 0 4px 16px -4px rgba(0, 0, 0, 0.4);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-smart-search__input::placeholder {
  color: var(--tz-text-muted) !important;
}

.spoke-smart-search__input:focus,
.spoke-smart-search__input:focus-visible {
  outline: none;
  border-color: var(--tz-form-control-focus-border) !important;
  box-shadow: 0 0 0 1px var(--tz-form-control-focus-ring) !important;
}

.spoke-smart-search__retry {
  min-height: 2.75rem;
  padding: 0.5rem 1rem;
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-form-control-surface);
}

.spoke-smart-search__card {
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  background: var(--tz-form-panel-surface);
  box-shadow: 0 8px 22px -14px rgba(20, 32, 43, 0.14);
}

.spoke-smart-search__card:hover {
  border-color: var(--tz-border-strong);
  background: var(--tz-surface-muted);
}

.spoke-smart-search__chip {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
}

.spoke-smart-search__result-stack {
  border-color: var(--tz-border-subtle);
}

.spoke-smart-search__empty {
  border: 1px dashed var(--tz-border-strong);
  border-radius: 0.5rem;
  background: var(--tz-surface-subtle);
}

.list-results-enter-active,
.list-results-leave-active {
  transition: all 0.3s ease;
}
.list-results-enter-from,
.list-results-leave-to {
  opacity: 0;
  transform: translateY(10px);
}
</style>
