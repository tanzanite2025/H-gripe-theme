<template>
  <section
    class="product-recommendations"
    :class="{ 'product-recommendations--compact': density === 'compact' }"
    :aria-label="sectionTitle"
  >
    <header class="product-recommendations__header">
      <h2>{{ sectionTitle }}</h2>
      <div v-if="showCarouselControls" class="product-recommendations__controls">
        <button
          type="button"
          class="tz-directional-arrow tz-directional-arrow--compact product-recommendations__control"
          :disabled="!canScrollBackward"
          :aria-label="$t('recommendations.previous', 'Previous recommended products')"
          :title="$t('recommendations.previous', 'Previous recommended products')"
          @click="scrollRecommendations(-1)"
        >
          <Icon name="lucide:chevron-left" aria-hidden="true" />
        </button>
        <button
          type="button"
          class="tz-directional-arrow tz-directional-arrow--compact product-recommendations__control"
          :disabled="!canScrollForward"
          :aria-label="$t('recommendations.next', 'Next recommended products')"
          :title="$t('recommendations.next', 'Next recommended products')"
          @click="scrollRecommendations(1)"
        >
          <Icon name="lucide:chevron-right" aria-hidden="true" />
        </button>
      </div>
    </header>

    <div
      v-if="recommendationsLoading && displayedProductCards.length === 0"
      class="product-recommendations__grid"
      aria-hidden="true"
    >
      <article
        v-for="index in skeletonCount"
        :key="`recommendation-skeleton-${index}`"
        class="product-recommendations__card product-recommendations__card--skeleton"
      >
        <span class="product-recommendations__image product-recommendations__skeleton-block"></span>
        <span class="product-recommendations__body">
          <span class="product-recommendations__skeleton-line"></span>
          <span class="product-recommendations__skeleton-line product-recommendations__skeleton-line--short"></span>
        </span>
      </article>
    </div>

    <div
      v-else-if="displayedProductCards.length > 0"
      ref="recommendationViewport"
      class="product-recommendations__viewport"
      @scroll="updateScrollControls"
    >
      <div
        class="product-recommendations__grid"
        :class="{
          'product-recommendations__grid--filled': displayedProductCards.length >= minimumVisibleCards,
          'product-recommendations__grid--carousel': showCarouselControls,
        }"
      >
        <NuxtLink
          v-for="(product, index) in displayedProductCards"
          :key="product.id"
          :to="product.url"
          class="product-recommendations__card"
          @click="trackRecommendationClick(product, index)"
        >
          <span class="product-recommendations__image">
            <img
              v-if="product.thumbnail"
              :src="product.thumbnail"
              :alt="product.title"
              loading="lazy"
            />
            <span v-else class="product-recommendations__image-placeholder" aria-hidden="true">
              <Icon name="lucide:image" />
            </span>
          </span>
          <span class="product-recommendations__body">
            <span class="product-recommendations__title">{{ product.title }}</span>
            <span v-if="product.priceLabel" class="product-recommendations__price">{{ product.priceLabel }}</span>
          </span>
        </NuxtLink>
      </div>
    </div>

    <div v-else class="product-recommendations__empty" role="status">
      <Icon name="lucide:sparkles" aria-hidden="true" />
      <span>{{ emptyText }}</span>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import type { BehaviorEventMetadata } from '~/types/behavior'
import type { RecommendationProductCard } from '~/types/recommendation'
import { useBehaviorEvents } from '~/composables/useBehaviorEvents'
import { useSmartRecommendations } from '~/composables/useSmartRecommendations'

const props = withDefaults(defineProps<{
  surface: string
  title?: string
  productId?: number | null
  categoryId?: number | null
  query?: string | null
  excludeProductIds?: Array<number | null | undefined>
  limit?: number
  density?: 'default' | 'compact'
  emptyText?: string
}>(), {
  title: '',
  productId: null,
  categoryId: null,
  query: '',
  excludeProductIds: () => [],
  limit: 6,
  density: 'default',
  emptyText: '',
})

const route = useRoute()
const { t } = useI18n()
const { track } = useBehaviorEvents()
const {
  displayedProductCards,
  recommendationsLoading,
  loadBaselineRecommendations,
  recommendationRequestId,
  recommendationAlgorithmVersion,
  recommendationSource,
} = useSmartRecommendations()

const hasRequested = ref(false)
const lastImpressionSignature = ref('')
const recommendationViewport = ref<HTMLElement | null>(null)
const canScrollBackward = ref(false)
const canScrollForward = ref(false)
const minimumVisibleCards = 5

