<template>
  <!-- Dock 菜单容器 (统一胶囊风格) -->
  <div class="dock-bar fixed inset-x-0 bottom-0 w-full z-[101] pointer-events-auto transition-all duration-300">
    <div class="dock-surface mx-auto w-full md:max-w-[500px] rounded-none px-1 py-2.5 md:px-4 md:py-3 items-center transition-all duration-300">
      
      <!-- 1. Menu (Sidebar) -->
      <button 
        class="dock-icon-button h-11 md:h-12 tz-text-secondary hover:text-white transition-colors"
        @click="openSidebarLeft"
        :aria-label="$t('dockMenu.openSidebar')"
      >
        <Icon name="lucide:user-round-check" class="dock-account-icon w-7 h-7 md:w-9 md:h-9 transition-all" />
      </button>

      <!-- 2. Chat -->
      <button 
        :class="[
          'dock-icon-button h-11 md:h-12 transition-colors',
          isChatOpen ? 'text-[#B5FF6D]' : 'tz-text-secondary hover:text-white'
        ]"
        @click="toggleChatFromDock()" 
        :aria-label="$t('dockMenu.chat')"
      >
        <span class="relative inline-flex h-7 w-7 md:h-9 md:w-9 items-center justify-center">
          <svg
            class="dock-chat-icon w-full h-full transition-all"
            viewBox="0 0 48 48"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <g transform="translate(24 24) scale(1.2) translate(-24 -24)">
              <path
                d="M31 12H15.5C12.46 12 10 14.46 10 17.5V30C10 33.04 12.46 35.5 15.5 35.5H20V42L28 35.5H37C40.04 35.5 42.5 33.04 42.5 30V21"
                stroke="currentColor"
                stroke-width="3.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                d="M29.5 18.5L40 8M33 8H40V15"
                stroke="currentColor"
                stroke-width="3.5"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </g>
          </svg>
          <!-- Unread Badge -->
          <span
            v-if="totalUnreadCount > 0"
            class="absolute top-0 right-0 w-2 h-2 md:w-2.5 md:h-2.5 bg-red-500 rounded-full border border-[#0b1020]"
          ></span>
        </span>
      </button>

      <!-- 3. Quick Buy -->
      <button 
        ref="quickBuyAnchorRef"
        class="dock-icon-button dock-quick-buy-button h-11 md:h-12 tz-text-secondary hover:text-[#B5FF6D] transition-colors"
        :class="{ 'dock-quick-buy-button--active': quickActive }"
        @click="openQuick()" 
        aria-haspopup="dialog" 
        :aria-expanded="quickActive"
        :aria-label="$t('dockMenu.quickBuy')"
      >
        <span class="dock-quick-buy-frame">
          <svg
            class="dock-quick-buy-icon w-7 h-7 md:w-9 md:h-9 transition-all"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
          >
            <title>Honeybadger</title>
            <path
              d="M11.999 0c-.346 0-.691.131-.955.395L.394 11.045a1.35 1.35 0 0 0 0 1.91l6.243 6.24.915-1.95L2.306 12l9.693-9.693 1.158 1.157 1.432-1.432L12.954.395A1.346 1.346 0 0 0 11.999 0Zm5.54 1.106a.331.331 0 0 0-.218.102l-1.777 1.778-1.432 1.432-8.393 8.392h4.726l-3.76 9.26c-.139.34.29.626.55.366l1.321-1.32v-.001l1.432-1.432h.001l8.56-8.561h-4.727l2.083-4.91v.001l.854-2.012 1.112-2.623c.108-.256-.108-.485-.333-.472Zm.25 4.125-.853 2.012 4.756 4.756L12 21.693l-1.056-1.055-1.432 1.432 1.533 1.534a1.35 1.35 0 0 0 1.91 0l10.65-10.65a1.35 1.35 0 0 0 0-1.91z"
              fill="currentColor"
            />
          </svg>
        </span>
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
          <span class="dock-cart-currency notranslate" translate="no" aria-hidden="true">
            <Icon :name="currencyIconName" class="h-full w-full" />
          </span>
          <span class="dock-cart-total notranslate" translate="no">{{ priceDisplay }}</span>
          <span class="dock-cart-count">{{ itemsCount }}</span>
        </span>
      </button>

    </div>
  </div>
  
  <!-- Quick Buy entry popover from Dock -->
  <LazyQuickBuyEntryRouterPopover
    v-if="quickOpen"
    :config="quickBuyConfig"
    :anchor="quickBuyAnchorRef"
    @close="closeQuickEntry"
    @direct-select="openQuickDirectSelect"
    @contact-service="openQuickContactService"
    @wheelset-selection-assistant="openQuickWheelsetSelectionAssistant"
  />

  <QuickBuyModal
    v-if="quickDirectSelectOpen"
    :config="quickBuyConfig"
    @close="closeQuickDirectSelect"
  />

  <QuickBuyContactServiceModal
    v-if="quickContactServiceOpen"
    @close="closeQuickContactService"
  />

  <WheelsetSelectionAssistantModal
    v-if="quickWheelsetSelectionAssistantOpen"
    :model-value="true"
    source="quick-buy/wheelset-selection-assistant"
    description=""
    @update:model-value="handleQuickWheelsetSelectionAssistantModelUpdate"
    @close="closeQuickWheelsetSelectionAssistant"
  >
    <WheelsetSelectionAssistantFlow
      source="quick-buy/wheelset-selection-assistant"
      @contact-support="openWheelsetSelectionSupportChat"
    />
  </WheelsetSelectionAssistantModal>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted, onBeforeUnmount, watchEffect } from 'vue'
