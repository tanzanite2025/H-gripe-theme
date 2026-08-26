import { ref, computed, watch } from 'vue'
import { useRuntimeConfig } from '#imports'
import type { CartItem } from '~~/types/cart'
import { useAuth } from '~/composables/useAuth'
import { useCartCalculation } from '~/composables/useCartCalculation'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import { isExpectedAnonymousApiMiss, logUnexpectedApiError } from '~/utils/storefrontApiFailures'
import {
  createStorefrontMediaContext,
  normalizeStorefrontMediaUrl,
} from '~/utils/storefrontMedia'

export interface ShippingAddress {
  name: string
  phone: string
  address: string
  city: string
  state: string
  zip: string
  country: string
}

const cartItems = ref<CartItem[]>([])
const isCartOpen = ref(false)
const isCheckoutOpen = ref(false)
const cartVariant = ref<'default' | 'checkout-bottom' | 'lever-bottom' | 'chat-bottom'>('default')
const preferredCheckoutPaymentMethod = ref('')
const shippingAddress = ref<ShippingAddress | null>(null)
const isLoadingCart = ref(false)

let eventListenersAdded = false
let cartBackendLoaded = false

const cartItemKey = (productId: number, variantId?: number | null) => variantId || productId

const normalizeCurrencyCode = (value: unknown) => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : ''
}

const normalizeCheckoutPaymentMethod = (value?: string | null) => {
  const method = String(value || '').trim().toLowerCase()
  if (['stripe', 'credit_card', 'credit-card'].includes(method)) return 'card'
  if (['card', 'paypal', 'alipay', 'wechat'].includes(method)) return method
  return ''
}

const extractCartSummaryItems = (payload: unknown): unknown[] => {
  let current = payload
  for (let depth = 0; depth < 3; depth += 1) {
    if (!current || typeof current !== 'object' || Array.isArray(current)) return []
    const record = current as Record<string, unknown>
    if (Array.isArray(record.items)) return record.items
    current = record.data
  }
  return []
}

const resolveProductThumbnail = (
  product: any,
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): string => {
  const media = Array.isArray(product?.media) ? product.media : []
  const imageMedia = media.filter((item: any) => {
    return item?.media_type === 'image' && item?.url && item?.is_visible !== false
  })
  const primaryImage =
    imageMedia.find((item: any) => item?.is_primary || item?.role === 'primary') ||
    imageMedia[0]
  const imageVariants: Record<string, any> = primaryImage?.image_variants && typeof primaryImage.image_variants === 'object'
    ? primaryImage.image_variants
    : {}

  return normalizeStorefrontMediaUrl(
    product?.thumbnail ||
      product?.featured_image ||
      imageVariants.thumbnail?.url ||
      primaryImage?.thumbnail_url ||
      imageVariants.card?.url ||
      primaryImage?.url,
    mediaContext,
  )
}

const normalizeBackendCartItem = (
  item: any,
  fallbackCurrency = 'USD',
  mediaContext: ReturnType<typeof createStorefrontMediaContext>,
): CartItem => {
  const productId = item.product_id
  const variantId = item.variant_id || null
  const product = item.product || {}
  const variant = item.variant || {}
  const thumbnail = resolveProductThumbnail(product, mediaContext)
  const itemCurrency = normalizeCurrencyCode(item.currency || variant.currency || product.currency) || normalizeCurrencyCode(fallbackCurrency) || 'USD'

  return {
    id: cartItemKey(productId, variantId),
    product_id: productId,
    variant_id: variantId,
    name: product.name || 'Unknown Product',
    title: product.name || 'Unknown Product',
    slug: product.slug || '',
    price: item.price,
    currency: itemCurrency,
    sale_price: product.sale_price,
    quantity: item.quantity,
    image: thumbnail,
    thumbnail,
    categories: product.categories || [],
  }
}

