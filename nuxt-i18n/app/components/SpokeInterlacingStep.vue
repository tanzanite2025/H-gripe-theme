<template>
  <section class="spoke-interlacing-step" aria-labelledby="spoke-interlacing-step-title">
    <SpokeStepProgress
      :current-step="currentStep"
      @select="emit('select-step', $event)"
    />

    <div class="spoke-interlacing-step__intro">
      <span class="spoke-interlacing-step__eyebrow">06</span>
      <div>
        <h2 id="spoke-interlacing-step-title" class="spoke-interlacing-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSixPrompt', '输入交叉压条折线补偿') }}
        </h2>
        <p class="spoke-interlacing-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSixSubtitle', '确认交叉编法是否在交叉点压条；如果压条，填写由微小折线带来的长度补偿。') }}
        </p>
      </div>
    </div>

    <div class="spoke-interlacing-step__input-panel">
      <div class="spoke-interlacing-step__input-copy">
        <span class="spoke-interlacing-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacingInputKicker', '第 6 步输入') }}
        </span>
        <strong class="spoke-interlacing-step__input-label">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacingInputLabel', '交叉压条折线') }}
        </strong>
        <p class="spoke-interlacing-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacingPairHelp', '请分别设置前轮和后轮是否在交叉点压条；开启后再填写对应的长度补偿。') }}
        </p>
        <div class="spoke-interlacing-step__input-note" role="note">
          <span class="spoke-interlacing-step__input-note-icon" aria-hidden="true">⌁</span>
          <p>
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacingInputNote', '2X / 3X 编法在最外侧交叉点做内压外或外压内时，辐条会形成微小折线；没有实际压条时请关闭，不要填入补偿。') }}
          </p>
        </div>
      </div>

      <div class="spoke-interlacing-step__wheel-grid">
        <fieldset class="spoke-interlacing-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.frontWheel') }}</legend>
          <p class="spoke-interlacing-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.frontInterlacingNote', '前轮交叉点') }}
          </p>
          <label for="spoke-interlacing-front" class="spoke-interlacing-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacing', '交叉处是否压条') }}</span>
            <SpokeCalculatorSelect
              id="spoke-interlacing-front"
              :model-value="props.frontInterlacing"
              :options="interlacingOptions"
              @update:model-value="updateInterlacing('front', $event)"
            />
          </label>
          <label
            v-if="props.frontInterlacing === 'on' && (props.frontCrossing ?? 0) > 0"
            for="spoke-interlace-compensation-front"
            class="spoke-interlacing-step__field"
          >
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensation', '压条长度补偿（毫米）') }}</span>
            <span class="spoke-interlacing-step__unit-field">
              <input
                id="spoke-interlace-compensation-front"
                :value="props.frontCompensation ?? ''"
                type="number"
                min="0"
                max="5"
                step="0.05"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensationPlaceholder', '例如 0.45')"
                @input="updateCompensation('front', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>

        <fieldset class="spoke-interlacing-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.rearWheel') }}</legend>
          <p class="spoke-interlacing-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rearInterlacingNote', '后轮交叉点') }}
          </p>
          <label for="spoke-interlacing-rear" class="spoke-interlacing-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlacing', '交叉处是否压条') }}</span>
            <SpokeCalculatorSelect
              id="spoke-interlacing-rear"
              :model-value="props.rearInterlacing"
              :options="interlacingOptions"
              @update:model-value="updateInterlacing('rear', $event)"
            />
          </label>
          <label
            v-if="props.rearInterlacing === 'on' && (props.rearCrossing ?? 0) > 0"
            for="spoke-interlace-compensation-rear"
            class="spoke-interlacing-step__field"
          >
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensation', '压条长度补偿（毫米）') }}</span>
            <span class="spoke-interlacing-step__unit-field">
              <input
                id="spoke-interlace-compensation-rear"
                :value="props.rearCompensation ?? ''"
                type="number"
                min="0"
                max="5"
                step="0.05"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensationPlaceholder', '例如 0.45')"
                @input="updateCompensation('rear', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>
      </div>
    </div>

    <SpokePhysicsDiagrams class="spoke-interlacing-step__reference" />

    <div class="spoke-interlacing-step__actions">
      <button type="button" class="spoke-interlacing-step__previous" @click="emit('previous')">
        <span aria-hidden="true">←</span>
        <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepPrevious', '上一步') }}</span>
      </button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'
import type { SpokeInterlacing } from '~/types/spokeCalculator'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontInterlacing?: SpokeInterlacing
  rearInterlacing?: SpokeInterlacing
  frontCompensation?: number | null
  rearCompensation?: number | null
  frontCrossing?: number
  rearCrossing?: number
}>(), {
  currentStep: 6,
  frontInterlacing: 'off',
  rearInterlacing: 'off',
  frontCompensation: 0.45,
  rearCompensation: 0.45,
  frontCrossing: 3,
  rearCrossing: 3,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontInterlacing': [value: SpokeInterlacing]
  'update:rearInterlacing': [value: SpokeInterlacing]
  'update:frontCompensation': [value: number | null]
  'update:rearCompensation': [value: number | null]
  previous: []
}>()

