<template>
  <TirePressureGuide />
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import TirePressureGuide from '~/components/tireguides/TirePressureGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  path: '/guides/tireguides/tire-pressure',
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
  feedbackThreadKey: 'guides-tire-pressure',
  feedbackTitleKey: 'guidesTirePressure.feedbackTitle',
  feedbackTitle: 'Share your feedback about the tire-pressure calculator',
  pageTitleKey: 'guidesTirePressure.title',
  pageTitle: 'Tire pressure and cornering-force demonstration',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesTirePressure')

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))

useHead(() => ({
  title: t('guidesTirePressure.title'),
  meta: [{ name: 'description', content: t('guidesTirePressure.description') }],
}))
</script>
