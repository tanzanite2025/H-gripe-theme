export const normalizeLocaleCode = (value) => {
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

export const normalizeLanguage = (language) => ({
  code: normalizeLocaleCode(language.code),
  name: String(language.name || '').trim(),
  native_name: String(language.native_name || language.nativeName || '').trim(),
  enabled: language.enabled !== false
})

export const languageLabel = (language) => {
  const primary = language.native_name || language.name || language.code
  return primary ? `${primary} · ${language.code}` : language.code
}

export const enabledLanguageList = (languages = []) => (
  languages.filter((language) => language.enabled !== false && language.code)
)

export const buildLanguageOptions = (languages = []) => (
  enabledLanguageList(languages)
    .map((language) => ({ label: languageLabel(language), value: language.code }))
)

export const buildLocaleFilterOptions = (languages = []) => [
  { label: '全部语言', value: 'all' },
  ...buildLanguageOptions(languages)
]

export const localeNameFromLanguages = (locale, languages = []) => {
  const normalizedLocale = normalizeLocaleCode(locale)
  const language = enabledLanguageList(languages)
    .find((item) => item.code === normalizedLocale)
  return language ? languageLabel(language) : (locale || '-')
}
