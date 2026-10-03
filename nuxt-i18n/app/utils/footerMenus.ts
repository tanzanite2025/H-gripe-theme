import type { RouteRecordNormalized } from 'vue-router'
import localeManifest from '../i18n/locales.manifest.js'
import { virtualPageSubNavigationEntries } from './pageSubNavigationData.js'

export interface FooterNavigationItemLabel {
  /** Optional i18n key declared by route metadata or a virtual tab. */
  labelKey?: string
  /** Fallback label used when the locale has not loaded the translation yet. */
  fallback?: string
}

/**
 * A footer node can be a link or a group heading without a landing page.
 * Children keep every nested route discoverable when a folder has no index.
 */
export interface FooterNavigationItem extends FooterNavigationItemLabel {
  /** Canonical route path passed to localePath(), when this node is clickable. */
  to?: string
  /** Render as an external <a> instead of NuxtLink when true. */
  external?: boolean
  /** Nested route or virtual-tab entries rendered below this node. */
  children?: FooterNavigationItem[]
}

/** A footer item that is guaranteed to be a clickable link. */
export interface FooterLink extends FooterNavigationItemLabel {
  to: string
  external?: boolean
}

export const isFooterLink = (item: FooterNavigationItem): item is FooterLink => (
  typeof item.to === 'string' && item.to.trim().length > 0
)

export interface FooterSection {
  /** Stable id for this column, e.g. 'resources' or 'siteOverview'. */
  id: string
  /** i18n key for the column title. */
  titleKey: string
  /** Fallback title used when the locale has not translated the key yet. */
  fallback?: string
  links: FooterNavigationItem[]
}

export interface FooterProductCategory {
  name?: string
  slug?: string
  routePath?: string
  children?: readonly FooterProductCategory[]
}

export interface FooterMenuRouteOptions {
  /** Runtime SHOP links supplied by the product-category API. */
  shopLinks?: readonly FooterNavigationItem[]
}

interface FooterRouteSectionDefinition {
  id: string
  path: string
  titleKey: string
  titleFallback: string
  /** Include the section landing page when the section is data-driven. */
  includeRoot?: boolean
}

type FooterRouteMeta = {
  footer?: boolean
  footerLabelKey?: string
  footerLabelFallback?: string
  footerGroupLabelKey?: string
  footerGroupLabelFallback?: string
}

interface StaticFooterRouteCandidate {
  route: RouteRecordNormalized
  path: string
  order: number
}

interface MutableFooterNavigationNode {
  path: string
  segment: string
  directRoute?: StaticFooterRouteCandidate
  groupLabelRoute?: StaticFooterRouteCandidate
  virtualLink?: FooterLink
  virtualOrder?: number
  order: number
  children: Map<string, MutableFooterNavigationNode>
}

