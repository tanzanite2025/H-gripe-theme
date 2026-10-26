import type { AdminNavigationItem } from '@/lib/adminNavigation'
import type { AdminGlobalSearchResult, AdminGlobalSearchSource } from './adminGlobalSearchResultAndSourceContracts'
import { buildPermissionScopedAdminPageSearchCatalog, filterAdminPageSearchCatalogByQuery, type AdminPageSearchCatalogEntry, type AdminPageSearchTranslator } from './adminPermissionScopedPageSearchCatalog'
import { selectPermittedCommonAdminPageRecommendations } from './adminCommonPageRecommendationSelection'

function convertAdminPageCatalogEntryToGlobalSearchResult(entry: AdminPageSearchCatalogEntry): AdminGlobalSearchResult {
  return {
    id: entry.id,
    sourceId: 'pages',
    title: entry.title,
    description: entry.description,
    categoryLabel: entry.moduleTitle,
    linkLabel: entry.path,
    routeTarget: { name: entry.routeName },
    icon: entry.icon,
  }
}

export function createAdminNavigationPageSearchSource(
  permittedNavigationItems: AdminNavigationItem[],
  translate: AdminPageSearchTranslator,
): AdminGlobalSearchSource {
  const catalog = buildPermissionScopedAdminPageSearchCatalog(permittedNavigationItems, translate)
  return {
    id: 'pages',
    label: translate('adminPageSearch.pages'),
    recommendationsLabel: translate('adminPageSearch.commonPages'),
    getRecommendations: () => selectPermittedCommonAdminPageRecommendations(catalog).map(convertAdminPageCatalogEntryToGlobalSearchResult),
    search: (query) => filterAdminPageSearchCatalogByQuery(catalog, query).map(convertAdminPageCatalogEntryToGlobalSearchResult),
  }
}
