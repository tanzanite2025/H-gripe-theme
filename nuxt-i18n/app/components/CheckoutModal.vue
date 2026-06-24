<template>
  <teleport to="body">
    <!-- 遮罩层 -->
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
            <!-- 头部：再次压缩高度，为下方内容留出更多空间 -->
            <div class="relative flex items-center justify-center px-2 md:px-6 py-3.5 border-b border-transparent bg-[rgba(4,7,20,0.92)] backdrop-blur-sm shadow-[0_10px_30px_rgba(0,0,0,0.5)] md:border-transparent md:bg-transparent md:backdrop-blur-0 md:shadow-none">
              <div class="flex items-center gap-2 md:gap-3">
                <h2 class="sr-only">Checkout</h2>
                <div class="flex items-center gap-2 overflow-x-auto sm:overflow-visible">
                  <button
                    type="button"
                    @click="openCartFromCheckout"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                  >
                    View cart
                  </button>
                  <button
                    type="button"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                    @click="handleOpenShippingChat"
                  >
                    Livechat
                  </button>
                  <button
                    type="button"
                    class="shrink-0 inline-flex items-center justify-center px-3 py-1.5 rounded-full text-xs font-semibold text-slate-900 bg-white shadow-[4px_6px_16px_rgba(0,0,0,0.45)] hover:bg-white/90 transition-all"
                  >
                    Email
                  </button>
                </div>
              </div>
              <button
                @click="closeCheckout"
                class="absolute right-2 md:right-4 top-1/2 -translate-y-1/2 w-9 h-9 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                aria-label="Close checkout"
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
                  alt="Secure SSL"
                  class="h-12 w-auto md:h-16 -my-2 md:-my-3"
                  loading="lazy"
                  decoding="async"
                />
                <p class="leading-tight">
                  Pages use HTTPS with trusted SSL; we only store the result, not your card details.
                </p>
              </div>
            </div>

            <div class="flex-1 overflow-y-auto px-0 md:px-6 pb-1 md:pb-4">
              <CheckoutStepper
                :initial-step="currentStepperStep"
                :initial-method="activePaymentTab"
                :coupon-input="couponCode"
                :is-applying-coupon="isApplyingCoupon"
                :applied-coupon="calculation.appliedCoupon.value"
                :points-available="calculation.userPoints.value?.available || 0"
                :is-using-points="calculation.usePointsDiscount.value"
                :points-to-use="calculation.pointsToUse.value"
                :max-points-to-use="calculation.userPoints.value?.available || 0"
                :points-hint="'1 point = $0.01, max 50% of order'"
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
import { ref, computed, watch, onMounted, type ComputedRef } from 'vue'
import { useCart } from '#imports'
import type { useCartCalculation } from '~/composables/useCartCalculation'
import ChatStartButton from '~/components/ChatStartButton.vue'
import { COUNTRIES } from '~/data/countries'
import { useShippingValidation } from '~/composables/useShippingValidation'
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

const typedPriceBreakdown = priceBreakdown as ComputedRef<CartPriceBreakdown>

const shippingState = computed<CartPriceBreakdown['shippingState']>(() => {
  if (!form.value.country) return 'select'
  if (!shippingValidation.value?.isShippable) return 'unavailable'
  return 'available'
})

const stepperOrderSummary = computed(() => {
  const items = (cartItems.value || []).map((item: any) => ({
    id: item.id ?? item.sku ?? item.title ?? '',
    title: item.title ?? 'Item',
    quantity: item.quantity ?? 1,
    price: item.price ?? 0,
    thumbnail: item.thumbnail ?? null,
  }))

  const totals = typedPriceBreakdown.value || ({} as CartPriceBreakdown)
  const matchedRule = shippingValidation.value?.matchedRule as { service_label?: string } | undefined

  return {
    items,
    totals: {
      subtotal: totals.subtotal ?? 0,
      shipping: totals.shipping ?? null,
      shippingLabel: matchedRule?.service_label,
      shippingState: shippingState.value,
      tax: totals.tax ?? 0,
      pointsDiscount: totals.pointsDiscount ?? 0,
      couponDiscount: totals.couponDiscount ?? 0,
      giftCardDiscount: 0,
      total: totals.total ?? 0,
    },
  }
})

