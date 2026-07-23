export type PrimaryMegaNavId = 'products' | 'support' | 'company' | 'guides'

export type PrimaryMegaNavCardSize = 'feature' | 'wide' | 'standard' | 'compact'

export type PrimaryMegaNavAccent =
  | 'mint'
  | 'blue'
  | 'violet'
  | 'amber'
  | 'rose'
  | 'slate'

export interface PrimaryMegaNavCard {
  /** stable id for rendering and analytics */
  id: string
  /** existing i18n label key; rendered with `labelFallback` when missing */
  labelKey: string
  /** safe fallback while detailed mega-menu copy is not localized yet */
  labelFallback: string
  /** display title; defaults to the translated label when omitted */
  title?: string
  /** short card description for the desktop mega menu */
  description: string
  /** route path; hash/query are supported and localized by the component */
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
   * Keep explicit route families in their own section: /guides stays in Guides,
   * /support stays in Support, and /company stays in Company.
   */
  routePrefixes: string[]
  cards: PrimaryMegaNavCard[]
}

export const normalizePrimaryMegaNavPath = (path: string, localeCodes: string[] = []) => {
  const pathWithoutHash = (path || '/').split('#')[0] || '/'
  const pathWithoutQuery = pathWithoutHash.split('?')[0] || '/'
  const absolutePath = pathWithoutQuery.startsWith('/') ? pathWithoutQuery : `/${pathWithoutQuery}`
  const segments = absolutePath.split('/').filter(Boolean)
  const withoutLocale =
    segments.length >= 1 && localeCodes.includes(segments[0])
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
 * Legacy hub routes redirect; header panels only list concrete child routes.
 * Do not cross-list routes with another section prefix here; the current route
 * ownership and the sticky section tab bar both derive from this file.
 */
export const primaryMegaNavSections: PrimaryMegaNavSection[] = [
  {
    id: 'products',
    labelKey: 'footer.menus.products',
    labelFallback: 'Products',
    routePrefixes: ['/shop', '/products', '/spoke-calculator', '/membershipandpoints', '/picture-warehouse'],
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
        accent: 'blue',
      },
      {
        id: 'spoke-calculator',
        labelKey: 'support.nav.spokeCalculator',
        labelFallback: 'Spoke Calculator',
        description: 'Calculate spoke length using hub, rim, lacing, and ERD parameters.',
        to: '/spoke-calculator',
        icon: 'lucide:calculator',
        size: 'wide',
        accent: 'mint',
      },
      {
        id: 'membership-and-points',
        labelKey: 'company.nav.membershipPoints',
        labelFallback: 'Membership & Points',
        description: 'Membership, rewards, and points information.',
        to: '/membershipandpoints',
        icon: 'lucide:gem',
        size: 'compact',
        accent: 'violet',
      },
      {
        id: 'picture-warehouse',
        labelKey: 'company.nav.pictureWarehouse',
        labelFallback: 'Picture Warehouse',
        description: 'Visual assets, product photos, and useful reference images.',
        to: '/picture-warehouse',
        icon: 'lucide:images',
        size: 'compact',
        accent: 'rose',
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
        description: 'Common questions and quick answers for product, order, and service topics.',
        to: '/support/faqs',
        icon: 'lucide:circle-help',
        size: 'feature',
        accent: 'mint',
      },
      {
        id: 'payment',
        labelKey: 'support.nav.payment',
        labelFallback: 'Payment',
        description: 'Payment methods, checkout notes, processing, and purchase guidance.',
        to: '/support/payment',
        icon: 'lucide:credit-card',
        size: 'wide',
        accent: 'violet',
      },
      {
        id: 'shipping',
        labelKey: 'support.nav.shipping',
        labelFallback: 'Shipping',
        description: 'Shipping policy, delivery expectations, tracking, and logistics notes.',
        to: '/support/shipping',
        icon: 'lucide:truck',
        size: 'feature',
        accent: 'amber',
      },
      {
        id: 'warranty',
        labelKey: 'support.nav.warranty',
        labelFallback: 'Warranty',
        description: 'Warranty coverage, claim flow, and service rules.',
        to: '/support/warranty',
        icon: 'lucide:shield-check',
        size: 'standard',
        accent: 'mint',
      },
      {
        id: 'warranty-check',
        labelKey: 'support.nav.warrantyCheck',
        labelFallback: 'Warranty Check',
        description: 'Check product serial number and warranty status.',
        to: '/support/warranty-check',
        icon: 'lucide:badge-check',
        size: 'wide',
        accent: 'blue',
      },
      {
        id: 'product-feedback',
        labelKey: 'support.nav.productFeedback',
        labelFallback: 'Product Feedback',
        description: 'Send product feedback and riding experience notes.',
        to: '/support/product-feedback',
        icon: 'lucide:message-square-text',
        size: 'compact',
        accent: 'rose',
      },
      {
        id: 'test-report',
        labelKey: 'support.nav.testReport',
        labelFallback: 'Test Report',
        description: 'Rim and wheelset test reports, safety references, and certificates.',
        to: '/support/test-report',
        icon: 'lucide:file-check-2',
        size: 'compact',
        accent: 'slate',
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
        description: 'Brand story, factory capability, and the path behind Tanzanite products.',
        to: '/company/ourstory',
        icon: 'lucide:sparkles',
        size: 'feature',
        accent: 'mint',
      },
      {
        id: 'about',
        labelKey: 'company.nav.about',
        labelFallback: 'About',
        description: 'A short company introduction and entry point for visitors.',
        to: '/company/about',
        icon: 'lucide:landmark',
        size: 'wide',
        accent: 'blue',
      },
      {
        id: 'global-partners',
        labelKey: 'company.nav.globalPartners',
        labelFallback: 'Global Partners',
        description: 'Global cooperation, partner resources, and regional relationships.',
        to: '/company/global-partners',
        icon: 'lucide:globe-2',
        size: 'wide',
        accent: 'amber',
      },
      {
        id: 'oem-odm',
        labelKey: 'company.nav.oemOdm',
        labelFallback: 'OEM / ODM',
        description: 'Custom manufacturing, private label, and wheel project cooperation.',
        to: '/company/oem-odm',
        icon: 'lucide:factory',
        size: 'feature',
        accent: 'rose',
      },
      {
        id: 'certificates',
        labelKey: 'company.nav.certificates',
        labelFallback: 'Certificates',
        description: 'Company certificates and qualification references.',
        to: '/company/certificates',
        icon: 'lucide:award',
        size: 'standard',
        accent: 'blue',
      },
      {
        id: 'contact',
        labelKey: 'company.nav.contact',
        labelFallback: 'Contact',
        description: 'Reach Tanzanite for sales, service, and business communication.',
        to: '/company/contact',
        icon: 'lucide:mail',
        size: 'wide',
        accent: 'mint',
      },
    ],
  },
  {
    id: 'guides',
    labelKey: 'breadcrumbs.guides',
    labelFallback: 'Guides',
    routePrefixes: ['/guides', '/blog'],
    cards: [
      {
        id: 'tire-guides',
        labelKey: 'products.nav.tireSizeCharts',
        labelFallback: 'Tire Guides',
        description: 'Tire sizing, pressure, inner tube, tubeless, and installation topics.',
        to: '/guides/tireguides',
        icon: 'lucide:ruler',
        size: 'feature',
        accent: 'blue',
      },
      {
        id: 'wheelset-buyers',
        labelKey: 'products.nav.wheelsetBuyersGuide',
        labelFallback: 'Wheelset Buyers Guide',
        description: 'How to think through wheel selection, components, and riding requirements.',
        to: '/guides/wheelset-buyers',
        icon: 'lucide:map',
        size: 'wide',
        accent: 'violet',
      },
      {
        id: 'blog',
        labelKey: 'blog.nav.all',
        labelFallback: 'Blog',
        title: 'Wheelbuild blog',
        description: 'Articles, updates, and long-form wheelbuilding knowledge.',
        to: '/blog',
        icon: 'lucide:newspaper',
        size: 'wide',
        accent: 'blue',
      },
    ],
  },
]
