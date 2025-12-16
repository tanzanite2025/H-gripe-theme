import type { ProductsNavItem } from '~/utils/productsNav'

/**
 * Top-level Company navigation definition used when viewing /company/* pages.
 * Reuses the same item shape as productsNavItems.
 */
export const companyNavItems: ProductsNavItem[] = [
  {
    id: 'about',
    labelKey: 'company.nav.about',
    to: '/company/about',
  },
  {
    id: 'our-story',
    labelKey: 'company.nav.ourStory',
    to: '/company/ourstory',
  },
  {
    id: 'global-partners',
    labelKey: 'company.nav.globalPartners',
    to: '/company/global-partners',
  },
  {
    id: 'oem-odm',
    labelKey: 'company.nav.oemOdm',
    to: '/company/oem-odm',
  },
  {
    id: 'certificates',
    labelKey: 'company.nav.certificates',
    to: '/company/certificates',
  },
  {
    id: 'contact',
    labelKey: 'company.nav.contact',
    to: '/company/contact',
  },
]
