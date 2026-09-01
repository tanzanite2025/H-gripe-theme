<template>
  <div class="flex flex-col h-full min-h-0">
    <!-- 消息列表区域 (Conversation History) -->
    <div 
      ref="messagesContainer"
      class="flex-1 overflow-y-auto space-y-3 px-1 md:px-6 md:py-7 md:space-y-4"
    >
      <!-- 空状态 -->
      <div v-if="messages.length === 0" class="flex flex-col items-center justify-center h-full tz-text-secondary text-sm">
        <svg class="w-12 h-12 md:w-16 md:h-16 mb-2 md:mb-4 tz-text-muted" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
        <p>No messages yet</p>
      </div>

      <!-- 消息循环 -->
      <div
        v-for="message in messages"
        :key="message.id"
        class="flex"
        :class="message.is_agent ? 'justify-end' : 'justify-start'"
      >
        <!-- 配置确认消息 -->
        <div
          v-if="isConfigConfirmMessage(message)"
            class="max-w-[82%] md:max-w-[76%] rounded-2xl border tz-border-subtle tz-surface-subtle p-3 tz-text-primary shadow-lg"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="tz-caption font-semibold uppercase tracking-[0.16em] text-emerald-700">
              Configuration request
            </div>
            <div class="tz-micro-label opacity-50 whitespace-nowrap">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
          <div class="mt-2 flex gap-3">
            <StorefrontImage
              v-if="configProduct(message).thumbnail"
              :src="configProduct(message).thumbnail"
              alt="Product"
              class="h-16 w-16 shrink-0 rounded-xl object-cover"
              preset="thumbnail"
            />
            <div class="min-w-0 flex-1">
              <a
                v-if="configProduct(message).url"
                :href="configProduct(message).url"
                target="_blank"
                rel="noopener"
                class="block truncate text-sm font-semibold tz-text-primary hover:underline"
              >
                {{ configProduct(message).title || message.message }}
              </a>
              <div v-else class="truncate text-sm font-semibold tz-text-primary">
                {{ configProduct(message).title || message.message }}
              </div>
              <div v-if="configProduct(message).price" class="mt-1 text-xs text-[#059669]">
                {{ configProduct(message).price }}
              </div>
              <div v-if="configProduct(message).sku" class="mt-1 tz-caption tz-text-primary/55">
                SKU: {{ configProduct(message).sku }}
              </div>
            </div>
          </div>
          <div class="mt-3 rounded-xl border tz-border-subtle tz-surface-subtle px-3 py-2 tz-caption leading-5 tz-text-muted">
            <div v-if="configSelection(message).variant_title" class="font-semibold tz-text-primary/85">
              Selected: {{ configSelection(message).variant_title }}
            </div>
            <div
              v-if="configOptionRows(message).length"
              class="mt-2 grid grid-cols-1 gap-1.5 sm:grid-cols-2"
            >
              <div
                v-for="option in configOptionRows(message)"
                :key="option.key"
                class="rounded-lg tz-surface-subtle px-2 py-1.5"
              >
                <span class="block tz-compact-label tz-text-primary/45">
                  {{ option.label }}
                </span>
                <span class="font-semibold tz-text-secondary">
                  {{ option.value }}<span v-if="option.unit"> {{ option.unit }}</span>
                </span>
              </div>
            </div>
            <div v-if="configSelection(message).weight_grams" class="mt-2 tz-caption tz-text-muted">
              Weight: {{ configSelection(message).weight_grams }}g
            </div>
            <div
              v-if="!configOptionRows(message).length && !configSelection(message).variant_title && !configSelection(message).weight_grams"
            >
              Customer asked staff to confirm this product configuration.
            </div>
          </div>
        </div>

        <!-- 订单确认消息 -->
        <div
          v-else-if="isOrderMessage(message)"
            class="max-w-[82%] md:max-w-[76%] rounded-2xl border border-[#059669]/35 bg-emerald-50 p-3 tz-text-primary shadow-lg"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="tz-caption font-semibold uppercase tracking-[0.16em] text-[#059669]">
              Order confirmation
            </div>
            <div class="tz-micro-label opacity-50 whitespace-nowrap">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
          <div class="mt-2">
            <a
              v-if="orderPayload(message).url"
              :href="orderPayload(message).url"
              target="_blank"
              rel="noopener"
              class="block truncate text-sm font-semibold tz-text-primary hover:underline"
            >
              {{ orderPayload(message).title || message.message }}
            </a>
            <div v-else class="truncate text-sm font-semibold tz-text-primary">
              {{ orderPayload(message).title || message.message }}
            </div>
            <div class="mt-1 flex flex-wrap gap-x-2 gap-y-1 tz-caption tz-text-primary/55">
              <span v-if="orderPayload(message).status">Status: {{ orderPayload(message).status }}</span>
              <span v-if="orderPayload(message).payment_status">Payment: {{ orderPayload(message).payment_status }}</span>
              <span v-if="orderPayload(message).shipping_status">Shipping: {{ orderPayload(message).shipping_status }}</span>
            </div>
            <div class="mt-2 text-xs font-semibold text-[#059669]">
              {{ formatOrderTotal(orderPayload(message)) }}
            </div>
          </div>
          <div v-if="orderItems(message).length" class="mt-3 space-y-1.5">
            <div
              v-for="item in orderItems(message).slice(0, 3)"
              :key="item.id || `${item.product_id}-${item.sku}`"
              class="rounded-xl border tz-border-subtle tz-surface-subtle px-3 py-2 tz-caption"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="truncate font-semibold tz-text-primary/85">{{ item.title || item.product_name || 'Product' }}</span>
                <span class="shrink-0 tz-text-muted">x{{ item.quantity || 1 }}</span>
              </div>
              <div v-if="item.sku" class="mt-1 tz-text-primary/45">SKU: {{ item.sku }}</div>
            </div>
            <div v-if="orderItems(message).length > 3" class="tz-micro-label tz-text-primary/45">
              +{{ orderItems(message).length - 3 }} more item{{ orderItems(message).length - 3 > 1 ? 's' : '' }}
            </div>
          </div>
        </div>

        <!-- FAQ 引用消息 -->
        <div
          v-else-if="isFaqMessage(message)"
            class="max-w-[82%] md:max-w-[76%] rounded-2xl border tz-border-subtle tz-surface-muted p-3 tz-text-primary shadow-lg"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="tz-caption font-semibold uppercase tracking-[0.16em] text-[#059669]">
              {{ t('chatModal.tabs.faq') }}
            </div>
            <div class="tz-micro-label whitespace-nowrap opacity-50">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>

          <div class="mt-3 flex gap-3">
            <StorefrontImage
              v-if="faqPayload(message).answer_image_url"
              :src="faqPayload(message).answer_image_url"
              :alt="faqPayload(message).answer_image_alt || faqQuestion(message)"
              class="h-16 w-16 shrink-0 rounded-xl border tz-border-subtle object-cover"
              preset="thumbnail"
            />
            <div class="min-w-0 flex-1">
              <div class="mb-1 flex flex-wrap gap-x-2 gap-y-1 tz-micro-label tz-text-primary/45">
                <span v-if="faqPayload(message).page_title || faqPayload(message).page_id">
                  {{ faqPayload(message).page_title || faqPayload(message).page_id }}
                </span>
                <span v-if="faqPayload(message).category_label || faqPayload(message).category">
                  · {{ faqPayload(message).category_label || faqPayload(message).category }}
                </span>
              </div>
              <div class="break-words text-sm font-semibold leading-5 tz-text-primary">
                {{ faqQuestion(message) }}
              </div>
              <div v-if="faqExcerpt(message)" class="mt-2 line-clamp-3 text-xs leading-5 tz-text-muted">
                {{ faqExcerpt(message) }}
              </div>
            </div>
          </div>

          <a
            v-if="faqUrl(message)"
            :href="faqUrl(message)"
            :target="isExternalFaqUrl(message) ? '_blank' : undefined"
            rel="noopener"
            class="mt-3 flex items-center justify-between gap-3 rounded-xl border tz-border-subtle tz-surface-subtle px-3 py-2 text-xs font-semibold tz-text-primary/85 transition-colors hover:border-[#059669]/45 hover:text-[#059669]"
          >
            <span>{{ t('faq.ui.viewAll') }}</span>
            <span aria-hidden="true">↗</span>
          </a>
        </div>

        <!-- 卡片类型消息 (Card) -->
        <a
          v-else-if="message.type === 'card'"
          :href="message.url || '#'"
          target="_blank"
          rel="noopener"
          class="flex gap-2.5 p-2 border tz-border-subtle rounded-2xl tz-surface-subtle hover:tz-surface-muted transition-colors max-w-[75%] md:max-w-[70%]"
        >
          <StorefrontImage
            v-if="message.thumbnail"
            :src="message.thumbnail"
            alt="thumbnail"
            class="w-14 h-14 object-cover rounded-xl md:rounded-lg"
            preset="thumbnail"
          />
          <div class="text-xs md:text-sm tz-text-primary">{{ message.title || message.message }}</div>
        </a>

        <!-- 普通/图片消息 -->
        <div
          v-else
          class="max-w-[75%] md:max-w-[70%] rounded-2xl md:rounded-xl px-3 py-2 tz-text-primary shadow-lg relative group"
          :class="[
            message.is_agent ? '' : 'tz-surface-subtle border tz-border-subtle md:bg-blue-50 md:border-blue-200'
          ]"
          :style="message.is_agent 
            ? { backgroundColor: 'var(--tz-surface-muted)', border: `1px solid ${currentThemeColor}` }
          : isDesktop ? {} : { backgroundColor: 'var(--tz-surface-subtle)', border: '1px solid var(--tz-border-subtle)' } // Mobile Visitor 样式
          "
          @touchstart="handleMessageTouchStart(message)"
          @touchend="handleMessageTouchEnd"
          @touchcancel="handleMessageTouchEnd"
          @mousedown="handleMessageMouseDown(message)"
          @mouseup="handleMessageMouseUp"
          @mouseleave="handleMessageMouseUp"
          @contextmenu.prevent="handleMessageContextMenu(message)"
        >
          <!-- 桌面端 Agent 样式覆盖 (如果需要精确对齐原版) -->
          <!-- 注意: :style 优先级高于 class。原版 PC 端 visitor 是蓝色背景。 -->
          
          <div class="tz-caption md:text-xs mb-1 opacity-70">
            {{ message.is_agent ? 'Agent' : message.sender_name }}
          </div>
          
          <div class="flex flex-col md:flex-row md:items-end gap-1 md:gap-2">
            <div class="text-sm whitespace-pre-wrap break-words flex-1">
              {{ message.message }}
            </div>
            <div class="tz-micro-label opacity-50 md:opacity-60 whitespace-nowrap self-end md:self-auto">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
          
          <div v-if="messageAttachments(message).length" class="mt-2 grid gap-2">
            <StorefrontImage
              v-for="attachmentUrl in messageAttachments(message)"
              :key="attachmentUrl"
              :src="attachmentUrl"
              alt="附件"
              class="max-w-full rounded-xl"
              preset="content"
            />
          </div>

          <a
            v-if="message.message_type === 'link' && messageMetadata(message).url"
            :href="messageMetadata(message).url"
            target="_blank"
            rel="noopener noreferrer"
            class="mt-2 block rounded-xl border tz-border-subtle tz-surface-subtle px-3 py-2 text-xs text-[#059669] underline-offset-4 hover:underline"
          >
            <span class="block font-semibold tz-text-secondary">
              {{ messageMetadata(message).title || message.message }}
            </span>
            <span class="mt-1 block truncate tz-text-primary/55">
              {{ messageMetadata(message).url }}
            </span>
          </a>

          <div
            v-if="isMessageSending(message) || isMessageFailed(message)"
            class="mt-2 flex items-center gap-2 text-[11px] font-semibold leading-4"
            :class="isMessageFailed(message) ? 'text-red-600' : 'tz-text-primary/45'"
          >
            <span>{{ isMessageFailed(message) ? 'Not delivered' : 'Sending...' }}</span>
            <button
              v-if="isMessageFailed(message)"
              type="button"
              class="rounded-full border border-red-200 bg-red-50 px-2 py-0.5 text-[11px] font-semibold text-red-700 transition-colors hover:bg-red-100 disabled:opacity-50"
              :disabled="isSending"
              @click.stop.prevent="$emit('retryMessage', message)"
              @mousedown.stop
              @touchstart.stop
            >
              Retry
            </button>
          </div>
        </div>
      </div>

      <div v-if="agentTyping?.active" class="flex justify-end">
        <div class="max-w-[75%] rounded-2xl border tz-border-subtle tz-surface-subtle px-3 py-2 tz-caption tz-text-muted shadow-lg">
          <span class="font-semibold tz-text-primary/85">{{ agentTyping.displayName || 'Agent' }}</span>
          <span class="ml-1">is typing</span>
          <span class="ml-1 inline-flex gap-0.5 align-middle">
            <span class="h-1 w-1 animate-pulse rounded-full bg-white/60"></span>
            <span class="h-1 w-1 animate-pulse rounded-full bg-white/60 [animation-delay:120ms]"></span>
            <span class="h-1 w-1 animate-pulse rounded-full bg-white/60 [animation-delay:240ms]"></span>
          </span>
        </div>
      </div>
    </div>

    <!-- 底部输入栏 -->
    <div class="chat-composer-bar tz-mobile-chrome-bottom px-3 pb-4 md:px-5 md:pt-5 md:pb-6 border-t tz-border-subtle md:tz-border-strong/[0.08]">
      <div v-if="showVisitorEmailCapture" class="chat-email-capture mb-2 rounded-2xl border tz-border-subtle tz-surface-subtle px-3 py-2">
        <label class="block tz-compact-label tz-text-primary/55">Email for follow-up · optional</label>
        <input
          :value="visitorEmail"
          @input="$emit('update:visitorEmail', ($event.target as HTMLInputElement).value)"
          type="email"
          inputmode="email"
          autocomplete="email"
          placeholder="you@example.com"
          class="chat-email-capture__input mt-1 h-8 w-full bg-transparent tz-caption tz-text-primary placeholder:tz-text-primary/35 focus:outline-none"
        />
      </div>

      <div
        v-if="pendingProductReference"
            class="mb-2 border border-[#059669]/35 bg-emerald-50 px-3 py-2 tz-text-primary shadow-[0_8px_24px_rgba(15,23,42,0.1)]"
      >
        <div class="flex items-center gap-3">
          <div class="h-12 w-12 shrink-0 overflow-hidden tz-surface-subtle">
            <StorefrontImage
              v-if="pendingProductReference.thumbnail"
              :src="pendingProductReference.thumbnail"
              :alt="pendingProductTitle"
              class="h-full w-full object-cover"
              preset="thumbnail"
            />
          </div>

          <div class="min-w-0 flex-1">
            <div class="tz-micro-label uppercase tracking-[0.14em] text-[#059669]">
              {{ t('chatModal.productDraft.selected') }}
            </div>
            <div class="truncate text-sm font-semibold tz-text-primary">
              {{ pendingProductTitle }}
            </div>
            <div class="mt-0.5 flex min-w-0 gap-2 text-xs tz-text-primary/50">
              <span v-if="pendingProductPrice" class="shrink-0 text-[#059669]">{{ pendingProductPrice }}</span>
              <span v-if="pendingProductSku" class="min-w-0 truncate">SKU: {{ pendingProductSku }}</span>
            </div>
          </div>

          <button
            type="button"
            class="flex h-8 w-8 shrink-0 items-center justify-center border tz-border-subtle tz-text-muted transition-colors hover:tz-border-strong/50 hover:tz-text-primary"
            :aria-label="t('chatModal.productDraft.remove')"
            :title="t('chatModal.productDraft.remove')"
            @click="$emit('clearPendingProductReference')"
          >
            <Icon name="lucide:x" class="h-4 w-4" />
          </button>
        </div>
      </div>

      <form
        ref="attachmentActionsRoot"
        @submit.prevent="handleSendMessage"
        class="relative flex items-center gap-2"
      >
        <input
          ref="messageInputElement"
          :value="newMessage"
          @input="$emit('update:newMessage', ($event.target as HTMLInputElement).value)"
          type="text"
          placeholder="Type a message..."
          class="flex-1 h-11 rounded-full border tz-border-strong/14 tz-surface-input px-4 tz-caption tz-text-primary placeholder:tz-text-primary/58 shadow-[inset_0_0_0_1px_rgba(15,23,42,0.04),0_2px_6px_rgba(15,23,42,0.08)] transition-colors focus:border-[#059669]/70 focus:tz-surface-card focus:outline-none md:text-base"
          :style="{ borderColor: currentThemeColor }"
          :disabled="isSending"
        />
        
        <div class="relative shrink-0">
          <input
            ref="imageLibraryInput"
            type="file"
            :accept="uploadSpecAccept('customer_service_attachment')"
            multiple
            class="hidden"
            @change="handleImageUpload($event, 'library')"
          />

          <input
            ref="cameraInput"
            type="file"
            :accept="uploadSpecAccept('customer_service_attachment')"
            capture="environment"
            class="hidden"
            @change="handleImageUpload($event, 'camera')"
          />

          <button
            type="button"
            @click="attachmentHubOpen = !attachmentHubOpen"
            :disabled="isUploadingImage || isSending"
            class="shrink-0 w-10 h-10 md:w-11 md:h-11 rounded-full tz-surface-subtle hover:bg-white/[0.18] tz-text-primary flex items-center justify-center shadow-sm disabled:opacity-50 transition-colors"
            :aria-label="t('chatModal.attachments.open')"
            :title="t('chatModal.attachments.open')"
          >
            <Icon
              :name="isUploadingImage ? 'lucide:loader-circle' : 'lucide:plus'"
              class="h-5 w-5"
              :class="{ 'animate-spin': isUploadingImage }"
            />
          </button>
        </div>

        <ChatAttachmentHub
          :open="attachmentHubOpen"
          @close="attachmentHubOpen = false"
          @select="handleAttachmentAction"
        />
        
        <button
          type="submit"
          :disabled="(!newMessage.trim() && !pendingProductReference) || isSending"
          class="shrink-0 px-4 md:px-6 h-11 rounded-full bg-[#059669] font-semibold text-sm md:text-base text-white flex items-center justify-center transition-colors hover:bg-[#047857] disabled:cursor-not-allowed disabled:bg-[#059669] disabled:text-white [&_svg]:text-white"
          title="Send message"
        >
          <span v-if="!isSending">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12l14-8-4 16-3-7-7-1z" />
            </svg>
          </span>
          <span v-else class="flex items-center gap-2">
            <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span class="hidden md:inline">Sending...</span>
          </span>
        </button>
      </form>
      <p class="mt-1 px-1 text-[10px] leading-4 tz-text-muted">
        Image attachments: {{ uploadSpecHint('customer_service_attachment') }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useThrottleFn } from '@vueuse/core'
