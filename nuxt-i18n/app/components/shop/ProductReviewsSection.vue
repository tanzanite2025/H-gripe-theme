<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
import { useI18n } from '#imports'
import {
  useProductReviews,
  type ProductReview,
} from '~/composables/useProductReviews'
import type { ShopProductReviewSummary } from '~/composables/useShopProducts'

const props = withDefaults(defineProps<{
  productId: number
  initialSummary?: ShopProductReviewSummary | null
  compact?: boolean
}>(), {
  initialSummary: null,
  compact: false,
})

const { t, locale } = useI18n()
const {
  summary,
  reviews,
  pagination,
  isLoading,
  isLoadingMore,
  error,
  hasMore,
  loadProductReviews,
  loadMoreProductReviews,
} = useProductReviews()

summary.value = props.initialSummary || null

const headingId = computed(() => `product-reviews-heading-${props.productId}`)
const averageRating = computed(() => Math.min(5, Math.max(0, Number(summary.value?.averageRating || 0))))
const roundedAverageRating = computed(() => Math.round(averageRating.value))
const reviewCount = computed(() => Math.max(
  Number(summary.value?.totalReviews || 0),
  Number(pagination.value.total || 0),
))

const ratingDistribution = computed(() => {
  const currentSummary = summary.value
  return [5, 4, 3, 2, 1].map((rating) => ({
    rating,
    count: Number(currentSummary?.[`rating${rating}Count` as keyof ShopProductReviewSummary] || 0),
  }))
})

const ratingBarWidth = (count: number) => {
  if (!reviewCount.value) return '0%'
  return `${Math.min(100, Math.round((count / reviewCount.value) * 100))}%`
}

const localeForIntl = computed(() => String(locale.value || 'en').replace('_', '-'))

const formatAverageRating = (value: number) => new Intl.NumberFormat(localeForIntl.value, {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
}).format(value)

const formatReviewDate = (value: string) => {
  if (!value) return ''
  const parsed = new Date(value)
  if (Number.isNaN(parsed.getTime())) return ''
  return new Intl.DateTimeFormat(localeForIntl.value, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  }).format(parsed)
}

const reviewRatingLabel = (review: ProductReview) => t('productReviews.ratingLabel', {
  rating: review.rating,
})

const summaryRatingLabel = computed(() => t('productReviews.summaryLabel', {
  rating: formatAverageRating(averageRating.value),
  count: reviewCount.value,
}))

const loadReviews = () => {
  void loadProductReviews(props.productId, {
    initialSummary: props.initialSummary,
    pageSize: props.compact ? 3 : 5,
    refreshSummary: false,
  })
}

onMounted(loadReviews)

watch(
  () => props.productId,
  () => loadReviews(),
)

watch(
  () => props.initialSummary,
  (nextSummary) => {
    if (!summary.value || !reviews.value.length) {
      summary.value = nextSummary || null
    }
  },
)
</script>

