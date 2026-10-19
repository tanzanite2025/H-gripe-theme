<template>
  <section class="space-y-3 border-t border-dashed pt-5">
    <div class="flex items-center justify-between gap-3">
      <h3 class="text-sm font-black uppercase text-foreground">客户附件</h3>
      <span class="text-xs text-muted-foreground">{{ attachments.length }} 个</span>
    </div>

    <div v-if="loading" class="rounded-lg border border-dashed p-5 text-center text-sm text-muted-foreground">
      附件加载中
    </div>
    <div v-else-if="!attachments.length" class="rounded-lg border border-dashed p-5 text-center text-sm text-muted-foreground">
      暂无客户附件
    </div>
    <div v-else class="grid gap-3 sm:grid-cols-2">
      <article
        v-for="attachment in attachments"
        :key="String(attachment.id)"
        class="overflow-hidden rounded-lg border bg-background"
      >
        <AuthorizedPrivateOrderAttachmentPreviewAndDownload class="p-3" :path="attachmentUrl(attachment)" :label="attachment.filename || '售后附件'" :filename="attachment.filename || 'attachment'" />
        <p class="px-3 pb-3 text-xs text-muted-foreground">{{ attachment.content_type || attachment.kind || '文件附件' }} · {{ attachment.size_bytes || 0 }} B</p>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import AuthorizedPrivateOrderAttachmentPreviewAndDownload from '@/components/admin/AuthorizedPrivateOrderAttachmentPreviewAndDownload.vue'
import type { AfterSalesCaseAttachment, AfterSalesCaseID } from '@/api/afterSales'

const props = withDefaults(defineProps<{
  caseId?: AfterSalesCaseID | null
  attachments?: AfterSalesCaseAttachment[] | null
  loading?: boolean
}>(), {
  caseId: null,
  attachments: () => [],
  loading: false,
})

const attachments = computed(() => props.attachments || [])

const attachmentUrl = (attachment: AfterSalesCaseAttachment): string => {
  if (!props.caseId || !attachment.id) return '#'
  return `/api/admin/after-sales/${encodeURIComponent(String(props.caseId))}/attachments/${encodeURIComponent(String(attachment.id))}`
}
</script>
