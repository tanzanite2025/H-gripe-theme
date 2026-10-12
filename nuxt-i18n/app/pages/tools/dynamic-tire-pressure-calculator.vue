<template>
  <TirePressureGuide />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TirePressureGuide from '~/components/tireguides/TirePressureGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useTireGuidePageSearchAndSocialMetadata } from '~/composables/seo/useTireGuidePageSearchAndSocialMetadata'

definePageMeta({
  path: '/tools/dynamic-tire-pressure-calculator',
  layout: 'products',
  footerLabelKey: 'toolsNavigation.dynamicTirePressureCalculatorLabel',
  footerLabelFallback: 'Dynamic tire pressure calculator',
  breadcrumbLabelKey: 'toolsNavigation.dynamicTirePressureCalculatorLabel',
  breadcrumbLabelFallback: 'Dynamic tire pressure calculator',
  feedbackThreadKey: 'guides-tire-pressure',
  feedbackTitleKey: 'guidesTirePressure.feedbackTitle',
  feedbackTitle: 'Share your feedback about the tire-pressure calculator',
  pageTitleKey: 'guidesTirePressure.title',
  pageTitle: 'Tire pressure and cornering-force demonstration',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTirePressure')
const { loadPageMessages: loadTireGuideSearchAndSocialMessages } = usePageMessages('guidesTireguides')

await Promise.all([
  loadPageMessages(locale.value),
  loadTireGuideSearchAndSocialMessages(locale.value),
])
watch(locale, (nextLocale) => {
  void Promise.all([
    loadPageMessages(nextLocale),
    loadTireGuideSearchAndSocialMessages(nextLocale),
  ])
})

useTireGuidePageSearchAndSocialMetadata({
  fallbackPageTitle: () => t('guidesTirePressure.title'),
  activeGuideTabId: 'tire-pressure',
})
</script>
