<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="fixed inset-0 z-[10000] flex items-center justify-center md:justify-end p-0 md:p-6 pointer-events-none"
      >
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm md:hidden pointer-events-auto"></div>
        <!-- 聊天窗口容器 - 右下角定位 -->
        
        <!-- 客服模式 / 访客模式 切换 -->
        <Transition name="fade-scale" mode="out-in">
          <AgentChatPanel
            v-if="agentMode"
            key="agent-mode"
            v-model:show-status-dropdown="showStatusDropdown"
            v-model:new-message="newMessage"
            :user="user"
            :selected-conversation="selectedConversation"
            :is-loading-conversations="isLoadingConversations"
            :agent-conversations="agentConversations"
            :current-agent-status="currentAgentStatus"
            :agent-status-colors="agentStatusColors"
            :agent-status-labels="agentStatusLabels"
            :messages="messages"
            :is-sending="isSending"
            @close="handleClose"
            @change-status="changeAgentStatus"
            @select-conversation="selectConversation"
            @refresh-conversations="fetchAgentConversations"
            @back-to-conversation-list="backToConversationList"
            @send-message="sendMessage"
          />

          <!-- 访客模式：欢迎页 -->
          <ChatWelcomePanel
            v-else-if="showWelcomeScreen && !agentMode"
            key="welcome"
            :welcome-agents="welcomeAgents"
            :selected-agent="selectedAgent"
            :online-agents-count="onlineAgentsCount"
            :has-history-chat="hasHistoryChat"
            :email-settings="emailSettings"
            @close="handleClose"
            @select-agent="selectAgentFromWelcome"
            @enter-chat="enterChat"
          />

          <!-- 聊天窗口 - 简化布局 -->
          <div
            v-else
            key="chat"
            class="sidebar-panel chat-modal-shell relative w-full md:w-[420px] max-w-full md:max-w-[calc(100vw-2rem)] h-[95vh] md:h-[85vh] max-h-[800px] rounded-2xl overflow-hidden flex flex-col border-2 border-[#6b73ff]/40 ring-1 ring-white/10 bg-slate-950/80 backdrop-blur-xl shadow-[0_0_30px_rgba(107,115,255,0.6)] transition-colors duration-300 pointer-events-auto"
          >
            <!-- 背景装饰 -->
            <div class="absolute inset-x-0 top-0 h-[200px] bg-gradient-to-br from-indigo-600/20 to-teal-600/20 blur-3xl pointer-events-none"></div>
            <!-- 聊天区域 -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <!-- 头部 - 当前客服信息 -->
              <div class="border-b border-white/[0.08] bg-white/[0.03] backdrop-blur-md">
                <div class="px-4 py-3 flex items-center gap-3">
                  <!-- 返回欢迎页按钮 -->
                  <button
                    type="button"
                    class="w-9 h-9 rounded-full border border-white/20 tz-text-secondary flex items-center justify-center hover:border-white/40 hover:text-white transition-colors"
                    @click="showWelcomeScreen = true"
                    :title="t('chatModal.actions.backToWelcome')"
                    :aria-label="t('chatModal.actions.backToWelcome')"
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M15 18l-6-6 6-6"/>
                    </svg>
                  </button>
                  
                  <!-- 当前客服头像 -->
                  <div
                    class="w-10 h-10 rounded-full bg-white/20 flex items-center justify-center text-xs font-semibold text-white overflow-hidden shadow-[0_0_12px_rgba(15,23,42,0.95)] flex-shrink-0"
                  >
                    <img
                      v-if="selectedAgent?.avatar"
                      :src="selectedAgent.avatar"
                      :alt="selectedAgent.name"
                      class="w-full h-full rounded-full object-cover"
                    />
                    <span v-else>{{ selectedAgent ? getInitials(selectedAgent.name) : '?' }}</span>
                  </div>
                  
                  <!-- 客服信息 -->
                  <div class="flex-1 min-w-0">
                    <div class="text-white font-medium text-sm truncate">{{ selectedAgent?.name || t('chatModal.fallback.agent') }}</div>
                    <div class="tz-text-muted text-xs truncate">{{ selectedAgent?.email }}</div>
                  </div>
                  
                  <!-- Chat 按钮 -->
                  <button
                    type="button"
                    @click="activeTab = 'chat'"
                    class="w-9 h-9 rounded-full bg-[linear-gradient(135deg,#2dd4bf_0%,#3b82f6_100%)] text-white flex items-center justify-center hover:opacity-90 transition-opacity shadow-lg mr-2"
                    :title="t('chatModal.actions.switchToChat')"
                    :aria-label="t('chatModal.actions.switchToChat')"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
                    </svg>
                  </button>

                  <!-- WhatsApp 按钮 - 官方图标样式 -->
                  <a
                    v-if="selectedAgent?.whatsapp"
                    :href="`https://wa.me/${selectedAgent.whatsapp.replace('+', '')}`"
                    target="_blank"
                    class="w-9 h-9 rounded-full bg-[#25D366] text-white flex items-center justify-center hover:bg-[#20BA5A] transition-colors shadow-lg"
                    :title="t('chatModal.actions.contactViaWhatsApp')"
                    :aria-label="t('chatModal.actions.contactViaWhatsApp')"
                  >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                    </svg>
                  </a>
                  
                  <!-- 关闭按钮 -->
                  <button
                    type="button"
                    class="w-9 h-9 rounded-full border-2 border-white/20 tz-text-secondary flex items-center justify-center hover:border-red-500 hover:text-red-500 transition-colors"
                    :aria-label="t('chatModal.actions.close')"
                    @click="handleClose"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M18 6L6 18M6 6l12 12"/>
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 统一的聊天主体 (Mobile + Desktop) -->
              <UserChatBody
                class="flex-1 min-h-0"
                v-model:activeTab="activeTab"
                v-model:newMessage="newMessage"
                v-model:searchQuery="searchQuery"
                :currentThemeColor="currentThemeColor"
                :messages="messages"
                :isSending="isSending"
                :isUploadingImage="isUploadingImage"
                :isSearching="isSearching"
                :searchResults="searchResults"
                :productDrawerVisible="productDrawerVisible"
                :ordersList="ordersList"
                :isLoadingOrders="isLoadingOrders"
                :isMemberLogged="isMemberLogged"
                :levelName="levelName"
                :points="points"
                :tierInfo="tierInfo"
                :levelDiscounts="levelDiscounts"
                :userCoupons="userCoupons"
                :userPointCards="userPointCards"
                :isLoggedInForWarranty="isLoggedInForWarranty"
                @sendMessage="handleSendMessage"
                @uploadImage="handleImageUpload"
                @deleteMessage="handleMessageContextMenu"
                @search="searchProducts"
                @shareProduct="shareProductToChat"
                @openHistory="historyDrawerVisible = true"
                @openCart="openCartFromChat"
                @openWishlist="wishlistDrawerVisible = true"
                @shareOrder="shareOrderToChat"
                @openAuth="openMemberAuth"
                @loginRequest="handleWarrantyLoginRequest"
                @openTestReport="handleOpenTestReport"
              />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
    
    <ChatTransferModal
      v-model="showTransferModal"
      v-model:transfer-to-agent="transferToAgent"
      v-model:transfer-note="transferNote"
      :agents="agents"
      :selected-agent="selectedAgent"
      :is-transferring="isTransferring"
      @submit="handleTransfer"
    />
    <!-- Toast 提示 -->
    <Transition name="fade">
      <div
        v-if="showToast"
        class="fixed bottom-20 left-1/2 -translate-x-1/2 z-[10001] px-4 py-2 bg-black/90 text-white text-sm rounded-lg shadow-lg backdrop-blur-sm"
      >
        {{ toastMessage }}
      </div>
    </Transition>

    <WhatsAppProductSearchResultDrawer
      v-model="productDrawerVisible"
      :loading="isSearching"
      :results="searchResults"
      :error="productDrawerError"
      :agent="selectedAgent"
      :query="productDrawerQuery"
      @close="handleProductDrawerClose"
      @select="shareProductToChat"
      @add-to-cart="handleAddProductToCart"
    />

    <TestReportDrawer
      v-model="testReportDrawerVisible"
      :agent="selectedAgent"
    />

    <WishlistDrawer
      v-model="wishlistDrawerVisible"
      variant="bottom"
      @share-to-chat="handleShareProductFromHistory"
    />

    <!-- 聊天内登录弹窗（复用全局 AuthModal，嵌入模式） -->
    <LazyAuthModal
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleChatAuthSuccess"
    />

    <Transition name="wa-drawer">
      <div
        v-if="historyDrawerVisible"
        class="wa-drawer-mask"
        @click.self="handleHistoryDrawerClose"
      >
        <!-- Backdrop -->
        <div 
          class="wa-drawer-backdrop md:hidden"
          @click="handleHistoryDrawerClose"
        />

        <div class="wa-drawer-shell">
          <div class="wa-drawer-header">
            <div class="wa-drawer-title">
              {{ t('chatModal.history.title') }}
            </div>
            <button
              type="button"
              class="wa-drawer-close-btn"
              :aria-label="t('chatModal.actions.closeHistory')"
              @click="handleHistoryDrawerClose"
            >
              <span class="text-lg leading-none">x</span>
            </button>
          </div>

          <div class="wa-drawer-content">
            <BrowsingHistoryDark @share-to-chat="handleShareProductFromHistory" />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { useWhatsAppState } from '~/composables/chat/useWhatsAppState'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import AgentChatPanel from '~/components/whatsapp/AgentChatPanel.vue'
