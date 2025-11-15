import { ref, computed } from 'vue'

export interface CartItem {
  id: number
  title: string
  slug: string
  price: number
  quantity: number
  thumbnail?: string
  sku?: string
  maxStock?: number
  weight?: number // 商品重量（克）
}

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
const shippingAddress = ref<ShippingAddress | null>(null)
const selectedPaymentMethod = ref<string>('')
const isSyncing = ref(false)

export const useCart = () => {
  // 保存到 localStorage
  const saveCart = () => {
    if (import.meta.client) {
      localStorage.setItem('tanzanite_cart', JSON.stringify(cartItems.value))
    }
  }

  // 从服务器加载购物车
  const loadCartFromServer = async () => {
    try {
      const response: any = await $fetch('/wp-json/tanzanite/v1/cart', {
        method: 'GET',
        credentials: 'include'
      })
      
      if (response.success && response.data) {
        cartItems.value = response.data
        saveCart() // 同步到 localStorage
      }
    } catch (error) {
      console.error('Failed to load cart from server', error)
      // 如果服务器失败，使用本地缓存
    }
  }
  
  // 从 localStorage 加载购物车（快速响应）
  if (import.meta.client) {
    const saved = localStorage.getItem('tanzanite_cart')
    if (saved) {
      try {
        cartItems.value = JSON.parse(saved)
      } catch (e) {
        console.error('Failed to load cart from localStorage', e)
      }
    }
    
    // 页面加载时从服务器恢复购物车
    loadCartFromServer()
  }

  // 同步到服务器（后台）
  const syncToServer = async () => {
    if (isSyncing.value) return
    
    try {
      isSyncing.value = true
      const response: any = await $fetch('/wp-json/tanzanite/v1/cart/sync', {
        method: 'POST',
        body: {
          cart_items: cartItems.value
        },
        credentials: 'include'
      })
      
      if (response.success && response.data) {
        cartItems.value = response.data
        saveCart()
      }
    } catch (error) {
      console.error('Failed to sync cart to server', error)
    } finally {
      isSyncing.value = false
    }
  }

  // 计算属性
  const cartCount = computed(() => {
    return cartItems.value.reduce((sum, item) => sum + item.quantity, 0)
  })

  // 集成高级计算系统
  const calculation = useCartCalculation()

  const subtotal = computed(() => {
    return calculation.calculateSubtotal(cartItems.value)
  })

  const shipping = computed(() => {
    return calculation.calculateShipping(cartItems.value, subtotal.value)
  })

  const tax = computed(() => {
    return calculation.calculateTax(subtotal.value, shipping.value)
  })

  const total = computed(() => {
    const result = calculation.calculateTotal(cartItems.value)
    return result.total
  })

  // 完整的价格明细
  const priceBreakdown = computed(() => {
    return calculation.calculateTotal(cartItems.value)
  })

  // 添加到购物车（混合方案）
  const addToCart = async (product: Omit<CartItem, 'quantity'>) => {
    // 1. 先添加到本地（快速响应）
    const existingItem = cartItems.value.find(item => item.id === product.id)
    
    if (existingItem) {
      // 检查库存
      if (existingItem.maxStock && existingItem.quantity >= existingItem.maxStock) {
        return { success: false, message: 'Stock limit reached' }
      }
      existingItem.quantity++
    } else {
      cartItems.value.push({ ...product, quantity: 1 })
    }
    
    saveCart() // 保存到 localStorage
    
    // 2. 同步到服务器（后台）
    try {
      await $fetch('/wp-json/tanzanite/v1/cart/add', {
        method: 'POST',
        body: {
          product_id: product.id,
          quantity: 1
        },
        credentials: 'include'
      })
    } catch (error) {
      console.error('Failed to sync add to server', error)
    }
    
    return { success: true, message: 'Added to cart' }
  }

  // 更新数量
  const updateQuantity = async (id: number, quantity: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return

    if (quantity <= 0) {
      await removeFromCart(id)
      return
    }

    // 检查库存
    if (item.maxStock && quantity > item.maxStock) {
      quantity = item.maxStock
    }

    // 更新本地
    item.quantity = quantity
    saveCart()
    
    // 同步到服务器
    try {
      await $fetch('/wp-json/tanzanite/v1/cart/update', {
        method: 'PUT',
        body: {
          cart_item_id: id,
          quantity: quantity
        },
        credentials: 'include'
      })
    } catch (error) {
      console.error('Failed to update cart on server', error)
    }
  }

  // 增加数量
  const incrementQuantity = async (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return

    if (item.maxStock && item.quantity >= item.maxStock) {
      return { success: false, message: 'Stock limit reached' }
    }

    item.quantity++
    saveCart()
    
    // 同步到服务器
    syncToServer()
    
    return { success: true }
  }

  // 减少数量
  const decrementQuantity = async (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return

    if (item.quantity <= 1) {
      await removeFromCart(id)
      return
    }

    item.quantity--
    saveCart()
    
    // 同步到服务器
    syncToServer()
  }

  // 从购物车移除
  const removeFromCart = async (id: number) => {
    const index = cartItems.value.findIndex(item => item.id === id)
    if (index > -1) {
      cartItems.value.splice(index, 1)
      saveCart()
      
      // 同步到服务器
      try {
        await $fetch('/wp-json/tanzanite/v1/cart/remove', {
          method: 'DELETE',
          body: {
            cart_item_id: id
          },
          credentials: 'include'
        })
      } catch (error) {
        console.error('Failed to remove from server', error)
      }
    }
  }

  // 清空购物车
  const clearCart = async () => {
    cartItems.value = []
    saveCart()
    
    // 同步到服务器
    try {
      await $fetch('/wp-json/tanzanite/v1/cart/clear', {
        method: 'DELETE',
        credentials: 'include'
      })
    } catch (error) {
      console.error('Failed to clear cart on server', error)
    }
  }

  // 打开/关闭购物车
  const openCart = () => {
    isCartOpen.value = true
  }

  const closeCart = () => {
    isCartOpen.value = false
  }

  const toggleCart = () => {
    isCartOpen.value = !isCartOpen.value
  }

  // 打开/关闭结账
  const openCheckout = () => {
    isCartOpen.value = false
    isCheckoutOpen.value = true
  }

  const closeCheckout = () => {
    isCheckoutOpen.value = false
  }

  const backToCart = () => {
    isCheckoutOpen.value = false
    isCartOpen.value = true
  }

  // 设置收货地址
  const setShippingAddress = (address: ShippingAddress) => {
    shippingAddress.value = address
  }

  // 设置支付方式
  const setPaymentMethod = (method: string) => {
    selectedPaymentMethod.value = method
  }

  // 创建订单
  const createOrder = async () => {
    if (!shippingAddress.value) {
      return { success: false, message: 'Shipping address required' }
    }
    
    try {
      const response: any = await $fetch('/wp-json/tanzanite/v1/orders/create', {
        method: 'POST',
        body: {
          cart_items: cartItems.value,
          shipping_address: shippingAddress.value,
          payment_method: selectedPaymentMethod.value || 'cod'
        },
        credentials: 'include'
      })
      
      if (response.success) {
        // 清空购物车
        await clearCart()
        return response
      }
      
      return { success: false, message: 'Order creation failed' }
    } catch (error) {
      console.error('Failed to create order', error)
      return { success: false, message: 'Order creation failed' }
    }
  }

  // 格式化价格
  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  return {
    // 状态
    cartItems,
    isCartOpen,
    isCheckoutOpen,
    shippingAddress,
    selectedPaymentMethod,
    isSyncing,
    
    // 计算属性
    cartCount,
    subtotal,
    shipping,
    tax,
    total,
    priceBreakdown,
    
    // 高级计算系统
    calculation,
    
    // 方法
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
    setShippingAddress,
    setPaymentMethod,
    createOrder,
    formatPrice,
    loadCartFromServer,
    syncToServer,
  }
}
