import {
  SPOKE_CALCULATOR_OPTIONS,
  type Brand,
  type HubModel,
  type RimModel,
  type SpokeCatalog,
  type SpokeCatalogOptions,
  type WheelBuildActualLengths,
  type WheelBuildPreset,
} from '../data/spoke-calculator/database'

const normalizeOptions = (options?: Partial<SpokeCatalogOptions> | null): SpokeCatalogOptions => ({
  spokeCounts: Array.isArray(options?.spokeCounts) && options.spokeCounts.length
    ? options.spokeCounts
    : SPOKE_CALCULATOR_OPTIONS.spokeCounts,
  crossings: Array.isArray(options?.crossings) && options.crossings.length
    ? options.crossings
    : SPOKE_CALCULATOR_OPTIONS.crossings,
  nippleTypes: Array.isArray(options?.nippleTypes) && options.nippleTypes.length
    ? options.nippleTypes
    : SPOKE_CALCULATOR_OPTIONS.nippleTypes,
  wheelPositions: Array.isArray(options?.wheelPositions) && options.wheelPositions.length
    ? options.wheelPositions
    : SPOKE_CALCULATOR_OPTIONS.wheelPositions,
})

const normalizeBrands = <T>(value: unknown): Brand<T>[] => (
  Array.isArray(value)
    ? value.flatMap((brand) => {
      if (!brand || typeof brand !== 'object') return []
      const record = brand as Brand<T>
      if (!record.id || !record.name || !Array.isArray(record.items)) return []
      return [{
        id: String(record.id),
        name: String(record.name),
        items: record.items,
      }]
    })
    : []
)

const normalizeActualLengths = (value: unknown): WheelBuildActualLengths | null => {
  if (!value || typeof value !== 'object') return null
  const record = value as Partial<WheelBuildActualLengths>
  const actual: WheelBuildActualLengths = {
    frontLeft: typeof record.frontLeft === 'number' ? record.frontLeft : null,
    frontRight: typeof record.frontRight === 'number' ? record.frontRight : null,
    rearLeft: typeof record.rearLeft === 'number' ? record.rearLeft : null,
    rearRight: typeof record.rearRight === 'number' ? record.rearRight : null,
    notes: typeof record.notes === 'string' ? record.notes : '',
  }

  return actual.frontLeft == null &&
    actual.frontRight == null &&
    actual.rearLeft == null &&
    actual.rearRight == null &&
    !actual.notes
    ? null
    : actual
}

const normalizePresets = (value: unknown): WheelBuildPreset[] => (
  Array.isArray(value)
    ? value.flatMap((preset) => {
      if (!preset || typeof preset !== 'object') return []
      const record = preset as WheelBuildPreset
      if (!record.id || !record.name) return []
      return [{
        ...record,
        id: String(record.id),
        name: String(record.name),
        keywords: Array.isArray(record.keywords) ? record.keywords : [],
        wheelPosition: record.wheelPosition || 'auto',
        actualLengths: normalizeActualLengths(record.actualLengths),
      }]
    })
    : []
)

export const normalizeSpokeCatalogPayload = (payload: unknown): SpokeCatalog => {
  const record = payload && typeof payload === 'object' ? payload as Partial<SpokeCatalog> : {}

  return {
    options: normalizeOptions(record.options),
    rims: normalizeBrands<RimModel>(record.rims),
    hubs: normalizeBrands<HubModel>(record.hubs),
    presets: normalizePresets(record.presets),
  }
}
