type StatusTone = 'gray' | 'amber' | 'blue' | 'green' | 'coral'

// This matches the backend's admin-facing currency catalog. The dashboard
// formatter deliberately supports a broader ISO set, whose ISK rule does not
// match the commercial-money contract used by Merchant.
const zeroDecimalCommercialCurrencies = new Set(['JPY', 'KRW', 'CLP'])

export const commercialMinorUnitsForCurrency = (currency?: string | null): number => (
  zeroDecimalCommercialCurrencies.has(String(currency || '').trim().toUpperCase()) ? 0 : 2
)

interface CouponLike {
  type?: string | null
  value_minor?: number | string | null
  value_rate_decimal?: number | string | null
  currency?: string | null
  enabled?: boolean
  start_date?: string | null
  end_date?: string | null
}

interface StatusDisplay {
  label: string
  tone: StatusTone
}

export const formatDate = (dateString?: string | number | Date | null): string => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
export const formatMoney = (amount?: number | string | null): string => Number(amount || 0).toFixed(2)

export const formatCurrency = (amount?: number | string | null, currency = ''): string => {
  const normalizedCurrency = String(currency || '').trim().toUpperCase()
  try {
    if (!normalizedCurrency) throw new Error('missing currency')
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: normalizedCurrency }).format(Number(amount || 0))
  } catch {
    return `${normalizedCurrency || '币种缺失'} ${formatMoney(amount)}`.trim()
  }
}

export const formatRate = (rate?: number | string | null): string => `${Number(rate || 0).toFixed(2)}%`
export const minorToMajor = (minor: number | string | null | undefined, currency?: string | null): number => {
  const amountMinor = Number(minor)
  return Number.isFinite(amountMinor) ? amountMinor / (10 ** commercialMinorUnitsForCurrency(currency)) : 0
}

const parseMinorInteger = (value: number | string | bigint | null | undefined): bigint | null => {
  if (typeof value === 'bigint') return value
  if (typeof value === 'number') return Number.isSafeInteger(value) ? BigInt(value) : null
  const text = String(value ?? '').trim()
  return /^[-+]?\d+$/.test(text) ? BigInt(text) : null
}

/** Returns an exact display decimal for an integer minor-unit amount. */
export const minorToMajorDecimal = (minor: number | string | bigint | null | undefined, currency?: string | null): string => {
  const amountMinor = parseMinorInteger(minor)
  if (amountMinor == null) return ''

  const units = commercialMinorUnitsForCurrency(currency)
  const negative = amountMinor < 0n
  const absolute = negative ? -amountMinor : amountMinor
  const scale = 10n ** BigInt(units)
  const whole = absolute / scale
  const fraction = absolute % scale
  const suffix = units > 0 ? `.${fraction.toString().padStart(units, '0')}` : ''
  return `${negative ? '-' : ''}${whole.toString()}${suffix}`
}

/**
 * Parses a user-entered decimal without floating-point rounding. A null result
 * means the input is malformed, exceeds the currency precision, or cannot be
 * represented safely in the JSON number accepted by the API.
 */
export const majorToMinor = (amount: number | string | null | undefined, currency?: string | null): number | null => {
  const text = String(amount ?? '').trim()
  if (!/^\d+(?:\.\d+)?$/.test(text)) return null

  const units = commercialMinorUnitsForCurrency(currency)
  const [wholePart, fractionPart = ''] = text.split('.')
  const excessFraction = fractionPart.slice(units)
  if (excessFraction && /[1-9]/.test(excessFraction)) return null

  const scale = 10n ** BigInt(units)
  const normalizedFraction = fractionPart.slice(0, units).padEnd(units, '0')
  const minor = BigInt(wholePart) * scale + BigInt(normalizedFraction || '0')
  return minor <= BigInt(Number.MAX_SAFE_INTEGER) ? Number(minor) : null
}
export const couponValue = (coupon: CouponLike): string => coupon.type === 'percentage'
  ? `${formatMoney(coupon.value_rate_decimal)}%`
  : formatCurrency(minorToMajor(coupon.value_minor, coupon.currency), coupon.currency || 'USD')

export const couponStatus = (coupon: CouponLike): StatusDisplay => {
  const now = Date.now()
  if (!coupon.enabled) return { label: '已停用', tone: 'gray' }
  if (coupon.end_date && now > new Date(coupon.end_date).getTime()) return { label: '已过期', tone: 'amber' }
  if (coupon.start_date && now < new Date(coupon.start_date).getTime()) return { label: '未开始', tone: 'blue' }
  return { label: '生效中', tone: 'green' }
}

export const loyaltyTypeName = (type?: string | null): string => ({ earn: '获得', spend: '消费', expire: '过期', adjust: '调整', refund: '退回' })[type || ''] || type || '-'

export const toDateTimeLocal = (value?: string | number | Date | null): string => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

export const toISO = (value?: string | number | Date | null): string | null => value ? new Date(value).toISOString() : null
