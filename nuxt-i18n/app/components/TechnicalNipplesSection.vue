<template>
  <div class="space-y-6">
    <!-- Tabs Header -->
    <div class="nav-pill-tabs mb-6" role="tablist">
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'brass-vs-alloy' }"
        @click="activeTab = 'brass-vs-alloy'"
      >
        Brass VS Alloy
      </button>
      <button
        type="button"
        class="nav-pill-item"
        :class="{ 'nav-pill-item--active': activeTab === 'choose-nipple' }"
        @click="activeTab = 'choose-nipple'"
      >
        Choose Nipple
      </button>
    </div>

    <!-- Tab 1: Brass VS Alloy -->
    <div v-show="activeTab === 'brass-vs-alloy'" class="space-y-8">
       <!-- Intro -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-amber-500">
          <h3 class="text-xl font-bold text-slate-100 mb-2">Brass vs. Alloy Nipples: How to Choose</h3>
          <p class="text-sm text-slate-400 max-w-2xl mx-auto">
             When configuring a wheelset, the choice between brass and alloy affects weight, durability, and aesthetics. Here is a breakdown to help you decide.
          </p>
       </div>

       <!-- Comparison Grid -->
       <div class="grid gap-6 md:grid-cols-2">
          <!-- 1. Weight -->
          <div class="bg-slate-800/30 p-5 rounded-xl border border-slate-700/50 hover:border-slate-600 transition-colors">
             <div class="flex items-center gap-3 mb-4">
                <span class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-slate-300 font-bold text-sm">1</span>
                <h4 class="text-lg font-bold text-slate-200">Weight</h4>
             </div>
             <ul class="space-y-3 text-sm text-slate-400">
                <li class="flex justify-between border-b border-slate-700/50 pb-2">
                   <span>Brass</span>
                   <span class="font-mono text-slate-200">~1.0g / pc</span>
                </li>
                 <li class="flex justify-between border-b border-slate-700/50 pb-2">
                   <span>Alloy</span>
                   <span class="font-mono text-emerald-400">~0.33g / pc</span>
                </li>
                <li class="text-xs text-slate-500 italic pt-1">
                   Impact: Significant rotational weight savings across 44-72 nipples positioned at the rim perimeter.
                </li>
             </ul>
          </div>

           <!-- 2. Color -->
          <div class="bg-slate-800/30 p-5 rounded-xl border border-slate-700/50 hover:border-slate-600 transition-colors">
             <div class="flex items-center gap-3 mb-4">
                <span class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-slate-300 font-bold text-sm">2</span>
                <h4 class="text-lg font-bold text-slate-200">Color Options</h4>
             </div>
              <ul class="space-y-3 text-sm text-slate-400">
                <li class="flex items-start gap-2">
                   <strong class="text-amber-500 shrink-0">Brass:</strong>
                   <span>Limited (Black/Silver). Black coating may fade over time.</span>
                </li>
                 <li class="flex items-start gap-2">
                   <strong class="text-fuchsia-400 shrink-0">Alloy:</strong>
                   <span>Anodized in vibrant colors (Red, Blue, Gold, Purple, etc).</span>
                </li>
             </ul>
          </div>

           <!-- 3. Strength -->
          <div class="bg-slate-800/30 p-5 rounded-xl border border-slate-700/50 hover:border-slate-600 transition-colors">
             <div class="flex items-center gap-3 mb-4">
                <span class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-slate-300 font-bold text-sm">3</span>
                <h4 class="text-lg font-bold text-slate-200">Strength & Engagement</h4>
             </div>
             <p class="text-sm text-slate-400 mb-3">
                <strong>Standard Brass (12mm):</strong> Strong overall, traditional choice.
             </p>
             <p class="text-sm text-slate-400">
                <strong>Extended Alloy (16mm+):</strong> Double Square/Hex heads allow full thread engagement, offering higher shear strength than standard 12mm nipples (which often only engage 2/3rds).
             </p>
          </div>

           <!-- 4. Building -->
          <div class="bg-slate-800/30 p-5 rounded-xl border border-slate-700/50 hover:border-slate-600 transition-colors">
             <div class="flex items-center gap-3 mb-4">
                <span class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-slate-300 font-bold text-sm">4</span>
                <h4 class="text-lg font-bold text-slate-200">Assembly</h4>
             </div>
             <ul class="space-y-2 text-sm text-slate-400 list-disc list-inside marker:text-slate-600">
                <li>Alloy has less threading friction (less wind-up).</li>
                <li>Alloy requires precise tools (4-sided/Squorx) to prevent rounding.</li>
                <li>High-end Alloy (Sapim Polyax) uses 7075-T6 heat treatment.</li>
                <li>Locking tech (Secure Lock) prevents loosening.</li>
             </ul>
          </div>

           <!-- 5. Corrosion -->
          <div class="bg-slate-800/30 p-5 rounded-xl border border-slate-700/50 hover:border-slate-600 transition-colors md:col-span-2">
             <div class="flex items-center gap-3 mb-4">
                <span class="flex h-8 w-8 items-center justify-center rounded-full bg-slate-700 text-slate-300 font-bold text-sm">5</span>
                <h4 class="text-lg font-bold text-slate-200">Corrosion Resistance</h4>
             </div>
             <div class="grid md:grid-cols-2 gap-4">
                <div class="bg-amber-500/5 p-3 rounded-lg">
                   <strong class="block text-amber-500 mb-1 text-sm">Brass</strong>
                   <p class="text-xs text-slate-400">Naturally robust against corrosion. Best for coastal/winter riding.</p>
                </div>
                 <div class="bg-indigo-500/5 p-3 rounded-lg">
                   <strong class="block text-indigo-400 mb-1 text-sm">High-End Alloy</strong>
                   <p class="text-xs text-slate-400">Cheap alloy corrodes easily. Premium alloy (Sapim) has special treatments outperforming basic plated brass.</p>
                </div>
             </div>
          </div>
       </div>

       <!-- Summary -->
       <div class="rounded-2xl bg-gradient-to-r from-emerald-900/20 to-emerald-800/20 border border-emerald-500/30 p-6">
          <h4 class="text-emerald-400 font-bold text-lg mb-4 flex items-center gap-2">
             ✅ Summary Recommendations
          </h4>
          <div class="space-y-3">
             <div class="flex items-center gap-3 bg-slate-900/50 p-3 rounded-lg">
                <span class="text-slate-300 text-sm">Lightweight & Color</span>
                <span class="flex-1 border-b border-dashed border-slate-700 mx-2"></span>
                <strong class="text-emerald-400 text-sm">Choose Alloy</strong>
             </div>
             <div class="flex items-center gap-3 bg-slate-900/50 p-3 rounded-lg">
                <span class="text-slate-300 text-sm">Max Durability</span>
                <span class="flex-1 border-b border-dashed border-slate-700 mx-2"></span>
                <strong class="text-amber-500 text-sm">Choose Brass</strong>
             </div>
             <div class="flex items-center gap-3 bg-slate-900/50 p-3 rounded-lg">
                <span class="text-slate-300 text-sm">Carbon Rims</span>
                <span class="flex-1 border-b border-dashed border-slate-700 mx-2"></span>
                <strong class="text-slate-200 text-sm">High-Quality Alloy or Brass</strong>
             </div>
          </div>
       </div>
    </div>

    <!-- Tab 2: Choose Nipple -->
    <div v-show="activeTab === 'choose-nipple'" class="space-y-10">
       
       <!-- Section 1: Deep Rims & Angles -->
       <section>
          <div class="mb-6">
            <h3 class="text-xl font-bold text-slate-100 mb-3">Rim Depth & Spoke Entry Angles</h3>
            <p class="text-sm text-slate-400 leading-relaxed max-w-3xl">
               As rims get deeper (≥40-50mm), the angle at which spokes enter the rim becomes sharper. Standard nipples may bind against the rim bed, causing spoke bending. Using <strong>Polyax (rounded head)</strong> nipples and appropriate internal/extended designs helps align the spoke line.
            </p>
          </div>
          
          <div class="grid md:grid-cols-2 gap-6">
             <div class="space-y-4">
                <div class="bg-indigo-500/5 p-4 rounded-xl border border-indigo-500/10">
                   <strong class="block text-indigo-400 text-sm mb-2">The Deep Rim Challenge</strong>
                   <p class="text-xs text-slate-400 leading-relaxed">
                      On deep rims, nipples have less room to tilt. If the nipple cannot align with the spoke hub flange, the spoke bends at the threaded section, leading to fatigue failures.
                   </p>
                </div>
                 <div class="bg-slate-800/40 p-4 rounded-xl">
                    <strong class="block text-slate-200 text-sm mb-2">Matched Drilling</strong>
                    <p class="text-xs text-slate-400 leading-relaxed">
                       We recommend rims with "matched drilling" (angled holes) or using Polyax nipples + washers to strictly minimize spoke bending.
                    </p>
                 </div>
             </div>
             
             <!-- Images -->
             <div class="grid grid-cols-2 gap-3">
                <GuideImage
                   src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-doublesquare-nipple.webp"
                   alt="Sapim Double Square Nipple"
                   :zoomOnClick="true"
                   caption="Double Square (High Engagement)"
                   class="rounded-lg shadow-lg"
                />
                 <GuideImage
                   src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-upsidedown-nipple.webp"
                   alt="Sapim Upside Down Internal Nipple"
                   :zoomOnClick="true"
                   caption="Upside Down (Internal)"
                   class="rounded-lg shadow-lg"
                />
             </div>
          </div>
       </section>
       
       <!-- Section 2: Special Nipple Types -->
       <section>
          <h3 class="text-lg font-bold text-slate-100 mb-6 flex items-center gap-2">
             <span class="w-1 h-6 bg-sky-500 rounded-full"></span>
             Advanced Nipple Options
          </h3>
          
          <div class="grid md:grid-cols-3 gap-6">
             <!-- Double Square -->
             <div class="bg-[#11151e] rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col">
                <div class="p-4 flex-1">
                   <strong class="block text-sky-400 mb-2">Double Square</strong>
                   <p class="text-xs text-slate-400 mb-3">
                      Extended head with a square tool interface. Allows tightening from inside the rim.
                   </p>
                   <ul class="text-xs text-slate-500 space-y-1 list-disc list-inside">
                      <li>Max thread engagement</li>
                      <li>Reduced head stripping</li>
                      <li>Best for most builds</li>
                   </ul>
                </div>
             </div>

             <!-- Upside Down (Internal) -->
             <div class="bg-[#11151e] rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col">
                <div class="p-4 flex-1">
                   <strong class="block text-slate-200 mb-2">Internal (Upside Down)</strong>
                   <p class="text-xs text-slate-400 mb-3">
                      Hidden inside the rim. Clean aero look.
                   </p>
                   <ul class="text-xs text-slate-500 space-y-1 list-disc list-inside">
                      <li>Aerodynamic advantage</li>
                      <li>Requires tire removal to true</li>
                      <li>Avoid on >50mm deep rims (tool access issues)</li>
                   </ul>
                </div>
             </div>
             
             <!-- Secure Lock -->
             <div class="bg-[#11151e] rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col">
                <div class="h-32 overflow-hidden bg-white/5">
                    <GuideImage
                      src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-securelock-nipple.webp"
                      alt="Sapim Secure Lock Nipple"
                      :zoomOnClick="true"
                      class="w-full h-full object-cover"
                   />
                </div>
                <div class="p-4 flex-1">
                   <strong class="block text-amber-500 mb-2">Secure Lock</strong>
                   <p class="text-xs text-slate-400">
                      Mechanical locking mechanism stamped into the nipple. Prevents loosening without messy glue.
                   </p>
                </div>
             </div>
          </div>
       </section>
       
       <!-- Section 3: Washers -->
       <section class="bg-slate-800/30 rounded-2xl p-6 border border-slate-700/30">
          <div class="flex flex-col md:flex-row gap-8 items-center">
             <div class="flex-1">
                <h3 class="text-lg font-bold text-slate-100 mb-3">Why Washers Matter</h3>
                <p class="text-sm text-slate-400 leading-relaxed mb-4">
                   Washers increase the contact area between the nipple and rim bed. They act like a micro spherical seat, allowing the nipple to better align with the spoke angle, reducing bending stress and preventing rim bed damage.
                </p>
                <div class="flex flex-wrap gap-2">
                   <span class="px-3 py-1 rounded-full bg-slate-700 text-xs text-slate-300">Load Spreading</span>
                   <span class="px-3 py-1 rounded-full bg-slate-700 text-xs text-slate-300">Angle Adaptation</span>
                   <span class="px-3 py-1 rounded-full bg-slate-700 text-xs text-slate-300">Rim Protection</span>
                </div>
             </div>
             <div class="flex gap-4 shrink-0">
                <GuideImage
                   src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/sapim-hm-washer.webp"
                   alt="Sapim HM Washer"
                   :zoomOnClick="true"
                   caption="HM Washer"
                   class="w-24 h-24 rounded-lg object-cover shadow-lg"
                />
                 <GuideImage
                   src="/public/wheelsetbuyersguide/wheelcomponents/nipple/choosenipple/tanzanite-sapim-nipple-washer.webp"
                   alt="Tanzanite Nipple Washer"
                   :zoomOnClick="true"
                   caption="Nipple + Washer"
                   class="w-24 h-24 rounded-lg object-cover shadow-lg"
                />
             </div>
          </div>
       </section>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import GuideImage from '~/components/GuideImage.vue'

const activeTab = ref<'brass-vs-alloy' | 'choose-nipple'>('brass-vs-alloy')
</script>
