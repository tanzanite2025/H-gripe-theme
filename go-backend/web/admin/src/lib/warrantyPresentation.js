export const REGISTRATION_STATUS_OPTIONS = [
  { value: 'active', label: '有效' },
  { value: 'expired', label: '已过期' },
  { value: 'claimed', label: '已申请' },
  { value: 'cancelled', label: '已取消' }
]

export const CLAIM_STATUS_OPTIONS = [
  { value: 'submitted', label: '已提交' },
  { value: 'reviewing', label: '审核中' },
  { value: 'approved', label: '已批准' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'completed', label: '已完成' }
]

export const ISSUE_TYPE_LABELS = {
  warranty: '保修问题',
  defect: '质量缺陷',
  damage: '损坏',
  malfunction: '功能异常'
}

export const WARRANTY_SERVICE_TYPE_OPTIONS = [
  { value: 'inspection', label: '检测' },
  { value: 'repair', label: '维修' },
  { value: 'replacement', label: '更换' },
  { value: 'refund', label: '退款' },
  { value: 'shipping', label: '物流处理' }
]

export const WARRANTY_SERVICE_STATUS_OPTIONS = [
  { value: 'open', label: '待处理' },
  { value: 'processing', label: '处理中' },
  { value: 'resolved', label: '已解决' },
  { value: 'closed', label: '已关闭' }
]

export const registrationStatusLabel = (status) =>
  REGISTRATION_STATUS_OPTIONS.find((item) => item.value === status)?.label || status || '-'

export const registrationStatusTone = (status) => {
  const tones = {
    active: 'green',
    expired: 'amber',
    claimed: 'blue',
    cancelled: 'gray'
  }
  return tones[status] || 'gray'
}

export const claimStatusLabel = (status) =>
  CLAIM_STATUS_OPTIONS.find((item) => item.value === status)?.label || status || '-'

export const claimStatusTone = (status) => {
  const tones = {
    submitted: 'amber',
    reviewing: 'blue',
    approved: 'green',
    rejected: 'coral',
    completed: 'gray'
  }
  return tones[status] || 'gray'
}

export const issueTypeLabel = (issueType) => ISSUE_TYPE_LABELS[issueType] || issueType || '-'

export const serviceTypeLabel = (serviceType) =>
  WARRANTY_SERVICE_TYPE_OPTIONS.find((item) => item.value === serviceType)?.label || serviceType || '-'

export const serviceStatusLabel = (status) =>
  WARRANTY_SERVICE_STATUS_OPTIONS.find((item) => item.value === status)?.label || status || '-'

export const serviceStatusTone = (status) => {
  const tones = {
    open: 'amber',
    processing: 'blue',
    resolved: 'green',
    closed: 'gray'
  }
  return tones[status] || 'gray'
}

export const productName = (product) => product?.name || product?.sku || '未关联商品'

export const registrationProductName = (claim) => {
  if (claim?.registration?.product) return productName(claim.registration.product)
  if (claim?.registration_id) return `Registration #${claim.registration_id}`
  return '未绑定注册记录'
}

export const userName = (user) => {
  if (!user) return '未关联用户'
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim()
  return fullName || user.username || user.email || `User #${user.id}`
}

export const claimImages = (claim) => {
  if (!claim?.images) return []
  try {
    const parsed = JSON.parse(claim.images)
    return Array.isArray(parsed) ? parsed.filter(Boolean) : []
  } catch {
    return []
  }
}

export const orderItemLabel = (item) => {
  if (!item) return '未绑定订单行'
  const variant = item.variant_id ? ` / variant ${item.variant_id}` : ''
  const sku = item.sku ? ` / ${item.sku}` : ''
  return `#${item.id} ${item.product_name || `Product ${item.product_id}`}${sku}${variant} × ${item.quantity || 1}`
}

export const formatMoney = (amount, currency = '') => {
  const value = Number(amount || 0)
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  return `${normalizedCurrency || '币种缺失'} ${value.toFixed(2)}`
}

export const formatDate = (value) => value ? new Date(value).toLocaleDateString('zh-CN') : '-'

export const formatDateTime = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'
