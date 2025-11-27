<template>
  <div>
    <h2 class="support-page__title">All FAQ'S</h2>
    <p class="support-page__intro">
      Browse common questions and quick answers about orders, products, and service.
      This page is a placeholder and will be expanded with structured FAQs later.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'support',
})

useHead({
  title: "All FAQ'S",
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