export const useCart = () => {
  const runtimeConfig = useRuntimeConfig()
  const mediaContext = createStorefrontMediaContext(runtimeConfig)
  const auth = useAuth()
  const calculation = useCartCalculation()
  const { track: trackBehaviorEvent } = useBehaviorEvents()
  const { baseCurrency } = useStorefrontContext()
  const overlayBackStack = useOverlayBackStack()

  const loadCartFromBackend = async () => {
    isLoadingCart.value = true
    try {
      const summary = await auth.request<any>('/cart/summary')
      cartItems.value = extractCartSummaryItems(summary)
        .map((item: any) => normalizeBackendCartItem(item, baseCurrency.value, mediaContext))
      cartBackendLoaded = true
    } catch (e) {
      if (isExpectedAnonymousApiMiss(e)) {
        cartItems.value = []
        cartBackendLoaded = true
        return
      }
      logUnexpectedApiError('Failed to load cart from backend', e, isExpectedAnonymousApiMiss)
    } finally {
      isLoadingCart.value = false
    }
  }

  const loadCartForInteraction = async () => {
    if (cartBackendLoaded || isLoadingCart.value) return
    if (!auth.initialized.value) {
      await auth.ensureSession()
    }
    const saved = localStorage.getItem('commerce_platform_cart')
    if (saved && !auth.isAuthenticated.value) {
      await syncGuestCart()
      return
    }
    await loadCartFromBackend()
  }

  const syncGuestCart = async (): Promise<{
    success: boolean
    error?: string
    itemsCount?: number
  }> => {
    if (!import.meta.client) {
      return { success: false, error: 'Not in client' }
    }

    const saved = localStorage.getItem('commerce_platform_cart')
    if (!saved) {
      return { success: true, itemsCount: 0 }
    }

    try {
      const items = JSON.parse(saved)
      if (!items || items.length === 0) {
        localStorage.removeItem('commerce_platform_cart')
        return { success: true, itemsCount: 0 }
      }

      const payload = items.map((item: any) => ({
        product_id: item.product_id || item.id,
        variant_id: item.variant_id || null,
        quantity: item.quantity,
      }))

      let lastError: unknown
      for (let attempt = 1; attempt <= 3; attempt++) {
        try {
          await auth.request('/cart/sync', {
            method: 'POST',
            body: JSON.stringify(payload),
          })

          localStorage.removeItem('commerce_platform_cart')
          await loadCartFromBackend()

          return {
            success: true,
            itemsCount: items.length,
          }
        } catch (e) {
          lastError = e
          console.warn(`[Cart] Sync attempt ${attempt}/3 failed:`, e)

          if (attempt < 3) {
            await new Promise(resolve => setTimeout(resolve, attempt * 1000))
          }
        }
      }

      console.error('[Cart] Failed to sync guest cart after 3 attempts:', lastError)

      return {
        success: false,
        error: lastError instanceof Error ? lastError.message : 'Sync failed',
        itemsCount: items.length,
      }
    } catch (e) {
      console.error('[Cart] Failed to parse guest cart', e)

      return {
        success: false,
        error: 'Failed to parse cart data',
      }
    }
  }

  const syncAction = async (
    action: 'add' | 'update' | 'remove' | 'clear',
    productId?: number,
    quantity?: number,
    variantId?: number | null,
  ) => {
    try {
      if (action === 'add') {
        await auth.request('/cart/add', {
          method: 'POST',
          body: JSON.stringify({ product_id: productId, variant_id: variantId || null, quantity }),
        })
      } else if (action === 'update') {
        await auth.request(`/cart/items/${productId}`, {
          method: 'PUT',
          body: JSON.stringify({ variant_id: variantId || null, quantity }),
        })
      } else if (action === 'remove') {
        const suffix = variantId ? `?variant_id=${variantId}` : ''
        await auth.request(`/cart/items/${productId}${suffix}`, { method: 'DELETE' })
      } else if (action === 'clear') {
        await auth.request('/cart/clear', { method: 'POST' })
      }
    } catch (e) {
      console.error('Cart sync failed', e)
      await loadCartFromBackend()
    }
  }

  if (import.meta.client && !eventListenersAdded) {
    eventListenersAdded = true

    window.addEventListener('open-cart-drawer', () => {
      openCart()
    })
  }

  watch(() => auth.isAuthenticated.value, async (newVal, oldVal) => {
    if (newVal && !oldVal) {
      const result = await syncGuestCart()

      if (!result.success && result.itemsCount && result.itemsCount > 0) {
        console.error('[Cart] Failed to sync cart:', result.error)

        if (typeof window !== 'undefined') {
          const retry = window.confirm(
            `Cart sync failed for ${result.itemsCount} item(s). Local data is kept.\n\nRefresh and retry?`,
          )

          if (retry) {
            window.location.reload()
          }
        }
      } else if (!result.itemsCount) {
        await loadCartFromBackend()
      }
    } else if (!newVal && oldVal) {
      cartItems.value = []
      cartBackendLoaded = false
    }
  })

  const cartCount = computed(() => cartItems.value.reduce((sum, item) => sum + item.quantity, 0))
  const subtotal = computed(() => calculation.calculateSubtotal(cartItems.value))
  const shipping = computed(() => calculation.calculateShipping(cartItems.value, subtotal.value))
  const tax = computed(() => calculation.calculateTax(subtotal.value, shipping.value))
  const total = computed(() => calculation.calculateTotal(cartItems.value).total)
  const priceBreakdown = computed(() => calculation.calculateTotal(cartItems.value))
  const cartCurrency = computed(() => {
    const firstCurrency = cartItems.value.map(item => normalizeCurrencyCode(item.currency)).find(Boolean)
    return firstCurrency || baseCurrency.value || 'USD'
  })

  const addToCart = (product: Omit<CartItem, 'quantity'>, quantity = 1) => {
    const productId = product.product_id || product.id
    const variantId = product.variant_id || null
    const itemId = cartItemKey(productId, variantId)
    const existingItem = cartItems.value.find(item => item.id === itemId)
    const quantityToAdd = Math.max(1, Math.floor(Number(quantity) || 1))
    const itemCurrency = normalizeCurrencyCode(product.currency) || normalizeCurrencyCode(baseCurrency.value) || 'USD'
    const cartCurrencies = new Set(
      cartItems.value
        .map(item => normalizeCurrencyCode(item.currency))
        .filter(Boolean),
    )
    if (cartCurrencies.size > 0 && [...cartCurrencies].some(currency => currency !== itemCurrency)) {
      return {
        success: false,
        message: `购物车已有 ${[...cartCurrencies].join(', ')} 商品，不能加入 ${itemCurrency} 商品。请先清空购物车或分开结算。`,
      }
    }
    const normalizedProduct = {
      ...product,
      currency: itemCurrency,
      weight: product.weight ?? (product.weight_grams ? product.weight_grams / 1000 : undefined),
    }
    let syncPromise: Promise<void>

    if (existingItem) {
      existingItem.quantity += quantityToAdd
      syncPromise = syncAction('update', productId, existingItem.quantity, variantId)
    } else {
      cartItems.value.push({ ...normalizedProduct, id: itemId, product_id: productId, variant_id: variantId, quantity: quantityToAdd })
      syncPromise = syncAction('add', productId, quantityToAdd, variantId)
    }

    trackBehaviorEvent({
      eventType: 'add_to_cart',
      productId,
      metadata: {
        source: 'cart_action',
        variant_id: variantId || 0,
        quantity: quantityToAdd,
        cart_action: existingItem ? 'increment' : 'add',
      },
    })

    return { success: true, message: 'Added to cart', syncPromise }
  }

  const updateQuantity = (id: number, quantity: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (quantity <= 0) {
      removeFromCart(id)
      return
    }
    item.quantity = quantity
    syncAction('update', item.product_id || item.id, quantity, item.variant_id || null)
  }

  const incrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    item.quantity++
    syncAction('update', item.product_id || item.id, item.quantity, item.variant_id || null)
    return { success: true }
  }

  const decrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (item.quantity <= 1) {
      removeFromCart(id)
      return
    }
    item.quantity--
    syncAction('update', item.product_id || item.id, item.quantity, item.variant_id || null)
  }

  const removeFromCart = (id: number) => {
    const index = cartItems.value.findIndex(item => item.id === id)
    if (index > -1) {
      const item = cartItems.value[index]
      if (!item) return
      cartItems.value.splice(index, 1)
      syncAction('remove', item.product_id || item.id, undefined, item.variant_id || null)
    }
  }

  const clearCart = () => {
    cartItems.value = []
    syncAction('clear')
  }

  const closeCartState = () => {
    isCartOpen.value = false
  }

  const closeCheckoutState = () => {
    isCheckoutOpen.value = false
  }

  const openCart = () => {
    cartVariant.value = 'default'
    isCartOpen.value = true
    void loadCartForInteraction()
    overlayBackStack.open('cart-drawer', closeCartState)
  }

  const closeCart = () => {
    void overlayBackStack.close('cart-drawer')
    closeCartState()
  }

  const toggleCart = () => {
    if (isCartOpen.value) {
      closeCart()
    } else {
      openCart()
    }
  }
  const openCheckout = (paymentMethod?: string) => {
    preferredCheckoutPaymentMethod.value = normalizeCheckoutPaymentMethod(paymentMethod)
    void loadCartForInteraction()

    const restoreCartOnClose = isCartOpen.value
    if (restoreCartOnClose) {
      overlayBackStack.setResume('cart-drawer', () => {
        isCartOpen.value = true
      })
      overlayBackStack.open('checkout-modal', closeCheckoutState, {
        mode: 'push',
      })
    } else {
      overlayBackStack.open('checkout-modal', closeCheckoutState)
    }

    isCartOpen.value = false
    isCheckoutOpen.value = true
  }

  const closeCheckout = async () => {
    const closePromise = overlayBackStack.close('checkout-modal')
    closeCheckoutState()
    await closePromise
  }

  const backToCart = async () => {
    await closeCheckout()
    openCartFromCheckout()
  }
  const openCartFromCheckout = () => {
    cartVariant.value = 'checkout-bottom'
    isCartOpen.value = true
    void loadCartForInteraction()
    overlayBackStack.open('cart-drawer', closeCartState)
  }
  const openCartFromLever = () => {
    cartVariant.value = 'lever-bottom'
    isCartOpen.value = true
    void loadCartForInteraction()
    overlayBackStack.open('cart-drawer', closeCartState, { mode: 'push' })
  }
  const openCartFromChat = () => {
    cartVariant.value = 'chat-bottom'
    isCartOpen.value = true
    void loadCartForInteraction()
    overlayBackStack.open('cart-drawer', closeCartState, { mode: 'push' })
  }

  const setShippingAddress = (address: ShippingAddress) => { shippingAddress.value = address }

  const formatPrice = (price: number, currency = cartCurrency.value || baseCurrency.value || 'USD') => {
    try {
      if (!currency) return new Intl.NumberFormat('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 }).format(price)
      return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(price)
    } catch {
      return `${currency || ''} ${Number(price || 0).toFixed(2)}`.trim()
    }
  }

  return {
    cartItems,
    isCartOpen,
    isCheckoutOpen,
    cartVariant,
    preferredCheckoutPaymentMethod,
    shippingAddress,
    isLoadingCart,

    cartCount,
    subtotal,
    shipping,
    tax,
    total,
    priceBreakdown,
    cartCurrency,
    calculation,

    addToCart,
    updateQuantity,
    incrementQuantity,
    decrementQuantity,
    removeFromCart,
    clearCart,

    openCart,
    closeCart,
    toggleCart,
    openCheckout,
    closeCheckout,
    backToCart,
    openCartFromCheckout,
    openCartFromLever,
    openCartFromChat,

    setShippingAddress,
    formatPrice,
  }
}
