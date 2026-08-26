<template>
  <div class="company-page">
    <h1 class="sr-only">Refund & Return Policy</h1>

    <div class="policies-content">
      <RefundReturnPolicyContent />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useHead, definePageMeta } from '#imports'
import RefundReturnPolicyContent from '~/components/RefundReturnPolicyContent.vue'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.refundReturn',
  footerLabelFallback: 'Refund & Return',
})

useHead({
  title: 'Refund & Return Policy',
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

@media (max-width: 767px) {
  .company-page {
    max-width: 100%;
    padding-inline: 0;
  }
}
</style>
