<template>
  <section id="hole-patterns" class="company-section hole-pattern-page">
      <div class="hole-pattern-video-description-row">
         <div class="hole-pattern-video-description min-w-0">
           <div class="hole-pattern-alignment-copy">
             <p class="hole-pattern-alignment-eyebrow">{{ t('companyAboutHolePatterns.alignmentEyebrow') }}</p>
             <h1 class="hole-pattern-alignment-headline">{{ t('companyAboutHolePatterns.title') }}</h1>
             <p class="hole-pattern-alignment-introduction">{{ t('companyAboutHolePatterns.intro') }}</p>

             <div class="hole-pattern-alignment-principle">
               <h2>{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.title') }}</h2>
               <p>{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.description') }}</p>
               <p class="hole-pattern-alignment-design-limit">{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.manufacturingLimit') }}</p>
             </div>
           </div>
           <ReusablePhotoWallWithGreenDotPagination
             class="hole-pattern-drilled-rim-lacing-photo-wall"
             :images="localizedCarbonRimDrillingAndLacingPhotoWallImages"
             :images-per-page="5"
             :title="t('companyAboutHolePatterns.lacingPhotoWall.title')"
             :aria-label="t('companyAboutHolePatterns.lacingPhotoWall.title')"
             :pagination-label="t('companyAboutHolePatterns.lacingPhotoWall.paginationLabel')"
             :show-page-label="formatHolePatternLacingPhotoWallPageLabel"
           />
         </div>

         <ManufacturingProcessVideoPreviewAndAccessibleDialog
           class="hole-pattern-manufacturing-video"
           :video-source-path="holePatternAutomatedDrillingVideoMetadata.contentPath"
           :video-poster-path="holePatternAutomatedDrillingVideoMetadata.poster.path"
           :video-poster-description="t('companyAboutHolePatterns.seo.imageAlt')"
           preview-mode="poster-image"
           video-aspect-ratio="1 / 1"
           video-dialog-max-width="32rem"
           :preview-caption="t('companyAboutHolePatterns.videoCaption')"
           :open-dialog-label="t('companyAboutHolePatterns.openVideo')"
           :close-dialog-label="t('companyAboutHolePatterns.closeVideo')"
         />

         <div class="hole-pattern-drilling-plane-grid">
            <article v-for="drillingPlaneKey in holePatternDrillingPlaneKeys" :key="drillingPlaneKey" class="hole-pattern-drilling-plane-card">
               <p class="hole-pattern-drilling-plane-reference">{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.reference`) }}</p>
               <h2>{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.label`) }}</h2>
               <p class="hole-pattern-drilling-plane-description">{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.description`) }}</p>
            </article>
         </div>
      </div>

      <HolePatternSelectableStaticDrillingDirectionIllustrations class="hole-pattern-direction-examples" />

      <section class="hole-pattern-drawing-simulation-and-manufacturing-tolerances" aria-labelledby="hole-pattern-drawing-simulation-and-manufacturing-tolerances-title">
         <header class="hole-pattern-workflow-overview">
            <p class="hole-pattern-section-eyebrow">{{ t('companyAboutHolePatterns.layoutLabels.manufacturingWorkflow') }}</p>
            <h2 id="hole-pattern-drawing-simulation-and-manufacturing-tolerances-title" class="hole-pattern-section-title">
              {{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.title') }}
            </h2>
            <div class="hole-pattern-manufacturing-consistency">
               <h3>{{ t('companyAboutHolePatterns.sections.commitment.title') }}</h3>
               <p>{{ t('companyAboutHolePatterns.sections.commitment.body') }}</p>
            </div>
         </header>
         <div class="hole-pattern-workflow-details">
            <ol class="hole-pattern-drilling-workflow">
               <li v-for="(workflowStepKey, workflowStepIndex) in holePatternDrawingSimulationAndDrillingWorkflowKeys" :key="workflowStepKey" class="hole-pattern-drilling-workflow-step">
                  <span class="hole-pattern-drilling-workflow-step-number" aria-hidden="true">{{ String(workflowStepIndex + 1).padStart(2, '0') }}</span>
                  <div>
                     <h3>{{ t(`companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.steps.${workflowStepKey}.title`) }}</h3>
                     <p>{{ t(`companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.steps.${workflowStepKey}.description`) }}</p>
                  </div>
               </li>
            </ol>
            <aside class="hole-pattern-carbon-rim-interior-manufacturing-limit" aria-labelledby="hole-pattern-carbon-rim-interior-manufacturing-limit-title">
               <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <circle cx="12" cy="12" r="9" />
                  <path d="M12 11v6m0-10v1" stroke-linecap="round" />
               </svg>
               <div>
                  <h3 id="hole-pattern-carbon-rim-interior-manufacturing-limit-title">{{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.carbonRimInterior.title') }}</h3>
                  <p>{{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.carbonRimInterior.description') }}</p>
               </div>
            </aside>
         </div>
      </section>

      <section class="hole-pattern-angle-alignment-benefits" aria-labelledby="hole-pattern-angle-alignment-benefits-title">
         <header>
            <p class="hole-pattern-section-eyebrow">{{ t('companyAboutHolePatterns.layoutLabels.alignmentBenefits') }}</p>
            <h2 id="hole-pattern-angle-alignment-benefits-title" class="hole-pattern-section-title">
              {{ t('companyAboutHolePatterns.angleAlignmentBenefits.title') }}
            </h2>
            <p class="hole-pattern-misalignment-explanation">{{ t('companyAboutHolePatterns.angleAlignmentBenefits.misalignmentExplanation') }}</p>
         </header>
         <div class="hole-pattern-mechanical-benefits-grid">
            <article v-for="benefitKey in holePatternAngleAlignmentBenefitKeys" :key="benefitKey" class="hole-pattern-mechanical-risk-card">
               <svg class="hole-pattern-mechanical-benefit-icon" aria-hidden="true" viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
                  <g v-if="benefitKey === 'rimHoleLocalDamage'">
                     <path d="M4 11a24 24 0 0 1 24 0m-24 7a24 24 0 0 1 24 0M16 6v5m0 8v9" />
                     <circle cx="16" cy="15" r="4" />
                  </g>
                  <g v-else-if="benefitKey === 'spokeEndFatigue'">
                     <path d="M16 3v13m-3 0h6v12h-6zM12 19l8-2m-8 6 8-2m-8 6 8-2" />
                  </g>
                  <g v-else>
                     <path d="M5 25h22M16 3v12m-4 0h8v7h-8zM9 22h14v3H9z" />
                  </g>
               </svg>
               <h3>{{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.title`) }}</h3>
               <p class="hole-pattern-mechanical-risk-mechanism">{{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.mechanism`) }}</p>
               <p class="hole-pattern-mechanical-risk-description">{{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.description`) }}</p>
            </article>
         </div>
         <p class="hole-pattern-assembly-context">{{ t('companyAboutHolePatterns.angleAlignmentBenefits.assemblyContext') }}</p>
      </section>

      <div class="hole-pattern-customization-details-grid">
         <section class="hole-pattern-customization-scope" aria-labelledby="hole-pattern-customization-scope-title">
            <p class="hole-pattern-section-eyebrow">{{ t('companyAboutHolePatterns.layoutLabels.customization') }}</p>
            <h2 id="hole-pattern-customization-scope-title" class="hole-pattern-section-title">
              {{ t('companyAboutHolePatterns.sections.support.title') }}
            </h2>
            <dl class="hole-pattern-supported-customization-list">
               <div v-for="supportItemKey in holePatternSupportedCustomizationItemKeys" :key="supportItemKey">
                  <dt>{{ t(`companyAboutHolePatterns.sections.support.items.${supportItemKey}.label`) }}</dt>
                  <dd>{{ t(`companyAboutHolePatterns.sections.support.items.${supportItemKey}.description`) }}</dd>
               </div>
            </dl>
         </section>

         <section class="hole-pattern-customization-request-panel" aria-labelledby="hole-pattern-customization-request-title">
            <h2 id="hole-pattern-customization-request-title" class="hole-pattern-section-title">
              {{ t('companyAboutHolePatterns.sections.customization.title') }}
            </h2>
            <p class="hole-pattern-customization-request-introduction">
              {{ t('companyAboutHolePatterns.sections.customization.intro') }}
            </p>
            <ul class="hole-pattern-customization-request-checklist">
               <li v-for="requestDetailKey in holePatternCustomizationRequestDetailKeys" :key="requestDetailKey">
                  <svg aria-hidden="true" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5">
                     <circle cx="10" cy="10" r="8" />
                     <path d="m6.5 10 2.3 2.3 4.7-4.6" stroke-linecap="round" stroke-linejoin="round" />
                  </svg>
                  <span>{{ t(`companyAboutHolePatterns.sections.customization.items.${requestDetailKey}`) }}</span>
               </li>
            </ul>
            <section class="hole-pattern-drawing-consultation" aria-labelledby="hole-pattern-drawing-consultation-title">
               <h3 id="hole-pattern-drawing-consultation-title">{{ t('companyAboutHolePatterns.consultation.title') }}</h3>
               <p>{{ t('companyAboutHolePatterns.consultation.description') }}</p>
               <div class="hole-pattern-drawing-consultation-actions">
                  <NuxtLink :to="localePath('/company/contact')" class="hole-pattern-drawing-consultation-primary-action">
                     {{ t('companyAboutHolePatterns.consultation.provideDrawings') }}
                  </NuxtLink>
                  <button type="button" @click="openHolePatternDrillingConsultationChat">
                     {{ t('companyAboutHolePatterns.consultation.contactSupport') }}
                  </button>
               </div>
            </section>
         </section>
      </div>

  </section>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'