import { useI18n, useLocalePath } from '#imports'
import ChatAttachmentHub from './ChatAttachmentHub.vue'
import type { ChatAttachmentActionId } from '~/composables/chat/useChatAttachmentActions'
import { getStorefrontLocaleEntry } from '~/utils/storefrontLocales'
import { uploadSpecAccept, uploadSpecHint } from '~/utils/uploadSpecs'

const props = defineProps<{
  messages: any[]
  newMessage: string
  visitorEmail: string
  showVisitorEmailCapture: boolean
  isSending: boolean
  isUploadingImage: boolean
  pendingProductReference?: any | null
  agentTyping?: { active: boolean; displayName?: string }
  currentThemeColor: string
}>()

const emit = defineEmits<{
  'update:newMessage': [value: string]
  'update:visitorEmail': [value: string]
  'sendMessage': []
  'uploadImage': [event: Event, source: 'library' | 'camera']
  'openOrderPicker': []
  'openCustomerServiceProductSearchModal': []
  'clearPendingProductReference': []
  'deleteMessage': [message: any]
  'retryMessage': [message: any]
}>()

const { t, locale } = useI18n()
const localePath = useLocalePath()
const messagesContainer = ref<HTMLElement | null>(null)
const messageInputElement = ref<HTMLInputElement | null>(null)
const imageLibraryInput = ref<HTMLInputElement | null>(null)
const cameraInput = ref<HTMLInputElement | null>(null)
const attachmentActionsRoot = ref<HTMLElement | null>(null)
const attachmentHubOpen = ref(false)
const isDesktop = ref(false)

