<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="isCheckoutOpen"
        class="fixed inset-0 bg-black z-[9998] flex items-center justify-center p-4"
        @click.self="closeCheckout"
      >
        <!-- 结账弹窗 -->
        <Transition name="scale">
          <div
            v-if="isCheckoutOpen"
            class="bg-black border-2 border-[#6e6ee9] rounded-2xl max-w-4xl w-full max-h-[90vh] overflow-hidden shadow-2xl"
          >
            <!-- 头部 -->
            <div class="flex items-center justify-between px-6 py-4 border-b border-white/10">
              <h2 class="text-2xl font-bold text-white">
                💳 Checkout
              </h2>
              <button
                @click="closeCheckout"
                class="w-10 h-10 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
                aria-label="Close checkout"
              >
                <svg class="w-6 h-6 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <!-- 内容区域 -->
            <div class="overflow-y-auto max-h-[calc(90vh-140px)]">
              <div class="grid md:grid-cols-2 gap-6 p-6">
                <!-- 左侧：表单 -->
                <div class="space-y-6">
                  <!-- Shipping Address -->
                  <div class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                      <svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                      </svg>
                      Shipping Address
                    </h3>
                    
                    <div class="space-y-3">
                      <div>
                        <label class="block text-sm font-medium text-white/80 mb-1">Recipient</label>
                        <input
                          v-model="form.name"
                          type="text"
                          placeholder="Enter recipient name"
                          class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white/80 mb-1">Phone</label>
                        <input
                          v-model="form.phone"
                          type="tel"
                          placeholder="Enter phone number"
                          class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                        />
                      </div>
                      
                      <div>
                        <label class="block text-sm font-medium text-white/80 mb-1">Address</label>
                        <textarea
                          v-model="form.address"
                          rows="3"
                          placeholder="Enter detailed address"
                          class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff] resize-none"
                        />
                      </div>
                      
                      <div class="grid grid-cols-2 gap-3">
                        <div>
                          <label class="block text-sm font-medium text-white/80 mb-1">City</label>
                          <input
                            v-model="form.city"
                            type="text"
                            placeholder="City"
                            class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                          />
                        </div>
                        <div>
                          <label class="block text-sm font-medium text-white/80 mb-1">Zip Code</label>
                          <input
                            v-model="form.zip"
                            type="text"
                            placeholder="Zip code"
                            class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                          />
                        </div>
                      </div>
                    </div>
                  </div>

                  <!-- Payment Method -->
                  <div class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                      <svg class="w-5 h-5 text-purple-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h18M7 15h1m4 0h1m-7 4h12a3 3 0 003-3V8a3 3 0 00-3-3H6a3 3 0 00-3 3v8a3 3 0 003 3z" />
                      </svg>
                      Payment Method
                    </h3>
                    
                    <div class="space-y-2">
                      <label
                        v-for="method in paymentMethods"
                        :key="method.id"
                        class="flex items-center gap-3 p-3 border-2 rounded-lg cursor-pointer transition-all"
                        :class="form.paymentMethod === method.id ? 'border-[#6b73ff] bg-[#6b73ff]/10' : 'border-white/20 hover:border-white/30'"
                      >
                        <input
                          v-model="form.paymentMethod"
                          type="radio"
                          :value="method.id"
                          class="w-4 h-4 text-[#6b73ff]"
                        />
                        <span class="text-2xl">{{ method.icon }}</span>
                        <span class="font-medium text-white">{{ method.name }}</span>
                      </label>
                    </div>
                  </div>
                </div>

                <!-- 右侧：订单摘要 -->
                <div class="space-y-6">
                  <div class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <h3 class="text-lg font-semibold text-white mb-4 flex items-center gap-2">
                      <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
                      </svg>
                      Order Summary
                    </h3>

                    <!-- 商品列表 -->
                    <div class="space-y-3 mb-4 max-h-60 overflow-y-auto">
                      <div
                        v-for="item in cartItems"
                        :key="item.id"
                        class="flex gap-3 p-3 bg-white/[0.04] border border-white/10 rounded-lg"
                      >
                        <div class="w-16 h-16 flex-shrink-0 bg-white/5 rounded-lg overflow-hidden border border-white">
                          <img
                            v-if="item.thumbnail"
                            :src="item.thumbnail"
                            :alt="item.title"
                            class="w-full h-full object-cover"
                          />
                        </div>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium text-white truncate">{{ item.title }}</p>
                          <p class="text-xs text-white/60 mt-1">Qty: {{ item.quantity }}</p>
                          <p class="text-sm font-semibold text-white mt-1">
                            {{ formatPrice(item.price * item.quantity) }}
                          </p>
                        </div>
                      </div>
                    </div>

                    <!-- 费用明细 -->
                    <div class="space-y-2 pt-4 border-t border-white/10">
                      <div class="flex justify-between text-sm">
                        <span class="text-white/70">Subtotal</span>
                        <span class="font-medium text-white">{{ formatPrice(priceBreakdown.subtotal) }}</span>
                      </div>
                      
                      <div class="flex justify-between text-sm">
                        <span class="text-white/70">Shipping</span>
                        <span class="font-medium text-white">
                          {{ priceBreakdown.shipping === 0 ? 'Free' : formatPrice(priceBreakdown.shipping) }}
                        </span>
                      </div>
                      <div class="flex justify-between text-sm">
                        <span class="text-white/70">Tax</span>
                        <span class="font-medium text-white">{{ formatPrice(priceBreakdown.tax) }}</span>
                      </div>
                      <div class="flex justify-between text-lg font-bold pt-3 border-t border-white/20">
                        <span class="text-white">Total Amount</span>
                        <span class="text-[#6b73ff]">{{ formatPrice(priceBreakdown.total) }}</span>
                      </div>
                    </div>
                  </div>

                  <!-- Coupon -->
                  <div class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <h3 class="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                      <svg class="w-4 h-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M5 5a3 3 0 015-2.236A3 3 0 0114.83 6H16a2 2 0 110 4h-5V9a1 1 0 10-2 0v1H4a2 2 0 110-4h1.17C5.06 5.687 5 5.35 5 5zm4 1V5a1 1 0 10-1 1h1zm3 0a1 1 0 10-1-1v1h1z" clip-rule="evenodd" />
                      </svg>
                      Coupon
                    </h3>
                    <div class="flex gap-2">
                      <input
                        v-model="couponCode"
                        type="text"
                        placeholder="Enter coupon code"
                        class="flex-1 px-3 py-2 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff] text-sm"
                      />
                      <button
                        @click="handleApplyCoupon"
                        :disabled="!couponCode || isApplyingCoupon"
                        class="px-4 py-2 bg-[#6b73ff] text-white rounded-lg hover:brightness-110 transition-all text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {{ isApplyingCoupon ? 'Applying...' : 'Apply' }}
                      </button>
                    </div>
                    <p v-if="calculation.appliedCoupon.value" class="mt-2 text-xs text-green-400 flex items-center gap-1">
                      <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
                      </svg>
                      Coupon applied: {{ calculation.appliedCoupon.value.code }}
                    </p>
                  </div>

                  <!-- Points Discount -->
                  <div v-if="calculation.userPoints.value" class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <h3 class="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                      <svg class="w-4 h-4 text-purple-500" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M8.433 7.418c.155-.103.346-.196.567-.267v1.698a2.305 2.305 0 01-.567-.267C8.07 8.34 8 8.114 8 8c0-.114.07-.34.433-.582zM11 12.849v-1.698c.22.071.412.164.567.267.364.243.433.468.433.582 0 .114-.07.34-.433.582a2.305 2.305 0 01-.567.267z" />
                        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-13a1 1 0 10-2 0v.092a4.535 4.535 0 00-1.676.662C6.602 6.234 6 7.009 6 8c0 .99.602 1.765 1.324 2.246.48.32 1.054.545 1.676.662v1.941c-.391-.127-.68-.317-.843-.504a1 1 0 10-1.51 1.31c.562.649 1.413 1.076 2.353 1.253V15a1 1 0 102 0v-.092a4.535 4.535 0 001.676-.662C13.398 13.766 14 12.991 14 12c0-.99-.602-1.765-1.324-2.246A4.535 4.535 0 0011 9.092V7.151c.391.127.68.317.843.504a1 1 0 101.511-1.31c-.563-.649-1.413-1.076-2.354-1.253V5z" clip-rule="evenodd" />
                      </svg>
                      Use Points Discount
                    </h3>
                    <div class="flex items-center gap-3 mb-2">
                      <input
                        v-model="calculation.usePointsDiscount.value"
                        type="checkbox"
                        class="w-4 h-4 text-[#6b73ff] rounded"
                      />
                      <span class="text-sm text-white/80">Use points (Available: {{ calculation.userPoints.value.available }} pts)</span>
                    </div>
                    <div v-if="calculation.usePointsDiscount.value" class="mt-3">
                      <label class="block text-xs text-white/70 mb-1">Points to use</label>
                      <input
                        :value="calculation.pointsToUse.value"
                        @input="calculation.setPointsUsage(parseInt(($event.target as HTMLInputElement).value) || 0)"
                        type="number"
                        :max="calculation.userPoints.value.available"
                        min="0"
                        class="w-full px-3 py-2 bg-white/5 border border-white rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-[#6b73ff] text-sm"
                      />
                      <p class="mt-1 text-xs text-white/60">1 point = $0.01, max 50% of order</p>
                    </div>
                  </div>

                  <!-- Notes -->
                  <div class="bg-white/[0.06] border border-white/10 rounded-xl p-5">
                    <label class="block text-sm font-medium text-white/80 mb-2">Order Notes (Optional)</label>
                    <textarea
                      v-model="form.notes"
                      rows="3"
                      placeholder="Any special requests..."
                      class="w-full px-4 py-2.5 bg-white/5 border border-white rounded-lg text-white placeholder:text-white/40 focus:outline-none focus:ring-2 focus:ring-[#6b73ff] resize-none"
                    />
                  </div>
                </div>
              </div>
            </div>

            <!-- 底部按钮 -->
            <div class="border-t border-white/10 px-6 py-4 flex gap-3">
              <button
                @click="backToCart"
                class="px-6 py-3 border border-white text-white rounded-lg hover:bg-white/10 transition-colors font-medium"
              >
                ← Back to Cart
              </button>
              <button
                @click="handleSubmit"
                :disabled="!isFormValid || isSubmitting"
                class="flex-1 px-6 py-3 bg-[#6b73ff] text-white rounded-lg hover:brightness-110 transition-all font-semibold shadow-lg disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <span v-if="!isSubmitting">Confirm Payment {{ formatPrice(priceBreakdown.total) }}</span>
                <span v-else class="flex items-center justify-center gap-2">
                  <svg class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Processing...
                </span>
              </button>
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const {
  cartItems,
  isCheckoutOpen,
  priceBreakdown,
  closeCheckout,
  backToCart,
  formatPrice,
  clearCart,
  calculation,
} = useCart()

