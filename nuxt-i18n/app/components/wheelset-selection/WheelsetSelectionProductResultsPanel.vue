<template>
  <div
    class="wheelset-selection-product-results-panel"
    :class="{ 'wheelset-selection-product-results-panel--mobile-expanded': isMobileResultsExpanded }"
  >
    <header class="wheelset-selection-product-results-panel__header wheelset-selection-product-results-panel__header--desktop">
      <div class="min-w-0">
        <span>{{ categorySlug }}</span>
        <h3>{{ t('wheelsetSelectionAssistant.results.title') }}</h3>
      </div>
      <strong v-if="selectedLabel" class="truncate">{{ selectedLabel }}</strong>
    </header>

    <button
      type="button"
      class="wheelset-selection-product-results-panel__mobile-toggle"
      :aria-expanded="isMobileResultsExpanded"
      :aria-controls="mobileResultsContentId"
      @click="toggleMobileResults"
    >
      <span class="wheelset-selection-product-results-panel__mobile-toggle-copy">
        <span>{{ categorySlug }}</span>
        <strong>{{ t('wheelsetSelectionAssistant.results.title') }}</strong>
      </span>
      <span class="wheelset-selection-product-results-panel__mobile-toggle-meta">
        <span class="wheelset-selection-product-results-panel__count">
          {{ matchingProductCount }}
        </span>
        <Icon
          name="lucide:chevron-down"
          class="wheelset-selection-product-results-panel__toggle-icon"
          aria-hidden="true"
        />
      </span>
    </button>

    <div
      :id="mobileResultsContentId"
      class="wheelset-selection-product-results-panel__content"
      :class="{ 'wheelset-selection-product-results-panel__content--collapsed': !isMobileResultsExpanded }"
    >
      <div v-if="loading" class="wheelset-selection-product-results-panel__state">
        <Icon name="lucide:loader-circle" class="h-5 w-5 animate-spin" />
        <span>{{ t('wheelsetSelectionAssistant.results.loading') }}</span>
      </div>
      <div v-else-if="error" class="wheelset-selection-product-results-panel__state wheelset-selection-product-results-panel__state--error">
        <span>{{ error }}</span>
        <button
          type="button"
          class="wheelset-selection-product-results-panel__retry"
          @click="emit('retry')"
        >
          {{ t('wheelsetSelectionAssistant.states.retry') }}
        </button>
      </div>
      <div v-else-if="products.length === 0" class="wheelset-selection-product-results-panel__state">
        <span>{{ t('wheelsetSelectionAssistant.results.empty') }}</span>
      </div>
      <div v-else class="wheelset-selection-product-results-panel__grid">
        <ShopProductDisplayCard
          v-for="product in products"
          :key="product.id"
          :product="product"
          body-click-action="details"
          :show-rating="false"
          show-details-action
          show-wishlist-action
          show-share-action
          :show-view-action="false"
          @details="openWheelsetSelectionProductDetails"
          @wishlist="handleAddToWishlist"
        />
      </div>

      <nav
        class="wheelset-selection-product-results-panel__pagination"
        :aria-label="t('wheelsetSelectionAssistant.results.pagination')"
      >
        <button
          type="button"
          class="tz-directional-arrow tz-directional-arrow--small"
          :aria-label="t('wheelsetSelectionAssistant.results.previousPage')"
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
          :aria-label="t('wheelsetSelectionAssistant.results.nextPage')"
          :disabled="!hasMore || loading"
          @click="emit('nextPage')"
        >
          <Icon name="lucide:chevron-right" aria-hidden="true" />
        </button>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from '#imports'
import { useGlobalProductDetailBottomSheet } from '~/composables/useGlobalProductDetailBottomSheet'
import ShopProductDisplayCard from '~/components/shop/ShopProductDisplayCard.vue'
import { useWheelsetSelectionAccordion } from '~/composables/useWheelsetSelectionAccordion'
import { useWishlist } from '~/composables/useWishlist'
import type { ShopProduct } from '~/composables/useShopProducts'

const props = defineProps<{
  categorySlug: string
  selectedLabel?: string
  products: ShopProduct[]
  total?: number
  loading: boolean
  error?: string | null
  page: number
  hasMore: boolean
}>()

const emit = defineEmits<{
  previousPage: []
  nextPage: []
  retry: []
}>()

const { t } = useI18n()
const accordion = useWheelsetSelectionAccordion()
const { addToWishlist } = useWishlist()
const { openGlobalProductDetailBottomSheet } = useGlobalProductDetailBottomSheet()
const localMobileResultsExpanded = ref(false)
const isMobileResultsExpanded = computed(() => (
  accordion?.isExpanded('results').value ?? localMobileResultsExpanded.value
))
const mobileResultsContentId = 'wheelset-selection-results-content'
const matchingProductCount = computed(() => Math.max(0, Number(props.total ?? props.products.length)))

const openWheelsetSelectionProductDetails = (product: ShopProduct) => {
  if (!product?.slug) return

  openGlobalProductDetailBottomSheet({
    id: product.id,
    slug: product.slug,
    title: product.title,
    thumbnail: product.thumbnail,
  })
}

const handleAddToWishlist = async (product: ShopProduct) => {
  if (!product?.id) return
  await addToWishlist(product.id)
}

