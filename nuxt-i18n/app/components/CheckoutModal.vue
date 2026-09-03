<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="isCheckoutOpen"
        class="fixed inset-0 z-[12000] flex items-center justify-center bg-slate-900/20 p-0 backdrop-blur-sm md:p-5 tz-mobile-dialog-mask"
        role="dialog"
        aria-modal="true"
        @click.self="closeCheckout"
      >
        <section class="checkout-shell tz-mobile-dialog-surface relative flex h-full w-full max-w-6xl flex-col overflow-hidden border tz-border-subtle tz-surface-card tz-text-primary shadow-lg md:h-[min(92vh,900px)] md:rounded-2xl">
          <header class="flex shrink-0 items-center justify-between border-b tz-border-subtle px-4 py-3 md:px-6">
            <div>
              <p class="text-[10px] uppercase tracking-[0.24em] tz-text-primary/45">{{ t('checkout.modal.title', 'Checkout') }}</p>
              <h2 class="mt-1 text-base font-semibold">{{ t('checkout.stepper.review.continueToCheckout', 'Complete your order') }}</h2>
            </div>
            <div class="flex items-center gap-2">
              <button
                type="button"
                class="inline-flex items-center gap-1.5 rounded-lg border tz-border-subtle px-3 py-2 text-xs tz-text-primary/75 transition hover:tz-surface-subtle"
                @click="backToCart"
              >
                <Icon name="lucide:shopping-cart" class="h-3.5 w-3.5" />
                {{ t('checkout.modal.actions.viewCart', 'View cart') }}
              </button>
              <button
                type="button"
                class="tz-global-close-btn"
                :aria-label="t('checkout.modal.closeAriaLabel', 'Close checkout')"
                @click="closeCheckout"
              >
                <Icon name="lucide:x" class="h-3.5 w-3.5" />
              </button>
            </div>
          </header>

          <div class="min-h-0 flex-1 overflow-y-auto">
            <div class="grid gap-5 p-4 md:grid-cols-[minmax(0,1fr)_300px] md:p-6">
              <main class="space-y-5">
                <section class="border-b tz-border-subtle pb-5">
                  <div class="mb-3 flex items-center justify-between gap-3">
                    <div>
                      <h3 class="text-sm font-semibold">{{ t('checkout.steps.payment', 'Choose payment method') }}</h3>
                      <p class="mt-1 text-xs tz-text-primary/55">
                        {{ t('checkout.stepper.payment.pickProvider', 'Choose a configured payment provider') }}
                      </p>
                    </div>
                    <span v-if="paymentMethodsLoading" class="text-xs tz-text-primary/45">
                      {{ t('common.loading', 'Loading...') }}
                    </span>
                  </div>

                  <div class="grid gap-2 sm:grid-cols-2">
                    <button
                      v-for="option in visiblePaymentOptions"
                      :key="option.id"
                      type="button"
                      class="flex min-h-20 items-center gap-3 rounded-xl border px-3 py-3 text-left transition"
                      :class="paymentOptionClass(option)"
                      :disabled="!isPaymentOptionAvailable(option)"
                      :aria-disabled="!isPaymentOptionAvailable(option)"
                      @click="selectPaymentOption(option)"
                    >
                      <span class="checkout-payment-logos" aria-hidden="true">
                        <img
                          v-for="logo in paymentLogos(option)"
                          :key="logo.src"
                          :src="logo.src"
                          :alt="logo.alt"
                          :width="logo.width"
                          :height="logo.height"
                          :class="logo.className"
                          loading="lazy"
                        />
                      </span>
                      <span class="min-w-0 flex-1">
                        <span class="flex flex-wrap items-center gap-2 text-sm font-semibold">
                          {{ paymentTitle(option) }}
                          <span
                            v-if="!isPaymentOptionAvailable(option)"
                            class="rounded-full bg-amber-50 px-2 py-0.5 text-[10px] font-medium text-amber-700"
                          >
                            {{ unavailableLabel(option) }}
                          </span>
                        </span>
                        <span class="mt-1 block text-xs tz-text-primary/50">{{ paymentDescription(option) }}</span>
                      </span>
                      <Icon
                        v-if="selectedMethod === option.id"
                        name="lucide:check"
                        class="h-4 w-4 shrink-0 tz-text-primary"
                      />
                    </button>
                  </div>

                  <p v-if="paymentMethodsError" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
                    {{ paymentMethodsError }}
                  </p>
                  <p v-if="selectedOption && !selectedPaymentAvailable" class="mt-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
                    {{ unavailableLabel(selectedOption) }}
                  </p>
                </section>

                <section class="space-y-3">
                  <div>
                    <h3 class="text-sm font-semibold">{{ t('checkout.stepper.shipping.addressTitle', 'Shipping address') }}</h3>
                    <p class="mt-1 text-xs tz-text-primary/55">
                      {{ t('checkout.stepper.shipping.addressHelp', 'Use the address where this order should be delivered.') }}
                    </p>
                  </div>

                  <div class="grid gap-3 sm:grid-cols-2">
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.countryRegion', 'Country / region') }}</span>
                      <select v-model="form.country" class="checkout-input">
                        <option value="" disabled>{{ t('checkout.stepper.shipping.selectCountry', 'Select country') }}</option>
                        <option v-for="country in COUNTRIES" :key="country.code" :value="country.code">
                          {{ countryLabel(country) }}
                        </option>
                      </select>
                    </label>
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.recipient', 'Recipient') }}</span>
                      <input v-model.trim="form.name" class="checkout-input" type="text" autocomplete="name" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.phone', 'Phone') }}</span>
                      <input v-model.trim="form.phone" class="checkout-input" type="tel" autocomplete="tel" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.city', 'City') }}</span>
                      <input v-model.trim="form.city" class="checkout-input" type="text" autocomplete="address-level2" />
                    </label>
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.address', 'Address') }}</span>
                      <input v-model.trim="form.address" class="checkout-input" type="text" autocomplete="street-address" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.zip', 'Postal code') }}</span>
                      <input
                        v-model.trim="form.zip"
                        class="checkout-input"
                        type="text"
                        autocomplete="postal-code"
                        :placeholder="zipPlaceholder"
                      />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.review.orderNotes', 'Order notes') }}</span>
                      <input v-model.trim="form.notes" class="checkout-input" type="text" />
                    </label>
                  </div>

                  <p v-if="zipHint" class="text-xs tz-text-primary/45">{{ zipHint }}</p>
                  <p v-if="shippingValidation.reason && form.country" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-700">
                    {{ shippingValidation.reason }}
                  </p>
                </section>

                <section v-if="stripePaymentSession" class="border-t tz-border-subtle pt-5">
                  <div class="mb-3">
                    <h3 class="text-sm font-semibold">{{ t('checkout.payment.stripe.title', 'Secure card payment') }}</h3>
                    <p class="mt-1 text-xs tz-text-primary/55">
                      {{ t('checkout.payment.stripe.description', 'Your card details are handled by Stripe.') }}
                    </p>
                  </div>
                  <StripePaymentElement
                    :session="stripePaymentSession"
                    :confirm-label="t('checkout.payment.stripe.confirm', 'Confirm payment')"
                    :confirming-label="t('checkout.payment.stripe.confirming', 'Confirming...')"
                    :disabled="isSubmitting"
                    @confirmed="handleStripeConfirmed"
                    @error="handleStripeError"
                  />
                </section>

                 <p v-if="checkoutError" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-xs text-rose-700">
                  {{ checkoutError }}
                </p>

                <section
                  v-if="gatewayFallbackOptions.length"
                   class="rounded-xl border border-amber-200 bg-amber-50 p-3"
                  aria-live="polite"
                >
                   <p class="text-xs font-medium text-amber-800">
                    {{ t('checkout.payment.gatewayFallback.title', 'This payment provider is having trouble') }}
                  </p>
                   <p class="mt-1 text-xs text-amber-700">
                    {{ t('checkout.payment.gatewayFallback.description', 'The previous attempt may still be processing. Choose another available method only if you did not complete it.') }}
                  </p>
                  <div class="mt-3 flex flex-wrap gap-2">
                    <button
                      v-for="option in gatewayFallbackOptions"
                      :key="`fallback-${option.id}`"
                      type="button"
                       class="inline-flex items-center gap-2 rounded-lg border border-amber-200 bg-white px-3 py-2 text-xs font-medium text-amber-800 transition hover:bg-amber-100"
                      @click="selectGatewayFallbackPaymentOption(option)"
                    >
                      <span class="checkout-payment-logos" aria-hidden="true">
                        <img
                          v-for="logo in paymentLogos(option)"
                          :key="logo.src"
                          :src="logo.src"
                          :alt="logo.alt"
                          :width="logo.width"
                          :height="logo.height"
                          :class="logo.className"
                          loading="lazy"
                        />
                      </span>
                      {{ paymentTitle(option) }}
                    </button>
                  </div>
                </section>
              </main>

              <aside class="h-fit space-y-4 border-t tz-border-subtle pt-5 md:sticky md:top-0 md:border-t-0 md:border-l md:pl-5 md:pt-0">
                <div>
                  <h3 class="text-sm font-semibold">{{ t('checkout.stepper.summary.title', 'Order summary') }}</h3>
                  <div class="mt-3 space-y-3">
                    <article v-for="item in cartItems" :key="item.id" class="flex gap-3">
                      <div class="h-12 w-12 shrink-0 overflow-hidden rounded-lg border tz-border-subtle tz-surface-subtle">
                        <StorefrontImage v-if="item.thumbnail" :src="item.thumbnail" :alt="item.title" class="h-full w-full object-cover" preset="thumbnail" />
                        <div v-else class="flex h-full items-center justify-center tz-text-primary/35">
                          <Icon name="lucide:image" class="h-4 w-4" />
                        </div>
                      </div>
                      <div class="min-w-0 flex-1">
                        <p class="truncate text-xs font-medium">{{ item.title }}</p>
                        <p class="mt-1 text-xs tz-text-primary/50">× {{ item.quantity }}</p>
                      </div>
                      <span class="text-xs font-medium">{{ formatPrice(item.price * item.quantity, item.currency) }}</span>
                    </article>
                  </div>
                </div>

                <div class="space-y-2 border-t tz-border-subtle pt-4 text-sm">
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.subtotal', 'Subtotal') }}</span>
                    <span>{{ formatPrice(orderTotals.subtotal, cartCurrency) }}</span>
                  </div>
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.shipping', 'Shipping') }}</span>
                    <span>{{ shippingLabel }}</span>
                  </div>
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.tax', 'Tax') }}</span>
                    <span>{{ checkoutAmountLabel(orderTotals.tax) }}</span>
                  </div>
                  <div class="flex justify-between gap-3 border-t tz-border-subtle pt-3 text-base font-semibold">
                    <span>{{ t('checkout.stepper.summary.total', 'Total') }}</span>
                    <span>{{ checkoutAmountLabel(orderTotals.total) }}</span>
                  </div>
                </div>

                <button
                  v-if="!stripePaymentSession"
                  type="button"
                  class="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-white px-4 py-3 text-sm font-semibold text-slate-900 transition hover:bg-white/90 disabled:cursor-not-allowed disabled:opacity-45"
                  :disabled="isSubmitting || !canSubmit"
                  @click="submitOrder"
                >
                  <Icon v-if="isSubmitting" name="lucide:loader-circle" class="h-4 w-4 animate-spin" />
                  {{ isSubmitting ? t('checkout.common.processing', 'Processing...') : paymentCtaLabel }}
                </button>
                <p v-if="!selectedPaymentAvailable" class="text-center text-xs tz-text-primary/45">
                  {{ t('checkout.modal.messages.paymentUnavailable', 'This payment method is temporarily unavailable.') }}
                </p>
              </aside>
            </div>
          </div>
        </section>
      </div>
    </Transition>

    <AuthModal v-model="showAuthModal" @success="handleAuthSuccess" />
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { COUNTRIES, getCountryName, getZipFormatHint } from '~/data/countries'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useAlipayPayment } from '~/composables/useAlipayPayment'
import { usePayPalPayment } from '~/composables/usePayPalPayment'
import { useWeChatPayment, type WeChatPaymentSession } from '~/composables/useWeChatPayment'
import { useShippingValidation } from '~/composables/useShippingValidation'
import type { StripeConfirmationResult, StripePaymentSession } from '~/composables/useStripePayment'
import { ApiRequestError } from '~/composables/useApiRequest'
import type { CheckoutPaymentOption, PaymentGatewayFallbackMethod } from '~/types/payment'
import {
  isPaymentOptionAvailable,
  normalizeStorefrontPaymentMethod,
  paymentMethodFromOption,
  paymentPresentation,
  type PaymentLogoAsset,
} from '~/utils/paymentPresentation'
import { createIdempotencyKey } from '~/utils/idempotency'
import StripePaymentElement from '~/components/StripePaymentElement.vue'

