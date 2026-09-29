/**
 * Dimensions that can be derived without adding new facts to the catalog.
 *
 * Schwalbe's ETRTO value is the source of truth for these two dimensions:
 * `nominalTireWidthMm-beadSeatDiameterMm` (for example, `50-622`).
 * The parser is intentionally strict so malformed or future non-standard
 * values do not silently become filterable dimensions.
 */
export interface SchwalbeTireCatalogDerivedDimensionFields {
  nominalTireWidthMm: number
  beadSeatDiameterMm: number
}

const SCHWALBE_ETRTO_DIMENSION_PATTERN = /^(\d+)-(\d+)$/

export const parseSchwalbeTireCatalogEtrtoDimensions = (
  etrto: string,
): SchwalbeTireCatalogDerivedDimensionFields | null => {
  const normalizedEtrto = etrto.trim()
  const match = SCHWALBE_ETRTO_DIMENSION_PATTERN.exec(normalizedEtrto)
  if (!match) return null

  const nominalTireWidthMm = Number(match[1])
  const beadSeatDiameterMm = Number(match[2])
  if (!Number.isSafeInteger(nominalTireWidthMm) || !Number.isSafeInteger(beadSeatDiameterMm)) {
    return null
  }

  return {
    nominalTireWidthMm,
    beadSeatDiameterMm,
  }
}