<template>
  <section
    class="product-reviews"
    :class="{ 'product-reviews--compact': compact }"
    :aria-labelledby="headingId"
  >
    <header class="product-reviews__header">
      <div>
        <p class="product-reviews__eyebrow">{{ t('productReviews.eyebrow') }}</p>
        <h2 :id="headingId" class="product-reviews__title">{{ t('productReviews.title') }}</h2>
      </div>
      <div v-if="summary" class="product-reviews__summary" role="img" :aria-label="summaryRatingLabel">
        <strong class="product-reviews__average">{{ formatAverageRating(averageRating) }}</strong>
        <span class="product-reviews__stars" aria-hidden="true">
          <Icon
            v-for="star in 5"
            :key="star"
            name="lucide:star"
            :class="{ 'product-reviews__star--filled': star <= roundedAverageRating }"
          />
        </span>
        <span class="product-reviews__count">
          {{ t('productReviews.reviewCount', { count: reviewCount }) }}
        </span>
      </div>
    </header>

    <div v-if="summary && reviewCount > 0" class="product-reviews__overview">
      <div class="product-reviews__distribution" aria-label="Rating distribution">
        <div v-for="item in ratingDistribution" :key="item.rating" class="product-reviews__distribution-row">
          <span>{{ item.rating }}</span>
          <Icon name="lucide:star" aria-hidden="true" />
          <span class="product-reviews__distribution-track">
            <span
              class="product-reviews__distribution-fill"
              :style="{ width: ratingBarWidth(item.count) }"
            ></span>
          </span>
          <span class="product-reviews__distribution-count">{{ item.count }}</span>
        </div>
      </div>
    </div>

    <div v-if="isLoading" class="product-reviews__state" role="status">
      <Icon name="lucide:loader-circle" class="product-reviews__spinner" aria-hidden="true" />
      <span>{{ t('productReviews.loading') }}</span>
    </div>

    <div v-else-if="error" class="product-reviews__state product-reviews__state--error" role="alert">
      <span>{{ t('productReviews.loadError') }}</span>
      <button type="button" class="product-reviews__retry" @click="loadReviews">
        <Icon name="lucide:refresh-cw" aria-hidden="true" />
        <span>{{ t('productReviews.retry') }}</span>
      </button>
    </div>

    <div v-else-if="!reviews.length" class="product-reviews__state">
      <Icon name="lucide:message-square" aria-hidden="true" />
      <span>{{ t('productReviews.empty') }}</span>
    </div>

    <div v-else class="product-reviews__list">
      <article v-for="review in reviews" :key="review.id" class="product-reviews__item">
        <header class="product-reviews__item-header">
          <div>
            <div
              class="product-reviews__item-stars"
              role="img"
              :aria-label="reviewRatingLabel(review)"
            >
              <Icon
                v-for="star in 5"
                :key="star"
                name="lucide:star"
                :class="{ 'product-reviews__star--filled': star <= review.rating }"
                aria-hidden="true"
              />
            </div>
            <h3 v-if="review.title" class="product-reviews__item-title">{{ review.title }}</h3>
          </div>
          <time
            v-if="formatReviewDate(review.createdAt)"
            class="product-reviews__item-date"
            :datetime="review.createdAt"
          >
            {{ formatReviewDate(review.createdAt) }}
          </time>
        </header>

        <p v-if="review.content" class="product-reviews__item-content">{{ review.content }}</p>

        <div v-if="review.images.length" class="product-reviews__images">
          <img
            v-for="image in review.images"
            :key="image"
            :src="image"
            :alt="t('productReviews.imageAlt')"
            loading="lazy"
          />
        </div>

        <div class="product-reviews__item-meta">
          <span v-if="review.verified" class="product-reviews__verified">
            <Icon name="lucide:badge-check" aria-hidden="true" />
            {{ t('productReviews.verifiedPurchase') }}
          </span>
          <span v-if="review.helpfulCount > 0" class="product-reviews__helpful">
            {{ t('productReviews.helpfulCount', { count: review.helpfulCount }) }}
          </span>
        </div>

        <div v-if="review.replyContent" class="product-reviews__reply">
          <strong>{{ t('productReviews.storeReply') }}</strong>
          <p>{{ review.replyContent }}</p>
        </div>
      </article>
    </div>

    <button
      v-if="hasMore && !isLoading && !error"
      type="button"
      class="product-reviews__load-more"
      :disabled="isLoadingMore"
      @click="loadMoreProductReviews"
    >
      <Icon
        v-if="isLoadingMore"
        name="lucide:loader-circle"
        class="product-reviews__spinner"
        aria-hidden="true"
      />
      <span>{{ isLoadingMore ? t('productReviews.loadingMore') : t('productReviews.loadMore') }}</span>
    </button>
  </section>
</template>

<style scoped>
.product-reviews {
  display: grid;
  gap: 1rem;
  border-top: 1px solid var(--tz-border-subtle);
  padding-top: 1.75rem;
  color: var(--tz-text-primary);
}

.product-reviews--compact {
  gap: 0.8rem;
  padding-top: 1rem;
}

.product-reviews__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.product-reviews__eyebrow {
  margin: 0 0 0.25rem;
  color: var(--tz-text-muted);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.product-reviews__title {
  margin: 0;
  font-size: 1.45rem;
  font-weight: 700;
}

.product-reviews__summary {
  display: grid;
  min-width: 10rem;
  justify-items: end;
  gap: 0.2rem;
}

.product-reviews__average {
  color: #059669;
  font-size: 1.45rem;
  line-height: 1;
}

.product-reviews__stars,
.product-reviews__item-stars {
  display: inline-flex;
  gap: 0.12rem;
  color: var(--tz-text-disabled);
}

.product-reviews__stars :deep(svg),
.product-reviews__item-stars :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
  fill: transparent;
}

