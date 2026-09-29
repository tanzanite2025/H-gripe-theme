<template>
  <section class="spoke-pcd-step" aria-labelledby="spoke-pcd-step-title">
    <SpokeStepProgress
      :current-step="currentStep"
      :available-step="3"
      @select="emit('select-step', $event)"
    />

    <div class="spoke-pcd-step__intro">
      <span class="spoke-pcd-step__eyebrow">03</span>
      <div>
        <h2 id="spoke-pcd-step-title" class="spoke-pcd-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepThreePrompt', '输入花鼓 PCD 与 WL / WR') }}
        </h2>
        <p class="spoke-pcd-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepThreeSubtitle', '确认花鼓节圆直径，以及左右法兰到轮组中心线的距离。') }}
        </p>
      </div>
    </div>

    <!-- Reuse only the existing PCD / WL / WR engineering panel for this step. -->
    <SpokePhysicsDiagrams class="spoke-pcd-step__reference" />

    <div class="spoke-pcd-step__actions">
      <button type="button" class="spoke-pcd-step__previous" @click="emit('previous')">
        <span aria-hidden="true">←</span>
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'

const { t } = useI18n()

withDefaults(defineProps<{
  currentStep?: number
}>(), {
  currentStep: 3,
})

const emit = defineEmits<{
  'select-step': [step: number]
  previous: []
}>()
</script>

<style scoped>
.spoke-pcd-step {
  --pcd-step-border: rgba(15, 23, 42, 0.12);
  --pcd-step-text: #0f172a;
  --pcd-step-muted: #475569;
  --pcd-step-accent: #059669;
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
  color: var(--pcd-step-text);
}

.spoke-pcd-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-pcd-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--pcd-step-accent);
  color: #ffffff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-pcd-step__title {
  margin: 0;
  color: var(--pcd-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-pcd-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--pcd-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-pcd-step__reference {
  margin: 0 0 14px;
}

/* Keep the source component untouched while showing only its PCD/WL/WR panel. */
.spoke-pcd-step__reference :deep(.physics-collapse-toggle),
.spoke-pcd-step__reference :deep(.schematic-nav),
.spoke-pcd-step__reference :deep(.diagram-panel:not(#diagram-pcd)) {
  display: none !important;
}

.spoke-pcd-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-pcd-step__reference :deep(.diagram-panel#diagram-pcd) {
  display: grid !important;
}

.spoke-pcd-step__actions {
  display: flex;
  justify-content: flex-start;
  margin-top: 14px;
}

.spoke-pcd-step__previous {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 1px solid var(--pcd-step-border);
  border-radius: 9999px;
  background: #ffffff;
  color: var(--pcd-step-text);
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-pcd-step__previous:hover {
  border-color: var(--pcd-step-accent);
  color: var(--pcd-step-accent);
}

.spoke-pcd-step__previous:focus-visible {
  outline: 2px solid var(--pcd-step-accent);
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .spoke-pcd-step {
    padding: 16px;
    border-radius: 20px;
  }
}
</style>
