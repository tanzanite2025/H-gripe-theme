<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from 'reka-ui'

import type { HTMLAttributes } from 'vue'
import { XIcon } from '@lucide/vue'
import { reactiveOmit } from '@vueuse/core'
import {
  DialogClose,
  DialogContent,
  DialogPortal,
  useForwardPropsEmits,
} from 'reka-ui'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import DialogOverlay from './DialogOverlay.vue'

defineOptions({
  inheritAttrs: false,
})

type DialogSize = 'sm' | 'md' | 'lg' | 'xl' | 'full'

const props = withDefaults(defineProps<DialogContentProps & {
  class?: HTMLAttributes['class']
  showCloseButton?: boolean
  size?: DialogSize
}>(), {
  showCloseButton: true,
  size: 'md',
})
const emits = defineEmits<DialogContentEmits>()

const delegatedProps = reactiveOmit(props, 'class', 'size')

const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <DialogPortal>
    <DialogOverlay />
    <DialogContent
      data-slot="dialog-content"
      :data-size="size"
      v-bind="{ ...$attrs, ...forwarded }"
      :class="cn('bg-popover text-popover-foreground data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 shadow-2xl relative overflow-hidden grid max-h-[calc(100dvh-2rem)] min-w-0 gap-4 overflow-x-hidden overflow-y-auto overscroll-contain rounded-[32px] p-5 text-sm duration-100 data-[size=sm]:w-[min(max(20rem,32dvw),40rem,calc(100dvw-2rem))] data-[size=md]:w-[min(max(28rem,44dvw),56rem,calc(100dvw-2rem))] data-[size=lg]:w-[min(max(40rem,64dvw),72rem,calc(100dvw-2rem))] data-[size=xl]:w-[min(max(56rem,78dvw),90rem,calc(100dvw-2rem))] data-[size=full]:w-[min(96dvw,112rem,calc(100dvw-1rem))] fixed top-1/2 left-1/2 z-50 -translate-x-1/2 -translate-y-1/2 outline-none', props.class)"
    >
      <div class="uds-glow-bg pointer-events-none" />
      <slot />

      <DialogClose
        v-if="showCloseButton"
        data-slot="dialog-close"
        as-child
      >
        <Button variant="ghost" class="absolute right-3 top-3 z-30 rounded-full" size="icon-sm">
          <XIcon />
          <span class="sr-only">关闭</span>
        </Button>
      </DialogClose>
    </DialogContent>
  </DialogPortal>
</template>
