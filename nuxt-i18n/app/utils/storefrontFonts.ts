import { getStorefrontLocaleEntry } from '~/utils/storefrontLocales'

export const storefrontFontFamily = 'StorefrontSystemLatin, StorefrontSystem'
export const storefrontLatinAccentFontFamily = `StorefrontSystemLatinAccents, ${storefrontFontFamily}`
export const storefrontArabicFontFamily = `StorefrontSystemArabic, ${storefrontFontFamily}`
export const storefrontDevanagariFontFamily = `StorefrontSystemDevanagari, ${storefrontFontFamily}`
export const storefrontThaiFontFamily = `StorefrontSystemThai, ${storefrontFontFamily}`
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
