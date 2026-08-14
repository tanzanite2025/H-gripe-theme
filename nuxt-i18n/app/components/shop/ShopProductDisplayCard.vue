<script setup lang="ts">
import type { ShopProduct } from '~/composables/useShopProducts'

type ShopProductDisplayCardDensity = 'catalog' | 'quick-buy'

const props = withDefaults(defineProps<{
  product: ShopProduct
  density?: ShopProductDisplayCardDensity
  selectable?: boolean
  selected?: boolean
  showDetailsAction?: boolean
  showWishlistAction?: boolean
  showViewAction?: boolean
}>(), {
  density: 'catalog',
  selectable: false,
  selected: false,
  showDetailsAction: false,
  showWishlistAction: false,
  showViewAction: true,
})

const emit = defineEmits<{
  select: [product: ShopProduct]
  details: [product: ShopProduct]
  wishlist: [product: ShopProduct]
}>()

const handleProductBodyClick = () => {
  if (props.selectable) {
    emit('select', props.product)
  }
}

const handleProductBodyKeydown = (event: KeyboardEvent) => {
  if (!props.selectable || (event.key !== 'Enter' && event.key !== ' ')) return
  event.preventDefault()
  emit('select', props.product)
}

const handleDetailsClick = () => {
  emit('details', props.product)
}

const handleWishlistClick = () => {
  emit('wishlist', props.product)
}
</script>

<template>
  <article
    class="shop-product-display-card"
    :class="[
      `shop-product-display-card--${density}`,
      { 'shop-product-display-card--selected': selectable && selected },
    ]"
  >
    <div
      class="shop-product-display-card__body"
      :class="{ 'shop-product-display-card__body--selectable': selectable }"
      :role="selectable ? 'button' : undefined"
      :tabindex="selectable ? 0 : undefined"
      :aria-pressed="selectable ? selected : undefined"
      @click="handleProductBodyClick"
      @keydown="handleProductBodyKeydown"
    >
      <span
        v-if="selectable"
        class="shop-product-display-card__selection-indicator"
        :class="{ 'shop-product-display-card__selection-indicator--selected': selected }"
        aria-hidden="true"
      >
        <Icon v-if="selected" name="lucide:check" class="h-3.5 w-3.5" />
      </span>

      <div class="shop-product-display-card__image">
        <img
          v-if="product.thumbnail"
          :src="product.thumbnail"
          :alt="product.title"
          loading="lazy"
        />
        <span v-else class="shop-product-display-card__image-empty">
          <Icon name="lucide:package" class="h-7 w-7" aria-hidden="true" />
        </span>
      </div>

      <div class="shop-product-display-card__content">
        <h3 class="shop-product-display-card__title">
          {{ product.title }}
        </h3>
        <p v-if="product.priceLabel" class="shop-product-display-card__price">
          {{ product.priceLabel }}
        </p>
      </div>
    </div>

    <button
      v-if="showDetailsAction"
      type="button"
      class="shop-product-display-card__details-action"
      :aria-label="`${$t('products.detail.viewDetails', 'View details')}: ${product.title}`"
      :title="$t('products.detail.viewDetails', 'View details')"
      @click.stop="handleDetailsClick"
    >
      <Icon name="lucide:info" class="h-4 w-4" aria-hidden="true" />
    </button>

    <div
      v-if="showWishlistAction || (showViewAction && product.url)"
      class="shop-product-display-card__actions"
    >
      <button
        v-if="showWishlistAction"
        type="button"
        class="shop-product-display-card__wishlist-action"
        :title="$t('shopPage.actions.addToWishlist', 'Add to wishlist')"
        :aria-label="$t('shopPage.actions.addToWishlist', 'Add to wishlist')"
        @click="handleWishlistClick"
      >
        <Icon name="lucide:heart" class="h-3.5 w-3.5" aria-hidden="true" />
      </button>

      <NuxtLink
        v-if="showViewAction && product.url"
        :to="product.url"
        class="shop-product-display-card__view-action"
      >
        {{ $t('shopPage.actions.view', 'View') }}
      </NuxtLink>
    </div>
  </article>
</template>

<style scoped>
.shop-product-display-card {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  overflow: hidden;
  border-radius: 0.75rem;
  background: rgba(0, 0, 0, 0.4);
  box-shadow: 8px 8px 22px rgba(0, 0, 0, 0.92);
  transition: background-color 160ms ease, box-shadow 160ms ease, transform 160ms ease;
}

.shop-product-display-card:hover {
  background: rgba(0, 0, 0, 0.6);
}

