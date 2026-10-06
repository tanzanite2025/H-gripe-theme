<template>
  <section class="spoke-rim-geometry-step" aria-labelledby="spoke-rim-geometry-step-title">
    <SpokeStepNavigation
      :current-step="currentStep"
      :show-previous="true"
      @select="emit('select-step', $event)"
      @previous="emit('previous')"
      @next="emit('next')"
    />

    <div class="spoke-rim-geometry-step__intro">
      <span class="spoke-rim-geometry-step__eyebrow">02</span>
      <div>
        <h2 id="spoke-rim-geometry-step-title" class="spoke-rim-geometry-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepTwoPrompt') }}
        </h2>
        <p class="spoke-rim-geometry-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepTwoSubtitle') }}
        </p>
      </div>
    </div>

    <div class="spoke-rim-geometry-step__inputs">
      <section class="spoke-rim-geometry-step__input-group" aria-labelledby="spoke-rim-geometry-erd-heading">
        <h3 id="spoke-rim-geometry-erd-heading" class="spoke-rim-geometry-step__section-title">
          {{ t('resourcesSpokeCalculator.parameter.items.erd.title') }}
        </h3>
        <SpokeERDStep
          presentation="inputs"
          :front-erd="props.frontErd"
          :rear-erd="props.rearErd"
          @update:front-erd="emit('update:frontErd', $event)"
          @update:rear-erd="emit('update:rearErd', $event)"
        />
      </section>

      <section class="spoke-rim-geometry-step__input-group" aria-labelledby="spoke-rim-geometry-offset-heading">
        <h3 id="spoke-rim-geometry-offset-heading" class="spoke-rim-geometry-step__section-title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rimGeometryOffsetHeading') }}
        </h3>
        <SpokeAlternatingDrillingStep
          presentation="inputs"
          :front-offset="props.frontOffset"
          :rear-offset="props.rearOffset"
          :front-rim-offset="props.frontRimOffset"
          :rear-rim-offset="props.rearRimOffset"
          @update:front-offset="emit('update:frontOffset', $event)"
          @update:rear-offset="emit('update:rearOffset', $event)"
          @update:front-rim-offset="emit('update:frontRimOffset', $event)"
          @update:rear-rim-offset="emit('update:rearRimOffset', $event)"
        />
      </section>
    </div>

    <div class="spoke-rim-geometry-step__guidance">
      <SpokeERDStep presentation="guidance" />
      <SpokeAlternatingDrillingStep presentation="guidance" />
    </div>

    <SpokePhysicsDiagrams class="spoke-rim-geometry-step__reference" />
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokeAlternatingDrillingStep from '~/components/SpokeAlternatingDrillingStep.vue'
import SpokeERDStep from '~/components/SpokeERDStep.vue'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepNavigation from '~/components/SpokeStepNavigation.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontErd?: number | null
  rearErd?: number | null
  frontOffset?: number | null
  rearOffset?: number | null
  frontRimOffset?: number | null
  rearRimOffset?: number | null
}>(), {
  currentStep: 2,
  frontErd: null,
  rearErd: null,
  frontOffset: 0,
  rearOffset: 0,
  frontRimOffset: 0,
  rearRimOffset: 0,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontErd': [value: number | null]
  'update:rearErd': [value: number | null]
  'update:frontOffset': [value: number | null]
  'update:rearOffset': [value: number | null]
  'update:frontRimOffset': [value: number | null]
  'update:rearRimOffset': [value: number | null]
  previous: []
  next: []
}>()
</script>

<style scoped>
.spoke-rim-geometry-step {
  --rim-geometry-step-text: #0f172a;
  --rim-geometry-step-muted: #475569;
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
  color: var(--rim-geometry-step-text);
}

.spoke-rim-geometry-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 16px;
}

.spoke-rim-geometry-step__eyebrow {
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

.spoke-rim-geometry-step__title {
  margin: 0;
  color: var(--rim-geometry-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-rim-geometry-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--rim-geometry-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-rim-geometry-step__inputs {
  display: grid;
  gap: 18px;
}

.spoke-rim-geometry-step__input-group {
  display: grid;
  gap: 9px;
  min-width: 0;
  padding: 14px 16px;
  border: 1px solid rgba(5, 150, 105, 0.22);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.72);
}

.spoke-rim-geometry-step__section-title {
  margin: 0;
  color: #047857;
  font-size: 13px;
  font-weight: 900;
  line-height: 1.35;
}

.spoke-rim-geometry-step__guidance {
  display: grid;
  gap: 16px;
  margin-top: 20px;
}

.spoke-rim-geometry-step__reference {
  margin-top: 16px;
}

.spoke-rim-geometry-step__reference :deep(.physics-collapse-toggle),
.spoke-rim-geometry-step__reference :deep(.schematic-nav),
.spoke-rim-geometry-step__reference :deep(.diagram-panel:not(#diagram-erd):not(#diagram-drill)) {
  display: none !important;
}

.spoke-rim-geometry-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-rim-geometry-step__reference :deep(.diagram-panel#diagram-erd),
.spoke-rim-geometry-step__reference :deep(.diagram-panel#diagram-drill) {
  display: grid !important;
}

@media (max-width: 767px) {
  .spoke-rim-geometry-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-rim-geometry-step__input-group {
    gap: 7px;
    padding: 12px;
  }
}
</style>
