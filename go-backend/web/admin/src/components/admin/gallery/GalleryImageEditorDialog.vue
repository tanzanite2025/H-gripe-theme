<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '添加图片' : '编辑图片' }}</DialogTitle>
          <DialogDescription>维护原图、缩略图和用于检索的图片信息。</DialogDescription>
        </DialogHeader>

        <div v-if="form.url" class="flex h-40 items-center justify-center overflow-hidden rounded-lg border bg-muted">
          <img :src="form.thumbnail || form.url" :alt="form.title || '图片预览'" class="size-full object-contain" />
        </div>

        <div class="grid gap-4 sm:grid-cols-2">
          <AdminFormField label="图片 URL" required :error="errors.url" class="sm:col-span-2">
            <Input v-model="form.url" type="url" placeholder="https://example.com/image.jpg" @input="emit('clear-error', 'url')" />
          </AdminFormField>
          <AdminFormField label="缩略图 URL" class="sm:col-span-2">
            <Input v-model="form.thumbnail" type="url" placeholder="可选" />
          </AdminFormField>
          <AdminFormField label="标题" required :error="errors.title">
            <Input v-model="form.title" placeholder="请输入图片标题" @input="emit('clear-error', 'title')" />
          </AdminFormField>
          <AdminFormField label="排序">
            <Input v-model.number="form.order" type="number" min="0" max="9999" step="1" />
          </AdminFormField>
          <AdminFormField label="标签" class="sm:col-span-2">
            <Input v-model="form.tags" placeholder="多个标签用逗号分隔" />
          </AdminFormField>
          <AdminFormField label="描述" class="sm:col-span-2">
            <Textarea v-model="form.description" class="min-h-24" placeholder="请输入图片描述" />
          </AdminFormField>
        </div>

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
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
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
}>()
</script>
