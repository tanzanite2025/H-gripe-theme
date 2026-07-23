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
        <!-- 卡片类型消息 (Card) -->
        <a
          v-if="message.type === 'card'"
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
          
          <div class="text-[11px] md:text-xs mb-1 opacity-70">
            {{ message.is_agent ? 'Agent' : message.sender_name }}
          </div>
          
          <div class="flex flex-col md:flex-row md:items-end gap-1 md:gap-2">
            <div class="text-sm whitespace-pre-wrap break-words flex-1">
              {{ message.message }}
            </div>
            <div class="text-[10px] opacity-50 md:opacity-60 whitespace-nowrap self-end md:self-auto">
              {{ formatMessageTime(message.created_at) }}
            </div>
          </div>
          
          <div v-if="message.attachment_url" class="mt-2">
            <img :src="message.attachment_url" alt="附件" class="max-w-full rounded-xl" />
          </div>
        </div>
      </div>
    </div>

    <!-- 底部输入栏 -->
    <div class="px-3 pb-4 md:p-4 border-t border-white/15 md:border-white/[0.08] md:bg-white/[0.02]">
      <form @submit.prevent="handleSendMessage" class="flex items-center gap-2">
        <input
          :value="newMessage"
          @input="$emit('update:newMessage', ($event.target as HTMLInputElement).value)"
          type="text"
          placeholder="Type a message..."
          class="flex-1 h-11 px-4 rounded-full text-sm md:text-base text-white bg-[linear-gradient(135deg,rgba(15,23,42,0.98),rgba(15,23,42,0.96))] shadow-[0_2px_6px_-3px_rgba(0,0,0,0.9),0_0_6px_rgba(15,23,42,0.7)] focus:outline-none focus:[box-shadow:0_0_0_1px_rgba(56,189,248,0.9)] transition-colors"
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
  isSending: boolean
  isUploadingImage: boolean
  currentThemeColor: string
}>()

const emit = defineEmits<{
  'update:newMessage': [value: string]
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
