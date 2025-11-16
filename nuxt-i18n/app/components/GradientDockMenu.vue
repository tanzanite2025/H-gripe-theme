<template>
  <!-- 背景条 -->
  <div class="fixed left-0 right-0 bottom-0 w-full h-[72px] bg-[#0b1020]/70 backdrop-blur-md border-t border-white/5 z-[100] pointer-events-none"></div>
  
  <!-- Dock 菜单 -->
  <div class="fixed left-0 right-0 bottom-[8px] w-full flex items-center justify-center gap-[11px] max-md:gap-[5px] min-[769px]:max-[1024px]:grid min-[769px]:max-[1024px]:grid-cols-5 min-[769px]:max-[1024px]:gap-[7px] min-[769px]:max-[1024px]:justify-items-center min-[1025px]:max-[1200px]:grid min-[1025px]:max-[1200px]:grid-flow-col min-[1025px]:max-[1200px]:auto-cols-max min-[1025px]:max-[1200px]:gap-3 pb-[max(env(safe-area-inset-bottom),0px)] z-[101] pointer-events-none" ref="dockRef">
    <!-- 第一个按钮：客服聊天 -->
    <div class="relative inline-flex items-center">
      <button 
        class="pointer-events-auto text-[#cfd6ff] bg-[rgba(31,41,55,0.9)] shadow-[0_4px_24px_rgba(0,0,0,0.35)] transition-[transform,box-shadow,background] duration-[180ms] ease-in-out hover:-translate-y-0.5 hover:shadow-[0_10px_28px_rgba(0,0,0,0.45)] focus-visible:-translate-y-0.5 focus-visible:shadow-[0_10px_28px_rgba(0,0,0,0.45)] w-12 h-12 max-md:w-12 max-md:h-12 rounded-full inline-flex items-center justify-center bg-[#0b1020] text-white relative border-2 border-transparent bg-clip-padding [background-image:linear-gradient(#0b1020,#0b1020),linear-gradient(to_right,#40ffaa,#6b73ff)] [background-origin:border-box] [background-clip:padding-box,border-box]" 
        @click="openChatModal()" 
        aria-label="客服聊天"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 24 24" aria-hidden="true" fill="none" stroke="#25D366" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
        </svg>
        <!-- 未读消息徽章 -->
        <span
          v-if="totalUnreadCount > 0"
          class="absolute -top-1 -right-1 w-5 h-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center font-bold"
        >
          {{ totalUnreadCount > 9 ? '9+' : totalUnreadCount }}
        </span>
      </button>
    </div>
    
    <button 
      class="pointer-events-auto text-[#cfd6ff] bg-white/10 shadow-[0_4px_24px_rgba(0,0,0,0.35)] transition-[transform,box-shadow,background] duration-[180ms] ease-in-out hover:-translate-y-0.5 hover:shadow-[0_10px_28px_rgba(0,0,0,0.45)] focus-visible:-translate-y-0.5 focus-visible:shadow-[0_10px_28px_rgba(0,0,0,0.45)] min-h-[52px] h-[52px] max-md:h-12 max-md:min-h-12 w-[250px] px-3.5 rounded-full font-semibold tracking-wider text-white uppercase bg-[#0b1020] shadow-none hover:shadow-none focus-visible:shadow-none inline-flex items-center justify-between border-2 border-transparent bg-clip-padding [background-image:linear-gradient(#0b1020,#0b1020),linear-gradient(to_right,#40ffaa,#6b73ff)] [background-origin:border-box] [background-clip:padding-box,border-box]" 
      type="button" 
      @click="openCartDrawer" 
      :aria-label="ctaLabel"
    >
      <span class="inline-flex items-center after:content-['•'] after:opacity-70 after:text-xs after:ml-1.5">
        <span class="inline-flex flex-col items-center justify-center gap-0.5">
          <svg aria-hidden="true" viewBox="0 0 24 24" class="w-4 h-4 block"><g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M6 6h13l-1.2 8H8.2L6 6z"/><circle cx="9" cy="19" r="1.2"/><circle cx="17" cy="19" r="1.2"/></g></svg>
          <span class="normal-case leading-none text-[13px]">{{ itemsCount }}</span>
        </span>
      </span>
      <span class="inline-flex items-center after:content-['•'] after:opacity-70 after:text-xs after:ml-1.5">
        <span class="inline-flex flex-col items-center justify-center gap-0.5">
          <svg aria-hidden="true" viewBox="0 0 24 24" class="w-4 h-4 block"><g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 18h16"/><path d="M7 18l2-8h6l2 8"/><path d="M9 10a3 3 0 0 1 6 0"/></g></svg>
          <span class="normal-case leading-none text-[13px]">{{ weightDisplay }}</span>
        </span>
      </span>
      <span class="inline-flex items-center">
        <span class="inline-flex flex-col items-center justify-center gap-0.5">
          <svg aria-hidden="true" viewBox="0 0 24 24" class="w-4 h-4 block"><g fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M7 7h8l4 4-8 8-4-4z"/><path d="M10 10h.01"/></g></svg>
          <span class="normal-case leading-none text-[13px]">{{ priceDisplay }}</span>
        </span>
      </span>
    </button>
    <button 
      class="pointer-events-auto text-[#cfd6ff] bg-white/10 shadow-[0_4px_24px_rgba(0,0,0,0.35)] transition-[transform,box-shadow,background] duration-[180ms] ease-in-out hover:-translate-y-0.5 hover:shadow-[0_10px_28px_rgba(0,0,0,0.45)] focus-visible:-translate-y-0.5 focus-visible:shadow-[0_10px_28px_rgba(0,0,0,0.45)] w-12 h-12 max-md:w-12 max-md:h-12 rounded-full inline-flex items-center justify-center bg-[#0b1020] text-white border-2 border-transparent bg-clip-padding [background-image:linear-gradient(#0b1020,#0b1020),linear-gradient(to_right,#40ffaa,#6b73ff)] [background-origin:border-box] [background-clip:padding-box,border-box]" 
      @click="openQuick()" 
      aria-haspopup="dialog" 
      :aria-expanded="quickOpen" 
      aria-label="Open quick buy"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="30" height="30" viewBox="0 0 16 16" aria-hidden="true" class="fill-current"><g fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="#c6a0f6" d="m11 5.5l3.5-1l-6.5 11l-6.5-11l3.5 1" stroke-width="1"/><path stroke="#eed49f" d="m6 1.5l-.5 5l2-1l-1 3L8 8v3l4-7.5l-2 .5L11.5.5Z" stroke-width="1"/></g></svg>
    </button>
    
  </div>
  
  <!-- Quick Buy Modal from Dock -->
  <QuickBuyModal v-if="quickOpen" :config="props.config || null" @close="quickOpen = false" />
  
  <!-- WhatsApp 聊天弹窗 -->
  <WhatsAppChatModal 
    v-if="currentConversation"
    :conversation="currentConversation"
    @close="handleCloseChat"
  />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n, useRuntimeConfig } from '#imports'
