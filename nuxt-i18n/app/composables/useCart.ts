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

export const useCart = () => {
  const auth = useAuth()
  const calculation = useCartCalculation()

  // 1. 从云端拉取购物车
  const loadCartFromBackend = async () => {
    isLoadingCart.value = true
    try {
      const summary = await auth.request<any>('/cart/summary')
      if (summary && summary.items) {
        cartItems.value = summary.items.map((item: any) => ({
          id: item.product_id,
          name: item.product?.name || 'Unknown Product',
          slug: item.product?.slug || '',
          sku: item.product?.sku || '',
          price: item.price,
          sale_price: item.product?.sale_price,
          quantity: item.quantity,
          image: item.product?.images?.[0]?.url || '',
          categories: (() => { if (!item.product?.categories) throw new Error("[CRITICAL] item.product.categories missing"); return item.product.categories; })(),
          stock: item.product?.stock || 0,
          maxStock: item.product?.stock || 0,
        }))
      } else {
        cartItems.value = []
      }
    } catch (e) {
      console.error('Failed to load cart from backend', e)
    } finally {
      isLoadingCart.value = false
    }
  }

  // 2. 游客本地购物车合并到云端（带重试和失败处理）
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
        product_id: item.id,
        quantity: item.quantity
      }))

      // 使用重试机制同步到后端（最多3次）
      let lastError: any
      for (let attempt = 1; attempt <= 3; attempt++) {
        try {
          await auth.request('/cart/sync', {
            method: 'POST',
            body: JSON.stringify(payload)
          })

          // ✅ 同步成功，删除本地数据
          localStorage.removeItem('tanzanite_cart')
          await loadCartFromBackend()

          return {
            success: true,
            itemsCount: items.length
          }
        } catch (e) {
          lastError = e
          console.warn(`[Cart] Sync attempt ${attempt}/3 failed:`, e)

          if (attempt < 3) {
            // 等待后重试（延迟递增：1s, 2s）
            await new Promise(resolve => setTimeout(resolve, attempt * 1000))
          }
        }
      }

      // ❌ 3次重试后仍然失败，保留本地数据
      console.error('[Cart] Failed to sync guest cart after 3 attempts:', lastError)

      return {
        success: false,
        error: lastError instanceof Error ? lastError.message : 'Sync failed',
        itemsCount: items.length
      }
    } catch (e) {
      console.error('[Cart] Failed to parse guest cart', e)

      return {
        success: false,
        error: 'Failed to parse cart data'
      }
    }
  }

  // 3. 乐观更新后同步到云端
  const syncAction = async (action: 'add' | 'update' | 'remove' | 'clear', productId?: number, quantity?: number) => {
    try {
      if (action === 'add') {
        await auth.request('/cart/add', { method: 'POST', body: JSON.stringify({ product_id: productId, quantity }) })
      } else if (action === 'update') {
        await auth.request(`/cart/items/${productId}`, { method: 'PUT', body: JSON.stringify({ quantity }) })
      } else if (action === 'remove') {
        await auth.request(`/cart/items/${productId}`, { method: 'DELETE' })
      } else if (action === 'clear') {
        await auth.request('/cart/clear', { method: 'POST' })
      }
    } catch (e) {
      console.error('Cart sync failed', e)
      // 同步失败时回滚本地状态
      loadCartFromBackend()
    }
  }

  // 客户端初始化
  if (import.meta.client) {
    if (!eventListenersAdded) {
      eventListenersAdded = true
      
      // 初始加载云端购物车
      loadCartFromBackend()
      
      // 游客旧版 localStorage 数据迁移到云端 session
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
  }

  // 监听登录状态变化
  watch(() => auth.isAuthenticated.value, async (newVal, oldVal) => {
    if (newVal && !oldVal) {
      // 用户刚登录，将本地购物车与云端合并
      const result = await syncGuestCart()

      if (!result.success && result.itemsCount && result.itemsCount > 0) {
        // 同步失败，显示错误提示
        console.error('[Cart] Failed to sync cart:', result.error)

        // 使用浏览器原生提示（因为可能还没加载ElMessage）
        if (typeof window !== 'undefined' && window.alert) {
          const retry = window.confirm(
            `购物车同步失败（${result.itemsCount}件商品），本地数据已保留。\n\n是否刷新页面重试？`
          )

          if (retry) {
            window.location.reload()
          }
        }
      } else if (result.success && result.itemsCount && result.itemsCount > 0) {
        // 同步成功，显示成功提示
      }
    } else if (!newVal && oldVal) {
      // 用户登出，重新拉取游客空车
      await loadCartFromBackend()
    }
  })

  // 计算属性
  const cartCount = computed(() => cartItems.value.reduce((sum, item) => sum + item.quantity, 0))
  const subtotal = computed(() => calculation.calculateSubtotal(cartItems.value))
  const shipping = computed(() => calculation.calculateShipping(cartItems.value, subtotal.value))
  const tax = computed(() => calculation.calculateTax(subtotal.value, shipping.value))
  const total = computed(() => calculation.calculateTotal(cartItems.value).total)
  const priceBreakdown = computed(() => calculation.calculateTotal(cartItems.value))

  // 添加到购物车
  const addToCart = (product: Omit<CartItem, 'quantity'>) => {
    const existingItem = cartItems.value.find(item => item.id === product.id)
    if (existingItem) {
      if (existingItem.maxStock && existingItem.quantity >= existingItem.maxStock) {
        return { success: false, message: 'Stock limit reached' }
      }
      existingItem.quantity++
      syncAction('update', product.id, existingItem.quantity)
    } else {
      cartItems.value.push({ ...product, quantity: 1 })
      syncAction('add', product.id, 1)
    }
    return { success: true, message: 'Added to cart' }
  }

  // 更新数量
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
    syncAction('update', id, quantity)
  }

  // 增加数量
  const incrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (item.maxStock && item.quantity >= item.maxStock) {
      return { success: false, message: 'Stock limit reached' }
    }
    item.quantity++
    syncAction('update', id, item.quantity)
    return { success: true }
  }

  // 减少数量
  const decrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return
    if (item.quantity <= 1) {
      removeFromCart(id)
      return
    }
    item.quantity--
    syncAction('update', id, item.quantity)
  }

  // 从购物车移除
  const removeFromCart = (id: number) => {
    const index = cartItems.value.findIndex(item => item.id === id)
    if (index > -1) {
      cartItems.value.splice(index, 1)
      syncAction('remove', id)
    }
  }

  // 清空购物车
  const clearCart = () => {
    cartItems.value = []
    syncAction('clear')
  }

  // UI 交互方法
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
