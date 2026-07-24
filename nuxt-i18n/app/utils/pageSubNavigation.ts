import type { PrimaryMegaNavCard, PrimaryMegaNavSection } from '~/utils/primaryMegaNav'
import {
  normalizePrimaryMegaNavPath,
  primaryMegaNavPathMatches,
} from '~/utils/primaryMegaNav'

export interface PageSubNavigationTab {
  id: string
  label?: string
  labelKey?: string
  fallback?: string
  to?: string
}

export interface PageSubNavigationEntry {
  /**
   * Canonical page route that owns the tabs.
   * Hashes are generated from tab ids unless a tab provides its own `to`.
   */
  path: string
  tabs: readonly PageSubNavigationTab[]
}

export type PageSubNavigationChild = PageSubNavigationTab & {
  to: string
}

export const tireGuideTabs = [
  { id: 'size', label: 'Tire size' },
  { id: 'match', label: 'Match' },
  { id: 'tubeless', label: 'Tubeless tires' },
  { id: 'installation', label: 'Installation' },
  { id: 'choose', label: 'How to choose' },
  { id: 'rims', label: 'Tire pressure' },
  { id: 'tube', label: 'Inner tube' },
] as const satisfies readonly PageSubNavigationTab[]

export type TireGuideTabId = (typeof tireGuideTabs)[number]['id']

export const wheelsetBuyerTabs = [
  { id: 'safety-instructions', label: 'Safety instructions' },
  { id: 'sample-assembly', label: 'Sample assembly' },
  { id: 'special-order', label: 'Special order' },
  { id: 'appearance-logo', label: 'Appearance Logo' },
  { id: 'choose-freehub', label: 'Choose freehub' },
  { id: 'wheel-components', label: 'Wheel Components' },
  { id: 'optional', label: 'Optional' },
] as const satisfies readonly PageSubNavigationTab[]

export type WheelsetBuyerTabId = (typeof wheelsetBuyerTabs)[number]['id']

export const companyAboutTabs = [
  { id: 'factory', label: 'Factory' },
  { id: 'appearance', label: 'Appearance' },
  { id: 'hole-patterns', label: 'Hole Patterns' },
  { id: 'facility', label: 'Facility' },
  { id: 'manufacture', label: 'Manufacture' },
  { id: 'qualitycontrol', label: 'Quality control' },
] as const satisfies readonly PageSubNavigationTab[]

export type CompanyAboutTabId = (typeof companyAboutTabs)[number]['id']

export const warrantyTabs = [
  { id: 'change-cancel', label: 'Change / Cancel' },
  { id: 'damaged-lost', label: 'Damaged or Lost Goods' },
  { id: 'returns', label: 'Returns' },
  { id: 'warranty', label: 'Warranty' },
  { id: 'accidental-damage', label: 'Accidental Damage' },
  { id: 'protection', label: 'Protection' },
  { id: 'submit-warranty', label: 'Submit Warranty' },
] as const satisfies readonly PageSubNavigationTab[]

export type WarrantyTabId = (typeof warrantyTabs)[number]['id']

export const testReportTabs = [
  { id: 'rim-test-report', label: 'Rim Test Report' },
  { id: 'wheelset-test-report', label: 'Wheelset Test Report' },
  { id: 'tension', label: 'Tension' },
  { id: 'wheelset-assembly', label: 'Wheelset Assembly' },
] as const satisfies readonly PageSubNavigationTab[]

export type TestReportTabId = (typeof testReportTabs)[number]['id']

export const spokeCalculatorTabs = [
  { id: 'calculator', labelKey: 'spokeCalculator.tabs.calculator', fallback: 'Calculator' },
  { id: 'parameter', labelKey: 'spokeCalculator.tabs.parameter', fallback: 'Parameter' },
] as const satisfies readonly PageSubNavigationTab[]

export type SpokeCalculatorTabId = (typeof spokeCalculatorTabs)[number]['id']

export const membershipAndPointsTabs = [
  { id: 'myinfo', labelKey: 'member.tabs.myInfo', fallback: 'My info' },
  { id: 'levers', labelKey: 'member.tabs.levers', fallback: 'Levers' },
  { id: 'exchange', labelKey: 'member.tabs.exchange', fallback: 'Exchange' },
] as const satisfies readonly PageSubNavigationTab[]