.shop-product-display-card--quick-buy {
  height: 100%;
  border: 0;
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #17171b), #0e0e11);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.026),
    inset 0 0 0 1px rgba(0, 0, 0, 0.56),
    0 8px 22px rgba(0, 0, 0, 0.24);
}

.shop-product-display-card--quick-buy:hover {
  background:
    linear-gradient(180deg, #1e1e24, #111115);
  transform: translateY(-1px);
}

.shop-product-display-card--quick-buy .shop-product-display-card__image {
  aspect-ratio: 16 / 9;
}

.shop-product-display-card--quick-buy .shop-product-display-card__content {
  padding: 0.5rem 0.625rem 0.625rem;
}

.shop-product-display-card--quick-buy .shop-product-display-card__title {
  font-size: 0.6875rem;
}

.shop-product-display-card--quick-buy .shop-product-display-card__price {
  font-size: 0.6875rem;
}

.shop-product-display-card--quick-buy.shop-product-display-card--selected {
  background:
    linear-gradient(180deg, rgba(181, 255, 109, 0.13), rgba(181, 255, 109, 0.055)),
    var(--quickbuy-panel-surface-raised, #17171b);
  box-shadow:
    inset 0 0 0 1px rgba(181, 255, 109, 0.28),
    0 0 0 3px rgba(181, 255, 109, 0.055),
    0 10px 26px rgba(0, 0, 0, 0.28);
}

.shop-product-display-card__body {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: column;
}

.shop-product-display-card__body--selectable {
  cursor: pointer;
  outline: none;
}

.shop-product-display-card__body--selectable:focus-visible {
  box-shadow: inset 0 0 0 2px rgba(181, 255, 109, 0.82);
}

.shop-product-display-card__selection-indicator {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  z-index: 2;
  display: grid;
  width: 1.25rem;
  height: 1.25rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: #0b1020;
  background: rgba(0, 0, 0, 0.64);
  box-shadow:
    inset 0 0 0 1px rgba(0, 0, 0, 0.68),
    0 3px 10px rgba(0, 0, 0, 0.28);
}

.shop-product-display-card__selection-indicator--selected {
  background: #b5ff6d;
  box-shadow:
    0 0 0 3px rgba(181, 255, 109, 0.12),
    0 4px 12px rgba(0, 0, 0, 0.3);
}

.shop-product-display-card__image {
  display: grid;
  width: 100%;
  aspect-ratio: 1;
  place-items: center;
  overflow: hidden;
  background: rgba(0, 0, 0, 0.36);
}

.shop-product-display-card__image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.shop-product-display-card__image-empty {
  display: grid;
  place-items: center;
  color: rgba(255, 255, 255, 0.45);
}

.shop-product-display-card__content {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: column;
  padding: 0.625rem 0.75rem 0.75rem;
}

.shop-product-display-card__title {
  display: -webkit-box;
  margin: 0 0 0.25rem;
  overflow: hidden;
  color: white;
  font-size: 0.75rem;
  font-weight: 600;
  line-height: 1.35;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.shop-product-display-card__price {
  margin: auto 0 0;
  color: #b5ff6d;
  font-size: 0.75rem;
  line-height: 1.25;
}

.shop-product-display-card__details-action {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  z-index: 2;
  display: grid;
  width: 1.5rem;
  height: 1.5rem;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: white;
  background: rgba(0, 0, 0, 0.7);
  box-shadow:
    inset 0 0 0 1px rgba(0, 0, 0, 0.7),
    0 4px 12px rgba(0, 0, 0, 0.28);
  transition: background-color 160ms ease, color 160ms ease;
}

.shop-product-display-card__details-action:hover {
  color: #b5ff6d;
  background: rgba(0, 0, 0, 0.82);
}

.shop-product-display-card__actions {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0 0.75rem 0.75rem;
}

.shop-product-display-card__wishlist-action {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.72);
  transition: background-color 160ms ease, color 160ms ease;
}

.shop-product-display-card__wishlist-action:hover {
  color: white;
  background: rgba(255, 255, 255, 0.15);
}

.shop-product-display-card__view-action {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  align-items: center;
  justify-content: center;
  padding: 0.375rem 0.5rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 0.25rem;
  color: white;
  background: rgba(255, 255, 255, 0.1);
  font-size: 0.6875rem;
  line-height: 1.25;
  text-align: center;
  transition: background-color 160ms ease, border-color 160ms ease;
}

.shop-product-display-card__view-action:hover {
  border-color: rgba(255, 255, 255, 0.4);
  background: rgba(255, 255, 255, 0.2);
}

@media (max-width: 767px) {
  .shop-product-display-card__content {
    padding-inline: 0.625rem;
  }

  .shop-product-display-card__actions {
    padding-inline: 0.625rem;
  }
}
</style>
