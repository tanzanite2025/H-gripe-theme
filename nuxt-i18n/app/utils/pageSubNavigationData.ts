export interface PageSubNavigationTab {
  id: string
  label?: string
  labelKey?: string
  fallback?: string
  description?: string
  descriptionKey?: string
  pageTitle?: string
  pageTitleKey?: string
  pageIntro?: string
  pageIntroKey?: string
  seoTitle?: string
  seoTitleKey?: string
  seoDescription?: string
  seoDescriptionKey?: string
  feedbackThreadKey?: string
  feedbackTitle?: string
  feedbackTitleKey?: string
  feedbackSubtitle?: string
  feedbackSubtitleKey?: string
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
    pageTitleKey: 'member.pages.myInfo.title',
    pageTitle: 'My Membership Info',
    pageIntroKey: 'member.pages.myInfo.intro',
    pageIntro: 'View your profile, membership level, points balance, coupons, gift cards, and warranty tools.',
    seoTitleKey: 'member.pages.myInfo.seoTitle',
    seoTitle: 'My Membership Info - Points and Benefits',
    seoDescriptionKey: 'member.pages.myInfo.seoDescription',
    seoDescription: 'View your member profile, loyalty level, points balance, coupons, gift cards, and warranty tools.',
    feedbackThreadKey: 'membership-myinfo',
    feedbackTitleKey: 'member.pages.myInfo.feedbackTitle',
    feedbackTitle: 'Share your feedback about your membership info',
    feedbackSubtitleKey: 'member.pages.myInfo.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether account details, benefits, coupons, gift cards, or warranty tools are clear.',
  },
  {
    id: 'levers',
    labelKey: 'member.tabs.levers',
    fallback: 'Levers',
    description: 'Reward levels, points rules, and upgrade status.',
    pageTitleKey: 'member.pages.levers.title',
    pageTitle: 'Membership Levels and Point Rules',
    pageIntroKey: 'member.pages.levers.intro',
    pageIntro: 'Review member tiers, discount levels, earning rules, redemption rates, referral rewards, and check-in rules.',
    seoTitleKey: 'member.pages.levers.seoTitle',
    seoTitle: 'Membership Levels and Point Rules',
    seoDescriptionKey: 'member.pages.levers.seoDescription',
    seoDescription: 'Review membership levels, points earning rules, redemption rates, referral rewards, and daily check-in rewards.',
    feedbackThreadKey: 'membership-levers',
    feedbackTitleKey: 'member.pages.levers.feedbackTitle',
    feedbackTitle: 'Share your feedback about membership levels and points',
    feedbackSubtitleKey: 'member.pages.levers.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether tier rules, earning rules, redemption, referral rewards, or check-in details are clear.',
  },
  {
    id: 'exchange',
    labelKey: 'member.tabs.exchange',
    fallback: 'Exchange',
    description: 'Points redemption, exchange options, and records.',
    pageTitleKey: 'member.pages.exchange.title',
    pageTitle: 'Gift Card Exchange',
    pageIntroKey: 'member.pages.exchange.intro',
    pageIntro: 'Redeem eligible loyalty points for gift cards and track redeemed cards from your member center.',
    seoTitleKey: 'member.pages.exchange.seoTitle',
    seoTitle: 'Gift Card Exchange - Redeem Loyalty Points',
    seoDescriptionKey: 'member.pages.exchange.seoDescription',
    seoDescription: 'Redeem loyalty points for available gift card rewards and manage redeemed gift cards from the member center.',
    feedbackThreadKey: 'membership-exchange',
    feedbackTitleKey: 'member.pages.exchange.feedbackTitle',
    feedbackTitle: 'Share your feedback about gift card exchange',
    feedbackSubtitleKey: 'member.pages.exchange.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether point redemption, available gift cards, or redeemed card records are easy to understand.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type MembershipTabId = (typeof membershipAndPointsTabs)[number]['id']

export const pictureWarehouseTabs = [
  { id: 'riders', label: 'Riders photos', description: 'Customer and rider photo references.' },
  { id: 'brand', label: 'Brand photos', description: 'Product and brand image library.' },
] as const satisfies readonly PageSubNavigationTab[]

export type PictureWarehouseTabId = (typeof pictureWarehouseTabs)[number]['id']

/**
 * Navigation for pages whose child routes are virtual tabs generated by Nuxt.
 * Real child page directories are discovered separately by pageSubNavigation.ts.
 */
export const virtualPageSubNavigationEntries = [
  { path: '/guides/tireguides', tabs: tireGuideTabs },
  { path: '/guides/wheelset-buyers', tabs: wheelsetBuyerTabs },
  { path: '/company/about', tabs: companyAboutTabs },
  { path: '/support/warranty', tabs: warrantyTabs },
  { path: '/support/test-report', tabs: testReportTabs },
  { path: '/resources/spoke-calculator', tabs: spokeCalculatorTabs },
  { path: '/resources/membershipandpoints', tabs: membershipAndPointsTabs },
  { path: '/resources/picture-warehouse', tabs: pictureWarehouseTabs },
] as const satisfies readonly PageSubNavigationEntry[]
