<template>
  <Teleport to="body">
    <!-- 遮罩层 -->
    <Transition name="fade">
      <div
        v-if="conversation"
        class="chat-drawer-mask fixed inset-0 z-[10050] flex items-stretch justify-end pointer-events-none"
      >
        <div
          class="chat-drawer-backdrop"
          :class="{ 'chat-drawer-backdrop--hidden': isChatCollapsed }"
          aria-hidden="true"
          @click="handleClose"
        ></div>
        <div
          class="chat-drawer-stage"
          :class="{ 'chat-drawer-stage--collapsed': isChatCollapsed }"
        >
          <div
            id="chat-drawer-dialog"
            class="chat-drawer-panel-region"
            role="dialog"
            :aria-modal="!isChatCollapsed"
            :aria-hidden="isChatCollapsed"
            :inert="isChatCollapsed"
            :aria-labelledby="showAgentSelectionPanel ? 'chat-agent-selection-title' : 'chat-current-agent-title'"
          >
            <Transition name="chat-drawer-panel" mode="out-in">
              <ChatAgentSelectionPanel
                v-if="showAgentSelectionPanel"
                key="agent-selection"
                :agents="agentSelectionAgents"
                :selected-agent="selectedAgent"
                :online-agents-count="onlineAgentsCount"
                :has-history-chat="hasHistoryChat"
                :email-settings="emailSettings"
                @close="handleClose"
                @select-agent="selectAgentFromAgentSelectionPanel"
                @enter-chat="enterChat"
              />

              <div
                v-else
                key="chat"
                class="chat-modal-shell chat-drawer-panel-surface relative flex h-full w-full min-w-0 flex-col overflow-hidden tz-surface-card transition-colors duration-300 pointer-events-auto"
              >
            <!-- 聊天区域 -->
            <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
              <div
                class="chat-drawer-top-bar tz-mobile-chrome-top border-b tz-border-strong/[0.08] backdrop-blur-md"
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
              <div class="relative z-30 border-b tz-border-strong/[0.08] tz-mobile-chrome-bottom">
                <div class="px-4 py-3 flex items-center gap-3">
                  <div
                    class="flex min-w-0 flex-1 items-center gap-2 rounded-full tz-surface-subtle px-2.5 py-2 transition-colors hover:tz-surface-muted"
                  >
                    <button
                      type="button"
                      class="flex min-w-0 flex-1 items-center gap-3 text-left"
                      :aria-expanded="agentPickerOpen"
                      aria-haspopup="listbox"
                      @click="toggleAgentPicker"
                    >
                      <span
                        class="w-10 h-10 rounded-full bg-slate-100 flex items-center justify-center text-xs font-semibold tz-text-primary overflow-hidden flex-shrink-0"
                      >
                        <StorefrontImage
                          v-if="selectedAgent?.avatar"
                          :src="selectedAgent.avatar"
                          :alt="selectedAgent.name"
                          class="w-full h-full rounded-full object-cover"
                          preset="avatar"
                        />
                        <span v-else>{{ selectedAgentPresentation.initials }}</span>
                      </span>

                      <span class="flex-1 min-w-0">
                        <span class="flex min-w-0 items-center gap-2">
                          <span id="chat-current-agent-title" class="block min-w-0 tz-text-primary font-medium text-sm truncate">{{ selectedAgent?.name || t('chatModal.fallback.agent') }}</span>
                          <span
                            class="inline-flex max-w-[8.5rem] shrink-0 items-center rounded-full border border-border/60 bg-white/70 px-2 py-0.5 text-[10px] font-semibold leading-none tz-text-secondary truncate"
                            :title="selectedAgentPresentation.groupLabel"
                          >
                            {{ selectedAgentPresentation.groupLabel }}
                          </span>
                        </span>
                        <span class="block tz-text-muted text-xs truncate">{{ selectedAgentContactLabel }}</span>
                      </span>

                    </button>

                    <span class="flex shrink-0 items-center gap-1">
                      <a
                        v-if="getAgentEmailHref(selectedAgent)"
                        :href="getAgentEmailHref(selectedAgent)"
                        class="flex h-8 w-8 items-center justify-center rounded-full border tz-border-subtle tz-surface-card tz-text-primary transition-colors hover:tz-border-strong/35 hover:tz-surface-muted"
                        :title="t('chatModal.actions.email')"
                        :aria-label="t('chatModal.actions.email')"
                        @click.stop
                      >
                        <Icon name="lucide:mail" class="h-4 w-4" />
                      </a>
                      <span
                        v-else
                        class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border tz-border-strong/[0.07] tz-surface-muted tz-text-primary/24"
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
                        class="flex h-8 w-8 items-center justify-center rounded-full border tz-border-subtle tz-surface-card tz-text-primary transition-colors hover:tz-border-strong/35 hover:tz-surface-muted"
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
                        class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border tz-border-strong/[0.07] tz-surface-muted tz-text-primary/24"
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
                      class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border tz-border-strong/[0.07] tz-surface-muted tz-text-primary/55 transition-colors hover:tz-border-subtle hover:tz-surface-card hover:tz-text-primary"
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
                  class="absolute left-4 right-4 top-[calc(100%+0.5rem)] z-[80] overflow-hidden rounded-[30px] tz-surface-card opacity-100 shadow-[0_18px_50px_rgba(15,23,42,0.18)]"
                >
                  <div class="max-h-[min(380px,55vh)] overflow-y-auto tz-surface-card p-2" role="listbox">
                    <div
                      v-for="agentEntry in agentEntries"
                      :key="agentEntry.agent.id"
                      class="w-full rounded-full border transition-colors"
                      :class="isSelectedAgent(agentEntry.agent) ? 'border-[#059669]/70 tz-surface-subtle' : 'border-transparent tz-surface-card hover:tz-border-strong/14 hover:tz-surface-muted'"
                      role="option"
                      :aria-selected="isSelectedAgent(agentEntry.agent)"
                    >
                      <div class="flex items-center gap-2 px-3 py-2.5">
                        <button
                          type="button"
                          class="flex min-w-0 flex-1 items-center gap-3 text-left"
                          @click="handleAgentPickerSelect(agentEntry.agent)"
                        >
                          <span class="relative h-10 w-10 flex-shrink-0 overflow-hidden rounded-full bg-slate-100 text-xs font-semibold tz-text-primary flex items-center justify-center">
                          <StorefrontImage v-if="agentEntry.agent.avatar" :src="agentEntry.agent.avatar" :alt="agentEntry.agent.name" class="h-full w-full object-cover" preset="avatar" />
                          <span v-else>{{ agentEntry.presentation.initials }}</span>
                          <span class="absolute bottom-0 right-0 h-2.5 w-2.5 rounded-full border border-black" :class="getAgentStatusDotClass(agentEntry.agent)"></span>
                          </span>
                          <span class="min-w-0 flex-1">
                            <span class="flex min-w-0 items-center gap-2">
                              <span class="block min-w-0 truncate text-sm font-semibold tz-text-primary">{{ agentEntry.agent.name || t('chatModal.fallback.agent') }}</span>
                              <span
                                class="inline-flex max-w-[8.5rem] shrink-0 items-center rounded-full border border-border/60 bg-white/70 px-2 py-0.5 text-[10px] font-semibold leading-none tz-text-secondary truncate"
                                :title="agentEntry.presentation.groupLabel"
                              >
                                {{ agentEntry.presentation.groupLabel }}
                              </span>
                            </span>
                            <span class="block truncate text-xs tz-text-primary/58">{{ agentEntry.presentation.contactLabel || t('chatModal.agentSelector.descriptions.default') }}</span>
                          </span>
                        </button>
                        <span v-if="isSelectedAgent(agentEntry.agent)" class="h-2 w-2 rounded-full bg-[#059669]"></span>
                        <span class="flex shrink-0 items-center gap-1">
                          <a
                            v-if="getAgentEmailHref(agentEntry.agent)"
                            :href="getAgentEmailHref(agentEntry.agent)"
                            class="flex h-8 w-8 items-center justify-center rounded-full border tz-border-subtle tz-surface-card tz-text-primary transition-colors hover:tz-border-strong/35 hover:tz-surface-muted"
                            :title="t('chatModal.actions.email')"
                            :aria-label="t('chatModal.actions.email')"
                            @click.stop
                          >
                            <Icon name="lucide:mail" class="h-4 w-4" />
                          </a>
                          <span
                            v-else
                            class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border tz-border-strong/[0.07] tz-surface-muted tz-text-primary/24"
                            :title="t('chatModal.actions.email')"
                            :aria-label="t('chatModal.actions.email')"
                            aria-disabled="true"
                          >
                            <Icon name="lucide:mail" class="h-4 w-4" />
                          </span>

                          <a
                            v-if="getAgentWhatsAppHref(agentEntry.agent)"
                            :href="getAgentWhatsAppHref(agentEntry.agent)"
                            target="_blank"
                            rel="noopener noreferrer"
                            class="flex h-8 w-8 items-center justify-center rounded-full border tz-border-subtle tz-surface-card tz-text-primary transition-colors hover:tz-border-strong/35 hover:tz-surface-muted"
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
                            class="flex h-8 w-8 cursor-not-allowed items-center justify-center rounded-full border tz-border-strong/[0.07] tz-surface-muted tz-text-primary/24"
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
                :hasOlderMessages="hasOlderMessages"
                :isLoadingOlderMessages="isLoadingOlderMessages"
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
                :isLoggedInForWarranty="isLoggedInForWarranty"
                @sendMessage="handleSendMessage"
                @loadOlderMessages="loadOlderMessages"
                @update:visitorEmail="visitorEmail = $event"
                @uploadImage="handleImageUpload"
                @openOrderPicker="openOrderPicker"
                @openCustomerServiceProductSearchModal="openCustomerServiceProductSearchModal"
                @clearPendingProductReference="clearPendingProductReference"
                @deleteMessage="handleMessageContextMenu"
                @retryMessage="retryMessage"
                @shareOrder="shareOrderToChat"
                @openAuth="openMemberAuth"
                @loginRequest="handleWarrantyLoginRequest"
              />
            </div>
              </div>
            </Transition>
          </div>
          <button
            type="button"
            class="chat-drawer-handle"
            :aria-label="t(isChatCollapsed ? 'chatModal.actions.expand' : 'chatModal.actions.collapse')"
            :title="t(isChatCollapsed ? 'chatModal.actions.expand' : 'chatModal.actions.collapse')"
            :aria-expanded="!isChatCollapsed"
            aria-controls="chat-drawer-dialog"
            @click="toggleChatDrawerCollapse"
          >
            <Icon
              :name="isChatCollapsed ? 'lucide:chevron-left' : 'lucide:chevron-right'"
              class="h-5 w-5"
              aria-hidden="true"
            />
          </button>
        </div>
      </div>
    </Transition>
    
    <!-- Toast 提示 -->
    <Transition name="fade">
      <div
        v-if="showToast"
        class="chat-toast fixed bottom-20 left-1/2 -translate-x-1/2 z-[10051] px-4 py-2 tz-surface-card tz-text-primary text-sm rounded-lg shadow-lg backdrop-blur-sm"
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useI18n } from '#imports'
import { useWhatsAppState } from '~/composables/chat/useWhatsAppState'
import { buildChatAgentPresentation, buildChatAgentPresentationList } from '~/lib/chatAgentPresentation'
import CustomerServiceProductSearchModal from '~/components/CustomerServiceProductSearchModal.vue'
import WhatsAppProductSearchResultDrawer from '~/components/WhatsAppProductSearchResultDrawer.vue'
import WishlistDrawer from '~/components/WishlistDrawer.vue'
import ChatAgentSelectionPanel from '~/components/whatsapp/ChatAgentSelectionPanel.vue'
import UserChatBody from '~/components/whatsapp/UserChatBody.vue'
import { createDialogStackId, useDialogStack } from '~/composables/useDialogStack'
import { useChatWidget } from '~/composables/useChatWidget'

