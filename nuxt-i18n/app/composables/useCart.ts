import { ref, computed, watch } from 'vue'
import type { CartItem } from '~/types/cart'
import { useAuth } from '~/composables/useAuth'
import { useCartCalculation } from '~/composables/useCartCalculation'

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
const shippingAddress = ref<ShippingAddress | null>(null)
const selectedPaymentMethod = ref<string>('')
const isLoadingCart = ref(false)

let eventListenersAdded = false

const cartItemKey = (productId: number, variantId?: number | null) => variantId || productId

const normalizeBackendCartItem = (item: any): CartItem => {
  const productId = item.product_id
  const variantId = item.variant_id || null
  const product = item.product || {}
  const variant = item.variant || {}
  const stock = variant.stock ?? product.stock ?? 0
  const weightGrams = variant.weight_grams ?? product.weight_grams ?? null

  return {
    id: cartItemKey(productId, variantId),
    product_id: productId,
    variant_id: variantId,
    name: product.name || 'Unknown Product',
    title: product.name || 'Unknown Product',
    slug: product.slug || '',
    sku: variant.sku || product.sku || '',
    product_type_id: product.product_type_id ?? null,
    price: item.price,
    sale_price: variant.sale_price ?? product.sale_price,
    quantity: item.quantity,
    image: product.images?.[0]?.url || '',
    categories: product.categories || [],
    stock,
    maxStock: stock,
    weight_grams: weightGrams,
    weight: weightGrams ? weightGrams / 1000 : undefined,
  }
}

export const useCart = () => {
  const auth = useAuth()
  const calculation = useCartCalculation()

  const loadCartFromBackend = async () => {
    isLoadingCart.value = true
    try {
      const summary = await auth.request<any>('/cart/summary')
      cartItems.value = summary?.items?.map(normalizeBackendCartItem) || []
    } catch (e) {
      console.error('Failed to load cart from backend', e)
    } finally {
      isLoadingCart.value = false
    }
  }

  const syncGuestCart = async (): Promise<{
    success: boolean
    error?: string
    itemsCount?: number
  }> => {
    if (!import.meta.client) {
      return { success: false, error: 'Not in client' }
    }

    const saved = localStorage.getItem('tanzanite_cart')
    if (!saved) {
      return { success: true, itemsCount: 0 }
    }

    try {
      const items = JSON.parse(saved)
      if (!items || items.length === 0) {
        localStorage.removeItem('tanzanite_cart')
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

          localStorage.removeItem('tanzanite_cart')
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
    loadCartFromBackend()

    const saved = localStorage.getItem('tanzanite_cart')
    if (saved && !auth.isAuthenticated.value) {
      syncGuestCart()
    }

    window.addEventListener('open-cart-drawer', () => {
      isCartOpen.value = true
    })
    window.addEventListener('open-checkout-modal', () => {
      isCartOpen.value = false
      isCheckoutOpen.value = true
    })
  }

  watch(() => auth.isAuthenticated.value, async (newVal, oldVal) => {
    if (newVal && !oldVal) {
      const result = await syncGuestCart()

      if (!result.success && result.itemsCount && result.itemsCount > 0) {
        console.error('[Cart] Failed to sync cart:', result.error)

        if (typeof window !== 'undefined' && window.alert) {
          const retry = window.confirm(
            `Cart sync failed for ${result.itemsCount} item(s). Local data is kept.\n\nRefresh and retry?`,
          )

          if (retry) {
            window.location.reload()
          }
        }
      }
    } else if (!newVal && oldVal) {
      await loadCartFromBackend()
    }
  })

  const cartCount = computed(() => cartItems.value.reduce((sum, item) => sum + item.quantity, 0))
  const subtotal = computed(() => calculation.calculateSubtotal(cartItems.value))
  const shipping = computed(() => calculation.calculateShipping(cartItems.value, subtotal.value))
  const tax = computed(() => calculation.calculateTax(subtotal.value, shipping.value))
  const total = computed(() => calculation.calculateTotal(cartItems.value).total)
  const priceBreakdown = computed(() => calculation.calculateTotal(cartItems.value))

  const addToCart = (product: Omit<CartItem, 'quantity'>) => {
    const productId = product.product_id || product.id
    const variantId = product.variant_id || null
    const itemId = cartItemKey(productId, variantId)
    const existingItem = cartItems.value.find(item => item.id === itemId)
    const normalizedProduct = {
      ...product,
      weight: product.weight ?? (product.weight_grams ? product.weight_grams / 1000 : undefined),
    }

    if (existingItem) {
      if (existingItem.maxStock && existingItem.quantity >= existingItem.maxStock) {
        return { success: false, message: 'Stock limit reached' }
      }
      existingItem.quantity++
      syncAction('update', productId, existingItem.quantity, variantId)
    } else {
      cartItems.value.push({ ...normalizedProduct, id: itemId, product_id: productId, variant_id: variantId, quantity: 1 })
      syncAction('add', productId, 1, variantId)
    }

    return { success: true, message: 'Added to cart' }
  }

  const updateQuantity = (id: number, quantity: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (quantity <= 0) {
      removeFromCart(id)
      return
    }
    if (item.maxStock && quantity > item.maxStock) {
      quantity = item.maxStock
    }
    item.quantity = quantity
    syncAction('update', item.product_id || item.id, quantity, item.variant_id || null)
  }

  const incrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (item.maxStock && item.quantity >= item.maxStock) {
      return { success: false, message: 'Stock limit reached' }
    }
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
      cartItems.value.splice(index, 1)
      syncAction('remove', item.product_id || item.id, undefined, item.variant_id || null)
    }
  }

  const clearCart = () => {
    cartItems.value = []
    syncAction('clear')
  }

  const openCart = () => { cartVariant.value = 'default'; isCartOpen.value = true }
  const closeCart = () => { isCartOpen.value = false }
  const toggleCart = () => { isCartOpen.value = !isCartOpen.value }
  const openCartFromCheckout = () => { cartVariant.value = 'checkout-bottom'; isCartOpen.value = true }
  const openCartFromLever = () => { cartVariant.value = 'lever-bottom'; isCartOpen.value = true }
  const openCartFromChat = () => { cartVariant.value = 'chat-bottom'; isCartOpen.value = true }

  const openCheckout = () => { isCartOpen.value = false; isCheckoutOpen.value = true }
  const closeCheckout = () => { isCheckoutOpen.value = false }
  const backToCart = () => { isCheckoutOpen.value = false; isCartOpen.value = true }

  const setShippingAddress = (address: ShippingAddress) => { shippingAddress.value = address }
  const setPaymentMethod = (method: string) => { selectedPaymentMethod.value = method }

  const formatPrice = (price: number) => new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(price)

  return {
    cartItems,
    isCartOpen,
    isCheckoutOpen,
    cartVariant,
    shippingAddress,
    selectedPaymentMethod,
    isLoadingCart,

    cartCount,
    subtotal,
    shipping,
    tax,
    total,
    priceBreakdown,
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
    openCartFromCheckout,
    openCartFromLever,
    openCartFromChat,
    openCheckout,
    closeCheckout,
    backToCart,

    setShippingAddress,
    setPaymentMethod,
    formatPrice,
  }
}
