import { SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS } from './schwalbeTireCatalogFilterModel.js'

/**
 * The selector exposes weight ordering to users. The legacy values are kept
 * here only so older API consumers can be migrated without a type break; the
 * page parser normalizes them to the new default below.
 */
export type SchwalbeCatalogSort = 'weight_asc' | 'weight_desc' | 'model' | 'etrto' | 'article'

export interface SchwalbeTireCatalogFilterQueryState {
  modelName: string | null
  /** User-entered rim inner width used for official possible-combination matching. */
  innerRimWidthMm: number | null
  /**
   * Inclusive nominal-width range used by the selector UI. Empty endpoints
   * mean that side of the range is unbounded.
   */
  nominalTireWidthMinMm: number | null
  nominalTireWidthMaxMm: number | null
  /** @deprecated Kept for old shared links and API clients. */
  nominalTireWidthsMm: number[]
  /** New user-facing size facet. Each value pairs wheel diameter with BSD. */
  wheelSizeKeys: string[]
  /** @deprecated Kept for old shared links and API clients. */
  beadSeatDiametersMm: number[]
  minimumLoadKg: number | null
  casingConstructions: string[]
  radialOnly: boolean
  beads: string[]
  seals: string[]
  eBikeRatings: (string | null)[]
  colors: string[]
  compounds: string[]
  sortBy: SchwalbeCatalogSort
}

export type SchwalbeTireCatalogFilterQuery = Record<string, unknown>

export const SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS = {
  modelName: 'model',
  innerRimWidthMm: 'inner_rim_width_mm',
  nominalTireWidthMinMm: 'tire_width_min_mm',
  nominalTireWidthMaxMm: 'tire_width_max_mm',
  /** @deprecated Use the inclusive min/max keys for new links. */
  nominalTireWidthsMm: 'tire_width_mm',
  wheelSizeKeys: 'wheel_size',
  /** @deprecated Use wheel_size for new selector interactions. */
  beadSeatDiametersMm: 'bead_seat_diameter_mm',
  minimumLoadKg: 'min_load_kg',
  casingConstructions: 'casing',
  radialOnly: 'radial',
  beads: 'bead',
  seals: 'seal',
  eBikeRatings: 'e_bike_rating',
  colors: 'color',
  compounds: 'compound',
  sortBy: 'sort',
} as const

/** Reserved URL value for a catalog row with no published E-Bike rating. */
export const SCHWALBE_E_BIKE_UNRATED_QUERY_VALUE = 'none'

const SCHWALBE_CATALOG_SORT_VALUES: readonly SchwalbeCatalogSort[] = [
  'weight_asc',
  'weight_desc',
]

const readFirstString = (value: unknown): string | null => {
  const candidate = Array.isArray(value) ? value[0] : value
  if (typeof candidate !== 'string') return null
  return candidate.trim() || null
}

const readEnabledFlag = (value: unknown): boolean => {
  const candidate = Array.isArray(value) ? value[0] : value
  return candidate === true || candidate === '1' || candidate === 'true'
}

const readPositiveIntegerValues = (value: unknown): number[] => {
  const candidates = Array.isArray(value) ? value : [value]
  const parsed = candidates.flatMap((candidate) => {
    if (typeof candidate !== 'string' || !/^[1-9]\d*$/.test(candidate.trim())) return []
    const number = Number(candidate)
    return Number.isSafeInteger(number) ? [number] : []
  })
  return [...new Set(parsed)].sort((left, right) => left - right)
}

const readPositiveIntegerValue = (value: unknown): number | null => {
  const candidate = readFirstString(value)
  if (candidate === null || !/^[1-9]\d*$/.test(candidate)) return null
  const number = Number(candidate)
  return Number.isSafeInteger(number) ? number : null
}

const readPositiveFiniteNumber = (value: unknown): number | null => {
  const candidate = readFirstString(value)
  if (candidate === null) return null
  const number = Number(candidate)
  return Number.isFinite(number) && number > 0 ? number : null
}

const SCHWALBE_WHEEL_SIZE_KEY_PATTERN = /^[1-9]\d*(?:\.\d+)?-[1-9]\d*$/

const readStringValues = (value: unknown): string[] => {
  const candidates = Array.isArray(value) ? value : [value]
  return [...new Set(candidates.flatMap((candidate) => {
    if (typeof candidate !== 'string') return []
    const normalized = candidate.trim()
    return normalized ? [normalized] : []
  }))].sort((left, right) => left.localeCompare(right))
}

const readWheelSizeValues = (value: unknown): string[] => (
  readStringValues(value).filter(candidate => SCHWALBE_WHEEL_SIZE_KEY_PATTERN.test(candidate))
)

