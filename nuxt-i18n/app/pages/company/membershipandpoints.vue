<template>
  <div>
    <h2 class="company-page__title">Membership and Points</h2>
    <p class="company-page__intro">
      This page will describe membership tiers, benefits, and how the points system works.
    </p>

    <!-- FAQ Section -->
    <section class="company-section">
      <PageFaq 
        page-id="company-membership"
        theme="dark"
        :show-categories="true"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import PageFaq from '~/components/PageFaq.vue'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Membership and Points',
})
</script>

<style scoped>
.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}
</style>
