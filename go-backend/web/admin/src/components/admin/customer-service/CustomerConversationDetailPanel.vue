<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <template v-if="selectedConversation">
      <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0">
            <CardTitle class="truncate">{{ selectedConversation.customer_name || '匿名客户' }}</CardTitle>
            <CardDescription class="break-all">
              {{ selectedConversation.conversation_id || selectedConversation.ticket_number || selectedConversation.id }}
            </CardDescription>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <AdminStatusBadge :tone="statusTone(selectedConversation.display_status)">
              {{ statusLabel(selectedConversation.display_status) }}
            </AdminStatusBadge>
            <AdminStatusBadge :tone="conversationIsMember(selectedConversation) ? 'green' : 'amber'">
              {{ customerIdentityLabel(selectedConversation) }}
            </AdminStatusBadge>
            <span
              v-if="memberTier(selectedConversation)"
              class="inline-flex h-5 items-center gap-1 rounded-full border border-amber-500/20 bg-amber-500/10 px-2 text-[10px] font-black text-amber-700"
              :style="memberTierStyle(selectedConversation)"
            >
              <span v-if="memberTierIcon(selectedConversation)" class="leading-none">{{ memberTierIcon(selectedConversation) }}</span>
              {{ memberTierName(selectedConversation) }}
            </span>
            <AdminStatusBadge tone="gray">{{ customerRegionLabel(selectedConversation) }}</AdminStatusBadge>
          </div>
        </div>
      </CardHeader>

      <CardContent class="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)_auto] gap-0 p-0">
        <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
          <div class="text-xs text-muted-foreground">
            当前负责人：
            <strong class="text-foreground">{{ assigneeName(selectedConversation.assigned_to, assignableAgents) }}</strong>
          </div>

          <div v-if="canEdit" class="flex flex-wrap items-center gap-2">
            <Select v-model="transferModel">
              <SelectTrigger class="h-9 w-44"><SelectValue placeholder="选择客服" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="agent in assignableAgents" :key="agent.user_id || agent.id" :value="String(agent.user_id || agent.id)">
                  {{ agent.name || agent.email || `用户 ${agent.user_id || agent.id}` }}
                </SelectItem>
              </SelectContent>
            </Select>
            <Button variant="outline" size="sm" class="rounded-full" :disabled="transferring || !transferModel" @click="emit('transfer')">
              <ArrowRightLeft v-if="!transferring" class="size-3.5" />
              <LoaderCircle v-else class="size-3.5 animate-spin" />
              转接
            </Button>
          </div>
        </div>

        <div class="relative min-h-0 overflow-y-auto px-4 py-4">
          <div v-if="messagesLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-card/75">
            <LoaderCircle class="size-5 animate-spin text-primary" />
          </div>
          <div v-else-if="messages.length === 0" class="flex h-72 flex-col items-center justify-center text-muted-foreground">
            <MessageCircleOff class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无消息记录</span>
          </div>
          <div v-else class="space-y-3">
            <article
              v-for="message in messages"
              :key="message.id"
              class="max-w-[86%] rounded-2xl border px-4 py-3 text-sm"
              :class="message.is_agent ? 'ml-auto border-blue-200 bg-blue-50/75' : 'mr-auto border-border bg-muted/45'"
            >
              <header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                <div class="flex items-center gap-2">
                  <span class="text-xs font-black">{{ message.sender_name || (message.is_agent ? '客服' : '客户') }}</span>
                  <AdminStatusBadge :tone="message.is_agent ? 'blue' : 'gray'">
                    {{ message.is_agent ? '客服' : '客户' }}
                  </AdminStatusBadge>
                </div>
                <time class="text-[11px] text-muted-foreground">{{ formatDate(message.created_at) }}</time>
              </header>

              <div
                v-if="isConfigConfirmMessage(message)"
                class="mt-3 rounded-2xl border border-indigo-200 bg-indigo-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-indigo-600">
                  配置确认请求
                </div>
                <div class="flex gap-3">
                  <div class="size-16 shrink-0 overflow-hidden rounded-xl bg-muted">
                    <img
                      v-if="configProduct(message).thumbnail"
                      :src="configProduct(message).thumbnail"
                      :alt="configProduct(message).title || 'Product'"
                      class="size-full object-cover"
                    />
                  </div>
                  <div class="min-w-0 flex-1">
                    <a
                      v-if="configProduct(message).url"
                      :href="configProduct(message).url"
                      target="_blank"
                      rel="noreferrer"
                      class="block truncate text-sm font-black text-foreground underline-offset-4 hover:underline"
                    >
                      {{ configProduct(message).title || message.content || message.message }}
                    </a>
                    <p v-else class="truncate text-sm font-black text-foreground">
                      {{ configProduct(message).title || message.content || message.message }}
                    </p>
                    <p v-if="configProduct(message).price" class="mt-1 text-xs font-bold text-emerald-600">
                      {{ configProduct(message).price }}
                    </p>
                    <p v-if="configProduct(message).sku" class="mt-1 text-[11px] text-muted-foreground">
                      SKU：{{ configProduct(message).sku }}
                    </p>
                  </div>
                </div>
                <div class="mt-3 rounded-xl bg-white/70 px-3 py-2 text-xs leading-5 text-muted-foreground">
                  <p v-if="configSelection(message).variant_title" class="font-bold text-foreground">
                    已选：{{ configSelection(message).variant_title }}
                  </p>
                  <div v-if="configOptionRows(message).length" class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
                    <div
                      v-for="option in configOptionRows(message)"
                      :key="option.key"
                      class="rounded-xl border border-indigo-100 bg-indigo-50/50 px-2.5 py-2"
                    >
                      <span class="block text-[10px] font-black uppercase tracking-wider text-indigo-500">
                        {{ option.label }}
                      </span>
                      <span class="mt-0.5 block font-bold text-foreground">
                        {{ option.value }}<span v-if="option.unit"> {{ option.unit }}</span>
                      </span>
                    </div>
                  </div>
                  <p v-if="configSelection(message).weight_grams" class="mt-2">
                    重量：{{ configSelection(message).weight_grams }}g
                  </p>
                  <p
                    v-if="!configOptionRows(message).length && !configSelection(message).variant_title && !configSelection(message).weight_grams"
                  >
                    客户请求客服确认该产品配置。
                  </p>
                </div>
              </div>

              <div
                v-else-if="isOrderMessage(message)"
                class="mt-3 rounded-2xl border border-emerald-200 bg-emerald-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-emerald-600">
                  订单确认请求
                </div>
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0 flex-1">
                    <a
                      v-if="orderPayload(message).url"
                      :href="orderPayload(message).url"
                      target="_blank"
                      rel="noreferrer"
                      class="block truncate text-sm font-black text-foreground underline-offset-4 hover:underline"
                    >
                      {{ orderPayload(message).title || message.content || message.message }}
                    </a>
                    <p v-else class="truncate text-sm font-black text-foreground">
                      {{ orderPayload(message).title || message.content || message.message }}
                    </p>
                    <p class="mt-1 text-xs font-bold text-emerald-700">
                      {{ formatOrderTotal(orderPayload(message)) }}
                    </p>
                  </div>
                  <AdminStatusBadge tone="green">
                    {{ orderPayload(message).status || 'order' }}
                  </AdminStatusBadge>
                </div>
                <div class="mt-2 flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                  <span v-if="orderPayload(message).payment_status">支付：{{ orderPayload(message).payment_status }}</span>
                  <span v-if="orderPayload(message).shipping_status">物流：{{ orderPayload(message).shipping_status }}</span>
                  <span v-if="orderPayload(message).item_count">商品：{{ orderPayload(message).item_count }} 件</span>
                </div>
                <div v-if="orderItems(message).length" class="mt-3 space-y-2">
                  <article
                    v-for="item in orderItems(message).slice(0, 4)"
                    :key="item.id || `${item.product_id}-${item.sku}`"
                    class="rounded-xl border border-emerald-100 bg-white/70 px-3 py-2 text-xs"
                  >
                    <div class="flex items-center justify-between gap-2">
                      <p class="truncate font-bold text-foreground">{{ item.title || item.product_name || 'Product' }}</p>
                      <span class="shrink-0 font-mono text-muted-foreground">x{{ item.quantity || 1 }}</span>
                    </div>
                    <p v-if="item.sku" class="mt-1 text-[11px] text-muted-foreground">SKU：{{ item.sku }}</p>
                  </article>
                  <p v-if="orderItems(message).length > 4" class="text-[11px] text-muted-foreground">
                    还有 {{ orderItems(message).length - 4 }} 个商品
                  </p>
                </div>
              </div>

              <p v-else class="mt-2 whitespace-pre-wrap break-words leading-6">{{ message.content || message.message }}</p>
              <div v-if="message.message_type === 'image' && messageAttachments(message).length" class="mt-2 grid gap-2">
                <img
                  v-for="attachmentUrl in messageAttachments(message)"
                  :key="attachmentUrl"
                  :src="attachmentUrl"
                  alt="客服图片附件"
                  class="max-h-64 max-w-full rounded-xl object-contain"
                />
              </div>
              <a
                v-if="message.message_type === 'link' && message.metadata?.url"
                :href="message.metadata.url"
                target="_blank"
                rel="noreferrer"
                class="mt-2 block rounded-xl border border-emerald-200 bg-emerald-50/60 px-3 py-2 text-xs text-emerald-700 underline-offset-4 hover:underline"
              >
                <span class="block font-black">{{ message.metadata.title || message.content || message.message }}</span>
                <span class="mt-1 block truncate text-emerald-700/70">{{ message.metadata.url }}</span>
              </a>
              <a
                v-for="attachmentUrl in messageAttachments(message)"
                :key="`link-${attachmentUrl}`"
                class="mt-2 inline-flex text-xs font-bold text-primary underline-offset-4 hover:underline"
                :href="attachmentUrl"
                target="_blank"
                rel="noreferrer"
              >
                查看附件
              </a>
            </article>
          </div>
          <div
            v-if="selectedCustomerTyping?.active"
            class="mt-3 inline-flex items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700"
          >
            <span>{{ selectedCustomerTyping.displayName || '客户' }} 正在输入</span>
            <span class="flex gap-0.5">
              <span class="size-1 animate-pulse rounded-full bg-emerald-500"></span>
              <span class="size-1 animate-pulse rounded-full bg-emerald-500 [animation-delay:120ms]"></span>
              <span class="size-1 animate-pulse rounded-full bg-emerald-500 [animation-delay:240ms]"></span>
            </span>
          </div>
        </div>

        <form v-if="canEdit" class="border-t p-4" @submit.prevent="emit('send-reply')">
          <Textarea
            v-model="replyModel"
            class="min-h-24 resize-none"
            placeholder="输入回复内容，发送后客户侧可在原会话中看到"
            @input="emit('typing-input')"
          />
          <div class="mt-3 flex justify-end">
            <Button type="submit" class="rounded-full" :disabled="replying || !replyModel.trim()">
              <LoaderCircle v-if="replying" class="size-4 animate-spin" />
              <Send v-else class="size-4" />
              发送回复
            </Button>
          </div>
        </form>
      </CardContent>
    </template>

    <div v-else class="flex h-full min-h-0 flex-col items-center justify-center p-8 text-center text-muted-foreground">
      <Headset class="mb-3 size-10 opacity-50" />
      <h2 class="text-sm font-black text-foreground">选择一个客户会话</h2>
      <p class="mt-2 max-w-sm text-xs leading-6">
        左侧每个卡片是一条独立 Public Chat 会话。客户之间不会共用消息窗口。
      </p>
    </div>
  </Card>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowRightLeft, Headset, LoaderCircle, MessageCircleOff, Send } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  assigneeName,
  configOptionRows,
  configProduct,
  configSelection,
  conversationIsMember,
  customerIdentityLabel,
  customerRegionLabel,
  formatDate,
  formatOrderTotal,
  isConfigConfirmMessage,
  isOrderMessage,
  memberTier,
  memberTierIcon,
  memberTierName,
  memberTierStyle,
  orderItems,
  orderPayload,
  statusLabel,
  statusTone,
} from '@/lib/customerServicePresentation'

const props = defineProps({
  selectedConversation: { type: Object, default: null },
  messages: { type: Array, default: () => [] },
  messagesLoading: { type: Boolean, default: false },
  selectedCustomerTyping: { type: Object, default: null },
  assignableAgents: { type: Array, default: () => [] },
  transferTo: { type: String, default: '' },
  replyMessage: { type: String, default: '' },
  transferring: { type: Boolean, default: false },
  replying: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:transferTo',
  'update:replyMessage',
  'transfer',
  'send-reply',
  'typing-input',
])

const transferModel = computed({
  get: () => props.transferTo,
  set: (value) => emit('update:transferTo', value),
})

const replyModel = computed({
  get: () => props.replyMessage,
  set: (value) => emit('update:replyMessage', value),
})

const messageAttachments = (message) => {
  const attachments = Array.isArray(message?.attachments)
    ? message.attachments
    : (message?.attachment_url ? [message.attachment_url] : [])
  const unique = new Set()
  attachments.forEach((value) => {
    const attachment = String(value || '').trim()
    if (attachment) unique.add(attachment)
  })
  return Array.from(unique)
}
</script>
