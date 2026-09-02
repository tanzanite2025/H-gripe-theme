<template>
  <div class="tz-light-theme">
    <NuxtLoadingIndicator
      color="#059669"
      :height="3"
      :throttle="80"
    />
    <SiteHeader ref="siteHeaderRef" />
    <NuxtLayout>
      <!-- Render the current page inside the active layout -->
      <NuxtPage />
    </NuxtLayout>
    
    <StorefrontClientOverlaysDeferred />
    
    <!-- Cookie 同意弹窗 -->
    <CookieConsentDeferred />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useHead, useI18n, useRequestURL, useRuntimeConfig } from '#imports'
import SiteHeader from '~/components/SiteHeader.vue'
import { useAuth } from '~/composables/useAuth'
import { useProductCategories } from '~/composables/useProductCategories'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useShopCategories } from '~/composables/useShopCategories'
import localeManifest from '~/i18n/locales.manifest'
import CookieConsentDeferred from '~/components/CookieConsentDeferred.vue'
import StorefrontClientOverlaysDeferred from '~/components/StorefrontClientOverlaysDeferred.vue'
import { scheduleDeferredClientWork } from '~/utils/clientDeferredWork'
import {
  createStorefrontPreconnectLinks,
  MAX_STOREFRONT_PRECONNECT_ORIGINS,
  STOREFRONT_CATEGORY_PREFETCH_WARMUP,
  STOREFRONT_SESSION_WARMUP,
  splitStorefrontConfiguredOrigins,
} from '~/utils/storefrontLoadingPolicy'
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

const preconnectLinks = computed(() => {
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
    ...splitStorefrontConfiguredOrigins(publicConfig.imageDomains),
  ]

  return createStorefrontPreconnectLinks(candidates, currentOrigin, {
    siteOrigin: publicConfig.siteUrl,
    maxOrigins: MAX_STOREFRONT_PRECONNECT_ORIGINS,
  })
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
  }, STOREFRONT_SESSION_WARMUP)

  scheduleDeferredClientWork(() => {
    void Promise.allSettled([
      prefetchShopCategories(),
      prefetchProductCategories(),
    ])
  }, STOREFRONT_CATEGORY_PREFETCH_WARMUP)
})
</script>
