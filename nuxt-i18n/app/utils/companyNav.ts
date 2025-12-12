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
    id: 'membership-and-points',
    labelKey: 'company.nav.membershipPoints',
    to: '/company/membershipandpoints',
  },
  {
    id: 'picture-warehouse',
    labelKey: 'company.nav.pictureWarehouse',
    to: '/company/picture-warehouse',
  },
  {
    id: 'contact',
    labelKey: 'company.nav.contact',
    to: '/company/contact',
  },
]
