import { computed, ref } from 'vue'
import { i18nApi } from '@/api/i18n'
import {
  buildLanguageOptions,
  buildLocaleFilterOptions,
  enabledLanguageList,
  localeNameFromLanguages,
  normalizeLanguage,
  STOREFRONT_SUPPORTED_LANGUAGES
} from '@/lib/languages'
import type { AdminLanguage } from '@/lib/languages'

interface LanguagesPayload {
  languages?: AdminLanguage[]
}

export function useSupportedLanguages() {
  const storefrontLanguages = () => STOREFRONT_SUPPORTED_LANGUAGES.map(normalizeLanguage)
  const mergeStorefrontLanguages = (incomingLanguages: AdminLanguage[]): AdminLanguage[] => {
    const incomingByCode = new Map(
      incomingLanguages
        .map(normalizeLanguage)
        .filter((language) => language.code)
        .map((language) => [language.code, language])
    )

    return storefrontLanguages().map((fallbackLanguage) => {
      const incomingLanguage = incomingByCode.get(fallbackLanguage.code)
      return incomingLanguage
        ? {
            ...fallbackLanguage,
            name: incomingLanguage.name || fallbackLanguage.name,
            native_name: incomingLanguage.native_name || fallbackLanguage.native_name,
          }
        : fallbackLanguage
    })
  }

  const languages = ref<AdminLanguage[]>(storefrontLanguages())
  const loading = ref(false)

  const enabledLanguages = computed(() => enabledLanguageList(languages.value))
  const defaultLocale = computed(() => (
    enabledLanguages.value.find((language) => language.code === 'zh_cn')?.code
    || enabledLanguages.value[0]?.code
    || ''
  ))
  const languageOptions = computed(() => buildLanguageOptions(enabledLanguages.value))
  const localeFilterOptions = computed(() => buildLocaleFilterOptions(enabledLanguages.value))

  const localeName = (locale?: string | null): string => localeNameFromLanguages(locale, enabledLanguages.value)

  const fetchLanguages = async (): Promise<void> => {
    loading.value = true
    try {
      const payload = await i18nApi.listLanguages() as LanguagesPayload
      languages.value = mergeStorefrontLanguages(payload.languages || [])
    } catch (error) {
      console.error('Failed to fetch supported languages:', error)
      languages.value = storefrontLanguages()
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
