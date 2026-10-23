<template>
  <Dialog :open="Boolean(conversation)" @update:open="closeConversationHistoryWhenDismissed">
    <DialogContent
      size="md"
      class="flex h-[760px] max-w-[calc(100dvw-2rem)] flex-col gap-0 overflow-hidden p-0 data-[size=md]:w-[800px]"
    >
      <DialogHeader class="shrink-0 border-b px-5 py-4 pr-14 text-left">
        <DialogTitle>聊天记录</DialogTitle>
        <DialogDescription class="break-words text-xs leading-5">
          <span class="font-bold text-foreground">{{ conversationDisplayName(conversation) }}</span>
          <span class="mx-1.5">·</span>
          #{{ conversation?.id }}<span v-if="conversation?.ticket_number || conversation?.conversation_id"> · {{ conversation.ticket_number || conversation.conversation_id }}</span>
        </DialogDescription>
      </DialogHeader>

      <div
        :key="`${conversation?.id}-${pageNumber}`"
        class="min-h-0 flex-1 overflow-x-hidden overflow-y-auto overscroll-contain bg-muted/15 p-4 [scrollbar-gutter:stable] sm:p-5"
        :aria-busy="loading"
        aria-label="当前会话聊天记录"
        tabindex="0"
      >
        <div v-if="loading" role="status" class="flex h-full min-h-32 items-center justify-center gap-2 text-sm text-muted-foreground">
          <LoaderCircle class="size-4 animate-spin" />
          正在加载聊天记录……
        </div>
        <div v-else-if="loadError" role="alert" class="flex h-full min-h-32 flex-col items-center justify-center gap-3 text-center">
          <p class="text-sm text-destructive">{{ loadError }}</p>
          <Button variant="outline" size="sm" @click="loadConversationHistoryPage(pageNumber)">重试</Button>
        </div>
        <p v-else-if="!messages.length" role="status" class="flex h-full min-h-32 items-center justify-center text-sm text-muted-foreground">该会话暂无聊天记录。</p>
        <ol v-else class="space-y-5">
          <li v-for="entry in historyMessageEntries" :key="entry.message.id" :class="['flex', entry.message.is_agent ? 'justify-end' : 'justify-start']">
            <article class="min-w-0 max-w-[92%] sm:max-w-[85%]">
              <div :class="['mb-1.5 flex flex-wrap items-baseline gap-x-2 gap-y-1 text-xs text-muted-foreground', { 'justify-end': entry.message.is_agent }]">
                <span class="break-words font-bold text-foreground">{{ entry.message.sender_name || (entry.message.is_agent ? '客服' : conversationDisplayName(conversation)) }}</span>
                <span>{{ entry.message.is_agent ? '客服' : '客户' }}</span>
                <time>{{ formatConversationHistoryMessageTime(entry.message.created_at) }}</time>
              </div>
              <div :class="['min-w-0 rounded-2xl border px-4 py-3 text-sm leading-6 [overflow-wrap:anywhere]', entry.message.is_agent ? 'border-primary/10 bg-primary/5' : 'border-border/60 bg-background']">
                <p v-if="entry.content" class="whitespace-pre-wrap">{{ entry.content }}</p>
                <p v-if="entry.details" :class="['whitespace-pre-wrap text-xs leading-5', { 'mt-2 border-t border-border/50 pt-2': entry.content }]">{{ entry.details }}</p>
                <a v-if="entry.link" :href="entry.link" target="_blank" rel="noopener noreferrer" class="mt-2 block text-xs text-primary underline underline-offset-4">{{ entry.link }}</a>
                <div v-if="entry.message.message_type === 'image'" class="mt-2 grid gap-2">
                  <a v-for="(attachment, index) in entry.attachments" :key="attachment" :href="attachment" target="_blank" rel="noopener noreferrer" :aria-label="`查看图片附件 ${index + 1}`">
                    <img :src="attachment" :alt="`图片附件 ${index + 1}`" loading="lazy" class="max-h-64 max-w-full rounded-lg object-contain" />
                  </a>
                </div>
                <video v-else-if="entry.videoUrl" :src="entry.videoUrl" controls playsinline preload="none" class="mt-2 max-h-64 max-w-full rounded-lg" />
                <template v-else>
                  <a v-for="(attachment, index) in entry.attachments" :key="attachment" :href="attachment" target="_blank" rel="noopener noreferrer" class="mt-2 mr-3 inline-block text-xs text-primary underline underline-offset-4">查看附件 {{ index + 1 }}</a>
                </template>
              </div>
            </article>
          </li>
        </ol>
      </div>

      <div class="flex shrink-0 flex-wrap items-center justify-between gap-2 border-t px-5 py-3 text-xs text-muted-foreground">
        <span aria-live="polite">共 {{ totalMessages }} 条 · 第 {{ pageNumber }} / {{ totalPages }} 页</span>
        <div class="flex gap-2">
          <Button variant="outline" size="sm" :disabled="loading || pageNumber <= 1" @click="loadConversationHistoryPage(pageNumber - 1)">上一页</Button>
          <Button variant="outline" size="sm" :disabled="loading || pageNumber >= totalPages" @click="loadConversationHistoryPage(pageNumber + 1)">下一页</Button>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, toRef } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useReadOnlyConversationHistoryWithIsolatedPagination } from '@/composables/customerService/useReadOnlyConversationHistoryWithIsolatedPagination'
