<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div
    data-slot="alert-dialog-footer"
    :class="
      cn(
        'bg-muted/20 -mx-5 -mb-5 rounded-b-[32px] border-t border-dashed border-border/70 px-5 py-3 flex flex-col-reverse gap-2 group-data-[size=sm]/alert-dialog-content:grid group-data-[size=sm]/alert-dialog-content:grid-cols-2 sm:flex-row sm:justify-end',
        props.class,
      )
    "
  >
    <slot />
  </div>
</template>
