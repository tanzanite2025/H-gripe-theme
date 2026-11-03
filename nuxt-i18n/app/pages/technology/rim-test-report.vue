<template>
  <div class="technology-rim-test-report-page">
    <SupportRimTestReportSection heading-tag="h1" :open-spoke-hole-video="openRimSpokeHoleTestVideoDialog" />
    <TechnologyTestReportVideoDialog
      v-if="isRimSpokeHoleTestVideoDialogOpen"
      video-source-path="/testreport/rimtestreport/Spoke-hole-strength-test.webm"
      :dialog-label="t('supportTestReport.tabs.rim.label')"
      @close="isRimSpokeHoleTestVideoDialogOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import SupportRimTestReportSection from '~/components/SupportRimTestReportSection.vue'
import TechnologyTestReportVideoDialog from '~/components/TechnologyTestReportVideoDialog.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'supportTestReport.tabs.rim.label',
  footerLabelFallback: 'Rim Test Report',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('supportTestReport')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

useHead(() => ({
  title: t('supportTestReport.tabs.rim.label'),
  meta: [{ name: 'description', content: t('supportTestReport.tabs.rim.description') }],
}))

const isRimSpokeHoleTestVideoDialogOpen = ref(false)
function openRimSpokeHoleTestVideoDialog(): void {
  isRimSpokeHoleTestVideoDialogOpen.value = true
}
</script>
