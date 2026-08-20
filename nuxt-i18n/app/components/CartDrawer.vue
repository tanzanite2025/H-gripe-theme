<template>
  <Teleport to="body">
    <!-- 购物车弹窗 -->
    <Transition name="wa-drawer">
      <div
        v-if="isCartOpen"
        class="wa-drawer-mask"
        :class="{ '!items-end': cartVariant === 'lever-bottom' }"
        @click.self="closeCart"
      >
        <!-- Backdrop -->
        <!-- 
             Standard wa-drawer-backdrop is md:hidden. 
             If cartVariant === 'default', we likely want a backdrop on desktop too (modal mode).
             We can add 'md:block' if variant is default.
        -->
        <div
          class="wa-drawer-backdrop"
          :class="{ 'md:block': cartVariant === 'default' }"
        ></div>

        <!-- 弹窗内容 -->
        <div
          class="wa-drawer-shell cart-drawer-shell"
          aria-modal="true"
          role="dialog"
          :aria-label="t('cartDrawer.ariaLabel')"
        >
        <!-- 头部 -->
        <div class="wa-drawer-header relative z-10">
          <h2 class="wa-drawer-title text-base sm:text-xl">
            🛒 {{ t('cartDrawer.title') }} ({{ cartCount }})
          </h2>
          <button
            @click="closeCart"
            class="wa-drawer-close-btn"
            :aria-label="t('cartDrawer.closeAriaLabel')"
          >
            <Icon name="lucide:x" class="h-3.5 w-3.5" />
          </button>
        </div>

        <!-- 购物车内容 -->
        <div v-if="cartItems.length > 0" class="wa-drawer-content cart-drawer-content relative z-10">
          <section class="cart-drawer-cart-section" :aria-label="t('cartDrawer.title')">
          <div class="cart-drawer-items-scroll">
            <div
              v-for="item in cartItems"
              :key="item.id"
              class="cart-drawer-item-card bg-white/[0.06] border border-white rounded-2xl"
            >
              <!-- 商品图片 -->
              <div class="w-20 h-20 flex-shrink-0 bg-white/[0.06] rounded-lg overflow-hidden border border-white">
                <StorefrontImage
                  v-if="item.thumbnail"
                  :src="item.thumbnail"
                  :alt="item.title"
                  class="w-full h-full object-cover"
                  preset="thumbnail"
                />
              <div v-else class="w-full h-full flex items-center justify-center tz-text-muted">
                  <Icon name="lucide:image" class="w-8 h-8" />
                </div>
              </div>

              <!-- 商品信息 -->
              <div class="flex-1 min-w-0">
                <h3 class="text-sm font-medium text-white truncate">
                  {{ item.title }}
                </h3>
                <p class="text-sm font-semibold text-white mt-2">
                  {{ formatPrice(item.price, item.currency) }}
                </p>

                <!-- 数量控制 -->
                <div class="flex items-center gap-2 mt-3">
                  <button
                    @click="decrementQuantity(item.id)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white/[0.18] hover:bg-white/10 transition-colors text-white"
                    :disabled="item.quantity <= 1"
                  >
                    <Icon name="lucide:minus" class="w-4 h-4" />
                  </button>
                  
                  <input
                    type="number"
                    :value="item.quantity"
                    @input="onQuantityInput(item.id, $event)"
                    class="w-12 h-7 text-center border border-white rounded bg-white/[0.06] text-white focus:outline-none focus:ring-2 focus:ring-white/35"
                    min="1"
                  />
                  
                  <button
                    @click="incrementQuantity(item.id)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white hover:bg-white/10 transition-colors text-white"
                  >
                    <Icon name="lucide:plus" class="w-4 h-4" />
                  </button>

                  <button
                    @click="handleAddToWishlist(item)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white/[0.18] hover:bg-white/10 transition-colors text-white"
                    :title="t('cartDrawer.actions.addToWishlist')"
                    :aria-label="t('cartDrawer.actions.addToWishlist')"
                  >
                    <Icon name="lucide:heart" class="w-4 h-4" />
                  </button>

                  <button
                    @click="removeFromCart(item.id)"
                    class="ml-auto text-red-400 hover:text-red-300 text-sm font-medium"
                  >
                    {{ t('cartDrawer.actions.remove') }}
                  </button>
                </div>
              </div>
            </div>
          </div>
          </section>

          <!-- 浏览历史组件 -->
          <div class="cart-drawer-history-section">
            <BrowsingHistoryDark density="cart" />
          </div>
        </div>

        <!-- 空购物车 -->
        <div v-else class="wa-drawer-content cart-drawer-content cart-drawer-content--empty relative z-10">
          <div class="cart-drawer-empty-state">
            <Icon name="lucide:shopping-bag" class="w-24 h-24 tz-text-muted mb-4" />
            <p class="tz-text-primary text-lg font-medium mb-2">{{ t('cartDrawer.empty.title') }}</p>
            <p class="tz-text-secondary text-sm mb-6">{{ t('cartDrawer.empty.description') }}</p>
            <button
              @click="closeCart"
              class="px-6 py-2 bg-white text-black rounded-lg hover:bg-white/90 transition-colors"
            >
              {{ t('cartDrawer.actions.continueShopping') }}
            </button>
          </div>

          <!-- 浏览历史组件 -->
          <div class="cart-drawer-history-section cart-drawer-history-section--empty">
            <BrowsingHistoryDark density="cart" />
          </div>
        </div>

        <!-- 底部汇总 -->
        <div v-if="cartItems.length > 0" class="cart-drawer-summary border-t border-white/10 px-6 py-4 bg-white/[0.03] relative z-10">
          <div class="space-y-2 mb-4">
            <div class="flex justify-between text-sm">
              <span class="tz-text-secondary">{{ t('cartDrawer.summary.subtotal') }}</span>
              <span class="font-medium text-white">{{ formatPrice(subtotal) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="tz-text-secondary">{{ t('cartDrawer.summary.shipping') }}</span>
              <span class="font-medium text-white text-right">
                {{ t('cartDrawer.summary.calculatedAtCheckout') }}
              </span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="tz-text-secondary">{{ t('cartDrawer.summary.tax') }}</span>
              <span class="font-medium text-white text-right">
                {{ t('cartDrawer.summary.calculatedAtCheckout') }}
              </span>
            </div>
            <div class="flex justify-between text-base font-semibold pt-2 border-t border-white/10">
              <span class="text-white">{{ t('cartDrawer.summary.estimatedTotal') }}</span>
              <span class="text-white text-right">
                {{ t('cartDrawer.summary.calculatedAtCheckout') }}
              </span>
            </div>
          </div>

          <p class="text-xs tz-text-secondary mb-3 text-center">
            {{ t('cartDrawer.summary.finalShippingNote') }}
          </p>

          <div
            v-if="shouldPrepareStripeExpressCheckout"
            class="mb-3"
          >
            <StripeExpressCheckoutElement
              ref="stripeExpressCheckoutElementRef"
              :publishable-key="stripeExpressCheckoutPublishableKey"
              :amount="stripeExpressCheckoutAmount"
              :currency="cartCurrency"
              :line-items="stripeExpressCheckoutLineItems"
              :allowed-shipping-countries="stripeExpressCheckoutAllowedShippingCountries"
              :disabled="isStripeExpressCheckoutProcessing"
              @ready="handleStripeExpressCheckoutAvailability"
              @available-payment-methods-change="handleStripeExpressCheckoutAvailability"
              @confirm="handleStripeExpressCheckoutConfirm"
              @shipping-address-change="handleStripeExpressCheckoutShippingAddressChange"
              @shipping-rate-change="handleStripeExpressCheckoutShippingRateChange"
              @error="handleStripeExpressCheckoutError"
            />
            <p
              v-if="stripeExpressCheckoutError"
              class="mt-2 rounded-lg border border-rose-300/20 bg-rose-300/10 px-3 py-2 text-xs text-rose-100"
            >
              {{ stripeExpressCheckoutError }}
            </p>
          </div>

          <div class="grid gap-3 sm:grid-cols-2">
            <button
              type="button"
              class="w-full px-4 py-3 border border-white text-white rounded-lg hover:bg-white/10 transition-colors font-medium"
              @click="closeCart"
            >
              {{ t('cartDrawer.actions.continueShopping') }}
            </button>
            <button
              type="button"
              class="w-full px-4 py-3 rounded-lg bg-white text-slate-900 transition-colors font-semibold hover:bg-white/90 disabled:cursor-not-allowed disabled:opacity-50"
              :disabled="!cartItems.length"
              @click="() => openCheckout()"
            >
              {{ t('checkout.modal.actions.checkout', 'Checkout') }}
            </button>
          </div>
        </div>
          </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch, onBeforeUnmount } from 'vue'
import type {
  StripeExpressCheckoutElementConfirmEvent,
  StripeExpressCheckoutElementShippingAddressChangeEvent,
  StripeExpressCheckoutElementShippingRateChangeEvent,
} from '@stripe/stripe-js'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'
import { useWishlist } from '~/composables/useWishlist'
import { useCart } from '~/composables/useCart'
import { useAuth } from '~/composables/useAuth'
import { usePaymentMethods } from '~/composables/usePaymentMethods'
import { useStripeExpressCheckoutOrder } from '~/composables/useStripeExpressCheckoutOrder'
import {
  convertMajorAmountToStripeMinorAmount,
  type StripeExpressCheckoutAvailablePaymentMethods,
} from '~/composables/useStripeExpressCheckout'
import { COUNTRIES } from '~/data/countries'
import BrowsingHistoryDark from '~/components/BrowsingHistoryDark.vue'
import StripeExpressCheckoutElement from '~/components/StripeExpressCheckoutElement.vue'

const {
  cartItems,
  isCartOpen,
  cartVariant,
  cartCount,
  subtotal,
  cartCurrency,
  clearCart,
  closeCart,
  openCheckout,
  updateQuantity,
  incrementQuantity,
  decrementQuantity,
  removeFromCart,
  formatPrice,
} = useCart()

const { addToWishlist } = useWishlist()
const { t } = useI18n()
const auth = useAuth()
const { countryCode } = useStorefrontContext()
const {
  paymentMethodOptions,
  loadPaymentMethods,
} = usePaymentMethods()
const {
  loadStripeExpressCheckoutPublishableKey,
  createStripeExpressCheckoutOrderAndPaymentSession,
  buildExpressCheckoutShippingQuoteAddress,
} = useStripeExpressCheckoutOrder()

const SIDEBAR_TOKEN_CART = 'cart-drawer'
const stripeExpressCheckoutPublishableKey = ref('')
const stripeExpressCheckoutError = ref('')
const isStripeExpressCheckoutProcessing = ref(false)
const isStripeExpressCheckoutWalletAvailable = ref(true)
const stripeExpressCheckoutElementRef = ref<{
  submitExpressCheckoutPayment: () => Promise<void>
  confirmExpressCheckoutPayment: (clientSecret: string, returnUrl: string) => Promise<{ status: string; paymentIntentId?: string }>
  resetExpressCheckoutPaymentState: () => void
} | null>(null)

const stripeExpressCheckoutAmount = computed(() => Number(subtotal.value || 0))
const stripeExpressCheckoutLineItems = computed(() =>
  cartItems.value.map(item => ({
    name: item.title,
    amount: convertMajorAmountToStripeMinorAmount(
      Number(item.price || 0) * Math.max(1, Number(item.quantity || 1)),
      cartCurrency.value,
    ),
  })),
)
const stripeExpressCheckoutAllowedShippingCountries = computed(() =>
  COUNTRIES.map(country => country.code),
)
const stripeCardPaymentAvailable = computed(() =>
  paymentMethodOptions.value.some((option) => {
    const provider = String(option.provider || '').trim().toLowerCase()
    const code = String(option.code || option.id || '').trim().toLowerCase()
    return (provider === 'stripe' || ['card', 'credit_card', 'credit-card', 'stripe'].includes(code))
      && option.enabled !== false
      && option.available === true
  }),
)
const shouldPrepareStripeExpressCheckout = computed(() =>
  Boolean(
    isCartOpen.value
      && cartItems.value.length
      && auth.isAuthenticated.value
      && stripeCardPaymentAvailable.value
      && stripeExpressCheckoutPublishableKey.value
      && stripeExpressCheckoutAmount.value > 0
      && isStripeExpressCheckoutWalletAvailable.value,
  ),
)

watch(isCartOpen, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_CART, open)
  if (!open) {
    stripeExpressCheckoutError.value = ''
    isStripeExpressCheckoutProcessing.value = false
    isStripeExpressCheckoutWalletAvailable.value = true
  }
})

