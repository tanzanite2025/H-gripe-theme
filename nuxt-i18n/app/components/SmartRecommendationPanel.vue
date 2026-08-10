<template>
  <div class="smart-recommendation-panel">
    <section class="recommend-panel recommend-panel--categories">
      <ProductCategoryNavigationCards
        class="smart-recommendation-panel__category-navigation"
        density="compact"
        :columns="6"
        :product-categories="displayedCategories"
        :product-categories-loading="categoriesLoading"
        :heading="categoryNavigationHeading"
        :all-product-categories-label="viewAllLabel"
        @category-navigate="emit('category-click', $event)"
        @navigate="emit('view-all')"
      />
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ProductCategoryNavigationCards from '~/components/shop/ProductCategoryNavigationCards.vue'
import type { ShopCategory } from '~/composables/useShopCategories'

const props = withDefaults(defineProps<{
  categories?: ShopCategory[]
  categoriesLoading?: boolean
  categoryLimit?: number
  categoriesTitle?: string
}>(), {
  categories: () => [],
  categoriesLoading: false,
  categoryLimit: 6,
  categoriesTitle: '',
})

const emit = defineEmits<{
  (event: 'category-click', category: ShopCategory): void
  (event: 'view-all'): void
}>()

const { t } = useI18n()

const categoryNavigationHeading = computed(() => {
  return props.categoriesTitle || (t('search.recommendedCategoriesTitle', 'Browse by type') as string)
})

const viewAllLabel = computed(() => {
  return t('common.viewAll', 'View all') as string
})

const displayedCategories = computed(() => (
  props.categories
    .filter((category) => category && category.slug && category.name)
    .slice(0, props.categoryLimit)
))
</script>

<style scoped>
.smart-recommendation-panel {
  display: block;
  width: 100%;
}

.recommend-panel {
  min-width: 0;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: #020202;
  padding: 12px;
}

.recommend-panel--categories {
  display: flex;
  min-height: 0;
}

.smart-recommendation-panel__category-navigation {
  width: 100%;
  min-height: 0;
}
</style>