const pendingProductTitle = computed(() => {
  return String(props.pendingProductReference?.title || props.pendingProductReference?.name || 'Product').trim()
})

const pendingProductPrice = computed(() => {
  return String(props.pendingProductReference?.priceLabel || props.pendingProductReference?.price || '').trim()
})

const pendingProductSku = computed(() => {
  return String(props.pendingProductReference?.sku || '').trim()
})

// 简单的视口检测，用于样式判断
const checkDesktop = () => {
  isDesktop.value = window.innerWidth >= 768
}

const throttledCheckDesktop = useThrottleFn(checkDesktop, 150)

onMounted(() => {
  checkDesktop()
  window.addEventListener('resize', throttledCheckDesktop)
  document.addEventListener('pointerdown', handleDocumentPointerDown)
  scrollToBottom()
})

onUnmounted(() => {
  window.removeEventListener('resize', throttledCheckDesktop)
  document.removeEventListener('pointerdown', handleDocumentPointerDown)
})

const handleSendMessage = () => {
  emit('sendMessage')
}

const handleImageUpload = (event: Event, source: 'library' | 'camera') => {
  emit('uploadImage', event, source)
  const target = event.target as HTMLInputElement
  if (target) target.value = ''
}

const handleAttachmentAction = (action: ChatAttachmentActionId) => {
  attachmentHubOpen.value = false

  if (action === 'image_library') {
    imageLibraryInput.value?.click()
    return
  }

  if (action === 'camera_capture') {
    cameraInput.value?.click()
    return
  }

  if (action === 'order_reference') {
    emit('openOrderPicker')
    return
  }

  emit('openCustomerServiceProductSearchModal')
}

