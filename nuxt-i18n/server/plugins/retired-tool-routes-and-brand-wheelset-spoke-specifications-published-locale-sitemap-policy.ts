import { defineNitroPlugin } from 'nitropack/runtime'
import storefrontLocales from '../../app/i18n/locales.manifest'
import { storefrontRetiredPagePaths } from '../../config/storefront/route-rules'
import storefrontRouteManifest from '../../public/storefront-route-manifest.json'

const storefrontLocaleCodes = new Set(storefrontLocales.map(locale => locale.code))
const wheelsetSpokeSpecificationsDeclaration = storefrontRouteManifest.routes.find(
  route => route.key === 'guides-spokeguides-brand-wheelset-spoke-specs',
)!
const wheelsetSpokeSpecificationsPublishedLocaleCodes = new Set<string>(
  wheelsetSpokeSpecificationsDeclaration.sitemap_locales,
)
const wheelsetSpokeSpecificationsPublishedAlternatives = [
  ...storefrontLocales
    .filter(locale => wheelsetSpokeSpecificationsPublishedLocaleCodes.has(locale.code))
    .map(locale => ({
      hreflang: locale.iso,
      href: `${locale.code === 'en' ? '' : `/${locale.code}`}${wheelsetSpokeSpecificationsDeclaration.path}`,
    })),
  { hreflang: 'x-default', href: wheelsetSpokeSpecificationsDeclaration.path },
]

function readUnlocalizedSitemapRouteAndLocale(routePathOrUrl: string) {
  const path = new URL(routePathOrUrl, 'https://sitemap-route.invalid').pathname.replace(/\/+$/, '')
  const firstSegment = path.split('/')[1] || ''
  const hasLocalePrefix = storefrontLocaleCodes.has(firstSegment)
  return {
    path: hasLocalePrefix ? path.slice(firstSegment.length + 1) : path,
    locale: hasLocalePrefix ? firstSegment : 'en',
  }
}

function isPublishedAndActiveToolSitemapRoute(routePathOrUrl: string) {
  const route = readUnlocalizedSitemapRouteAndLocale(routePathOrUrl)
  if (storefrontRetiredPagePaths.includes(route.path)) return false
  return route.path !== wheelsetSpokeSpecificationsDeclaration.path
    || wheelsetSpokeSpecificationsPublishedLocaleCodes.has(route.locale)
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('sitemap:resolved', (context) => {
    context.urls = context.urls
      .filter(entry => isPublishedAndActiveToolSitemapRoute(entry.loc))
      .map((entry) => {
        if (readUnlocalizedSitemapRouteAndLocale(entry.loc).path === wheelsetSpokeSpecificationsDeclaration.path) {
          return { ...entry, alternatives: wheelsetSpokeSpecificationsPublishedAlternatives }
        }
        return {
          ...entry,
          alternatives: entry.alternatives?.filter(alternate => isPublishedAndActiveToolSitemapRoute(alternate.href)),
        }
      })
  })
})