import QuickBuyModal from '@/components/QuickBuy.vue'
import WhatsAppChatModal from '~/components/WhatsAppChatModal.vue'

// floating submenu state
const isOpen = ref(false)
const quickOpen = ref(false)
const currentConversation = ref<any>(null)

// mutually exclusive open helpers
const closeAll = () => {
  isOpen.value = false
  quickOpen.value = false
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

// 直接打开聊天窗口（WhatsAppChatModal 内部会显示客服列表）
const openChatModal = () => {
  closeAll()
  // 打开聊天窗口，不需要预先选择客服
  currentConversation.value = { showAgentList: true }
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'whatsapp-chat' } }))
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
const weightLabel = computed(() => $t('cart.summary.weight', 'Weight'))
const priceLabel = computed(() => $t('cart.summary.price', 'Price'))
const ctaLabel = computed(() => $t('cart.summary.openCart', 'View cart summary'))

// 集成购物车系统
const { cartCount, cartItems, total, openCart, formatPrice } = useCart()

const itemsCount = computed(() => cartCount.value)

const weightDisplay = computed(() => {
  const weight = cartItems.value.reduce((sum: number, item: any) => {
    return sum + (item.weight || 0) * item.quantity
  }, 0)
  return `${weight.toFixed(1)} g`
})

const priceDisplay = computed(() => {
  return formatPrice(total.value)
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

