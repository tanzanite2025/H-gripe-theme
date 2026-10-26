import type { Component } from 'vue'
import type { AdminNavigationItem } from '@/lib/adminNavigation'
import { adminPageSearchDescriptionAndKeywordMetadata } from './adminPageSearchDescriptionAndKeywordMetadata'

export type AdminPageSearchTranslator = (
  key: string,
  params?: Record<string, string | number>,
  fallback?: string,
) => string

export interface AdminPageSearchCatalogEntry {
  id: string
  title: string
  moduleTitle: string
  description: string
  path: string
  routeName: string
  icon?: Component
  searchableText: string
}

// Callers supply the canonical navigation tree after permission filtering, before translation.
export function buildPermissionScopedAdminPageSearchCatalog(
  permittedNavigationItems: AdminNavigationItem[],
  translate: AdminPageSearchTranslator,
): AdminPageSearchCatalogEntry[] {
  const entries: AdminPageSearchCatalogEntry[] = []

  function collectPermittedAdminNavigationLeafPages(items: AdminNavigationItem[], ancestors: AdminNavigationItem[] = []): void {
    for (const item of items) {
      if (item.children?.length) {
        collectPermittedAdminNavigationLeafPages(item.children, [...ancestors, item])
        continue
      }
      if (!item.path || !item.routeName) continue

      const metadata = adminPageSearchDescriptionAndKeywordMetadata[item.id]
      const title = translate(`nav.${item.id}`, {}, item.label)
      const moduleTitle = ancestors.map((ancestor) => translate(`nav.${ancestor.id}`, {}, ancestor.label)).join(' / ')
      const description = metadata
        ? translate(`adminPageSearch.descriptions.${metadata.descriptionKey}`)
        : translate('adminPageSearch.defaultDescription', { title, module: moduleTitle || translate('adminPageSearch.pages') })
      entries.push({
        id: item.id,
        title,
        moduleTitle,
        description,
        path: item.path,
        routeName: item.routeName,
        icon: item.icon || ancestors[ancestors.length - 1]?.icon,
        searchableText: [title, moduleTitle, item.label, ...ancestors.map((ancestor) => ancestor.label), description, item.path, metadata?.keywords].join(' '),
      })
    }
  }

  collectPermittedAdminNavigationLeafPages(permittedNavigationItems)
  return entries
}

export function filterAdminPageSearchCatalogByQuery(entries: AdminPageSearchCatalogEntry[], query: string): AdminPageSearchCatalogEntry[] {
  const searchTerms = query.normalize('NFKC').toLowerCase().trim().split(/\s+/).filter(Boolean)
  if (!searchTerms.length) return entries

  return entries.filter((entry) => {
    const searchableText = entry.searchableText.normalize('NFKC').toLowerCase()
    return searchTerms.every((term) => searchableText.includes(term))
  })
}
