import type { PrimaryMegaNavCard, PrimaryMegaNavSection } from '~/utils/primaryMegaNav'
import {
  normalizePrimaryMegaNavPath,
  primaryMegaNavPathMatches,
} from '~/utils/primaryMegaNav'
import localeManifest from '~/i18n/locales.manifest'

export interface PageSubNavigationTab {
  id: string
  label?: string
  labelKey?: string
  fallback?: string
  description?: string
  descriptionKey?: string
  to?: string
}

export interface PageSubNavigationEntry {
  /**
   * Canonical page route that owns the tabs.
   * Child route paths are generated from tab ids unless a tab provides its own `to`.
   */
  path: string
  tabs: readonly PageSubNavigationTab[]
}

export type PageSubNavigationChild = PageSubNavigationTab & {
  to: string
}

export type PageSubNavigationPathMatchMode = 'exact' | 'nested'

export interface PageSubNavigationTabFromPathOptions {
  localeCodes?: string[]
  match?: PageSubNavigationPathMatchMode
}

export const tireGuideTabs = [
  { id: 'size', label: 'Tire size', description: 'Size charts by tire width, rim range, and fit reference.' },
  { id: 'match', label: 'Match', description: 'Match tires with rim profiles and riding conditions.' },
  { id: 'tubeless', label: 'Tubeless tires', description: 'Tubeless setup notes, sealant basics, and compatibility.' },
  { id: 'installation', label: 'Installation', description: 'Mounting steps and practical installation checks.' },
  { id: 'choose', label: 'How to choose', description: 'Selection tips for terrain, clearance, and use case.' },
  { id: 'rims', label: 'Tire pressure', description: 'Recommended pressure ranges and adjustment cues.' },
  { id: 'tube', label: 'Inner tube', description: 'Tube selection notes, valve types, and sizing basics.' },
] as const satisfies readonly PageSubNavigationTab[]

export type TireGuideTabId = (typeof tireGuideTabs)[number]['id']

export const wheelsetBuyerTabs = [
  { id: 'overview', label: 'Buying overview', description: 'Start here for the wheelset buying path and key checks.' },
  { id: 'safety-instructions', label: 'Safety instructions', description: 'Core safety checks before riding and servicing.' },
  { id: 'sample-assembly', label: 'Sample assembly', description: 'Assembly example with parts and setup references.' },
  { id: 'special-order', label: 'Special order', description: 'Custom order options and request details.' },
  { id: 'appearance-logo', label: 'Appearance Logo', description: 'Decal, finish, and logo customization notes.' },
  { id: 'choose-freehub', label: 'Choose freehub', description: 'Freehub choices and drivetrain compatibility.' },
  { id: 'wheel-components', label: 'Wheel Components', description: 'Hubs, rims, spokes, nipples, and build parts.' },
  { id: 'optional', label: 'Optional', description: 'Optional upgrades and configuration add-ons.' },
] as const satisfies readonly PageSubNavigationTab[]

export type WheelsetBuyerTabId = (typeof wheelsetBuyerTabs)[number]['id']

export const companyAboutTabs = [
  { id: 'factory', label: 'Factory', description: 'Production scale, workshop flow, and factory context.' },
  { id: 'appearance', label: 'Appearance', description: 'Brand visuals, finishes, decals, and surface details.' },
  { id: 'hole-patterns', label: 'Hole Patterns', description: 'Drilling layouts for rim, spoke, and hub matching.' },
  { id: 'facility', label: 'Facility', description: 'Equipment, work areas, and production capacity.' },
  { id: 'manufacture', label: 'Manufacture', description: 'Manufacturing process, build steps, and workflow.' },
  { id: 'qualitycontrol', label: 'Quality control', description: 'Inspection standards, testing, and QC checkpoints.' },
] as const satisfies readonly PageSubNavigationTab[]

export type CompanyAboutTabId = (typeof companyAboutTabs)[number]['id']

export const warrantyTabs = [
  { id: 'change-cancel', label: 'Change / Cancel', description: 'How to update, pause, or cancel an order.' },
  { id: 'damaged-lost', label: 'Damaged or Lost Goods', description: 'What to do when goods arrive damaged or missing.' },
  { id: 'returns', label: 'Returns', description: 'Return conditions, timing, and handling process.' },
  { id: 'warranty', label: 'Warranty', description: 'Coverage scope, duration, and claim basics.' },
  { id: 'accidental-damage', label: 'Accidental Damage', description: 'Support for accidental riding or handling damage.' },
  { id: 'protection', label: 'Protection', description: 'Protection options and service expectations.' },
  { id: 'submit-warranty', label: 'Submit Warranty', description: 'Submit documents and start a warranty request.' },
] as const satisfies readonly PageSubNavigationTab[]

