<template>
  <div class="technology-rim-test-report-page">
    <SupportRimTestReportSection heading-tag="h1" />
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import SupportRimTestReportSection from '~/components/SupportRimTestReportSection.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'supportTestReport.tabs.rim.label',
  footerLabelFallback: 'Rim Test',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('supportTestReport')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

useHead(() => ({
  title: t('supportTestReport.tabs.rim.label'),
  meta: [{ name: 'description', content: t('supportTestReport.tabs.rim.description') }],
}))
</script>
