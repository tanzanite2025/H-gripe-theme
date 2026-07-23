<template>
  <div>
    <!-- Tabs Header -->
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'technology' }"
        @click="activeTab = 'technology'"
      >
        Technology
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'how-to-choose' }"
        @click="activeTab = 'how-to-choose'"
      >
        How to choose
      </button>
    </div>

    <!-- Tab: Technology (Existing Content) -->
    <div v-show="activeTab === 'technology'">
      <!-- Intro Card -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center mb-8 border-t-4 border-indigo-500">
        <h3 class="text-xl font-bold text-slate-100 mb-4 flex items-center justify-center gap-2">
          Rim Technology
        </h3>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-2xl mx-auto">
          The rim is the rotating mass furthest from the center. Its weight, aerodynamics, and width directly define the riding character—responsiveness, comfort, and speed.
        </p>

        <!-- Key Metrics Row -->
        <div class="grid grid-cols-3 gap-2 md:gap-4 max-w-lg mx-auto mb-4">
          <div class="bg-indigo-500/10 rounded-lg p-3">
             <span class="block text-xl font-bold text-indigo-400">Toray</span>
             <span class="text-[10px] uppercase tracking-wider tz-text-muted">Carbon Fiber</span>
          </div>
          <div class="bg-indigo-500/10 rounded-lg p-3">
             <span class="block text-xl font-bold text-indigo-400">UCI</span>
             <span class="text-[10px] uppercase tracking-wider tz-text-muted">Approved</span>
          </div>
          <div class="bg-indigo-500/10 rounded-lg p-3">
             <span class="block text-xl font-bold text-indigo-400">3-Year</span>
             <span class="text-[10px] uppercase tracking-wider tz-text-muted">Warranty</span>
          </div>
        </div>
      </div>

      <!-- 1. Profiles & Depths -->
      <div class="mb-10">
         <div class="flex items-center justify-center gap-2 mb-6">
            <div class="h-px w-8 bg-slate-700"></div>
        <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">Rim Profiles</h4>
            <div class="h-px w-8 bg-slate-700"></div>
         </div>

         <div class="grid md:grid-cols-2 gap-6">
            <!-- Climbing / Shallow -->
            <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 transition-transform hover:-translate-y-1">
               <div class="flex justify-between items-start mb-4">
                  <span class="text-sm font-bold text-slate-100">Shallow Depth</span>
          <span class="text-xs bg-slate-800 tz-text-muted px-2 py-1 rounded">30mm - 40mm</span>
               </div>
          <p class="text-xs tz-text-secondary leading-relaxed mb-3">
                 Minimized rotational weight for instant acceleration and climbing efficiency. Less affected by crosswinds.
               </p>
               <div class="text-xs text-indigo-400 font-medium uppercase tracking-wide">
                 Best for: Climbing, Windy Conditions
               </div>
            </div>
            <!-- Aero / Deep -->
            <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 transition-transform hover:-translate-y-1">
               <div class="flex justify-between items-start mb-4">
                  <span class="text-sm font-bold text-slate-100">Aero Depth</span>
          <span class="text-xs bg-slate-800 tz-text-muted px-2 py-1 rounded">45mm - 65mm+</span>
               </div>
          <p class="text-xs tz-text-secondary leading-relaxed mb-3">
                 Maximizes aerodynamic advantage by smoothing airflow. Maintains momentum at high speeds (35km/h+).
               </p>
               <div class="text-xs text-indigo-400 font-medium uppercase tracking-wide">
                 Best for: Flats, Racing, Time Trials
               </div>
            </div>
         </div>
      </div>

      <!-- 2. Width Concepts -->
      <div class="mb-10">
         <div class="flex items-center justify-center gap-2 mb-6">
            <div class="h-px w-8 bg-slate-700"></div>
        <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">Rim Width & Tire Fit</h4>
            <div class="h-px w-8 bg-slate-700"></div>
         </div>
         
         <div class="bg-slate-800/30 rounded-2xl p-6 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
           <div class="flex flex-col md:flex-row gap-8 items-center">
              <div class="flex-1 space-y-4">
                 <div>
          <strong class="block tz-text-primary text-sm mb-1">Internal Width (ID)</strong>
          <p class="text-xs tz-text-secondary leading-relaxed">
                     Determines the tire shape. Wider ID (e.g., 21mm+) allows tires to "balloon" properly, creating a wider contact patch, lower rolling resistance, and better grip.
                   </p>
                 </div>
                 <div>
          <strong class="block tz-text-primary text-sm mb-1">External Width (OD)</strong>
          <p class="text-xs tz-text-secondary leading-relaxed">
                     Should ideally be slightly wider than the tire (Rule of 105%) for optimal straight-line aerodynamics.
                   </p>
                 </div>
                 <div class="p-3 bg-indigo-500/10 border-l-2 border-indigo-500 text-xs text-indigo-200">
                    <strong>Trend:</strong> Modern road rims are moving towards 21-25mm ID to optimized 28mm-32mm tires.
                 </div>
              </div>
              <div class="w-full md:w-1/3">
                 <!-- Placeholder for Rim graphical cross section if available, else a simple CSS representation -->
                  <GuideImage
                    src="/public/wheelsetbuyersguide/wheelcomponents/rim/carbon-rim-hooked-vs-hookless.webp"
                    alt="Rim Width and Hookless profile explanation"
                    :zoomOnClick="true"
                    caption="Modern Rim Profiles"
                    class="rounded-lg w-full shadow-lg"
                 />
              </div>
           </div>
         </div>
      </div>

      <!-- 3. Hooked vs Hookless -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
         <!-- Hooked -->
         <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center border-b-2 border-slate-700">
        <h5 class="tz-text-primary font-bold mb-3">Hooked (Crotchet)</h5>
            <p class="text-xs tz-text-secondary leading-relaxed mb-4 min-h-[3rem]">
              Traditional design with bead hooks. Compatible with almost all clincher and tubeless tires. High pressure tolerance.
            </p>
            <span class="inline-block px-3 py-1 bg-slate-800 tz-text-secondary text-[10px] uppercase font-bold rounded-full">Universal Compatibility</span>
         </div>

         <!-- Hookless -->
         <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-5 md:p-6 text-center border-b-2 border-indigo-500">
            <h5 class="text-indigo-400 font-bold mb-3">Hookless (TSS)</h5>
            <p class="text-xs tz-text-secondary leading-relaxed mb-4 min-h-[3rem]">
              Modern straight-wall design. Lighter, stronger rim wall, and more aero transition to tire.
              <br/><span class="text-rose-400/80 mt-1 block">*Requires Tubeless-Ready Tires & Max 73psi (5 bar).</span>
            </p>
            <span class="inline-block px-3 py-1 bg-indigo-500/10 text-indigo-400 text-[10px] uppercase font-bold rounded-full">Performance Choice</span>
         </div>
      </div>
    </div>

    <!-- Tab: How to choose (New Content) -->
    <div v-show="activeTab === 'how-to-choose'">
       <div class="space-y-10">
          
          <!-- 1. Rim Bed Type (Table Comparison) -->
          <div>
             <div class="flex items-center justify-center gap-2 mb-6">
                <div class="h-px w-8 bg-slate-700"></div>
                <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">1. Rim Bed Drilling Options</h4>
                <div class="h-px w-8 bg-slate-700"></div>
             </div>

             <div class="overflow-x-auto rounded-lg shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-slate-800">
                <table class="w-full text-sm text-left tz-text-secondary">
                   <thead class="text-xs tz-text-primary uppercase bg-slate-800/80">
                      <tr>
                         <th scope="col" class="px-6 py-4 font-bold">Feature</th>
                         <th scope="col" class="px-6 py-4 font-bold text-indigo-400">No-Tape (Undrilled)</th>
                         <th scope="col" class="px-6 py-4 font-bold text-sky-400">Standard Drilled (Default)</th>
                      </tr>
                   </thead>
                   <tbody class="divide-y divide-slate-800">
                      <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium tz-text-primary">Structure</td>
                         <td class="px-6 py-4">No spoke access holes in rim bed.</td>
                         <td class="px-6 py-4">Has spoke access holes.</td>
                      </tr>
                      <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium tz-text-primary">Airtightness</td>
                         <td class="px-6 py-4">Better (Inherently airtight).</td>
                         <td class="px-6 py-4"><span class="text-amber-500">Requires tubeless tape.</span></td>
                      </tr>
                      <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium tz-text-primary">Weight</td>
                         <td class="px-6 py-4">Adds ~1-3g (Negligible).</td>
                         <td class="px-6 py-4">Standard weight + Tape weight.</td>
                      </tr>
                       <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium tz-text-primary">Assembly Difficulty</td>
                         <td class="px-6 py-4 text-amber-500">More difficult (Nipple guidance required).</td>
                         <td class="px-6 py-4 text-emerald-500">Easier assembly.</td>
                      </tr>
                      <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium tz-text-primary">Compabitility</td>
                         <td class="px-6 py-4">
                            <div class="space-y-1">
                               <div class="flex items-center gap-2"><span class="text-emerald-500">✓</span> Tubeless Tires</div>
                               <div class="flex items-center gap-2"><span class="text-emerald-500">✓</span> Tube + Tire</div>
                            </div>
                         </td>
                         <td class="px-6 py-4">
                            <div class="space-y-1">
                               <div class="flex items-center gap-2"><span class="text-emerald-500">✓</span> Tubeless Tires</div>
                               <div class="flex items-center gap-2"><span class="text-emerald-500">✓</span> Tube + Tire</div>
                            </div>
                         </td>
                      </tr>
                      <tr class="bg-[#11151e] hover:bg-slate-800/30">
                         <td class="px-6 py-4 font-medium text-rose-400">Caution</td>
                         <td class="px-6 py-4 text-xs italic" colspan="2">
                            Strongly <strong>not recommended</strong> to use low-end wire bead tires (potential rim damage). High-end wire bead tires may differ, but caution is advised.
                         </td>
                      </tr>
                   </tbody>
                </table>
             </div>
          </div>

          <!-- 2. Symmetry & Drilling Pattern -->
          <div class="grid md:grid-cols-2 gap-6">
             <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5">
                <h5 class="flex items-center gap-2 tz-text-primary font-bold mb-3">
                   <span class="w-2 h-2 rounded-full bg-indigo-500"></span> 2. Symmetrical Comparison
                </h5>
                <div class="space-y-4">
                   <div class="bg-indigo-500/5 p-3 rounded-lg">
                      <strong class="block text-indigo-400 text-sm mb-1">Symmetrical (Default 1:1)</strong>
                      <p class="text-xs tz-text-secondary">Most stable load distribution. Recommended standard option.</p>
                   </div>
                   <div class="bg-slate-800/30 p-3 rounded-lg border border-slate-700/50">
                      <strong class="block tz-text-primary text-sm mb-1">Asymmetrical (e.g., 2:1)</strong>
                      <p class="text-xs tz-text-secondary">
                         Requires specific hub compatibility. Custom option (CAD drawing needed). Please contact support.
                      </p>
                   </div>
                </div>
             </div>

             <!-- 4. Drain Holes -->
             <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5">
                <h5 class="flex items-center gap-2 tz-text-primary font-bold mb-3">
                   <span class="w-2 h-2 rounded-full bg-sky-500"></span> 4. Drain Holes
                </h5>
                <div class="space-y-3">
                   <div class="flex items-start gap-3">
                      <span class="mt-1 w-1.5 h-1.5 rounded-full bg-sky-500 shrink-0"></span>
                      <div>
                         <strong class="tz-text-primary text-sm block">With Drain Holes</strong>
                         <p class="text-xs tz-text-secondary">Recommended for frequent wet or rainy riding conditions.</p>
                      </div>
                   </div>
                   <div class="flex items-start gap-3">
                      <span class="mt-1 w-1.5 h-1.5 rounded-full bg-amber-500 shrink-0"></span>
                      <div>
                         <strong class="tz-text-primary text-sm block">Without Drain Holes</strong>
                         <p class="text-xs tz-text-secondary">Suitable for dry environments. Reduces extra drilling.</p>
                      </div>
                   </div>
                </div>
             </div>
          </div>

          <!-- 3. Valve Hole -->
          <div>
            <div class="flex items-center justify-center gap-2 mb-6">
                <div class="h-px w-8 bg-slate-700"></div>
                <h4 class="tz-text-primary font-bold text-base uppercase tracking-wider">3. Valve Hole Standards</h4>
                <div class="h-px w-8 bg-slate-700"></div>
             </div>
             
             <div class="grid md:grid-cols-2 gap-6">
                <!-- Presta -->
                <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 border-l-4 border-emerald-500">
                   <div class="flex justify-between items-start mb-3">
                   <h5 class="text-lg font-bold tz-text-primary">Presta Valve</h5>
                      <span class="text-[10px] bg-emerald-500/20 text-emerald-400 px-2 py-1 rounded uppercase font-bold">Default</span>
                   </div>
                   <ul class="space-y-2 text-sm tz-text-secondary mb-4">
                      <li>• Hole Diameter: ~6.5mm</li>
                      <li>• Lightweight standard for Road/MTB/Tubeless.</li>
                      <li>• All our rims default to Presta.</li>
                   </ul>
                   <div class="bg-rose-900/20 p-3 rounded border border-rose-500/20 text-xs text-rose-300">
                      <strong>Note:</strong> Tubeless profiles may make installing certain wire bead tires difficult. Caution advised.
                   </div>
                </div>

                <!-- Schrader -->
                <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 opacity-75 grayscale hover:grayscale-0 transition-all duration-300 border-l-4 border-slate-600">
                   <div class="flex justify-between items-start mb-3">
                      <h5 class="text-lg font-bold tz-text-primary">Schrader Valve</h5>
                      <span class="text-[10px] bg-slate-700 tz-text-muted px-2 py-1 rounded uppercase font-bold">Rare</span>
                   </div>
                   <ul class="space-y-2 text-sm tz-text-secondary">
                      <li>• Hole Diameter: ~8.5mm</li>
                      <li>• Common in city/kids/fat bikes.</li>
                      <li>• Rarely used in high-end performance rims.</li>
                   </ul>
                </div>
             </div>
          </div>

          <!-- 5. Notes & Risks -->
          <div class="bg-amber-900/10 border border-amber-500/20 rounded-2xl p-6">
             <h5 class="text-amber-500 font-bold mb-4 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z"/></svg>
                Critical Notes & Risks
             </h5>
             <div class="grid md:grid-cols-2 gap-6 text-sm">
                <div>
                   <strong class="block tz-text-primary mb-2">Incorrect Drilling Angle</strong>
                   <p class="tz-text-secondary leading-relaxed mb-2">
                      May cause uneven spoke tension and increase breakage risk.
                   </p>
                   <p class="text-emerald-400 text-xs">
                      ✓ Our rims are drilled precisely according to hub specifications.
                   </p>
                </div>
                <div>
                   <strong class="block tz-text-primary mb-2">Closed Tubeless Rim Bed</strong>
                   <p class="tz-text-secondary leading-relaxed">
                      More difficult to assemble. Requires professional tools/experience. We recommend professional installation or our pre-nipple installation service.
                   </p>
                </div>
             </div>
          </div>

       </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import GuideImage from '~/components/GuideImage.vue'

const activeTab = ref<'technology' | 'how-to-choose'>('technology')
</script>