// Props - 现在不需要预先传入conversation
const props = defineProps<{
  conversation?: {
    showAgentList?: boolean
    pendingSelectionRequest?: import('~/types/wheelsetSelectionAssistant').WheelsetSelectionRequestDraft | null
    pendingProductReference?: Record<string, any> | null
  }
}>()

// Emits
const emit = defineEmits<{
  close: []
}>()

const { t } = useI18n()
const dialogStack = useDialogStack()
const { isChatCollapsed, toggleChatCollapse } = useChatWidget()
const dialogStackId = createDialogStackId('whatsapp-chat-modal')
const agentPickerOpen = ref(false)
let unregisterDialogStack: (() => void) | null = null

const {
  user,
  showAgentSelectionPanel,
  hasHistoryChat,
  agents,
  selectedAgent,
  agentSelectionAgents,
  onlineAgentsCount,
  emailSettings,
  visitorEmail,
  showVisitorEmailCapture,
  isSending,
  messages,
  hasOlderMessages,
  isLoadingOlderMessages,
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
  isLoggedInForWarranty,
  showAuthModal,
  authMode,
  currentThemeColor,
  openMemberAuth,
  openOrderPicker,
  openCustomerServiceProductSearchModal,
  closeCustomerServiceProductSearchModal,
  dismissDetachedChatOverlays,
  clearPendingProductReference,
  handleWarrantyLoginRequest,
  handleChatAuthSuccess,
  handleClose,
  enterChat,
  selectAgent,
  selectAgentFromAgentSelectionPanel,
  handleMessageContextMenu,
  retryMessage,
  handleSendMessage,
  loadOlderMessages,
  handleAddProductToCart,
  handleProductDrawerClose,
  handleHistoryDrawerClose,
  shareProductToChat,
  handleSelectCustomerServiceProductFromSearchModal,
  shareProductConfigConfirmToChat,
  handleShareProductFromHistory,
  shareOrderToChat,
  handleImageUpload
} = useWhatsAppState(emit, {
  initialSelectionRequest: props.conversation?.pendingSelectionRequest || null,
})

