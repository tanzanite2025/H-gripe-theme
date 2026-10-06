<template>
  <section class="spoke-physical-corrections-step" aria-labelledby="spoke-physical-corrections-step-title">
    <SpokeStepNavigation
      :current-step="currentStep"
      :show-previous="true"
      @select="emit('select-step', $event)"
      @previous="emit('previous')"
      @next="emit('next')"
    />

    <div class="spoke-physical-corrections-step__intro">
      <span class="spoke-physical-corrections-step__eyebrow">04</span>
      <div>
        <h2 id="spoke-physical-corrections-step-title" class="spoke-physical-corrections-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepFourPrompt') }}
        </h2>
        <p class="spoke-physical-corrections-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepFourSubtitle') }}
        </p>
      </div>
    </div>

    <div class="spoke-physical-corrections-step__sections">
      <SpokeHoleEngagementStep
        embedded
        :front-hole-diameter="props.frontHoleDiameter"
        :rear-hole-diameter="props.rearHoleDiameter"
        @update:front-hole-diameter="emit('update:frontHoleDiameter', $event)"
        @update:rear-hole-diameter="emit('update:rearHoleDiameter', $event)"
      />

      <SpokeInterlacingStep
        embedded
        :front-interlacing="props.frontInterlacing"
        :rear-interlacing="props.rearInterlacing"
        :front-compensation="props.frontCompensation"
        :rear-compensation="props.rearCompensation"
        :spoke-elongation-compensation="props.spokeElongationCompensation"
        :front-crossing="props.frontCrossing"
        :rear-crossing="props.rearCrossing"
        @update:front-interlacing="emit('update:frontInterlacing', $event)"
        @update:rear-interlacing="emit('update:rearInterlacing', $event)"
        @update:front-compensation="emit('update:frontCompensation', $event)"
        @update:rear-compensation="emit('update:rearCompensation', $event)"
        @update:spoke-elongation-compensation="emit('update:spokeElongationCompensation', $event)"
      />
    </div>

    <SpokePhysicsDiagrams class="spoke-physical-corrections-step__reference" />
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeHoleEngagementStep from '~/components/SpokeHoleEngagementStep.vue'
import SpokeInterlacingStep from '~/components/SpokeInterlacingStep.vue'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepNavigation from '~/components/SpokeStepNavigation.vue'
import type { SpokeInterlacing } from '~/types/spokeCalculator'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontHoleDiameter?: number | null
  rearHoleDiameter?: number | null
  frontInterlacing?: SpokeInterlacing
  rearInterlacing?: SpokeInterlacing
  frontCompensation?: number | null
  rearCompensation?: number | null
  spokeElongationCompensation?: number | null
  frontCrossing?: number
  rearCrossing?: number
}>(), {
  currentStep: 4,
  frontHoleDiameter: 2.5,
  rearHoleDiameter: 2.5,
  frontInterlacing: 'off',
  rearInterlacing: 'off',
  frontCompensation: null,
  rearCompensation: null,
  spokeElongationCompensation: null,
  frontCrossing: 3,
  rearCrossing: 3,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontHoleDiameter': [value: number | null]
  'update:rearHoleDiameter': [value: number | null]
  'update:frontInterlacing': [value: SpokeInterlacing]
  'update:rearInterlacing': [value: SpokeInterlacing]
  'update:frontCompensation': [value: number | null]
  'update:rearCompensation': [value: number | null]
  'update:spokeElongationCompensation': [value: number | null]
  previous: []
  next: []
}>()
</script>

<style scoped>
.spoke-physical-corrections-step {
  --corrections-step-text: #0f172a;
  --corrections-step-muted: #475569;
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
  color: var(--corrections-step-text);
}

.spoke-physical-corrections-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-physical-corrections-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: #059669;
  color: #ffffff;
  font-family: var(--tz-font-ui);
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-physical-corrections-step__title {
  margin: 0;
  color: var(--corrections-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-physical-corrections-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--corrections-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-physical-corrections-step__sections {
  display: grid;
  gap: 18px;
}

.spoke-physical-corrections-step__reference {
  margin-top: 18px;
}

.spoke-physical-corrections-step__reference :deep(.schematic-nav),
.spoke-physical-corrections-step__reference :deep(.diagram-panel:not(#diagram-hole):not(#diagram-interlace):not(#diagram-stretch)) {
  display: none !important;
}

.spoke-physical-corrections-step__reference :deep(.diagram-panel#diagram-hole),
.spoke-physical-corrections-step__reference :deep(.diagram-panel#diagram-interlace),
.spoke-physical-corrections-step__reference :deep(.diagram-panel#diagram-stretch) {
  display: grid !important;
}

@media (max-width: 767px) {
  .spoke-physical-corrections-step {
    padding: 16px;
    border-radius: 20px;
  }
}
</style>
