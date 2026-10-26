export type PrimaryMegaNavId = 'products' | 'technology' | 'resources' | 'support' | 'company' | 'tools' | 'guides'

export type PrimaryMegaNavCardSize = 'feature' | 'wide' | 'standard' | 'compact'

export type PrimaryMegaNavAccent =
  | 'emerald'

export interface PrimaryMegaNavCard {
  /** stable id for rendering and analytics */
  id: string
  /** existing i18n label key; rendered with `labelFallback` when missing */
  labelKey: string
  /** safe fallback while detailed mega-menu copy is not localized yet */
  labelFallback: string
  /** display title; defaults to the translated label when omitted */
  title?: string
  /** discover direct child index pages under this card route automatically */
  discoverChildPages?: boolean
  /** optional translation namespace for automatically discovered children */
  childLabelNamespace?: string
  /** short card description for the desktop mega menu */
  description: string
  /** route path; query is supported and localized by the component */
  to: string
  /** Nuxt Icon name */
  icon: string
  /** mixed card sizes used by the header mega menu layout */
  size: PrimaryMegaNavCardSize
  /** visual accent token, mapped in HeaderMegaMenu.vue */
  accent: PrimaryMegaNavAccent
}

export interface PrimaryMegaNavSection {
  id: PrimaryMegaNavId
  labelKey: string
  labelFallback: string
  /**
   * Canonical path prefixes that make a route belong to this top-level section.
   * Keep explicit route families in their own section: /resources stays in
   * Resources, /technology stays in Technology, /guides stays in Guides,
   * /support stays in Support, and /company stays in Company.
   */
  routePrefixes: string[]
  cards: PrimaryMegaNavCard[]
}

export const normalizePrimaryMegaNavPath = (path: string, localeCodes: string[] = []) => {
  const pathWithoutQuery = (path || '/').split('?')[0] || '/'
  const absolutePath = pathWithoutQuery.startsWith('/') ? pathWithoutQuery : `/${pathWithoutQuery}`
  const segments = absolutePath.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''
  const withoutLocale =
    firstSegment && localeCodes.includes(firstSegment)
      ? `/${segments.slice(1).join('/')}`
      : absolutePath

  return withoutLocale.replace(/\/+$/, '') || '/'
}

export const primaryMegaNavPathMatches = (
  currentPath: string,
  targetPath: string,
  localeCodes: string[] = []
) => {
  const current = normalizePrimaryMegaNavPath(currentPath, localeCodes)
  const target = normalizePrimaryMegaNavPath(targetPath, localeCodes)

  return (
    current === target ||
    (current.startsWith(target) && current[target.length] === '/')
  )
}

export const findPrimaryMegaNavSectionByPath = (
  path: string,
  sections: PrimaryMegaNavSection[],
  localeCodes: string[] = []
) => {
  const currentPath = normalizePrimaryMegaNavPath(path, localeCodes)

  return sections.find((section) =>
    section.routePrefixes.some((prefix) => primaryMegaNavPathMatches(currentPath, prefix))
  ) || null
}

/**
 * Desktop header mega-menu data.
 *
 * This is the single source of truth for top-level header navigation.
 * Header panels only list concrete canonical child routes. Do not cross-list
 * routes with another section prefix here; the current route ownership and the
 * sticky section tab bar both derive from this file.
 */
