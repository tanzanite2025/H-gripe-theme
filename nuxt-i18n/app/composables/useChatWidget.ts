import { ref, computed } from 'vue'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'

// 对话对象目前只在前端控制聊天窗口展示用，保持为 any 以兼容现有调用
export type ChatWidgetConversation = any

// 全局单例状态：是否打开聊天窗口 & 当前会话配置
const currentConversation = ref<ChatWidgetConversation | null>(null)
const overlayBackStack = useOverlayBackStack()

const isChatOpen = computed(() => !!currentConversation.value)

const closeChatState = () => {
  currentConversation.value = null
}

/**
 * 打开聊天窗口
 * @param conversation 用于传递给 WhatsAppChatModal 的会话配置，默认展示客服列表
 */
const openChat = (conversation: ChatWidgetConversation = { showAgentList: true }) => {
  currentConversation.value = conversation
  overlayBackStack.open('whatsapp-chat', closeChatState)
}

/** 关闭聊天窗口 */
const closeChat = () => {
  void overlayBackStack.close('whatsapp-chat')
  closeChatState()
}

/** 切换聊天窗口打开 / 关闭 */
const toggleChat = (conversation?: ChatWidgetConversation) => {
  if (isChatOpen.value) {
    closeChat()
  } else {
    openChat(conversation)
  }
}

export const useChatWidget = () => {
  return {
    // 状态
    currentConversation,
    isChatOpen,
    // 操作方法
    openChat,
    closeChat,
    toggleChat,
  }
}
