<template>
  <section class="spoke-erd-step" aria-labelledby="spoke-erd-step-title">
    <SpokeStepProgress
      :current-step="currentStep"
      :available-step="2"
      @select="emit('select-step', $event)"
    />

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

    <div class="spoke-erd-step__input-panel">
      <div class="spoke-erd-step__input-copy">
        <span class="spoke-erd-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.erdInputKicker', 'STEP 2 INPUT') }}
        </span>
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

    <section class="spoke-erd-step__measurement" aria-labelledby="spoke-erd-measurement-title">
      <div class="spoke-erd-step__measurement-heading">
        <span class="spoke-erd-step__measurement-badge">01</span>
        <h3 id="spoke-erd-measurement-title" class="spoke-erd-step__measurement-title">
          {{ t('resourcesSpokeCalculator.parameter.workflow.stepOneTitle') }}
        </h3>
      </div>

      <div class="spoke-erd-step__measurement-grid">
        <div class="spoke-erd-step__measurement-copy">
          <p>
            {{ t('resourcesSpokeCalculator.parameter.workflow.stepOneFormula') }}
            <strong>{{ t('resourcesSpokeCalculator.parameter.workflow.stepOneFormulaValue') }}</strong>.
          </p>
          <ul>
            <li
              v-for="index in 4"
              :key="`erd-measurement-item-${index}`"
            >
              {{ t(`resourcesSpokeCalculator.parameter.workflow.stepOneItems.${index - 1}`) }}
            </li>
          </ul>
          <p class="spoke-erd-step__measurement-note">
            {{ t('resourcesSpokeCalculator.parameter.workflow.stepOneNote') }}
          </p>
        </div>

        <div class="spoke-erd-step__measurement-illustration">
          <GuideImage
            src="/public/technical/what-is-erd.webp"
            :alt="t('resourcesSpokeCalculator.parameter.workflow.stepOneAlt')"
            :zoomOnClick="true"
            :caption="t('resourcesSpokeCalculator.parameter.workflow.stepOneCaption')"
          />
        </div>
      </div>
    </section>

    <SpokePhysicsDiagrams class="spoke-erd-step__reference" />

    <div class="spoke-erd-step__actions">
      <button type="button" class="spoke-erd-step__previous" @click="emit('previous')">
        <span aria-hidden="true">←</span>
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'

const { t } = useI18n()
const props = withDefaults(defineProps<{
  currentStep?: number
  modelValue?: number | null
}>(), {
  currentStep: 2,
  modelValue: null,
})

const emit = defineEmits<{
  'update:modelValue': [value: number | null]
  'select-step': [step: number]
  previous: []
}>()

const erdMm = computed({
  get: () => props.modelValue,
  set: value => emit('update:modelValue', value),
})
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

.spoke-erd-step__actions {
  display: flex;
  justify-content: flex-start;
  margin-top: 14px;
}

.spoke-erd-step__previous {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 1px solid var(--erd-step-border);
  border-radius: 9999px;
  background: #ffffff;
  color: var(--erd-step-text);
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-erd-step__previous:hover {
  border-color: var(--erd-step-accent);
  color: var(--erd-step-accent);
}

.spoke-erd-step__previous:focus-visible {
  outline: 2px solid var(--erd-step-accent);
  outline-offset: 3px;
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

.spoke-erd-step__measurement {
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 20px;
  background: #ffffff;
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.06);
}

.spoke-erd-step__measurement-heading {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.spoke-erd-step__measurement-badge {
  display: inline-flex;
  width: 28px;
  height: 28px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-radius: 9999px;
  background: rgba(5, 150, 105, 0.1);
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
}

.spoke-erd-step__measurement-title {
  margin: 0;
  color: var(--erd-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-erd-step__measurement-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.85fr) minmax(320px, 1.15fr);
  gap: 24px;
  align-items: center;
}

.spoke-erd-step__measurement-copy {
  min-width: 0;
  color: var(--erd-step-muted);
  font-size: 13px;
  line-height: 1.6;
}

.spoke-erd-step__measurement-copy p {
  margin: 0;
}

.spoke-erd-step__measurement-copy strong {
  color: var(--erd-step-text);
  font-weight: 800;
}

.spoke-erd-step__measurement-copy ul {
  display: grid;
  gap: 6px;
  margin: 12px 0 0;
  padding-left: 1.2rem;
  list-style: disc;
}

.spoke-erd-step__measurement-note {
  margin-top: 12px !important;
  color: var(--erd-step-muted);
  font-size: 12px;
  font-style: italic;
}

.spoke-erd-step__measurement-illustration {
  min-width: 0;
  overflow: hidden;
  border: 1px solid rgba(15, 23, 42, 0.1);
  border-radius: 16px;
  background: #f8fafc;
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
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--erd-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-erd-step__input-copy {
  min-width: 0;
}

.spoke-erd-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-erd-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
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
  border: 2px solid rgba(5, 150, 105, 0.48);
  border-radius: 9999px;
  background: #ffffff;
  box-shadow: 0 3px 10px rgba(5, 150, 105, 0.08);
  overflow: hidden;
}

.spoke-erd-step__input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--erd-step-text);
  padding: 11px 12px 11px 16px;
  font-size: 15px;
  font-weight: 700;
}

.spoke-erd-step__input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--erd-step-accent);
}

.spoke-erd-step__unit {
  flex: 0 0 auto;
  padding: 0 14px 0 4px;
  color: #047857;
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
    margin-bottom: 18px;
    padding: 15px;
  }

  .spoke-erd-step__measurement {
    padding: 15px;
    border-radius: 18px;
  }

  .spoke-erd-step__measurement-grid {
    grid-template-columns: 1fr;
    gap: 16px;
  }
}
</style>
