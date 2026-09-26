<template>
  <div class="tire-pressure-section">
    <div class="mt-1 space-y-4 text-sm tz-text-primary">
      <p>
        {{ t('guidesTirePressure.intro.matching') }}
      </p>
      <p>
        {{ t('guidesTirePressure.intro.width') }}
      </p>
      <p>
        {{ t('guidesTirePressure.intro.etrto') }}
      </p>

      <TirePressureCalculator />

      <p class="guide-section__cta-wrapper">
        <button
          type="button"
          class="tire-pressure-section__cta"
          @click="emit('openTireProducts')"
        >
          {{ t('guidesTirePressure.actions.checkTires') }}
        </button>
      </p>

      <div>
        <h3 class="tire-pressure-section__heading">
          {{ t('guidesTirePressure.sections.why.title') }}
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>
            <span class="font-semibold">{{ t('guidesTirePressure.sections.why.performanceLabel') }}:</span>
            {{ t('guidesTirePressure.sections.why.performance') }}
          </li>
          <li>
            <span class="font-semibold">{{ t('guidesTirePressure.sections.why.safetyLabel') }}:</span>
            {{ t('guidesTirePressure.sections.why.safety') }}
          </li>
          <li>
            <span class="font-semibold">{{ t('guidesTirePressure.sections.why.comfortLabel') }}:</span>
            {{ t('guidesTirePressure.sections.why.comfort') }}
          </li>
        </ul>
      </div>

      <div>
        <h3 class="tire-pressure-section__heading">
          {{ t('guidesTirePressure.sections.factors.title') }}
        </h3>
        <div class="tire-pressure-section__table-shell">
          <table class="tire-pressure-section__table">
            <thead class="tire-pressure-section__table-head">
              <tr>
                <th class="px-3 py-2 font-semibold">{{ t('guidesTirePressure.sections.factors.headers.factor') }}</th>
                <th class="px-3 py-2 font-semibold">{{ t('guidesTirePressure.sections.factors.headers.explanation') }}</th>
                <th class="px-3 py-2 font-semibold">{{ t('guidesTirePressure.sections.factors.headers.adjustment') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.riderWeight.factor') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.riderWeight.explanation') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.riderWeight.adjustment') }}</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.tireWidth.factor') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.tireWidth.explanation') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.tireWidth.adjustment') }}</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.terrain.factor') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.terrain.explanation') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.terrain.adjustment') }}</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.ridingStyle.factor') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.ridingStyle.explanation') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.ridingStyle.adjustment') }}</td>
              </tr>
              <tr>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.rimDesign.factor') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.rimDesign.explanation') }}</td>
                <td class="px-3 py-2 align-top">{{ t('guidesTirePressure.sections.factors.rimDesign.adjustment') }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div>
        <h3 class="tire-pressure-section__heading">
          {{ t('guidesTirePressure.sections.typical.title') }}
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>{{ t('guidesTirePressure.sections.typical.road') }}</li>
          <li>{{ t('guidesTirePressure.sections.typical.mountain') }}</li>
          <li>{{ t('guidesTirePressure.sections.typical.city') }}</li>
        </ul>
      </div>

      <div>
        <h3 class="tire-pressure-section__heading">
          {{ t('guidesTirePressure.sections.tips.title') }}
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>{{ t('guidesTirePressure.sections.tips.sidewall') }}</li>
          <li>{{ t('guidesTirePressure.sections.tips.gauge') }}</li>
          <li>{{ t('guidesTirePressure.sections.tips.inspect') }}</li>
          <li>{{ t('guidesTirePressure.sections.tips.extremes') }}</li>
          <li>{{ t('guidesTirePressure.sections.tips.seasonal') }}</li>
        </ul>
      </div>

      <div>
        <h3 class="tire-pressure-section__heading">
          {{ t('guidesTirePressure.sections.warnings.title') }}
        </h3>
        <ul class="sizecharts-section__list mt-1">
          <li>{{ t('guidesTirePressure.sections.warnings.high') }}</li>
          <li>{{ t('guidesTirePressure.sections.warnings.low') }}</li>
          <li>{{ t('guidesTirePressure.sections.warnings.neglect') }}</li>
          <li class="tire-pressure-section__safety-warning">
            {{ t('guidesTirePressure.sections.warnings.hookless') }}
          </li>
        </ul>
      </div>
    </div>

    <div
      class="tire-pressure-section__comparison"
    >
      <h3 class="mb-3 text-sm font-semibold tz-text-primary">
        {{ t('guidesTirePressure.comparison.title') }}
      </h3>

      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <div
            class="tire-pressure-section__comparison-label"
          >
            {{ t('guidesTirePressure.comparison.high.title') }}
          </div>
          <ul class="sizecharts-section__list tz-text-primary">
            <li>{{ t('guidesTirePressure.comparison.high.stability') }}</li>
            <li>{{ t('guidesTirePressure.comparison.high.puncture') }}</li>
            <li>{{ t('guidesTirePressure.comparison.high.grip') }}</li>
            <li>{{ t('guidesTirePressure.comparison.high.comfort') }}</li>
            <li>{{ t('guidesTirePressure.comparison.high.rolling') }}</li>
          </ul>
        </div>

        <div>
          <div
            class="tire-pressure-section__comparison-label"
          >
            {{ t('guidesTirePressure.comparison.low.title') }}
          </div>
          <ul class="sizecharts-section__list tz-text-primary">
            <li>{{ t('guidesTirePressure.comparison.low.grip') }}</li>
            <li>{{ t('guidesTirePressure.comparison.low.rolling') }}</li>
            <li>{{ t('guidesTirePressure.comparison.low.comfort') }}</li>
            <li>{{ t('guidesTirePressure.comparison.low.stability') }}</li>
            <li>{{ t('guidesTirePressure.comparison.low.puncture') }}</li>
          </ul>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import TirePressureCalculator from '~/components/tireguides/tirepressure/TirePressureCalculator.vue'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const emit = defineEmits<{
  (e: 'openTireProducts'): void
}>()
</script>

<style scoped>
.tire-pressure-section__cta {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
  cursor: pointer;
  font-size: 0.75rem;
  font-weight: 700;
  padding: 0.375rem 1rem;
  transition: background-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.tire-pressure-section__cta:hover {
  background: var(--tz-action-primary-hover);
  box-shadow: 0 0.5rem 1.35rem -0.5rem rgba(0, 0, 0, 0.9);
  transform: translateY(-1px);
}

.tire-pressure-section__calculator {
  margin-top: 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-form-panel-surface);
  padding: 1rem;
}

.tire-pressure-section__calculator-description {
  margin: 0.45rem auto 0;
  max-width: 48rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  line-height: 1.5;
  text-align: center;
}

.tire-pressure-section__calculator-controls {
  display: flex;
  flex-wrap: wrap;
  align-items: end;
  justify-content: center;
  gap: 0.65rem;
  margin-top: 0.85rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.tire-pressure-section__calculator-controls label,
.tire-pressure-section__calculator-system-label {
  display: block;
  font-weight: 600;
}

.tire-pressure-section__calculator-controls input {
  width: 8rem;
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.4rem;
  background: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
  padding: 0.4rem 0.55rem;
  font-size: 0.75rem;
  outline: none;
}

.tire-pressure-section__calculator-controls input:focus {
  border-color: var(--tz-site-accent);
}

.tire-pressure-section__toggle-group {
  display: inline-flex;
  gap: 0.25rem;
  border-radius: 999px;
  background: var(--tz-form-control-surface);
  padding: 0.2rem;
}

.tire-pressure-section__toggle-group button {
  border: 1px solid transparent;
  border-radius: 999px;
  background: transparent;
  color: var(--tz-text-secondary);
  cursor: pointer;
  padding: 0.35rem 0.7rem;
  font-size: 0.7rem;
  font-weight: 600;
}

.tire-pressure-section__toggle-group button:hover,
.tire-pressure-section__toggle-group button:focus-visible {
  border-color: var(--tz-site-accent);
  outline: none;
}

.tire-pressure-section__toggle-group button.tire-pressure-section__toggle--active {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: #ffffff;
}

.tire-pressure-section__calculator-result {
  margin: 0.85rem 0 0;
  color: var(--tz-text-primary);
  font-size: 0.9rem;
  font-weight: 700;
  text-align: center;
}

.tire-pressure-section__safety-warning {
  color: #b91c1c;
  font-weight: 700;
}

.tire-pressure-section__calculator .tire-pressure-section__safety-warning {
  margin: 0.45rem 0 0;
  font-size: 0.85rem;
  line-height: 1.5;
  text-align: center;
}

.tire-pressure-section__heading {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.875rem;
  font-weight: 600;
  line-height: 1.35;
  text-align: center;
}

.tire-pressure-section__table-shell {
  margin-top: 0.5rem;
  overflow-x: auto;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-form-panel-surface);
}

.tire-pressure-section__table {
  width: 100%;
  min-width: 40rem;
  border-collapse: collapse;
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  text-align: left;
}

.tire-pressure-section__table-head {
  background: var(--tz-card-surface);
}

.tire-pressure-section__table th,
.tire-pressure-section__table td {
  padding: 0.5rem 0.75rem;
  vertical-align: top;
}

.tire-pressure-section__table th {
  font-weight: 600;
}

.tire-pressure-section__table tbody tr + tr {
  border-top: 1px solid var(--tz-border-subtle);
}

.tire-pressure-section__comparison {
  margin-top: 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-form-panel-surface);
  box-shadow: 3px 3px 10px rgba(20, 32, 43, 0.1);
  padding: 1rem;
}

.tire-pressure-section__comparison-label {
  margin-bottom: 0.5rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.35rem;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 1.2;
  padding: 0.375rem 0.75rem;
  text-transform: uppercase;
}

@media (min-width: 640px) {
  .tire-pressure-section__table {
    font-size: 0.875rem;
  }
}
</style>
