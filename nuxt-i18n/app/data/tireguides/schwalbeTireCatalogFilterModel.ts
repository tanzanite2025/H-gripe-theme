import type { SchwalbeTireCatalogItem } from '~/data/tireguides/schwalbeCatalog'
import { parseSchwalbeTireCatalogEtrtoDimensions } from '~/data/tireguides/schwalbeTireCatalogDimensionNormalization'

/**
 * Filter dimensions are a selector-data contract, independent from the
 * `is_filterable` flag on the sales-product template. Values remain the
 * official catalog strings; this model does not translate, split, or infer
 * protection levels from compound or version labels.
 */
export interface SchwalbeTireCatalogFilterDimensions {
  modelName: string
  nominalTireWidthMm: number | null
  beadSeatDiameterMm: number | null
  versionLabel: string | null
  compound: string | null
  color: string | null
  bead: string | null
  seal: string | null
  eBikeRating: string | null
}

export interface SchwalbeTireCatalogFilterCriteria {
  modelNames?: readonly string[]
  nominalTireWidthMm?: readonly number[]
  beadSeatDiameterMm?: readonly number[]
  versionLabels?: readonly string[]
  compounds?: readonly string[]
  colors?: readonly string[]
  beads?: readonly string[]
  seals?: readonly string[]
  /** Includes `null` for the official blank value (ordinary bicycle models). */
  eBikeRatings?: readonly (string | null)[]
}

export interface SchwalbeTireCatalogFilterOption<Value extends string | number | null> {
  value: Value
}

export interface SchwalbeTireCatalogFilterOptions {
  modelNames: readonly SchwalbeTireCatalogFilterOption<string>[]
  nominalTireWidthsMm: readonly SchwalbeTireCatalogFilterOption<number>[]
  beadSeatDiametersMm: readonly SchwalbeTireCatalogFilterOption<number>[]
  versionLabels: readonly SchwalbeTireCatalogFilterOption<string>[]
  compounds: readonly SchwalbeTireCatalogFilterOption<string>[]
  colors: readonly SchwalbeTireCatalogFilterOption<string>[]
  beads: readonly SchwalbeTireCatalogFilterOption<string>[]
  seals: readonly SchwalbeTireCatalogFilterOption<string>[]
  eBikeRatings: readonly SchwalbeTireCatalogFilterOption<string | null>[]
}

/**
 * API-neutral form of migration 362's official possible-combination rule.
 * Inclusive ranges describe width guidance; they do not certify a specific
 * tire model or replace a frame-clearance and rim-maker check.
 */
export interface SchwalbeTireRimWidthCombinationRule {
  tireWidthMinMm: number
  tireWidthMaxMm: number
  innerRimWidthMinMm: number
  innerRimWidthMaxMm: number
}

const normalizedOptionalString = (value: string | undefined): string | null => {
  const normalized = value?.trim() || ''
  return normalized || null
}

const normalizedStringSelection = (values: readonly string[] | undefined): readonly string[] => (
  values
    ?.map(value => value.trim())
    .filter(Boolean) || []
)

const normalizedNumberSelection = (values: readonly number[] | undefined): readonly number[] => (
  values?.filter(value => Number.isFinite(value)) || []
)

/**
 * Converts one official catalog row into the dimensions that the selector can
 * safely expose. The two numeric ETRTO dimensions are derived only when the
 * complete value follows the strict `width-diameter` format.
 */
export const deriveSchwalbeTireCatalogFilterDimensions = (
  item: SchwalbeTireCatalogItem,
): SchwalbeTireCatalogFilterDimensions => {
  const etrtoDimensions = parseSchwalbeTireCatalogEtrtoDimensions(item.etrto)

  return {
    modelName: item.model_name.trim(),
    nominalTireWidthMm: etrtoDimensions?.nominalTireWidthMm ?? null,
    beadSeatDiameterMm: etrtoDimensions?.beadSeatDiameterMm ?? null,
    versionLabel: normalizedOptionalString(item.version_label),
    compound: normalizedOptionalString(item.compound),
    color: normalizedOptionalString(item.color),
    bead: normalizedOptionalString(item.bead),
    seal: normalizedOptionalString(item.seal),
    eBikeRating: normalizedOptionalString(item.e_bike_rating),
  }
}

const matchesSelectedStrings = (
  selectedValues: readonly string[] | undefined,
  actualValue: string | null,
): boolean => {
  const normalizedSelection = normalizedStringSelection(selectedValues)
  return normalizedSelection.length === 0 || (
    actualValue !== null && normalizedSelection.includes(actualValue)
  )
}

const matchesSelectedNumbers = (
  selectedValues: readonly number[] | undefined,
  actualValue: number | null,
): boolean => {
  const normalizedSelection = normalizedNumberSelection(selectedValues)
  return normalizedSelection.length === 0 || (
    actualValue !== null && normalizedSelection.includes(actualValue)
  )
}

const matchesSelectedEBikeRatings = (
  selectedRatings: readonly (string | null)[] | undefined,
  actualRating: string | null,
): boolean => {
  const normalizedSelection = selectedRatings
    ?.map(value => value === null ? null : value.trim())
    .filter(value => value === null || Boolean(value)) || []
  if (normalizedSelection.length === 0) return true

  return normalizedSelection.includes(actualRating)
}

/**
 * Applies the selector contract without sorting or paginating the result.
 * Multiple values inside one dimension use OR; selected dimensions combine as
 * AND. Empty selections mean that the dimension is not constrained.
 */