import ChatTransferModal from '~/components/whatsapp/ChatTransferModal.vue'
import ChatWelcomePanel from '~/components/whatsapp/ChatWelcomePanel.vue'
import UserChatBody from '~/components/whatsapp/UserChatBody.vue'
import TestReportDrawer from '~/components/whatsapp/TestReportDrawer.vue'

// Props - 现在不需要预先传入conversation
const props = defineProps<{
  conversation?: {
    showAgentList?: boolean
  }
}>()

// Emits
const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()

const {
  user,
  agentMode,
  agentConversations,
  isLoadingConversations,
  selectedConversation,
  currentAgentStatus,
  showStatusDropdown,
  agentStatusColors,
  agentStatusLabels,
  showWelcomeScreen,
  hasHistoryChat,
  agents,
  selectedAgent,
  welcomeAgents,
  onlineAgentsCount,
  emailSettings,
  isSending,
  messages,
  activeTab,
  newMessage,
  searchQuery,
  searchResults,
  isSearching,
  ordersList,
  isLoadingOrders,
  productDrawerVisible,
  productDrawerError,
  productDrawerQuery,
  historyDrawerVisible,
  wishlistDrawerVisible,
  showTransferModal,
  transferToAgent,
  transferNote,
  isTransferring,
  isUploadingImage,
  testReportDrawerVisible,
  showToast,
  toastMessage,
  isMemberLogged,
  levelName,
  points,
  tierInfo,
  levelDiscounts,
  userCoupons,
  userPointCards,
  isLoggedInForWarranty,
  showAuthModal,
  authMode,
  currentThemeColor,
  openMemberAuth,
  handleWarrantyLoginRequest,
  handleChatAuthSuccess,
  handleOpenTestReport,
  handleClose,
  enterChat,
  selectAgentFromWelcome,
  handleMessageContextMenu,
  handleSendMessage,
  searchProducts,
  handleAddProductToCart,
  handleProductDrawerClose,
  handleHistoryDrawerClose,
  shareProductToChat,
  handleShareProductFromHistory,
  shareOrderToChat,
  openCartFromChat,
  getInitials,
  handleImageUpload,
  handleTransfer,
  fetchAgentConversations,
  selectConversation,
  backToConversationList,
  sendMessage,
  changeAgentStatus
} = useWhatsAppState(emit)
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

