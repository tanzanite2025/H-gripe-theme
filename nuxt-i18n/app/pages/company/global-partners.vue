<template>
  <div class="w-full pb-6">
    <div class="w-full max-w-none">
      <section class="text-center">
        <h1 class="mb-2 text-xl md:text-2xl font-black tracking-tighter uppercase italic leading-tight tz-text-primary">
          {{ t('companyGlobalPartners.title') }}
        </h1>
        <p class="mx-auto max-w-3xl text-sm leading-relaxed tz-text-secondary">
          {{ t('companyGlobalPartners.hero.tagline') }}
        </p>
        <section
          id="global-partners-collaboration-channels"
          aria-labelledby="global-partners-collaboration-channels-heading"
          class="mb-6 mt-6"
        >
          <h2
            id="global-partners-collaboration-channels-heading"
            class="sr-only"
          >
            {{ t('companyGlobalPartners.partnershipChannels.title') }}
          </h2>
          <div class="grid gap-4 lg:grid-cols-3">
            <article
              v-for="channel in globalPartnerCollaborationChannelDefinitions"
              :key="channel.id"
              :data-partnership-channel="channel.id"
              class="premium-card flex min-w-0 flex-col rounded-2xl p-5 text-left sm:p-6"
            >
              <span
                class="inline-flex h-10 w-10 items-center justify-center rounded-xl tz-surface-muted tz-text-accent"
                aria-hidden="true"
              >
                <Icon :name="channel.icon" class="h-5 w-5" />
              </span>
              <p class="mt-4 text-xs font-medium leading-relaxed tz-text-accent">
                {{ t(channel.audienceKey) }}
              </p>
              <h3 class="mt-2 text-base font-semibold leading-snug tz-text-primary">
                {{ t(channel.titleKey) }}
              </h3>
              <p class="mt-3 text-sm leading-relaxed tz-text-secondary">
                {{ t(channel.descriptionKey) }}
              </p>
              <ul class="my-4 space-y-2 text-xs leading-relaxed tz-text-secondary">
                <li v-for="featureKey in channel.featureKeys" :key="featureKey" class="flex items-start gap-2">
                  <span class="mt-1.5 h-1 w-1 shrink-0 rounded-full bg-current tz-text-accent" aria-hidden="true"></span>
                  <span>{{ t(featureKey) }}</span>
                </li>
              </ul>
              <div class="mt-auto border-t tz-border-subtle pt-4">
                <NuxtLink
                  :to="localePath({ path: '/company/contact', query: { partnership: channel.id } })"
                  class="premium-button premium-button--active min-h-12 w-full justify-center gap-2 text-center leading-relaxed"
                >
                  <span>{{ t(channel.ctaKey) }}</span>
                  <Icon name="lucide:arrow-right" class="h-4 w-4 shrink-0" aria-hidden="true" />
                </NuxtLink>
              </div>
            </article>
          </div>
          <p class="mt-3 text-xs leading-relaxed tz-text-muted">
            {{ t('companyGlobalPartners.partnershipChannels.conditionsNote') }}
          </p>
        </section>
        <WorldMapXiamen />

        <!-- Data highlights cards -->
        <section class="mt-6 grid gap-4 sm:grid-cols-3">
          <!-- Countries and regions served at the recorded statistics date -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.countries.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              {{ globalPartnerStatisticsSnapshot.countriesAndRegionsMinimumCount }}<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-2 text-xs tz-text-muted">
              {{ t('companyGlobalPartners.stats.asOf') }}
              <time :datetime="globalPartnerStatisticsSnapshot.asOf">
                {{ formattedGlobalPartnerStatisticsDate }}
              </time>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.countries.description') }}
            </p>
          </div>

          <!-- Brand partner count at the recorded statistics date -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.brandPartners.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              {{ globalPartnerStatisticsSnapshot.brandPartnersMinimumCount }}<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-2 text-xs tz-text-muted">
              {{ t('companyGlobalPartners.stats.asOf') }}
              <time :datetime="globalPartnerStatisticsSnapshot.asOf">
                {{ formattedGlobalPartnerStatisticsDate }}
              </time>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.brandPartners.description') }}
            </p>
          </div>

          <!-- Products delivered at the recorded statistics date -->
          <div class="premium-card rounded-2xl p-4">
            <p class="text-xs font-medium uppercase tracking-wide tz-text-muted">
              {{ t('companyGlobalPartners.stats.productsDelivered.label') }}
            </p>
            <p class="mt-2 text-3xl font-semibold tz-text-primary">
              {{ formattedProductsDeliveredMinimumCount }}<span class="align-top text-lg">+</span>
            </p>
            <p class="mt-2 text-xs tz-text-muted">
              {{ t('companyGlobalPartners.stats.asOf') }}
              <time :datetime="globalPartnerStatisticsSnapshot.asOf">
                {{ formattedGlobalPartnerStatisticsDate }}
              </time>
            </p>
            <p class="mt-1 text-xs tz-text-secondary">
              {{ t('companyGlobalPartners.stats.productsDelivered.description') }}
            </p>
          </div>
        </section>

        <section
          id="global-partners-certifications-and-standards"
          aria-labelledby="global-partners-certifications-and-standards-heading"
          class="premium-card mt-4 rounded-2xl p-4 text-left sm:p-5"
        >
          <div class="flex items-start gap-3">
            <span
              class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full tz-surface-muted tz-text-accent"
              aria-hidden="true"
            >
              <Icon name="lucide:award" class="h-5 w-5" />
            </span>
            <div class="min-w-0">
              <h2
                id="global-partners-certifications-and-standards-heading"
                class="text-sm font-semibold tz-text-primary"
              >
                {{ t('companyGlobalPartners.certificationsAndStandards.title') }}
              </h2>
              <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
                {{ t('companyGlobalPartners.certificationsAndStandards.description') }}
              </p>
            </div>
          </div>

          <ul class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            <li
              v-for="badge in globalPartnerCertificationAndStandardsBadges"
              :key="badge.titleKey"
              class="flex min-w-0 items-start gap-3 rounded-xl border tz-border-subtle p-3"
            >
              <span
                class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-lg tz-surface-muted tz-text-accent"
                aria-hidden="true"
              >
                <Icon :name="badge.icon" class="h-4 w-4" />
              </span>
              <div class="min-w-0">
                <p class="text-sm font-semibold leading-snug tz-text-accent">
                  {{ badge.standardNames }}
                </p>
                <h3 class="mt-1 text-xs font-semibold tz-text-primary">
                  {{ t(badge.titleKey) }}
                </h3>
                <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
                  {{ t(badge.descriptionKey) }}
                </p>
              </div>
            </li>
          </ul>

          <div class="mt-4 flex flex-col gap-3 border-t tz-border-subtle pt-3 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-xs leading-relaxed tz-text-muted">
              {{ t('companyGlobalPartners.certificationsAndStandards.scopeNote') }}
            </p>
            <NuxtLink
              :to="localePath('/company/contact')"
              class="inline-flex shrink-0 items-center gap-1 text-xs font-semibold tz-text-accent hover:underline"
            >
              {{ t('companyGlobalPartners.certificationsAndStandards.contactCta') }}
              <Icon name="lucide:arrow-right" class="h-3.5 w-3.5" aria-hidden="true" />
            </NuxtLink>
          </div>
        </section>

        <!-- Confidentiality notice for partner brands -->
        <div
          class="premium-card mt-4 rounded-2xl p-3 text-center text-sm tz-text-accent"
        >
          <p class="m-0">
            {{ t('companyGlobalPartners.confidentiality') }}
          </p>
        </div>

        <!-- Value proposition: Why partner with us -->
        <section class="mt-8">
          <h2 class="mb-3 text-sm font-semibold tz-text-primary text-center">
            {{ t('companyGlobalPartners.whyPartner.title') }}
          </h2>
          <div class="grid gap-4 sm:grid-cols-3">
            <!-- Project communication -->
            <div
              class="premium-card rounded-2xl p-4 text-center"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                01
              </div>
              <h3 class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.trust.title') }}
              </h3>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.trust.description') }}
              </p>
            </div>

            <!-- Product development -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                02
              </div>
              <h3 class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.technology.title') }}
              </h3>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.technology.description') }}
              </p>
            </div>

            <!-- Delivery coordination -->
            <div
              class="premium-card rounded-2xl p-4"
            >
              <div
                class="mb-3 inline-flex h-8 w-8 items-center justify-center rounded-full tz-surface-muted text-xs font-semibold tz-text-accent"
              >
                03
              </div>
              <h3 class="text-sm font-semibold tz-text-primary">
                {{ t('companyGlobalPartners.whyPartner.support.title') }}
              </h3>
              <p class="mt-1 text-xs tz-text-secondary">
                {{ t('companyGlobalPartners.whyPartner.support.description') }}
              </p>
            </div>
          </div>
        </section>
      </section>
    </div>

    <section
      id="global-partners-b2b-partnership-faq"
      aria-labelledby="global-partners-b2b-partnership-faq-heading"
      class="premium-card mt-8 w-full rounded-2xl p-4 text-left sm:p-5"
    >
      <div class="flex items-start gap-3">
        <span
          class="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-full tz-surface-muted tz-text-accent"
          aria-hidden="true"
        >
          <Icon name="lucide:briefcase" class="h-5 w-5" />
        </span>
        <div class="min-w-0">
          <h2
            id="global-partners-b2b-partnership-faq-heading"
            class="text-base font-semibold tz-text-primary"
          >
            {{ t('companyGlobalPartners.partnershipFaq.title') }}
          </h2>
          <p class="mt-1 text-xs leading-relaxed tz-text-secondary">
            {{ t('companyGlobalPartners.partnershipFaq.description') }}
          </p>
        </div>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-2">
        <article
          v-for="question in globalPartnerFrequentlyAskedQuestionTranslationKeys"
          :key="question.id"
          class="min-w-0 rounded-xl border tz-border-subtle p-4"
        >
          <h3 class="text-sm font-semibold leading-relaxed tz-text-primary">
            {{ t(question.questionKey) }}
          </h3>
          <p class="mt-3 inline-flex rounded-full px-2 py-1 text-xs font-medium tz-surface-muted tz-text-muted">
            {{ t('companyGlobalPartners.partnershipFaq.pendingUpdate') }}
          </p>
          <p class="mt-2 text-xs leading-relaxed tz-text-secondary">
            {{ t(question.answerKey) }}
          </p>
        </article>
      </div>

      <div class="mt-4 flex flex-col gap-3 border-t tz-border-subtle pt-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-xs leading-relaxed tz-text-secondary">
          {{ t('companyGlobalPartners.partnershipFaq.contactDescription') }}
        </p>
        <NuxtLink
          :to="localePath('/company/contact')"
          class="inline-flex shrink-0 items-center gap-1 text-xs font-semibold tz-text-accent hover:underline"
        >
          {{ t('companyGlobalPartners.partnershipFaq.contactCta') }}
          <Icon name="lucide:arrow-right" class="h-3.5 w-3.5" aria-hidden="true" />
        </NuxtLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'
