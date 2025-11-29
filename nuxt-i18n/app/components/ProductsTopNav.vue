<template>
  <nav class="products-top-nav" aria-label="Products navigation">
    <div class="products-top-nav__scroll">
      <NuxtLink
        v-for="item in items"
        :key="item.id"
        class="products-top-nav__link"
        :class="{ 'products-top-nav__link--active': isActive(item) }"
        :to="localePath(item.to)"
      >
        {{ $t(item.labelKey) }}
      </NuxtLink>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath, useRoute } from '#imports'
import type { ProductsNavItem } from '~/utils/productsNav'
import { productsNavItems } from '~/utils/productsNav'
import { companyNavItems } from '~/utils/companyNav'

const props = defineProps<{
  /** Optional override for products nav items. */
  itemsOverride?: ProductsNavItem[]
}>()

const route = useRoute()
const localePath = useLocalePath()

const items = computed<ProductsNavItem[]>(() => {
  if (props.itemsOverride && props.itemsOverride.length) {
    return props.itemsOverride
  }

  const path = route.path || ''

  // 当处于 /company 下的页面时，使用 Company 导航（Our Story / Membership and Points / Picture warehouse）
  if (path.startsWith('/company/')) {
    return companyNavItems
  }

  // 默认使用 Products 导航（Shop / Guides / blog 等）
  return productsNavItems
})

const isActive = (item: ProductsNavItem) => {
  const targetPath = localePath(item.to)
  const currentPath = route.path

  return (
    currentPath === targetPath ||
    (currentPath.startsWith(targetPath) && currentPath[targetPath.length] === '/')
  )
}
</script>

<style scoped>
.products-top-nav {
  width: 100%;
  border-bottom: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.92);
  backdrop-filter: blur(12px);
}

.products-top-nav__scroll {
  max-width: 960px;
  margin: 0 auto;
  padding: 0.5rem 1rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  overflow-x: auto;
  scrollbar-width: thin;
}

.products-top-nav__link {
  flex-shrink: 0;
  font-size: 0.875rem;
  color: #ffffff !important;
  text-decoration: none;
  padding-bottom: 0.25rem;
  border-bottom: 2px solid transparent;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.products-top-nav__link:hover,
.products-top-nav__link:focus-visible {
  color: #e5f2ff;
}

.products-top-nav__link--active {
  color: #ffffff;
  border-color: #38bdf8;
}

@media (max-width: 768px) {
  .products-top-nav__scroll {
    padding-inline: 0.75rem;
  }
}
</style>
