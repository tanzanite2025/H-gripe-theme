<template>
  <div class="spoke-calculator">
    <div class="grid gap-6 items-start">
      <section class="spoke-calculator__shell">
        <!-- Action row -->
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <p class="tz-description tz-text-muted max-w-md">
            {{ t('resourcesSpokeCalculator.calculator.action.description') }}
          </p>
          <div class="flex items-center gap-3">
            <button
              type="button"
              class="inline-flex items-center rounded-full bg-[var(--tz-action-primary)] px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-[var(--tz-action-primary-hover)] focus:outline-none focus:ring-2 focus:ring-[color:var(--tz-site-accent)] focus:ring-offset-2 focus:ring-offset-[var(--tz-card-surface)] disabled:opacity-50 disabled:cursor-not-allowed"
              :disabled="loading"
              @click="onCalculate"
            >
              <span v-if="loading">{{ t('resourcesSpokeCalculator.calculator.action.calculating') }}</span>
              <span v-else>{{ t('resourcesSpokeCalculator.calculator.action.recalculate') }}</span>
            </button>
            <p v-if="error" class="tz-caption text-rose-400">{{ error }}</p>
          </div>
        </div>

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
import { useSpokeCalculatorRun } from '~/composables/useSpokeCalculatorRun'
import type { SpokeWheelBuildConfig } from '~/types/spokeCalculator'
import { useI18n } from '#imports'

const { t } = useI18n()

const props = defineProps<{
  frontConfig: SpokeWheelBuildConfig
  rearConfig: SpokeWheelBuildConfig
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
  --spoke-shell-surface: var(--tz-card-surface);
  --spoke-panel-surface: var(--tz-form-panel-surface);
  --spoke-control-surface: var(--tz-input-surface);
  --spoke-result-surface: var(--tz-surface-subtle);
  --spoke-border: var(--tz-border-subtle);
  --spoke-border-strong: var(--tz-border-strong);
  --spoke-focus-ring: var(--tz-form-control-focus-ring);
  color: var(--tz-text-primary);
}

.spoke-calculator__shell {
  border: 1px solid var(--spoke-border);
  border-radius: 0.5rem;
  background: var(--spoke-shell-surface);
  box-shadow: 0 10px 26px -14px rgba(20, 32, 43, 0.12);
}

.spoke-calculator__shell {
  padding: 1.25rem;
}

</style>
