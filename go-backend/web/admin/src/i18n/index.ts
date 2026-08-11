import { computed, ref, watch } from 'vue'
import zhCN from './locales/zh-CN'
import enUS from './locales/en-US'

export type AdminLocale = 'zh-CN' | 'en-US'
type TranslationValue = string | Record<string, unknown>
type TranslationParams = Record<string, string | number>

const STORAGE_KEY = 'commerce_platform.admin.locale'
const messages: Record<AdminLocale, Record<string, TranslationValue>> = {
  'zh-CN': zhCN,
  'en-US': enUS,
}

const initialLocale = (): AdminLocale => {
  if (typeof window === 'undefined') return 'zh-CN'
  const stored = window.localStorage.getItem(STORAGE_KEY)
  return stored === 'en-US' || stored === 'zh-CN' ? stored : 'zh-CN'
}

const localeState = ref<AdminLocale>(initialLocale())

const resolveMessage = (locale: AdminLocale, key: string): string | undefined => {
  const value = key.split('.').reduce<TranslationValue | undefined>(
    (current, segment) => current && typeof current === 'object' ? current[segment] as TranslationValue : undefined,
    messages[locale],
  )
  return typeof value === 'string' ? value : undefined
}

const interpolate = (message: string, params: TranslationParams = {}): string => (
  message.replace(/\{(\w+)\}/g, (_, name: string) => String(params[name] ?? `{${name}}`))
)

export const translateAdmin = (
  key: string,
  params: TranslationParams = {},
  fallback = key,
): string => {
  const message = resolveMessage(localeState.value, key) || resolveMessage('zh-CN', key) || fallback
  return interpolate(message, params)
}

export const translateAdminNavigation = <T extends { id: string; label: string; children?: T[] }>(
  items: T[],
  translate: (key: string, params?: TranslationParams, fallback?: string) => string = translateAdmin,
): T[] => items.map((item) => ({
  ...item,
  label: translate(`nav.${item.id}`, {}, item.label),
  children: item.children?.length ? translateAdminNavigation(item.children, translate) : undefined,
}))

export const useAdminI18n = () => {
  const locale = computed(() => localeState.value)
  const t = (key: string, params: TranslationParams = {}, fallback = key): string => translateAdmin(key, params, fallback)
  const setLocale = (nextLocale: AdminLocale): void => {
    localeState.value = nextLocale
  }

  watch(localeState, (nextLocale) => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, nextLocale)
      document.documentElement.lang = nextLocale
    }
  }, { immediate: true })

  return {
    locale,
    t,
    setLocale,
  }
}
