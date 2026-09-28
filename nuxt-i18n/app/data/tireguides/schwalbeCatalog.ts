import type { ApiRequestFunction } from '~/composables/useApiRequest'

export interface SchwalbeTireCatalogItem {
  article_no: string
  ean?: string
  model_name: string
  etrto: string
  inch_designation?: string
  weight_g?: number
  version_label?: string
  compound?: string
  color?: string
  bead?: string
  e_bike_rating?: string
  epi?: number
  load_kg?: number
  seal?: string
  tread?: string
  min_pressure_bar?: number
  max_pressure_bar?: number
  min_pressure_psi?: number
  max_pressure_psi?: number
  source_url: string
  source_checked_at: string
  product_exists: boolean
}

// `useApiRequest` already prefixes paths with the configured API base
// (`/api/v1` in the storefront). Keep this endpoint relative to that base so
// development and production do not request `/api/v1/api/v1/...`.
const endpoint = '/products/schwalbe-tire-catalog'

const asRecord = (value: unknown): Record<string, unknown> | null => {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  return value as Record<string, unknown>
}

const requiredString = (value: unknown, field: string): string => {
  const normalized = String(value ?? '').trim()
  if (!normalized) throw new Error(`Schwalbe catalog response is missing ${field}`)
  return normalized
}

const optionalString = (value: unknown): string | undefined => {
  const normalized = String(value ?? '').trim()
  return normalized || undefined
}

const optionalNumber = (value: unknown): number | undefined => {
  if (value === null || value === undefined || value === '') return undefined
  const number = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(number) ? number : undefined
}

const readItem = (value: unknown): SchwalbeTireCatalogItem => {
  const item = asRecord(value)
  if (!item) throw new Error('Schwalbe catalog response contains an invalid item')

  return {
    article_no: requiredString(item.article_no, 'article_no'),
    ...(optionalString(item.ean) ? { ean: optionalString(item.ean) } : {}),
    model_name: requiredString(item.model_name, 'model_name'),
    etrto: requiredString(item.etrto, 'etrto'),
    ...(optionalString(item.inch_designation) ? { inch_designation: optionalString(item.inch_designation) } : {}),
    ...(optionalNumber(item.weight_g) !== undefined ? { weight_g: optionalNumber(item.weight_g) } : {}),
    ...(optionalString(item.version_label) ? { version_label: optionalString(item.version_label) } : {}),
    ...(optionalString(item.compound) ? { compound: optionalString(item.compound) } : {}),
    ...(optionalString(item.color) ? { color: optionalString(item.color) } : {}),
    ...(optionalString(item.bead) ? { bead: optionalString(item.bead) } : {}),
    ...(optionalString(item.e_bike_rating) ? { e_bike_rating: optionalString(item.e_bike_rating) } : {}),
    ...(optionalNumber(item.epi) !== undefined ? { epi: optionalNumber(item.epi) } : {}),
    ...(optionalNumber(item.load_kg) !== undefined ? { load_kg: optionalNumber(item.load_kg) } : {}),
    ...(optionalString(item.seal) ? { seal: optionalString(item.seal) } : {}),
    ...(optionalString(item.tread) ? { tread: optionalString(item.tread) } : {}),
    ...(optionalNumber(item.min_pressure_bar) !== undefined ? { min_pressure_bar: optionalNumber(item.min_pressure_bar) } : {}),
    ...(optionalNumber(item.max_pressure_bar) !== undefined ? { max_pressure_bar: optionalNumber(item.max_pressure_bar) } : {}),
    ...(optionalNumber(item.min_pressure_psi) !== undefined ? { min_pressure_psi: optionalNumber(item.min_pressure_psi) } : {}),
    ...(optionalNumber(item.max_pressure_psi) !== undefined ? { max_pressure_psi: optionalNumber(item.max_pressure_psi) } : {}),
    source_url: requiredString(item.source_url, 'source_url'),
    source_checked_at: requiredString(item.source_checked_at, 'source_checked_at'),
    product_exists: item.product_exists === true,
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

export const fetchSchwalbeTireCatalog = async (
  request: ApiRequestFunction,
  search = '',
): Promise<SchwalbeTireCatalogItem[]> => {
  const normalizedSearch = search.trim()
  const response = await request<unknown>(
    endpoint,
    normalizedSearch ? { params: { search: normalizedSearch } } : {},
    'Schwalbe catalog is temporarily unavailable',
  )
  return extractItems(response).map(readItem)
}

export const schwalbeTireCatalogEndpoint = endpoint
