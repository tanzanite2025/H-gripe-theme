<template>
  <teleport to="body">
    <!-- 遮罩�?-->
    <transition name="fade">
      <div
        v-if="isCheckoutOpen"
        class="fixed inset-0 z-[9998] flex items-center justify-center p-0 md:p-4"
        @click.self="closeCheckout"
        @save-cart="saveCartForLater"
      >
        <!-- 半透明背景遮罩 -->
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm"></div>
        <!-- 结账弹窗 -->
        <transition name="scale">
          <div
            v-if="isCheckoutOpen"
            class="relative flex flex-col w-full max-w-[1400px] overflow-hidden rounded-2xl bg-[radial-gradient(circle_at_top_left,rgba(15,23,42,0.98),rgba(0,0,0,1))] backdrop-blur-xl border-2 border-[#6b73ff]/40 shadow-[0_0_30px_rgba(107,115,255,0.6)] checkout-modal-shell h-[95vh] md:h-[780px] max-h-[95vh]"
          >
            <!-- 头部：再次压缩高度，为下方内容留出更多空�?-->
            <div class="relative flex items-center justify-center px-2 md:px-6 py-3.5 border-b border-transparent bg-[rgba(4,7,20,0.92)] backdrop-blur-sm shadow-[0_10px_30px_rgba(0,0,0,0.5)] md:border-transparent md:bg-transparent md:backdrop-blur-0 md:shadow-none">
              <div class="flex items-center gap-2 md:gap-3">
                <h2 class="sr-only">{{ t('checkout.modal.title') }}</h2>
                <div class="flex items-center gap-2 overflow-x-auto sm:overflow-visible">
                  <button
                    type="button"
                    @click="openCartFromCheckout"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                  >
                    {{ t('checkout.modal.actions.viewCart') }}
                  </button>
                  <button
                    type="button"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                    @click="handleOpenShippingChat"
                  >
                    {{ t('checkout.modal.actions.livechat') }}
                  </button>
                  <button
                    type="button"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                  >
                    {{ t('checkout.modal.actions.email') }}
                  </button>
                </div>
              </div>
              <button
                @click="closeCheckout"
                class="absolute right-2 md:right-4 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                :aria-label="t('checkout.modal.closeAriaLabel')"
              >
                <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- 顶部提示：保留说明文字但移除标题 -->
            <div class="px-2 md:px-6 pt-0.5 pb-0 md:pb-1">
              <div class="checkout-modal-ssl-banner flex items-center justify-center gap-1.5 text-[11px] md:text-xs text-emerald-100 max-w-[420px] mx-auto text-center leading-tight">
                <img
                  src="/checkout/secured_ssl-preview.png"
                  :alt="t('checkout.modal.sslAlt')"
                  class="h-12 w-auto md:h-16 -my-2 md:-my-3"
                  loading="lazy"
                  decoding="async"
                />
                <p class="leading-tight">
                  {{ t('checkout.modal.sslNote') }}
                </p>
              </div>
            </div>

            <div class="flex-1 overflow-y-auto px-0 md:px-6 pb-1 md:pb-4">
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
                :currency="'USD'"
                :show-shipping-form="activePaymentTab !== 'bank' && activePaymentTab !== 'worldfirst'"
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
              />
            </div>
          </div>
        </transition>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, type ComputedRef } from 'vue'
import { useCart, useI18n } from '#imports'
import type { useCartCalculation } from '~/composables/useCartCalculation'
import ChatStartButton from '~/components/ChatStartButton.vue'
import { COUNTRIES } from '~/data/countries'
import { useShippingValidation } from '~/composables/useShippingValidation'
import type { ShippingQuoteResult } from '~/composables/useShippingQuote'
import { useChatWidget } from '~/composables/useChatWidget'
import { useAuth } from '~/composables/useAuth'

type CartCalculation = ReturnType<typeof useCartCalculation>
type CartPriceBreakdownBase = ReturnType<CartCalculation['calculateTotal']>
type CartPriceBreakdown = CartPriceBreakdownBase & {
  shippingLabel?: string
  shippingState?: 'select' | 'available' | 'unavailable' | 'checking'
  pointsDiscount?: number
  couponDiscount?: number
  giftCardDiscount?: number
}

type ApiResponse<T> = T | { data?: T }

