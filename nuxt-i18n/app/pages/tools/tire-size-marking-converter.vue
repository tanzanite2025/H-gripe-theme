<template>
  <TireSizeGuide />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import TireSizeGuide from '~/components/tireguides/TireSizeGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useTireGuidePageSearchAndSocialMetadata } from '~/composables/seo/useTireGuidePageSearchAndSocialMetadata'

definePageMeta({
  path: '/tools/tire-size-marking-converter',
  layout: 'products',
  footerLabelKey: 'toolsNavigation.tireSizeMarkingConverterLabel',
  footerLabelFallback: 'Tire size marking converter',
  breadcrumbLabelKey: 'toolsNavigation.tireSizeMarkingConverterLabel',
  breadcrumbLabelFallback: 'Tire size marking converter',
  feedbackThreadKey: 'guides-tireguides',
  feedbackTitleKey: 'guidesTireguides.feedbackTitle',
  feedbackTitle: 'Share your feedback about the tire size marking converter',
  feedbackSubtitleKey: 'feedback.defaultSubtitle',
  pageTitleKey: 'guidesTireSize.title',
  pageTitle: 'Tire size marking converter',
})

const { locale, t } = useI18n()
const { loadPageMessages: loadTireSizeMessages } = usePageMessages('guidesTireSize')
const { loadPageMessages: loadTireGuideMessages } = usePageMessages('guidesTireguides')

await Promise.all([
  loadTireSizeMessages(locale.value),
  loadTireGuideMessages(locale.value),
])

watch(locale, (nextLocale) => {
  void Promise.all([
    loadTireSizeMessages(nextLocale),
    loadTireGuideMessages(nextLocale),
  ])
})

useTireGuidePageSearchAndSocialMetadata({
  fallbackPageTitle: () => t('guidesTireSize.title'),
  activeGuideTabId: 'tire-size-markings',
})
</script>
