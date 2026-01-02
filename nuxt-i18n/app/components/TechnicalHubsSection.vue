<template>
  <div class="space-y-6">
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
        :class="{ 'nav-pill-item--active': activeTab === 'choose-hubs' }"
        @click="activeTab = 'choose-hubs'"
      >
        Choose hubs
      </button>
    </div>

    <!-- Tab: Technology (Existing Content) -->
    <div v-show="activeTab === 'technology'" class="space-y-8">
    <!-- Intro Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-slate-500">
       <h3 class="text-lg font-bold text-slate-100 mb-2">Bicycle Hub Key Specifications</h3>
       <p class="text-sm text-slate-400 leading-relaxed max-w-2xl mx-auto">
          Understanding the core specifications of your hubs is critical for wheel building and frame compatibility.
       </p>
    </div>

    <!-- 1. Spoke Interface Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6">
       <div class="flex items-center gap-3 mb-6 pb-3 border-b border-sky-500/10">
          <h3 class="text-lg font-bold text-slate-200">1. Spoke Type: J-Bend vs. Straight-Pull</h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-2 gap-8">
          <!-- J-Bend -->
          <div class="bg-indigo-500/5 rounded-xl p-4 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
              <h4 class="text-indigo-400 font-bold mb-3 flex items-center gap-2">
                 J-Bend (Classic)
              </h4>
              <ul class="space-y-2 text-sm text-slate-300 mb-4">
                 <li>• Spoke head is bent (elbow) to fit flange.</li>
                 <li>• <strong>Pros:</strong> Easy to service, replacement spokes widely available.</li>
                 <li>• <strong>Cons:</strong> Traditional look, stress concentration at elbow.</li>
              </ul>
              <div class="grid grid-cols-2 gap-3">
                <GuideImage
                    src="/public/technical/hubs/hubs/tanzanite-dt240-6bolt-jbend-hub.webp"
                    alt="Tanzanite DT240 6-bolt J-bend hub"
                    :zoomOnClick="true"
                    caption="6-Bolt J-Bend"
                    class="rounded-lg"
                />
                <GuideImage
                    src="/public/technical/hubs/hubs/tanzanite-dt240-centerlock-jbend-hub.webp"
                    alt="Tanzanite DT240 Center Lock J-bend hub"
                    :zoomOnClick="true"
                    caption="Center Lock J-Bend"
                    class="rounded-lg"
                />
              </div>
          </div>

          <!-- Straight-Pull -->
          <div class="bg-emerald-500/5 rounded-xl p-4 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
              <h4 class="text-emerald-400 font-bold mb-3 flex items-center gap-2">
                 Straight-Pull (Modern)
              </h4>
               <ul class="space-y-2 text-sm text-slate-300 mb-4">
                 <li>• Spokes are straight, no elbow.</li>
                 <li>• <strong>Pros:</strong> Higher tension potential, sleek aesthetic.</li>
                 <li>• <strong>Cons:</strong> Harder to find specific replacements on the road.</li>
              </ul>
              <div class="grid grid-cols-2 gap-3">
                <GuideImage
                    src="/public/technical/hubs/hubs/tanzanite-dt240-6bolt-straightpull-hub.webp"
                    alt="Tanzanite DT240 6-bolt straight-pull hub"
                    :zoomOnClick="true"
                    caption="6-Bolt SP"
                    class="rounded-lg"
                />
                <GuideImage
                    src="/public/technical/hubs/hubs/tanzanite-dt240-centerlock-straightpull-hub.webp"
                    alt="Tanzanite DT240 Center Lock straight-pull hub"
                    :zoomOnClick="true"
                    caption="Center Lock SP"
                    class="rounded-lg"
                />
              </div>
          </div>
       </div>
    </div>

    <!-- 2. Tech Specs Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- Axle Specs -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-amber-500 mb-4 flex items-center gap-2">
              Axle Standards
           </h3>
           <div class="space-y-4 text-sm text-slate-300">
              <div class="bg-slate-800/50 p-3 rounded-lg">
                 <strong class="block text-slate-200 mb-1">Front Hubs</strong>
                 <p>QR (9x100), 12x100 (Road/Gravel), 15x100, <strong>15x110 (Boost)</strong></p>
              </div>
              <div class="bg-slate-800/50 p-3 rounded-lg">
                 <strong class="block text-slate-200 mb-1">Rear Hubs</strong>
                 <p>QR (135mm), 12x142, <strong>12x148 (Boost)</strong></p>
              </div>
              <p class="text-xs text-slate-500 mt-2 italic">
                 Note: Thru-axles (12mm/15mm) provide superior stiffness compared to Quick Release.
              </p>
           </div>
       </div>

       <!-- Brake Interface -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-rose-500 mb-4 flex items-center gap-2">
              Brake Interface
           </h3>
           <ul class="space-y-3 text-sm text-slate-300">
              <li class="flex items-start gap-3">
                 <span class="mt-1 w-1.5 h-1.5 rounded-full bg-rose-500 shrink-0"></span>
                 <div>
                    <strong class="text-slate-200 block">6-Bolt (IS)</strong>
                    Rotor attached via 6 Torx bolts. Universal and robust.
                 </div>
              </li>
              <li class="flex items-start gap-3">
                 <span class="mt-1 w-1.5 h-1.5 rounded-full bg-rose-500 shrink-0"></span>
                 <div>
                    <strong class="text-slate-200 block">Center Lock (CL)</strong>
                    Splined interface with a lockring. Faster installation, lighter, favored by Shimano/Road.
                 </div>
              </li>
           </ul>
       </div>
    </div>

    <!-- 3. Ratchet & Holes -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- Hole Count -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-fuchsia-500 mb-4 flex items-center gap-2">
              Hole Count Logic
           </h3>
           <div class="space-y-4">
              <div class="flex items-center justify-between text-sm border-b border-white/5 pb-2">
                  <span class="text-slate-400">Lightweight / Road</span>
                  <span class="text-slate-200 font-bold">20H / 24H</span>
              </div>
              <div class="flex items-center justify-between text-sm border-b border-white/5 pb-2">
                  <span class="text-slate-400">Standard MTB / Gravel</span>
                  <span class="text-slate-200 font-bold">28H / 32H</span>
              </div>
              <div class="flex items-center justify-between text-sm pb-2">
                  <span class="text-slate-400">Heavy Duty / E-Bike</span>
                  <span class="text-slate-200 font-bold">32H / 36H</span>
              </div>
           </div>
       </div>

       <!-- Ratchet System -->
       <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300">
           <h3 class="text-lg font-bold text-cyan-500 mb-4 flex items-center gap-2">
              Ratchet Engagement
           </h3>
           <p class="text-sm text-slate-300 mb-4">
              Higher tooth counts = faster engagement (less lag when you start pedaling).
           </p>
           <div class="grid grid-cols-3 gap-2 text-center">
              <div class="bg-cyan-500/10 rounded-lg p-2 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <div class="text-lg font-bold text-cyan-400">18T</div>
                 <div class="text-[10px] text-slate-400">Standard</div>
              </div>
              <div class="bg-cyan-500/10 rounded-lg p-2 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <div class="text-lg font-bold text-cyan-400">36T</div>
                 <div class="text-[10px] text-slate-400">Balanced</div>
              </div>
              <div class="bg-cyan-500/10 rounded-lg p-2 shadow-[0_4px_16px_rgba(0,0,0,0.5)]">
                 <div class="text-lg font-bold text-cyan-400">54T</div>
                 <div class="text-[10px] text-slate-400">Instant</div>
              </div>
           </div>
       </div>
    </div>

    <!-- 4. Freehub Summary Card (Link) -->
    <div class="rounded-2xl bg-gradient-to-r from-slate-900 to-slate-800 shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-6 flex flex-col md:flex-row items-center justify-between gap-6">
       <div class="text-center md:text-left">
          <h3 class="text-lg font-bold text-slate-100 mb-1">Freehub Body Compatibility</h3>
          <p class="text-sm text-slate-400">HG vs MS vs XD vs XDR vs N3W? Check our detailed guide.</p>
       </div>
       <button
        type="button"
        class="shrink-0 inline-flex items-center justify-center rounded-full bg-sky-500 px-6 py-2.5 text-sm font-bold text-white shadow-lg hover:bg-sky-400 hover:shadow-sky-500/25 transition-all"
        @click="goToChooseFreehub"
      >
        View Guide
      </button>
    </div>
    </div>

    <!-- Tab: Choose hubs (New Content) -->
    <div v-show="activeTab === 'choose-hubs'" class="space-y-8">
      <!-- Intro Card -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-5 md:p-6 text-center border-t-4 border-indigo-500">
         <h3 class="text-xl font-bold text-slate-100 mb-2">Hub Selection Guide</h3>
         <p class="text-sm text-slate-400 max-w-2xl mx-auto">
            Choosing the right hub is a balance between serviceability, performance, and compatibility.
         </p>
      </div>

      <!-- 1 & 2: J-bend vs Straight-pull -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
         <div class="space-y-6">
            <!-- J-bend -->
            <div class="bg-indigo-500/5 rounded-xl p-5 shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-indigo-500/10">
               <strong class="block text-indigo-400 text-lg mb-3">1. J-bend Hub</strong>
               <p class="text-sm text-slate-300 mb-2"><strong>Structure:</strong> Flange holes designed for spokes with a 90° bend (elbow).</p>
               <p class="text-sm text-slate-300 mb-2"><strong>Advantages:</strong> Traditional design, easy to source replacements, convenient maintenance.</p>
               <p class="text-sm text-slate-300 mb-4"><strong>Best for:</strong> Riders valuing universality and ease of future maintenance.</p>
               <div class="bg-indigo-500/10 px-3 py-1.5 rounded text-xs text-indigo-300 inline-block font-medium">Compatible: J-bend spokes</div>
            </div>

            <!-- Straight-pull -->
            <div class="bg-emerald-500/5 rounded-xl p-5 shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-emerald-500/10">
               <strong class="block text-emerald-400 text-lg mb-3">2. Straight-pull Hub</strong>
               <p class="text-sm text-slate-300 mb-2"><strong>Structure:</strong> Flange holes designed for spokes that insert directly without bending.</p>
               <p class="text-sm text-slate-300 mb-2"><strong>Advantages:</strong> Stiffer structure, more even tension, high-performance aesthetic.</p>
               <p class="text-sm text-slate-300 mb-4"><strong>Best for:</strong> Racing or performance use seeking rigidity and weight savings.</p>
               <div class="bg-emerald-500/10 px-3 py-1.5 rounded text-xs text-emerald-300 inline-block font-medium">Compatible: Straight-pull spokes</div>
            </div>
         </div>

         <!-- Image -->
         <div class="h-full">
            <GuideImage
               src="/public/wheelsetbuyersguide/wheelcomponents/hubs/bicycle-hub-spoke-type-comparison-jbend-straightpull.webp"
               alt="Comparison of J-bend and Straight-pull hub flange designs"
               :zoomOnClick="true"
               caption="J-bend (Left) vs Straight-pull (Right) Interface"
               class="rounded-xl overflow-hidden shadow-[0_8px_30px_rgba(0,0,0,0.6)] h-full object-cover"
            />
         </div>
      </div>

      <!-- 3. Hole Count -->
      <div class="rounded-2xl bg-[#11151e] shadow-[0_4px_16px_rgba(0,0,0,0.5)] p-6">
         <h3 class="text-lg font-bold text-fuchsia-400 mb-4 flex items-center gap-2">
            3. Hole Count Selection
         </h3>
         <div class="grid md:grid-cols-2 gap-6">
            <div class="bg-slate-800/30 p-4 rounded-xl border border-slate-700/50">
               <strong class="block text-slate-200 mb-2">Lightweight (20H / 24H)</strong>
               <p class="text-sm text-slate-400 leading-relaxed">
                  Lighter weight, strictly for road bikes and climbing sets. Requires higher spoke tension and stiffer rims.
               </p>
            </div>
            <div class="bg-slate-800/30 p-4 rounded-xl border border-slate-700/50">
               <strong class="block text-slate-200 mb-2">Durability (28H / 32H)</strong>
               <p class="text-sm text-slate-400 leading-relaxed">
                  More durable, suitable for MTB, heavy loads, or touring. Distributes stress across more spokes.
               </p>
            </div>
         </div>
         <p class="mt-4 text-xs text-fuchsia-400/80 bg-fuchsia-500/5 p-3 rounded-lg border border-fuchsia-500/20">
            <strong>Key Principle:</strong> Hub hole count must match rim hole count exactly.
         </p>
      </div>

      <!-- Notes -->
      <div class="bg-amber-500/10 border-l-4 border-amber-500 p-4 rounded-r-xl">
         <h4 class="text-amber-500 font-bold mb-2 flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/></svg>
            Critical Notes
         </h4>
         <ul class="space-y-1 text-sm text-amber-500/80 list-disc pl-5 marker:text-amber-500/50">
            <li>Hub type must match spoke type (J-bend to J-bend, SP to SP).</li>
            <li>Hole count should be chosen based on riding scenario: lightweight vs durability.</li>
            <li>High-performance wheelsets often use straight-pull hubs with fewer holes; durable wheelsets often use J-bend hubs with more holes.</li>
         </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useLocalePath, useRouter } from '#imports'
import GuideImage from '~/components/GuideImage.vue'

const router = useRouter()
const localePath = useLocalePath()
const activeTab = ref<'technology' | 'choose-hubs'>('technology')

const goToChooseFreehub = async () => {
  await router.push(`${localePath('/guides/wheelset-buyers')}#choose-freehub`)
}
</script>
