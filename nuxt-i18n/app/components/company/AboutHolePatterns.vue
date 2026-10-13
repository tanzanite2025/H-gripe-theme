<template>
  <section id="hole-patterns" class="company-section">
      <!-- Drilling video and description -->
      <div class="hole-pattern-video-description-row grid grid-cols-1 md:grid-cols-[minmax(0,1fr)_minmax(0,2fr)] items-start gap-6 lg:gap-8 mb-10">
         <div class="hole-pattern-video-column min-w-0">
         <button
           type="button"
           class="hole-pattern-video-preview relative group cursor-pointer rounded-xl overflow-hidden shadow-md border tz-border-subtle w-full focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--tz-action-primary)]"
           :aria-label="t('companyAboutHolePatterns.openVideo')"
           @click="showHolePatternVideo = true"
         >
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
                  <svg aria-hidden="true" class="w-8 h-8 tz-text-primary ml-1" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
               </div>
            </div>
         </button>
         <p class="hole-pattern-video-caption mt-3 text-sm tz-text-secondary leading-relaxed">
           {{ t('companyAboutHolePatterns.videoCaption') }}
         </p>
         </div>

         <div class="hole-pattern-video-description min-w-0">
            <p class="hole-pattern-alignment-eyebrow">{{ t('companyAboutHolePatterns.alignmentEyebrow') }}</p>
            <h2 class="hole-pattern-alignment-headline">{{ t('companyAboutHolePatterns.alignmentHeadline') }}</h2>
            <p class="hole-pattern-alignment-introduction">{{ t('companyAboutHolePatterns.intro') }}</p>

            <div class="hole-pattern-alignment-principle">
               <h3>{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.title') }}</h3>
               <p>{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.description') }}</p>
               <p class="hole-pattern-alignment-design-limit">{{ t('companyAboutHolePatterns.drillingGeometry.alignmentPrinciple.manufacturingLimit') }}</p>
            </div>

            <div class="hole-pattern-drilling-plane-grid">
               <article v-for="drillingPlaneKey in holePatternDrillingPlaneKeys" :key="drillingPlaneKey" class="hole-pattern-drilling-plane-card">
                  <p class="hole-pattern-drilling-plane-reference">{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.reference`) }}</p>
                  <h3>{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.label`) }}</h3>
                  <p class="hole-pattern-drilling-plane-description">{{ t(`companyAboutHolePatterns.drillingGeometry.${drillingPlaneKey}.description`) }}</p>
               </article>
            </div>
         </div>
      </div>

      <section class="hole-pattern-drawing-simulation-and-manufacturing-tolerances mb-10" aria-labelledby="hole-pattern-drawing-simulation-and-manufacturing-tolerances-title">
         <h3 id="hole-pattern-drawing-simulation-and-manufacturing-tolerances-title" class="text-xl font-semibold tz-text-primary mb-5">
           {{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.title') }}
         </h3>
         <ol class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <li v-for="(workflowStepKey, workflowStepIndex) in holePatternDrawingSimulationAndDrillingWorkflowKeys" :key="workflowStepKey" class="hole-pattern-drilling-workflow-step bg-[var(--tz-card-surface)] p-5 rounded-xl border tz-border-subtle">
               <span class="hole-pattern-drilling-workflow-step-number" aria-hidden="true">{{ String(workflowStepIndex + 1).padStart(2, '0') }}</span>
               <h4 class="text-base font-semibold tz-text-primary mb-2">
                 {{ t(`companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.steps.${workflowStepKey}.title`) }}
               </h4>
               <p class="text-sm tz-text-secondary leading-relaxed">
                 {{ t(`companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.steps.${workflowStepKey}.description`) }}
               </p>
            </li>
         </ol>
         <div class="hole-pattern-carbon-rim-interior-manufacturing-limit">
            <h4 class="text-sm font-semibold tz-text-primary mb-2">
              {{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.carbonRimInterior.title') }}
            </h4>
            <p class="text-sm tz-text-secondary leading-relaxed">
              {{ t('companyAboutHolePatterns.drawingSimulationAndManufacturingTolerances.carbonRimInterior.description') }}
            </p>
         </div>
      </section>

      <section class="hole-pattern-angle-alignment-benefits mb-10" aria-labelledby="hole-pattern-angle-alignment-benefits-title">
         <h3 id="hole-pattern-angle-alignment-benefits-title" class="text-xl font-semibold tz-text-primary mb-3">
           {{ t('companyAboutHolePatterns.angleAlignmentBenefits.title') }}
         </h3>
         <p class="hole-pattern-misalignment-explanation text-sm tz-text-secondary leading-relaxed mb-5">
           {{ t('companyAboutHolePatterns.angleAlignmentBenefits.misalignmentExplanation') }}
         </p>
         <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <article v-for="benefitKey in holePatternAngleAlignmentBenefitKeys" :key="benefitKey" class="hole-pattern-mechanical-risk-card bg-[var(--tz-card-surface)] p-5 rounded-xl border tz-border-subtle">
               <p class="hole-pattern-mechanical-risk-mechanism">{{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.mechanism`) }}</p>
               <h4 class="text-base font-semibold tz-text-primary mb-2">
                 {{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.title`) }}
               </h4>
               <p class="text-sm tz-text-secondary leading-relaxed">
                 {{ t(`companyAboutHolePatterns.angleAlignmentBenefits.items.${benefitKey}.description`) }}
               </p>
            </article>
         </div>
         <p class="text-sm tz-text-secondary mt-3 leading-relaxed">
           {{ t('companyAboutHolePatterns.angleAlignmentBenefits.assemblyContext') }}
         </p>
      </section>

      <!-- Info Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
         <!-- Card 1: Our Support -->
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-xl border tz-border-subtle">
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
         <div class="bg-[var(--tz-card-surface)] p-6 rounded-xl border tz-border-subtle">
            <h3 class="text-lg font-bold tz-text-secondary mb-4 border-b tz-border-subtle pb-2">
              {{ t('companyAboutHolePatterns.sections.customization.title') }}
            </h3>
            <p class="text-sm tz-text-secondary mb-3">
              {{ t('companyAboutHolePatterns.sections.customization.intro') }}
            </p>
            <ul class="space-y-3 text-sm tz-text-secondary">
               <li v-for="key in holePatternCustomizationServiceItemKeys" :key="key" class="flex items-start gap-3">
                  <span class="mt-1.5 w-1.5 h-1.5 rounded-full bg-emerald-400 shrink-0"></span>
                  <span>{{ t(`companyAboutHolePatterns.sections.customization.items.${key}`) }}</span>
               </li>
            </ul>
         </div>

         <div class="hole-pattern-manufacturing-consistency md:col-span-2 border-t tz-border-subtle pt-5">
            <h3 class="text-sm font-semibold tz-text-primary mb-2">
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

const holePatternCustomizationServiceItemKeys = ['holeLayout', 'lacingCompatibility', 'oemOdm'] as const
const holePatternDrillingPlaneKeys = ['axialCapability', 'longitudinalCapability'] as const
const holePatternAngleAlignmentBenefitKeys = ['rimHoleLocalDamage', 'spokeEndFatigue', 'nippleSideLoading'] as const
const holePatternDrawingSimulationAndDrillingWorkflowKeys = ['hubDrawingGeometry', 'matchedDrillingAngles', 'manufacturingAndAssemblyTolerances'] as const
</script>

<style scoped>
.hole-pattern-alignment-eyebrow {
  margin: 0 0 0.6rem;
  color: #047857;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.06em;
}

.hole-pattern-alignment-headline {
  max-width: 30ch;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.4rem, 2vw, 1.9rem);
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
  border-left: 3px solid #059669;
  border-radius: 0 0.6rem 0.6rem 0;
  background: var(--tz-card-surface);
}

.hole-pattern-alignment-principle h3 {
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

.hole-pattern-drilling-workflow-step-number {
  display: block;
  margin-bottom: 0.65rem;
  color: #047857;
  font-size: 0.75rem;
  font-weight: 600;
}

.hole-pattern-carbon-rim-interior-manufacturing-limit {
  margin-top: 1.25rem;
  padding-left: 1rem;
  border-left: 2px solid var(--tz-border-subtle);
}

.hole-pattern-drilling-plane-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 1rem;
  margin-top: 1.25rem;
}

.hole-pattern-drilling-plane-card {
  padding-top: 0.85rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.hole-pattern-drilling-plane-reference {
  margin: 0 0 0.4rem;
  color: #047857;
  font-size: 0.75rem;
  font-weight: 600;
}

.hole-pattern-drilling-plane-card h3 {
  margin: 0 0 0.4rem;
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.hole-pattern-misalignment-explanation {
  max-width: 70rem;
}

.hole-pattern-mechanical-risk-mechanism {
  margin: 0 0 0.65rem;
  color: #047857;
  font-size: 0.75rem;
  font-weight: 600;
}

@media (min-width: 640px) {
  .hole-pattern-drilling-plane-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.hole-pattern-video-close {
  top: max(1rem, calc(1rem + var(--tz-safe-area-top, 0px)));
  right: max(1rem, calc(1rem + var(--tz-safe-area-right, 0px)));
}

.hole-pattern-video-content {
  max-height: var(--tz-mobile-safe-viewport-height, 100vh);
}
</style>
