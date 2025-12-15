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
      // Products column: product- and guide-related links
      { labelKey: 'footer.links.shop', to: '/shop' },
      { labelKey: 'footer.links.aboutTools', to: '/guides/tools' },
      { labelKey: 'footer.links.tireSizeCharts', to: '/guides/sizecharts' },
      { labelKey: 'footer.links.wheelsbuildBlog', to: '/blog' },
      { labelKey: 'footer.links.technicalDocs', to: '/guides/technical' },
      { labelKey: 'footer.links.wheelsetBuyersGuide', to: '/guides/wheelset-buyers' },
      { labelKey: 'footer.links.membershipPoints', to: '/membershipandpoints' },
      { labelKey: 'footer.links.pictureWarehouse', to: '/picture-warehouse' },
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
      { labelKey: 'footer.links.aboutUs', to: '/company/about' },
      { labelKey: 'footer.links.ourStory', to: '/company/ourstory' },
    ],
  },
  {
    id: 'resources',
    titleKey: 'footer.menus.resources',
    // You can fill links for this column later
    links: [],
  },
]