// 配送验证
const {
  loadShippingTemplates,
  validateShipping,
  getShippableCountries,
  getEstimatedDeliveryText,
  getZipFormatHint,
} = useShippingValidation()

// 初始化计算系统和配送模板
onMounted(async () => {
  calculation.initialize()
  await loadShippingTemplates()
})

// 聊天窗口（WhatsAppChatModal）
const { openChat } = useChatWidget()

const handleOpenShippingChat = () => {
  openChat({ showAgentList: true })
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

// 支付方式列表（暂未直接用于 Tabs，但保留给后续逻辑使用）
const paymentMethods = [
  { id: 'credit_card', name: 'Credit Card', icon: '💳' },
  { id: 'paypal', name: 'PayPal', icon: '🅿️' },
  { id: 'alipay', name: 'Alipay', icon: '💙' },
  { id: 'wechat', name: 'WeChat Pay', icon: '💚' },
]

type PaymentTab = 'card' | 'paypal' | 'alipay' | 'stripe' | 'bank' | 'worldfirst'
type StepperStep = 1 | 2 | 3

const activePaymentTab = ref<PaymentTab>('card')
const currentStepperStep = ref<StepperStep>(1)

const paymentCopy: Record<PaymentTab, { title: string; description: string; cta: string }> = {
  card: {
    title: 'Credit / Debit cards',
    description: 'Card details are entered on a secure payment page from our provider; we only store your order result.',
    cta: 'Continue to secure payment',
  },
  paypal: {
    title: 'Pay with PayPal',
    description: 'You will be redirected to PayPal. We never see or store your PayPal login or card information.',
    cta: 'Continue to PayPal',
  },
  alipay: {
    title: 'Alipay / WeChat / UnionPay',
    description: 'You approve the payment in your local wallet. We still collect shipping info here to prepare dispatch.',
    cta: 'Pay with local wallet',
  },
  stripe: {
    title: 'Pay with Stripe',
    description: 'Stripe handles your card details and any extra verification. We only store the payment status.',
    cta: 'Continue to Stripe',
  },
  bank: {
    title: 'Bank transfer',
    description: 'Place the order first, then send a bank transfer using the reference we show you on the next step.',
    cta: 'Place order for bank transfer',
  },
  worldfirst: {
    title: 'WorldFirst',
    description: 'Send funds to a local WorldFirst account in your currency. We never access your banking credentials.',
    cta: 'Continue with WorldFirst',
  },
}

type StepperOption = {
  id: PaymentTab
  title: string
  subtitle: string
  description: string
}

type ShippingField = 'country' | 'name' | 'phone' | 'address' | 'city' | 'zip' | 'paymentMethod' | 'notes'

const stepperOptions = computed<StepperOption[]>(() => {
  const priceText = `≈ ${formatPrice(priceBreakdown.value.total)}`
  return [
    {
      id: 'card',
      title: 'Credit / Debit cards',
      subtitle: `${priceText} · Secure card page`,
      description:
        'Enter full shipping info and contact number here, then we redirect you to the PCI-compliant card page. We never store your card number or CVC.',
    },
    {
      id: 'alipay',
      title: 'Alipay / WeChat / UnionPay',
      subtitle: `${priceText} · Local wallets`,
      description:
        'Approve payment inside your usual wallet. We still gather shipping details in this tab to prepare dispatch and confirm customs information.',
    },
    {
      id: 'paypal',
      title: 'PayPal',
      subtitle: `${priceText} · Express checkout`,
      description:
        'We rely on the shipping address saved in your PayPal account whenever possible. Choose a supported country first so we can confirm fulfillment.',
    },
    {
      id: 'stripe',
      title: 'Stripe',
      subtitle: `${priceText} · 3D Secure ready`,
      description:
        'Stripe handles card input and any extra verification. We keep the form minimal here and let Stripe manage the sensitive fields.',
    },
    {
      id: 'bank',
      title: 'Bank transfer',
      subtitle: `${priceText} · Manual invoice`,
      description:
        'Place the order, then transfer the funds using the bank instructions shown on the next step. Remember to include your order number as reference.',
    },
    {
      id: 'worldfirst',
      title: 'WorldFirst',
      subtitle: `${priceText} · Cross-border`,
      description:
        'Ideal for international or business orders. You pay into a regional WorldFirst account in your currency; they settle the converted funds to us.',
    },
  ]
})

const setActivePaymentTab = (tab: PaymentTab) => {
  activePaymentTab.value = tab
}

const handleStepperSelect = (tab: PaymentTab) => {
  setActivePaymentTab(tab)
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

const mobilePaymentTitle = computed(() => paymentCopy[activePaymentTab.value].title)
const mobilePaymentDescription = computed(() => paymentCopy[activePaymentTab.value].description)
const paymentCtaLabel = computed(() => paymentCopy[activePaymentTab.value].cta)
const desktopCtaDescription = computed(() => paymentCopy[activePaymentTab.value].description)

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

// 配送验证结果
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
    reason: val.reason,
    matchedRule,
  }
})

