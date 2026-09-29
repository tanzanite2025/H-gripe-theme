import type { ApiRequestFunction } from '~/composables/useApiRequest'
import type { SchwalbeTireRimWidthCombinationRule } from '~/data/tireguides/schwalbeTireCatalogFilterModel'

export interface SchwalbeTireRimWidthCombinationRuleRecord extends SchwalbeTireRimWidthCombinationRule {
  sourceBasis: string
  sourceVersion: string
  sourceUrl: string
  sourceCheckedAt: string
}

const endpoint = '/products/schwalbe-tire-rim-width-combination-rules'

const asRecord = (value: unknown): Record<string, unknown> | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  return value as Record<string, unknown>
}

const requiredString = (value: unknown, field: string): string => {
  const normalized = String(value ?? '').trim()
  if (!normalized) throw new Error(`Schwalbe rim-width rule response is missing ${field}`)
  return normalized
}

const requiredPositiveInteger = (value: unknown, field: string): number => {
  const parsed = typeof value === 'number' ? value : Number(value)
  if (!Number.isSafeInteger(parsed) || parsed <= 0) {
    throw new Error(`Schwalbe rim-width rule response has an invalid ${field}`)
  }
  return parsed
}

const readRule = (value: unknown): SchwalbeTireRimWidthCombinationRuleRecord => {
  const item = asRecord(value)
  if (!item) throw new Error('Schwalbe rim-width rule response contains an invalid item')

  return {
    tireWidthMinMm: requiredPositiveInteger(item.tire_width_min_mm, 'tire_width_min_mm'),
    tireWidthMaxMm: requiredPositiveInteger(item.tire_width_max_mm, 'tire_width_max_mm'),
    innerRimWidthMinMm: requiredPositiveInteger(item.inner_rim_width_min_mm, 'inner_rim_width_min_mm'),
    innerRimWidthMaxMm: requiredPositiveInteger(item.inner_rim_width_max_mm, 'inner_rim_width_max_mm'),
    sourceBasis: requiredString(item.source_basis, 'source_basis'),
    sourceVersion: requiredString(item.source_version, 'source_version'),
    sourceUrl: requiredString(item.source_url, 'source_url'),
    sourceCheckedAt: requiredString(item.source_checked_at, 'source_checked_at'),
  }
}

const extractItems = (value: unknown): unknown[] => {
  if (Array.isArray(value)) return value
  const root = asRecord(value)
  if (!root) return []
  if (Array.isArray(root.data)) return root.data
  if (Array.isArray(root.items)) return root.items
  const nested = asRecord(root.data)
  if (nested && Array.isArray(nested.items)) return nested.items
  return []
}

/**
 * Fetches migration-362 rules without hardcoding the official ranges in UI
 * code. The paginated storefront selector remains the authoritative result
 * filter; this adapter is for source context and complete-data consumers.
 */
export const fetchSchwalbeTireRimWidthCombinationRules = async (
  request: ApiRequestFunction,
): Promise<SchwalbeTireRimWidthCombinationRuleRecord[]> => {
  const response = await request<unknown>(
    endpoint,
    {},
    'Schwalbe rim-width guidance is temporarily unavailable',
  )
  return extractItems(response).map(readRule)
}

export const schwalbeTireRimWidthCombinationRulesEndpoint = endpoint
