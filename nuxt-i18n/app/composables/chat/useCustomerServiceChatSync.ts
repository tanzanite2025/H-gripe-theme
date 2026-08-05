import { ref } from 'vue'
import type { ComputedRef, Ref, WritableComputedRef } from 'vue'

type AuthRequest = <T = any>(path: string, options?: any, errorMessage?: string) => Promise<T>

interface CustomerServiceChatSyncOptions {
  publicApiBase: ComputedRef<string>
  locale: Ref<string>
  conversationId: Ref<string>
  selectedAgent: Ref<any>
  messages: WritableComputedRef<any[]>
  user: Ref<any>
  visitorEmail: Ref<string>
  authRequest: AuthRequest
  saveMessagesToStorage: () => void
  scrollToBottom: () => void
}

const realtimeEventTypes = [
  'conversation.message.created',
  'conversation.messages.read',
  'conversation.assigned',
  'conversation.status.changed',
  'conversation.context.updated',
  'conversation.typing'
]

export const useCustomerServiceChatSync = ({
  publicApiBase,
  locale,
  conversationId,
  selectedAgent,
  messages,
  user,
  visitorEmail,
  authRequest,
  saveMessagesToStorage,
  scrollToBottom
}: CustomerServiceChatSyncOptions) => {
  const realtimeSource = ref<EventSource | null>(null)
  const realtimeConversationId = ref('')
  const agentTyping = ref<{ active: boolean; displayName: string }>({ active: false, displayName: '' })
  let realtimeReconnectTimer: number | null = null
  let realtimeSyncTimer: number | null = null
  let remoteTypingTimer: number | null = null

  const rememberConversationId = (payload: any) => {
    const id = payload?.conversation_id || payload?.conversationId || payload?.data?.conversation_id || payload?.data?.conversationId
    if (typeof id === 'string' && id.length > 0) {
      conversationId.value = id
    }
    return conversationId.value
  }

  const currentSenderEmail = () => {
    if (user.value?.email) return user.value.email
    const normalized = visitorEmail.value.trim().toLowerCase()
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(normalized) ? normalized : ''
  }

  const normalizeServerMessage = (message: any) => {
    if (!message || typeof message !== 'object') return null
    const normalized = {
      ...message,
      id: message.id,
      conversation_id: message.conversation_id || conversationId.value,
      sender_id: message.sender_id || 0,
      sender_name: message.sender_name || (message.is_agent ? 'Agent' : (user.value?.display_name || '访客')),
      sender_email: message.sender_email || '',
      message: message.message || message.content || '',
      message_type: message.message_type || 'text',
      metadata: message.metadata || null,
      attachment_url: message.attachment_url || '',
      attachments: Array.isArray(message.attachments)
        ? message.attachments
        : (message.attachment_url ? [message.attachment_url] : []),
      source: message.source || '',
      created_at: message.created_at || new Date().toISOString(),
      is_agent: !!message.is_agent,
      sync_state: 'persisted'
    }

    if (normalized.message_type === 'product' && normalized.metadata) {
      return {
        ...normalized,
        type: 'card',
        title: normalized.metadata.title || normalized.message,
        url: normalized.metadata.url || '#',
        thumbnail: normalized.metadata.thumbnail || ''
      }
    }

    if (normalized.message_type === 'config_confirm') {
      return {
        ...normalized,
        type: 'config_confirm'
      }
    }

    if (normalized.message_type === 'order') {
      return {
        ...normalized,
        type: 'order'
      }
    }

    if (normalized.message_type === 'faq') {
      return {
        ...normalized,
        type: 'faq'
      }
    }

    return normalized
  }

  const extractMessageItems = (payload: any) => {
    if (Array.isArray(payload)) return payload
    if (Array.isArray(payload?.data)) return payload.data
    if (Array.isArray(payload?.messages)) return payload.messages
    if (Array.isArray(payload?.data?.messages)) return payload.data.messages
    return []
  }

  const mergePersistedMessages = (serverMessages: any[]) => {
    const localById = new Map(messages.value.map((message) => [String(message.id), message]))
    const persisted = serverMessages
      .map((message) => {
        const normalized = normalizeServerMessage(message)
        if (!normalized) return null

        const local = localById.get(String(normalized.id))
        if (!local) return normalized

        return {
          ...normalized,
          metadata: normalized.metadata || local.metadata || null,
          message_type: normalized.message_type || local.message_type || 'text',
          attachment_url: normalized.attachment_url || local.attachment_url || '',
          attachments: normalized.attachments?.length ? normalized.attachments : (local.attachments || []),
          source: normalized.source || local.source || '',
          type: local.type || normalized.type,
          title: local.title || normalized.title,
          url: local.url || normalized.url,
          thumbnail: local.thumbnail || normalized.thumbnail
        }
      })
      .filter(Boolean)

    const localOnly = messages.value.filter((message) => ['sending', 'failed'].includes(message.sync_state))
    const merged = [...persisted, ...localOnly]
    const uniqueById = new Map<string, any>()
    merged.forEach((message) => {
      uniqueById.set(String(message.id), message)
    })

    messages.value = Array.from(uniqueById.values()).sort((left, right) => {
      return new Date(left.created_at || 0).getTime() - new Date(right.created_at || 0).getTime()
    })
    saveMessagesToStorage()
  }

  const replaceLocalMessageWithServerMessage = (localId: number | string, payload: any) => {
    const persisted = normalizeServerMessage(payload?.data || payload?.message || payload)
    if (!persisted) return

    const index = messages.value.findIndex((message) => String(message.id) === String(localId))
    if (index < 0) return

    const local = messages.value[index]
    messages.value[index] = {
      ...persisted,
      metadata: persisted.metadata || local.metadata || null,
      message_type: persisted.message_type || local.message_type || 'text',
      attachment_url: persisted.attachment_url || local.attachment_url || '',
      attachments: persisted.attachments?.length ? persisted.attachments : (local.attachments || []),
      source: persisted.source || local.source || '',
      type: local.type || persisted.type,
      title: local.title || persisted.title,
      url: local.url || persisted.url,
      thumbnail: local.thumbnail || persisted.thumbnail
    }
    saveMessagesToStorage()
  }

  const markLocalMessageFailed = (localId: number | string) => {
    const message = messages.value.find((item) => String(item.id) === String(localId))
    if (!message) return
    message.sync_state = 'failed'
    saveMessagesToStorage()
  }

  const ensureCustomerServiceConversation = async () => {
    if (conversationId.value) return conversationId.value
    const response = await $fetch<any>(`${publicApiBase.value}/customer-service/conversations`, {
      method: 'POST',
      credentials: 'include',
      body: {
        agent_id: selectedAgent.value?.id ? String(selectedAgent.value.id) : '',
        locale: locale.value
      }
    })
    const id = rememberConversationId(response)
    if (!id) {
      throw new Error('[CRITICAL] conversation_id missing in customer-service conversation response')
    }
    return id
  }

  const loadMessagesFromAPI = async () => {
    if (!conversationId.value) return
    try {
      const response = await $fetch<any>(`${publicApiBase.value}/customer-service/messages/${encodeURIComponent(conversationId.value)}`, {
        credentials: 'include',
        params: {
          limit: 100,
          offset: 0
        }
      })
      mergePersistedMessages(extractMessageItems(response))
      scrollToBottom()
    } catch (error) {
      console.warn('同步客服消息失败:', error)
    }
  }

  const sendMessageToAPI = async (messageData: any) => {
    try {
      const currentConversationId = await ensureCustomerServiceConversation()
      const response = await authRequest<any>(
        '/customer-service/messages',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            conversation_id: currentConversationId,
            message: messageData.message,
            sender_type: user.value ? 'user' : 'visitor',
            sender_name: user.value?.display_name || '访客',
            sender_email: currentSenderEmail(),
            agent_id: selectedAgent.value?.id || '',
            locale: locale.value,
            message_type: messageData.message_type || 'text',
            metadata: messageData.metadata || null,
            attachment_url: messageData.attachment_url || '',
            attachments: Array.isArray(messageData.attachments) ? messageData.attachments : []
          })
        },
        'Failed to send customer-service message'
      )
      rememberConversationId(response)
      return response
    } catch (error) {
      console.error('发送消息到API失败:', error)
      throw error
    }
  }

  const uploadCustomerServiceAttachment = async (
    file: File,
    source: 'library' | 'camera' = 'library'
  ) => {
    const currentConversationId = await ensureCustomerServiceConversation()
    const formData = new FormData()
    formData.append('conversation_id', currentConversationId)
    formData.append('source', source)
    formData.append('file', file)

    const response = await $fetch<any>(`${publicApiBase.value}/customer-service/attachments`, {
      method: 'POST',
      credentials: 'include',
      body: formData
    })

    const asset = response?.asset || response?.data?.asset || response?.data || null
    if (!asset?.url) {
      throw new Error('[CRITICAL] customer-service attachment upload returned no asset URL')
    }
    return asset
  }

  const customerServiceEventURL = (currentConversationId: string) => {
    const query = new URLSearchParams({ conversation_id: currentConversationId })
    return `${publicApiBase.value}/customer-service/events?${query.toString()}`
  }

  const clearAgentTyping = () => {
    agentTyping.value = { active: false, displayName: '' }
    if (import.meta.client && remoteTypingTimer) {
      window.clearTimeout(remoteTypingTimer)
      remoteTypingTimer = null
    }
  }

  const closeCustomerServiceRealtimeSource = () => {
    if (realtimeSource.value) {
      realtimeSource.value.close()
      realtimeSource.value = null
    }
  }

  const closeCustomerServiceRealtime = () => {
    closeCustomerServiceRealtimeSource()
    realtimeConversationId.value = ''

    if (import.meta.client && realtimeReconnectTimer) {
      window.clearTimeout(realtimeReconnectTimer)
      realtimeReconnectTimer = null
    }
    if (import.meta.client && realtimeSyncTimer) {
      window.clearTimeout(realtimeSyncTimer)
      realtimeSyncTimer = null
    }
    clearAgentTyping()
  }

  const scheduleRealtimeMessageSync = () => {
    if (!import.meta.client) return
    if (realtimeSyncTimer) {
      window.clearTimeout(realtimeSyncTimer)
    }

    realtimeSyncTimer = window.setTimeout(async () => {
      realtimeSyncTimer = null
      await loadMessagesFromAPI()
    }, 300)
  }

  const handleCustomerServiceRealtimeEvent = (event: MessageEvent) => {
    try {
      const payload = JSON.parse(event.data || '{}')
      if (payload?.conversation_id && payload.conversation_id !== conversationId.value) return
      if (payload?.type === 'conversation.typing') {
        handleCustomerServiceTypingEvent(payload)
        return
      }
      scheduleRealtimeMessageSync()
    } catch (error) {
      console.warn('解析客服实时事件失败:', error)
    }
  }

  const handleCustomerServiceTypingEvent = (event: any) => {
    if (event?.actor?.kind !== 'agent') return

    const payload = event.payload || {}
    if (payload.is_typing === false) {
      clearAgentTyping()
      return
    }

    agentTyping.value = {
      active: true,
      displayName: payload.display_name || selectedAgent.value?.name || 'Agent'
    }

    if (import.meta.client) {
      if (remoteTypingTimer) {
        window.clearTimeout(remoteTypingTimer)
      }
      const expiresAt = Date.parse(payload.expires_at || '')
      const timeout = Number.isFinite(expiresAt)
        ? Math.max(1200, Math.min(5000, expiresAt - Date.now()))
        : 3500
      remoteTypingTimer = window.setTimeout(() => {
        clearAgentTyping()
      }, timeout)
    }
  }

  const sendTypingIndicator = async (isTyping = true) => {
    if (!import.meta.client) return
    try {
      const currentConversationId = isTyping
        ? await ensureCustomerServiceConversation()
        : conversationId.value
      if (!currentConversationId) return

      await authRequest(
        '/customer-service/typing',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            conversation_id: currentConversationId,
            is_typing: isTyping,
            display_name: user.value?.display_name || user.value?.username || 'Visitor'
          })
        },
        'Failed to send customer-service typing signal'
      )
    } catch (error) {
      // Typing is transient UI state. Message send/read must keep working even if
      // this best-effort signal is unavailable.
      console.warn('客服 typing 状态上报失败:', error)
    }
  }

  const connectCustomerServiceRealtime = () => {
    if (!import.meta.client || typeof window === 'undefined' || !('EventSource' in window)) return
    if (!conversationId.value) return
    if (realtimeSource.value && realtimeConversationId.value === conversationId.value) return

    closeCustomerServiceRealtimeSource()
    realtimeConversationId.value = conversationId.value

    const source = new EventSource(customerServiceEventURL(conversationId.value), { withCredentials: true })
    realtimeSource.value = source

    realtimeEventTypes.forEach((eventType) => {
      source.addEventListener(eventType, handleCustomerServiceRealtimeEvent)
    })

    source.onerror = () => {
      const reconnectConversationId = realtimeConversationId.value
      if (realtimeSource.value === source) {
        realtimeSource.value = null
      }
      source.close()

      if (!reconnectConversationId || reconnectConversationId !== conversationId.value) return
      if (!realtimeReconnectTimer) {
        realtimeReconnectTimer = window.setTimeout(() => {
          realtimeReconnectTimer = null
          if (conversationId.value === reconnectConversationId) {
            connectCustomerServiceRealtime()
          }
        }, 5000)
      }
    }
  }

  return {
    rememberConversationId,
    currentSenderEmail,
    ensureCustomerServiceConversation,
    loadMessagesFromAPI,
    sendMessageToAPI,
    uploadCustomerServiceAttachment,
    sendTypingIndicator,
    replaceLocalMessageWithServerMessage,
    markLocalMessageFailed,
    connectCustomerServiceRealtime,
    closeCustomerServiceRealtime,
    agentTyping,
    clearAgentTyping
  }
}
