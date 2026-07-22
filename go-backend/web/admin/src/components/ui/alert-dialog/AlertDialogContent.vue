<script setup lang="ts">
import type { AlertDialogContentEmits, AlertDialogContentProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import {
  AlertDialogContent,
  AlertDialogOverlay,
  AlertDialogPortal,
  useForwardPropsEmits,
} from 'reka-ui'
import { cn } from '@/lib/utils'

defineOptions({
  inheritAttrs: false,
})

const props = withDefaults(
  defineProps<AlertDialogContentProps & {
    class?: HTMLAttributes['class']
    size?: 'default' | 'sm'
  }>(),
  {
    size: 'default',
  },
)
const emits = defineEmits<AlertDialogContentEmits>()

const delegatedProps = reactiveOmit(props, 'class', 'size')

const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <AlertDialogPortal>
    <AlertDialogOverlay
      data-slot="alert-dialog-overlay"
      class="data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 bg-black/10 duration-100 supports-backdrop-filter:backdrop-blur-xs fixed inset-0 z-50"
    />
    <AlertDialogContent
      data-slot="alert-dialog-content"
      :data-size="size"
      v-bind="{ ...$attrs, ...forwarded }"
      :class="
        cn(
          'data-open:animate-in data-closed:animate-out data-closed:fade-out-0 data-open:fade-in-0 data-closed:zoom-out-95 data-open:zoom-in-95 bg-popover text-popover-foreground shadow-2xl gap-4 rounded-[32px] p-5 duration-100 data-[size=default]:w-[min(max(18rem,26dvw),24rem,calc(100dvw-2rem))] data-[size=sm]:w-[min(max(16rem,22dvw),20rem,calc(100dvw-2rem))] group/alert-dialog-content fixed top-1/2 left-1/2 z-50 grid max-h-[70dvh] min-w-0 -translate-x-1/2 -translate-y-1/2 overflow-x-hidden overflow-y-auto overscroll-contain outline-none relative',
          props.class,
        )
      "
    >
      <div class="uds-glow-bg pointer-events-none" />
      <slot />
    </AlertDialogContent>
  </AlertDialogPortal>
</template>
