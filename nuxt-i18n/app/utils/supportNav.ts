export interface SupportNavItem {
  /** stable id for this support nav entry */
  id: string
  /** i18n key for the label */
  labelKey: string
  /** route path; will be passed through localePath() when rendered */
  to: string
}

/**
 * Top-level Support navigation definition used by the support layout.
 *
 * Each item should correspond to a concrete support page route so that
 * Google/Bing can index them individually.
 */
export const supportNavItems: SupportNavItem[] = [
  {
    id: 'faqs',
    labelKey: 'support.nav.faqs',
    to: '/support/faqs',
  },
  {
    id: 'payment',
    labelKey: 'support.nav.payment',
    to: '/support/payment',
  },
  {
    id: 'shipping',
    labelKey: 'support.nav.shipping',
    to: '/support/shipping',
  },
  {
    id: 'after-sales',
    labelKey: 'support.nav.afterSales',
    to: '/support/after-sales',
  },
  {
    id: 'warranty',
    labelKey: 'support.nav.warranty',
    to: '/support/warranty',
  },
  {
    id: 'warranty-check',
    labelKey: 'support.nav.warrantyCheck',
    to: '/support/warranty-check',
  },
  {
    id: 'product-feedback',
    labelKey: 'support.nav.productFeedback',
    to: '/support/product-feedback',
  },
  {
    id: 'test-report',
    labelKey: 'support.nav.testReport',
    to: '/support/test-report',
  },
]
