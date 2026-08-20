<template>
  <section class="paypal-cancel-page">
    <div class="paypal-cancel-panel">
      <Icon name="lucide:circle-x" class="paypal-cancel-icon" />
      <h1>{{ t('checkout.paypalCancel.title') }}</h1>
      <p>{{ t('checkout.paypalCancel.message') }}</p>
      <p v-if="orderNumber" class="paypal-cancel-order">{{ orderNumber }}</p>
      <div class="paypal-cancel-actions">
        <NuxtLink class="paypal-cancel-button" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const firstQueryValue = (value: unknown) => Array.isArray(value) ? String(value[0] || '') : String(value || '')
const orderNumber = computed(() => firstQueryValue(route.query.order_number).trim())
</script>

<style scoped>
.paypal-cancel-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
  background: #000;
  color: #fff;
}

.paypal-cancel-panel {
  width: min(100%, 520px);
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.86);
  padding: 2rem;
  text-align: center;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.42);
}

.paypal-cancel-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
  color: #fda4af;
}

.paypal-cancel-panel h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.paypal-cancel-panel p {
  margin: 0.75rem 0 0;
  color: rgba(255, 255, 255, 0.72);
}

.paypal-cancel-order {
  font-family: var(--tz-font-ui);
  color: #fff !important;
}

.paypal-cancel-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.paypal-cancel-button {
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 8px;
  padding: 0.7rem 1rem;
  color: #fff;
  font-size: 0.875rem;
  font-weight: 700;
}

.paypal-cancel-button--primary {
  background: #fff;
  color: #020617;
}
</style>
