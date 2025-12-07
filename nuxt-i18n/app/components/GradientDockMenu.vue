<template>
  <!-- Dock 菜单容器 (统一胶囊风格) -->
  <div class="fixed left-1/2 -translate-x-1/2 bottom-6 w-[96%] max-w-[380px] md:max-w-[500px] z-[101] pointer-events-auto transition-all duration-300">
    <div class="w-full bg-[#0b1020]/80 backdrop-blur-md rounded-[24px] md:rounded-full border border-white/10 shadow-[0_10px_30px_rgba(0,0,0,0.5)] px-1 py-2 md:px-4 md:py-3 flex items-center justify-between transition-all duration-300">
      
      <!-- 1. Menu (Sidebar) -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 text-white/60 hover:text-white transition-colors py-1 min-w-[40px]"
        @click="openSidebarLeft"
        aria-label="Open sidebar"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="md:w-6 md:h-6 transition-all"><line x1="3" y1="12" x2="21" y2="12"/><line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="18" x2="21" y2="18"/></svg>
        <span class="text-[9px] md:text-xs font-medium tracking-tight">Menu</span>
      </button>

      <!-- 2. Chat -->
      <button 
        :class="[
          'flex-1 flex flex-col items-center gap-0.5 transition-colors py-1 min-w-[40px] relative',
          isChatOpen ? 'text-[#40ffaa]' : 'text-white/60 hover:text-white'
        ]"
        @click="toggleChatModal()" 
        aria-label="Chat"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="md:w-6 md:h-6 transition-all"><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/></svg>
        <span class="text-[9px] md:text-xs font-medium tracking-tight">Chat</span>
        <!-- Unread Badge -->
        <span
          v-if="totalUnreadCount > 0"
          class="absolute top-0 right-1 md:right-4 w-2 h-2 md:w-2.5 md:h-2.5 bg-red-500 rounded-full border border-[#0b1020]"
        ></span>
      </button>

      <!-- 3. Checkout (Main Action) -->
      <button 
        class="h-10 px-2 mx-0.5 md:h-12 md:px-6 md:mx-2 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-[#0b1020] font-bold text-xs md:text-sm flex items-center justify-center shadow-lg shadow-[#40ffaa]/20 hover:shadow-[#40ffaa]/40 transition-all transform hover:-translate-y-0.5 min-w-[64px] md:min-w-[100px]"
        @click="openCheckoutFromDock"
      >
        <span>{{ priceDisplay }}</span>
      </button>

      <!-- 4. Quick Buy -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 text-white/60 hover:text-[#40ffaa] transition-colors py-1 min-w-[40px]"
        @click="openQuick()" 
        aria-haspopup="dialog" 
        :aria-expanded="quickOpen" 
        aria-label="Quick buy"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="md:w-6 md:h-6 transition-all"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
        <span class="text-[9px] md:text-xs font-medium tracking-tight">Quick</span>
      </button>

      <!-- 5. Cart -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 text-white/60 hover:text-white transition-colors py-1 min-w-[40px]"
        @click="openCartDrawer" 
        aria-label="Open cart"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="md:w-6 md:h-6 transition-all"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
        <span class="text-[9px] md:text-xs font-medium tracking-tight">Cart</span>
      </button>

      <!-- 6. Saved (Wishlist) -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 text-white/60 hover:text-white transition-colors py-1 min-w-[40px] relative"
        @click="openWishlist" 
        aria-label="Open wishlist"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="md:w-6 md:h-6 transition-all"><path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/></svg>
        <span class="text-[9px] md:text-xs font-medium tracking-tight">Saved</span>
        <span
          v-if="wishlistCount > 0"
          class="absolute top-0 right-1 md:right-5 w-2 h-2 md:w-2.5 md:h-2.5 bg-[#40ffaa] rounded-full border border-[#0b1020]"
        ></span>
      </button>

    </div>
  </div>
  
  <!-- Quick Buy Modal from Dock -->
  <QuickBuyModal v-if="quickOpen" :config="props.config || null" @close="quickOpen = false" />
  
  <!-- WhatsApp 聊天弹窗 -->
  <WhatsAppChatModal 
    v-if="currentConversation"
    :conversation="currentConversation"
    @close="handleCloseChat"
  />

  <!-- Wishlist 抽屉弹窗 -->
  <WishlistDrawer v-model="wishlistDrawerVisible" />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n, useRuntimeConfig } from '#imports'
import QuickBuyModal from '@/components/QuickBuy.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import { useWishlist } from '~/composables/useWishlist'

// floating submenu state
const isOpen = ref(false)
const quickOpen = ref(false)
const currentConversation = ref<any>(null)
const wishlistDrawerVisible = ref(false)

// mutually exclusive open helpers
const closeAll = () => {
  isOpen.value = false
  quickOpen.value = false
  wishlistDrawerVisible.value = false
}

// 关闭聊天窗口
const handleCloseChat = () => {
  currentConversation.value = null
}

const openQuick = () => {
  closeAll()
  quickOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'dock-quick' } }))
  }
}

// removed old share popup and outside-click listeners; modal closes by overlay click

// Types aligned with CartSummaryBar.vue
interface QuickBuyConfig {
  steps?: unknown[]
  storeApiBase?: string
  cartUrl?: string
  checkoutUrl?: string
}

interface CartTotals {
  total_price?: string | number
  currency_symbol?: string
}

interface CartResponse {
  items_count?: number
  items_weight?: number
  totals?: CartTotals
}