const footerRouteSections: readonly FooterRouteSectionDefinition[] = [
  {
    id: 'shop',
    path: '/shop',
    titleKey: 'products.nav.shop',
    titleFallback: 'Shop',
    includeRoot: true,
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
    : '/' + pathWithoutQuery
  const segments = absolutePath.split('/').filter(Boolean)
  const firstSegment = segments[0] || ''

  if (
    localeCodes.includes(firstSegment) ||
    firstSegment === ':locale' ||
    firstSegment.startsWith(':locale')
  ) {
    segments.shift()
  }

  return '/' + segments.join('/').replace(/\/+$/, '') || '/'
}

const isDynamicRouteSegment = (segment: string) => (
  segment.startsWith(':') ||
  segment.startsWith('[') ||
  segment.includes('*') ||
  segment.includes('(')
)

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

const humanizeRouteGroupSegment = (segment: string) => {
  const withGuideSuffixSpacing = segment.replace(
    /^(.*?)(guides|tools)$/i,
    '$1 $2',
  )

  return humanizeRouteSegment(withGuideSuffixSpacing)
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

const getStaticFooterRouteCandidates = (
  routes: readonly RouteRecordNormalized[],
  section: FooterRouteSectionDefinition,
): StaticFooterRouteCandidate[] => {
  const normalizedSectionPath = normalizeRoutePath(section.path)
  const sectionSegments = normalizedSectionPath.split('/').filter(Boolean)
  const candidatesByPath = new Map<string, StaticFooterRouteCandidate>()

  routes.forEach((route, order) => {
    const meta = (route.meta || {}) as FooterRouteMeta
    const path = normalizeRoutePath(route.path)
    const segments = path.split('/').filter(Boolean)
    const isSectionRoot = path === normalizedSectionPath
    const isDescendantOfSection = segments.length > sectionSegments.length &&
      sectionSegments.every((segment, index) => segments[index] === segment)

    if (
      meta.footer === false ||
      !route.name ||
      route.redirect ||
      !((section.includeRoot && isSectionRoot) || isDescendantOfSection) ||
      segments.some(isDynamicRouteSegment)
    ) {
      return
    }

    const candidate = { route, path, order }
    const previous = candidatesByPath.get(path)
    if (!previous || order < previous.order) {
      candidatesByPath.set(path, candidate)
    }
  })

  return [...candidatesByPath.values()].sort((left, right) => left.order - right.order)
}

/** Keep non-GUIDES columns compact at their first route level. */
const createFlatFooterSectionLinks = (
  candidates: readonly StaticFooterRouteCandidate[],
  section: FooterRouteSectionDefinition,
): FooterLink[] => {
  const sectionSegments = normalizeRoutePath(section.path).split('/').filter(Boolean)
  const routeGroups = new Map<string, {
    route: RouteRecordNormalized
    path: string
    order: number
    hasDirectRoute: boolean
  }>()

  for (const candidate of candidates) {
    const segments = candidate.path.split('/').filter(Boolean)
    const groupPath = '/' + segments.slice(0, sectionSegments.length + 1).join('/')
    const hasDirectRoute = candidate.path === groupPath
    const previous = routeGroups.get(groupPath)

    if (
      !previous ||
      (hasDirectRoute && !previous.hasDirectRoute) ||
      (hasDirectRoute === previous.hasDirectRoute && candidate.order < previous.order)
    ) {
      routeGroups.set(groupPath, {
        route: candidate.route,
        path: candidate.path,
        order: candidate.order,
        hasDirectRoute,
      })
    }
  }

  return [...routeGroups.entries()]
    .sort(([, left], [, right]) => left.order - right.order)
    .map(([groupPath, group]) => {
      if (group.hasDirectRoute) return routeLabel(group.route, groupPath)

      const meta = (group.route.meta || {}) as FooterRouteMeta
      const groupSegment = groupPath.split('/').filter(Boolean).at(-1) || ''
      return {
        ...(meta.footerGroupLabelKey ? { labelKey: meta.footerGroupLabelKey } : {}),
        fallback: meta.footerGroupLabelFallback || humanizeRouteGroupSegment(groupSegment),
        to: group.path,
      }
    })
}

const createEmptyFooterNavigationNode = (
  path: string,
  segment: string,
): MutableFooterNavigationNode => ({
  path,
  segment,
  order: Number.POSITIVE_INFINITY,
  children: new Map(),
})

const isPathWithinSection = (path: string, sectionPath: string) => {
  const pathSegments = normalizeRoutePath(path).split('/').filter(Boolean)
  const sectionSegments = normalizeRoutePath(sectionPath).split('/').filter(Boolean)

  return sectionSegments.every((segment, index) => pathSegments[index] === segment) &&
    pathSegments.length > sectionSegments.length
}

const ensureFooterNavigationNodePath = (
  root: MutableFooterNavigationNode,
  path: string,
  sectionPath: string,
) => {
  const pathSegments = normalizeRoutePath(path).split('/').filter(Boolean)
  const sectionSegments = normalizeRoutePath(sectionPath).split('/').filter(Boolean)
  const relativeSegments = pathSegments.slice(sectionSegments.length)
  let currentNode = root

  for (let index = 0; index < relativeSegments.length; index += 1) {
    const segment = relativeSegments[index]
    if (!segment) continue
    const nodePath = '/' + pathSegments.slice(0, sectionSegments.length + index + 1).join('/')
    let childNode = currentNode.children.get(segment)
    if (!childNode) {
      childNode = createEmptyFooterNavigationNode(nodePath, segment)
      currentNode.children.set(segment, childNode)
    }
    currentNode = childNode
  }

  return currentNode
}

const createVirtualFooterLink = (
  tab: { labelKey?: string; label?: string; fallback?: string },
  targetPath: string,
): FooterLink => {
  const segment = targetPath.split('/').filter(Boolean).at(-1) || ''

  return {
    ...(tab.labelKey ? { labelKey: tab.labelKey } : {}),
    fallback: tab.fallback || tab.label || humanizeRouteSegment(segment),
    to: targetPath,
  }
}

const createGuidesNavigationTree = (
  routes: readonly RouteRecordNormalized[],
): FooterNavigationItem[] => {
  const section = footerRouteSections.find(sectionDefinition => sectionDefinition.id === 'guides')
  if (!section) return []

  const candidates = getStaticFooterRouteCandidates(routes, section)
  const candidatesByPath = new Map(candidates.map(candidate => [candidate.path, candidate]))
  const root = createEmptyFooterNavigationNode(normalizeRoutePath(section.path), '')
  const sectionSegments = normalizeRoutePath(section.path).split('/').filter(Boolean)

  for (const candidate of candidates) {
    const node = ensureFooterNavigationNodePath(root, candidate.path, section.path)
    node.order = Math.min(node.order, candidate.order)
    node.directRoute = candidate

    let groupNode = root
    const pathSegments = candidate.path.split('/').filter(Boolean)
    for (let index = 0; index < pathSegments.length - sectionSegments.length; index += 1) {
      const segment = pathSegments[index + sectionSegments.length]
      if (!segment) continue
      const childNode = groupNode.children.get(segment)
      if (!childNode) break
      groupNode = childNode
      if (!groupNode.groupLabelRoute || candidate.order < groupNode.groupLabelRoute.order) {
        groupNode.groupLabelRoute = candidate
      }
    }
  }

  for (const entry of virtualPageSubNavigationEntries) {
    if (!isPathWithinSection(entry.path, section.path)) continue

    const basePath = normalizeRoutePath(entry.path)
    ensureFooterNavigationNodePath(root, basePath, section.path)

    entry.tabs.forEach((tab, tabIndex) => {
      const explicitTabPath = 'to' in tab ? tab.to : undefined
      const targetPath = normalizeRoutePath(
        explicitTabPath || basePath + '/' + encodeURIComponent(tab.id),
      )
      if (!isPathWithinSection(targetPath, section.path)) return

      const node = ensureFooterNavigationNodePath(root, targetPath, section.path)
      node.virtualLink = createVirtualFooterLink(tab, targetPath)
      node.virtualOrder = tabIndex
      node.order = Math.min(node.order, tabIndex)

      const targetCandidate = candidatesByPath.get(targetPath)
      if (targetCandidate && !node.groupLabelRoute) {
        node.groupLabelRoute = targetCandidate
      }
    })
  }

  const sortNodes = (nodes: Iterable<MutableFooterNavigationNode>) => [...nodes].sort((left, right) => {
    if (left.virtualOrder !== undefined && right.virtualOrder !== undefined) {
      return left.virtualOrder - right.virtualOrder
    }
    if (left.virtualOrder !== undefined) return -1
    if (right.virtualOrder !== undefined) return 1
    return left.order - right.order || left.path.localeCompare(right.path)
  })

  const createNavigationItem = (node: MutableFooterNavigationNode): FooterNavigationItem => {
    const children = sortNodes(node.children.values()).map(createNavigationItem)
    const routeMeta = (node.groupLabelRoute?.route.meta || {}) as FooterRouteMeta
    const item: FooterNavigationItem = node.virtualLink
      ? { ...node.virtualLink }
      : node.directRoute
        ? routeLabel(node.directRoute.route, node.path)
        : {
            ...(routeMeta.footerGroupLabelKey ? { labelKey: routeMeta.footerGroupLabelKey } : {}),
            fallback: routeMeta.footerGroupLabelFallback || humanizeRouteGroupSegment(node.segment),
          }

    if (children.length) item.children = children
    return item
  }

  return sortNodes(root.children.values()).map(createNavigationItem)
}

/** Build all footer columns, with GUIDES kept as a complete route tree. */
export const createFooterMenusFromRoutes = (
  routes: readonly RouteRecordNormalized[],
  options: FooterMenuRouteOptions = {},
): FooterSection[] => {
  const routeSections = footerRouteSections
    .map(section => {
      if (section.id === 'guides') {
        return {
          id: section.id,
          titleKey: section.titleKey,
          fallback: section.titleFallback,
          links: createGuidesNavigationTree(routes),
        }
      }

      const candidates = getStaticFooterRouteCandidates(routes, section)
      return {
        id: section.id,
        titleKey: section.titleKey,
        fallback: section.titleFallback,
        links: section.id === 'shop' && options.shopLinks && options.shopLinks.length > 0
          ? [...options.shopLinks]
          : createFlatFooterSectionLinks(candidates, section),
      }
    })
    .filter(section => section.links.length > 0)

  return routeSections
}

const createFooterShopLinkFromProductCategory = (
  category: FooterProductCategory,
): FooterNavigationItem | null => {
  const slug = String(category.slug || '').trim()
  const name = String(category.name || '').trim()
  if (!slug || !name) return null

  const suppliedRoutePath = String(category.routePath || '').trim()
  const normalizedSuppliedRoutePath = suppliedRoutePath
    ? normalizeRoutePath(suppliedRoutePath)
    : ''
  const routePath = normalizedSuppliedRoutePath === '/shop'
    || normalizedSuppliedRoutePath.startsWith('/shop/')
    ? normalizedSuppliedRoutePath
    : '/shop/' + encodeURIComponent(slug)
  const children = (category.children || [])
    .map(createFooterShopLinkFromProductCategory)
    .filter((child): child is FooterNavigationItem => Boolean(child))

  return {
    fallback: name,
    to: routePath,
    ...(children.length > 0 ? { children } : {}),
  }
}

/** Build the SHOP column from the public product-category tree. */
export const createFooterShopLinksFromProductCategories = (
  categories: readonly FooterProductCategory[],
): FooterNavigationItem[] => {
  const links = categories
    .map(createFooterShopLinkFromProductCategory)
    .filter((link): link is FooterNavigationItem => Boolean(link))

  return links.length > 0
    ? links
    : [{ labelKey: 'products.nav.shop', fallback: 'Shop', to: '/shop' }]
}
