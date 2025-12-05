<template>
  <aside class="w-full text-xs text-white/80">
    <h3 class="mb-3 text-sm font-semibold text-white/90">Categories</h3>

    <div v-if="loading" class="py-4 text-white/40">
      Loading categories...
    </div>
    <div v-else-if="error" class="py-4 text-red-300">
      {{ error }}
    </div>
    <ul v-else class="space-y-1">
      <li>
        <button
          type="button"
          class="w-full text-left px-2 py-1.5 rounded-md transition-colors"
          :class="!selected ? 'bg-white/15 text-[#40ffaa]' : 'hover:bg-white/10'"
          @click="handleSelect(null)"
        >
          All
        </button>
      </li>
      <li
        v-for="cat in categories"
        :key="cat.id"
      >
        <button
          type="button"
          class="w-full text-left px-2 py-1.5 rounded-md transition-colors flex items-center justify-between gap-2"
          :class="selected && selected.id === cat.id ? 'bg-white/15 text-[#40ffaa]' : 'hover:bg-white/10'"
          @click="handleSelect(cat)"
        >
          <span class="truncate">{{ cat.name }}</span>
          <span v-if="typeof cat.count === 'number'" class="text-[10px] text-white/40">
            {{ cat.count }}
          </span>
        </button>
      </li>
    </ul>
  </aside>
</template>

<script setup lang="ts">
import type { ShopCategory } from '~/composables/useShopCategories'

const props = defineProps<{
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
</style>