import { useChatWidget } from '~/composables/useChatWidget'
import { useCompanyAboutSectionSearchSocialAndStructuredData } from '~/composables/seo/useCompanyAboutSectionSearchSocialAndStructuredData'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'
import { holePatternAutomatedDrillingVideoMetadata } from '~/data/company/holePatternAutomatedDrillingVideoMetadata'
import { carbonRimDrillingAndLacingPhotoWallImages } from '~/data/company/carbonRimDrillingAndLacingPhotoWallImages'
import ReusablePhotoWallWithGreenDotPagination from '~/components/ReusablePhotoWallWithGreenDotPagination.vue'
import HolePatternSelectableStaticDrillingDirectionIllustrations from '~/components/company/HolePatternSelectableStaticDrillingDirectionIllustrations.vue'
import ManufacturingProcessVideoPreviewAndAccessibleDialog from '~/components/company/ManufacturingProcessVideoPreviewAndAccessibleDialog.vue'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { openChat } = useChatWidget()
const { loadPageMessages, resolvedLocale } = usePageMessages('companyAboutHolePatterns')

await loadPageMessages(locale.value)

const localizedCarbonRimDrillingAndLacingPhotoWallImages = computed(() => (
  carbonRimDrillingAndLacingPhotoWallImages.map(image => ({
    ...image,
    alt: image.alt[isSimplifiedChineseStorefrontLocale(resolvedLocale.value) ? 'zh_cn' : 'en'],
  }))
))