import WorldMapXiamen from '~/components/WorldMapXiamen.vue'
import { globalPartnerCollaborationChannelDefinitions } from '~/data/company/globalPartnerCollaborationChannelDefinitions'
import { usePageMessages } from '~/composables/usePageMessages'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { getStorefrontLocaleLanguageTag } from '~/utils/storefrontLocales'

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages, resolvedLocale } = usePageMessages('companyGlobalPartners')
const { siteTitle } = useSiteTitle()
const { canonicalUrl } = useStorefrontSeoLinks()

const globalPartnerFrequentlyAskedQuestionTranslationKeys = [
  {
    id: 'minimum-order-quantity',
    questionKey: 'companyGlobalPartners.partnershipFaq.minimumOrderQuantity.question',
    answerKey: 'companyGlobalPartners.partnershipFaq.minimumOrderQuantity.answer',
  },
  {
    id: 'production-and-delivery-lead-time',
    questionKey: 'companyGlobalPartners.partnershipFaq.leadTime.question',
    answerKey: 'companyGlobalPartners.partnershipFaq.leadTime.answer',
  },
  {
    id: 'sample-request-costs-and-shipping-process',
    questionKey: 'companyGlobalPartners.partnershipFaq.samples.question',
    answerKey: 'companyGlobalPartners.partnershipFaq.samples.answer',
  },
  {
    id: 'customs-import-duties-and-taxes',
    questionKey: 'companyGlobalPartners.partnershipFaq.customsAndDuties.question',
    answerKey: 'companyGlobalPartners.partnershipFaq.customsAndDuties.answer',
  },
] as const

