export interface RiskStrategyPagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export const formatDate = (dateString: unknown): string => (
  dateString ? new Date(dateString as string | number | Date).toLocaleString('zh-CN') : '-'
)

export const formatMoney = (amount: unknown, currency = ''): string => {
  const value = Number(amount || 0)
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  try {
    if (!normalizedCurrency) throw new Error('missing currency')
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: normalizedCurrency }).format(value)
  } catch {
    return `${normalizedCurrency || '币种缺失'} ${value.toFixed(2)}`
  }
}

export const isEvidenceSoon = (dateString: unknown): boolean => {
  if (!dateString) return false
  const due = new Date(dateString as string | number | Date).getTime()
  return Number.isFinite(due) && due - Date.now() < 3 * 24 * 60 * 60 * 1000
}

export const applyPaged = <T>(
  target: { value: T[] },
  pagination: RiskStrategyPagination,
  payload: {
    data?: T[]
    pagination?: Partial<RiskStrategyPagination>
  },
): void => {
  target.value = payload.data || []
  Object.assign(pagination, {
    page: payload.pagination?.page || 1,
    page_size: payload.pagination?.page_size || 20,
    total: payload.pagination?.total || 0,
    total_pages: payload.pagination?.total_pages || 1,
  })
}

export const refundRecommendationSourceLabel = (sourceKind?: string): string => {
  if (sourceKind === 'early_fraud_warning') return '早期欺诈预警'
  if (sourceKind === 'dispute') return '争议/拒付'
  return sourceKind || '-'
}
