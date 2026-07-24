<template>
  <nav class="shop-category-menu" :aria-label="t('shopCategoryMenu.ariaLabel')">
    <div v-if="loading" class="shop-category-menu__state">
      {{ t('shopCategoryMenu.loading') }}
    </div>

    <div v-else-if="error" class="shop-category-menu__state shop-category-menu__state--error">
      {{ error }}
    </div>

    <ul v-else class="shop-category-menu__list">
      <li>
        <button
          type="button"
          class="shop-category-menu__item"
          :class="{ 'shop-category-menu__item--active': !selected }"
          @click="handleSelect(null)"
        >
          <span class="shop-category-menu__arrow" aria-hidden="true">→</span>
          <span class="shop-category-menu__label">{{ t('shopCategoryMenu.all') }}</span>
        </button>
      </li>

      <li v-for="category in categories" :key="category.id">
        <button
          type="button"
          class="shop-category-menu__item"
          :class="{ 'shop-category-menu__item--active': selected?.id === category.id }"
          @click="handleSelect(category)"
        >
          <span class="shop-category-menu__arrow" aria-hidden="true">→</span>
          <span class="shop-category-menu__label">{{ category.name }}</span>
        </button>
      </li>
    </ul>
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
  (event: 'select', category: ShopCategory | null): void
}>()

const handleSelect = (category: ShopCategory | null) => {
  emit('select', category)
}
</script>

<style scoped>
.shop-category-menu {
  --shop-category-accent: #ff6a00;
  width: min(100%, 18rem);
  color: #f8fafc;
}

.shop-category-menu__list {
  display: flex;
  flex-direction: column;
  gap: clamp(0.55rem, 0.85vw, 0.95rem);
  margin: 0;
  padding: 0;
  list-style: none;
}

.shop-category-menu__item {
  position: relative;
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 0 0 0 clamp(1.75rem, 2.2vw, 2.6rem);
  border: 0;
  background: transparent;
  color: rgba(248, 250, 252, 0.94);
  cursor: pointer;
  font-size: clamp(0.95rem, 1.05vw, 1.5rem);
  font-weight: 800;
  line-height: 1.12;
  letter-spacing: -0.035em;
  text-align: left;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.shop-category-menu__item:hover,
.shop-category-menu__item:focus-visible {
  color: #ffffff;
  transform: translateX(0.2rem);
}

.shop-category-menu__item:focus-visible {
  outline: 2px solid rgba(255, 106, 0, 0.78);
  outline-offset: 0.35rem;
  border-radius: 0.35rem;
}

.shop-category-menu__item--active {
  color: var(--shop-category-accent);
}

.shop-category-menu__item--active:hover,
.shop-category-menu__item--active:focus-visible {
  color: var(--shop-category-accent);
}

.shop-category-menu__arrow {
  position: absolute;
  left: 0;
  color: var(--shop-category-accent);
  opacity: 0;
  transform: translateX(-0.35rem);
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.shop-category-menu__item--active .shop-category-menu__arrow {
  opacity: 1;
  transform: translateX(0);
}

.shop-category-menu__label {
  min-width: 0;
  overflow-wrap: anywhere;
}

.shop-category-menu__state {
  padding-left: clamp(1.75rem, 2.2vw, 2.6rem);
  color: rgba(248, 250, 252, 0.56);
  font-size: 0.9rem;
  line-height: 1.5;
}

.shop-category-menu__state--error {
  color: #fca5a5;
}
</style>
