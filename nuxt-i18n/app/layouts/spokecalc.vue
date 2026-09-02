<template>
  <div class="layout">
    <main id="main-content" class="layout-main" role="main">
      <!-- Dedicated layout for the spoke calculator page.
           We intentionally do NOT render the homepage hero content here
           so that the calculator UI can take the full viewport without
           being pushed down by the global header content. -->
      <slot />
    </main>

    <!-- Reuse the same deferred chrome as the storefront layouts. -->
    <LazyAppFooter :hydrate-on-visible="footerHydrationOptions" />
    <GradientDockMenuDeferred />
  </div>
</template>

<script setup lang="ts">
import GradientDockMenuDeferred from '~/components/GradientDockMenuDeferred.vue'
import { STOREFRONT_FOOTER_HYDRATION_OPTIONS } from '~/utils/storefrontLoadingPolicy'

const footerHydrationOptions = STOREFRONT_FOOTER_HYDRATION_OPTIONS
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--tz-surface-page);
}

.layout-main {
  flex: 1;
}
</style>
