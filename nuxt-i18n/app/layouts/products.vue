<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <div class="site-header-layout-spacer" aria-hidden="true"></div>

      <!-- Products page content -->
      <section class="products-content">
        <div class="products-content__inner page-content-shell">
          <slot />
          <PageFaqSlot />
          <PageFeedbackSlot />
        </div>
      </section>
    </main>

    <StorefrontFooterWithLocaleSafeDeferredHydration />
    <GradientDockMenuDeferred />
    <BehaviorAttributionBootstrapDeferred />
  </div>
</template>

<script setup lang="ts">
import { useHead, useRoute } from '#imports'
import { computed } from 'vue'
import BehaviorAttributionBootstrapDeferred from '~/components/BehaviorAttributionBootstrapDeferred.vue'
import GradientDockMenuDeferred from '~/components/GradientDockMenuDeferred.vue'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'

const { canonicalUrl, alternateLinks, xDefaultLink } = useStorefrontSeoLinks()
const route = useRoute()
const resolvedCanonicalUrl = computed(() => {
  const variant = typeof route.query.variant === 'string' ? route.query.variant.trim() : ''
  if (!variant || !/^\d+$/.test(variant) || !route.path.includes('/products/')) {
    return canonicalUrl.value
  }
  return `${canonicalUrl.value}?variant=${encodeURIComponent(variant)}`
})

useHead(() => ({
  link: [
    { rel: 'canonical', href: resolvedCanonicalUrl.value },
    ...alternateLinks.value.map((link) => ({
      rel: 'alternate',
      hreflang: link.hreflang,
      href: link.href,
    })),
    { rel: 'alternate', hreflang: 'x-default', href: xDefaultLink.value },
  ],
}))
</script>

<style scoped>
.layout--products {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--tz-surface-page);
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.products-content {
  padding: var(--tz-page-content-top-gap) 0 3rem;
}

.products-content__inner {
  max-width: none;
}

</style>
