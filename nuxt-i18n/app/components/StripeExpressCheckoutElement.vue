<template>
  <section
    v-if="shouldRenderExpressCheckout"
    class="stripe-express-checkout"
    aria-live="polite"
  >
    <div
      ref="expressCheckoutContainer"
      class="min-h-12"
      :class="{ 'pointer-events-none opacity-60': disabled || isConfirming }"
    />

    <p
      v-if="errorMessage"
      class="mt-2 rounded-lg border border-rose-300/25 bg-rose-300/10 px-3 py-2 text-xs text-rose-100"
    >
      {{ errorMessage }}
    </p>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import type {
  StripeExpressCheckoutElementConfirmEvent,
  StripeExpressCheckoutElementShippingAddressChangeEvent,
  StripeExpressCheckoutElementShippingRateChangeEvent,
} from '@stripe/stripe-js'
import {
  useStripeExpressCheckout,
  type StripeExpressCheckoutAvailablePaymentMethods,
  type StripeExpressCheckoutConfirmationResult,
  type StripeExpressCheckoutLineItem,
  type StripeExpressCheckoutShippingRate,
} from '~/composables/useStripeExpressCheckout'

const props = defineProps<{
  publishableKey: string
  amount: number
  currency: string
  lineItems: StripeExpressCheckoutLineItem[]
  shippingRates?: StripeExpressCheckoutShippingRate[]
  allowedShippingCountries?: string[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  (event: 'ready', availablePaymentMethods: StripeExpressCheckoutAvailablePaymentMethods): void
  (event: 'available-payment-methods-change', availablePaymentMethods: StripeExpressCheckoutAvailablePaymentMethods): void
  (event: 'confirm', confirmationEvent: StripeExpressCheckoutElementConfirmEvent): void
  (event: 'shipping-address-change', shippingEvent: StripeExpressCheckoutElementShippingAddressChangeEvent): void
  (event: 'shipping-rate-change', shippingEvent: StripeExpressCheckoutElementShippingRateChangeEvent): void
  (event: 'error', message: string): void
  (event: 'cancel'): void
  (event: 'confirmed', result: StripeExpressCheckoutConfirmationResult): void
}>()

const expressCheckoutContainer = ref<HTMLElement | null>(null)
const errorMessage = ref('')
const isConfirming = ref(false)
const {
  isReady,
  hasResolvedAvailability,
  availablePaymentMethods,
  mount,
  submit,
  confirmPayment,
  destroy,
} = useStripeExpressCheckout()

const hasAvailableWallet = computed(() =>
  availablePaymentMethods.value.applePay || availablePaymentMethods.value.googlePay,
)
const shouldRenderExpressCheckout = computed(() =>
  !hasResolvedAvailability.value || hasAvailableWallet.value,
)

const buildMountSession = () => ({
  publishableKey: props.publishableKey,
  amount: props.amount,
  currency: props.currency,
  lineItems: props.lineItems,
  shippingRates: props.shippingRates,
  allowedShippingCountries: props.allowedShippingCountries,
})

let hasMounted = false
let mountRequestVersion = 0

const mountExpressCheckoutElement = async () => {
  if (!expressCheckoutContainer.value || !props.publishableKey || props.amount <= 0) return

  const requestVersion = ++mountRequestVersion
  errorMessage.value = ''
  try {
    await mount(expressCheckoutContainer.value, buildMountSession(), {
      onReady: (event) => emit('ready', availablePaymentMethods.value),
      onAvailablePaymentMethodsChange: () => emit('available-payment-methods-change', availablePaymentMethods.value),
      onConfirm: (event) => emit('confirm', event),
      onShippingAddressChange: (event) => emit('shipping-address-change', event),
      onShippingRateChange: (event) => emit('shipping-rate-change', event),
      onCancel: () => emit('cancel'),
      onError: (error) => {
        const message = error instanceof Error ? error.message : 'Stripe Express Checkout could not be loaded'
        errorMessage.value = message
        emit('error', message)
      },
    })
    if (requestVersion !== mountRequestVersion) return
  } catch (error) {
    if (requestVersion !== mountRequestVersion) return
    const message = error instanceof Error ? error.message : 'Stripe Express Checkout could not be loaded'
    errorMessage.value = message
    emit('error', message)
  }
}

const submitExpressCheckoutPayment = async () => {
  isConfirming.value = true
  errorMessage.value = ''
  try {
    await submit()
  } catch (error) {
    isConfirming.value = false
    throw error
  }
}

const confirmExpressCheckoutPayment = async (clientSecret: string, returnUrl: string) => {
  try {
    const result = await confirmPayment(clientSecret, returnUrl)
    emit('confirmed', result)
    return result
  } finally {
    isConfirming.value = false
  }
}

const resetExpressCheckoutPaymentState = () => {
  isConfirming.value = false
}

watch(
  () => [
    props.publishableKey,
    props.amount,
    props.currency,
    JSON.stringify(props.lineItems),
    JSON.stringify(props.shippingRates || []),
  ],
  async () => {
    if (!hasMounted) return
    await nextTick()
    mountRequestVersion += 1
    destroy()
    await mountExpressCheckoutElement()
  },
)

onMounted(() => {
  hasMounted = true
  void mountExpressCheckoutElement()
})

onBeforeUnmount(() => {
  mountRequestVersion += 1
  destroy()
})

defineExpose({
  submitExpressCheckoutPayment,
  confirmExpressCheckoutPayment,
  resetExpressCheckoutPaymentState,
  isReady,
  isConfirming,
})
</script>
