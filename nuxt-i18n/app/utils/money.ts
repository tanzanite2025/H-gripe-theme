/**
 * Storefront money helpers.
 *
 * Cart and checkout calculations use minor-unit integers. Conversion to a
 * major-unit number is kept at the rendering/protocol boundary only.
 */
const ZERO_DECIMAL_CURRENCIES = new Set([
  'BIF', 'CLP', 'DJF', 'GNF', 'JPY', 'KMF', 'KRW', 'MGA', 'PYG', 'RWF',
  'UGX', 'VND', 'VUV', 'XAF', 'XOF', 'XPF',
])

const THREE_DECIMAL_CURRENCIES = new Set([
  'BHD', 'IQD', 'JOD', 'KWD', 'LYD', 'OMR', 'TND',
])

export const normalizeMoneyCurrency = (value: unknown, fallback = 'USD') => {
  const code = String(value || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(code) ? code : fallback
}

export const currencyMinorUnits = (value: unknown) => {
  const currency = normalizeMoneyCurrency(value)
  if (THREE_DECIMAL_CURRENCIES.has(currency)) return 3
  if (ZERO_DECIMAL_CURRENCIES.has(currency)) return 0
  return 2
}

/** Convert a display-only major amount into a minor integer at an input boundary. */
export const majorToMinor = (value: unknown, currency: unknown) => {
  const amount = Number(value)
  if (!Number.isFinite(amount)) return 0
  const minor = Math.round(amount * (10 ** currencyMinorUnits(currency)))
  return Number.isSafeInteger(minor) ? minor : 0
}

/** Convert a minor integer for display/protocol formatting only. */
export const minorToMajor = (value: unknown, currency: unknown) => {
  const amountMinor = Number(value)
  if (!Number.isFinite(amountMinor)) return 0
  return amountMinor / (10 ** currencyMinorUnits(currency))
}

export const formatMinorMoney = (
  value: unknown,
  currency: unknown,
  locale = 'en-US',
) => {
  const normalizedCurrency = normalizeMoneyCurrency(currency)
  const amount = minorToMajor(value, normalizedCurrency)
  try {
    return new Intl.NumberFormat(locale, {
      style: 'currency',
      currency: normalizedCurrency,
    }).format(amount)
  } catch {
    return `${normalizedCurrency} ${amount}`
  }
}

