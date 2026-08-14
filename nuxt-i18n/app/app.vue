<template>
  <div>
    <NuxtLoadingIndicator
      color="repeating-linear-gradient(90deg, #22d3ee 0%, #8b5cf6 50%, #22d3ee 100%)"
      :height="3"
      :throttle="80"
    />
    <SiteHeader ref="siteHeaderRef" />
    <NuxtLayout>
      <SidePanel>
        <template #left>
          <AccountSidebarPanel />
        </template>
      </SidePanel>
      <!-- Render the current page inside the active layout -->
      <NuxtPage />
    </NuxtLayout>
    
    <!-- 购物车与搜索面板 -->
    <LazyCartDrawer />
    <LazyCheckoutModal />
    <LazyShopSearchSheet />
    <LazyGlobalProductDetailBottomSheet />
    
    <!-- 全局聊天弹窗 -->
    <LazyWhatsAppChatModal
      v-if="currentConversation"
      :conversation="currentConversation"
      @close="closeChat"
    />
    
    <!-- Cookie 同意弹窗 -->
    <CookieConsent />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useHead, useI18n } from '#imports'
import SidePanel from './components/SidePanel.vue'
import AccountSidebarPanel from '~/components/account/AccountSidebarPanel.vue'
import SiteHeader from '~/components/SiteHeader.vue'
import { useChatWidget } from '~/composables/useChatWidget'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useShopCategories } from '~/composables/useShopCategories'
import localeManifest from '~/i18n/locales.manifest'

// 全局聊天状态
const { currentConversation, closeChat } = useChatWidget()
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
const siteFavicon = computed(() => {
  const configuredFavicon = (siteSettings.value.siteFavicon || '').toString().trim()
  if (configuredFavicon) return configuredFavicon

  const configuredLogo = (siteSettings.value.siteLogo || '').toString().trim()
  return configuredLogo || '/favicon.svg'
})

useHead(() => ({
  htmlAttrs: {
    lang: htmlLanguage.value,
    dir: htmlDirection.value,
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
  void prefetchShopCategories().catch(() => {})
})
</script>
