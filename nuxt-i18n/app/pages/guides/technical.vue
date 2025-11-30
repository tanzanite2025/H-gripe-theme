<template>
  <div>
    <h2 class="products-page__title">Technical</h2>
    <p class="products-page__intro">
      Technical notes, formulas, and reference docs for advanced wheelbuilding.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Technical',
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
