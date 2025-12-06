<template>
  <div>
    <h2 class="support-page__title">Warranty</h2>
    <p class="support-page__intro">
      Placeholder for Tanzanite warranty terms, coverage, and claim process.
      Use this page to explain how long components are covered and what is
      included or excluded.
    </p>

    <!-- FAQ Section -->
    <section class="support-section">
      <PageFaq 
        page-id="support-warranty"
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
  title: 'Warranty',
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
  margin: 0;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}
</style>