const prepareStripeExpressCheckout = async () => {
  if (!isCartOpen.value || !cartItems.value.length) return

  const session = await auth.ensureSession()
  if (!session) return

  await loadPaymentMethods(countryCode.value !== 'ZZ' ? countryCode.value : undefined)
  if (!stripeCardPaymentAvailable.value) return

  try {
    stripeExpressCheckoutPublishableKey.value = await loadStripeExpressCheckoutPublishableKey()
  } catch (error) {
    stripeExpressCheckoutPublishableKey.value = ''
    stripeExpressCheckoutError.value = error instanceof Error ? error.message : 'Express Checkout is unavailable'
  }
}

const handleStripeExpressCheckoutAvailability = (
  availablePaymentMethods: StripeExpressCheckoutAvailablePaymentMethods,
) => {
  isStripeExpressCheckoutWalletAvailable.value = availablePaymentMethods.applePay || availablePaymentMethods.googlePay
}

const unwrapCheckoutQuote = (payload: any) => {
  let current = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object') return null
    if (!('data' in current)) return current
    current = current.data
  }
  return null
}

const handleStripeExpressCheckoutShippingAddressChange = async (
  shippingEvent: StripeExpressCheckoutElementShippingAddressChangeEvent,
) => {
  try {
    const session = await auth.ensureSession()
    if (!session) {
      shippingEvent.reject()
      return
    }

    const shippingAddress = buildExpressCheckoutShippingQuoteAddress(
      shippingEvent,
      String(session.email || ''),
      String(session.profile?.phone || ''),
    )
    const quoteResponse = await auth.request('/checkout/quote', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify({
        shipping_address: shippingAddress,
        display_currency: cartCurrency.value,
      }),
    })
    const quote = unwrapCheckoutQuote(quoteResponse)
    const shippingAmount = convertMajorAmountToStripeMinorAmount(
      Number(quote?.shipping_fee || 0),
      cartCurrency.value,
    )
    const taxAmount = convertMajorAmountToStripeMinorAmount(
      Number(quote?.tax_amount || 0),
      cartCurrency.value,
    )
    const lineItems = [
      ...stripeExpressCheckoutLineItems.value,
      ...(shippingAmount > 0 ? [{ name: t('checkout.stepper.summary.shipping', 'Shipping'), amount: shippingAmount }] : []),
      ...(taxAmount > 0 ? [{ name: t('checkout.stepper.summary.tax', 'Tax'), amount: taxAmount }] : []),
    ]
    shippingEvent.resolve({ lineItems })
  } catch (error) {
    stripeExpressCheckoutError.value = error instanceof Error ? error.message : 'Shipping could not be calculated'
    shippingEvent.reject()
  }
}

