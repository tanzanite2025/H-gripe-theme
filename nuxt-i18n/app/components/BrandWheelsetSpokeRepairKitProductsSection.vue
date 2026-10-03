<template>
  <section
    id="brand-wheelset-spoke-repair-kit-products"
    class="brand-wheelset-spoke-repair-kit-products"
    :aria-labelledby="sectionTitleId"
    :aria-busy="pending || undefined"
  >
    <header class="brand-wheelset-spoke-repair-kit-products__header">
      <p class="brand-wheelset-spoke-repair-kit-products__eyebrow">
        {{ t('brandWheelsetSpokeSpecs.products.eyebrow') }}
      </p>
      <h2 :id="sectionTitleId">
        {{ t('brandWheelsetSpokeSpecs.products.title') }}
      </h2>
      <p class="brand-wheelset-spoke-repair-kit-products__description">
        {{ t('brandWheelsetSpokeSpecs.products.description') }}
      </p>
    </header>

    <div v-if="products.length" class="brand-wheelset-spoke-repair-kit-products__grid">
      <ShopProductDisplayCard
        v-for="product in products"
        :key="product.id"
        :product="product"
        density="catalog"
        :show-rating="false"
        :show-share-action="false"
        :show-wishlist-action="false"
        :show-view-action="true"
      />
    </div>

    <p
      v-if="requestError"
      class="brand-wheelset-spoke-repair-kit-products__status"
      role="status"
    >
      {{ t('brandWheelsetSpokeSpecs.products.loadError') }}
    </p>
    <p
      v-else-if="!pending && products.length === 0"
      class="brand-wheelset-spoke-repair-kit-products__status"
      role="status"
    >
      {{ t('brandWheelsetSpokeSpecs.products.emptyDescription') }}
    </p>
    <p
      v-else-if="pending"
      class="brand-wheelset-spoke-repair-kit-products__status"
      role="status"
      aria-live="polite"
    >
      {{ t('products.loading', 'Loading...') }}
    </p>

    <aside class="brand-wheelset-spoke-repair-kit-products__note">
      <strong>{{ t('brandWheelsetSpokeSpecs.shimanoNoteTitle') }}</strong>
      <p>{{ t('brandWheelsetSpokeSpecs.shimanoNoteBody') }}</p>
    </aside>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAsyncData, useI18n, useId } from '#imports'
import ShopProductDisplayCard from '~/components/shop/ShopProductDisplayCard.vue'
import { useShopProducts, type ShopProduct } from '~/composables/useShopProducts'

const { locale, t } = useI18n()
const sectionTitleId = `brand-wheelset-spoke-repair-kit-products-${useId()}`
const { fetchPublicShopProducts } = useShopProducts()

const { data: productResponse, pending, error } = await useAsyncData(
  'brand-wheelset-spoke-repair-kit-products',
  () => fetchPublicShopProducts({
    product_category: 'spoke-repair-kits',
    status: 'active',
    page_size: 4,
  }),
  {
    default: () => ({ items: [] as ShopProduct[] }),
    watch: [locale],
  },
)

const products = computed(() => (productResponse.value?.items || []).slice(0, 4))
const requestError = computed(() => Boolean(error.value))
</script>

<style scoped>
.brand-wheelset-spoke-repair-kit-products {
  display: grid;
  gap: 1rem;
  min-width: 0;
  margin-top: 0.25rem;
  scroll-margin-top: 5rem;
  padding: clamp(1.1rem, 2vw, 1.5rem);
  border: 1px solid var(--spoke-line, #dbe3ec);
  border-radius: 1.35rem;
  background:
    radial-gradient(circle at top right, rgb(37 99 235 / 0.07), transparent 38%),
    var(--spoke-panel, #fff);
}

.brand-wheelset-spoke-repair-kit-products__note {
  display: grid;
  gap: 0.3rem;
  padding: 0.9rem 1.2rem;
  border-left: 3px solid #dc2626;
  background: #fff;
}

.brand-wheelset-spoke-repair-kit-products__note strong {
  color: var(--spoke-ink, #17212b);
  font-size: 0.78rem;
  font-weight: 850;
}

.brand-wheelset-spoke-repair-kit-products__note p {
  margin: 0;
  color: var(--spoke-muted, #64748b);
  font-size: 0.76rem;
  line-height: 1.55;
}

.brand-wheelset-spoke-repair-kit-products__header {
  display: grid;
  gap: 0.45rem;
}

.brand-wheelset-spoke-repair-kit-products__eyebrow {
  margin: 0;
  color: var(--spoke-blue, #2563eb);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.16em;
  text-transform: uppercase;
}

.brand-wheelset-spoke-repair-kit-products__header h2 {
  margin: 0;
  color: var(--spoke-ink, #17212b);
  font-size: clamp(1.15rem, 1.2vw + 0.85rem, 1.55rem);
}

.brand-wheelset-spoke-repair-kit-products__description,
.brand-wheelset-spoke-repair-kit-products__status {
  margin: 0;
  color: var(--spoke-muted, #64748b);
  font-size: 0.9rem;
  line-height: 1.6;
}

.brand-wheelset-spoke-repair-kit-products__grid {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.85rem;
}

.brand-wheelset-spoke-repair-kit-products__grid > * {
  min-width: 0;
}

.brand-wheelset-spoke-repair-kit-products__grid :deep(.shop-product-display-card) {
  height: 100%;
}

@media (max-width: 980px) {
  .brand-wheelset-spoke-repair-kit-products__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .brand-wheelset-spoke-repair-kit-products {
    padding: 1rem;
  }

  .brand-wheelset-spoke-repair-kit-products__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.55rem;
  }
}
</style>
