<template>
  <Teleport to="body">
    <Transition name="fade">
      <div
        v-if="isCheckoutOpen"
        class="fixed inset-0 z-[12000] flex items-center justify-center p-0 md:p-4 tz-mobile-safe-modal-mask"
        aria-modal="true"
        role="dialog"
        @click.self="closeCheckout"
      >
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>

        <Transition name="scale" appear>
          <div
            v-if="isCheckoutOpen"
            class="checkout-modal-shell relative flex w-full max-w-[1400px] flex-col overflow-hidden rounded-2xl border border-white/15 bg-black text-white shadow-[0_18px_44px_rgba(0,0,0,0.72)]"
          >
            <header class="relative flex items-center justify-center border-b border-white/10 px-3 py-3 md:px-6">
              <h2 class="sr-only">{{ t('checkout.modal.title') }}</h2>
              <div class="flex items-center gap-2 overflow-x-auto">
                <button
                  type="button"
                  class="shrink-0 rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-slate-900 transition hover:bg-white/90"
                  @click="openCartFromCheckout"
                >
                  {{ t('checkout.modal.actions.viewCart') }}
                </button>
                <button
                  type="button"
                  class="shrink-0 rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-slate-900 transition hover:bg-white/90"
                  @click="handleOpenShippingChat"
                >
                  {{ t('checkout.modal.actions.livechat') }}
                </button>
                <button
                  type="button"
                  class="shrink-0 rounded-full bg-white px-3 py-1.5 text-xs font-semibold text-slate-900 transition hover:bg-white/90"
                  @click="openContactSupport"
                >
                  {{ t('checkout.modal.actions.email') }}
                </button>
              </div>
              <button
                type="button"
                class="tz-global-close-btn absolute right-3 top-1/2 -translate-y-1/2"
                :aria-label="t('checkout.modal.closeAriaLabel')"
                @click="closeCheckout"
              >
                <Icon name="lucide:x" class="h-3.5 w-3.5" />
              </button>
            </header>

            <div class="px-3 pt-2 md:px-6">
              <div class="checkout-modal-ssl-banner mx-auto flex max-w-[480px] items-center justify-center gap-2 text-center text-xs leading-tight text-emerald-100">
                <img
                  src="/checkout/secured_ssl-preview.png"
                  :alt="t('checkout.modal.sslAlt')"
                  class="h-12 w-auto md:h-16"
                  loading="lazy"
                  decoding="async"
                />
                <p>{{ t('checkout.modal.sslNote') }}</p>
              </div>
            </div>

            <p
              v-if="checkoutError"
              class="mx-3 mt-2 rounded-xl border border-rose-300/25 bg-rose-300/10 px-3 py-2 text-xs text-rose-100 md:mx-6"
            >
              {{ checkoutError }}
            </p>

            <div class="min-h-0 flex-1 overflow-y-auto px-0 pb-2 md:px-6 md:pb-4">
              <CheckoutStepper
                :initial-step="currentStepperStep"
                :initial-method="activePaymentTab"
                :coupon-input="couponCode"
                :is-applying-coupon="isApplyingCoupon"
                :applied-coupon="appliedCouponDisplayPayload"
                :points-available="calculation.userPoints.value?.available || 0"
                :is-using-points="calculation.usePointsDiscount.value"
                :points-to-use="calculation.pointsToUse.value"
                :max-points-to-use="calculation.userPoints.value?.available || 0"
                :points-hint="t('checkout.modal.pointsHint')"
                :payment-options="stepperOptions"
                :order-summary="stepperOrderSummary"
                :currency="checkoutCurrency"
                :show-shipping-form="true"
                :shipping-form="form"
                :country-search="countrySearch"
                :shippable-countries="filteredShippableCountries"
                :non-shippable-countries="filteredNonShippableCountries"
                :shipping-validation="normalizedShippingValidation"
                :estimated-delivery="estimatedDelivery"
                :zip-placeholder="zipPlaceholder"
                :zip-hint="zipHint"
                :desktop-cta-label="paymentCtaLabel"
                :cta-description="desktopCtaDescription"
                :mobile-payment-title="mobilePaymentTitle"
                :mobile-payment-description="mobilePaymentDescription"
                :is-submitting="isSubmitting"
                :stripe-payment-session="stripePaymentSession"
                :stripe-payment-confirm-label="stripePaymentConfirmLabel"
                :stripe-payment-confirming-label="stripePaymentConfirmingLabel"
                @update:step="handleStepperStepChange"
                @update:method="handleStepperSelect"
                @coupon-input="handleStepperCouponInput"
                @apply-coupon="handleApplyCoupon"
                @toggle-points="handleStepperTogglePoints"
                @points-input="handleStepperPointsInput"
                @update-shipping-field="handleStepperShippingField"
                @country-search="handleStepperCountrySearch"
                @open-contact="openContactSupport"
                @open-freight="openFreightForwarder"
                @save-cart="saveCartForLater"
                @submit="handleSubmit"
                @stripe-confirmed="handleStripeConfirmed"
                @stripe-error="handleStripeError"
              />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>

    <AuthModal v-model="showAuthModal" @success="handleAuthSuccess" />
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch, type ComputedRef } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { COUNTRIES } from '~/data/countries'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { usePaymentCurrencies } from '~/composables/usePaymentCurrencies'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useAlipayPayment } from '~/composables/useAlipayPayment'
import { usePayPalPayment } from '~/composables/usePayPalPayment'
import { useWeChatPayment, type WeChatPaymentSession } from '~/composables/useWeChatPayment'
import { useShippingValidation } from '~/composables/useShippingValidation'
import { useChatWidget } from '~/composables/useChatWidget'
import type { ShippingQuoteResult } from '~/composables/useShippingQuote'
import type { StripeConfirmationResult, StripePaymentSession } from '~/composables/useStripePayment'
import type { CheckoutPaymentOption } from '~/types/payment'

