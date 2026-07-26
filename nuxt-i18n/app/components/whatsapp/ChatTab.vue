<template>
  <div class="flex flex-col h-full min-h-0">
    <!-- 消息列表区域 (Conversation History) -->
    <div 
      ref="messagesContainer"
      class="flex-1 overflow-y-auto space-y-3 px-1 md:p-6 md:space-y-4"
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
          class="max-w-[82%] md:max-w-[76%] rounded-2xl border border-[#a5b4fc]/45 bg-[#111827]/90 p-3 text-white shadow-lg"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="tz-caption font-semibold uppercase tracking-[0.16em] text-[#a5b4fc]">
              Configuration request
            </div>
            <div class="tz-micro-label opacity-50 whitespace-nowrap">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
          <div class="mt-2 flex gap-3">
            <img
              v-if="configProduct(message).thumbnail"
              :src="configProduct(message).thumbnail"
              alt="Product"
              class="h-16 w-16 shrink-0 rounded-xl object-cover"
            />
            <div class="min-w-0 flex-1">
              <a
                v-if="configProduct(message).url"
                :href="configProduct(message).url"
                target="_blank"
                rel="noopener"
                class="block truncate text-sm font-semibold text-white hover:underline"
              >
                {{ configProduct(message).title || message.message }}
              </a>
              <div v-else class="truncate text-sm font-semibold text-white">
                {{ configProduct(message).title || message.message }}
              </div>
              <div v-if="configProduct(message).price" class="mt-1 text-xs text-[#40ffaa]">
                {{ configProduct(message).price }}
              </div>
              <div v-if="configProduct(message).sku" class="mt-1 tz-caption text-white/55">
                SKU: {{ configProduct(message).sku }}
              </div>
            </div>
          </div>
          <div class="mt-3 rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2 tz-caption leading-5 text-white/70">
            <div v-if="configSelection(message).variant_title" class="font-semibold text-white/85">
              Selected: {{ configSelection(message).variant_title }}
            </div>
            <div
              v-if="configOptionRows(message).length"
              class="mt-2 grid grid-cols-1 gap-1.5 sm:grid-cols-2"
            >
              <div
                v-for="option in configOptionRows(message)"
                :key="option.key"
                class="rounded-lg bg-white/[0.05] px-2 py-1.5"
              >
                <span class="block tz-compact-label text-white/45">
                  {{ option.label }}
                </span>
                <span class="font-semibold text-white/80">
                  {{ option.value }}<span v-if="option.unit"> {{ option.unit }}</span>
                </span>
              </div>
            </div>
            <div v-if="configSelection(message).weight_grams" class="mt-2 tz-caption text-white/60">
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
          class="max-w-[82%] md:max-w-[76%] rounded-2xl border border-[#40ffaa]/35 bg-[#07120b]/85 p-3 text-white shadow-lg"
        >
          <div class="flex items-center justify-between gap-3">
            <div class="tz-caption font-semibold uppercase tracking-[0.16em] text-[#40ffaa]">
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
              class="block truncate text-sm font-semibold text-white hover:underline"
            >
              {{ orderPayload(message).title || message.message }}
            </a>
            <div v-else class="truncate text-sm font-semibold text-white">
              {{ orderPayload(message).title || message.message }}
            </div>
            <div class="mt-1 flex flex-wrap gap-x-2 gap-y-1 tz-caption text-white/55">
              <span v-if="orderPayload(message).status">Status: {{ orderPayload(message).status }}</span>
              <span v-if="orderPayload(message).payment_status">Payment: {{ orderPayload(message).payment_status }}</span>
              <span v-if="orderPayload(message).shipping_status">Shipping: {{ orderPayload(message).shipping_status }}</span>
            </div>
            <div class="mt-2 text-xs font-semibold text-[#40ffaa]">
              {{ formatOrderTotal(orderPayload(message)) }}
            </div>
          </div>
          <div v-if="orderItems(message).length" class="mt-3 space-y-1.5">
            <div
              v-for="item in orderItems(message).slice(0, 3)"
              :key="item.id || `${item.product_id}-${item.sku}`"
              class="rounded-xl border border-white/10 bg-white/[0.04] px-3 py-2 tz-caption"
            >
              <div class="flex items-center justify-between gap-2">
                <span class="truncate font-semibold text-white/85">{{ item.title || item.product_name || 'Product' }}</span>
                <span class="shrink-0 text-white/60">x{{ item.quantity || 1 }}</span>
              </div>
              <div v-if="item.sku" class="mt-1 text-white/45">SKU: {{ item.sku }}</div>
            </div>
            <div v-if="orderItems(message).length > 3" class="tz-micro-label text-white/45">
              +{{ orderItems(message).length - 3 }} more item{{ orderItems(message).length - 3 > 1 ? 's' : '' }}
            </div>
          </div>
        </div>

        <!-- 卡片类型消息 (Card) -->
        <a
          v-else-if="message.type === 'card'"
          :href="message.url || '#'"
          target="_blank"
          rel="noopener"
          class="flex gap-2.5 p-2 border border-white/20 rounded-2xl bg-black/40 hover:bg-white/[0.10] transition-colors max-w-[75%] md:max-w-[70%]"
        >
          <img
            v-if="message.thumbnail"
            :src="message.thumbnail"
            alt="thumbnail"
            class="w-14 h-14 object-cover rounded-xl md:rounded-lg"
          />
          <div class="text-xs md:text-sm text-white">{{ message.title || message.message }}</div>
        </a>

        <!-- 普通/图片消息 -->
        <div
          v-else
          class="max-w-[75%] md:max-w-[70%] rounded-2xl md:rounded-xl px-3 py-2 text-white shadow-lg relative group"
          :class="[
            message.is_agent ? '' : 'bg-[rgba(255,255,255,0.08)] border border-[rgba(255,255,255,0.2)] md:bg-[rgba(64,122,255,0.35)] md:border-[rgba(64,122,255,0.6)]'
          ]"
          :style="message.is_agent 
            ? { backgroundColor: 'rgba(0,0,0,0.4)', border: `1px solid ${currentThemeColor}` } // Agent 样式: 深色背景 + 主题色边框
            : isDesktop ? {} : { backgroundColor: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.2)' } // Mobile Visitor 样式
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
          
          <div v-if="message.attachment_url" class="mt-2">
            <img :src="message.attachment_url" alt="附件" class="max-w-full rounded-xl" />
          </div>
        </div>
      </div>

      <div v-if="agentTyping?.active" class="flex justify-end">
        <div class="max-w-[75%] rounded-2xl border border-white/15 bg-black/35 px-3 py-2 tz-caption text-white/70 shadow-lg">
          <span class="font-semibold text-white/85">{{ agentTyping.displayName || 'Agent' }}</span>
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
    <div class="px-3 pb-4 md:p-4 border-t border-white/15 md:border-white/[0.08] md:bg-white/[0.02]">
      <div v-if="showVisitorEmailCapture" class="mb-2 rounded-2xl border border-white/10 bg-white/[0.04] px-3 py-2">
        <label class="block tz-compact-label text-white/55">Email for follow-up · optional</label>
        <input
          :value="visitorEmail"
          @input="$emit('update:visitorEmail', ($event.target as HTMLInputElement).value)"
          type="email"
          inputmode="email"
          autocomplete="email"
          placeholder="you@example.com"
          class="mt-1 h-8 w-full bg-transparent tz-caption text-white placeholder:text-white/35 focus:outline-none"
        />
      </div>
      <form @submit.prevent="handleSendMessage" class="flex items-center gap-2">
        <input
          :value="newMessage"
          @input="$emit('update:newMessage', ($event.target as HTMLInputElement).value)"
          type="text"
          placeholder="Type a message..."
          class="flex-1 h-11 px-4 rounded-full tz-caption md:text-base text-white bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)] transition-colors"
          :style="{ borderColor: currentThemeColor }"
          :disabled="isSending"
        />
        
        <input
          ref="imageInput"
          type="file"
          accept="image/*"
          class="hidden"
          @change="handleImageUpload"
        />
        
        <button
          type="button"
          @click="imageInput?.click()"
          :disabled="isUploadingImage"
          class="shrink-0 w-10 h-10 md:w-11 md:h-11 rounded-full bg-white/[0.08] hover:bg-white/[0.18] text-white flex items-center justify-center shadow-sm shadow-black/40 disabled:opacity-50 transition-colors"
          title="Upload image"
        >
          <svg v-if="!isUploadingImage" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
          <svg v-else class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </button>
        
        <button
          type="submit"
          :disabled="!newMessage.trim() || isSending"
          class="shrink-0 px-4 md:px-6 h-11 rounded-full font-semibold text-sm md:text-base text-black flex items-center justify-center transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          :style="{ backgroundColor: currentThemeColor }"
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
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useThrottleFn } from '@vueuse/core'