type ApiResponse<T> = T | { data?: T | { data?: T } }

interface CheckoutQuote {
  subtotal_amount?: number
  shipping_fee?: number
  tax_amount?: number
  total_amount?: number
  shipping_quote?: { selected_option?: { service_name?: string; service_code?: string } }
}

interface OrderResponse {
  order_number: string
}

const { t, locale } = useI18n()
const localePath = useLocalePath()
const auth = useAuth()
const {
  cartItems,
  cartCurrency,
  isCheckoutOpen,
  preferredCheckoutPaymentMethod,
  priceBreakdown,
  formatPrice,
  clearCart,
  closeCheckout,
  backToCart,
} = useCart()
const {
  paymentMethodOptions,
  paymentMethodsLoading,
  paymentMethodsError,
  loadPaymentMethods,
} = usePaymentMethods()
const { createPayPalOrder, redirectToPayPal } = usePayPalPayment()
const { createAlipayOrder, redirectToAlipay } = useAlipayPayment()
const { createWeChatOrder } = useWeChatPayment()
const {
  loadShippingTemplates,
  validateShipping,
  getZipFormatHint: getShippingZipFormatHint,
} = useShippingValidation()
const { baseCurrency } = useStorefrontContext()

const selectedMethod = ref('card')
const checkoutError = ref('')
const gatewayFallbackOptions = ref<CheckoutPaymentOption[]>([])
const isSubmitting = ref(false)
const showAuthModal = ref(false)
const stripePaymentSession = ref<StripePaymentSession | null>(null)
const checkoutQuote = ref<CheckoutQuote | null>(null)
const checkoutSubmissionKey = ref('')
let quoteTimer: ReturnType<typeof setTimeout> | null = null

