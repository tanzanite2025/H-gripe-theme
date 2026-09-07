import { computed, ref } from 'vue'
import { useNuxtApp } from '#imports'

type PageMessageValue = Record<string, unknown>
type PageMessageModule = { default?: unknown }
type PageMessageLoader = () => Promise<PageMessageModule>

const pageMessageModules = import.meta.glob('../i18n/page-messages/**/*.json') as Record<
  string,
  PageMessageLoader
>

const normalizeLocale = (locale: string): string => (
  locale.trim().replace(/_/g, '-').toLowerCase()
)

const pageMessageLoaders = new Map<string, PageMessageLoader>()

for (const [modulePath, loader] of Object.entries(pageMessageModules)) {
  const match = modulePath.match(/(?:^|\/)page-messages\/([^/]+)\/([^/]+)\.json$/)
  if (!match) continue

  const namespace = match[1]
  const localeFile = match[2]
  if (!namespace || !localeFile) continue
  const locale = localeFile.replace(/\.json$/, '')
  pageMessageLoaders.set(`${namespace}:${normalizeLocale(locale)}`, loader)
}

export const pageMessageNamespaces = Array.from(new Set(
  [...pageMessageLoaders.keys()].map(key => key.split(':')[0]).filter(Boolean),
))

const isPageMessageValue = (value: unknown): value is PageMessageValue => (
  value !== null && typeof value === 'object' && !Array.isArray(value)
)

const mergePageMessageValues = (
  base: PageMessageValue,
  override: PageMessageValue,
): PageMessageValue => {
  const merged: PageMessageValue = { ...base }

  for (const [key, value] of Object.entries(override)) {
    const baseValue = merged[key]
    if (isPageMessageValue(baseValue) && isPageMessageValue(value)) {
      merged[key] = mergePageMessageValues(baseValue, value)
    } else {
      merged[key] = value
    }
  }

  return merged
}

const resolveLocaleLoader = (namespace: string, locale: string) => {
  const normalizedLocale = normalizeLocale(locale)
  const baseLocale = normalizedLocale.split('-')[0] || normalizedLocale
  const candidates = [
    normalizedLocale,
    baseLocale,
  ]

  for (const candidate of candidates) {
    const loader = pageMessageLoaders.get(`${namespace}:${candidate}`)
    if (loader) {
      return {
        loader,
        locale: candidate,
      }
    }
  }

  return null
}

export function usePageMessages(namespace: string) {
  const nuxtApp = useNuxtApp()
  const loading = ref(false)
  const loaded = ref(false)
  const error = ref<Error | null>(null)
  const resolvedLocale = ref('')
  const loadedRequests = new Map<string, Promise<void>>()

  const loadPageMessages = async (requestedLocale: string): Promise<void> => {
    const locale = String(requestedLocale || 'en').trim() || 'en'
    const requestKey = `${namespace}:${normalizeLocale(locale)}`
    const existingRequest = loadedRequests.get(requestKey)
    if (existingRequest) {
      await existingRequest
      return
    }

    const request = (async () => {
      loading.value = true
      error.value = null

      try {
        const baseLoader = pageMessageLoaders.get(`${namespace}:en`)
        if (!baseLoader) {
          throw new Error(`English page message resource not found: ${namespace}/en`)
        }

        const resolvedOverride = resolveLocaleLoader(namespace, locale)
        const [baseModule, overrideModule] = await Promise.all([
          baseLoader(),
          resolvedOverride?.loader(),
        ])
        const baseMessage = baseModule.default ?? baseModule
        const rawOverrideMessage = overrideModule?.default ?? overrideModule

        if (!isPageMessageValue(baseMessage)) {
          throw new Error(`Page message resource must be an object: ${namespace}/en`)
        }

        let overrideMessage: PageMessageValue | undefined
        if (overrideModule) {
          if (!isPageMessageValue(rawOverrideMessage)) {
            throw new Error(
              `Page message resource must be an object: ${namespace}/${resolvedOverride?.locale}`,
            )
          }
          overrideMessage = rawOverrideMessage
        }

        const message = overrideMessage
          ? mergePageMessageValues(baseMessage, overrideMessage)
          : baseMessage
        nuxtApp.$i18n.mergeLocaleMessage(locale, {
          [namespace]: message,
        })
        resolvedLocale.value = resolvedOverride?.locale || 'en'
        loaded.value = true
      } catch (cause) {
        error.value = cause instanceof Error ? cause : new Error(String(cause))
        throw error.value
      } finally {
        loading.value = false
      }
    })()

    loadedRequests.set(requestKey, request)
    try {
      await request
    } catch (cause) {
      loadedRequests.delete(requestKey)
      throw cause
    }
  }

  return {
    isLoading: computed(() => loading.value),
    isLoaded: computed(() => loaded.value),
    error: computed(() => error.value),
    resolvedLocale: computed(() => resolvedLocale.value),
    loadPageMessages,
  }
}
