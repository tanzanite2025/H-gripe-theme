<template>
  <div class="spoke-calculator__panel space-y-4">
    <h3 class="text-sm font-semibold text-[var(--tz-site-accent)] uppercase tracking-wide">
      {{ side === 'front'
        ? t('resourcesSpokeCalculator.calculator.frontWheel')
        : t('resourcesSpokeCalculator.calculator.rearWheel') }}
    </h3>

    <div class="spoke-calculator__blueprint-sheet">
      <div class="spoke-calculator__blueprint-grid">
        <div class="spoke-calculator__legend-table">
          <div class="spoke-calculator__legend-row">
            <div class="spoke-calculator__legend-badge">4</div>
            <div class="spoke-calculator__legend-heading">
              <strong>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.label') }}</strong>
              <span>{{ t('resourcesSpokeCalculator.calculator.schematic.rimOffset.description') }}</span>
            </div>
            <div class="spoke-calculator__legend-control">
              <div class="spoke-calculator__unit-field">
                <input
                  :id="`${side}-blueprint-rim-offset`"
                  v-model.number="config.rimOffsetMm"
                  type="number"
                  min="-20"
                  max="20"
                  step="0.1"
                  placeholder="0"
                  class="spoke-calculator__control spoke-calculator__control--with-unit"
                />
                <span class="spoke-calculator__unit">{{ t('resourcesSpokeCalculator.calculator.results.unit') }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <SpokeCalculatorBuildSettings
      :side="side"
      :config="config"
      :options="options"
    />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeCalculatorBuildSettings from '~/components/SpokeCalculatorBuildSettings.vue'
import type {
  SpokeCalculatorManualOptions,
  SpokeWheelBuildConfig,
  SpokeWheelSide,
} from '~/types/spokeCalculator'

const props = defineProps<{
  side: SpokeWheelSide
  config: SpokeWheelBuildConfig
  options: SpokeCalculatorManualOptions
}>()

const { t } = useI18n()
const side = props.side
const config = props.config
const options = props.options
</script>

<style scoped>
.spoke-calculator__panel {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
  padding: 1rem;
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

.spoke-calculator__control:disabled {
  cursor: not-allowed;
  opacity: 0.55;
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

.spoke-calculator__blueprint-sheet {
  display: grid;
  gap: 0.75rem;
  padding: 0.875rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.88), rgba(248, 250, 252, 0.96)),
    var(--spoke-shell-surface);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.85),
    0 12px 28px rgba(20, 32, 43, 0.08);
}

.spoke-calculator__blueprint-grid {
  display: grid;
  gap: 0.85rem;
}

.spoke-calculator__legend-table {
  display: grid;
  gap: 0.55rem;
  padding: 0.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background: rgba(255, 255, 255, 0.7);
}

.spoke-calculator__legend-row {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr) minmax(0, 1.55fr);
  gap: 0.75rem;
  align-items: start;
  padding: 0.65rem 0.7rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-panel-surface);
}

.spoke-calculator__legend-badge {
  display: inline-flex;
  width: 1.8rem;
  height: 1.8rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.28);
  border-radius: 50%;
  background: rgba(5, 150, 105, 0.08);
  color: var(--tz-text-accent);
  font-size: 0.8rem;
  font-weight: 800;
  line-height: 1;
}

.spoke-calculator__legend-heading {
  display: grid;
  min-width: 0;
  gap: 0.08rem;
}

.spoke-calculator__legend-heading strong {
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 750;
  line-height: 1.2;
}

.spoke-calculator__legend-heading span {
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.35;
}

.spoke-calculator__legend-control {
  min-width: 0;
}

@media (max-width: 767px) {
  .spoke-calculator__panel {
    padding: 0.5rem 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__panel + .spoke-calculator__panel {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--spoke-border);
  }

  .spoke-calculator__blueprint-sheet {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
  }

  .spoke-calculator__legend-table {
    padding: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    gap: 0.45rem;
  }

  .spoke-calculator__legend-row {
    grid-template-columns: 1fr;
    padding: 0.4rem;
    gap: 0.4rem;
  }
}
</style>
