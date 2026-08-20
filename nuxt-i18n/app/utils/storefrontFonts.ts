import { getStorefrontLocaleEntry } from '~/utils/storefrontLocales'
import type { LocaleFontFamily } from '~/i18n/locales.manifest'

export const storefrontFontFamily = 'MapleUILatin, MapleUICJK'
export const storefrontLatinFontFamily = storefrontFontFamily
export const storefrontMapleUIFontFamily = storefrontFontFamily
export const storefrontLatinAccentFontFamily = `MapleUICoverageNotoSansLatinAccents, ${storefrontFontFamily}`
export const storefrontArabicFontFamily = `MapleUICoverageNotoSansArabic, ${storefrontFontFamily}`
export const storefrontDevanagariFontFamily = `MapleUICoverageNotoSansDevanagari, ${storefrontFontFamily}`
export const storefrontThaiFontFamily = `MapleUICoverageNotoSansThai, ${storefrontFontFamily}`
export const storefrontFontStylesheetPath = '/fonts/maple-ui.css'

export const storefrontFontFamilyByShard: Record<LocaleFontFamily, string> = {
  latin: storefrontLatinFontFamily,
  'latin-accent': storefrontLatinAccentFontFamily,
  'maple-ui': storefrontMapleUIFontFamily,
  arabic: storefrontArabicFontFamily,
  devanagari: storefrontDevanagariFontFamily,
  thai: storefrontThaiFontFamily,
}

export const storefrontFontFamilyForLocale = (locale: unknown) => {
  const fontFamily = getStorefrontLocaleEntry(locale)?.fontFamily || 'latin'
  return storefrontFontFamilyByShard[fontFamily] || storefrontFontFamily
}

export const storefrontFontStylesheetUrl = (origin?: string) => {
  const resolvedOrigin = origin || (typeof window !== 'undefined' ? window.location.origin : '')
  if (!resolvedOrigin) return storefrontFontStylesheetPath

  return new URL(storefrontFontStylesheetPath, resolvedOrigin).toString()
}