const handleStripeExpressCheckoutShippingRateChange = (
  shippingEvent: StripeExpressCheckoutElementShippingRateChangeEvent,
) => {
  shippingEvent.resolve({
    lineItems: stripeExpressCheckoutLineItems.value,
  })
}

const handleStripeExpressCheckoutConfirm = async (
  confirmationEvent: StripeExpressCheckoutElementConfirmEvent,
) => {
  if (isStripeExpressCheckoutProcessing.value) {
    confirmationEvent.paymentFailed({ reason: 'fail', message: 'Another payment is already being processed' })
    return
  }

  const expressCheckoutElement = stripeExpressCheckoutElementRef.value
  if (!expressCheckoutElement) {
    confirmationEvent.paymentFailed({ reason: 'fail', message: 'Express Checkout is not ready' })
    return
  }

  isStripeExpressCheckoutProcessing.value = true
  stripeExpressCheckoutError.value = ''
  try {
    await expressCheckoutElement.submitExpressCheckoutPayment()
    const session = await createStripeExpressCheckoutOrderAndPaymentSession(
      confirmationEvent,
      cartItems.value,
    )
    const returnUrl = new URL(window.location.href)
    returnUrl.searchParams.set('order_number', session.orderNumber)
    const result = await expressCheckoutElement.confirmExpressCheckoutPayment(
      session.clientSecret,
      returnUrl.toString(),
    )
    if (['succeeded', 'processing', 'requires_capture'].includes(result.status)) {
      clearCart()
      closeCart()
      return
    }
    throw new Error(t('checkout.modal.messages.paymentPending', 'Payment is not complete yet. Please check your order status later.'))
  } catch (error) {
    expressCheckoutElement.resetExpressCheckoutPaymentState()
    const message = error instanceof Error
      ? error.message
      : t('checkout.modal.messages.orderFailed', 'Order submission failed, please try again')
    stripeExpressCheckoutError.value = message
    confirmationEvent.paymentFailed({ reason: 'fail', message })
  } finally {
    isStripeExpressCheckoutProcessing.value = false
  }
}

