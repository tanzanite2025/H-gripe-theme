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
        <div class="aspect-video overflow-hidden border-b bg-muted/20">
          <img
            v-if="isImageAttachment(attachment)"
            :src="attachmentUrl(attachment)"
            :alt="attachment.filename || '售后图片附件'"
            class="size-full object-cover"
            loading="lazy"
          >
          <video
            v-else-if="isVideoAttachment(attachment)"
            class="size-full bg-black object-contain"
            controls
            preload="metadata"
          >
            <source :src="attachmentUrl(attachment)" :type="attachment.content_type || undefined">
          </video>
          <div v-else class="flex size-full items-center justify-center">
            <FileText class="size-5 text-muted-foreground" />
          </div>
        </div>

        <div class="flex items-start justify-between gap-3 p-3">
          <div class="min-w-0 flex-1">
            <p class="truncate text-xs font-bold text-foreground" :title="attachment.filename || '未命名附件'">
              {{ attachment.filename || '未命名附件' }}
            </p>
            <div class="mt-1 flex items-center gap-1 text-[10px] font-bold text-muted-foreground">
              <ImageIcon v-if="isImageAttachment(attachment)" class="size-3" />
              <Video v-else-if="isVideoAttachment(attachment)" class="size-3" />
              <FileText v-else class="size-3" />
              <span>{{ attachmentKindLabel(attachment) }}</span>
              <span>·</span>
              <span>{{ formatFileSize(attachment.size_bytes) }}</span>
            </div>
          </div>
          <Button
            as="a"
            variant="outline"
            size="icon"
            :href="attachmentUrl(attachment)"
            target="_blank"
            rel="noopener noreferrer"
            :aria-label="`打开附件 ${attachment.filename || attachment.id}`"
            title="打开附件"
          >
            <ExternalLink class="size-3.5" />
          </Button>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { ExternalLink, FileText, Image as ImageIcon, Video } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { afterSalesAttachmentUrl } from '@/api/afterSales'
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
  return afterSalesAttachmentUrl(props.caseId, attachment.id)
}
const isImageAttachment = (attachment: AfterSalesCaseAttachment): boolean => (
  attachment.kind === 'image' || String(attachment.content_type || '').startsWith('image/')
)
const isVideoAttachment = (attachment: AfterSalesCaseAttachment): boolean => (
  attachment.kind === 'video' || String(attachment.content_type || '').startsWith('video/')
)
const attachmentKindLabel = (attachment: AfterSalesCaseAttachment): string => {
  if (isImageAttachment(attachment)) return '图片附件'
  if (isVideoAttachment(attachment)) return '视频附件'
  return '文件附件'
}
const formatFileSize = (value?: number | null): string => {
  const size = Number(value || 0)
  if (!Number.isFinite(size) || size <= 0) return '未知大小'
  if (size < 1024) return `${size} B`
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
}
</script>
