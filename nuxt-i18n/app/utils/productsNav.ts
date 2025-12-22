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
    id: 'tire-size-charts',
    labelKey: 'products.nav.tireSizeCharts',
    to: '/guides/tireguides',
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
  {
    id: 'test-report',
    labelKey: 'support.nav.testReport',
    to: '/support/test-report',
  },
  {
    id: 'membership-and-points',
    labelKey: 'company.nav.membershipPoints',
    to: '/membershipandpoints',
  },
  {
    id: 'picture-warehouse',
    labelKey: 'company.nav.pictureWarehouse',
    to: '/picture-warehouse',
  },
]
