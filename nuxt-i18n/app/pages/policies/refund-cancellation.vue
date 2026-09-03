<template>
  <div class="company-page">
    <h1 class="sr-only">{{ pageTitle }}</h1>

    <div class="policies-content">
      <RefundCancellationPolicyContent />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, definePageMeta, useI18n } from '#imports'
import RefundCancellationPolicyContent from '~/components/RefundCancellationPolicyContent.vue'
import { getRefundCancellationPolicyContent } from '~/data/refundCancellationPolicy'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.refundCancellation',
  footerLabelFallback: 'Refund & Cancellation Policy',
})

const { locale } = useI18n()
const pageTitle = computed(() => getRefundCancellationPolicyContent(locale.value).title)

useHead(() => {
  return {
    title: pageTitle.value,
  }
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
