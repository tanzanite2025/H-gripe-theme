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
          categories: item.product?.categories || [],
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

  // 2. 游客本地购物车合并到云端
  const syncGuestCart = async () => {
    if (import.meta.client) {
      const saved = localStorage.getItem('tanzanite_cart')
      if (saved) {
        try {
          const items = JSON.parse(saved)
          if (items && items.length > 0) {
            const payload = items.map((item: any) => ({ product_id: item.id, quantity: item.quantity }))
            await auth.request('/cart/sync', {
              method: 'POST',
              body: JSON.stringify(payload)
            })
          }
          localStorage.removeItem('tanzanite_cart')
        } catch (e) {
          console.error('Failed to sync guest cart', e)
        }
      }
      await loadCartFromBackend()
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
  watch(() => auth.isAuthenticated.value, (newVal) => {
    if (newVal) {
      // 用户登录后，将 session 的购物车与 user 的购物车合并
      syncGuestCart()
    } else {
      // 登出后重新拉取游客空车
      loadCartFromBackend()
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
