import localeManifest from '~/i18n/locales.manifest'

const FAQ_AUTO_INSERT_DISABLED_PATHS = new Set([
  '/faq',
  '/support/faqs',
])

export function normalizeFaqRoutePath(routePath: string) {
  const [pathWithoutQuery = '/'] = String(routePath || '/').split('?')
  const [pathWithoutHash = '/'] = pathWithoutQuery.split('#')
  let normalized = pathWithoutHash.trim()
  const localeCodes = localeManifest
    .map((locale: { code: string }) => String(locale.code || '').trim())
    .filter(Boolean)
    .map((code: string) => code.replace(/[.*+?^${}()|[\]\\]/g, '\\$&'))
    .join('|')
  if (localeCodes) {
    normalized = normalized.replace(new RegExp(`^/(${localeCodes})(?=/|$)`, 'i'), '')
  }
  if (!normalized.startsWith('/')) normalized = `/${normalized}`
  normalized = normalized.replace(/\/+$/, '')
  return normalized || '/'
}

export function resolveFaqRouteLookupPath(routePath: string) {
  const normalizedPath = normalizeFaqRoutePath(routePath)

  if (/^\/shop\/[^/]+$/.test(normalizedPath)) {
    return '/shop/:slug'
  }

  if (/^\/products\/[^/]+$/.test(normalizedPath)) {
    return '/products/:slug'
  }

  return normalizedPath
}

export function shouldAutoInsertFaqForRoute(routePath: string) {
  return !FAQ_AUTO_INSERT_DISABLED_PATHS.has(normalizeFaqRoutePath(routePath))
}
