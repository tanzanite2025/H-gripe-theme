import type { ApiRequestFunction } from '~/composables/useApiRequest'
import type {
  SchwalbeTireCatalogFilterOptions,
  SchwalbeTireCatalogWheelSizeOption,
} from '~/data/tireguides/schwalbeTireCatalogFilterModel'
import type { SchwalbeCatalogSort, SchwalbeTireCatalogFilterQueryState } from '~/data/tireguides/schwalbeTireCatalogFilterQuery'
import type { TireRimWidthReferenceSuggestion } from '~/data/tireguides/tireRimWidthReferencePresentation'

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
  source_checked_at: string
  product_exists: boolean
  wheel_size?: SchwalbeTireCatalogWheelSizeOption
  rim_compatibility?: SchwalbeTireCatalogRimCompatibility[]
  rim_width_guidance?: TireRimWidthReferenceSuggestion[]
}

/**
 * The pressure calculator receives only the product identity and pressure
 * reference it can display or use as a calculation input.
 */
export interface SchwalbeTirePressureReferenceCatalogItem {
  article_no: string
  model_name: string
  etrto: string
  inch_designation?: string
  min_pressure_bar?: number
  max_pressure_bar?: number
  min_pressure_psi?: number
  max_pressure_psi?: number
  source_checked_at: string
}

export type SchwalbeTireRimSystem = 'hooked' | 'hookless'

export interface SchwalbeTireCatalogRimCompatibility {
  rim_system: SchwalbeTireRimSystem
  status: 'supported'
}

// `useApiRequest` already prefixes paths with the configured API base
// (`/api/v1` in the storefront). Keep this endpoint relative to that base so
// development and production do not request `/api/v1/api/v1/...`.
const endpoint = '/products/schwalbe-tire-catalog'
const selectorEndpoint = '/products/schwalbe-tire-catalog/selector'
const tirePressureReferenceCatalogEndpoint = '/engineering/tire-pressure/reference-data/catalog'

export interface SchwalbeTireCatalogSelectorRequest extends SchwalbeTireCatalogFilterQueryState {
  search: string
  page: number
}

export interface SchwalbeTireCatalogSelectorPage {
  items: SchwalbeTireCatalogItem[]
  page: number
  page_size: number
  total: number
  total_pages: number
  filter_options: SchwalbeTireCatalogFilterOptions
}

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

const readWheelSizeProjection = (value: unknown): SchwalbeTireCatalogWheelSizeOption | undefined => {
  const item = asRecord(value)
  const key = optionalString(item?.value)
  const wheelDiameterIn = optionalString(item?.wheel_diameter_in)
  const beadSeatDiameterMm = optionalNumber(item?.bsd_mm)
  if (
    !key
    || !wheelDiameterIn
    || beadSeatDiameterMm === undefined
    || !Number.isSafeInteger(beadSeatDiameterMm)
    || beadSeatDiameterMm <= 0
    || key !== `${wheelDiameterIn}-${beadSeatDiameterMm}`
  ) return undefined

  return {
    value: key,
    wheelDiameterIn,
    beadSeatDiameterMm,
  }
}

const readRimCompatibility = (value: unknown): SchwalbeTireCatalogRimCompatibility[] => {
  if (!Array.isArray(value)) return []
  const seen = new Set<SchwalbeTireRimSystem>()
  return value.flatMap((candidate) => {
    const item = asRecord(candidate)
    const rimSystem = optionalString(item?.rim_system)
    const status = optionalString(item?.status)
    if (
      (rimSystem !== 'hooked' && rimSystem !== 'hookless')
      || status !== 'supported'
      || seen.has(rimSystem)
    ) return []
    seen.add(rimSystem)
    return [{ rim_system: rimSystem, status: 'supported' as const }]
  })
}

