<template>
  <div class="flex items-center gap-2 overflow-x-auto no-scrollbar pb-1">
    <button
      type="button"
      class="px-3 py-1 rounded-full border text-[11px] whitespace-nowrap transition-colors"
      :class="!selected ? 'bg-white text-black border-transparent' : 'bg-transparent text-white/80 border-white/30 hover:bg-white/10'"
      @click="handleSelect(null)"
    >
      All
    </button>
    <button
      v-for="cat in categories"
      :key="cat.id"
      type="button"
      class="px-3 py-1 rounded-full border text-[11px] whitespace-nowrap transition-colors"
      :class="selected && selected.id === cat.id ? 'bg-white text-black border-transparent' : 'bg-transparent text-white/80 border-white/30 hover:bg-white/10'"
      @click="handleSelect(cat)"
    >
      {{ cat.name }}
    </button>
  </div>
</template>

<script setup lang="ts">
import type { ShopCategory } from '~/composables/useShopCategories'

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
.no-scrollbar::-webkit-scrollbar {
  display: none;
}

.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
