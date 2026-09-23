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
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.state', 'State / province') }}</span>
                      <input v-model.trim="form.state" class="checkout-input" type="text" autocomplete="address-level1" />
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

                  <div v-if="checkoutQuote?.shipping_quote?.plans?.length" class="space-y-2 pt-2">
                    <p class="checkout-label">{{ t('checkout.stepper.shipping.method', 'Delivery method') }}</p>
                    <label
                      v-for="plan in checkoutQuote.shipping_quote.plans"
                      :key="plan.id"
                      class="flex cursor-pointer items-center justify-between gap-3 rounded-lg border tz-border-subtle px-3 py-2 text-sm transition hover:tz-surface-subtle"
                      :class="selectedQuotePlanID === plan.id ? 'tz-border-strong bg-white/[0.06]' : ''"
                    >
                      <span class="flex min-w-0 items-center gap-2">
                        <input
                          v-model="selectedQuotePlanID"
                          type="radio"
                          name="checkout-shipping-method"
                          :value="plan.id"
                          @change="selectShippingPlan(plan.id)"
                        />
                        <span class="min-w-0">
                          <span class="block font-medium">{{ shippingPlanLabel(plan) }}</span>
                          <span v-if="plan.eta_min_days || plan.eta_max_days" class="block text-xs tz-text-primary/50">
                            {{ plan.eta_min_days }}-{{ plan.eta_max_days }} days
                          </span>
                        </span>
                      </span>
                      <span class="shrink-0 font-medium">{{ formatMinorPrice(plan.shipping_fee_minor, checkoutCurrency) }}</span>
                    </label>
                  </div>
                </section>

                <section class="border-t tz-border-subtle pt-5">
                  <label class="checkout-policy-confirmation">
                    <input v-model="billingSameAsShipping" type="checkbox" />
                    <span>{{ t('checkout.stepper.billing.sameAsShipping', 'Billing address is the same as shipping address') }}</span>
                  </label>

                  <div v-if="!billingSameAsShipping" class="mt-4 grid gap-3 sm:grid-cols-2">
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.billing.countryRegion', 'Billing country / region') }}</span>
                      <select v-model="billingForm.country" class="checkout-input">
                        <option value="" disabled>{{ t('checkout.stepper.shipping.selectCountry', 'Select country') }}</option>
                        <option v-for="country in COUNTRIES" :key="`billing-${country.code}`" :value="country.code">
                          {{ countryLabel(country) }}
                        </option>
                      </select>
                    </label>
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.billing.recipient', 'Billing name') }}</span>
                      <input v-model.trim="billingForm.name" class="checkout-input" type="text" autocomplete="billing name" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.phone', 'Phone') }}</span>
                      <input v-model.trim="billingForm.phone" class="checkout-input" type="tel" autocomplete="billing tel" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.city', 'City') }}</span>
                      <input v-model.trim="billingForm.city" class="checkout-input" type="text" autocomplete="billing address-level2" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.state', 'State / province') }}</span>
                      <input v-model.trim="billingForm.state" class="checkout-input" type="text" autocomplete="billing address-level1" />
                    </label>
                    <label class="sm:col-span-2">
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.address', 'Address') }}</span>
                      <input v-model.trim="billingForm.address" class="checkout-input" type="text" autocomplete="billing street-address" />
                    </label>
                    <label>
                      <span class="checkout-label">{{ t('checkout.stepper.shipping.zip', 'Postal code') }}</span>
                      <input
                        v-model.trim="billingForm.zip"
                        class="checkout-input"
                        type="text"
                        autocomplete="billing postal-code"
                        :placeholder="billingZipPlaceholder"
                      />
                    </label>
                  </div>
                  <p v-if="!billingSameAsShipping && billingZipHint" class="mt-2 text-xs tz-text-primary/45">{{ billingZipHint }}</p>
                  <p v-if="!billingSameAsShipping && !billingAddressComplete" class="mt-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-800">
                    {{ t('checkout.stepper.billing.completeAddress', 'Please complete your billing address before continuing.') }}
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
                    :billing-details="stripeBillingDetails"
                    :return-url="stripeReturnUrl"
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
                      <span class="text-xs font-medium">{{ formatMinorPrice(item.price_minor * item.quantity, item.currency || checkoutCurrency) }}</span>
                    </article>
                  </div>
                </div>

                <section class="checkout-fulfillment-notice" aria-live="polite">
                  <div class="checkout-fulfillment-notice__heading">
                    <Icon name="lucide:factory" class="h-4 w-4" aria-hidden="true" />
                    <strong>{{ t('checkout.fulfillment.title', 'Fulfillment and delivery') }}</strong>
                  </div>
                  <p v-if="hasMadeToOrderItems">
                    {{ t('checkout.fulfillment.madeToOrder', 'This cart includes made-to-order items. Production begins after payment confirmation, and the production schedule is added to the shipping time.') }}
                  </p>
                  <p v-if="hasMadeToOrderItems">
                    {{ t('checkout.fulfillment.cancellation', 'Cancellation is available only before production or material cutting begins. After production starts, cancellation may be restricted and a custom handling fee may apply.') }}
                  </p>
                  <p class="checkout-signature-note">
                    <Icon name="lucide:pen-line" class="h-3.5 w-3.5" aria-hidden="true" />
                    {{ t('checkout.fulfillment.signature', 'Orders totaling $750 USD or more require a signature at delivery.') }}
                  </p>
                </section>

                <label
                  v-if="fulfillmentDisclosureRequired"
                  class="checkout-policy-confirmation"
                >
                  <input v-model="policyDisclosureAcknowledged" type="checkbox" />
                  <span>{{ t('checkout.fulfillment.confirmation', 'I understand the production, cancellation, delivery, and signature requirements for this order.') }}</span>
                </label>

                <div class="space-y-2 border-t tz-border-subtle pt-4 text-sm">
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.subtotal', 'Subtotal') }}</span>
                    <span>{{ formatMinorPrice(orderTotals.subtotalMinor, checkoutCurrency) }}</span>
                  </div>
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.shipping', 'Shipping') }}</span>
                    <span>{{ shippingLabel }}</span>
                  </div>
                  <div class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.tax', 'Tax') }}</span>
                    <span>{{ checkoutAmountLabel(orderTotals.taxMinor) }}</span>
                  </div>
                  <div v-if="orderTotals.couponDiscountMinor > 0" class="flex justify-between gap-3 tz-text-muted">
                    <span>{{ t('checkout.stepper.summary.couponDiscount', 'Coupon discount') }}</span>
                    <span>-{{ formatMinorPrice(orderTotals.couponDiscountMinor, checkoutCurrency) }}</span>
                  </div>
                  <div class="flex justify-between gap-3 border-t tz-border-subtle pt-3 text-base font-semibold">
                    <span>{{ t('checkout.stepper.summary.total', 'Total') }}</span>
                    <span>{{ checkoutAmountLabel(orderTotals.totalMinor) }}</span>
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
import { navigateTo, useI18n, useLocalePath } from '#imports'
import { COUNTRIES, getCountryName, getZipFormatHint, validateZipFormat } from '~/data/countries'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useAlipayPayment } from '~/composables/useAlipayPayment'
import { usePayPalPayment } from '~/composables/usePayPalPayment'
import { useWeChatPayment, type WeChatPaymentSession } from '~/composables/useWeChatPayment'
import { useShippingValidation } from '~/composables/useShippingValidation'
import {
  useShippingQuote,
  type CheckoutQuoteResult,
  type ShippingQuotePlan,
} from '~/composables/useShippingQuote'
import type {
  StripeConfirmationResult,
  StripePaymentBillingDetails,
  StripePaymentSession,
} from '~/composables/useStripePayment'
import { ApiRequestError } from '~/composables/useApiRequest'
import type { CheckoutPaymentOption, PaymentGatewayFallbackMethod } from '~/types/payment'
import {
  HIGH_VALUE_SIGNATURE_THRESHOLD_USD,
  isMadeToOrderFulfillment,
} from '~/utils/fulfillmentPresentation'
import {
  isPaymentOptionAvailable,
  normalizeStorefrontPaymentMethod,
  paymentMethodFromOption,
  paymentPresentation,
  type PaymentLogoAsset,
} from '~/utils/paymentPresentation'
import { createIdempotencyKey } from '~/utils/idempotency'
import {
  STRIPE_RETURN_PATH,
  clearStripeReturnSession,
  saveStripeReturnSession,
} from '~/utils/stripeReturn'
import StripePaymentElement from '~/components/StripePaymentElement.vue'
import { formatMinorMoney, minorToMajor } from '~/utils/money'

