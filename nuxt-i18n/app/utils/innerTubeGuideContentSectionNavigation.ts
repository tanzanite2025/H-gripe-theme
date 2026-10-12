import type { LocationQuery } from 'vue-router'
import { hasInnerTubeValveFitmentShareQueryParameters } from './innerTubeValveFitmentShareUrlState'

export type InnerTubeGuideContentPanel = 'introduction' | 'calculator'

export const innerTubeGuideContentSectionLinks = [
  { id: 'inner-tube-size', labelKey: 'size' },
  { id: 'inner-tube-valves', labelKey: 'valves' },
  { id: 'inner-tube-materials', labelKey: 'materials' },
  { id: 'inner-tube-installation-details', labelKey: 'installation' },
  { id: 'inner-tube-fitment-calculator', labelKey: 'calculator' },
  { id: 'inner-tube-guide-references', labelKey: 'references' },
] as const

export function resolveInnerTubeGuideContentPanelFromSectionHash(hash: string): InnerTubeGuideContentPanel | null {
  const sectionId = hash.replace(/^#/, '')
  if (sectionId === 'introduction' || sectionId === 'calculator') return sectionId
  if (['inner-tube-size', 'inner-tube-valves', 'inner-tube-materials', 'inner-tube-installation-details', 'inner-tube-introduction-panel'].includes(sectionId)) return 'introduction'
  if (['inner-tube-fitment-calculator', 'inner-tube-fitment-how-to', 'inner-tube-fitment-matrix', 'inner-tube-calculator-panel'].includes(sectionId)
    || /^inner-tube-fitment-how-to-step-[1-6]$/.test(sectionId)) return 'calculator'
  return null
}

export function resolveInnerTubeGuideContentPanelFromTabQuery(query: LocationQuery): InnerTubeGuideContentPanel | null {
  const requestedPanel = Array.isArray(query.tab) ? query.tab[0] : query.tab
  return requestedPanel === 'introduction' || requestedPanel === 'calculator' ? requestedPanel : null
}

export function resolveInnerTubeGuideContentPanelFromRouteLocation(
  query: LocationQuery, hash = '',
): InnerTubeGuideContentPanel {
  // An explicit section wins in the browser. SSR calls this without a fragment.
  return resolveInnerTubeGuideContentPanelFromSectionHash(hash)
    ?? resolveInnerTubeGuideContentPanelFromTabQuery(query)
    ?? (hasInnerTubeValveFitmentShareQueryParameters(query) ? 'calculator' : 'introduction')
}

export function resolveInnerTubeGuideContentSectionIdFromHash(hash: string): string {
  const sectionId = hash.replace(/^#/, '')
  return sectionId === 'introduction' || sectionId === 'calculator'
    ? `inner-tube-${sectionId}-panel`
    : sectionId
}

export function isInnerTubeGuideContentSectionHash(hash: string): boolean {
  return resolveInnerTubeGuideContentPanelFromSectionHash(hash) !== null
    || hash === '#inner-tube-guide' || hash === '#inner-tube-guide-references'
}
