<template>
  <!-- Dock 菜单容器 (统一胶囊风格) -->
  <div class="fixed left-0 md:left-1/2 md:-translate-x-1/2 bottom-0 md:bottom-6 w-full md:w-[96%] md:max-w-[500px] z-[101] pointer-events-auto transition-all duration-300">
    <div class="w-full bg-[radial-gradient(circle_at_top,rgba(31,41,55,0.98),rgba(15,23,42,0.98),rgba(15,23,42,1))] backdrop-blur-md rounded-none md:rounded-full shadow-[0_24px_56px_-16px_rgba(0,0,0,1)] px-1 py-2 md:px-4 md:py-3 flex items-center justify-between transition-all duration-300">
      
      <!-- 1. Menu (Sidebar) -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 tz-text-secondary hover:text-white transition-colors py-1 min-w-[40px]"
        @click="openSidebarLeft"
        :aria-label="$t('dockMenu.openSidebar')"
      >
        <Icon name="lucide:menu" class="w-6 h-6 md:w-6 md:h-6 transition-all" />
        <span class="text-[9px] md:text-xs font-medium tracking-tight">{{ $t('dockMenu.menu') }}</span>
      </button>

      <!-- 2. Chat -->
      <button 
        :class="[
          'flex-1 flex flex-col items-center gap-0.5 transition-colors py-1 min-w-[40px] relative',
          isChatOpen ? 'text-[#40ffaa]' : 'tz-text-secondary hover:text-white'
        ]"
        @click="toggleChatFromDock()" 
        :aria-label="$t('dockMenu.chat')"
      >
        <Icon name="lucide:message-circle" class="w-6 h-6 md:w-6 md:h-6 transition-all" />
        <span class="text-[9px] md:text-xs font-medium tracking-tight">{{ $t('dockMenu.chat') }}</span>
        <!-- Unread Badge -->
        <span
          v-if="totalUnreadCount > 0"
          class="absolute top-0 right-1 md:right-4 w-2 h-2 md:w-2.5 md:h-2.5 bg-red-500 rounded-full border border-[#0b1020]"
        ></span>
      </button>

      <!-- 3. Checkout (Main Action) -->
      <button 
        class="h-10 px-2 mx-0.5 md:h-12 md:px-6 md:mx-2 rounded-full bg-gradient-to-r from-[#40ffaa] to-[#6b73ff] text-[#0b1020] font-bold text-xs md:text-sm flex items-center justify-center border border-black shadow-[3px_3px_2px_rgba(0,0,0,0.95)] hover:shadow-[4px_4px_3px_rgba(0,0,0,1)] transition-all transform hover:-translate-y-0.5 min-w-[64px] md:min-w-[100px]"
        @click="openCheckoutFromDock"
      >
        <span>{{ priceDisplay }}</span>
      </button>

      <!-- 4. Quick Buy -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 tz-text-secondary hover:text-[#40ffaa] transition-colors py-1 min-w-[40px]"
        @click="openQuick()" 
        aria-haspopup="dialog" 
        :aria-expanded="quickOpen" 
        :aria-label="$t('dockMenu.quickBuy')"
      >
        <Icon name="lucide:zap" class="w-6 h-6 md:w-6 md:h-6 transition-all" />
        <span class="text-[9px] md:text-xs font-medium tracking-tight">{{ $t('dockMenu.quick') }}</span>
      </button>

      <!-- 5. Cart -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 tz-text-secondary hover:text-white transition-colors py-1 min-w-[40px]"
        @click="openCartDrawer" 
        :aria-label="$t('dockMenu.openCart')"
      >
        <Icon name="lucide:shopping-cart" class="w-6 h-6 md:w-6 md:h-6 transition-all" />
        <span class="text-[9px] md:text-xs font-medium tracking-tight">{{ $t('dockMenu.cart') }}</span>
      </button>

      <!-- 6. Saved (Wishlist) -->
      <button 
        class="flex-1 flex flex-col items-center gap-0.5 tz-text-secondary hover:text-white transition-colors py-1 min-w-[40px] relative"
        @click="openWishlist" 
        :aria-label="$t('dockMenu.openWishlist')"
      >
        <Icon name="lucide:heart" class="w-6 h-6 md:w-6 md:h-6 transition-all" />
        <span class="text-[9px] md:text-xs font-medium tracking-tight">{{ $t('dockMenu.saved') }}</span>
        <span
          v-if="wishlistCount > 0"
          class="absolute top-0 right-1 md:right-5 w-2 h-2 md:w-2.5 md:h-2.5 bg-[#40ffaa] rounded-full border border-[#0b1020]"
        ></span>
      </button>

    </div>
  </div>
  
  <!-- Quick Buy Modal from Dock -->
  <QuickBuyModal v-if="quickOpen" :config="quickBuyConfig" @close="quickOpen = false" />

  <!-- Wishlist 抽屉弹窗 -->
  <WishlistDrawer v-model="wishlistDrawerVisible" />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n, useRuntimeConfig } from '#imports'
import QuickBuyModal from '@/components/QuickBuy.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import { useWishlist } from '~/composables/useWishlist'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuySettings } from '~/composables/usePublicSettings'

// floating submenu state
const isOpen = ref(false)
const quickOpen = ref(false)
const wishlistDrawerVisible = ref(false)

// 全局聊天窗口状态（在多个布局之间保持一致）
const { currentConversation, isChatOpen, openChat, closeChat } = useChatWidget()

// mutually exclusive open helpers
const closeAll = () => {
  isOpen.value = false
  quickOpen.value = false
  wishlistDrawerVisible.value = false
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
const { quickBuySettings } = useQuickBuySettings()
const quickBuyConfig = computed(() => props.config || quickBuySettings.value || null)

// i18n and runtime config
const runtimeConfig = useRuntimeConfig()
const { t: $t } = useI18n()

// 未读消息数（从 localStorage 跟踪）
const totalUnreadCount = ref(0)

// Dock 内部控制聊天开关：需要兼顾全局状态和现有事件
const toggleChatFromDock = () => {
  if (isChatOpen.value) {
    closeChat()
  } else {
    closeAll()
    openChat({ showAgentList: true })
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
    }
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
let unreadInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  calculateUnreadCount()
  
  // 每30秒更新一次未读消息数
  unreadInterval = setInterval(calculateUnreadCount, 30000)
})

// cart summary data
const summary = ref<CartResponse | null>(null)
const loading = ref(false)

const apiBase = computed(() => {
  const fromProp = quickBuyConfig.value?.storeApiBase?.replace(/\/$/, '')
  if (fromProp) return fromProp
  const fallback = (runtimeConfig.public as { storeApiBase?: string }).storeApiBase
  return fallback ? String(fallback).replace(/\/$/, '') : ''
})

const cartUrl = computed(() => {
  if (quickBuyConfig.value?.cartUrl) return quickBuyConfig.value.cartUrl
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

  if (unreadInterval) {
    clearInterval(unreadInterval)
  }
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