const handleDocumentPointerDown = (event: PointerEvent) => {
  if (!attachmentHubOpen.value) return
  if (attachmentActionsRoot.value?.contains(event.target as Node)) return
  attachmentHubOpen.value = false
}

// 自动滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

watch(() => props.messages, () => {
  scrollToBottom()
}, { deep: true })

watch(() => props.pendingProductReference, (value) => {
  if (!value) return
  nextTick(() => messageInputElement.value?.focus())
})

// 格式化消息时间
const formatMessageTime = (time: string) => {
  const date = new Date(time)
  const localeIso = getStorefrontLocaleEntry(locale.value)?.iso || 'en-US'
  return date.toLocaleTimeString(localeIso, { hour: '2-digit', minute: '2-digit' })
}

const messageMetadata = (message: any) => {
  if (!message?.metadata) return {}
  if (typeof message.metadata === 'string') {
    try {
      return JSON.parse(message.metadata)
    } catch {
      return {}
    }
  }
  return message.metadata
}

const messageAttachments = (message: any) => {
  const attachments = Array.isArray(message?.attachments)
    ? message.attachments
    : (message?.attachment_url ? [message.attachment_url] : [])
  const unique = new Set<string>()
  attachments.forEach((value: any) => {
    const attachment = String(value || '').trim()
    if (attachment) unique.add(attachment)
  })
  return Array.from(unique)
}

