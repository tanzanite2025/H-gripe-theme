<template>
  <section id="hole-patterns" class="company-section">
      <div class="company-section__header flex flex-col items-center justify-center text-center mb-10 w-full px-4">
        <h2 class="sr-only">{{ t('companyAboutHolePatterns.title') }}</h2>
        <p class="text-[0.95rem] tz-text-secondary max-w-4xl w-full text-center leading-relaxed">
          {{ t('companyAboutHolePatterns.intro') }}
        </p>
      </div>

      <!-- Video Section -->
      <div class="flex justify-center mb-10">
         <div class="relative group cursor-pointer rounded-xl overflow-hidden shadow-md border tz-border-subtle max-w-3xl w-full" @click="showHolePatternVideo = true">
            <video 
              class="w-full h-auto object-cover opacity-90 group-hover:opacity-100 transition-opacity duration-300"
              src="/company/ourstory/holepatterns/automated-wheel-rimhole-drilling–customizable-Patterns_medium.webm?v=2"
              muted 
              loop 
              playsinline
              autoplay
            ></video>
            <!-- Play Overlay -->
            <div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/10 transition-colors duration-300">
               <div class="w-16 h-16 rounded-full tz-surface-subtle backdrop-blur-sm flex items-center justify-center border tz-border-strong/30 group-hover:scale-110 transition-transform duration-300">
                  <svg class="w-8 h-8 tz-text-primary ml-1" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
         </div>
      </div>

      <!-- Info Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-5xl mx-auto">
         <!-- Card 1: Our Support -->
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-2xl shadow-md border tz-border-subtle">
            <h3 class="text-lg font-bold tz-text-secondary mb-4 border-b tz-border-subtle pb-2">
              {{ t('companyAboutHolePatterns.sections.support.title') }}
            </h3>
            <ul class="space-y-3 text-sm tz-text-secondary">
               <li v-for="item in supportItems" :key="item.key" class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0"></span>
                  <span>
                    <strong class="tz-text-primary">{{ item.label }}:</strong>
                    {{ item.description }}
                  </span>
               </li>
            </ul>
         </div>

         <!-- Card 2: Customization Services -->
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-2xl shadow-md border tz-border-subtle">
            <h3 class="text-lg font-bold tz-text-secondary mb-4 border-b tz-border-subtle pb-2">
              {{ t('companyAboutHolePatterns.sections.customization.title') }}
            </h3>
            <p class="text-sm tz-text-secondary mb-3">
              {{ t('companyAboutHolePatterns.sections.customization.intro') }}
            </p>
            <ul class="space-y-3 text-sm tz-text-secondary">
               <li v-for="key in customizationItemKeys" :key="key" class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0"></span>
                  <span>{{ t(`companyAboutHolePatterns.sections.customization.items.${key}`) }}</span>
               </li>
            </ul>
         </div>

         <!-- Card 3: Advantages -->
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-2xl shadow-md border tz-border-subtle">
            <h3 class="text-lg font-bold tz-text-secondary mb-4 border-b tz-border-subtle pb-2">
              {{ t('companyAboutHolePatterns.sections.automated.title') }}
            </h3>
            <p class="text-sm tz-text-secondary mb-3">
              {{ t('companyAboutHolePatterns.sections.automated.intro') }}
            </p>
            <ul class="space-y-3 text-sm tz-text-secondary">
               <li v-for="key in automatedItemKeys" :key="key" class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-amber-400 shrink-0"></span>
                  <span>{{ t(`companyAboutHolePatterns.sections.automated.items.${key}`) }}</span>
               </li>
            </ul>
         </div>

         <!-- Card 4: Our Commitment -->
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-2xl shadow-md border tz-border-subtle">
            <h3 class="text-lg font-bold tz-text-secondary mb-4 border-b tz-border-subtle pb-2">
              {{ t('companyAboutHolePatterns.sections.commitment.title') }}
            </h3>
             <p class="text-sm tz-text-secondary leading-relaxed">
               {{ t('companyAboutHolePatterns.sections.commitment.body') }}
             </p>
         </div>
      </div>

    <!-- Video Modal -->
    <div v-if="showHolePatternVideo" class="tz-standard-modal-mask fixed inset-0 z-[9999] flex items-center justify-center p-4 tz-mobile-safe-modal-mask" @click="showHolePatternVideo = false">
       <button
         type="button"
         class="tz-global-close-btn absolute top-4 right-4 z-10"
         :aria-label="t('companyAboutHolePatterns.closeVideo')"
         @click="showHolePatternVideo = false"
       >
          ×
       </button>
       <div class="hole-pattern-video-content tz-standard-modal-surface relative w-full max-w-5xl aspect-video tz-surface-card overflow-hidden" @click.stop>
          <video 
            class="w-full h-full object-contain"
            src="/company/ourstory/holepatterns/automated-wheel-rimhole-drilling–customizable-Patterns_medium.webm?v=2"
            controls
            autoplay
          ></video>
       </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('companyAboutHolePatterns')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const showHolePatternVideo = ref(false)

const supportItemKeys = ['symmetrical', 'asymmetrical', 'customCounts', 'angled'] as const
const supportItems = computed(() => supportItemKeys.map((key) => ({
  key,
  label: t(`companyAboutHolePatterns.sections.support.items.${key}.label`),
  description: t(`companyAboutHolePatterns.sections.support.items.${key}.description`),
})))

const customizationItemKeys = ['specialCounts', 'customLacing', 'exclusiveDesigns'] as const
const automatedItemKeys = ['precision', 'consistency', 'scalableCapacity'] as const
</script>

<style scoped>
.hole-pattern-video-close {
  top: max(1rem, calc(1rem + var(--tz-safe-area-top, 0px)));
  right: max(1rem, calc(1rem + var(--tz-safe-area-right, 0px)));
}

.hole-pattern-video-content {
  max-height: var(--tz-mobile-safe-viewport-height, 100vh);
}
</style>
