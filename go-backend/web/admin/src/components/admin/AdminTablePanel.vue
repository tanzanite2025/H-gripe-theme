<template>
  <Card class="relative gap-0 overflow-hidden py-0 shadow-none rounded-[24px] border-dashed border-border/80 bg-card">
    <div class="absolute inset-0 bg-gradient-to-br from-primary/5 via-transparent to-transparent pointer-events-none" />
    <div v-if="batchVisible && $slots.batch" class="relative z-10 border-b border-dashed border-border/70 bg-muted/20 px-4 py-2.5">
      <slot name="batch" />
    </div>

    <div class="relative z-10 min-h-40 overflow-x-auto">
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex items-center justify-center bg-card/80 backdrop-blur-[1px]"
      >
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载" />
      </div>
      <slot />
    </div>

    <div v-if="$slots.footer" class="relative z-10 border-t border-dashed border-border/70 px-4 py-3">
      <slot name="footer" />
    </div>
  </Card>
</template>

<script setup>
import { LoaderCircle } from '@lucide/vue'
import { Card } from '@/components/ui/card'

defineProps({
  loading: {
    type: Boolean,
    default: false
  },
  batchVisible: {
    type: Boolean,
    default: false
  }
})
</script>
