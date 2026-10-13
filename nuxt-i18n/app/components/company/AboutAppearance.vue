<template>
  <section id="appearance" class="company-section">
    <div class="text-center mb-10 max-w-3xl mx-auto">
      <h1 class="text-xl sm:text-2xl font-bold tz-text-primary mb-4">
        {{ t('companyAboutAppearance.title') }}
      </h1>
      <p class="tz-text-secondary leading-relaxed">
        {{ t('companyAboutAppearance.intro') }}
      </p>
    </div>

    <div class="mb-10">
      <h2 class="text-xl font-bold tz-text-primary mb-6">
        {{ t('companyAboutAppearance.sections.primaryAppearance') }}
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div v-for="key in primaryAppearanceOptionKeys" :key="key" class="premium-card p-5">
          <span class="inline-block text-xs font-medium tz-text-secondary tz-surface-panel px-2 py-1 rounded border tz-border-subtle mb-3">
            {{ t('companyAboutAppearance.primaryAppearance.label') }}
          </span>
          <h3 class="font-bold tz-text-primary mb-2">
            {{ t(`companyAboutAppearance.primaryAppearance.${key}.name`) }}
          </h3>
          <p class="text-sm tz-text-secondary leading-relaxed">
            {{ t(`companyAboutAppearance.primaryAppearance.${key}.description`) }}
          </p>
        </div>
      </div>
      <p class="mt-4 text-sm tz-text-secondary leading-relaxed">
        {{ t('companyAboutAppearance.customizationLeadTime') }}
      </p>
    </div>

    <!-- Brake track heat resistance and surface treatment tradeoffs -->
    <div class="mb-16">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-1 h-6 bg-orange-500 rounded-full"></div>
        <h2 class="text-xl font-bold tz-text-primary">
          {{ t('companyAboutAppearance.sections.brakeTracks') }}
          <span class="text-sm font-normal tz-text-muted ml-2">
            {{ t('companyAboutAppearance.rimBrakeOnly') }}
          </span>
        </h2>
      </div>
      <div class="premium-card p-6 mb-4">
        <h3 class="font-bold tz-text-primary mb-2">
          {{ t('companyAboutAppearance.brakeTracks.resinTitle') }}
        </h3>
        <p class="text-sm tz-text-secondary leading-relaxed">
          {{ t('companyAboutAppearance.brakeTracks.resinDescription') }}
        </p>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="premium-card p-6 border-l-4 border-l-emerald-500">
          <span class="inline-block text-xs font-medium tz-text-secondary tz-surface-panel px-2 py-1 rounded border tz-border-subtle mb-3">
            {{ t('companyAboutAppearance.brakeTracks.recommendedLabel') }}
          </span>
          <h3 class="font-bold tz-text-primary mb-2">
            {{ t('companyAboutAppearance.brakeTracks.standardTitle') }}
          </h3>
          <p class="text-sm tz-text-secondary leading-relaxed">
            {{ t('companyAboutAppearance.brakeTracks.standardDescription') }}
          </p>
        </div>
        <div class="premium-card p-6">
          <span class="inline-block text-xs font-medium tz-text-secondary tz-surface-panel px-2 py-1 rounded border tz-border-subtle mb-3">
            {{ t('companyAboutAppearance.brakeTracks.optionalLabel') }}
          </span>
          <h3 class="font-bold tz-text-primary mb-2">
            {{ t('companyAboutAppearance.brakeTracks.laserTitle') }}
          </h3>
          <p class="text-sm tz-text-secondary leading-relaxed mb-3">
            {{ t('companyAboutAppearance.brakeTracks.laserBenefit') }}
          </p>
          <p class="text-sm tz-text-secondary leading-relaxed">
            {{ t('companyAboutAppearance.brakeTracks.laserTradeoff') }}
          </p>
        </div>
      </div>
    </div>

    <!-- Laser Engraving -->
    <div class="mb-16 grid grid-cols-1 lg:grid-cols-2 gap-8 lg:gap-12 items-center">
      <div>
         <div class="flex items-center gap-3 mb-4">
            <div class="w-1 h-6 bg-emerald-400 rounded-full"></div>
            <h2 class="text-xl font-bold tz-text-primary">
              {{ t('companyAboutAppearance.sections.laserEngraving') }}
            </h2>
         </div>
        <div class="text-sm tz-text-secondary leading-relaxed">
           <p>{{ t('companyAboutAppearance.laser.body') }}</p>
           <ul class="my-4 space-y-2 list-none pl-0">
             <li class="flex gap-2">
               <span aria-hidden="true" class="tz-text-muted">•</span> {{ t('companyAboutAppearance.laser.bullets.0') }}
             </li>
             <li class="flex gap-2">
               <span aria-hidden="true" class="tz-text-muted">•</span> {{ t('companyAboutAppearance.laser.bullets.1') }}
             </li>
             <li class="flex gap-2">
               <span aria-hidden="true" class="tz-text-muted">•</span> {{ t('companyAboutAppearance.laser.bullets.2') }}
             </li>
           </ul>
           <p class="tz-surface-panel p-3 rounded border tz-border-subtle">
             {{ t('companyAboutAppearance.laser.guideNote') }}
           </p>
         </div>
      </div>
      
      <!-- Video Thumbnail -->
      <button
        type="button"
        class="premium-card p-2 group cursor-pointer text-left w-full"
        :aria-label="t('companyAboutAppearance.laser.watchVideo')"
        @click="showLaserVideo = true"
      >
         <div class="relative rounded-lg overflow-hidden aspect-video tz-surface-panel">
            <img
              src="/company/aboutus/appearance/carbon-rim-laser-engraving-LOGO.webp"
              :alt="t('companyAboutAppearance.laser.videoAlt')"
              class="w-full h-full object-cover opacity-80 group-hover:opacity-100 transition-opacity duration-300"
              loading="lazy"
              width="1280"
              height="720"
            />
            <div class="absolute inset-0 flex items-center justify-center bg-black/30 group-hover:bg-black/10 transition-colors">
               <div class="w-14 h-14 flex items-center justify-center rounded-full bg-emerald-500 text-white pl-1 shadow-[0_0_30px_rgba(20,184,166,0.5)] group-hover:scale-110 transition-transform">
                 <svg class="w-6 h-6" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
            <div class="absolute bottom-0 left-0 right-0 p-4 bg-gradient-to-t from-black/90 to-transparent">
              <span class="text-sm font-medium text-white">
                {{ t('companyAboutAppearance.laser.watchVideo') }}
              </span>
            </div>
         </div>
      </button>
    </div>

    <details class="premium-card mb-16">
      <summary class="cursor-pointer p-6 tz-text-primary">
        <h2 class="inline text-lg font-bold">
          {{ t('companyAboutAppearance.sections.customAppearance') }}
        </h2>
        <span class="block text-sm font-normal tz-text-secondary mt-2 leading-relaxed">
          {{ t('companyAboutAppearance.customizationLeadTime') }}
        </span>
      </summary>
      <div class="px-6 pb-6">
        <h3 class="text-lg font-bold tz-text-primary mb-4">
          {{ t('companyAboutAppearance.sections.weavePatterns') }}
        </h3>
        <GuideImage
          src="/company/aboutus/appearance/carbon-rim-carbon-weave-pattern-comparison-without-brake-track-options.webp"
          :alt="t('companyAboutAppearance.images.weavePatternsAlt')"
          :zoomOnClick="true"
          :caption="t('companyAboutAppearance.images.weavePatternsCaption')"
          aspectRatio="16 / 5"
          class="rounded-xl overflow-hidden border tz-border-subtle mb-6"
        />
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
          <div v-for="key in customWeavePatternKeys" :key="key" class="premium-card p-4">
            <h4 class="font-bold tz-text-primary mb-2">
              {{ t(`companyAboutAppearance.weaves.${key}.name`) }}
            </h4>
            <p class="text-sm tz-text-secondary">
              {{ t(`companyAboutAppearance.weaves.${key}.description`) }}
            </p>
          </div>
        </div>
        <GuideImage
          src="/company/aboutus/appearance/carbonrims-finish2.webp"
          :alt="t('companyAboutAppearance.images.surfaceFinishesAlt')"
          :zoomOnClick="true"
          :caption="t('companyAboutAppearance.images.surfaceFinishesCaption')"
          class="rounded-xl overflow-hidden border tz-border-subtle mb-6"
        />
        <div class="grid grid-cols-1 lg:grid-cols-2 gap-8">
          <div>
            <h3 class="text-lg font-bold tz-text-primary mb-4">
              {{ t('companyAboutAppearance.sections.glossLevels') }}
            </h3>
            <div class="grid grid-cols-1 gap-3">
              <div v-for="key in customGlossLevelKeys" :key="key" class="premium-card p-4">
                <h4 class="font-bold tz-text-primary mb-2">
                  {{ t(`companyAboutAppearance.gloss.${key}.name`) }}
                </h4>
                <p class="text-sm tz-text-secondary">
                  {{ t(`companyAboutAppearance.gloss.${key}.description`) }}
                </p>
              </div>
            </div>
          </div>
          <div>
            <h3 class="text-lg font-bold tz-text-primary mb-4">
              {{ t('companyAboutAppearance.sections.crystalEffects') }}
            </h3>
            <p class="text-sm tz-text-secondary mb-4 leading-relaxed">
              {{ t('companyAboutAppearance.crystalIntro') }}
            </p>
            <div class="flex flex-wrap gap-2">
              <span v-for="key in crystalEffectKeys" :key="key" class="px-3 py-1.5 rounded-full text-sm font-medium tz-surface-panel border tz-border-subtle tz-text-secondary">
                {{ t(`companyAboutAppearance.crystalEffects.${key}`) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </details>

    <div class="premium-card p-6 border-l-4 border-l-emerald-500">
      <h2 class="text-lg font-bold tz-text-primary mb-3">
        {{ t('companyAboutAppearance.selectionAdvice.title') }}
      </h2>
      <ul class="space-y-2 text-sm tz-text-secondary list-disc pl-5 leading-relaxed">
        <li>{{ t('companyAboutAppearance.selectionAdvice.items.0') }}</li>
        <li>{{ t('companyAboutAppearance.selectionAdvice.items.1') }}</li>
        <li>{{ t('companyAboutAppearance.selectionAdvice.items.2') }}</li>
      </ul>
      <p class="mt-4 text-sm tz-text-secondary pt-3 border-t tz-border-subtle leading-relaxed">
        {{ t('companyAboutAppearance.selectionAdvice.footer') }}
      </p>
    </div>

    <div class="mt-8">
      <h2 class="text-lg font-bold tz-text-primary mb-4">
        {{ t('companyAboutAppearance.relatedResources.title') }}
      </h2>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <NuxtLink
          v-for="resource in appearanceRelatedResources"
          :key="resource.key"
          :to="localePath(resource.path)"
          class="premium-card p-5 block hover:tz-surface-panel transition-colors"
        >
          <h3 class="font-bold tz-text-primary mb-2">
            {{ t(`companyAboutAppearance.relatedResources.${resource.key}.title`) }}
          </h3>
          <p class="text-sm tz-text-secondary leading-relaxed">
            {{ t(`companyAboutAppearance.relatedResources.${resource.key}.description`) }}
          </p>
        </NuxtLink>
      </div>
    </div>

    <!-- Video Modal -->
    <div
      v-if="showLaserVideo"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 tz-mobile-safe-modal-mask"
      role="dialog"
      aria-modal="true"
      :aria-label="t('companyAboutAppearance.laser.watchVideo')"
    >
      <div class="tz-standard-modal-backdrop absolute inset-0" @click="showLaserVideo = false" />
      <div class="appearance-video-modal-content tz-standard-modal-surface relative w-full max-w-4xl tz-surface-card overflow-hidden ring-1 ring-black/10">
        <button
          type="button"
          class="tz-global-close-btn absolute top-4 right-4 z-10"
          :aria-label="t('companyAboutAppearance.closeVideo')"
          @click="showLaserVideo = false"
        >
          ×
        </button>
        <video
          class="w-full h-auto aspect-video"
          controls
          autoplay
          playsinline
        >
          <source src="/company/aboutus/appearance/carbonrim-appearancelogo.mp4" type="video/mp4" />
        </video>
      </div>
    </div>

  </section>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import { usePageMessages } from '~/composables/usePageMessages'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('companyAboutAppearance')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const showLaserVideo = ref(false)

const primaryAppearanceOptionKeys = ['ud', 'paintless', 'laserLogo'] as const

const customWeavePatternKeys = [
  'butterfly',
  'marble',
  'threeK',
  'threeKTwillVertical',
  'threeKTwillHorizontal',
  'twelveK',
  'twelveKTwillVertical',
  'twelveKTwillHorizontal',
] as const

const customGlossLevelKeys = ['matte', 'satin', 'glossy'] as const

const crystalEffectKeys = ['black', 'silver', 'orange', 'purple', 'blue'] as const

const appearanceRelatedResources = [
  { key: 'buyingGuide', path: '/guides/wheelset-buyers/appearance-logo' },
  { key: 'oemCustomization', path: '/company/oem-odm' },
  { key: 'customizationContact', path: '/company/contact' },
] as const
</script>

<style scoped>
.appearance-video-modal-content {
  max-height: var(--tz-mobile-safe-viewport-height, 100vh);
}
</style>
