<template>
  <div class="spoke-smart-search w-full max-w-2xl mx-auto">
    <div class="spoke-smart-search__header relative mb-6">
      <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
        <svg class="h-5 w-5 tz-text-muted" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clip-rule="evenodd" />
        </svg>
      </div>
      <input
        v-model="query"
        type="text"
        placeholder="Type a hub model (e.g. '350', '240', 'Mavic')..."
        class="spoke-smart-search__input block w-full pl-10 pr-4 py-3"
      />
      <div v-if="query.length > 1 && matchingConfigs.length > 0" class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
        <span class="spoke-smart-search__badge inline-flex items-center px-2 py-0.5 rounded text-xs font-medium">
          {{ matchingConfigs.length }} builds found
        </span>
      </div>
    </div>

    <TransitionGroup 
      name="list-results" 
      tag="div" 
      class="space-y-4"
    >
      <div 
        v-for="config in matchingConfigs" 
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
              Verified
            </div>
            <div class="grid grid-cols-2 gap-x-4 gap-y-2">
              <div v-for="cell in resultCells(config)" :key="cell.label" class="text-center">
                <div class="tz-compact-label tz-text-muted mb-0.5">{{ cell.label }}</div>
                <div class="text-lg font-mono font-bold text-[var(--tz-site-accent)]">
                  {{ cell.value }}<span class="tz-micro-label tz-text-muted ml-0.5">mm</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <div 
        v-if="query.length > 1 && matchingConfigs.length === 0"
        class="spoke-smart-search__empty text-center py-12 tz-text-muted"
      >
        <p class="text-sm">No verified build result found for "{{ query }}".</p>
      </div>

    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { WheelBuildPreset } from '~/data/spoke-calculator/database'
import { useSpokeCalculatorCatalog } from '~/composables/useSpokeCalculatorCatalog'

const query = ref('')
const { presets, options: catalogOptions } = useSpokeCalculatorCatalog()

interface LengthCell {
  label: string
  value: string
}

const nippleTypeLabels = computed(() => new Map(
  catalogOptions.value.nippleTypes.map(option => [option.value, option.label])
))

const matchingConfigs = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (q.length < 2) return [] // Minimum 2 chars to search

  return presets.value.filter(preset => {
    if (actualResultCells(preset).length === 0) return false
    const matchName = preset.name.toLowerCase().includes(q)
    const matchKeywords = preset.keywords.some(k => k.toLowerCase().includes(q))
    return matchName || matchKeywords
  })
})

function resultCells(preset: WheelBuildPreset): LengthCell[] {
  return actualResultCells(preset)
}

function actualResultCells(preset: WheelBuildPreset): LengthCell[] {
  const actual = preset.actualLengths
  if (!actual) return []

  return [
    { label: 'F Left', value: actual.frontLeft },
    { label: 'F Right', value: actual.frontRight },
    { label: 'R Left', value: actual.rearLeft },
    { label: 'R Right', value: actual.rearRight },
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

function nippleTypeLabel(value: WheelBuildPreset['nippleType']) {
  return nippleTypeLabels.value.get(value) || value
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

.spoke-smart-search__badge {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-secondary);
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
