<template>
  <div
    class="spoke-calculator__build-settings"
    :class="{ 'spoke-calculator__build-settings--inline': props.layout === 'inline' }"
  >
    <div v-if="props.layout === 'section'" class="spoke-calculator__build-settings-header">
      <p class="spoke-calculator__build-settings-title">
        {{ t('resourcesSpokeCalculator.calculator.buildSettings.title') }}
      </p>
    </div>

    <div
      class="spoke-calculator__build-settings-grid"
      :class="{ 'spoke-calculator__build-settings-grid--inline': props.layout === 'inline' }"
    >
      <div
        class="spoke-calculator__setting-field"
        :class="{ 'spoke-calculator__setting-field--inline': props.layout === 'inline' }"
      >
        <label :for="fieldId('topology')" class="block text-xs font-medium tz-text-secondary">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.topology') }}
        </label>
        <SpokeCalculatorSelect
          :id="fieldId('topology')"
          :model-value="config.topologyId"
          :options="props.options.topologyOptions"
          :disabled="props.options.topologyOptions.length === 0"
          :aria-label="t('resourcesSpokeCalculator.calculator.buildSettings.topology')"
          @update:model-value="updateTopology"
        />
      </div>

      <div
        v-if="selectedTopology"
        class="spoke-calculator__topology-summary"
        :class="{ 'spoke-calculator__topology-summary--inline': props.layout === 'inline' }"
        role="status"
      >
        <span class="spoke-calculator__topology-summary-title">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.topologySummary') }}
        </span>
        <div class="spoke-calculator__topology-summary-grid">
          <span>{{ t('resourcesSpokeCalculator.calculator.buildSettings.totalHoles') }} <strong>{{ selectedTopology.holeCount }}</strong></span>
          <span>{{ t('resourcesSpokeCalculator.calculator.buildSettings.sideACount') }} <strong>{{ selectedTopology.sideACount }}</strong></span>
          <span>{{ t('resourcesSpokeCalculator.calculator.buildSettings.sideBCount') }} <strong>{{ selectedTopology.sideBCount }}</strong></span>
          <span>{{ t('resourcesSpokeCalculator.calculator.buildSettings.crossing') }} <strong>{{ selectedTopology.cross }}X</strong></span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import type {
  SpokeCalculatorManualOptions,
  SpokeWheelBuildConfig,
  SpokeWheelSide,
} from '~/types/spokeCalculator'
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  side: SpokeWheelSide
  config: SpokeWheelBuildConfig
  options: SpokeCalculatorManualOptions
  layout?: 'section' | 'inline'
}>(), {
  layout: 'section',
})

const { t } = useI18n()
const config = props.config
const fieldId = (name: string) => `${props.side}-${name}`

const selectedTopology = computed(() => props.options.topologySummaries.find(
  topology => topology.topologyId === config.topologyId,
))

const updateTopology = (value: string | number | null) => {
  if (typeof value !== 'string') return
  const topology = props.options.topologySummaries.find(item => item.topologyId === value)
  if (!topology) return
  config.topologyId = topology.topologyId
  config.spokeCount = topology.holeCount
  config.crossing = topology.cross
}
</script>

<style scoped>
.spoke-calculator__build-settings {
  display: grid;
  gap: 0.75rem;
  padding: 0.75rem 0 0;
  border-top: 1px solid var(--spoke-border, var(--tz-border-subtle));
}

.spoke-calculator__build-settings--inline,
.spoke-calculator__build-settings-grid--inline {
  display: contents;
}

.spoke-calculator__build-settings-header {
  display: flex;
  min-width: 0;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.75rem;
}

.spoke-calculator__build-settings-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.spoke-calculator__build-settings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
  min-width: 0;
}

.spoke-calculator__setting-field {
  display: grid;
  gap: 0.3rem;
  min-width: 0;
}

.spoke-calculator__setting-field--inline,
.spoke-calculator__topology-summary--inline {
  grid-column: 1 / -1;
}

.spoke-calculator__setting-field label {
  color: var(--tz-text-secondary);
}

.spoke-calculator__topology-summary {
  display: grid;
  gap: 0.45rem;
  align-content: start;
  min-width: 0;
  border: 1px solid var(--spoke-border, var(--tz-border-subtle));
  border-radius: 0.5rem;
  background: var(--spoke-result-surface, var(--tz-surface-subtle));
  padding: 0.65rem 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  line-height: 1.35;
}

.spoke-calculator__topology-summary-title {
  color: var(--tz-text-muted);
  font-size: 0.65rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.spoke-calculator__topology-summary-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem 0.75rem;
}

.spoke-calculator__topology-summary-grid strong {
  color: var(--tz-text-primary);
  font-weight: 800;
}

.spoke-calculator__control {
  display: block;
  width: 100%;
  min-width: 0;
  border: 1px solid var(--spoke-border) !important;
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface) !important;
  background-image: none !important;
  color: var(--tz-text-primary) !important;
  padding: 0.75rem 0.875rem;
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__control:focus,
.spoke-calculator__control:focus-visible {
  outline: none;
  border-color: var(--spoke-border-strong) !important;
  box-shadow: 0 0 0 1px var(--spoke-focus-ring) !important;
}

.spoke-calculator__control--with-unit {
  flex: 1 1 auto;
  border: 0 !important;
  border-radius: 0;
  background: transparent !important;
  box-shadow: none !important;
  padding-right: 0.75rem;
}

.spoke-calculator__unit-field {
  display: flex;
  width: 100%;
  align-items: stretch;
  gap: 0;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background-color: var(--spoke-control-surface);
  transition:
    border-color 0.18s ease,
    box-shadow 0.18s ease,
    background-color 0.18s ease;
}

.spoke-calculator__unit-field:focus-within {
  border-color: var(--spoke-border-strong);
  box-shadow: 0 0 0 1px var(--spoke-focus-ring);
}

.spoke-calculator__unit-field .spoke-calculator__control:focus,
.spoke-calculator__unit-field .spoke-calculator__control:focus-visible {
  box-shadow: none !important;
}

.spoke-calculator__unit {
  display: inline-flex;
  flex: 0 0 auto;
  min-width: 2.75rem;
  align-items: center;
  justify-content: center;
  border-left: 1px solid var(--spoke-border);
  padding: 0 0.75rem;
  color: var(--tz-text-muted);
  font-size: var(--tz-type-caption);
  line-height: 1;
  white-space: nowrap;
}

@media (max-width: 767px) {
  .spoke-calculator__build-settings-grid {
    grid-template-columns: 1fr;
  }

  .spoke-calculator__build-settings-header {
    flex-direction: column;
  }
}
</style>
