import type { AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'

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

export const isValidCustomerTimezone = (timezone: any) => {
  const normalized = String(timezone || '').trim()
  if (!normalized) return false
  try {
    new Intl.DateTimeFormat('en-US', { timeZone: normalized }).format()
    return true
  } catch {
    return false
  }
}

const customerTimezoneHour = (dateValue: any, timezone: any) => {
  const normalized = String(timezone || '').trim()
  const date = dateValue instanceof Date ? dateValue : new Date(dateValue)
  if (!isValidCustomerTimezone(normalized) || Number.isNaN(date.getTime())) return null

  const hourPart = new Intl.DateTimeFormat('en-US', {
    timeZone: normalized,
    hour: '2-digit',
    hour12: false,
  }).formatToParts(date).find((part) => part.type === 'hour')
  const hour = Number(hourPart?.value)
  return Number.isFinite(hour) ? hour % 24 : null
}

export const formatCustomerLocalTime = (dateValue: any, timezone: any) => {
  const normalized = String(timezone || '').trim()
  const date = dateValue instanceof Date ? dateValue : new Date(dateValue)
  if (!isValidCustomerTimezone(normalized) || Number.isNaN(date.getTime())) return '未采集时区'

  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: normalized,
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date)
}

export const formatCustomerLocalDate = (dateValue: any, timezone: any) => {
  const normalized = String(timezone || '').trim()
  const date = dateValue instanceof Date ? dateValue : new Date(dateValue)
  if (!isValidCustomerTimezone(normalized) || Number.isNaN(date.getTime())) return ''

  return new Intl.DateTimeFormat('zh-CN', {
    timeZone: normalized,
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  }).format(date)
}

export const customerLocalTimePhase = (dateValue: any, timezone: any) => {
  const hour = customerTimezoneHour(dateValue, timezone)
  if (hour === null) return '未采集'
  if (hour < 6) return '深夜'
  if (hour < 9) return '早晨'
  if (hour < 12) return '上午'
  if (hour < 14) return '午间'
  if (hour < 18) return '当地工作时间'
  if (hour < 22) return '晚上'
  return '深夜'
}

export const customerLocalTimeHint = (dateValue: any, timezone: any) => {
  const hour = customerTimezoneHour(dateValue, timezone)
  if (hour === null) return '暂未采集客户时区，先按普通客服语气沟通'
  if (hour < 6 || hour >= 22) return '可能不方便回复，建议保持简短'
  if (hour < 9) return '当地刚开始一天，适合先礼貌问候'
  if (hour < 18) return '当地处于工作时段，可直接说明事项'
  return '当地已进入晚上，建议保持简洁'
}

export const customerTimezoneSourceLabel = (source: any) => ({
  visitor_profile: '来自访客档案',
  account: '来自账号档案',
  request: '来自本次请求',
}[String(source || '').trim()] || (String(source || '').trim() || '未采集'))

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
export const isProductMessage = (message: any) => message?.message_type === 'product'
export const isVideoMessage = (message: any) => message?.message_type === 'video'

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

export const productPayload = (message: any) => messageMetadata(message) || {}

export const videoPayload = (message: any) => messageMetadata(message) || {}

export const orderItems = (message: any) => {
  const items = orderPayload(message)?.items
  return Array.isArray(items) ? items : []
}

export const formatOrderTotal = (order: any) => {
  const total = Number(order?.total || 0)
  const currency = String(order?.currency || '').trim().toUpperCase()
  if (!Number.isFinite(total) || total <= 0) return currency || '-'
  return currency ? `${currency} ${total.toFixed(2)}` : total.toFixed(2)
}

export const formatProductPrice = (product: any) => {
  const priceText = String(product?.price || '').trim()
  const currency = String(product?.currency || '').trim().toUpperCase()
  if (priceText) {
    if (currency && !priceText.toUpperCase().startsWith(currency)) {
      return `${currency} ${priceText}`
    }
    return priceText
  }

  const priceValue = Number(product?.price_value ?? product?.priceValue ?? 0)
  if (!Number.isFinite(priceValue) || priceValue <= 0) return ''

  return currency ? `${currency} ${priceValue.toFixed(2)}` : priceValue.toFixed(2)
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

export const statusTone = (status: any): AdminStatusTone => ({
  active: 'blue',
  pending: 'amber',
  closed: 'gray',
  open: 'amber',
  in_progress: 'blue',
  resolved: 'green',
} as Record<string, AdminStatusTone>)[status] || 'gray'

export const signalTone = (status: any): AdminStatusTone => {
  if (['captured', 'verified', 'bound', 'created', 'linked'].includes(status)) return 'green'
  if (['missing', 'not_captured', 'not_linked', 'not_created', 'missing_user'].includes(status)) return 'amber'
  return 'gray'
}

export const agentDisplayName = (agent: any, fallbackPrefix = '客服') => {
  const name = String(agent?.name || '').trim()
  if (name) return name

  const email = String(agent?.email || '').trim()
  if (email) return email

  const agentID = agent?.user_id || agent?.id
  if (agentID !== undefined && agentID !== null && String(agentID).trim()) {
    return `${fallbackPrefix} #${agentID}`
  }

  return fallbackPrefix
}

export const assigneeName = (assignedTo: any, agents: any[] = [], currentUser: any = null) => {
  const assignedToID = Number(assignedTo)
  if (!Number.isFinite(assignedToID) || assignedToID <= 0) return '未分配'

  const agent = agents.find((item) => Number(item.user_id || item.id) === assignedToID)
  if (agent) return agentDisplayName(agent)

  if (currentUser && Number(currentUser.id) === assignedToID) {
    const currentUserName = [
      currentUser.first_name,
      currentUser.last_name,
    ].filter(Boolean).map((item) => String(item).trim()).filter(Boolean).join(' ')

    const fallbackName = String(currentUser.display_name || currentUser.name || '').trim()
    return currentUserName || fallbackName || '当前客服'
  }

  return '未知客服'
}
