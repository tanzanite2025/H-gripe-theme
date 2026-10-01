<template>
  <section class="spoke-nipple-step" aria-labelledby="spoke-nipple-step-title">
    <SpokeStepNavigation
      :current-step="currentStep"
      :show-previous="true"
      :show-next="false"
      @select="emit('select-step', $event)"
      @previous="emit('previous')"
    />

    <div class="spoke-nipple-step__intro">
      <span class="spoke-nipple-step__eyebrow">07</span>
      <div>
        <h2 id="spoke-nipple-step-title" class="spoke-nipple-step__title">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSevenPrompt', '选择外置或内置辐条帽') }}
        </h2>
        <p class="spoke-nipple-step__subtitle">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSevenSubtitle', '外置辐条帽作为长度基准；本计算器按 SAPIM POLYAX 14 mm 外置辐条帽建立基准。选择内置式后，请输入实际辐条帽长度。') }}
        </p>
      </div>
    </div>

    <div class="spoke-nipple-step__input-panel">
      <div class="spoke-nipple-step__input-copy">
        <span class="spoke-nipple-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleInputKicker', '第 7 步输入') }}
        </span>
        <strong class="spoke-nipple-step__input-label">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleInputLabel', '辐条帽位置与长度') }}
        </strong>
        <p class="spoke-nipple-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nipplePairHelp', '请分别设置前轮和后轮的辐条帽类型；选择内置式后，再填写实际长度。') }}
        </p>
        <div class="spoke-nipple-step__input-note" role="note">
          <span class="spoke-nipple-step__input-note-icon" aria-hidden="true">14</span>
          <p>
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleInputNote', '外置式以 SAPIM POLYAX 14 mm 为测长基准，不额外填写长度。内置式请按实际型号或官方图纸填写辐条帽长度；不要直接用外置式的 14 mm 代替。') }}
          </p>
        </div>
      </div>

      <div class="spoke-nipple-step__wheel-grid">
        <fieldset class="spoke-nipple-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.frontWheel') }}</legend>
          <p class="spoke-nipple-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.frontNippleNote', '前轮辐条帽') }}
          </p>
          <label for="spoke-nipple-type-front" class="spoke-nipple-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleTypeFieldLabel', '辐条帽类型') }}</span>
            <SpokeCalculatorSelect
              id="spoke-nipple-type-front"
              :model-value="props.frontNippleType"
              :options="nippleOptions"
              @update:model-value="updateNippleType('front', $event)"
            />
          </label>
          <label
            v-if="props.frontNippleType === 'hidden'"
            for="spoke-nipple-length-front"
            class="spoke-nipple-step__field"
          >
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleLengthFieldLabel', '内置辐条帽长度') }}</span>
            <span class="spoke-nipple-step__unit-field">
              <input
                id="spoke-nipple-length-front"
                :value="props.frontNippleLength ?? ''"
                type="number"
                min="3"
                max="40"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleLengthPlaceholder', '例如 12 或 14')"
                @input="updateNippleLength('front', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>

        <fieldset class="spoke-nipple-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.rearWheel') }}</legend>
          <p class="spoke-nipple-step__wheel-note">
            {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.rearNippleNote', '后轮辐条帽') }}
          </p>
          <label for="spoke-nipple-type-rear" class="spoke-nipple-step__field">
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleTypeFieldLabel', '辐条帽类型') }}</span>
            <SpokeCalculatorSelect
              id="spoke-nipple-type-rear"
              :model-value="props.rearNippleType"
              :options="nippleOptions"
              @update:model-value="updateNippleType('rear', $event)"
            />
          </label>
          <label
            v-if="props.rearNippleType === 'hidden'"
            for="spoke-nipple-length-rear"
            class="spoke-nipple-step__field"
          >
            <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleLengthFieldLabel', '内置辐条帽长度') }}</span>
            <span class="spoke-nipple-step__unit-field">
              <input
                id="spoke-nipple-length-rear"
                :value="props.rearNippleLength ?? ''"
                type="number"
                min="3"
                max="40"
                step="0.1"
                inputmode="decimal"
                :placeholder="t('resourcesSpokeCalculator.calculator.physicalCorrections.nippleLengthPlaceholder', '例如 12 或 14')"
                @input="updateNippleLength('rear', $event)"
              />
              <span>mm</span>
            </span>
          </label>
        </fieldset>
      </div>
    </div>

    <SpokeNippleGuide />
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import SpokeCalculatorSelect from '~/components/SpokeCalculatorSelect.vue'
import SpokeNippleGuide from '~/components/SpokeNippleGuide.vue'
import SpokeStepNavigation from '~/components/SpokeStepNavigation.vue'
import type { SpokeNippleType } from '~/types/spokeCalculator'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontNippleType?: SpokeNippleType
  rearNippleType?: SpokeNippleType
  frontNippleLength?: number | null
  rearNippleLength?: number | null
}>(), {
  currentStep: 7,
  frontNippleType: 'standard',
  rearNippleType: 'standard',
  frontNippleLength: 12,
  rearNippleLength: 12,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontNippleType': [value: SpokeNippleType]
  'update:rearNippleType': [value: SpokeNippleType]
  'update:frontNippleLength': [value: number | null]
  'update:rearNippleLength': [value: number | null]
  previous: []
}>()