export type MembershipTabId = (typeof membershipAndPointsTabs)[number]['id']

export const pictureWarehouseTabs = [
  { id: 'riders', label: 'Riders photos' },
  { id: 'brand', label: 'Tanzanite photos' },
] as const satisfies readonly PageSubNavigationTab[]

export type PictureWarehouseTabId = (typeof pictureWarehouseTabs)[number]['id']

/**
 * Third-level page navigation registry.
 *
 * Header/mobile mega menus derive card children from this list by matching
 * `entry.path` to a second-level card `to`. When a tabbed page gains another
 * tab, update the page by importing the matching exported tab array here; the
 * mega menu will pick it up automatically without editing menu components.
 */
export const pageSubNavigationEntries = [
  { path: '/guides/tireguides', tabs: tireGuideTabs },
  { path: '/guides/wheelset-buyers', tabs: wheelsetBuyerTabs },
  { path: '/company/about', tabs: companyAboutTabs },
  { path: '/support/warranty', tabs: warrantyTabs },
  { path: '/support/test-report', tabs: testReportTabs },
  { path: '/spoke-calculator', tabs: spokeCalculatorTabs },
  { path: '/membershipandpoints', tabs: membershipAndPointsTabs },
  { path: '/picture-warehouse', tabs: pictureWarehouseTabs },
] as const satisfies readonly PageSubNavigationEntry[]

export const isPageSubNavigationTabId = <Tabs extends readonly PageSubNavigationTab[]>(
  tabs: Tabs,
  id: string
): id is Tabs[number]['id'] => tabs.some((tab) => tab.id === id)

const routePathFromTo = (to: string) => {
  return to.split('#')[0]?.split('?')[0] || '/'
}

const childTargetForTab = (entry: PageSubNavigationEntry, tab: PageSubNavigationTab) => {
  return tab.to || `${entry.path}#${tab.id}`
}

const belongsToSection = (
  section: PrimaryMegaNavSection,
  path: string,
  localeCodes: string[] = []
) => {
  return section.routePrefixes.some((prefix) => primaryMegaNavPathMatches(path, prefix, localeCodes))
}

const belongsToCard = (
  card: PrimaryMegaNavCard,
  path: string,
  localeCodes: string[] = []
) => {
  const normalizedPath = normalizePrimaryMegaNavPath(path, localeCodes)
  const cardPath = normalizePrimaryMegaNavPath(routePathFromTo(card.to), localeCodes)

  return normalizedPath === cardPath || primaryMegaNavPathMatches(normalizedPath, cardPath, localeCodes)
}

export const getPageSubNavigationForPath = (
  path: string,
  localeCodes: string[] = []
) => {
  const normalizedPath = normalizePrimaryMegaNavPath(routePathFromTo(path), localeCodes)

  return pageSubNavigationEntries.find((entry) =>
    normalizePrimaryMegaNavPath(entry.path, localeCodes) === normalizedPath
  ) || null
}

export const pageSubNavigationBelongsToCard = (
  section: PrimaryMegaNavSection,
  card: PrimaryMegaNavCard,
  entry: PageSubNavigationEntry,
  localeCodes: string[] = []
) => {
  const entryPath = normalizePrimaryMegaNavPath(entry.path, localeCodes)

  return belongsToSection(section, entryPath, localeCodes) && belongsToCard(card, entryPath, localeCodes)
}

export const getPrimaryMegaNavCardChildren = (
  section: PrimaryMegaNavSection,
  card: PrimaryMegaNavCard,
  localeCodes: string[] = []
): PageSubNavigationChild[] => {
  const children: PageSubNavigationChild[] = []

  for (const entry of pageSubNavigationEntries) {
    if (!pageSubNavigationBelongsToCard(section, card, entry, localeCodes)) continue

    for (const tab of entry.tabs) {
      const to = childTargetForTab(entry, tab)
      const childPath = normalizePrimaryMegaNavPath(routePathFromTo(to), localeCodes)
      if (!belongsToSection(section, childPath, localeCodes) || !belongsToCard(card, childPath, localeCodes)) {
        continue
      }

      children.push({ ...tab, to })
    }
  }

  return children
}
