<template>
  <section class="stripe-return-page">
    <div class="stripe-return-panel">
      <Icon
        :name="statusIcon"
        class="stripe-return-icon"
        :class="status === 'success' ? 'text-emerald-600' : status === 'error' ? 'text-rose-600' : 'text-blue-600'"
        aria-hidden="true"
      />
      <h1>{{ title }}</h1>
      <p>{{ message }}</p>
      <p v-if="orderNumber" class="stripe-return-order">{{ orderNumber }}</p>
      <div v-if="status === 'error'" class="stripe-return-actions">
        <NuxtLink class="stripe-return-button stripe-return-button--primary" :to="localePath('/')">
          {{ t('checkout.paypalReturn.actions.continueShopping', 'Continue shopping') }}
        </NuxtLink>
        <button class="stripe-return-button" type="button" @click="openCart">
          {{ t('checkout.modal.actions.viewCart', 'View cart') }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { navigateTo, useI18n, useLocalePath, useRoute } from '#imports'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useStripePayment } from '~/composables/useStripePayment'
import {
  clearStripeReturnSession,
  readStripeReturnSession,
} from '~/utils/stripeReturn'

type ReturnStatus = 'loading' | 'success' | 'error'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()
const auth = useAuth()
const { clearCart, reloadCartFromBackend, openCart } = useCart()
const { loadPublishableKey, retrievePaymentIntent } = useStripePayment()

const status = ref<ReturnStatus>('loading')
const message = ref(t(
  'checkout.stripeReturn.messages.confirming',
  'Please wait while we confirm your Stripe payment.',
))

const firstQueryValue = (value: unknown) =>
  Array.isArray(value) ? String(value[0] || '') : String(value || '')

const orderNumber = computed(() => firstQueryValue(route.query.order_number).trim())
const paymentIntentId = computed(() => firstQueryValue(route.query.payment_intent).trim())
const redirectStatus = computed(() => firstQueryValue(route.query.redirect_status).trim().toLowerCase())
const queryClientSecret = computed(() => firstQueryValue(
  route.query.payment_intent_client_secret || route.query.client_secret,
).trim())

const title = computed(() => {
  if (status.value === 'success') {
    return t('checkout.stripeReturn.title.success', 'Payment confirmed')
  }
  if (status.value === 'error') {
    return t('checkout.stripeReturn.title.error', 'Payment could not be confirmed')
  }
  return t('checkout.stripeReturn.title.loading', 'Confirming Stripe payment')
})

const statusIcon = computed(() => {
  if (status.value === 'success') return 'lucide:circle-check'
  if (status.value === 'error') return 'lucide:circle-alert'
  return 'lucide:loader-circle'
})

onMounted(async () => {
  try {
    const session = await auth.ensureSession()
    if (!session) {
      throw new Error(t(
        'checkout.stripeReturn.messages.loginRequired',
        'Please sign in with the account that placed this order.',
      ))
    }
    if (!orderNumber.value) {
      throw new Error(t(
        'checkout.stripeReturn.messages.missingData',
        'Stripe returned without the required payment data.',
      ))
    }

    const storedSession = readStripeReturnSession(orderNumber.value)
    const clientSecret = queryClientSecret.value || storedSession?.clientSecret || ''
    if (!clientSecret) {
      throw new Error(t(
        'checkout.stripeReturn.messages.missingData',
        'Stripe returned without the required payment data.',
      ))
    }
    if (redirectStatus.value === 'failed') {
      throw new Error(t(
        'checkout.stripeReturn.messages.incomplete',
        'Stripe payment was not completed.',
      ))
    }

    const publishableKey = storedSession?.publishableKey || await loadPublishableKey()
    const result = await retrievePaymentIntent(clientSecret, publishableKey)
    if (paymentIntentId.value && result.paymentIntentId !== paymentIntentId.value) {
      throw new Error(t(
        'checkout.stripeReturn.messages.mismatch',
        'The returned payment does not match this order.',
      ))
    }
    if (!['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
      throw new Error(t(
        'checkout.stripeReturn.messages.incomplete',
        'Stripe payment is not complete yet.',
      ))
    }

    await clearCart()
    await reloadCartFromBackend()
    clearStripeReturnSession(orderNumber.value)
    status.value = 'success'
    message.value = t(
      'checkout.stripeReturn.messages.success',
      'Your payment has been confirmed. We are preparing your order.',
    )
    await navigateTo({
      path: localePath('/checkout/success'),
      query: { order_number: orderNumber.value },
    }, { replace: true })
  } catch (error) {
    status.value = 'error'
    message.value = error instanceof Error
      ? error.message
      : t('checkout.stripeReturn.messages.failed', 'Unable to confirm this Stripe payment.')
  }
})
</script>

<style scoped>
.stripe-return-page {
  min-height: 72vh;
  display: grid;
  place-items: center;
  padding: 8rem 1.25rem 4rem;
  background: var(--tz-surface-page);
  color: var(--tz-text-primary);
}

.stripe-return-panel {
  width: min(100%, 520px);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 8px;
  background: var(--tz-card-surface);
  padding: 2rem;
  text-align: center;
  box-shadow: 0 18px 44px rgba(15, 23, 42, 0.14);
}

.stripe-return-icon {
  width: 3rem;
  height: 3rem;
  margin: 0 auto 1rem;
}

.stripe-return-panel h1 {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 700;
}

.stripe-return-panel p {
  margin: 0.75rem 0 0;
  color: var(--tz-text-secondary);
}

.stripe-return-order {
  color: var(--tz-text-primary) !important;
  font-family: var(--tz-font-ui);
}

.stripe-return-actions {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
  flex-wrap: wrap;
  margin-top: 1.5rem;
}

.stripe-return-button {
  border: 1px solid var(--tz-border-strong);
  border-radius: 8px;
  padding: 0.7rem 1rem;
  color: var(--tz-text-primary);
  font-size: 0.875rem;
  font-weight: 700;
}

.stripe-return-button--primary {
  background: #fff;
  color: #020617;
}
</style>
