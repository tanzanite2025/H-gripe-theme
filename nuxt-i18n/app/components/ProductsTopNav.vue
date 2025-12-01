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
  if (path === '/company' || path.startsWith('/company/')) {
    return companyNavItems
  }

  let base: ProductsNavItem[]

  // 仅在 /guides 总览页精简为 Guides 子分类入口；/guides/* 仍使用完整 Products 导航
  if (path === '/guides') {
    base = productsNavItems.filter((item) => item.to.startsWith('/guides/'))
  } else {
    base = productsNavItems
  }

  // 只要当前导航中包含 Shop，就隐藏 About Tools 与 Wheelsbuild blog 两个入口，
  // 保持其它 Products 布局页面导航一致。
  if (base.some((item) => item.id === 'shop')) {
    return base.filter(
      (item) => item.id !== 'about-tools' && item.id !== 'wheelsbuild-blog'
    )
  }

  return base
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
  padding: 0.75rem 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  overflow-x: auto;
  scrollbar-width: thin;
}

.products-top-nav__link {
  flex-shrink: 0;
  font-size: 1rem;
  font-weight: 500;
  color: #ffffff !important;
  text-decoration: none;
  padding-bottom: 0.3rem;
  border-bottom: 3px solid transparent;
  transition: color 0.15s ease, border-color 0.15s ease;
}

.products-top-nav__link:hover,
.products-top-nav__link:focus-visible {
  color: #e5f2ff;
}

.products-top-nav__link--active {
  color: #ffffff;
  font-weight: 600;
  border-color: #38bdf8;
}

@media (max-width: 768px) {
  .products-top-nav__scroll {
    padding-inline: 0.75rem;
    justify-content: flex-start;
  }
}
</style>
