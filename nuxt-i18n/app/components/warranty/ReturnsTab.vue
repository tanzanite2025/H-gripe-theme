<template>
  <section id="returns" class="support-section">
    <RefundCancellationPolicyContent
      :contact-email="contactEmail"
      compact
    />
  </section>
</template>

<script setup lang="ts">
import RefundCancellationPolicyContent from '~/components/RefundCancellationPolicyContent.vue'

defineProps<{
  contactEmail?: string
}>()
</script>