const handleStripeExpressCheckoutError = (message: string) => {
  stripeExpressCheckoutError.value = message
}

onBeforeUnmount(() => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_CART, false)
})

const handleAddToWishlist = async (item: any) => {
  if (!item || !item.id) return
  try {
    await addToWishlist(item.id)
  } catch (error) {
    console.error('Failed to add to wishlist from cart:', error)
  }
}

const onQuantityInput = (id: number, event: Event) => {
  const target = event.target as HTMLInputElement | null
  const raw = target ? target.value : ''
  const parsed = parseInt(raw, 10) || 1
  updateQuantity(id, parsed)
}

</script>

<style src="~/assets/css/components/whatsapp-mobile-drawer.css"></style>

<style scoped>
.cart-drawer-shell {
  height: min(92vh, var(--tz-mobile-safe-viewport-height, 92vh));
  max-height: min(92vh, var(--tz-mobile-safe-viewport-height, 92vh));
  background: #000 !important;
  background-image: none !important;
}

.cart-drawer-shell > .wa-drawer-header {
  flex: 0 0 auto;
}

@supports (height: 100dvh) {
  .cart-drawer-shell {
    height: min(92dvh, var(--tz-mobile-safe-viewport-height, 92dvh));
    max-height: min(92dvh, var(--tz-mobile-safe-viewport-height, 92dvh));
  }
}

