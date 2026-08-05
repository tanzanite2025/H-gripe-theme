<template>
  <div class="how-to-choose-guide">
    <!-- Main Premium Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-5 text-center shadow-[0_8px_30px_rgba(0,0,0,0.6)] md:p-6">
      <h2 class="mb-6 flex items-center justify-center gap-2 text-xl font-bold text-slate-100">
        How to Choose the Right Size
      </h2>

      <!-- Interactive Helper -->
      <div class="mb-8">
        <TireRimHelper />
      </div>

      <div class="mt-8 border-t border-slate-800 pt-8">
        <div class="mb-6 flex items-center justify-center gap-2">
          <span class="h-px w-8 bg-slate-700"></span>
          <h3 class="text-lg font-bold uppercase tracking-wider tz-text-primary">Manufacturer Standards</h3>
          <span class="h-px w-8 bg-slate-700"></span>
        </div>

        <p class="mx-auto mb-3 max-w-3xl text-sm tz-text-secondary">
          The following hookless (TSS) and hooked (TC) rim inner width charts from DT Swiss show
          recommended (dark) and possible (light) combinations. The HTML tables below transcribe
          the chart rows for search engines and screen readers. Always stay within the limits
          specified by your tire and rim manufacturers.
        </p>

        <div class="mb-8 flex flex-wrap items-center justify-center gap-x-5 gap-y-2 text-xs tz-text-muted">
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--recommended"></span>
            Recommended combination
          </span>
          <span class="inline-flex items-center gap-2">
            <span class="tire-chart-legend__swatch tire-chart-legend__swatch--possible"></span>
            Possible combination
          </span>
        </div>
        <p class="mx-auto mb-6 max-w-3xl text-xs tz-text-muted">
          Recommended combinations are the preferred tire and rim pairings in the chart. Possible
          combinations are allowed chart pairings outside the preferred range.
        </p>

        <div class="tire-chart-table-grid">
          <section class="tire-chart-table-panel" aria-labelledby="hookless-chart-table-title">
            <h4 id="hookless-chart-table-title" class="tire-chart-table-panel__title">
              Hookless / Tubeless Straight Side (TSS)
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  DT Swiss hookless TSS recommended and possible tire width to rim inner width
                  combinations
                </caption>
                <thead>
                  <tr>
                    <th scope="col">Tire width (mm)</th>
                    <th scope="col">Tire width (inch)</th>
                    <th scope="col">Recommended rim inner width (mm)</th>
                    <th scope="col">Possible rim inner width (mm)</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hooklessChartRows" :key="`hookless-${row.tire}`">
                    <th scope="row">{{ row.tire }}</th>
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
              Hooked / Tubeless Crotchet (TC)
            </h4>
            <div class="tire-chart-table-scroll">
              <table class="tire-chart-table">
                <caption>
                  DT Swiss hooked TC recommended and possible tire width to rim inner width
                  combinations
                </caption>
                <thead>
                  <tr>
                    <th scope="col">Tire width (mm)</th>
                    <th scope="col">Tire width (inch)</th>
                    <th scope="col">Recommended rim inner width (mm)</th>
                    <th scope="col">Possible rim inner width (mm)</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in hookedChartRows" :key="`hooked-${row.tire}`">
                    <th scope="row">{{ row.tire }}</th>
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
          Source chart images
        </h4>
        <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
          <figure class="overflow-hidden rounded-xl border border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)] transition-colors hover:border-slate-700">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
              alt="DT Swiss hookless TSS rim inner width recommendation chart"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="bg-slate-900/50 py-2 text-xs tracking-wider tz-text-muted">
              HOOKLESS (TSS)
            </figcaption>
          </figure>
          <figure class="overflow-hidden rounded-xl border border-slate-800/50 shadow-[0_4px_16px_rgba(0,0,0,0.5)] transition-colors hover:border-slate-700">
            <img
              src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
              alt="DT Swiss hooked TC rim inner width recommendation chart"
              class="block h-auto w-full"
              loading="lazy"
            />
            <figcaption class="bg-slate-900/50 py-2 text-xs tracking-wider tz-text-muted">
              HOOKED (TC)
            </figcaption>
          </figure>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import TireRimHelper from '~/components/TireRimHelper.vue'

interface ChartRow {
  tire: number
  inch: string
  recommended: string[]
  possible: string[]
}

