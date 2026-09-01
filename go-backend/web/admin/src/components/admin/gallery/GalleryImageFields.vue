<template>
  <div class="space-y-3 rounded-xl bg-muted/25 p-3">
    <div v-if="form.url" class="flex h-28 items-center justify-center overflow-hidden rounded-lg bg-muted">
      <img :src="form.thumbnail || form.url" :alt="form.title || '图片预览'" class="size-full object-contain" />
    </div>

    <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      <AdminFormField label="媒体库图片" required :error="errors.media_asset_id" class="sm:col-span-2 lg:col-span-3">
        <div class="flex flex-wrap items-center gap-2 rounded-xl bg-background/70 p-2">
          <span class="min-w-0 flex-1 truncate text-xs font-bold text-muted-foreground">
            {{ form.media_asset_id ? `媒体资源 #${form.media_asset_id}` : '尚未选择图片' }}
          </span>
          <Button type="button" variant="outline" size="sm" @click="emit('pick-media')">
            <ImagePlus class="size-3.5" />
            选择图片
          </Button>
        </div>
      </AdminFormField>

      <AdminFormField label="标题" required :error="errors.title" class="lg:col-span-2">
        <Input v-model="form.title" placeholder="请输入图片标题" @input="emit('clear-error', 'title')" />
      </AdminFormField>
      <AdminFormField label="排序">
        <Input v-model.number="form.order" type="number" min="0" max="9999" step="1" />
      </AdminFormField>
      <AdminFormField label="标签">
        <Input v-model="form.tags" placeholder="多个标签用逗号分隔" />
      </AdminFormField>
      <AdminFormField label="描述" class="sm:col-span-2 lg:col-span-2">
        <Textarea v-model="form.description" class="min-h-16" placeholder="请输入图片描述" />
      </AdminFormField>
    </div>

    <div v-if="showRemove" class="flex justify-end border-t border-border/60 pt-2">
      <Button type="button" variant="ghost" size="sm" class="text-destructive hover:text-destructive" @click="emit('remove')">
        <Trash2 class="size-3.5" />
        移除图片
      </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ImagePlus, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { GalleryFormErrors, GalleryImageForm } from '@/modules/gallery/galleryTypes'

withDefaults(defineProps<{
  form: GalleryImageForm
  errors?: GalleryFormErrors
  showRemove?: boolean
}>(), {
  errors: () => ({}),
  showRemove: false
})

const emit = defineEmits<{
  (event: 'pick-media'): void
  (event: 'clear-error', key: string): void
  (event: 'remove'): void
}>()
</script>

