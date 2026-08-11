<template>
  <!-- Dock 菜单容器 (统一胶囊风格) -->
  <div class="dock-bar fixed inset-x-0 bottom-0 w-full z-[101] pointer-events-auto transition-all duration-300">
    <div class="dock-surface mx-auto w-full md:max-w-[500px] rounded-none px-1 py-2.5 md:px-4 md:py-3 flex items-center justify-between transition-all duration-300">
      
      <!-- 1. Menu (Sidebar) -->
      <button 
        class="flex-1 h-11 md:h-12 flex items-center justify-center tz-text-secondary hover:text-white transition-colors min-w-[40px]"
        @click="openSidebarLeft"
        :aria-label="$t('dockMenu.openSidebar')"
      >
        <Icon name="lucide:circle-user-round" class="w-7 h-7 md:w-9 md:h-9 transition-all" />
      </button>

      <!-- 2. Chat -->
      <button 
        :class="[
          'flex-1 h-11 md:h-12 flex items-center justify-center transition-colors min-w-[40px]',
          isChatOpen ? 'text-[#B5FF6D]' : 'tz-text-secondary hover:text-white'
        ]"
        @click="toggleChatFromDock()" 
        :aria-label="$t('dockMenu.chat')"
      >
        <span class="relative inline-flex h-7 w-7 md:h-9 md:w-9 items-center justify-center">
          <Icon name="lucide:message-circle" class="w-full h-full transition-all" />
          <!-- Unread Badge -->
          <span
            v-if="totalUnreadCount > 0"
            class="absolute top-0 right-0 w-2 h-2 md:w-2.5 md:h-2.5 bg-red-500 rounded-full border border-[#0b1020]"
          ></span>
        </span>
      </button>

      <!-- 3. Quick Buy -->
      <button 
        class="flex-1 h-11 md:h-12 flex items-center justify-center tz-text-secondary hover:text-[#B5FF6D] transition-colors min-w-[40px]"
        @click="openQuick()" 
        aria-haspopup="dialog" 
        :aria-expanded="quickOpen" 
        :aria-label="$t('dockMenu.quickBuy')"
      >
        <Icon name="lucide:zap" class="w-7 h-7 md:w-9 md:h-9 transition-all" />
      </button>

      <!-- 4. Cart -->
      <button
        class="dock-cart-button"
        type="button"
        :class="{ 'dock-cart-button--active': itemsCount > 0 }"
        @click="openCartDrawer"
        :aria-label="cartActionAriaLabel"
      >
        <span class="dock-cart-content">
          <span class="dock-cart-total">{{ priceDisplay }}</span>
          <span class="dock-cart-count">{{ itemsCount }}</span>
        </span>
      </button>

    </div>
  </div>
  
  <!-- Quick Buy Modal from Dock -->
  <QuickBuyModal v-if="quickOpen" :config="quickBuyConfig" @close="quickOpen = false" />
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n } from '#imports'
import QuickBuyModal from '@/components/QuickBuy.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuyFlow } from '~/composables/useQuickBuyFlow'

// floating submenu state
const isOpen = ref(false)
const quickOpen = ref(false)

// 全局聊天窗口状态（在多个布局之间保持一致）
const { currentConversation, isChatOpen, openChat, closeChat } = useChatWidget()

// mutually exclusive open helpers
const closeAll = () => {
  isOpen.value = false
  quickOpen.value = false
}


const openQuick = () => {
  closeAll()
  quickOpen.value = true
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'dock-quick' } }))
  }
}

// removed old share popup and outside-click listeners; modal closes by overlay click

const { quickBuyFlowConfig } = useQuickBuyFlow('dock')
const quickBuyConfig = computed(() => quickBuyFlowConfig.value)

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

// 集成购物车系统
const { cartCount, total, openCart, formatPrice } = useCart()

const itemsCount = computed(() => cartCount.value)

const priceDisplay = computed(() => {
  return formatPrice(total.value)
})

const cartActionAriaLabel = computed(() => {
  const openCartLabel = String($t('dockMenu.openCart'))
  if (itemsCount.value <= 0) return openCartLabel
  return `${openCartLabel}: ${itemsCount.value} ${String($t('cart.summary.items', 'Items'))}, ${priceDisplay.value}`
})

const openCartDrawer = () => {
  openCart()
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('ui:popup-open', { detail: { id: 'cart-drawer' } }))
  }
}

onMounted(() => {
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

<style scoped>
.dock-bar {
  min-height: var(--tz-bottom-dock-height, 4.5rem);
  background: rgba(17, 17, 22, 0.78);
  background: color-mix(in srgb, var(--tz-card-surface) 78%, transparent);
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  -webkit-backdrop-filter: blur(18px) saturate(120%);
  backdrop-filter: blur(18px) saturate(120%);
  box-shadow: 0 -12px 36px rgba(0, 0, 0, 0.18);
}

.dock-surface {
  max-width: min(100%, 500px);
  background: transparent;
  -webkit-backdrop-filter: none;
  backdrop-filter: none;
}

.dock-cart-button {
  display: flex;
  flex: 1.18 1 0;
  min-width: 6.8rem;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  padding-inline: 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 999px;
  background: #ffffff;
  color: #050505;
  font-weight: 900;
  line-height: 1;
  white-space: nowrap;
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.82);
  transition:
    color 180ms ease,
    border-color 180ms ease,
    background-color 180ms ease,
    transform 180ms ease;
}

.dock-cart-content {
  display: inline-flex;
  min-width: 0;
  max-width: 100%;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
}

.dock-cart-button:hover {
  background: #f8fafc;
  color: #050505;
  transform: translateY(-0.125rem);
}

.dock-cart-button--active {
  border-color: #ffffff;
  background: #ffffff;
  color: #050505;
}

.dock-cart-button--active:hover {
  background: #f8fafc;
  color: #050505;
}

.dock-cart-count {
  display: inline-flex;
  width: 1.65rem;
  min-width: 1.65rem;
  height: 1.65rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: var(--tz-brand-primary, #b5ff6d);
  color: #050505;
  font-size: 0.8rem;
  font-weight: 900;
  padding: 0;
}

.dock-cart-total {
  min-width: 0;
  max-width: 6rem;
  overflow: hidden;
  font-size: 1.12rem;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (min-width: 768px) {
  .dock-cart-button {
    height: 2.75rem;
    min-width: 7.4rem;
  }

  .dock-cart-total {
    max-width: 6.5rem;
    font-size: 1.22rem;
  }
}

@media (max-width: 767px) {
  .dock-surface {
    background: transparent !important;
    padding-bottom: max(0.625rem, calc(0.625rem + var(--tz-safe-area-bottom, 0px)));
  }

  .dock-cart-button {
    min-width: 6.35rem;
    height: 2.5rem;
    padding-inline: 0.45rem;
  }

  .dock-cart-content {
    gap: 0.3rem;
  }

  .dock-cart-total {
    max-width: 4.4rem;
    font-size: 1rem;
  }
}
</style>
