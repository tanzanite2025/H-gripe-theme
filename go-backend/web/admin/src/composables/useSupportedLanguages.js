import { computed, ref } from 'vue'
import { i18nApi } from '@/api/i18n'
import {
  buildLanguageOptions,
  buildLocaleFilterOptions,
  enabledLanguageList,
  localeNameFromLanguages,
  normalizeLanguage
} from '@/lib/languages'

export function useSupportedLanguages() {
  const languages = ref([])
  const loading = ref(false)

  const enabledLanguages = computed(() => enabledLanguageList(languages.value))
  const defaultLocale = computed(() => (
    enabledLanguages.value.find((language) => language.code === 'zh_cn')?.code
    || enabledLanguages.value[0]?.code
    || ''
  ))
  const languageOptions = computed(() => buildLanguageOptions(enabledLanguages.value))
  const localeFilterOptions = computed(() => buildLocaleFilterOptions(enabledLanguages.value))

  const localeName = (locale) => localeNameFromLanguages(locale, enabledLanguages.value)

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
