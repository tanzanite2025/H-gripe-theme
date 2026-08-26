import type {
  PageSubNavigationChild,
  PageSubNavigationEntry,
  PageSubNavigationPathMatchMode,
  PageSubNavigationTab,
  PageSubNavigationTabFromPathOptions,
} from './pageSubNavigationData'
import {
  virtualPageSubNavigationEntries,
} from './pageSubNavigationData'
import {
  normalizePrimaryMegaNavPath,
  primaryMegaNavPathMatches,
  primaryMegaNavSections,
  type PrimaryMegaNavCard,
  type PrimaryMegaNavSection,
} from '~/utils/primaryMegaNav'
import localeManifest from '~/i18n/locales.manifest'

export * from './pageSubNavigationData'

const defaultPageSubNavigationLocaleCodes = localeManifest.map(locale => locale.code)

const routePathFromTo = (to: string) => {
  return to.split('?')[0] || '/'
}

export const pageSubNavigationChildPath = (basePath: string, tabId: string) => {
  const normalizedBasePath = routePathFromTo(basePath).replace(/\/+$/, '') || '/'
  return normalizedBasePath === '/' ? `/${tabId}` : `${normalizedBasePath}/${tabId}`
}

/**
 * Vite exposes page files as a static module map. The map keeps the build
 * aware of new page routes without importing page components at runtime.
 */
const pageModulePaths = Object.keys(
  import.meta.glob('../pages/**/*.vue')
)

const pagePathFromModule = (modulePath: string) => {
  const normalizedModulePath = modulePath.replace(/\\/g, '/')
  const pagesMarker = '/pages/'
  const markerIndex = normalizedModulePath.indexOf(pagesMarker)
  if (markerIndex < 0) return null

  const relativePagePath = normalizedModulePath.slice(markerIndex + pagesMarker.length)
  const withoutExtension = relativePagePath.replace(/\.vue$/, '')
  const routePath = withoutExtension === 'index'
    ? ''
    : withoutExtension.replace(/\/index$/, '')

  if (!routePath) return '/'

  const segments = routePath.split('/').filter(Boolean)
  if (segments.some(segment => segment.startsWith('['))) {
    return null
  }

  return `/${segments.join('/')}`
}

const discoveredPagePaths = [...new Set(
  pageModulePaths
    .map(pagePathFromModule)
    .filter((path): path is string => Boolean(path))
)]

const humanizeRouteSegment = (segment: string) => {
  let decodedSegment = segment
  try {
    decodedSegment = decodeURIComponent(segment)
  } catch {
    // Keep the route segment when decoding is not possible.
  }

  return decodedSegment
    .replace(/[-_]+/g, ' ')
    .replace(/\b\w/g, character => character.toUpperCase())
}

const getDiscoveredChildTabs = (card: PrimaryMegaNavCard): PageSubNavigationTab[] => {
  if (card.discoverChildPages === false) return []

  const cardPath = normalizePrimaryMegaNavPath(routePathFromTo(card.to))
  const cardSegments = cardPath.split('/').filter(Boolean)

  return discoveredPagePaths
    .filter((pagePath) => {
      const pageSegments = pagePath.split('/').filter(Boolean)
      return (
        pageSegments.length === cardSegments.length + 1 &&
        primaryMegaNavPathMatches(pagePath, cardPath)
      )
    })
    .sort((left, right) => left.localeCompare(right))
    .map((pagePath) => {
      const segment = pagePath.split('/').filter(Boolean).at(-1) || ''
      const labelKey = card.childLabelNamespace
        ? `${card.childLabelNamespace}.${segment}`
        : undefined

      return {
        id: segment,
        ...(labelKey ? { labelKey } : {}),
        fallback: humanizeRouteSegment(segment),
        to: pagePath,
      }
    })
}

const discoveredPageSubNavigationEntries: PageSubNavigationEntry[] =
  primaryMegaNavSections.flatMap((section) =>
    section.cards.flatMap((card) => {
      const tabs = getDiscoveredChildTabs(card)
      return tabs.length > 0
        ? [{ path: routePathFromTo(card.to), tabs }]
        : []
    })
  )

const tabTargetPath = (entry: PageSubNavigationEntry, tab: PageSubNavigationTab) => {
  return normalizePrimaryMegaNavPath(
    routePathFromTo(tab.to || pageSubNavigationChildPath(entry.path, tab.id))
  )
}