type ApiResponse<T> = T | { data?: T | { data?: T } }

interface OrderResponse {
  order_number: string
  total_minor?: number | string | null
  payment_status?: string | null
}

interface CheckoutAddressForm {
  country: string
  name: string
  phone: string
  address: string
  city: string
  state: string
  zip: string
}

const { t, locale } = useI18n()
const localePath = useLocalePath()
const auth = useAuth()
const shippingQuoteApi = useShippingQuote()
const {
  cartItems,
  cartCurrency,
  isCheckoutOpen,
  preferredCheckoutPaymentMethod,
  priceBreakdown,
  clearCart,
  reloadCartFromBackend,
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
  validateShipping,
  getZipFormatHint: getShippingZipFormatHint,
} = useShippingValidation()
const { displayCurrency } = useStorefrontContext()

const selectedMethod = ref('card')
const checkoutError = ref('')
const gatewayFallbackOptions = ref<CheckoutPaymentOption[]>([])
const isSubmitting = ref(false)
const showAuthModal = ref(false)
const stripePaymentSession = ref<StripePaymentSession | null>(null)
const checkoutQuote = ref<CheckoutQuoteResult | null>(null)
const selectedQuotePlanID = ref<string | null>(null)
const checkoutSubmissionKey = ref('')
const policyDisclosureAcknowledged = ref(false)
const billingSameAsShipping = ref(true)
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

