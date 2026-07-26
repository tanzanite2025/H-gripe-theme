import { computed, ref } from 'vue'
import { i18nApi } from '@/api/i18n'

const normalizeLocaleCode = (value) => String(value || '')
  .trim()
  .toLowerCase()
  .replace(/-/g, '_')

const normalizeLanguage = (language) => ({
  code: normalizeLocaleCode(language.code),
  name: String(language.name || '').trim(),
  native_name: String(language.native_name || language.nativeName || '').trim(),
  enabled: language.enabled !== false
})

const languageLabel = (language) => {
  const primary = language.native_name || language.name || language.code
  return primary ? `${primary} · ${language.code}` : language.code
}

export function useSupportedLanguages() {
  const languages = ref([])
  const loading = ref(false)

  const enabledLanguages = computed(() => languages.value.filter((language) => language.enabled && language.code))
  const defaultLocale = computed(() => (
    enabledLanguages.value.find((language) => language.code === 'zh_cn')?.code
    || enabledLanguages.value[0]?.code
    || ''
  ))
  const languageOptions = computed(() => enabledLanguages.value.map((language) => ({
    label: languageLabel(language),
    value: language.code
  })))
  const localeFilterOptions = computed(() => [
    { label: '全部语言', value: 'all' },
    ...languageOptions.value
  ])

  const localeName = (locale) => {
    const normalizedLocale = normalizeLocaleCode(locale)
    const language = enabledLanguages.value.find((item) => item.code === normalizedLocale)
    return language ? languageLabel(language) : (locale || '-')
  }

  const fetchLanguages = async () => {
    loading.value = true
    try {
      const payload = await i18nApi.listLanguages()
      languages.value = (payload.languages || [])
        .map(normalizeLanguage)
        .filter((language) => language.code)
    } catch (error) {
      console.error('Failed to fetch supported languages:', error)
      languages.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    languages,
    enabledLanguages,
    defaultLocale,
    languageOptions,
    localeFilterOptions,
    localeName,
    loading,
    fetchLanguages
  }
}
