import type { PageSubNavigationEntry, PageSubNavigationTab } from './pageSubNavigationData.js'
import { normalizePrimaryMegaNavPath } from './primaryMegaNav.js'

const routePathFromTo = (to: string) => to.split('?')[0] || '/'

const pageSubNavigationChildPath = (basePath: string, tabId: string) => {
  const normalizedBasePath = routePathFromTo(basePath).replace(/\/+$/, '') || '/'
  return normalizedBasePath === '/' ? `/${tabId}` : `${normalizedBasePath}/${tabId}`
}

const getTabIdFromExactPath = (
  entry: PageSubNavigationEntry,
  path: string,
  localeCodes: string[],
) => {
  const normalizedPath = normalizePrimaryMegaNavPath(routePathFromTo(path), localeCodes)

  return entry.tabs.find((tab) => {
    const tabPath = tab.to || pageSubNavigationChildPath(entry.path, tab.id)
    return normalizePrimaryMegaNavPath(routePathFromTo(tabPath), localeCodes) === normalizedPath
  }) || null
}

export type PageSubNavigationBreadcrumbMatch =
  | { kind: 'canonical'; entry: PageSubNavigationEntry }
  | { kind: 'tab'; entry: PageSubNavigationEntry; tab: PageSubNavigationTab }

/**
 * Resolves breadcrumb ownership with an explicit canonical/tab split.
 * Canonical paths belong to their parent page list; only an exact child-tab
 * path can open the owning page's internal tab navigation.
 */
export const resolvePageSubNavigationBreadcrumb = (
  path: string,
  entries: readonly PageSubNavigationEntry[],
  localeCodes: string[] = [],
): PageSubNavigationBreadcrumbMatch | null => {
  const normalizedPath = normalizePrimaryMegaNavPath(routePathFromTo(path), localeCodes)

  const canonicalEntry = entries.find((entry) => (
    normalizePrimaryMegaNavPath(routePathFromTo(entry.path), localeCodes) === normalizedPath
  ))
  if (canonicalEntry) return { kind: 'canonical', entry: canonicalEntry }

  for (const entry of entries) {
    const tab = getTabIdFromExactPath(entry, normalizedPath, localeCodes)
    if (tab) return { kind: 'tab', entry, tab }
  }

  return null
}
