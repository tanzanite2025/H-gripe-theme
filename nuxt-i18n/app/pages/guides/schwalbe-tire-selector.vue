<template>
  <div class="schwalbe-page">
    <SchwalbeTireSelector />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, useI18n } from '#imports'
import SchwalbeTireSelector from '~/components/tireguides/schwalbe/SchwalbeTireSelector.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  // Keep this page at the Schwalbe selector URL while keeping it out of the
  // `tireguides.vue` component hierarchy. The parent page renders the default
  // Tubeless guide itself and has no `<NuxtPage />` outlet.
  path: '/guides/tireguides/schwalbe-tire-selector',
  breadcrumbLabelKey: 'guidesSchwalbeTireSelector.title',
  breadcrumbLabelFallback: 'Schwalbe tire selector',
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
