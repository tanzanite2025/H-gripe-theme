<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>{{ currentGallery ? galleryTitle(currentGallery) : '图库图片' }}</DialogTitle>
        <DialogDescription>查看和维护图库内图片、缩略图、标签及展示顺序。</DialogDescription>
      </DialogHeader>

      <div class="space-y-4 px-5 py-5">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs text-muted-foreground">共 {{ images.length }} 张图片</span>
          <Button v-if="canCreate" size="sm" @click="emit('create')">
            <Plus class="size-3.5" />
            添加图片
          </Button>
        </div>

        <GalleryImageTablePanel
          :loading="loading"
          :images="images"
          :selected-images="selectedImages"
          :image-selection-state="imageSelectionState"
          :can-edit="canEdit"
          :can-delete="canDelete"
          @batch-delete="emit('batch-delete')"
          @toggle-all="emit('toggle-all', $event)"
          @toggle-image="(...args) => emit('toggle-image', ...args)"
          @preview="(...args) => emit('preview', ...args)"
          @edit="emit('edit', $event)"
          @delete="emit('delete', $event)"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { Plus } from '@lucide/vue'
import GalleryImageTablePanel from '@/components/admin/gallery/GalleryImageTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'

defineProps({
  open: { type: Boolean, default: false },
  currentGallery: { type: Object, default: null },
  images: { type: Array, default: () => [] },
  selectedImages: { type: Array, default: () => [] },
  imageSelectionState: { type: [Boolean, String], default: false },
  loading: { type: Boolean, default: false },
  canCreate: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
  galleryTitle: { type: Function, required: true },
})

const emit = defineEmits([
  'update:open',
  'create',
  'batch-delete',
  'toggle-all',
  'toggle-image',
  'preview',
  'edit',
  'delete',
])
</script>
