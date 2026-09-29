export type SchwalbeCatalogSort = 'model' | 'etrto' | 'article'

export interface SchwalbeTireCatalogFilterQueryState {
  modelName: string | null
  nominalTireWidthsMm: number[]
  beadSeatDiametersMm: number[]
  beads: string[]
  seals: string[]
  eBikeRatings: (string | null)[]
  sortBy: SchwalbeCatalogSort
}

export type SchwalbeTireCatalogFilterQuery = Record<string, unknown>

export const SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS = {
  modelName: 'model',
  nominalTireWidthsMm: 'tire_width_mm',
  beadSeatDiametersMm: 'bead_seat_diameter_mm',
  beads: 'bead',
  seals: 'seal',
  eBikeRatings: 'e_bike_rating',
  sortBy: 'sort',
} as const

/** Reserved URL value for a catalog row with no published E-Bike rating. */
export const SCHWALBE_E_BIKE_UNRATED_QUERY_VALUE = 'none'

const SCHWALBE_CATALOG_SORT_VALUES: readonly SchwalbeCatalogSort[] = [
  'model',
  'etrto',
  'article',
]

const readFirstString = (value: unknown): string | null => {
  const candidate = Array.isArray(value) ? value[0] : value
  if (typeof candidate !== 'string') return null
  return candidate.trim() || null
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

const readStringValues = (value: unknown): string[] => {
  const candidates = Array.isArray(value) ? value : [value]
  return [...new Set(candidates.flatMap((candidate) => {
    if (typeof candidate !== 'string') return []
    const normalized = candidate.trim()
    return normalized ? [normalized] : []
  }))].sort((left, right) => left.localeCompare(right))
}

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
    : 'model'
}

/** Reads only selector state; search, page, and unrelated query keys stay separate. */
export const parseSchwalbeTireCatalogFilterQuery = (
  query: SchwalbeTireCatalogFilterQuery,
): SchwalbeTireCatalogFilterQueryState => ({
  modelName: readFirstString(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.modelName]),
  nominalTireWidthsMm: readPositiveIntegerValues(
    query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.nominalTireWidthsMm],
  ),
  beadSeatDiametersMm: readPositiveIntegerValues(
    query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.beadSeatDiametersMm],
  ),
  beads: readStringValues(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.beads]),
  seals: readStringValues(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.seals]),
  eBikeRatings: readEBikeRatingValues(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.eBikeRatings]),
  sortBy: readSort(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.sortBy]),
})

const normalizedPositiveIntegers = (values: readonly number[]): number[] => (
  [...new Set(values.filter(value => Number.isSafeInteger(value) && value > 0))]
    .sort((left, right) => left - right)
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
  delete nextQuery[keys.nominalTireWidthsMm]
  delete nextQuery[keys.beadSeatDiametersMm]
  delete nextQuery[keys.beads]
  delete nextQuery[keys.seals]
  delete nextQuery[keys.eBikeRatings]
  delete nextQuery[keys.sortBy]

  if (state.modelName?.trim()) nextQuery[keys.modelName] = state.modelName.trim()

  const tireWidthsMm = normalizedPositiveIntegers(state.nominalTireWidthsMm)
  if (tireWidthsMm.length > 0) nextQuery[keys.nominalTireWidthsMm] = tireWidthsMm.map(String)

  const beadSeatDiametersMm = normalizedPositiveIntegers(state.beadSeatDiametersMm)
  if (beadSeatDiametersMm.length > 0) {
    nextQuery[keys.beadSeatDiametersMm] = beadSeatDiametersMm.map(String)
  }

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

  if (state.sortBy !== 'model') nextQuery[keys.sortBy] = state.sortBy
  return nextQuery
}
