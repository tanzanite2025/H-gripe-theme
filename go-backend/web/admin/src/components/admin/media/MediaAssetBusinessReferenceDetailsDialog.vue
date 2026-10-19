<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>媒体使用位置</DialogTitle>
        <DialogDescription>{{ assetTitle(asset) }} · {{ asset ? mediaAssetLibraryOwnershipLabel(asset) : '' }}</DialogDescription>
      </DialogHeader>
      <p v-if="loading" class="py-6 text-center text-xs text-muted-foreground">正在查询使用位置</p>
      <p v-else-if="loadError" role="alert" class="py-6 text-center text-xs text-destructive">{{ loadError }}</p>
      <p v-else-if="!references.length" class="py-6 text-center text-xs text-muted-foreground">暂无已识别使用位置</p>
      <div v-else class="max-h-80 space-y-2 overflow-y-auto">
        <div v-for="(reference, index) in references" :key="`${reference.resource_type}-${reference.resource_id}-${index}`" class="rounded-lg border p-3">
          <p class="text-xs font-bold">{{ reference.label || mediaReferenceTypeLabel(reference.resource_type) }}</p>
          <p class="mt-1 text-[10px] text-muted-foreground">{{ mediaReferenceCategoryLabel(reference.category) }} · {{ mediaReferenceTypeLabel(reference.resource_type) }}</p>
          <p v-if="reference.usage_state === 'retained_deleted_parent'" class="mt-1 text-xs text-amber-700">已删除内容保留引用，文件仍受保护</p>
          <p v-else-if="reference.usage_state === 'retained_deleted_placement'" class="mt-1 text-xs text-amber-700">已删除图片记录保留引用，文件仍受保护</p>
          <p v-else-if="reference.usage_state === 'retained_unpublished_parent'" class="mt-1 text-xs text-amber-700">未发布内容保留引用，文件仍受保护</p>
          <p v-else-if="reference.usage_state === 'retained_upload_receipt'" class="mt-1 text-xs text-amber-700">已绑定的上传收据仍保留对象，不可作为待发送附件重复使用</p>
          <p v-else-if="reference.usage_state === 'missing_parent'" class="mt-1 text-xs text-amber-700">关联的业务记录已不存在，需要核查关联</p>
          <p v-else-if="reference.usage_state === 'disabled'" class="mt-1 text-xs text-amber-700">已停用的自动回复仍保留此媒体，移除引用后才能清理</p>
        </div>
      </div>
      <DialogFooter>
        <Button v-if="asset" variant="outline" @click="emit('diagnose', asset)">异常关联诊断</Button>
        <Button variant="outline" @click="emit('update:open', false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { onScopeDispose, ref, watch } from 'vue'
import mediaApi, { type MediaAsset, type MediaReference } from '@/api/media'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { assetTitle } from '@/lib/mediaPresentation'
import { mediaReferenceCategoryLabel, mediaReferenceTypeLabel } from '@/lib/mediaReferencePresentation'
import { mediaAssetLibraryOwnershipLabel } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'
const props = defineProps<{ open: boolean; asset: MediaAsset | null }>()
const emit = defineEmits<{ (event: 'update:open', value: boolean): void; (event: 'diagnose', asset: MediaAsset): void }>()
const loading = ref(false)
const loadError = ref('')
const references = ref<MediaReference[]>([])
let requestGeneration = 0
onScopeDispose(() => { requestGeneration++ })
watch(() => [props.open, props.asset?.id], async () => {
  const generation = ++requestGeneration
  references.value = []
  loadError.value = ''
  if (!props.open || !props.asset) { loading.value = false; return }
  loading.value = true
  try {
    const result = await mediaApi.getAssetReferences(props.asset.id)
    if (generation === requestGeneration) references.value = result.references
  } catch {
    if (generation === requestGeneration) loadError.value = '使用位置查询失败，请关闭后重试'
  } finally {
    if (generation === requestGeneration) loading.value = false
  }
})
</script>