const readRimWidthReferenceGuidance = (value: unknown): TireRimWidthReferenceSuggestion[] => {
  if (!Array.isArray(value)) return []
  return value.flatMap((candidate) => {
    const item = asRecord(candidate)
    const rimSystem = optionalString(item?.rim_system)
    const resultKind = optionalString(item?.result_kind)
    const ranges = item?.rim_width_ranges
    if (
      (rimSystem !== 'hooked' && rimSystem !== 'hookless')
      || !resultKind
      || !Array.isArray(ranges)
      || ranges.length === 0
    ) return []
    const normalizedRanges = ranges.flatMap((range) => {
      const rangeRecord = asRecord(range)
      const min = optionalNumber(rangeRecord?.min)
      const max = optionalNumber(rangeRecord?.max)
      if (
        min === undefined
        || max === undefined
        || !Number.isSafeInteger(min)
        || !Number.isSafeInteger(max)
        || min < 0
        || max < min
      ) return []
      return [{ min, max }]
    })
    if (normalizedRanges.length !== ranges.length) return []
    return [item as unknown as TireRimWidthReferenceSuggestion]
  })
}

const readItem = (value: unknown): SchwalbeTireCatalogItem => {
  const item = asRecord(value)
  if (!item) throw new Error('Schwalbe catalog response contains an invalid item')
  const wheelSize = readWheelSizeProjection(item.wheel_size)
  const rimCompatibility = readRimCompatibility(item.rim_compatibility)
  const rimWidthGuidance = readRimWidthReferenceGuidance(item.rim_width_guidance)

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
    source_checked_at: requiredString(item.source_checked_at, 'source_checked_at'),
    product_exists: item.product_exists === true,
    ...(wheelSize ? { wheel_size: wheelSize } : {}),
    ...(rimCompatibility.length > 0
      ? { rim_compatibility: rimCompatibility }
      : {}),
    ...(rimWidthGuidance.length > 0
      ? { rim_width_guidance: rimWidthGuidance }
      : {}),
  }
}

/**
 * Reads the locally persisted Schwalbe pressure reference catalog in one
 * request. The endpoint intentionally has a different response shape from
 * the general selector, so its compact product/reference contract is parsed
 * into a calculator-specific item type.
 */
const readSchwalbeTirePressureReferenceItem = (value: unknown): SchwalbeTirePressureReferenceCatalogItem => {
  const item = asRecord(value)
  if (!item) throw new Error('Schwalbe tire pressure reference response contains an invalid product')

  return {
    article_no: requiredString(item.article_no, 'article_no'),
    model_name: requiredString(item.model_name, 'model_name'),
    etrto: requiredString(item.etrto, 'etrto'),
    ...(optionalString(item.inch_designation) ? { inch_designation: optionalString(item.inch_designation) } : {}),
    ...(optionalNumber(item.min_pressure_bar) !== undefined ? { min_pressure_bar: optionalNumber(item.min_pressure_bar) } : {}),
    ...(optionalNumber(item.max_pressure_bar) !== undefined ? { max_pressure_bar: optionalNumber(item.max_pressure_bar) } : {}),
    ...(optionalNumber(item.min_pressure_psi) !== undefined ? { min_pressure_psi: optionalNumber(item.min_pressure_psi) } : {}),
    ...(optionalNumber(item.max_pressure_psi) !== undefined ? { max_pressure_psi: optionalNumber(item.max_pressure_psi) } : {}),
    source_checked_at: requiredString(item.source_checked_at, 'source_checked_at'),
  }
}