const normalizeCheckoutPaymentMethod = (value?: string | null) => {
  return normalizeStorefrontPaymentMethod(value)
}

const resetCheckoutSubmissionKey = () => {
  checkoutSubmissionKey.value = ''
}

const ensureCheckoutSubmissionKey = () => {
  if (!checkoutSubmissionKey.value) {
    checkoutSubmissionKey.value = createIdempotencyKey('checkout')
  }
  return checkoutSubmissionKey.value
}

const form = ref({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  zip: '',
  notes: '',
})

const fallbackPaymentOptions = computed<CheckoutPaymentOption[]>(() => [
  { id: 'card', code: 'card', provider: 'stripe', title: 'Credit / Debit cards', subtitle: '', description: '', enabled: true, available: false, unavailableReason: 'gateway_not_configured' },
  { id: 'paypal', code: 'paypal', provider: 'paypal', title: 'PayPal', subtitle: '', description: '', enabled: true, available: false, unavailableReason: 'gateway_not_configured' },
  { id: 'alipay', code: 'alipay', provider: 'alipay', title: 'Alipay', subtitle: '', description: '', enabled: true, available: false, unavailableReason: 'gateway_not_configured' },
  { id: 'wechat', code: 'wechat', provider: 'wechat', title: 'WeChat Pay', subtitle: '', description: '', enabled: true, available: false, unavailableReason: 'gateway_not_configured' },
])

