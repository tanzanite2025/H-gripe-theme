<template>
  <div class="layout">
    <main id="main-content" class="layout-main" role="main">
      <!-- Dedicated layout for the spoke calculator page.
           We intentionally do NOT render the homepage hero content here
           so that the calculator UI can take the full viewport without
           being pushed down by the global header content. -->
      <slot />
    </main>

    <!-- Reuse the same global footer and dock menu so the overall
         site chrome stays consistent, but keep the layout itself
         as light as possible. -->
    <AppFooter />
    <GradientDockMenu />
  </div>
</template>

<script setup lang="ts">
import AppFooter from '~/components/AppFooter.vue'
import GradientDockMenu from '~/components/GradientDockMenu.vue'
</script>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #000;
}

.layout-main {
  flex: 1;
}
</style>
