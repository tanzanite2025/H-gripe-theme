const UNASSIGNED_GROUP_LABEL = '未配置分组'
const UNSET_CONTACT_LABEL = ''
const DEFAULT_INITIALS = 'CS'

export type ChatAgentPresentation = {
  groupLabel: string
  contactLabel: string
  initials: string
}

export type ChatAgentPresentationItem<T = any> = {
  agent: T
  presentation: ChatAgentPresentation
}

export const getChatAgentGroupNames = (agent: any) => {
  const groups = Array.isArray(agent?.groups) ? agent.groups : []
  return groups
    .map((group: any) => String(group?.name || '').trim())
    .filter(Boolean)
}

export const formatChatAgentGroupLabel = (agent: any, fallback = UNASSIGNED_GROUP_LABEL) => {
  const names = getChatAgentGroupNames(agent)
  if (names.length) {
    return names.slice(0, 3).join(' · ')
  }
  return fallback
}

export const formatChatAgentContactLabel = (agent: any, fallback = UNSET_CONTACT_LABEL) => {
  const email = String(agent?.email || '').trim()
  const whatsapp = String(agent?.whatsapp || '').trim()
  if (email && whatsapp) return `${email} / ${whatsapp}`
  return email || whatsapp || fallback
}

export const formatChatAgentInitials = (agent: any, fallback = DEFAULT_INITIALS) => {
  const name = String(agent?.name || agent?.display_name || agent?.email || '').trim()
  if (!name) return fallback

  const words = name.split(/[\s._-]+/).filter(Boolean)
  return words.slice(0, 2).map((word) => word.charAt(0).toUpperCase()).join('') || fallback
}

export const buildChatAgentPresentation = (agent: any, options?: {
  groupFallback?: string
  contactFallback?: string
  initialsFallback?: string
}): ChatAgentPresentation => {
  return {
    groupLabel: formatChatAgentGroupLabel(agent, options?.groupFallback),
    contactLabel: formatChatAgentContactLabel(agent, options?.contactFallback),
    initials: formatChatAgentInitials(agent, options?.initialsFallback),
  }
}

export const buildChatAgentPresentationList = <T = any>(
  agents: T[] | null | undefined,
  options?: {
    groupFallback?: string
    contactFallback?: string
    initialsFallback?: string
  },
): ChatAgentPresentationItem<T>[] => {
  if (!Array.isArray(agents)) return []

  return agents.map((agent) => ({
    agent,
    presentation: buildChatAgentPresentation(agent, options),
  }))
}