const visiblePaymentOptions = computed(() =>
  paymentMethodOptions.value.length ? paymentMethodOptions.value : fallbackPaymentOptions.value,
)

const selectedOption = computed(() =>
  gatewayFallbackOptions.value.find(option => option.id === selectedMethod.value)
    || visiblePaymentOptions.value.find(option => option.id === selectedMethod.value)
    || null,
)
const selectedPaymentAvailable = computed(() =>
  Boolean(selectedOption.value && selectedOption.value.enabled !== false && selectedOption.value.available === true),
)

const shippingValidation = computed(() =>
  validateShipping(form.value.country, form.value.zip),
)
const checkoutEmail = computed(() => String(auth.user.value?.email || '').trim())
const zipHint = computed(() => {
  if (!form.value.country) return ''
  return getShippingZipFormatHint(form.value.country)?.hint || ''
})
const zipPlaceholder = computed(() => {
  if (!form.value.country) return ''
  return getZipFormatHint(form.value.country)?.placeholder || ''
})

const orderTotals = computed(() => {
  const local = priceBreakdown.value as {
    subtotal?: number
  }
  const quote = checkoutQuote.value
  return {
    subtotal: Number(quote?.subtotal_amount ?? local.subtotal ?? 0),
    shipping: quote ? Number(quote.shipping_fee ?? 0) : null,
    tax: quote ? Number(quote.tax_amount ?? 0) : null,
    total: quote ? Number(quote.total_amount ?? 0) : null,
  }
})

