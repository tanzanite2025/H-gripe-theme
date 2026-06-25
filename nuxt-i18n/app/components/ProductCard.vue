<template>
  <div class="flex gap-2.5 p-2 border border-white/[0.18] rounded-[10px] bg-white/[0.06] hover:bg-white/[0.12] transition-colors">
    <a :href="item.url" target="_blank" rel="noopener" class="flex gap-2.5 flex-1 min-w-0">
      <NuxtImg v-if="item.thumbnail" :src="item.thumbnail" :alt="item.title || 'product'" class="w-14 h-14 object-cover rounded-lg flex-shrink-0" loading="lazy" format="webp" />
      <div class="text-sm text-white flex-1 min-w-0 truncate">{{ item.title }}</div>
    </a>
    <div class="flex items-center ml-auto flex-shrink-0">
      <slot name="actions">
        <button class="px-3 py-1.5 bg-[#6b73ff] hover:bg-[#5d65e8] text-white text-sm rounded-md transition-colors" type="button" @click="emit('share', item)">
          share to live chat
        </button>
      </slot>
    </div>
  </div>
</template>

<script setup lang="ts">
interface ProductItem {
  id?: number | string
  title?: string
  url?: string
  thumbnail?: string
  [key: string]: unknown
}

const props = defineProps<{ item: ProductItem }>()
const emit = defineEmits<{ (event: 'share', item: ProductItem): void }>()
</script>
