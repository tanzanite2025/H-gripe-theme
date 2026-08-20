<template>
  <section id="returns" class="support-section">
    <RefundReturnPolicyContent
      :contact-email="contactEmail"
      compact
    />
  </section>
</template>

<script setup lang="ts">
import RefundReturnPolicyContent from '~/components/RefundReturnPolicyContent.vue'

defineProps<{
  contactEmail?: string
}>()
</script>
