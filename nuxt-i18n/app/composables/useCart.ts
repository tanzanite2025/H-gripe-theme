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
// cartVariant 用来区分不同父级打开购物车时的停靠方式：
// - default: 全屏遮罩 + 居中弹窗（导航、小球等默认入口）
// - checkout-bottom: 结账页面底部抽屉模式
// - lever-bottom: LeverAndPoint 弹窗内专用底部抽屉模式
// - chat-bottom: WhatsApp 聊天页专用底部抽屉模式
const cartVariant = ref<'default' | 'checkout-bottom' | 'lever-bottom' | 'chat-bottom'>('default')
const shippingAddress = ref<ShippingAddress | null>(null)
const selectedPaymentMethod = ref<string>('')

// 全局事件监听器标志，确保只添加一次
let eventListenersAdded = false

export const useCart = () => {
  // 从 localStorage 加载购物车
  if (import.meta.client) {
    const saved = localStorage.getItem('tanzanite_cart')
    if (saved) {
      try {
        cartItems.value = JSON.parse(saved)
      } catch (e) {
        console.error('Failed to load cart from localStorage', e)
      }
    }
    
    // 只添加一次全局事件监听器
    if (!eventListenersAdded) {
      eventListenersAdded = true
      
      // 监听全局事件打开购物车
      const handleOpenCart = () => {
        isCartOpen.value = true
      }
      window.addEventListener('open-cart-drawer', handleOpenCart)
      
      // 监听全局事件打开结账页面
      const handleOpenCheckout = () => {
        isCartOpen.value = false
        isCheckoutOpen.value = true
      }
      window.addEventListener('open-checkout-modal', handleOpenCheckout)
    }
  }

  // 保存到 localStorage
  const saveCart = () => {
    if (import.meta.client) {
      localStorage.setItem('tanzanite_cart', JSON.stringify(cartItems.value))
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

  // 添加到购物车
  const addToCart = (product: Omit<CartItem, 'quantity'>) => {
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
    
    saveCart()
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

    // 检查库存
    if (item.maxStock && quantity > item.maxStock) {
      quantity = item.maxStock
    }

    item.quantity = quantity
    saveCart()
  }

  // 增加数量
  const incrementQuantity = (id: number) => {
    const item = cartItems.value.find(item => item.id === id)
    if (!item) return

    if (item.maxStock && item.quantity >= item.maxStock) {
      return { success: false, message: 'Stock limit reached' }
    }

    item.quantity++
    saveCart()
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
    saveCart()
  }

  // 从购物车移除
  const removeFromCart = (id: number) => {
    const index = cartItems.value.findIndex(item => item.id === id)
    if (index > -1) {
      cartItems.value.splice(index, 1)
      saveCart()
    }
  }

  // 清空购物车
  const clearCart = () => {
    cartItems.value = []
    saveCart()
  }

  // 打开/关闭购物车
  const openCart = () => {
    // 默认入口：使用居中弹窗样式
    cartVariant.value = 'default'
    isCartOpen.value = true
  }

  const closeCart = () => {
    isCartOpen.value = false
  }

  const toggleCart = () => {
    isCartOpen.value = !isCartOpen.value
  }

  // 从 Checkout 打开购物车：保持结账弹窗打开，仅在底部以抽屉形式展示购物车
  const openCartFromCheckout = () => {
    cartVariant.value = 'checkout-bottom'
    isCartOpen.value = true
  }

  // 从 LeverAndPoint 等父级打开购物车的专用底部抽屉模式
  const openCartFromLever = () => {
    cartVariant.value = 'lever-bottom'
    isCartOpen.value = true
  }

  // 从 WhatsApp 聊天页等入口打开购物车：专用底部抽屉模式
  const openCartFromChat = () => {
    cartVariant.value = 'chat-bottom'
    isCartOpen.value = true
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
    cartVariant,
    shippingAddress,
    selectedPaymentMethod,
    
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
