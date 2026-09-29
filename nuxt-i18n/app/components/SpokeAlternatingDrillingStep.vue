<template>
  <section class="spoke-alternating-step" aria-labelledby="spoke-alternating-step-title">
    <SpokeStepProgress
      :current-step="currentStep"
      @select="emit('select-step', $event)"
    />

    <div class="spoke-alternating-step__intro">
      <span class="spoke-alternating-step__eyebrow">03</span>
      <div>
        <h2 id="spoke-alternating-step-title" class="spoke-alternating-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepThreePrompt', '输入轮圈偏心量与交错钻孔偏移') }}
        </h2>
        <p class="spoke-alternating-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepThreeSubtitle', '继续录入轮圈整体偏心量和左右交错钻孔相对于轮圈中心基准的偏移量。') }}
        </p>
      </div>
    </div>

    <div class="spoke-alternating-step__input-panel">
      <div class="spoke-alternating-step__input-copy">
        <span class="spoke-alternating-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingInputKicker', 'STEP 3 INPUT') }}
        </span>
        <strong class="spoke-alternating-step__input-label">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingInputLabel', '轮圈交错钻孔偏移') }}
        </strong>
        <p class="spoke-alternating-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingPairHelp', '请分别填写前轮和后轮的交错钻孔偏移；没有交错钻孔时填 0。') }}
        </p>
        <div class="spoke-alternating-step__input-note" role="note">
          <span class="spoke-alternating-step__input-note-icon" aria-hidden="true">±</span>
          <p>
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingInputNote', '正值表示孔位向左侧偏移，负值表示向右侧偏移；优先以轮圈厂商数据为准。') }}
          </p>
        </div>
        <div class="spoke-alternating-step__input-note spoke-alternating-step__input-note--rim" role="note">
          <span class="spoke-alternating-step__input-note-icon" aria-hidden="true">↔</span>
          <p>
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetInputNote', '偏心圈的轮圈床整体偏离轮组中心线；对称轮圈填 0。偏心量与交错钻孔偏移是两项独立参数。') }}
          </p>
        </div>
      </div>

      <div class="spoke-alternating-step__wheel-grid">
        <fieldset class="spoke-alternating-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.frontWheel') }}</legend>
          <p class="spoke-alternating-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.frontAlternatingNote', '前轮轮圈交错孔偏移') }}
          </p>
          <label for="spoke-alternating-front-offset" class="spoke-alternating-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingOffsetFieldLabel', '交错钻孔偏移（正值偏左）') }}</span>
            <span class="spoke-alternating-step__unit-field">
              <input
                id="spoke-alternating-front-offset"
                :value="props.frontOffset ?? ''"
                type="number"
                min="-5"
                max="5"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingOffsetPlaceholder', '例如 0.75')"
                @input="updateOffset('front', $event)"
              />
              <span>mm</span>
            </span>
          </label>
          <label for="spoke-rim-offset-front" class="spoke-alternating-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetFieldLabel', '偏心量（正值向右）') }}</span>
            <span class="spoke-alternating-step__unit-field">
              <input
                id="spoke-rim-offset-front"
                :value="props.frontRimOffset ?? ''"
                type="number"
                min="-20"
                max="20"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetPlaceholder', '例如 2.5')"
                @input="updateRimOffset('front', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>

        <fieldset class="spoke-alternating-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.rearWheel') }}</legend>
          <p class="spoke-alternating-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rearAlternatingNote', '后轮轮圈交错孔偏移') }}
          </p>
          <label for="spoke-alternating-rear-offset" class="spoke-alternating-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingOffsetFieldLabel', '交错钻孔偏移（正值偏左）') }}</span>
            <span class="spoke-alternating-step__unit-field">
              <input
                id="spoke-alternating-rear-offset"
                :value="props.rearOffset ?? ''"
                type="number"
                min="-5"
                max="5"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.alternatingOffsetPlaceholder', '例如 0.75')"
                @input="updateOffset('rear', $event)"
              />
              <span>mm</span>
            </span>
          </label>
          <label for="spoke-rim-offset-rear" class="spoke-alternating-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetFieldLabel', '偏心量（正值向右）') }}</span>
            <span class="spoke-alternating-step__unit-field">
              <input
                id="spoke-rim-offset-rear"
                :value="props.rearRimOffset ?? ''"
                type="number"
                min="-20"
                max="20"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetPlaceholder', '例如 2.5')"
                @input="updateRimOffset('rear', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>
      </div>
    </div>

    <SpokePhysicsDiagrams class="spoke-alternating-step__reference" />

    <div class="spoke-alternating-step__actions">
      <button type="button" class="spoke-alternating-step__previous" @click="emit('previous')">
        <span aria-hidden="true">←</span>
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}</span>
      </button>
      <button type="button" class="spoke-alternating-step__next" @click="emit('next')">
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepNext', '下一步') }}</span>
        <span aria-hidden="true">→</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontOffset?: number | null
  rearOffset?: number | null
  frontRimOffset?: number | null
  rearRimOffset?: number | null
}>(), {
  currentStep: 3,
  frontOffset: 0,
  rearOffset: 0,
  frontRimOffset: 0,
  rearRimOffset: 0,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontOffset': [value: number | null]
  'update:rearOffset': [value: number | null]
  'update:frontRimOffset': [value: number | null]
  'update:rearRimOffset': [value: number | null]
  previous: []
  next: []
}>()

