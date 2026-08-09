<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>删除媒体资源</DialogTitle>
        <DialogDescription>{{ assetTitle }}</DialogDescription>
      </DialogHeader>

      <div class="space-y-4">
        <div v-if="loading" class="flex min-h-24 items-center justify-center gap-2 text-xs font-bold text-muted-foreground">
          <LoaderCircle class="size-3.5 animate-spin" />
          正在检查引用关系
        </div>

        <div v-else-if="hasReferences" class="space-y-3">
          <div class="rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-bold text-amber-800">
            当前资源仍被 {{ total }} 处内容引用。请先到对应内容里替换或移除图片，再回到媒体库删除。
          </div>
          <div class="max-h-72 space-y-2 overflow-y-auto rounded-lg border bg-muted/20 p-2">
            <div
              v-for="(item, index) in references"
              :key="`${item.resource_type}-${item.resource_id || index}-${item.field}`"
              class="rounded-md border bg-background px-3 py-2"
            >
              <div class="truncate text-xs font-black">{{ item.label || referenceTypeLabel(item.resource_type) }}</div>
              <div class="mt-1 flex flex-wrap gap-2 text-[10px] font-bold text-muted-foreground">
                <span>{{ referenceCategoryLabel(item.category) }}</span>
                <span>{{ referenceTypeLabel(item.resource_type) }}</span>
                <span v-if="item.field">字段：{{ item.field }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="space-y-3">
          <div class="rounded-lg border border-destructive/25 bg-destructive/5 px-3 py-2 text-xs font-bold text-destructive">
            未发现引用。确认后会永久删除数据库记录和服务器/对象存储里的物理文件，版权证据包也将不可再导出。
          </div>
          <label class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              输入 {{ expectedConfirmation }} 确认删除
            </span>
            <Input
              :model-value="confirmation"
              autocomplete="off"
              :placeholder="expectedConfirmation"
              @update:model-value="(value) => emit('update:confirmation', String(value))"
            />
          </label>
        </div>
      </div>

      <DialogFooter>
        <Button variant="outline" :disabled="deleting" @click="emit('update:open', false)">取消</Button>
        <Button
          variant="destructive"
          :disabled="loading || deleting || hasReferences || confirmation !== expectedConfirmation"
          @click="emit('confirm')"
        >
          <LoaderCircle v-if="deleting" class="size-3.5 animate-spin" />
          <Trash2 v-else class="size-3.5" />
          永久删除
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle, Trash2 } from '@lucide/vue'
import type { MediaAsset, MediaReference } from '@/api/media'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { mediaReferenceCategoryLabel as referenceCategoryLabel, mediaReferenceTypeLabel as referenceTypeLabel } from '@/lib/mediaReferencePresentation'

const props = withDefaults(defineProps<{
  open?: boolean
  asset?: MediaAsset | null
  references?: MediaReference[]
  total?: number
  loading?: boolean
  deleting?: boolean
  confirmation?: string
}>(), {
  open: false,
  asset: null,
  references: () => [],
  total: 0,
  loading: false,
  deleting: false,
  confirmation: ''
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:confirmation', value: string): void
  (event: 'confirm'): void
}>()

const assetTitle = computed(() => {
  const asset = props.asset
  if (!asset) return '未选择媒体资源'
  return asset.alt || asset.original_filename || asset.filename || `媒体 #${asset.id || '-'}`
})
const expectedConfirmation = computed(() => props.asset?.id ? `DELETE ${props.asset.id}` : '')
const hasReferences = computed(() => Number(props.total || 0) > 0)
</script>