const shippingLabel = computed(() => {
  if (!form.value.country) return t('checkout.stepper.shipping.state.selectCountry', 'Select country')
  if (checkoutQuote.value?.shipping_quote?.selected_option) {
    const option = checkoutQuote.value.shipping_quote.selected_option
    return option.service_name || option.service_code || t('checkout.stepper.shipping.state.calculating', 'Calculating...')
  }
  return orderTotals.value.shipping !== null && orderTotals.value.shipping > 0
    ? formatPrice(orderTotals.value.shipping, cartCurrency.value)
    : t('checkout.stepper.shipping.state.calculating', 'Calculating...')
})

const checkoutAmountLabel = (amount: number | null) =>
  amount === null
    ? t('cartDrawer.summary.calculatedAtCheckout', 'Calculated at checkout')
    : formatPrice(amount, cartCurrency.value)

const canSubmit = computed(() =>
  cartItems.value.length > 0 &&
  selectedPaymentAvailable.value &&
  Boolean(
    checkoutEmail.value &&
    form.value.country &&
    form.value.name.trim() &&
    form.value.phone.trim() &&
    form.value.address.trim() &&
    form.value.city.trim() &&
    shippingValidation.value.isShippable,
  ),
)

const paymentCtaLabel = computed(() => {
  const method = normalizeCheckoutPaymentMethod(selectedMethod.value) || 'card'
  const presentation = paymentPresentation(method)
  return t(presentation.ctaKey, presentation.cta)
})

const paymentTitle = (option: CheckoutPaymentOption) => {
  const method = paymentMethodFromOption(option)
  if (!method) return option.title || option.code || option.id
  const presentation = paymentPresentation(method)
  return t(presentation.titleKey, presentation.title)
}

const paymentDescription = (option: CheckoutPaymentOption) => {
  if (option.description) return option.description
  const method = paymentMethodFromOption(option)
  if (!method) return option.subtitle || ''
  const presentation = paymentPresentation(method)
  return t(presentation.descriptionKey, presentation.description)
}

const paymentLogos = (option: CheckoutPaymentOption): PaymentLogoAsset[] => {
  const method = paymentMethodFromOption(option)
  return method
    ? paymentPresentation(method).logos
    : [{ src: '/icons/payment/default.svg', alt: paymentTitle(option), width: 750, height: 471 }]
}

const unavailableLabel = (option: CheckoutPaymentOption) => {
  const reason = String(option.unavailableReason || option.unavailable_reason || '').trim()
  if (reason === 'gateway_not_configured') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  if (reason === 'gateway_config_invalid') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  if (reason === 'disabled') {
    return t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
  }
  return reason ? reason.replace(/_/g, ' ') : t('checkout.payment.temporarilyUnavailable', 'Temporarily unavailable')
}

const paymentOptionClass = (option: CheckoutPaymentOption) => {
  if (!isPaymentOptionAvailable(option)) {
    return 'cursor-not-allowed tz-border-subtle tz-surface-subtle opacity-60'
  }
  return selectedMethod.value === option.id
    ? 'tz-border-strong/60 bg-white/[0.10]'
    : 'tz-border-subtle tz-surface-subtle hover:tz-border-strong/30 hover:tz-surface-subtle'
}