import { useI18n } from '#imports'
import QuickBuyModal from '~/components/QuickBuy.vue'
import QuickBuyContactServiceModal from '~/components/quick-buy/QuickBuyContactServiceModal.vue'
import WheelsetSelectionAssistantModal from '~/components/WheelsetSelectionAssistantModal.vue'
import WheelsetSelectionAssistantFlow from '~/components/wheelset-selection/WheelsetSelectionAssistantFlow.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useQuickBuyFlow } from '~/composables/useQuickBuyFlow'
import type { WheelsetSelectionRequestDraft } from '~/types/wheelsetSelectionAssistant'

// floating submenu state
const isOpen = ref(false)
const quickOpen = ref(false)
const quickDirectSelectOpen = ref(false)
const quickContactServiceOpen = ref(false)
const quickWheelsetSelectionAssistantOpen = ref(false)
const quickBuyAnchorRef = ref<HTMLElement | null>(null)

// 全局聊天窗口状态（在多个布局之间保持一致）
const { currentConversation, isChatOpen, openChat, closeChat } = useChatWidget()

// mutually exclusive open helpers
const closeAll = () => {
  isOpen.value = false
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
}

const quickActive = computed(() =>
  quickOpen.value || quickDirectSelectOpen.value || quickContactServiceOpen.value || quickWheelsetSelectionAssistantOpen.value,
)

const openQuickEntry = () => {
  closeAll()
  quickOpen.value = true
}

const openQuick = () => {
  if (quickActive.value) {
    closeAll()
    return
  }

  openQuickEntry()
}

const openQuickFromGlobalEvent = () => {
  openQuickEntry()
}

const closeQuickEntry = () => {
  quickOpen.value = false
}

const openQuickDirectSelect = () => {
  quickOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
  quickDirectSelectOpen.value = true
}

const closeQuickDirectSelect = () => {
  quickDirectSelectOpen.value = false
}

const openQuickContactService = () => {
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = false
  quickContactServiceOpen.value = true
}

const closeQuickContactService = () => {
  quickContactServiceOpen.value = false
}

const openQuickWheelsetSelectionAssistant = () => {
  quickOpen.value = false
  quickDirectSelectOpen.value = false
  quickContactServiceOpen.value = false
  quickWheelsetSelectionAssistantOpen.value = true
}

const closeQuickWheelsetSelectionAssistant = () => {
  quickWheelsetSelectionAssistantOpen.value = false
}

const handleQuickWheelsetSelectionAssistantModelUpdate = (value: boolean) => {
  if (!value) {
    closeQuickWheelsetSelectionAssistant()
  }
}

const openWheelsetSelectionSupportChat = async (draft?: WheelsetSelectionRequestDraft) => {
  closeQuickWheelsetSelectionAssistant()
  openChat({
    showAgentList: true,
    source: 'wheelset-selection-assistant',
    pendingSelectionRequest: draft || null,
  })
  await nextTick()
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
  window.addEventListener('quickbuy:open-entry', openQuickFromGlobalEvent)
  
  // 每30秒更新一次未读消息数
  unreadInterval = setInterval(calculateUnreadCount, 30000)
})

// 集成购物车系统
const { cartCount, total, cartCurrency, openCart } = useCart()

const itemsCount = computed(() => cartCount.value)

const priceDisplay = computed(() => {
  try {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: cartCurrency.value || 'USD',
    }).formatToParts(total.value)
      .filter(part => part.type !== 'currency' && part.type !== 'literal')
      .map(part => part.value)
      .join('')
      .trim()
  } catch {
    return Number(total.value || 0).toFixed(2)
  }
})

const currencyIconName = computed(() => {
  switch (cartCurrency.value) {
    case 'EUR':
      return 'lucide:badge-euro'
    case 'GBP':
      return 'lucide:badge-pound-sterling'
    case 'JPY':
    case 'CNY':
      return 'lucide:badge-japanese-yen'
    case 'INR':
      return 'lucide:badge-indian-rupee'
    case 'CHF':
      return 'lucide:badge-swiss-franc'
    case 'TRY':
      return 'lucide:badge-turkish-lira'
    case 'RUB':
      return 'lucide:badge-russian-ruble'
    case 'USD':
    case 'AUD':
    case 'CAD':
    case 'HKD':
    case 'NZD':
    case 'SGD':
      return 'lucide:badge-dollar-sign'
    default:
      return 'lucide:banknote'
  }
})