const form = ref<CheckoutAddressForm & { notes: string }>({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  state: '',
  zip: '',
  notes: '',
})

const billingForm = ref<CheckoutAddressForm>({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  state: '',
  zip: '',
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
const billingZipHint = computed(() => {
  if (!billingForm.value.country) return ''
  return getShippingZipFormatHint(billingForm.value.country)?.hint || ''
})
const billingZipPlaceholder = computed(() => {
  if (!billingForm.value.country) return ''
  return getZipFormatHint(billingForm.value.country)?.placeholder || ''
})
const billingAddressComplete = computed(() => {
  if (billingSameAsShipping.value) return true
  return Boolean(
    billingForm.value.country &&
    billingForm.value.name.trim() &&
    billingForm.value.phone.trim() &&
    billingForm.value.address.trim() &&
    billingForm.value.city.trim() &&
    billingForm.value.zip.trim() &&
    validateZipFormat(billingForm.value.country, billingForm.value.zip),
  )
})

const orderTotals = computed(() => {
  const local = priceBreakdown.value as { subtotal_minor?: number; subtotal?: number }
  const quote = checkoutQuote.value
  return {
    subtotalMinor: quote
      ? Number(quote.subtotal_minor ?? 0)
      : Number(local.subtotal_minor ?? local.subtotal ?? 0),
    shippingMinor: quote ? Number(quote.shipping_fee_minor ?? 0) : null,
    taxMinor: quote ? Number(quote.tax_minor ?? 0) : null,
    couponDiscountMinor: quote ? Number(quote.coupon_discount_minor ?? 0) : 0,
    totalMinor: quote ? Number(quote.total_minor ?? 0) : null,
  }
})

const checkoutCurrency = computed(() => String(
  checkoutQuote.value?.currency || cartCurrency.value || displayCurrency.value || 'USD',
).trim().toUpperCase())

const formatMinorPrice = (minor: number | string | null | undefined, currency: string) => (
  formatMinorMoney(minor, currency)
)

const shippingLabel = computed(() => {
  if (!form.value.country) return t('checkout.stepper.shipping.state.selectCountry', 'Select country')
  if (checkoutQuote.value?.shipping_quote?.selected_plan) {
    return shippingPlanLabel(checkoutQuote.value.shipping_quote.selected_plan)
  }
  return orderTotals.value.shippingMinor !== null && orderTotals.value.shippingMinor > 0
    ? formatMinorPrice(orderTotals.value.shippingMinor, checkoutCurrency.value)
    : t('checkout.stepper.shipping.state.calculating', 'Calculating...')
})

const shippingPlanLabel = (plan?: ShippingQuotePlan | null) => {
  const labels = (plan?.legs || []).map(leg => (
    leg.service_name || leg.service_code || leg.template_name || ''
  )).filter(Boolean)
  return labels.join(' + ') || t('checkout.stepper.shipping.state.calculating', 'Calculating...')
}

const selectShippingPlan = (planID: string) => {
  const normalized = String(planID || '').trim()
  if (!normalized) return
  selectedQuotePlanID.value = normalized
  resetCheckoutSubmissionKey()
  scheduleQuoteRefresh()
}

const checkoutAmountLabel = (amountMinor: number | null) =>
  amountMinor === null
    ? t('cartDrawer.summary.calculatedAtCheckout', 'Calculated at checkout')
    : formatMinorPrice(amountMinor, checkoutCurrency.value)

const hasMadeToOrderItems = computed(() => cartItems.value.some(item => (
  isMadeToOrderFulfillment(item.fulfillment_mode)
)))

const signatureCheckAmount = computed(() => Number(
  minorToMajor(
    orderTotals.value.totalMinor ?? orderTotals.value.subtotalMinor ?? 0,
    checkoutCurrency.value,
  ),
))

const highValueSignatureRequired = computed(() => (
  cartCurrency.value === 'USD'
  && signatureCheckAmount.value >= HIGH_VALUE_SIGNATURE_THRESHOLD_USD
))

const fulfillmentDisclosureRequired = computed(() => (
  hasMadeToOrderItems.value || highValueSignatureRequired.value
))

const canSubmit = computed(() =>
  cartItems.value.length > 0 &&
  selectedPaymentAvailable.value &&
  (!fulfillmentDisclosureRequired.value || policyDisclosureAcknowledged.value) &&
  Boolean(
    checkoutEmail.value &&
    form.value.country &&
    form.value.name.trim() &&
    form.value.phone.trim() &&
    form.value.address.trim() &&
    form.value.city.trim() &&
    form.value.zip.trim() &&
    validateZipFormat(form.value.country, form.value.zip) &&
    shippingValidation.value.isShippable &&
    billingAddressComplete.value,
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

const addressPayloadFromForm = (addressForm: CheckoutAddressForm) => {
  const nameParts = addressForm.name.trim().split(/\s+/).filter(Boolean)
  return {
    first_name: nameParts[0] || 'Customer',
    last_name: nameParts.length > 1 ? nameParts.slice(1).join(' ') : 'User',
    address1: addressForm.address.trim(),
    city: addressForm.city.trim(),
    state: addressForm.state.trim(),
    postal_code: addressForm.zip.trim(),
    country: addressForm.country.trim().toUpperCase(),
    phone: addressForm.phone.trim(),
    email: checkoutEmail.value,
  }
}

const buildShippingAddressPayload = () => addressPayloadFromForm(form.value)

const buildBillingAddressPayload = () => {
  if (billingSameAsShipping.value) return buildShippingAddressPayload()
  return addressPayloadFromForm(billingForm.value)
}

const stripeBillingDetails = computed<StripePaymentBillingDetails>(() => {
  const addressForm = billingSameAsShipping.value ? form.value : billingForm.value
  return {
    name: addressForm.name.trim(),
    email: checkoutEmail.value,
    phone: addressForm.phone.trim(),
    address: {
      line1: addressForm.address.trim(),
      city: addressForm.city.trim(),
      ...(addressForm.state.trim() ? { state: addressForm.state.trim() } : {}),
      postal_code: addressForm.zip.trim(),
      country: addressForm.country.trim().toUpperCase(),
    },
  }
})

const shippingUnavailableMessage = () => t(
  'checkout.stepper.shipping.unavailableFallback',
  'Shipping unavailable for this country.',
)

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
    const nextQuote = await shippingQuoteApi.quoteCheckout({
      shipping_address: buildShippingAddressPayload(),
      display_currency: String(displayCurrency.value || '').trim().toUpperCase(),
      payment_method: selectedMethod.value === 'card' ? 'card' : selectedMethod.value,
      ...(selectedQuotePlanID.value && checkoutQuote.value?.shipping_quote?.id
        ? {
            shipping_quote_id: checkoutQuote.value.shipping_quote.id,
            selected_quote_plan_id: selectedQuotePlanID.value,
          }
        : {}),
    })
    checkoutQuote.value = nextQuote
    if (checkoutError.value === shippingUnavailableMessage()) {
      checkoutError.value = ''
    }
    const available = nextQuote?.shipping_quote?.plans || []
    if (!selectedQuotePlanID.value || !available.some(plan => plan.id === selectedQuotePlanID.value)) {
      selectedQuotePlanID.value = nextQuote?.shipping_quote?.selected_plan?.id || null
    }
  } catch (error) {
    checkoutQuote.value = null
    if (error instanceof ApiRequestError && error.code === 'shipping_rate_unavailable') {
      checkoutError.value = shippingUnavailableMessage()
    }
  }
}

const scheduleQuoteRefresh = () => {
  if (quoteTimer) clearTimeout(quoteTimer)
  quoteTimer = setTimeout(() => { void refreshCheckoutQuote() }, 300)
}

const ensureCheckoutData = async () => {
  await Promise.all([
    loadPaymentMethods(form.value.country || undefined, checkoutCurrency.value),
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
  const expectedTotalMinor = Number(checkoutQuote.value?.total_minor)
  if (!Number.isSafeInteger(expectedTotalMinor) || expectedTotalMinor < 0) {
    throw new Error(t('checkout.modal.messages.unableRefreshQuote', 'Unable to refresh checkout quote'))
  }
  const shippingQuoteID = checkoutQuote.value?.shipping_quote?.id
  if (!shippingQuoteID || !selectedQuotePlanID.value) {
    throw new Error(shippingUnavailableMessage())
  }

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
      billing_address: buildBillingAddressPayload(),
      payment_method: selectedMethod.value === 'card' ? 'card' : selectedMethod.value,
      display_currency: String(displayCurrency.value || '').trim().toUpperCase(),
      shipping_method: 'standard',
      shipping_quote_id: shippingQuoteID,
      selected_quote_plan_id: selectedQuotePlanID.value,
      expected_total_minor: expectedTotalMinor,
      policy_disclosure_acknowledged: policyDisclosureAcknowledged.value,
    }),
  })
  const order = unwrapApiData<OrderResponse>(response)
  if (!order?.order_number) throw new Error(t('checkout.modal.messages.orderFailed', 'Order submission failed'))
  return order
}