const toPositiveInteger = (value: unknown) => {
  const numberValue = Number(value)
  if (!Number.isInteger(numberValue) || numberValue <= 0) return null
  return numberValue
}

const normalizedLimit = computed(() => {
  const limit = toPositiveInteger(props.limit) || 6
  return Math.min(Math.max(limit, 1), 12)
})

const contextProductId = computed(() => toPositiveInteger(props.productId))
const contextCategoryId = computed(() => toPositiveInteger(props.categoryId))

const excludeProductIds = computed(() => {
  const seen = new Set<number>()
  for (const value of props.excludeProductIds || []) {
    const productId = toPositiveInteger(value)
    if (productId) seen.add(productId)
  }
  if (contextProductId.value) seen.add(contextProductId.value)
  return Array.from(seen)
})

const sectionTitle = computed(() => {
  return props.title || (t('recommendations.title', 'Recommended products') as string)
})

const emptyText = computed(() => {
  return props.emptyText || (t('recommendations.empty', 'No recommendations yet') as string)
})

const skeletonCount = computed(() => Math.min(normalizedLimit.value, 6))
const showCarouselControls = computed(() => displayedProductCards.value.length > minimumVisibleCards)

const updateScrollControls = () => {
  const viewport = recommendationViewport.value
  if (!viewport) {
    canScrollBackward.value = false
    canScrollForward.value = false
    return
  }

  const maxScrollLeft = Math.max(0, viewport.scrollWidth - viewport.clientWidth)
  canScrollBackward.value = viewport.scrollLeft > 2
  canScrollForward.value = viewport.scrollLeft < maxScrollLeft - 2
}

const resetScrollControls = async () => {
  await nextTick()
  if (recommendationViewport.value) {
    recommendationViewport.value.scrollLeft = 0
  }
  updateScrollControls()
}

const scrollRecommendations = (direction: -1 | 1) => {
  const viewport = recommendationViewport.value
  if (!viewport) return

  viewport.scrollBy({
    left: direction * viewport.clientWidth,
    behavior: 'smooth',
  })
}

const requestKey = computed(() => JSON.stringify({
  surface: props.surface,
  productId: contextProductId.value,
  categoryId: contextCategoryId.value,
  query: props.query || '',
  excludeProductIds: excludeProductIds.value,
  limit: normalizedLimit.value,
  route: route.fullPath,
}))

const buildMetadata = (extra: Record<string, string | number | boolean | null | undefined> = {}) => {
  const metadata: BehaviorEventMetadata = {
    surface: props.surface,
    source: recommendationSource.value,
  }
  if (recommendationRequestId.value) metadata.request_id = recommendationRequestId.value
  if (recommendationAlgorithmVersion.value) metadata.algorithm_version = recommendationAlgorithmVersion.value
  if (contextProductId.value) metadata.context_product_id = contextProductId.value
  if (props.query) metadata.query = String(props.query).slice(0, 256)

  for (const [key, value] of Object.entries(extra)) {
    if (value !== undefined) metadata[key] = value
  }

  return metadata
}

const loadRecommendations = async () => {
  if (!import.meta.client) return
  hasRequested.value = true
  await loadBaselineRecommendations({
    surface: props.surface,
    productId: contextProductId.value,
    categoryId: contextCategoryId.value,
    query: props.query,
    route: route.fullPath,
    limit: normalizedLimit.value,
    excludeProductIds: excludeProductIds.value,
  })
}

const trackRecommendationImpression = () => {
  const cards = displayedProductCards.value
  if (!hasRequested.value || recommendationsLoading.value || cards.length === 0) return

  const signature = [
    requestKey.value,
    recommendationRequestId.value || recommendationSource.value,
    cards.map((product) => String(product.id)).join(','),
  ].join(':')
  if (signature === lastImpressionSignature.value) return

  lastImpressionSignature.value = signature
  track({
    eventType: 'recommendation_impression',
    productId: contextProductId.value || undefined,
    categoryId: contextCategoryId.value || undefined,
    metadata: buildMetadata({
      item_count: cards.length,
      item_ids: cards.map((product) => String(product.id)).join(',').slice(0, 512),
    }),
  })
}

const trackRecommendationClick = (product: RecommendationProductCard, index: number) => {
  const recommendedProductId = toPositiveInteger(product.id)
  track({
    eventType: 'recommendation_click',
    productId: recommendedProductId || undefined,
    categoryId: contextCategoryId.value || undefined,
    metadata: buildMetadata({
      position: index + 1,
      slot: product.slot,
      reason: product.reason,
      target_url: product.url,
    }),
  })
}

onMounted(() => {
  window.addEventListener('resize', updateScrollControls)
  void resetScrollControls()
  void loadRecommendations()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateScrollControls)
})

