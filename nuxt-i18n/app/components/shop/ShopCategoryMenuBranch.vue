<template>
  <li
    class="shop-category-menu__branch"
    :style="{ '--category-indent': `${Math.max(0, category.depth - 1) * 0.75}rem` }"
  >
    <button
      type="button"
      class="shop-category-menu__item"
      :class="{
        'shop-category-menu__item--active': selected?.id === category.id,
        'shop-category-menu__item--child': category.depth > 1,
      }"
      :title="category.name"
      @click="emit('select', category)"
    >
      <span class="shop-category-menu__copy">
        <span class="shop-category-menu__label-track">
          <span class="shop-category-menu__label-line">{{ category.name }}</span>
          <span class="shop-category-menu__label-line shop-category-menu__label-line--roll" aria-hidden="true">
            {{ category.name }}
          </span>
        </span>
      </span>

      <span class="shop-category-menu__meta" aria-hidden="true">
        {{ categoryMetaLabel }}
      </span>

    </button>

    <ul v-if="category.children.length" class="shop-category-menu__children">
      <ShopCategoryMenuBranch
        v-for="child in category.children"
        :key="child.id"
        :category="child"
        :selected="selected"
        @select="emit('select', $event)"
      />
    </ul>
  </li>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { ProductCategory } from '~/composables/useProductCategories'

const props = defineProps<{
  category: ProductCategory
  selected: ProductCategory | null
}>()

const emit = defineEmits<{
  (event: 'select', category: ProductCategory): void
}>()

const categoryMetaLabel = computed(() => {
  const level = String(Math.max(1, props.category.depth)).padStart(2, '0')
  if (props.category.children.length) {
    return `${level} / ${props.category.children.length} ${props.category.children.length === 1 ? 'item' : 'items'}`
  }

  return `${level} / View`
})
</script>