// 可配送国家列表
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
  return zipFormatHint.value?.placeholder || 'Zip code'
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

// 保存购物车稍后再买
const saveCartForLater = () => {
  // 可以保存到 localStorage 或用户账户
  alert('Your cart has been saved. We\'ll notify you when shipping becomes available to your region.')
}

watch(isCheckoutOpen, value => {
  if (!value) {
    currentStepperStep.value = 1
  }
})

// 基于地区的运费计算
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

// 应用优惠券
const handleApplyCoupon = async () => {
  if (!couponCode.value) return
  
  isApplyingCoupon.value = true
  const result = await calculation.applyCoupon(couponCode.value)
  isApplyingCoupon.value = false
  
  if (result.success) {
    alert('Coupon applied successfully!')
  } else {
    alert(result.message)
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
    // 在线卡支付 / 本地钱包：需要完整可配送地址 + 联系方式
    if (!isFormValid.value) {
      alert('Please complete your shipping address and contact details so we can confirm shipping to your region before continuing.')
      return
    }
  } else if (tab === 'paypal') {
    // PayPal：至少需要选择一个在可配送列表中的国家，用于判断能否发货
    if (!form.value.country || !shippableCountryCodes.value.includes(form.value.country)) {
      alert('Please select a country/region we can ship to before continuing to PayPal.')
      return
    }
  }

  // Bank transfer / WorldFirst：不强制前置地址校验，由后续人工沟通确认

  isSubmitting.value = true

  try {
    // 调用 Go 后端订单 API
    const auth = useAuth()
    
    // 适配后端 CreateOrderRequest 数据结构
    const nameParts = form.value.name.trim().split(' ')
    const firstName = nameParts[0] || 'Guest'
    const lastName = nameParts.length > 1 ? nameParts.slice(1).join(' ') : 'User'

    const orderPayload = {
      items: cartItems.value.map((item: any) => ({
        product_id: item.id || 0,
        quantity: item.quantity || 1
      })),
      shipping_address: {
        first_name: firstName,
        last_name: lastName,
        address1: form.value.address || 'N/A',
        city: form.value.city || 'N/A',
        postal_code: form.value.zip || '000000',
        country: form.value.country || 'US',
        phone: form.value.phone || '00000000',
        email: auth.user.value?.email || 'guest@example.com'
      },
      payment_method: form.value.paymentMethod || 'credit_card',
      shipping_method: 'standard',
      coupon_code: couponCode.value || ''
    }

    const response = await auth.request('/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(orderPayload)
    })

    // 成功后清空购物车
    clearCart()
    closeCheckout()

    // 显示成功消息
    alert('Order submitted successfully!')

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
    alert('Order submission failed, please try again')
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
