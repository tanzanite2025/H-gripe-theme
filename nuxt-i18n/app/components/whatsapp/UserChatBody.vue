<template>
  <div class="flex flex-col h-full min-h-0">
    <!-- 移动端样式容器 / 桌面端普通容器 -->
    <div 
      class="flex flex-col h-full min-h-0 md:rounded-none md:border-none md:bg-transparent md:shadow-none transition-all duration-300"
      :class="!isDesktop ? 'rounded-[28px] border-2 overflow-hidden' : ''"
      :style="!isDesktop ? mobilePanelStyle : {}"
    >
      <!-- 二级导航栏 (Products, Orders, etc.) - 不包含 Chat -->
      <div class="flex-none px-2 pt-3 pb-2 md:py-3 md:px-4 md:border-b md:border-white/[0.08] md:bg-white/[0.02]">
        <div class="flex flex-wrap gap-1 md:gap-2 justify-center">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="$emit('update:activeTab', tab.id)"
            class="flex-1 md:flex-none h-10 md:h-8 md:px-4 rounded-full text-[11px] md:text-xs font-semibold transition-all whitespace-nowrap flex items-center justify-center"
            :class="activeTab === tab.id
              ? 'bg-[linear-gradient(135deg,#2dd4bf_0%,#3b82f6_100%)] text-white shadow-[0_4px_12px_rgba(45,212,191,0.3)]'
              : 'bg-[rgba(31,41,55,0.9)] text-white shadow-[0_3px_9px_rgba(0,0,0,0.9)] hover:bg-[rgba(51,65,85,0.95)]'"
          >
            {{ tab.label }}
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
          :is-sending="isSending"
          :is-uploading-image="isUploadingImage"
          :current-theme-color="currentThemeColor"
          @update:new-message="$emit('update:newMessage', $event)"
          @send-message="$emit('sendMessage')"
          @upload-image="$emit('uploadImage', $event)"
          @delete-message="$emit('deleteMessage', $event)"
        />

        <!-- 商品 Tab -->
        <ProductTab
          v-else-if="activeTab === 'share'"
          :search-query="searchQuery"
          :is-searching="isSearching"
          :search-results="searchResults"
          :product-drawer-visible="productDrawerVisible"
          :current-theme-color="currentThemeColor"
          @update:search-query="$emit('update:searchQuery', $event)"
          @search="$emit('search')"
          @share-product="$emit('shareProduct', $event)"
          @open-history="$emit('openHistory')"
          @open-cart="$emit('openCart')"
          @open-wishlist="$emit('openWishlist')"
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

        <!-- FAQ Tab -->
        <FaqTab v-else-if="activeTab === 'faq'" />

        <!-- 保修 Tab -->
        <WarrantyTab
          v-else-if="activeTab === 'warranty'"
          :is-logged-in="isLoggedInForWarranty"
          @login-request="$emit('loginRequest')"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import ChatTab from './ChatTab.vue'
import ProductTab from './ProductTab.vue'
import OrderTab from './OrderTab.vue'
import MemberTab from './MemberTab.vue'
import FaqTab from './FaqTab.vue'
import WarrantyTab from './WarrantyTab.vue'

const props = defineProps<{
  activeTab: string
  currentThemeColor: string
  // Chat Props
  messages: any[]
  newMessage: string
  isSending: boolean
  isUploadingImage: boolean
  // Product Props
  searchQuery: string
  isSearching: boolean
  searchResults: any[]
  productDrawerVisible: boolean
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

defineEmits<{
  'update:activeTab': [value: string]
  // Chat Emits
  'update:newMessage': [value: string]
  'sendMessage': []
  'uploadImage': [event: Event]
  'deleteMessage': [message: any]
  // Product Emits
  'update:searchQuery': [value: string]
  'search': []
  'shareProduct': [product: any]
  'openHistory': []
  'openCart': []
  'openWishlist': []
  // Order Emits
  'shareOrder': [order: any]
  // Member Emits
  'openAuth': [mode: 'login' | 'register']
  // Warranty Emits
  'loginRequest': []
}>()

const tabs = [
  { id: 'share', label: 'Products' },
  { id: 'orders', label: 'Orders' },
  { id: 'faq', label: 'FAQ' },
  { id: 'warranty', label: 'Warranty' },
  { id: 'member', label: 'Member' },
]

const isDesktop = ref(false)
const checkDesktop = () => {
  isDesktop.value = window.innerWidth >= 768
}

onMounted(() => {
  checkDesktop()
  window.addEventListener('resize', checkDesktop)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkDesktop)
})

const mobilePanelStyle = computed(() => {
  const color = props.currentThemeColor
  return {
    borderColor: color,
    background: `linear-gradient(180deg, ${color}33 0%, rgba(0,0,0,0.85) 100%)`,
    boxShadow: `0 15px 40px ${color}40`
  }
})
</script>
