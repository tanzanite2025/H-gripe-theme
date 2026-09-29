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

    <div class="spoke-pcd-step__input-panel">
      <div class="spoke-pcd-step__input-copy">
        <span class="spoke-pcd-step__input-kicker">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdInputKicker', 'STEP 3 INPUT') }}
        </span>
        <strong class="spoke-pcd-step__input-label">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdInputLabel', '花鼓 PCD / WL / WR') }}
        </strong>
        <p class="spoke-pcd-step__input-help">
          {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdPairHelp', '请分别填写前轮和后轮花鼓的 PCD、WL、WR。') }}
        </p>
      </div>

      <div class="spoke-pcd-step__wheel-grid">
        <fieldset class="spoke-pcd-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.frontWheel') }}</legend>
          <div class="spoke-pcd-step__fields">
            <label for="spoke-pcd-front-left" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdLeftLabel', 'PCDL（左法兰）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-front-left"
                  :value="frontGeometry.leftFlangePcd ?? ''"
                  type="number"
                  min="30"
                  max="80"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateFrontGeometry('leftFlangePcd', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-front-right" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdRightLabel', 'PCDR（右法兰）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-front-right"
                  :value="frontGeometry.rightFlangePcd ?? ''"
                  type="number"
                  min="30"
                  max="80"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateFrontGeometry('rightFlangePcd', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-front-wl" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.wlLabel', 'WL（左法兰至中心线）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-front-wl"
                  :value="frontGeometry.leftFlange ?? ''"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateFrontGeometry('leftFlange', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-front-wr" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.wrLabel', 'WR（右法兰至中心线）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-front-wr"
                  :value="frontGeometry.rightFlange ?? ''"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateFrontGeometry('rightFlange', $event)"
                />
                <span>mm</span>
              </span>
            </label>
          </div>
        </fieldset>

        <fieldset class="spoke-pcd-step__wheel-card">
          <legend>{{ t('resourcesSpokeCalculator.calculator.rearWheel') }}</legend>
          <div class="spoke-pcd-step__fields">
            <label for="spoke-pcd-rear-left" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdLeftLabel', 'PCDL（左法兰）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-rear-left"
                  :value="rearGeometry.leftFlangePcd ?? ''"
                  type="number"
                  min="30"
                  max="80"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateRearGeometry('leftFlangePcd', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-rear-right" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.pcdRightLabel', 'PCDR（右法兰）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-rear-right"
                  :value="rearGeometry.rightFlangePcd ?? ''"
                  type="number"
                  min="30"
                  max="80"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateRearGeometry('rightFlangePcd', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-rear-wl" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.wlLabel', 'WL（左法兰至中心线）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-rear-wl"
                  :value="rearGeometry.leftFlange ?? ''"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateRearGeometry('leftFlange', $event)"
                />
                <span>mm</span>
              </span>
            </label>

            <label for="spoke-pcd-rear-wr" class="spoke-pcd-step__field">
              <span>{{ t('resourcesSpokeCalculator.calculator.physicalCorrections.wrLabel', 'WR（右法兰至中心线）') }}</span>
              <span class="spoke-pcd-step__unit-field">
                <input
                  id="spoke-pcd-rear-wr"
                  :value="rearGeometry.rightFlange ?? ''"
                  type="number"
                  min="0"
                  max="100"
                  step="0.1"
                  inputmode="decimal"
                  @input="updateRearGeometry('rightFlange', $event)"
                />
                <span>mm</span>
              </span>
            </label>
          </div>
        </fieldset>
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
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import SpokeStepProgress from '~/components/SpokeStepProgress.vue'

const { t } = useI18n()

const props = withDefaults(defineProps<{
  currentStep?: number
  frontGeometry?: HubGeometry | null
  rearGeometry?: HubGeometry | null
}>(), {
  currentStep: 3,
  frontGeometry: null,
  rearGeometry: null,
})

const emit = defineEmits<{
  'select-step': [step: number]
  'update:frontGeometry': [value: HubGeometry]
  'update:rearGeometry': [value: HubGeometry]
  previous: []
}>()

const emptyGeometry = (): HubGeometry => ({
  leftFlange: null,
  rightFlange: null,
  leftFlangePcd: null,
  rightFlangePcd: null,
})

const frontGeometry = computed(() => props.frontGeometry ?? emptyGeometry())
const rearGeometry = computed(() => props.rearGeometry ?? emptyGeometry())

type GeometryField = 'leftFlange' | 'rightFlange' | 'leftFlangePcd' | 'rightFlangePcd'

const readNumber = (event: Event): number | null => {
  const rawValue = (event.target as HTMLInputElement).value.trim()
  if (!rawValue) return null
  const value = Number(rawValue)
  return Number.isFinite(value) ? value : null
}

const updateFrontGeometry = (field: GeometryField, event: Event) => {
  emit('update:frontGeometry', {
    ...(props.frontGeometry ?? emptyGeometry()),
    [field]: readNumber(event),
  })
}

const updateRearGeometry = (field: GeometryField, event: Event) => {
  emit('update:rearGeometry', {
    ...(props.rearGeometry ?? emptyGeometry()),
    [field]: readNumber(event),
  })
}
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

.spoke-pcd-step__input-panel {
  display: grid;
  grid-template-columns: minmax(210px, 0.55fr) minmax(0, 1.45fr);
  gap: 18px;
  align-items: start;
  margin-bottom: 24px;
  padding: 18px 20px;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-left: 5px solid var(--pcd-step-accent);
  border-radius: 18px;
  background: linear-gradient(135deg, #ecfdf5 0%, #f0fdfa 54%, #ffffff 100%);
  box-shadow: 0 10px 24px rgba(5, 150, 105, 0.1);
}

.spoke-pcd-step__input-copy {
  min-width: 0;
}

.spoke-pcd-step__input-kicker {
  display: block;
  margin-bottom: 4px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  line-height: 1.2;
}

.spoke-pcd-step__input-label {
  display: block;
  color: #064e3b;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.35;
}

.spoke-pcd-step__input-help {
  margin: 4px 0 0;
  color: var(--pcd-step-muted);
  font-size: 12px;
  line-height: 1.45;
}

.spoke-pcd-step__wheel-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  min-width: 0;
}

.spoke-pcd-step__wheel-card {
  min-width: 0;
  margin: 0;
  padding: 12px;
  border: 1px solid rgba(5, 150, 105, 0.2);
  border-radius: 15px;
  background: rgba(255, 255, 255, 0.78);
}

.spoke-pcd-step__wheel-card legend {
  padding: 0 6px;
  color: #047857;
  font-size: 12px;
  font-weight: 900;
}

.spoke-pcd-step__fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.spoke-pcd-step__field {
  display: grid;
  gap: 5px;
  min-width: 0;
  color: #047857;
  font-size: 10px;
  font-weight: 900;
  line-height: 1.25;
}

.spoke-pcd-step__unit-field {
  display: flex;
  align-items: center;
  min-width: 0;
  overflow: hidden;
  border: 2px solid rgba(5, 150, 105, 0.38);
  border-radius: 9999px;
  background: #ffffff;
}

.spoke-pcd-step__unit-field input {
  min-width: 0;
  width: 100%;
  border: 0;
  outline: 0;
  background: transparent;
  color: var(--pcd-step-text);
  padding: 8px 5px 8px 11px;
  font-size: 13px;
  font-weight: 700;
}

.spoke-pcd-step__unit-field input:focus-visible {
  box-shadow: inset 0 0 0 2px var(--pcd-step-accent);
}

.spoke-pcd-step__unit-field > span {
  flex: 0 0 auto;
  padding-right: 10px;
  color: #047857;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 10px;
  font-weight: 800;
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

  .spoke-pcd-step__input-panel {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 18px;
    padding: 15px;
  }

  .spoke-pcd-step__wheel-grid {
    grid-template-columns: 1fr;
  }
}
</style>
