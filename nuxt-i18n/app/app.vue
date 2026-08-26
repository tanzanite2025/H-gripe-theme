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
import { useHead, useI18n, useRuntimeConfig } from '#imports'
import SidePanel from './components/SidePanel.vue'
import SiteHeader from '~/components/SiteHeader.vue'
import { useAuth } from '~/composables/useAuth'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useShopCategories } from '~/composables/useShopCategories'
import localeManifest from '~/i18n/locales.manifest'

const auth = useAuth()
const runtimeConfig = useRuntimeConfig()
const { siteSettings } = useSiteSettings()
const { loadCategories: prefetchShopCategories } = useShopCategories()
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
    { rel: 'icon', href: siteFavicon.value },
    { rel: 'shortcut icon', href: siteFavicon.value },
  ],
}))

if (import.meta.server) {
  await prefetchShopCategories().catch(() => [])
}

onMounted(() => {
  void auth.ensureSession().catch(() => {})
  void prefetchShopCategories().catch(() => {})
})
</script>
