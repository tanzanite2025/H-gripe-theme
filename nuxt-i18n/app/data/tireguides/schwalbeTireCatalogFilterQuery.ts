export type SchwalbeCatalogSort = 'model' | 'etrto' | 'article'

export interface SchwalbeTireCatalogFilterQueryState {
  modelName: string | null
  nominalTireWidthsMm: number[]
  beadSeatDiametersMm: number[]
  sortBy: SchwalbeCatalogSort
}

export type SchwalbeTireCatalogFilterQuery = Record<string, unknown>

export const SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS = {
  modelName: 'model',
  nominalTireWidthsMm: 'tire_width_mm',
  beadSeatDiametersMm: 'bead_seat_diameter_mm',
  sortBy: 'sort',
} as const

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
  sortBy: readSort(query[SCHWALBE_TIRE_CATALOG_FILTER_QUERY_KEYS.sortBy]),
})

const normalizedPositiveIntegers = (values: readonly number[]): number[] => (
  [...new Set(values.filter(value => Number.isSafeInteger(value) && value > 0))]
    .sort((left, right) => left - right)
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
  delete nextQuery[keys.sortBy]

  if (state.modelName?.trim()) nextQuery[keys.modelName] = state.modelName.trim()

  const tireWidthsMm = normalizedPositiveIntegers(state.nominalTireWidthsMm)
  if (tireWidthsMm.length > 0) nextQuery[keys.nominalTireWidthsMm] = tireWidthsMm.map(String)

  const beadSeatDiametersMm = normalizedPositiveIntegers(state.beadSeatDiametersMm)
  if (beadSeatDiametersMm.length > 0) {
    nextQuery[keys.beadSeatDiametersMm] = beadSeatDiametersMm.map(String)
  }

  if (state.sortBy !== 'model') nextQuery[keys.sortBy] = state.sortBy
  return nextQuery
}
