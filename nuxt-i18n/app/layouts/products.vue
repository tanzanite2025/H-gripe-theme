<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <div class="products-header-spacer" aria-hidden="true"></div>

      <!-- Products page content -->
      <section class="products-content">
        <div class="products-content__inner page-content-shell">
          <slot />
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
  height: 145px;
}



.products-content {
  padding: 1rem 0 3rem;
}

.products-content__inner {
  max-width: none;
}

@media (min-width: 768px) {
  .products-header-spacer {
    height: 130px;
  }
}

/* tablet-768: 768x1024 等宽度段，使用 160px 的顶部空白 */
@media (min-width: 768px) and (max-width: 819px) {
  .products-header-spacer {
    height: 145px;
  }
}

/* tablet-820: 820x1180 等宽度段，使用 160px 的顶部空白 */
@media (min-width: 820px) and (max-width: 1023px) {
  .products-header-spacer {
    height: 145px;
  }
}
</style>
