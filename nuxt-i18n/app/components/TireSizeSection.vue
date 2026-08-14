<template>
  <div class="space-y-8">
    <!-- 1. Intro Card: Basics & Definitions -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
      <div class="mb-6 flex justify-center">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-[var(--tz-brand-primary)] px-6 py-2 text-xs font-bold uppercase tracking-wider text-slate-950 shadow-[0_4px_14px_rgba(181,255,109,0.22)] hover:shadow-[0_8px_22px_-6px_rgba(181,255,109,0.45)] hover:-translate-y-0.5 transition-all"
          @click="emit('openTireProducts')"
        >
          Check out tires
        </button>
      </div>

      <div class="w-full mb-8 max-w-4xl mx-auto">
        <GuideImage
          src="/public/tiresizecharts/tiresize/schwalbe-tiresize.webp"
          alt="Carbon and Schwalbe tire size chart overview"
          :zoomOnClick="true"
          class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
        />
      </div>

      <div class="mb-8">
         <h3 class="text-xl font-bold text-slate-100 mb-4 flex items-center justify-center gap-2">
           Tire Size Standards
         </h3>
         <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-2xl mx-auto">
           Bicycle tires use various sizing standards. Understanding the difference between ETRTO, Inch, and French notations is key to finding the perfect fit.
         </p>

         <!-- Definitions Grid -->
         <div class="grid grid-cols-1 md:grid-cols-3 gap-4 text-left md:text-center">
            <!-- ETRTO -->
            <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
               <strong class="block text-[var(--tz-brand-primary)] text-sm font-bold uppercase tracking-wider mb-2">ETRTO (ISO)</strong>
               <div class="text-xs tz-text-primary font-mono bg-slate-900/50 px-2 py-1 rounded mb-3">e.g. 37-622</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 The modern European standard. Indicates <strong>Width (37mm)</strong> and <strong>Inner Diameter (622mm)</strong>. This is the most unambiguous and recommended method.
               </p>
            </div>

            <!-- Inch -->
            <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
               <strong class="block text-amber-400 text-sm font-bold uppercase tracking-wider mb-2">Inch (Imperial)</strong>
               <div class="text-xs tz-text-primary font-mono bg-slate-900/50 px-2 py-1 rounded mb-3">e.g. 28 x 1.40</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 Traditional notation. Indicates approx. <strong>Outer Diameter (28")</strong> and <strong>Width (1.40")</strong>. Often imprecise (e.g., 26" can refer to multiple different diameters).
               </p>
            </div>

            <!-- French -->
            <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
               <strong class="block text-emerald-300 text-sm font-bold uppercase tracking-wider mb-2">French</strong>
               <div class="text-xs tz-text-primary font-mono bg-slate-900/50 px-2 py-1 rounded mb-3">e.g. 700 x 35C</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 Indicates approx. <strong>Outer Diameter (700mm)</strong> and <strong>Width (35mm)</strong>. The letter (C) hints at the inner diameter (622mm). Common in Road/Gravel.
               </p>
            </div>
         </div>
      </div>
    </div>

    <!-- 2. Comparison Table Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold text-slate-100 mb-6 flex items-center justify-center gap-2">
         Notation Comparison Table
       </h3>
       <div class="overflow-x-auto rounded-xl bg-slate-950/40 shadow-[0_4px_16px_rgba(0,0,0,0.5)] inline-block w-full max-w-4xl">
        <table class="min-w-full text-left text-xs sm:text-sm tz-text-secondary">
          <thead class="bg-slate-900/80">
            <tr>
              <th class="px-4 py-3 font-bold tz-text-primary text-center">Dimension</th>
              <th class="px-4 py-3 font-bold text-[var(--tz-brand-primary)] text-center">ETRTO</th>
              <th class="px-4 py-3 font-bold text-amber-400 text-center">Inch</th>
              <th class="px-4 py-3 font-bold text-emerald-300 text-center">French</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800 bg-slate-800/20">
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Example Size</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">37-622</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">
                28 x 1.40<br />
                <span class="tz-caption tz-text-muted">or 28 x 1 5/8 x 1 3/8</span>
              </td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">700 x 35C</td>
            </tr>
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Outer Diameter</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">~ 28 Inch</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">~ 700 mm</td>
            </tr>
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Inner Diameter</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">622 mm</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
            </tr>
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Tire Width</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">~ 37 mm</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">
                ~ 1 3/8 Inch<br />
                <span class="tz-caption tz-text-muted">(1.40 Inch)</span>
              </td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">~ 35 mm</td>
            </tr>
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Tire Height</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">~ 1 5/8 Inch</td>
              <td class="px-4 py-3 text-center tz-text-muted">-</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 3. Market Sizes Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-2 mb-4">
         <div class="h-px w-8 bg-slate-700"></div>
         <h3 class="text-lg font-bold text-slate-100 uppercase tracking-wider">WHAT TIRE SIZES ARE AVAILABLE</h3>
         <div class="h-px w-8 bg-slate-700"></div>
      </div>

      <div class="tire-size-availability-grid">
        <div
          v-for="(column, columnIndex) in tireSizeAvailabilityColumns"
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
              <h4 :id="`tire-size-group-${group.key}`">{{ group.label }}</h4>
              <span>
                {{ group.rows.length }} size{{ group.rows.length === 1 ? '' : 's' }}
                <span class="tire-size-availability-card__chevron" aria-hidden="true"></span>
              </span>
            </summary>
            <div class="overflow-x-auto">
              <table class="tire-size-availability-table">
                <thead>
                  <tr>
                    <th scope="col">ETRTO</th>
                    <th scope="col">Inch size</th>
                    <th scope="col">French / alt.</th>
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

    <!-- 4. 28 vs 29 Comparison Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center border-t-4 border-[var(--tz-brand-primary)]">
       <h3 class="text-xl font-bold text-slate-100 mb-6">
         The "28 Inch" vs "29 Inch" Mystery
       </h3>

       <div class="grid md:grid-cols-2 gap-8 items-center max-w-4xl mx-auto">
          <div class="bg-slate-900/55 rounded-xl p-5 shadow-[0_4px_16px_rgba(0,0,0,0.5)] h-full flex flex-col justify-center">
             <strong class="block text-[var(--tz-brand-primary)] text-3xl font-bold mb-2">SAME</strong>
             <span class="text-xs uppercase tracking-widest tz-text-muted mb-3">Inner Diameter</span>
             <p class="tz-text-primary font-mono text-lg">622 mm</p>
             <p class="text-xs tz-text-muted mt-2 leading-relaxed">
                Both 28" (Touring/Road) and 29" (MTB) tires fit on the same 622mm rim diameter.
             </p>
          </div>

          <div class="text-left space-y-4">
             <div class="space-y-2">
                <strong class="text-[var(--tz-brand-primary)] text-sm font-bold uppercase tracking-wider block">History & Context</strong>
                <p class="text-sm tz-text-secondary leading-relaxed">
                   <strong>28 Inches:</strong> The traditional European size for touring bikes.
                   <br>
                   <strong>29 Inches:</strong> A marketing term introduced in the US for voluminous MTB tires that, when mounted, have an approx. outer diameter of 29".
                </p>
             </div>
             <div class="space-y-2">
                <strong class="text-rose-400 text-sm font-bold uppercase tracking-wider block">Reality Check</strong>
                <p class="text-sm tz-text-secondary leading-relaxed">
                   Both designations are imprecise. A narrow 28" road tire (23mm) is actually ~26" outer diameter. A wide MTB tire (60mm+) is almost 30".
                </p>
             </div>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

const emit = defineEmits<{
  (e: 'openTireProducts'): void
}>()

const tireSizeAvailabilityGroups = [
  {
    key: '12-inch',
    label: '12 inch',
    rows: [
      { etrto: '47-203', inch: '12 x 1.75', french: '' },
      { etrto: '50-203', inch: '12 x 2.00', french: '' },
      { etrto: '54-203', inch: '12 x 1.95', french: '' },
      { etrto: '57-203', inch: '12.5 x 2.25', french: '' },
      { etrto: '62-203', inch: '12.5 x 2.25', french: '' },
    ],
  },
  {
    key: '14-inch',
    label: '14 inch',
    rows: [
      { etrto: '47-254', inch: '14 x 1.75 / 1.90', french: '' },
      { etrto: '50-254', inch: '14 x 2.00', french: '' },
      { etrto: '40-279', inch: '14 x 1.50', french: '350 x 38B' },
      { etrto: '37-288', inch: '14 x 1.375', french: '350 x 35A' },
      { etrto: '40-288', inch: '14 x 1.50', french: '350 x 38A' },
      { etrto: '44-288', inch: '14 x 1.625', french: '350 x 42A' },
      { etrto: '32-298', inch: '14 x 1.25', french: '350 x 32A' },
      { etrto: '47-288', inch: '14 x 1.75', french: '' },
    ],
  },
  {
    key: '16-inch',
    label: '16 inch',
    rows: [
      { etrto: '40-305', inch: '16 x 1.50', french: '' },
      { etrto: '47-305', inch: '16 x 1.75 / 1.90', french: '' },
      { etrto: '50-305', inch: '16 x 2.00', french: '' },
      { etrto: '54-305', inch: '16 x 1.95 / 2.00', french: '' },
      { etrto: '57-305', inch: '16 x 2.125', french: '' },
      { etrto: '28-349', inch: '16 x 1.125', french: '' },
      { etrto: '30-349', inch: '16 x 1.20', french: '' },
      { etrto: '32-349', inch: '16 x 1.25', french: '' },
      { etrto: '35-349', inch: '16 x 1.35', french: '' },
      { etrto: '37-349', inch: '16 x 1.375', french: '' },
      { etrto: '40-330', inch: '16 x 1.50', french: '400 x 38B' },
      { etrto: '28-340', inch: '', french: '400 x 30A' },
      { etrto: '32-340', inch: '16 x 1.25', french: '400 x 32A' },
      { etrto: '37-340', inch: '16 x 1.375', french: '400 x 35A' },
      { etrto: '44-340', inch: '16 x 1.75', french: '' },
    ],
  },
  {
    key: '17-inch',
    label: '17 inch',
    rows: [
      { etrto: '32-357', inch: '17 x 1.25', french: '' },
      { etrto: '32-369', inch: '17 x 1.25', french: '' },
    ],
  },
  {
    key: '18-inch',
    label: '18 inch',
    rows: [
      { etrto: '28-355', inch: '18 x 1.125', french: '' },
      { etrto: '32-355', inch: '18 x 1.25', french: '' },
      { etrto: '35-355', inch: '18 x 1.35', french: '' },
      { etrto: '40-355', inch: '18 x 1.50', french: '' },
      { etrto: '42-355', inch: '18 x 1.60', french: '' },
      { etrto: '47-355', inch: '18 x 1.75 / 1.90', french: '' },
      { etrto: '50-355', inch: '18 x 2.00', french: '' },
      { etrto: '37-387', inch: '18 x 1.50', french: '' },
      { etrto: '28-390', inch: '18 x 1.125', french: '450 x 28A' },
      { etrto: '37-390', inch: '18 x 1.375', french: '450 x 35A' },
      { etrto: '55-390', inch: '', french: '450 x 55A' },
      { etrto: '57-390', inch: '', french: '450 x 55A' },
      { etrto: '37-400', inch: '18 x 1.375', french: '' },
    ],
  },
  {
    key: '20-inch',
    label: '20 inch',
    rows: [
      { etrto: '54-400', inch: '20 x 2 x 1.75', french: '' },
      { etrto: '23-406', inch: '20 x 0.90', french: '' },
      { etrto: '25-406', inch: '20 x 1.00', french: '' },
      { etrto: '28-406', inch: '20 x 1.10', french: '' },
      { etrto: '32-406', inch: '20 x 1.25', french: '' },
      { etrto: '35-406', inch: '20 x 1.35', french: '' },
      { etrto: '37-406', inch: '20 x 1.40', french: '' },
      { etrto: '40-406', inch: '20 x 1.50', french: '' },
      { etrto: '42-406', inch: '20 x 1.60', french: '' },
      { etrto: '44-406', inch: '20 x 1.50 / 1.625', french: '' },
      { etrto: '47-406', inch: '20 x 1.75 / 1.90', french: '' },
      { etrto: '50-406', inch: '20 x 2.00', french: '' },
      { etrto: '54-406', inch: '20 x 2.10 / 2.00', french: '' },
      { etrto: '57-406', inch: '20 x 2.25 / 2.125', french: '' },
      { etrto: '60-406', inch: '20 x 2.35', french: '' },
      { etrto: '62-406', inch: '20 x 2.40', french: '' },
      { etrto: '65-406', inch: '20 x 2.60', french: '' },
      { etrto: '70-406', inch: '20 x 2.80', french: '' },
      { etrto: '100-406', inch: '20 x 4.00', french: '' },
      { etrto: '54-428', inch: '20 x 2.00', french: '' },
      { etrto: '40-432', inch: '20 x 1.50', french: '' },
      { etrto: '37-438', inch: '20 x 1.375', french: '' },
      { etrto: '40-438', inch: '20 x 1.375 x 1.50', french: '' },
      { etrto: '28-440', inch: '', french: '500 x 28A' },
      { etrto: '37-440', inch: '', french: '500 x 35A' },
      { etrto: '40-440', inch: '20 x 1.50 NL', french: '500 x 38A' },
      { etrto: '23-451', inch: '20 x 0.90', french: '' },
      { etrto: '25-451', inch: '20 x 1.00', french: '' },
      { etrto: '28-451', inch: '20 x 1.125', french: '' },
      { etrto: '37-451', inch: '20 x 1.375 B.S.', french: '' },
    ],
  },
  {
    key: '22-inch',
    label: '22 inch',
    rows: [
      { etrto: '44-457', inch: '22 x 1.75', french: '' },
      { etrto: '44-484', inch: '22 x 1.50', french: '' },
      { etrto: '25-489', inch: '22 x 1.00', french: '' },
      { etrto: '37-489', inch: '22 x 1.375 NL', french: '' },
      { etrto: '40-489', inch: '22 x 1.375 x 1.50', french: '' },
      { etrto: '50-489', inch: '22 x 2.00', french: '' },
      { etrto: '28-490', inch: '', french: '550 x 28A' },
      { etrto: '32-490', inch: '22 x 1.375 x 1.25', french: '550 x 32A' },
      { etrto: '37-490', inch: '22 x 1.375', french: '550 x 35A' },
      { etrto: '47-498', inch: '22 x 1.375 x 1.25', french: '' },
      { etrto: '25-501', inch: '22 x 1.00 B.S.', french: '' },
      { etrto: '32-501', inch: '22 x 1.25', french: '' },
      { etrto: '37-501', inch: '22 x 1.375', french: '' },
    ],
  },
  {
    key: '24-inch',
    label: '24 inch',
    rows: [
      { etrto: '40-507', inch: '24 x 1.50', french: '' },
      { etrto: '44-507', inch: '24 x 1.625 / 1.75', french: '' },
      { etrto: '47-507', inch: '24 x 1.75 / 1.85 / 1.90', french: '' },
      { etrto: '50-507', inch: '24 x 2.00 / 1.90 / 2.125', french: '' },
      { etrto: '54-507', inch: '24 x 2.10', french: '' },
      { etrto: '55-507', inch: '24 x 2.15', french: '' },
      { etrto: '57-507', inch: '24 x 2.25 / 2.125', french: '' },
      { etrto: '60-507', inch: '24 x 2.35', french: '' },
      { etrto: '62-507', inch: '24 x 2.40', french: '' },
      { etrto: '65-507', inch: '24 x 2.60', french: '' },
      { etrto: '23-520', inch: '24 x 0.90', french: '' },
      { etrto: '44-531', inch: '24 x 1.375 x 1.50', french: '' },
      { etrto: '40-534', inch: '24 x 1.50', french: '' },
      { etrto: '20-540', inch: '24 x 0.75', french: '' },
      { etrto: '23-540', inch: '24 x 0.90', french: '' },
      { etrto: '25-540', inch: '24 x 1.00', french: '' },
      { etrto: '30-540', inch: '24 x 1.20', french: '' },
      { etrto: '32-540', inch: '24 x 1.25', french: '' },
      { etrto: '37-540', inch: '24 x 1.375', french: '' },
      { etrto: '40-540', inch: '24 x 1.375 x 1.50', french: '' },
      { etrto: '25-541', inch: '', french: '600 x 25A' },
      { etrto: '28-541', inch: '', french: '600 x 28A' },
      { etrto: '32-541', inch: '24 x 1.375 x 1.25 NL', french: '600 x 32A' },
      { etrto: '37-541', inch: '', french: '600 x 35A' },
    ],
  },
  {
    key: '25-inch',
    label: '25 inch',
    rows: [
      { etrto: '57-520', inch: '25 x 2.25', french: '' },
    ],
  },
  {
    key: '26-inch',
    label: '26 inch',
    rows: [
      { etrto: '20-559', inch: '26 x 0.75', french: '' },
      { etrto: '23-559', inch: '26 x 0.90', french: '' },
      { etrto: '25-559', inch: '26 x 1.00', french: '' },
      { etrto: '28-559', inch: '26 x 1.10', french: '' },
      { etrto: '30-559', inch: '26 x 1.20', french: '' },
      { etrto: '32-559', inch: '26 x 1.25', french: '' },
      { etrto: '35-559', inch: '26 x 1.35', french: '' },
      { etrto: '37-559', inch: '26 x 1.40', french: '' },
      { etrto: '40-559', inch: '26 x 1.50', french: '' },
      { etrto: '42-559', inch: '26 x 1.60', french: '' },
      { etrto: '44-559', inch: '26 x 1.625 / 1.75', french: '' },
      { etrto: '47-559', inch: '26 x 1.75 / 1.80 / 1.85 / 1.90', french: '' },
      { etrto: '50-559', inch: '26 x 2.00 / 1.90 / 1.95', french: '' },
      { etrto: '54-559', inch: '26 x 2.10 / 1.95 / 2.125', french: '' },
      { etrto: '55-559', inch: '26 x 2.15', french: '' },
      { etrto: '57-559', inch: '26 x 2.25 / 2.125 / 2.20', french: '' },
      { etrto: '60-559', inch: '26 x 2.35', french: '' },
      { etrto: '62-559', inch: '26 x 2.40 / 2.50', french: '' },
      { etrto: '64-559', inch: '26 x 2.50', french: '' },
      { etrto: '65-559', inch: '26 x 2.60', french: '' },
      { etrto: '70-559', inch: '26 x 2.75', french: '' },
      { etrto: '75-559', inch: '26 x 3.00', french: '' },
      { etrto: '95-559', inch: '26 x 3.70 / 3.80', french: '' },
      { etrto: '100-559', inch: '26 x 4.00', french: '' },
      { etrto: '110-559', inch: '26 x 4.40', french: '' },
      { etrto: '115-559', inch: '26 x 4.50', french: '' },
      { etrto: '120-559', inch: '26 x 4.80', french: '' },
      { etrto: '20-571', inch: '26 x 0.75', french: '' },
      { etrto: '23-571', inch: '26 x 0.90', french: '650 x 23C' },
      { etrto: '40-571', inch: '26 x 1.50 CS', french: '650 x 38C' },
      { etrto: '47-571', inch: '26 x 1.75', french: '650 x 45C' },
      { etrto: '54-571', inch: '26 x 2 x 1.75', french: '650 x 50C' },
      { etrto: '20-590', inch: '26 x 0.75', french: '650 x 20A' },
      { etrto: '25-590', inch: '26 x 1.00', french: '650 x 25A' },
      { etrto: '32-590', inch: '26 x 1.375 x 1.25', french: '650 x 32A' },
      { etrto: '37-590', inch: '26 x 1.375', french: '650 x 35A' },
      { etrto: '40-590', inch: '26 x 1.50 / 1.375 x 1.50', french: '650 x 38A' },
      { etrto: '42-590', inch: '26 x 1.625', french: '650 x 40A' },
      { etrto: '28-590', inch: '26 x 1.125', french: '650 x 28A' },
      { etrto: '32-597', inch: '26 x 1.25', french: '' },
    ],
  },
  {
    key: '27-inch',
    label: '27 inch',
    rows: [
      { etrto: '40-609', inch: '27 x 1.50', french: '' },
      { etrto: '20-630', inch: '27 x 0.75', french: '' },
      { etrto: '22-630', inch: '27 x 0.875', french: '' },
      { etrto: '25-630', inch: '27 x 1.00', french: '' },
      { etrto: '28-630', inch: '27 x 1.125', french: '' },
      { etrto: '32-630', inch: '27 x 1.25', french: '' },
      { etrto: '35-630', inch: '27 x 1.375', french: '' },
    ],
  },
  {
    key: '27-5-inch',
    label: '27.5 inch / 650B',
    rows: [
      { etrto: '28-584', inch: '26 x 1.375 x 1.50', french: '650 x 28B' },
      { etrto: '32-584', inch: '', french: '650 x 32B' },
      { etrto: '35-584', inch: '27.5 x 1.35', french: '650B / 650 x 35B' },
      { etrto: '37-584', inch: '27.5 x 1.40', french: '650B / 650 x 35B' },
      { etrto: '40-584', inch: '27.5 x 1.50', french: '650B / 650 x 38B' },
      { etrto: '44-584', inch: '27.5 x 1.65', french: '650B / 650 x 42B' },
      { etrto: '47-584', inch: '27.5 x 1.75', french: '650B' },
      { etrto: '50-584', inch: '27.5 x 2.00', french: '650B' },
      { etrto: '54-584', inch: '27.5 x 2.10', french: '650B' },
      { etrto: '57-584', inch: '27.5 x 2.25', french: '650B' },
      { etrto: '60-584', inch: '27.5 x 2.35', french: '650B' },
      { etrto: '62-584', inch: '27.5 x 2.40', french: '650B' },
      { etrto: '64-584', inch: '27.5 x 2.50', french: '650B' },
      { etrto: '65-584', inch: '27.5 x 2.60', french: '650B' },
      { etrto: '70-584', inch: '27.5 x 2.75', french: '650B' },
      { etrto: '74-584', inch: '27.5 x 2.90', french: '650B' },
      { etrto: '75-584', inch: '27.5 x 3.00', french: '650B' },
    ],
  },
  {
    key: '28-inch',
    label: '28 inch / 700C',
    rows: [
      { etrto: '18-622', inch: '28 x 0.75', french: '700 x 18C' },
      { etrto: '19-622', inch: '', french: '700 x 19C / 700 x 19' },
      { etrto: '20-622', inch: '28 x 0.75', french: '700 x 20C' },
      { etrto: '22-622', inch: '28 x 0.875', french: '700 x 22C / 700 x 22' },
      { etrto: '23-622', inch: '28 x 0.90', french: '700 x 23C' },
      { etrto: '24-622', inch: '', french: '700 x 24C' },
      { etrto: '25-622', inch: '28 x 1.00', french: '700 x 25C' },
      { etrto: '26-622', inch: '', french: '700 x 26C' },
      { etrto: '28-622', inch: '28 x 1.10', french: '700 x 28C' },
      { etrto: '30-622', inch: '28 x 1.20', french: '700 x 30C' },
      { etrto: '32-622', inch: '28 x 1.25', french: '700 x 32C' },
      { etrto: '33-622', inch: '28 x 1.30', french: '700 x 33C' },
      { etrto: '35-622', inch: '28 x 1.35', french: '700 x 35C' },
      { etrto: '37-622', inch: '28 x 1.40', french: '700 x 35C' },
      { etrto: '40-622', inch: '28 x 1.50', french: '700 x 38C' },
      { etrto: '42-622', inch: '28 x 1.60', french: '700 x 40C' },
      { etrto: '44-622', inch: '28 x 1.625', french: '700 x 42C' },
      { etrto: '47-622', inch: '28 x 1.75', french: '700 x 45C' },
      { etrto: '50-622', inch: '28 x 2.00 / 29 x 2.00 / 28 x 1.90', french: '' },
      { etrto: '55-622', inch: '28 x 2.15 / 29 x 2.15', french: '' },
      { etrto: '60-622', inch: '28 x 2.35 / 29 x 2.35', french: '' },
      { etrto: '32-635', inch: '28 x 1.50 x 1.25', french: '770 x 28B / 700 x 28B' },
      { etrto: '40-635', inch: '28 x 1.50', french: '700 x 38B / 700B Standard' },
      { etrto: '44-635', inch: '28 x 1.625 x 1.50', french: '700 x 42B' },
      { etrto: '28-642', inch: '28 x 1.375 x 1.50', french: '700 x 28A' },
      { etrto: '37-642', inch: '28 x 1.375', french: '700 x 35A' },
    ],
  },
  {
    key: '29-inch',
    label: '29 inch',
    rows: [
      { etrto: '50-622', inch: '29 x 2.00 / 28 x 2.00', french: '' },
      { etrto: '54-622', inch: '29 x 2.10 / 28 x 2.10', french: '' },
      { etrto: '55-622', inch: '29 x 2.15 / 28 x 2.15', french: '' },
      { etrto: '57-622', inch: '29 x 2.25', french: '' },
      { etrto: '60-622', inch: '29 x 2.35 / 28 x 2.35', french: '' },
      { etrto: '62-622', inch: '29 x 2.40', french: '' },
      { etrto: '65-622', inch: '29 x 2.60', french: '' },
      { etrto: '75-622', inch: '29 x 3.00', french: '' },
    ],
  },
]

const tireSizeAvailabilityColumns = [
  tireSizeAvailabilityGroups.slice(0, Math.ceil(tireSizeAvailabilityGroups.length / 2)),
  tireSizeAvailabilityGroups.slice(Math.ceil(tireSizeAvailabilityGroups.length / 2)),
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
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0.75rem;
  background: rgba(13, 15, 20, 0.86);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}

.tire-size-availability-card[open] {
  box-shadow: 0 8px 22px rgba(0, 0, 0, 0.45);
}

.tire-size-availability-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  background: rgba(255, 255, 255, 0.035);
  cursor: pointer;
  list-style: none;
  user-select: none;
}

.tire-size-availability-card__header::-webkit-details-marker {
  display: none;
}

.tire-size-availability-card__header h4 {
  margin: 0;
  color: var(--tz-brand-primary);
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
  border-right: 2px solid var(--tz-brand-primary);
  border-bottom: 2px solid var(--tz-brand-primary);
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
  background: rgba(255, 255, 255, 0.05);
}

.tire-size-availability-table th,
.tire-size-availability-table td {
  padding: 0.55rem 0.7rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
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
  font-family: var(--tz-font-system);
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
  background: rgba(255, 255, 255, 0.04);
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