const nippleOptions = computed(() => [
  {
    value: 'standard',
    label: t(
      'resourcesSpokeCalculator.calculator.physicalCorrections.nippleExternal',
      '外置式（SAPIM POLYAX 14 mm 基准）',
    ),
  },
  {
    value: 'hidden',
    label: t(
      'resourcesSpokeCalculator.calculator.physicalCorrections.nippleInternal',
      '内置式 / 隐藏式',
    ),
  },
])

const readNumber = (event: Event): number | null => {
  const rawValue = (event.target as HTMLInputElement).value.trim()
  if (!rawValue) return null
  const value = Number(rawValue)
  return Number.isFinite(value) ? value : null
}

const updateNippleType = (side: 'front' | 'rear', value: string | number | null) => {
  if (value !== 'standard' && value !== 'hidden') return
  if (side === 'front') {
    emit('update:frontNippleType', value)
  } else {
    emit('update:rearNippleType', value)
  }
}

const updateNippleLength = (side: 'front' | 'rear', event: Event) => {
  const value = readNumber(event)
  if (side === 'front') {
    emit('update:frontNippleLength', value)
  } else {
    emit('update:rearNippleLength', value)
  }
}
</script>

<style scoped>
.spoke-nipple-step {
  --nipple-step-border: rgba(15, 23, 42, 0.12);
  --nipple-step-text: #0f172a;
  --nipple-step-muted: #475569;
  --nipple-step-accent: #059669;
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
  color: var(--nipple-step-text);
}

.spoke-nipple-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 14px;
}

.spoke-nipple-step__eyebrow {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: var(--nipple-step-accent);
  color: #ffffff;
  font-family: var(--tz-font-ui);
  font-size: 0.7rem;
  font-weight: 900;
}

.spoke-nipple-step__title {
  margin: 0;
  color: var(--nipple-step-text);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-nipple-step__subtitle {
  max-width: 70rem;
  margin: 3px 0 0;
  color: var(--nipple-step-muted);
  font-size: 12px;
  line-height: 1.5;
}

.spoke-nipple-step__input-panel {
  display: grid;
  grid-template-columns: minmax(210px, 0.55fr) minmax(0, 1.45fr);
  gap: 18px;
  align-items: start;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--nipple-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-nipple-step__input-copy {
  min-width: 0;
}

.spoke-nipple-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: var(--tz-font-ui);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-nipple-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-nipple-step__input-help {
  margin: 4px 0 0;
  color: var(--nipple-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-nipple-step__input-note {
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

.spoke-nipple-step__input-note-icon {
  display: inline-flex;
  width: 1.35rem;
  height: 1.35rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgba(217, 119, 6, 0.14);
  color: #b45309;
  font-family: var(--tz-font-ui);
  font-size: 0.62rem;
  font-weight: 900;
}

.spoke-nipple-step__input-note p {
  margin: 0;
}

.spoke-nipple-step__wheel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 0;
}

.spoke-nipple-step__wheel-card {
  min-width: 0;
  margin: 0;
  padding: 12px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.78);
}

.spoke-nipple-step__wheel-card legend {
  padding: 0 6px;
  color: #047857;
  font-size: 12px;
  font-weight: 900;
}

.spoke-nipple-step__wheel-note {
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

.spoke-nipple-step__field {
  display: grid;
  gap: 5px;
  min-width: 0;
  margin-top: 10px;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.25;
}

.spoke-nipple-step__field:first-of-type {
  margin-top: 0;
}

.spoke-nipple-step__field :deep(.spoke-calculator-select__button) {
  min-height: 38px;
  border: 2px solid rgba(5, 150, 105, 0.38);
  padding: 8px 11px;
  color: var(--nipple-step-text);
  font-size: 13px;
  font-weight: 700;
}

.spoke-nipple-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 2px solid rgba(5, 150, 105, 0.38);
  border-radius: 9999px;
  background: #ffffff;
}

.spoke-nipple-step__unit-field input {
  width: 100%;
  min-width: 0;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--nipple-step-text);
  padding: 8px 5px 8px 11px;
  font-family: var(--tz-font-ui);
  font-size: 13px;
  font-weight: 700;
}

.spoke-nipple-step__unit-field input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--nipple-step-accent);
}

.spoke-nipple-step__unit-field > span {
  flex: 0 0 auto;
  padding-right: 10px;
  color: #047857;
  font-family: var(--tz-font-ui);
  font-size: 10px;
  font-weight: 800;
}

@media (max-width: 767px) {
  .spoke-nipple-step {
    padding: 16px;
    border-radius: 20px;
  }

  .spoke-nipple-step__input-panel {
    grid-template-columns: 1fr;
    gap: 12px;
    padding: 15px;
  }

  .spoke-nipple-step__wheel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
