<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>售后单详情</DialogTitle>
        <DialogDescription>
          售后单 #{{ record?.id || '-' }} · {{ record?.order_number || `订单 #${record?.order_id || '-'}` }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="record" class="space-y-5">
        <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-muted/25 p-3">
          <div class="flex flex-wrap items-center gap-2">
            <span class="type-mark">{{ getTypeName(record.type) }}</span>
            <span class="status-mark" :class="statusClass(record.status)">
              {{ getStatusName(record.status) }}
            </span>
          </div>
          <RouterLink
            :to="{ name: 'OrdersList', query: record.order_number ? { search: record.order_number } : undefined }"
            class="font-mono text-xs font-bold text-primary hover:underline"
          >
            查看订单
          </RouterLink>
        </div>

        <dl class="grid overflow-hidden rounded-lg border text-sm sm:grid-cols-2">
          <DetailItem label="申请原因">{{ record.reason || '-' }}</DetailItem>
          <DetailItem label="创建时间">{{ formatDate(record.created_at) }}</DetailItem>
          <DetailItem label="补充说明" class="sm:col-span-2">{{ record.description || '-' }}</DetailItem>
          <DetailItem label="处理说明" class="sm:col-span-2">{{ record.resolution || '尚未填写' }}</DetailItem>
        </dl>

        <AfterSalesCaseAttachmentsPanel
          :case-id="record.id"
          :attachments="record.attachments"
          :loading="loading"
        />

        <AfterSalesRefundReviewPanel
          :record="record"
          :refund-submitting="refundSubmitting"
          @save-refund-review="forwardRefundReviewSave"
          @decide-refund-review="forwardRefundReviewDecision"
          @create-pending-refund="emit('create-pending-refund')"
        />

        <section class="space-y-3 border-t border-dashed pt-5">
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-sm font-black uppercase text-foreground">状态记录</h3>
            <span class="text-xs text-muted-foreground">{{ record.events?.length || 0 }} 条</span>
          </div>
          <div v-if="loading" class="rounded-lg border border-dashed p-5 text-center text-sm text-muted-foreground">
            状态记录加载中
          </div>
          <div v-else-if="!record.events?.length" class="rounded-lg border border-dashed p-5 text-center text-sm text-muted-foreground">
            暂无状态记录
          </div>
          <ol v-else class="space-y-4 border-l border-dashed border-border pl-4">
            <li v-for="event in record.events" :key="String(event.id || `${event.created_at}-${event.to_status}`)" class="relative">
              <span class="absolute -left-[1.31rem] top-1 size-2 rounded-full border border-background bg-primary" />
              <div class="flex flex-wrap items-center justify-between gap-2">
                <span class="text-xs font-black text-foreground">{{ eventLabel(event) }}</span>
                <time class="font-mono text-[10px] text-muted-foreground">{{ formatDate(event.created_at) }}</time>
              </div>
              <p class="mt-1 whitespace-pre-wrap text-xs text-muted-foreground">{{ event.resolution || '未填写处理说明' }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">操作人 {{ eventOperatorLabel(event) }}</p>
            </li>
          </ol>
        </section>

        <section class="space-y-3">
          <h3 class="text-sm font-black uppercase text-foreground">涉及商品</h3>
          <div class="overflow-x-auto rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>商品</TableHead>
                  <TableHead>SKU</TableHead>
                  <TableHead class="text-right">数量</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in record.items || []" :key="String(item.id || item.order_item_id)">
                  <TableCell class="font-medium">{{ item.product_name || '-' }}</TableCell>
                  <TableCell class="font-mono text-xs text-muted-foreground">{{ item.sku || '-' }}</TableCell>
                  <TableCell class="text-right font-mono">{{ item.quantity || 0 }}</TableCell>
                </TableRow>
                <TableEmpty v-if="!record.items?.length" :colspan="3">暂无商品明细</TableEmpty>
              </TableBody>
            </Table>
          </div>
        </section>

        <form v-if="availableStatuses.length" class="space-y-3 border-t border-dashed pt-5" @submit.prevent="submit">
          <h3 class="text-sm font-black uppercase text-foreground">处理售后单</h3>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block space-y-1.5">
              <span class="field-label">NEXT STATUS / 下一状态</span>
              <select v-model="form.status" class="field-select" :disabled="submitting">
                <option v-for="status in availableStatuses" :key="status" :value="status">
                  {{ getStatusName(status) }}
                </option>
              </select>
            </label>
            <div class="rounded-lg border border-dashed bg-muted/20 p-3 text-xs text-muted-foreground">
              当前状态：<span class="font-bold text-foreground">{{ getStatusName(record.status) }}</span>
              <p class="mt-1">本次状态流转必须填写新的处理说明。</p>
            </div>
          </div>
          <label class="block space-y-1.5">
            <span class="field-label">RESOLUTION / 处理说明</span>
            <Textarea
              v-model="form.resolution"
              rows="4"
              maxlength="2000"
              :disabled="submitting"
              placeholder="填写本次审核结论、异常原因或后续处理备注"
            />
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" :disabled="submitting" @click="emit('update:open', false)">
              关闭
            </Button>
            <Button type="submit" :disabled="submitting || !canSubmit">
              <Check :class="['size-4', submitting ? 'animate-pulse' : '']" />
              {{ submitting ? '保存中' : '保存状态' }}
            </Button>
          </DialogFooter>
        </form>

        <DialogFooter v-else>
          <Button type="button" variant="outline" @click="emit('update:open', false)">关闭</Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, watch } from 'vue'
