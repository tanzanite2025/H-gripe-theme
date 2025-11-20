<template>
  <!-- 背景条 -->
  <div class="fixed left-0 right-0 bottom-0 w-full h-[72px] bg-[#0b1020]/70 backdrop-blur-md border-t border-white/5 z-[100] pointer-events-none"></div>
  
  <!-- Dock 菜单 -->
  <div class="fixed left-0 right-0 bottom-[8px] w-full flex items-center justify-center gap-[11px] max-md:gap-[5px] min-[769px]:max-[1024px]:grid min-[769px]:max-[1024px]:grid-cols-5 min-[769px]:max-[1024px]:gap-[7px] min-[769px]:max-[1024px]:justify-items-center min-[1025px]:max-[1200px]:grid min-[1025px]:max-[1200px]:grid-flow-col min-[1025px]:max-[1200px]:auto-cols-max min-[1025px]:max-[1200px]:gap-3 pb-[max(env(safe-area-inset-bottom),0px)] z-[101] pointer-events-none" ref="dockRef">
    <!-- 第一个按钮：客服聊天 -->
    <div class="relative inline-flex items-center">
      <button 
        class="pointer-events-auto text-[#cfd6ff] shadow-[0_4px_24px_rgba(0,0,0,0.35)] transition-[transform,box-shadow,background] duration-[180ms] ease-in-out hover:-translate-y-0.5 hover:shadow-[0_10px_28px_rgba(0,0,0,0.45)] focus-visible:-translate-y-0.5 focus-visible:shadow-[0_10px_28px_rgba(0,0,0,0.45)] w-[52px] h-[52px] max-md:w-[52px] max-md:h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020] text-white relative" 
        @click="openChatModal()" 
        aria-label="客服聊天"
      >
        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true" class="w-[56px] h-[56px]">
          <g fill="none">
            <path fill="#241d5e" d="M11.915 18.714c3.509 0 6.354-2.278 6.354-5.09c0-2.81-2.845-5.089-6.354-5.089s-6.354 2.28-6.354 5.09s2.844 5.09 6.354 5.09" />
            <path fill="#f5c249" d="M20.685 11.694c-.148-1.53-.643-6.215-.652-6.39c-.01-.203-.302-.464-.302-.09c0 .265-.297 3.042-.297 5.21h-.522c-.976-2.375-3.555-3.937-6.628-4.04h-.594c-3.074.099-5.634 1.665-6.606 4.04h-.54c0-2.168-.27-4.949-.27-5.205c0-.379-.293-.118-.297.085c-.01.175-.504 4.86-.653 6.39a1.62 1.62 0 0 0-.324.981c0 .567.378 1.17.837 1.17c.122.324.28.797.477 1.08h.747c1.004 2.453 3.726 4.05 6.921 4.05s5.909-1.598 6.912-4.05h.779c.198-.284.373-.756.49-1.08c.463 0 .837-.599.837-1.17a1.64 1.64 0 0 0-.32-.981zm-8.698 6.853c-3.321 0-6.012-2.142-6.012-4.792c0-1.98 1.507-3.681 3.654-4.41l.391.805a.45.45 0 0 0 .167.203a.45.45 0 0 0 .252.072h3.105a.45.45 0 0 0 .252-.072a.45.45 0 0 0 .166-.203l.392-.8c2.142.733 3.64 2.43 3.64 4.41c0 2.646-2.69 4.788-6.007 4.788" />
            <path fill="#f0f3fa" d="M14.03 12.225a.67.67 0 0 0-.477.207a.74.74 0 0 0-.198.504v1.728c0 .189.072.369.198.504a.65.65 0 0 0 .477.207c.18 0 .35-.077.477-.207a.74.74 0 0 0 .198-.504v-1.728a.74.74 0 0 0-.198-.504a.65.65 0 0 0-.477-.207m-4.05 0a.67.67 0 0 0-.477.207a.74.74 0 0 0-.198.504v1.728c0 .189.072.369.198.504a.65.65 0 0 0 .477.207c.18 0 .35-.077.477-.207a.74.74 0 0 0 .198-.504v-1.728a.74.74 0 0 0-.198-.504a.65.65 0 0 0-.477-.207m2.025 4.95a.9.9 0 0 1-.9-.9h.292c0 .315.27.576.608.576c.333 0 .607-.261.607-.576h.293a.9.9 0 0 1-.9.9" />
          </g>
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
      class="pointer-events-auto text-[#cfd6ff] shadow-[0_4px_24px_rgba(0,0,0,0.35)] transition-[transform,box-shadow,background] duration-[180ms] ease-in-out hover:-translate-y-0.5 hover:shadow-[0_10px_28px_rgba(0,0,0,0.45)] focus-visible:-translate-y-0.5 focus-visible:shadow-[0_10px_28px_rgba(0,0,0,0.45)] w-[52px] h-[52px] max-md:w-[52px] max-md:h-[52px] rounded-full inline-flex items-center justify-center bg-[#0b1020] text-white" 
      @click="openQuick()" 
      aria-haspopup="dialog" 
      :aria-expanded="quickOpen" 
      aria-label="Open quick buy"
    >
      <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" aria-hidden="true" class="w-9 h-9">
        <g fill="none">
          <path fill="url(#SVGa3lGXcfD)" d="M15.037 9.504c-.505-.042-.739-.06-1.213.204a2070 2070 0 0 0-6.72 3.72a4.3 4.3 0 0 1-1.074.468c-1.699.354-3.205-1.26-2.593-2.832c.312-.804.798-1.05 1.657-1.524q5.508-3.058 11.035-6.09c1.14-.625 2.226-.6 3.27.083a2.73 2.73 0 0 1 1.2 1.609c.072.276.108.78.102 1.5v11.875c0 1.242-.822 2.43-2.148 2.478c-.36.012-.72.012-1.05-.12a2.84 2.84 0 0 1-1.602-2.46V10.95c0-.618-.084-1.08-.63-1.38a.6.6 0 0 0-.235-.066" />
          <path fill="url(#SVGw2FHhcGq)" d="M10.674 20.551a7 7 0 0 1-3.126-1.626c-.198-.18-.42-.384-.468-.636c-.084-.42.228-.624.57-.816l2.988-1.668c.372-.21.852-.085 1.02.3c.043.09.043.24.043.462v3.156q0 .405-.109.576c-.198.318-.552.342-.918.252" />
          <path fill="url(#SVGHT3QCcQu)" d="M14.965 9.498c-.456-.042-.69-.042-1.14.21a2056 2056 0 0 0-6.721 3.72a4.3 4.3 0 0 1-1.074.468c-1.699.354-3.205-1.26-2.593-2.832c.222-.57.534-.858 1.002-1.152c3.18-1.038 8.281-2.112 10.526-.42z" />
          <path fill="url(#SVGT01Vfe6i)" d="m14.845 9.487l.192.018c.096.006.18.03.24.06c.48.27.6.624.618 1.128c.078-4.825-3.756-4.483-5.16-4.273l-5.64 3.12l-.085.049c-.21.114-.396.216-.558.318c3.15-1.038 8.173-2.1 10.393-.42" />
          <defs>
            <linearGradient id="SVGa3lGXcfD" x1="12" x2="12" y1="2.999" y2="21.001" gradientUnits="userSpaceOnUse">
              <stop stop-color="#4af78b" />
              <stop offset="1" stop-color="#23bfef" />
            </linearGradient>
            <linearGradient id="SVGw2FHhcGq" x1="4.905" x2="18.018" y1="8.691" y2="18.997" gradientUnits="userSpaceOnUse">
              <stop stop-color="#3cd1c6" />
              <stop offset="1" stop-color="#34c2d8" />
            </linearGradient>
            <linearGradient id="SVGHT3QCcQu" x1="15.241" x2="1.197" y1="3.008" y2="8.016" gradientUnits="userSpaceOnUse">
              <stop stop-color="#226a82" />
              <stop offset="1" stop-color="#55deb0" />
            </linearGradient>
            <linearGradient id="SVGT01Vfe6i" x1="12.474" x2="10.38" y1="6.319" y2="11.641" gradientUnits="userSpaceOnUse">
              <stop stop-color="#2a687f" />
              <stop offset="1" stop-color="#6cf7b6" />
            </linearGradient>
          </defs>
        </g>
      </svg>
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