function formatHolePatternLacingPhotoWallPageLabel(pageNumber: number, pageCount: number): string {
  return t('companyAboutHolePatterns.lacingPhotoWall.showPage', { page: pageNumber, total: pageCount })
}

useCompanyAboutSectionSearchSocialAndStructuredData({
  sectionId: 'hole-patterns',
  messageNamespace: 'companyAboutHolePatterns',
  resolvedContentLocale: resolvedLocale,
  socialPreviewImage: holePatternAutomatedDrillingVideoMetadata.poster,
  video: holePatternAutomatedDrillingVideoMetadata,
})

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

function openHolePatternDrillingConsultationChat(): void {
  openChat({ showAgentList: true, source: 'company-about-hole-patterns' })
}

const holePatternSupportedCustomizationItemKeys = ['customCounts', 'holeLayout', 'oemOdm'] as const
const holePatternCustomizationRequestDetailKeys = ['hubDrawingAndRevision', 'rimDimensionsAndHoleCount', 'lacingAndNippleAssembly'] as const
const holePatternDrillingPlaneKeys = ['axialCapability', 'longitudinalCapability'] as const
const holePatternAngleAlignmentBenefitKeys = ['rimHoleLocalDamage', 'spokeEndFatigue', 'nippleSideLoading'] as const
const holePatternDrawingSimulationAndDrillingWorkflowKeys = ['hubDrawingGeometry', 'matchedDrillingAngles', 'manufacturingAndAssemblyTolerances'] as const
</script>

<style scoped>
.hole-pattern-page {
  display: flex;
  flex-direction: column;
  gap: clamp(3rem, 5.5vw, 5rem);
  width: 100%;
}

.hole-pattern-page > * {
  min-width: 0;
}

.hole-pattern-video-description-row {
  display: grid;
  align-items: stretch;
  gap: 2rem;
}

.hole-pattern-video-description {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.hole-pattern-drilled-rim-lacing-photo-wall {
  flex: 1;
}

.hole-pattern-manufacturing-video {
  width: min(100%, 512px);
  justify-self: center;
}

.hole-pattern-direction-examples {
  margin-bottom: 0;
}

.hole-pattern-section-eyebrow {
  margin: 0 0 0.75rem;
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.06em;
}

.hole-pattern-section-title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.35rem, 1.8vw, 1.65rem);
  font-weight: 650;
  line-height: 1.45;
  text-wrap: balance;
}