const interlacingOptions = computed(() => [
  {
    value: 'off',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.off', '否'),
  },
  {
    value: 'on',
    label: t('resourcesSpokeCalculator.calculator.physicalCorrections.on', '是'),
  },
])

const readNumber = (event: Event): number | null => {
  const rawValue = (event.target as HTMLInputElement).value.trim()
  if (!rawValue) return null
  const value = Number(rawValue)
  return Number.isFinite(value) ? value : null
}

const updateInterlacing = (side: 'front' | 'rear', value: string | number | null) => {
  if (value !== 'off' && value !== 'on') return
  if (side === 'front') {
    emit('update:frontInterlacing', value)
  } else {
    emit('update:rearInterlacing', value)
  }
}

const updateCompensation = (side: 'front' | 'rear', event: Event) => {
  const value = readNumber(event)
  if (side === 'front') {
    emit('update:frontCompensation', value)
  } else {
    emit('update:rearCompensation', value)
  }
}
</script>

<style scoped>
.spoke-interlacing-step {
  --interlacing-step-border: rgba(15, 23, 42, 0.12);
  --interlacing-step-text: #0f172a;
  --interlacing-step-muted: #475569;
  --interlacing-step-accent: #059669;
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
  color: var(--interlacing-step-text);
}

.spoke-interlacing-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-interlacing-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--interlacing-step-accent);
  color: #ffffff;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-interlacing-step__title {
  margin: 0;
  color: var(--interlacing-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-interlacing-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--interlacing-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-interlacing-step__input-panel {
  display: grid;
  grid-template-columns: minmax(210px, 0.55fr) minmax(0, 1.45fr);
  gap: 18px;
  align-items: start;
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--interlacing-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-interlacing-step__input-copy {
  min-width: 0;
}

.spoke-interlacing-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-interlacing-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-interlacing-step__input-help {
  margin: 4px 0 0;
  color: var(--interlacing-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-interlacing-step__input-note {
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

.spoke-interlacing-step__input-note-icon {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgba(217, 119, 6, 0.14);
  color: #b45309;
  font-size: 0.85rem;
  font-weight: 900;
}

.spoke-interlacing-step__input-note p {
  margin: 0;
}

.spoke-interlacing-step__wheel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 0;
}

.spoke-interlacing-step__wheel-card {
  min-width: 0;
  margin: 0;
  padding: 12px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.78);
}

.spoke-interlacing-step__wheel-card legend {
  padding: 0 6px;
  color: #047857;
  font-size: 12px;
  font-weight: 900;
}

.spoke-interlacing-step__wheel-note {
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

.spoke-interlacing-step__field {
  display: grid;
  gap: 5px;
  min-width: 0;
  margin-top: 10px;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.25;
}

.spoke-interlacing-step__field:first-of-type {
  margin-top: 0;
}

.spoke-interlacing-step__field :deep(.spoke-calculator-select__button) {
  min-height: 38px;
  border: 2px solid rgba(5, 150, 105, 0.38);
  padding: 8px 11px;
  color: var(--interlacing-step-text);
  font-size: 13px;
  font-weight: 700;
}

.spoke-interlacing-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 2px solid rgba(5, 150, 105, 0.38);
  border-radius: 9999px;
  background: #ffffff;
}

.spoke-interlacing-step__unit-field input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--interlacing-step-text);
  padding: 8px 5px 8px 11px;
  font-size: 13px;
  font-weight: 700;
}

.spoke-interlacing-step__unit-field input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--interlacing-step-accent);
}

.spoke-interlacing-step__unit-field > span {
  flex: 0 0 auto;
  padding-right: 10px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 800;
}

.spoke-interlacing-step__reference {
  margin: 0 0 14px;
}

/* Reuse the exact interlacing diagram, legend, and two-column arrangement
 * from the seven-tab engineering reference while showing only tab 07 here. */
.spoke-interlacing-step__reference :deep(.physics-collapse-toggle),
.spoke-interlacing-step__reference :deep(.schematic-nav),
.spoke-interlacing-step__reference :deep(.diagram-panel:not(#diagram-interlace)) {
  display: none !important;
}

.spoke-interlacing-step__reference :deep(.physics-collapse-content) {
  display: block !important;
}

.spoke-interlacing-step__reference :deep(.diagram-panel#diagram-interlace) {
  display: grid !important;
}

.spoke-interlacing-step__actions {
  display: flex;
  justify-content: flex-start;
  margin-top: 14px;
}

.spoke-interlacing-step__previous {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 36px;
  padding: 8px 14px;
  border: 1px solid var(--interlacing-step-border);
  border-radius: 9999px;
  background: #ffffff;
  color: var(--interlacing-step-text);
  font-size: 12px;
  font-weight: 800;
  cursor: pointer;
}

.spoke-interlacing-step__previous:hover {
  border-color: var(--interlacing-step-accent);
  color: var(--interlacing-step-accent);
}

.spoke-interlacing-step__previous:focus-visible {
  outline: 2px solid var(--interlacing-step-accent);
  outline-offset: 3px;
}

@media (max-width: 767px) {
  .spoke-interlacing-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-interlacing-step__input-panel {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 18px;
    padding: 15px;
  }

  .spoke-interlacing-step__wheel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
