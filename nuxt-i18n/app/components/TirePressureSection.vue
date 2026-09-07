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
