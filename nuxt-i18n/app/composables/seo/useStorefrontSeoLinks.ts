import { computed, toValue, watch, type MaybeRefOrGetter } from 'vue'
import {
  useI18n,
  useRequestURL,
  useRoute,
  useRuntimeConfig,
  useState,
  useSwitchLocalePath,
} from '#imports'
import { toAbsoluteSeoUrl } from '~/utils/seo/urls'

export interface StorefrontSeoAlternateLinkEntry {
  code: string
  path: string
}

interface StorefrontSeoLinksOptions {
  siteOrigin?: MaybeRefOrGetter<string>
}

type RawLocale = string | { code: string; iso?: string }

interface LocaleEntry {
  code: string
  iso?: string
}

const alternateLinksStateKey = 'alternateLinksOverride'

const trimTrailingSlash = (value: string): string => value.replace(/\/+$/, '')

const resolveLocaleEntries = (locales: unknown): LocaleEntry[] => {
  const source = Array.isArray(locales)
    ? locales
    : typeof locales === 'object' && locales !== null && 'value' in locales
      ? (locales as { value?: RawLocale[] }).value
      : []

  if (!Array.isArray(source)) return []

  return source
    .map((entry) => typeof entry === 'string'
      ? { code: entry }
      : { code: entry.code, iso: entry.iso })
    .filter((entry) => entry.code)
}

export const useStorefrontSeoRouteOverride = (
  entries: MaybeRefOrGetter<StorefrontSeoAlternateLinkEntry[] | null | undefined>,
) => {
  const state = useState<StorefrontSeoAlternateLinkEntry[] | null>(
    alternateLinksStateKey,
    () => null,
  )

  watch(
    () => toValue(entries),
    (value) => {
      const normalized = Array.isArray(value)
        ? value
          .map((entry) => ({
            code: String(entry.code || '').trim(),
            path: String(entry.path || '').trim(),
          }))
          .filter((entry) => entry.code && entry.path)
        : []

      state.value = normalized.length ? normalized : null
    },
    { immediate: true, deep: true },
  )

  return state
}

export const useStorefrontSeoLinks = (options: StorefrontSeoLinksOptions = {}) => {
  const config = useRuntimeConfig()
  const requestUrl = useRequestURL()
  const route = useRoute()
  const { locales, defaultLocale, locale } = useI18n()
  const switchLocalePath = useSwitchLocalePath()
  const alternateLinksOverride = useState<StorefrontSeoAlternateLinkEntry[] | null>(
    alternateLinksStateKey,
    () => null,
  )

  watch(
    () => route.fullPath,
    () => {
      alternateLinksOverride.value = null
    },
    { immediate: true },
  )

  const siteOrigin = computed(() => {
    const configured = options.siteOrigin ? String(toValue(options.siteOrigin) || '').trim() : ''
    if (configured) return trimTrailingSlash(configured)

    const runtimeSiteUrl = String((config.public as { siteUrl?: string }).siteUrl || '').trim()
    if (runtimeSiteUrl) return trimTrailingSlash(runtimeSiteUrl)

    return trimTrailingSlash(requestUrl.origin)
  })

  const resolvedLocales = computed(() => resolveLocaleEntries(locales))
  const currentLocaleCode = computed(() => String(toValue(locale) || '').trim())
  const defaultLocaleCode = computed(() => String(toValue(defaultLocale) || 'en').trim() || 'en')

  const makeAbsoluteUrl = (path: string) => toAbsoluteSeoUrl(siteOrigin.value, path)

  const overrideEntries = computed(() => {
    if (!Array.isArray(alternateLinksOverride.value) || !alternateLinksOverride.value.length) {
      return null
    }
    return alternateLinksOverride.value
  })

  const canonicalPath = computed(() => {
    const currentOverride = overrideEntries.value?.find(
      (entry) => entry.code === currentLocaleCode.value,
    )
    return currentOverride?.path || route.path || '/'
  })

  const canonicalUrl = computed(() => makeAbsoluteUrl(canonicalPath.value))

  const alternateLinks = computed(() => {
    if (overrideEntries.value) {
      return overrideEntries.value.map((override) => {
        const localeEntry = resolvedLocales.value.find((entry) => entry.code === override.code)
        return {
          hreflang: localeEntry?.iso || override.code,
          href: makeAbsoluteUrl(override.path),
        }
      })
    }

    return resolvedLocales.value.map((entry) => {
      const targetPath = switchLocalePath(entry.code as any) || '/'
      return {
        hreflang: entry.iso || entry.code,
        href: makeAbsoluteUrl(targetPath),
      }
    })
  })

  const xDefaultLink = computed(() => {
    const overrideDefault = overrideEntries.value?.find(
      (entry) => entry.code === defaultLocaleCode.value,
    ) || overrideEntries.value?.[0]
    if (overrideDefault?.path) {
      return makeAbsoluteUrl(overrideDefault.path)
    }

    const targetPath = switchLocalePath(defaultLocaleCode.value as any) || '/'
    return makeAbsoluteUrl(targetPath)
  })

  return {
    canonicalUrl,
    alternateLinks,
    xDefaultLink,
  }
}
