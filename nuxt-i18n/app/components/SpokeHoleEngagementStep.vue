<template>
  <section class="spoke-hole-step" aria-labelledby="spoke-hole-step-title">
    <SpokeStepProgress
      :current-step="currentStep"
      @select="emit('select-step', $event)"
    />

    <div class="spoke-hole-step__intro">
      <span class="spoke-hole-step__eyebrow">05</span>
      <div>
        <h2 id="spoke-hole-step-title" class="spoke-hole-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepFivePrompt', '输入法兰孔径') }}
        </h2>
        <p class="spoke-hole-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepFiveSubtitle', '确认法兰孔径，用于从理论孔心扣除孔半径；J 弯头和直拉式花鼓的长度测量都适用。') }}
        </p>
      </div>
    </div>

    <div class="spoke-hole-step__input-panel">
      <div class="spoke-hole-step__input-copy">
        <span class="spoke-hole-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.holeInputKicker', 'STEP 5 INPUT') }}
        </span>
        <strong class="spoke-hole-step__input-label">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.holeInputLabel', '法兰孔径') }}
        </strong>
        <p class="spoke-hole-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.holePairHelp', '请分别填写前轮和后轮花鼓的法兰孔径，常见 14G 辐条花鼓约为 2.4–2.6 mm。') }}
        </p>
        <div class="spoke-hole-step__input-note" role="note">
          <span class="spoke-hole-step__input-note-icon" aria-hidden="true">Ø</span>
          <p>
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.holeInputNote', '部分花鼓没有官方 CAD 图纸，无法直接确认准确数值；有些花鼓的沉头或卡位区域做了内凹处理，有些没有。J 弯头按实际孔径填写，直拉式按实际结构填写；如果有官方 CAD 且已包含这部分结构，请填写 0。') }}
          </p>
        </div>
      </div>

      <div class="spoke-hole-step__wheel-grid">
        <fieldset class="spoke-hole-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.frontWheel') }}</legend>
          <p class="spoke-hole-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.frontHoleNote', '前轮花鼓法兰孔') }}
          </p>
          <label for="spoke-hole-diameter-front" class="spoke-hole-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.spokeHoleDiameter', '法兰孔径（毫米）') }}</span>
            <span class="spoke-hole-step__unit-field">
              <input
                id="spoke-hole-diameter-front"
                :value="props.frontHoleDiameter ?? ''"
                type="number"
                min="0"
                max="10"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.holeDiameterPlaceholder', '例如 2.5')"
                @input="updateHoleDiameter('front', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>

        <fieldset class="spoke-hole-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.rearWheel') }}</legend>
          <p class="spoke-hole-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rearHoleNote', '后轮花鼓法兰孔') }}
          </p>
          <label for="spoke-hole-diameter-rear" class="spoke-hole-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.spokeHoleDiameter', '法兰孔径（毫米）') }}</span>
            <span class="spoke-hole-step__unit-field">
              <input
                id="spoke-hole-diameter-rear"
                :value="props.rearHoleDiameter ?? ''"
                type="number"
                min="0"
                max="10"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.holeDiameterPlaceholder', '例如 2.5')"
                @input="updateHoleDiameter('rear', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>
      </div>
    </div>

    <SpokePhysicsDiagrams class="spoke-hole-step__reference" />

    <div class="spoke-hole-step__actions">
      <button type="button" class="spoke-hole-step__previous" @click="emit('previous')">
        <span aria-hidden="true">←</span>
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}</span>
      </button>
      <button type="button" class="spoke-hole-step__next" @click="emit('next')">
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
  frontHoleDiameter?: number | null
  rearHoleDiameter?: number | null
}>(), {
  currentStep: 5,
  frontHoleDiameter: 2.5,
  rearHoleDiameter: 2.5,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontHoleDiameter': [value: number | null]
  'update:rearHoleDiameter': [value: number | null]
  previous: []
  next: []
}>()

const readNumber = (event: Event): number | null => {
  const rawValue = (event.target as HTMLInputElement).value.trim()
  if (!rawValue) return null
  const value = Number(rawValue)
  return Number.isFinite(value) ? value : null
}

