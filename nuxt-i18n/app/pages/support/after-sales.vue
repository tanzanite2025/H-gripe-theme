<template>
  <div>
    <h2 class="support-page__title">After sales</h2>
    <p class="support-page__intro">
      Overview of after-sales support including returns, exchanges, and repair
      requests. This page is a placeholder and will later describe policies and
      step-by-step procedures.
    </p>

    <!-- FAQ Section -->
    <section class="support-section">
      <PageFaq 
        page-id="support-after-sales"
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
  title: 'After sales',
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