const isConfigConfirmMessage = (message: any) => {
  return message?.message_type === 'config_confirm'
}

const isOrderMessage = (message: any) => {
  return message?.message_type === 'order'
}

const isFaqMessage = (message: any) => {
  return message?.message_type === 'faq' || message?.type === 'faq'
}

const isMessageSending = (message: any) => {
  return message?.sync_state === 'sending'
}

const isMessageFailed = (message: any) => {
  return message?.sync_state === 'failed'
}

const configProduct = (message: any) => {
  const metadata = messageMetadata(message)
  return metadata?.product || {}
}

const configSelection = (message: any) => {
  const metadata = messageMetadata(message)
  return metadata?.selections || {}
}

const configOptionRows = (message: any) => {
  const options = configSelection(message)?.options
  return Array.isArray(options) ? options : []
}

const orderPayload = (message: any) => {
  return messageMetadata(message) || {}
}

const orderItems = (message: any) => {
  const items = orderPayload(message)?.items
  return Array.isArray(items) ? items : []
}

const formatOrderTotal = (order: any) => {
  const total = Number(order?.total || 0)
  const currency = String(order?.currency || '').trim().toUpperCase()
  if (!Number.isFinite(total) || total <= 0) return currency
  return `${currency || 'Currency missing'} ${total.toFixed(2)}`
}

