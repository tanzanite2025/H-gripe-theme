<template>
  <div class="technology-wheelset-test-report-page">
    <SupportWheelsetTestReportSection heading-tag="h1" :open-wheelset-video="openWheelsetTestVideoDialog" />
    <TechnologyTestReportVideoDialog
      v-if="isWheelsetTestVideoDialogOpen"
      video-source-path="/testreport/wheelsettestreport/wheelsettestroport.webm"
      :dialog-label="t('supportTestReport.tabs.wheelset.label')"
      @close="isWheelsetTestVideoDialogOpen = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { definePageMeta, useHead, useI18n } from '#imports'
import SupportWheelsetTestReportSection from '~/components/SupportWheelsetTestReportSection.vue'
import TechnologyTestReportVideoDialog from '~/components/TechnologyTestReportVideoDialog.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'supportTestReport.tabs.wheelset.label',
  footerLabelFallback: 'Wheelset Test Report',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('supportTestReport')
await loadPageMessages(locale.value)
watch(locale, nextLocale => { void loadPageMessages(nextLocale) })

useHead(() => ({
  title: t('supportTestReport.tabs.wheelset.label'),
  meta: [{ name: 'description', content: t('supportTestReport.tabs.wheelset.description') }],
}))

const isWheelsetTestVideoDialogOpen = ref(false)
function openWheelsetTestVideoDialog(): void {
  isWheelsetTestVideoDialogOpen.value = true
}
</script>
