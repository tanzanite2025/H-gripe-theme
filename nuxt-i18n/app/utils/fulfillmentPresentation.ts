export type StorefrontFulfillmentMode = 'stock' | 'made_to_order' | 'mixed'

export type StorefrontProductionStatus =
  | 'not_applicable'
  | 'not_started'
  | 'started'
  | 'completed'
  | 'cancelled'

export const HIGH_VALUE_SIGNATURE_THRESHOLD_USD = 750

export const normalizeFulfillmentMode = (value: unknown): StorefrontFulfillmentMode => {
  const normalized = String(value || '').trim().toLowerCase()
  if (normalized === 'made_to_order') return 'made_to_order'
  if (normalized === 'mixed') return 'mixed'
  return 'stock'
}

export const normalizeProductionStatus = (value: unknown): StorefrontProductionStatus => {
  const normalized = String(value || '').trim().toLowerCase()
  switch (normalized) {
    case 'not_started':
    case 'started':
    case 'completed':
    case 'cancelled':
      return normalized
    default:
      return 'not_applicable'
  }
}

export const isMadeToOrderFulfillment = (value: unknown): boolean => {
  const mode = normalizeFulfillmentMode(value)
  return mode === 'made_to_order' || mode === 'mixed'
}

export const isProductionStarted = (value: unknown): boolean => {
  const status = normalizeProductionStatus(value)
  return status === 'started' || status === 'completed'
}