const handleWhatsAppEscape = () => {
  if (agentPickerOpen.value) {
    agentPickerOpen.value = false
    return
  }

  handleClose()
}

const syncDialogStack = (isOpen: boolean) => {
  if (isOpen && !unregisterDialogStack) {
    unregisterDialogStack = dialogStack.register(dialogStackId, handleWhatsAppEscape, {
      priority: 10050,
    })
    return
  }

  if (!isOpen && unregisterDialogStack) {
    unregisterDialogStack()
    unregisterDialogStack = null
  }
}

watch(
  () => Boolean(props.conversation),
  value => syncDialogStack(value),
  { immediate: true }
)

onBeforeUnmount(() => {
  syncDialogStack(false)
})

const queuedConversationProductReference = ref<Record<string, any> | null>(
  props.conversation?.pendingProductReference || null
)

const applyQueuedConversationProductReference = () => {
  const productReference = queuedConversationProductReference.value
  if (!productReference || !selectedAgent.value?.id) return
  pendingProductReference.value = productReference
  activeTab.value = 'chat'
  queuedConversationProductReference.value = null
}

watch(
  () => props.conversation?.pendingProductReference,
  (productReference) => {
    if (!productReference) return
    queuedConversationProductReference.value = productReference
    applyQueuedConversationProductReference()
  },
  { immediate: true }
)

