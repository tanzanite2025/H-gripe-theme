<template>
  <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
    <DialogHeader>
      <DialogTitle>礼品卡详情</DialogTitle>
      <DialogDescription v-if="currentGiftCard" class="font-mono">{{ currentGiftCard.code }}</DialogDescription>
    </DialogHeader>

    <div v-if="loading" class="flex h-52 items-center justify-center">
      <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载礼品卡详情" />
    </div>
    <div v-else-if="currentGiftCard" class="space-y-6">
      <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-3">
        <DetailItem label="状态">
          <AdminStatusBadge :tone="giftCardStatusTone(currentGiftCard.status)">{{ giftCardStatusName(currentGiftCard.status) }}</AdminStatusBadge>
        </DetailItem>
        <DetailItem label="初始金额">{{ formatCurrency(currentGiftCard.initial_value, currentGiftCard.currency) }}</DetailItem>
        <DetailItem label="当前余额"><strong>{{ formatCurrency(currentGiftCard.balance, currentGiftCard.currency) }}</strong></DetailItem>
        <DetailItem label="收件人">{{ currentGiftCard.recipient_name || '-' }}</DetailItem>
        <DetailItem label="收件邮箱">{{ currentGiftCard.recipient_email || '-' }}</DetailItem>
        <DetailItem label="发送人">{{ currentGiftCard.sender_name || '-' }}</DetailItem>
        <DetailItem label="到期时间">{{ formatDate(currentGiftCard.expires_at) }}</DetailItem>
        <DetailItem label="创建时间">{{ formatDate(currentGiftCard.created_at) }}</DetailItem>
        <DetailItem label="更新时间">{{ formatDate(currentGiftCard.updated_at) }}</DetailItem>
      </dl>

      <div v-if="currentGiftCard.message" class="space-y-1.5">
        <h3 class="text-sm font-black tracking-tighter uppercase">祝福语</h3>
        <p class="whitespace-pre-wrap rounded-lg border bg-muted/30 p-3 text-sm leading-6">{{ currentGiftCard.message }}</p>
      </div>

      <div v-if="canEdit" class="flex flex-col gap-2 border-t pt-5 sm:flex-row sm:items-end">
        <label class="w-full space-y-1.5 sm:w-52">
          <span class="text-xs font-medium">更新状态</span>
          <Select v-model="statusUpdateModel">
            <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem v-for="option in giftCardStatusOptions(currentGiftCard)" :key="option.value" :value="option.value">
                {{ option.label }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>
        <Button :disabled="statusSubmitting || statusUpdateModel === currentGiftCard.status" @click="emit('update-status')">
          <LoaderCircle v-if="statusSubmitting" class="size-4 animate-spin" />
          更新状态
        </Button>
      </div>

      <section class="space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-black tracking-tighter uppercase">交易记录</h3>
          <span class="text-xs text-muted-foreground">{{ transactions.length }} 条</span>
        </div>
        <div class="overflow-x-auto rounded-lg border">
          <Table class="min-w-[620px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-24">类型</TableHead>
                <TableHead class="w-32 text-right">金额</TableHead>
                <TableHead class="w-32 text-right">交易后余额</TableHead>
                <TableHead>备注</TableHead>
                <TableHead class="w-44">时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="transactions.length === 0" :colspan="5">暂无交易记录</TableEmpty>
              <TableRow v-for="transaction in transactions" :key="transaction.id">
                <TableCell>{{ transactionTypeName(transaction.type) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatCurrency(transaction.amount, currentGiftCard.currency) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatCurrency(transaction.balance, currentGiftCard.currency) }}</TableCell>
 <TableCell class="text-muted-foreground">{{ transaction.note || '-'}}</TableCell>
                <TableCell class="text-xs text-muted-foreground">{{ formatDate(transaction.created_at) }}</TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
      </section>
    </div>
  </DialogContent>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const DetailItem = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots }) {
 return () => h('div', { class: 'border-b p-3 last:border-b-0 sm:border-b sm:border-r sm:nth-[3n]:border-r-0'}, [
 h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block'}, props.label),
 h('dd', { class: 'mt-1 break-words text-xs font-bold'}, slots.default?.()),
    ])
  },
})

export interface GiftCardRecord {
  id?: string | number
  code?: string
  status?: string
  initial_value?: number | string
  balance?: number | string
  currency?: string
  recipient_name?: string
  recipient_email?: string
  sender_name?: string
  expires_at?: string
  created_at?: string
  updated_at?: string
  message?: string
}

export interface GiftCardTransaction {
  id: string | number
  type?: string
  amount?: number | string
  balance?: number | string
  note?: string
  created_at?: string
}

export interface GiftCardStatusOption {
  value: string
  label: string
}

const props = withDefaults(defineProps<{
  currentGiftCard?: GiftCardRecord | null
  loading?: boolean
  transactions?: GiftCardTransaction[]
  statusUpdate?: string
  statusSubmitting?: boolean
  canEdit?: boolean
  formatCurrency: (value: unknown, currency?: string) => string
  formatDate: (value: unknown) => string
  giftCardStatusName: (status?: string) => string
  giftCardStatusTone: (status?: string) => AdminStatusTone
  giftCardStatusOptions: (giftCard: GiftCardRecord) => GiftCardStatusOption[]
  transactionTypeName: (type?: string) => string
}>(), {
  currentGiftCard: null,
  loading: false,
  transactions: () => [],
  statusUpdate: 'active',
  statusSubmitting: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:statusUpdate', value: string): void
  (event: 'update-status'): void
}>()

const statusUpdateModel = computed<string>({
  get: () => props.statusUpdate,
  set: (value: string) => emit('update:statusUpdate', value),
})
</script>