const updateHoleDiameter = (side: 'front' | 'rear', event: Event) => {
  const value = readNumber(event)
  if (side === 'front') {
    emit('update:frontHoleDiameter', value)
  } else {
    emit('update:rearHoleDiameter', value)
  }
}

</script>

<style scoped>
.spoke-hole-step {
  --hole-step-border: rgba(15, 23, 42, 0.12);
  --hole-step-text: #0f172a;
  --hole-step-muted: #475569;
  --hole-step-accent: #059669;
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
  color: var(--hole-step-text);
}

.spoke-hole-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-hole-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--hole-step-accent);
  color: #ffffff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-hole-step__title {
  margin: 0;
  color: var(--hole-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-hole-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--hole-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-hole-step__input-panel {
  display: grid;
  grid-template-columns: minmax(210px, 0.55fr) minmax(0, 1.45fr);
  gap: 18px;
  align-items: start;
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--hole-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-hole-step__input-copy {
  min-width: 0;
}

.spoke-hole-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-hole-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-hole-step__input-help {
  margin: 4px 0 0;
  color: var(--hole-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-hole-step__input-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-top: 10px;
  padding: 9px 10px;
  border: 1px solid rgba(225, 29, 72, 0.24);
  border-radius: 12px;
  background: rgba(255, 241, 242, 0.82);
  color: #9f1239;
  font-size: 10px;
  line-height: 1.45;
}

.spoke-hole-step__input-note-icon {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgba(225, 29, 72, 0.14);
  color: #be123c;
  font-size: 0.75rem;
  font-weight: 900;
}

.spoke-hole-step__input-note p {
  margin: 0;
}

.spoke-hole-step__wheel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 0;
}

.spoke-hole-step__wheel-card {
  min-width: 0;
  margin: 0;
  padding: 12px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.78);
}

.spoke-hole-step__wheel-card legend {
  padding: 0 6px;
  color: #047857;
  font-size: 12px;
  font-weight: 900;
}

.spoke-hole-step__wheel-note {
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

.spoke-hole-step__field {
  display: grid;
  gap: 5px;
  min-width: 0;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.25;
}

.spoke-hole-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 2px solid rgba(5, 150, 105, 0.38);
  border-radius: 9999px;
  background: #ffffff;
}

.spoke-hole-step__unit-field input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--hole-step-text);
  padding: 8px 5px 8px 11px;
  font-size: 13px;
  font-weight: 700;
}

.spoke-hole-step__unit-field input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--hole-step-accent);
}

.spoke-hole-step__unit-field > span {
  flex: 0 0 auto;
  padding-right: 10px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 800;
}

.spoke-hole-step__reference {
  margin: 0 0 14px;
}

/* Reuse the exact hole-contact diagram, legend, and two-column arrangement
 * from the seven-tab engineering reference while showing only tab 03 here. */
.spoke-hole-step__reference :deep(.physics-collapse-toggle),
.spoke-hole-step__reference :deep(.schematic-nav),
.spoke-hole-step__reference :deep(.diagram-panel:not(#diagram-hole)) {
  display: none !important;
}

.spoke-hole-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-hole-step__reference :deep(.diagram-panel#diagram-hole) {
  display: grid !important;
}

.spoke-hole-step__actions {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  margin-top: 14px;
}

.spoke-hole-step__previous {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 1px solid var(--hole-step-border);
  border-radius: 9999px;
  background: #ffffff;
  color: var(--hole-step-text);
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-hole-step__previous:hover {
  border-color: var(--hole-step-accent);
  color: var(--hole-step-accent);
}

.spoke-hole-step__previous:focus-visible {
  outline: 2px solid var(--hole-step-accent);
  outline-offset: 3px;
}

.spoke-hole-step__next {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 0;
  border-radius: 9999px;
  background: var(--hole-step-accent);
  color: #ffffff;
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-hole-step__next:hover {
  background: #047857;
  box-shadow: 0 8px 18px rgba(5, 150, 105, 0.22);
}

.spoke-hole-step__next:focus-visible {
  outline: 2px solid var(--hole-step-accent);
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .spoke-hole-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-hole-step__input-panel {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 18px;
    padding: 15px;
  }

  .spoke-hole-step__wheel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
