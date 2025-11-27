<template>
  <div>
    <h2 class="support-page__title">Product Feedback</h2>
    <p class="support-page__intro">
      Share feedback about Tanzanite products and your riding experience. This
      placeholder page can later be connected to a form or survey system.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Product Feedback',
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
