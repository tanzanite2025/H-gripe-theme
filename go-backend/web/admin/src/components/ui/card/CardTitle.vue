<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div
    data-slot="card-title"
    :class="cn('text-sm font-black tracking-tighter italic uppercase group-data-[size=sm]/card:text-xs', props.class)"
  >
    <slot />
  </div>
</template>
