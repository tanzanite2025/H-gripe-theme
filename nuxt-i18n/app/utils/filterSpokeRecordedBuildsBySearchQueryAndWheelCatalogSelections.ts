import type { SpokeCatalog, SpokeRecordedResult } from '../data/spoke-calculator/database'
import type { SpokeWheelCatalogSelection } from '../types/spokeCalculator'

interface SpokeRecordedBuildWheelCatalogSelections {
  front?: SpokeWheelCatalogSelection | null
  rear?: SpokeWheelCatalogSelection | null
}

function normalizeSpokeRecordedBuildSearchText(value: string): string {
  return value.normalize('NFKC').toLowerCase().replace(/[^\p{L}\p{N}]+/gu, ' ').trim()
}

function hasSpokeRecordedBuildWheelCatalogSelection(selection?: SpokeWheelCatalogSelection | null): boolean {
  return Boolean(selection && (
    selection.rimBrandId || selection.rimModelId || selection.hubBrandId || selection.hubModelId
  ))
}

function matchesSpokeRecordedBuildWheelCatalogSelection(
  preset: SpokeRecordedResult,
  selection?: SpokeWheelCatalogSelection | null,
): boolean {
  return !selection || (
    (!selection.rimBrandId || preset.rimBrandId === selection.rimBrandId)
    && (!selection.rimModelId || preset.rimModelId === selection.rimModelId)
    && (!selection.hubBrandId || preset.hubBrandId === selection.hubBrandId)
    && (!selection.hubModelId || preset.hubModelId === selection.hubModelId)
  )
}

export function filterSpokeRecordedBuildsBySearchQueryAndWheelCatalogSelections(
  presets: readonly SpokeRecordedResult[],
  query: string,
  selections: SpokeRecordedBuildWheelCatalogSelections,
  catalog: Pick<SpokeCatalog, 'rims' | 'hubs'>,
): SpokeRecordedResult[] {
  const trimmedQuery = query.trim()
  const activeSelections = [selections.front, selections.rear].filter(hasSpokeRecordedBuildWheelCatalogSelection)
  if (trimmedQuery.length < 2 && (trimmedQuery.length > 0 || activeSelections.length === 0)) return []

  const searchTerms = normalizeSpokeRecordedBuildSearchText(trimmedQuery).split(/\s+/).filter(Boolean)
  if (trimmedQuery && searchTerms.length === 0) return []

  return presets.filter(preset => {
    if (!Object.values(preset.actualLengths).some(value => value != null)) return false

    if (preset.wheelPosition === 'front' || preset.wheelPosition === 'rear') {
      if (!matchesSpokeRecordedBuildWheelCatalogSelection(preset, selections[preset.wheelPosition])) return false
    } else if (activeSelections.length && !activeSelections.some(selection => (
      matchesSpokeRecordedBuildWheelCatalogSelection(preset, selection)
    ))) {
      return false
    }

    const rimBrand = catalog.rims.find(brand => brand.id === preset.rimBrandId)
    const hubBrand = catalog.hubs.find(brand => brand.id === preset.hubBrandId)
    const searchableText = normalizeSpokeRecordedBuildSearchText([
      preset.name,
      preset.description,
      ...preset.keywords,
      rimBrand?.name,
      rimBrand?.items.find(rim => rim.id === preset.rimModelId)?.name,
      hubBrand?.name,
      hubBrand?.items.find(hub => hub.id === preset.hubModelId)?.name,
    ].filter(Boolean).join(' '))

    return searchTerms.every(term => searchableText.includes(term))
  })
}
