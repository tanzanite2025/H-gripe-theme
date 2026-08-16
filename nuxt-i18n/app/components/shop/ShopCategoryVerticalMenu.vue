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

    <ul v-else class="shop-category-menu__list">
      <li>
        <button
          type="button"
          class="shop-category-menu__item shop-category-menu__item--all"
          :class="{ 'shop-category-menu__item--active': !selected }"
          @click="handleSelect(null)"
        >
          <span class="shop-category-menu__arrow" aria-hidden="true">→</span>
          <span class="shop-category-menu__label">{{ t('shopCategoryMenu.all') }}</span>
        </button>
      </li>

      <ShopCategoryMenuBranch
        v-for="category in categories"
        :key="category.id"
        :category="category"
        :selected="selected"
        :expanded-keys="expandedKeys"
        @select="handleSelect"
        @toggle="toggleCategory"
      />
    </ul>
  </nav>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { ProductCategory } from '~/composables/useProductCategories'
import ShopCategoryMenuBranch from './ShopCategoryMenuBranch.vue'

const { t } = useI18n()

const props = defineProps<{
  categories: ProductCategory[]
  selected: ProductCategory | null
  loading?: boolean
  error?: string | null
}>()

const emit = defineEmits<{
  (event: 'select', category: ProductCategory | null): void
}>()

const expandedKeys = ref<Set<number>>(new Set())

const findCategoryPath = (
  items: ProductCategory[],
  targetID: number,
  path: ProductCategory[] = [],
): ProductCategory[] => {
  for (const item of items) {
    const nextPath = [...path, item]
    if (item.id === targetID) return nextPath
    const nestedPath = findCategoryPath(item.children, targetID, nextPath)
    if (nestedPath.length) return nestedPath
  }
  return []
}

const expandPathTo = (category: ProductCategory | null): void => {
  if (!category) {
    expandedKeys.value = new Set()
    return
  }

  const path = findCategoryPath(props.categories, category.id)
  if (!path.length) return

  const next = new Set(expandedKeys.value)
  path.slice(0, -1).forEach((item) => next.add(item.id))
  expandedKeys.value = next
}

const handleSelect = (category: ProductCategory | null): void => {
  expandPathTo(category)
  emit('select', category)
}

const toggleCategory = (category: ProductCategory): void => {
  const next = new Set(expandedKeys.value)
  if (next.has(category.id)) {
    next.delete(category.id)
  } else {
    if (category.depth === 1) {
      props.categories.forEach((root) => {
        if (root.id !== category.id) next.delete(root.id)
      })
    }
    next.add(category.id)
  }
  expandedKeys.value = next
}

watch(
  () => props.selected?.id,
  () => expandPathTo(props.selected),
  { immediate: true },
)
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
  gap: 0.4rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.shop-category-menu :deep(.shop-category-menu__children) {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.shop-category-menu :deep(.shop-category-menu__branch) {
  min-width: 0;
}

.shop-category-menu :deep(.shop-category-menu__row) {
  display: flex;
  min-width: 0;
  align-items: stretch;
  gap: 0.25rem;
}

.shop-category-menu :deep(.shop-category-menu__item) {
  position: relative;
  display: flex;
  min-width: 0;
  min-height: 2.6rem;
  flex: 1 1 auto;
  align-items: center;
  gap: 0.55rem;
  padding-top: 0.35rem;
  padding-right: 0.35rem;
  padding-bottom: 0.35rem;
  border: 0;
  background: transparent;
  color: rgba(248, 250, 252, 0.84);
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1.12;
  text-align: left;
  transition: color 0.18s ease, background-color 0.18s ease;
}

.shop-category-menu :deep(.shop-category-menu__item--all) {
  min-height: 2.25rem;
}

.shop-category-menu :deep(.shop-category-menu__item:hover),
.shop-category-menu :deep(.shop-category-menu__item:focus-visible) {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.05);
}

.shop-category-menu :deep(.shop-category-menu__item:focus-visible),
.shop-category-menu :deep(.shop-category-menu__expand:focus-visible) {
  outline: 2px solid rgba(255, 106, 0, 0.78);
  outline-offset: 0.15rem;
  border-radius: 0.35rem;
}

.shop-category-menu :deep(.shop-category-menu__item--active) {
  color: var(--shop-category-accent);
}

.shop-category-menu :deep(.shop-category-menu__item--active:hover),
.shop-category-menu :deep(.shop-category-menu__item--active:focus-visible) {
  color: var(--shop-category-accent);
}

.shop-category-menu :deep(.shop-category-menu__arrow) {
  position: absolute;
  left: 0;
  width: 1.15rem;
  color: var(--shop-category-accent);
  opacity: 0;
  transform: translateX(-0.35rem);
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.shop-category-menu :deep(.shop-category-menu__item--active .shop-category-menu__arrow) {
  opacity: 1;
  transform: translateX(0);
}

.shop-category-menu :deep(.shop-category-menu__label) {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.shop-category-menu :deep(.shop-category-menu__thumbnail) {
  display: inline-flex;
  width: 4rem;
  height: 2.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  margin-left: 1.55rem;
  border: 1px solid rgba(255, 255, 255, 0.13);
  border-radius: 0.35rem;
  background: rgba(255, 255, 255, 0.06);
}

.shop-category-menu :deep(.shop-category-menu__thumbnail img) {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.shop-category-menu :deep(.shop-category-menu__thumbnail--placeholder) {
  border-style: dashed;
}

.shop-category-menu :deep(.shop-category-menu__expand) {
  display: inline-flex;
  width: 2.25rem;
  min-height: 2.25rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: rgba(248, 250, 252, 0.58);
  cursor: pointer;
  transition: color 0.18s ease, background-color 0.18s ease;
}

.shop-category-menu :deep(.shop-category-menu__expand:hover) {
  color: #ffffff;
  background: rgba(255, 255, 255, 0.06);
}

.shop-category-menu :deep(.shop-category-menu__expand svg) {
  width: 1rem;
  height: 1rem;
}

.shop-category-menu__state {
  padding-left: 1.75rem;
  color: rgba(248, 250, 252, 0.56);
  font-size: 0.9rem;
  line-height: 1.5;
}

.shop-category-menu__state--error {
  color: #fca5a5;
}
</style>
