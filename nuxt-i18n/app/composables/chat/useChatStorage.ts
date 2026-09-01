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

export const getChatStorageKey = (conversationId: string, agentId: number | string | undefined) => {
  return `tz_chat_${conversationId || 'pending'}_agent_${agentId || 'default'}`
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

export const hasLocalChatHistory = () => {
  if (typeof window === 'undefined') return false

  const chatKeys = Object.keys(localStorage).filter(key => key.startsWith('tz_chat_'))
  for (const key of chatKeys) {
    const stored = localStorage.getItem(key)
    if (!stored) continue

    const parsed = JSON.parse(stored)
    if (Array.isArray(parsed?.messages) && parsed.messages.length > 0) {
      return true
    }
  }

  return false
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

export const getLastSelectedAgentId = () => {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(LAST_AGENT_STORAGE_KEY)
}

export const saveLastSelectedAgentId = (agentId: number | string) => {
  if (typeof window === 'undefined') return
  localStorage.setItem(LAST_AGENT_STORAGE_KEY, String(agentId))
}
