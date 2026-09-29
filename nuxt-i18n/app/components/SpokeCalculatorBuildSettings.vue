<template>
  <div class="spoke-calculator__build-settings">
    <div class="spoke-calculator__build-settings-header">
      <p class="spoke-calculator__build-settings-title">
        {{ t('resourcesSpokeCalculator.calculator.buildSettings.title') }}
      </p>
    </div>

    <div class="spoke-calculator__build-settings-grid">
      <div class="spoke-calculator__setting-field">
        <label :for="fieldId('spoke-count')" class="block text-xs font-medium tz-text-secondary">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.spokeCount') }}
        </label>
        <SpokeCalculatorSelect
          :id="fieldId('spoke-count')"
          v-model="config.spokeCount"
          :options="options.spokeCountOptions"
        />
      </div>

      <div class="spoke-calculator__setting-field">
        <label :for="fieldId('lacing')" class="block text-xs font-medium tz-text-secondary">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.lacingPattern') }}
        </label>
        <SpokeCalculatorSelect
          :id="fieldId('lacing')"
          v-model="config.crossing"
          :options="options.lacingOptions"
        />
      </div>

      <div class="spoke-calculator__setting-field">
        <label :for="fieldId('nipple')" class="block text-xs font-medium tz-text-secondary">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.nippleType') }}
        </label>
        <SpokeCalculatorSelect
          :id="fieldId('nipple')"
          v-model="config.nippleType"
          :options="options.nippleTypeOptions"
        />
      </div>

      <div v-if="config.nippleType === 'hidden'" class="spoke-calculator__setting-field">
        <label :for="fieldId('nipple-length')" class="block text-xs font-medium tz-text-secondary">
          {{ t('resourcesSpokeCalculator.calculator.buildSettings.nippleLength') }}
        </label>
        <div class="spoke-calculator__unit-field">
          <input
            :id="fieldId('nipple-length')"
            v-model.number="config.nippleLength"
            type="number"
            min="0"
            max="30"
            :placeholder="t('resourcesSpokeCalculator.calculator.buildSettings.nippleLengthPlaceholder')"
            class="spoke-calculator__control spoke-calculator__control--with-unit"
          />
          <span class="spoke-calculator__unit">{{ t('resourcesSpokeCalculator.calculator.results.unit') }}</span>
        </div>
      </div>

    </div>

    <details open class="spoke-calculator__physical-settings">
      <summary>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.title') }}</summary>
      <div class="spoke-calculator__physical-settings-grid">
        <div class="spoke-calculator__setting-field">
          <label :for="fieldId('spoke-head-type')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.spokeHeadType') }}</label>
          <SpokeCalculatorSelect
            :id="fieldId('spoke-head-type')"
            v-model="config.spokeHeadType"
            :options="options.spokeHeadTypeOptions"
          />
        </div>
        <div v-if="config.spokeHeadType === 'straight_pull'" class="spoke-calculator__setting-field">
          <label :for="fieldId('straight-pull-offset')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPullOffset') }}</label>
          <input
            :id="fieldId('straight-pull-offset')"
            v-model.number="config.straightPullTangentOffsetMm"
            class="spoke-calculator__physical-number"
            type="number"
            min="-20"
            max="20"
            step="0.1"
          />
        </div>
        <div class="spoke-calculator__setting-field">
          <label :for="fieldId('spoke-profile')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.spokeProfile') }}</label>
          <SpokeCalculatorSelect
            :id="fieldId('spoke-profile')"
            v-model="config.spokeProfile"
            :options="options.spokeProfileOptions"
          />
        </div>
        <div class="spoke-calculator__setting-field">
          <label :for="fieldId('target-tension')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.targetTension') }}</label>
          <input
            :id="fieldId('target-tension')"
            v-model.number="config.targetTensionN"
            class="spoke-calculator__physical-number"
            type="number"
            min="0"
            max="3000"
            step="50"
          />
        </div>
        <div class="spoke-calculator__setting-field">
          <label :for="fieldId('alternating-offset')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingOffset') }}</label>
          <input
            :id="fieldId('alternating-offset')"
            v-model.number="config.alternatingDrillingOffsetMm"
            class="spoke-calculator__physical-number"
            type="number"
            min="-5"
            max="5"
            step="0.1"
          />
        </div>
        <div class="spoke-calculator__setting-field">
          <label :for="fieldId('interlacing')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacing') }}</label>
          <SpokeCalculatorSelect
            :id="fieldId('interlacing')"
            v-model="config.interlacing"
            :options="options.interlacingOptions"
          />
        </div>
        <div v-if="config.interlacing === 'on' && config.crossing > 0" class="spoke-calculator__setting-field">
          <label :for="fieldId('interlace-compensation')">{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensation') }}</label>
          <input
            :id="fieldId('interlace-compensation')"
            v-model.number="config.interlaceCompensationMm"
            class="spoke-calculator__physical-number"
            type="number"
            min="0"
            max="5"
            step="0.05"
          />
        </div>
      </div>
    </details>
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

const props = defineProps<{
  side: SpokeWheelSide
  config: SpokeWheelBuildConfig
  options: SpokeCalculatorManualOptions
}>()

const { t } = useI18n()
const config = props.config
const options = props.options
const fieldId = (name: string) => `${props.side}-${name}`
</script>

<style scoped>
.spoke-calculator__build-settings {
  display: grid;
  gap: 0.75rem;
  padding: 0.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 0.625rem;
  background: rgba(255, 255, 255, 0.72);
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

.spoke-calculator__setting-field label {
  color: var(--tz-text-secondary);
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

.spoke-calculator__physical-settings {
  border-top: 1px solid var(--spoke-border);
  padding-top: 0.65rem;
}

.spoke-calculator__physical-settings > summary {
  color: var(--tz-text-secondary);
  cursor: pointer;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.35;
}

.spoke-calculator__physical-settings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.65rem 0.75rem;
  margin-top: 0.75rem;
}

.spoke-calculator__physical-settings-grid label {
  font-size: 0.7rem;
  line-height: 1.3;
}

.spoke-calculator__physical-number {
  width: 100%;
  min-width: 0;
  min-height: 2.75rem;
  border: 1px solid var(--spoke-border);
  border-radius: 9999px;
  background: var(--spoke-control-surface);
  color: var(--tz-text-primary);
  padding: 0.6rem 0.85rem;
  font-size: 0.875rem;
  line-height: 1.25rem;
}

.spoke-calculator__physical-number:focus-visible {
  outline: none;
  border-color: var(--spoke-border-strong);
  box-shadow: 0 0 0 1px var(--spoke-focus-ring);
}

@media (max-width: 767px) {
  .spoke-calculator__build-settings-grid,
  .spoke-calculator__physical-settings-grid {
    grid-template-columns: 1fr;
  }

  .spoke-calculator__build-settings-header {
    flex-direction: column;
  }
}
</style>