const isSettledOrder = (order: OrderResponse) => {
  const paymentStatus = String(order.payment_status || '').trim().toLowerCase()
  if (paymentStatus === 'paid') return true

  if (order.total_minor === null || order.total_minor === undefined || order.total_minor === '') {
    return false
  }

  const totalMinor = Number(order.total_minor)
  return Number.isSafeInteger(totalMinor) && totalMinor <= 0
}

const completeSettledOrder = async (orderNumber: string) => {
  await clearCart()
  await reloadCartFromBackend()
  await closeCheckout()
  await navigateTo({
    path: localePath('/checkout/success'),
    query: { order_number: orderNumber },
  })
}

const checkoutUrl = (path: string, orderNumber: string) => {
  if (!import.meta.client) return ''
  const target = new URL(localePath(path), window.location.origin)
  target.searchParams.set('order_number', orderNumber)
  return target.toString()
}

const stripeReturnUrl = computed(() => {
  const orderNumber = stripePaymentSession.value?.orderNumber
  return orderNumber ? checkoutUrl(STRIPE_RETURN_PATH, orderNumber) : ''
})

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
  stripePaymentSession.value = { clientSecret, publishableKey, orderNumber }
  saveStripeReturnSession({
    orderNumber,
    clientSecret,
    publishableKey,
  })
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
    if (fulfillmentDisclosureRequired.value && !policyDisclosureAcknowledged.value) {
      checkoutError.value = t(
        'checkout.fulfillment.confirmRequired',
        'Please confirm the custom-order and delivery requirements before continuing.',
      )
      return
    }
    checkoutError.value = t('checkout.modal.messages.completeShipping', 'Please complete your shipping address and contact details.')
    return
  }

  isSubmitting.value = true
  try {
    const idempotencyKey = ensureCheckoutSubmissionKey()
    await refreshCheckoutQuote()
    if (!checkoutQuote.value) {
      if (!checkoutError.value) {
        checkoutError.value = t('checkout.modal.messages.unableRefreshQuote', 'Unable to refresh checkout quote')
      }
      return
    }
    const order = await createLocalOrder(idempotencyKey)
    if (isSettledOrder(order)) {
      await completeSettledOrder(order.order_number)
      return
    }
    await startProviderPayment(order.order_number, idempotencyKey)
  } catch (error) {
    if (applyPaymentGatewayFallbackRecommendation(error)) {
      checkoutError.value = t(
        'checkout.payment.gatewayFallback.error',
        'The selected payment provider is temporarily unavailable. Please choose another available payment method.',
      )
    } else if (error instanceof ApiRequestError && error.code === 'order_total_changed') {
      resetCheckoutSubmissionKey()
      await refreshCheckoutQuote()
      checkoutError.value = error.message || t(
        'checkout.modal.messages.priceUpdated',
        'The price has been updated. Please review the new order total.',
      )
    } else if (error instanceof ApiRequestError && error.code === 'checkout_cart_already_consumed') {
      await reloadCartFromBackend()
      resetCheckoutSubmissionKey()
      checkoutError.value = error.message || t(
        'checkout.modal.messages.checkoutAlreadySubmitted',
        'This cart was already submitted as an order. Please check your order status before trying again.',
      )
    } else if (error instanceof ApiRequestError && error.code === 'shipping_rate_unavailable') {
      checkoutError.value = shippingUnavailableMessage()
    } else {
      checkoutError.value = error instanceof Error
        ? error.message
        : t('checkout.modal.messages.orderFailed', 'Order submission failed')
    }
  } finally {
    isSubmitting.value = false
  }
}

