<template>
  <li class="shop-category-menu__branch">
    <div class="shop-category-menu__row">
      <button
        type="button"
        class="shop-category-menu__item"
        :class="{ 'shop-category-menu__item--active': selected?.id === category.id }"
        :style="{ paddingLeft: `${1.75 + Math.max(0, category.depth - 1) * 0.9}rem` }"
        :title="category.name"
        @click="emit('select', category)"
      >
        <span class="shop-category-menu__arrow" aria-hidden="true">→</span>
        <span
          v-if="category.depth === 1"
          class="shop-category-menu__thumbnail"
          :class="{ 'shop-category-menu__thumbnail--placeholder': !category.imageUrl }"
          aria-hidden="true"
        >
          <img v-if="category.imageUrl" :src="category.imageUrl" :alt="category.name" loading="lazy" />
        </span>
        <span class="shop-category-menu__label">{{ category.name }}</span>
      </button>

      <button
        v-if="category.children.length"
        type="button"
        class="shop-category-menu__expand"
        :aria-expanded="expanded"
        :aria-label="expanded ? `收起 ${category.name}` : `展开 ${category.name}`"
        :title="expanded ? `收起 ${category.name}` : `展开 ${category.name}`"
        @click.stop="emit('toggle', category)"
      >
        <Icon :name="expanded ? 'lucide:chevron-up' : 'lucide:chevron-down'" aria-hidden="true" />
      </button>
    </div>

    <ul v-if="expanded && category.children.length" class="shop-category-menu__children">
      <ShopCategoryMenuBranch
        v-for="child in category.children"
        :key="child.id"
        :category="child"
        :selected="selected"
        :expanded-keys="expandedKeys"
        @select="emit('select', $event)"
        @toggle="emit('toggle', $event)"
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
  expandedKeys: Set<number>
}>()

const emit = defineEmits<{
  (event: 'select', category: ProductCategory): void
  (event: 'toggle', category: ProductCategory): void
}>()

const expanded = computed(() => props.expandedKeys.has(props.category.id))
</script>
