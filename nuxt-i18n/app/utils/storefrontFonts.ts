import { getStorefrontLocaleEntry } from '~/utils/storefrontLocales'

export const storefrontFontFamily = 'StorefrontSystem'
export const storefrontLatinAccentFontFamily = 'StorefrontSystemLatinAccents, StorefrontSystem'
export const storefrontArabicFontFamily = 'StorefrontSystemArabic, StorefrontSystem'
export const storefrontDevanagariFontFamily = 'StorefrontSystemDevanagari, StorefrontSystem'
export const storefrontThaiFontFamily = 'StorefrontSystemThai, StorefrontSystem'
export const storefrontFontStylesheetPath = '/fonts/storefront-system.css'

export const storefrontFontFamilyForLocale = (locale: unknown) => {
  const fontFamily = getStorefrontLocaleEntry(locale)?.fontFamily

  if (fontFamily === 'arabic') return storefrontArabicFontFamily
  if (fontFamily === 'devanagari') return storefrontDevanagariFontFamily
  if (fontFamily === 'thai') return storefrontThaiFontFamily
  if (fontFamily === 'latin-accent') return storefrontLatinAccentFontFamily

  return storefrontFontFamily
}

export const storefrontFontStylesheetUrl = (origin?: string) => {
  const resolvedOrigin = origin || (typeof window !== 'undefined' ? window.location.origin : '')
  if (!resolvedOrigin) return storefrontFontStylesheetPath

  return new URL(storefrontFontStylesheetPath, resolvedOrigin).toString()
}