export const filterSchwalbeTireCatalogItems = (
  items: readonly SchwalbeTireCatalogItem[],
  criteria: SchwalbeTireCatalogFilterCriteria = {},
): SchwalbeTireCatalogItem[] => items.filter((item) => {
  const dimensions = deriveSchwalbeTireCatalogFilterDimensions(item)

  return (
    matchesSelectedStrings(criteria.modelNames, dimensions.modelName)
    && matchesSelectedNumbers(criteria.nominalTireWidthMm, dimensions.nominalTireWidthMm)
    && matchesSelectedNumbers(criteria.beadSeatDiameterMm, dimensions.beadSeatDiameterMm)
    && matchesSelectedStrings(criteria.versionLabels, dimensions.versionLabel)
    && matchesSelectedStrings(criteria.compounds, dimensions.compound)
    && matchesSelectedStrings(criteria.colors, dimensions.color)
    && matchesSelectedStrings(criteria.beads, dimensions.bead)
    && matchesSelectedStrings(criteria.seals, dimensions.seal)
    && matchesSelectedEBikeRatings(criteria.eBikeRatings, dimensions.eBikeRating)
  )
})

/**
 * Tests the official width-range guidance for one catalog item and one rim
 * inner width. A missing or invalid ETRTO/rim value is never treated as a
 * match. The caller must provide rules loaded from the migration-362 table.
 */
export const isSchwalbeTireCatalogItemCompatibleWithInnerRimWidth = (
  item: SchwalbeTireCatalogItem,
  innerRimWidthMm: number,
  rules: readonly SchwalbeTireRimWidthCombinationRule[],
): boolean => {
  if (!Number.isFinite(innerRimWidthMm) || innerRimWidthMm <= 0) return false

  const nominalTireWidthMm = deriveSchwalbeTireCatalogFilterDimensions(item).nominalTireWidthMm
  if (nominalTireWidthMm === null) return false

  return rules.some(rule => (
    Number.isFinite(rule.tireWidthMinMm)
    && Number.isFinite(rule.tireWidthMaxMm)
    && Number.isFinite(rule.innerRimWidthMinMm)
    && Number.isFinite(rule.innerRimWidthMaxMm)
    && rule.tireWidthMinMm <= nominalTireWidthMm
    && nominalTireWidthMm <= rule.tireWidthMaxMm
    && rule.innerRimWidthMinMm <= innerRimWidthMm
    && innerRimWidthMm <= rule.innerRimWidthMaxMm
  ))
}

/**
 * Applies the migration-362 possible-combination guidance as a separate
 * compatibility pass, so it can be reused by a selector or product modal.
 */
export const filterSchwalbeTireCatalogItemsByInnerRimWidth = (
  items: readonly SchwalbeTireCatalogItem[],
  innerRimWidthMm: number,
  rules: readonly SchwalbeTireRimWidthCombinationRule[],
): SchwalbeTireCatalogItem[] => items.filter(item => (
  isSchwalbeTireCatalogItemCompatibleWithInnerRimWidth(item, innerRimWidthMm, rules)
))

const buildStringFilterOptions = (
  dimensions: readonly SchwalbeTireCatalogFilterDimensions[],
  readValue: (dimensions: SchwalbeTireCatalogFilterDimensions) => string | null,
): readonly SchwalbeTireCatalogFilterOption<string>[] => (
  [...new Set(dimensions.map(readValue).filter((value): value is string => value !== null))]
    .sort((left, right) => left.localeCompare(right))
    .map(value => ({ value }))
)

const buildNumberFilterOptions = (
  dimensions: readonly SchwalbeTireCatalogFilterDimensions[],
  readValue: (dimensions: SchwalbeTireCatalogFilterDimensions) => number | null,
): readonly SchwalbeTireCatalogFilterOption<number>[] => (
  [...new Set(dimensions.map(readValue).filter((value): value is number => value !== null))]
    .sort((left, right) => left - right)
    .map(value => ({ value }))
)

/**
 * Builds stable option lists from the currently loaded catalog snapshot.
 * Missing values are omitted from ordinary dimensions. The official blank
 * E-Bike value is exposed as `null` because it represents ordinary bicycle
 * models in this snapshot, not an unknown value.
 */
export const buildSchwalbeTireCatalogFilterOptions = (
  items: readonly SchwalbeTireCatalogItem[],
): SchwalbeTireCatalogFilterOptions => {
  const dimensions = items.map(deriveSchwalbeTireCatalogFilterDimensions)

  return {
    modelNames: buildStringFilterOptions(dimensions, value => value.modelName || null),
    nominalTireWidthsMm: buildNumberFilterOptions(dimensions, value => value.nominalTireWidthMm),
    beadSeatDiametersMm: buildNumberFilterOptions(dimensions, value => value.beadSeatDiameterMm),
    versionLabels: buildStringFilterOptions(dimensions, value => value.versionLabel),
    compounds: buildStringFilterOptions(dimensions, value => value.compound),
    colors: buildStringFilterOptions(dimensions, value => value.color),
    beads: buildStringFilterOptions(dimensions, value => value.bead),
    seals: buildStringFilterOptions(dimensions, value => value.seal),
    eBikeRatings: [
      ...buildStringFilterOptions(dimensions, value => value.eBikeRating),
      ...(dimensions.some(value => value.eBikeRating === null)
        ? [{ value: null }]
        : []),
    ].sort((left, right) => {
      if (left.value === null) return 1
      if (right.value === null) return -1
      return left.value.localeCompare(right.value)
    }),
  }
}
