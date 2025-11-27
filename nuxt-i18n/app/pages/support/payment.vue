<template>
  <div>
    <h2 class="support-page__title">Payment</h2>
    <p class="support-page__intro">
      Learn about available payment methods, processing times, and basic checkout
      tips. This is a placeholder page; detailed payment guides will be added
      here later.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Payment',
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
