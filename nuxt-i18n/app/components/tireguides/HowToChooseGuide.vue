<template>
  <div class="how-to-choose-guide">
    <!-- Main Premium Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-md md:p-6">
      <h2 class="mb-6 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
        {{ t('guidesTireChoose.title') }}
      </h2>

      <!-- Interactive Helper -->
      <div class="mb-8">
        <TireRimHelper />
      </div>

      <div class="mt-8 border-t tz-border-subtle pt-8">
        <div class="mb-6 flex items-center justify-center gap-2">
          <span class="h-px w-8 tz-surface-panel"></span>
          <h3 class="text-lg font-bold uppercase tracking-wider tz-text-primary">{{ t('guidesTireChoose.standards.title') }}</h3>
          <span class="h-px w-8 tz-surface-panel"></span>
        </div>

        <p class="mx-auto mb-3 max-w-3xl text-sm tz-text-secondary">
          {{ t('guidesTireChoose.standards.description') }}
        </p>

        <div class="mb-8 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-xs tz-text-muted">
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--recommended"></span>
            {{ t('guidesTireChoose.standards.recommended') }}
          </span>
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--possible"></span>
            {{ t('guidesTireChoose.standards.possible') }}
          </span>
        </div>
        <p class="mx-auto mb-6 max-w-3xl text-xs tz-text-muted">
          {{ t('guidesTireChoose.standards.legendDescription') }}
        </p>

        <div class="tire-chart-table-grid">
          <section class="tire-chart-table-panel" aria-labelledby="hookless-chart-table-title">
            <h4 id="hookless-chart-table-title" class="tire-chart-table-panel__title">
              {{ t('guidesTireChoose.standards.hooklessTitle') }}
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  {{ t('guidesTireChoose.standards.captionHookless') }}
                </caption>
                <thead>
                  <tr>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthMm') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthInch') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.recommendedRim') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.possibleRim') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hooklessChartRows" :key="`hookless-${row.tire_width_mm}`">
                    <th scope="row">{{ row.tire_width_mm }}</th>
                    <td>{{ row.inch }}</td>
                    <td class="tire-chart-table__recommended">
                      {{ formatWidthList(row.recommended) }}
                    </td>
                    <td class="tire-chart-table__possible">
                      {{ formatWidthList(row.possible) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>

          <section class="tire-chart-table-panel" aria-labelledby="hooked-chart-table-title">
            <h4 id="hooked-chart-table-title" class="tire-chart-table-panel__title">
              {{ t('guidesTireChoose.standards.hookedTitle') }}
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  {{ t('guidesTireChoose.standards.captionHooked') }}
                </caption>
                <thead>
                  <tr>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthMm') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.tireWidthInch') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.recommendedRim') }}</th>
                    <th scope="col">{{ t('guidesTireChoose.standards.headers.possibleRim') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hookedChartRows" :key="`hooked-${row.tire_width_mm}`">
                    <th scope="row">{{ row.tire_width_mm }}</th>
                    <td>{{ row.inch }}</td>
                    <td class="tire-chart-table__recommended">
                      {{ formatWidthList(row.recommended) }}
                    </td>
                    <td class="tire-chart-table__possible">
                      {{ formatWidthList(row.possible) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </section>
        </div>

        <h4 class="mt-10 mb-4 text-left text-sm font-semibold uppercase tracking-wider tz-text-primary">
          {{ t('guidesTireChoose.images.sourceTitle') }}
        </h4>
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <figure class="overflow-hidden rounded-xl border tz-border-subtle shadow-md transition-colors hover:tz-border-subtle">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
              :alt="t('guidesTireChoose.images.hooklessAlt')"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="tz-surface-panel py-2 text-xs tracking-wider tz-text-muted">
              {{ t('guidesTireChoose.images.hooklessCaption') }}
            </figcaption>
          </figure>
          <figure class="overflow-hidden rounded-xl border tz-border-subtle shadow-md transition-colors hover:tz-border-subtle">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
              :alt="t('guidesTireChoose.images.hookedAlt')"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="tz-surface-panel py-2 text-xs tracking-wider tz-text-muted">
              {{ t('guidesTireChoose.images.hookedCaption') }}
            </figcaption>
          </figure>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useAsyncData, useI18n } from '#imports'
import TireRimHelper from '~/components/TireRimHelper.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import {
  formatTireRimWidthReferenceRanges,
  type TireRimWidthRange,
  type TireRimWidthReferenceMatrixRow,
  type TireRimWidthReferenceMetadata,
} from '~/data/tireguides/tireRimWidthReferencePresentation'
import { useApiRequest } from '~/composables/useApiRequest'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTireChoose')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const { request } = useApiRequest()
const { data: tireRimWidthReferenceMetadata } = await useAsyncData<TireRimWidthReferenceMetadata>(
  'tire-rim-width-reference-metadata',
  async () => {
    const response = await request<{ data?: TireRimWidthReferenceMetadata }>(
      '/engineering/tire-rim/matrix',
      {},
      'Tire/rim reference matrix is temporarily unavailable',
    )
    if (!response.data) throw new Error('Tire/rim reference matrix response is missing data')
    return response.data
  },
  { default: () => ({ rows: [] } as unknown as TireRimWidthReferenceMetadata) },
)

const referenceRows = computed<TireRimWidthReferenceMatrixRow[]>(() => (
  tireRimWidthReferenceMetadata.value?.rows || []
))
const hooklessChartRows = computed(() => referenceRows.value.filter(row => row.rim_system === 'hookless'))
const hookedChartRows = computed(() => referenceRows.value.filter(row => row.rim_system === 'hooked'))

const formatWidthList = (widths: TireRimWidthRange[]) =>
  widths.length > 0
    ? `${formatTireRimWidthReferenceRanges(widths)} mm`
    : t('guidesTireChoose.helper.noneShown')
</script>

<style scoped>
.tire-chart-legend__swatch {
  display: inline-block;
  height: 0.75rem;
  width: 0.75rem;
  border-radius: 0.15rem;
}

.tire-chart-legend__swatch--recommended {
  background: var(--tz-text-accent);
}

.tire-chart-legend__swatch--possible {
  background: var(--tz-text-secondary);
}

.tire-chart-table-panel {
  min-width: 0;
}

.tire-chart-table-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.5rem;
  align-items: start;
  text-align: left;
}

.tire-chart-table-panel__title {
  margin: 0 0 0.65rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.tire-chart-table-scroll {
  max-width: 100%;
  overflow-x: auto;
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 0.5rem;
}

.tire-chart-table {
  min-width: 34rem;
  width: 100%;
  border-collapse: collapse;
  font-size: 0.78rem;
  table-layout: fixed;
}

.tire-chart-table caption {
  padding: 0.75rem 0.85rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1.45;
  text-align: left;
}

.tire-chart-table th,
.tire-chart-table td {
  border-top: 1px solid rgba(148, 163, 184, 0.14);
  padding: 0.5rem 0.6rem;
  text-align: left;
  vertical-align: top;
  overflow-wrap: anywhere;
}

.tire-chart-table thead th {
  background: rgba(15, 23, 42, 0.65);
  color: var(--tz-text-secondary);
  font-size: 0.68rem;
  font-weight: 700;
  line-height: 1.35;
}

.tire-chart-table th:nth-child(1),
.tire-chart-table td:nth-child(1) {
  width: 18%;
}

.tire-chart-table th:nth-child(2),
.tire-chart-table td:nth-child(2) {
  width: 17%;
}

.tire-chart-table th:nth-child(3),
.tire-chart-table td:nth-child(3) {
  width: 31%;
}

.tire-chart-table th:nth-child(4),
.tire-chart-table td:nth-child(4) {
  width: 34%;
}

.tire-chart-table tbody th {
  color: var(--tz-text-primary);
  font-weight: 700;
  white-space: nowrap;
}

.tire-chart-table tbody td {
  color: var(--tz-text-secondary);
}

.tire-chart-table__recommended {
  color: var(--tz-text-accent) !important;
  font-weight: 600;
}

.tire-chart-table__possible {
  color: var(--tz-text-secondary) !important;
}

@media (min-width: 1024px) {
  .tire-chart-table-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .tire-chart-table {
    min-width: 0;
  }

  .tire-chart-table caption {
    min-height: 3.35rem;
  }
}
</style>