watch(
  () => selectedAgent.value?.id,
  () => applyQueuedConversationProductReference(),
  { flush: 'post' }
)

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

const getAgentStatusDotClass = (agent: any) => {
  const status = String(agent?.online_status || agent?.status || '').trim().toLowerCase()
  if (status === 'online') return 'bg-[#059669]'
  if (status === 'busy') return 'bg-white/70'
  if (status === 'away') return 'bg-white/45'
  return 'bg-white/25'
}

const isSelectedAgent = (agent: any) => String(selectedAgent.value?.id ?? '') === String(agent?.id ?? '')

const agentEntries = computed(() => buildChatAgentPresentationList(agents.value))

const selectedAgentPresentation = computed(() => buildChatAgentPresentation(selectedAgent.value))
const selectedAgentContactLabel = computed(() => selectedAgentPresentation.value.contactLabel || t('chatModal.agentSelector.descriptions.default'))

const toggleAgentPicker = () => {
  agentPickerOpen.value = !agentPickerOpen.value
}

const toggleChatDrawerCollapse = () => {
  if (!isChatCollapsed.value) {
    agentPickerOpen.value = false
    dismissDetachedChatOverlays()
  }
  toggleChatCollapse()
}

const handleAgentPickerSelect = (agent: any) => {
  selectAgent(agent)
  showAgentSelectionPanel.value = false
  agentPickerOpen.value = false
}
</script>