const readNumber = (event: Event): number | null => {
  const rawValue = (event.target as HTMLInputElement).value.trim()
  if (!rawValue) return null
  const value = Number(rawValue)
  return Number.isFinite(value) ? value : null
}

const updateOffset = (side: 'front' | 'rear', event: Event) => {
  const value = readNumber(event)
  if (side === 'front') {
    emit('update:frontOffset', value)
  } else {
    emit('update:rearOffset', value)
  }
}

const updateRimOffset = (side: 'front' | 'rear', event: Event) => {
  const value = readNumber(event)
  if (side === 'front') {
    emit('update:frontRimOffset', value)
  } else {
    emit('update:rearRimOffset', value)
  }
}
</script>

<style scoped>
.spoke-alternating-step {
  --alternating-step-border: rgba(15, 23, 42, 0.12);
  --alternating-step-text: #0f172a;
  --alternating-step-muted: #475569;
  --alternating-step-accent: #059669;
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
  color: var(--alternating-step-text);
}

.spoke-alternating-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-alternating-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--alternating-step-accent);
  color: #ffffff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-alternating-step__title {
  margin: 0;
  color: var(--alternating-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-alternating-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--alternating-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-alternating-step__input-panel {
  display: grid;
  grid-template-columns: minmax(210px, 0.55fr) minmax(0, 1.45fr);
  gap: 18px;
  align-items: start;
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--alternating-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-alternating-step__input-copy {
  min-width: 0;
}

.spoke-alternating-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-alternating-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-alternating-step__input-help {
  margin: 4px 0 0;
  color: var(--alternating-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-alternating-step__input-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 10px;
  padding: 9px 10px;
  border: 1px solid rgba(217, 119, 6, 0.24);
  border-radius: 12px;
  background: rgba(255, 251, 235, 0.82);
  color: #92400e;
  font-size: 10px;
  line-height: 1.45;
}

.spoke-alternating-step__input-note-icon {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgba(217, 119, 6, 0.14);
  color: #b45309;
  font-size: 0.8rem;
  font-weight: 900;
}

.spoke-alternating-step__input-note p {
  margin: 0;
}

.spoke-alternating-step__input-note--rim {
  border-color: rgba(5, 150, 105, 0.24);
  background: rgba(236, 253, 245, 0.82);
  color: #065f46;
}

.spoke-alternating-step__input-note--rim .spoke-alternating-step__input-note-icon {
  background: rgba(5, 150, 105, 0.14);
  color: #047857;
}

.spoke-alternating-step__wheel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 0;
}

.spoke-alternating-step__wheel-card {
  min-width: 0;
  margin: 0;
  padding: 12px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.78);
}

.spoke-alternating-step__wheel-card legend {
  padding: 0 6px;
  color: #047857;
  font-size: 12px;
  font-weight: 900;
}

.spoke-alternating-step__wheel-note {
  margin: 0 0 10px;
  padding: 6px 8px;
  border-radius: 9999px;
  background: rgba(5, 150, 105, 0.08);
  color: #047857;
  font-size: 10px;
  font-weight: 800;
  line-height: 1.35;
  text-align: center;
}

.spoke-alternating-step__field {
  display: grid;
  gap: 5px;
  min-width: 0;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.25;
}

.spoke-alternating-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 2px solid rgba(5, 150, 105, 0.38);
  border-radius: 9999px;
  background: #ffffff;
}

.spoke-alternating-step__unit-field input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--alternating-step-text);
  padding: 8px 5px 8px 11px;
  font-size: 13px;
  font-weight: 700;
}

.spoke-alternating-step__unit-field input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--alternating-step-accent);
}

.spoke-alternating-step__unit-field > span {
  flex: 0 0 auto;
  padding-right: 10px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 800;
}

.spoke-alternating-step__reference {
  margin: 0 0 14px;
}

/* Reuse the exact drill diagram, legend, and two-column arrangement from the
 * seven-tab engineering reference while showing only tab 06 in this step. */
.spoke-alternating-step__reference :deep(.physics-collapse-toggle),
.spoke-alternating-step__reference :deep(.schematic-nav),
.spoke-alternating-step__reference :deep(.diagram-panel:not(#diagram-drill)) {
  display: none !important;
}

.spoke-alternating-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-alternating-step__reference :deep(.diagram-panel#diagram-drill) {
  display: grid !important;
}

.spoke-alternating-step__actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 14px;
}

.spoke-alternating-step__previous {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 1px solid var(--alternating-step-border);
  border-radius: 9999px;
  background: #ffffff;
  color: var(--alternating-step-text);
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-alternating-step__previous:hover {
  border-color: var(--alternating-step-accent);
  color: var(--alternating-step-accent);
}

.spoke-alternating-step__previous:focus-visible {
  outline: 2px solid var(--alternating-step-accent);
  outline-offset: 3px;
}

.spoke-alternating-step__next {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 0;
  border-radius: 9999px;
  background: var(--alternating-step-accent);
  color: #ffffff;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-alternating-step__next:hover {
  background: #047857;
  box-shadow: 0 8px 18px rgba(5, 150, 105, 0.22);
}

.spoke-alternating-step__next:focus-visible {
  outline: 2px solid var(--alternating-step-accent);
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .spoke-alternating-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-alternating-step__input-panel {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 18px;
    padding: 15px;
  }

  .spoke-alternating-step__wheel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