const globalPartnerCertificationAndStandardsBadges = [
  {
    standardNames: 'UCI',
    titleKey: 'companyGlobalPartners.certificationsAndStandards.uci.title',
    descriptionKey: 'companyGlobalPartners.certificationsAndStandards.uci.description',
    icon: 'lucide:award',
  },
  {
    standardNames: 'ISO 4210 / ISO 5775',
    titleKey: 'companyGlobalPartners.certificationsAndStandards.iso.title',
    descriptionKey: 'companyGlobalPartners.certificationsAndStandards.iso.description',
    icon: 'lucide:ruler',
  },
  {
    standardNames: 'SGS',
    titleKey: 'companyGlobalPartners.certificationsAndStandards.sgs.title',
    descriptionKey: 'companyGlobalPartners.certificationsAndStandards.sgs.description',
    icon: 'lucide:clipboard-list',
  },
  {
    standardNames: 'TÜV / CE',
    titleKey: 'companyGlobalPartners.certificationsAndStandards.exportCompliance.title',
    descriptionKey: 'companyGlobalPartners.certificationsAndStandards.exportCompliance.description',
    icon: 'lucide:globe',
  },
] as const

// Update these counts and their date together when new statistics are confirmed.
// Keep this date fixed so older counts are never presented as today's statistics.
const globalPartnerStatisticsSnapshot = {
  countriesAndRegionsMinimumCount: 5,
  brandPartnersMinimumCount: 10,
  productsDeliveredMinimumCount: 60_000,
  asOf: '2026-10-12',
} as const

