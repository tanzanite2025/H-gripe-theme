export interface ChatEmailSettings {
  preSalesEmail: string
  afterSalesEmail: string
}

export interface ChatAgentCacheData {
  agents: any[]
  groups: any[]
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
  version?: number
}

const AGENTS_CACHE_KEY = 'public_chat_agents_cache'
const AGENTS_CACHE_VERSION = 3
const AGENTS_CACHE_TTL_MS = 30 * 60 * 1000
const VALID_ONLINE_STATUSES = new Set(['online', 'busy', 'away', 'offline'])

export const normalizeChatAgentOnlineStatus = (agent: any) => {
  const value = String(agent?.online_status ?? agent?.status ?? '').trim().toLowerCase()
  return VALID_ONLINE_STATUSES.has(value) ? value : 'offline'
}

export const normalizeChatAgent = (agent: any) => ({
  ...agent,
  avatar: String(agent?.avatar || '').trim(),
  email: String(agent?.email || '').trim(),
  whatsapp: String(agent?.whatsapp || '').trim(),
  groups: Array.isArray(agent?.groups) ? agent.groups : [],
  primary_group: agent?.primary_group || (Array.isArray(agent?.groups) ? agent.groups[0] : null),
  online_status: normalizeChatAgentOnlineStatus(agent)
})

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
  if (parsed.version !== AGENTS_CACHE_VERSION) return null
  if (Date.now() - parsed.timestamp >= AGENTS_CACHE_TTL_MS) return null

  return {
    agents: filterAgentsForUser(parsed.data.agents || [], currentUserId).map(normalizeChatAgent),
    groups: Array.isArray(parsed.data.groups) ? parsed.data.groups : [],
    emailSettings: parsed.data.emailSettings
  }
}

export const saveChatAgentsCache = (data: ChatAgentCacheData) => {
  if (typeof window === 'undefined' || data.agents.length === 0) return

  localStorage.setItem(AGENTS_CACHE_KEY, JSON.stringify({
    data: {
      ...data,
      agents: data.agents.map(normalizeChatAgent)
    },
    timestamp: Date.now(),
    version: AGENTS_CACHE_VERSION
  }))
}

export const getDevFallbackAgentDirectory = (): ChatAgentCacheData => ({
  agents: [
    { id: 'CS001', name: 'Sales', email: 'sales@tanzanite.site', avatar: '', whatsapp: '+8613800138001', user_id: null, online_status: 'offline', groups: [{ id: 1, code: 'sales', name: 'Sales' }] },
    { id: 'CS002', name: 'Tech Support', email: 'tech@tanzanite.site', avatar: '', whatsapp: '+8613800138002', user_id: null, online_status: 'offline', groups: [{ id: 2, code: 'technical_support', name: 'Technical Support' }] },
    { id: 'CS003', name: 'After Sales', email: 'support@tanzanite.site', avatar: '', whatsapp: '+8613800138003', user_id: null, online_status: 'offline', groups: [{ id: 3, code: 'after_sales', name: 'After Sales' }] },
  ],
  groups: [
    { id: 1, code: 'sales', name: 'Sales', status: 'active', sort_order: 10 },
    { id: 2, code: 'technical_support', name: 'Technical Support', status: 'active', sort_order: 20 },
    { id: 3, code: 'after_sales', name: 'After Sales', status: 'active', sort_order: 30 }
  ],
  emailSettings: {
    preSalesEmail: 'sales@tanzanite.site',
    afterSalesEmail: 'support@tanzanite.site'
  }
})

export const fetchChatAgentDirectory = async (apiBase: string): Promise<ChatAgentCacheData> => {
  const response = await $fetch<any>(`${apiBase}/customer-service/agents`)
  return {
    agents: response?.success && Array.isArray(response.data) ? response.data.map(normalizeChatAgent) : [],
    groups: response?.success && Array.isArray(response.groups) ? response.groups : [],
    emailSettings: response?.emailSettings || null
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
    groups: [],
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
    agents: filterAgentsForUser(directory.agents, currentUserId).map(normalizeChatAgent),
    groups: directory.groups,
    emailSettings: directory.emailSettings
  }
}