.product-reviews__star--filled {
  color: #fbbf24;
}

.product-reviews__stars :deep(.product-reviews__star--filled),
.product-reviews__item-stars :deep(.product-reviews__star--filled) {
  fill: currentColor;
}

.product-reviews__count,
.product-reviews__item-date,
.product-reviews__helpful {
  color: var(--tz-text-secondary);
  font-size: 0.75rem;
}

.product-reviews__overview {
  display: grid;
  grid-template-columns: minmax(0, 22rem);
}

.product-reviews__distribution {
  display: grid;
  gap: 0.35rem;
}

.product-reviews__distribution-row {
  display: grid;
  grid-template-columns: 0.75rem 0.85rem minmax(0, 1fr) 2rem;
  align-items: center;
  gap: 0.35rem;
  color: var(--tz-text-secondary);
  font-size: 0.72rem;
}

.product-reviews__distribution-row > svg {
  width: 0.75rem;
  height: 0.75rem;
  color: #fbbf24;
  fill: currentColor;
}

.product-reviews__distribution-track {
  display: block;
  height: 0.35rem;
  overflow: hidden;
  border-radius: 999px;
  background: var(--tz-border-subtle);
}

.product-reviews__distribution-fill {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: #059669;
}

.product-reviews__distribution-count {
  text-align: right;
}

.product-reviews__state {
  display: grid;
  min-height: 7rem;
  place-items: center;
  align-content: center;
  gap: 0.55rem;
  color: var(--tz-text-secondary);
  text-align: center;
}

.product-reviews__state--error {
  color: var(--tz-status-danger-text);
}

.product-reviews__spinner {
  animation: product-reviews-spin 0.9s linear infinite;
}

.product-reviews__retry,
.product-reviews__load-more {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  min-height: 2.25rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 0.5rem;
  color: var(--tz-text-primary);
  background: var(--tz-surface-muted);
  padding: 0.45rem 0.75rem;
  font-size: 0.78rem;
  font-weight: 700;
}

.product-reviews__retry :deep(svg),
.product-reviews__load-more :deep(svg) {
  width: 0.9rem;
  height: 0.9rem;
}

.product-reviews__list {
  display: grid;
  gap: 0.75rem;
}

.product-reviews__item {
  display: grid;
  gap: 0.65rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.7rem;
  background: var(--tz-card-surface);
  padding: 0.9rem;
}

.product-reviews__item-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.75rem;
}

.product-reviews__item-stars {
  color: var(--tz-text-disabled);
}

.product-reviews__item-title {
  margin: 0.4rem 0 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 700;
}

.product-reviews__item-content {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.88rem;
  line-height: 1.6;
  white-space: pre-wrap;
}

.product-reviews__images {
  display: flex;
  gap: 0.45rem;
  overflow-x: auto;
}

.product-reviews__images img {
  width: 4.5rem;
  height: 4.5rem;
  flex: 0 0 auto;
  border-radius: 0.45rem;
  object-fit: cover;
}

.product-reviews__item-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.65rem;
  align-items: center;
}

.product-reviews__verified {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  color: #059669;
  font-size: 0.75rem;
  font-weight: 700;
}

.product-reviews__verified :deep(svg) {
  width: 0.9rem;
  height: 0.9rem;
}

.product-reviews__reply {
  display: grid;
  gap: 0.25rem;
  border-left: 2px solid rgba(5, 150, 105, 0.68);
  color: var(--tz-text-secondary);
  padding-left: 0.7rem;
  font-size: 0.78rem;
  line-height: 1.5;
}

.product-reviews__reply strong {
  color: #059669;
}

.product-reviews__reply p {
  margin: 0;
}

.product-reviews__load-more {
  justify-self: start;
}

.product-reviews__retry:focus-visible,
.product-reviews__load-more:focus-visible {
  outline: 2px solid #059669;
  outline-offset: 3px;
}

.product-reviews__retry:hover,
.product-reviews__load-more:hover:not(:disabled) {
  border-color: rgba(5, 150, 105, 0.62);
  background: rgba(5, 150, 105, 0.12);
}

.product-reviews__load-more:disabled {
  cursor: wait;
  opacity: 0.62;
}

@keyframes product-reviews-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 640px) {
  .product-reviews__header {
    display: grid;
  }

  .product-reviews__summary {
    justify-items: start;
  }

  .product-reviews__item-header {
    display: grid;
  }
}
</style>
