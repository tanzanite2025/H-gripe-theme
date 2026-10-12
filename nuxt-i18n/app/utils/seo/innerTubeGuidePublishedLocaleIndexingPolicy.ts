import storefrontLocales from '../../i18n/locales.manifest'

export const innerTubeGuideCanonicalRoutePath = '/tools/valve-length-fitment-calculator'
export const innerTubeGuidePublishedLocaleCodes = ['en', 'zh_cn'] as const

const storefrontLocaleCodeSet = new Set(storefrontLocales.map(locale => locale.code))
const innerTubeGuidePublishedLocaleCodeSet = new Set<string>(innerTubeGuidePublishedLocaleCodes)

export const innerTubeGuidePublishedLocaleSeoRoutes = innerTubeGuidePublishedLocaleCodes.map(code => ({
  code,
  path: `${code === 'en' ? '' : `/${code}`}${innerTubeGuideCanonicalRoutePath}`,
  hreflang: storefrontLocales.find(locale => locale.code === code)!.iso,
}))

function normalizeInnerTubeGuideRoutePathWithoutParameters(routePathOrUrl: string): string {
  let routePath = String(routePathOrUrl || '').trim()
  if (/^https?:\/\//i.test(routePath)) {
    try {
      routePath = new URL(routePath).pathname
    } catch {
      return ''
    }
  }
  return (routePath.split(/[?#]/, 1)[0] || '').replace(/\/+$/, '')
}

export function readInnerTubeGuideRouteLocaleCode(routePathOrUrl: string): string | null {
  const routePath = normalizeInnerTubeGuideRoutePathWithoutParameters(routePathOrUrl)
  if (routePath === innerTubeGuideCanonicalRoutePath) return 'en'
  const firstPathSegment = routePath.split('/')[1] || ''
  return storefrontLocaleCodeSet.has(firstPathSegment)
    && routePath === `/${firstPathSegment}${innerTubeGuideCanonicalRoutePath}`
    ? firstPathSegment
    : null
}

export function isInnerTubeGuidePublishedLocale(localeCode: string): boolean {
  return innerTubeGuidePublishedLocaleCodeSet.has(localeCode)
}

export function getInnerTubeGuideActualContentLanguageTag(localeCode: string): string {
  return innerTubeGuidePublishedLocaleSeoRoutes.find(entry => entry.code === localeCode)?.hreflang || 'en-US'
}

export function isInnerTubeGuideSitemapRouteAllowed(routePathOrUrl: string): boolean {
  if (readInnerTubeGuideRouteLocaleCode(routePathOrUrl) === null) return true
  const routePath = normalizeInnerTubeGuideRoutePathWithoutParameters(routePathOrUrl)
  return innerTubeGuidePublishedLocaleSeoRoutes.some(entry => entry.path === routePath)
}

export const innerTubeGuideUnpublishedLocaleSitemapExcludes = storefrontLocales
  .filter(locale => !isInnerTubeGuidePublishedLocale(locale.code))
  .map(locale => `/${locale.code}${innerTubeGuideCanonicalRoutePath}`)
