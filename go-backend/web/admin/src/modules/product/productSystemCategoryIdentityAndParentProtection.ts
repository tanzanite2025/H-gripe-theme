import { carbonRimSystemCategorySlug } from './carbonRimRequiredSystemCategorySelection'

export const isProtectedSystemProductCategory = (slug: string): boolean => (
  ['wheelset', 'wheel-components', 'tire', 'spokes', 'spoke-repair-kits', carbonRimSystemCategorySlug].includes(slug)
)

export const hasProtectedSystemProductCategoryParent = (slug: string): boolean => (
  isProtectedSystemProductCategory(slug) && slug !== 'wheelset'
)
