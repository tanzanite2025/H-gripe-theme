export type ChatTab = 'chat' | 'share' | 'orders' | 'faq' | 'warranty' | 'member' | 'test' | 'tire'

export interface ChatRoomState {
  messages: any[]
  activeTab: ChatTab
  newMessage: string
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

export const createEmptyChatRoom = (): ChatRoomState => ({
  messages: [],
  activeTab: 'chat',
  newMessage: '',
  searchQuery: '',
  searchResults: [],
  ordersList: [],
  isLoadingOrders: false,
  isSearching: false
})

export const getChatStorageKey = (conversationId: string, agentId: number | string | undefined) => {
  return `tz_chat_${conversationId || 'pending'}_agent_${agentId || 'default'}`
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
  const messages = storedMessages.filter((message: any) => {
    const messageTime = new Date(message.created_at).getTime()
    return (now - messageTime) < expiryTime
  })

  return {
    room: {
      messages,
      activeTab: data.activeTab || 'chat',
      newMessage: data.newMessage || '',
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

  localStorage.setItem(storageKey, JSON.stringify({
    messages: room.messages,
    activeTab: room.activeTab,
    newMessage: room.newMessage,
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
