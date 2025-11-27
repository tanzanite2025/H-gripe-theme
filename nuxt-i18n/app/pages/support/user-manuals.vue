<template>
  <div>
    <h2 class="support-page__title">User Manuals</h2>
    <p class="support-page__intro">
      Central place for downloading or viewing user manuals for Tanzanite
      products. This page is a placeholder and will later link to PDF guides and
      setup instructions.
    </p>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'support',
})

useHead({
  title: 'User Manuals',
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