const hooklessChartRows: ChartRow[] = [
  { tire: 30, inch: '1.20', recommended: [], possible: ['23-25'] },
  { tire: 32, inch: '1.25', recommended: ['23-25'], possible: [] },
  { tire: 34, inch: '1.35', recommended: ['23-25'], possible: [] },
  { tire: 36, inch: '1.40', recommended: ['23-25'], possible: ['26-27'] },
  { tire: 38, inch: '1.50', recommended: ['23-25'], possible: ['26-27'] },
  { tire: 41, inch: '1.60', recommended: ['23-25'], possible: ['26-27'] },
  { tire: 43, inch: '1.70', recommended: ['23-25', '26-27'], possible: [] },
  { tire: 47, inch: '1.85', recommended: ['23-25', '26-27'], possible: ['28-30'] },
  { tire: 50, inch: '2.00', recommended: ['23-25', '26-27'], possible: ['28-30'] },
  { tire: 52, inch: '2.00', recommended: ['26-27'], possible: ['23-25', '28-30'] },
  { tire: 53, inch: '2.10', recommended: ['26-27', '28-30'], possible: ['23-25'] },
  { tire: 56, inch: '2.20', recommended: ['26-27', '28-30'], possible: ['23-25'] },
  { tire: 60, inch: '2.40', recommended: ['26-27', '28-30'], possible: ['23-25', '31-35'] },
  { tire: 64, inch: '2.50', recommended: ['28-30', '31-35'], possible: ['23-25', '26-27'] },
  { tire: 66, inch: '2.60', recommended: ['28-30', '31-35'], possible: ['23-25', '26-27', '36-40'] },
  { tire: 69, inch: '2.70', recommended: ['28-30', '31-35'], possible: ['26-27', '36-40'] },
  { tire: 71, inch: '2.80', recommended: ['31-35', '36-40'], possible: ['26-27', '28-30'] },
  { tire: 75, inch: '3.00', recommended: ['36-40'], possible: ['31-35'] },
  { tire: 85, inch: '3.30', recommended: ['36-40'], possible: ['31-35'] },
  { tire: 96, inch: '3.70', recommended: [], possible: ['76'] },
  { tire: 102, inch: '4.00', recommended: ['76'], possible: [] },
  { tire: 110, inch: '4.30', recommended: ['76'], possible: [] },
  { tire: 115, inch: '4.50', recommended: ['76'], possible: [] },
  { tire: 122, inch: '4.80', recommended: ['76'], possible: [] },
  { tire: 127, inch: '5.00', recommended: ['76'], possible: [] },
]

const hookedChartRows: ChartRow[] = [
  { tire: 20, inch: '0.80', recommended: [], possible: ['15-17'] },
  { tire: 23, inch: '0.90', recommended: ['15-17'], possible: ['18-20'] },
  { tire: 25, inch: '1.00', recommended: ['15-17', '18-20'], possible: ['21-22'] },
  { tire: 26, inch: '1.00', recommended: ['18-20', '21-22'], possible: ['15-17'] },
  { tire: 28, inch: '1.10', recommended: ['18-20', '21-22'], possible: ['15-17'] },
  { tire: 29, inch: '1.15', recommended: ['18-20', '21-22'], possible: ['15-17', '23-25'] },
  { tire: 30, inch: '1.20', recommended: ['18-20', '21-22'], possible: ['15-17', '23-25'] },
  { tire: 32, inch: '1.25', recommended: ['21-22', '23-25'], possible: ['15-17', '18-20'] },
  { tire: 34, inch: '1.35', recommended: ['21-22', '23-25'], possible: ['15-17', '18-20'] },
  { tire: 35, inch: '1.40', recommended: ['21-22', '23-25'], possible: ['15-17', '18-20', '26-27'] },
  { tire: 36, inch: '1.40', recommended: ['23-25'], possible: ['18-20', '21-22', '26-27'] },
  { tire: 38, inch: '1.50', recommended: ['23-25'], possible: ['18-20', '21-22', '26-27'] },
  { tire: 41, inch: '1.60', recommended: ['23-25'], possible: ['18-20', '21-22', '26-27'] },
  { tire: 43, inch: '1.70', recommended: ['23-25', '26-27'], possible: ['18-20', '21-22'] },
  { tire: 47, inch: '1.85', recommended: ['23-25', '26-27'], possible: ['18-20', '21-22', '28-30'] },
  { tire: 50, inch: '2.00', recommended: ['23-25', '26-27'], possible: ['18-20', '21-22', '28-30'] },
  { tire: 52, inch: '2.00', recommended: ['26-27'], possible: ['18-20', '21-22', '23-25', '28-30'] },
  { tire: 53, inch: '2.10', recommended: ['26-27', '28-30'], possible: ['18-20', '21-22', '23-25'] },
  { tire: 56, inch: '2.20', recommended: ['26-27', '28-30'], possible: ['18-20', '21-22', '23-25'] },
  { tire: 60, inch: '2.40', recommended: ['26-27', '28-30'], possible: ['21-22', '23-25', '31-35'] },
  { tire: 64, inch: '2.50', recommended: ['28-30', '31-35'], possible: ['21-22', '23-25', '26-27'] },
  { tire: 66, inch: '2.60', recommended: ['28-30', '31-35'], possible: ['23-25', '26-27', '36-40'] },
  { tire: 69, inch: '2.70', recommended: ['28-30', '31-35'], possible: ['26-27', '36-40'] },
  { tire: 71, inch: '2.80', recommended: ['31-35', '36-40'], possible: ['26-27', '28-30'] },
  { tire: 75, inch: '3.00', recommended: ['36-40'], possible: ['31-35'] },
  { tire: 85, inch: '3.30', recommended: ['36-40'], possible: ['31-35'] },
  { tire: 96, inch: '3.70', recommended: [], possible: ['76'] },
  { tire: 102, inch: '4.00', recommended: ['76'], possible: [] },
  { tire: 110, inch: '4.30', recommended: ['76'], possible: [] },
  { tire: 115, inch: '4.50', recommended: ['76'], possible: [] },
  { tire: 122, inch: '4.80', recommended: ['76'], possible: [] },
  { tire: 127, inch: '5.00', recommended: ['76'], possible: [] },
]

const formatWidthList = (widths: string[]) =>
  widths.length > 0 ? widths.map((width) => `${width} mm`).join(', ') : 'None shown'
</script>

<style scoped>
.tire-chart-legend__swatch {
  display: inline-block;
  height: 0.75rem;
  width: 0.75rem;
  border-radius: 0.15rem;
}

.tire-chart-legend__swatch--recommended {
  background: #9b9b9b;
}

.tire-chart-legend__swatch--possible {
  background: #cecece;
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
  color: #d7ffbc !important;
  font-weight: 600;
}

.tire-chart-table__possible {
  color: #d4d9de !important;
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
