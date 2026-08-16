<template>
  <div class="wheelset-selection-product-results-panel">
    <header class="wheelset-selection-product-results-panel__header">
      <div class="min-w-0">
        <span>{{ categorySlug }}</span>
        <h3>Wheelset products</h3>
      </div>
      <strong v-if="selectedLabel" class="truncate">{{ selectedLabel }}</strong>
    </header>

    <div v-if="loading" class="wheelset-selection-product-results-panel__state">
      <Icon name="lucide:loader-circle" class="h-5 w-5 animate-spin" />
      <span>Matching wheelsets...</span>
    </div>
    <div v-else-if="error" class="wheelset-selection-product-results-panel__state wheelset-selection-product-results-panel__state--error">
      {{ error }}
    </div>
    <div v-else-if="products.length === 0" class="wheelset-selection-product-results-panel__state">
      <span>No wheelsets match these answers yet.</span>
    </div>
    <div v-else class="wheelset-selection-product-results-panel__grid">
      <article
        v-for="product in products"
        :key="product.id"
        class="wheelset-selection-product-results-panel__card"
      >
        <div class="wheelset-selection-product-results-panel__media">
          <img v-if="product.thumbnail" :src="product.thumbnail" :alt="product.title" loading="lazy">
          <Icon v-else name="lucide:circle-dashed" class="h-7 w-7" />
        </div>
        <div class="wheelset-selection-product-results-panel__title">{{ product.title }}</div>
        <div class="wheelset-selection-product-results-panel__price">
          {{ product.displayPriceLabel || product.priceLabel }}
        </div>
      </article>
    </div>

    <nav
      class="wheelset-selection-product-results-panel__pagination"
      aria-label="Wheelset product result pages"
    >
      <button
        type="button"
        class="tz-directional-arrow tz-directional-arrow--small"
        aria-label="Previous page"
        :disabled="page <= 1 || loading"
        @click="emit('previousPage')"
      >
        <Icon name="lucide:chevron-left" aria-hidden="true" />
      </button>
      <span class="wheelset-selection-product-results-panel__page wheelset-selection-product-results-panel__page--active">
        {{ page }}
      </span>
      <button
        type="button"
        class="tz-directional-arrow tz-directional-arrow--small"
        aria-label="Next page"
        :disabled="!hasMore || loading"
        @click="emit('nextPage')"
      >
        <Icon name="lucide:chevron-right" aria-hidden="true" />
      </button>
    </nav>
  </div>
</template>

<script setup lang="ts">
import type { ShopProduct } from '~/composables/useShopProducts'

defineProps<{
  categorySlug: string
  selectedLabel?: string
  products: ShopProduct[]
  loading: boolean
  error?: string | null
  page: number
  hasMore: boolean
}>()

const emit = defineEmits<{
  previousPage: []
  nextPage: []
}>()
</script>

<style scoped>
.wheelset-selection-product-results-panel {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem;
}

.wheelset-selection-product-results-panel__header {
  display: flex;
  min-height: 2.5rem;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.wheelset-selection-product-results-panel__header span,
.wheelset-selection-product-results-panel__header strong {
  color: var(--tz-brand-primary, #b5ff6d);
  font-size: 0.63rem;
  font-weight: 850;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.wheelset-selection-product-results-panel__header h3 {
  margin-top: 0.2rem;
  color: var(--tz-text-primary, #f8fafc);
  font-size: 0.95rem;
  font-weight: 800;
}

.wheelset-selection-product-results-panel__grid {
  display: grid;
  flex: 1;
  min-height: 0;
  gap: 0.55rem;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  grid-template-rows: repeat(4, minmax(0, 1fr));
}

.wheelset-selection-product-results-panel__card {
  display: grid;
  min-width: 0;
  min-height: 0;
  align-content: start;
  gap: 0.3rem;
  overflow: hidden;
  border-radius: 0.65rem;
  background: var(--quickbuy-panel-surface-raised, #22242c);
  padding: 0.4rem;
}

.wheelset-selection-product-results-panel__media {
  display: grid;
  width: 100%;
  aspect-ratio: 1.45;
  place-items: center;
  overflow: hidden;
  border-radius: 0.45rem;
  background: rgba(255, 255, 255, 0.055);
  color: rgba(255, 255, 255, 0.35);
}

.wheelset-selection-product-results-panel__media img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.wheelset-selection-product-results-panel__title {
  overflow: hidden;
  color: var(--tz-text-primary, #f8fafc);
  font-size: 0.68rem;
  font-weight: 750;
  line-height: 1.25;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wheelset-selection-product-results-panel__price {
  overflow: hidden;
  color: var(--tz-brand-primary, #b5ff6d);
  font-size: 0.65rem;
  font-weight: 800;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.wheelset-selection-product-results-panel__state {
  display: grid;
  flex: 1;
  min-height: 14rem;
  place-items: center;
  gap: 0.5rem;
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.8rem;
  text-align: center;
}

.wheelset-selection-product-results-panel__state--error {
  color: #fca5a5;
}

.wheelset-selection-product-results-panel__pagination {
  display: flex;
  min-height: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border-top: 1px solid var(--quickbuy-divider, rgba(255, 255, 255, 0.085));
  padding-top: 0.55rem;
}

.wheelset-selection-product-results-panel__page {
  display: inline-flex;
  width: 1.85rem;
  height: 1.85rem;
  align-items: center;
  justify-content: center;
  border-radius: 0.45rem;
  color: var(--tz-text-secondary, #cbd5e1);
  background: var(--quickbuy-control-surface-raised, #1a1c24);
  font-size: 0.75rem;
  font-weight: 800;
}

.wheelset-selection-product-results-panel__page--active {
  color: #101014;
  background: var(--tz-brand-primary, #b5ff6d);
}

@media (max-width: 767px) {
  .wheelset-selection-product-results-panel__grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-template-rows: repeat(6, minmax(0, 1fr));
  }
}
</style>
