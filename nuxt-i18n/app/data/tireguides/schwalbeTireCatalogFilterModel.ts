import type { SchwalbeTireCatalogItem } from '~/data/tireguides/schwalbeCatalog'
import { parseSchwalbeTireCatalogEtrtoDimensions } from '~/data/tireguides/schwalbeTireCatalogDimensionNormalization'

/** Official Version tokens verified as Schwalbe casing constructions. */
export const SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS = [
  'Super Race',
  'Super Ground',
  'Super Trail',
  'Super Downhill',
  'TRAIL',
  'TRAIL PRO',
  'GRAVITY',
  'GRAVITY PRO',
] as const

export type SchwalbeTireCasingConstructionLabel = typeof SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS[number]

/**
 * Filter dimensions are a selector-data contract, independent from the
 * `is_filterable` flag on the sales-product template. Raw official values are
 * preserved. A separate casing facet recognizes only reviewed, exact
 * comma-delimited Version tokens; it does not infer puncture protection levels.
 */
export interface SchwalbeTireCatalogFilterDimensions {
  modelName: string
  nominalTireWidthMm: number | null
  beadSeatDiameterMm: number | null
  wheelDiameterIn: string | null
  wheelSizeKey: string | null
  loadKg: number | null
  versionLabel: string | null
  casingConstructions: readonly SchwalbeTireCasingConstructionLabel[]
  isRadial: boolean
  compound: string | null
  color: string | null
  bead: string | null
  seal: string | null
  eBikeRating: string | null
}

export interface SchwalbeTireCatalogFilterCriteria {
  modelNames?: readonly string[]
  /** Inclusive nominal-width endpoints; null/undefined leaves that side open. */
  nominalTireWidthMinMm?: number | null
  nominalTireWidthMaxMm?: number | null
  /** @deprecated Exact-value selection retained for old links/consumers. */
  nominalTireWidthMm?: readonly number[]
  beadSeatDiameterMm?: readonly number[]
  wheelSizeKeys?: readonly string[]
  minimumLoadKg?: number | null
  versionLabels?: readonly string[]
  casingConstructions?: readonly string[]
  radialOnly?: boolean
  compounds?: readonly string[]
  colors?: readonly string[]
  beads?: readonly string[]
  seals?: readonly string[]
  /** Includes `null` for catalog rows with no listed E-Bike rating. */
  eBikeRatings?: readonly (string | null)[]
}

export interface SchwalbeTireCatalogFilterOption<Value extends string | number | null> {
  value: Value
}

export interface SchwalbeTireCatalogWheelSizeOption {
  value: string
  wheelDiameterIn: string
  beadSeatDiameterMm: number
}

