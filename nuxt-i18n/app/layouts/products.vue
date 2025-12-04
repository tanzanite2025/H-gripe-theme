<template>
  <div class="layout layout--products">
    <main class="layout-main">
      <!-- Products hero / banner (copy support spacing & background) -->
      <section class="products-hero">
        <div class="products-hero__inner">
          <h1 class="products-hero__title">Products</h1>
        </div>
      </section>

      <!-- Top horizontal Products navigation (same style as Support) -->
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

.products-hero {
  margin-top: 2.75rem;
  padding: 3rem 1.5rem 0.75rem;
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
    margin-top: 5.5rem;
    padding: 2rem 1.25rem 0.75rem;
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

/* Mirror Support layout's extra spacing for key mobile widths */
@media (min-width: 360px) and (max-width: 374px) {
  .products-hero {
    margin-top: 7rem;
  }
}

@media (min-width: 375px) and (max-width: 389px) {
  .products-hero {
    margin-top: 7rem;
  }
}

@media (min-width: 390px) and (max-width: 411px) {
  .products-hero {
    margin-top: 7rem;
  }
}

@media (min-width: 412px) and (max-width: 413px) {
  .products-hero {
    margin-top: 7rem;
  }
}

@media (min-width: 414px) and (max-width: 429px) {
  .products-hero {
    margin-top: 7.5rem;
  }
}

@media (min-width: 430px) and (max-width: 767px) {
  .products-hero {
    margin-top: 7rem;
  }
}

@media (min-width: 768px) and (max-width: 1023px) {
  .products-hero {
    margin-top: 5.5rem;
  }
}

/* Very narrow tall phones (e.g. iPhone SE portrait): ensure horizontal nav is fully visible below fixed header */
@media (max-width: 359px) and (min-height: 700px) {
  .products-hero {
    margin-top: 7rem;
  }
}

/* Tall 20:9 Android phones (e.g. Pixel 8a) where the fixed header is slightly higher:
   add extra spacing so the horizontal products nav is fully visible below the header. */
@media (min-width: 400px) and (max-width: 430px) and (min-height: 850px) {
  .products-hero {
    margin-top: 8.25rem;
  }
}
</style>
