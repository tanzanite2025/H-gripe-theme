import { ref, computed } from 'vue'

export interface ChatMessage {
  id: number
  conversation_id: number
  sender_id: number
  sender_name: string
  sender_avatar?: string
  message: string
  attachment_url?: string
  created_at: string
  is_read: boolean
  is_agent: boolean // true = 客服，false = 客户
}

export interface Conversation {
  id: number
  customer_id: number
  customer_name: string
  customer_avatar?: string
  agent_id?: number
  status: 'active' | 'closed' | 'pending'
  unread_count: number
  last_message?: string
  last_message_time?: string
  created_at: string
  updated_at: string
}

// 全局状态
const conversations = ref<Conversation[]>([])
const currentConversation = ref<Conversation | null>(null)
const messages = ref<ChatMessage[]>([])
const isCustomerListOpen = ref(false)
const isChatOpen = ref(false)

export const useChat = () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => {
    const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    return base.replace(/\/$/, '')
  })

  /**
   * 加载客户列表（会话列表）
   */
  const loadConversations = async () => {
    try {
      const response = await $fetch<{ data?: { items?: Conversation[] }, conversations?: Conversation[] }>(
        `${apiBase.value}/tanzanite/v1/customer-service/agent/conversations`,
        {
          credentials: 'include',
          headers: { accept: 'application/json' }
        }
      )
      conversations.value = response.data?.items || response.conversations || []
    } catch (error) {
      console.error('Failed to load conversations:', error)
    }
  }

  /**
   * 加载某个会话的消息列表
   */
  const loadMessages = async (conversationId: number) => {
    try {
      const response = await $fetch<{ data?: { items?: ChatMessage[] }, messages?: ChatMessage[] }>(
        `${apiBase.value}/tanzanite/v1/customer-service/agent/conversations/${conversationId}/messages`,
        {
          credentials: 'include',
          headers: { accept: 'application/json' }
        }
      )
      messages.value = response.data?.items || response.messages || []

      // 标记为已读
      await markAsRead(conversationId)
    } catch (error) {
      console.error('Failed to load messages:', error)
    }
  }

  /**
   * 发送消息
   */
  const sendMessage = async (conversationId: number, message: string, attachmentUrl?: string) => {
    try {
      const response = await $fetch<{ message: ChatMessage }>(
        `${apiBase.value}/tanzanite/v1/customer-service/agent/messages`,
        {
          method: 'POST',
          credentials: 'include',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: {
            conversation_id: conversationId,
            message,
            attachment_url: attachmentUrl
          }
        }
      )

      // 添加到消息列表
      if (response.message) {
        messages.value.push(response.message)
      }

      return { success: true, message: response.message }
    } catch (error) {
      console.error('Failed to send message:', error)
      return { success: false, error }
    }
  }

  /**
   * 标记会话为已读
   */
  const markAsRead = async (conversationId: number) => {
    try {
      await $fetch(
        `${apiBase.value}/tanzanite/v1/customer-service/agent/messages/read`,
        {
          method: 'POST',
          credentials: 'include',
          headers: { accept: 'application/json' },
          body: { conversation_id: conversationId }
        }
      )

      // 更新本地未读数
      const conv = conversations.value.find(c => c.id === conversationId)
      if (conv) {
        conv.unread_count = 0
      }
    } catch (error) {
      console.error('Failed to mark as read:', error)
    }
  }

  /**
   * 获取未读消息总数
   */
  const totalUnreadCount = computed(() => {
    return conversations.value.reduce((sum, conv) => sum + conv.unread_count, 0)
  })

  /**
   * 打开客户列表
   */
  const openCustomerList = async () => {
    await loadConversations()
    isCustomerListOpen.value = true
  }

  /**
   * 关闭客户列表
   */
  const closeCustomerList = () => {
    isCustomerListOpen.value = false
  }

  /**
   * 打开聊天窗口
   */
  const openChat = async (conversation: Conversation) => {
    currentConversation.value = conversation
    await loadMessages(conversation.id)
    isChatOpen.value = true
    isCustomerListOpen.value = false
  }

  /**
   * 关闭聊天窗口
   */
  const closeChat = () => {
    isChatOpen.value = false
    currentConversation.value = null
    messages.value = []
  }

  /**
   * 返回客户列表
   */
  const backToCustomerList = () => {
    isChatOpen.value = false
    currentConversation.value = null
    messages.value = []
    isCustomerListOpen.value = true
  }

  /**
   * 实时轮询新消息（可选）
   */
  const startPolling = (interval = 5000) => {
    const pollInterval = setInterval(async () => {
      if (isChatOpen.value && currentConversation.value) {
        await loadMessages(currentConversation.value.id)
      }
      if (isCustomerListOpen.value || isChatOpen.value) {
        await loadConversations()
      }
    }, interval)

    return () => clearInterval(pollInterval)
  }

  return {
    // 状态
    conversations,
    currentConversation,
    messages,
    isCustomerListOpen,
    isChatOpen,
    totalUnreadCount,

    // 方法
    loadConversations,
    loadMessages,
    sendMessage,
    markAsRead,
    openCustomerList,
    closeCustomerList,
    openChat,
    closeChat,
    backToCustomerList,
    startPolling,
  }
}
