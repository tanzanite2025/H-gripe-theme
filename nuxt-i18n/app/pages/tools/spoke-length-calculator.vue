<template>
  <div>
    <div class="spoke-page">
      <header class="spoke-page__intro">
        <h1 class="spoke-page__title">{{ t('resourcesSpokeCalculator.title') }}</h1>
        <p class="spoke-page__description">{{ t('resourcesSpokeCalculator.seo.description') }}</p>
      </header>

      <section>
        <SpokeHeadTypeStep
          v-if="activeWizardStep === 1"
          class="spoke-page__head-step"
          v-model="selectedSpokeHeadType"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @next="nextStep"
        />
        <SpokeRimGeometryStep
          v-else-if="activeWizardStep === 2"
          v-model:front-erd="frontErdMm"
          v-model:rear-erd="rearErdMm"
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
          v-else-if="activeWizardStep === 3"
          v-model:front-geometry="frontGeometry"
          v-model:rear-geometry="rearGeometry"
          :front-config="spokeWizardDraft.front"
          :rear-config="spokeWizardDraft.rear"
          :options="spokeCalculatorOptions"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />
        <SpokePhysicalCorrectionsStep
          v-else-if="activeWizardStep === 4"
          v-model:front-hole-diameter="frontHoleDiameterMm"
          v-model:rear-hole-diameter="rearHoleDiameterMm"
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
          v-else-if="activeWizardStep === 5"
          v-model:front-nipple-type="frontNippleType"
          v-model:rear-nipple-type="rearNippleType"
          v-model:front-nipple-length="frontNippleLengthMm"
          v-model:rear-nipple-length="rearNippleLengthMm"
          :current-step="activeWizardStep"
          @select-step="goToStep"
          @previous="previousStep"
          @next="nextStep"
        />

        <div class="support-page__calculator-wrapper">
          <div v-show="activeWizardStep === 6" class="spoke-result-step">
            <SpokeCalculatorBlueprint
              :front-config="spokeWizardDraft.front"
              :rear-config="spokeWizardDraft.rear"
              :current-step="activeWizardStep"
              @select-step="goToStep"
              @previous="previousStep"
            >
              <template #intro>
                <header class="spoke-result-step__intro">
                  <span class="spoke-result-step__number" aria-hidden="true">06</span>
                  <div>
                    <h2 class="spoke-result-step__title">
                      {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSixPrompt') }}
                    </h2>
                    <p class="spoke-result-step__subtitle">
                      {{ t('resourcesSpokeCalculator.calculator.physicalCorrections.stepSixSubtitle') }}
                    </p>
                  </div>
                </header>
              </template>
            </SpokeCalculatorBlueprint>
          </div>

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
import SpokeHeadTypeStep from '~/components/SpokeHeadTypeStep.vue'
import SpokeNippleStep from '~/components/SpokeNippleStep.vue'
import SpokePCDStep from '~/components/SpokePCDStep.vue'
import SpokePhysicalCorrectionsStep from '~/components/SpokePhysicalCorrectionsStep.vue'
import SpokeRimGeometryStep from '~/components/SpokeRimGeometryStep.vue'
import UserFeedbackThread from '~/components/UserFeedbackThread.vue'

import { useStorefrontSeoRouteOverride } from '~/composables/seo/useStorefrontSeoLinks'
import { usePageMessages } from '~/composables/usePageMessages'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useSpokeCalculatorManualOptions } from '~/composables/useSpokeCalculatorManualOptions'
import { useSpokeCalculatorWizard } from '~/composables/useSpokeCalculatorWizard'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import type { HubGeometry } from '~/data/spoke-calculator/database'
import localeManifest from '~/i18n/locales.manifest'
import type { SpokeHeadType, SpokeInterlacing, SpokeNippleType } from '~/types/spokeCalculator'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { definePageMeta, useAsyncData, useHead, useI18n, useRoute, useSwitchLocalePath } from '#imports'
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
const spokeCalculatorRoute = useRoute()
const switchLocalePath = useSwitchLocalePath()
const spokeCalculatorPublishedLocaleCodes = new Set(['en', 'zh_cn'])
const spokeCalculatorPublishedSeoRoutes = computed(() => [...spokeCalculatorPublishedLocaleCodes].map(code => ({
  code,
  path: new URL(switchLocalePath(code as any) || '/tools/spoke-length-calculator', 'https://spoke-length-seo.invalid').pathname,
})))
const spokeCalculatorSeoRouteOverride = useStorefrontSeoRouteOverride(spokeCalculatorPublishedSeoRoutes)
// Shared SEO consumers clear overrides during client navigation. Restore this
// page's published languages after those consumers process query/hash changes.
watch(() => spokeCalculatorRoute.fullPath, () => {
  spokeCalculatorSeoRouteOverride.value = spokeCalculatorPublishedSeoRoutes.value
}, { flush: 'post' })
useHead(() => ({
  meta: [{
    key: 'robots',
    name: 'robots',
    content: spokeCalculatorPublishedLocaleCodes.has(locale.value) ? 'index,follow' : 'noindex,follow',
  }],
}))
const { loadPageMessages } = usePageMessages('resourcesSpokeCalculator')
const { siteTitle } = useSiteTitle()
const { canonicalUrl } = useStorefrontSeoLinks()
const spokeCalculatorSearchTitle = computed(() => {
  const calculatorTitle = t('resourcesSpokeCalculator.seo.title')
  const configuredBrandName = siteTitle.value.trim()
  return configuredBrandName ? `${calculatorTitle} | ${configuredBrandName}` : calculatorTitle
})
const spokeCalculatorLanguage = computed(() => (
  localeManifest.find(entry => entry.code === locale.value)?.iso || locale.value.replace(/_/g, '-')
))
const spokeCalculatorSocialImageUrl = computed(() => (
  new URL('/public/technical/j-bend-spokeandstraightpull-spoke.webp', canonicalUrl.value).toString()
))

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

