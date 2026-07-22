<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div
    data-slot="card-description"
    :class="cn('text-[9px] font-black uppercase tracking-widest text-muted-foreground/70 mt-0.5', props.class)"
  >
    <slot />
  </div>
</template>
