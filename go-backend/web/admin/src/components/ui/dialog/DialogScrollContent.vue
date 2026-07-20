<script setup lang="ts">
import type { DialogContentEmits, DialogContentProps } from 'reka-ui'

import type { HTMLAttributes } from 'vue'
import { XIcon } from '@lucide/vue'
import { reactiveOmit } from '@vueuse/core'
import {
  DialogClose,
  DialogContent,
  DialogOverlay,
  DialogPortal,
  useForwardPropsEmits,
} from 'reka-ui'
import { cn } from '@/lib/utils'

defineOptions({
  inheritAttrs: false,
})

type DialogSize = 'sm' | 'md' | 'lg' | 'xl' | 'full'

const props = withDefaults(defineProps<DialogContentProps & {
  class?: HTMLAttributes['class']
  size?: DialogSize
}>(), {
  size: 'md',
})
const emits = defineEmits<DialogContentEmits>()

const delegatedProps = reactiveOmit(props, 'class', 'size')

const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <DialogPortal>
    <DialogOverlay
      class="fixed inset-0 z-50 grid place-items-center overflow-x-hidden overflow-y-auto overscroll-contain bg-black/80 px-4 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0"
    >
      <DialogContent
        :data-size="size"
        :class="
          cn(
            'relative z-50 my-8 grid gap-4 border border-border bg-background p-6 shadow-lg duration-200 data-[size=sm]:w-[min(max(20rem,32dvw),40rem,calc(100dvw-2rem))] data-[size=md]:w-[min(max(28rem,44dvw),56rem,calc(100dvw-2rem))] data-[size=lg]:w-[min(max(40rem,64dvw),72rem,calc(100dvw-2rem))] data-[size=xl]:w-[min(max(56rem,78dvw),90rem,calc(100dvw-2rem))] data-[size=full]:w-[min(96dvw,112rem,calc(100dvw-1rem))] sm:rounded-lg',
            props.class,
          )
        "
        v-bind="{ ...$attrs, ...forwarded }"
        @pointer-down-outside="(event) => {
          const originalEvent = event.detail.originalEvent;
          const target = originalEvent.target as HTMLElement;
          if (originalEvent.offsetX > target.clientWidth || originalEvent.offsetY > target.clientHeight) {
            event.preventDefault();
          }
        }"
      >
        <slot />

        <DialogClose
          class="absolute top-4 right-4 p-0.5 transition-colors rounded-md hover:bg-secondary"
        >
          <XIcon class="w-4 h-4" />
          <span class="sr-only">Close</span>
        </DialogClose>
      </DialogContent>
    </DialogOverlay>
  </DialogPortal>
</template>
