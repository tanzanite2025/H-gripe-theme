export type LocaleDirection = 'ltr' | 'rtl'
export type LocaleFontFamily = 'latin' | 'latin-accent' | 'maple-ui' | 'arabic' | 'devanagari' | 'thai'

export interface LocaleManifestEntry {
  code: string
  iso: string
  name: string
  file: string
  files?: string[]
  language?: string
  dir?: LocaleDirection
  fontFamily: LocaleFontFamily
}

export const locales: LocaleManifestEntry[] = [
  { code: 'en', iso: 'en-US', name: 'English', file: 'en.json', fontFamily: 'latin' },
  { code: 'zh_cn', iso: 'zh-CN', name: '简体中文', file: 'zh_cn.json', fontFamily: 'maple-ui' },
  { code: 'fr', iso: 'fr-FR', name: 'Français', file: 'fr.json', fontFamily: 'latin-accent' },
  { code: 'de', iso: 'de-DE', name: 'Deutsch', file: 'de.json', fontFamily: 'latin-accent' },
  { code: 'es', iso: 'es-ES', name: 'Español', file: 'es.json', fontFamily: 'latin-accent' },
  { code: 'ja', iso: 'ja-JP', name: '日本語', file: 'ja.json', fontFamily: 'maple-ui' },
  { code: 'ko', iso: 'ko-KR', name: '한국어', file: 'ko.json', fontFamily: 'maple-ui' },
  { code: 'it', iso: 'it-IT', name: 'Italiano', file: 'it.json', fontFamily: 'latin-accent' },
  { code: 'pt', iso: 'pt-PT', name: 'Português', file: 'pt.json', fontFamily: 'latin-accent' },
  { code: 'ru', iso: 'ru-RU', name: 'Русский', file: 'ru.json', fontFamily: 'maple-ui' },
  { code: 'ar', iso: 'ar-SA', name: 'العربية', file: 'ar.json', dir: 'rtl', fontFamily: 'arabic' },
  { code: 'nl', iso: 'nl-NL', name: 'Nederlands', file: 'nl.json', fontFamily: 'latin-accent' },
  { code: 'tr', iso: 'tr-TR', name: 'Türkçe', file: 'tr.json', fontFamily: 'latin-accent' },
  { code: 'id', iso: 'id-ID', name: 'Bahasa Indonesia', file: 'id.json', fontFamily: 'latin' },
  { code: 'th', iso: 'th-TH', name: 'ไทย', file: 'th.json', fontFamily: 'thai' },
  { code: 'sv', iso: 'sv-SE', name: 'Svenska', file: 'sv.json', fontFamily: 'latin-accent' },
  { code: 'da', iso: 'da-DK', name: 'Dansk', file: 'da.json', fontFamily: 'latin-accent' },
  { code: 'fi', iso: 'fi-FI', name: 'Suomi', file: 'fi.json', fontFamily: 'latin-accent' },
  { code: 'hi', iso: 'hi-IN', name: 'हिन्दी', file: 'hi.json', fontFamily: 'devanagari' },
  { code: 'ms', iso: 'ms-MY', name: 'Bahasa Melayu', file: 'ms.json', fontFamily: 'latin' },
]

export default locales
