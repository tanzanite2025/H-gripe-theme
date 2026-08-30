<template>
  <div class="grid gap-3 rounded-xl border border-border/80 bg-background p-3 lg:grid-cols-[9rem_minmax(0,1fr)] lg:items-start">
    <div class="min-w-0 space-y-2">
      <div class="aspect-[3/4] overflow-hidden rounded-lg bg-muted">
        <img
          v-if="item.image_url"
          :src="item.image_url"
          :alt="item.alt_text"
          class="size-full object-cover"
        />
        <div v-else class="flex size-full items-center justify-center text-muted-foreground">
          <ImagePlus class="size-6 opacity-50" />
        </div>
      </div>
      <div class="flex items-center justify-between gap-2">
        <Badge variant="outline">#{{ index + 1 }}</Badge>
        <Badge :variant="item.is_published ? 'default' : 'outline'">
          {{ item.is_published ? 'Published' : 'Draft' }}
        </Badge>
      </div>
      <div class="space-y-1.5">
        <Button
          type="button"
          variant="outline"
          size="sm"
          class="w-full justify-center"
          :disabled="!canEdit || uploading"
          title="上传或替换图片"
          @click="fileInput?.click()"
        >
          <LoaderCircle v-if="uploading" class="size-4 animate-spin" />
          <ImagePlus v-else class="size-3.5" />
          上传图片
        </Button>
        <input
          ref="fileInput"
          type="file"
          class="hidden"
          :accept="uploadSpecAccept('visual_showcase_editorial')"
          :disabled="!canEdit || uploading"
          @change="handleUploadFile"
        />
      </div>
      <p class="truncate text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
        {{ item.storage_key || '未上传到专用目录' }}
      </p>
    </div>

    <div class="grid min-w-0 gap-3 md:grid-cols-2">
      <AdminFormField label="标题" required class="md:col-span-2">
        <Input
          :model-value="item.title"
          :disabled="!canEdit"
          @update:model-value="updateText('title', $event)"
        />
      </AdminFormField>
      <AdminFormField label="ALT 文本" required class="md:col-span-2">
        <Input
          :model-value="item.alt_text"
          :disabled="!canEdit"
          @update:model-value="updateText('alt_text', $event)"
        />
      </AdminFormField>
      <AdminFormField label="桌面顺序">
        <Input
          type="number"
          min="1"
          :model-value="item.desktop_order"
          :disabled="!canEdit"
          @update:model-value="updateNumber('desktop_order', $event)"
        />
      </AdminFormField>
      <AdminFormField label="发布">
        <div class="flex h-9 items-center">
          <Switch
            :checked="item.is_published"
            :disabled="!canEdit"
            @update:checked="updatePublished"
          />
        </div>
      </AdminFormField>
      <AdminFormField label="跳转链接" class="md:col-span-2">
        <Input
          :model-value="item.target_url"
          :disabled="!canEdit"
          @update:model-value="updateText('target_url', $event)"
        />
      </AdminFormField>
      <AdminFormField label="跳转文字" class="md:col-span-2">
        <Input
          :model-value="item.target_label"
          :disabled="!canEdit"
          @update:model-value="updateText('target_label', $event)"
        />
      </AdminFormField>
      <AdminFormField label="备注" class="md:col-span-2">
        <Textarea
          :model-value="item.caption"
          class="min-h-16"
          :disabled="!canEdit"
          @update:model-value="updateText('caption', $event)"
        />
      </AdminFormField>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { toast } from 'vue-sonner'
import { ImagePlus, LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { uploadSpecAccept, validateUploadFile } from '@/lib/uploadSpecs'
import type {
  VisualShowcaseAdministrationItemFormState,
  VisualShowcaseAdministrationUploadRequest,
} from './visualShowcaseTypes'

const props = withDefaults(defineProps<{
  item: VisualShowcaseAdministrationItemFormState
  index: number
  canEdit?: boolean
  uploading?: boolean
}>(), {
  canEdit: false,
  uploading: false,
})

const emit = defineEmits<{
  (event: 'update:item', value: VisualShowcaseAdministrationItemFormState): void
  (event: 'upload-image', value: VisualShowcaseAdministrationUploadRequest): void
}>()

const fileInput = ref<HTMLInputElement | null>(null)

const patchItem = (patch: Partial<VisualShowcaseAdministrationItemFormState>): void => {
  emit('update:item', { ...props.item, ...patch })
}

const updateText = (
  key: 'title' | 'caption' | 'alt_text' | 'target_url' | 'target_label',
  value: string | number,
): void => {
  patchItem({ [key]: String(value ?? '') })
}

const updateNumber = (
  key: 'desktop_order',
  value: string | number,
): void => {
  const parsed = Number(value)
  patchItem({ [key]: Number.isFinite(parsed) ? Math.max(1, Math.trunc(parsed)) : 1 })
}

const updatePublished = (value: boolean): void => {
  patchItem({ is_published: Boolean(value) })
}

const handleUploadFile = async (event: Event): Promise<void> => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  if (!file) return
  const validation = await validateUploadFile(file, 'visual_showcase_editorial')
  if (!validation.ok) {
    toast.error(validation.error || '竖版视觉图片不符合上传规范')
    return
  }
  if (validation.warning) toast.warning(validation.warning)
  emit('upload-image', { index: props.index, file })
}
</script>