// accept optional config; keep flexible to match QuickBuyModal expected shape
const props = defineProps<{ config?: any }>()

// i18n and runtime config
const runtimeConfig = useRuntimeConfig()
const { t: $t } = useI18n()

// 未读消息数（从 localStorage 跟踪）
const totalUnreadCount = ref(0)

// 聊天窗口是否打开
const isChatOpen = computed(() => !!currentConversation.value)

// 直接打开聊天窗口（WhatsAppChatModal 内部会显示客服列表）
const openChatModal = () => {
  closeAll()
  // 打开聊天窗口，不需要预先选择客服
  currentConversation.value = { showAgentList: true }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
  }
}

// 切换聊天窗口打开 / 关闭
const toggleChatModal = () => {
  if (isChatOpen.value) {
    handleCloseChat()
  } else {
    openChatModal()
  }
}

// 打开心愿单抽屉
const openWishlist = () => {
  closeAll()
  wishlistDrawerVisible.value = true
}

// 打开左侧 Sidebar（通过全局自定义事件通知 SidePanel）
const openSidebarLeft = () => {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:sidebar-open', { detail: { side: 'left' } }))
  }
}

// 计算未读消息数（从 localStorage）
const calculateUnreadCount = () => {
  try {
    let total = 0
    const keys = Object.keys(localStorage)
    const chatKeys = keys.filter(key => key.startsWith('tz_chat_'))
    
    chatKeys.forEach(key => {
      const data = localStorage.getItem(key)
      if (data) {
        const parsed = JSON.parse(data)
        // 统计未读消息（这里简单处理，可以根据实际需求调整）
        const unread = parsed.messages?.filter((msg: any) => !msg.is_read && msg.is_agent)
        total += unread?.length || 0
      }
    })
    
    totalUnreadCount.value = total
  } catch (error) {
    console.error('计算未读消息失败:', error)
  }
}

// 组件挂载时计算未读消息数
onMounted(() => {
  calculateUnreadCount()
  
  // 每30秒更新一次未读消息数
  setInterval(calculateUnreadCount, 30000)
})

// cart summary data
const summary = ref<CartResponse | null>(null)
const loading = ref(false)

const apiBase = computed(() => {
  const fromProp = props.config?.storeApiBase?.replace(/\/$/, '')
  if (fromProp) return fromProp
  const fallback = (runtimeConfig.public as { storeApiBase?: string }).storeApiBase
  return fallback ? String(fallback).replace(/\/$/, '') : ''
})

const cartUrl = computed(() => {
  if (props.config?.cartUrl) return props.config.cartUrl
  const fallback = (runtimeConfig.public as { cartUrl?: string }).cartUrl
  return fallback && fallback.trim().length ? fallback : '/cart'
})

const itemsLabel = computed(() => $t('cart.summary.items', 'Items'))
const priceLabel = computed(() => $t('cart.summary.price', 'Price'))
const ctaLabel = computed(() => $t('cart.summary.openCart', 'View cart summary'))

// 集成购物车系统
const { cartCount, total, openCart, formatPrice } = useCart()

const itemsCount = computed(() => cartCount.value)

// 心愿单数量（使用全局 wishlist composable）
const { items: wishlistItems, loadWishlist, loadedOnce } = useWishlist()
const wishlistCount = computed(() => wishlistItems.value.length)

const priceDisplay = computed(() => {
  return formatPrice(total.value)
})

// 初次挂载时，如果尚未加载过心愿单，则触发一次加载
onMounted(() => {
  try {
    if (!loadedOnce.value) {
      loadWishlist()
    }
  } catch {}
})

const fetchSummary = async () => {
  if (!apiBase.value || loading.value) return
  loading.value = true
  try {
    const response = await $fetch<CartResponse>(`${apiBase.value}/cart`, { credentials: 'include' })
    summary.value = response
  } catch (e) {
    console.warn('Dock summary fetch failed', e)
  } finally {
    loading.value = false
  }
}

const openCartDrawer = () => {
  openCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'cart-drawer' } }))
  }
}

// 从 Dock 直接打开结账弹窗
const openCheckoutFromDock = () => {
  closeAll()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('open-checkout-modal'))
  }
}

watch(apiBase, () => {
  summary.value = null
  if (apiBase.value) fetchSummary()
}, { immediate: true })

onMounted(() => {
  if (summary.value === null) fetchSummary()
  // global popup listener: close this component's popups when others open
  const onGlobalPopup = (e: any) => {
    try {
      const id = e?.detail?.id as string | undefined
      if (!id) return
      if (id === 'dock-fab') {
        quickOpen.value = false
      } else if (id === 'dock-quick') {
        isOpen.value = false
      } else {
        // opened by other components (e.g., language switcher) -> close all dock popups
        closeAll()
      }
    } catch {}
  }
  window.addEventListener('ui:popup-open', onGlobalPopup)
  ;(window as any)._dockOnGlobalPopup = onGlobalPopup
})
onBeforeUnmount(() => {
  // remove global listener with stored reference
  const ref = (window as any)._dockOnGlobalPopup
  if (ref) window.removeEventListener('ui:popup-open', ref)
})

// defensive: ensure mutual exclusivity if any state is toggled externally
watchEffect(() => {
  const openCount = [isOpen.value, quickOpen.value].filter(Boolean).length
  if (openCount > 1) {
    // prefer the most recently opened by simple priority: quick > fab
    if (quickOpen.value) {
      isOpen.value = false
    } else if (isOpen.value) {
      quickOpen.value = false
    }
  }
})
</script>

