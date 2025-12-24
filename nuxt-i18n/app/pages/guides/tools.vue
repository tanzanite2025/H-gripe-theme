<template>
  <div>
    <h2 class="products-page__title">About Tools</h2>
    <p class="products-page__intro">
      Overview of tools and utilities that support your wheelbuilding workflow.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'products',
})

useHead({
  title: 'About Tools',
})
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
  text-align: center;
}

.products-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
  text-align: center;
}
</style>
