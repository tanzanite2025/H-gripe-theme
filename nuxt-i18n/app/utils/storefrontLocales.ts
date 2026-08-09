import localeManifest from '~/i18n/locales.manifest'
import type { LocaleManifestEntry } from '~/i18n/locales.manifest'

const normalizeLocaleLookupKey = (value: unknown): string => {
  return String(value || '').trim().replace(/_/g, '-').toLowerCase()
}

const localeByCode = new Map<string, LocaleManifestEntry>(
  localeManifest.map((entry) => [entry.code, entry]),
)

const localeAliases = new Map<string, string>()

for (const entry of localeManifest) {
  localeAliases.set(normalizeLocaleLookupKey(entry.code), entry.code)
  localeAliases.set(normalizeLocaleLookupKey(entry.iso), entry.code)

  if (entry.language) {
    localeAliases.set(normalizeLocaleLookupKey(entry.language), entry.code)
  }
}

if (localeByCode.has('zh_cn')) {
  localeAliases.set('zh', 'zh_cn')
}

export const STOREFRONT_LOCALE_CODES = localeManifest.map((entry) => entry.code)

export function normalizeStorefrontLocaleCode(value: unknown): string {
  const lookupKey = normalizeLocaleLookupKey(value)
  if (!lookupKey) return ''
  return localeAliases.get(lookupKey) || ''
}

export function isSupportedStorefrontLocale(value: unknown): boolean {
  return normalizeStorefrontLocaleCode(value) !== ''
}

export function isSimplifiedChineseStorefrontLocale(value: unknown): boolean {
  return normalizeStorefrontLocaleCode(value) === 'zh_cn'
}

export function getStorefrontLocaleEntry(value: unknown): LocaleManifestEntry | undefined {
  const code = normalizeStorefrontLocaleCode(value)
  return code ? localeByCode.get(code) : undefined
}

export function getStorefrontLocaleName(value: unknown, fallback = ''): string {
  const entry = getStorefrontLocaleEntry(value)
  return entry?.name || fallback || String(value || '')
}

export function getStorefrontLocaleFlag(value: unknown): string {
  const entry = getStorefrontLocaleEntry(value)
  const regionCode = entry?.iso.split('-')[1]?.toUpperCase() || ''
  if (!/^[A-Z]{2}$/.test(regionCode)) return ''

  return Array.from(regionCode)
    .map((letter) => String.fromCodePoint(127397 + letter.charCodeAt(0)))
    .join('')
}
