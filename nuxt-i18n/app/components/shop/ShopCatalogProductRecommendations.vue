<template>
  <ProductRecommendations
    :surface="surface"
    :title="$t('recommendations.shopTitle', 'Recommended products')"
    :query="normalizedSearchQuery"
    :exclude-product-ids="visibleCatalogProductIds"
    :limit="8"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import ProductRecommendations from '~/components/shop/ProductRecommendations.vue'
import type { ShopProduct } from '~/composables/useShopProducts'

const props = withDefaults(defineProps<{
  surface: 'shop_index_bottom' | 'shop_category_bottom'
  visibleProducts: ShopProduct[]
  searchQuery?: string | null
}>(), {
  searchQuery: '',
})

const normalizedSearchQuery = computed(() => String(props.searchQuery || '').trim())
const visibleCatalogProductIds = computed(() => props.visibleProducts
  .map(product => product.id)
  .filter(productId => Number.isInteger(productId) && productId > 0))

// Recommendation API category IDs refer to legacy specification templates,
// so catalog category IDs must not be forwarded as recommendation category IDs.
</script>
