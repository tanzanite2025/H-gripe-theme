export const orderStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
  { label: '已退款', value: 'refunded' }
]

export const paymentStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '未支付', value: 'unpaid' },
  { label: '已支付', value: 'paid' },
  { label: '已退款', value: 'refunded' }
]

export const shippingStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已送达', value: 'delivered' }
]

export const editableOrderStatusOptions = orderStatusOptions.filter((option) => !['all', 'paid', 'refunded'].includes(option.value))
export const editableShippingStatusOptions = shippingStatusOptions.filter((option) => option.value !== 'all')

export const getOrderStatusName = (status) => orderStatusOptions.find((option) => option.value === status)?.label || status
export const orderStatusTone = (status) => ({
  pending: 'gray',
  paid: 'green',
  processing: 'amber',
  shipped: 'blue',
  completed: 'green',
  cancelled: 'coral',
  refunded: 'amber'
})[status] || 'gray'

export const getPaymentStatusName = (status) => paymentStatusOptions.find((option) => option.value === status)?.label || status
export const paymentStatusTone = (status) => ({ unpaid: 'gray', paid: 'green', refunded: 'amber' })[status] || 'gray'
export const getShippingStatusName = (status) => shippingStatusOptions.find((option) => option.value === status)?.label || status
export const shippingStatusTone = (status) => ({ pending: 'gray', processing: 'amber', shipped: 'blue', delivered: 'green' })[status] || 'gray'

export const trackingSyncStatusName = (status) => ({
  pending: '待同步',
  syncing: '同步中',
  synced: '已同步',
  failed: '同步失败'
})[status] || '未建立'
export const trackingSyncStatusTone = (status) => ({
  pending: 'gray',
  syncing: 'blue',
  synced: 'green',
  failed: 'coral'
})[status] || 'gray'
export const trackingRegistrationStatusName = (status) => ({
  pending: '待登记',
  registered: '已登记',
  failed: '登记失败'
})[status] || '未建立'

export const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const formatMoney = (amount) => Number(amount || 0).toFixed(2)
export const shippingName = (address) => [address?.first_name, address?.last_name].filter(Boolean).join(' ') || '-'
export const shippingAddressLine = (address) => [address?.address_1, address?.address_2].filter(Boolean).join(' ') || '-'

export const numericSelectID = (value) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

export const selectValueFromID = (value) => numericSelectID(value) ? String(value) : 'none'
