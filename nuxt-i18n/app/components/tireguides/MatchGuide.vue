<template>
  <div class="match-guide space-y-8">
    <section class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-4 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
        {{ t('guidesTireMatch.title') }}
      </h2>
      <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        {{ t('guidesTireMatch.intro') }}
      </p>

      <section class="match-guide__measurement-reference" aria-labelledby="measurement-reference-title">
        <div class="mx-auto w-full max-w-3xl">
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame3.webp"
            :alt="t('guidesTireMatch.images.measurementAlt')"
            :caption="t('guidesTireMatch.images.measurementCaption')"
            :zoomOnClick="true"
          />
        </div>
        <div class="mt-5">
          <h3 id="measurement-reference-title" class="match-guide__measurement-title">
            {{ t('guidesTireMatch.measurement.title') }}
          </h3>
          <p class="mx-auto mt-2 max-w-3xl text-sm leading-relaxed tz-text-secondary">
            {{ t('guidesTireMatch.measurement.intro') }}
          </p>
          <p class="mx-auto mt-3 max-w-3xl text-sm leading-relaxed tz-text-secondary">
            {{ t('guidesTireMatch.measurement.important') }}
          </p>
        </div>
        <div class="match-guide__measurement-grid">
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">A</span>
            <div>
              <h4>{{ t('guidesTireMatch.measurement.items.width.title') }}</h4>
              <p>
                {{ t('guidesTireMatch.measurement.items.width.body') }}
              </p>
            </div>
          </div>
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">B</span>
            <div>
              <h4>{{ t('guidesTireMatch.measurement.items.diameter.title') }}</h4>
              <p>
                {{ t('guidesTireMatch.measurement.items.diameter.body') }}
              </p>
            </div>
          </div>
          <div class="match-guide__measurement-item">
            <span class="match-guide__measurement-badge">C</span>
            <div>
              <h4>{{ t('guidesTireMatch.measurement.items.shoulder.title') }}</h4>
              <p>
                {{ t('guidesTireMatch.measurement.items.shoulder.body') }}
              </p>
            </div>
          </div>
        </div>
      </section>

      <div class="match-guide__table-list">
        <section
          v-for="chart in tireFrameClearanceTables"
          :key="chart.key"
          class="match-guide__table-section"
          :aria-labelledby="`${chart.key}-title`"
        >
          <h3 :id="`${chart.key}-title`" class="match-guide__table-title">
            {{ t(`guidesTireMatch.clearance.tables.${chart.key}`) }}
          </h3>
          <div class="match-guide__table-scroll">
            <table class="match-guide__table">
              <caption>
                {{ t(`guidesTireMatch.clearance.tables.${chart.key}`) }}
                {{ t('guidesTireMatch.clearance.captionSuffix') }}
              </caption>
              <thead>
                <tr>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.wheelSize') }}</th>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.etrto') }}</th>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.tire') }}</th>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.maxWidth') }}</th>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.maxDiameter') }}</th>
                  <th scope="col">{{ t('guidesTireMatch.clearance.headers.shoulderDiameter') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in chart.rows" :key="`${chart.key}-${row.inch}-${row.etrto}-${row.tire}`">
                  <td class="match-guide__dimension-cell">{{ row.inch }}</td>
                  <td class="match-guide__dimension-cell">{{ row.etrto }}</td>
                  <th scope="row">{{ row.tire }}</th>
                  <td class="match-guide__dimension-cell">{{ row.maxWidthMm }} mm</td>
                  <td class="match-guide__dimension-cell">{{ row.maxDiameterMm }} mm</td>
                  <td class="match-guide__dimension-cell">{{ row.shoulderDiameterMm }} mm</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <div class="mt-8 border-t tz-border-subtle pt-6">
        <h3 class="match-guide__table-title">{{ t('guidesTireMatch.sourceChartsTitle') }}</h3>
        <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame1.webp"
            :alt="t('guidesTireMatch.images.chart2729Alt')"
            :caption="t('guidesTireMatch.images.chart2729Caption')"
            :zoomOnClick="true"
          />
          <GuideImage
            class="overflow-hidden rounded-xl shadow-md"
            src="/public/tiresizecharts/match/schwalbe-tire-fit-frame2.webp"
            :alt="t('guidesTireMatch.images.chart2426Alt')"
            :caption="t('guidesTireMatch.images.chart2426Caption')"
            :zoomOnClick="true"
          />
        </div>
      </div>
    </section>

    <section class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-4 flex items-center justify-center gap-2 text-lg font-bold uppercase tracking-wide tz-text-secondary">
        {{ t('guidesTireMatch.circumference.title') }}
      </h2>
      <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        {{ t('guidesTireMatch.circumference.intro') }}
      </p>
      <p class="mx-auto mb-8 max-w-3xl text-sm leading-relaxed tz-text-secondary">
        {{ t('guidesTireMatch.circumference.test') }}
      </p>

      <div class="match-guide__table-scroll text-left">
        <table class="match-guide__table match-guide__table--circumference">
          <caption>
            {{ t('guidesTireMatch.circumference.caption') }}
          </caption>
          <thead>
            <tr>
              <th scope="col">{{ t('guidesTireMatch.circumference.headers.wheelSize') }}</th>
              <th scope="col">{{ t('guidesTireMatch.circumference.headers.etrto') }}</th>
              <th scope="col">{{ t('guidesTireMatch.circumference.headers.circumference') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in tireCircumferenceRows" :key="`${row.inch}-${row.etrto}`">
              <td class="match-guide__dimension-cell">{{ row.inch }}</td>
              <th scope="row">{{ row.etrto }}</th>
              <td class="match-guide__dimension-cell">{{ row.circumferenceMm }} mm</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mx-auto mt-8 w-full max-w-4xl">
        <h3 class="match-guide__table-title">{{ t('guidesTireMatch.circumference.visualTitle') }}</h3>
        <GuideImage
          class="mt-4 overflow-hidden rounded-xl shadow-md"
          src="/public/tiresizecharts/match/exact-circumference-of-tire.webp"
          :alt="t('guidesTireMatch.images.circumferenceAlt')"
          :caption="t('guidesTireMatch.images.circumferenceCaption')"
          :zoomOnClick="true"
        />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  tireCircumferenceRows,
  tireFrameClearanceTables,
} from '~/data/tireguides/match'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireMatch')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})
</script>

<style scoped>
.match-guide__measurement-reference {
  margin: 0 auto;
  max-width: 72rem;
  border: 1px solid rgba(5, 150, 105, 0.24);
  border-radius: 1rem;
  background: var(--tz-surface-subtle);
  padding: 1rem;
  text-align: left;
}

.match-guide__measurement-title {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 1rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.4;
  text-align: center;
  text-transform: uppercase;
}

.match-guide__measurement-grid {
  display: grid;
  gap: 1rem;
  margin-top: 1.25rem;
}

.match-guide__measurement-item {
  display: grid;
  grid-template-columns: 2rem minmax(0, 1fr);
  gap: 0.7rem;
  border-top: 1px solid rgba(5, 150, 105, 0.2);
  padding-top: 0.85rem;
}

.match-guide__measurement-badge {
  display: inline-flex;
  width: 2rem;
  height: 2rem;
  align-items: center;
  justify-content: center;
  border: 1px solid rgba(5, 150, 105, 0.32);
  border-radius: 0.5rem;
  background: rgba(5, 150, 105, 0.12);
  color: var(--tz-text-accent);
  font-size: 1rem;
  font-weight: 800;
  line-height: 1;
}

.match-guide__measurement-item h4 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.88rem;
  font-weight: 700;
  line-height: 1.35;
}

.match-guide__measurement-item p {
  margin: 0.35rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.55;
}

.match-guide__table-list {
  display: grid;
  gap: 1.5rem;
  text-align: left;
}

.match-guide__table-section {
  min-width: 0;
}

.match-guide__table-title {
  margin: 0 0 0.65rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  line-height: 1.35;
  text-transform: uppercase;
}

.match-guide__table-scroll {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.match-guide__table {
  min-width: 60rem;
  width: 100%;
  border-collapse: collapse;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.4;
  table-layout: fixed;
}

.match-guide__table--circumference {
  min-width: 38rem;
}

.match-guide__table caption {
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
  text-align: left;
}

.match-guide__table th,
.match-guide__table td {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  padding: 0.55rem 0.65rem;
  text-align: left;
  vertical-align: top;
  overflow-wrap: anywhere;
}

.match-guide__table thead th {
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1.35;
}

.match-guide__table tbody tr:hover {
  background: var(--tz-surface-subtle);
}

.match-guide__table tbody th {
  color: var(--tz-text-primary);
  font-weight: 700;
}

.match-guide__dimension-cell {
  color: var(--tz-text-secondary);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.match-guide__table th:nth-child(1),
.match-guide__table td:nth-child(1) {
  width: 12%;
}

.match-guide__table th:nth-child(2),
.match-guide__table td:nth-child(2) {
  width: 14%;
}

.match-guide__table th:nth-child(3),
.match-guide__table td:nth-child(3) {
  width: 22%;
}

.match-guide__table th:nth-child(4),
.match-guide__table td:nth-child(4),
.match-guide__table th:nth-child(5),
.match-guide__table td:nth-child(5) {
  width: 15%;
}

.match-guide__table th:nth-child(6),
.match-guide__table td:nth-child(6) {
  width: 22%;
}

.match-guide__table--circumference th:nth-child(1),
.match-guide__table--circumference td:nth-child(1) {
  width: 25%;
}

.match-guide__table--circumference th:nth-child(2),
.match-guide__table--circumference td:nth-child(2) {
  width: 30%;
}

.match-guide__table--circumference th:nth-child(3),
.match-guide__table--circumference td:nth-child(3) {
  width: 45%;
}

@media (min-width: 768px) {
  .match-guide__measurement-reference {
    padding: 1.25rem;
  }

  .match-guide__measurement-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1.25rem;
  }

  .match-guide__measurement-item {
    display: block;
    border-top-width: 2px;
    padding: 0.85rem 0.25rem 0;
  }

  .match-guide__measurement-item h4 {
    margin-top: 0.7rem;
  }

  .match-guide__table {
    min-width: 0;
  }
}
</style>
