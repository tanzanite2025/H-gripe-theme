<template>
  <section class="spoke-erd-step" aria-labelledby="spoke-erd-step-title">
    <div class="spoke-erd-step__intro">
      <span class="spoke-erd-step__eyebrow">02</span>
      <div>
        <h2 id="spoke-erd-step-title" class="spoke-erd-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepTwoPrompt', '输入轮圈 ERD') }}
        </h2>
        <p class="spoke-erd-step__subtitle">
          {{ t('resourcesSpokeCalculator.parameter.items.erd.desc') }}
        </p>
      </div>
    </div>

    <SpokePhysicsDiagrams class="spoke-erd-step__reference" />

    <div class="spoke-erd-step__input-panel">
      <div class="spoke-erd-step__input-copy">
        <label for="spoke-erd-step-input" class="spoke-erd-step__input-label">
          {{ t('resourcesSpokeCalculator.parameter.items.erd.title') }}
        </label>
        <p class="spoke-erd-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.erdInputHelp', 'Enter the measured ERD in millimetres.') }}
        </p>
      </div>

      <div class="spoke-erd-step__unit-field">
        <input
          id="spoke-erd-step-input"
          v-model.number="erdMm"
          class="spoke-erd-step__input"
          type="number"
          min="250"
          max="800"
          step="0.1"
          inputmode="decimal"
          :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.erdPlaceholder', '例如 598')"
        />
        <span class="spoke-erd-step__unit">mm</span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from '#imports'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'

const { t } = useI18n()
const erdMm = ref<number | null>(null)
</script>

<style scoped>
.spoke-erd-step {
  --erd-step-border: rgba(15, 23, 42, 0.12);
  --erd-step-text: #0f172a;
  --erd-step-muted: #475569;
  --erd-step-accent: #059669;
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
  color: var(--erd-step-text);
}

.spoke-erd-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-erd-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--erd-step-accent);
  color: #ffffff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-erd-step__title {
  margin: 0;
  color: var(--erd-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-erd-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--erd-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-erd-step__reference {
  margin: 0 0 14px;
}

/* Reuse the existing ERD panel exactly while leaving the seven-tab card below intact. */
.spoke-erd-step__reference :deep(.physics-collapse-toggle),
.spoke-erd-step__reference :deep(.schematic-nav),
.spoke-erd-step__reference :deep(.diagram-panel:not(#diagram-erd)) {
  display: none !important;
}

.spoke-erd-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-erd-step__reference :deep(.diagram-panel#diagram-erd) {
  display: grid !important;
}

.spoke-erd-step__input-panel {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(220px, 300px);
  gap: 18px;
  align-items: center;
  padding: 16px;
  border: 1px solid var(--erd-step-border);
  border-radius: 18px;
  background: #ffffff;
}

.spoke-erd-step__input-label {
  display: block;
  color: var(--erd-step-text);
  font-size: 14px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-erd-step__input-help {
  margin: 4px 0 0;
  color: var(--erd-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-erd-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  border: 1px solid var(--erd-step-border);
  border-radius: 9999px;
  background: #f8fafc;
  overflow: hidden;
}

.spoke-erd-step__input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--erd-step-text);
  padding: 10px 12px 10px 16px;
  font-size: 14px;
  font-weight: 700;
}

.spoke-erd-step__input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--erd-step-accent);
}

.spoke-erd-step__unit {
  flex: 0 0 auto;
  padding: 0 14px 0 4px;
  color: var(--erd-step-muted);
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 12px;
  font-weight: 800;
}

@media (max-width: 767px) {
  .spoke-erd-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-erd-step__input-panel {
    grid-template-columns: 1fr;
    gap: 10px;
  }
}
</style>
