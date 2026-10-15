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
        <div class="premium-card p-6 border-l-4 border-l-[var(--tz-site-accent)]">
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
            <div class="w-1 h-6 bg-[var(--tz-site-accent)] rounded-full"></div>
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
      
      <ManufacturingProcessVideoPreviewAndAccessibleDialog
        :video-source-path="carbonRimLogoLaserEngravingVideoMetadata.contentPath"
        :video-poster-path="carbonRimLogoLaserEngravingVideoMetadata.poster.path"
        :video-poster-description="t('companyAboutAppearance.laser.videoAlt')"
        :preview-caption="t('companyAboutAppearance.laser.watchVideo')"
        :open-dialog-label="t('companyAboutAppearance.laser.watchVideo')"
        :close-dialog-label="t('companyAboutAppearance.closeVideo')"
        preview-mode="poster-image"
        video-aspect-ratio="16 / 9"
      />
    </div>

    <section class="premium-card mb-16" aria-labelledby="appearance-custom-options-title">
      <header class="p-6 tz-text-primary">
        <h2 id="appearance-custom-options-title" class="text-lg font-bold">
          {{ t('companyAboutAppearance.sections.customAppearance') }}
        </h2>
        <span class="block text-sm font-normal tz-text-secondary mt-2 leading-relaxed">
          {{ t('companyAboutAppearance.customizationLeadTime') }}
        </span>
      </header>
      <div class="px-6 pb-6 space-y-8">
        <div class="appearance-customization-image-and-description-row grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
          <GuideImage
            src="/company/aboutus/appearance/carbon-rim-carbon-weave-pattern-comparison-without-brake-track-options.webp"
            :alt="t('companyAboutAppearance.images.weavePatternsAlt')"
            :zoomOnClick="true"
            :caption="t('companyAboutAppearance.images.weavePatternsCaption')"
            aspectRatio="16 / 5"
            class="min-w-0 rounded-xl overflow-hidden border tz-border-subtle"
          />
          <div class="min-w-0 lg:col-span-2">
            <h3 id="appearance-weave-patterns" class="appearance-section-anchor text-lg font-bold tz-text-primary mb-4">
              {{ t('companyAboutAppearance.sections.weavePatterns') }}
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-4">
              <div v-for="key in customWeavePatternKeys" :key="key" class="premium-card p-4">
                <h4 class="font-bold tz-text-primary mb-2">
                  {{ t(`companyAboutAppearance.weaves.${key}.name`) }}
                </h4>
                <p class="text-sm tz-text-secondary">
                  {{ t(`companyAboutAppearance.weaves.${key}.description`) }}
                </p>
              </div>
            </div>
          </div>
        </div>
        <div class="appearance-customization-image-and-description-row grid grid-cols-1 lg:grid-cols-3 gap-6 items-start">
          <GuideImage
            src="/company/aboutus/appearance/carbonrims-finish2.webp"
            :alt="t('companyAboutAppearance.images.surfaceFinishesAlt')"
            :zoomOnClick="true"
            :caption="t('companyAboutAppearance.images.surfaceFinishesCaption')"
            class="min-w-0 rounded-xl overflow-hidden border tz-border-subtle"
          />
          <div class="min-w-0 lg:col-span-2">
            <h3 id="appearance-gloss-levels" class="appearance-section-anchor text-lg font-bold tz-text-primary mb-4">
              {{ t('companyAboutAppearance.sections.glossLevels') }}
            </h3>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
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
        </div>
      </div>
    </section>

    <p class="text-sm tz-text-secondary leading-relaxed">
      {{ t('companyAboutAppearance.selectionAdvice.footer') }}
    </p>

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


  </section>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import GuideImage from '~/components/GuideImage.vue'
import ManufacturingProcessVideoPreviewAndAccessibleDialog from '~/components/company/ManufacturingProcessVideoPreviewAndAccessibleDialog.vue'
import { carbonRimLogoLaserEngravingVideoMetadata } from '~/data/company/carbonRimLogoLaserEngravingVideoMetadata'
import { usePageMessages } from '~/composables/usePageMessages'
import { useCompanyAboutSectionSearchSocialAndStructuredData } from '~/composables/seo/useCompanyAboutSectionSearchSocialAndStructuredData'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages, resolvedLocale } = usePageMessages('companyAboutAppearance')

await loadPageMessages(locale.value)

useCompanyAboutSectionSearchSocialAndStructuredData({
  sectionId: 'appearance',
  messageNamespace: 'companyAboutAppearance',
  resolvedContentLocale: resolvedLocale,
  socialPreviewImage: carbonRimLogoLaserEngravingVideoMetadata.poster,
  video: carbonRimLogoLaserEngravingVideoMetadata,
})

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

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

const appearanceRelatedResources = [
  { key: 'buyingGuide', path: '/guides/wheelset-buyers/appearance-logo' },
  { key: 'oemCustomization', path: '/company/oem-odm' },
  { key: 'customizationContact', path: '/company/contact' },
] as const
</script>

<style scoped>
.appearance-section-anchor {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height, 6rem) + 1rem);
}
</style>