const { options: spokeCalculatorOptions } = useSpokeCalculatorManualOptions()

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
  path: '/tools/spoke-length-calculator',
  layout: 'products',
  breadcrumbLabelKey: 'resourcesSpokeCalculator.title',
  breadcrumbLabelFallback: 'Spoke Length Calculator',
  footerLabelKey: 'toolsNavigation.spokeLengthCalculatorLabel',
  footerLabelFallback: 'Spoke length calculator',
})

useHead(() => ({
  title: spokeCalculatorSearchTitle.value,
  meta: [
    { key: 'description', name: 'description', content: t('resourcesSpokeCalculator.seo.description') },
    { key: 'og:type', property: 'og:type', content: 'website' },
    { key: 'og:title', property: 'og:title', content: spokeCalculatorSearchTitle.value },
    { key: 'og:description', property: 'og:description', content: t('resourcesSpokeCalculator.seo.description') },
    { key: 'og:url', property: 'og:url', content: canonicalUrl.value },
    { key: 'og:locale', property: 'og:locale', content: spokeCalculatorLanguage.value.replace(/-/g, '_') },
    ...(siteTitle.value.trim()
      ? [{ key: 'og:site_name', property: 'og:site_name', content: siteTitle.value.trim() }]
      : []),
    { key: 'og:image', property: 'og:image', content: spokeCalculatorSocialImageUrl.value },
    { key: 'og:image:type', property: 'og:image:type', content: 'image/webp' },
    { key: 'og:image:width', property: 'og:image:width', content: '1280' },
    { key: 'og:image:height', property: 'og:image:height', content: '720' },
    { key: 'og:image:alt', property: 'og:image:alt', content: t('resourcesSpokeCalculator.seo.imageAlt') },
    { key: 'twitter:card', name: 'twitter:card', content: 'summary_large_image' },
    { key: 'twitter:title', name: 'twitter:title', content: spokeCalculatorSearchTitle.value },
    { key: 'twitter:description', name: 'twitter:description', content: t('resourcesSpokeCalculator.seo.description') },
    { key: 'twitter:image', name: 'twitter:image', content: spokeCalculatorSocialImageUrl.value },
    { key: 'twitter:image:alt', name: 'twitter:image:alt', content: t('resourcesSpokeCalculator.seo.imageAlt') },
  ],
  script: [{
    key: 'spoke-length-calculator-web-application-schema',
    id: 'spoke-length-calculator-web-application-schema',
    ...createSeoJsonLdScript({
      '@context': 'https://schema.org',
      '@type': 'WebApplication',
      '@id': `${canonicalUrl.value}#spoke-length-calculator`,
      url: canonicalUrl.value,
      mainEntityOfPage: canonicalUrl.value,
      name: t('resourcesSpokeCalculator.title'),
      description: t('resourcesSpokeCalculator.seo.description'),
      image: spokeCalculatorSocialImageUrl.value,
      applicationCategory: 'EngineeringApplication',
      operatingSystem: 'Web',
      isAccessibleForFree: true,
      offers: {
        '@type': 'Offer',
        price: '0',
        priceCurrency: 'USD',
        url: canonicalUrl.value,
      },
      featureList: [
        t('resourcesSpokeCalculator.calculator.physicalCorrections.straightPull'),
        t('resourcesSpokeCalculator.calculator.physicalCorrections.jBend'),
        t('resourcesSpokeCalculator.calculator.physicalCorrections.rimOffsetFieldLabel'),
        t('resourcesSpokeCalculator.calculator.physicalCorrections.elongationCompensationLabel'),
        t('resourcesSpokeCalculator.calculator.physicalCorrections.interlaceCompensation'),
      ],
      author: {
        '@type': 'Organization',
        name: 'Guangengwang Engineering Lab',
      },
      inLanguage: spokeCalculatorLanguage.value,
      ...(spokeCalculatorEngineeringMetadata.value?.model_version
        ? { softwareVersion: spokeCalculatorEngineeringMetadata.value.model_version }
        : {}),
      ...(spokeCalculatorEngineeringMetadata.value?.knowledge_as_of
        ? { dateModified: spokeCalculatorEngineeringMetadata.value.knowledge_as_of }
        : {}),
      ...(spokeCalculatorEngineeringMetadata.value?.formula_version
        ? {
            identifier: [{
              '@type': 'PropertyValue',
              propertyID: 'formulaVersion',
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
.spoke-page__intro {
  margin-bottom: 1.5rem;
}

.spoke-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: var(--tz-text-primary);
}

.spoke-page__description {
  max-width: 78ch;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.9375rem;
  line-height: 1.65;
}

.support-page__calculator-wrapper {
  margin-top: 0;
}

.spoke-page__head-step {
  margin-bottom: 1.5rem;
}

.spoke-result-step {
  margin-bottom: 1.5rem;
}

.spoke-result-step__intro {
  display: flex;
  align-items: flex-start;
  gap: 0.875rem;
  margin: 0 0 1.25rem;
}

.spoke-result-step__number {
  display: inline-flex;
  width: 1.7rem;
  height: 1.7rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: rgba(5, 150, 105, 0.12);
  color: #047857;
  font-size: 0.68rem;
  font-weight: 800;
}

.spoke-result-step__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 15px;
  font-weight: 800;
  line-height: 1.3;
}

.spoke-result-step__subtitle {
  margin: 3px 0 0;
  color: var(--tz-text-muted);
  font-size: 12px;
  line-height: 1.5;
}

 .spoke-page {
   margin: 0 auto;
   width: 100%;
   max-width: none;
 }

</style>