const cartActionAriaLabel = computed(() => {
  const openCartLabel = String($t('dockMenu.openCart'))
  if (itemsCount.value <= 0) return openCartLabel
  return `${openCartLabel}: ${itemsCount.value} ${String($t('cart.summary.items', 'Items'))}, ${priceDisplay.value} ${cartCurrency.value || 'USD'}`
})

const openCartDrawer = () => {
  openCart()
}

onBeforeUnmount(() => {
  if (unreadInterval) {
    clearInterval(unreadInterval)
  }
  window.removeEventListener('quickbuy:open-entry', openQuickFromGlobalEvent)
})

// defensive: ensure mutual exclusivity if any state is toggled externally
watchEffect(() => {
  const openCount = [isOpen.value, quickActive.value].filter(Boolean).length
  if (openCount > 1) {
    // prefer the most recently opened by simple priority: quick > fab
    if (quickActive.value) {
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
  background: linear-gradient(180deg, #0c0c10 0%, #111116 42%, #1a1a20 100%);
  box-shadow:
    0 -12px 32px rgba(0, 0, 0, 0.28),
    inset 0 1px 0 rgba(255, 255, 255, 0.035);
}

.dock-surface {
  display: grid;
  grid-template-columns: repeat(3, minmax(3rem, 1fr)) minmax(9.25rem, 1.85fr);
  column-gap: 0.25rem;
  max-width: min(100%, 500px);
}

.dock-icon-button {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
}

.dock-quick-buy-button {
  --dock-quickbuy-active-edge: color-mix(in srgb, var(--tz-brand-primary, #b5ff6d) 74%, transparent);
}

.dock-quick-buy-frame {
  display: grid;
  width: 2.5rem;
  height: 2.5rem;
  place-items: center;
  border-radius: 999px;
  transition:
    background-color 180ms ease,
    box-shadow 180ms ease,
    color 180ms ease,
    transform 180ms ease;
}

.dock-quick-buy-button--active {
  color: var(--tz-brand-primary, #b5ff6d);
}

.dock-quick-buy-button--active .dock-quick-buy-frame {
  background: rgba(181, 255, 109, 0.08);
  box-shadow:
    inset 0 0 0 1px var(--dock-quickbuy-active-edge),
    0 0 0 4px rgba(181, 255, 109, 0.075);
}

.dock-quick-buy-button:hover .dock-quick-buy-frame {
  transform: translateY(-0.0625rem);
}

.dock-cart-button {
  display: flex;
  width: 100%;
  min-width: 0;
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

.dock-cart-currency {
  display: inline-flex;
  width: 1.25rem;
  height: 1.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
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
  min-width: 1.875rem;
  height: 1.875rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 999px;
  background: var(--tz-brand-primary, #b5ff6d);
  color: #050505;
  font-size: 1rem;
  font-weight: 900;
  padding: 0 0.35rem;
}

.dock-cart-total {
  min-width: 0;
  max-width: 7rem;
  overflow: hidden;
  font-size: 1.2rem;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (min-width: 768px) {
  .dock-quick-buy-frame {
    width: 3rem;
    height: 3rem;
  }

  .dock-cart-button {
    height: 2.75rem;
  }

  .dock-cart-total {
    max-width: 7.75rem;
    font-size: 1.3rem;
  }

  .dock-cart-currency {
    width: 1.35rem;
    height: 1.35rem;
  }
}

@media (max-width: 767px) {
  .dock-bar {
    background: linear-gradient(180deg, #303034 0%, #2b2b2f 22%, #242428 42%, #1d1d21 62%, #18181c 82%, #151519 100%);
  }

  .dock-surface {
    grid-template-columns: repeat(3, minmax(2.75rem, 1fr)) minmax(9.15rem, 1.85fr);
    column-gap: 0.25rem;
    padding-bottom: max(0.625rem, calc(0.625rem + var(--tz-safe-area-bottom, 0px)));
  }

  .dock-cart-button {
    height: 2.5rem;
    padding-inline: 0.4rem;
  }

  .dock-cart-content {
    gap: 0.3rem;
  }

  .dock-cart-total {
    max-width: 5.75rem;
    font-size: 1.05rem;
  }
}

@media (max-width: 360px) {
  .dock-surface {
    grid-template-columns: repeat(3, minmax(2.55rem, 1fr)) minmax(8.65rem, 1.9fr);
    column-gap: 0.125rem;
  }

  .dock-cart-button {
    padding-inline: 0.35rem;
  }

  .dock-cart-content {
    gap: 0.25rem;
  }

  .dock-cart-currency {
    width: 1.15rem;
    height: 1.15rem;
  }

  .dock-cart-count {
    min-width: 1.75rem;
    height: 1.75rem;
    font-size: 0.95rem;
    padding-inline: 0.3rem;
  }

  .dock-cart-total {
    max-width: 5.25rem;
    font-size: 1rem;
  }
}
</style>
