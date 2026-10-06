<template>
  <div class="spoke-calculator">
    <SpokeStepNavigation
      :current-step="currentStep"
      :show-previous="true"
      :show-next="false"
      @select="emit('select-step', $event)"
      @previous="emit('previous')"
    >
      <template #trailing>
        <div class="spoke-calculator__action">
          <button
            type="button"
            class="spoke-calculator__calculate-button"
            :disabled="loading"
            @click="onCalculate"
          >
            <span v-if="loading">{{ t('resourcesSpokeCalculator.calculator.action.calculating') }}</span>
            <span v-else>
              {{ frontResult || rearResult
                ? t('resourcesSpokeCalculator.calculator.action.recalculate')
                : t('resourcesSpokeCalculator.calculator.action.calculate') }}
            </span>
          </button>
          <p v-if="error" class="tz-caption text-rose-400">{{ error }}</p>
        </div>
      </template>
    </SpokeStepNavigation>

    <slot name="intro" />

    <div class="grid gap-6 items-start">
      <section class="spoke-calculator__shell">
        <SpokeCalculatorResults
          :front-result="frontResult"
          :rear-result="rearResult"
        />
      </section>
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorResults from '~/components/SpokeCalculatorResults.vue'
import SpokeStepNavigation from '~/components/SpokeStepNavigation.vue'
import { useSpokeCalculatorRun } from '~/composables/useSpokeCalculatorRun'
import type { SpokeWheelBuildConfig } from '~/types/spokeCalculator'
import { useI18n } from '#imports'

const { t } = useI18n()

const props = defineProps<{
  frontConfig: SpokeWheelBuildConfig
  rearConfig: SpokeWheelBuildConfig
  currentStep: number
}>()

const emit = defineEmits<{
  'select-step': [step: number]
  previous: []
}>()

const frontConfig = props.frontConfig
const rearConfig = props.rearConfig

const {
  loading,
  error,
  frontResult,
  rearResult,
  onCalculate,
} = useSpokeCalculatorRun(frontConfig, rearConfig)
</script>

<style scoped>
.spoke-calculator {
  --spoke-step-text: #0f172a;
  --spoke-step-accent: #059669;
  --spoke-result-surface: rgba(255, 255, 255, 0.78);
  --spoke-border: rgba(5, 150, 105, 0.2);
  width: 100%;
  margin-bottom: 1.5rem;
  padding: 20px;
  border: 1px dashed rgba(15, 23, 42, 0.16);
  border-radius: 24px;
  background:
    linear-gradient(rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(90deg, rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    #f8fafc;
  background-size: 28px 28px;
  color: var(--spoke-step-text);
}

.spoke-calculator__shell {
  min-width: 0;
}

.spoke-calculator__action {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 0.65rem;
}

.spoke-calculator__calculate-button {
  display: inline-flex;
  min-height: 2.25rem;
  max-width: 100%;
  align-items: center;
  justify-content: center;
  padding: 0.45rem 0.8rem;
  border: 1px solid var(--tz-action-primary);
  border-radius: 9999px;
  background: var(--tz-action-primary);
  color: #ffffff;
  font-family: var(--tz-font-ui);
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1;
  white-space: nowrap;
  cursor: pointer;
  transition: background-color 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.spoke-calculator__calculate-button:hover {
  border-color: var(--tz-action-primary-hover);
  background: var(--tz-action-primary-hover);
  box-shadow: 0 6px 14px rgba(5, 150, 105, 0.16);
}

.spoke-calculator__calculate-button:focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}

.spoke-calculator__calculate-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

@media (max-width: 767px) {
  .spoke-calculator {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-calculator__action {
    gap: 0.35rem;
  }

  .spoke-calculator__calculate-button {
    min-height: 2rem;
    padding: 0.35rem 0.6rem;
    font-size: 0.68rem;
  }
}

</style>
