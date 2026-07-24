<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <div class="products-header-spacer" aria-hidden="true"></div>

      <!-- Products page content -->
      <section class="products-content">
        <div class="products-content__inner page-content-shell">
          <slot />
          <PageFaqSlot />
        </div>
      </section>
    </main>

    <AppFooter />
    <GradientDockMenu />
  </div>
</template>

<script setup lang="ts">
import AppFooter from '~/components/AppFooter.vue'
import GradientDockMenu from '~/components/GradientDockMenu.vue'
import PageFaqSlot from '~/components/PageFaqSlot.vue'
</script>

<style scoped>
.layout--products {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: radial-gradient(circle at top, #020617 0, #020617 24%, #000000 100%);
}

.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.products-header-spacer {
  height: var(--site-header-offset, 145px);
}

.products-content {
  padding: var(--tz-page-content-top-gap) 0 3rem;
}

.products-content__inner {
  max-width: none;
}

@media (min-width: 768px) {
  .products-header-spacer {
    height: var(--site-header-offset, 112px);
  }
}

</style>
