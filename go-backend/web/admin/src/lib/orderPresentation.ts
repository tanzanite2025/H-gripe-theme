type StatusTone = 'gray' | 'green' | 'amber' | 'blue' | 'coral'

interface StatusOption {
  label: string
  value: string
}

interface ShippingAddressLike {
  first_name?: string | null
  last_name?: string | null
  address_1?: string | null
  address_2?: string | null
}

export const orderStatusOptions: StatusOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' },
  { label: '支付超时', value: 'payment_expired' },
  { label: '已取消', value: 'cancelled' },
  { label: '已退款', value: 'refunded' }
]

export const paymentStatusOptions: StatusOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '未支付', value: 'unpaid' },
  { label: '已支付', value: 'paid' },
  { label: '已超时', value: 'expired' },
  { label: '已退款', value: 'refunded' }
]

export const shippingStatusOptions: StatusOption[] = [
  { label: '全部状态', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已送达', value: 'delivered' }
]

export const editableOrderStatusOptions = orderStatusOptions.filter((option) => !['all', 'paid', 'payment_expired', 'refunded'].includes(option.value))
export const editableShippingStatusOptions = shippingStatusOptions.filter((option) => option.value !== 'all')

export const getOrderStatusName = (status?: string | null): string => orderStatusOptions.find((option) => option.value === status)?.label || status || '-'
export const orderStatusTone = (status?: string | null): StatusTone => ({
  pending: 'gray',
  paid: 'green',
  processing: 'amber',
  shipped: 'blue',
  completed: 'green',
  payment_expired: 'amber',
  cancelled: 'coral',
  refunded: 'amber'
} as Record<string, StatusTone>)[status || ''] || 'gray'

export const getPaymentStatusName = (status?: string | null): string => paymentStatusOptions.find((option) => option.value === status)?.label || status || '-'
export const paymentStatusTone = (status?: string | null): StatusTone => ({ unpaid: 'gray', paid: 'green', expired: 'amber', refunded: 'amber' } as Record<string, StatusTone>)[status || ''] || 'gray'
export const getShippingStatusName = (status?: string | null): string => shippingStatusOptions.find((option) => option.value === status)?.label || status || '-'
export const shippingStatusTone = (status?: string | null): StatusTone => ({ pending: 'gray', processing: 'amber', shipped: 'blue', delivered: 'green' } as Record<string, StatusTone>)[status || ''] || 'gray'

export const trackingSyncStatusName = (status?: string | null): string => ({
  pending: '待同步',
  syncing: '同步中',
  synced: '已同步',
  failed: '同步失败'
})[status || ''] || '未建立'
export const trackingSyncStatusTone = (status?: string | null): StatusTone => ({
  pending: 'gray',
  syncing: 'blue',
  synced: 'green',
  failed: 'coral'
} as Record<string, StatusTone>)[status || ''] || 'gray'
export const trackingRegistrationStatusName = (status?: string | null): string => ({
  pending: '待登记',
  registered: '已登记',
  failed: '登记失败'
})[status || ''] || '未建立'

export const formatDate = (dateString?: string | number | Date | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const formatMoney = (amount?: number | string | null): string => Number(amount || 0).toFixed(2)
export const shippingName = (address?: ShippingAddressLike | null): string => [address?.first_name, address?.last_name].filter(Boolean).join(' ') || '-'
export const shippingAddressLine = (address?: ShippingAddressLike | null): string => [address?.address_1, address?.address_2].filter(Boolean).join(' ') || '-'

export const numericSelectID = (value?: number | string | null): number | null => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

export const selectValueFromID = (value?: number | string | null): string => numericSelectID(value) ? String(value) : 'none'
