<template>
  <div class="space-y-5">
    <!-- 1. Interactive marking mapping calculator -->
    <TireSizeMarkingMappingCalculator
      :labels="{
        title: t('guidesTireSize.calculator.title'),
        description: t('guidesTireSize.calculator.description'),
        etrto: t('guidesTireSize.calculator.labels.etrto'),
        inch: t('guidesTireSize.calculator.labels.inch'),
        french: t('guidesTireSize.calculator.labels.french'),
        resultLabel: t('guidesTireSize.calculator.resultLabel'),
        note: t('guidesTireSize.calculator.note'),
      }"
      initial-etrto="37-622"
    />

    <!-- 2. Intro Card: Basics & Definitions -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] p-4 text-center shadow-md md:p-5">
      <div class="mb-4 flex justify-center">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-[var(--tz-action-primary)] px-6 py-2 text-xs font-bold uppercase tracking-wider text-white shadow-[0_4px_14px_rgba(15,23,42,0.16)] hover:bg-[var(--tz-action-primary-hover)] hover:shadow-[0_8px_22px_-6px_rgba(15,23,42,0.24)] hover:-translate-y-0.5 transition-all"
          @click="emit('openTireProducts')"
        >
          {{ t('guidesTireSize.actions.checkTires') }}
        </button>
      </div>

      <div class="mx-auto mb-5 w-full max-w-3xl">
        <GuideImage
          src="/public/tiresizecharts/tiresize/schwalbe-tiresize.webp"
          :alt="t('guidesTireSize.images.overviewAlt')"
          :zoomOnClick="true"
          class="rounded-xl overflow-hidden shadow-md"
        />
      </div>

      <div class="mb-5">
         <h3 class="mb-3 flex items-center justify-center gap-2 text-xl font-bold tz-text-secondary">
           {{ t('guidesTireSize.standards.title') }}
         </h3>
         <p class="mx-auto mb-4 max-w-2xl text-sm leading-relaxed tz-text-secondary">
           {{ t('guidesTireSize.standards.description') }}
         </p>

         <!-- Definitions Grid -->
         <div class="grid grid-cols-1 gap-3 text-left md:grid-cols-3 md:text-center">
            <!-- ETRTO -->
            <div class="tz-surface-panel flex flex-col items-center rounded-xl p-3 shadow-md">
               <strong class="mb-1.5 block text-[var(--tz-site-accent)] text-sm font-bold uppercase tracking-wider">{{ t('guidesTireSize.standards.etrto.name') }}</strong>
               <div class="mb-2 rounded px-2 py-1 text-xs font-mono tz-surface-panel tz-text-primary">{{ t('guidesTireSize.standards.etrto.example') }}</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 {{ t('guidesTireSize.standards.etrto.description') }}
               </p>
            </div>

            <!-- Inch -->
            <div class="tz-surface-panel flex flex-col items-center rounded-xl p-3 shadow-md">
               <strong class="mb-1.5 block text-amber-400 text-sm font-bold uppercase tracking-wider">{{ t('guidesTireSize.standards.inch.name') }}</strong>
               <div class="mb-2 rounded px-2 py-1 text-xs font-mono tz-surface-panel tz-text-primary">{{ t('guidesTireSize.standards.inch.example') }}</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 {{ t('guidesTireSize.standards.inch.description') }}
               </p>
            </div>

            <!-- French -->
            <div class="tz-surface-panel flex flex-col items-center rounded-xl p-3 shadow-md">
               <strong class="mb-1.5 block text-emerald-700 text-sm font-bold uppercase tracking-wider">{{ t('guidesTireSize.standards.french.name') }}</strong>
               <div class="mb-2 rounded px-2 py-1 text-xs font-mono tz-surface-panel tz-text-primary">{{ t('guidesTireSize.standards.french.example') }}</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 {{ t('guidesTireSize.standards.french.description') }}
               </p>
            </div>
         </div>

         <div class="mx-auto mt-4 max-w-3xl rounded-xl border border-[var(--tz-site-accent)]/30 bg-[var(--tz-surface-muted)] p-3 text-left">
           <strong class="block text-sm font-bold tz-text-primary mb-1">
             {{ t('guidesTireSize.standards.readFirst') }}
           </strong>
           <p class="text-sm tz-text-secondary leading-relaxed">
             {{ t('guidesTireSize.standards.readFirstDescription') }}
           </p>
         </div>
      </div>
    </div>

    <!-- 3. Comparison Table Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold tz-text-secondary mb-6 flex items-center justify-center gap-2">
         {{ t('guidesTireSize.comparison.title') }}
       </h3>
       <div class="overflow-x-auto rounded-xl tz-surface-panel shadow-md inline-block w-full max-w-4xl">
        <table class="min-w-full text-left text-xs sm:text-sm tz-text-secondary">
          <thead class="tz-surface-panel">
            <tr>
              <th class="px-4 py-3 font-bold tz-text-primary text-center">{{ t('guidesTireSize.comparison.headers.dimension') }}</th>
              <th class="px-4 py-3 font-bold text-[var(--tz-site-accent)] text-center">{{ t('guidesTireSize.comparison.headers.etrto') }}</th>
              <th class="px-4 py-3 font-bold text-amber-400 text-center">{{ t('guidesTireSize.comparison.headers.inch') }}</th>
              <th class="px-4 py-3 font-bold text-emerald-700 text-center">{{ t('guidesTireSize.comparison.headers.french') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 tz-surface-panel">
            <tr class="hover:tz-surface-panel transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">{{ t('guidesTireSize.comparison.rows.exampleSize') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">37-622</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">
                28 x 1.40<br />
                <span class="tz-caption tz-text-muted">{{ t('guidesTireSize.comparison.inchAltSize') }}</span>
              </td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">700 x 35C</td>
            </tr>
            <tr class="hover:tz-surface-panel transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">{{ t('guidesTireSize.comparison.rows.outerDiameter') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.etrtoFormat') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.inchDiameter') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.frenchDiameter') }}</td>
            </tr>
            <tr class="hover:tz-surface-panel transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">{{ t('guidesTireSize.comparison.rows.innerDiameter') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">622 mm</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.frenchCode') }}</td>
            </tr>
            <tr class="hover:tz-surface-panel transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">{{ t('guidesTireSize.comparison.rows.tireWidth') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.etrtoWidth') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">
                {{ t('guidesTireSize.comparison.inchWidth') }}<br />
                <span class="tz-caption tz-text-muted">{{ t('guidesTireSize.comparison.inchWidthAlt') }}</span>
              </td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.frenchWidth') }}</td>
            </tr>
            <tr class="hover:tz-surface-panel transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">{{ t('guidesTireSize.comparison.rows.tireHeight') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.etrtoReminder') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.inchHeight') }}</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">{{ t('guidesTireSize.comparison.frenchReminder') }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 4. Market Sizes Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-2 mb-4">
         <div class="h-px w-8 tz-surface-panel"></div>
         <h3 class="text-lg font-bold tz-text-secondary uppercase tracking-wider">{{ t('guidesTireSize.availability.title') }}</h3>
          <div class="h-px w-8 tz-surface-panel"></div>
       </div>
       <p class="mx-auto mb-6 max-w-3xl text-sm leading-relaxed tz-text-secondary">
         {{ t('guidesTireSize.availability.description') }}
       </p>

       <div class="tire-size-availability-grid">
        <div
           v-for="(column, columnIndex) in tireSizeMarkingMappingColumns"
          :key="`tire-size-availability-column-${columnIndex}`"
          class="tire-size-availability-column"
        >
          <details
            v-for="group in column"
            :key="group.label"
            class="tire-size-availability-card"
            :aria-labelledby="`tire-size-group-${group.key}`"
          >
            <summary class="tire-size-availability-card__header">
              <h4 :id="`tire-size-group-${group.key}`">{{ t(`guidesTireSize.availability.groups.${group.key}`) }}</h4>
              <span>
                {{
                  t(
                    group.rows.length === 1
                      ? 'guidesTireSize.availability.sizeCountSingular'
                      : 'guidesTireSize.availability.sizeCount',
                    { count: group.rows.length },
                  )
                }}
                <span class="tire-size-availability-card__chevron" aria-hidden="true"></span>
              </span>
            </summary>
            <div class="overflow-x-auto">
              <table class="tire-size-availability-table">
                <thead>
                  <tr>
                    <th scope="col">{{ t('guidesTireSize.availability.headers.etrto') }}</th>
                    <th scope="col">{{ t('guidesTireSize.availability.headers.inch') }}</th>
                    <th scope="col">{{ t('guidesTireSize.availability.headers.french') }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="row in group.rows" :key="`${group.key}-${row.etrto}-${row.inch}`">
                    <td>{{ row.etrto }}</td>
                    <td>{{ row.inch }}</td>
                    <td>{{ row.french || '-' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </details>
        </div>
      </div>
    </div>

    <!-- 5. 28 vs 29 Comparison Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-[var(--tz-site-accent)]">
       <h3 class="text-xl font-bold tz-text-secondary mb-6">
         {{ t('guidesTireSize.mystery.title') }}
       </h3>

       <div class="grid md:grid-cols-2 gap-8 items-center max-w-4xl mx-auto">
          <div class="tz-surface-panel rounded-xl p-5 shadow-md h-full flex flex-col justify-center">
             <strong class="block text-[var(--tz-site-accent)] text-3xl font-bold mb-2">{{ t('guidesTireSize.mystery.same') }}</strong>
             <span class="text-xs uppercase tracking-widest tz-text-muted mb-3">{{ t('guidesTireSize.mystery.innerDiameter') }}</span>
             <p class="tz-text-primary font-mono text-lg">622 mm</p>
             <p class="text-xs tz-text-muted mt-2 leading-relaxed">
                {{ t('guidesTireSize.mystery.sameDescription') }}
             </p>
          </div>

          <div class="text-left space-y-4">
             <div class="space-y-2">
                <strong class="text-[var(--tz-site-accent)] text-sm font-bold uppercase tracking-wider block">{{ t('guidesTireSize.mystery.contextTitle') }}</strong>
                <p class="text-sm tz-text-secondary leading-relaxed">
                   {{ t('guidesTireSize.mystery.contextBody') }}
                </p>
             </div>
             <div class="space-y-2">
                <strong class="text-rose-400 text-sm font-bold uppercase tracking-wider block">{{ t('guidesTireSize.mystery.realityTitle') }}</strong>
                <p class="text-sm tz-text-secondary leading-relaxed">
                   {{ t('guidesTireSize.mystery.realityBody') }}
                </p>
             </div>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import TireSizeMarkingMappingCalculator from '~/components/tireguides/TireSizeMarkingMappingCalculator.vue'
import { tireSizeMarkingMappingGroups } from '~/data/tireguides/tireSizeMarkingMappingData'

const { t } = useI18n()

const emit = defineEmits<{
  (e: 'openTireProducts'): void
}>()

const tireSizeMarkingMappingColumns = [
  tireSizeMarkingMappingGroups.slice(0, Math.ceil(tireSizeMarkingMappingGroups.length / 2)),
  tireSizeMarkingMappingGroups.slice(Math.ceil(tireSizeMarkingMappingGroups.length / 2)),
]

</script>

<style scoped>
.tire-size-availability-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: start;
  gap: 1rem;
  text-align: left;
}

.tire-size-availability-column {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 1rem;
}

.tire-size-availability-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  box-shadow: 0 4px 16px rgb(15 23 42 / 0.06);
}

.tire-size-availability-card[open] {
  box-shadow: 0 8px 22px rgb(15 23 42 / 0.1);
}

.tire-size-availability-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-muted);
  cursor: pointer;
  list-style: none;
  user-select: none;
}

.tire-size-availability-card__header::-webkit-details-marker {
  display: none;
}

.tire-size-availability-card__header h4 {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.82rem;
  font-weight: 800;
  letter-spacing: 0;
  text-transform: uppercase;
}

.tire-size-availability-card__header span {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  flex: 0 0 auto;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 700;
}

.tire-size-availability-card__chevron {
  width: 0.48rem;
  height: 0.48rem;
  border-right: 2px solid var(--tz-site-accent);
  border-bottom: 2px solid var(--tz-site-accent);
  transform: rotate(45deg) translateY(-0.08rem);
  transition: transform 0.2s ease;
}

.tire-size-availability-card[open] .tire-size-availability-card__chevron {
  transform: rotate(225deg) translate(-0.05rem, -0.05rem);
}

.tire-size-availability-table {
  width: 100%;
  min-width: 25rem;
  border-collapse: collapse;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.35;
}

.tire-size-availability-table thead {
  background: var(--tz-surface-muted);
}

.tire-size-availability-table th,
.tire-size-availability-table td {
  padding: 0.55rem 0.7rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  vertical-align: top;
}

.tire-size-availability-table th {
  color: var(--tz-text-primary);
  font-size: 0.7rem;
  font-weight: 800;
  text-transform: uppercase;
}

.tire-size-availability-table td:first-child {
  color: var(--tz-text-primary);
  font-family: var(--tz-font-ui);
  font-weight: 800;
  white-space: nowrap;
}

.tire-size-availability-table td:nth-child(2),
.tire-size-availability-table td:nth-child(3) {
  color: var(--tz-text-secondary);
  font-variant-numeric: tabular-nums;
}

.tire-size-availability-table tbody tr:last-child td {
  border-bottom: 0;
}

.tire-size-availability-table tbody tr:hover {
  background: var(--tz-surface-subtle);
}

@media (max-width: 767px) {
  .tire-size-availability-grid {
    grid-template-columns: 1fr;
  }

  .tire-size-availability-table {
    min-width: 22rem;
    font-size: 0.74rem;
  }

  .tire-size-availability-table th,
  .tire-size-availability-table td {
    padding: 0.5rem 0.58rem;
  }
}
</style>
