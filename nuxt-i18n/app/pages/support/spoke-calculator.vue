<template>
  <div>
    <h2 class="support-page__title">Spoke Calculator</h2>
    <p class="support-page__intro">
      Use the Tanzanite spoke length calculator directly inside Support to design wheelsets and
      estimate spoke lengths before you order.
    </p>

    <div class="support-page__calculator-wrapper">
      <SpokeCalculatorCore />
    </div>
  </div>
</template>

<script setup lang="ts">
import SpokeCalculatorCore from '~/components/SpokeCalculatorCore.vue'

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Spoke Calculator',
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
  margin: 0 0 0.75rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-page__intro--secondary {
  margin-top: 0.5rem;
}

.support-page__link {
  color: #38bdf8;
  text-decoration: underline;
}

.support-page__calculator-wrapper {
  margin-top: 1.5rem;
}
</style>