const countryLabel = (country: { code: string; name: string }) =>
  getCountryName(country.code, String(locale.value || 'en'))

const selectPaymentOption = (option: CheckoutPaymentOption) => {
  if (!isPaymentOptionAvailable(option)) return
  selectedMethod.value = option.id
  stripePaymentSession.value = null
  checkoutError.value = ''
  gatewayFallbackOptions.value = []
  resetCheckoutSubmissionKey()
}

const selectGatewayFallbackPaymentOption = (option: CheckoutPaymentOption) => {
  selectedMethod.value = option.id
  stripePaymentSession.value = null
  checkoutError.value = ''
  resetCheckoutSubmissionKey()
}

const buildShippingAddressPayload = () => {
  const nameParts = form.value.name.trim().split(/\s+/).filter(Boolean)
  return {
    first_name: nameParts[0] || 'Customer',
    last_name: nameParts.length > 1 ? nameParts.slice(1).join(' ') : 'User',
    address1: form.value.address.trim(),
    city: form.value.city.trim(),
    postal_code: form.value.zip.trim(),
    country: form.value.country.trim().toUpperCase(),
    phone: form.value.phone.trim(),
    email: checkoutEmail.value,
  }
}

const unwrapApiData = <T,>(payload: ApiResponse<T> | null | undefined): T | null => {
  let current: unknown = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') return (current as T) || null
    if (!('data' in current)) return current as T
    current = (current as { data?: unknown }).data
  }
  return null
}

const refreshCheckoutQuote = async () => {
  if (!isCheckoutOpen.value || !cartItems.value.length || !form.value.country || !auth.isAuthenticated.value) {
    checkoutQuote.value = null
    return
  }

  try {
    const response = await auth.request<ApiResponse<CheckoutQuote>>('/checkout/quote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({ shipping_address: buildShippingAddressPayload() }),
    })
    checkoutQuote.value = unwrapApiData<CheckoutQuote>(response)
  } catch {
    checkoutQuote.value = null
  }
}

const scheduleQuoteRefresh = () => {
  if (quoteTimer) clearTimeout(quoteTimer)
  quoteTimer = setTimeout(() => { void refreshCheckoutQuote() }, 300)
}

const ensureCheckoutData = async () => {
  await Promise.all([
    loadShippingTemplates(),
    loadPaymentMethods(form.value.country || undefined),
  ])
}

const requireAuthenticatedUser = async () => {
  const session = await auth.ensureSession()
  if (session) return true
  showAuthModal.value = true
  checkoutError.value = t('checkout.modal.messages.loginRequired', 'Please sign in before checkout.')
  return false
}

const createLocalOrder = async (idempotencyKey: string): Promise<OrderResponse> => {
  const response = await auth.request<ApiResponse<OrderResponse>>('/orders', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({
      items: cartItems.value.map(item => ({
        product_id: Number(item.product_id || item.id || 0),
        variant_id: item.variant_id || null,
        quantity: Math.max(1, Number(item.quantity || 1)),
      })),
      shipping_address: buildShippingAddressPayload(),
      payment_method: selectedMethod.value === 'card' ? 'card' : selectedMethod.value,
      shipping_method: 'standard',
    }),
  })
  const order = unwrapApiData<OrderResponse>(response)
  if (!order?.order_number) throw new Error(t('checkout.modal.messages.orderFailed', 'Order submission failed'))
  return order
}

const checkoutUrl = (path: string, orderNumber: string) => {
  if (!import.meta.client) return ''
  const target = new URL(localePath(path), window.location.origin)
  target.searchParams.set('order_number', orderNumber)
  return target.toString()
}

