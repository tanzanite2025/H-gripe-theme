<template>
  <section class="checkout-success-page">
    <div class="checkout-success-panel">
      <Icon name="lucide:circle-check" class="checkout-success-icon" aria-hidden="true" />
      <p class="checkout-success-eyebrow">{{ t('checkout.paypalReturn.title.success') }}</p>
      <h1>{{ t('checkout.modal.messages.orderSuccess') }}</h1>
      <p>{{ t('checkout.paypalReturn.messages.success') }}</p>
      <p v-if="orderNumber" class="checkout-success-order">
        {{ orderNumber }}
      </p>
      <div class="checkout-success-actions">
        <NuxtLink class="checkout-success-button checkout-success-button--primary" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
        <button class="checkout-success-button" type="button" @click="openCart">
          {{ t('checkout.modal.actions.viewCart', 'View cart') }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'
import { useCart } from '~/composables/useCart'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const { openCart } = useCart()

const firstQueryValue = (value: unknown) =>
  Array.isArray(value) ? String(value[0] || '') : String(value || '')

const orderNumber = computed(() => firstQueryValue(route.query.order_number).trim())
</script>

<style scoped>
.checkout-success-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
}

.checkout-success-panel {
  width: min(100%, 520px);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 8px;
  background: var(--tz-card-surface);
  padding: 2rem;
  text-align: center;
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.14);
}

.checkout-success-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
  color: rgb(5 150 105);
}

.checkout-success-eyebrow {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.checkout-success-panel h1 {
  margin: 0.5rem 0 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.checkout-success-panel p {
  margin: 0.75rem 0 0;
  color: var(--tz-text-secondary);
}

.checkout-success-order {
  color: var(--tz-text-primary) !important;
  font-family: var(--tz-font-ui);
}

.checkout-success-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.checkout-success-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.5rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.5rem;
  padding: 0.65rem 1rem;
  color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  transition: background-color 0.2s ease, border-color 0.2s ease;
}

.checkout-success-button:hover {
  background: var(--tz-surface-subtle);
  border-color: var(--tz-border-strong);
}

.checkout-success-button--primary {
  background: var(--tz-text-primary);
  border-color: var(--tz-text-primary);
  color: var(--tz-surface-page);
}

.checkout-success-button--primary:hover {
  background: var(--tz-text-primary);
}
</style>
