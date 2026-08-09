<template>
  <div
    v-if="(thumbnail.kind === 'image' || thumbnail.kind === 'video') && thumbnail.src"
    class="relative flex h-14 w-14 shrink-0 overflow-hidden rounded-xl border border-dashed bg-muted/30 shadow-sm"
    :title="thumbnail.alt"
  >
    <img :src="thumbnail.src" :alt="thumbnail.alt" class="h-full w-full object-cover" />
    <span class="absolute left-1 top-1 rounded-full bg-black/55 px-1.5 py-0.5 text-[10px] font-bold leading-none text-white backdrop-blur">
      {{ thumbnail.label }}
    </span>
  </div>
  <div
    v-else
    class="flex h-14 w-14 items-center justify-center rounded-xl border border-dashed bg-muted/20 text-muted-foreground"
    :title="thumbnail.alt"
  >
    <div class="flex flex-col items-center gap-0.5">
      <Video v-if="thumbnail.kind === 'video'" class="size-5" />
      <ImageIcon v-else class="size-5" />
      <span class="text-[10px] font-medium leading-none">{{ thumbnail.label }}</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ImageIcon, Video } from '@lucide/vue'
import {
  getProductThumbnail,
  type ProductThumbnail,
  type ProductThumbnailProduct,
} from '@/lib/productMedia'

interface ProductThumbnailProps {
  product: ProductThumbnailProduct
}

const props = defineProps<ProductThumbnailProps>()
const thumbnail = computed<ProductThumbnail>(() => getProductThumbnail(props.product))
</script>
