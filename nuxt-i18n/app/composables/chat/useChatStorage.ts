export type ChatTab = 'chat' | 'orders' | 'faq' | 'warranty' | 'member' | 'tire' | 'calculator'

export interface ChatRoomState {
  messages: any[]
  activeTab: ChatTab
  newMessage: string
  pendingProductReference: any | null
  searchQuery: string
  searchResults: any[]
  ordersList: any[]
  isLoadingOrders: boolean
  isSearching: boolean
}

interface StoredChatRoomState extends Partial<ChatRoomState> {
  lastUpdated?: string
}

export const CHAT_STORAGE_EXPIRY_DAYS = 5
export const LAST_AGENT_STORAGE_KEY = 'tz_last_selected_agent'
export const CHAT_SENDING_MESSAGE_STALE_MS = 2 * 60 * 1000
export const CHAT_DRAFT_STORAGE_VERSION = 1

const ANONYMOUS_CHAT_OWNER_ID_STORAGE_KEY = 'tz_chat_anonymous_owner_id_v1'
const CHAT_SESSION_STORAGE_VERSION = 1

const chatTabs: readonly ChatTab[] = ['chat', 'orders', 'faq', 'warranty', 'member', 'tire', 'calculator']

const normalizeChatTab = (value: unknown): ChatTab => {
  if (value === 'share') return 'chat'
  return chatTabs.includes(value as ChatTab) ? value as ChatTab : 'chat'
}

export const createEmptyChatRoom = (): ChatRoomState => ({
  messages: [],
  activeTab: 'chat',
  newMessage: '',
  pendingProductReference: null,
  searchQuery: '',
  searchResults: [],
  ordersList: [],
  isLoadingOrders: false,
  isSearching: false
})

export const getChatStorageKey = (
  conversationId: string,
  agentId: number | string | undefined,
  ownerStorageScope = '',
) => {
  const ownerPrefix = ownerStorageScope ? `owner_${encodeURIComponent(ownerStorageScope)}_` : ''
  const normalizedConversationId = encodeURIComponent(conversationId || 'pending')
  const normalizedAgentId = encodeURIComponent(String(agentId || 'default'))
  return `tz_chat_${ownerPrefix}${normalizedConversationId}_agent_${normalizedAgentId}`
}

export const getChatOwnerStorageScope = (userId: unknown, anonymousOwnerId: string) => {
  const normalizedUserId = String(userId ?? '').trim()
  if (normalizedUserId) return `member:${normalizedUserId}`
  return anonymousOwnerId ? `guest:${anonymousOwnerId}` : 'guest:unavailable'
}

export const getOrCreateAnonymousChatOwnerId = () => {
  if (typeof window === 'undefined') return ''

  try {
    const storedOwnerId = window.localStorage.getItem(ANONYMOUS_CHAT_OWNER_ID_STORAGE_KEY)
    if (storedOwnerId) return storedOwnerId

    const randomId = globalThis.crypto?.randomUUID?.()
      || `${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 14)}`
    const anonymousOwnerId = `visitor_${randomId}`
    window.localStorage.setItem(ANONYMOUS_CHAT_OWNER_ID_STORAGE_KEY, anonymousOwnerId)
    return anonymousOwnerId
  } catch {
    return ''
  }
}

export const getChatSessionStorageKey = (ownerStorageScope: string) => {
  return `tz_chat_session_v1_${encodeURIComponent(ownerStorageScope)}`
}

export const loadChatConversationIdFromStorage = (ownerStorageScope: string) => {
  if (typeof window === 'undefined') return ''

  try {
    const stored = window.localStorage.getItem(getChatSessionStorageKey(ownerStorageScope))
    if (!stored) return ''

    const record = JSON.parse(stored)
    if (record?.version !== CHAT_SESSION_STORAGE_VERSION) return ''
    return typeof record.conversationId === 'string' ? record.conversationId.trim() : ''
  } catch {
    return ''
  }
}

export const saveChatConversationIdToStorage = (ownerStorageScope: string, conversationId: string) => {
  if (typeof window === 'undefined') return
  const normalizedConversationId = String(conversationId || '').trim()
  if (!normalizedConversationId) return

  window.localStorage.setItem(getChatSessionStorageKey(ownerStorageScope), JSON.stringify({
    version: CHAT_SESSION_STORAGE_VERSION,
    conversationId: normalizedConversationId,
    updatedAt: new Date().toISOString(),
  }))
}

export const clearStoredChatConversationId = (ownerStorageScope: string, expectedConversationId = '') => {
  if (typeof window === 'undefined') return
  const storageKey = getChatSessionStorageKey(ownerStorageScope)
  if (expectedConversationId) {
    const storedConversationId = loadChatConversationIdFromStorage(ownerStorageScope)
    if (storedConversationId && storedConversationId !== expectedConversationId) return
  }
  window.localStorage.removeItem(storageKey)
}