import { Check } from '@lucide/vue'
import AfterSalesCaseAttachmentsPanel from '@/components/admin/order/AfterSalesCaseAttachmentsPanel.vue'
import AfterSalesRefundReviewPanel from '@/components/admin/order/AfterSalesRefundReviewPanel.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import type { AfterSalesCase } from '@/api/afterSales'
import {
  afterSalesStatusClass as statusClass,
  getAfterSalesNextStatuses as nextStatuses,
  getAfterSalesStatusName as getStatusName,
  getAfterSalesTypeName as getTypeName,
} from '@/lib/afterSalesPresentation'

const DetailItem = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots, attrs }) {
    return () => h('div', { ...attrs, class: ['border-b p-3 last:border-b-0 sm:border-r sm:last:border-r-0', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
      h('dd', { class: 'mt-1 whitespace-pre-wrap text-xs font-bold' }, slots.default?.()),
    ])
  },
})

const props = withDefaults(defineProps<{
  open?: boolean
  record?: AfterSalesCase | null
  loading?: boolean
  submitting?: boolean
  refundSubmitting?: boolean
}>(), {
  open: false,
  record: null,
  loading: false,
  submitting: false,
  refundSubmitting: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit', status: string, resolution: string): void
  (event: 'save-refund-review', proposedAmount: number, currency: string, requestNotes: string): void
  (event: 'decide-refund-review', status: 'approved' | 'rejected' | 'cancelled', decisionNotes: string): void
  (event: 'create-pending-refund'): void
}>()

const form = reactive({ status: '', resolution: '' })
const availableStatuses = computed(() => nextStatuses(props.record?.status))
const canSubmit = computed(() => Boolean(form.status && form.resolution.trim()))

const formatDate = (value?: string | null): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
const eventLabel = (event: { from_status?: string | null; to_status?: string | null }): string => (
  event.from_status
    ? `${getStatusName(event.from_status)} -> ${getStatusName(event.to_status)}`
    : `创建售后单 · ${getStatusName(event.to_status)}`
)
const eventOperatorLabel = (event: { operator_name?: string | null; updated_by?: string | number | null }): string => {
  if (event.operator_name) return event.operator_name
  const userID = Number(event.updated_by || 0)
  return userID > 0 ? `账号 #${userID}` : '系统'
}
const resetForm = (): void => {
  form.status = availableStatuses.value[0] || ''
  form.resolution = ''
}

const submit = (): void => {
  if (!canSubmit.value) return
  emit('submit', form.status, form.resolution.trim())
}

const forwardRefundReviewSave = (
  proposedAmount: number,
  currency: string,
  requestNotes: string,
): void => emit('save-refund-review', proposedAmount, currency, requestNotes)

const forwardRefundReviewDecision = (
  status: 'approved' | 'rejected' | 'cancelled',
  decisionNotes: string,
): void => emit('decide-refund-review', status, decisionNotes)

watch(
  () => [props.open, props.record?.id, props.record?.status, props.record?.resolution],
  () => {
    if (props.open) resetForm()
  },
  { immediate: true },
)
</script>

<style scoped>
.field-label {
  display: block;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.field-select {
  height: 2.5rem;
  width: 100%;
  border: 1px dashed hsl(var(--border) / 0.8);
  border-radius: 0.375rem;
  background: hsl(var(--background));
  padding: 0 0.75rem;
  color: hsl(var(--foreground));
  font-size: 0.875rem;
}

.type-mark,
.status-mark {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font-size: 0.6875rem;
  font-weight: 800;
  white-space: nowrap;
}

.type-mark {
  border: 1px dashed hsl(var(--border));
  color: hsl(var(--muted-foreground));
}

.status-gray { background: rgb(148 163 184 / 0.12); color: rgb(71 85 105); }
.status-green { background: rgb(16 185 129 / 0.12); color: rgb(4 120 87); }
.status-amber { background: rgb(245 158 11 / 0.14); color: rgb(180 83 9); }
.status-blue { background: rgb(59 130 246 / 0.12); color: rgb(29 78 216); }
.status-coral { background: rgb(244 63 94 / 0.12); color: rgb(190 24 93); }
</style>
