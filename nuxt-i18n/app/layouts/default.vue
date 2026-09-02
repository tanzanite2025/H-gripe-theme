<template>
  <div class="layout">
    <main id="main-content" class="layout-main" role="main">
      <slot />
    </main>

    <LazyAppFooter :hydrate-on-visible="footerHydrationOptions" />
    <GradientDockMenuDeferred />
    <BehaviorAttributionBootstrapDeferred />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  useRuntimeConfig,
  useHead,
  useRequestURL
} from '#imports'
import BehaviorAttributionBootstrapDeferred from '~/components/BehaviorAttributionBootstrapDeferred.vue'
import GradientDockMenuDeferred from '~/components/GradientDockMenuDeferred.vue'
import { useAuth } from '~/composables/useAuth'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useSeoSettings } from '~/composables/useSeoSettings'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'
import { normalizeSocialLinkItems, type SocialLinkViewModel } from '~/utils/socialLinks'
import { STOREFRONT_FOOTER_HYDRATION_OPTIONS } from '~/utils/storefrontLoadingPolicy'

const config = useRuntimeConfig()
const requestUrl = useRequestURL()
const auth = useAuth()
const authUser = computed<Record<string, unknown> | null>(() => (auth.user.value as Record<string, unknown> | null) ?? null)
const footerHydrationOptions = STOREFRONT_FOOTER_HYDRATION_OPTIONS

const { siteSettings: resolvedSettings } = useSiteSettings()

// Use a single source of truth for site title (Customizer preview -> API)
const { siteTitle } = useSiteTitle()
const { seoSettings } = useSeoSettings()

const defaultMetaTitle = computed(() => {
  return seoSettings.value.metaTitle || siteTitle.value
})

const siteUrl = computed(() => {
  const value = (config.public as { siteUrl?: string }).siteUrl
  if (value && value.trim().length) return value.replace(/\/$/, '')
  return requestUrl.origin.replace(/\/$/, '')
})

const defaultDescription = computed(() => {
  const fromSeo = seoSettings.value.metaDescription
  if (fromSeo.length) {
    return fromSeo
  }

  const fromSettings = (resolvedSettings.value.siteDescription || '').toString().trim()
  if (fromSettings.length) {
    return fromSettings
  }
  const value = (config.public as { siteDescription?: string }).siteDescription
  return value && value.trim().length ? value.trim() : ''
})

const siteLogo = computed(() => {
  const fromSettings = (resolvedSettings.value.siteLogo || '').toString().trim()
  if (fromSettings.length) {
    return fromSettings
  }
  const value = (config.public as { siteLogo?: string }).siteLogo
  return value && value.trim().length ? value : `${siteUrl.value}/logo.png`
})

const socialLinks = computed<SocialLinkViewModel[]>(() => {
  const apiLinks = normalizeSocialLinkItems(resolvedSettings.value.socialLinks)
  if (apiLinks.length > 0) {
    return apiLinks
  }

  const socials = (config.public as { socialLinks?: unknown }).socialLinks
  if (!Array.isArray(socials)) {
    return []
  }

  return normalizeSocialLinkItems(socials)
})

const organizationSchema = computed(() => {
  const sameAs = socialLinks.value
    .filter((item) => item.url)
    .map((item) => item.url)

  return {
    '@context': 'https://schema.org',
    '@type': 'Organization',
    name: siteTitle.value,
    url: siteUrl.value,
    description: defaultDescription.value,
    logo: siteLogo.value,
    sameAs
  }
})

const { canonicalUrl, alternateLinks, xDefaultLink } = useStorefrontSeoLinks({
  siteOrigin: siteUrl,
})

useHead(() => ({
  title: defaultMetaTitle.value,
  titleTemplate: (chunk?: string) => {
    const title = siteTitle.value.trim()
    const pageTitle = chunk?.trim()
    if (pageTitle && title && pageTitle !== title) return `${pageTitle} · ${title}`
    return pageTitle || title || defaultMetaTitle.value
  },
  link: [
    { rel: 'canonical', href: canonicalUrl.value },
    ...alternateLinks.value.map((link) => ({
      rel: 'alternate',
      hreflang: link.hreflang,
      href: link.href
    })),
    { rel: 'alternate', hreflang: 'x-default', href: xDefaultLink.value }
  ],
  meta: [
    { name: 'description', content: defaultDescription.value },
    { property: 'og:site_name', content: siteTitle.value },
    { property: 'og:type', content: 'website' },
    { property: 'og:title', content: defaultMetaTitle.value },
    { property: 'og:description', content: defaultDescription.value },
    { property: 'og:url', content: canonicalUrl.value },
    { name: 'twitter:card', content: 'summary_large_image' },
    { name: 'twitter:title', content: defaultMetaTitle.value },
    { name: 'twitter:description', content: defaultDescription.value }
  ],
  script: [createSeoJsonLdScript(organizationSchema.value)]
}))
</script>

<style scoped>
.layout {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--tz-surface-page);
}

.layout-main {
  flex: 1;
}
</style>
