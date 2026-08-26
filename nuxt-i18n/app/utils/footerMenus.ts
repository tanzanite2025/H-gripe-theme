import type { RouteRecordNormalized } from 'vue-router'
import localeManifest from '~/i18n/locales.manifest'

export interface FooterLink {
  /** Optional i18n key declared by the route page metadata. */
  labelKey?: string
  /** Fallback label used when the route has no translation metadata yet. */
  fallback?: string
  /** Canonical route path passed to localePath(). */
  to: string
  /** Render as external <a> instead of NuxtLink when true. */
  external?: boolean
}

export interface FooterSection {
  /** Stable id for this column, e.g. 'resources' or 'siteOverview'. */
  id: string
  /** i18n key for the column title. */
  titleKey: string
  /** Fallback title used when a locale has not translated the key yet. */
  fallback?: string
  links: FooterLink[]
}

interface FooterRouteSectionDefinition {
  id: string
  path: string
  titleKey: string
  titleFallback: string
  includeRoot?: boolean
}

type FooterRouteMeta = {
  footer?: boolean
  footerLabelKey?: string
  footerLabelFallback?: string
}

const footerRouteSections: readonly FooterRouteSectionDefinition[] = [
  {
    id: 'shop',
    path: '/shop',
    titleKey: 'products.nav.shop',
    titleFallback: 'Shop',
  },
  {
    id: 'resources',
    path: '/resources',
    titleKey: 'footer.menus.resources',
    titleFallback: 'Resources',
  },
  {
    id: 'support',
    path: '/support',
    titleKey: 'footer.menus.support',
    titleFallback: 'Support',
  },
  {
    id: 'company',
    path: '/company',
    titleKey: 'footer.menus.company',
    titleFallback: 'Company',
  },
  {
    id: 'guides',
    path: '/guides',
    titleKey: 'breadcrumbs.guides',
    titleFallback: 'Guides',
  },
  {
    id: 'policies',
    path: '/policies',
    titleKey: 'footer.menus.policies',
    titleFallback: 'Policies',
  },
  {
    id: 'siteOverview',
    path: '/website',
    titleKey: 'footer.menus.siteOverview',
    titleFallback: 'Site Overview',
  },
]

const localeCodes = localeManifest.map(locale => locale.code)

const normalizeRoutePath = (path: string) => {
  const pathWithoutQuery = (path || '/').split(/[?#]/)[0] || '/'
  const absolutePath = pathWithoutQuery.startsWith('/')
    ? pathWithoutQuery
    : `/${pathWithoutQuery}`
  const segments = absolutePath.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''

  if (
    localeCodes.includes(firstSegment) ||
    firstSegment === ':locale' ||
    firstSegment.startsWith(':locale')
  ) {
    segments.shift()
  }

  return `/${segments.join('/')}`.replace(/\/+$/, '') || '/'
}

const isDynamicRouteSegment = (segment: string) => {
  return (
    segment.startsWith(':') ||
    segment.startsWith('[') ||
    segment.includes('*') ||
    segment.includes('(')
  )
}

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

const routeLabel = (route: RouteRecordNormalized, path: string): FooterLink => {
  const meta = (route.meta || {}) as FooterRouteMeta
  const segment = path.split('/').filter(Boolean).at(-1) || ''

  return {
    ...(meta.footerLabelKey ? { labelKey: meta.footerLabelKey } : {}),
    fallback: meta.footerLabelFallback || humanizeRouteSegment(segment),
    to: path,
  }
}

const directStaticChildRoutes = (
  routes: readonly RouteRecordNormalized[],
  section: FooterRouteSectionDefinition
) => {
  const normalizedSectionPath = normalizeRoutePath(section.path)
  const sectionSegments = normalizedSectionPath.split('/').filter(Boolean)
  const seenPaths = new Set<string>()

  return routes
    .map((route, order) => ({
      route,
      order,
      path: normalizeRoutePath(route.path),
    }))
    .filter(({ route, path }) => {
      const meta = (route.meta || {}) as FooterRouteMeta
      const segments = path.split('/').filter(Boolean)

      return (
        meta.footer !== false &&
        Boolean(route.name) &&
        !route.redirect &&
        (section.includeRoot
          ? path === normalizedSectionPath || segments.length === sectionSegments.length + 1
          : path !== normalizedSectionPath && segments.length === sectionSegments.length + 1) &&
        segments.slice(0, sectionSegments.length).every(
          (segment, index) => segment === sectionSegments[index]
        ) &&
        !segments.some(isDynamicRouteSegment)
      )
    })
    .filter(({ path }) => {
      if (seenPaths.has(path)) return false
      seenPaths.add(path)
      return true
    })
    .sort((left, right) => left.order - right.order)
    .map(({ route, path }) => routeLabel(route, path))
}

/**
 * Build footer columns from the router's direct child pages.
 *
 * The header owns presentation cards; the footer owns route discovery. Keeping
 * these sources separate prevents a root page such as /shop from becoming a
 * duplicate "Shop" child link under the SHOP column.
 */
export const createFooterMenusFromRoutes = (
  routes: readonly RouteRecordNormalized[]
): FooterSection[] => {
  const routeSections = footerRouteSections
    .map(section => ({
      id: section.id,
      titleKey: section.titleKey,
      fallback: section.titleFallback,
      links: directStaticChildRoutes(routes, section),
    }))
    .filter(section => section.links.length > 0)

  return routeSections
}
