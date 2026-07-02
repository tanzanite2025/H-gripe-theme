export interface ChatEmailSettings {
  preSalesEmail: string
  afterSalesEmail: string
}

export interface ChatAgentCacheData {
  agents: any[]
  emailSettings: ChatEmailSettings | null
}

interface LoadChatAgentDirectoryOptions {
  apiBase: string
  currentUserId?: number | string | null
  allowDevFallback?: boolean
}

interface StoredChatAgentCache {
  data?: ChatAgentCacheData
  timestamp?: number
}

const AGENTS_CACHE_KEY = 'public_chat_agents_cache'
const AGENTS_CACHE_TTL_MS = 30 * 60 * 1000

export const filterAgentsForUser = (agents: any[], currentUserId?: number | string | null) => {
  return agents.filter((agent: any) => {
    return !agent.user_id || String(agent.user_id) !== String(currentUserId)
  })
}

export const loadCachedChatAgents = (currentUserId?: number | string | null) => {
  if (typeof window === 'undefined') return null

  const cached = localStorage.getItem(AGENTS_CACHE_KEY)
  if (!cached) return null

  const parsed = JSON.parse(cached) as StoredChatAgentCache
  if (!parsed?.data || !parsed.timestamp) return null
  if (Date.now() - parsed.timestamp >= AGENTS_CACHE_TTL_MS) return null

  return {
    agents: filterAgentsForUser(parsed.data.agents || [], currentUserId),
    emailSettings: parsed.data.emailSettings
  }
}

export const saveChatAgentsCache = (data: ChatAgentCacheData) => {
  if (typeof window === 'undefined' || data.agents.length === 0) return

  localStorage.setItem(AGENTS_CACHE_KEY, JSON.stringify({
    data,
    timestamp: Date.now()
  }))
}

export const getDevFallbackAgentDirectory = (): ChatAgentCacheData => ({
  agents: [
    { id: 'CS001', name: 'Sales', email: 'sales@tanzanite.site', avatar: '', whatsapp: '+8613800138001', user_id: null },
    { id: 'CS002', name: 'Tech Support', email: 'tech@tanzanite.site', avatar: '', whatsapp: '+8613800138002', user_id: null },
    { id: 'CS003', name: 'After Sales', email: 'support@tanzanite.site', avatar: '', whatsapp: '+8613800138003', user_id: null },
  ],
  emailSettings: {
    preSalesEmail: 'sales@tanzanite.site',
    afterSalesEmail: 'support@tanzanite.site'
  }
})

export const fetchChatAgentDirectory = async (apiBase: string): Promise<ChatAgentCacheData> => {
  const response = await $fetch<any>(`${apiBase}/customer-service/agents`)
  return {
    agents: response?.success && Array.isArray(response.data) ? response.data : [],
    emailSettings: null
  }
}

export const loadChatAgentDirectory = async ({
  apiBase,
  currentUserId,
  allowDevFallback = false
}: LoadChatAgentDirectoryOptions) => {
  try {
    const cachedDirectory = loadCachedChatAgents(currentUserId)
    if (cachedDirectory) return cachedDirectory
  } catch (e) {
    // 缓存解析失败，继续请求
  }

  let directory: ChatAgentCacheData = {
    agents: [],
    emailSettings: null
  }

  try {
    directory = await fetchChatAgentDirectory(apiBase)
  } catch (error) {
    console.warn('Failed to fetch agents from API', error)
  }

  if (directory.agents.length === 0 && allowDevFallback) {
    directory = getDevFallbackAgentDirectory()
  }

  saveChatAgentsCache(directory)

  return {
    agents: filterAgentsForUser(directory.agents, currentUserId),
    emailSettings: directory.emailSettings
  }
}
