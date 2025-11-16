<template>
  <Teleport to="body">
    <!-- 购物车弹窗 -->
    <Transition
      enter-active-class="transition-opacity duration-300 ease-out"
      leave-active-class="transition-opacity duration-200 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <div v-if="isCartOpen" class="fixed inset-0 z-[9999] flex items-center justify-center" @click.self="closeCart">
        <!-- 不透明背景遮罩 -->
        <div class="absolute inset-0 bg-black"></div>
        <!-- 弹窗内容 -->
        <div class="relative w-[min(95vw,1650px)] max-h-[90vh] bg-black border-2 border-[#6e6ee9] rounded-2xl shadow-2xl flex flex-col overflow-hidden" aria-modal="true" role="dialog" aria-label="Shopping Cart">
        <!-- 头部 -->
        <div class="flex items-center justify-between px-6 py-4 border-b border-white/10">
          <h2 class="text-xl font-semibold text-white">
            🛒 Cart ({{ cartCount }})
          </h2>
          <button
            @click="closeCart"
            class="w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/10 transition-colors"
            aria-label="Close cart"
          >
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- 购物车内容 -->
        <div v-if="cartItems.length > 0" class="flex-1 overflow-y-auto px-6 py-4">
          <div class="space-y-4">
            <div
              v-for="item in cartItems"
              :key="item.id"
              class="flex gap-4 p-4 bg-white/[0.06] border border-white rounded-2xl"
            >
              <!-- 商品图片 -->
              <div class="w-20 h-20 flex-shrink-0 bg-white/[0.06] rounded-lg overflow-hidden border border-white">
                <img
                  v-if="item.thumbnail"
                  :src="item.thumbnail"
                  :alt="item.title"
                  class="w-full h-full object-cover"
                />
                <div v-else class="w-full h-full flex items-center justify-center text-white/50">
                  <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
              </div>

              <!-- 商品信息 -->
              <div class="flex-1 min-w-0">
                <h3 class="text-sm font-medium text-white truncate">
                  {{ item.title }}
                </h3>
                <p v-if="item.sku" class="text-xs text-white/50 mt-1">
                  SKU: {{ item.sku }}
                </p>
                <p class="text-sm font-semibold text-white mt-2">
                  {{ formatPrice(item.price) }}
                </p>

                <!-- 数量控制 -->
                <div class="flex items-center gap-2 mt-3">
                  <button
                    @click="decrementQuantity(item.id)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white/[0.18] hover:bg-white/10 transition-colors text-white"
                    :disabled="item.quantity <= 1"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4" />
                    </svg>
                  </button>
                  
                  <input
                    type="number"
                    :value="item.quantity"
                    @input="updateQuantity(item.id, parseInt(($event.target as HTMLInputElement).value) || 1)"
                    class="w-12 h-7 text-center border border-white rounded bg-white/[0.06] text-white focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                    min="1"
                    :max="item.maxStock"
                  />
                  
                  <button
                    @click="incrementQuantity(item.id)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white hover:bg-white/10 transition-colors text-white"
                    :disabled="item.maxStock ? item.quantity >= item.maxStock : false"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                    </svg>
                  </button>

                  <button
                    @click="removeFromCart(item.id)"
                    class="ml-auto text-red-400 hover:text-red-300 text-sm font-medium"
                  >
                    Remove
                  </button>
                </div>
              </div>
            </div>
          </div>

          <!-- 浏览历史组件 -->
          <div class="mt-6">
            <BrowsingHistoryDark />
          </div>
        </div>

        <!-- 空购物车 -->
        <div v-else class="flex-1 overflow-y-auto px-6 py-4">
          <div class="flex flex-col items-center justify-center py-12">
            <svg class="w-24 h-24 text-white/30 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z" />
            </svg>
            <p class="text-white/70 text-lg font-medium mb-2">Your cart is empty</p>
            <p class="text-white/50 text-sm mb-6">Add some products to get started!</p>
            <button
              @click="closeCart"
              class="px-6 py-2 bg-[#6b73ff] text-white rounded-lg hover:bg-[#5d65e8] transition-colors"
            >
              Continue Shopping
            </button>
          </div>

          <!-- 浏览历史组件 -->
          <div class="mt-6">
            <BrowsingHistoryDark />
          </div>
        </div>

        <!-- 底部汇总 -->
        <div v-if="cartItems.length > 0" class="border-t border-white/10 px-6 py-4 bg-white/[0.03]">
          <div class="space-y-2 mb-4">
            <div class="flex justify-between text-sm">
              <span class="text-white/70">Subtotal</span>
              <span class="font-medium text-white">{{ formatPrice(subtotal) }}</span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-white/70">Shipping</span>
              <span class="font-medium text-white">
                {{ shipping === 0 ? 'Free' : formatPrice(shipping) }}
              </span>
            </div>
            <div class="flex justify-between text-sm">
              <span class="text-white/70">Tax</span>
              <span class="font-medium text-white">{{ formatPrice(tax) }}</span>
            </div>
            <div class="flex justify-between text-base font-semibold pt-2 border-t border-white/10">
              <span class="text-white">Total</span>
              <span class="text-white">{{ formatPrice(total) }}</span>
            </div>
          </div>

          <div class="flex gap-3">
            <button
              @click="closeCart"
              class="flex-1 px-4 py-3 border border-white text-white rounded-lg hover:bg-white/10 transition-colors font-medium"
            >
              Continue Shopping
            </button>
            <button
              @click="openCheckout"
              class="flex-1 px-4 py-3 bg-gradient-to-r from-[#6b73ff] to-[#a78bfa] text-white rounded-lg hover:brightness-110 transition-all font-medium shadow-lg"
            >
              Checkout →
            </button>
          </div>
        </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
const {
  cartItems,
  isCartOpen,
  cartCount,
  subtotal,
  shipping,
  tax,
  total,
  closeCart,
  updateQuantity,
  incrementQuantity,
  decrementQuantity,
  removeFromCart,
  openCheckout,
  formatPrice,
} = useCart()
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

/* 右侧滑入动画 */
.slide-right-enter-active,
.slide-right-leave-active {
  transition: transform 0.3s ease;
}

.slide-right-enter-from,
.slide-right-leave-to {
  transform: translateX(100%);
}
</style>
