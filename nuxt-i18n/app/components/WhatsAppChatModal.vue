<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="fixed inset-0 z-[10000] flex items-center justify-center md:items-end md:justify-end p-0 md:pr-6 md:pb-8 pointer-events-none tz-mobile-safe-modal-mask tz-mobile-dialog-mask"
      >
        <div class="absolute inset-0 bg-black/80 backdrop-blur-sm md:hidden pointer-events-auto"></div>
        <!-- 聊天窗口容器 - 右下角定位 -->
        
        <!-- 客户侧聊天：前台只负责访客/会员与客服 Profile 建立会话 -->
        <Transition name="fade-scale" mode="out-in">
          <!-- 访客模式：欢迎页 -->
          <ChatWelcomePanel
            v-if="showWelcomeScreen"
            key="welcome"
            ref="chatModalRef"
            class="chat-modal-draggable-shell"
            :class="{ 'chat-modal-shell--dragging': isDraggingChatModal }"
            :style="chatModalDragStyle"
            :welcome-agents="welcomeAgents"
            :selected-agent="selectedAgent"
            :online-agents-count="onlineAgentsCount"
            :has-history-chat="hasHistoryChat"
            :email-settings="emailSettings"
            @drag-start="handleChatModalDragStart"
            @close="handleClose"
            @select-agent="selectAgentFromWelcome"
            @enter-chat="enterChat"
          />

          <!-- 聊天窗口 - 简化布局 -->
          <div
            v-else
            key="chat"
            ref="chatModalRef"
            class="chat-modal-draggable-shell chat-modal-shell tz-mobile-dialog-surface relative w-full md:w-[560px] max-w-full md:max-w-[calc(100vw-3rem)] tz-mobile-safe-full-height rounded-none md:rounded-2xl overflow-hidden flex flex-col bg-[#111111] shadow-[-12px_0_28px_rgba(0,0,0,0.45)] transition-colors duration-300 pointer-events-auto"
            :class="{ 'chat-modal-shell--dragging': isDraggingChatModal }"
            :style="chatModalDragStyle"
          >
            <!-- 聊天区域 -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <div
                class="chat-modal-drag-handle border-b border-white/[0.08] bg-white/[0.03] backdrop-blur-md"
                @pointerdown="handleChatModalDragStart"
              >
                <div class="px-3 py-2 flex items-center justify-end gap-3">
                  <button
                    type="button"
                    class="tz-global-close-btn"
                    :aria-label="t('chatModal.actions.close')"
                    @click="handleClose"
                  >
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <path d="M18 6L6 18M6 6l12 12"/>
                    </svg>
                  </button>
                </div>
              </div>

              <!-- 头部 - 当前客服信息 -->
              <div class="relative z-30 border-b border-white/[0.08] bg-[#050505]">
                <div class="px-4 py-3 flex items-center gap-3">
                  <div
                    data-no-drag
                    class="flex min-w-0 flex-1 items-center gap-2 rounded-full bg-[#111111] px-2.5 py-2 transition-colors hover:bg-[#171717]"
                  >
                    <button
                      type="button"
                      class="flex min-w-0 flex-1 items-center gap-3 text-left"
                      :aria-expanded="agentPickerOpen"
                      aria-haspopup="listbox"
                      @click="toggleAgentPicker"
                    >
                      <span
                        class="w-10 h-10 rounded-full bg-white/12 flex items-center justify-center text-xs font-semibold text-white overflow-hidden flex-shrink-0"
                      >
                        <img
                          v-if="selectedAgent?.avatar"
                          :src="selectedAgent.avatar"
                          :alt="selectedAgent.name"
                          class="w-full h-full rounded-full object-cover"
                        />
                        <span v-else>{{ selectedAgentInitials }}</span>
                      </span>

                      <span class="flex-1 min-w-0">
                        <span class="block text-white font-medium text-sm truncate">{{ selectedAgent?.name || t('chatModal.fallback.agent') }}</span>
                        <span class="block tz-text-muted text-xs truncate">{{ selectedAgentContactLabel }}</span>
                      </span>

                    </button>

                    <span class="flex shrink-0 items-center gap-1">
                      <a
                        v-if="getAgentEmailHref(selectedAgent)"
                        :href="getAgentEmailHref(selectedAgent)"
                        class="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-[#151515] text-white transition-colors hover:border-white/35 hover:bg-[#202020]"
                        :title="t('chatModal.actions.email')"
                        :aria-label="t('chatModal.actions.email')"
                        @click.stop
                      >
                        <Icon name="lucide:mail" class="h-4 w-4" />
                      </a>
                      <span
                        v-else
                        class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border border-white/[0.07] bg-[#0d0d0d] text-white/24"
                        :title="t('chatModal.actions.email')"
                        :aria-label="t('chatModal.actions.email')"
                        aria-disabled="true"
                      >
                        <Icon name="lucide:mail" class="h-4 w-4" />
                      </span>

                      <a
                        v-if="getAgentWhatsAppHref(selectedAgent)"
                        :href="getAgentWhatsAppHref(selectedAgent)"
                        target="_blank"
                        rel="noopener noreferrer"
                        class="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-[#151515] text-white transition-colors hover:border-white/35 hover:bg-[#202020]"
                        :title="t('chatModal.actions.contactViaWhatsApp')"
                        :aria-label="t('chatModal.actions.contactViaWhatsApp')"
                        @click.stop
                      >
                        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                          <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                        </svg>
                      </a>
                      <span
                        v-else
                        class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border border-white/[0.07] bg-[#0d0d0d] text-white/24"
                        :title="t('chatModal.actions.contactViaWhatsApp')"
                        :aria-label="t('chatModal.actions.contactViaWhatsApp')"
                        aria-disabled="true"
                      >
                        <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                          <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                        </svg>
                      </span>
                    </span>
                    <button
                      type="button"
                      class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-white/[0.07] bg-[#0d0d0d] text-white/55 transition-colors hover:border-white/20 hover:bg-[#151515] hover:text-white"
                      :aria-expanded="agentPickerOpen"
                      aria-haspopup="listbox"
                      @click="toggleAgentPicker"
                    >
                      <svg class="h-4 w-4 transition-transform" :class="{ 'rotate-180': agentPickerOpen }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="m6 9 6 6 6-6"/>
                      </svg>
                    </button>
                  </div>

                </div>

                <div
                  v-if="agentPickerOpen"
                  data-no-drag
                  class="absolute left-4 right-4 top-[calc(100%+0.5rem)] z-[80] overflow-hidden rounded-[30px] bg-[#050505] opacity-100 shadow-[0_18px_50px_rgba(0,0,0,0.85)]"
                >
                  <div class="max-h-[min(380px,55vh)] overflow-y-auto bg-[#050505] p-2" role="listbox">
                    <div
                      v-for="agent in agents"
                      :key="agent.id"
                      class="w-full rounded-full border transition-colors"
                      :class="isSelectedAgent(agent) ? 'border-[#B5FF6D]/70 bg-[#151515]' : 'border-transparent bg-[#080808] hover:border-white/14 hover:bg-[#101010]'"
                      role="option"
                      :aria-selected="isSelectedAgent(agent)"
                    >
                      <div class="flex items-center gap-2 px-3 py-2.5">
                        <button
                          type="button"
                          class="flex min-w-0 flex-1 items-center gap-3 text-left"
                          @click="handleAgentPickerSelect(agent)"
                        >
                          <span class="relative h-10 w-10 flex-shrink-0 overflow-hidden rounded-full bg-white/12 text-xs font-semibold text-white flex items-center justify-center">
                          <img v-if="agent.avatar" :src="agent.avatar" :alt="agent.name" class="h-full w-full object-cover" />
                          <span v-else>{{ getAgentInitials(agent) }}</span>
                          <span class="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border border-black" :class="getAgentStatusDotClass(agent)"></span>
                          </span>
                          <span class="min-w-0 flex-1">
                            <span class="block truncate text-sm font-semibold text-white">{{ agent.name || t('chatModal.fallback.agent') }}</span>
                            <span class="block truncate text-xs text-white/58">{{ getAgentContactLabel(agent) }}</span>
                          </span>
                        </button>
                        <span v-if="isSelectedAgent(agent)" class="h-2 w-2 rounded-full bg-[#B5FF6D]"></span>
                        <span class="flex shrink-0 items-center gap-1">
                          <a
                            v-if="getAgentEmailHref(agent)"
                            :href="getAgentEmailHref(agent)"
                            class="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-[#151515] text-white transition-colors hover:border-white/35 hover:bg-[#202020]"
                            :title="t('chatModal.actions.email')"
                            :aria-label="t('chatModal.actions.email')"
                            @click.stop
                          >
                            <Icon name="lucide:mail" class="h-4 w-4" />
                          </a>
                          <span
                            v-else
                            class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border border-white/[0.07] bg-[#0d0d0d] text-white/24"
                            :title="t('chatModal.actions.email')"
                            :aria-label="t('chatModal.actions.email')"
                            aria-disabled="true"
                          >
                            <Icon name="lucide:mail" class="h-4 w-4" />
                          </span>

                          <a
                            v-if="getAgentWhatsAppHref(agent)"
                            :href="getAgentWhatsAppHref(agent)"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-[#151515] text-white transition-colors hover:border-white/35 hover:bg-[#202020]"
                            :title="t('chatModal.actions.contactViaWhatsApp')"
                            :aria-label="t('chatModal.actions.contactViaWhatsApp')"
                            @click.stop
                          >
                            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                              <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                            </svg>
                          </a>
                          <span
                            v-else
                            class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border border-white/[0.07] bg-[#0d0d0d] text-white/24"
                            :title="t('chatModal.actions.contactViaWhatsApp')"
                            :aria-label="t('chatModal.actions.contactViaWhatsApp')"
                            aria-disabled="true"
                          >
                            <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
                              <path d="M17.472 14.382c-.297-.149-1.758-.867-2.03-.967-.273-.099-.471-.148-.67.15-.197.297-.767.966-.94 1.164-.173.199-.347.223-.644.075-.297-.15-1.255-.463-2.39-1.475-.883-.788-1.48-1.761-1.653-2.059-.173-.297-.018-.458.13-.606.134-.133.298-.347.446-.52.149-.174.198-.298.298-.497.099-.198.05-.371-.025-.52-.075-.149-.669-1.612-.916-2.207-.242-.579-.487-.5-.669-.51-.173-.008-.371-.01-.57-.01-.198 0-.52.074-.792.372-.272.297-1.04 1.016-1.04 2.479 0 1.462 1.065 2.875 1.213 3.074.149.198 2.096 3.2 5.077 4.487.709.306 1.262.489 1.694.625.712.227 1.36.195 1.871.118.571-.085 1.758-.719 2.006-1.413.248-.694.248-1.289.173-1.413-.074-.124-.272-.198-.57-.347m-5.421 7.403h-.004a9.87 9.87 0 01-5.031-1.378l-.361-.214-3.741.982.998-3.648-.235-.374a9.86 9.86 0 01-1.51-5.26c.001-5.45 4.436-9.884 9.888-9.884 2.64 0 5.122 1.03 6.988 2.898a9.825 9.825 0 012.893 6.994c-.003 5.45-4.437 9.884-9.885 9.884m8.413-18.297A11.815 11.815 0 0012.05 0C5.495 0 .16 5.335.157 11.892c0 2.096.547 4.142 1.588 5.945L.057 24l6.305-1.654a11.882 11.882 0 005.683 1.448h.005c6.554 0 11.89-5.335 11.893-11.893a11.821 11.821 0 00-3.48-8.413z"/>
                            </svg>
                          </span>
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 统一的聊天主体 (Mobile + Desktop) -->
              <UserChatBody
                class="flex-1 min-h-0"
                v-model:activeTab="activeTab"
                v-model:newMessage="newMessage"
                :currentThemeColor="currentThemeColor"
                :messages="messages"
                :visitorEmail="visitorEmail"
                :showVisitorEmailCapture="showVisitorEmailCapture"
                :isSending="isSending"
                :isUploadingImage="isUploadingImage"
                :pendingProductReference="pendingProductReference"
                :agentTyping="agentTyping"
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
                @update:visitorEmail="visitorEmail = $event"
                @uploadImage="handleImageUpload"
                @openOrderPicker="openOrderPicker"
                @openCustomerServiceProductSearchModal="openCustomerServiceProductSearchModal"
                @clearPendingProductReference="clearPendingProductReference"
                @deleteMessage="handleMessageContextMenu"
                @shareOrder="shareOrderToChat"
                @openAuth="openMemberAuth"
                @loginRequest="handleWarrantyLoginRequest"
              />
            </div>
          </div>
        </Transition>
      </div>
    </Transition>
    
    <!-- Toast 提示 -->
    <Transition name="fade">
      <div
        v-if="showToast"
        class="chat-toast fixed bottom-20 left-1/2 -translate-x-1/2 z-[10001] px-4 py-2 bg-black/90 text-white text-sm rounded-lg shadow-lg backdrop-blur-sm"
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
      @confirm-config="shareProductConfigConfirmToChat"
    />

    <CustomerServiceProductSearchModal
      v-model="customerServiceProductSearchModalVisible"
      @close="closeCustomerServiceProductSearchModal"
      @select-customer-service-product="handleSelectCustomerServiceProductFromSearchModal"
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
              <Icon name="lucide:x" class="h-3.5 w-3.5" />
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from '#imports'
import { useWhatsAppState } from '~/composables/chat/useWhatsAppState'
import CustomerServiceProductSearchModal from '~/components/CustomerServiceProductSearchModal.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import ChatWelcomePanel from '~/components/whatsapp/ChatWelcomePanel.vue'
import UserChatBody from '~/components/whatsapp/UserChatBody.vue'

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

type ChatModalElementRef = HTMLElement | { rootElement?: HTMLElement | { value?: HTMLElement | null } | null }

const chatModalRef = ref<ChatModalElementRef | null>(null)
const chatModalPosition = ref<{ left: number; top: number } | null>(null)
const isDraggingChatModal = ref(false)
const agentPickerOpen = ref(false)
const CHAT_MODAL_DRAG_MARGIN = 16

let chatModalDragState: {
  startX: number
  startY: number
  startLeft: number
  startTop: number
  width: number
  height: number
  pointerId: number
  handleElement: HTMLElement | null
} | null = null

const getChatModalElement = () => {
  const rawRef = chatModalRef.value
  if (!rawRef) return null

  if (rawRef instanceof HTMLElement) return rawRef

  const exposedRoot = rawRef.rootElement
  if (!exposedRoot) return null

  return exposedRoot instanceof HTMLElement ? exposedRoot : exposedRoot.value || null
}

const chatModalDragStyle = computed(() => {
  if (!chatModalPosition.value) return undefined

  return {
    position: 'fixed' as const,
    left: `${Math.round(chatModalPosition.value.left)}px`,
    top: `${Math.round(chatModalPosition.value.top)}px`,
    right: 'auto',
    bottom: 'auto'
  }
})

const clampChatModalDragValue = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max)

const removeChatModalDragListeners = () => {
  if (typeof window === 'undefined') return
  window.removeEventListener('pointermove', handleChatModalDragMove)
  window.removeEventListener('pointerup', handleChatModalDragEnd)
  window.removeEventListener('pointercancel', handleChatModalDragEnd)
}

const handleChatModalDragStart = (event: PointerEvent) => {
  if (typeof window === 'undefined' || window.innerWidth < 768) return
  if (event.pointerType === 'mouse' && event.button !== 0) return

  const target = event.target as HTMLElement | null
  if (target?.closest('button, a, input, textarea, select, [data-no-drag]')) return

  const modalElement = getChatModalElement()
  if (!modalElement) return
  const rect = modalElement.getBoundingClientRect()
  const handleElement = event.currentTarget instanceof HTMLElement ? event.currentTarget : null

  chatModalDragState = {
    startX: event.clientX,
    startY: event.clientY,
    startLeft: rect.left,
    startTop: rect.top,
    width: rect.width,
    height: rect.height,
    pointerId: event.pointerId,
    handleElement
  }
  chatModalPosition.value = { left: Math.round(rect.left), top: Math.round(rect.top) }

  isDraggingChatModal.value = true
  handleElement?.setPointerCapture?.(event.pointerId)
  window.addEventListener('pointermove', handleChatModalDragMove, { passive: false })
  window.addEventListener('pointerup', handleChatModalDragEnd)
  window.addEventListener('pointercancel', handleChatModalDragEnd)
  event.preventDefault()
}

const handleChatModalDragMove = (event: PointerEvent) => {
  if (!chatModalDragState || typeof window === 'undefined') return

  event.preventDefault()

  const state = chatModalDragState
  const rawLeft = state.startLeft + event.clientX - state.startX
  const rawTop = state.startTop + event.clientY - state.startY
  const maxLeft = Math.max(CHAT_MODAL_DRAG_MARGIN, window.innerWidth - state.width - CHAT_MODAL_DRAG_MARGIN)
  const maxTop = Math.max(CHAT_MODAL_DRAG_MARGIN, window.innerHeight - state.height - CHAT_MODAL_DRAG_MARGIN)

  chatModalPosition.value = {
    left: Math.round(clampChatModalDragValue(rawLeft, CHAT_MODAL_DRAG_MARGIN, maxLeft)),
    top: Math.round(clampChatModalDragValue(rawTop, CHAT_MODAL_DRAG_MARGIN, maxTop))
  }
}

const handleChatModalDragEnd = () => {
  const state = chatModalDragState
  if (state?.handleElement?.hasPointerCapture?.(state.pointerId)) {
    state.handleElement.releasePointerCapture(state.pointerId)
  }
  chatModalDragState = null
  isDraggingChatModal.value = false
  removeChatModalDragListeners()
}

const keepChatModalInsideViewport = () => {
  if (typeof window === 'undefined') return
  if (window.innerWidth < 768) {
    chatModalPosition.value = null
    handleChatModalDragEnd()
    return
  }

  const modalElement = getChatModalElement()
  if (!modalElement) return

  const rect = modalElement.getBoundingClientRect()
  const currentLeft = chatModalPosition.value?.left ?? rect.left
  const currentTop = chatModalPosition.value?.top ?? rect.top
  const maxLeft = Math.max(CHAT_MODAL_DRAG_MARGIN, window.innerWidth - rect.width - CHAT_MODAL_DRAG_MARGIN)
  const maxTop = Math.max(CHAT_MODAL_DRAG_MARGIN, window.innerHeight - rect.height - CHAT_MODAL_DRAG_MARGIN)

  chatModalPosition.value = {
    left: Math.round(clampChatModalDragValue(currentLeft, CHAT_MODAL_DRAG_MARGIN, maxLeft)),
    top: Math.round(clampChatModalDragValue(currentTop, CHAT_MODAL_DRAG_MARGIN, maxTop))
  }
}

onMounted(() => {
  if (typeof window === 'undefined') return
  window.addEventListener('resize', keepChatModalInsideViewport)
})

onBeforeUnmount(() => {
  removeChatModalDragListeners()
  if (typeof window !== 'undefined') {
    window.removeEventListener('resize', keepChatModalInsideViewport)
  }
})

const {
  user,
  showWelcomeScreen,
  hasHistoryChat,
  agents,
  selectedAgent,
  welcomeAgents,
  onlineAgentsCount,
  emailSettings,
  visitorEmail,
  showVisitorEmailCapture,
  isSending,
  messages,
  activeTab,
  newMessage,
  searchResults,
  isSearching,
  ordersList,
  isLoadingOrders,
  productDrawerVisible,
  productDrawerError,
  productDrawerQuery,
  customerServiceProductSearchModalVisible,
  historyDrawerVisible,
  wishlistDrawerVisible,
  isUploadingImage,
  pendingProductReference,
  agentTyping,
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
  openOrderPicker,
  openCustomerServiceProductSearchModal,
  closeCustomerServiceProductSearchModal,
  clearPendingProductReference,
  handleWarrantyLoginRequest,
  handleChatAuthSuccess,
  handleClose,
  enterChat,
  selectAgent,
  selectAgentFromWelcome,
  handleMessageContextMenu,
  handleSendMessage,
  handleAddProductToCart,
  handleProductDrawerClose,
  handleHistoryDrawerClose,
  shareProductToChat,
  handleSelectCustomerServiceProductFromSearchModal,
  shareProductConfigConfirmToChat,
  handleShareProductFromHistory,
  shareOrderToChat,
  handleImageUpload
} = useWhatsAppState(emit)

const getAgentInitials = (agent: any) => {
  const name = String(agent?.name || agent?.display_name || agent?.email || '').trim()
  if (!name) return 'CS'
  const words = name.split(/[\s._-]+/).filter(Boolean)
  return words.slice(0, 2).map((word) => word.charAt(0).toUpperCase()).join('') || 'CS'
}

const normalizeWhatsAppNumber = (value: unknown) => {
  const digits = String(value || '').replace(/[^0-9]/g, '')
  return digits.length >= 6 ? digits : ''
}

const getAgentWhatsAppHref = (agent: any) => {
  const number = normalizeWhatsAppNumber(agent?.whatsapp)
  return number ? `https://wa.me/${number}` : ''
}

const getAgentEmailHref = (agent: any) => {
  const email = String(agent?.email || '').trim()
  return email ? `mailto:${email}` : ''
}

const getAgentContactLabel = (agent: any) => {
  const email = String(agent?.email || '').trim()
  const whatsapp = String(agent?.whatsapp || '').trim()
  if (email && whatsapp) return `${email} / ${whatsapp}`
  return email || whatsapp || t('chatModal.agentSelector.descriptions.default')
}

const getAgentStatusDotClass = (agent: any) => {
  const status = String(agent?.online_status || agent?.status || '').trim().toLowerCase()
  if (status === 'online') return 'bg-[#B5FF6D]'
  if (status === 'busy') return 'bg-white/70'
  if (status === 'away') return 'bg-white/45'
  return 'bg-white/25'
}

const isSelectedAgent = (agent: any) => String(selectedAgent.value?.id ?? '') === String(agent?.id ?? '')

const selectedAgentInitials = computed(() => getAgentInitials(selectedAgent.value))
const selectedAgentContactLabel = computed(() => getAgentContactLabel(selectedAgent.value))

const toggleAgentPicker = () => {
  agentPickerOpen.value = !agentPickerOpen.value
}

const handleAgentPickerSelect = (agent: any) => {
  selectAgent(agent)
  showWelcomeScreen.value = false
  agentPickerOpen.value = false
}
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
              linear-gradient(to right, #fff, #fff) border-box;
  border: 2px solid transparent;
}

/* 渐变文字 */
.gradient-text {
  background: linear-gradient(to right, #fff, #fff);
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
  height: 100vh;
  max-height: 100vh;
  background: #000 !important;
  box-shadow:
    0 0 1px rgba(255, 255, 255, 0.42),
    0 0 8px rgba(255, 255, 255, 0.12),
    0 0 18px rgba(255, 255, 255, 0.055);
}

.chat-history-shell {
  height: 100vh;
  max-height: 100vh;
}

.chat-modal-drag-handle {
  touch-action: auto;
}

@supports (height: 100dvh) {
  .chat-modal-shell {
    height: 100dvh;
    max-height: 100dvh;
  }

  .chat-history-shell {
    height: 100dvh;
    max-height: 100dvh;
  }
}

@media (max-width: 767px) {
  .chat-modal-shell,
  .chat-history-shell {
    height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }

  .chat-toast {
    bottom: calc(5rem + var(--tz-safe-area-bottom, 0px));
  }
}

@media (min-width: 768px) {
  .chat-modal-shell {
    height: min(900px, calc(100vh - 56px));
    max-height: calc(100vh - 56px);
  }

  .chat-modal-draggable-shell {
    position: fixed;
    right: 1.5rem;
    bottom: 2rem;
  }

  .chat-modal-drag-handle {
    cursor: grab;
    touch-action: none;
    user-select: none;
  }

  .chat-modal-shell--dragging,
  .chat-modal-shell--dragging .chat-modal-drag-handle {
    cursor: grabbing;
  }

  .chat-history-shell {
    height: 850px;
    max-height: 92vh;
  }

  @supports (height: 100dvh) {
    .chat-modal-shell {
      height: min(900px, calc(100dvh - 56px));
      max-height: calc(100dvh - 56px);
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
