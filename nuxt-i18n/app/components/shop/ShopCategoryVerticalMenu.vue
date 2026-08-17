<template>
  <nav class="shop-category-menu" :aria-label="t('shopCategoryMenu.ariaLabel')">
    <div v-if="loading" class="shop-category-menu__state">
      {{ t('shopCategoryMenu.loading') }}
    </div>

    <div v-else-if="error" class="shop-category-menu__state shop-category-menu__state--error">
      {{ error }}
    </div>

    <div v-else-if="categories.length === 0" class="shop-category-menu__state">
      {{ t('shopCategoryMenu.empty') }}
    </div>

    <div v-else class="shop-category-menu__content">
      <p class="shop-category-menu__eyebrow">
        {{ t('filter.categories', 'Categories') }}
      </p>

      <ul class="shop-category-menu__list">
        <li class="shop-category-menu__branch shop-category-menu__branch--all">
          <button
            type="button"
            class="shop-category-menu__item shop-category-menu__item--all"
            :class="{ 'shop-category-menu__item--active': !selected }"
            :title="t('shopCategoryMenu.all')"
            @click="emit('select', null)"
          >
            <span class="shop-category-menu__copy">
              <span class="shop-category-menu__label-track">
                <span class="shop-category-menu__label-line">{{ t('shopCategoryMenu.all') }}</span>
                <span class="shop-category-menu__label-line shop-category-menu__label-line--roll" aria-hidden="true">
                  {{ t('shopCategoryMenu.all') }}
                </span>
              </span>
            </span>
            <span class="shop-category-menu__meta" aria-hidden="true">00 / View</span>
          </button>
        </li>

        <ShopCategoryMenuBranch
          v-for="category in categories"
          :key="category.id"
          :category="category"
          :selected="selected"
          @select="emit('select', $event)"
        />
      </ul>
    </div>
  </nav>
</template>

<script setup lang="ts">
import type { ProductCategory } from '~/composables/useProductCategories'
import ShopCategoryMenuBranch from './ShopCategoryMenuBranch.vue'

const { t } = useI18n()

defineProps<{
  categories: ProductCategory[]
  selected: ProductCategory | null
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  (event: 'select', category: ProductCategory | null): void
}>()
</script>

<style scoped>
.shop-category-menu {
  --shop-category-accent: #b5ff6d;

  width: 100%;
  min-width: 0;
  color: #f8fafc;
}

.shop-category-menu__content {
  display: flex;
  min-width: 0;
  flex-direction: column;
}

.shop-category-menu__eyebrow {
  margin: 0 0 1.4rem;
  color: rgba(248, 250, 252, 0.62);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.1em;
  line-height: 1.2;
  text-align: center;
  text-transform: uppercase;
}

.shop-category-menu__list,
.shop-category-menu :deep(.shop-category-menu__children) {
  margin: 0;
  padding: 0;
  list-style: none;
}

.shop-category-menu :deep(.shop-category-menu__branch) {
  min-width: 0;
  padding-left: var(--category-indent, 0rem);
}

.shop-category-menu :deep(.shop-category-menu__children) {
  margin-left: 0;
}

.shop-category-menu :deep(.shop-category-menu__item) {
  position: relative;
  display: grid;
  width: 100%;
  min-width: 0;
  min-height: clamp(2rem, 2.9vw, 2.6rem);
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  gap: 1rem;
  overflow: hidden;
  border: 0;
  border-bottom: 1px solid rgba(248, 250, 252, 0.15);
  background: transparent;
  color: #f8fafc;
  cursor: pointer;
  padding: 0.2rem 0.15rem;
  text-align: left;
}

.shop-category-menu :deep(.shop-category-menu__item:focus-visible) {
  outline: 2px solid var(--shop-category-accent);
  outline-offset: -2px;
}

.shop-category-menu :deep(.shop-category-menu__copy) {
  position: relative;
  z-index: 2;
  min-width: 0;
  height: clamp(1rem, 1.3vw, 1.4rem);
  overflow: hidden;
}

.shop-category-menu :deep(.shop-category-menu__label-track) {
  display: flex;
  flex-direction: column;
  line-height: 1.06;
  transition: transform 0.38s cubic-bezier(0.22, 1, 0.36, 1);
}

.shop-category-menu :deep(.shop-category-menu__label-line) {
  display: block;
  min-width: 0;
  overflow: hidden;
  font-size: clamp(0.82rem, 1.15vw, 1.25rem);
  font-weight: 850;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shop-category-menu :deep(.shop-category-menu__label-line--roll) {
  color: var(--shop-category-accent);
}

.shop-category-menu :deep(.shop-category-menu__item--child .shop-category-menu__label-line) {
  font-size: clamp(0.75rem, 0.9vw, 1rem);
}

.shop-category-menu :deep(.shop-category-menu__item--child .shop-category-menu__copy) {
  height: clamp(0.9rem, 1.05vw, 1.12rem);
}

.shop-category-menu :deep(.shop-category-menu__item:hover .shop-category-menu__label-track),
.shop-category-menu :deep(.shop-category-menu__item:focus-visible .shop-category-menu__label-track),
.shop-category-menu :deep(.shop-category-menu__item--active .shop-category-menu__label-track) {
  transform: translateY(-50%);
}

.shop-category-menu :deep(.shop-category-menu__meta) {
  position: relative;
  z-index: 2;
  color: rgba(248, 250, 252, 0.64);
  font-size: 0.7rem;
  font-weight: 800;
  letter-spacing: 0.06em;
  line-height: 1.2;
  text-align: right;
  text-transform: uppercase;
  transition: opacity 0.18s ease;
  white-space: nowrap;
}

.shop-category-menu :deep(.shop-category-menu__item:hover .shop-category-menu__meta),
.shop-category-menu :deep(.shop-category-menu__item:focus-visible .shop-category-menu__meta),
.shop-category-menu :deep(.shop-category-menu__item--active .shop-category-menu__meta) {
  opacity: 0;
}

.shop-category-menu__state {
  padding: 1rem;
  color: rgba(248, 250, 252, 0.56);
  font-size: 0.9rem;
  line-height: 1.5;
  text-align: center;
}

.shop-category-menu__state--error {
  color: #fca5a5;
}

@media (max-width: 1023px) {
  .shop-category-menu__eyebrow {
    margin-bottom: 0.9rem;
    text-align: left;
  }

  .shop-category-menu :deep(.shop-category-menu__item) {
    min-height: 3.4rem;
    padding: 0.45rem 0.1rem;
  }

  .shop-category-menu :deep(.shop-category-menu__label-line),
  .shop-category-menu :deep(.shop-category-menu__item--child .shop-category-menu__label-line) {
    font-size: 1.05rem;
    font-weight: 780;
  }

  .shop-category-menu :deep(.shop-category-menu__copy),
  .shop-category-menu :deep(.shop-category-menu__item--child .shop-category-menu__copy) {
    height: 1.2rem;
  }

  .shop-category-menu :deep(.shop-category-menu__meta) {
    display: none;
  }
}
</style>
