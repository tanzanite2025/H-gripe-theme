<template>
  <div class="installation-guide space-y-12">
    <!-- Intro / Overview Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center">
      <h2 class="text-xl font-bold tz-text-secondary mb-4 flex items-center justify-center gap-2">
        Tubeless Installation Guide
      </h2>
      <p class="tz-text-secondary text-sm leading-relaxed mb-6 max-w-3xl mx-auto">
         The tubeless system relies on an airtight seal between the tire, rim, and valve, replacing the traditional inner tube.
         <br>
         The process involves 3 main steps: <strong>Installation</strong>, <strong>Sealant Filling</strong>, and <strong>Inflation</strong>.
      </p>
      
      <div class="flex justify-center mb-4">
         <button
            type="button"
            class="inline-flex items-center justify-center rounded-full tz-surface-panel border tz-border-subtle px-6 py-2 text-xs font-bold uppercase tracking-wider tz-text-secondary hover:tz-surface-subtle transition-colors"
            @click="$emit('changeTab', 'tubeless')"
          >
            View Required Accessories
          </button>
      </div>

       <div class="mt-8">
        <h4 class="text-xs font-bold uppercase tracking-wider tz-text-muted mb-4">Essential Components</h4>
        <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle tubeless tire pad.webp"
            alt="Bicycle tubeless tire pad"
            :zoomOnClick="true"
            caption="Tire Pad"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle vacuum self-replenishing fluid.webp"
            alt="Bicycle cycling sealant"
            :zoomOnClick="true"
            caption="Sealant"
          />
          <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/installation/Bicycle vacuum valve.webp"
            alt="Bicycle tubeless valve"
            :zoomOnClick="true"
             caption="Tubeless Valve"
          />
        </div>
      </div>
    </div>
    
    <!-- Timeline Steps -->
    <div class="relative space-y-8 before:absolute before:inset-0 before:ml-5 before:h-full before:w-0.5 before:-translate-x-px before:bg-slate-300 md:before:mx-auto md:before:translate-x-0">
      
      <!-- Step 0: Preparation -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] tz-surface-panel tz-text-secondary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">0</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">Preparation & Rim Taping</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <ul class="space-y-2 list-none pl-0">
                 <li class="flex gap-2"><span class="text-emerald-600">1.</span> Clean the rim bed thoroughly.</li>
                 <li class="flex gap-2"><span class="text-emerald-600">2.</span> Apply tubeless tape tightly, covering spoke holes.</li>
                 <li class="flex gap-2"><span class="text-emerald-600">3.</span> Overlap tape by ~10cm at the valve hole.</li>
                 <li class="flex gap-2"><span class="text-emerald-600">4.</span> Pierce a small hole for the valve and tighten the nut by hand only.</li>
               </ul>

               <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/Bicycle-tubeless-tire-pads.webp"
                    alt="Taping the rim"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/howtoopenthe vacuumnozzle.webp"
                    alt="Valve hole preparation"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/tubeless-tirepad-interface-overlaps.webp"
                    alt="Tape overlap"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 1: Install Tire -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">1</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">Install the Tire</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <p>Install the tire like a standard clincher. Push the bead into the center channel (drop center) of the rim to make mounting easier. "Massage" the tire to ensure even seating.</p>
               <div class="bg-emerald-50 p-3 rounded-lg border border-emerald-200 text-xs">
                  <strong>Pro Tip:</strong> Use soapy water or mounting fluid on the bead to help it slip into place.
               </div>

                <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/first-step-to-install-tubelesstires.webp"
                    alt="Installing tire bead"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/Tubelesstire-installation-diagram-2.webp"
                    alt="Seating the tire"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/3/Picture-of-whether-the-vacuum-nozzle-is-installed-properly.webp"
                    alt="Valve check"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 2: Add Sealant -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">2</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">Inject Sealant</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <ul class="space-y-2 list-none pl-0">
                 <li class="flex gap-2"><span class="text-emerald-600">1.</span> Remove the valve core.</li>
                 <li class="flex gap-2"><span class="text-emerald-600">2.</span> Inject ~30-60ml (Road) of sealant through the valve.</li>
                 <li class="flex gap-2"><span class="text-emerald-600">3.</span> Re-install the valve core.</li>
               </ul>

                <div class="grid grid-cols-2 lg:grid-cols-3 gap-3 mt-4">
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant1.webp"
                    alt="Remove valve core"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Pour-vacuum-tire-sealant2.webp"
                    alt="Inject sealant"
                    :zoomOnClick="true"
                  />
                  <GuideImage
                    class="rounded-lg overflow-hidden shadow-md aspect-square object-cover"
                    src="/public/tiresizecharts/installation/4/Separate-vacuum-nozzle-tool.webp"
                    alt="Valve tool"
                    :zoomOnClick="true"
                  />
               </div>
            </div>
         </div>
      </div>

       <!-- Step 3: Inflate -->
      <div class="relative flex items-center justify-between md:justify-normal md:odd:flex-row-reverse group is-active">
         <!-- Icon/Marker -->
         <div class="flex items-center justify-center w-10 h-10 rounded-full border-4 border-[var(--tz-card-surface)] bg-emerald-500 tz-text-primary shadow shrink-0 md:order-1 md:group-odd:-translate-x-1/2 md:group-even:translate-x-1/2">
            <span class="font-bold text-sm">3</span>
         </div>
         <!-- Card -->
         <div class="w-[calc(100%-4rem)] md:w-[calc(50%-2.5rem)] rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 p-6">
            <h3 class="text-lg font-bold tz-text-secondary mb-3">Inflate & Shake</h3>
          <div class="tz-text-secondary text-sm leading-relaxed space-y-4">
               <p>Inflate quickly (using a pump with a reservoir or CO2 helps) to snap the beads into place. You should hear a satisfying "pop".</p>
               <p>Shake the wheel to distribute sealant evenly.</p>
               <div class="flex items-center justify-center gap-2 text-emerald-600 font-bold mt-2">
                  <span>✓</span> <span>Ready to Roll!</span>
               </div>
            </div>
         </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import GuideImage from '~/components/GuideImage.vue'

// Match SizeChartsTabId from parent
type GuideTabId = 'size' | 'match' | 'tubeless' | 'installation' | 'choose' | 'rims' | 'tube'

defineEmits<{
  (e: 'changeTab', tab: GuideTabId): void
}>()
</script>
