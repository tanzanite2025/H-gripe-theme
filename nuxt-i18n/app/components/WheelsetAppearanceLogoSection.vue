<template>
  <div class="space-y-6">
    <!-- Intro / Recommendation Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 text-center border-t-4 border-emerald-500">
       <h3 class="text-lg font-bold tz-text-primary mb-3">
         {{ t('guidesWheelsetBuyersAppearance.recommendation.title') }}
       </h3>
       <p class="text-sm tz-text-secondary leading-relaxed max-w-2xl mx-auto mb-4">
          {{ t('guidesWheelsetBuyersAppearance.recommendation.body') }}
       </p>
       
       <button
        type="button"
        class="inline-flex items-center gap-2 px-6 py-2 rounded-full tz-surface-panel hover:tz-surface-panel text-emerald-600 text-xs font-bold uppercase tracking-wider transition-all shadow-lg hover:shadow-emerald-500/20"
        @click="goToAboutAppearance"
      >
        <span>{{ t('guidesWheelsetBuyersAppearance.recommendation.button') }}</span>
        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
      </button>
    </div>

    <!-- Options Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
       <!-- Option 1: Custom Decals (Violet) -->
       <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300 group">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-emerald-200">
             <h3 class="text-base font-bold tz-text-primary">
               {{ t('guidesWheelsetBuyersAppearance.decals.title') }}
             </h3>
          </div>
          
           <p class="text-sm tz-text-secondary leading-relaxed mb-4">
             {{ t('guidesWheelsetBuyersAppearance.decals.body') }}
          </p>
          <ul class="space-y-2 mb-4">
              <li
                v-for="(item, index) in decalsItems"
                :key="index"
                class="flex items-center gap-2 text-xs tz-text-secondary"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> {{ rt(item) }}
             </li>
          </ul>
          <p class="text-xs tz-text-muted italic">
             {{ t('guidesWheelsetBuyersAppearance.decals.note') }}
          </p>
       </div>

       <!-- Option 2: Laser Engraving (Teal) -->
       <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 hover:translate-y-[-2px] transition-transform duration-300 group">
          <div class="flex items-center gap-3 mb-4 pb-3 border-b border-emerald-200">
             <h3 class="text-base font-bold tz-text-primary">
               {{ t('guidesWheelsetBuyersAppearance.laser.title') }}
             </h3>
          </div>
          
          <div class="absolute top-5 right-5">
             <span class="inline-flex items-center gap-1 px-2 py-1 rounded tz-micro-label font-bold uppercase tracking-wider bg-emerald-50 text-emerald-600 border border-emerald-200">
                <svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z"/></svg>
                {{ t('guidesWheelsetBuyersAppearance.laser.badge') }}
             </span>
          </div>

           <p class="text-sm tz-text-secondary leading-relaxed mb-4">
              {{ t('guidesWheelsetBuyersAppearance.laser.body') }}
          </p>
          <ul class="space-y-2 mb-4">
              <li
                v-for="(item, index) in laserItems"
                :key="index"
                class="flex items-center gap-2 text-xs tz-text-secondary"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> {{ rt(item) }}
             </li>
          </ul>
          <p class="text-xs tz-text-muted italic">
             {{ t('guidesWheelsetBuyersAppearance.laser.note') }}
          </p>
       </div>
    </div>

    <!-- Gallery 1: Laser Examples -->
    <div>
        <h4 class="text-xs font-bold tz-text-muted uppercase tracking-widest mb-3 pl-1">
          {{ t('guidesWheelsetBuyersAppearance.gallery.title') }}
        </h4>
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO.webp"
            :alt="t('guidesWheelsetBuyersAppearance.gallery.detailAlt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetBuyersAppearance.gallery.detailCaption')"
            class="rounded-xl overflow-hidden shadow-lg"
          />
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO1.webp"
            :alt="t('guidesWheelsetBuyersAppearance.gallery.processAlt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetBuyersAppearance.gallery.processCaption')"
            class="rounded-xl overflow-hidden shadow-lg"
          />
          <GuideImage
            src="/public/wheelsetbuyersguide/appearancelogo/Carbon-rim-laser-engraving-LOGO2.webp"
            :alt="t('guidesWheelsetBuyersAppearance.gallery.finishedAlt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetBuyersAppearance.gallery.finishedCaption')"
            class="rounded-xl overflow-hidden shadow-lg"
          />
        </div>
    </div>

    <!-- Advanced Finishes Card -->
    <div class="rounded-2xl bg-[var(--tz-card-surface)] shadow-md p-5 md:p-6 mt-6">
       <div class="flex items-center justify-between mb-4 pb-3 border-b tz-border-subtle">
           <h3 class="text-base font-bold tz-text-primary">
             {{ t('guidesWheelsetBuyersAppearance.advanced.title') }}
           </h3>
           <span class="tz-micro-label tz-surface-panel tz-text-muted px-2 py-1 rounded border tz-border-subtle">
             {{ t('guidesWheelsetBuyersAppearance.advanced.badge') }}
           </span>
       </div>
       
        <p class="text-sm tz-text-secondary mb-6">
          {{ t('guidesWheelsetBuyersAppearance.advanced.body') }}
       </p>

       <div class="grid gap-4 sm:grid-cols-2">
          <GuideImage
            src="/company/aboutus/appearance/carbon-rim-finish1.webp"
            :alt="t('guidesWheelsetBuyersAppearance.advanced.matteAlt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetBuyersAppearance.advanced.matteCaption')"
            class="rounded-xl overflow-hidden border tz-border-subtle"
          />
          <GuideImage
            src="/company/aboutus/appearance/carbonrims-finish2.webp"
            :alt="t('guidesWheelsetBuyersAppearance.advanced.glossyAlt')"
            :zoomOnClick="true"
            :caption="t('guidesWheelsetBuyersAppearance.advanced.glossyCaption')"
            class="rounded-xl overflow-hidden border tz-border-subtle"
          />
       </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t, tm, rt } = useI18n()
const { loadPageMessages } = usePageMessages('guidesWheelsetBuyersAppearance')
await loadPageMessages(locale.value)
watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const decalsItems = computed<string[]>(() => (
  tm('guidesWheelsetBuyersAppearance.decals.items') as string[]
))
const laserItems = computed<string[]>(() => (
  tm('guidesWheelsetBuyersAppearance.laser.items') as string[]
))

const { goToAboutAppearance } = defineProps<{
  goToAboutAppearance: () => void | Promise<void>
}>()
</script>
