<template>
  <div class="tire-rim-calculator-guidance">
    <div class="tire-rim-calculator-guidance__checklist">
      <h4>{{ t('guidesTireChoose.engineering.checklist.title') }}</h4>
      <ul>
        <li v-for="checklistKey in checklistKeys" :key="checklistKey">
          {{ t(`guidesTireChoose.engineering.checklist.items.${checklistKey}`) }}
        </li>
      </ul>
    </div>

    <section v-if="!suggestion || suggestion.aeroRuleApplies" class="tire-rim-calculator-guidance__aero-board">
      <div class="tire-rim-calculator-guidance__board-section-heading">
        <h4>{{ t('guidesTireChoose.engineering.board.aero.title') }}</h4>
        <span>{{ t('guidesTireChoose.engineering.board.aero.source') }}</span>
      </div>
      <p class="tire-rim-calculator-guidance__board-section-intro">
        {{ t('guidesTireChoose.engineering.board.aero.intro') }}
      </p>
      <div class="tire-rim-calculator-guidance__aero-grid">
        <article v-for="aeroKey in aeroKeys" :key="aeroKey">
          <h5>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.title`) }}</h5>
          <strong>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.formula`) }}</strong>
          <p>{{ t(`guidesTireChoose.engineering.board.aero.cards.${aeroKey}.body`) }}</p>
        </article>
      </div>
      <div class="tire-rim-calculator-guidance__aero-formula">
        <strong>{{ t('guidesTireChoose.engineering.board.aero.formulaLabel') }}</strong>
        <code>W<sub>tire</sub> ≈ W<sub>rim outer</sub> × 1.05</code>
        <span v-if="rimOuterWidthEstimate && aeroTargetTireWidthMm !== null">
          {{ t('guidesTireChoose.engineering.board.aero.currentPairing', {
            rimWidth: rimOuterWidthEstimate.rimInnerWidthMm,
            wallThickness: rimOuterWidthEstimate.wallThicknessPerSideEstimateMm,
            rimOuterWidth: rimOuterWidthEstimate.rimOuterWidthEstimateMm.toFixed(1),
            targetTireWidth: aeroTargetTireWidthMm,
          }) }}
        </span>
      </div>
    </section>

    <section class="tire-rim-calculator-guidance__traps-board">
      <div class="tire-rim-calculator-guidance__board-section-heading">
        <h4>{{ t('guidesTireChoose.engineering.board.traps.title') }}</h4>
        <span>{{ t('guidesTireChoose.engineering.board.traps.source') }}</span>
      </div>
      <div class="tire-rim-calculator-guidance__traps-grid">
        <article v-for="trapKey in trapKeys" :key="trapKey">
          <span class="tire-rim-calculator-guidance__trap-number">
            {{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.number`) }}
          </span>
          <h5>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.title`) }}</h5>
          <p>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.body`) }}</p>
          <strong>{{ t(`guidesTireChoose.engineering.board.traps.items.${trapKey}.rule`) }}</strong>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import type { TireRimSuggestion, TireRimReverseSuggestion } from '~/composables/useTireRimWidthReferenceRecommendation'

const props = defineProps<{ suggestion: TireRimSuggestion | TireRimReverseSuggestion | null }>()
const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireChoose')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

const checklistKeys = ['exactSpecifications', 'pressureLimits', 'installationCheck'] as const
const aeroKeys = ['bulbTire', 'ruleOf105', 'modernSweetSpot'] as const
const trapKeys = ['nonTleOnHookless', 'emptyTpuInflation', 'oldSealantResidue', 'co2ColdBrittle', 'metalLever', 'beadPinch'] as const

// Present only values returned for this calculator's active lookup direction.
const rimOuterWidthEstimate = computed(() => {
  const suggestion = props.suggestion
  if (!suggestion?.aeroRuleApplies) return null
  if ('rimOuterWidthEstimate' in suggestion) return suggestion.rimOuterWidthEstimate
  const bestRimInnerWidthMm = suggestion.bestFitment?.rimInnerWidthMm
  return suggestion.physical.rimOuterWidthEstimates.find(
    estimate => estimate.rimInnerWidthMm === bestRimInnerWidthMm,
  ) ?? null
})
const aeroTargetTireWidthMm = computed(() => {
  const suggestion = props.suggestion
  if (!suggestion?.aeroRuleApplies) return null
  return 'aeroTargetTireWidthMm' in suggestion
    ? suggestion.aeroTargetTireWidthMm
    : suggestion.bestFitment?.targetTireWidthMm ?? null
})
</script>

<style scoped>
.tire-rim-calculator-guidance {
  display: grid;
  gap: 1.25rem;
  margin-top: 1.25rem;
  text-align: left;
}

.tire-rim-calculator-guidance__checklist {
  padding: 0.8rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.9rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-calculator-guidance__checklist h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  font-weight: 700;
}

.tire-rim-calculator-guidance__checklist ul {
  display: grid;
  gap: 0.2rem;
  margin: 0.45rem 0 0;
  padding-left: 1.1rem;
}

.tire-rim-calculator-guidance__checklist li {
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.7rem;
  line-height: 1.6;
}

.tire-rim-calculator-guidance__aero-board,
.tire-rim-calculator-guidance__traps-board {
  padding: 1rem;
  border: 1px dashed var(--tz-border-subtle);
  border-radius: 0.95rem;
  background: var(--tz-form-panel-surface);
}

.tire-rim-calculator-guidance__board-section-heading {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  justify-content: space-between;
  padding-bottom: 0.55rem;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.tire-rim-calculator-guidance__board-section-heading h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 900;
}

.tire-rim-calculator-guidance__board-section-heading span {
  color: var(--tz-text-muted);
  font-family: var(--tz-font-ui);
  font-size: 0.58rem;
  font-weight: 800;
  text-align: right;
}

.tire-rim-calculator-guidance__board-section-intro {
  margin: 0.65rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
  line-height: 1.6;
}

.tire-rim-calculator-guidance__aero-grid,
.tire-rim-calculator-guidance__traps-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.7rem;
  margin-top: 0.8rem;
}

.tire-rim-calculator-guidance__aero-grid article,
.tire-rim-calculator-guidance__traps-grid article {
  min-width: 0;
  padding: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.tire-rim-calculator-guidance__aero-grid h5,
.tire-rim-calculator-guidance__traps-grid h5 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.72rem;
  line-height: 1.4;
}

.tire-rim-calculator-guidance__aero-grid strong {
  display: block;
  margin-top: 0.45rem;
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-size: 0.66rem;
  line-height: 1.45;
}

.tire-rim-calculator-guidance__aero-grid p,
.tire-rim-calculator-guidance__traps-grid p {
  margin: 0.45rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
  line-height: 1.55;
}

.tire-rim-calculator-guidance__aero-formula {
  display: flex;
  gap: 0.7rem;
  align-items: baseline;
  flex-wrap: wrap;
  margin-top: 0.8rem;
  padding: 0.7rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-secondary);
  font-size: 0.66rem;
}

.tire-rim-calculator-guidance__aero-formula strong,
.tire-rim-calculator-guidance__aero-formula code {
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-weight: 900;
}

.tire-rim-calculator-guidance__aero-formula code {
  color: #0284c7;
}

.tire-rim-calculator-guidance__aero-formula span {
  flex: 1 1 100%;
  line-height: 1.5;
}

.tire-rim-calculator-guidance__trap-number {
  display: inline-flex;
  padding: 0.2rem 0.38rem;
  border-radius: 0.25rem;
  background: #fee2e2;
  color: #b91c1c;
  font-family: var(--tz-font-ui);
  font-size: 0.56rem;
  font-weight: 900;
}

.tire-rim-calculator-guidance__traps-grid article strong {
  display: block;
  margin-top: 0.55rem;
  padding: 0.45rem;
  border-left: 3px solid #0284c7;
  border-radius: 0.3rem;
  background: rgb(239 246 255 / 0.8);
  color: #075985;
  font-size: 0.62rem;
  line-height: 1.45;
}

@media (max-width: 900px) {
  .tire-rim-calculator-guidance__aero-grid,
  .tire-rim-calculator-guidance__traps-grid {
    grid-template-columns: 1fr;
  }
}
</style>
