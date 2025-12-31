<template>
  <div class="inner-tube-guide space-y-8">
    <!-- Step 1: Search / Intro Card -->
     <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-3 mb-6">
         <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-500/10 text-sky-500 text-lg font-bold">1</span>
           <h3 class="text-xl font-bold text-slate-100 uppercase tracking-wide">
             Choose Your Size
           </h3>
        </div>
      
       <p class="text-slate-400 text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
         All necessary size information (ETRTO or Inch) can be found on your tire sidewall.
       </p>

       <div class="flex justify-center">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-8 py-3 text-sm font-bold uppercase tracking-wider text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] hover:-translate-y-0.5 transition-all"
          @click="openInnerTubeSearch"
        >
          Find My Tube
        </button>
       </div>
    </div>


    <!-- Step 2: Valve Selection Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-3 mb-8">
         <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-500/10 text-sky-500 text-lg font-bold">2</span>
           <h3 class="text-xl font-bold text-slate-100 uppercase tracking-wide">
             Select Valve Type
           </h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <!-- AV -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
             <strong class="text-sky-400 text-lg font-bold mb-1">AV</strong>
             <span class="text-[10px] uppercase font-bold text-slate-500 tracking-wider mb-2">Auto / Schrader</span>
             <p class="text-xs text-slate-400">Standard car valve. Length typical 40mm.</p>
          </div>
          <!-- DV -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
             <strong class="text-sky-400 text-lg font-bold mb-1">DV</strong>
             <span class="text-[10px] uppercase font-bold text-slate-500 tracking-wider mb-2">Dunlop / Woods</span>
             <p class="text-xs text-slate-400">Classic bicycle valve. Length 32/40mm.</p>
          </div>
          <!-- SV -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center border border-sky-500/30">
             <strong class="text-sky-400 text-lg font-bold mb-1">SV</strong>
             <span class="text-[10px] uppercase font-bold text-slate-500 tracking-wider mb-2">Presta / French</span>
             <p class="text-xs text-slate-400">Sport valve. Available in 40/50/60/80mm.</p>
          </div>
       </div>

        <div class="w-full max-w-2xl mx-auto mb-8">
           <GuideImage
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
            src="/public/tiresizecharts/innertube/tanzanite-schwalbe-innertube-valve.webp"
            alt="Schwalbe inner tube valve types and lengths overview"
            :zoomOnClick="true"
          />
        </div>

        <!-- Compatibility Table -->
         <h4 class="text-base font-bold text-slate-200 mb-4 flex items-center justify-center gap-2">
            <span>🔧</span> Pump & Rim Compatibility
         </h4>
        <div class="overflow-x-auto rounded-xl bg-slate-950/40 shadow-[0_4px_16px_rgba(0,0,0,0.5)] inline-block w-full max-w-4xl">
           <table class="min-w-full text-center text-xs sm:text-sm text-slate-100">
            <thead class="bg-slate-900/80">
              <tr>
                <th class="px-4 py-3 font-bold text-slate-400">Rim Height</th>
                <th class="px-4 py-3 font-bold text-sky-400">Rec. Valve</th>
                <th class="px-4 py-3 font-bold text-slate-400">Notes</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800 bg-slate-800/20">
              <tr class="hover:bg-slate-800/40 transition-colors">
                <td class="px-4 py-3 font-semibold text-slate-300">≤25mm</td>
                <td class="px-4 py-3 text-sky-300 font-mono">40mm</td>
                <td class="px-4 py-3 text-slate-400 text-xs">Standard</td>
              </tr>
              <tr class="hover:bg-slate-800/40 transition-colors">
                <td class="px-4 py-3 font-semibold text-slate-300">30–45mm</td>
                <td class="px-4 py-3 text-sky-300 font-mono">60mm</td>
                <td class="px-4 py-3 text-slate-400 text-xs">Deep rim</td>
              </tr>
              <tr class="hover:bg-slate-800/40 transition-colors">
                <td class="px-4 py-3 font-semibold text-slate-300">50–65mm</td>
                <td class="px-4 py-3 text-sky-300 font-mono">80mm</td>
                <td class="px-4 py-3 text-slate-400 text-xs">Aero rim</td>
              </tr>
               <tr class="hover:bg-slate-800/40 transition-colors">
                <td class="px-4 py-3 font-semibold text-slate-300">≥70mm</td>
                <td class="px-4 py-3 text-amber-300 font-mono">Extender</td>
                <td class="px-4 py-3 text-slate-400 text-xs">Valve extender needed</td>
              </tr>
            </tbody>
           </table>
        </div>
    </div>

    <!-- Step 3: Tube Type Selection -->
     <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
       <div class="flex items-center justify-center gap-3 mb-8">
         <span class="flex h-8 w-8 items-center justify-center rounded-lg bg-sky-500/10 text-sky-500 text-lg font-bold">3</span>
           <h3 class="text-xl font-bold text-slate-100 uppercase tracking-wide">
             Choose Model Type
           </h3>
       </div>

       <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <!-- Standard -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
              <strong class="block text-slate-200 text-sm font-bold mb-2">Standard</strong>
              <p class="text-xs text-slate-400 leading-relaxed">
                Reliable hold. Good purity rubber. Uniform wall thickness.
              </p>
          </div>

          <!-- Air Plus -->
          <div class="bg-indigo-500/10 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-indigo-500/20">
              <strong class="block text-indigo-400 text-sm font-bold mb-2">Air Plus</strong>
              <p class="text-xs text-slate-400 leading-relaxed">
                Thicker walls. Holds air longer. Better puncture protection.
              </p>
          </div>

          <!-- Extralight -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
              <strong class="block text-amber-400 text-sm font-bold mb-2">Extralight / XXLight</strong>
              <p class="text-xs text-slate-400 leading-relaxed">
                Reduced weight for racing. Available in select sizes.
              </p>
          </div>

           <!-- Freeride -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
              <strong class="block text-slate-200 text-sm font-bold mb-2">Freeride</strong>
              <p class="text-xs text-slate-400 leading-relaxed">
                Large volume, low weight.
              </p>
          </div>

           <!-- Downhill -->
          <div class="bg-slate-800/40 p-4 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
              <strong class="block text-slate-200 text-sm font-bold mb-2">Downhill</strong>
              <p class="text-xs text-slate-400 leading-relaxed">
                Extremely robust. Reinforced base. Valve shear protection.
              </p>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
import { useShopSearchSheet } from '~/composables/useShopSearchSheet'

const { open: openShopSearchSheet } = useShopSearchSheet()

const openInnerTubeSearch = () => {
  openShopSearchSheet({
    presetCategorySlug: 'inner-tube',
    presetKeywords: ['Inner tube'],
  })
}
</script>
