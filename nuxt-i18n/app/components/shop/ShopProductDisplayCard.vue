<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLocalePath } from '#imports'
import ProductRatingCompact from '~/components/shop/ProductRatingCompact.vue'
import ProductSharePopover from '~/components/shop/ProductSharePopover.vue'
import { resolveShopProductImage, type ShopProduct } from '~/composables/useShopProducts'

type ShopProductDisplayCardDensity = 'catalog' | 'quick-buy'

const props = withDefaults(defineProps<{
  product: ShopProduct
  density?: ShopProductDisplayCardDensity
  selectable?: boolean
  selected?: boolean
  showDetailsAction?: boolean
  showRating?: boolean
  showShareAction?: boolean
  showWishlistAction?: boolean
  showViewAction?: boolean
}>(), {
  density: 'catalog',
  selectable: false,
  selected: false,
  showDetailsAction: false,
  showRating: true,
  showShareAction: true,
  showWishlistAction: false,
  showViewAction: true,
})

const emit = defineEmits<{
  select: [product: ShopProduct]
  details: [product: ShopProduct]
  wishlist: [product: ShopProduct]
  view: [product: ShopProduct]
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

const handleViewClick = () => {
  emit('view', props.product)
}

const shareDialogOpen = ref(false)
const shareButtonElement = ref<HTMLElement | null>(null)
const localePath = useLocalePath()

const handleShareClick = () => {
  shareDialogOpen.value = !shareDialogOpen.value
}

const handleShareDialogClose = () => {
  shareDialogOpen.value = false
}

const productImageSrc = computed(() => resolveShopProductImage(props.product, 'card'))
const productWeightGrams = computed(() => {
  const productWeight = Number(props.product.weightGrams || 0)
  if (productWeight > 0) return Math.round(productWeight)

  const defaultVariant = props.product.variants.find(variant => variant.isDefault) || props.product.variants[0] || null
  const variantWeight = Number(defaultVariant?.weightGrams || 0)
  return variantWeight > 0 ? Math.round(variantWeight) : 0
})
const productWeightLabel = computed(() => {
  const grams = productWeightGrams.value
  if (grams <= 0) return ''
  if (grams < 1000) return `${grams}g`

  const kilograms = grams / 1000
  return `${kilograms.toFixed(kilograms >= 10 ? 1 : 2).replace(/\.?0+$/, '')}kg`
})

const productDetailUrl = computed(() => {
  const rawUrl = String(props.product.url || '').trim()
  if (!rawUrl) return ''
  if (/^https?:\/\//i.test(rawUrl)) return rawUrl
  if (/^\/[a-z]{2}(?:[_-][a-z]{2})?\/shop\/[^/?#]+(?:[?#].*)?$/i.test(rawUrl)) return rawUrl
  return localePath(rawUrl)
})
</script>

<template>
  <article
    class="shop-product-display-card tz-product-card"
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

      <div class="shop-product-display-card__image tz-product-card__image">
        <StorefrontImage
          v-if="productImageSrc"
          :src="productImageSrc"
          :alt="product.title"
          preset="card"
        />
        <span v-else class="shop-product-display-card__image-empty">
          <Icon name="lucide:package" class="h-7 w-7" aria-hidden="true" />
        </span>
      </div>

      <div class="shop-product-display-card__content tz-product-card__body">
        <h3 class="shop-product-display-card__title">
          {{ product.title }}
        </h3>
        <ProductRatingCompact
          v-if="showRating"
          class="shop-product-display-card__rating"
          :summary="product.reviewSummary"
          show-empty
          :size="density === 'quick-buy' ? 'xs' : 'sm'"
          :show-count="density !== 'quick-buy'"
        />
        <div
          v-if="productWeightLabel || product.priceLabel"
          class="shop-product-display-card__facts"
        >
          <span
            v-if="productWeightLabel"
            class="shop-product-display-card__fact shop-product-display-card__fact--weight"
            :aria-label="`${$t('quickBuy.summary.weight', 'Weight')}: ${productWeightLabel}`"
            :title="`${$t('quickBuy.summary.weight', 'Weight')}: ${productWeightLabel}`"
          >
            <Icon name="lucide:weight" class="h-3.5 w-3.5" aria-hidden="true" />
            <span class="shop-product-display-card__fact-value">{{ productWeightLabel }}</span>
          </span>
          <span
            v-else
            class="shop-product-display-card__fact-placeholder"
            aria-hidden="true"
          />

          <span
            v-if="product.priceLabel"
            class="shop-product-display-card__fact shop-product-display-card__fact--price"
            :aria-label="`${$t('quickBuy.summary.price', 'Price')}: ${product.priceLabel}`"
            :title="`${$t('quickBuy.summary.price', 'Price')}: ${product.priceLabel}`"
          >
            <Icon name="lucide:badge-dollar-sign" class="h-3.5 w-3.5" aria-hidden="true" />
            <span class="shop-product-display-card__fact-value">{{ product.priceLabel }}</span>
          </span>
        </div>
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
      v-if="showWishlistAction || (showViewAction && product.url) || (showShareAction && product.url)"
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

      <!-- Product-detail link: opens the current product's detail page. -->
      <NuxtLink
        v-if="showViewAction && product.url"
        :to="productDetailUrl"
        class="shop-product-display-card__view-action"
        :aria-label="`${$t('products.detail.viewDetails', 'View product details')}: ${product.title}`"
        :title="`${$t('products.detail.viewDetails', 'View product details')}: ${product.title}`"
        @click="handleViewClick"
      >
        <Icon name="lucide:eye" class="h-4 w-4" aria-hidden="true" />
      </NuxtLink>

      <div
        v-if="showShareAction && product.url"
        class="shop-product-display-card__share-group"
      >
        <button
          ref="shareButtonElement"
          type="button"
          class="shop-product-display-card__share-action"
          :class="{ 'shop-product-display-card__share-action--active': shareDialogOpen }"
          :aria-expanded="shareDialogOpen"
          :aria-label="`${$t('common.share', 'Share')}: ${product.title}`"
          :title="`${$t('common.share', 'Share')}: ${product.title}`"
          @click.stop="handleShareClick"
        >
          <Icon name="lucide:share-2" class="h-4 w-4" aria-hidden="true" />
        </button>

        <ProductSharePopover
          v-if="shareDialogOpen"
          :product="product"
          :anchor-el="shareButtonElement"
          @close="handleShareDialogClose"
        />
      </div>
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
  --tz-product-card-width: 100%;
  --tz-product-card-content-height: auto;
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
  height: auto;
}

.shop-product-display-card--quick-buy .shop-product-display-card__content {
  height: auto;
  min-height: 0;
  padding: 0.5rem 0.625rem 0.625rem;
}

.shop-product-display-card--quick-buy .shop-product-display-card__title {
  font-size: 0.6875rem;
}

.shop-product-display-card--quick-buy .shop-product-display-card__facts {
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
  --product-rating-text: rgba(226, 232, 240, 0.72);
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

.shop-product-display-card__rating {
  margin: 0.05rem 0 0.4rem;
}

.shop-product-display-card__facts {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  gap: 0.5rem;
  margin: auto 0 0;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.25;
}

.shop-product-display-card__fact,
.shop-product-display-card__fact-placeholder {
  min-width: 0;
}

.shop-product-display-card__fact {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.shop-product-display-card__fact--weight {
  color: rgba(226, 232, 240, 0.78);
}

.shop-product-display-card__fact--price {
  justify-content: flex-end;
  color: #b5ff6d;
}

.shop-product-display-card__fact-value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  justify-content: center;
  gap: 0.625rem;
  padding: 0 0.75rem 0.75rem;
}

.shop-product-display-card__share-group {
  position: relative;
  display: inline-flex;
  min-width: 0;
  flex: 0 1 auto;
  align-items: center;
  justify-content: center;
}

.shop-product-display-card__wishlist-action,
.shop-product-display-card__view-action,
.shop-product-display-card__share-action {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid rgba(255, 255, 255, 0.25);
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.72);
  background: rgba(255, 255, 255, 0.08);
  transition: background-color 160ms ease, color 160ms ease;
}

.shop-product-display-card__wishlist-action:hover,
.shop-product-display-card__view-action:hover,
.shop-product-display-card__share-action:hover {
  color: white;
  background: rgba(255, 255, 255, 0.15);
}

.shop-product-display-card__share-action:hover,
.shop-product-display-card__share-action--active {
  border-color: var(--tz-brand-primary, #b5ff6d);
  color: var(--tz-brand-primary, #b5ff6d);
  background: rgba(181, 255, 109, 0.1);
}

.shop-product-display-card__view-action {
  color: white;
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
