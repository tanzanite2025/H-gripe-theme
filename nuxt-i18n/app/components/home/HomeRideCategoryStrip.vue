<template>
  <section
    id="home-ride-category-strip"
    class="bg-transparent pb-4 pt-0 tz-text-primary sm:pb-5 lg:pb-6"
    aria-label="Product categories"
  >
    <div class="page-content-shell px-0 md:px-6">
      <div
        v-if="visibleCategories.length || loading"
        class="home-ride-category-strip__shell"
      >
        <ul
          v-if="activePageItems.length || loading"
          class="home-ride-category-strip__list"
          :aria-busy="loading ? 'true' : 'false'"
        >
          <li
            v-if="!visibleCategories.length && loading"
            class="home-ride-category-strip__item"
            aria-hidden="true"
          >
            <div class="home-ride-category-strip__link home-ride-category-strip__link--skeleton">
              <span class="home-ride-category-strip__media home-ride-category-strip__media--skeleton" />
              <span class="home-ride-category-strip__name home-ride-category-strip__name--skeleton" />
            </div>
          </li>

          <li
            v-for="category in activePageItems"
            :key="category.id"
            class="home-ride-category-strip__item"
          >
            <NuxtLink
              :to="targetFor(category)"
              class="home-ride-category-strip__link group"
            >
              <span class="home-ride-category-strip__media" aria-hidden="true">
                <StorefrontImage
                  v-if="category.imageUrl"
                  :src="category.imageUrl"
                  :alt="category.name"
                  preset="thumbnail"
                  sizes="(max-width: 767px) 72px, 112px"
                  class="home-ride-category-strip__image"
                />
                <span v-else class="home-ride-category-strip__placeholder">
                  <Icon name="lucide:disc-3" aria-hidden="true" />
                </span>
              </span>

              <span class="home-ride-category-strip__name">{{ category.name }}</span>
            </NuxtLink>
          </li>
        </ul>

        <nav
          v-if="pageCount > 1"
          class="tz-carousel-pagination home-ride-category-strip__pagination"
          role="tablist"
          :aria-label="paginationLabel"
        >
          <button
            v-for="page in pageCount"
            :key="page"
            type="button"
            class="tz-carousel-pagination__dot"
            :class="{ 'is-active': activePage === page - 1 }"
            :aria-label="paginationButtonLabel(page)"
            :aria-selected="activePage === page - 1"
            :aria-current="activePage === page - 1 ? 'true' : undefined"
            role="tab"
            @click="goToPage(page - 1)"
          />
        </nav>
      </div>

      <p v-else class="home-ride-category-strip__state">
        {{ emptyStateText }}
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, onServerPrefetch, ref, watch } from 'vue'
import { useLocalePath } from '#imports'
import StorefrontImage from '~/components/StorefrontImage.vue'
import { useProductCategories, type ProductCategory } from '~/composables/useProductCategories'

type ProductCategoryRoute = {
  path: string
  query: Record<string, string>
}

const localePath = useLocalePath()
const {
  tree,
  loading,
  error,
  loadCategories,
} = useProductCategories()

const isMobileViewport = ref(false)
const activePage = ref(0)

const pageSize = computed(() => (isMobileViewport.value ? 3 : 6))

const visibleCategories = computed(() => {
  const seenIds = new Set<number>()
  return [...tree.value]
    .sort((left, right) => {
      if (left.sortOrder !== right.sortOrder) return left.sortOrder - right.sortOrder
      return left.name.localeCompare(right.name)
    })
    .filter((category) => {
      if (seenIds.has(category.id)) return false
      seenIds.add(category.id)
      return true
    })
})

const chunkCategories = (categories: ProductCategory[], size: number): ProductCategory[][] => {
  if (size <= 0) return [categories]

  const pages: ProductCategory[][] = []
  for (let index = 0; index < categories.length; index += size) {
    pages.push(categories.slice(index, index + size))
  }
  return pages
}

const categoryPages = computed(() => chunkCategories(visibleCategories.value, pageSize.value))
const pageCount = computed(() => Math.max(1, categoryPages.value.length))
const activePageItems = computed(() => categoryPages.value[activePage.value] || [])

const targetFor = (category: ProductCategory): ProductCategoryRoute => {
  return {
    path: localePath('/shop'),
    query: {
      product_category: category.slug,
    },
  }
}

const paginationLabel = 'Product category pages'
const paginationButtonLabel = (page: number) => `Go to category group ${page}`

const emptyStateText = computed(() => {
  if (error.value) {
    return 'Unable to load product categories right now.'
  }
  return 'Product categories are not available yet.'
})

