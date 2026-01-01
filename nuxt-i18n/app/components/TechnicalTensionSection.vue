<template>
  <div class="space-y-8">
    <!-- Intro Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-amber-500">
       <h3 class="text-lg font-bold text-slate-100 mb-2">Wheel Spoke Tension Standard</h3>
       <p class="text-sm text-slate-400 leading-relaxed max-w-2xl mx-auto mb-6">
          Tension is the soul of a wheel. We customize tension based on **Rider Weight**, **Usage**, and **Rim Material** to balance stiffness, durability, and comfort.
       </p>
       <div class="flex justify-center flex-col items-center gap-4">
         <GuideImage
            src="/public/technical/tension/tanzanite-wheel-spoke-tension.webp"
            alt="Wheel Spoke Tension Distribution"
            :zoomOnClick="true"
            caption="Tanzanite Workshop: Tension Distribution Reference"
            class="rounded-xl max-w-lg w-full"
         />
         <NuxtLink
           :to="localePath('/support/test-report') + '#wheelset-assembly'"
           class="premium-button"
         >
           Check the wheelset assembly guide
         </NuxtLink>
       </div>
    </div>

    <!-- 1. Key Factors Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- Rider Weight -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-amber-500 mb-4 flex items-center gap-2">
              Effect of Rider Weight
           </h3>
           <ul class="space-y-3 text-sm text-slate-300">
              <li class="flex items-start gap-3">
                 <span class="mt-1 w-1.5 h-1.5 rounded-full bg-amber-500 shrink-0"></span>
                 <span><strong>Static Load:</strong> A 75kg rider increases upper spoke tension by ~15%, while lower spokes lose tension. (De-tensioning cycle).</span>
              </li>
              <li class="flex items-start gap-3">
                 <span class="mt-1 w-1.5 h-1.5 rounded-full bg-amber-500 shrink-0"></span>
                 <span><strong>Heavy Riders:</strong> Need higher base tension and thicker spokes (14G) to prevent spokes from going slack, which causes fatigue failure.</span>
              </li>
           </ul>
       </div>

       <!-- Usage Scenarios -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-sky-500 mb-4 flex items-center gap-2">
              Standard by Usage
           </h3>
           <div class="space-y-3 text-sm">
              <div class="p-2 rounded-lg bg-indigo-500/10 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <strong class="block text-indigo-400">Road Racing</strong>
                 <span class="text-slate-400 text-xs">High stiffness Focus. 120-130kgf (Drive Side).</span>
              </div>
              <div class="p-2 rounded-lg bg-emerald-500/10 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <strong class="block text-emerald-400">MTB / Gravel</strong>
                 <span class="text-slate-400 text-xs">Durability Focus. Slightly lower tension for compliance and impact absorption.</span>
              </div>
              <div class="p-2 rounded-lg bg-teal-500/10 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <strong class="block text-teal-400">Touring / Commuting</strong>
                 <span class="text-slate-400 text-xs">Reliability Focus. Moderate tension, thicker spokes.</span>
              </div>
           </div>
       </div>
    </div>

    <!-- 2. Tension Chart -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] overflow-hidden">
       <div class="p-4 bg-slate-800/50 border-b border-slate-700/50 flex items-center justify-between">
          <h3 class="font-bold text-slate-200">Recommended Drive-Side Tension (kgf)</h3>
          <span class="text-xs font-mono text-slate-500 bg-slate-900 px-2 py-1 rounded">Unit: kgf</span>
       </div>
       <div class="overflow-x-auto">
          <table class="min-w-full text-left text-sm text-slate-400">
            <thead class="bg-slate-800/30 font-semibold text-slate-300">
              <tr>
                <th class="px-5 py-3">Rider Weight</th>
                <th class="px-5 py-3 text-indigo-300">Road Racing</th>
                <th class="px-5 py-3 text-emerald-300">MTB</th>
                <th class="px-5 py-3 text-teal-300">Touring</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/50">
              <tr class="hover:bg-slate-800/20 transition-colors">
                <td class="px-5 py-3 font-medium text-slate-200">≤ 60 kg</td>
                <td class="px-5 py-3">115–120</td>
                <td class="px-5 py-3">95–105</td>
                <td class="px-5 py-3">90–100</td>
              </tr>
              <tr class="hover:bg-slate-800/20 transition-colors">
                <td class="px-5 py-3 font-medium text-slate-200">60–80 kg</td>
                <td class="px-5 py-3">120–125</td>
                <td class="px-5 py-3">100–115</td>
                <td class="px-5 py-3">95–105</td>
              </tr>
              <tr class="hover:bg-slate-800/20 transition-colors">
                <td class="px-5 py-3 font-medium text-slate-200">80–100 kg</td>
                <td class="px-5 py-3">125–130</td>
                <td class="px-5 py-3">110–120</td>
                <td class="px-5 py-3">100–110</td>
              </tr>
              <tr class="hover:bg-slate-800/20 transition-colors">
                <td class="px-5 py-3 font-medium text-slate-200">≥ 100 kg</td>
                <td class="px-5 py-3">130–135</td>
                <td class="px-5 py-3">115–125</td>
                <td class="px-5 py-3">105–115</td>
              </tr>
            </tbody>
          </table>
       </div>
       <div class="p-4 bg-amber-500/5 border-t border-amber-500/10 text-xs text-amber-200/80 leading-relaxed">
         <strong>Note:</strong> These are baseline values for Drive-Side spokes. Non-Drive Side tension is typically 60-70% of Drive Side, determined by hub geometry. Max tension for Carbon rims is usually ~135kgf; Alloy rims ~125kgf.
       </div>
    </div>

    <!-- 3. Workshop Process (Timeline) -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6">
       <h3 class="text-lg font-bold text-fuchsia-500 mb-6 flex items-center gap-2">
          Our Workshop Building Process
       </h3>
       
       <div class="relative pl-6 border-l-2 border-slate-700 space-y-8">
          <!-- Step 1 -->
          <div class="relative">
             <span class="absolute -left-[31px] bg-[#11151e] p-1">
                <span class="w-3 h-3 rounded-full bg-slate-500 block"></span>
             </span>
             <h4 class="text-sm font-bold text-slate-200">1. Baseline Calculation</h4>
             <p class="text-xs text-slate-400 mt-1">Determine base tension range based on Rider Weight classification.</p>
          </div>
          <!-- Step 2 -->
          <div class="relative">
             <span class="absolute -left-[31px] bg-[#11151e] p-1">
                <span class="w-3 h-3 rounded-full bg-indigo-500 block"></span>
             </span>
             <h4 class="text-sm font-bold text-slate-200">2. Usage Adjustment</h4>
             <p class="text-xs text-slate-400 mt-1">Adjust for stiffness (Road) or compliance (MTB).</p>
          </div>
          <!-- Step 3 -->
          <div class="relative">
             <span class="absolute -left-[31px] bg-[#11151e] p-1">
                <span class="w-3 h-3 rounded-full bg-rose-500 block"></span>
             </span>
             <h4 class="text-sm font-bold text-slate-200">3. Safety Cap Check</h4>
             <p class="text-xs text-slate-400 mt-1">Ensure target does not exceed Rim Limits (Carbon 135kgf / Alloy 125kgf).</p>
          </div>
          <!-- Step 4 -->
          <div class="relative">
             <span class="absolute -left-[31px] bg-[#11151e] p-1">
                <span class="w-3 h-3 rounded-full bg-emerald-500 block"></span>
             </span>
             <h4 class="text-sm font-bold text-slate-200">4. Stress Relief & Balancing</h4>
             <p class="text-xs text-slate-400 mt-1">Repeated de-stressing cycles to ensure tension does not drop after first ride.</p>
          </div>
       </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'

const localePath = useLocalePath()
</script>
