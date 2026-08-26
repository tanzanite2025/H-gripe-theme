<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'
import type { ShopProductReviewSummary } from '~/composables/useShopProducts'

const props = withDefaults(defineProps<{
  summary?: ShopProductReviewSummary | null
  showCount?: boolean
  showEmpty?: boolean
  emptyLabel?: string
  size?: 'xs' | 'sm'
}>(), {
  summary: null,
  showCount: true,
  showEmpty: false,
  emptyLabel: '',
  size: 'sm',
})

const { t, locale } = useI18n()

const averageRating = computed(() => Math.min(5, Math.max(0, Number(props.summary?.averageRating || 0))))
const reviewCount = computed(() => Math.max(0, Math.floor(Number(props.summary?.totalReviews || 0))))
const hasRating = computed(() => reviewCount.value > 0 && averageRating.value > 0)
const localeForIntl = computed(() => String(locale.value || 'en').replace('_', '-'))

const formatNumber = (value: number, fractionDigits = 0) => new Intl.NumberFormat(localeForIntl.value, {
  minimumFractionDigits: fractionDigits,
  maximumFractionDigits: fractionDigits,
}).format(value)

const formattedRating = computed(() => formatNumber(averageRating.value, 1))
const formattedCount = computed(() => formatNumber(reviewCount.value))
const ratingLabel = computed(() => t('productReviews.summaryLabel', {
  rating: formattedRating.value,
  count: formattedCount.value,
}))
const resolvedEmptyLabel = computed(() => (
  props.emptyLabel || t('productReviews.empty', 'No reviews yet.')
))

const starFillPercentages = computed(() => {
  return [0, 1, 2, 3, 4].map((index) => {
    const fill = (averageRating.value - index) * 100
    return Math.min(100, Math.max(0, fill))
  })
})
</script>

<template>
  <span
    v-if="hasRating"
    class="product-rating-compact"
    :class="`product-rating-compact--${size}`"
    role="img"
    :aria-label="ratingLabel"
  >
    <span class="product-rating-compact__stars" aria-hidden="true">
      <span
        v-for="(fillPercentage, index) in starFillPercentages"
        :key="index"
        class="product-rating-compact__star"
      >
        <Icon
          name="lucide:star"
          class="product-rating-compact__star-icon product-rating-compact__star-icon--empty"
          aria-hidden="true"
        />
        <span
          class="product-rating-compact__star-fill"
          :style="{ width: `${fillPercentage}%` }"
        >
          <Icon
            name="lucide:star"
            class="product-rating-compact__star-icon product-rating-compact__star-icon--filled"
            aria-hidden="true"
          />
        </span>
      </span>
    </span>
    <span class="product-rating-compact__score">{{ formattedRating }}</span>
    <span v-if="showCount" class="product-rating-compact__count">({{ formattedCount }})</span>
  </span>
  <span
    v-else-if="showEmpty"
    class="product-rating-compact product-rating-compact--empty"
    :class="`product-rating-compact--${size}`"
    role="img"
    :aria-label="resolvedEmptyLabel"
  >
    <span class="product-rating-compact__empty-label">{{ resolvedEmptyLabel }}</span>
  </span>
</template>

<style scoped>
.product-rating-compact {
  --product-rating-star-size: 0.86rem;
  display: inline-flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  gap: 0.28rem;
  color: var(--product-rating-text, var(--tz-text-secondary));
  font-size: 0.72rem;
  font-weight: 700;
  line-height: 1;
  white-space: nowrap;
}

.product-rating-compact--xs {
  --product-rating-star-size: 0.76rem;
  gap: 0.22rem;
  font-size: 0.66rem;
}

.product-rating-compact--empty {
  width: 100%;
  justify-content: center;
  text-align: center;
  color: var(--product-rating-empty-text, var(--tz-text-muted));
}

.product-rating-compact__stars {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.08rem;
}

.product-rating-compact__empty-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.product-rating-compact__star {
  position: relative;
  display: inline-block;
  width: var(--product-rating-star-size);
  height: var(--product-rating-star-size);
  flex: 0 0 auto;
}

.product-rating-compact__star-icon {
  position: absolute;
  top: 0;
  left: 0;
  width: var(--product-rating-star-size);
  height: var(--product-rating-star-size);
}

.product-rating-compact__star-icon--empty {
  color: var(--product-rating-empty, rgba(20, 32, 43, 0.24));
  fill: transparent;
}

.product-rating-compact__star-icon--filled {
  color: var(--product-rating-star, #f59e0b);
  fill: currentColor;
}

.product-rating-compact__star-fill {
  position: absolute;
  inset: 0;
  display: block;
  overflow: hidden;
}

.product-rating-compact__score,
.product-rating-compact__count {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
