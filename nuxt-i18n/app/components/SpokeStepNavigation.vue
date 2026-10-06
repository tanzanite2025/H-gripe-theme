<template>
  <div
    class="spoke-step-navigation"
    :class="{ 'spoke-step-navigation--has-trailing-slot': $slots.trailing }"
  >
    <button
      v-if="showPrevious"
      type="button"
      class="spoke-step-navigation__button spoke-step-navigation__button--previous"
      :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步')"
      @click="emit('previous')"
    >
      <span aria-hidden="true">←</span>
      <span class="spoke-step-navigation__label">
        {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}
      </span>
    </button>
    <span v-else class="spoke-step-navigation__spacer" aria-hidden="true" />

    <SpokeStepProgress
      :current-step="currentStep"
      @select="emit('select', $event)"
    />

    <div class="spoke-step-navigation__trailing">
      <slot name="trailing">
        <button
          v-if="showNext"
          type="button"
          class="spoke-step-navigation__button spoke-step-navigation__button--next"
          :aria-label="t('resourcesSpokeCalculator.calculator.physicalCorrections.stepNext', '下一步')"
          @click="emit('next')"
        >
          <span class="spoke-step-navigation__label">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepNext', '下一步') }}
          </span>
          <span aria-hidden="true">→</span>
        </button>
        <span v-else class="spoke-step-navigation__spacer" aria-hidden="true" />
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'

withDefaults(defineProps<{
  currentStep: number
  showPrevious?: boolean
  showNext?: boolean
}>(), {
  showPrevious: false,
  showNext: true,
})

const emit = defineEmits<{
  select: [step: number]
  previous: []
  next: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.spoke-step-navigation {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
  align-items: center;
  gap: 0.75rem;
  width: 100%;
  margin-bottom: 12px;
}

.spoke-step-navigation__button,
.spoke-step-navigation__spacer {
  min-width: 0;
}

.spoke-step-navigation__trailing {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
}

.spoke-step-navigation__button {
  display: inline-flex;
  min-height: 2.25rem;
  align-items: center;
  gap: 0.45rem;
  padding: 0.45rem 0.8rem;
  border: 1px solid rgba(5, 150, 105, 0.24);
  border-radius: 9999px;
  background: rgba(5, 150, 105, 0.08);
  color: #047857;
  font-family: var(--tz-font-ui);
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1;
  cursor: pointer;
  transition: transform 0.18s ease, border-color 0.18s ease, background-color 0.18s ease, box-shadow 0.18s ease;
}

.spoke-step-navigation__button--previous {
  justify-self: start;
}

.spoke-step-navigation__button--next {
  justify-self: end;
  background: #059669;
  border-color: #059669;
  color: #ffffff;
}

.spoke-step-navigation__button:hover {
  border-color: #059669;
  box-shadow: 0 6px 14px rgba(5, 150, 105, 0.16);
  transform: translateY(-1px);
}

.spoke-step-navigation__button--next:hover {
  background: #047857;
}

.spoke-step-navigation__button:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .spoke-step-navigation {
    grid-template-columns: 2rem minmax(0, 1fr) 2rem;
    gap: 0.3rem;
  }

  .spoke-step-navigation--has-trailing-slot {
    grid-template-columns: 2rem minmax(0, 1fr) minmax(0, auto);
  }

  .spoke-step-navigation__button {
    width: 2rem;
    height: 2rem;
    min-height: 2rem;
    justify-content: center;
    padding: 0;
    font-size: 1rem;
  }

  .spoke-step-navigation__label {
    display: none;
  }

  .spoke-step-navigation :deep(.spoke-step-progress) {
    gap: 0.2rem;
  }

  .spoke-step-navigation :deep(.spoke-step-progress__item) {
    width: 1.6rem;
    height: 1.6rem;
  }

  .spoke-step-navigation :deep(.spoke-step-progress__button) {
    width: 1.3rem;
    height: 1.3rem;
  }
}
</style>