type CheckoutQuote = {
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

const unwrapApiData = <T,>(payload: ApiResponse<T> | null | undefined): T | null => {
  if (!payload || typeof payload !== 'object') {
    return (payload as T) || null
  }
  if ('data' in payload && payload.data !== undefined) {
    return payload.data as T
  }
  return payload as T
}

const {
  cartItems,
  isCheckoutOpen,
  priceBreakdown,
  closeCheckout,
  backToCart,
  formatPrice,
  clearCart,
  calculation,
  openCartFromCheckout,
} = useCart()

const auth = useAuth()
const { t } = useI18n()
const typedPriceBreakdown = priceBreakdown as ComputedRef<CartPriceBreakdown>
const checkoutQuote = ref<CheckoutQuote | null>(null)
const isFetchingCheckoutQuote = ref(false)
const checkoutQuoteError = ref<string | null>(null)

const shippingOptionLabel = (option: ShippingQuoteResult['selected_option'] | null | undefined) => {
  if (!option) return ''
  const carrierName = option.carrier_name?.trim()
  const routeName = option.route_name?.trim()
  const serviceName = option.service_name?.trim()
  const serviceCode = option.service_code?.trim()
  const serviceLabel = serviceName
    ? serviceCode ? `${serviceName} (${serviceCode})` : serviceName
    : serviceCode || ''
  return [carrierName, routeName && routeName !== serviceName ? routeName : '', serviceLabel]
    .filter(Boolean)
    .join(' / ')
}

const shippingState = computed<CartPriceBreakdown['shippingState']>(() => {
  if (!form.value.country) return 'select'
  if (checkoutQuoteError.value) return 'unavailable'
  if (!shippingValidation.value?.isShippable) return 'unavailable'
  if (isFetchingCheckoutQuote.value || !checkoutQuote.value) return 'checking'
  return 'available'
})

const stepperOrderSummary = computed(() => {
  if (!cartItems.value) throw new Error("[CRITICAL] cartItems missing")
  const items = cartItems.value.map((item: any) => ({
    id: item.id ?? item.sku ?? item.title ?? '',
    title: item.title ?? t('checkout.order.itemFallback'),
    quantity: item.quantity ?? 1,
    price: item.price ?? 0,
    thumbnail: item.thumbnail ?? null,
  }))

  const localTotals = typedPriceBreakdown.value || ({} as CartPriceBreakdown)
  const quote = checkoutQuote.value
  const matchedRule = shippingValidation.value?.matchedRule as { service_label?: string } | undefined
  const selectedOptionLabel = shippingOptionLabel(quote?.shipping_quote?.selected_option)
  const shippingQuoteLabels = Array.from(new Set(
    quote?.shipping_quote?.items
      ?.map(item => item.template_name)
      .filter(Boolean) || []
  ))
  const shippingLabel = selectedOptionLabel || (shippingQuoteLabels.length ? shippingQuoteLabels.join(', ') : matchedRule?.service_label)

  return {
    items,
    totals: {
      subtotal: quote?.subtotal_amount ?? localTotals.subtotal ?? 0,
      shipping: quote ? quote.shipping_fee : null,
      shippingLabel,
      shippingState: shippingState.value,
      tax: quote?.tax_amount ?? localTotals.tax ?? 0,
      pointsDiscount: quote?.points_discount ?? localTotals.pointsDiscount ?? 0,
      couponDiscount: quote?.coupon_discount ?? localTotals.couponDiscount ?? 0,
      giftCardDiscount: 0,
      total: quote?.total_amount ?? localTotals.total ?? 0,
    },
  }
})

// 配送验�?
const {
  loadShippingTemplates,
  validateShipping,
  getShippableCountries,
  getEstimatedDeliveryText,
  getZipFormatHint,
} = useShippingValidation()

// 初始化计算系统和配送模�?
onMounted(async () => {
  calculation.initialize()
  await loadShippingTemplates()
})

// 聊天窗口（WhatsAppChatModal�?
const { openChat } = useChatWidget()

const handleOpenShippingChat = () => {
  openChat({ showAgentList: true })
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

type PaymentTab = 'card' | 'paypal' | 'alipay' | 'stripe' | 'bank' | 'worldfirst'
type StepperStep = 1 | 2 | 3

const activePaymentTab = ref<PaymentTab>('card')
const currentStepperStep = ref<StepperStep>(1)

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

type StepperOption = {
  id: PaymentTab
  title: string
  subtitle: string
  description: string
  points?: string[]
}

type ShippingField = 'country' | 'name' | 'phone' | 'address' | 'city' | 'zip' | 'paymentMethod' | 'notes'

const stepperOptions = computed<StepperOption[]>(() => {
  const priceText = formatPrice(checkoutQuote.value?.total_amount ?? priceBreakdown.value.total)
  return [
    {
      id: 'card',
      title: t('checkout.payment.card.title'),
      subtitle: `${priceText} · ${t('checkout.payment.card.subtitle')}`,
      description: t('checkout.payment.card.stepperDescription'),
      points: [
        t('checkout.payment.card.points.shipping'),
        t('checkout.payment.card.points.immediate'),
      ],
    },
    {
      id: 'alipay',
      title: t('checkout.payment.alipay.title'),
      subtitle: `${priceText} · ${t('checkout.payment.alipay.subtitle')}`,
      description: t('checkout.payment.alipay.stepperDescription'),
      points: [
        t('checkout.payment.alipay.points.recipient'),
        t('checkout.payment.alipay.points.wallets'),
      ],
    },
    {
      id: 'paypal',
      title: t('checkout.payment.paypal.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.paypal.subtitle')}`,
      description: t('checkout.payment.paypal.stepperDescription'),
      points: [
        t('checkout.payment.paypal.points.country'),
      ],
    },
    {
      id: 'stripe',
      title: t('checkout.payment.stripe.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.stripe.subtitle')}`,
      description: t('checkout.payment.stripe.stepperDescription'),
      points: [
        t('checkout.payment.stripe.points.sca'),
      ],
    },
    {
      id: 'bank',
      title: t('checkout.payment.bank.title'),
      subtitle: `${priceText} · ${t('checkout.payment.bank.subtitle')}`,
      description: t('checkout.payment.bank.stepperDescription'),
      points: [
        t('checkout.payment.bank.points.reference'),
      ],
    },
    {
      id: 'worldfirst',
      title: t('checkout.payment.worldfirst.optionTitle'),
      subtitle: `${priceText} · ${t('checkout.payment.worldfirst.subtitle')}`,
      description: t('checkout.payment.worldfirst.stepperDescription'),
      points: [
        t('checkout.payment.worldfirst.points.b2b'),
      ],
    },
  ]
})

const setActivePaymentTab = (tab: PaymentTab) => {
  activePaymentTab.value = tab
}

const handleStepperSelect = (tab: string) => {
  setActivePaymentTab(tab as PaymentTab)
  currentStepperStep.value = 1
}

const handleStepperStepChange = (step: StepperStep) => {
  currentStepperStep.value = step
}

const handleStepperShippingField = ({ field, value }: { field: ShippingField; value: string }) => {
  form.value[field] = value
}

const handleStepperCountrySearch = (value: string) => {
  countrySearch.value = value
}

const mobilePaymentTitle = computed(() => paymentCopy.value[activePaymentTab.value].title)
const mobilePaymentDescription = computed(() => paymentCopy.value[activePaymentTab.value].description)
const paymentCtaLabel = computed(() => paymentCopy.value[activePaymentTab.value].cta)
const desktopCtaDescription = computed(() => paymentCopy.value[activePaymentTab.value].description)

const handleStepperCouponInput = (value: string) => {
  couponCode.value = value
}

const handleStepperTogglePoints = (value: boolean) => {
  if (calculation?.usePointsDiscount) {
    calculation.usePointsDiscount.value = value
  }
}

const handleStepperPointsInput = (value: number) => {
  if (calculation?.pointsToUse) {
    calculation.pointsToUse.value = value
  }
}

// 表单数据
const form = ref({
  country: '',
  name: '',
  phone: '',
  address: '',
  city: '',
  zip: '',
  paymentMethod: 'credit_card',
  notes: '',
})

const isSubmitting = ref(false)
const couponCode = ref('')
const isApplyingCoupon = ref(false)
const appliedCouponDisplayPayload = computed(() => checkoutQuote.value?.coupon_code || calculation.appliedCoupon.value)
let checkoutQuoteTimer: ReturnType<typeof setTimeout> | null = null

const selectedPointsToUse = computed(() => {
  return calculation?.usePointsDiscount?.value ? (calculation.pointsToUse.value || 0) : 0
})

const buildShippingAddressPayload = () => {
  const nameParts = form.value.name.trim().split(' ').filter(Boolean)
  const firstName = nameParts[0] || 'Guest'
  const lastName = nameParts.length > 1 ? nameParts.slice(1).join(' ') : 'User'

  return {
    first_name: firstName,
    last_name: lastName,
    address1: form.value.address || '',
    city: form.value.city || '',
    postal_code: form.value.zip || '',
    country: form.value.country || '',
    phone: form.value.phone || '',
    email: auth.user.value?.email || 'guest@example.com'
  }
}

const fetchCheckoutQuote = async (showError = false) => {
  if (!isCheckoutOpen.value || !cartItems.value?.length) {
    checkoutQuote.value = null
    checkoutQuoteError.value = null
    return false
  }

  if (!form.value.country) {
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
      })
    })
    const quote = unwrapApiData<CheckoutQuote>(response)
    if (!quote) {
      throw new Error(t('checkout.modal.messages.invalidQuote'))
    }
    checkoutQuote.value = quote
    checkoutQuoteError.value = null
    return true
  } catch (error) {
    const message = error instanceof Error ? error.message : t('checkout.modal.messages.unableRefreshQuote')
    checkoutQuote.value = null
    checkoutQuoteError.value = message
    if (showError) {
      alert(message)
    } else {
      console.warn('Checkout quote refresh failed:', message)
    }
    return false
  } finally {
    isFetchingCheckoutQuote.value = false
  }
}