const startProviderPayment = async (orderNumber: string, idempotencyKey: string) => {
  if (selectedMethod.value === 'paypal') {
    const session = await createPayPalOrder({
      orderNumber,
      returnUrl: checkoutUrl('/checkout/paypal/return', orderNumber),
      cancelUrl: checkoutUrl('/checkout/paypal/cancel', orderNumber),
      idempotencyKey,
    })
    redirectToPayPal(session)
    return
  }

  if (selectedMethod.value === 'alipay') {
    const session = await createAlipayOrder({
      orderNumber,
      returnUrl: checkoutUrl('/checkout/alipay/return', orderNumber),
      cancelUrl: '',
      idempotencyKey,
    })
    redirectToAlipay(session)
    return
  }

  if (selectedMethod.value === 'wechat') {
    const session: WeChatPaymentSession = await createWeChatOrder({ orderNumber, idempotencyKey })
    if (import.meta.client) {
      window.sessionStorage.setItem(`checkout:wechat:${orderNumber}`, JSON.stringify(session))
      const target = new URL(localePath('/checkout/wechat/pay'), window.location.origin)
      target.searchParams.set('order_number', orderNumber)
      window.location.assign(target.toString())
    }
    return
  }

  const response = await auth.request<ApiResponse<StripePaymentSession & { client_secret?: string; publishable_key?: string }>>('/payment/stripe/payment-intents', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({ order_number: orderNumber }),
  })
  const session = unwrapApiData<StripePaymentSession & { client_secret?: string; publishable_key?: string }>(response)
  const clientSecret = session?.clientSecret || session?.client_secret || ''
  const publishableKey = session?.publishableKey || session?.publishable_key || ''
  if (!clientSecret || !publishableKey) throw new Error('Stripe payment response is incomplete')
  stripePaymentSession.value = { clientSecret, publishableKey }
}

const paymentGatewayFallbackMethodKey = (method: PaymentGatewayFallbackMethod) =>
  String(method.code || method.provider || '').trim().toLowerCase()

const buildGatewayFallbackPaymentOptions = (methods: PaymentGatewayFallbackMethod[]) => {
  const options: CheckoutPaymentOption[] = []
  const seen = new Set<string>()

  for (const method of methods) {
    const methodKey = paymentGatewayFallbackMethodKey(method)
    if (!methodKey || seen.has(methodKey)) continue
    seen.add(methodKey)

    const existingOption = visiblePaymentOptions.value.find((option) => {
      const code = String(option.code || '').trim().toLowerCase()
      const provider = String(option.provider || '').trim().toLowerCase()
      return code === methodKey || provider === methodKey
    })
    if (existingOption) {
      options.push({
        ...existingOption,
        available: true,
        unavailableReason: undefined,
        unavailable_reason: undefined,
      })
      continue
    }

    options.push({
      id: String(method.code || method.provider || '').trim(),
      code: method.code,
      provider: method.provider,
      title: method.name || method.code || method.provider || '',
      subtitle: method.provider || '',
      description: '',
      enabled: true,
      available: true,
    })
  }
  return options
}

const applyPaymentGatewayFallbackRecommendation = (error: unknown) => {
  if (!(error instanceof ApiRequestError) || error.code !== 'payment_gateway_degraded') {
    return false
  }

  const details = error.details && typeof error.details === 'object'
    ? error.details as { fallback_payment_methods?: PaymentGatewayFallbackMethod[] }
    : null
  const fallbackMethods = Array.isArray(details?.fallback_payment_methods)
    ? details.fallback_payment_methods
    : []
  gatewayFallbackOptions.value = buildGatewayFallbackPaymentOptions(fallbackMethods)
  return true
}

const submitOrder = async () => {
  if (isSubmitting.value) return
  checkoutError.value = ''

  if (!selectedPaymentAvailable.value) {
    checkoutError.value = unavailableLabel(selectedOption.value || fallbackPaymentOptions.value[0]!)
    return
  }
  if (!cartItems.value.length) {
    checkoutError.value = t('checkout.modal.messages.emptyCart', 'Your cart is empty.')
    return
  }
  if (!(await requireAuthenticatedUser())) return
  if (!checkoutEmail.value) {
    checkoutError.value = t(
      'checkout.modal.messages.emailRequired',
      'Add an email address to your account before checkout.',
    )
    return
  }
  if (!canSubmit.value) {
    checkoutError.value = t('checkout.modal.messages.completeShipping', 'Please complete your shipping address and contact details.')
    return
  }

  isSubmitting.value = true
  try {
    const idempotencyKey = ensureCheckoutSubmissionKey()
    await refreshCheckoutQuote()
    const order = await createLocalOrder(idempotencyKey)
    await startProviderPayment(order.order_number, idempotencyKey)
  } catch (error) {
    if (applyPaymentGatewayFallbackRecommendation(error)) {
      checkoutError.value = t(
        'checkout.payment.gatewayFallback.error',
        'The selected payment provider is temporarily unavailable. Please choose another available payment method.',
      )
    } else {
      checkoutError.value = error instanceof Error
        ? error.message
        : t('checkout.modal.messages.orderFailed', 'Order submission failed')
    }
  } finally {
    isSubmitting.value = false
  }
}

