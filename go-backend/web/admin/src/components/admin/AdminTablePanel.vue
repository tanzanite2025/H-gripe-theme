<template>
  <Card class="relative gap-0 overflow-hidden py-0 shadow-none">
    <div v-if="batchVisible && $slots.batch" class="border-b bg-muted/35 px-3 py-2.5 sm:px-4">
      <slot name="batch" />
    </div>

    <div class="relative min-h-40 overflow-x-auto">
      <div
        v-if="loading"
        class="absolute inset-0 z-10 flex items-center justify-center bg-card/80 backdrop-blur-[1px]"
      >
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载" />
      </div>
      <slot />
    </div>

    <div v-if="$slots.footer" class="border-t px-3 py-3 sm:px-4">
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
