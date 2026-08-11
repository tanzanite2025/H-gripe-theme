<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '添加图片' : '编辑图片' }}</DialogTitle>
          <DialogDescription>从媒体仓库选择图片，图库只维护标题、描述和排序。</DialogDescription>
        </DialogHeader>

        <GalleryImageFields
          :form="form"
          :errors="errors"
          @pick-media="emit('pick-media')"
          @clear-error="emit('clear-error', $event)"
        />

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存图片' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import GalleryImageFields from './GalleryImageFields.vue'
import type { GalleryDialogMode, GalleryFormErrors, GalleryImageForm } from './galleryTypes'

withDefaults(defineProps<{
  open?: boolean
  mode?: GalleryDialogMode
  form: GalleryImageForm
  errors?: GalleryFormErrors
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  errors: () => ({}),
  submitting: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: string): void
  (event: 'pick-media'): void
}>()
</script>
