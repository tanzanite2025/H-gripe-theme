<template>
  <div class="inner-tube-guide space-y-5 md:space-y-6">
    <!-- Step 1: Search / Intro Card -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-4 md:p-5">
       <div class="grid gap-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-center">
        <div class="text-center md:text-left">
          <div class="sizecharts-section__step-header sizecharts-section__step-header--compact md:justify-start">
            <span class="sizecharts-section__step-badge">1</span>
            <h3 class="sizecharts-section__step-title">
              {{ $t('guidesTireguides.innerTube.steps.size.title') }}
            </h3>
          </div>

          <p class="tz-text-secondary text-sm leading-relaxed max-w-3xl md:max-w-none">
            {{ $t('guidesTireguides.innerTube.steps.size.description') }}
          </p>
        </div>

       <div class="flex justify-center md:justify-end">
         <button
          type="button"
          class="inline-flex items-center justify-center rounded-full bg-[var(--tz-card-surface)] border tz-border-subtle px-7 py-2.5 text-xs font-bold uppercase tracking-wider tz-text-primary shadow-md hover:border-[rgba(5, 150, 105,0.28)] hover:tz-surface-subtle hover:text-[var(--tz-site-accent)] transition-colors"
          @click="openInnerTubeSearch"
        >
          {{ $t('guidesTireguides.innerTube.actions.findTube') }}
        </button>
       </div>
       </div>
    </div>


    <!-- Step 2: Valve Selection Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-4 md:p-5 text-center">
       <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
         <span class="sizecharts-section__step-badge">2</span>
           <h3 class="sizecharts-section__step-title">
             {{ $t('guidesTireguides.innerTube.steps.valve.title') }}
           </h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
          <!-- AV -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md flex flex-col items-center border tz-border-subtle">
             <strong class="sizecharts-section__card-title tz-text-primary mb-1">AV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.av.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.av.description') }}</p>
          </div>
          <!-- DV -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md flex flex-col items-center border tz-border-subtle">
             <strong class="sizecharts-section__card-title tz-text-primary mb-1">DV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.dv.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.dv.description') }}</p>
          </div>
          <!-- SV -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md flex flex-col items-center border tz-border-subtle">
             <strong class="sizecharts-section__card-title tz-text-primary mb-1">SV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.sv.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.sv.description') }}</p>
          </div>
       </div>

       <div class="grid gap-5 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:items-start">
        <div class="w-full max-w-xl mx-auto lg:max-w-none">
           <GuideImage
            class="rounded-xl overflow-hidden shadow-md"
            src="/public/tiresizecharts/innertube/schwalbe-innertube-valve.webp"
            :alt="$t('guidesTireguides.innerTube.images.valveOverviewAlt')"
            :zoomOnClick="true"
          />
        </div>

        <div class="w-full">
        <!-- Compatibility Table -->
              <h4 class="text-sm font-semibold tz-text-primary mb-3 flex items-center justify-center gap-2">
            <span>🔧</span> {{ $t('guidesTireguides.innerTube.compatibility.title') }}
         </h4>
        <div class="overflow-x-auto rounded-xl bg-[var(--tz-card-surface)] shadow-md inline-block w-full border tz-border-subtle">
           <table class="min-w-full text-center text-xs sm:text-sm tz-text-secondary">
            <thead class="tz-surface-muted">
              <tr>
                    <th class="px-3 py-2.5 font-semibold tz-text-primary">{{ $t('guidesTireguides.innerTube.compatibility.headers.rimHeight') }}</th>
                <th class="px-3 py-2.5 font-semibold tz-text-primary">{{ $t('guidesTireguides.innerTube.compatibility.headers.recommendedValve') }}</th>
                    <th class="px-3 py-2.5 font-semibold tz-text-primary">{{ $t('guidesTireguides.innerTube.compatibility.headers.notes') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y tz-border-subtle tz-surface-subtle">
              <tr class="hover:tz-surface-subtle transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">≤25mm</td>
                <td class="px-3 py-2.5 tz-text-primary font-mono">40mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.standard') }}</td>
              </tr>
              <tr class="hover:tz-surface-subtle transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">30–45mm</td>
                <td class="px-3 py-2.5 tz-text-primary font-mono">60mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.deepRim') }}</td>
              </tr>
              <tr class="hover:tz-surface-subtle transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">50–65mm</td>
                <td class="px-3 py-2.5 tz-text-primary font-mono">80mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.aeroRim') }}</td>
              </tr>
               <tr class="hover:tz-surface-subtle transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">≥70mm</td>
                <td class="px-3 py-2.5 text-amber-300 font-mono">{{ $t('guidesTireguides.innerTube.compatibility.valves.extender') }}</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.extenderNeeded') }}</td>
              </tr>
            </tbody>
           </table>
        </div>
        </div>
       </div>
    </div>

    <!-- Step 3: Tube Type Selection -->
     <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-4 md:p-5 text-center">
       <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
         <span class="sizecharts-section__step-badge">3</span>
           <h3 class="sizecharts-section__step-title">
             {{ $t('guidesTireguides.innerTube.steps.model.title') }}
           </h3>
       </div>

       <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <!-- Standard -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.standard.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.standard.description') }}
              </p>
          </div>

          <!-- Air Plus -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
              <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.airPlus.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.airPlus.description') }}
              </p>
          </div>

          <!-- Extralight -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
              <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.extralight.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.extralight.description') }}
              </p>
          </div>

           <!-- Freeride -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.freeride.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.freeride.description') }}
              </p>
          </div>

           <!-- Downhill -->
          <div class="rounded-xl bg-[var(--tz-card-surface)] p-3 shadow-md border tz-border-subtle hover:tz-surface-subtle transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.downhill.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.downhill.description') }}
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
