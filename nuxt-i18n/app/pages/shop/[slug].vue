<template>
  <ProductCategoryPage
    v-if="category"
    :category-route-path="route.path"
  />
</template>

<script setup lang="ts">
import {
  createError,
  useRoute,
} from '#imports'
import ProductCategoryPage from '~/components/shop/ProductCategoryPage.vue'
import {
  useProductCategories,
  type ProductCategory,
} from '~/composables/useProductCategories'

definePageMeta({
  layout: 'products',
})

const route = useRoute()
const { categories, loadCategories } = useProductCategories()

const normalizePath = (value: unknown): string => {
  const path = (String(value || '').split(/[?#]/, 1)[0] || '').trim()
  if (!path) return '/'
  const normalized = `/${path.replace(/^\/+|\/+$/g, '')}`
  return normalized === '//' ? '/' : normalized
}

await loadCategories()

const category = computed<ProductCategory | null>(() => (
  categories.value.find((item) => normalizePath(item.routePath) === normalizePath(route.path)) || null
))

if (!category.value) {
  throw createError({
    statusCode: 404,
    statusMessage: 'Product category not found',
  })
}
</script>