export const fetchSchwalbeTirePressureReferenceCatalog = async (
  request: ApiRequestFunction,
): Promise<SchwalbeTirePressureReferenceCatalogItem[]> => {
  const response = await request<unknown>(
    tirePressureReferenceCatalogEndpoint,
    {},
    'Schwalbe tire pressure reference data is temporarily unavailable',
  )
  const envelope = asRecord(response)
  const payload = asRecord(envelope?.data)
  if (!payload || !Array.isArray(payload.items)) {
    throw new Error('Schwalbe tire pressure reference response is missing items')
  }

  return payload.items.map((candidate) => {
    const row = asRecord(candidate)
    const product = asRecord(row?.product)
    const pressure = asRecord(row?.pressure)
    if (!row || !product || !pressure) {
      throw new Error('Schwalbe tire pressure reference response contains an invalid item')
    }

    return readSchwalbeTirePressureReferenceItem({
      ...product,
      min_pressure_bar: pressure.min_pressure_bar,
      max_pressure_bar: pressure.max_pressure_bar,
      min_pressure_psi: pressure.min_pressure_psi,
      max_pressure_psi: pressure.max_pressure_psi,
      source_checked_at: row.source_checked_at,
    })
  })
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

const readStringFilterOptions = (value: unknown): { value: string }[] => {
  if (!Array.isArray(value)) return []
  return value.flatMap((candidate) => {
    const option = asRecord(candidate)
    const normalized = optionalString(option?.value as string | undefined)
    return normalized ? [{ value: normalized }] : []
  })
}

const readNumberFilterOptions = (value: unknown): { value: number }[] => {
  if (!Array.isArray(value)) return []
  return value.flatMap((candidate) => {
    const option = asRecord(candidate)
    const normalized = optionalNumber(option?.value as number | undefined)
    return normalized !== undefined && Number.isSafeInteger(normalized) ? [{ value: normalized }] : []
  })
}

const readWheelSizeFilterOptions = (value: unknown): SchwalbeTireCatalogWheelSizeOption[] => {
  if (!Array.isArray(value)) return []
  return value.flatMap((candidate) => {
    const option = asRecord(candidate)
    const key = optionalString(option?.value as string | undefined)
    const wheelDiameterIn = optionalString(option?.wheel_diameter_in as string | undefined)
    const bsdMm = optionalNumber(option?.bsd_mm)
    if (
      !key
      || !wheelDiameterIn
      || bsdMm === undefined
      || !Number.isSafeInteger(bsdMm)
      || bsdMm <= 0
      || key !== `${wheelDiameterIn}-${bsdMm}`
    ) return []
    return [{ value: key, wheelDiameterIn, beadSeatDiameterMm: bsdMm }]
  })
}

const readEBikeRatingFilterOptions = (value: unknown): { value: string | null }[] => {
  if (!Array.isArray(value)) return []
  const options: { value: string | null }[] = []
  value.forEach((candidate) => {
    const option = asRecord(candidate)
    if (!option || !Object.prototype.hasOwnProperty.call(option, 'value')) return
    if (option.value === null) {
      options.push({ value: null })
      return
    }
    const normalized = optionalString(typeof option.value === 'string' ? option.value : undefined)
    if (normalized) options.push({ value: normalized })
  })
  return options
}

const readSafeNonNegativeInteger = (value: unknown, fallback: number): number => {
  const normalized = typeof value === 'number' ? value : Number(value)
  return Number.isSafeInteger(normalized) && normalized >= 0 ? normalized : fallback
}

const readSelectorPage = (value: unknown): SchwalbeTireCatalogSelectorPage => {
  const envelope = asRecord(value)
  const payload = asRecord(envelope?.data) || envelope
  if (!payload || !Array.isArray(payload.items)) {
    throw new Error('Schwalbe selector response is missing items')
  }
  const rawOptions = asRecord(payload.filter_options)
  if (!rawOptions) throw new Error('Schwalbe selector response is missing filter_options')

  return {
    items: payload.items.map(readItem),
    page: readSafeNonNegativeInteger(payload.page, 1) || 1,
    page_size: readSafeNonNegativeInteger(payload.page_size, 20) || 20,
    total: readSafeNonNegativeInteger(payload.total, 0),
    total_pages: Math.max(1, readSafeNonNegativeInteger(payload.total_pages, 1)),
    filter_options: {
      modelNames: readStringFilterOptions(rawOptions.model_names),
      nominalTireWidthsMm: readNumberFilterOptions(rawOptions.nominal_tire_widths_mm),
      beadSeatDiametersMm: readNumberFilterOptions(rawOptions.bead_seat_diameters_mm),
      wheelSizes: readWheelSizeFilterOptions(rawOptions.wheel_sizes),
      versionLabels: [],
      casingConstructions: readStringFilterOptions(rawOptions.casing_constructions) as SchwalbeTireCatalogFilterOptions['casingConstructions'],
      compounds: readStringFilterOptions(rawOptions.compounds),
      colors: readStringFilterOptions(rawOptions.colors),
      beads: readStringFilterOptions(rawOptions.beads),
      seals: readStringFilterOptions(rawOptions.seals),
      eBikeRatings: readEBikeRatingFilterOptions(rawOptions.e_bike_ratings),
    },
  }
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

export const fetchSchwalbeTireCatalogSelectorPage = async (
  request: ApiRequestFunction,
  selectorQuery: SchwalbeTireCatalogSelectorRequest,
): Promise<SchwalbeTireCatalogSelectorPage> => {
  const params: Record<string, string | string[]> = {
    page: String(selectorQuery.page),
    sort: selectorQuery.sortBy as SchwalbeCatalogSort,
  }
  const search = selectorQuery.search.trim()
  if (search) params.search = search
  if (selectorQuery.modelName) params.model = selectorQuery.modelName
  if (selectorQuery.nominalTireWidthMinMm !== null && selectorQuery.nominalTireWidthMinMm !== undefined) {
    params.tire_width_min_mm = String(selectorQuery.nominalTireWidthMinMm)
  }
  if (selectorQuery.nominalTireWidthMaxMm !== null && selectorQuery.nominalTireWidthMaxMm !== undefined) {
    params.tire_width_max_mm = String(selectorQuery.nominalTireWidthMaxMm)
  }
  // Keep the old exact-value request available for legacy direct links and
  // non-upgraded consumers. New selector state uses the range endpoints.
  if (
    (selectorQuery.nominalTireWidthMinMm === null || selectorQuery.nominalTireWidthMinMm === undefined)
    && (selectorQuery.nominalTireWidthMaxMm === null || selectorQuery.nominalTireWidthMaxMm === undefined)
    && selectorQuery.nominalTireWidthsMm.length
  ) {
    params.tire_width_mm = selectorQuery.nominalTireWidthsMm.map(String)
  }
  if (selectorQuery.wheelSizeKeys.length) params.wheel_size = [...selectorQuery.wheelSizeKeys]
  if (selectorQuery.beadSeatDiametersMm.length) params.bead_seat_diameter_mm = selectorQuery.beadSeatDiametersMm.map(String)
  if (selectorQuery.minimumLoadKg !== null) params.min_load_kg = String(selectorQuery.minimumLoadKg)
  if (selectorQuery.casingConstructions.length) params.casing = [...selectorQuery.casingConstructions]
  if (selectorQuery.radialOnly) params.radial = '1'
  if (selectorQuery.beads.length) params.bead = [...selectorQuery.beads]
  if (selectorQuery.seals.length) params.seal = [...selectorQuery.seals]
  if (selectorQuery.eBikeRatings.length) {
    params.e_bike_rating = selectorQuery.eBikeRatings.map(rating => rating ?? 'none')
  }
  if (selectorQuery.colors.length) params.color = [...selectorQuery.colors]
  if (selectorQuery.compounds.length) params.compound = [...selectorQuery.compounds]

  const response = await request<unknown>(
    selectorEndpoint,
    { params },
    'Schwalbe catalog is temporarily unavailable',
  )
  return readSelectorPage(response)
}

export const schwalbeTireCatalogEndpoint = endpoint
export const schwalbeTireCatalogSelectorEndpoint = selectorEndpoint
export const schwalbeTirePressureReferenceCatalogEndpoint = tirePressureReferenceCatalogEndpoint
