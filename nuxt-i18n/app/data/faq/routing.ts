import localeManifest from '~/i18n/locales.manifest'

export function normalizeFaqRoutePath(routePath: string) {
  let normalized = String(routePath || '/').split('?')[0].split('#')[0].trim()
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
