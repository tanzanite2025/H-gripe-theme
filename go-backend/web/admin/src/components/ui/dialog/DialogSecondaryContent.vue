<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import DialogContent from './DialogContent.vue'

defineOptions({
  inheritAttrs: false,
})

type DialogSize = 'sm' | 'md' | 'lg' | 'xl' | 'full'

const props = withDefaults(defineProps<{
  class?: HTMLAttributes['class']
  showCloseButton?: boolean
  size?: DialogSize
}>(), {
  showCloseButton: true,
  size: 'xl',
})
</script>

<template>
  <DialogContent
    :size="size"
    :show-close-button="showCloseButton"
    v-bind="$attrs"
    :class="cn('h-[calc(100dvh-1.5rem)] max-h-[calc(100dvh-1.5rem)] grid-rows-[auto_minmax(0,1fr)] gap-0 overflow-hidden rounded-[24px] p-0 sm:h-[min(78dvh,50rem)] sm:max-h-[calc(100dvh-3rem)]', props.class)"
  >
    <slot />
  </DialogContent>
</template>
