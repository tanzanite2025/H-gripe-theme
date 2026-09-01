import { ref } from 'vue'
import type { ComputedRef, Ref, WritableComputedRef } from 'vue'
import { normalizeStoredChatMessage } from '~/composables/chat/useChatStorage'
import { validateStorefrontUploadFile } from '~/utils/uploadSpecs'

type AuthRequest = <T = any>(path: string, options?: any, errorMessage?: string) => Promise<T>

interface CustomerServiceChatSyncOptions {
  apiBase: ComputedRef<string>
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

export const useCustomerServiceChatSync = ({
  apiBase,
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
  const realtimeSocket = ref<WebSocket | null>(null)
  const realtimeConversationId = ref('')
  const realtimeCursorConversationId = ref('')
  const agentTyping = ref<{ active: boolean; displayName: string }>({ active: false, displayName: '' })
  let realtimeReconnectTimer: number | null = null
  let realtimeSyncTimer: number | null = null
  let remoteTypingTimer: number | null = null
  let realtimeLastEventId = ''
  let realtimeReconnectEnabled = false
  let realtimeReconnectAttempt = 0
  let browserRecoveryListenersAttached = false
  const seenRealtimeEventIds = new Set<string>()

  const customerServiceTimezoneHeaders = (): Record<string, string> => {
    if (!import.meta.client) return {}
    try {
      const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone
      return timezone ? { 'X-Timezone': timezone } : {}
    } catch {
      return {}
    }
  }

  const getHttpStatus = (error: unknown): number | null => {
    if (!error || typeof error !== 'object') return null

    const candidates = [
      (error as any).status,
      (error as any).statusCode,
      (error as any).response?.status,
      (error as any).response?.statusCode,
      (error as any).cause?.status,
      (error as any).cause?.statusCode,
    ]

    for (const candidate of candidates) {
      const status = Number(candidate)
      if (Number.isInteger(status)) return status
    }

    return null
  }

  const isAccessDeniedError = (error: unknown): boolean => {
    const status = getHttpStatus(error)
    if (status === 401 || status === 403) return true

    const message = error instanceof Error ? error.message : String(error || '')
    return /access denied|forbidden|unauthori[sz]ed/i.test(message)
  }

  const clearCustomerServiceConversationState = () => {
    closeCustomerServiceRealtime()
    realtimeReconnectAttempt = 0
    realtimeCursorConversationId.value = ''
    realtimeLastEventId = ''
    seenRealtimeEventIds.clear()
    conversationId.value = ''
  }

  const postCustomerServiceMessage = async (currentConversationId: string, messageData: any) => {
    const response = await authRequest<any>(
      '/customer-service/messages',
      {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
          ...customerServiceTimezoneHeaders()
        },
        body: JSON.stringify({
          conversation_id: currentConversationId,
          message: messageData.message,
          sender_type: user.value ? 'user' : 'visitor',
          sender_name: user.value?.display_name || '访客',
          sender_email: currentSenderEmail(),
          agent_id: selectedAgent.value?.id != null ? String(selectedAgent.value.id) : '',
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
  }

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

    const now = Date.now()
    const localOnly = messages.value
      .map((message) => normalizeStoredChatMessage(message, now))
      .filter((message) => message && ['sending', 'failed'].includes(message.sync_state))
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
    const response = await authRequest<any>(
      '/customer-service/conversations',
      {
        method: 'POST',
        headers: {
          accept: 'application/json',
          'Content-Type': 'application/json',
          ...customerServiceTimezoneHeaders()
        },
        body: JSON.stringify({
          agent_id: selectedAgent.value?.id ? String(selectedAgent.value.id) : '',
          locale: locale.value
        })
      },
      'Failed to start customer-service conversation'
    )
    const id = rememberConversationId(response)
    if (!id) {
      throw new Error('[CRITICAL] conversation_id missing in customer-service conversation response')
    }
    return id
  }

  const loadMessagesFromAPI = async () => {
    if (!conversationId.value) return
    try {
      const response = await authRequest<any>(`/customer-service/messages/${encodeURIComponent(conversationId.value)}`, {
        credentials: 'include',
        headers: customerServiceTimezoneHeaders(),
        params: {
          limit: 100,
          offset: 0
        }
      }, 'Failed to sync customer-service messages')
      mergePersistedMessages(extractMessageItems(response))
      scrollToBottom()
    } catch (error) {
      if (isAccessDeniedError(error)) {
        clearCustomerServiceConversationState()
        return
      }
      console.warn('同步客服消息失败:', error)
    }
  }

  const sendMessageToAPI = async (messageData: any) => {
    try {
      const currentConversationId = await ensureCustomerServiceConversation()
      return await postCustomerServiceMessage(currentConversationId, messageData)
    } catch (error) {
      if (!isAccessDeniedError(error)) {
        console.error('发送消息到API失败:', error)
        throw error
      }

      clearCustomerServiceConversationState()

      try {
        const currentConversationId = await ensureCustomerServiceConversation()
        return await postCustomerServiceMessage(currentConversationId, messageData)
      } catch (retryError) {
        console.error('发送消息到API失败:', retryError)
        throw retryError
      }
    }
  }

  const uploadCustomerServiceAttachment = async (
    file: File,
    source: 'library' | 'camera' = 'library'
  ) => {
    const validation = await validateStorefrontUploadFile(file, 'customer_service_attachment')
    if (!validation.ok) {
      throw new Error(validation.error || 'Customer service image does not meet the upload requirements.')
    }
    const currentConversationId = await ensureCustomerServiceConversation()
    const formData = new FormData()
    formData.append('conversation_id', currentConversationId)
    formData.append('source', source)
    formData.append('file', file)

    const response = await authRequest<any>(
      '/customer-service/attachments',
      {
        method: 'POST',
        headers: customerServiceTimezoneHeaders(),
        body: formData
      },
      'Failed to upload customer-service attachment'
    )

    const asset = response?.asset || response?.data?.asset || response?.data || null
    if (!asset?.url) {
      throw new Error('[CRITICAL] customer-service attachment upload returned no asset URL')
    }
    return asset
  }

  const customerServiceWebSocketURL = (currentConversationId: string, lastEventId = '') => {
    const query = new URLSearchParams({ conversation_id: currentConversationId })
    if (lastEventId) {
      query.set('last_event_id', lastEventId)
    }

    const baseURL = new URL(apiBase.value || '/api/v1', window.location.origin)
    baseURL.pathname = `${baseURL.pathname.replace(/\/$/, '')}/customer-service/ws`
    baseURL.search = query.toString()
    baseURL.hash = ''
    baseURL.protocol = baseURL.protocol === 'https:' ? 'wss:' : 'ws:'
    return baseURL.toString()
  }

  const rememberRealtimeEvent = (eventId: unknown) => {
    if (typeof eventId !== 'string' || !eventId) return true
    if (seenRealtimeEventIds.has(eventId)) return false
    if (seenRealtimeEventIds.size >= 2048) {
      seenRealtimeEventIds.clear()
    }
    seenRealtimeEventIds.add(eventId)
    return true
  }

  const clearAgentTyping = () => {
    agentTyping.value = { active: false, displayName: '' }
    if (import.meta.client && remoteTypingTimer) {
      window.clearTimeout(remoteTypingTimer)
      remoteTypingTimer = null
    }
  }

  const closeCustomerServiceRealtimeSocket = () => {
    const socket = realtimeSocket.value
    realtimeSocket.value = null
    if (socket && socket.readyState < WebSocket.CLOSING) {
      socket.close()
    }
  }

  const clearRealtimeReconnectTimer = () => {
    if (import.meta.client && realtimeReconnectTimer) {
      window.clearTimeout(realtimeReconnectTimer)
      realtimeReconnectTimer = null
    }
  }

  const recoverCustomerServiceRealtime = () => {
    if (!realtimeReconnectEnabled || !conversationId.value) return
    void loadMessagesFromAPI()
    connectCustomerServiceRealtime()
  }

  const handleCustomerServiceRealtimeVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      recoverCustomerServiceRealtime()
    }
  }

  const attachCustomerServiceRealtimeRecoveryListeners = () => {
    if (!import.meta.client || browserRecoveryListenersAttached) return
    window.addEventListener('online', recoverCustomerServiceRealtime)
    document.addEventListener('visibilitychange', handleCustomerServiceRealtimeVisibilityChange)
    browserRecoveryListenersAttached = true
  }

  const removeCustomerServiceRealtimeRecoveryListeners = () => {
    if (!import.meta.client || !browserRecoveryListenersAttached) return
    window.removeEventListener('online', recoverCustomerServiceRealtime)
    document.removeEventListener('visibilitychange', handleCustomerServiceRealtimeVisibilityChange)
    browserRecoveryListenersAttached = false
  }

  const scheduleCustomerServiceRealtimeReconnect = (expectedConversationId: string) => {
    if (!import.meta.client || !realtimeReconnectEnabled || realtimeReconnectTimer) return
    if (navigator.onLine === false) return

    const exponentialDelay = Math.min(30_000, 1_000 * (2 ** realtimeReconnectAttempt))
    const jitteredDelay = Math.round(exponentialDelay * (0.8 + Math.random() * 0.4))
    realtimeReconnectAttempt = Math.min(realtimeReconnectAttempt + 1, 5)
    realtimeReconnectTimer = window.setTimeout(() => {
      realtimeReconnectTimer = null
      if (realtimeReconnectEnabled && conversationId.value === expectedConversationId) {
        connectCustomerServiceRealtime()
      }
    }, jitteredDelay)
  }

  const closeCustomerServiceRealtime = () => {
    realtimeReconnectEnabled = false
    closeCustomerServiceRealtimeSocket()
    realtimeConversationId.value = ''

    clearRealtimeReconnectTimer()
    if (import.meta.client && realtimeSyncTimer) {
      window.clearTimeout(realtimeSyncTimer)
      realtimeSyncTimer = null
    }
    clearAgentTyping()
    removeCustomerServiceRealtimeRecoveryListeners()
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

  const handleCustomerServiceRealtimeEvent = (rawFrame: unknown) => {
    try {
      if (typeof rawFrame !== 'string') return
      const frame = JSON.parse(rawFrame || '{}')
      if (typeof frame?.cursor === 'string' && frame.cursor) {
        realtimeLastEventId = frame.cursor
      }
      if (frame?.type !== 'event' || !frame.event || typeof frame.event !== 'object') return

      const payload = frame.event
      if (payload?.conversation_id && payload.conversation_id !== conversationId.value) return
      if (!rememberRealtimeEvent(payload?.event_id)) return
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

      if (realtimeConversationId.value !== currentConversationId || !realtimeSocket.value) {
        connectCustomerServiceRealtime()
        return
      }

      if (realtimeSocket.value.readyState !== WebSocket.OPEN) return
      realtimeSocket.value.send(JSON.stringify({ type: 'typing', is_typing: isTyping }))
    } catch (error) {
      // Typing is transient UI state. Message send/read must keep working even
      // when the WebSocket is reconnecting.
      console.warn('Failed to send customer-service WebSocket typing signal:', error)
    }
  }

  const connectCustomerServiceRealtime = () => {
    if (!import.meta.client || typeof window === 'undefined' || !('WebSocket' in window)) return
    if (!conversationId.value) return
    const currentConversationId = conversationId.value
    realtimeReconnectEnabled = true
    attachCustomerServiceRealtimeRecoveryListeners()
    if (realtimeCursorConversationId.value !== conversationId.value) {
      realtimeCursorConversationId.value = conversationId.value
      realtimeLastEventId = ''
      seenRealtimeEventIds.clear()
    }

    const existingSocket = realtimeSocket.value
    if (existingSocket && realtimeConversationId.value === currentConversationId && (
      existingSocket.readyState === WebSocket.OPEN || existingSocket.readyState === WebSocket.CONNECTING
    )) {
      return
    }

    clearRealtimeReconnectTimer()
    closeCustomerServiceRealtimeSocket()
    realtimeConversationId.value = currentConversationId

    const socket = new WebSocket(customerServiceWebSocketURL(currentConversationId, realtimeLastEventId))
    realtimeSocket.value = socket

    socket.onopen = () => {
      if (realtimeSocket.value !== socket || realtimeConversationId.value !== currentConversationId) return
      realtimeReconnectAttempt = 0
      void loadMessagesFromAPI()
    }

    socket.onmessage = (event: MessageEvent) => {
      if (realtimeSocket.value !== socket) return
      handleCustomerServiceRealtimeEvent(event.data)
    }

    socket.onclose = () => {
      if (realtimeSocket.value !== socket) return
      realtimeSocket.value = null
      if (!realtimeReconnectEnabled || conversationId.value !== currentConversationId) return
      scheduleCustomerServiceRealtimeReconnect(currentConversationId)
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
