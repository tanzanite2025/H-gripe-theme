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
  {
    id: 'size',
    labelKey: 'guidesTireguides.tabs.size.label',
    fallback: 'Tire size',
    descriptionKey: 'guidesTireguides.tabs.size.description',
    description: 'Size charts by tire width, rim range, and fit reference.',
  },
  {
    id: 'match',
    labelKey: 'guidesTireguides.tabs.match.label',
    fallback: 'Match',
    descriptionKey: 'guidesTireguides.tabs.match.description',
    description: 'Match tires with rim profiles and riding conditions.',
  },
  {
    id: 'tubeless',
    labelKey: 'guidesTireguides.tabs.tubeless.label',
    fallback: 'Tubeless tires',
    descriptionKey: 'guidesTireguides.tabs.tubeless.description',
    description: 'Tubeless setup notes, sealant basics, and compatibility.',
  },
  {
    id: 'installation',
    labelKey: 'guidesTireguides.tabs.installation.label',
    fallback: 'Installation',
    descriptionKey: 'guidesTireguides.tabs.installation.description',
    description: 'Mounting steps and practical installation checks.',
  },
  {
    id: 'choose',
    labelKey: 'guidesTireguides.tabs.choose.label',
    fallback: 'How to choose',
    descriptionKey: 'guidesTireguides.tabs.choose.description',
    description: 'Selection tips for terrain, clearance, and use case.',
  },
  {
    id: 'rims',
    labelKey: 'guidesTireguides.tabs.pressure.label',
    fallback: 'Tire pressure',
    descriptionKey: 'guidesTireguides.tabs.pressure.description',
    description: 'Recommended pressure ranges and adjustment cues.',
  },
  {
    id: 'tube',
    labelKey: 'guidesTireguides.tabs.innerTube.label',
    fallback: 'Inner tube',
    descriptionKey: 'guidesTireguides.tabs.innerTube.description',
    description: 'Tube selection notes, valve types, and sizing basics.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type TireGuideTabId = (typeof tireGuideTabs)[number]['id']

export const wheelsetBuyerTabs = [
  {
    id: 'overview',
    labelKey: 'guidesWheelsetBuyers.tabs.overview.label',
    fallback: 'Buying overview',
    descriptionKey: 'guidesWheelsetBuyers.tabs.overview.description',
    description: 'Start here for the wheelset buying path and key checks.',
  },
  {
    id: 'safety-instructions',
    labelKey: 'guidesWheelsetBuyers.tabs.safetyInstructions.label',
    fallback: 'Safety instructions',
    descriptionKey: 'guidesWheelsetBuyers.tabs.safetyInstructions.description',
    description: 'Core safety checks before riding and servicing.',
  },
  {
    id: 'sample-assembly',
    labelKey: 'guidesWheelsetBuyers.tabs.sampleAssembly.label',
    fallback: 'Sample assembly',
    descriptionKey: 'guidesWheelsetBuyers.tabs.sampleAssembly.description',
    description: 'Assembly example with parts and setup references.',
  },
  {
    id: 'special-order',
    labelKey: 'guidesWheelsetBuyers.tabs.specialOrder.label',
    fallback: 'Special order',
    descriptionKey: 'guidesWheelsetBuyers.tabs.specialOrder.description',
    description: 'Custom order options and request details.',
  },
  {
    id: 'appearance-logo',
    labelKey: 'guidesWheelsetBuyers.tabs.appearanceLogo.label',
    fallback: 'Appearance Logo',
    descriptionKey: 'guidesWheelsetBuyers.tabs.appearanceLogo.description',
    description: 'Decal, finish, and logo customization notes.',
  },
  {
    id: 'choose-freehub',
    labelKey: 'guidesWheelsetBuyers.tabs.chooseFreehub.label',
    fallback: 'Choose freehub',
    descriptionKey: 'guidesWheelsetBuyers.tabs.chooseFreehub.description',
    description: 'Freehub choices and drivetrain compatibility.',
  },
  {
    id: 'wheel-components',
    labelKey: 'guidesWheelsetBuyers.tabs.wheelComponents.label',
    fallback: 'Wheel Components',
    descriptionKey: 'guidesWheelsetBuyers.tabs.wheelComponents.description',
    description: 'Hubs, rims, spokes, nipples, and build parts.',
  },
  {
    id: 'optional',
    labelKey: 'guidesWheelsetBuyers.tabs.optional.label',
    fallback: 'Optional',
    descriptionKey: 'guidesWheelsetBuyers.tabs.optional.description',
    description: 'Optional upgrades and configuration add-ons.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type WheelsetBuyerTabId = (typeof wheelsetBuyerTabs)[number]['id']

export const companyAboutTabs = [
  {
    id: 'factory',
    labelKey: 'company.aboutTabs.factory.label',
    fallback: 'Factory',
    descriptionKey: 'company.aboutTabs.factory.description',
    description: 'Production scale, workshop flow, and factory context.',
  },
  {
    id: 'appearance',
    labelKey: 'company.aboutTabs.appearance.label',
    fallback: 'Appearance',
    descriptionKey: 'company.aboutTabs.appearance.description',
    description: 'Brand visuals, finishes, decals, and surface details.',
  },
  {
    id: 'hole-patterns',
    labelKey: 'company.aboutTabs.holePatterns.label',
    fallback: 'Hole Patterns',
    descriptionKey: 'company.aboutTabs.holePatterns.description',
    description: 'Drilling layouts for rim, spoke, and hub matching.',
  },
  {
    id: 'facility',
    labelKey: 'company.aboutTabs.facility.label',
    fallback: 'Facility',
    descriptionKey: 'company.aboutTabs.facility.description',
    description: 'Equipment, work areas, and production capacity.',
  },
  {
    id: 'manufacture',
    labelKey: 'company.aboutTabs.manufacture.label',
    fallback: 'Manufacture',
    descriptionKey: 'company.aboutTabs.manufacture.description',
    description: 'Manufacturing process, build steps, and workflow.',
  },
  {
    id: 'qualitycontrol',
    labelKey: 'company.aboutTabs.qualityControl.label',
    fallback: 'Quality control',
    descriptionKey: 'company.aboutTabs.qualityControl.description',
    description: 'Inspection standards, testing, and QC checkpoints.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type CompanyAboutTabId = (typeof companyAboutTabs)[number]['id']

export const warrantyTabs = [
  {
    id: 'damaged-lost',
    labelKey: 'supportWarranty.tabs.damagedLost.label',
    fallback: 'Damaged or Lost Goods',
    descriptionKey: 'supportWarranty.tabs.damagedLost.description',
    description: 'What to do when goods arrive damaged or missing.',
  },
  {
    id: 'warranty',
    labelKey: 'supportWarranty.tabs.warranty.label',
    fallback: 'Warranty',
    descriptionKey: 'supportWarranty.tabs.warranty.description',
    description: 'Coverage scope, duration, and claim basics.',
  },
  {
    id: 'accidental-damage',
    labelKey: 'supportWarranty.tabs.accidentalDamage.label',
    fallback: 'Accidental Damage',
    descriptionKey: 'supportWarranty.tabs.accidentalDamage.description',
    description: 'Support for accidental riding or handling damage.',
  },
  {
    id: 'protection',
    labelKey: 'supportWarranty.tabs.protection.label',
    fallback: 'Protection',
    descriptionKey: 'supportWarranty.tabs.protection.description',
    description: 'Protection options and service expectations.',
  },
  {
    id: 'submit-warranty',
    labelKey: 'supportWarranty.tabs.submitWarranty.label',
    fallback: 'Submit Warranty',
    descriptionKey: 'supportWarranty.tabs.submitWarranty.description',
    description: 'Submit documents and start a warranty request.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type WarrantyTabId = (typeof warrantyTabs)[number]['id']

export const testReportTabs = [
  {
    id: 'rim-test-report',
    labelKey: 'supportTestReport.tabs.rim.label',
    fallback: 'Rim Test Report',
    descriptionKey: 'supportTestReport.tabs.rim.description',
    description: 'Rim testing documents and compliance references.',
  },
  {
    id: 'wheelset-test-report',
    labelKey: 'supportTestReport.tabs.wheelset.label',
    fallback: 'Wheelset Test Report',
    descriptionKey: 'supportTestReport.tabs.wheelset.description',
    description: 'Wheelset testing reports and validation data.',
  },
  {
    id: 'tension',
    labelKey: 'supportTestReport.tabs.tension.label',
    fallback: 'Tension',
    descriptionKey: 'supportTestReport.tabs.tension.description',
    description: 'Spoke tension targets, ranges, and build checks.',
  },
  {
    id: 'wheelset-assembly',
    labelKey: 'supportTestReport.tabs.assembly.label',
    fallback: 'Wheelset Assembly',
    descriptionKey: 'supportTestReport.tabs.assembly.description',
    description: 'Assembly records, process notes, and final checks.',
  },
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
    pageTitleKey: 'resourcesMembershipMyInfo.title',
    pageTitle: 'My Membership Info',
    pageIntroKey: 'resourcesMembershipMyInfo.intro',
    pageIntro: 'View your profile, membership level, points balance, coupons, and warranty tools.',
    seoTitleKey: 'resourcesMembershipMyInfo.seoTitle',
    seoTitle: 'My Membership Info - Points and Benefits',
    seoDescriptionKey: 'resourcesMembershipMyInfo.seoDescription',
    seoDescription: 'View your member profile, loyalty level, points balance, coupons, and warranty tools.',
    feedbackThreadKey: 'membership-myinfo',
    feedbackTitleKey: 'resourcesMembershipMyInfo.feedbackTitle',
    feedbackTitle: 'Share your feedback about your membership info',
    feedbackSubtitleKey: 'resourcesMembershipMyInfo.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether account details, benefits, coupons, or warranty tools are clear.',
  },
  {
    id: 'levers',
    labelKey: 'member.tabs.levers',
    fallback: 'Levers',
    description: 'Reward levels, points rules, and upgrade status.',
    pageTitleKey: 'resourcesMembershipLevers.title',
    pageTitle: 'Membership Levels and Point Rules',
    pageIntroKey: 'resourcesMembershipLevers.intro',
    pageIntro: 'Review member tiers, earning rules, referral rewards, and check-in rules.',
    seoTitleKey: 'resourcesMembershipLevers.seoTitle',
    seoTitle: 'Membership Levels and Point Rules',
    seoDescriptionKey: 'resourcesMembershipLevers.seoDescription',
    seoDescription: 'Review membership levels, points earning rules, referral rewards, and daily check-in rewards.',
    feedbackThreadKey: 'membership-levers',
    feedbackTitleKey: 'resourcesMembershipLevers.feedbackTitle',
    feedbackTitle: 'Share your feedback about membership levels and points',
    feedbackSubtitleKey: 'resourcesMembershipLevers.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether tier rules, earning rules, referral rewards, or check-in details are clear.',
  },
  {
    id: 'referral',
    labelKey: 'member.tabs.referral',
    fallback: 'Referral code',
    description: 'Your referral code, sharing link, and referral activity.',
    pageTitleKey: 'resourcesMembershipReferral.title',
    pageTitle: 'Referral Code and Rewards',
    pageIntroKey: 'resourcesMembershipReferral.intro',
    pageIntro: 'View your automatically generated referral code and share it with friends. Registration points are added to the same points balance as every other loyalty point.',
    seoTitleKey: 'resourcesMembershipReferral.seoTitle',
    seoTitle: 'Referral Code and Rewards - Membership Points',
    seoDescriptionKey: 'resourcesMembershipReferral.seoDescription',
    seoDescription: 'View your referral code, share link, and referral activity from the membership and points center.',
    feedbackThreadKey: 'membership-referral',
    feedbackTitleKey: 'resourcesMembershipReferral.feedbackTitle',
    feedbackTitle: 'Share your feedback about referral rewards',
    feedbackSubtitleKey: 'resourcesMembershipReferral.feedbackSubtitle',
    feedbackSubtitle: 'Tell us whether your referral code, sharing link, and points rules are clear.',
  },
] as const satisfies readonly PageSubNavigationTab[]

export type MembershipTabId = (typeof membershipAndPointsTabs)[number]['id']

export const pictureWarehouseTabs = [
  {
    id: 'riders',
    labelKey: 'resourcesPictureWarehouseRiders.navLabel',
    fallback: 'Rider photos',
    descriptionKey: 'resourcesPictureWarehouseRiders.navDescription',
    description: 'Customer and rider photo references.',
  },
  {
    id: 'brand',
    labelKey: 'resourcesPictureWarehouseBrand.navLabel',
    fallback: 'Brand photos',
    descriptionKey: 'resourcesPictureWarehouseBrand.navDescription',
    description: 'Product and brand image library.',
  },
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
