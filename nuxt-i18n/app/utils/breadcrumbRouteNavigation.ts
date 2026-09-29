import { normalizePrimaryMegaNavPath } from './primaryMegaNav.js'

const pathWithoutQueryOrHash = (path: string) => path.split(/[?#]/, 1)[0] || '/'

const splitNormalizedPath = (path: string, localeCodes: string[]) => (
  normalizePrimaryMegaNavPath(pathWithoutQueryOrHash(path), localeCodes)
    .split('/')
    .filter(Boolean)
)

interface DynamicRouteParameter {
  constraint: string
  modifier: '' | '?' | '*' | '+'
}

const parseDynamicRouteParameter = (segment: string): DynamicRouteParameter | null => {
  if (segment.startsWith('[') && segment.endsWith(']')) {
    const catchAll = segment.startsWith('[...')
    return { constraint: '.*', modifier: catchAll ? '*' : '' }
  }

  if (segment.startsWith('*')) {
    return { constraint: '.*', modifier: '*' }
  }

  if (!segment.startsWith(':')) return null

  const openIndex = segment.indexOf('(')
  if (openIndex < 0) {
    const rawParameter = segment.slice(1)
    const modifier = rawParameter.slice(-1)
    return {
      constraint: '[^/]+',
      modifier: ['?', '*', '+'].includes(modifier) ? modifier as DynamicRouteParameter['modifier'] : '',
    }
  }

  const closeIndex = segment.lastIndexOf(')')
  if (closeIndex < openIndex) return { constraint: '.*', modifier: '' }

  const modifier = segment.slice(closeIndex + 1)
  return {
    constraint: segment.slice(openIndex + 1, closeIndex) || '.*',
    modifier: ['?', '*', '+'].includes(modifier) ? modifier as DynamicRouteParameter['modifier'] : '',
  }
}

const dynamicParameterMatches = (parameter: DynamicRouteParameter, value: string) => {
  try {
    return new RegExp(`^(?:${parameter.constraint})$`).test(value)
  } catch {
    // An unparseable custom route constraint is not safe to reuse for a switch.
    return false
  }
}

/** Match a concrete path against a Nuxt/Vue Router route path pattern. */
export const breadcrumbRoutePatternMatches = (
  patternPath: string,
  targetPath: string,
  localeCodes: string[] = [],
) => {
  const patternSegments = splitNormalizedPath(patternPath, localeCodes)
  const targetSegments = splitNormalizedPath(targetPath, localeCodes)

  const matchesFrom = (patternIndex: number, targetIndex: number): boolean => {
    if (patternIndex === patternSegments.length) {
      return targetIndex === targetSegments.length
    }

    const patternSegment = patternSegments[patternIndex] || ''
    const parameter = parseDynamicRouteParameter(patternSegment)
    if (!parameter) {
      return patternSegment === targetSegments[targetIndex]
        && matchesFrom(patternIndex + 1, targetIndex + 1)
    }

    const remainingTargetSegments = targetSegments.length - targetIndex
    const minimumSegments = parameter.modifier === '*' || parameter.modifier === '?'
      ? 0
      : 1
    const maximumSegments = parameter.modifier === '*' || parameter.modifier === '+'
      ? remainingTargetSegments
      : Math.min(1, remainingTargetSegments)

    for (let count = maximumSegments; count >= minimumSegments; count--) {
      const value = targetSegments.slice(targetIndex, targetIndex + count).join('/')
      if (
        (count === 0 || dynamicParameterMatches(parameter, value))
        && matchesFrom(patternIndex + 1, targetIndex + count)
      ) {
        return true
      }
    }

    return false
  }

  return matchesFrom(0, 0)
}

export interface BreadcrumbSiblingTargetOptions {
  /** Full active route, including any descendants below the expanded crumb. */
  currentPath: string
  /** Path represented by the crumb whose sibling menu is open. */
  breadcrumbPath: string
  /** Logical path of the selected sibling group, which may be a route prefix. */
  siblingPath: string
  /** Registered canonical route to open when no matching descendant exists. */
  fallbackPath: string
  /** Current Nuxt route paths, including constrained dynamic route patterns. */
  routePatterns: readonly string[]
  localeCodes?: string[]
}

/**
 * Keep the route suffix below a breadcrumb when the selected sibling branch
 * has a registered route for that same suffix. Otherwise use its canonical
 * target. Route additions are picked up from the router's route table.
 */
export const resolveBreadcrumbSiblingTarget = ({
  currentPath,
  breadcrumbPath,
  siblingPath,
  fallbackPath,
  routePatterns,
  localeCodes = [],
}: BreadcrumbSiblingTargetOptions) => {
  const currentSegments = splitNormalizedPath(currentPath, localeCodes)
  const breadcrumbSegments = splitNormalizedPath(breadcrumbPath, localeCodes)
  const siblingSegments = splitNormalizedPath(siblingPath, localeCodes)
  const isBreadcrumbPrefix = breadcrumbSegments.length <= currentSegments.length
    && breadcrumbSegments.every((segment, index) => segment === currentSegments[index])

  if (!isBreadcrumbPrefix || currentSegments.length === breadcrumbSegments.length) {
    return fallbackPath
  }

  const suffix = currentSegments.slice(breadcrumbSegments.length)
  const candidatePath = `/${[...siblingSegments, ...suffix].join('/')}`
  const hasRegisteredRoute = routePatterns.some(patternPath => (
    breadcrumbRoutePatternMatches(patternPath, candidatePath, localeCodes)
  ))

  return hasRegisteredRoute ? candidatePath : fallbackPath
}
