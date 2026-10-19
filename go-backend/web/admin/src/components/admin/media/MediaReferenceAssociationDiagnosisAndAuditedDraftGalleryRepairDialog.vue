<template>
  <Dialog :open="review.open.value" @update:open="(value) => { if (!value) review.closeMediaAssociationDiagnosis() }">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>异常关联诊断</DialogTitle>
        <DialogDescription>媒体 #{{ review.assetID.value }} · 仅核查已识别业务来源，历史引用继续保护文件。</DialogDescription>
      </DialogHeader>
      <p v-if="review.error.value" role="alert" class="text-xs text-destructive">{{ review.error.value }}</p>
      <p v-if="review.notice.value" role="status" class="text-xs">{{ review.notice.value }}</p>
      <p v-if="review.loading.value" class="text-xs text-muted-foreground">正在核查</p>
      <div v-if="review.report.value" class="max-h-[55vh] space-y-3 overflow-y-auto">
        <p v-if="review.report.value.has_restricted_references" class="text-xs text-amber-700">含受限引用，继续保留；当前核查不提供修复操作。</p>
        <p v-if="!review.report.value.entries.length" class="text-xs text-muted-foreground">本次可见范围未发现缺失父记录、删除保留或登记冲突。</p>
        <div v-for="(entry, index) in review.report.value.entries" :key="index" class="space-y-2 rounded-lg border p-3 text-xs">
          <p class="font-bold">{{ entry.label }}</p>
          <p v-if="entry.diagnosis === 'deleted_history_must_remain_retained'">已删除业务或图片仍保留引用。请在原业务中核查恢复需求。</p>
          <p v-else-if="entry.diagnosis === 'registration_conflict_requires_ownership_evidence'">须先核对原始登记与存储所有权，继续保留全部登记。</p>
          <p v-else-if="entry.resource_type === 'ticket_message_attachment_binding'">附件绑定仍在，但原消息已缺失。继续保留文件；需核对原始记录和审计，不能凭附件推断客户或迁入其他会话。</p>
          <p v-else>父记录缺失，需人工核查正确归属。商品、工单及其他类型请在原业务处理。</p>
          <Button v-if="canRepair && !review.report.value.has_restricted_references && entry.resource_type === 'gallery_image' && entry.usage_state === 'missing_parent'"
            size="sm" variant="outline" :disabled="review.loading.value || review.saving.value" @click="review.selectMissingGalleryImageForReviewedRepair(entry)">审核迁入草稿图库</Button>
        </div>
        <p v-if="review.report.value.entries_limited" class="text-xs text-amber-700">仅展示前 200 项；其余关联继续保留。</p>
        <div v-if="review.selectedImageID.value && canRepair" class="space-y-3 rounded-lg border p-3">
          <p class="text-xs">图片 #{{ review.selectedImageID.value }}。请先在图库中新建或确认目标草稿；仅支持精确原图绑定且未删除的图片。</p>
          <label class="block text-xs">目标草稿图库 ID
            <Input v-model="review.destinationGalleryID.value" inputmode="numeric" :disabled="review.saving.value" class="mt-1" />
          </label>
          <Button size="sm" variant="outline" :disabled="review.loading.value || review.saving.value" @click="review.previewMissingGalleryParentRepair">生成修复预览</Button>
          <div v-if="review.preview.value" class="space-y-3 text-xs">
            <p>确认将“{{ review.preview.value.image_title || '图片' }}”从缺失图库 #{{ review.preview.value.missing_gallery_id }} 迁入草稿“{{ review.preview.value.destination_gallery_name }}” #{{ review.preview.value.destination_gallery_id }}。原图、缩略图、排序和说明均保留。</p>
            <label class="block">审核理由（3 至 1000 字）<Input v-model="review.reason.value" :disabled="review.saving.value" class="mt-1" /></label>
            <Button size="sm" :disabled="review.loading.value || review.saving.value || [...review.reason.value.trim()].length < 3 || [...review.reason.value.trim()].length > 1000" @click="review.applyReviewedMissingGalleryParentRepair">确认迁入并记录审计</Button>
          </div>
        </div>
      </div>
      <DialogFooter>
        <Button variant="outline" :disabled="review.loading.value || review.saving.value" @click="review.loadMediaAssociationDiagnosis">重新诊断</Button>
        <Button variant="outline" @click="review.closeMediaAssociationDiagnosis">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
<script setup lang="ts">
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation } from '@/composables/media/useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation'
defineProps<{ review: ReturnType<typeof useMediaAssociationDiagnosisAndReviewedRepairWithCancellationAndScopeInvalidation>; canRepair: boolean }>()
</script>