type ApiResponse<T> = T | { data?: T | { data?: T } }
type PaymentTab = 'card' | 'paypal' | 'alipay' | 'wechat' | 'stripe' | 'bank' | 'worldfirst'
type StepperStep = 1 | 2 | 3
type ShippingField = 'country' | 'name' | 'phone' | 'address' | 'city' | 'zip' | 'notes'

interface CheckoutQuote {
  subtotal_amount: number
  shipping_fee: number
  shipping_quote?: ShippingQuoteResult
  tax_amount: number
  member_discount: number
  points_discount: number
  coupon_discount: number
  discount_amount: number
  total_amount: number
  coupon_code?: string
  points_to_use: number
}

interface PublicOrderResponse {
  order_number: string
  payment_method: string
  payment_status: string
  total_amount: number
  currency: string
}

type StripePaymentIntentResponse = StripePaymentSession & {
  client_secret?: string
  publishable_key?: string
}

type CartPriceBreakdown = {
  subtotal?: number
  shipping?: number
  tax?: number
  total: number
  pointsDiscount?: number
  couponDiscount?: number
  giftCardDiscount?: number
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

const {
  cartItems,
  isCheckoutOpen,
  priceBreakdown,
  closeCheckout,
  formatPrice,
  clearCart,
  calculation,
  openCartFromCheckout,
} = useCart()
const auth = useAuth()
const { t } = useI18n()
const localePath = useLocalePath()
const { defaultOrderCurrency } = usePaymentCurrencies()
const { loadPaymentMethods, availabilityByCode } = usePaymentMethods()
const { createAlipayOrder, redirectToAlipay } = useAlipayPayment()
const { createPayPalOrder, redirectToPayPal } = usePayPalPayment()
const { createWeChatOrder } = useWeChatPayment()
const { openChat } = useChatWidget()
const {
  loadShippingTemplates,
  validateShipping,
  getShippableCountries,
  getEstimatedDeliveryText,
  getZipFormatHint,
} = useShippingValidation()

const checkoutCurrency = computed(() => defaultOrderCurrency.value || 'USD')
const typedPriceBreakdown = priceBreakdown as ComputedRef<CartPriceBreakdown>
const activePaymentTab = ref<PaymentTab>('card')
const currentStepperStep = ref<StepperStep>(1)
const checkoutQuote = ref<CheckoutQuote | null>(null)
const checkoutQuoteError = ref<string | null>(null)
const isFetchingCheckoutQuote = ref(false)
const isSubmitting = ref(false)
const isApplyingCoupon = ref(false)
const couponCode = ref('')
const countrySearch = ref('')
const checkoutError = ref('')
const showAuthModal = ref(false)
const stripePaymentSession = ref<StripePaymentSession | null>(null)
let checkoutQuoteTimer: ReturnType<typeof setTimeout> | null = null

const form = ref({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  zip: '',
  notes: '',
})

const selectedPointsToUse = computed(() =>
  calculation.usePointsDiscount.value ? calculation.pointsToUse.value || 0 : 0,
)

const paymentCopy = computed<Record<PaymentTab, { title: string; description: string; cta: string }>>(() => ({
  card: {
    title: t('checkout.payment.card.title'),
    description: t('checkout.payment.card.description'),
    cta: t('checkout.payment.card.cta'),
  },
  paypal: {
    title: t('checkout.payment.paypal.title'),
    description: t('checkout.payment.paypal.description'),
    cta: t('checkout.payment.paypal.cta'),
  },
  alipay: {
    title: t('checkout.payment.alipay.title'),
    description: t('checkout.payment.alipay.description'),
    cta: t('checkout.payment.alipay.cta'),
  },
  wechat: {
    title: t('checkout.payment.wechat.title'),
    description: t('checkout.payment.wechat.description'),
    cta: t('checkout.payment.wechat.cta'),
  },
  stripe: {
    title: t('checkout.payment.stripe.title'),
    description: t('checkout.payment.stripe.description'),
    cta: t('checkout.payment.stripe.cta'),
  },
  bank: {
    title: t('checkout.payment.bank.title'),
    description: t('checkout.payment.bank.description'),
    cta: t('checkout.payment.bank.cta'),
  },
  worldfirst: {
    title: t('checkout.payment.worldfirst.title'),
    description: t('checkout.payment.worldfirst.description'),
    cta: t('checkout.payment.worldfirst.cta'),
  },
}))

const paymentAvailability = (keys: string[]) => {
  for (const key of keys) {
    const state = availabilityByCode.value[key.toLowerCase()]
    if (state) return state
  }
  return { available: true, reason: '' }
}

const withPaymentAvailability = (
  option: CheckoutPaymentOption,
  keys: string[],
): CheckoutPaymentOption => {
  const state = paymentAvailability(keys)
  return {
    ...option,
    code: option.code || option.id,
    available: state.available,
    unavailableReason: state.reason || undefined,
  }
}

const stepperOptions = computed<CheckoutPaymentOption[]>(() => {
  const priceText = formatPrice(checkoutQuote.value?.total_amount ?? typedPriceBreakdown.value.total, checkoutCurrency.value)
  return [
    withPaymentAvailability({
      id: 'card',
      title: t('checkout.payment.card.title'),
      subtitle: `${priceText} · ${t('checkout.payment.card.subtitle')}`,
      description: t('checkout.payment.card.stepperDescription'),
      points: [t('checkout.payment.card.points.shipping'), t('checkout.payment.card.points.immediate')],
      provider: 'stripe',
    }, ['card', 'credit_card', 'stripe']),
    withPaymentAvailability({
      id: 'paypal',
      title: t('checkout.payment.paypal.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.paypal.subtitle')}`,
      description: t('checkout.payment.paypal.stepperDescription'),
      points: [t('checkout.payment.paypal.points.country')],
      provider: 'paypal',
    }, ['paypal']),
    withPaymentAvailability({
      id: 'stripe',
      title: t('checkout.payment.stripe.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.stripe.subtitle')}`,
      description: t('checkout.payment.stripe.stepperDescription'),
      points: [t('checkout.payment.stripe.points.sca')],
      provider: 'stripe',
    }, ['stripe', 'card']),
    withPaymentAvailability({
      id: 'alipay',
      title: t('checkout.payment.alipay.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.alipay.subtitle')}`,
      description: t('checkout.payment.alipay.stepperDescription'),
      points: [t('checkout.payment.alipay.points.recipient'), t('checkout.payment.alipay.points.wallets')],
      provider: 'alipay',
    }, ['alipay']),
    withPaymentAvailability({
      id: 'wechat',
      title: t('checkout.payment.wechat.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.wechat.subtitle')}`,
      description: t('checkout.payment.wechat.stepperDescription'),
      points: [t('checkout.payment.wechat.points.recipient'), t('checkout.payment.wechat.points.scan')],
      provider: 'wechat',
    }, ['wechat', 'wechatpay', 'wechat_pay']),
    withPaymentAvailability({
      id: 'bank',
      title: t('checkout.payment.bank.title'),
      subtitle: `${priceText} · ${t('checkout.payment.bank.subtitle')}`,
      description: t('checkout.payment.bank.stepperDescription'),
      points: [t('checkout.payment.bank.points.reference')],
    }, ['bank']),
    withPaymentAvailability({
      id: 'worldfirst',
      title: t('checkout.payment.worldfirst.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.worldfirst.subtitle')}`,
      description: t('checkout.payment.worldfirst.stepperDescription'),
      points: [t('checkout.payment.worldfirst.points.b2b')],
    }, ['worldfirst']),
  ]
})

const selectedPaymentOption = computed(() =>
  stepperOptions.value.find(option => option.id === activePaymentTab.value),
)
const selectedPaymentAvailable = computed(() => selectedPaymentOption.value?.available !== false)
const mobilePaymentTitle = computed(() => paymentCopy.value[activePaymentTab.value].title)
const mobilePaymentDescription = computed(() => paymentCopy.value[activePaymentTab.value].description)
const paymentCtaLabel = computed(() => paymentCopy.value[activePaymentTab.value].cta)
const desktopCtaDescription = computed(() => paymentCopy.value[activePaymentTab.value].description)
const stripePaymentConfirmLabel = computed(() => t('checkout.payment.stripe.confirm', 'Confirm payment'))
const stripePaymentConfirmingLabel = computed(() => t('checkout.payment.stripe.confirming', 'Confirming...'))
const appliedCouponDisplayPayload = computed(() => checkoutQuote.value?.coupon_code || '')

const shippingOptionLabel = (option: ShippingQuoteResult['selected_option'] | null | undefined) => {
  if (!option) return ''
  const serviceLabel = option.service_name
    ? option.service_code ? `${option.service_name} (${option.service_code})` : option.service_name
    : option.service_code || ''
  return [option.carrier_name, option.route_name && option.route_name !== option.service_name ? option.route_name : '', serviceLabel]
    .filter(Boolean)
    .join(' / ')
}

const shippingValidation = computed(() => validateShipping(form.value.country, form.value.zip))
const shippingState = computed<'select' | 'available' | 'unavailable' | 'checking'>(() => {
  if (!form.value.country) return 'select'
  if (checkoutQuoteError.value) return 'unavailable'
  if (!shippingValidation.value.isShippable) return 'unavailable'
  if (isFetchingCheckoutQuote.value || !checkoutQuote.value) return 'checking'
  return 'available'
})

const stepperOrderSummary = computed(() => {
  const localTotals = typedPriceBreakdown.value
  const quote = checkoutQuote.value
  const selectedOptionLabel = shippingOptionLabel(quote?.shipping_quote?.selected_option)
  const shippingQuoteLabels = Array.from(new Set(
    quote?.shipping_quote?.items?.map(item => item.template_name).filter(Boolean) || [],
  ))
  const shippingLabel = selectedOptionLabel || shippingQuoteLabels.join(', ') || shippingValidation.value.matchedRule?.service_label

  return {
    items: cartItems.value.map(item => ({
      id: item.id ?? item.sku ?? item.title ?? '',
      title: item.title ?? t('checkout.order.itemFallback'),
      quantity: item.quantity ?? 1,
      price: item.price ?? 0,
      thumbnail: item.thumbnail ?? null,
    })),
    totals: {
      subtotal: quote?.subtotal_amount ?? localTotals.subtotal ?? 0,
      shipping: quote ? quote.shipping_fee : null,
      shippingLabel,
      shippingState: shippingState.value,
      tax: quote?.tax_amount ?? localTotals.tax ?? 0,
      pointsDiscount: quote?.points_discount ?? localTotals.pointsDiscount ?? 0,
      couponDiscount: quote?.coupon_discount ?? localTotals.couponDiscount ?? 0,
      giftCardDiscount: localTotals.giftCardDiscount ?? 0,
      total: quote?.total_amount ?? localTotals.total ?? 0,
    },
  }
})

const normalizedShippingValidation = computed(() => ({
  isShippable: Boolean(shippingValidation.value.isShippable),
  reason: checkoutQuoteError.value || shippingValidation.value.reason,
  matchedRule: shippingValidation.value.matchedRule
    ? {
        service_label: shippingValidation.value.matchedRule.service_label,
        free_over: shippingValidation.value.matchedRule.free_over ?? undefined,
      }
    : undefined,
}))

const shippableCountryCodes = computed(() => getShippableCountries())
const shippableCountries = computed(() => COUNTRIES.filter(country => shippableCountryCodes.value.includes(country.code)))
const nonShippableCountries = computed(() => COUNTRIES.filter(country => !shippableCountryCodes.value.includes(country.code)))
const filteredShippableCountries = computed(() => filterCountries(shippableCountries.value, countrySearch.value))
const filteredNonShippableCountries = computed(() => filterCountries(nonShippableCountries.value, countrySearch.value))
const estimatedDelivery = computed(() =>
  shippingValidation.value.isShippable ? getEstimatedDeliveryText(shippingValidation.value.matchedRule) : null,
)
const zipFormatHint = computed(() => form.value.country ? getZipFormatHint(form.value.country) : null)
const zipPlaceholder = computed(() => zipFormatHint.value?.placeholder || t('checkout.stepper.shipping.zipPlaceholder'))
const zipHint = computed(() => zipFormatHint.value?.hint || '')

function filterCountries(countries: Array<{ code: string; name: string }>, term: string) {
  const normalized = term.trim().toLowerCase()
  if (!normalized) return countries
  return countries.filter(country =>
    country.name.toLowerCase().includes(normalized) || country.code.toLowerCase().includes(normalized),
  )
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
    email: String(auth.user.value?.email || '').trim() || 'customer@example.com',
  }
}

const isFormValid = computed(() =>
  Boolean(
    form.value.country &&
    shippingValidation.value.isShippable &&
    form.value.name.trim() &&
    form.value.phone.trim() &&
    form.value.address.trim() &&
    form.value.city.trim(),
  ),
)

const fetchCheckoutQuote = async (showError = false) => {
  if (!isCheckoutOpen.value || !cartItems.value.length || !form.value.country) {
    checkoutQuote.value = null
    checkoutQuoteError.value = null
    return false
  }

  isFetchingCheckoutQuote.value = true
  checkoutQuoteError.value = null
  try {
    const response = await auth.request<ApiResponse<CheckoutQuote>>('/checkout/quote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({
        shipping_address: buildShippingAddressPayload(),
        coupon_code: couponCode.value.trim(),
        points_to_use: selectedPointsToUse.value,
      }),
    })
    const quote = unwrapApiData<CheckoutQuote>(response)
    if (!quote) throw new Error(t('checkout.modal.messages.invalidQuote'))
    checkoutQuote.value = quote
    return true
  } catch (error) {
    const message = error instanceof Error ? error.message : t('checkout.modal.messages.unableRefreshQuote')
    checkoutQuote.value = null
    checkoutQuoteError.value = message
    if (showError) checkoutError.value = message
    return false
  } finally {
    isFetchingCheckoutQuote.value = false
  }
}

const scheduleCheckoutQuoteRefresh = () => {
  if (checkoutQuoteTimer) clearTimeout(checkoutQuoteTimer)
  checkoutQuoteTimer = setTimeout(() => {
    void fetchCheckoutQuote(false)
  }, 300)
}

const requireAuthenticatedUser = async () => {
  const session = await auth.ensureSession()
  if (session) return true
  showAuthModal.value = true
  checkoutError.value = t('checkout.modal.messages.loginRequired', 'Please sign in before checkout.')
  return false
}

const paymentMethodForOrder = () => activePaymentTab.value === 'card' ? 'card' : activePaymentTab.value
const isStripeMethod = () => activePaymentTab.value === 'card' || activePaymentTab.value === 'stripe'

const createLocalOrder = async (): Promise<PublicOrderResponse> => {
  const orderPayload = {
    items: cartItems.value.map(item => ({
      product_id: Number(item.product_id || item.id || 0),
      variant_id: item.variant_id || null,
      quantity: Math.max(1, Number(item.quantity || 1)),
    })),
    shipping_address: buildShippingAddressPayload(),
    payment_method: paymentMethodForOrder(),
    shipping_method: 'standard',
    coupon_code: couponCode.value.trim(),
    points_to_use: selectedPointsToUse.value,
    client_risk: {
      billing_country: form.value.country.trim().toUpperCase(),
      timezone: Intl.DateTimeFormat().resolvedOptions().timeZone || '',
    },
  }

  const response = await auth.request<ApiResponse<PublicOrderResponse>>('/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify(orderPayload),
  })
  const order = unwrapApiData<PublicOrderResponse>(response)
  if (!order?.order_number) throw new Error(t('checkout.modal.messages.orderFailed'))
  return order
}

const createStripePaymentIntent = async (orderNumber: string) => {
  const response = await auth.request<ApiResponse<StripePaymentIntentResponse>>('/payment/stripe/payment-intents', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
    body: JSON.stringify({ order_number: orderNumber }),
  })
  const session = unwrapApiData<StripePaymentIntentResponse>(response)
  const clientSecret = session?.clientSecret || session?.client_secret || ''
  const publishableKey = session?.publishableKey || session?.publishable_key || ''
  if (!clientSecret || !publishableKey) {
    throw new Error('Invalid Stripe payment response')
  }
  stripePaymentSession.value = { clientSecret, publishableKey }
}

const checkoutUrl = (path: string, orderNumber: string) => {
  if (!import.meta.client) return ''
  const target = new URL(localePath(path), window.location.origin)
  target.searchParams.set('order_number', orderNumber)
  return target.toString()
}

const startPayPalPayment = async (orderNumber: string) => {
  const session = await createPayPalOrder({
    orderNumber,
    returnUrl: checkoutUrl('/checkout/paypal/return', orderNumber),
    cancelUrl: checkoutUrl('/checkout/paypal/cancel', orderNumber),
  })
  redirectToPayPal(session)
}

const startAlipayPayment = async (orderNumber: string) => {
  const session = await createAlipayOrder({
    orderNumber,
    returnUrl: checkoutUrl('/checkout/alipay/return', orderNumber),
    cancelUrl: '',
  })
  redirectToAlipay(session)
}

const weChatSessionStorageKey = (orderNumber: string) => `checkout:wechat:${orderNumber}`

const startWeChatPayment = async (orderNumber: string) => {
  const session: WeChatPaymentSession = await createWeChatOrder({ orderNumber })
  if (import.meta.client) {
    window.sessionStorage.setItem(weChatSessionStorageKey(orderNumber), JSON.stringify(session))
    const target = new URL(localePath('/checkout/wechat/pay'), window.location.origin)
    target.searchParams.set('order_number', orderNumber)
    window.location.assign(target.toString())
  }
}

const completeOfflineOrder = () => {
  clearCart()
  closeCheckout()
  resetCheckoutState()
  alert(t('checkout.modal.messages.orderSuccess'))
}

const resetCheckoutState = () => {
  currentStepperStep.value = 1
  stripePaymentSession.value = null
  checkoutQuote.value = null
  checkoutQuoteError.value = null
  checkoutError.value = ''
  couponCode.value = ''
}

const handleSubmit = async () => {
  if (isSubmitting.value) return
  checkoutError.value = ''

  if (!selectedPaymentAvailable.value) {
    checkoutError.value = selectedPaymentOption.value?.unavailableReason || t('checkout.modal.messages.paymentUnavailable', 'This payment method is temporarily unavailable.')
    return
  }
  if (!cartItems.value.length) {
    checkoutError.value = t('checkout.modal.messages.emptyCart', 'Your cart is empty.')
    return
  }
  if (!(await requireAuthenticatedUser())) return
  if (!isFormValid.value) {
    checkoutError.value = t('checkout.modal.messages.completeShipping')
    return
  }

  isSubmitting.value = true
  try {
    const quoteReady = await fetchCheckoutQuote(true)
    if (!quoteReady) return
    const order = await createLocalOrder()

    if (activePaymentTab.value === 'paypal') {
      await startPayPalPayment(order.order_number)
      return
    }
    if (activePaymentTab.value === 'alipay') {
      await startAlipayPayment(order.order_number)
      return
    }
    if (activePaymentTab.value === 'wechat') {
      await startWeChatPayment(order.order_number)
      return
    }
    if (isStripeMethod()) {
      await createStripePaymentIntent(order.order_number)
      return
    }

    completeOfflineOrder()
  } catch (error) {
    checkoutError.value = error instanceof Error ? error.message : t('checkout.modal.messages.orderFailed')
  } finally {
    isSubmitting.value = false
  }
}

const handleStripeConfirmed = (result: StripeConfirmationResult) => {
  stripePaymentSession.value = null
  if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
    clearCart()
    closeCheckout()
    resetCheckoutState()
    alert(t('checkout.modal.messages.orderSuccess'))
    return
  }
  checkoutError.value = t('checkout.modal.messages.paymentPending', 'Payment is not complete yet. Please check your order status later.')
}

const handleStripeError = (message: string) => {
  checkoutError.value = message
}

const handleStepperSelect = (tab: string) => {
  activePaymentTab.value = tab as PaymentTab
  currentStepperStep.value = 1
  stripePaymentSession.value = null
}
const handleStepperStepChange = (step: StepperStep) => { currentStepperStep.value = step }
const handleStepperShippingField = ({ field, value }: { field: ShippingField; value: string }) => { form.value[field] = value }
const handleStepperCountrySearch = (value: string) => { countrySearch.value = value }
const handleStepperCouponInput = (value: string) => { couponCode.value = value }
const handleStepperTogglePoints = (value: boolean) => { calculation.usePointsDiscount.value = value }
const handleStepperPointsInput = (value: number) => { calculation.pointsToUse.value = value }

const handleApplyCoupon = async () => {
  if (!couponCode.value.trim()) return
  isApplyingCoupon.value = true
  const success = await fetchCheckoutQuote(true)
  isApplyingCoupon.value = false
  if (success) alert(t('checkout.modal.messages.couponApplied'))
}

const handleOpenShippingChat = () => {
  openChat({ showAgentList: true })
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}
const openContactSupport = () => { if (import.meta.client) window.open(localePath('/company/contact'), '_blank') }
const openFreightForwarder = () => { if (import.meta.client) window.open('/help/freight-forwarder', '_blank') }
const saveCartForLater = () => { alert(t('checkout.modal.messages.cartSaved')) }
const handleAuthSuccess = () => {
  showAuthModal.value = false
  checkoutError.value = ''
  void fetchCheckoutQuote(false)
}

onMounted(async () => {
  calculation.initialize()
  await loadShippingTemplates()
})

watch(isCheckoutOpen, (open) => {
  if (!open) {
    resetCheckoutState()
    return
  }
  void requireAuthenticatedUser()
  void loadPaymentMethods(form.value.country || undefined)
  scheduleCheckoutQuoteRefresh()
})

watch(
  () => [
    isCheckoutOpen.value,
    cartItems.value.map(item => `${item.id}:${item.quantity}`).join('|'),
    form.value.country,
    form.value.city,
    form.value.zip,
    couponCode.value.trim(),
    selectedPointsToUse.value,
  ],
  () => {
    if (isCheckoutOpen.value) {
      void loadPaymentMethods(form.value.country || undefined)
      scheduleCheckoutQuoteRefresh()
    }
  },
)

onUnmounted(() => {
  if (checkoutQuoteTimer) clearTimeout(checkoutQuoteTimer)
})
</script>

<style scoped>
.checkout-modal-shell {
  height: min(95vh, calc(100vh - 16px));
  max-height: min(95vh, calc(100vh - 16px));
}

.checkout-modal-ssl-banner {
  border: 1px solid rgba(148, 255, 223, 0.35);
  border-radius: 9999px;
  background: linear-gradient(135deg, rgba(54, 213, 149, 0.22), rgba(59, 130, 246, 0.16));
  padding: 0.2rem 1.1rem;
}

@supports (height: 100dvh) {
  .checkout-modal-shell {
    height: min(95dvh, calc(100dvh - 16px));
    max-height: min(95dvh, calc(100dvh - 16px));
  }
}

@media (min-width: 768px) {
  .checkout-modal-shell {
    height: min(780px, 95vh);
  }
}

.fade-enter-active,
.fade-leave-active,
.scale-enter-active,
.scale-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to,
.scale-enter-from,
.scale-leave-to {
  opacity: 0;
}

.scale-enter-from,
.scale-leave-to {
  transform: scale(0.98);
}
</style>
