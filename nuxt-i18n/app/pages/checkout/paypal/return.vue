<template>
  <section class="paypal-return-page">
    <div class="paypal-return-panel">
      <Icon
        :name="statusIcon"
        class="paypal-return-icon"
        :class="status === 'success' ? 'text-emerald-300' : status === 'error' ? 'text-rose-300' : 'text-sky-300'"
      />
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <p v-if="orderNumber" class="paypal-return-order">{{ orderNumber }}</p>
      <div class="paypal-return-actions">
        <NuxtLink class="paypal-return-button paypal-return-button--primary" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
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
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePayPalPayment } from '~/composables/usePayPalPayment'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const auth = useAuth()
const { clearCart, openCart } = useCart()
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

onMounted(async () => {
  try {
    const user = await auth.ensureSession()
    if (!user) {
      throw new Error(t('checkout.paypalReturn.messages.loginRequired'))
    }
    if (!orderNumber.value || !paypalOrderId.value) {
      throw new Error(t('checkout.paypalReturn.messages.missingData'))
    }

    const result = await capturePayPalOrder({
      orderNumber: orderNumber.value,
      paypalOrderId: paypalOrderId.value,
    })
    if (String(result.status || '').toUpperCase() !== 'COMPLETED') {
      throw new Error(t('checkout.paypalReturn.messages.incomplete'))
    }

    clearCart()
    status.value = 'success'
    message.value = t('checkout.paypalReturn.messages.success')
  } catch (error) {
    status.value = 'error'
    message.value = error instanceof Error ? error.message : t('checkout.paypalReturn.messages.failed')
  }
})
</script>

<style scoped>
.paypal-return-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
  background: #000;
  color: #fff;
}

.paypal-return-panel {
  width: min(100%, 520px);
  border: 1px solid rgba(255, 255, 255, 0.14);
  border-radius: 8px;
  background: rgba(15, 23, 42, 0.86);
  padding: 2rem;
  text-align: center;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.42);
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
  color: rgba(255, 255, 255, 0.72);
}

.paypal-return-order {
  font-family: 'StorefrontSystem';
  color: #fff !important;
}

.paypal-return-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.paypal-return-button {
  border: 1px solid rgba(255, 255, 255, 0.22);
  border-radius: 8px;
  padding: 0.7rem 1rem;
  color: #fff;
  font-size: 0.875rem;
  font-weight: 700;
}

.paypal-return-button--primary {
  background: #fff;
  color: #020617;
}
</style>
