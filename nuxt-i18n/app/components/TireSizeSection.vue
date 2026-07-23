<template>
  <div class="space-y-8">
    <!-- 1. Intro Card: Basics & Definitions -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
      <div class="mb-6 flex justify-center">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-6 py-2 text-xs font-bold uppercase tracking-wider text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] hover:-translate-y-0.5 transition-all"
          @click="emit('openTireProducts')"
        >
          Check out tires
        </button>
      </div>

      <div class="w-full mb-8 max-w-4xl mx-auto">
        <GuideImage
          src="/public/tiresizecharts/tiresize/tanzanite-schwalbe-tiresize.webp"
          alt="Tanzanite and Schwalbe tire size chart overview"
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
               <strong class="block text-sky-400 text-sm font-bold uppercase tracking-wider mb-2">ETRTO (ISO)</strong>
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
               <strong class="block text-indigo-400 text-sm font-bold uppercase tracking-wider mb-2">French</strong>
               <div class="text-xs tz-text-primary font-mono bg-slate-900/50 px-2 py-1 rounded mb-3">e.g. 700 x 35C</div>
               <p class="text-xs tz-text-secondary leading-relaxed">
                 Indicates approx. <strong>Outer Diameter (700mm)</strong> and <strong>Width (35mm)</strong>. The letter (C) hints at the inner diameter (622mm). Common in Road/Gravel.
               </p>
            </div>
         </div>
      </div>
    </div>

    <!-- 2. Comparison Table Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <h3 class="text-lg font-bold text-slate-100 mb-6 flex items-center justify-center gap-2">
         Notation Comparison Table
       </h3>
       <div class="overflow-x-auto rounded-xl bg-slate-950/40 shadow-[0_4px_16px_rgba(0,0,0,0.5)] inline-block w-full max-w-4xl">
        <table class="min-w-full text-left text-xs sm:text-sm tz-text-secondary">
          <thead class="bg-slate-900/80">
            <tr>
              <th class="px-4 py-3 font-bold tz-text-primary text-center">Dimension</th>
              <th class="px-4 py-3 font-bold text-sky-400 text-center">ETRTO</th>
              <th class="px-4 py-3 font-bold text-amber-400 text-center">Inch</th>
              <th class="px-4 py-3 font-bold text-indigo-400 text-center">French</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800 bg-slate-800/20">
            <tr class="hover:bg-slate-800/40 transition-colors">
              <td class="px-4 py-3 font-semibold tz-text-primary text-center">Example Size</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">37-622</td>
              <td class="px-4 py-3 text-center font-mono tz-text-secondary">
                28 x 1.40<br />
                <span class="text-[10px] tz-text-muted">or 28 x 1 5/8 x 1 3/8</span>
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
                <span class="text-[10px] tz-text-muted">(1.40 Inch)</span>
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
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-2 mb-4">
         <div class="h-px w-8 bg-slate-700"></div>
         <h3 class="text-lg font-bold text-slate-100 uppercase tracking-wider">Market Landscape</h3>
         <div class="h-px w-8 bg-slate-700"></div>
       </div>

      <p class="sizecharts-section__intro max-w-3xl mx-auto mb-6 text-center">
        This list encompasses almost every known size designation found in the market, including historical and modern standards.
        Classic fractional sizes are often found on vintage Schwalbe tires or specific heritage models.
      </p>

      <div
        class="rounded-2xl bg-slate-900/80 shadow-[0_4px_16px_rgba(0,0,0,0.5)] overflow-hidden inline-block"
      >
        <img
          src="/public/tiresizecharts/tiresize/what-tire-sizes-are-available-schwalbe.svg"
          alt="Overview table comparing ETRTO, inch and French tire size designations"
          class="block h-auto w-full max-w-4xl"
          loading="lazy"
        />
      </div>
    </div>

    <!-- 4. 28 vs 29 Comparison Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center border-t-4 border-indigo-500">
       <h3 class="text-xl font-bold text-slate-100 mb-6">
         The "28 Inch" vs "29 Inch" Mystery
       </h3>

       <div class="grid md:grid-cols-2 gap-8 items-center max-w-4xl mx-auto">
          <div class="bg-indigo-500/5 rounded-xl p-5 shadow-[0_4px_16px_rgba(0,0,0,0.5)] h-full flex flex-col justify-center">
             <strong class="block text-indigo-400 text-3xl font-bold mb-2">SAME</strong>
             <span class="text-xs uppercase tracking-widest tz-text-muted mb-3">Inner Diameter</span>
             <p class="tz-text-primary font-mono text-lg">622 mm</p>
             <p class="text-xs tz-text-muted mt-2 leading-relaxed">
                Both 28" (Touring/Road) and 29" (MTB) tires fit on the same 622mm rim diameter.
             </p>
          </div>

          <div class="text-left space-y-4">
             <div class="space-y-2">
                <strong class="text-sky-400 text-sm font-bold uppercase tracking-wider block">History & Context</strong>
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
</script>
