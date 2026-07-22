<template>
  <div class="support-test-report">
    <h1 class="support-page__title support-page__title--sr-only">Test report</h1>
    
    <TestReportContent :sync-with-url="true" />

    <!-- FAQ Section - 放在所有 tab 内容之后 -->
    <section class="support-section">
      <PageFaq
        page-id="support-test-report"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import PageFaq from '~/components/PageFaq.vue'
import TestReportContent from '~/components/TestReportContent.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Test report',
})
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__title--sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.support-test-report {
  margin-top: -1rem;
}

.support-section {
  margin-top: 1rem;
  text-align: center;
}
</style>