const mergePageSubNavigationEntries = (
  entries: readonly PageSubNavigationEntry[]
): PageSubNavigationEntry[] => {
  const mergedByPath = new Map<string, PageSubNavigationEntry>()

  for (const entry of entries) {
    const entryPath = normalizePrimaryMegaNavPath(routePathFromTo(entry.path))
    const current = mergedByPath.get(entryPath)

    if (!current) {
      mergedByPath.set(entryPath, {
        path: entry.path,
        tabs: [...entry.tabs],
      })
      continue
    }

    const tabs = [...current.tabs]
    for (const tab of entry.tabs) {
      const targetPath = tabTargetPath(entry, tab)
      if (!tabs.some(existingTab => tabTargetPath(current, existingTab) === targetPath)) {
        tabs.push(tab)
      }
    }

    mergedByPath.set(entryPath, { ...current, tabs })
  }

  return [...mergedByPath.values()]
}

/**
 * The shared page-navigation registry contains both:
 * - virtual tabs generated from a base page by Nuxt's pages:extend hook
 * - real child index pages discovered from app/pages at build time
 */
export const pageSubNavigationEntries = mergePageSubNavigationEntries([
  ...virtualPageSubNavigationEntries,
  ...discoveredPageSubNavigationEntries,
])

export const isPageSubNavigationTabId = <Tabs extends readonly PageSubNavigationTab[]>(
  tabs: Tabs,
  id: string
): id is Tabs[number]['id'] => tabs.some((tab) => tab.id === id)

export const getPageSubNavigationTabFromPath = <Tabs extends readonly PageSubNavigationTab[]>(
  tabs: Tabs,
  basePath: string,
  path: string | null | undefined,
  options: PageSubNavigationTabFromPathOptions = {}
): Tabs[number]['id'] | null => {
  const localeCodes = options.localeCodes || defaultPageSubNavigationLocaleCodes
  const match = options.match || 'nested'
  const normalizedPath = normalizePrimaryMegaNavPath(
    routePathFromTo(String(path || '/')),
    localeCodes
  )

  for (const tab of tabs) {
    const tabPath = tab.to || pageSubNavigationChildPath(basePath, tab.id)
    const normalizedTabPath = normalizePrimaryMegaNavPath(
      routePathFromTo(tabPath),
      localeCodes
    )
    const matches = match === 'exact'
      ? normalizedPath === normalizedTabPath
      : primaryMegaNavPathMatches(normalizedPath, normalizedTabPath, localeCodes)

    if (matches) return tab.id
  }

  return null
}

const belongsToSection = (
  section: PrimaryMegaNavSection,
  path: string,
  localeCodes: string[] = []
) => {
  return section.routePrefixes.some((prefix) => primaryMegaNavPathMatches(path, prefix, localeCodes))
}

const belongsToCard = (
  card: PrimaryMegaNavCard,
  path: string,
  localeCodes: string[] = []
) => {
  const normalizedPath = normalizePrimaryMegaNavPath(path, localeCodes)
  const cardPath = normalizePrimaryMegaNavPath(routePathFromTo(card.to), localeCodes)

  return normalizedPath === cardPath || primaryMegaNavPathMatches(normalizedPath, cardPath, localeCodes)
}

export const getPageSubNavigationForPath = (
  path: string,
  localeCodes: string[] = []
) => {
  const normalizedPath = normalizePrimaryMegaNavPath(routePathFromTo(path), localeCodes)

  return pageSubNavigationEntries.find((entry) => {
    const entryPath = normalizePrimaryMegaNavPath(entry.path, localeCodes)

    return (
      entryPath === normalizedPath ||
      Boolean(getPageSubNavigationTabFromPath(entry.tabs, entry.path, normalizedPath, {
        localeCodes,
        match: 'nested',
      }))
    )
  }) || null
}

export const pageSubNavigationBelongsToCard = (
  section: PrimaryMegaNavSection,
  card: PrimaryMegaNavCard,
  entry: PageSubNavigationEntry,
  localeCodes: string[] = []
) => {
  const entryPath = normalizePrimaryMegaNavPath(entry.path, localeCodes)

  return belongsToSection(section, entryPath, localeCodes) && belongsToCard(card, entryPath, localeCodes)
}

export const getPrimaryMegaNavCardChildren = (
  section: PrimaryMegaNavSection,
  card: PrimaryMegaNavCard,
  localeCodes: string[] = []
): PageSubNavigationChild[] => {
  const children: PageSubNavigationChild[] = []

  for (const entry of pageSubNavigationEntries) {
    if (!pageSubNavigationBelongsToCard(section, card, entry, localeCodes)) continue

    for (const tab of entry.tabs) {
      const childPath = normalizePrimaryMegaNavPath(
        routePathFromTo(tab.to || pageSubNavigationChildPath(entry.path, tab.id)),
        localeCodes
      )
      if (!belongsToSection(section, childPath, localeCodes) || !belongsToCard(card, childPath, localeCodes)) {
        continue
      }

      if (!children.some(child => child.to === childPath)) {
        children.push({ ...tab, to: childPath })
      }
    }
  }

  return children
}
