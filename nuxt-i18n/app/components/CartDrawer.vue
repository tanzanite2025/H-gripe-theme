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
          class="wa-drawer-shell"
          aria-modal="true"
          role="dialog"
          :aria-label="t('cartDrawer.ariaLabel')"
        >
        <!-- 背景装饰 -->
        <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none z-0"></div>
        
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
            <span class="text-lg leading-none">x</span>
          </button>
        </div>

        <!-- 购物车内容 -->
        <div v-if="cartItems.length > 0" class="wa-drawer-content relative z-10">
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
              <div v-else class="w-full h-full flex items-center justify-center tz-text-muted">
                  <Icon name="lucide:image" class="w-8 h-8" />
                </div>
              </div>

              <!-- 商品信息 -->
              <div class="flex-1 min-w-0">
                <h3 class="text-sm font-medium text-white truncate">
                  {{ item.title }}
                </h3>
              <p v-if="item.sku" class="text-xs tz-text-muted mt-1">
                  {{ t('cartDrawer.item.sku') }}: {{ item.sku }}
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
                    <Icon name="lucide:minus" class="w-4 h-4" />
                  </button>
                  
                  <input
                    type="number"
                    :value="item.quantity"
                    @input="onQuantityInput(item.id, $event)"
                    class="w-12 h-7 text-center border border-white rounded bg-white/[0.06] text-white focus:outline-none focus:ring-2 focus:ring-[#6b73ff]"
                    min="1"
                    :max="item.maxStock"
                  />
                  
                  <button
                    @click="incrementQuantity(item.id)"
                    class="w-7 h-7 flex items-center justify-center rounded border border-white hover:bg-white/10 transition-colors text-white"
                    :disabled="item.maxStock ? item.quantity >= item.maxStock : false"
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

          <!-- 浏览历史组件 -->
          <div class="mt-6">
            <BrowsingHistoryDark />
          </div>
        </div>

        <!-- 空购物车 -->
        <div v-else class="wa-drawer-content relative z-10 flex flex-col">
          <div class="flex flex-col items-center justify-center py-12">
            <Icon name="lucide:shopping-bag" class="w-24 h-24 tz-text-muted mb-4" />
            <p class="tz-text-primary text-lg font-medium mb-2">{{ t('cartDrawer.empty.title') }}</p>
            <p class="tz-text-secondary text-sm mb-6">{{ t('cartDrawer.empty.description') }}</p>
            <button
              @click="closeCart"
              class="px-6 py-2 bg-[#6b73ff] text-white rounded-lg hover:bg-[#5d65e8] transition-colors"
            >
              {{ t('cartDrawer.actions.continueShopping') }}
            </button>
          </div>

          <!-- 浏览历史组件 -->
          <div class="mt-6">
            <BrowsingHistoryDark />
          </div>
        </div>

        <!-- 底部汇总 -->
        <div v-if="cartItems.length > 0" class="border-t border-white/10 px-6 py-4 bg-white/[0.03] relative z-10">
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
              <span class="font-medium text-white">{{ formatPrice(tax) }}</span>
            </div>
            <div class="flex justify-between text-base font-semibold pt-2 border-t border-white/10">
              <span class="text-white">{{ t('cartDrawer.summary.estimatedTotal') }}</span>
              <span class="text-white">{{ formatPrice(total) }}</span>
            </div>
          </div>

          <p class="text-xs tz-text-secondary mb-3 text-center">
            {{ t('cartDrawer.summary.finalShippingNote') }}
          </p>

          <div class="flex gap-3">
            <button
              @click="closeCart"
              class="flex-1 px-4 py-3 border border-white text-white rounded-lg hover:bg-white/10 transition-colors font-medium"
            >
              {{ t('cartDrawer.actions.continueShopping') }}
            </button>
            <button
              @click="openCheckout"
              class="flex-1 px-4 py-3 bg-gradient-to-r from-[#6b73ff] to-[#a78bfa] text-white rounded-lg hover:brightness-110 transition-all font-medium shadow-lg"
            >
              {{ t('cartDrawer.actions.checkout') }} →
            </button>
          </div>
        </div>
          </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { watch, onBeforeUnmount } from 'vue'
import { setSidebarHandlesHidden } from '~/utils/sidebarHandles'
import { useWishlist } from '~/composables/useWishlist'
import { useCart } from '~/composables/useCart'
import BrowsingHistoryDark from '~/components/BrowsingHistoryDark.vue'

const {
  cartItems,
  isCartOpen,
  cartVariant,
  cartCount,
  subtotal,
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

const { addToWishlist } = useWishlist()
const { t } = useI18n()

const SIDEBAR_TOKEN_CART = 'cart-drawer'

watch(isCartOpen, (open) => {
  setSidebarHandlesHidden(SIDEBAR_TOKEN_CART, open)
}, { immediate: true })

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

<style scoped>
/* Inline styles removed in favor of global .wa-drawer classes */
</style>
