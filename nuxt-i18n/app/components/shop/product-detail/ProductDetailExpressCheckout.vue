<template>
  <div class="product-express-checkout">
    <StripeExpressCheckoutElement
      ref="stripeExpressCheckoutElementRef"
      :publishable-key="publishableKey"
      :amount="amount"
      :currency="currency"
      :line-items="lineItems"
      :allowed-shipping-countries="allowedShippingCountries"
      :disabled="disabled"
      @ready="emit('ready', $event)"
      @available-payment-methods-change="emit('available-payment-methods-change', $event)"
      @confirm="emit('confirm', $event)"
      @shipping-address-change="emit('shipping-address-change', $event)"
      @shipping-rate-change="emit('shipping-rate-change', $event)"
      @error="emit('error', $event)"
    />
    <p v-if="error" class="product-express-checkout__error">
      {{ error }}
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type {
  StripeExpressCheckoutElementConfirmEvent,
  StripeExpressCheckoutElementShippingAddressChangeEvent,
  StripeExpressCheckoutElementShippingRateChangeEvent,
} from '@stripe/stripe-js'
import type { StripeExpressCheckoutAvailablePaymentMethods } from '~/composables/useStripeExpressCheckout'
import StripeExpressCheckoutElement from '~/components/StripeExpressCheckoutElement.vue'

interface StripeExpressCheckoutElementExposed {
  submitExpressCheckoutPayment: () => Promise<void>
  confirmExpressCheckoutPayment: (clientSecret: string, returnUrl: string) => Promise<{ status: string; paymentIntentId?: string }>
  resetExpressCheckoutPaymentState: () => void
}

export interface ProductDetailExpressCheckoutExposed {
  submitExpressCheckoutPayment: () => Promise<void>
  confirmExpressCheckoutPayment: (clientSecret: string, returnUrl: string) => Promise<{ status: string; paymentIntentId?: string }>
  resetExpressCheckoutPaymentState: () => void
}

defineProps<{
  publishableKey: string
  amount: number
  currency: string
  lineItems: Array<{ name: string; amount: number }>
  allowedShippingCountries: string[]
  disabled: boolean
  error?: string
}>()

const emit = defineEmits<{
  (event: 'ready', value: StripeExpressCheckoutAvailablePaymentMethods): void
  (event: 'available-payment-methods-change', value: StripeExpressCheckoutAvailablePaymentMethods): void
  (event: 'confirm', value: StripeExpressCheckoutElementConfirmEvent): void
  (event: 'shipping-address-change', value: StripeExpressCheckoutElementShippingAddressChangeEvent): void
  (event: 'shipping-rate-change', value: StripeExpressCheckoutElementShippingRateChangeEvent): void
  (event: 'error', value: string): void
}>()

const stripeExpressCheckoutElementRef = ref<StripeExpressCheckoutElementExposed | null>(null)

defineExpose<ProductDetailExpressCheckoutExposed>({
  submitExpressCheckoutPayment: async () => {
    await stripeExpressCheckoutElementRef.value?.submitExpressCheckoutPayment()
  },
  confirmExpressCheckoutPayment: async (clientSecret, returnUrl) => {
    if (!stripeExpressCheckoutElementRef.value) {
      throw new Error('Express Checkout is not ready')
    }
    return stripeExpressCheckoutElementRef.value.confirmExpressCheckoutPayment(clientSecret, returnUrl)
  },
  resetExpressCheckoutPaymentState: () => {
    stripeExpressCheckoutElementRef.value?.resetExpressCheckoutPaymentState()
  },
})
</script>

<style scoped>
.product-express-checkout {
  display: grid;
  gap: 0.5rem;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 0.85rem;
}

.product-express-checkout__error {
  margin: 0;
  border: 1px solid rgba(251, 113, 133, 0.22);
  border-radius: 0.55rem;
  background: var(--tz-status-danger-bg);
  color: var(--tz-status-danger-text);
  font-size: 0.72rem;
  line-height: 1.4;
  padding: 0.55rem 0.7rem;
}
</style>
