<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <!-- Products hero / banner (copy support spacing & background) -->
      <!-- Top horizontal Products navigation (same style as Support) -->
      <div class="products-header-spacer" aria-hidden="true"></div>
      <ProductsTopNav />

      <!-- Products page content -->
      <section class="products-content">
        <div class="products-content__inner">
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
import ProductsTopNav from '~/components/ProductsTopNav.vue'
</script>

<style scoped>
.layout--products {
  min-height: 100vh;
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

.products-hero {
  margin-top: 0;
  padding: 1.5rem 1.5rem 0.75rem;
  background: transparent !important;
}

.products-hero__inner {
  max-width: 960px;
  margin: 0 auto;
}

.products-hero__title {
  margin: 0 0 0.5rem;
  font-size: 2rem;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #f9fafb;
  display: none !important;
}

.products-content {
  padding: 1rem 1.5rem 3rem;
}

.products-content__inner {
  max-width: 960px;
  margin: 0 auto;
}

@media (max-width: 768px) {
  .products-hero {
    margin-top: 0;
    padding: 1.25rem 1.25rem 0.75rem;
  }

  .products-content {
    padding-inline: 1rem;
  }
}

@media (max-width: 400px) {
  .products-content {
    padding-inline: 0.75rem;
  }
}

@media (min-width: 768px) {
  .products-header-spacer {
    /* Desktop: push ProductsTopNav slightly below the fixed SiteHeader */
    height: 5.75rem;
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
