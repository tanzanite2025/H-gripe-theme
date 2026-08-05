<template>
  <div class="flex flex-col h-full min-h-0">
    <!-- Container -->
    <div class="flex flex-col h-full min-h-0">
      <!-- 二级导航栏 -->
      <div class="flex-none px-2 pt-3 pb-2 md:py-3 md:px-4 md:border-b md:border-white/[0.08] md:bg-white/[0.02]">
        <div class="grid grid-cols-5 md:flex md:flex-wrap gap-1 md:gap-2 justify-center">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="handleTabClick(tab.id)"
            class="relative h-9 w-9 md:h-8 md:w-8 rounded-full border transition-all flex items-center justify-center"
            :class="[
              activeTab === tab.id
                ? 'border-white bg-white text-slate-950 shadow-[0_4px_12px_rgba(0,0,0,0.55)]'
                : 'border-white/15 bg-white/[0.08] text-white shadow-[0_3px_9px_rgba(0,0,0,0.55)] hover:border-[#B5FF6D]/60 hover:bg-white/[0.14]',
              tab.id === 'chat' && activeTab !== 'chat' ? 'chat-return-attention' : ''
            ]"
            :title="t(tab.labelKey)"
            :aria-label="t(tab.labelKey)"
          >
            <Icon :name="tab.icon" class="h-4 w-4" />
          </button>
        </div>
      </div>

      <!-- 内容区域 -->
      <div class="flex-1 min-h-0 overflow-hidden relative">
        <!-- 聊天 Tab -->
        <ChatTab
          v-if="activeTab === 'chat'"
          :messages="messages"
          :new-message="newMessage"
          :visitor-email="visitorEmail"
          :show-visitor-email-capture="showVisitorEmailCapture"
          :is-sending="isSending"
          :is-uploading-image="isUploadingImage"
          :pending-product-reference="pendingProductReference"
          :agent-typing="agentTyping"
          :current-theme-color="currentThemeColor"
          @update:new-message="$emit('update:newMessage', $event)"
          @update:visitor-email="$emit('update:visitorEmail', $event)"
          @send-message="$emit('sendMessage')"
          @upload-image="handleUploadImage"
          @open-order-picker="$emit('openOrderPicker')"
          @open-customer-service-product-search-modal="$emit('openCustomerServiceProductSearchModal')"
          @clear-pending-product-reference="$emit('clearPendingProductReference')"
          @delete-message="$emit('deleteMessage', $event)"
        />

        <!-- 订单 Tab -->
        <OrderTab
          v-else-if="activeTab === 'orders'"
          :orders-list="ordersList"
          :is-loading-orders="isLoadingOrders"
          @share-order="$emit('shareOrder', $event)"
        />

        <!-- 会员 Tab -->
        <MemberTab
          v-else-if="activeTab === 'member'"
          :is-member-logged="isMemberLogged"
          :level-name="levelName"
          :points="points"
          :tier-info="tierInfo"
          :level-discounts="levelDiscounts"
          :user-coupons="userCoupons"
          :user-point-cards="userPointCards"
          @open-auth="$emit('openAuth', $event)"
        />

        <!-- 保修 Tab -->
        <WarrantyTab
          v-else-if="activeTab === 'warranty'"
          :is-logged-in="isLoggedInForWarranty"
          @login-request="$emit('loginRequest')"
        />

        <!-- Calculator Tab -->
        <div v-else-if="activeTab === 'calculator'" class="h-full overflow-y-auto p-4 custom-scrollbar">
          <SpokeSmartSearch />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import ChatTab from './ChatTab.vue'
import OrderTab from './OrderTab.vue'
import MemberTab from './MemberTab.vue'
import WarrantyTab from './WarrantyTab.vue'
import SpokeSmartSearch from '~/components/SpokeSmartSearch.vue'

const { t } = useI18n()

const props = defineProps<{
  activeTab: string
  currentThemeColor: string
  // Chat Props
  messages: any[]
  newMessage: string
  visitorEmail: string
  showVisitorEmailCapture: boolean
  isSending: boolean
  isUploadingImage: boolean
  pendingProductReference?: any | null
  agentTyping?: { active: boolean; displayName?: string }
  // Order Props
  ordersList: any[]
  isLoadingOrders: boolean
  // Member Props
  isMemberLogged: boolean
  levelName: string | number
  points: number | string
  tierInfo: any
  levelDiscounts: any
  userCoupons: number
  userPointCards: number
  // Warranty Props
  isLoggedInForWarranty: boolean
}>()

const emit = defineEmits<{
  'update:activeTab': [value: string]
  // Chat Emits
  'update:newMessage': [value: string]
  'update:visitorEmail': [value: string]
  'sendMessage': []
  'uploadImage': [event: Event, source: 'library' | 'camera']
  'openOrderPicker': []
  'openCustomerServiceProductSearchModal': []
  'clearPendingProductReference': []
  'deleteMessage': [message: any]
  // Order Emits
  'shareOrder': [order: any]
  // Member Emits
  'openAuth': [mode: 'login' | 'register']
  // Warranty Emits
  'loginRequest': []
}>()

const tabs = computed(() => [
  { id: 'chat', labelKey: 'chatModal.actions.switchToChat', icon: 'lucide:message-circle' },
  { id: 'orders', labelKey: 'chatModal.tabs.orders', icon: 'lucide:receipt-text' },
  { id: 'warranty', labelKey: 'chatModal.tabs.warranty', icon: 'lucide:shield-check' },
  { id: 'member', labelKey: 'chatModal.tabs.member', icon: 'lucide:user-round' },
  { id: 'calculator', labelKey: 'chatModal.tabs.calculator', icon: 'lucide:calculator' },
])

const availableTabIds = computed(() => new Set(['chat', ...tabs.value.map((tab) => tab.id)]))

watch(
  () => props.activeTab,
  (activeTab) => {
    if (!availableTabIds.value.has(activeTab)) {
      emit('update:activeTab', 'chat')
    }
  },
  { immediate: true },
)

const handleTabClick = (id: string) => {
  emit('update:activeTab', id)
}

const handleUploadImage = (event: Event, source: 'library' | 'camera') => {
  emit('uploadImage', event, source)
}
</script>

<style scoped>
@keyframes chat-return-halo {
  0%,
  100% {
    opacity: 0;
    box-shadow: 0 0 0 0 rgba(181, 255, 109, 0);
  }

  18% {
    opacity: 1;
    box-shadow:
      0 0 0 1px rgba(181, 255, 109, 0.28),
      0 0 10px rgba(181, 255, 109, 0.18);
  }

  64% {
    opacity: 0.18;
    box-shadow:
      0 0 0 5px rgba(181, 255, 109, 0),
      0 0 14px rgba(181, 255, 109, 0.1);
  }
}

.chat-return-attention {
  isolation: isolate;
}

.chat-return-attention::after {
  content: '';
  position: absolute;
  inset: -3px;
  z-index: -1;
  border-radius: 9999px;
  pointer-events: none;
  animation: chat-return-halo 2.4s ease-out infinite;
}
</style>
