<template>
  <FreehubCassetteFitmentWholePage />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import FreehubCassetteFitmentWholePage from '~/components/FreehubCassetteFitmentWholePage.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'

definePageMeta({
  path: '/tools/freehub-cassette-fitment-engine',
  layout: 'products',
  footerLabelKey: 'toolsNavigation.freehubCassetteFitmentEngineLabel',
  footerLabelFallback: 'Freehub and cassette compatibility engine',
  breadcrumbLabelKey: 'toolsNavigation.freehubCassetteFitmentEngineLabel',
  breadcrumbLabelFallback: 'Freehub and cassette compatibility engine',
  feedbackThreadKey: 'guides-wheelset-buyers',
  feedbackTitleKey: 'guidesWheelsetBuyers.feedbackTitle',
  feedbackTitle: 'Share your feedback about the wheelset guide',
  pageTitleKey: 'guidesWheelsetBuyersChooseFreehub.seo.title',
  pageTitle: 'Bicycle freehub body compatibility guide',
})

const { locale, t } = useI18n()
const { canonicalUrl } = useStorefrontSeoLinks()
const { loadPageMessages: loadFreehubMessages } = usePageMessages('guidesWheelsetBuyersChooseFreehub')
const { loadPageMessages: loadWheelsetMessages } = usePageMessages('guidesWheelsetBuyers')
const { loadPageMessages: loadFreehubHelperMessages } = usePageMessages('wheelsetFreehubHelper')

await Promise.all([
  loadFreehubMessages(locale.value),
  loadWheelsetMessages(locale.value),
  loadFreehubHelperMessages(locale.value),
])

watch(locale, (nextLocale) => {
  void Promise.all([
    loadFreehubMessages(nextLocale),
    loadWheelsetMessages(nextLocale),
    loadFreehubHelperMessages(nextLocale),
  ])
})

useHead(() => ({
  title: t('guidesWheelsetBuyersChooseFreehub.seo.title'),
  meta: [
    { key: 'description', name: 'description', content: t('guidesWheelsetBuyersChooseFreehub.seo.description') },
    { key: 'og:type', property: 'og:type', content: 'article' },
    { key: 'og:title', property: 'og:title', content: t('guidesWheelsetBuyersChooseFreehub.seo.title') },
    { key: 'og:description', property: 'og:description', content: t('guidesWheelsetBuyersChooseFreehub.seo.description') },
    { key: 'og:url', property: 'og:url', content: canonicalUrl.value },
    { key: 'twitter:card', name: 'twitter:card', content: 'summary' },
    { key: 'twitter:title', name: 'twitter:title', content: t('guidesWheelsetBuyersChooseFreehub.seo.title') },
    { key: 'twitter:description', name: 'twitter:description', content: t('guidesWheelsetBuyersChooseFreehub.seo.description') },
  ],
}))
</script>