/* 滑入动画 - FAQ 从底部滑上来 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: transform 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  transform: translateY(100%);
}

/* 挥手动画 */
@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  25% { transform: rotate(20deg); }
  75% { transform: rotate(-10deg); }
}

.animate-wave {
  animation: wave 1.5s ease-in-out infinite;
}

/* 欢迎页/聊天窗口切换动画 */
.fade-scale-enter-active,
.fade-scale-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-scale-enter-from {
  opacity: 0;
  transform: scale(0.95);
}

.fade-scale-leave-to {
  opacity: 0;
  transform: scale(0.95);
}

/* 渐变边框按钮 */
.gradient-border-btn {
  background: linear-gradient(black, black) padding-box,
              linear-gradient(to right, #40ffaa, #6b73ff) border-box;
  border: 2px solid transparent;
}

/* 渐变文字 */
.gradient-text {
  background: linear-gradient(to right, #40ffaa, #6b73ff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 自定义滚动条 */
.overflow-y-auto::-webkit-scrollbar {
  width: 6px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: #888;
  border-radius: 10px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #555;
}

.overflow-y-auto::-webkit-scrollbar-thumb.social-btn {
  width: 3rem;
  height: 3rem;
  border-radius: 9999px;
}

.chat-modal-shell {
  height: min(95vh, calc(100vh - 16px));
  max-height: min(95vh, calc(100vh - 16px));
}

.chat-history-shell {
  height: 90vh;
  max-height: 80vh;
}

@supports (height: 100dvh) {
  .chat-modal-shell {
    height: min(95dvh, calc(100dvh - 16px));
    max-height: min(95dvh, calc(100dvh - 16px));
  }

  .chat-history-shell {
    height: 90dvh;
    max-height: 80dvh;
  }
}

@media (min-width: 768px) {
  .chat-modal-shell {
    height: 850px;
    max-height: 92vh;
  }

  .chat-history-shell {
    height: 850px;
    max-height: 92vh;
  }

  @supports (height: 100dvh) {
    .chat-modal-shell {
      height: 850px;
      max-height: 92dvh;
    }

    .chat-history-shell {
      height: 850px;
      max-height: 92dvh;
    }
  }
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: #555;
}
</style>