export type WarrantyTabId = (typeof warrantyTabs)[number]['id']

export const testReportTabs = [
  { id: 'rim-test-report', label: 'Rim Test Report', description: 'Rim testing documents and compliance references.' },
  { id: 'wheelset-test-report', label: 'Wheelset Test Report', description: 'Wheelset testing reports and validation data.' },
  { id: 'tension', label: 'Tension', description: 'Spoke tension targets, ranges, and build checks.' },
  { id: 'wheelset-assembly', label: 'Wheelset Assembly', description: 'Assembly records, process notes, and final checks.' },
] as const satisfies readonly PageSubNavigationTab[]

export type TestReportTabId = (typeof testReportTabs)[number]['id']

export const spokeCalculatorTabs = [
  {
    id: 'calculator',
    labelKey: 'spokeCalculator.tabs.calculator',
    fallback: 'Calculator',
    description: 'Enter wheel data and calculate spoke length.',
  },
  {
    id: 'parameter',
    labelKey: 'spokeCalculator.tabs.parameter',
    fallback: 'Parameter',
    description: 'Rim, hub, lacing, and offset reference data.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type SpokeCalculatorTabId = (typeof spokeCalculatorTabs)[number]['id']

export const membershipAndPointsTabs = [
  {
    id: 'myinfo',
    labelKey: 'member.tabs.myInfo',
    fallback: 'My info',
    description: 'Account profile, benefits, and member details.',
  },
  {
    id: 'levers',
    labelKey: 'member.tabs.levers',
    fallback: 'Levers',
    description: 'Reward levels, points rules, and upgrade status.',
  },
  {
    id: 'exchange',
    labelKey: 'member.tabs.exchange',
    fallback: 'Exchange',
    description: 'Points redemption, exchange options, and records.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type MembershipTabId = (typeof membershipAndPointsTabs)[number]['id']

export const pictureWarehouseTabs = [
  { id: 'riders', label: 'Riders photos', description: 'Customer and rider photo references.' },
  { id: 'brand', label: 'Brand photos', description: 'Product and brand image library.' },
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

const defaultPageSubNavigationLocaleCodes = localeManifest.map(locale => locale.code)

const routePathFromTo = (to: string) => {
  return to.split('?')[0] || '/'
}

export const pageSubNavigationChildPath = (basePath: string, tabId: string) => {
  const normalizedBasePath = routePathFromTo(basePath).replace(/\/+$/, '') || '/'
  return normalizedBasePath === '/' ? `/${tabId}` : `${normalizedBasePath}/${tabId}`
}

export const getPageSubNavigationTabFromPath = <Tabs extends readonly PageSubNavigationTab[]>(
  tabs: Tabs,
  basePath: string,
  path: string | null | undefined,
  options: PageSubNavigationTabFromPathOptions = {}
): Tabs[number]['id'] | null => {
  const localeCodes = options.localeCodes || defaultPageSubNavigationLocaleCodes
  const match = options.match || 'nested'
  const normalizedPath = normalizePrimaryMegaNavPath(routePathFromTo(String(path || '/')), localeCodes)

  for (const tab of tabs) {
    const tabPath = tab.to || pageSubNavigationChildPath(basePath, tab.id)
    const normalizedTabPath = normalizePrimaryMegaNavPath(routePathFromTo(tabPath), localeCodes)
    const matches = match === 'exact'
      ? normalizedPath === normalizedTabPath
      : primaryMegaNavPathMatches(normalizedPath, normalizedTabPath, localeCodes)

    if (matches) return tab.id
  }

  return null
}

const childTargetForTab = (entry: PageSubNavigationEntry, tab: PageSubNavigationTab) => {
  return tab.to || pageSubNavigationChildPath(entry.path, tab.id)
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

  return pageSubNavigationEntries.find((entry) => {
    const entryPath = normalizePrimaryMegaNavPath(entry.path, localeCodes)

    return (
      entryPath === normalizedPath ||
      Boolean(getPageSubNavigationTabFromPath(entry.tabs, entry.path, normalizedPath, {
        localeCodes,
        match: 'nested',
      }))
    )
  }) || null
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