watch(billingSameAsShipping, (same) => {
  if (!same) {
    billingForm.value = {
      country: form.value.country,
      name: form.value.name,
      phone: form.value.phone,
      address: form.value.address,
      city: form.value.city,
      state: form.value.state,
      zip: form.value.zip,
    }
  }
  resetCheckoutSubmissionKey()
})

watch(
  () => [
    billingForm.value.country,
    billingForm.value.name,
    billingForm.value.phone,
    billingForm.value.address,
    billingForm.value.city,
    billingForm.value.state,
    billingForm.value.zip,
  ],
  () => {
    if (isCheckoutOpen.value && !billingSameAsShipping.value) {
      resetCheckoutSubmissionKey()
    }
  },
)

const handleStripeConfirmed = async (result: StripeConfirmationResult) => {
  const orderNumber = stripePaymentSession.value?.orderNumber || ''
  stripePaymentSession.value = null
  gatewayFallbackOptions.value = []
  if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
    if (orderNumber) {
      clearStripeReturnSession(orderNumber)
      await completeSettledOrder(orderNumber)
    } else {
      await clearCart()
      await reloadCartFromBackend()
      await closeCheckout()
    }
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
    policyDisclosureAcknowledged.value = false
    billingSameAsShipping.value = true
    billingForm.value = {
      country: '',
      name: '',
      phone: '',
      address: '',
      city: '',
      state: '',
      zip: '',
    }
    resetCheckoutSubmissionKey()
    selectedQuotePlanID.value = null
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
    selectedQuotePlanID.value = null
    checkoutQuote.value = null
    resetCheckoutSubmissionKey()
    void loadPaymentMethods(form.value.country || undefined, checkoutCurrency.value)
    scheduleQuoteRefresh()
  }
})

