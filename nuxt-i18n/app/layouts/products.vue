<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <div class="products-header-spacer" aria-hidden="true"></div>

      <!-- Products page content -->
      <section class="products-content">
        <div class="products-content__inner page-content-shell">
          <slot />
          <PageFaqSlot />
          <PageFeedbackSlot />
        </div>
      </section>
    </main>

    <AppFooter />
    <GradientDockMenu />
    <BehaviorAttributionBootstrap />
  </div>
</template>

<script setup lang="ts">
import { useHead } from '#imports'
import { useStorefrontSeoLinks } from '~/composables/seo/useStorefrontSeoLinks'

const { canonicalUrl, alternateLinks, xDefaultLink } = useStorefrontSeoLinks()

useHead(() => ({
  link: [
    { rel: 'canonical', href: canonicalUrl.value },
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
  background: #000000;
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.products-header-spacer {
  height: 145px;
}

.products-content {
  padding: var(--tz-page-content-top-gap) 0 3rem;
}

.products-content__inner {
  max-width: none;
}

@media (min-width: 768px) {
  .products-header-spacer {
    height: 112px;
  }
}

</style>
