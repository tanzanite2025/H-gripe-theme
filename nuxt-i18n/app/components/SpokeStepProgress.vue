<template>
  <ol
    class="spoke-step-progress"
    :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.stepProgressLabel', 'Calculator steps')"
  >
    <li
      v-for="step in stepCount"
      :key="step"
      class="spoke-step-progress__item"
      :class="{
        'spoke-step-progress__item--active': step === currentStep,
        'spoke-step-progress__item--complete': step < currentStep,
      }"
    >
      <button
        type="button"
        class="spoke-step-progress__button"
        :class="{ 'spoke-step-progress__button--available': step <= availableStep }"
        :aria-current="step === currentStep ? 'step' : undefined"
        :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.stepNumber', { step }, `Step ${step}`)"
        :disabled="step > availableStep"
        @click="emit('select', step)"
      >
        <span class="spoke-step-progress__dot">{{ step }}</span>
      </button>
    </li>
  </ol>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { SPOKE_WIZARD_STEP_COUNT } from '~/types/spokeCalculator'

withDefaults(defineProps<{
  currentStep: number
  availableStep?: number
  stepCount?: number
}>(), {
  availableStep: SPOKE_WIZARD_STEP_COUNT,
  stepCount: SPOKE_WIZARD_STEP_COUNT,
})

const emit = defineEmits<{
  select: [step: number]
}>()

const { t } = useI18n()
</script>

<style scoped>
.spoke-step-progress {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.spoke-step-progress__item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  border-radius: 9999px;
}

.spoke-step-progress__button {
  display: inline-flex;
  width: 1.45rem;
  height: 1.45rem;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 1px solid color-mix(in srgb, #0f172a 18%, transparent);
  border-radius: 9999px;
  background: color-mix(in srgb, #0f172a 8%, transparent);
  color: #475569;
  font-family: var(--tz-font-ui);
  font-size: 0.62rem;
  font-weight: 900;
  line-height: 1;
  cursor: default;
  transition: transform 0.2s ease, background-color 0.2s ease, box-shadow 0.2s ease;
}

.spoke-step-progress__button--available {
  cursor: pointer;
}

.spoke-step-progress__button--available:hover {
  transform: translateY(-1px);
}

.spoke-step-progress__button:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 2px;
}

.spoke-step-progress__item--active .spoke-step-progress__button {
  border-color: #059669;
  background: #059669;
  color: #ffffff;
  box-shadow: 0 0 0 4px rgba(5, 150, 105, 0.15);
}

.spoke-step-progress__item--complete .spoke-step-progress__button {
  border-color: #059669;
  background: rgba(5, 150, 105, 0.12);
  color: #047857;
}

.spoke-step-progress__button:disabled {
  opacity: 0.7;
}
</style>
