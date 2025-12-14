<template>
  <nav class="products-top-nav" aria-label="Products navigation">
    <div class="products-top-nav__scroll">
      <NuxtLink
        v-for="item in items"
        :key="item.id"
        class="products-top-nav__link"
        :class="{ 'products-top-nav__link--active': isActive(item) }"
        :to="getTo(item)"
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

const guidesNavItems: ProductsNavItem[] = [
  {
    id: 'about-tools',
    labelKey: 'products.nav.aboutTools',
    to: '/guides/tools',
  },
  {
    id: 'tire-size-charts',
    labelKey: 'products.nav.tireSizeCharts',
    to: '/guides/sizecharts',
  },
  {
    id: 'technical-docs',
    labelKey: 'products.nav.technicalDocs',
    to: '/guides/technical',
  },
  {
    id: 'wheelset-buyers',
    labelKey: 'products.nav.wheelsetBuyersGuide',
    to: '/guides/wheelset-buyers',
  },
]

const blogNavItems: ProductsNavItem[] = [
  {
    id: 'blog',
    labelKey: 'blog.nav.all',
    to: '/blog',
  },
  {
    id: 'blog-news',
    labelKey: 'blog.nav.news',
    to: '/blog/news',
  },
  {
    id: 'blog-wheelsbuild',
    labelKey: 'blog.nav.wheelsbuild',
    to: '/blog/wheelsbuild',
  },
]

type NavContext = 'company' | 'guides' | 'products' | 'blog'

const navContext = computed<NavContext>(() => {
  const path = route.path || ''

  if (path === '/company' || path.startsWith('/company/')) {
    return 'company'
  }

  if (path === '/blog' || path.startsWith('/blog/')) {
    return 'blog'
  }

  const navQuery = route.query.nav
  const forcedProducts =
    typeof navQuery === 'string' && navQuery.toLowerCase() === 'products'

  if (!forcedProducts && (path === '/guides' || path.startsWith('/guides/'))) {
    return 'guides'
  }

  return 'products'
})

const items = computed<ProductsNavItem[]>(() => {
  if (props.itemsOverride && props.itemsOverride.length) {
    return props.itemsOverride
  }

  if (navContext.value === 'company') {
    return companyNavItems
  }

  if (navContext.value === 'guides') {
    return guidesNavItems
  }

  if (navContext.value === 'blog') {
    return blogNavItems
  }

  return productsNavItems
})

const getTo = (item: ProductsNavItem) => {
  const target = localePath(item.to)

  if (
    navContext.value === 'products' &&
    (item.to.startsWith('/guides/') || item.to.startsWith('/support/'))
  ) {
    return `${target}?nav=products`
  }

  return target
}

const isActive = (item: ProductsNavItem) => {
  const targetPath = localePath(item.to)
  const currentPath = route.path

  if (item.to === '/blog') {
    return currentPath === targetPath
  }

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
  -webkit-backdrop-filter: blur(12px);
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

  .products-top-nav__link {
    font-size: 0.875rem;
  }
}
</style>