.hole-pattern-alignment-eyebrow {
  margin: 0 0 0.6rem;
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.06em;
}

.hole-pattern-alignment-headline {
  max-width: 30ch;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.65rem, 2.5vw, 2.2rem);
  font-weight: 650;
  line-height: 1.4;
  text-wrap: balance;
}

.hole-pattern-alignment-introduction {
  max-width: 44rem;
  margin: 0.8rem 0 1.25rem;
  color: var(--tz-text-secondary);
  font-size: 0.95rem;
  line-height: 1.8;
}

.hole-pattern-alignment-principle {
  padding: 0.85rem 1rem;
  border-left: 3px solid var(--tz-site-accent);
  border-radius: 0 0.6rem 0.6rem 0;
  background: var(--tz-card-surface);
}

.hole-pattern-alignment-principle h2 {
  margin: 0 0 0.3rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 600;
}

.hole-pattern-alignment-principle p,
.hole-pattern-drilling-plane-description {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.75;
}

.hole-pattern-alignment-principle .hole-pattern-alignment-design-limit {
  margin-top: 0.65rem;
  padding-top: 0.65rem;
  border-top: 1px solid var(--tz-border-subtle);
  font-size: 0.8rem;
}

.hole-pattern-drilling-plane-grid {
  display: grid;
  grid-column: 1 / -1;
  grid-template-columns: minmax(0, 1fr);
  gap: 1.5rem 3rem;
}

