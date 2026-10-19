<template>
  <Dialog :open="ownershipReview.open.value" @update:open="ownershipReview.setMediaAssetOwnershipReviewDialogOpen">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>整理素材归属</DialogTitle>
        <DialogDescription>选择负责管理这份素材的业务分区，并记录整理理由。已有内容中的使用继续有效。</DialogDescription>
      </DialogHeader>
      <p class="truncate text-sm font-semibold">{{ ownershipReview.asset.value?.original_filename || ownershipReview.asset.value?.filename }}</p>
      <p v-if="ownershipReview.loading.value" role="status" class="py-6 text-sm text-muted-foreground">正在加载归属审核</p>
      <div v-else-if="ownershipReview.review.value" class="space-y-4">
        <dl class="grid grid-cols-[5rem_minmax(0,1fr)] gap-2 rounded-lg border bg-muted/30 p-3 text-sm">
          <dt class="text-muted-foreground">当前归属</dt><dd>{{ ownershipChoiceLabel(ownershipReview.review.value.current) }}</dd>
          <dt class="text-muted-foreground">上传来源</dt><dd>{{ mediaAssetUploadOriginLabel(ownershipReview.review.value.current.upload_origin) }}</dd>
          <dt class="text-muted-foreground">归属依据</dt><dd>{{ ownershipReview.review.value.current.classification_method === 'manual_reviewed' ? '人工整理' : ownershipReview.review.value.current.classification_method === 'recorded_upload_context' ? '上传入口记录' : '历史待确认' }}</dd>
        </dl>
        <div v-if="ownershipReview.review.value.can_reclassify" class="space-y-1 text-xs text-muted-foreground">
          <p>已识别可见使用 {{ ownershipReview.review.value.recognized_reference_total }} 处<span v-for="(count, category) in ownershipReview.review.value.reference_category_counts" :key="category"> · {{ referenceCategoryLabel(String(category)) }} {{ count }}</span></p>
          <p v-if="ownershipReview.review.value.has_restricted_references">另含受限引用，相关数量与来源不展示，请人工核查归属。</p>
          <p v-if="ownershipReview.review.value.retained_deleted_parent_references">已删除业务记录保留引用 {{ ownershipReview.review.value.retained_deleted_parent_references }} 处</p>
          <p v-if="ownershipReview.review.value.missing_parent_references" class="text-destructive">缺失业务父记录 {{ ownershipReview.review.value.missing_parent_references }} 处，请核对。</p>
          <p>使用位置仅作为整理参考，不能确认全部历史来源。</p>
        </div>
        <p v-if="!ownershipReview.review.value.can_reclassify" class="rounded-lg border p-3 text-sm">{{ ownershipRestrictionLabel(ownershipReview.review.value.restriction_reason) }}</p>
        <template v-else>
          <div v-if="ownershipReview.review.value.suggested_ownership" class="flex flex-wrap items-center justify-between gap-2 rounded-lg border p-3 text-xs">
            <span>当前已识别使用集中在{{ ownershipChoiceLabel(ownershipReview.review.value.suggested_ownership) }}，请人工核对。</span>
            <Button size="xs" variant="outline" :disabled="ownershipReview.saving.value || ownershipReview.revisionConflict.value" @click="ownershipReview.selectedPartition.value = ownershipReview.review.value.suggested_ownership.library_partition">采用建议</Button>
          </div>
          <label class="block space-y-2 text-sm"><span>管理归属</span>
            <Select :model-value="ownershipReview.selectedPartition.value" :disabled="ownershipReview.saving.value || ownershipReview.revisionConflict.value" @update:model-value="selectReviewedOwnershipPartition">
              <SelectTrigger aria-label="管理归属"><SelectValue placeholder="请选择管理分区" /></SelectTrigger>
              <SelectContent><SelectItem v-for="choice in ownershipReview.review.value.allowed_choices" :key="choice.library_partition" :value="choice.library_partition">{{ ownershipChoiceLabel(choice) }}</SelectItem></SelectContent>
            </Select>
          </label>
          <p v-if="ownershipReview.review.value.current.classification_method === 'manual_reviewed' && ownershipReview.selectedPartition.value === ownershipReview.review.value.current.library_partition" class="text-xs text-muted-foreground">当前归属已人工整理，选择不同归属后可保存变更。</p>
          <label class="block space-y-2 text-sm"><span>整理理由</span>
            <Textarea v-model="ownershipReview.reason.value" :disabled="ownershipReview.saving.value || ownershipReview.revisionConflict.value" aria-label="整理理由" placeholder="例如：已核对为商品详情通用说明图" class="min-h-20" />
            <span class="text-xs text-muted-foreground">3–1000 字，随归属变更保存到操作审计。</span>
          </label>
        </template>
      </div>
      <p v-if="ownershipReview.error.value" role="alert" class="text-sm text-destructive">{{ ownershipReview.error.value }}</p>
      <DialogFooter>
        <Button variant="outline" :disabled="ownershipReview.saving.value" @click="ownershipReview.setMediaAssetOwnershipReviewDialogOpen(false)">关闭</Button>
        <Button v-if="ownershipReview.error.value" variant="outline" :disabled="ownershipReview.saving.value || ownershipReview.loading.value" @click="ownershipReview.loadCurrentMediaAssetOwnershipReview">重新加载审核</Button>
        <Button :disabled="!ownershipReview.canSave.value" @click="ownershipReview.saveReviewedMediaAssetLibraryOwnership">{{ ownershipReview.saving.value ? '正在保存' : '保存归属并记录审计' }}</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import type { MediaAssetLibraryOwnershipChoice } from '@/api/media'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import type { useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave } from '@/composables/media/useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave'
import { mediaAssetUploadOriginLabel, mediaLibraryPartitionLabel } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'

const props = defineProps<{ ownershipReview: ReturnType<typeof useMediaAssetManualLibraryOwnershipReviewVersionConflictAndTransactionalSave> }>()

function ownershipChoiceLabel(choice: MediaAssetLibraryOwnershipChoice): string {
  return choice.customer_service_media_purpose === 'reusable_reply' ? '客服常用素材' : mediaLibraryPartitionLabel(choice.library_partition)
}

function selectReviewedOwnershipPartition(value: unknown): void {
  const choice = props.ownershipReview.review.value?.allowed_choices.find((item) => item.library_partition === value)
  if (choice) props.ownershipReview.selectedPartition.value = choice.library_partition
}

function referenceCategoryLabel(category: string): string {
  return ({ catalog: '商品', content: '页面', community: '社区', customer: '客户', support: '客服', registration: '媒体登记' } as Record<string, string>)[category] || '其他'
}

function ownershipRestrictionLabel(restriction?: string): string {
  return ({ conversation_attachments_require_conversation_lifecycle: '会话附件请在对应会话中管理。',
    private_assets_require_their_business_access_policy: '私有文件需要通过所属业务管理。', inactive_assets_cannot_be_reclassified: '停用素材暂不能整理归属。',
    customer_submissions_require_separate_publication_review: '客户提交的文件需要另行审核后才能作为运营素材。' } as Record<string, string>)[restriction || ''] || '这份素材暂不支持整理归属。'
}
</script>
