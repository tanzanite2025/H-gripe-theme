<template>
  <div class="layout layout--support">
    <main class="layout-main">
      <div class="site-header-layout-spacer" aria-hidden="true"></div>

      <!-- Support page content -->
      <section class="support-content">
        <div class="support-content__inner page-content-shell">
          <slot />
          <PageFaqSlot />
        </div>
      </section>
    </main>

    <AppFooter />
    <GradientDockMenu />
    <BehaviorAttributionBootstrap />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useHead, useRequestURL, useRuntimeConfig } from '#imports'
import AppFooter from '~/components/AppFooter.vue'
import BehaviorAttributionBootstrap from '~/components/BehaviorAttributionBootstrap.vue'
import GradientDockMenu from '~/components/GradientDockMenu.vue'
import PageFaqSlot from '~/components/PageFaqSlot.vue'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { useSiteTitle } from '~/composables/useSiteTitle'
import { createSeoJsonLdScript } from '~/utils/seo/jsonLd'

const config = useRuntimeConfig()
const requestUrl = useRequestURL()
const { siteSettings } = useSiteSettings()
const { siteTitle } = useSiteTitle()

const siteUrl = computed(() => {
  const configured = String((config.public as { siteUrl?: string }).siteUrl || '').trim()
  return (configured || requestUrl.origin).replace(/\/+$/, '')
})

const siteLogo = computed(() => {
  const configured = String(siteSettings.value.siteLogo || '').trim()
  return configured || `${siteUrl.value}/logo.png`
})

const organizationSchema = computed(() => ({
  '@context': 'https://schema.org',
  '@type': 'Organization',
  name: siteTitle.value,
  url: siteUrl.value,
  logo: siteLogo.value,
}))

useHead(() => ({
  script: [createSeoJsonLdScript(organizationSchema.value)],
}))
</script>

<style scoped>
.layout--support {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--tz-surface-page);
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.support-hero {
  margin-top: 0;
  /* 桌面端：略微压缩底部 padding，让下方 Support 导航更靠上 */
  padding: 1.5rem 1.5rem 0.75rem;
  /* 去掉单独的 hero 渐变背景，直接使用整体布局背景，避免形成一条额外的色带 */
  background: transparent !important;
}

.support-hero__inner {
  max-width: 960px;
  margin: 0 auto;
}

.support-hero__title {
  margin: 0 0 0.5rem;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: var(--tz-text-primary);
  display: none !important;
}

.support-hero__subtitle {
  margin: 0;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
}

.support-content {
  padding: var(--tz-page-content-top-gap) 0 3rem;
}

.support-content__inner {
  max-width: none;
}

@media (max-width: 768px) {
  .support-hero {
    /* 顶部固定 SiteHeader 由全局 spacer 处理，这里只负责内容与导航的间距 */
    margin-top: 0;
    padding: 1.25rem 1.25rem 0.75rem;
  }

  .support-hero__title {
    font-size: 1.5rem;
    /* 保持隐藏：移动端同样不展示 Support 标题行，避免占用垂直空间 */
    display: none;
  }

  .support-content {
    padding-inline: 0;
  }

  .support-content__inner {
    max-width: none;
  }
}

</style>
