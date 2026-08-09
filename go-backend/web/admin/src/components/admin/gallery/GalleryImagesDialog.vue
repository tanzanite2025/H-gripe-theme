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
          @toggle-image="forwardToggleImage"
          @preview="forwardPreview"
          @edit="emit('edit', $event)"
          @delete="emit('delete', $event)"
        />
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { Plus } from '@lucide/vue'
import GalleryImageTablePanel from '@/components/admin/gallery/GalleryImageTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import type {
  GalleryImage,
  GalleryRecord,
  GallerySelectionState,
  GalleryTitleResolver
} from './galleryTypes'

withDefaults(defineProps<{
  open?: boolean
  currentGallery?: GalleryRecord | null
  images?: GalleryImage[]
  selectedImages?: GalleryImage[]
  imageSelectionState?: GallerySelectionState
  loading?: boolean
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
  galleryTitle: GalleryTitleResolver
}>(), {
  open: false,
  currentGallery: null,
  images: () => [],
  selectedImages: () => [],
  imageSelectionState: false,
  loading: false,
  canCreate: false,
  canEdit: false,
  canDelete: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'create'): void
  (event: 'batch-delete'): void
  (event: 'toggle-all', checked: GallerySelectionState): void
  (event: 'toggle-image', image: GalleryImage, checked: GallerySelectionState): void
  (event: 'preview', url?: string | null, title?: string | null): void
  (event: 'edit', image: GalleryImage): void
  (event: 'delete', image: GalleryImage): void
}>()

const forwardToggleImage = (image: GalleryImage, checked: GallerySelectionState): void => {
  emit('toggle-image', image, checked)
}

const forwardPreview = (url?: string | null, title?: string | null): void => {
  emit('preview', url, title)
}
</script>
