export const customerSummary = (conversation: any) => conversation?.customer_summary || {}

export const conversationIsMember = (conversation: any) => {
  const summary = customerSummary(conversation)
  return ['member', 'account', 'user'].includes(String(summary.identity || summary.type || '').toLowerCase()) || conversation?.visitor_anonymous === false
}

export const conversationDisplayName = (conversation: any) => {
  const summaryName = String(customerSummary(conversation).display_name || '').trim()
  if (summaryName) return summaryName
  return conversation?.customer_name || (conversationIsMember(conversation) ? '会员客户' : '匿名客户')
}

export const customerIdentityLabel = (conversation: any) => {
  const label = String(customerSummary(conversation).identity_label || '').trim()
  if (label) return label
  return conversationIsMember(conversation) ? '会员' : '游客'
}

export const customerRegionLabel = (conversation: any) => {
  const label = String(customerSummary(conversation).region_label || '').trim()
  return label || '未知区域'
}

export const memberTier = (conversation: any) => customerSummary(conversation).member_tier || null
export const memberTierIcon = (conversation: any) => String(memberTier(conversation)?.icon || '').trim()
export const memberTierName = (conversation: any) => String(memberTier(conversation)?.name || '').trim() || '会员等级'

export const tierStyle = (tier: any) => {
  const color = String(tier?.color || '').trim()
  if (!color) return {}
  return {
    color,
    borderColor: `${color}33`,
    backgroundColor: `${color}14`,
  }
}

export const memberTierStyle = (conversation: any) => tierStyle(memberTier(conversation))

export const identityPillClass = (conversation: any) => [
  'inline-flex h-5 items-center gap-1 rounded-full border px-2',
  conversationIsMember(conversation)
    ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700'
    : 'border-amber-500/20 bg-amber-500/10 text-amber-700',
]

export const initials = (name: any) => String(name || '?').trim().slice(0, 2).toUpperCase()
export const conversationInitials = (conversation: any) => initials(conversationDisplayName(conversation))

export const formatDate = (dateString: any) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const formatShortDate = (dateString: any) => dateString ? new Date(dateString).toLocaleDateString('zh-CN') : '-'
export const formatMoney = (value: any) => `$${Number(value || 0).toFixed(2)}`

export const messageMetadata = (message: any) => {
  if (!message?.metadata) return {}
  if (typeof message.metadata === 'string') {
    try {
      return JSON.parse(message.metadata)
    } catch {
      return {}
    }
  }
  return message.metadata
}

export const isConfigConfirmMessage = (message: any) => message?.message_type === 'config_confirm'
export const isOrderMessage = (message: any) => message?.message_type === 'order'

export const configProduct = (message: any) => {
  const metadata = messageMetadata(message)
  return metadata?.product || {}
}

export const configSelection = (message: any) => {
  const metadata = messageMetadata(message)
  return metadata?.selections || {}
}

export const configOptionRows = (message: any) => {
  const options = configSelection(message)?.options
  return Array.isArray(options) ? options : []
}

export const orderPayload = (message: any) => messageMetadata(message) || {}

export const orderItems = (message: any) => {
  const items = orderPayload(message)?.items
  return Array.isArray(items) ? items : []
}

export const formatOrderTotal = (order: any) => {
  const total = Number(order?.total || 0)
  const currency = String(order?.currency || '').trim().toUpperCase()
  if (!Number.isFinite(total) || total <= 0) return currency
  return `${currency || '币种缺失'} ${total.toFixed(2)}`
}

export const statusDisplayValue = (status: any) => {
  if (['resolved', 'closed'].includes(status)) return 'closed'
  if (['in_progress', 'active'].includes(status)) return 'active'
  if (['open', 'pending'].includes(status)) return 'pending'
  return status || 'pending'
}

export const statusLabel = (status: any) => ({
  active: '进行中',
  pending: '待处理',
  closed: '已关闭',
  open: '待处理',
  in_progress: '处理中',
  resolved: '已解决',
} as Record<string, string>)[status] || status || '-'

export const statusTone = (status: any) => ({
  active: 'blue',
  pending: 'amber',
  closed: 'gray',
  open: 'amber',
  in_progress: 'blue',
  resolved: 'green',
} as Record<string, string>)[status] || 'gray'

export const signalTone = (status: any) => {
  if (['captured', 'verified', 'bound', 'created', 'linked'].includes(status)) return 'green'
  if (['missing', 'not_captured', 'not_linked', 'not_created', 'missing_user'].includes(status)) return 'amber'
  return 'gray'
}

export const assigneeName = (assignedTo: any, agents: any[] = []) => {
  if (!assignedTo) return '未分配'
  const agent = agents.find((item) => Number(item.user_id || item.id) === Number(assignedTo))
  return agent?.name || agent?.email || `用户 ${assignedTo}`
}
