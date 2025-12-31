<template>
  <div class="how-to-choose-guide">
    <!-- Main Premium Card -->
     <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center">
        <h2 class="text-xl font-bold text-slate-100 mb-6 flex items-center justify-center gap-2">
          How to Choose the Right Size
        </h2>

        <!-- Interactive Helper -->
        <div class="mb-8">
           <TireRimHelper />
        </div>

        <div class="mt-8 border-t border-slate-800 pt-8">
          <div class="flex items-center justify-center gap-2 mb-6">
             <span class="h-px w-8 bg-slate-700"></span>
             <h3 class="text-lg font-bold text-slate-200 uppercase tracking-wider">Manufacturer Standards</h3>
             <span class="h-px w-8 bg-slate-700"></span>
          </div>

          <p class="text-slate-400 text-sm max-w-3xl mx-auto mb-6">
            The following hookless (TSS) and hooked (TC) rim inner width charts from DT Swiss
            show recommended (dark) and possible (light) combinations. Always stay within the
            limits specified by your tire and rim manufacturers.
          </p>

          <div class="mt-3 grid grid-cols-1 gap-6 md:grid-cols-2">
            <div class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50 hover:border-slate-700 transition-colors">
              <img
                src="/public/tiresizecharts/howtochoose/dtswiss-hookless-tss-rim-table.webp"
                alt="DT Swiss hookless TSS rim inner width recommendation chart"
                class="block h-auto w-full"
                loading="lazy"
              />
              <div class="bg-slate-900/50 py-2 text-xs text-slate-500 font-mono tracking-wider">HOOKLESS (TSS)</div>
            </div>
            <div class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800/50 hover:border-slate-700 transition-colors">
              <img
                src="/public/tiresizecharts/howtochoose/dtswiss-hooked-tc-rim-table.webp"
                alt="DT Swiss hooked TC rim inner width recommendation chart"
                class="block h-auto w-full"
                loading="lazy"
              />
               <div class="bg-slate-900/50 py-2 text-xs text-slate-500 font-mono tracking-wider">HOOKED (TC)</div>
            </div>
          </div>
        </div>
     </div>
  </div>
</template>

<script setup lang="ts">
import TireRimHelper from '~/components/TireRimHelper.vue'
</script>
