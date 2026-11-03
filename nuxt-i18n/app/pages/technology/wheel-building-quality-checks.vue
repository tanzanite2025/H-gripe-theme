<template>
  <div class="technology-wheel-building-quality-checks-page">
    <h1 class="text-center text-xl font-semibold tz-text-primary mb-6">{{ t('supportTestReport.tabs.assembly.label') }}</h1>
    <SupportWheelsetAssemblySection />
    <section id="spoke-tension" class="wheel-building-spoke-tension-section">
      <TechnicalTensionSection />
    </section>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import SupportWheelsetAssemblySection from '~/components/SupportWheelsetAssemblySection.vue'
import TechnicalTensionSection from '~/components/TechnicalTensionSection.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  faqPageId: 'support-test-report',
  footerLabelKey: 'supportTestReport.tabs.assembly.label',
  footerLabelFallback: 'Wheel Building & Quality Checks',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('supportTestReport')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

useHead(() => ({
  title: t('supportTestReport.tabs.assembly.label'),
  meta: [{ name: 'description', content: t('supportTestReport.tabs.assembly.description') }],
}))
</script>

<style scoped>
.wheel-building-spoke-tension-section {
  margin-top: 2rem;
  scroll-margin-top: calc(var(--tz-site-header-spacer-height, 120px) + 1rem);
}
</style>
