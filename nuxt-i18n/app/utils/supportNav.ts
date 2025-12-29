export interface SupportNavItem {
  /** stable id for this support nav entry */
  id: string
  /** i18n key for the label (used in nav bar) */
  labelKey: string
  /** route path; will be passed through localePath() when rendered */
  to: string
  /** Icon character or name for the dashboard card */
  icon: string
  /** English fallback title for the dashboard card */
  title: string
  /** Short description for the dashboard card */
  description: string
  /** CTA text for the dashboard card */
  cta: string
}

/**
 * Top-level Support navigation definition used by the support layout and dashboard index.
 * Only `labelKey` uses i18n for now; other fields use English fallbacks.
 */
export const supportNavItems: SupportNavItem[] = [
  {
    id: 'faqs',
    labelKey: 'support.nav.faqs',
    to: '/support/faqs',
    icon: 'Q',
    title: "All FAQ'S",
    description: 'Browse common questions and quick answers about orders, products, and service.',
    cta: 'Browse FAQs',
  },
  {
    id: 'payment',
    labelKey: 'support.nav.payment',
    to: '/support/payment',
    icon: 'P',
    title: 'Payment',
    description: 'Available payment methods, processing times, and basic checkout tips.',
    cta: 'View payment info',
  },
  {
    id: 'shipping',
    labelKey: 'support.nav.shipping',
    to: '/support/shipping',
    icon: 'S',
    title: 'Shipping',
    description: 'Track your order, view shipping rates, delivery times, and global shipping policies.',
    cta: 'Check shipping details',
  },
  {
    id: 'after-sales',
    labelKey: 'support.nav.afterSales',
    to: '/support/after-sales',
    icon: 'A',
    title: 'After sales',
    description: 'Returns, exchanges, repairs, and other after-sales support options.',
    cta: 'View after-sales options',
  },
  {
    id: 'warranty',
    labelKey: 'support.nav.warranty',
    to: '/support/warranty',
    icon: 'W',
    title: 'Warranty',
    description: 'Overview of Tanzanite warranty coverage, terms, and claim process.',
    cta: 'Read warranty details',
  },
  {
    id: 'warranty-check',
    labelKey: 'support.nav.warrantyCheck',
    to: '/support/warranty-check',
    icon: 'C',
    title: 'Warranty Check',
    description: 'Verify the authenticity and warranty status of your specific product serial number.',
    cta: 'Check serial number',
  },
  {
    id: 'product-feedback',
    labelKey: 'support.nav.productFeedback',
    to: '/support/product-feedback',
    icon: 'F',
    title: 'Product Feedback',
    description: 'Share feedback about products and your riding experience with Tanzanite.',
    cta: 'Give feedback',
  },
  {
    id: 'test-report',
    labelKey: 'support.nav.testReport',
    to: '/support/test-report',
    icon: 'T',
    title: 'Test Report',
    description: 'View technical test reports and safety certifications for our wheelsets and rims.',
    cta: 'View test reports',
  },
]
