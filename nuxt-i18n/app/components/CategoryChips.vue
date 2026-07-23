<template>
  <nav class="shop-category-chips" :aria-label="t('shopCategoryMenu.ariaLabel')">
    <div class="shop-category-chips__list">
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

const props = defineProps<{
  categories: ShopCategory[]
  selected: ShopCategory | null
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
  border: 1px solid rgba(148, 163, 184, 0.12);
  border-radius: 1rem;
  background: rgba(15, 23, 42, 0.62);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.03);
}

.shop-category-chips__list {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  align-items: center;
  gap: 0.55rem 0.65rem;
  width: 100%;
}

.shop-category-chips__item {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-height: 2.35rem;
  max-width: 100%;
  padding: 0.48rem 0.9rem;
  border: 1px solid rgba(148, 163, 184, 0.32);
  border-radius: 9999px;
  background: rgba(15, 23, 42, 0.5);
  color: rgba(226, 232, 240, 0.9);
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
  border-color: rgba(255, 255, 255, 0.58);
  background: rgba(255, 255, 255, 0.1);
  color: #ffffff;
  transform: translateY(-1px);
}

.shop-category-chips__item:focus-visible {
  outline: 2px solid rgba(64, 255, 170, 0.78);
  outline-offset: 0.2rem;
}

.shop-category-chips__item--active {
  border-color: transparent;
  background: #ffffff;
  color: #020617;
}

.shop-category-chips__item--active:hover,
.shop-category-chips__item--active:focus-visible {
  border-color: transparent;
  background: #ffffff;
  color: #020617;
}
</style>
