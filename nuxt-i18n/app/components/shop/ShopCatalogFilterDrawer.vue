<template>
  <teleport to="body">
    <transition name="shop-catalog-filter-drawer">
      <div
        v-if="open"
        class="shop-catalog-filter-drawer"
        role="dialog"
        aria-modal="true"
        :aria-label="t('filter.filters', 'Filters')"
        @click.self="emit('close')"
      >
        <section class="shop-catalog-filter-drawer__panel">
          <header class="shop-catalog-filter-drawer__header">
            <span>{{ t('filter.filters', 'Filters') }}</span>
            <button
              type="button"
              class="shop-catalog-filter-drawer__close"
              :aria-label="t('common.close', 'Close')"
              @click="emit('close')"
            >
              <Icon name="lucide:x" />
            </button>
          </header>

          <div class="shop-catalog-filter-drawer__accordion">
            <section class="shop-catalog-filter-drawer__section">
              <button
                type="button"
                class="shop-catalog-filter-drawer__trigger"
                :class="{ 'shop-catalog-filter-drawer__trigger--active': activePanel === 'search' }"
                :aria-expanded="activePanel === 'search'"
                aria-controls="shop-catalog-filter-search-panel"
                @click="activePanel = 'search'"
              >
                <span class="shop-catalog-filter-drawer__trigger-label">
                  <Icon name="lucide:search" aria-hidden="true" />
                  <span>{{ t('sidebar.search', 'Search') }}</span>
                </span>
                <Icon name="lucide:chevron-down" class="shop-catalog-filter-drawer__chevron" aria-hidden="true" />
              </button>

              <div
                v-show="activePanel === 'search'"
                id="shop-catalog-filter-search-panel"
                class="shop-catalog-filter-drawer__body"
              >
                <ShopProductQuickSearchForm
                  density="drawer"
                  :initial-query="initialQuery"
                  :initial-price-range="initialPriceRange"
                  @submit="emitSearchSubmit"
                />
              </div>
            </section>

            <section class="shop-catalog-filter-drawer__section">
              <button
                type="button"
                class="shop-catalog-filter-drawer__trigger"
                :class="{ 'shop-catalog-filter-drawer__trigger--active': activePanel === 'categories' }"
                :aria-expanded="activePanel === 'categories'"
                aria-controls="shop-catalog-filter-categories-panel"
                @click="activePanel = 'categories'"
              >
                <span class="shop-catalog-filter-drawer__trigger-label">
                  <Icon name="lucide:list-tree" aria-hidden="true" />
                  <span>{{ t('filter.categories', 'Categories') }}</span>
                </span>
                <Icon name="lucide:chevron-down" class="shop-catalog-filter-drawer__chevron" aria-hidden="true" />
              </button>

              <div
                v-show="activePanel === 'categories'"
                id="shop-catalog-filter-categories-panel"
                class="shop-catalog-filter-drawer__body"
              >
                <ShopCategoryVerticalMenu
                  :categories="categories"
                  :selected="selected"
                  :loading="loading"
                  :error="error"
                  @select="emit('category-select', $event)"
                />
              </div>
            </section>
          </div>
        </section>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from '#imports'
import ShopProductQuickSearchForm from '~/components/shop/ShopProductQuickSearchForm.vue'
import ShopCategoryVerticalMenu from '~/components/shop/ShopCategoryVerticalMenu.vue'
import type { ProductCategory } from '~/composables/useProductCategories'
import type { ShopSearchFiltersPayload, ShopSearchPayload } from '~/composables/useShopSearchSheet'

type ShopCatalogFilterPanel = 'search' | 'categories'

const props = withDefaults(defineProps<{
  open: boolean
  categories: ProductCategory[]
  selected: ProductCategory | null
  loading?: boolean
  error?: string | null
  initialQuery?: string
  initialPriceRange?: [number, number]
  initialPanel?: ShopCatalogFilterPanel
}>(), {
  loading: false,
  error: null,
  initialQuery: '',
  initialPriceRange: () => [0, 5000] as [number, number],
  initialPanel: 'search',
})

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'search-submit', payload: ShopSearchPayload): void
  (event: 'category-select', category: ProductCategory | null): void
}>()

const { t } = useI18n()
const activePanel = ref<ShopCatalogFilterPanel>(props.initialPanel)

const cloneFilters = (filters?: ShopSearchFiltersPayload): ShopSearchFiltersPayload => ({
  priceRange: Array.isArray(filters?.priceRange)
    ? [...filters!.priceRange] as [number, number]
    : [...props.initialPriceRange] as [number, number],
  currency: filters?.currency,
  attributes: { ...(filters?.attributes || {}) },
})

