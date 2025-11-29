<template>
  <div>
    <h2 class="products-page__title">Wheelset Buyers Guide</h2>
    <p class="products-page__intro">
      A high-level guide to choosing the right wheelset, to be expanded with detailed steps.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Wheelset Buyers Guide',
})
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}
</style>