watch(requestKey, () => {
  void loadRecommendations()
})

watch(
  () => displayedProductCards.value.map((product) => String(product.id)).join(','),
  () => {
    void resetScrollControls()
  },
  { flush: 'post' }
)

watch(
  () => [
    recommendationsLoading.value,
    recommendationRequestId.value,
    recommendationSource.value,
    displayedProductCards.value.map((product) => String(product.id)).join(','),
  ],
  () => {
    trackRecommendationImpression()
  },
  { flush: 'post' }
)
</script>

<style scoped>
.product-recommendations {
  --recommendation-visible-columns: 2;
  --recommendation-card-width: calc((100% - 0.7rem) / 2);
  display: grid;
  width: 100%;
  gap: 0.9rem;
}

.product-recommendations__header {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.product-recommendations__header h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-section-title);
  font-weight: 800;
  line-height: 1.15;
}

.product-recommendations__controls {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.4rem;
}

.product-recommendations__viewport {
  min-width: 0;
  overflow-x: auto;
  scrollbar-width: none;
  scroll-behavior: smooth;
}

.product-recommendations__viewport::-webkit-scrollbar {
  display: none;
}

.product-recommendations__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.7rem;
}

.product-recommendations__grid--filled {
  grid-template-columns: repeat(var(--recommendation-visible-columns), minmax(0, 1fr));
}

.product-recommendations__grid--carousel {
  display: flex;
  align-items: stretch;
  width: max-content;
  min-width: 100%;
}

.product-recommendations__grid--carousel .product-recommendations__card {
  flex: 0 0 var(--recommendation-card-width);
}

.product-recommendations__card {
  display: grid;
  min-width: 0;
  grid-template-rows: auto 1fr;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: #050506;
  color: inherit;
  text-decoration: none;
  transition: border-color 0.2s ease, background-color 0.2s ease, transform 0.2s ease;
}

.product-recommendations__card:hover {
  border-color: rgba(181, 255, 109, 0.45);
  background: #0c0d0f;
  transform: translateY(-1px);
}

.product-recommendations__card:focus-visible {
  outline: 2px solid var(--tz-brand-primary);
  outline-offset: 3px;
}

.product-recommendations__image {
  display: block;
  aspect-ratio: 1;
  min-width: 0;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.055);
}

.product-recommendations__image img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.product-recommendations__image-placeholder {
  display: flex;
  width: 100%;
  height: 100%;
  align-items: center;
  justify-content: center;
  color: var(--tz-text-muted);
}

.product-recommendations__image-placeholder :deep(svg) {
  width: 1.45rem;
  height: 1.45rem;
}

.product-recommendations__body {
  display: grid;
  min-width: 0;
  align-content: start;
  gap: 0.42rem;
  min-height: 4.45rem;
  padding: 0.72rem;
}

.product-recommendations__title {
  display: -webkit-box;
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.82rem;
  font-weight: 750;
  line-height: 1.32;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.product-recommendations__price {
  color: var(--tz-brand-primary);
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1;
}

.product-recommendations__empty {
  display: flex;
  min-height: 8.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  box-sizing: border-box;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  background: #050506;
  color: var(--tz-text-muted);
  font-size: 0.82rem;
  font-weight: 700;
  line-height: 1.2;
  padding: 1rem;
  text-align: center;
}

.product-recommendations__empty :deep(svg) {
  width: 1rem;
  height: 1rem;
  color: var(--tz-brand-primary);
}

.product-recommendations__card--skeleton {
  pointer-events: none;
}

.product-recommendations__skeleton-block,
.product-recommendations__skeleton-line {
  position: relative;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.075);
}

.product-recommendations__skeleton-line {
  display: block;
  width: 100%;
  height: 0.76rem;
  border-radius: 999px;
}

.product-recommendations__skeleton-line--short {
  width: 48%;
}

.product-recommendations--compact {
  gap: 0.7rem;
}

.product-recommendations--compact .product-recommendations__grid {
  gap: 0.55rem;
}

.product-recommendations--compact .product-recommendations__body {
  min-height: 3.9rem;
  padding: 0.62rem;
}

@media (min-width: 640px) {
  .product-recommendations {
    --recommendation-visible-columns: 3;
    --recommendation-card-width: calc((100% - 1.4rem) / 3);
  }

  .product-recommendations__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) {
  .product-recommendations {
    --recommendation-visible-columns: 5;
    --recommendation-card-width: calc((100% - 2.8rem) / 5);
  }

  .product-recommendations__grid {
    grid-template-columns: repeat(auto-fit, minmax(9.5rem, 1fr));
  }
}
</style>
