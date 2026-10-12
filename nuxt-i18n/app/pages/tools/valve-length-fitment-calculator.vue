<template>
  <InnerTubeGuide />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useI18n } from '#imports'
import InnerTubeGuide from '~/components/tireguides/InnerTubeGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'
import { useTireGuidePageSearchAndSocialMetadata } from '~/composables/seo/useTireGuidePageSearchAndSocialMetadata'

definePageMeta({
  path: '/tools/valve-length-fitment-calculator',
  layout: 'products',
  footerLabelKey: 'toolsNavigation.valveLengthFitmentCalculatorLabel',
  footerLabelFallback: 'Inner tube valve-length fitment calculator',
  breadcrumbLabelKey: 'toolsNavigation.valveLengthFitmentCalculatorLabel',
  breadcrumbLabelFallback: 'Inner tube valve-length fitment calculator',
  pageTitleKey: 'guidesTireInnerTube.page.title',
  pageTitle: 'How to Choose a Bicycle Inner Tube',
})

const { locale, t } = useI18n()
const { loadPageMessages: loadInnerTubeMessages } = usePageMessages('guidesTireInnerTube')
const { loadPageMessages: loadTireGuideMessages } = usePageMessages('guidesTireguides')

await Promise.all([
  loadInnerTubeMessages(locale.value),
  loadTireGuideMessages(locale.value),
])

watch(locale, (nextLocale) => {
  void Promise.all([
    loadInnerTubeMessages(nextLocale),
    loadTireGuideMessages(nextLocale),
  ])
})

useTireGuidePageSearchAndSocialMetadata({
  fallbackPageTitle: () => t('guidesTireInnerTube.page.title'),
  activeGuideTabId: 'choose-inner-tube',
})
</script>
