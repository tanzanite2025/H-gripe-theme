import { defineNitroPlugin } from 'nitropack/runtime'
import storefrontLocales from '../../app/i18n/locales.manifest'
import { storefrontRetiredPagePaths } from '../../config/storefront/route-rules'
import storefrontRouteManifest from '../../public/storefront-route-manifest.json'

const storefrontLocaleCodes = new Set(storefrontLocales.map(locale => locale.code))
const publishedToolSitemapPolicies = new Map<string, {
  localeCodes: Set<string>
  alternatives: { hreflang: string; href: string }[]
}>()

for (const declaration of storefrontRouteManifest.routes) {
  if (!declaration.path.startsWith('/tools/') || !Array.isArray(declaration.sitemap_locales)) continue
  const localeCodes = new Set<string>(declaration.sitemap_locales)
  publishedToolSitemapPolicies.set(declaration.path, {
    localeCodes,
    alternatives: [
      ...storefrontLocales
        .filter(locale => localeCodes.has(locale.code))
        .map(locale => ({
          hreflang: locale.iso,
          href: `${locale.code === 'en' ? '' : `/${locale.code}`}${declaration.path}`,
        })),
      { hreflang: 'x-default', href: declaration.path },
    ],
  })
}

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
  const policy = publishedToolSitemapPolicies.get(route.path)
  return !policy || policy.localeCodes.has(route.locale)
}

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('sitemap:resolved', (context) => {
    context.urls = context.urls
      .filter(entry => isPublishedAndActiveToolSitemapRoute(entry.loc))
      .map((entry) => {
        const policy = publishedToolSitemapPolicies.get(readUnlocalizedSitemapRouteAndLocale(entry.loc).path)
        return {
          ...entry,
          alternatives: policy?.alternatives
            || entry.alternatives?.filter(alternate => isPublishedAndActiveToolSitemapRoute(alternate.href)),
        }
      })
  })
})
