type StatusTone = 'green' | 'amber' | 'blue' | 'gray' | 'coral'

interface StatusOption {
  value: string
  label: string
}

interface WarrantyProductLike {
  name?: string | null
  sku?: string | null
}

interface WarrantyClaimLike {
  order_item?: OrderItemLike | null
  images?: string | null
}

interface WarrantyUserLike {
  id?: number | string | null
  first_name?: string | null
  last_name?: string | null
  username?: string | null
  email?: string | null
}

interface OrderItemLike {
  id?: number | string | null
  product_id?: number | string | null
  product_name?: string | null
  sku?: string | null
  variant_id?: number | string | null
  quantity?: number | string | null
}

export const CLAIM_STATUS_OPTIONS: StatusOption[] = [
  { value: 'submitted', label: '已提交' },
  { value: 'reviewing', label: '审核中' },
  { value: 'approved', label: '已批准' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'completed', label: '已完成' }
]

export const ISSUE_TYPE_LABELS: Record<string, string> = {
  warranty: '保修问题',
  defect: '质量缺陷',
  damage: '损坏',
  malfunction: '功能异常'
}

export const WARRANTY_SERVICE_TYPE_OPTIONS: StatusOption[] = [
  { value: 'inspection', label: '检测' },
  { value: 'repair', label: '维修' },
  { value: 'replacement', label: '更换' },
  { value: 'refund', label: '退款' },
  { value: 'shipping', label: '物流处理' }
]

export const WARRANTY_SERVICE_STATUS_OPTIONS: StatusOption[] = [
  { value: 'open', label: '待处理' },
  { value: 'processing', label: '处理中' },
  { value: 'resolved', label: '已解决' },
  { value: 'closed', label: '已关闭' }
]

export const claimStatusLabel = (status?: string | null): string =>
  CLAIM_STATUS_OPTIONS.find((item) => item.value === status)?.label || status || '-'

export const claimStatusTone = (status?: string | null): StatusTone => {
  const tones: Record<string, StatusTone> = {
    submitted: 'amber',
    reviewing: 'blue',
    approved: 'green',
    rejected: 'coral',
    completed: 'gray'
  }
  return tones[status || ''] || 'gray'
}

export const issueTypeLabel = (issueType?: string | null): string => ISSUE_TYPE_LABELS[issueType || ''] || issueType || '-'

export const serviceTypeLabel = (serviceType?: string | null): string =>
  WARRANTY_SERVICE_TYPE_OPTIONS.find((item) => item.value === serviceType)?.label || serviceType || '-'

export const serviceStatusLabel = (status?: string | null): string =>
  WARRANTY_SERVICE_STATUS_OPTIONS.find((item) => item.value === status)?.label || status || '-'

export const serviceStatusTone = (status?: string | null): StatusTone => {
  const tones: Record<string, StatusTone> = {
    open: 'amber',
    processing: 'blue',
    resolved: 'green',
    closed: 'gray'
  }
  return tones[status || ''] || 'gray'
}

export const productName = (product?: WarrantyProductLike | null): string => product?.name || product?.sku || '未关联商品'

export const claimProductName = (claim?: WarrantyClaimLike | null): string =>
  claim?.order_item ? orderItemLabel(claim.order_item) : '未绑定订单行'

export const userName = (user?: WarrantyUserLike | null): string => {
  if (!user) return '未关联用户'
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim()
  return fullName || user.username || user.email || `User #${user.id}`
}

export const claimImages = (claim?: WarrantyClaimLike | null): string[] => {
  if (!claim?.images) return []
  try {
    const parsed = JSON.parse(claim.images)
    return Array.isArray(parsed)
      ? parsed.filter((item): item is string => typeof item === 'string' && item.length > 0)
      : []
  } catch {
    return []
  }
}

export const orderItemLabel = (item?: OrderItemLike | null): string => {
  if (!item) return '未绑定订单行'
  const variant = item.variant_id ? ` / variant ${item.variant_id}` : ''
  const sku = item.sku ? ` / ${item.sku}` : ''
  return `#${item.id} ${item.product_name || `Product ${item.product_id}`}${sku}${variant} × ${item.quantity || 1}`
}

export const formatMoney = (amount?: number | string | null, currency = ''): string => {
  const value = Number(amount || 0)
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  return `${normalizedCurrency || '币种缺失'} ${value.toFixed(2)}`
}

export const formatDate = (value?: string | number | Date | null): string => value ? new Date(value).toLocaleDateString('zh-CN') : '-'

export const formatDateTime = (value?: string | number | Date | null): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
