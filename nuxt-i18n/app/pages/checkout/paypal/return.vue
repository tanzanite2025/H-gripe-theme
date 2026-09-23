<template>
  <section class="paypal-return-page">
    <div class="paypal-return-panel">
      <Icon
        :name="statusIcon"
        class="paypal-return-icon"
:class="status === 'success' ? 'text-emerald-600' : status === 'error' ? 'text-rose-600' : 'text-blue-600'"
      />
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <p v-if="orderNumber" class="paypal-return-order">{{ orderNumber }}</p>
      <div class="paypal-return-actions">
        <NuxtLink class="paypal-return-button paypal-return-button--primary" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
        <button v-if="status === 'error'" class="paypal-return-button paypal-return-button--primary" type="button" @click="capture">
          {{ t('checkout.paypalReturn.actions.retry') }}
        </button>
        <button class="paypal-return-button" type="button" @click="openCart">
          {{ t('checkout.modal.actions.viewCart') }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'
import { useCart } from '~/composables/useCart'
import { usePayPalPayment } from '~/composables/usePayPalPayment'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const { reloadCartFromBackend, openCart } = useCart()
const { capturePayPalOrder } = usePayPalPayment()
const status = ref<'loading' | 'success' | 'error'>('loading')
const message = ref(t('checkout.paypalReturn.messages.capturing'))

const firstQueryValue = (value: unknown) => Array.isArray(value) ? String(value[0] || '') : String(value || '')
const orderNumber = computed(() => firstQueryValue(route.query.order_number).trim())
const paypalOrderId = computed(() =>
  firstQueryValue(route.query.token || route.query.paypal_order_id).trim(),
)
const title = computed(() => {
  if (status.value === 'success') return t('checkout.paypalReturn.title.success')
  if (status.value === 'error') return t('checkout.paypalReturn.title.error')
  return t('checkout.paypalReturn.title.loading')
})
const statusIcon = computed(() => {
  if (status.value === 'success') return 'lucide:circle-check'
  if (status.value === 'error') return 'lucide:circle-alert'
  return 'lucide:loader-circle'
})

const capture = async () => {
  if (status.value === 'loading') return
  try {
    if (!orderNumber.value || !paypalOrderId.value) {
      throw new Error(t('checkout.paypalReturn.messages.missingData'))
    }

    const result = await capturePayPalOrder({
      orderNumber: orderNumber.value,
      paypalOrderId: paypalOrderId.value,
      idempotencyKey: `paypal-capture-${paypalOrderId.value}`,
    })
    if (String(result.status || '').toUpperCase() !== 'COMPLETED') {
      throw new Error(t('checkout.paypalReturn.messages.incomplete'))
    }

    await reloadCartFromBackend()
    status.value = 'success'
    message.value = t('checkout.paypalReturn.messages.success')
  } catch (error) {
    status.value = 'error'
    message.value = error instanceof Error ? error.message : t('checkout.paypalReturn.messages.failed')
  }
}

onMounted(() => void capture())
</script>

<style scoped>
.paypal-return-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
   background: var(--tz-surface-page);
   color: var(--tz-text-primary);
}

.paypal-return-panel {
  width: min(100%, 520px);
   border: 1px solid var(--tz-border-subtle);
  border-radius: 8px;
   background: var(--tz-card-surface);
  padding: 2rem;
  text-align: center;
   box-shadow: 0 18px 44px rgba(15, 23, 42, 0.14);
}

.paypal-return-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
}

.paypal-return-panel h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.paypal-return-panel p {
  margin: 0.75rem 0 0;
   color: var(--tz-text-secondary);
}

.paypal-return-order {
  font-family: var(--tz-font-ui);
   color: var(--tz-text-primary) !important;
}

.paypal-return-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.paypal-return-button {
   border: 1px solid var(--tz-border-strong);
  border-radius: 8px;
  padding: 0.7rem 1rem;
   color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 700;
}

.paypal-return-button--primary {
  background: #fff;
  color: #020617;
}
</style>