const schwalbeCasingConstructionLabelSet = new Set<string>(SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS)

const readCasingConstructionValues = (value: unknown): string[] => (
  readStringValues(value).filter(label => schwalbeCasingConstructionLabelSet.has(label))
)

const readEBikeRatingValues = (value: unknown): (string | null)[] => {
  const ratings = readStringValues(value).map(rating => (
    rating === SCHWALBE_E_BIKE_UNRATED_QUERY_VALUE ? null : rating
  ))
  return [...new Set(ratings)].sort((left, right) => {
    if (left === null) return 1
    if (right === null) return -1
    return left.localeCompare(right)
  })
}

const readSort = (value: unknown): SchwalbeCatalogSort => {
  const candidate = readFirstString(value)
  return SCHWALBE_CATALOG_SORT_VALUES.includes(candidate as SchwalbeCatalogSort)
    ? candidate as SchwalbeCatalogSort
    : 'weight_asc'
}

/** Reads only selector state; search, page, and unrelated query keys stay separate. */
export const parseSchwalbeTireCatalogFilterQuery = (
  query: SchwalbeTireCatalogFilterQuery,
): SchwalbeTireCatalogFilterQueryState => {
  const keys = SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS
  let nominalTireWidthMinMm = readPositiveIntegerValue(query[keys.nominalTireWidthMinMm])
  let nominalTireWidthMaxMm = readPositiveIntegerValue(query[keys.nominalTireWidthMaxMm])
  // Treat a manually reversed URL as the same closed interval. This keeps
  // direct links deterministic and prevents the two slider thumbs crossing.
  if (
    nominalTireWidthMinMm !== null
    && nominalTireWidthMaxMm !== null
    && nominalTireWidthMinMm > nominalTireWidthMaxMm
  ) {
    [nominalTireWidthMinMm, nominalTireWidthMaxMm] = [nominalTireWidthMaxMm, nominalTireWidthMinMm]
  }
  const hasNominalTireWidthRange = nominalTireWidthMinMm !== null || nominalTireWidthMaxMm !== null

  return {
    modelName: readFirstString(query[keys.modelName]),
    innerRimWidthMm: readPositiveFiniteNumber(query[keys.innerRimWidthMm]),
    nominalTireWidthMinMm,
    nominalTireWidthMaxMm,
    // A range takes precedence when both contracts are present. This keeps a
    // manually edited URL deterministic while preserving old exact links.
    nominalTireWidthsMm: hasNominalTireWidthRange
      ? []
      : readPositiveIntegerValues(query[keys.nominalTireWidthsMm]),
    wheelSizeKeys: readWheelSizeValues(query[keys.wheelSizeKeys]),
    beadSeatDiametersMm: readPositiveIntegerValues(query[keys.beadSeatDiametersMm]),
    minimumLoadKg: readPositiveFiniteNumber(query[keys.minimumLoadKg]),
    casingConstructions: readCasingConstructionValues(query[keys.casingConstructions]),
    radialOnly: readEnabledFlag(query[keys.radialOnly]),
    beads: readStringValues(query[keys.beads]),
    seals: readStringValues(query[keys.seals]),
    eBikeRatings: readEBikeRatingValues(query[keys.eBikeRatings]),
    colors: readStringValues(query[keys.colors]),
    compounds: readStringValues(query[keys.compounds]),
    sortBy: readSort(query[keys.sortBy]),
  }
}

const normalizedPositiveIntegers = (values: readonly number[]): number[] => (
  [...new Set(values.filter(value => Number.isSafeInteger(value) && value > 0))]
    .sort((left, right) => left - right)
)

const normalizePositiveInteger = (value: number | null | undefined): number | null => (
  value !== null
  && value !== undefined
  && Number.isSafeInteger(value)
  && value > 0
    ? value
    : null
)

const normalizePositiveFiniteNumber = (value: number | null | undefined): number | null => (
  value !== null
  && value !== undefined
  && Number.isFinite(value)
  && value > 0
    ? value
    : null
)

const normalizedStrings = (values: readonly string[]): string[] => (
  [...new Set(values.map(value => value.trim()).filter(Boolean))]
    .sort((left, right) => left.localeCompare(right))
)

const normalizedEBikeRatings = (values: readonly (string | null)[]): (string | null)[] => (
  [...new Set(values.flatMap((value) => {
    if (value === null) return [null]
    const normalized = value.trim()
    return normalized ? [normalized] : []
  }))].sort((left, right) => {
    if (left === null) return 1
    if (right === null) return -1
    return left.localeCompare(right)
  })
)