const scheduleCheckoutQuoteRefresh = () => {
  if (checkoutQuoteTimer) {
    clearTimeout(checkoutQuoteTimer)
  }
  checkoutQuoteTimer = setTimeout(() => {
    void fetchCheckoutQuote(false)
  }, 300)
}

// 配送验证结�?
const shippingValidation = computed(() => {
  return validateShipping(form.value.country, form.value.zip)
})

const normalizedShippingValidation = computed(() => {
  const val = shippingValidation.value
  if (!val) return null
  const rule = val.matchedRule
  const matchedRule =
    rule && typeof rule === 'object'
      ? {
          service_label: (rule as any).service_label,
          free_over: (rule as any).free_over,
        }
      : undefined
  return {
    isShippable: Boolean(val.isShippable),
    reason: checkoutQuoteError.value || val.reason,
    matchedRule,
  }
})

// 可配送国家列�?
const shippableCountryCodes = computed(() => getShippableCountries())

const shippableCountries = computed(() => {
  return COUNTRIES.filter(c => shippableCountryCodes.value.includes(c.code))
})

const nonShippableCountries = computed(() => {
  return COUNTRIES.filter(c => !shippableCountryCodes.value.includes(c.code))
})

const countrySearch = ref('')

const filteredShippableCountries = computed(() => {
  const term = countrySearch.value.trim().toLowerCase()
  if (!term) return shippableCountries.value
  return shippableCountries.value.filter(country => {
    const name = country.name.toLowerCase()
    const code = country.code.toLowerCase()
    return name.includes(term) || code.includes(term)
  })
})