export const primaryMegaNavSections: PrimaryMegaNavSection[] = [
  {
    id: 'products',
    labelKey: 'products.nav.shop',
    labelFallback: 'Shop',
    routePrefixes: ['/shop', '/products'],
    cards: [
      {
        id: 'shop',
        labelKey: 'products.nav.shop',
        labelFallback: 'Shop',
        title: 'Shop',
        description: 'Browse complete wheel products and product listings.',
        to: '/shop',
        icon: 'lucide:shopping-bag',
        size: 'feature',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'technology',
    labelKey: 'footer.menus.technology',
    labelFallback: 'Technology',
    routePrefixes: ['/technology'],
    cards: [
      {
        id: 'test-report',
        labelKey: 'support.nav.testReport',
        labelFallback: 'Test Report',
        description: 'Rim and wheelset testing, spoke tension, and assembly references.',
        to: '/technology/test-report',
        icon: 'lucide:file-check-2',
        size: 'feature',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'resources',
    labelKey: 'footer.menus.resources',
    labelFallback: 'Resources',
    routePrefixes: ['/resources'],
    cards: [
      {
        id: 'membership-and-points',
        labelKey: 'company.nav.membershipPoints',
        labelFallback: 'Membership & Points',
        description: 'Membership, rewards, and points information.',
        to: '/resources/membershipandpoints',
        icon: 'lucide:gem',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'picture-warehouse',
        labelKey: 'company.nav.pictureWarehouse',
        labelFallback: 'Picture Warehouse',
        description: 'Visual assets, product photos, and useful reference images.',
        to: '/resources/picture-warehouse',
        icon: 'lucide:images',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'blog',
        labelKey: 'footer.links.blog',
        labelFallback: 'Blog',
        title: 'Blog',
        childLabelNamespace: 'blog.nav',
        description: 'Articles, updates, and long-form content.',
        to: '/resources/blog',
        icon: 'lucide:newspaper',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'workbench',
        labelKey: 'workbench.nav.feed',
        labelFallback: 'Workbench Feed',
        description: 'Short, real notes from wheel builds, repairs, measurements, and fitment work.',
        to: '/resources/workbench',
        icon: 'lucide:wrench',
        size: 'standard',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'support',
    labelKey: 'footer.menus.support',
    labelFallback: 'Support',
    routePrefixes: ['/support'],
    cards: [
      {
        id: 'faqs',
        labelKey: 'support.nav.faqs',
        labelFallback: 'All FAQs',
        description: 'Quick answers by topic.',
        to: '/support/faqs',
        icon: 'lucide:circle-help',
        size: 'wide',
        accent: 'emerald',
      },
      {
        id: 'payment',
        labelKey: 'support.nav.payment',
        labelFallback: 'Payment',
        description: 'Payment and checkout help.',
        to: '/support/payment',
        icon: 'lucide:credit-card',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'shipping',
        labelKey: 'support.nav.shipping',
        labelFallback: 'Shipping',
        description: 'Delivery and tracking help.',
        to: '/support/shipping',
        icon: 'lucide:truck',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'warranty',
        labelKey: 'support.nav.warranty',
        labelFallback: 'Warranty',
        description: 'Coverage and claim rules.',
        to: '/support/warranty',
        icon: 'lucide:shield-check',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'warranty-check',
        labelKey: 'support.nav.warrantyCheck',
        labelFallback: 'Warranty Check',
        description: 'Check warranty status.',
        to: '/support/warranty-check',
        icon: 'lucide:badge-check',
        size: 'standard',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'company',
    labelKey: 'footer.menus.company',
    labelFallback: 'Company',
    routePrefixes: ['/company'],
    cards: [
      {
        id: 'our-story',
        labelKey: 'company.nav.ourStory',
        labelFallback: 'Our Story',
        description: 'Brand story, factory capability, and the path behind our products.',
        to: '/company/ourstory',
        icon: 'lucide:sparkles',
        size: 'feature',
        accent: 'emerald',
      },
      {
        id: 'about',
        labelKey: 'company.nav.about',
        labelFallback: 'About',
        description: 'A short company introduction and entry point for visitors.',
        to: '/company/about',
        icon: 'lucide:landmark',
        size: 'wide',
        accent: 'emerald',
      },
      {
        id: 'global-partners',
        labelKey: 'company.nav.globalPartners',
        labelFallback: 'Global Partners',
        description: 'Global cooperation, partner resources, and regional relationships.',
        to: '/company/global-partners',
        icon: 'lucide:globe-2',
        size: 'wide',
        accent: 'emerald',
      },
      {
        id: 'oem-odm',
        labelKey: 'company.nav.oemOdm',
        labelFallback: 'OEM / ODM',
        description: 'Custom manufacturing, private label, and wheel project cooperation.',
        to: '/company/oem-odm',
        icon: 'lucide:factory',
        size: 'feature',
        accent: 'emerald',
      },
      {
        id: 'certificates',
        labelKey: 'company.nav.certificates',
        labelFallback: 'Certificates',
        description: 'Company certificates and qualification references.',
        to: '/company/certificates',
        icon: 'lucide:award',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'contact',
        labelKey: 'company.nav.contact',
        labelFallback: 'Contact',
        description: 'Reach us for sales, service, and business communication.',
        to: '/company/contact',
        icon: 'lucide:mail',
        size: 'wide',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'tools',
    labelKey: 'toolsNavigation.label',
    labelFallback: 'Tools',
    routePrefixes: ['/tools'],
    cards: [
      {
        id: 'interactive-tool-catalog',
        // Complete tool routes are listed as cards below, not inferred from
        // page filenames that can have an explicit public path override.
        discoverChildPages: false,
        labelKey: 'toolsNavigation.label',
        labelFallback: 'Tools',
        description: 'Browse interactive tools for bicycle component selection.',
        to: '/tools',
        icon: 'lucide:wrench',
        size: 'wide',
        accent: 'emerald',
      },
      {
        id: 'schwalbe-tire-selector',
        labelKey: 'toolsNavigation.schwalbeTireSelectorLabel',
        labelFallback: 'Schwalbe tire selector',
        description: 'Search and filter the Schwalbe tire catalog with its reference explanations.',
        to: '/tools/schwalbe-tire-selector',
        icon: 'lucide:search',
        size: 'feature',
        accent: 'emerald',
      },
      {
        id: 'brand-wheelset-spoke-specs',
        labelKey: 'toolsNavigation.brandWheelsetSpokeSpecsLabel',
        labelFallback: 'Wheelset spoke and nipple specifications',
        description: 'Find replacement spoke lengths and nipple specifications by brand and wheelset model.',
        to: '/tools/brand-wheelset-spoke-specs',
        icon: 'lucide:table-properties',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'spoke-length-calculator',
        labelKey: 'toolsNavigation.spokeLengthCalculatorLabel',
        labelFallback: 'Spoke length calculator',
        description: 'Calculate spoke lengths through the complete rim, hub, lacing, and nipple wizard.',
        to: '/tools/spoke-length-calculator',
        icon: 'lucide:calculator',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'spoke-lacing-topology-simulator',
        labelKey: 'toolsNavigation.spokeLacingTopologySimulatorLabel',
        labelFallback: 'Spoke lacing topology simulator',
        description: 'Preview complete spoke-hole mapping and flange geometry with the full workbench and explanations.',
        to: '/tools/spoke-lacing-topology-simulator',
        icon: 'lucide:network',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'dynamic-tire-pressure-calculator',
        labelKey: 'toolsNavigation.dynamicTirePressureCalculatorLabel',
        labelFallback: 'Dynamic tire pressure calculator',
        description: 'Explore how load, pressure, speed, and lean change estimated tire contact area and cornering force.',
        to: '/tools/dynamic-tire-pressure-calculator',
        icon: 'lucide:gauge',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'tire-rim-width-compatibility-engine',
        labelKey: 'toolsNavigation.tireRimWidthCompatibilityEngineLabel',
        labelFallback: 'Tire and rim width compatibility engine',
        description: 'Calculate tire and rim widths, inspect the fitment simulation, and compare published reference charts.',
        to: '/tools/tire-rim-width-compatibility-engine',
        icon: 'lucide:scan-line',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'tire-size-marking-converter',
        labelKey: 'toolsNavigation.tireSizeMarkingConverterLabel',
        labelFallback: 'Tire size marking converter',
        description: 'Convert ETRTO, inch, and French tire markings with the complete size reference and fitment limits.',
        to: '/tools/tire-size-marking-converter',
        icon: 'lucide:ruler',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'freehub-cassette-fitment-engine',
        labelKey: 'toolsNavigation.freehubCassetteFitmentEngineLabel',
        labelFallback: 'Freehub and cassette compatibility engine',
        description: 'Match cassette standards, freehub bodies, spacers, and mechanical fitment rules with the complete guide.',
        to: '/tools/freehub-cassette-fitment-engine',
        icon: 'lucide:settings-2',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'valve-length-fitment-calculator',
        labelKey: 'toolsNavigation.valveLengthFitmentCalculatorLabel',
        labelFallback: 'Inner tube valve-length fitment calculator',
        description: 'Match inner-tube size, valve, material, and deep-rim valve length with the complete guide and calculator.',
        to: '/tools/valve-length-fitment-calculator',
        icon: 'lucide:circle-gauge',
        size: 'standard',
        accent: 'emerald',
      },
      {
        id: 'spoke-stress-relief-mechanics',
        labelKey: 'toolsNavigation.spokeStressReliefMechanicsLabel',
        labelFallback: 'Stainless steel spoke stress-relief mechanics',
        description: 'Calculate axial spoke elongation from material, section, length, and working tension inputs with the full engineering context.',
        to: '/tools/spoke-stress-relief-mechanics',
        icon: 'lucide:atom',
        size: 'standard',
        accent: 'emerald',
      },
    ],
  },
  {
    id: 'guides',
    labelKey: 'breadcrumbs.guides',
    labelFallback: 'Guides',
    routePrefixes: ['/guides'],
    cards: [
      {
        id: 'tire-guides',
        labelKey: 'products.nav.tireSizeCharts',
        labelFallback: 'Tire Guides',
        description: 'Tire sizing, pressure, inner tube, and tubeless setup and installation topics.',
        to: '/guides/tireguides',
        icon: 'lucide:ruler',
        size: 'feature',
        accent: 'emerald',
      },
      {
        id: 'wheelset-buyers',
        labelKey: 'products.nav.wheelsetBuyersGuide',
        labelFallback: 'Wheelset Buyers Guide',
        description: 'How to think through wheel selection, components, and riding requirements.',
        to: '/guides/wheelset-buyers',
        icon: 'lucide:map',
        size: 'wide',
        accent: 'emerald',
      },
    ],
  },
]
