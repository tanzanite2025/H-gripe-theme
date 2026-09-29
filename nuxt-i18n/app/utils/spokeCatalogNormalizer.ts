import {
  SPOKE_CALCULATOR_OPTIONS,
  type Brand,
  type HubModel,
  type RimModel,
  type SpokeCatalog,
  type SpokeCatalogOptions,
  type SpokeRecordedActualLengths,
  type SpokeRecordedResult,
  type SpokeRecordedResultsResponse,
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

const normalizePublicRims = (value: unknown): Brand<RimModel>[] => (
  normalizeBrands<unknown>(value).map(brand => ({
    ...brand,
    items: brand.items.flatMap(item => {
      if (!item || typeof item !== 'object') return []
      const record = item as Partial<RimModel>
      if (!record.id || !record.name) return []
      // ERD/weight are deliberately discarded even if a misconfigured API
      // accidentally includes them in its public response.
      return [{ id: String(record.id), name: String(record.name), erd: null }]
    }),
  }))
)

const normalizePublicHubs = (value: unknown): Brand<HubModel>[] => (
  normalizeBrands<unknown>(value).map(brand => ({
    ...brand,
    items: brand.items.flatMap(item => {
      if (!item || typeof item !== 'object') return []
      const record = item as Partial<HubModel>
      if (!record.id || !record.name) return []
      return [{ id: String(record.id), name: String(record.name) }]
    }),
  }))
)

const normalizeNumber = (value: unknown): number | null => (
  typeof value === 'number' && Number.isFinite(value) ? value : null
)

const normalizeRecordedLengths = (value: unknown): SpokeRecordedActualLengths | null => {
  if (!value || typeof value !== 'object') return null
  const record = value as Partial<SpokeRecordedActualLengths>
  const actual: SpokeRecordedActualLengths = {
    frontLeft: normalizeNumber(record.frontLeft),
    frontRight: normalizeNumber(record.frontRight),
    rearLeft: normalizeNumber(record.rearLeft),
    rearRight: normalizeNumber(record.rearRight),
  }

  return actual.frontLeft == null &&
    actual.frontRight == null &&
    actual.rearLeft == null &&
    actual.rearRight == null
    ? null
    : actual
}

const normalizePresets = (value: unknown): WheelBuildPreset[] => (
  Array.isArray(value)
    ? value.flatMap((preset) => {
      if (!preset || typeof preset !== 'object') return []
      const record = preset as Partial<WheelBuildPreset>
      if (!record.id || !record.name) return []
      return [{
        id: String(record.id),
        name: String(record.name),
        keywords: Array.isArray(record.keywords)
          ? record.keywords.filter((keyword): keyword is string => typeof keyword === 'string')
          : [],
        description: typeof record.description === 'string' ? record.description : undefined,
        rimBrandId: typeof record.rimBrandId === 'string' ? record.rimBrandId : '',
        rimModelId: typeof record.rimModelId === 'string' ? record.rimModelId : '',
        hubBrandId: typeof record.hubBrandId === 'string' ? record.hubBrandId : '',
        hubModelId: typeof record.hubModelId === 'string' ? record.hubModelId : '',
        spokeCount: typeof record.spokeCount === 'number' ? record.spokeCount : 0,
        crossing: typeof record.crossing === 'number' ? record.crossing : 0,
        nippleType: record.nippleType === 'hidden' ? 'hidden' : 'standard',
        nippleLength: null,
        wheelPosition: record.wheelPosition === 'front' || record.wheelPosition === 'rear'
          ? record.wheelPosition
          : 'auto',
      }]
    })
    : []
)

const normalizeRecordedResults = (value: unknown): SpokeRecordedResult[] => (
  Array.isArray(value)
    ? value.flatMap((preset) => {
      if (!preset || typeof preset !== 'object') return []
      const record = preset as Partial<SpokeRecordedResult>
      const actualLengths = normalizeRecordedLengths(record.actualLengths)
      if (!record.id || !record.name || !actualLengths) return []

      return [{
        id: String(record.id),
        name: String(record.name),
        keywords: Array.isArray(record.keywords)
          ? record.keywords.filter((keyword): keyword is string => typeof keyword === 'string')
          : [],
        description: typeof record.description === 'string' ? record.description : undefined,
        rimBrandId: typeof record.rimBrandId === 'string' ? record.rimBrandId : '',
        rimModelId: typeof record.rimModelId === 'string' ? record.rimModelId : '',
        hubBrandId: typeof record.hubBrandId === 'string' ? record.hubBrandId : '',
        hubModelId: typeof record.hubModelId === 'string' ? record.hubModelId : '',
        spokeCount: typeof record.spokeCount === 'number' ? record.spokeCount : 0,
        crossing: typeof record.crossing === 'number' ? record.crossing : 0,
        nippleType: record.nippleType === 'hidden' ? 'hidden' : 'standard',
        wheelPosition: record.wheelPosition === 'front' || record.wheelPosition === 'rear'
          ? record.wheelPosition
          : 'auto',
        actualLengths,
      }]
    })
    : []
)

export const normalizeSpokeCatalogPayload = (payload: unknown): SpokeCatalog => {
  const record = payload && typeof payload === 'object' ? payload as Partial<SpokeCatalog> : {}

  return {
    options: normalizeOptions(record.options),
    rims: normalizePublicRims(record.rims),
    hubs: normalizePublicHubs(record.hubs),
    presets: normalizePresets(record.presets),
  }
}

export const normalizeSpokeRecordedResultsPayload = (payload: unknown): SpokeRecordedResultsResponse => {
  const record = payload && typeof payload === 'object'
    ? payload as Partial<SpokeRecordedResultsResponse>
    : {}

  return {
    presets: normalizeRecordedResults(record.presets),
  }
}