<style src="~/assets/css/components/whatsapp-mobile-drawer.css"></style>

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

/* 面板/聊天窗口切换动画 */
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

.chat-drawer-mask {
  box-sizing: border-box;
  isolation: isolate;
  padding: var(--tz-safe-area-top, 0px) var(--tz-safe-area-right, 0px)
    var(--tz-safe-area-bottom, 0px) var(--tz-safe-area-left, 0px);
}

.chat-drawer-backdrop {
  position: absolute;
  inset: 0;
  pointer-events: auto;
  background: rgba(15, 23, 42, 0.24);
  backdrop-filter: blur(2px);
  transition: opacity 320ms ease;
}

.chat-drawer-backdrop--hidden {
  opacity: 0;
  pointer-events: none;
  backdrop-filter: none;
}

.chat-drawer-stage {
  position: relative;
  z-index: 1;
  display: flex;
  width: clamp(22.5rem, 42vw, 47.5rem);
  max-width: min(47.5rem, calc(100vw - 2.5rem));
  height: 100%;
  min-height: 0;
  flex-direction: column;
  pointer-events: none;
  transition: transform 320ms cubic-bezier(0.2, 0.75, 0.25, 1);
}

.chat-drawer-stage--collapsed {
  transform: translateX(100%);
}

.chat-drawer-panel-region {
  position: relative;
  z-index: 1;
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  pointer-events: auto;
}

.chat-drawer-panel-surface {
  box-sizing: border-box;
  width: 100%;
  height: 100%;
  min-height: 0;
  max-height: 100%;
  border: 1px solid var(--tz-border-subtle);
  border-right: 0;
  border-radius: 1.25rem 0 0 1.25rem;
  background: var(--tz-card-surface) !important;
  box-shadow: -12px 0 32px rgba(15, 23, 42, 0.2);
  pointer-events: auto;
}

.chat-drawer-top-bar {
  flex: 0 0 auto;
}

.chat-drawer-handle {
  position: absolute;
  top: 50%;
  left: -2rem;
  z-index: 90;
  display: grid;
  width: 2rem;
  height: 4.75rem;
  transform: translateY(-50%);
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-right: 0;
  border-radius: 1.5rem 0 0 1.5rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-primary);
  box-shadow: none;
  cursor: pointer;
  pointer-events: auto;
  transition: color 160ms ease, background-color 160ms ease;
}

.chat-drawer-handle:hover {
  background: var(--tz-surface-subtle);
}

.chat-drawer-handle:focus-visible {
  outline: 2px solid var(--tz-form-control-focus-border);
  outline-offset: 3px;
}

@media (prefers-reduced-motion: reduce) {
  .chat-drawer-backdrop,
  .chat-drawer-stage {
    transition-duration: 0.01ms;
  }
}

.chat-drawer-panel-enter-active,
.chat-drawer-panel-leave-active {
  transition: transform 320ms cubic-bezier(0.2, 0.75, 0.25, 1), opacity 180ms ease;
}

.chat-drawer-panel-enter-from,
.chat-drawer-panel-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

.chat-modal-shell {
  height: 100%;
  max-height: 100%;
}

.chat-history-shell {
  height: 100vh;
  max-height: 100vh;
}

@supports (height: 100dvh) {
  .chat-history-shell {
    height: 100dvh;
    max-height: 100dvh;
  }
}

@media (max-width: 767px) {
  .chat-drawer-stage {
    width: min(100%, calc(100vw - 2.5rem));
  }

  .chat-drawer-panel-surface {
    border-radius: 1rem 0 0 1rem;
  }

  .chat-history-shell {
    height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
    max-height: calc(var(--tz-mobile-safe-viewport-height, 100dvh) - var(--tz-mobile-dialog-inset, 2px) * 2);
  }

  .chat-toast {
    bottom: calc(5rem + var(--tz-safe-area-bottom, 0px));
  }
}

@media (min-width: 768px) {
  .chat-history-shell {
    height: 850px;
    max-height: 92vh;
  }

  @supports (height: 100dvh) {
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