// 初始化计算系统
onMounted(() => {
  calculation.initialize()
})

// 支付方式列表
const paymentMethods = [
  { id: 'credit_card', name: 'Credit Card', icon: '💳' },
  { id: 'paypal', name: 'PayPal', icon: '🅿️' },
  { id: 'alipay', name: 'Alipay', icon: '💙' },
  { id: 'wechat', name: 'WeChat Pay', icon: '💚' },
]

// 表单数据
const form = ref({
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
    form.value.name.trim() !== '' &&
    form.value.phone.trim() !== '' &&
    form.value.address.trim() !== '' &&
    form.value.city.trim() !== '' &&
    form.value.paymentMethod !== ''
  )
})

// 提交订单
const handleSubmit = async () => {
  if (!isFormValid.value || isSubmitting.value) return

  isSubmitting.value = true

  try {
    // 这里调用你的订单 API
    // const response = await $fetch('/wp-json/tanzanite/v1/orders', {
    //   method: 'POST',
    //   body: {
    //     items: cartItems.value,
    //     shipping: form.value,
    //     payment_method: form.value.paymentMethod,
    //     notes: form.value.notes,
    //     total: total.value,
    //   }
    // })

    // 模拟 API 调用
    await new Promise(resolve => setTimeout(resolve, 2000))

    // 成功后清空购物车
    clearCart()
    closeCheckout()

    // 显示成功消息
    alert('Order submitted successfully!')

    // 重置表单
    form.value = {
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