.cart-drawer-content {
  display: flex;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 0.85rem;
  overflow: hidden;
}

.cart-drawer-cart-section {
  flex: 1 1 14rem;
  min-height: 0;
  overflow: hidden;
}

.cart-drawer-items-scroll {
  display: flex;
  height: 100%;
  min-height: 0;
  gap: 1rem;
  align-items: flex-start;
  overflow-x: auto;
  overflow-y: hidden;
  padding: 0.125rem 0.125rem 0.45rem;
  overscroll-behavior-x: contain;
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.72) transparent;
}

.cart-drawer-items-scroll::-webkit-scrollbar {
  height: 0.42rem;
}

.cart-drawer-items-scroll::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.72);
}

.cart-drawer-item-card {
  display: flex;
  flex: 0 0 min(82vw, 22rem);
  align-self: flex-start;
  min-height: 0;
  height: 8.75rem;
  max-height: none;
  gap: 1rem;
  overflow: hidden;
  padding: 0.8rem;
}

.cart-drawer-history-section {
  flex: 0 0 auto;
  min-height: 0;
  overflow: visible;
}

.cart-drawer-content--empty {
  justify-content: stretch;
}

.cart-drawer-empty-state {
  display: flex;
  flex: 0 0 auto;
  min-height: 12rem;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding-block: 1.5rem;
  text-align: center;
}

.cart-drawer-history-section--empty {
  flex: 0 0 auto;
}

.cart-drawer-summary {
  flex: 0 0 auto;
  min-height: 0;
  overflow: visible;
  padding: 0.9rem 1.5rem 1rem;
}

@media (min-width: 768px) {
  .cart-drawer-shell {
    width: 90vw;
    max-width: 90vw;
    height: min(92vh, 1040px);
    max-height: 92vh;
  }

  @supports (height: 100dvh) {
    .cart-drawer-shell {
      height: min(92dvh, 1040px);
      max-height: 92dvh;
    }
  }

  .cart-drawer-content {
    gap: 0.95rem;
    padding: 1rem 1.5rem;
  }

  .cart-drawer-item-card {
    flex-basis: min(36vw, 24rem);
  }
}

@media (max-width: 767px) {
  .cart-drawer-shell {
    width: 100%;
    max-width: 100%;
    height: calc(100vh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
    max-height: calc(100vh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
  }

  @supports (height: 100svh) {
    .cart-drawer-shell {
      height: calc(100svh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
      max-height: calc(100svh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
    }
  }

  @supports (height: 100dvh) {
    .cart-drawer-shell {
      height: calc(100dvh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
      max-height: calc(100dvh - var(--tz-safe-area-top) - var(--tz-safe-area-bottom));
    }
  }

  .cart-drawer-content {
    flex: 1 1 auto;
    gap: 0.65rem;
    padding: 0.85rem;
  }

  .cart-drawer-cart-section {
    min-height: 12rem;
  }

  .cart-drawer-history-section {
    flex-basis: auto;
  }

  .cart-drawer-item-card {
    flex-basis: min(86vw, 21rem);
    gap: 0.75rem;
    height: 8rem;
    padding: 0.7rem;
  }

  .cart-drawer-summary {
    max-height: 42%;
    overflow-y: auto;
    padding: 0.8rem 0.85rem var(--tz-mobile-modal-safe-padding-bottom, 1rem);
  }
}
</style>
