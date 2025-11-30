export interface ProductsNavItem {
  /** stable id for this products nav entry */
  id: string
  /** i18n key for the label */
  labelKey: string
  /** route path; will be passed through localePath() when rendered */
  to: string
}

/**
 * Top-level Products navigation definition used by the products layout.
 *
 * Each item corresponds to a concrete products / guides / blog page route.
 */
export const productsNavItems: ProductsNavItem[] = [
  {
    id: 'shop',
    labelKey: 'products.nav.shop',
    to: '/shop',
  },
  {
    id: 'about-tools',
    labelKey: 'products.nav.aboutTools',
    to: '/guides/tools',
  },
  {
    id: 'tire-size-charts',
    labelKey: 'products.nav.tireSizeCharts',
    to: '/guides/sizecharts',
  },
  {
    id: 'wheelsbuild-blog',
    labelKey: 'products.nav.wheelsbuildBlog',
    to: '/wheelsbuild',
  },
  {
    id: 'technical-docs',
    labelKey: 'products.nav.technicalDocs',
    to: '/guides/technical',
  },
  {
    id: 'wheelset-buyers',
    labelKey: 'products.nav.wheelsetBuyersGuide',
    to: '/guides/wheelset-buyers',
  },
  {
    id: 'spoke-calculator',
    labelKey: 'support.nav.spokeCalculator',
    to: '/spoke-calculator',
  },
]
