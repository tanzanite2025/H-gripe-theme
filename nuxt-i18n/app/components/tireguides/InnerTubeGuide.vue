<template>
  <div class="inner-tube-guide space-y-5 md:space-y-6">
    <!-- Step 1: Search / Intro Card -->
     <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-4 md:p-5">
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
          class="inline-flex items-center justify-center rounded-full bg-gradient-to-r from-sky-400 to-indigo-500 px-7 py-2.5 text-xs font-bold uppercase tracking-wider text-slate-950 shadow-[0_4px_14px_rgba(0,0,0,0.9)] hover:shadow-[0_8px_22px_-6px_rgba(0,0,0,1)] hover:-translate-y-0.5 transition-all"
          @click="openInnerTubeSearch"
        >
          {{ $t('guidesTireguides.innerTube.actions.findTube') }}
        </button>
       </div>
       </div>
    </div>


    <!-- Step 2: Valve Selection Card -->
    <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-4 md:p-5 text-center">
       <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
         <span class="sizecharts-section__step-badge">2</span>
           <h3 class="sizecharts-section__step-title">
             {{ $t('guidesTireguides.innerTube.steps.valve.title') }}
           </h3>
       </div>

       <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-5">
          <!-- AV -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
             <strong class="sizecharts-section__card-title text-sky-400 mb-1">AV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.av.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.av.description') }}</p>
          </div>
          <!-- DV -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center">
             <strong class="sizecharts-section__card-title text-sky-400 mb-1">DV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.dv.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.dv.description') }}</p>
          </div>
          <!-- SV -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] flex flex-col items-center border border-sky-500/30">
             <strong class="sizecharts-section__card-title text-sky-400 mb-1">SV</strong>
             <span class="sizecharts-section__card-meta tz-text-muted">{{ $t('guidesTireguides.innerTube.valves.sv.type') }}</span>
                <p class="text-xs tz-text-secondary">{{ $t('guidesTireguides.innerTube.valves.sv.description') }}</p>
          </div>
       </div>

       <div class="grid gap-5 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)] lg:items-start">
        <div class="w-full max-w-xl mx-auto lg:max-w-none">
           <GuideImage
            class="rounded-xl overflow-hidden shadow-[0_4px_16px_rgba(0,0,0,0.5)]"
            src="/public/tiresizecharts/innertube/tanzanite-schwalbe-innertube-valve.webp"
            :alt="$t('guidesTireguides.innerTube.images.valveOverviewAlt')"
            :zoomOnClick="true"
          />
        </div>

        <div class="w-full">
        <!-- Compatibility Table -->
              <h4 class="text-sm font-semibold tz-text-primary mb-3 flex items-center justify-center gap-2">
            <span>🔧</span> {{ $t('guidesTireguides.innerTube.compatibility.title') }}
         </h4>
        <div class="overflow-x-auto rounded-xl bg-slate-950/40 shadow-[0_4px_16px_rgba(0,0,0,0.5)] inline-block w-full">
           <table class="min-w-full text-center text-xs sm:text-sm text-slate-100">
            <thead class="bg-slate-900/80">
              <tr>
                    <th class="px-3 py-2.5 font-semibold tz-text-primary">{{ $t('guidesTireguides.innerTube.compatibility.headers.rimHeight') }}</th>
                <th class="px-3 py-2.5 font-semibold text-sky-400">{{ $t('guidesTireguides.innerTube.compatibility.headers.recommendedValve') }}</th>
                    <th class="px-3 py-2.5 font-semibold tz-text-primary">{{ $t('guidesTireguides.innerTube.compatibility.headers.notes') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800 bg-slate-800/20">
              <tr class="hover:bg-slate-800/40 transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">≤25mm</td>
                <td class="px-3 py-2.5 text-sky-300 font-mono">40mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.standard') }}</td>
              </tr>
              <tr class="hover:bg-slate-800/40 transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">30–45mm</td>
                <td class="px-3 py-2.5 text-sky-300 font-mono">60mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.deepRim') }}</td>
              </tr>
              <tr class="hover:bg-slate-800/40 transition-colors">
                    <td class="px-3 py-2.5 font-semibold tz-text-primary">50–65mm</td>
                <td class="px-3 py-2.5 text-sky-300 font-mono">80mm</td>
                    <td class="px-3 py-2.5 tz-text-secondary text-xs">{{ $t('guidesTireguides.innerTube.compatibility.notes.aeroRim') }}</td>
              </tr>
               <tr class="hover:bg-slate-800/40 transition-colors">
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
     <div class="rounded-2xl bg-[#11151e] shadow-[0_8px_30px_rgba(0,0,0,0.6)] p-4 md:p-5 text-center">
       <div class="sizecharts-section__step-header sizecharts-section__step-header--compact">
         <span class="sizecharts-section__step-badge">3</span>
           <h3 class="sizecharts-section__step-title">
             {{ $t('guidesTireguides.innerTube.steps.model.title') }}
           </h3>
       </div>

       <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <!-- Standard -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.standard.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.standard.description') }}
              </p>
          </div>

          <!-- Air Plus -->
          <div class="bg-indigo-500/10 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] border border-indigo-500/20">
              <strong class="block text-indigo-400 text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.airPlus.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.airPlus.description') }}
              </p>
          </div>

          <!-- Extralight -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
              <strong class="block text-amber-400 text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.extralight.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.extralight.description') }}
              </p>
          </div>

           <!-- Freeride -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
                <strong class="block tz-text-primary text-sm font-bold mb-2">{{ $t('guidesTireguides.innerTube.models.freeride.title') }}</strong>
                <p class="text-xs tz-text-secondary leading-relaxed">
                {{ $t('guidesTireguides.innerTube.models.freeride.description') }}
              </p>
          </div>

           <!-- Downhill -->
          <div class="bg-slate-800/40 p-3 rounded-xl shadow-[0_4px_16px_rgba(0,0,0,0.5)] hover:bg-slate-800/60 transition-colors">
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
