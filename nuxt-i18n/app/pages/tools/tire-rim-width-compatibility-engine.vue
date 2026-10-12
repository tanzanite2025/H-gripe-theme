<template>
  <HowToChooseGuide :seo-canonical-route-path="standaloneToolRoutePath" />
</template>

<script setup lang="ts">
import HowToChooseGuide from '~/components/tireguides/HowToChooseGuide.vue'

const standaloneToolRoutePath = '/tools/tire-rim-width-compatibility-engine'

definePageMeta({
  path: '/tools/tire-rim-width-compatibility-engine',
  layout: 'products',
  footerLabelKey: 'toolsNavigation.tireRimWidthCompatibilityEngineLabel',
  footerLabelFallback: 'Tire and rim width compatibility engine',
  breadcrumbLabelKey: 'toolsNavigation.tireRimWidthCompatibilityEngineLabel',
  breadcrumbLabelFallback: 'Tire and rim width compatibility engine',
  feedbackThreadKey: 'guides-tireguides',
  feedbackTitleKey: 'guidesTireguides.feedbackTitle',
  feedbackTitle: 'Share your feedback about the tire and rim width tool',
  feedbackSubtitleKey: 'feedback.defaultSubtitle',
  pageTitleKey: 'guidesTireguides.tabs.choose.label',
  pageTitle: 'Tire and rim width compatibility engine',
})
</script>