import { conversationDisplayName, formatMoneyWithCurrency, formatProductPrice, messageMetadata } from '@/lib/customerServicePresentation'
import type { CustomerConversation, CustomerConversationMessage } from '@/modules/customer-service/customerServiceTypes'

const props = defineProps<{ conversation: CustomerConversation | null }>()
const emit = defineEmits<{ (event: 'close'): void }>()
const { messages, loading, loadError, pageNumber, totalMessages, totalPages, loadConversationHistoryPage } =
  useReadOnlyConversationHistoryWithIsolatedPagination(toRef(props, 'conversation'))

function closeConversationHistoryWhenDismissed(open: boolean) {
  if (!open) emit('close')
}

function safeConversationHistoryMessageUrl(value: unknown): string {
  if (typeof value !== 'string' || !value.trim()) return ''
  const normalized = value.trim()
  if (/^\/(?!\/)/.test(normalized) && !normalized.includes('\\')) return normalized
  try {
    const url = new URL(normalized)
    return ['https:', 'http:'].includes(url.protocol) ? url.href : ''
  } catch { return '' }
}

function formatConversationHistoryMessageTime(value: CustomerConversationMessage['created_at']): string {
  if (!value) return '时间未知'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? '时间未知'
    : new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}

function describeStructuredConversationHistoryMessage(message: CustomerConversationMessage): string {
  const metadata = messageMetadata(message)
  switch (message.message_type) {
    case 'product':
      return ['产品', metadata.title, metadata.sku ? `SKU：${metadata.sku}` : '', formatProductPrice(metadata)].filter(Boolean).join('\n')
    case 'order':
      return [
        metadata.order_number ? `订单：${metadata.order_number}` : '订单',
        metadata.title,
        metadata.status ? `状态：${metadata.status}` : '',
        metadata.total != null ? `金额：${formatMoneyWithCurrency(metadata.total, metadata.currency)}` : '',
        ...(Array.isArray(metadata.items) ? metadata.items.map((item: Record<string, unknown>) => `${item.title || item.product_name || '商品'} × ${item.quantity || 1}`) : []),
      ].filter(Boolean).join('\n')
    case 'config_confirm': {
      const selection = metadata.selections || {}
      return [
        '配置确认', metadata.product?.title, selection.variant_title,
        ...(Array.isArray(selection.options) ? selection.options.map((option: Record<string, unknown>) => `${option.label || option.key || '选项'}：${option.value ?? ''}${option.unit ? ` ${option.unit}` : ''}`) : []),
        selection.weight_grams ? `重量：${selection.weight_grams}g` : '',
      ].filter(Boolean).join('\n')
    }
    default: return metadata.title && metadata.title !== (message.content || message.message) ? String(metadata.title) : ''
  }
}

const historyMessageEntries = computed(() => messages.value.map((message) => {
  const metadata = messageMetadata(message)
  const attachments = [...new Set([
    ...(Array.isArray(message.attachments) ? message.attachments : []),
    message.attachment_url,
    ...(['image', 'video'].includes(message.message_type || '') ? [metadata.url] : []),
  ].map(safeConversationHistoryMessageUrl).filter(Boolean))]
  return {
    message,
    content: message.content || message.message || '',
    details: describeStructuredConversationHistoryMessage(message),
    attachments,
    link: message.message_type === 'link' ? safeConversationHistoryMessageUrl(metadata.url) : '',
    videoUrl: message.message_type === 'video' ? attachments[0] || '' : '',
  }
}))
</script>
