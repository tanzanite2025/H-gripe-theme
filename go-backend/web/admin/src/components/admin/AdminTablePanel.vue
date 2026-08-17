<template>
  <Card class="relative gap-0 overflow-hidden py-0 shadow-none rounded-[24px] border-dashed border-border/80 bg-card">
    <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent pointer-events-none" />
    <div v-if="batchVisible && $slots.batch" class="relative z-10 border-b border-dashed border-border/70 bg-muted/20 px-4 py-2.5">
      <slot name="batch" />
    </div>
    <div v-if="$slots.header" class="relative z-10 border-b border-dashed border-border/70 px-4 py-3">
      <slot name="header" />
    </div>

    <div
      class="relative z-10 min-w-0"
 :class="scrollBody ? 'min-h-0 flex-1 overflow-auto': 'min-h-40 overflow-x-auto'"
    >
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex items-center justify-center bg-card/80 backdrop-blur-[1px]"
      >
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载" />
      </div>
      <slot />
    </div>

    <div v-if="$slots.footer" class="relative z-10 shrink-0 border-t border-dashed border-border/70 px-4 py-3">
      <slot name="footer" />
    </div>
  </Card>
</template>

<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'
import { Card } from '@/components/ui/card'

withDefaults(defineProps<{
  loading?: boolean
  batchVisible?: boolean
  scrollBody?: boolean
}>(), {
  loading: false,
  batchVisible: false,
  scrollBody: false
})
</script>