const toggleMobileResults = () => {
  if (accordion) {
    accordion.toggle('results')
    return
  }

  localMobileResultsExpanded.value = !localMobileResultsExpanded.value
}
</script>

<style scoped>
.wheelset-selection-product-results-panel {
  display: flex;
  box-sizing: border-box;
  height: 100%;
  min-height: 100%;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0.85rem;
}

.wheelset-selection-product-results-panel__content {
  display: flex;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 0.75rem;
}

.wheelset-selection-product-results-panel__mobile-toggle {
  display: none;
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
  color: var(--tz-site-accent, #059669);
  font-size: 0.63rem;
  font-weight: 850;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.wheelset-selection-product-results-panel__header h3 {
  margin-top: 0.2rem;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 800;
}

.wheelset-selection-product-results-panel__grid {
  --wheelset-selection-card-min-width: 16rem;
  --wheelset-selection-card-width: clamp(16rem, 100%, 18.75rem);
  display: grid;
  flex: 1 1 0;
  min-height: 0;
  min-width: 0;
  align-content: start;
  gap: 0.55rem;
  grid-template-columns: repeat(auto-fit, minmax(var(--wheelset-selection-card-min-width), 1fr));
  grid-auto-rows: max-content;
  justify-content: start;
  overflow-y: auto;
  scrollbar-gutter: stable;
  padding: 0.1rem;
}

.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card) {
  --tz-product-card-width: var(--wheelset-selection-card-width);
  width: min(100%, var(--wheelset-selection-card-width));
  height: auto;
  min-height: 0;
  align-self: start;
}

.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__image) {
  flex: 0 0 auto;
  width: 100%;
  height: auto;
  aspect-ratio: 1 / 1;
}

.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__actions) {
  gap: 0.25rem;
  padding-right: 0.4rem;
  padding-left: 0.4rem;
}

.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__wishlist-action),
.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__add-to-cart-action),
.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__view-action),
.wheelset-selection-product-results-panel__grid :deep(.shop-product-display-card__share-action) {
  width: 1.75rem;
  height: 1.75rem;
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

.wheelset-selection-product-results-panel__retry {
  min-height: 2.25rem;
  border-radius: 0.65rem;
  background: var(--tz-site-accent, #059669);
  padding: 0 0.9rem;
  color: #ffffff;
  font-size: 0.8rem;
  font-weight: 800;
}

.wheelset-selection-product-results-panel__pagination {
  display: flex;
  min-height: 2.25rem;
  flex-shrink: 0;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border-top: 1px solid var(--quickbuy-divider, var(--tz-border-subtle));
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
  background: var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
  font-size: 0.75rem;
  font-weight: 800;
}

.wheelset-selection-product-results-panel__page--active {
  color: #ffffff;
  background: var(--tz-site-accent, #059669);
}

@media (max-width: 767px) {
  .wheelset-selection-product-results-panel {
    min-height: 0;
    padding: 0.75rem;
  }

  .wheelset-selection-product-results-panel__header--desktop {
    display: none;
  }

  .wheelset-selection-product-results-panel__mobile-toggle {
    display: flex;
    width: 100%;
    min-height: 3.25rem;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    border: 1px solid var(--quickbuy-divider-strong, var(--tz-border-strong));
    border-radius: 0.75rem;
    background: var(--quickbuy-panel-surface-raised, var(--tz-surface-muted));
    padding: 0.65rem 0.75rem;
    text-align: left;
  }

  .wheelset-selection-product-results-panel__mobile-toggle-copy {
    display: grid;
    min-width: 0;
    gap: 0.2rem;
  }

  .wheelset-selection-product-results-panel__mobile-toggle-copy span {
    overflow: hidden;
    color: var(--tz-site-accent, #059669);
    font-size: 0.63rem;
    font-weight: 850;
    letter-spacing: 0.12em;
    text-overflow: ellipsis;
    text-transform: uppercase;
    white-space: nowrap;
  }

  .wheelset-selection-product-results-panel__mobile-toggle-copy strong {
    overflow: hidden;
    color: var(--tz-text-primary);
    font-size: 0.95rem;
    font-weight: 800;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .wheelset-selection-product-results-panel__mobile-toggle-meta {
    display: inline-flex;
    flex: 0 0 auto;
    align-items: center;
    gap: 0.5rem;
  }

  .wheelset-selection-product-results-panel__count {
    display: inline-flex;
    min-width: 1.75rem;
    height: 1.75rem;
    align-items: center;
    justify-content: center;
    border-radius: 999px;
    background: var(--tz-site-accent, #059669);
    color: #ffffff;
    font-size: 0.75rem;
    font-weight: 850;
  }

  .wheelset-selection-product-results-panel__toggle-icon {
    color: var(--tz-text-secondary, #cbd5e1);
    transition: transform 160ms ease;
  }

  .wheelset-selection-product-results-panel--mobile-expanded
    .wheelset-selection-product-results-panel__toggle-icon {
    transform: rotate(180deg);
  }

  .wheelset-selection-product-results-panel__content--collapsed {
    display: none;
  }

  .wheelset-selection-product-results-panel__grid {
    flex: 1 1 auto;
    --wheelset-selection-card-min-width: 0;
    --wheelset-selection-card-width: 18.75rem;
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
