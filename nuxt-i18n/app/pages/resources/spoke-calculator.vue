<template>
  <div>
    <h1 class="sr-only">{{ t('resourcesSpokeCalculator.title') }}</h1>

    <div class="spoke-page">
      <section>
        <SpokeHeadTypeStep
          v-if="activeWizardStep === 1"
          class="spoke-page__head-step"
          v-model="selectedSpokeHeadType"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @next="nextStep"
        />
        <SpokeERDStep
          v-else-if="activeWizardStep === 2"
          v-model:front-erd="frontErdMm"
          v-model:rear-erd="rearErdMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokePCDStep
          v-else-if="activeWizardStep === 3"
          v-model:front-geometry="frontGeometry"
          v-model:rear-geometry="rearGeometry"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokeAlternatingDrillingStep
          v-else-if="activeWizardStep === 4"
          v-model:front-offset="frontAlternatingOffsetMm"
          v-model:rear-offset="rearAlternatingOffsetMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
        />

        <!-- Standalone full-width reference card, above the calculator settings. -->
        <SpokePhysicsDiagrams class="spoke-page__physics-card" />

        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorBlueprint
            :front-config="spokeWizardDraft.front"
            :rear-config="spokeWizardDraft.rear"
          />

          <SpokeCalculatorCatalogPanel />


        </div>

        <div class="mt-10">
          <UserFeedbackThread
            threadKey="products-spoke-calculator"
            :title="t('resourcesSpokeCalculator.feedbackTitle')"
          />
        </div>
      </section>


    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorBlueprint from '~/components/SpokeCalculatorBlueprint.vue'
import SpokeCalculatorCatalogPanel from '~/components/SpokeCalculatorCatalogPanel.vue'
import SpokeAlternatingDrillingStep from '~/components/SpokeAlternatingDrillingStep.vue'
import SpokeERDStep from '~/components/SpokeERDStep.vue'
import SpokeHeadTypeStep from '~/components/SpokeHeadTypeStep.vue'
import SpokePCDStep from '~/components/SpokePCDStep.vue'
import SpokePhysicsDiagrams from '~/components/SpokePhysicsDiagrams.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import { usePageSubNavigationTab } from '~/composables/usePageSubNavigationTab'
import { spokeCalculatorTabs } from '~/utils/pageSubNavigation'
import { usePageMessages } from '~/composables/usePageMessages'
import { useSpokeCalculatorWizard } from '~/composables/useSpokeCalculatorWizard'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import type { SpokeHeadType } from '~/types/spokeCalculator'
import { definePageMeta, useHead, useI18n } from '#imports'
import { computed, watch } from 'vue'

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('resourcesSpokeCalculator')

const {
  draft: spokeWizardDraft,
  activeStep: activeWizardStep,
  goToStep,
  nextStep,
  previousStep,
  setHeadType,
  setErd,
  setHubGeometry,
  setAlternatingDrillingOffset,
} = useSpokeCalculatorWizard()

const selectedSpokeHeadType = computed<SpokeHeadType>({
  get: () => spokeWizardDraft.front.spokeHeadType,
  set: value => setHeadType(value),
})

const frontErdMm = computed<number | null>({
  get: () => spokeWizardDraft.front.erd,
  set: value => setErd('front', value),
})

const rearErdMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.erd,
  set: value => setErd('rear', value),
})

const frontGeometry = computed<HubGeometry>({
  get: () => ({
    leftFlange: spokeWizardDraft.front.leftFlange,
    rightFlange: spokeWizardDraft.front.rightFlange,
    leftFlangePcd: spokeWizardDraft.front.leftFlangePcd,
    rightFlangePcd: spokeWizardDraft.front.rightFlangePcd,
  }),
  set: value => setHubGeometry('front', value),
})

const rearGeometry = computed<HubGeometry>({
  get: () => ({
    leftFlange: spokeWizardDraft.rear.leftFlange,
    rightFlange: spokeWizardDraft.rear.rightFlange,
    leftFlangePcd: spokeWizardDraft.rear.leftFlangePcd,
    rightFlangePcd: spokeWizardDraft.rear.rightFlangePcd,
  }),
  set: value => setHubGeometry('rear', value),
})

const frontAlternatingOffsetMm = computed<number | null>({
  get: () => spokeWizardDraft.front.alternatingDrillingOffsetMm,
  set: value => setAlternatingDrillingOffset('front', value),
})

const rearAlternatingOffsetMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.alternatingDrillingOffsetMm,
  set: value => setAlternatingDrillingOffset('rear', value),
})

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

usePageSubNavigationTab({
  tabs: spokeCalculatorTabs,
  basePath: '/resources/spoke-calculator',
  defaultValue: 'calculator',
  redirectBasePathToDefaultTab: true,
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'support.nav.spokeCalculator',
  footerLabelFallback: 'Spoke Calculator',
})

useHead(() => ({
  title: t('resourcesSpokeCalculator.title'),
  script: [{
    type: 'application/ld+json',
    children: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'TechArticle',
      headline: t('resourcesSpokeCalculator.title'),
      version: 'V1.0-ENGINEERING',
      proficiencyLevel: 'Expert',
      author: {
        '@type': 'Organization',
        name: 'Guangengwang Engineering Lab',
      },
      inLanguage: locale.value,
      hasPart: [{
        '@type': 'Dataset',
        name: 'Spoke calculator engineering dataset',
        description: 'Server-side spoke length and tension-ratio calculations for validated rim and hub geometry.',
      }],
    }),
  }],
}))
</script>

<style src="~/assets/css/guide-sections.css"></style>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.support-page__calculator-wrapper {
  margin-top: 1.5rem;
}

.spoke-page__physics-card {
  margin-bottom: 1.5rem;
}

.spoke-page__head-step {
  margin-bottom: 1.5rem;
}

 .spoke-page {
   margin: 0 auto;
   width: 100%;
   max-width: none;
 }

</style>
