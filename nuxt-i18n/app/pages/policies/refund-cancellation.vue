<template>
  <div class="company-page">
    <h1 class="sr-only">{{ t('refundCancellation.title') }}</h1>

    <div class="policies-content">
      <RefundCancellationPolicyContent />
    </div>
  </div>
</template>

<script setup lang="ts">
import { watch } from 'vue'
import { useHead, definePageMeta, useI18n } from '#imports'
import RefundCancellationPolicyContent from '~/components/RefundCancellationPolicyContent.vue'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'products',
  footerLabelKey: 'policyTabs.refundCancellation',
  footerLabelFallback: 'Refund & Cancellation Policy',
})

const { locale, t } = useI18n()
const { loadPageMessages } = usePageMessages('refundCancellation')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('refundCancellation.title'),
}))
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