const updateViewportMode = () => {
  if (!import.meta.client) return
  isMobileViewport.value = window.innerWidth < 768
}

const goToPage = (pageIndex: number) => {
  activePage.value = pageIndex
}

watch(pageSize, () => {
  activePage.value = 0
})

watch(
  () => visibleCategories.value.length,
  () => {
    activePage.value = 0
  },
)

watch(pageCount, (count) => {
  if (activePage.value >= count) {
    activePage.value = Math.max(0, count - 1)
  }
})

onServerPrefetch(() => loadCategories())

onMounted(() => {
  updateViewportMode()
  window.addEventListener('resize', updateViewportMode)
  void loadCategories()
})

onBeforeUnmount(() => {
  if (!import.meta.client) return
  window.removeEventListener('resize', updateViewportMode)
})
</script>

<style scoped>
#home-ride-category-strip {
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}

.home-ride-category-strip__shell {
  display: flex;
  min-width: 0;
  flex-direction: column;
  gap: 0.3rem;
}

.home-ride-category-strip__list {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  justify-content: center;
  gap: clamp(0.5rem, 1.2vw, 0.95rem);
  margin: 0;
  padding: 0.25rem 0 0;
  list-style: none;
}

.home-ride-category-strip__item {
  flex: 1 1 0;
  min-width: 0;
  max-width: 7.25rem;
}

.home-ride-category-strip__link {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  color: inherit;
  text-decoration: none;
}

.home-ride-category-strip__media {
  display: grid;
  width: 100%;
  aspect-ratio: 1;
  overflow: hidden;
  place-items: center;
  border: 1px solid rgba(20, 32, 43, 0.1);
  border-radius: 999px;
  background: var(--tz-surface-inset);
  box-shadow: 0 12px 24px rgba(0, 0, 0, 0.16);
  transition:
    border-color 180ms ease,
    box-shadow 180ms ease,
    transform 180ms ease;
}

.home-ride-category-strip__link:hover .home-ride-category-strip__media {
  border-color: rgba(5, 150, 105, 0.4);
  box-shadow: 0 14px 28px rgba(0, 0, 0, 0.22);
  transform: translateY(-1px);
}

.home-ride-category-strip__link:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.92);
  outline-offset: 3px;
}

.home-ride-category-strip__image {
  display: block;
  width: 100%;
  height: 100%;
  padding: 0.45rem;
  object-fit: contain;
}

.home-ride-category-strip__placeholder {
  display: grid;
  width: 100%;
  height: 100%;
  place-items: center;
  background: linear-gradient(180deg, rgba(5, 150, 105, 0.15), rgba(5, 150, 105, 0.08));
  color: var(--tz-site-accent);
}

.home-ride-category-strip__placeholder :deep(svg) {
  width: 1.4rem;
  height: 1.4rem;
}

.home-ride-category-strip__name {
  min-width: 0;
  overflow: hidden;
  color: var(--tz-text-primary);
  font-size: 0.79rem;
  font-weight: 700;
  line-height: 1.15;
  text-align: center;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-ride-category-strip__link--skeleton {
  pointer-events: none;
}

.home-ride-category-strip__media--skeleton,
.home-ride-category-strip__name--skeleton {
  background: linear-gradient(90deg, rgba(20, 32, 43, 0.04), rgba(20, 32, 43, 0.1), rgba(20, 32, 43, 0.04));
  background-size: 220% 100%;
  animation: home-ride-category-strip-shimmer 1.4s ease-in-out infinite;
}

.home-ride-category-strip__media--skeleton {
  background-color: rgba(20, 32, 43, 0.04);
}

.home-ride-category-strip__name--skeleton {
  display: block;
  width: 78%;
  height: 0.7rem;
  margin: 0 auto;
  border-radius: 999px;
}

.home-ride-category-strip__pagination {
  display: flex;
  justify-content: center;
  gap: 0.1rem;
  margin-top: 0.05rem;
}

.home-ride-category-strip__state {
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.8rem;
  line-height: 1.35;
}

@media (max-width: 767px) {
  .home-ride-category-strip__item {
    max-width: 6rem;
  }

  .home-ride-category-strip__list {
    gap: 0.5rem;
  }

  .home-ride-category-strip__name {
    font-size: 0.74rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-ride-category-strip__media,
  .home-ride-category-strip__media--skeleton,
  .home-ride-category-strip__name--skeleton {
    animation: none;
    transition: none;
  }
}

@keyframes home-ride-category-strip-shimmer {
  0% {
    background-position: 0 50%;
  }

  100% {
    background-position: 200% 50%;
  }
}
</style>