export const getChatDraftStorageKey = (
  conversationId: string,
  agentId: number | string | undefined,
  ownerStorageScope = '',
) => {
  const ownerPrefix = ownerStorageScope ? `owner_${encodeURIComponent(ownerStorageScope)}_` : ''
  const normalizedConversationId = encodeURIComponent(conversationId || 'pending')
  const normalizedAgentId = encodeURIComponent(String(agentId || 'default'))
  return `tz_chat_draft_v1_${ownerPrefix}${normalizedConversationId}_agent_${normalizedAgentId}`
}

export const loadChatDraftFromStorage = (storageKey: string, expiryDays = CHAT_STORAGE_EXPIRY_DAYS, now = Date.now()) => {
  if (typeof window === 'undefined') return null

  try {
    const stored = window.localStorage.getItem(storageKey)
    if (!stored) return null

    const record = JSON.parse(stored)
    if (record?.version !== CHAT_DRAFT_STORAGE_VERSION || typeof record.draft !== 'string') return null

    const updatedAt = Date.parse(record.updatedAt || '')
    const expiryTime = expiryDays * 24 * 60 * 60 * 1000
    if (!Number.isFinite(updatedAt) || now - updatedAt >= expiryTime) {
      window.localStorage.removeItem(storageKey)
      return null
    }

    return record.draft as string
  } catch {
    return null
  }
}

export const restoreChatRoomDraftFromStoredSources = ({
  currentDraft,
  legacyRoomDraft,
  versionedDraft,
}: {
  currentDraft: string
  legacyRoomDraft: string
  versionedDraft: string | null
}) => {
  if (versionedDraft !== null) return versionedDraft
  return currentDraft || legacyRoomDraft || ''
}

export const saveChatDraftToStorage = (storageKey: string, draft: string) => {
  if (typeof window === 'undefined') return
  if (!draft) {
    window.localStorage.removeItem(storageKey)
    return
  }

  window.localStorage.setItem(storageKey, JSON.stringify({
    version: CHAT_DRAFT_STORAGE_VERSION,
    draft,
    updatedAt: new Date().toISOString(),
  }))
}

const messageCreatedAtMs = (message: any) => {
  const value = new Date(message?.created_at || 0).getTime()
  return Number.isFinite(value) ? value : 0
}

export const isStaleSendingChatMessage = (message: any, now = Date.now()) => {
  return message?.sync_state === 'sending' && (now - messageCreatedAtMs(message)) > CHAT_SENDING_MESSAGE_STALE_MS
}

export const normalizeStoredChatMessage = (message: any, now = Date.now()) => {
  if (!message || typeof message !== 'object') return null

  const normalized = { ...message }
  const id = String(normalized.id || '')
  if (!normalized.sync_state && id.startsWith('local-')) {
    normalized.sync_state = 'failed'
  }
  if (isStaleSendingChatMessage(normalized, now)) {
    normalized.sync_state = 'failed'
    normalized.sync_error = normalized.sync_error || 'Message was not confirmed by the server.'
  }
  return normalized
}

export const hasLocalChatHistory = (ownerStorageScope: string) => {
  if (typeof window === 'undefined') return false
  if (!ownerStorageScope) return false

  const ownerPrefix = `tz_chat_owner_${encodeURIComponent(ownerStorageScope)}_`
  const storage = window.localStorage
  const chatKeys = Array.from({ length: storage.length }, (_, index) => storage.key(index))
    .filter((key): key is string => Boolean(key?.startsWith(ownerPrefix)))
  for (const key of chatKeys) {
    const stored = storage.getItem(key)
    if (!stored) continue

    const parsed = JSON.parse(stored)
    if (Array.isArray(parsed?.messages) && parsed.messages.length > 0) {
      return true
    }
  }

  return false
}

export const countUnreadCustomerServiceMessagesForOwner = (ownerStorageScope: string) => {
  if (typeof window === 'undefined' || !ownerStorageScope) return 0

  const ownerPrefix = `tz_chat_owner_${encodeURIComponent(ownerStorageScope)}_`
  const storage = window.localStorage
  let unreadCount = 0

  for (let index = 0; index < storage.length; index += 1) {
    const key = storage.key(index)
    if (!key?.startsWith(ownerPrefix)) continue

    try {
      const stored = storage.getItem(key)
      if (!stored) continue

      const parsed = JSON.parse(stored)
      if (!Array.isArray(parsed?.messages)) continue
      unreadCount += parsed.messages.filter((message: any) => message?.is_agent && !message?.is_read).length
    } catch {
      continue
    }
  }

  return unreadCount
}

