<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>{{ assetTitle(asset) }}</DialogTitle>
        <DialogDescription>{{ asset?.url }}</DialogDescription>
      </DialogHeader>
      <div class="max-h-[76dvh] overflow-auto bg-black/95 p-4">
        <img
          v-if="asset?.media_type === 'image'"
          :src="assetAccessURL(asset)"
          :alt="asset?.alt || ''"
          class="mx-auto max-h-[70dvh] max-w-full rounded-xl object-contain"
        />
        <video
          v-else-if="asset?.media_type === 'video'"
          :src="assetAccessURL(asset)"
          controls
          class="mx-auto max-h-[70dvh] max-w-full rounded-xl"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import type { MediaAsset } from '@/api/media'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { assetAccessURL, assetTitle } from '@/lib/mediaPresentation'

withDefaults(defineProps<{
  open?: boolean
  asset?: MediaAsset | null
}>(), {
  open: false,
  asset: null
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
}>()
</script>