const faqPayload = (message: any) => {
  const metadata = messageMetadata(message)
  return metadata?.faq || metadata || {}
}

const faqQuestion = (message: any) => {
  const payload = faqPayload(message)
  return String(payload.question || payload.title || message?.message || '').trim()
}

const stripHtml = (value: any) => String(value || '')
  .replace(/<[^>]+>/g, ' ')
  .replace(/&nbsp;/g, ' ')
  .replace(/\s+/g, ' ')
  .trim()

const faqExcerpt = (message: any) => {
  const payload = faqPayload(message)
  const excerpt = stripHtml(payload.answer_excerpt || payload.answerExcerpt || payload.answer || '')
  return excerpt.length > 320 ? `${excerpt.slice(0, 317)}...` : excerpt
}

const isSafeFaqUrl = (value: any) => {
  const raw = String(value || '').trim()
  if (!raw) return false
  if (raw.startsWith('/') && !raw.startsWith('//')) return true
  try {
    const parsed = new URL(raw)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const localizeFaqUrl = (value: string) => {
  const [pathname, query = ''] = value.split('?', 2)
  if (pathname !== '/support/faqs' && pathname !== '/faq') return value

  const localizedPath = localePath('/support/faqs')
  return query ? `${localizedPath}?${query}` : localizedPath
}

const faqUrl = (message: any) => {
  const payload = faqPayload(message)
  if (isSafeFaqUrl(payload.url)) {
    const safeURL = String(payload.url).trim()
    return safeURL.startsWith('/') && !safeURL.startsWith('//')
      ? localizeFaqUrl(safeURL)
      : safeURL
  }
  const pageID = String(payload.page_id || payload.pageId || '').trim()
  const faqID = String(payload.faq_id || payload.faqId || '').trim()
  if (pageID && faqID) {
    return localizeFaqUrl(`/support/faqs?page=${encodeURIComponent(pageID)}&faq=${encodeURIComponent(faqID)}`)
  }
  return localePath('/support/faqs')
}

const isExternalFaqUrl = (message: any) => {
  const value = faqUrl(message)
  return value.startsWith('http://') || value.startsWith('https://')
}

// 长按逻辑
const messagePressTimer = ref<number | null>(null)
const pressedMessage = ref<any | null>(null)

const clearMessagePressTimer = () => {
  if (messagePressTimer.value) {
    clearTimeout(messagePressTimer.value)
    messagePressTimer.value = null
  }
  pressedMessage.value = null
}

const startMessagePress = (message: any) => {
  if (message.is_agent) return // 不能删除客服消息
  pressedMessage.value = message
  clearMessagePressTimer()
  messagePressTimer.value = window.setTimeout(() => {
    messagePressTimer.value = null
    if (pressedMessage.value) {
      emit('deleteMessage', pressedMessage.value)
      pressedMessage.value = null
    }
  }, 600)
}

const handleMessageTouchStart = (message: any) => {
  startMessagePress(message)
}

const handleMessageTouchEnd = () => {
  clearMessagePressTimer()
}

const handleMessageMouseDown = (message: any) => {
  // Only handle long press for non-touch devices when mouse button held
  if ((window as any)?.ontouchstart !== undefined) return
  startMessagePress(message)
}

const handleMessageMouseUp = () => {
  clearMessagePressTimer()
}

const handleMessageContextMenu = (message: any) => {
  emit('deleteMessage', message)
}
</script>

<style scoped>
.chat-email-capture {
  border-color: var(--tz-border-subtle);
  background: var(--tz-surface-subtle);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.8);
}

.chat-email-capture input.chat-email-capture__input {
  border: 0 !important;
  border-radius: 0.65rem;
  background: transparent !important;
  background-color: transparent !important;
  box-shadow: inset 0 -1px 0 var(--tz-border-subtle) !important;
  padding-inline: 0.25rem;
}

.chat-email-capture input.chat-email-capture__input:focus {
  background: var(--tz-card-surface) !important;
  background-color: var(--tz-card-surface) !important;
  box-shadow:
    inset 0 -1px 0 var(--tz-site-accent),
    0 0 0 1px rgba(5, 150, 105, 0.24) !important;
}

@media (max-width: 767px) {
  .chat-composer-bar {
    border-top-color: var(--tz-mobile-chrome-edge-border);
    padding-bottom: var(--tz-mobile-modal-safe-padding-bottom, 1rem);
  }
}
</style>
