<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="overflow-hidden p-0">
      <DialogHeader class="sr-only">
        <DialogTitle>{{ image.title || '图片预览' }}</DialogTitle>
        <DialogDescription>图库图片大图预览</DialogDescription>
      </DialogHeader>
      <div class="flex max-h-[85vh] min-h-64 items-center justify-center bg-black/95 p-3">
        <img :src="image.url" :alt="image.title" class="max-h-[80vh] max-w-full object-contain" />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'

defineProps({
  open: { type: Boolean, default: false },
  image: { type: Object, required: true },
})

const emit = defineEmits(['update:open'])
</script>
