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

// Risk snapshots expose exact minor-unit totals grouped by currency. Keep the
// display conversion string-based so large int64 values never pass through a
// lossy JavaScript Number.
export const formatMinorAmountsByCurrency = (amounts: unknown): string => {
  if (!amounts || typeof amounts !== 'object' || Array.isArray(amounts)) return '-'
  const entries = Object.entries(amounts as Record<string, unknown>)
    .map(([currency, value]) => [currency.trim().toUpperCase(), value] as const)
    .filter(([currency, value]) => currency && /^-?\d+$/.test(String(value ?? '')))
    .sort(([left], [right]) => left.localeCompare(right))
  if (entries.length === 0) return '-'

  return entries.map(([currency, value]) => {
    const raw = String(value)
    const negative = raw.startsWith('-')
    const digits = (negative ? raw.slice(1) : raw).replace(/^0+(?=\d)/, '') || '0'
    const minorUnits = ['JPY', 'KRW', 'CLP'].includes(currency) ? 0 : 2
    const sign = negative && digits !== '0' ? '-' : ''
    if (minorUnits === 0) return `${currency} ${sign}${digits}`
    const padded = digits.padStart(minorUnits + 1, '0')
    const split = padded.length - minorUnits
    return `${currency} ${sign}${padded.slice(0, split)}.${padded.slice(split)}`
  }).join(' · ')
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
