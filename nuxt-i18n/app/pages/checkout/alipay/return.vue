<template>
  <section class="wallet-return-page">
    <div class="wallet-return-panel">
      <Icon
        :name="statusIcon"
        class="wallet-return-icon"
:class="status === 'success' ? 'text-emerald-600' : status === 'error' ? 'text-rose-600' : 'text-blue-600'"
      />
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <p v-if="orderNumber" class="wallet-return-order">{{ orderNumber }}</p>
      <div class="wallet-return-actions">
        <NuxtLink class="wallet-return-button wallet-return-button--primary" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping') }}
        </NuxtLink>
        <button class="wallet-return-button" type="button" @click="openCart">
          {{ t('checkout.modal.actions.viewCart') }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'
import { useAlipayPayment } from '~/composables/useAlipayPayment'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const auth = useAuth()
const { clearCart, openCart } = useCart()
const { confirmAlipayOrder } = useAlipayPayment()
const status = ref<'loading' | 'success' | 'error'>('loading')
const message = ref(t('checkout.alipayReturn.messages.confirming'))

const firstQueryValue = (value: unknown) => Array.isArray(value) ? String(value[0] || '') : String(value || '')
const orderNumber = computed(() =>
  firstQueryValue(route.query.order_number || route.query.out_trade_no).trim(),
)
const title = computed(() => {
  if (status.value === 'success') return t('checkout.alipayReturn.title.success')
  if (status.value === 'error') return t('checkout.alipayReturn.title.error')
  return t('checkout.alipayReturn.title.loading')
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
      throw new Error(t('checkout.alipayReturn.messages.loginRequired'))
    }
    if (!orderNumber.value) {
      throw new Error(t('checkout.alipayReturn.messages.missingData'))
    }

    const result = await confirmAlipayOrder(
      orderNumber.value,
      `alipay-confirm-${orderNumber.value}`,
    )
    if (!['TRADE_SUCCESS', 'TRADE_FINISHED'].includes(String(result.status || '').toUpperCase())) {
      throw new Error(t('checkout.alipayReturn.messages.incomplete'))
    }

    clearCart()
    status.value = 'success'
    message.value = t('checkout.alipayReturn.messages.success')
  } catch (error) {
    status.value = 'error'
    message.value = error instanceof Error ? error.message : t('checkout.alipayReturn.messages.failed')
  }
})
</script>

<style scoped>
.wallet-return-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
   background: var(--tz-surface-page);
   color: var(--tz-text-primary);
}

.wallet-return-panel {
  width: min(100%, 520px);
   border: 1px solid var(--tz-border-subtle);
  border-radius: 8px;
   background: var(--tz-card-surface);
  padding: 2rem;
  text-align: center;
   box-shadow: 0 18px 44px rgba(15, 23, 42, 0.14);
}

.wallet-return-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
}

.wallet-return-panel h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.wallet-return-panel p {
  margin: 0.75rem 0 0;
   color: var(--tz-text-secondary);
}

.wallet-return-order {
  font-family: var(--tz-font-ui);
   color: var(--tz-text-primary) !important;
}

.wallet-return-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.wallet-return-button {
   border: 1px solid var(--tz-border-strong);
  border-radius: 8px;
  padding: 0.7rem 1rem;
   color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 700;
}

.wallet-return-button--primary {
  background: #fff;
  color: #020617;
}
</style>