const handleStripeConfirmed = (result: StripeConfirmationResult) => {
  stripePaymentSession.value = null
  gatewayFallbackOptions.value = []
  if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
    clearCart()
    closeCheckout()
    checkoutError.value = ''
    return
  }
  checkoutError.value = t('checkout.modal.messages.paymentPending', 'Payment is not complete yet. Please check your order status later.')
}

const handleStripeError = (message: string) => {
  checkoutError.value = message
}

const handleAuthSuccess = async () => {
  showAuthModal.value = false
  checkoutError.value = ''
  await refreshCheckoutQuote()
}

watch(isCheckoutOpen, (open) => {
  if (open) {
    const preferredMethod = normalizeCheckoutPaymentMethod(preferredCheckoutPaymentMethod.value)
    if (preferredMethod) {
      selectedMethod.value = preferredMethod
      stripePaymentSession.value = null
      checkoutError.value = ''
      gatewayFallbackOptions.value = []
    }
    void ensureCheckoutData()
    void auth.ensureSession()
  } else {
    stripePaymentSession.value = null
    checkoutError.value = ''
    gatewayFallbackOptions.value = []
    resetCheckoutSubmissionKey()
  }
}, { immediate: true })

watch(preferredCheckoutPaymentMethod, (method) => {
  if (!isCheckoutOpen.value) return
  const preferredMethod = normalizeCheckoutPaymentMethod(method)
  if (!preferredMethod) return
  selectedMethod.value = preferredMethod
  stripePaymentSession.value = null
  checkoutError.value = ''
  gatewayFallbackOptions.value = []
  resetCheckoutSubmissionKey()
})

watch(() => form.value.country, () => {
  if (isCheckoutOpen.value) {
    resetCheckoutSubmissionKey()
    void loadPaymentMethods(form.value.country || undefined)
    scheduleQuoteRefresh()
  }
})

watch(
  () => [form.value.name, form.value.phone, form.value.address, form.value.city, form.value.zip],
  () => {
    if (isCheckoutOpen.value) {
      resetCheckoutSubmissionKey()
      scheduleQuoteRefresh()
    }
  },
)

watch(
  () => [
    cartCurrency.value,
    ...cartItems.value.map(item => [
      item.product_id || item.id,
      item.variant_id || '',
      item.quantity || 0,
    ].join(':')),
  ],
  () => {
    if (isCheckoutOpen.value) {
      resetCheckoutSubmissionKey()
    }
  },
)

onBeforeUnmount(() => {
  if (quoteTimer) clearTimeout(quoteTimer)
})
</script>

<style scoped>
.checkout-shell {
  background-image: none;
}

.checkout-label {
  display: block;
  margin-bottom: 0.35rem;
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.checkout-input {
  width: 100%;
  min-height: 2.65rem;
  border: 1px solid var(--tz-form-control-border);
  border-radius: 0.7rem;
  background: var(--tz-form-control-surface);
  padding: 0.65rem 0.8rem;
  color: var(--tz-text-primary);
  font-size: 0.8rem;
  outline: none;
}

.checkout-input:focus {
  border-color: var(--tz-form-control-focus-border);
  box-shadow: 0 0 0 1px var(--tz-form-control-focus-ring);
}

.checkout-input option {
  background: var(--tz-form-control-surface);
  color: var(--tz-text-primary);
}

.checkout-payment-logos {
  display: inline-flex;
  min-width: 3.25rem;
  max-width: 4.9rem;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-start;
  gap: 0.18rem;
}

.checkout-payment-logos img {
  display: block;
  width: auto;
  max-width: 2.35rem;
  height: 1rem;
  object-fit: contain;
}

.checkout-payment-logos img.payment-logo--alipay {
  max-width: 2.75rem;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
