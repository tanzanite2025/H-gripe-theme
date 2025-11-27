export interface FooterLink {
  /** i18n key for the link label, e.g. 'footer.links.about' */
  labelKey: string
  /** route path or named route key; will be passed to localePath() */
  to: string
  /** render as external <a> instead of NuxtLink when true */
  external?: boolean
}

export interface FooterSection {
  /** stable id for this column, e.g. 'company', 'support' */
  id: string
  /** i18n key for the column title, e.g. 'footer.menus.company' */
  titleKey: string
  links: FooterLink[]
}

/**
 * Pure-i18n footer menu definition.
 *
 * - Structure (columns, links, targets) is owned by Nuxt.
 * - All visible text comes from nuxt-i18n via the labelKey/titleKey keys.
 * - There is no runtime dependency on WordPress menus.
 *
 * You can safely extend this array later with more columns / links without
 * changing the rendering component.
 */
export const footerMenus: FooterSection[] = [
  {
    id: 'products',
    titleKey: 'footer.menus.products',
    links: [
      { labelKey: 'footer.links.spokeCalculator', to: '/support/spoke-calculator' },
    ],
  },
  {
    id: 'support',
    titleKey: 'footer.menus.support',
    links: [
      { labelKey: 'footer.links.helpCenter', to: '/help-center' },
      { labelKey: 'footer.links.faq', to: '/support/faqs' },
    ],
  },
  {
    id: 'company',
    titleKey: 'footer.menus.company',
    links: [
      { labelKey: 'footer.links.about', to: '/about' },
      { labelKey: 'footer.links.contact', to: '/contact' },
    ],
  },
  {
    id: 'resources',
    titleKey: 'footer.menus.resources',
    // You can fill links for this column later
    links: [],
  },
]