export interface SchwalbeTireCatalogFilterOptions {
  modelNames: readonly SchwalbeTireCatalogFilterOption<string>[]
  nominalTireWidthsMm: readonly SchwalbeTireCatalogFilterOption<number>[]
  beadSeatDiametersMm: readonly SchwalbeTireCatalogFilterOption<number>[]
  wheelSizes: readonly SchwalbeTireCatalogWheelSizeOption[]
  versionLabels: readonly SchwalbeTireCatalogFilterOption<string>[]
  casingConstructions: readonly SchwalbeTireCatalogFilterOption<SchwalbeTireCasingConstructionLabel>[]
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

const SCHWALBE_INCH_DESIGNATION_PATTERN = /^([0-9]+(?:\.[0-9]+)?)\s*[x×]/i

const deriveWheelDiameterIn = (inchDesignation: string | undefined): string | null => {
  const match = SCHWALBE_INCH_DESIGNATION_PATTERN.exec(inchDesignation?.trim() || '')
  if (!match?.[1]) return null
  const numericValue = Number(match[1])
  return Number.isFinite(numericValue) && numericValue > 0
    ? String(numericValue)
    : null
}

const buildWheelSizeKey = (wheelDiameterIn: string | null, beadSeatDiameterMm: number | null): string | null => (
  wheelDiameterIn !== null && beadSeatDiameterMm !== null
    ? `${wheelDiameterIn}-${beadSeatDiameterMm}`
    : null
)

const deriveCasingConstructionTokens = (versionLabel: string | null): {
  casingConstructions: readonly SchwalbeTireCasingConstructionLabel[]
  isRadial: boolean
} => {
  const versionTokens = new Set(
    versionLabel?.split(',').map(token => token.trim()).filter(Boolean) || [],
  )

  return {
    casingConstructions: SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS
      .filter(label => versionTokens.has(label)),
    isRadial: versionTokens.has('Radial'),
  }
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
  const wheelDiameterIn = deriveWheelDiameterIn(item.inch_designation)
  const beadSeatDiameterMm = etrtoDimensions?.beadSeatDiameterMm ?? null
  const versionLabel = normalizedOptionalString(item.version_label)
  const casingDimensions = deriveCasingConstructionTokens(versionLabel)

  return {
    modelName: item.model_name.trim(),
    nominalTireWidthMm: etrtoDimensions?.nominalTireWidthMm ?? null,
    beadSeatDiameterMm,
    wheelDiameterIn,
    wheelSizeKey: buildWheelSizeKey(wheelDiameterIn, beadSeatDiameterMm),
    loadKg: Number.isFinite(item.load_kg) && (item.load_kg ?? 0) > 0 ? item.load_kg! : null,
    versionLabel,
    ...casingDimensions,
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

const matchesSelectedWheelSizes = (
  selectedValues: readonly string[] | undefined,
  actualValue: string | null,
): boolean => {
  const normalizedSelection = normalizedStringSelection(selectedValues)
  return normalizedSelection.length === 0 || (
    actualValue !== null && normalizedSelection.includes(actualValue)
  )
}

const matchesNominalTireWidthRange = (
  minimumMm: number | null | undefined,
  maximumMm: number | null | undefined,
  actualMm: number | null,
): boolean => {
  const hasMinimum = minimumMm !== null && minimumMm !== undefined
  const hasMaximum = maximumMm !== null && maximumMm !== undefined
  if (!hasMinimum && !hasMaximum) return true
  if (actualMm === null || !Number.isFinite(actualMm)) return false
  if (hasMinimum && (!Number.isFinite(minimumMm) || minimumMm! <= 0 || actualMm < minimumMm!)) return false
  if (hasMaximum && (!Number.isFinite(maximumMm) || maximumMm! <= 0 || actualMm > maximumMm!)) return false
  return true
}

const matchesMinimumLoadKg = (minimumLoadKg: number | null | undefined, actualLoadKg: number | null): boolean => {
  if (minimumLoadKg === undefined || minimumLoadKg === null) return true
  return Number.isFinite(minimumLoadKg)
    && minimumLoadKg > 0
    && actualLoadKg !== null
    && actualLoadKg >= minimumLoadKg
}

const matchesSelectedStringTokens = (
  selectedValues: readonly string[] | undefined,
  actualValues: readonly string[],
): boolean => {
  const normalizedSelection = normalizedStringSelection(selectedValues)
  return normalizedSelection.length === 0 || (
    actualValues.some(actualValue => normalizedSelection.includes(actualValue))
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
  const hasNominalTireWidthRange = (
    (criteria.nominalTireWidthMinMm !== undefined && criteria.nominalTireWidthMinMm !== null)
    || (criteria.nominalTireWidthMaxMm !== undefined && criteria.nominalTireWidthMaxMm !== null)
  )

  return (
    matchesSelectedStrings(criteria.modelNames, dimensions.modelName)
    && (hasNominalTireWidthRange
      ? matchesNominalTireWidthRange(
        criteria.nominalTireWidthMinMm,
        criteria.nominalTireWidthMaxMm,
        dimensions.nominalTireWidthMm,
      )
      : matchesSelectedNumbers(criteria.nominalTireWidthMm, dimensions.nominalTireWidthMm))
    && matchesSelectedNumbers(criteria.beadSeatDiameterMm, dimensions.beadSeatDiameterMm)
    && matchesSelectedWheelSizes(criteria.wheelSizeKeys, dimensions.wheelSizeKey)
    && matchesMinimumLoadKg(criteria.minimumLoadKg, dimensions.loadKg)
    && matchesSelectedStrings(criteria.versionLabels, dimensions.versionLabel)
    && matchesSelectedStringTokens(criteria.casingConstructions, dimensions.casingConstructions)
    && (!criteria.radialOnly || dimensions.isRadial)
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

const buildWheelSizeFilterOptions = (
  dimensions: readonly SchwalbeTireCatalogFilterDimensions[],
): readonly SchwalbeTireCatalogWheelSizeOption[] => {
  const options = new Map<string, SchwalbeTireCatalogWheelSizeOption>()
  dimensions.forEach((value) => {
    if (value.wheelSizeKey === null || value.wheelDiameterIn === null || value.beadSeatDiameterMm === null) return
    options.set(value.wheelSizeKey, {
      value: value.wheelSizeKey,
      wheelDiameterIn: value.wheelDiameterIn,
      beadSeatDiameterMm: value.beadSeatDiameterMm,
    })
  })
  return [...options.values()].sort((left, right) => {
    const diameterDifference = Number(left.wheelDiameterIn) - Number(right.wheelDiameterIn)
    if (diameterDifference !== 0) return diameterDifference
    if (left.beadSeatDiameterMm !== right.beadSeatDiameterMm) {
      return left.beadSeatDiameterMm - right.beadSeatDiameterMm
    }
    return left.value.localeCompare(right.value)
  })
}

/**
 * Builds stable option lists from the currently loaded catalog snapshot.
 * Missing values are omitted from ordinary dimensions. The official blank
 * E-Bike value is exposed as `null` so rows without a listed rating can be
 * filtered separately. The blank does not establish an E-Bike compatibility
 * or category conclusion.
 */
export const buildSchwalbeTireCatalogFilterOptions = (
  items: readonly SchwalbeTireCatalogItem[],
): SchwalbeTireCatalogFilterOptions => {
  const dimensions = items.map(deriveSchwalbeTireCatalogFilterDimensions)

  return {
    modelNames: buildStringFilterOptions(dimensions, value => value.modelName || null),
    nominalTireWidthsMm: buildNumberFilterOptions(dimensions, value => value.nominalTireWidthMm),
    beadSeatDiametersMm: buildNumberFilterOptions(dimensions, value => value.beadSeatDiameterMm),
    wheelSizes: buildWheelSizeFilterOptions(dimensions),
    versionLabels: buildStringFilterOptions(dimensions, value => value.versionLabel),
    casingConstructions: SCHWALBE_TIRE_CASING_CONSTRUCTION_LABELS
      .filter(label => dimensions.some(value => value.casingConstructions.includes(label)))
      .map(value => ({ value })),
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
