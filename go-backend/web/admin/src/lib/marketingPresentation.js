export const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const formatMoney = (amount) => Number(amount || 0).toFixed(2)

export const formatCurrency = (amount, currency = '') => {
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  try {
    if (!normalizedCurrency) throw new Error('missing currency')
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: normalizedCurrency }).format(Number(amount || 0))
  } catch {
    return `${normalizedCurrency || '币种缺失'} ${formatMoney(amount)}`.trim()
  }
}

export const formatRate = (rate) => `${Number(rate || 0).toFixed(2)}%`
export const couponValue = (coupon) => coupon.type === 'percentage' ? `${formatMoney(coupon.value)}%` : `¥${formatMoney(coupon.value)}`

export const couponStatus = (coupon) => {
  const now = Date.now()
  if (!coupon.enabled) return { label: '已停用', tone: 'gray' }
  if (coupon.end_date && now > new Date(coupon.end_date).getTime()) return { label: '已过期', tone: 'amber' }
  if (coupon.start_date && now < new Date(coupon.start_date).getTime()) return { label: '未开始', tone: 'blue' }
  return { label: '生效中', tone: 'green' }
}

export const giftCardStatusName = (status) => ({ active: '活跃', used: '已使用', expired: '已过期', cancelled: '已取消' })[status] || status || '-'
export const giftCardStatusTone = (status) => ({ active: 'green', used: 'blue', expired: 'amber', cancelled: 'coral' })[status] || 'gray'
export const transactionTypeName = (type) => ({ issue: '发行', use: '消费', refund: '退款' })[type] || type || '-'
export const loyaltyTypeName = (type) => ({ earn: '获得', spend: '消费', expire: '过期', adjust: '调整', refund: '退回' })[type] || type || '-'

export const giftCardStatusOptions = (card) => {
  const current = card?.status || 'active'
  const options = [{ value: current, label: giftCardStatusName(current) }]
  if (current === 'active') {
    if (Number(card?.balance || 0) <= 0) options.push({ value: 'used', label: '已使用' })
    options.push({ value: 'expired', label: '已过期' }, { value: 'cancelled', label: '已取消' })
  }
  return options
}

export const toDateTimeLocal = (value) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

export const toISO = (value) => value ? new Date(value).toISOString() : null
