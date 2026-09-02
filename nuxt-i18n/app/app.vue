<template>
  <div class="tz-light-theme">
    <NuxtLoadingIndicator
      color="#059669"
      :height="3"
      :throttle="80"
    />
    <SiteHeader ref="siteHeaderRef" />
    <NuxtLayout>
      <SidePanel>
        <template #left>
          <LazyAccountSidebarPanel />
        </template>
      </SidePanel>
      <!-- Render the current page inside the active layout -->
      <NuxtPage />
    </NuxtLayout>
    
    <ClientOnly>
      <StorefrontClientOverlays />
    </ClientOnly>
    
    <!-- Cookie 同意弹窗 -->
    <CookieConsent />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useHead, useI18n, useRequestURL, useRuntimeConfig } from '#imports'
import SidePanel from './components/SidePanel.vue'
import SiteHeader from '~/components/SiteHeader.vue'
import { useAuth } from '~/composables/useAuth'
import { useProductCategories } from '~/composables/useProductCategories'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useShopCategories } from '~/composables/useShopCategories'
import localeManifest from '~/i18n/locales.manifest'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import { storefrontFontPreloadLinkForLocale } from '~/utils/storefrontFonts'

const auth = useAuth()
const runtimeConfig = useRuntimeConfig()
const requestUrl = useRequestURL()
const { siteSettings } = useSiteSettings()
const { loadCategories: prefetchShopCategories } = useShopCategories()
const { loadCategories: prefetchProductCategories } = useProductCategories()
const { locale } = useI18n()

const siteHeaderRef = ref<InstanceType<typeof SiteHeader> | null>(null)
const activeLocaleEntry = computed(() => {
  const code = String(locale.value || 'en').trim().replace(/_/g, '-').toLowerCase()
  const baseCode = code.split('-')[0]

  return localeManifest.find((entry) => {
    const entryCode = entry.code.replace(/_/g, '-').toLowerCase()
    const entryIso = entry.iso.toLowerCase()

    return entryCode === code || entryCode === baseCode || entryIso === code
  }) || localeManifest.find(entry => entry.code === 'en')
})
const htmlLanguage = computed(() => (
  activeLocaleEntry.value?.iso || activeLocaleEntry.value?.language || String(locale.value || 'en').replace(/_/g, '-')
))
const htmlDirection = computed(() => activeLocaleEntry.value?.dir || 'ltr')
const htmlFontFamily = computed(() => activeLocaleEntry.value?.fontFamily || 'latin')
const htmlFontPreloadLink = computed(() => storefrontFontPreloadLinkForLocale(locale.value))
const MAX_PRECONNECT_ORIGINS = 4

type PreconnectLink = {
  key: string
  rel: 'preconnect'
  href: string
  crossorigin?: 'anonymous'
}

const resolveOrigin = (value: unknown, baseOrigin: string): string => {
  const candidate = String(value || '').trim()
  if (!candidate) return ''

  try {
    const url = candidate.includes('://')
      ? new URL(candidate)
      : new URL(candidate.startsWith('/') ? candidate : `https://${candidate}`, baseOrigin)
    return ['http:', 'https:'].includes(url.protocol) ? url.origin : ''
  } catch {
    return ''
  }
}

const splitConfiguredOrigins = (value: unknown): string[] => (
  String(value || '')
    .split(',')
    .map(item => item.trim())
    .filter(Boolean)
)

const preconnectLinks = computed<PreconnectLink[]>(() => {
  const currentOrigin = requestUrl.origin
  const publicConfig = runtimeConfig.public as {
    apiBase?: string
    imageDomains?: string
    siteUrl?: string
  }
  const appConfig = runtimeConfig.app as { cdnURL?: string }
  const candidates = [
    publicConfig.apiBase,
    publicConfig.siteUrl,
    appConfig.cdnURL,
    ...splitConfiguredOrigins(publicConfig.imageDomains),
  ]

  return Array.from(new Set(
    candidates
      .map(candidate => resolveOrigin(candidate, currentOrigin))
      .filter(origin => origin && origin !== currentOrigin)
  ))
    .slice(0, MAX_PRECONNECT_ORIGINS)
    .map(origin => ({
      key: `storefront-preconnect:${origin}`,
      rel: 'preconnect',
      href: origin,
      crossorigin: 'anonymous',
    }))
})

const resolveConfiguredAsset = (value: string) => {
  if (!/^\/uploads\//i.test(value)) return value

  const apiBase = String(
    (runtimeConfig.public as { apiBase?: string }).apiBase || ''
  ).trim()

  try {
    return new URL(value, apiBase).toString()
  } catch {
    return value
  }
}
const siteFavicon = computed(() => {
  const configuredFavicon = (siteSettings.value.siteFavicon || '').toString().trim()
  if (configuredFavicon) return resolveConfiguredAsset(configuredFavicon)

  const configuredLogo = (siteSettings.value.siteLogo || '').toString().trim()
  return configuredLogo ? resolveConfiguredAsset(configuredLogo) : '/favicon.svg'
})

useHead(() => ({
  htmlAttrs: {
    lang: htmlLanguage.value,
    dir: htmlDirection.value,
    'data-storefront-font-family': htmlFontFamily.value,
  },
  link: [
    ...preconnectLinks.value,
    // Preload the active first-paint shard so font-display:block can resolve sooner.
    htmlFontPreloadLink.value,
    { rel: 'icon', href: siteFavicon.value },
    { rel: 'shortcut icon', href: siteFavicon.value },
  ],
}))

onMounted(() => {
  scheduleDeferredClientWork(() => {
    void auth.ensureSession().catch(() => {})
  }, { delayMs: 6500, idleTimeoutMs: 3000 })

  scheduleDeferredClientWork(() => {
    void Promise.allSettled([
      prefetchShopCategories(),
      prefetchProductCategories(),
    ])
  }, { delayMs: 7500, idleTimeoutMs: 5000 })
})
</script>