const props = defineProps<{
  messages: any[]
  newMessage: string
  visitorEmail: string
  showVisitorEmailCapture: boolean
  isSending: boolean
  isUploadingImage: boolean
  agentTyping?: { active: boolean; displayName?: string }
  currentThemeColor: string
}>()

const emit = defineEmits<{
  'update:newMessage': [value: string]
  'update:visitorEmail': [value: string]
  'sendMessage': []
  'uploadImage': [event: Event]
  'deleteMessage': [message: any]
}>()

const messagesContainer = ref<HTMLElement | null>(null)
const imageInput = ref<HTMLInputElement | null>(null)
const isDesktop = ref(false)

// 简单的视口检测，用于样式判断
const checkDesktop = () => {
  isDesktop.value = window.innerWidth >= 768
}

const throttledCheckDesktop = useThrottleFn(checkDesktop, 150)

onMounted(() => {
  checkDesktop()
  window.addEventListener('resize', throttledCheckDesktop)
  scrollToBottom()
})

onUnmounted(() => {
  window.removeEventListener('resize', throttledCheckDesktop)
})

const handleSendMessage = () => {
  emit('sendMessage')
}

const handleImageUpload = (event: Event) => {
  emit('uploadImage', event)
  //重置 input
  if (imageInput.value) imageInput.value.value = ''
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

// 格式化消息时间
const formatMessageTime = (time: string) => {
  const date = new Date(time)
  return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
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

const isConfigConfirmMessage = (message: any) => {
  return message?.message_type === 'config_confirm'
}

const isOrderMessage = (message: any) => {
  return message?.message_type === 'order'
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
  const currency = order?.currency || 'USD'
  if (!Number.isFinite(total) || total <= 0) return currency
  return `${currency} ${total.toFixed(2)}`
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