.hole-pattern-drilling-plane-card {
  padding-top: 1.25rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.hole-pattern-drilling-plane-reference {
  margin: 0 0 0.4rem;
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  font-weight: 600;
}

.hole-pattern-drilling-plane-card h2 {
  margin: 0 0 0.4rem;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.hole-pattern-drawing-simulation-and-manufacturing-tolerances {
  display: grid;
  gap: 2rem;
  padding-top: clamp(2rem, 4vw, 3.5rem);
  border-top: 1px solid var(--tz-border-strong);
}

.hole-pattern-manufacturing-consistency {
  max-width: 35rem;
  margin-top: 1.5rem;
}

.hole-pattern-manufacturing-consistency h3 {
  margin: 0 0 0.5rem;
  color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 600;
}

.hole-pattern-manufacturing-consistency p,
.hole-pattern-drilling-workflow-step p,
.hole-pattern-mechanical-risk-description,
.hole-pattern-customization-request-introduction,
.hole-pattern-supported-customization-list dd {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.85;
}

.hole-pattern-drilling-workflow {
  margin: 0;
  padding: 0;
  list-style: none;
}

.hole-pattern-drilling-workflow-step {
  position: relative;
  display: grid;
  grid-template-columns: 2.25rem minmax(0, 1fr);
  gap: 1rem;
  padding-bottom: 1.75rem;
}

.hole-pattern-drilling-workflow-step:not(:last-child)::before {
  position: absolute;
  top: 2.25rem;
  bottom: 0;
  inset-inline-start: calc(1.125rem - 0.5px);
  width: 1px;
  background: var(--tz-border-strong);
  content: '';
}

.hole-pattern-drilling-workflow-step-number {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 50%;
  color: var(--tz-text-accent);
  background: var(--tz-card-surface);
  font-size: 0.75rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.hole-pattern-drilling-workflow-step h3 {
  margin: 0.25rem 0 0.5rem;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.5;
}

.hole-pattern-carbon-rim-interior-manufacturing-limit {
  display: grid;
  grid-template-columns: 1.1rem minmax(0, 1fr);
  gap: 0.75rem;
  padding: 1.25rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
}

.hole-pattern-carbon-rim-interior-manufacturing-limit svg {
  width: 1.1rem;
  height: 1.1rem;
  margin-top: 0.15rem;
  color: var(--tz-text-secondary);
}

.hole-pattern-carbon-rim-interior-manufacturing-limit h3 {
  margin: 0 0 0.5rem;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  font-weight: 600;
  line-height: 1.6;
}

.hole-pattern-carbon-rim-interior-manufacturing-limit p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.85;
}

.hole-pattern-angle-alignment-benefits {
  padding: clamp(1.25rem, 3vw, 2.5rem);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
}

.hole-pattern-misalignment-explanation {
  max-width: 54rem;
  margin: 1rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.85;
}

.hole-pattern-mechanical-benefits-grid {
  display: grid;
  margin-top: 2rem;
}

.hole-pattern-mechanical-risk-card {
  min-width: 0;
  padding-block: 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.hole-pattern-mechanical-benefit-icon {
  width: 2rem;
  height: 2rem;
  margin-bottom: 1rem;
  color: var(--tz-text-accent);
}

.hole-pattern-mechanical-risk-card h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.05rem;
  font-weight: 600;
  line-height: 1.55;
}

.hole-pattern-mechanical-risk-mechanism {
  margin: 0.6rem 0 0.75rem;
  color: var(--tz-text-accent);
  font-size: 0.75rem;
  line-height: 1.7;
}

.hole-pattern-assembly-context {
  margin: 0;
  padding-top: 1.25rem;
  border-top: 1px solid var(--tz-border-subtle);
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.8;
}

.hole-pattern-customization-details-grid {
  display: grid;
  gap: 2rem;
}

.hole-pattern-customization-scope {
  min-width: 0;
}

.hole-pattern-supported-customization-list {
  margin: 1.5rem 0 0;
}

.hole-pattern-supported-customization-list > div {
  padding-block: 1.1rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.hole-pattern-supported-customization-list dt {
  margin-bottom: 0.4rem;
  color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 600;
}

.hole-pattern-customization-request-panel {
  min-width: 0;
  padding: clamp(1.25rem, 3vw, 2rem);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
}

.hole-pattern-customization-request-introduction {
  margin-top: 0.75rem;
}

.hole-pattern-customization-request-checklist {
  display: grid;
  gap: 1rem;
  margin: 1.5rem 0 0;
  padding: 0;
  list-style: none;
}

.hole-pattern-customization-request-checklist li {
  display: grid;
  grid-template-columns: 1.15rem minmax(0, 1fr);
  gap: 0.75rem;
  color: var(--tz-text-secondary);
  font-size: 0.875rem;
  line-height: 1.8;
}

.hole-pattern-customization-request-checklist svg {
  width: 1.15rem;
  height: 1.15rem;
  margin-top: 0.15rem;
  color: var(--tz-text-accent);
}

.hole-pattern-drawing-consultation {
  margin-top: 1.75rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.hole-pattern-drawing-consultation h3 {
  margin: 0 0 0.5rem;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 600;
  line-height: 1.6;
}

.hole-pattern-drawing-consultation p {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.85;
}

.hole-pattern-drawing-consultation-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 1.25rem;
}

.hole-pattern-drawing-consultation-actions a,
.hole-pattern-drawing-consultation-actions button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: var(--tz-pill-button-min-height);
  padding: var(--tz-pill-button-padding-block) 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 9999px;
  font-size: 0.875rem;
  line-height: var(--tz-pill-button-line-height);
  font-weight: 600;
  text-decoration: none;
  color: var(--tz-text-primary);
  cursor: pointer;
}

.hole-pattern-drawing-consultation-actions .hole-pattern-drawing-consultation-primary-action {
  border-color: var(--tz-action-primary);
  background: var(--tz-action-primary);
  color: var(--tz-action-primary-foreground);
}

.hole-pattern-drawing-consultation-actions :is(a, button):focus-visible {
  outline: 2px solid var(--tz-site-accent);
  outline-offset: 3px;
}

@media (min-width: 640px) {
  .hole-pattern-drilling-plane-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 768px) {
  .hole-pattern-mechanical-benefits-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    margin-block: 2rem;
  }

  .hole-pattern-mechanical-risk-card {
    padding-block: 0;
    border-top: 0;
  }

  .hole-pattern-mechanical-risk-card + .hole-pattern-mechanical-risk-card {
    padding-inline-start: 1.5rem;
    border-inline-start: 1px solid var(--tz-border-subtle);
  }

  .hole-pattern-mechanical-risk-card:not(:last-child) {
    padding-inline-end: 1.5rem;
  }

  .hole-pattern-customization-details-grid {
    grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.15fr);
    gap: 3rem;
    align-items: start;
  }

  .hole-pattern-customization-scope {
    padding-top: 0.5rem;
  }
}

@media (min-width: 1024px) {
  .hole-pattern-video-description-row {
    grid-template-columns: minmax(0, 1fr) 512px;
    column-gap: 2.5rem;
  }

  .hole-pattern-video-description {
    min-height: 512px;
  }

  .hole-pattern-drawing-simulation-and-manufacturing-tolerances {
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.65fr);
    gap: 4rem;
  }
}

</style>
