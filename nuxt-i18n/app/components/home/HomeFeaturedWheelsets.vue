<template>
  <div>
    <div
      v-if="products.length > 0"
      class="home-featured-wheelsets__catalog"
      data-home-featured-wheelsets
    >
      <div class="home-featured-wheelsets__desktop-grid">
        <ShopProductDisplayCard
          v-for="product in desktopProducts"
          :key="product.id"
          :product="product"
          density="catalog"
          :show-rating="true"
          :show-view-action="true"
        />
      </div>

      <div class="home-featured-wheelsets__mobile-carousel">
        <div class="home-featured-wheelsets__mobile-grid">
          <ShopProductDisplayCard
            v-for="product in mobileVisibleProducts"
            :key="product.id"
            :product="product"
            density="catalog"
            :show-rating="true"
            :show-view-action="true"
          />
        </div>

        <div
          v-if="mobilePageCount > 1"
          class="tz-carousel-pagination home-featured-wheelsets__pagination"
          role="tablist"
          :aria-label="t('home.featuredProducts.mobilePagesLabel')"
        >
          <button
            v-for="page in mobilePageCount"
            :key="page"
            type="button"
            class="tz-carousel-pagination__dot"
            :class="{ 'is-active': activeMobilePage === page - 1 }"
            :aria-label="t('home.featuredProducts.mobilePageLabel', { page })"
            :aria-selected="activeMobilePage === page - 1"
            role="tab"
            @click="activeMobilePage = page - 1"
          />
        </div>
      </div>

      <NuxtLink
        v-if="hasAdditionalWheelsets"
        :to="shopPath"
        class="home-featured-wheelsets__all-link"
      >
        {{ moreAvailableLabel }}
        <Icon name="lucide:arrow-right" class="h-4 w-4" aria-hidden="true" />
      </NuxtLink>
    </div>

    <div v-else class="home-featured-wheelsets__state-shell">
      <StorefrontDataNotice
        v-if="pending"
        tone="empty"
        :title="t('storefrontDataNotice.featuredProducts.loading.title')"
        :description="t('storefrontDataNotice.featuredProducts.loading.description')"
      />

      <StorefrontDataNotice
        v-else-if="error"
        tone="error"
        role="alert"
        :title="t('storefrontDataNotice.featuredProducts.error.title')"
        :description="t('storefrontDataNotice.featuredProducts.error.description')"
      >
        <template #actions>
          <button
            type="button"
            class="storefront-data-notice-action"
            :disabled="pending"
            @click="emit('retry')"
          >
            <Icon name="lucide:refresh-cw" aria-hidden="true" />
            {{ t('common.retry') }}
          </button>
          <NuxtLink :to="shopPath" class="storefront-data-notice-action">
            <Icon name="lucide:shopping-bag" aria-hidden="true" />
            {{ t('common.shop') }}
          </NuxtLink>
        </template>
      </StorefrontDataNotice>

      <StorefrontDataNotice
        v-else
        tone="empty"
        :title="t('storefrontDataNotice.featuredProducts.empty.title')"
        :description="t('storefrontDataNotice.featuredProducts.empty.description')"
      >
        <template #actions>
          <NuxtLink :to="shopPath" class="storefront-data-notice-action">
            <Icon name="lucide:shopping-bag" aria-hidden="true" />
            {{ t('common.shop') }}
          </NuxtLink>
        </template>
      </StorefrontDataNotice>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from '#imports'
import StorefrontDataNotice from '~/components/StorefrontDataNotice.vue'
import ShopProductDisplayCard from '~/components/shop/ShopProductDisplayCard.vue'
import type { ShopProduct } from '~/composables/useShopProducts'

const props = defineProps<{
  products: ShopProduct[]
  pending: boolean
  error: Error | null
  hasMore: boolean
  total: number
  shopPath: string
}>()

const emit = defineEmits<{
  retry: []
}>()

const { t } = useI18n()
const activeMobilePage = ref(0)

const desktopProducts = computed(() => props.products.slice(0, 4))
const hasAdditionalWheelsets = computed(() => (
  props.hasMore || props.total > props.products.length
))
const moreAvailableLabel = computed(() => (
  props.total > props.products.length
    ? t('home.featuredProducts.moreAvailable', { count: props.total })
    : t('home.featuredProducts.moreAvailableUnknown')
))
const mobilePageCount = computed(() => (
  props.products.length <= 4
    ? 1
    : Math.ceil((props.products.length - 4) / 2) + 1
))
const mobileVisibleProducts = computed(() => {
  const start = activeMobilePage.value * 2
  return props.products.slice(start, start + 4)
})

watch(
  () => [props.products.length, props.error] as const,
  () => {
    activeMobilePage.value = 0
  },
)

watch(mobilePageCount, (pageCount) => {
  if (activeMobilePage.value >= pageCount) {
    activeMobilePage.value = Math.max(0, pageCount - 1)
  }
})
</script>

<style scoped>
.home-featured-wheelsets__state-shell {
  display: grid;
  min-height: 280px;
  align-items: center;
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 0.75rem;
  background: rgba(255, 255, 255, 0.025);
}

.home-featured-wheelsets__desktop-grid {
  display: grid;
  min-width: 0;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 1rem;
}

.home-featured-wheelsets__mobile-carousel {
  display: none;
}

.home-featured-wheelsets__pagination {
  margin-top: 0.75rem;
}

.home-featured-wheelsets__all-link {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  margin-top: 0.75rem;
  color: #b5ff6d;
  font-size: 0.8125rem;
  font-weight: 600;
  line-height: 1.25;
  text-decoration: none;
}

.home-featured-wheelsets__all-link:hover,
.home-featured-wheelsets__all-link:focus-visible {
  color: #fff;
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

@media (max-width: 767px) {
  .home-featured-wheelsets__desktop-grid {
    display: none;
  }

  .home-featured-wheelsets__mobile-carousel {
    display: block;
    min-width: 0;
  }

  .home-featured-wheelsets__mobile-grid {
    display: grid;
    min-width: 0;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-template-rows: repeat(2, minmax(0, 1fr));
    align-items: stretch;
    gap: 0.5rem;
  }

  .home-featured-wheelsets__mobile-grid > * {
    min-width: 0;
  }

  .home-featured-wheelsets__mobile-grid :deep(.shop-product-display-card) {
    width: 100%;
    min-width: 0;
  }
}

@media (min-width: 640px) {
  .home-featured-wheelsets__desktop-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (min-width: 1280px) {
  .home-featured-wheelsets__desktop-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
