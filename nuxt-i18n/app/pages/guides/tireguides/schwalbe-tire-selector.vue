<template>
  <div class="schwalbe-page">
    <SchwalbeTireSelector />
    <SchwalbeTelemetryGuide />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import SchwalbeTireSelector from '~/components/tireguides/schwalbe/SchwalbeTireSelector.vue'
import SchwalbeTelemetryGuide from '~/components/tireguides/schwalbe/SchwalbeTelemetryGuide.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'products.nav.tireSizeCharts',
  footerLabelFallback: 'Tire Guides',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('guidesSchwalbeTireSelector')

await loadPageMessages(locale.value)
watch(locale, (nextLocale) => void loadPageMessages(nextLocale))

useHead(() => ({
  title: t('guidesSchwalbeTireSelector.seo.title'),
  meta: [{
    name: 'description',
    content: t('guidesSchwalbeTireSelector.seo.description'),
  }],
}))
</script>

<style scoped>
.schwalbe-page {
  display: grid;
  gap: 2rem;
  width: 100%;
  max-width: 96rem;
  margin: 0 auto;
}
</style>
