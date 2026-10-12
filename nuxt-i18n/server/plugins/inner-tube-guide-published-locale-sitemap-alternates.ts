import { defineNitroPlugin } from 'nitropack/runtime'
import {
  innerTubeGuideCanonicalRoutePath,
  innerTubeGuidePublishedLocaleSeoRoutes,
  isInnerTubeGuideSitemapRouteAllowed,
  readInnerTubeGuideRouteLocaleCode,
} from '../../app/utils/seo/innerTubeGuidePublishedLocaleIndexingPolicy'

export default defineNitroPlugin((nitroApp) => {
  nitroApp.hooks.hook('sitemap:resolved', (context) => {
    context.urls = context.urls
      .filter(entry => isInnerTubeGuideSitemapRouteAllowed(entry.loc))
      .map((entry) => {
        if (readInnerTubeGuideRouteLocaleCode(entry.loc) === null) return entry

        // Sitemap's automatic locale transform can add fallback-language
        // alternatives even when their <loc> entries have been excluded.
        return {
          ...entry,
          alternatives: [
            ...innerTubeGuidePublishedLocaleSeoRoutes.map(localeRoute => ({
              hreflang: localeRoute.hreflang,
              href: localeRoute.path,
            })),
            { hreflang: 'x-default', href: innerTubeGuideCanonicalRoutePath },
          ],
        }
      })
  })
})
