<template>
  <nav class="shop-category-chips" :aria-label="t('shopCategoryMenu.ariaLabel')">
    <div v-if="loading" class="shop-category-chips__state">
      {{ t('shopCategoryMenu.loading') }}
    </div>

    <div v-else-if="error" class="shop-category-chips__state shop-category-chips__state--error">
      {{ error }}
    </div>

    <div v-else-if="categories.length === 0" class="shop-category-chips__state">
      {{ t('shopCategoryMenu.empty') }}
    </div>

    <div v-else class="shop-category-chips__list">
      <button
        type="button"
        class="shop-category-chips__item"
        :class="{ 'shop-category-chips__item--active': !selected }"
        :aria-pressed="!selected"
        @click="handleSelect(null)"
      >
        {{ t('shopCategoryMenu.all') }}
      </button>
      <button
        v-for="cat in categories"
        :key="cat.id"
        type="button"
        class="shop-category-chips__item"
        :class="{ 'shop-category-chips__item--active': selected?.id === cat.id }"
        :aria-pressed="selected?.id === cat.id"
        @click="handleSelect(cat)"
      >
        {{ cat.name }}
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import type { ShopCategory } from '~/composables/useShopCategories'

const { t } = useI18n()

defineProps<{
  categories: ShopCategory[]
  selected: ShopCategory | null
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  (e: 'select', category: ShopCategory | null): void
}>()

const handleSelect = (category: ShopCategory | null) => {
  emit('select', category)
}
</script>

<style scoped>
.shop-category-chips {
  width: 100%;
  padding: 0.65rem 0.55rem 0.75rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: var(--tz-card-surface);
  box-shadow: 0 8px 20px rgba(15, 23, 42, 0.06);
}

.shop-category-chips__list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 0.55rem 0.65rem;
  width: 100%;
}

.shop-category-chips__state {
  width: 100%;
  color: var(--tz-text-secondary);
  font-size: 0.86rem;
  line-height: 1.45;
  text-align: center;
}

.shop-category-chips__state--error {
  color: #fca5a5;
}

.shop-category-chips__item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  max-width: 100%;
  padding: 0.48rem 0.9rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 9999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  font-size: clamp(0.85rem, 2.8vw, 1rem);
  font-weight: 600;
  line-height: 1.1;
  white-space: nowrap;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.shop-category-chips__item:hover,
.shop-category-chips__item:focus-visible {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
  color: var(--tz-site-accent-hover);
  transform: translateY(-1px);
}

.shop-category-chips__item:focus-visible {
  outline: 2px solid rgba(5, 150, 105, 0.78);
  outline-offset: 0.2rem;
}

.shop-category-chips__item--active {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent);
  color: #ffffff;
}

.shop-category-chips__item--active:hover,
.shop-category-chips__item--active:focus-visible {
  border-color: var(--tz-site-accent-hover);
  background: var(--tz-site-accent-hover);
  color: #ffffff;
}
</style>