watch(selectedMethod, () => {
  if (isCheckoutOpen.value) {
    selectedQuotePlanID.value = null
    checkoutQuote.value = null
    resetCheckoutSubmissionKey()
    scheduleQuoteRefresh()
  }
})

watch(
  () => [form.value.name, form.value.phone, form.value.address, form.value.city, form.value.zip],
  () => {
    if (isCheckoutOpen.value) {
      selectedQuotePlanID.value = null
      checkoutQuote.value = null
      resetCheckoutSubmissionKey()
      scheduleQuoteRefresh()
    }
  },
)

watch(
  () => [
    cartCurrency.value,
    displayCurrency.value,
    ...cartItems.value.map(item => [
      item.product_id || item.id,
      item.variant_id || '',
      item.quantity || 0,
    ].join(':')),
  ],
  () => {
    if (isCheckoutOpen.value) {
      selectedQuotePlanID.value = null
      checkoutQuote.value = null
      resetCheckoutSubmissionKey()
      policyDisclosureAcknowledged.value = false
      scheduleQuoteRefresh()
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

.checkout-fulfillment-notice {
  display: grid;
  gap: 0.45rem;
  border: 1px solid rgb(245 158 11 / 0.35);
  border-left: 3px solid #f59e0b;
  border-radius: 0.65rem;
  background: rgb(245 158 11 / 0.08);
  padding: 0.85rem 0.95rem;
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  line-height: 1.55;
}

.checkout-fulfillment-notice p,
.checkout-signature-note {
  margin: 0;
}

.checkout-fulfillment-notice__heading,
.checkout-signature-note {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
}

.checkout-signature-note {
  color: var(--tz-text-secondary);
}

.checkout-policy-confirmation {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  color: var(--tz-text-secondary);
  font-size: 0.78rem;
  line-height: 1.55;
}

.checkout-policy-confirmation input {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
  margin-top: 0.12rem;
  accent-color: var(--tz-site-accent);
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
