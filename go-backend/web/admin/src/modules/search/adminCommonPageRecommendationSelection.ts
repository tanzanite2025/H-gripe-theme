import type { AdminPageSearchCatalogEntry } from './adminPermissionScopedPageSearchCatalog'

export const adminCommonPageRecommendationIds = [
  'dashboard',
  'orders-list',
  'catalog-products',
  'media-library',
  'content-blog',
  'shipping-tracking-shipments',
  'marketing-coupons',
  'support-analytics',
  'settings-markets',
  'settings-api',
] as const

export function selectPermittedCommonAdminPageRecommendations(catalog: AdminPageSearchCatalogEntry[]): AdminPageSearchCatalogEntry[] {
  const availablePagesById = new Map(catalog.map((entry) => [entry.id, entry]))
  return adminCommonPageRecommendationIds.flatMap((id) => {
    const entry = availablePagesById.get(id)
    return entry ? [entry] : []
  })
}