export const loadChatRoomFromStorage = (storageKey: string, expiryDays = CHAT_STORAGE_EXPIRY_DAYS) => {
  if (typeof window === 'undefined') return null

  const stored = localStorage.getItem(storageKey)
  if (!stored) return null

  const data = JSON.parse(stored) as StoredChatRoomState
  const storedMessages = Array.isArray(data.messages) ? data.messages : []
  const expiryTime = expiryDays * 24 * 60 * 60 * 1000
  const now = Date.now()
  const messages = storedMessages
    .map((message: any) => normalizeStoredChatMessage(message, now))
    .filter((message: any) => {
      const messageTime = messageCreatedAtMs(message)
      return messageTime > 0 && (now - messageTime) < expiryTime
    })

  return {
    room: {
      messages,
      activeTab: normalizeChatTab(data.activeTab),
      newMessage: data.newMessage || '',
      pendingProductReference: data.pendingProductReference || null,
      searchQuery: data.searchQuery || '',
      searchResults: Array.isArray(data.searchResults) ? data.searchResults : [],
      ordersList: Array.isArray(data.ordersList) ? data.ordersList : [],
      isSearching: !!data.isSearching,
      isLoadingOrders: !!data.isLoadingOrders
    },
    hasExpiredMessages: messages.length !== storedMessages.length
  }
}

export const saveChatRoomToStorage = (storageKey: string, room: ChatRoomState) => {
  if (typeof window === 'undefined') return
  const now = Date.now()

  localStorage.setItem(storageKey, JSON.stringify({
    messages: room.messages
      .map((message: any) => normalizeStoredChatMessage(message, now))
      .filter(Boolean),
    activeTab: room.activeTab,
    newMessage: room.newMessage,
    pendingProductReference: room.pendingProductReference || null,
    searchQuery: room.searchQuery,
    searchResults: room.searchResults,
    ordersList: room.ordersList,
    isSearching: room.isSearching,
    isLoadingOrders: room.isLoadingOrders,
    lastUpdated: new Date().toISOString()
  }))
}

export const migrateChatRoomStorage = (sourceStorageKey: string, destinationStorageKey: string, expiryDays = CHAT_STORAGE_EXPIRY_DAYS) => {
  if (typeof window === 'undefined' || !sourceStorageKey || sourceStorageKey === destinationStorageKey) return false

  const source = loadChatRoomFromStorage(sourceStorageKey, expiryDays)
  if (!source) return false

  const destination = loadChatRoomFromStorage(destinationStorageKey, expiryDays)
  const mergedMessagesById = new Map<string, any>()
  for (const message of destination?.room.messages || []) {
    mergedMessagesById.set(String(message.id), message)
  }
  for (const message of source.room.messages) {
    const messageId = String(message.id)
    const existingMessage = mergedMessagesById.get(messageId)
    if (!existingMessage || (message.sync_state === 'failed' && existingMessage.sync_state !== 'failed')) {
      mergedMessagesById.set(messageId, message)
    }
  }

  const messages = Array.from(mergedMessagesById.values()).sort((left, right) => {
    return messageCreatedAtMs(left) - messageCreatedAtMs(right)
  })
  const destinationRoom = destination?.room
  const mergedRoom: ChatRoomState = {
    ...source.room,
    ...destinationRoom,
    messages,
    newMessage: destinationRoom?.newMessage || source.room.newMessage,
    pendingProductReference: destinationRoom?.pendingProductReference || source.room.pendingProductReference,
  }

  saveChatRoomToStorage(destinationStorageKey, mergedRoom)
  window.localStorage.removeItem(sourceStorageKey)
  return true
}

export const migrateChatDraftStorage = (sourceStorageKey: string, destinationStorageKey: string) => {
  if (typeof window === 'undefined' || !sourceStorageKey || sourceStorageKey === destinationStorageKey) return false

  const sourceDraft = loadChatDraftFromStorage(sourceStorageKey)
  if (sourceDraft === null) return false

  const destinationDraft = loadChatDraftFromStorage(destinationStorageKey)
  if (destinationDraft === null) {
    saveChatDraftToStorage(destinationStorageKey, sourceDraft)
  }
  window.localStorage.removeItem(sourceStorageKey)
  return true
}

export const getLastSelectedAgentId = () => {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(LAST_AGENT_STORAGE_KEY)
}

export const saveLastSelectedAgentId = (agentId: number | string) => {
  if (typeof window === 'undefined') return
  localStorage.setItem(LAST_AGENT_STORAGE_KEY, String(agentId))
}
