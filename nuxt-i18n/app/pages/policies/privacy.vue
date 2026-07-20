<template>
  <div class="company-page">
    <h1 class="sr-only">Privacy Policy</h1>

    <div class="policies-content">
      <p class="text-sm text-slate-400 mb-6">
        This page describes how Tanzanite collects, uses, and protects your personal data when you browse our site or use our services.
      </p>
      <div class="mt-4">
        <PrivacyStatementContent />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import PrivacyStatementContent from '~/components/PrivacyStatementContent.vue'
import { useHead, definePageMeta } from '#imports'

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Privacy Policy',
})
</script>

<style scoped>
.company-page {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 0;
}

.policies-content {
  margin-top: 0.5rem;
}

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