const filteredNonShippableCountries = computed(() => {
  const term = countrySearch.value.trim().toLowerCase()
  if (!term) return nonShippableCountries.value
  return nonShippableCountries.value.filter(country => {
    const name = country.name.toLowerCase()
    const code = country.code.toLowerCase()
    return name.includes(term) || code.includes(term)
  })
})

// 预计送达时间
const estimatedDelivery = computed(() => {
  if (!shippingValidation.value.isShippable) return null
  return getEstimatedDeliveryText(shippingValidation.value.matchedRule)
})

// 邮编格式提示
const zipFormatHint = computed(() => {
  if (!form.value.country) return null
  return getZipFormatHint(form.value.country)
})

const zipPlaceholder = computed(() => {
  return zipFormatHint.value?.placeholder || t('checkout.stepper.shipping.zipPlaceholder')
})

const zipHint = computed(() => {
  return zipFormatHint.value?.hint || ''
})

// 联系客服
const openContactSupport = () => {
  window.open('/company/contact', '_blank')
}

// 货运代理服务
const openFreightForwarder = () => {
  window.open('/help/freight-forwarder', '_blank')
}

// 保存购物车稍后再�?
const saveCartForLater = () => {
  // 可以保存�?localStorage 或用户账�?
  alert(t('checkout.modal.messages.cartSaved'))
}

watch(isCheckoutOpen, value => {
  if (!value) {
    currentStepperStep.value = 1
    checkoutQuote.value = null
  } else {
    scheduleCheckoutQuoteRefresh()
  }
})

watch(
  () => [
    isCheckoutOpen.value,
    cartItems.value?.map((item: any) => `${item.id}:${item.quantity}`).join('|') || '',
    form.value.country,
    form.value.city,
    form.value.zip,
    couponCode.value.trim(),
    selectedPointsToUse.value,
  ],
  () => {
    scheduleCheckoutQuoteRefresh()
  }
)

onUnmounted(() => {
  if (checkoutQuoteTimer) {
    clearTimeout(checkoutQuoteTimer)
  }
})

