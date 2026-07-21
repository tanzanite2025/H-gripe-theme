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
  eyebrow: string
  description: string
  overviewTo: string
  overviewLabel: string
  cards: PrimaryMegaNavCard[]
}

/**
 * Desktop header mega-menu data.
 *
 * Keep this separate from page-level horizontal tab navigation:
 * - ProductsTopNav / SupportTopNav own in-page context switching.
 * - This file owns the top header's click-to-open card menu.
 */
export const primaryMegaNavSections: PrimaryMegaNavSection[] = [
  {
    id: 'products',
    labelKey: 'footer.menus.products',
    labelFallback: 'Products',
    eyebrow: 'Products & tools',
    description: 'Shop wheel products, compare technical guides, and jump into product-related utilities.',
    overviewTo: '/products',
    overviewLabel: 'Open products hub',
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
        id: 'tire-size-charts',
        labelKey: 'products.nav.tireSizeCharts',
        labelFallback: 'Tire Guides',
        description: 'Tire size, pressure, tubeless setup, installation, and compatibility references.',
        to: '/guides/tireguides',
        icon: 'lucide:ruler',
        size: 'feature',
        accent: 'violet',
      },
      {
        id: 'wheelset-buyers',
        labelKey: 'products.nav.wheelsetBuyersGuide',
        labelFallback: 'Wheelset Buyers Guide',
        description: 'Selection notes for riders choosing wheelsets, hubs, rims, spokes, and specs.',
        to: '/guides/wheelset-buyers',
        icon: 'lucide:route',
        size: 'wide',
        accent: 'amber',
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
        id: 'test-report',
        labelKey: 'support.nav.testReport',
        labelFallback: 'Test Report',
        description: 'Technical reports and test references for product confidence.',
        to: '/support/test-report',
        icon: 'lucide:clipboard-check',
        size: 'wide',
        accent: 'blue',
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
    eyebrow: 'Help center',
    description: 'Customer support, order help, warranty, shipping, payment, and feedback routes.',
    overviewTo: '/support',
    overviewLabel: 'Open support hub',
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
    eyebrow: 'About Tanzanite',
    description: 'Company story, capability, partners, certificates, and direct contact routes.',
    overviewTo: '/company',
    overviewLabel: 'Open company hub',
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
    eyebrow: 'Guides & knowledge',
    description: 'Learning content, technical references, buying guides, and wheelbuild articles.',
    overviewTo: '/guides',
    overviewLabel: 'Open guides hub',
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
        id: 'wheel-components',
        labelKey: 'products.nav.technicalDocs',
        labelFallback: 'Technical Docs',
        title: 'Wheel components',
        description: 'Jump directly to hubs, rims, spokes, nipples, and build references.',
        to: '/guides/wheelset-buyers#wheel-components',
        icon: 'lucide:boxes',
        size: 'wide',
        accent: 'amber',
      },
      {
        id: 'installation',
        labelKey: 'products.nav.tireSizeCharts',
        labelFallback: 'Installation',
        title: 'Installation',
        description: 'Tire and tubeless setup workflow from the tire guide.',
        to: '/guides/tireguides#installation',
        icon: 'lucide:wrench',
        size: 'wide',
        accent: 'rose',
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
      {
        id: 'news',
        labelKey: 'blog.nav.news',
        labelFallback: 'News',
        description: 'News and updates from Tanzanite.',
        to: '/blog/news',
        icon: 'lucide:radio',
        size: 'compact',
        accent: 'slate',
      },
      {
        id: 'wheelbuild',
        labelKey: 'blog.nav.wheelsbuild',
        labelFallback: 'Wheelbuild',
        description: 'Wheelbuild articles and technical stories.',
        to: '/blog/wheelsbuild',
        icon: 'lucide:pen-tool',
        size: 'compact',
        accent: 'mint',
      },
    ],
  },
]
