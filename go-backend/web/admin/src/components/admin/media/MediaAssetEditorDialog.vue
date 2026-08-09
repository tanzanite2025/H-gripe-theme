<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>编辑媒体信息</DialogTitle>
        <DialogDescription>只编辑资源元数据，不修改物理文件路径，避免引用链断裂。</DialogDescription>
      </DialogHeader>
      <div class="grid gap-4 md:grid-cols-[16rem_minmax(0,1fr)]">
        <div class="overflow-hidden rounded-2xl bg-muted">
          <img v-if="asset?.media_type === 'image'" :src="assetAccessURL(asset)" :alt="alt" class="aspect-[4/3] size-full object-cover" />
          <div v-else class="flex aspect-[4/3] items-center justify-center text-muted-foreground">
            <FileVideo class="size-8" />
          </div>
        </div>
        <div class="space-y-3">
          <label class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">ALT 文本</span>
            <Input :model-value="alt" @update:model-value="(value) => emit('update:alt', String(value))" />
          </label>
          <label class="block space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">说明</span>
            <Textarea :model-value="caption" class="min-h-24" @update:model-value="(value) => emit('update:caption', String(value))" />
          </label>
          <div class="grid grid-cols-2 gap-3">
            <label class="block space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">状态</span>
              <Select :model-value="status" @update:model-value="(value) => emit('update:status', String(value))">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">启用</SelectItem>
                  <SelectItem value="archived">归档</SelectItem>
                </SelectContent>
              </Select>
            </label>
            <label class="block space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">可见性</span>
              <Select :model-value="visibility" @update:model-value="(value) => emit('update:visibility', String(value))">
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="public">公开</SelectItem>
                  <SelectItem value="private">私有</SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
        <Button :disabled="saving" @click="emit('save')">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          保存
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { FileVideo, LoaderCircle, Save } from '@lucide/vue'
import type { MediaAsset } from '@/api/media'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { assetAccessURL } from '@/lib/mediaPresentation'

withDefaults(defineProps<{
  open?: boolean
  asset?: MediaAsset | null
  alt?: string
  caption?: string
  status?: string
  visibility?: string
  saving?: boolean
}>(), {
  open: false,
  asset: null,
  alt: '',
  caption: '',
  status: 'active',
  visibility: 'public',
  saving: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:alt', value: string): void
  (event: 'update:caption', value: string): void
  (event: 'update:status', value: string): void
  (event: 'update:visibility', value: string): void
  (event: 'save'): void
}>()
</script>