/** Merges selector state into a route query while preserving search and other keys. */
export const mergeSchwalbeTireCatalogFilterQuery = (
  query: SchwalbeTireCatalogFilterQuery,
  state: SchwalbeTireCatalogFilterQueryState,
): SchwalbeTireCatalogFilterQuery => {
  const nextQuery = { ...query }
  const keys = SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS

  delete nextQuery[keys.modelName]
  delete nextQuery[keys.innerRimWidthMm]
  delete nextQuery[keys.nominalTireWidthMinMm]
  delete nextQuery[keys.nominalTireWidthMaxMm]
  delete nextQuery[keys.nominalTireWidthsMm]
  delete nextQuery[keys.wheelSizeKeys]
  delete nextQuery[keys.beadSeatDiametersMm]
  delete nextQuery[keys.minimumLoadKg]
  delete nextQuery[keys.casingConstructions]
  delete nextQuery[keys.radialOnly]
  delete nextQuery[keys.beads]
  delete nextQuery[keys.seals]
  delete nextQuery[keys.eBikeRatings]
  delete nextQuery[keys.colors]
  delete nextQuery[keys.compounds]
  delete nextQuery[keys.sortBy]

  if (state.modelName?.trim()) nextQuery[keys.modelName] = state.modelName.trim()

  const innerRimWidthMm = normalizePositiveFiniteNumber(state.innerRimWidthMm)
  if (innerRimWidthMm !== null) nextQuery[keys.innerRimWidthMm] = String(innerRimWidthMm)

  let tireWidthMinMm = normalizePositiveInteger(state.nominalTireWidthMinMm)
  let tireWidthMaxMm = normalizePositiveInteger(state.nominalTireWidthMaxMm)
  if (tireWidthMinMm !== null && tireWidthMaxMm !== null && tireWidthMinMm > tireWidthMaxMm) {
    [tireWidthMinMm, tireWidthMaxMm] = [tireWidthMaxMm, tireWidthMinMm]
  }
  if (tireWidthMinMm !== null) nextQuery[keys.nominalTireWidthMinMm] = String(tireWidthMinMm)
  if (tireWidthMaxMm !== null) nextQuery[keys.nominalTireWidthMaxMm] = String(tireWidthMaxMm)

  // Keep generating the legacy exact-value parameter only when no range
  // endpoint is set. New selector interactions use the range keys above.
  if (tireWidthMinMm === null && tireWidthMaxMm === null) {
    const tireWidthsMm = normalizedPositiveIntegers(state.nominalTireWidthsMm)
    if (tireWidthsMm.length > 0) nextQuery[keys.nominalTireWidthsMm] = tireWidthsMm.map(String)
  }

  const wheelSizeKeys = normalizedStrings(state.wheelSizeKeys)
  if (wheelSizeKeys.length > 0) nextQuery[keys.wheelSizeKeys] = wheelSizeKeys

  const beadSeatDiametersMm = normalizedPositiveIntegers(state.beadSeatDiametersMm)
  if (beadSeatDiametersMm.length > 0) {
    nextQuery[keys.beadSeatDiametersMm] = beadSeatDiametersMm.map(String)
  }

  if (state.minimumLoadKg !== null && Number.isFinite(state.minimumLoadKg) && state.minimumLoadKg > 0) {
    nextQuery[keys.minimumLoadKg] = String(state.minimumLoadKg)
  }

  const casingConstructions = normalizedStrings(state.casingConstructions)
  if (casingConstructions.length > 0) {
    nextQuery[keys.casingConstructions] = casingConstructions
  }
  if (state.radialOnly) nextQuery[keys.radialOnly] = '1'

  const beads = normalizedStrings(state.beads)
  if (beads.length > 0) nextQuery[keys.beads] = beads

  const seals = normalizedStrings(state.seals)
  if (seals.length > 0) nextQuery[keys.seals] = seals

  const eBikeRatings = normalizedEBikeRatings(state.eBikeRatings)
  if (eBikeRatings.length > 0) {
    nextQuery[keys.eBikeRatings] = eBikeRatings.map(rating => (
      rating === null ? SCHWALBE_E_BIKE_UNRATED_QUERY_VALUE : rating
    ))
  }

  const colors = normalizedStrings(state.colors)
  if (colors.length > 0) nextQuery[keys.colors] = colors

  const compounds = normalizedStrings(state.compounds)
  if (compounds.length > 0) nextQuery[keys.compounds] = compounds

  if (state.sortBy !== 'weight_asc') nextQuery[keys.sortBy] = state.sortBy
  return nextQuery
}
