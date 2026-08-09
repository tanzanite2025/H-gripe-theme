export interface AdminLanguage {
  code: string
  name?: string
  native_name?: string
  nativeName?: string
  enabled?: boolean
}

export interface LanguageOption {
  label: string
  value: string
}

export const STOREFRONT_SUPPORTED_LANGUAGES: AdminLanguage[] = [
  { code: 'en', name: 'English', native_name: 'English', enabled: true },
  { code: 'zh_cn', name: 'Chinese (Simplified)', native_name: '简体中文', enabled: true },
  { code: 'fr', name: 'French', native_name: 'Français', enabled: true },
  { code: 'de', name: 'German', native_name: 'Deutsch', enabled: true },
  { code: 'es', name: 'Spanish', native_name: 'Español', enabled: true },
  { code: 'ja', name: 'Japanese', native_name: '日本語', enabled: true },
  { code: 'ko', name: 'Korean', native_name: '한국어', enabled: true },
  { code: 'it', name: 'Italian', native_name: 'Italiano', enabled: true },
  { code: 'pt', name: 'Portuguese', native_name: 'Português', enabled: true },
  { code: 'ru', name: 'Russian', native_name: 'Русский', enabled: true },
  { code: 'ar', name: 'Arabic', native_name: 'العربية', enabled: true },
  { code: 'nl', name: 'Dutch', native_name: 'Nederlands', enabled: true },
  { code: 'tr', name: 'Turkish', native_name: 'Türkçe', enabled: true },
  { code: 'id', name: 'Indonesian', native_name: 'Bahasa Indonesia', enabled: true },
  { code: 'th', name: 'Thai', native_name: 'ไทย', enabled: true },
  { code: 'sv', name: 'Swedish', native_name: 'Svenska', enabled: true },
  { code: 'da', name: 'Danish', native_name: 'Dansk', enabled: true },
  { code: 'fi', name: 'Finnish', native_name: 'Suomi', enabled: true },
  { code: 'hi', name: 'Hindi', native_name: 'हिन्दी', enabled: true },
  { code: 'ms', name: 'Malay', native_name: 'Bahasa Melayu', enabled: true },
]

export const normalizeLocaleCode = (value?: string | null): string => {
  const raw = String(value || '')
    .trim()
    .toLowerCase()
    .split(',')[0]
    .split(';')[0]
  const cleaned = raw
    .replace(/_/g, '-')
    .replace(/[^a-z-]/g, '')

  if (!cleaned) return ''
  if (['zh', 'zh-cn', 'zh-hans', 'zh-sg'].includes(cleaned)) return 'zh_cn'

  const [base] = cleaned.split('-')
  if (base && base !== 'zh') return base
  return cleaned.replace(/-/g, '_')
}

export const normalizeLanguage = (language: AdminLanguage): Required<Pick<AdminLanguage, 'code' | 'name' | 'native_name' | 'enabled'>> => ({
  code: normalizeLocaleCode(language.code),
  name: String(language.name || '').trim(),
  native_name: String(language.native_name || language.nativeName || '').trim(),
  enabled: language.enabled !== false
})

export const languageLabel = (language: AdminLanguage): string => {
  const primary = language.native_name || language.name || language.code
  return primary ? `${primary} · ${language.code}` : language.code
}

export const enabledLanguageList = <T extends AdminLanguage>(languages: T[] = []): T[] => (
  languages.filter((language) => language.enabled !== false && language.code)
)

export const buildLanguageOptions = (languages: AdminLanguage[] = []): LanguageOption[] => (
  enabledLanguageList(languages)
    .map((language) => ({ label: languageLabel(language), value: language.code }))
)

export const buildLocaleFilterOptions = (languages: AdminLanguage[] = []): LanguageOption[] => [
  { label: '全部语言', value: 'all' },
  ...buildLanguageOptions(languages)
]

export const localeNameFromLanguages = (locale?: string | null, languages: AdminLanguage[] = []): string => {
  const normalizedLocale = normalizeLocaleCode(locale)
  const language = enabledLanguageList(languages)
    .find((item) => item.code === normalizedLocale)
  return language ? languageLabel(language) : (locale || '-')
}