const emitSearchSubmit = (payload: { query: string; filters: ShopSearchFiltersPayload }) => {
  emit('search-submit', {
    query: String(payload.query || '').trim(),
    filters: cloneFilters(payload.filters),
  })
}

watch(() => props.open, (isOpen) => {
  if (isOpen) {
    activePanel.value = props.initialPanel
  }
})

watch(() => props.initialPanel, (panel) => {
  if (props.open) {
    activePanel.value = panel
  }
})
</script>

<style scoped>
.shop-catalog-filter-drawer {
  position: fixed;
  inset: 0;
  z-index: 1700;
  display: flex;
  background: rgb(15 23 42 / 0.2);
}

.shop-catalog-filter-drawer__panel {
  display: flex;
  width: min(92vw, 28rem);
  min-width: 17rem;
  height: 100%;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--tz-border-strong);
  background: var(--tz-card-surface);
  box-shadow: 24px 0 60px -28px rgb(15 23 42 / 0.16);
}

.shop-catalog-filter-drawer__header {
  display: flex;
  min-height: 3.5rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  color: var(--tz-text-primary);
  font-size: 14px;
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.shop-catalog-filter-drawer__close {
  display: inline-flex;
  width: 34px;
  height: 34px;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
}

.shop-catalog-filter-drawer__close:hover,
.shop-catalog-filter-drawer__close:focus-visible {
  border-color: rgba(5, 150, 105, 0.64);
  background: var(--tz-site-accent-soft-surface);
}

.shop-catalog-filter-drawer__close:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.72);
  outline-offset: 2px;
}

.shop-catalog-filter-drawer__close :deep(svg) {
  width: 18px;
  height: 18px;
}

.shop-catalog-filter-drawer__accordion {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  overflow-y: auto;
  padding: 0.75rem 0.75rem calc(1rem + env(safe-area-inset-bottom));
  scrollbar-width: thin;
  scrollbar-color: rgba(5, 150, 105, 0.46) transparent;
}

.shop-catalog-filter-drawer__section {
  min-width: 0;
  border-bottom: 1px solid var(--tz-border-subtle);
}

.shop-catalog-filter-drawer__section:first-child {
  border-top: 1px solid var(--tz-border-subtle);
}

.shop-catalog-filter-drawer__trigger {
  display: flex;
  width: 100%;
  min-height: 3.1rem;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 0;
  background: transparent;
  color: var(--tz-text-primary);
  padding: 0.75rem 0.25rem;
  text-align: left;
}

.shop-catalog-filter-drawer__trigger-label {
  display: inline-flex;
  min-width: 0;
  align-items: center;
  gap: 9px;
  font-size: 13px;
  font-weight: 850;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.shop-catalog-filter-drawer__trigger-label :deep(svg) {
  width: 16px;
  height: 16px;
  color: var(--tz-site-accent);
}

.shop-catalog-filter-drawer__chevron {
  width: 16px;
  height: 16px;
  flex: 0 0 auto;
  color: var(--tz-text-muted);
  transition: transform 0.18s ease;
}

.shop-catalog-filter-drawer__trigger--active .shop-catalog-filter-drawer__chevron {
  transform: rotate(180deg);
}

.shop-catalog-filter-drawer__body {
  min-width: 0;
  padding: 0.25rem 0 1rem;
}

.shop-catalog-filter-drawer__body :deep(.shop-category-menu) {
  width: 100%;
  min-height: 0;
}

.shop-catalog-filter-drawer__body :deep(.shop-category-menu__eyebrow) {
  display: none;
}

.shop-catalog-filter-drawer__body :deep(.shop-category-menu__list) {
  gap: 0.45rem;
}

.shop-catalog-filter-drawer-enter-active,
.shop-catalog-filter-drawer-leave-active {
  transition: opacity 0.2s ease;
}

.shop-catalog-filter-drawer-enter-active .shop-catalog-filter-drawer__panel,
.shop-catalog-filter-drawer-leave-active .shop-catalog-filter-drawer__panel {
  transition: transform 0.22s ease;
}

.shop-catalog-filter-drawer-enter-from,
.shop-catalog-filter-drawer-leave-to {
  opacity: 0;
}

.shop-catalog-filter-drawer-enter-from .shop-catalog-filter-drawer__panel,
.shop-catalog-filter-drawer-leave-to .shop-catalog-filter-drawer__panel {
  transform: translateX(-100%);
}
</style>
