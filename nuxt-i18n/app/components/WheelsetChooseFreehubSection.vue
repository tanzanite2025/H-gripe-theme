<template>
  <div class="space-y-8">
    <!-- Intro Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-slate-500">
       <div class="flex flex-col items-center">

         <h3 class="text-lg font-bold text-slate-100 mb-2">Select Your Freehub Body</h3>
         <p class="text-sm text-slate-400 leading-relaxed max-w-2xl mx-auto mb-4">
            A comprehensive guide to matching your freehub with your drivetrain cassettes (Shimano, SRAM, Campagnolo).
         </p>
         <FreehubGroupsetHelper />
       </div>
    </div>

    <!-- 1. Shimano Ecosystem (Blue) -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
       <div class="flex items-center gap-3 mb-6 pb-3 border-b border-blue-500/10">
          <h3 class="text-lg font-bold text-slate-200">Shimano Ecosystem</h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          <div class="bg-blue-500/5 rounded-xl p-3 border border-blue-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/shimano-micro-spline-11-12-speed-mountain-freehub.webp"
                alt="Shimano Micro Spline 11/12-speed mountain freehub body"
                :zoomOnClick="true"
                caption="For MTB 12-Speed"
                class="rounded-lg mb-2"
              />
              <strong class="block text-blue-400 text-sm mb-1 text-center">Micro Spline (MS)</strong>
              <p class="text-xs text-slate-400 text-center">Deore M6100, XT M8100, XTR M9100</p>
          </div>
          <div class="bg-blue-500/5 rounded-xl p-3 border border-blue-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-road-hyper-freehub.webp"
                alt="Shimano road HG freehub body"
                :zoomOnClick="true"
                caption="For Road 11/12-Speed"
                class="rounded-lg mb-2"
              />
              <strong class="block text-blue-400 text-sm mb-1 text-center">HG Road (HG-11)</strong>
              <p class="text-xs text-slate-400 text-center">105, Ultegra, Dura-Ace (8-12sp)</p>
          </div>
          <div class="bg-blue-500/5 rounded-xl p-3 border border-blue-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/shimano-8-9-10-11-speed-mountain-hyper-freehub.webp"
                alt="Shimano mountain HG freehub body"
                :zoomOnClick="true"
                caption="For MTB 8-11 Speed"
                class="rounded-lg mb-2"
              />
              <strong class="block text-blue-400 text-sm mb-1 text-center">HG Mountain</strong>
              <p class="text-xs text-slate-400 text-center">Older MTB & Entry Road setups</p>
          </div>
       </div>

       <div class="bg-slate-800/50 rounded-xl p-4">
          <h4 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">Compatibility Cheatsheet</h4>
          <ul class="space-y-3">
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-blue-500 shrink-0"></span>
                <span><strong>MS (Micro Spline):</strong> Necessary for Shimano 12-speed MTB cassettes (10T start). Not backward compatible.</span>
             </li>
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-blue-500 shrink-0"></span>
                <span><strong>HG (HyperGlide):</strong> The gold standard for decades. Fits 8-11sp MTB and 8-12sp Road cassettes. *Note: Road 11sp HG freehubs are longer; require 1.85mm spacer for MTB cassettes.*</span>
             </li>
          </ul>
       </div>
    </div>

    <!-- 2. SRAM Ecosystem (Red) -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
       <div class="flex items-center gap-3 mb-6 pb-3 border-b border-red-500/10">
          <h3 class="text-lg font-bold text-slate-200">SRAM Ecosystem</h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div class="bg-red-500/5 rounded-xl p-3 border border-red-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/sram-xd-11-12-speed-mountain-freehub.webp"
                alt="SRAM XD freehub body"
                :zoomOnClick="true"
                caption="For MTB 11/12-Speed"
                class="rounded-lg mb-2"
              />
              <strong class="block text-red-400 text-sm mb-1 text-center">XD Driver</strong>
              <p class="text-xs text-slate-400 text-center">GX Eagle, X01, XX1 (10T start)</p>
          </div>
          <div class="bg-red-500/5 rounded-xl p-3 border border-red-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/sram-xdr-road-11-12-speed-freehub.webp"
                alt="SRAM XDR freehub body"
                :zoomOnClick="true"
                caption="For Road 12-Speed"
                class="rounded-lg mb-2"
              />
              <strong class="block text-red-400 text-sm mb-1 text-center">XDR Driver</strong>
              <p class="text-xs text-slate-400 text-center">Rival, Force, Red eTap AXS</p>
          </div>
       </div>

       <div class="bg-slate-800/50 rounded-xl p-4">
          <h4 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">Compatibility Cheatsheet</h4>
          <ul class="space-y-3">
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-red-500 shrink-0"></span>
                <span><strong>XD:</strong> Compact design for MTB cassettes with a 10T small cog.</span>
             </li>
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-red-500 shrink-0"></span>
                <span><strong>XDR:</strong> "Road" version of XD, 1.85mm longer. Can run XD cassettes with a 1.85mm spacer. Designed for 12-speed AXS road groups.</span>
             </li>
          </ul>
       </div>
    </div>

    <!-- 3. Campagnolo Ecosystem (Indigo) -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
       <div class="flex items-center gap-3 mb-6 pb-3 border-b border-indigo-500/10">
          <h3 class="text-lg font-bold text-slate-200">Campagnolo Ecosystem</h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div class="bg-indigo-500/5 rounded-xl p-3 border border-indigo-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-N3W-freehub.webp"
                alt="Campagnolo N3W freehub"
                :zoomOnClick="true"
                caption="For 13-Speed / Gravel"
                class="rounded-lg mb-2"
              />
              <strong class="block text-indigo-400 text-sm mb-1 text-center">N3W</strong>
              <p class="text-xs text-slate-400 text-center">Ekar 13sp (Backward compatible w/ adapter)</p>
          </div>
          <div class="bg-indigo-500/5 rounded-xl p-3 border border-indigo-500/10">
              <GuideImage
                src="/public/wheelsetbuyersguide/choose%20freehub/Campagnolo-8-9-10-11-spd-freehub.webp"
                alt="Campagnolo Classic freehub"
                :zoomOnClick="true"
                caption="Classic Road"
                class="rounded-lg mb-2"
              />
              <strong class="block text-indigo-400 text-sm mb-1 text-center">Classic (ED/CP)</strong>
              <p class="text-xs text-slate-400 text-center">9/10/11/12 Speed Road</p>
          </div>
       </div>

       <div class="bg-slate-800/50 rounded-xl p-4">
          <h4 class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-3">Compatibility Cheatsheet</h4>
          <ul class="space-y-3">
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-indigo-500 shrink-0"></span>
                <span><strong>N3W:</strong> The future standard. Native for Ekar 13sp. Fits older 10-12sp cassettes with an adapter.</span>
             </li>
             <li class="flex items-start gap-3 text-xs text-slate-300">
                <span class="mt-0.5 w-1.5 h-1.5 rounded-full bg-indigo-500 shrink-0"></span>
                <span><strong>Classic:</strong> The traditional spline for Super Record, Record, Chorus (11/12sp).</span>
             </li>
          </ul>
       </div>
    </div>

    <!-- SEO & Technical Reference Tables (Restored) -->
    <div class="pt-6 border-t border-slate-800/50">
       <h3 class="text-base font-bold text-slate-200 mb-4 px-2 border-l-4 border-slate-500">
         Technical Reference: Freehub Standards & Groupsets
       </h3>
       <div class="overflow-x-auto rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] bg-[#11151e]">
          <table class="min-w-full text-left text-xs sm:text-sm text-slate-400">
            <thead class="bg-slate-800/50 font-bold text-slate-200 uppercase tracking-wider text-[10px] sm:text-xs">
              <tr>
                <th class="px-4 py-3">Type</th>
                <th class="px-4 py-3">Example Groupsets</th>
                <th class="px-4 py-3">Features</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/50">
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">MS (Micro Spline 12s)</td>
                <td class="px-4 py-3">Shimano Deore M6100, SLX M7100, XT M8100, XTR M9100</td>
                <td class="px-4 py-3">MTB 12-speed specific; supports 10T cog.</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">HG 8s</td>
                <td class="px-4 py-3">Claris R2000, Altus M310, Acera M360</td>
                <td class="px-4 py-3">Entry-level Road/MTB (11-32T).</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">HG 9-11s</td>
                <td class="px-4 py-3">Sora, Tiagra, 105, Ultegra, Dura-Ace; MTB Deore/SLX/XT</td>
                <td class="px-4 py-3">Most common standard. Strong cross-compatibility.</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">XD</td>
                <td class="px-4 py-3">SRAM GX Eagle, X01 Eagle, XX1 Eagle</td>
                <td class="px-4 py-3">MTB 11/12s, supports 10T cog, lightweight.</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">XDR</td>
                <td class="px-4 py-3">SRAM Force/Red eTap AXS</td>
                <td class="px-4 py-3">Road 12s specific, 1.85mm longer than XD.</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">N3W</td>
                <td class="px-4 py-3">Campagnolo Ekar 13s</td>
                <td class="px-4 py-3">Native 13s. Fits 10-12s w/ adapter.</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                <td class="px-4 py-3 font-medium text-slate-300">CP (Classic)</td>
                <td class="px-4 py-3">Super Record 11, Record 11, Chorus 11</td>
                <td class="px-4 py-3">Traditional 11s Road only. No 12/13s support.</td>
              </tr>
            </tbody>
          </table>
       </div>

       <h3 class="text-base font-bold text-slate-200 mt-8 mb-4 px-2 border-l-4 border-slate-500">
         Technical Reference: Compatibility & Spacer Matrix
       </h3>
       <div class="overflow-x-auto rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] bg-[#11151e]">
          <table class="min-w-full text-left text-xs sm:text-sm text-slate-400">
            <thead class="bg-slate-800/50 font-bold text-slate-200 uppercase tracking-wider text-[10px] sm:text-xs">
              <tr>
                <th class="px-4 py-3">Brand</th>
                <th class="px-4 py-3">Freehub Body</th>
                <th class="px-4 py-3">Cassette</th>
                <th class="px-4 py-3">Spacer Needed?</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/50">
              <!-- Shimano -->
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-blue-400">Shimano</td>
                 <td class="px-4 py-3">HG 8-11s MTB</td>
                 <td class="px-4 py-3">8/9/11s MTB</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-blue-400">Shimano</td>
                 <td class="px-4 py-3">HG 8-11s MTB</td>
                 <td class="px-4 py-3">10s Cassette</td>
                 <td class="px-4 py-3 text-amber-400">1mm</td>
              </tr>
               <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-blue-400">Shimano</td>
                 <td class="px-4 py-3">HG 11/12s Road</td>
                 <td class="px-4 py-3">8/9/11s MTB</td>
                 <td class="px-4 py-3 text-amber-400">1.85mm</td>
              </tr>
               <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-blue-400">Shimano</td>
                 <td class="px-4 py-3">HG 11/12s Road</td>
                 <td class="px-4 py-3">11/12s Road</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-blue-400">Shimano</td>
                 <td class="px-4 py-3">MS 12s MTB</td>
                 <td class="px-4 py-3">11/12s MS Cassette</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              
              <!-- SRAM -->
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-red-400">SRAM</td>
                 <td class="px-4 py-3">HG-Style 9/10s</td>
                 <td class="px-4 py-3">SRAM 9/10s / NX Eagle</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-red-400">SRAM</td>
                 <td class="px-4 py-3">XD 11/12s</td>
                 <td class="px-4 py-3">SRAM PhD/Eagle (GX+)</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-red-400">SRAM</td>
                 <td class="px-4 py-3">XDR 12s Road</td>
                 <td class="px-4 py-3">SRAM AXS Road</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
               <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-red-400">SRAM</td>
                 <td class="px-4 py-3">XDR 12s Road</td>
                 <td class="px-3 py-2">SRAM XD Cassette</td>
                 <td class="px-3 py-2 text-amber-400">1.85mm</td>
              </tr>

              <!-- Campy -->
               <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-indigo-400">Campagnolo</td>
                 <td class="px-4 py-3">N3W 13s</td>
                 <td class="px-4 py-3">Ekar 13s (9/10t start)</td>
                 <td class="px-4 py-3 text-emerald-400">No</td>
              </tr>
              <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-indigo-400">Campagnolo</td>
                 <td class="px-4 py-3">N3W 13s</td>
                 <td class="px-4 py-3">Classic 10-12s</td>
                 <td class="px-4 py-3 text-amber-400">AC21-N3W Adapter</td>
              </tr>

               <!-- Mavic -->
               <tr class="hover:bg-slate-800/30 transition-colors">
                 <td class="px-4 py-3 font-medium text-yellow-400">Mavic</td>
                 <td class="px-4 py-3">HG-Style 8-11s</td>
                 <td class="px-4 py-3">Shimano/SRAM 9s</td>
                 <td class="px-4 py-3 text-amber-400">1.75mm Mavic</td>
              </tr>
            </tbody>
          </table>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'
import FreehubGroupsetHelper from '~/components/FreehubGroupsetHelper.vue'
</script>
