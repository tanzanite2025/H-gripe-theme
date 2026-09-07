import type { BlogLocalizedRoute } from '~/utils/blog/types'
import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

export const buildBlogPath = (slug: string): string => {
  const cleanSlug = String(slug || '').trim()
  if (!cleanSlug) return '/resources/blog'

  return `/resources/blog/${encodeURIComponent(cleanSlug)}`
}

export const buildLocalizedBlogPath = (
  locale: string,
  slug: string,
): string => {
  const normalizedLocale = normalizeStorefrontLocaleCode(locale)
  const localePrefix = normalizedLocale && normalizedLocale !== 'en'
    ? `/${normalizedLocale}`
    : ''

  return `${localePrefix}${buildBlogPath(slug)}`
}

export const normalizeBlogLocalizedRoutes = (
  routes: unknown,
): BlogLocalizedRoute[] => {
  if (!Array.isArray(routes)) return []

  return routes
    .map((entry): BlogLocalizedRoute | null => {
      if (!entry || typeof entry !== 'object') return null
      const value = entry as Record<string, unknown>
      const locale = String(value.locale || '').trim()
      const slug = String(value.slug || '').trim()
      const path = String(value.path || '').trim()
      if (!locale || !slug || !path) return null

      const id = Number(value.id)
      return {
        ...(Number.isFinite(id) && id > 0 ? { id } : {}),
        locale,
        slug,
        path,
      }
    })
    .filter((entry): entry is BlogLocalizedRoute => Boolean(entry))
}
