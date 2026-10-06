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
        <SpokeAlternatingDrillingStep
          v-else-if="activeWizardStep === 3"
          v-model:front-offset="frontAlternatingOffsetMm"
          v-model:rear-offset="rearAlternatingOffsetMm"
          v-model:front-rim-offset="frontRimOffsetMm"
          v-model:rear-rim-offset="rearRimOffsetMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokePCDStep
          v-else-if="activeWizardStep === 4"
          v-model:front-geometry="frontGeometry"
          v-model:rear-geometry="rearGeometry"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokeHoleEngagementStep
          v-else-if="activeWizardStep === 5"
          v-model:front-hole-diameter="frontHoleDiameterMm"
          v-model:rear-hole-diameter="rearHoleDiameterMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokeInterlacingStep
          v-else-if="activeWizardStep === 6"
          v-model:front-interlacing="frontInterlacing"
          v-model:rear-interlacing="rearInterlacing"
          v-model:front-compensation="frontInterlaceCompensationMm"
          v-model:rear-compensation="rearInterlaceCompensationMm"
          v-model:spoke-elongation-compensation="spokeElongationCompensationMm"
          :front-crossing="spokeWizardDraft.front.crossing"
          :rear-crossing="spokeWizardDraft.rear.crossing"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokeNippleStep
          v-else-if="activeWizardStep === 7"
          v-model:front-nipple-type="frontNippleType"
          v-model:rear-nipple-type="rearNippleType"
          v-model:front-nipple-length="frontNippleLengthMm"
          v-model:rear-nipple-length="rearNippleLengthMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
        />

        <div class="support-page__calculator-wrapper">
          <SpokeCalculatorBlueprint
            :front-config="spokeWizardDraft.front"
            :rear-config="spokeWizardDraft.rear"
          />

          <SpokeCalculatorCatalogPanel />


        </div>

        <div class="mt-10">
          <UserFeedbackThread
            threadKey="guides-spokeguides-spoke-length-calculator"
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
import SpokeHoleEngagementStep from '~/components/SpokeHoleEngagementStep.vue'
import SpokeInterlacingStep from '~/components/SpokeInterlacingStep.vue'
import SpokeNippleStep from '~/components/SpokeNippleStep.vue'
import SpokePCDStep from '~/components/SpokePCDStep.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import { usePageMessages } from '~/composables/usePageMessages'
import { useSpokeCalculatorWizard } from '~/composables/useSpokeCalculatorWizard'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import type { SpokeHeadType, SpokeInterlacing, SpokeNippleType } from '~/types/spokeCalculator'
import { definePageMeta, useAsyncData, useHead, useI18n } from '#imports'
import { computed, watch } from 'vue'
import { useApiRequest } from '~/composables/useApiRequest'

interface SpokeCalculatorEngineeringMetadata {
  model_version: string
  formula_version: string
  knowledge_as_of: string
  calculation_status: string
  source_basis: string
  limitations: string[]
}

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
  setRimOffset,
  setSpokeHoleDiameter,
  setInterlacing,
  setInterlaceCompensation,
  setSpokeElongationCompensation,
  setNippleType,
  setNippleLength,
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

const frontRimOffsetMm = computed<number | null>({
  get: () => spokeWizardDraft.front.rimOffsetMm,
  set: value => setRimOffset('front', value),
})

const rearRimOffsetMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.rimOffsetMm,
  set: value => setRimOffset('rear', value),
})

const frontHoleDiameterMm = computed<number | null>({
  get: () => spokeWizardDraft.front.spokeHoleDiameterMm,
  set: value => setSpokeHoleDiameter('front', value),
})

const rearHoleDiameterMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.spokeHoleDiameterMm,
  set: value => setSpokeHoleDiameter('rear', value),
})

const frontInterlacing = computed<SpokeInterlacing>({
  get: () => spokeWizardDraft.front.interlacing,
  set: value => setInterlacing('front', value),
})

const rearInterlacing = computed<SpokeInterlacing>({
  get: () => spokeWizardDraft.rear.interlacing,
  set: value => setInterlacing('rear', value),
})

const frontInterlaceCompensationMm = computed<number | null>({
  get: () => spokeWizardDraft.front.interlaceCompensationMm,
  set: value => setInterlaceCompensation('front', value),
})

const rearInterlaceCompensationMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.interlaceCompensationMm,
  set: value => setInterlaceCompensation('rear', value),
})

const spokeElongationCompensationMm = computed<number | null>({
  get: () => spokeWizardDraft.front.spokeElongationCompensationMm
    ?? spokeWizardDraft.rear.spokeElongationCompensationMm,
  set: value => {
    setSpokeElongationCompensation('front', value)
    setSpokeElongationCompensation('rear', value)
  },
})

const frontNippleType = computed<SpokeNippleType>({
  get: () => spokeWizardDraft.front.nippleType,
  set: value => setNippleType('front', value),
})

const rearNippleType = computed<SpokeNippleType>({
  get: () => spokeWizardDraft.rear.nippleType,
  set: value => setNippleType('rear', value),
})

const frontNippleLengthMm = computed<number | null>({
  get: () => spokeWizardDraft.front.nippleLength,
  set: value => setNippleLength('front', value),
})

const rearNippleLengthMm = computed<number | null>({
  get: () => spokeWizardDraft.rear.nippleLength,
  set: value => setNippleLength('rear', value),
})

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const { request } = useApiRequest()
const { data: spokeCalculatorEngineeringMetadata } = await useAsyncData<SpokeCalculatorEngineeringMetadata | null>(
  'spoke-calculator-engineering-metadata',
  async () => {
    const response = await request<{ data?: SpokeCalculatorEngineeringMetadata }>(
      '/spoke/metadata',
      {},
      'Spoke calculator engineering metadata is temporarily unavailable',
    )
    return response.data || null
  },
  { default: () => null },
)

definePageMeta({
  layout: 'products',
  breadcrumbLabelKey: 'resourcesSpokeCalculator.title',
  breadcrumbLabelFallback: 'Spoke Length Calculator',
  footerLabelKey: 'support.nav.spokeCalculator',
  footerLabelFallback: 'Spoke Calculator',
  footerGroupLabelFallback: 'Spoke Guides',
})

useHead(() => ({
  title: t('resourcesSpokeCalculator.title'),
  script: [{
    type: 'application/ld+json',
    children: JSON.stringify({
      '@context': 'https://schema.org',
      '@type': 'TechArticle',
      headline: t('resourcesSpokeCalculator.title'),
      proficiencyLevel: 'Expert',
      description: t('resourcesSpokeCalculator.catalog.description'),
      author: {
        '@type': 'Organization',
        name: 'Guangengwang Engineering Lab',
      },
      inLanguage: locale.value,
      ...(spokeCalculatorEngineeringMetadata.value?.model_version
        ? { version: spokeCalculatorEngineeringMetadata.value.model_version }
        : {}),
      ...(spokeCalculatorEngineeringMetadata.value?.knowledge_as_of
        ? { dateModified: spokeCalculatorEngineeringMetadata.value.knowledge_as_of }
        : {}),
      ...(spokeCalculatorEngineeringMetadata.value?.formula_version
        ? {
            additionalProperty: [{
              '@type': 'PropertyValue',
              name: 'formulaVersion',
              value: spokeCalculatorEngineeringMetadata.value.formula_version,
            }],
          }
        : {}),
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

.spoke-page__head-step {
  margin-bottom: 1.5rem;
}

 .spoke-page {
   margin: 0 auto;
   width: 100%;
   max-width: none;
 }

</style>