// 基于地区的运费计�?
const regionShippingFee = computed(() => {
  if (!form.value.country || !shippingValidation.value.isShippable) {
    return 0
  }
  
  const rule = shippingValidation.value.matchedRule
  if (!rule) return 0
  
  // 检查免运费门槛
  if (rule.free_over && priceBreakdown.value.subtotal >= rule.free_over) {
    return 0
  }
  
  return rule.fee || 0
})

// 应用优惠�?
const handleApplyCoupon = async () => {
  if (!couponCode.value) return
  
  isApplyingCoupon.value = true
  const success = await fetchCheckoutQuote(true)
  isApplyingCoupon.value = false
  
  if (success) {
    alert(t('checkout.modal.messages.couponApplied'))
  }
}

// 表单验证
const isFormValid = computed(() => {
  return (
    form.value.country !== '' &&
    shippingValidation.value.isShippable &&
    form.value.name.trim() !== '' &&
    form.value.phone.trim() !== '' &&
    form.value.address.trim() !== '' &&
    form.value.city.trim() !== '' &&
    form.value.paymentMethod !== ''
  )
})

// 提交订单
const handleSubmit = async () => {
  if (isSubmitting.value) return

  const tab = activePaymentTab.value

  // 根据支付方式分别校验
  if (tab === 'card' || tab === 'alipay' || tab === 'stripe') {
    // 在线卡支�?/ 本地钱包：需要完整可配送地址 + 联系方式
    if (!isFormValid.value) {
      alert(t('checkout.modal.messages.completeShipping'))
      return
    }
  } else if (tab === 'paypal') {
    // PayPal：至少需要选择一个在可配送列表中的国家，用于判断能否发货
    if (!form.value.country || !shippableCountryCodes.value.includes(form.value.country)) {
      alert(t('checkout.modal.messages.selectPaypalCountry'))
      return
    }
  }

  // Bank transfer / WorldFirst：不强制前置地址校验，由后续人工沟通确�?

  isSubmitting.value = true

  try {
    const quoteReady = await fetchCheckoutQuote(true)
    if (!quoteReady) {
      return
    }

    // 调用 Go 后端订单 API
    
    // 适配后端 CreateOrderRequest 数据结构
    const orderPayload = {
      items: cartItems.value.map((item: any) => ({
        product_id: item.product_id || item.id || 0,
        variant_id: item.variant_id || null,
        quantity: item.quantity || 1
      })),
      shipping_address: buildShippingAddressPayload(),
      payment_method: form.value.paymentMethod || 'credit_card',
      shipping_method: 'standard',
      coupon_code: couponCode.value || '',
      points_to_use: selectedPointsToUse.value
    }

    await auth.request('/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(orderPayload)
    })

    // 成功后清空购物车
    clearCart()
    closeCheckout()

    // 显示成功消息
    alert(t('checkout.modal.messages.orderSuccess'))

    // 重置表单
    form.value = {
      country: '',
      name: '',
      phone: '',
      address: '',
      city: '',
      zip: '',
      paymentMethod: 'credit_card',
      notes: '',
    }
  } catch (error) {
    console.error('Order submission failed:', error)
    alert(t('checkout.modal.messages.orderFailed'))
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.checkout-modal-shell {
  height: min(95vh, calc(100vh - 16px));
  max-height: min(95vh, calc(100vh - 16px));
}

.checkout-modal-ssl-banner {
  background: linear-gradient(135deg, rgba(54, 213, 149, 0.25), rgba(59, 130, 246, 0.2));
  border: 1px solid rgba(148, 255, 223, 0.35);
  border-radius: 9999px;
  padding: 0.2rem 1.1rem;
  box-shadow:
    0 6px 24px rgba(59, 130, 246, 0.2),
    inset 0 0 12px rgba(13, 148, 136, 0.25);
}

@supports (height: 100dvh) {
  .checkout-modal-shell {
    height: min(95dvh, calc(100dvh - 16px));
    max-height: min(95dvh, calc(100dvh - 16px));
  }
}

@media (min-width: 768px) {
  .checkout-modal-shell {
    height: 780px;
    max-height: 95vh;
  }

  @supports (height: 100dvh) {
    .checkout-modal-shell {
      height: min(780px, 95dvh);
    }
  }
}

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 缩放动画 */
.scale-enter-active,
.scale-leave-active {
  transition: all 0.3s ease;
}

.scale-enter-from,
.scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

/* 自定义滚动条 */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style>
