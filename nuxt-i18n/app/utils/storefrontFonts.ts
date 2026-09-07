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

type StorefrontFontPreloadLink = {
  key: string
  rel: 'preload'
  as: 'font'
  type: 'font/woff2'
  href: string
  crossorigin: 'anonymous'
}

const storefrontLatinFontPreloadPath = '/fonts/MapleUI-Latin.00af3fec5b34.woff2'
const storefrontFontPreloadPathByShard: Record<LocaleFontFamily, string> = {
  latin: storefrontLatinFontPreloadPath,
  'latin-accent': '/fonts/MapleUI-Coverage-NotoSans-Latin-Accents.e645edc952b6.woff2',
  'maple-ui': storefrontLatinFontPreloadPath,
  arabic: '/fonts/MapleUI-Coverage-NotoSans-Arabic.ce85091f0209.woff2',
  devanagari: '/fonts/MapleUI-Coverage-NotoSans-Devanagari.3b3cae4d2600.woff2',
  thai: '/fonts/MapleUI-Coverage-NotoSans-Thai.1f5a173641bb.woff2',
}
const storefrontAdditionalFontPreloadPathByShard: Partial<Record<LocaleFontFamily, string>> = {
  'latin-accent': storefrontFontPreloadPathByShard['latin-accent'],
  arabic: storefrontFontPreloadPathByShard.arabic,
  devanagari: storefrontFontPreloadPathByShard.devanagari,
  thai: storefrontFontPreloadPathByShard.thai,
}

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

export const storefrontAdditionalFontPreloadLinkForLocale = (locale: unknown): StorefrontFontPreloadLink | null => {
  const fontFamily = getStorefrontLocaleEntry(locale)?.fontFamily || 'latin'
  const href = storefrontAdditionalFontPreloadPathByShard[fontFamily]
  if (!href) return null

  return {
    key: `storefront-font-preload-${fontFamily}`,
    rel: 'preload',
    as: 'font',
    type: 'font/woff2',
    href,
    crossorigin: 'anonymous',
  }
}

export const storefrontFontStylesheetUrl = (origin?: string) => {
  const resolvedOrigin = origin || (typeof window !== 'undefined' ? window.location.origin : '')
  if (!resolvedOrigin) return storefrontFontStylesheetPath

  return new URL(storefrontFontStylesheetPath, resolvedOrigin).toString()
}
