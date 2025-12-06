<template>
  <div>
    <h2 class="support-page__title">Test report</h2>
    <p class="support-page__intro">
      Here you will be able to find test reports, quality checks, and performance results
      related to Tanzanite products.
    </p>

    <section class="support-section">
      <h3 class="support-section__title">Overview</h3>
      <p class="support-section__body">
        This page is a placeholder for future detailed test reports. We will add structured
        summaries, downloadable documents, and links to third-party test data here so that
        you can review how our products perform.
      </p>
    </section>

    <section class="support-section">
      <h3 class="support-section__title">Coming soon</h3>
      <p class="support-section__body">
        If you need a specific report right now, please contact our support team and mention
        which product and test you are interested in. We can share available documents
        manually while this page is being completed.
      </p>
    </section>

    <!-- FAQ Section -->
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
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.support-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-section {
  margin-top: 1.75rem;
}

.support-section__title {
  margin: 0 0 0.5rem;
  font-size: 1.05rem;
  font-weight: 600;
  color: #e5e7eb;
}

.support-section__body {
  margin: 0;
  font-size: 0.9rem;
  line-height: 1.6;
  color: rgba(148, 163, 184, 0.9);
}
</style>
