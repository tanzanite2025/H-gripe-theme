<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { DialogClose } from 'reka-ui'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

const props = withDefaults(defineProps<{
  class?: HTMLAttributes['class']
  showCloseButton?: boolean
}>(), {
  showCloseButton: false,
})
</script>

<template>
  <div
    data-slot="dialog-footer"
    :class="cn('bg-muted/20 -mx-5 -mb-5 rounded-b-[32px] border-t border-dashed border-border/70 px-5 py-3 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end', props.class)"
  >
    <slot />
    <DialogClose v-if="showCloseButton" as-child>
      <Button variant="outline" class="rounded-full">
        关闭
      </Button>
    </DialogClose>
  </div>
</template>