const formattedGlobalPartnerStatisticsDate = computed(() => (
  new Intl.DateTimeFormat(getStorefrontLocaleLanguageTag(resolvedLocale.value, 'en'), {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    timeZone: 'UTC',
  }).format(new Date(`${globalPartnerStatisticsSnapshot.asOf}T00:00:00Z`))
))

const formattedProductsDeliveredMinimumCount = computed(() => (
  new Intl.NumberFormat(getStorefrontLocaleLanguageTag(resolvedLocale.value, 'en'))
    .format(globalPartnerStatisticsSnapshot.productsDeliveredMinimumCount)
))

const globalPartnerOrganizationName = computed(() => siteTitle.value.trim())
const globalPartnerOrganizationWebsiteUrl = computed(() => new URL('/', canonicalUrl.value).href)
const globalPartnerOrganizationEntityId = computed(() => `${globalPartnerOrganizationWebsiteUrl.value}#organization`)
const globalPartnerSearchTitle = computed(() => (
  `${t('companyGlobalPartners.seo.title')} | ${t('companyGlobalPartners.seo.factoryTitle', {
    brandName: globalPartnerOrganizationName.value,
  }).trim()}`
))
const globalPartnerSearchDescription = computed(() => t('companyGlobalPartners.seo.description', {
  brandPartnersMinimumCount: globalPartnerStatisticsSnapshot.brandPartnersMinimumCount,
  countriesAndRegionsMinimumCount: globalPartnerStatisticsSnapshot.countriesAndRegionsMinimumCount,
}))

const globalPartnerOrganizationAndAboutPageStructuredData = computed(() => ({
  '@context': 'https://schema.org',
  '@graph': [
    ...(globalPartnerOrganizationName.value ? [{
      '@type': 'Organization',
      '@id': globalPartnerOrganizationEntityId.value,
      name: globalPartnerOrganizationName.value,
      url: globalPartnerOrganizationWebsiteUrl.value,
      address: {
        '@type': 'PostalAddress',
        addressLocality: 'Xiamen',
        addressRegion: 'Fujian',
        addressCountry: 'CN',
      },
      areaServed: ['North America', 'Europe', 'Asia-Pacific'],
      knowsAbout: ['Carbon Bicycle Rims', 'Wheelbuilding', 'OEM/ODM Manufacturing', 'Silent Bicycle Hubs'],
    }] : []),
    {
      '@type': 'AboutPage',
      '@id': `${canonicalUrl.value}#about-page`,
      url: canonicalUrl.value,
      name: t('companyGlobalPartners.title'),
      description: globalPartnerSearchDescription.value,
      inLanguage: getStorefrontLocaleLanguageTag(resolvedLocale.value, 'en'),
      ...(globalPartnerOrganizationName.value ? {
        about: { '@id': globalPartnerOrganizationEntityId.value },
        mainEntity: { '@id': globalPartnerOrganizationEntityId.value },
      } : {}),
    },
  ],
}))

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

definePageMeta({
  layout: 'products',
  footerLabelKey: 'company.nav.globalPartners',
  footerLabelFallback: 'Global Partners',
})

useHead(() => ({
  title: globalPartnerSearchTitle.value,
  titleTemplate: null,
  meta: [
    { name: 'description', content: globalPartnerSearchDescription.value },
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: globalPartnerSearchTitle.value },
    { property: 'og:description', content: globalPartnerSearchDescription.value },
    { property: 'og:url', content: canonicalUrl.value },
    { name: 'twitter:card', content: 'summary' },
    { name: 'twitter:title', content: globalPartnerSearchTitle.value },
    { name: 'twitter:description', content: globalPartnerSearchDescription.value },
  ],
  script: [{
    key: 'global-partners-organization-and-about-page-structured-data',
    ...createSeoJsonLdScript(globalPartnerOrganizationAndAboutPageStructuredData.value),
  }],
}))
</script>

