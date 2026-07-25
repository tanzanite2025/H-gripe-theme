<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>{{ currentTicket?.ticket_number || '工单详情' }}</DialogTitle>
        <DialogDescription>{{ currentTicket?.subject || '查看工单信息和消息记录' }}</DialogDescription>
      </DialogHeader>

      <div class="relative min-h-80">
        <div v-if="detailLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载工单详情" />
        </div>

        <div v-if="currentTicket" class="grid lg:grid-cols-[320px_minmax(0,1fr)]">
          <aside class="space-y-6 border-b p-5 lg:border-b-0 lg:border-r">
            <section class="space-y-3">
              <h3 class="text-sm font-black tracking-tighter italic uppercase">工单信息</h3>
              <dl class="divide-y rounded-lg border">
                <DetailItem label="状态">
                  <AdminStatusBadge :tone="statusTone(currentTicket.status)">{{ statusName(currentTicket.status) }}</AdminStatusBadge>
                </DetailItem>
                <DetailItem label="优先级">
                  <AdminStatusBadge :tone="priorityTone(currentTicket.priority)">{{ priorityName(currentTicket.priority) }}</AdminStatusBadge>
                </DetailItem>
                <DetailItem label="分类">{{ categoryName(currentTicket.category) }}</DetailItem>
                <DetailItem label="客户">{{ customerName(currentTicket) }}</DetailItem>
                <DetailItem label="负责人">{{ assigneeName(currentTicket.assigned_to) }}</DetailItem>
                <DetailItem label="创建时间">{{ formatDate(currentTicket.created_at) }}</DetailItem>
                <DetailItem label="更新时间">{{ formatDate(currentTicket.updated_at) }}</DetailItem>
                <DetailItem v-if="currentTicket.tags" label="标签">{{ currentTicket.tags }}</DetailItem>
              </dl>
            </section>

            <section v-if="canEdit" class="space-y-3 border-t pt-5">
              <h3 class="text-sm font-black tracking-tighter italic uppercase">处理操作</h3>
              <label class="block space-y-1.5">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">状态</span>
                <Select :model-value="statusUpdate" @update:model-value="emit('update:statusUpdate', $event)">
                  <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="option in editableStatusOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </label>
              <Button class="w-full" :disabled="statusUpdating || statusUpdate === currentTicket.status" @click="emit('update-status')">
                <LoaderCircle v-if="statusUpdating" class="size-4 animate-spin" />
                更新状态
              </Button>
              <Button variant="outline" class="w-full" @click="emit('show-assign', currentTicket)">
                <UserRoundCog class="size-4" />
                {{ currentTicket.assigned_to ? '更换负责人' : '分配负责人' }}
              </Button>
            </section>
          </aside>

          <section class="flex min-h-[620px] min-w-0 flex-col">
            <div class="flex items-center justify-between border-b px-5 py-3">
              <h3 class="text-sm font-black tracking-tighter italic uppercase">消息记录</h3>
              <span class="text-xs text-muted-foreground">{{ messages.length }} 条</span>
            </div>

            <div class="relative min-h-64 flex-1 overflow-y-auto px-5 py-4">
              <div v-if="messagesLoading" class="absolute inset-0 flex items-center justify-center bg-background/75">
                <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载消息" />
              </div>
              <div v-else-if="messages.length === 0" class="flex h-52 flex-col items-center justify-center text-muted-foreground">
                <MessageCircleOff class="mb-2 size-7 opacity-55" />
                <span class="text-xs">暂无消息记录</span>
              </div>
              <div v-else class="space-y-3">
                <article
                  v-for="message in messages"
                  :key="message.id"
                  class="max-w-[88%] rounded-lg border px-3.5 py-3"
                  :class="message.is_staff ? 'ml-auto border-blue-200 bg-blue-50/70' : 'mr-auto bg-muted/40'"
                >
                  <header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                    <div class="flex items-center gap-2">
                      <span class="text-xs font-bold">{{ messageSender(message) }}</span>
                      <AdminStatusBadge :tone="message.is_staff ? 'blue' : 'gray'">
                        {{ message.is_staff ? '客服' : '客户' }}
                      </AdminStatusBadge>
                    </div>
                    <time class="text-[11px] text-muted-foreground">{{ formatDate(message.created_at) }}</time>
                  </header>
                  <p class="mt-2 whitespace-pre-wrap break-words text-sm leading-6">{{ message.content || message.message }}</p>
                </article>
              </div>
            </div>

            <form v-if="canEdit" class="border-t p-4" @submit.prevent="emit('send-reply')">
              <Textarea :model-value="replyMessage" class="min-h-24 resize-y" placeholder="输入回复内容" @update:model-value="emit('update:replyMessage', $event)" />
              <div class="mt-3 flex justify-end">
                <Button type="submit" :disabled="replying || !replyMessage.trim()">
                  <LoaderCircle v-if="replying" class="size-4 animate-spin" />
                  <Send v-else class="size-4" />
                  发送回复
                </Button>
              </div>
            </form>
          </section>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { defineComponent, h } from 'vue'
import { LoaderCircle, MessageCircleOff, Send, UserRoundCog } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  currentTicket: { type: Object, default: null },
  detailLoading: { type: Boolean, default: false },
  messages: { type: Array, default: () => [] },
  messagesLoading: { type: Boolean, default: false },
  replyMessage: { type: String, default: '' },
  replying: { type: Boolean, default: false },
  statusUpdate: { type: String, default: 'open' },
  statusUpdating: { type: Boolean, default: false },
  editableStatusOptions: { type: Array, default: () => [] },
  canEdit: { type: Boolean, default: false },
  categoryName: { type: Function, required: true },
  statusName: { type: Function, required: true },
  statusTone: { type: Function, required: true },
  priorityName: { type: Function, required: true },
  priorityTone: { type: Function, required: true },
  customerName: { type: Function, required: true },
  assigneeName: { type: Function, required: true },
  messageSender: { type: Function, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits([
  'update:open',
  'update:replyMessage',
  'update:statusUpdate',
  'update-status',
  'show-assign',
  'send-reply',
])

const DetailItem = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: '' }
  },
  setup(props, { slots }) {
    return () => h('div', { class: 'space-y-1' }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h('dd', { class: 'text-xs font-bold' }, slots.default ? slots.default() : (props.value || '-'))
    ])
  }
})
</script>
