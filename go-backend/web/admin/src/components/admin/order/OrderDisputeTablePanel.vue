<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[1260px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-24">渠道</TableHead>
          <TableHead class="w-56">拒付记录</TableHead>
          <TableHead class="w-44">订单</TableHead>
          <TableHead>客户</TableHead>
          <TableHead class="w-36">状态</TableHead>
          <TableHead class="w-32 text-right">金额</TableHead>
          <TableHead class="w-40">证据完整度</TableHead>
          <TableHead class="w-36">误操作判断</TableHead>
          <TableHead class="w-44">截止/提交</TableHead>
          <TableHead class="w-36 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="disputes.length === 0" :colspan="10">
          <div class="flex flex-col items-center text-muted-foreground">
            <ShieldCheck class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无拒付订单</span>
          </div>
        </TableEmpty>

        <TableRow v-for="dispute in disputes" :key="`${dispute.provider}-${dispute.dispute_id}`">
          <TableCell>
            <AdminStatusBadge :tone="providerTone(dispute.provider)">{{ providerLabel(dispute.provider) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
 <div class="max-w-[220px] truncate font-mono text-xs font-bold">{{ dispute.provider_dispute_id || '-'}}</div>
 <div class="mt-1 max-w-[220px] truncate text-[10px] text-muted-foreground">{{ dispute.reason || '-'}}</div>
          </TableCell>
          <TableCell>
            <button
              type="button"
              class="font-mono text-xs font-bold text-primary hover:underline disabled:text-muted-foreground disabled:no-underline"
              :disabled="!dispute.order_id"
              @click="emit('view-order', dispute)"
            >
              {{ dispute.order_number || orderIDLabel(dispute.order_id) }}
            </button>
            <div class="mt-1 text-[10px] text-muted-foreground">
              {{ dispute.order_status || '-' }} / {{ dispute.shipping_status || '-' }}
            </div>
          </TableCell>
          <TableCell>
 <div class="text-xs font-bold">{{ dispute.customer_name || '-'}}</div>
 <div class="mt-1 max-w-[220px] truncate font-mono text-[10px] text-muted-foreground">{{ dispute.customer_email || '-'}}</div>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(dispute)">{{ dispute.status || '-' }}</AdminStatusBadge>
            <div v-if="dispute.needs_response" class="mt-1 text-[10px] font-bold text-rose-600">需响应</div>
          </TableCell>
          <TableCell class="text-right font-mono text-xs font-bold tabular-nums">
            {{ formatMoney(dispute.amount, dispute.currency) }}
          </TableCell>
          <TableCell>
            <div class="text-xs font-black">
              {{ dispute.evidence_summary?.ready_count || 0 }}/{{ dispute.evidence_summary?.total_count || 0 }}
            </div>
            <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-muted">
              <div
                class="h-full rounded-full"
                :class="evidenceBarClass(dispute)"
                :style="{ width: evidencePercent(dispute) }"
              />
            </div>
            <div v-if="dispute.evidence_summary?.blocker_count" class="mt-1 text-[10px] font-bold text-rose-600">
              阻断 {{ dispute.evidence_summary.blocker_count }}
            </div>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="assessmentTone(dispute.mistake_assessment?.level)">
              {{ dispute.mistake_assessment?.label || '待判断' }}
            </AdminStatusBadge>
          </TableCell>
          <TableCell class="text-xs" :class="deadlineClass(dispute)">
            {{ deadlineLabel(dispute) }}
          </TableCell>
          <TableCell>
            <div class="flex justify-end gap-1.5">
              <Button
                variant="ghost"
                size="icon"
                :disabled="!dispute.order_id"
                :aria-label="`查看订单 ${dispute.order_number || dispute.order_id || ''}`"
                @click="emit('view-order', dispute)"
              >
                <Eye class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                :disabled="!canContact(dispute)"
                :aria-label="`联系客户 ${dispute.customer_email || ''}`"
                @click="emit('contact-customer', dispute)"
              >
                <Mail class="size-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                aria-label="打开支付拒付工作台"
                @click="emit('open-payment-workbench', dispute)"
              >
                <CreditCard class="size-4" />
              </Button>
            </div>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { CreditCard, Eye, Mail, ShieldCheck } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type {
  OrderDateFormatter,
  OrderDisputeCase,
  OrderID,
  OrderMoneyFormatter,
  OrderPagination,
  OrderStatusTone,
} from './orderTypes'

const props = withDefaults(defineProps<{
  loading?: boolean
  disputes?: OrderDisputeCase[]
  pagination: OrderPagination
  formatMoney: OrderMoneyFormatter
  formatDate: OrderDateFormatter
}>(), {
  loading: false,
  disputes: () => []
})

const emit = defineEmits<{
  (event: 'view-order', dispute: OrderDisputeCase): void
  (event: 'contact-customer', dispute: OrderDisputeCase): void
  (event: 'open-payment-workbench', dispute: OrderDisputeCase): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const providerLabel = (provider?: string | null): string => provider === 'paypal' ? 'PayPal' : 'Stripe'
const providerTone = (provider?: string | null): OrderStatusTone => provider === 'paypal' ? 'blue' : 'gray'
const orderIDLabel = (orderID?: OrderID | null): string => orderID ? `#${orderID}` : '未关联'
const canContact = (dispute: OrderDisputeCase): boolean => Boolean(dispute.order_id && dispute.contact_draft?.can_send)

const statusTone = (dispute: OrderDisputeCase): OrderStatusTone => {
  const status = String(dispute.status || '').toLowerCase()
  if (dispute.needs_response) return 'coral'
  if (['won', 'resolved', 'closed'].includes(status)) return 'green'
  if (['lost'].includes(status)) return 'amber'
  if (status.includes('review')) return 'blue'
  return 'gray'
}

const assessmentTone = (level?: string | null): OrderStatusTone => {
  if (level === 'likely_mistake') return 'amber'
  if (level === 'resolved') return 'green'
  if (level === 'evidence_gap' || level === 'no_email' || level === 'unlinked_order') return 'coral'
  return 'gray'
}

const evidencePercent = (dispute: OrderDisputeCase): string => {
  const ready = Number(dispute.evidence_summary?.ready_count || 0)
  const total = Number(dispute.evidence_summary?.total_count || 0)
  return `${total > 0 ? Math.max(6, Math.round((ready / total) * 100)) : 0}%`
}

const evidenceBarClass = (dispute: OrderDisputeCase): string => {
  if (dispute.evidence_summary?.blocker_count) return 'bg-rose-500'
  if (dispute.evidence_summary?.complete) return 'bg-emerald-500'
  return 'bg-amber-500'
}

const isSoon = (value?: string | null): boolean => {
  if (!value) return false
  const time = new Date(value).getTime()
  return Number.isFinite(time) && time - Date.now() < 3 * 24 * 60 * 60 * 1000
}

const deadlineClass = (dispute: OrderDisputeCase): string => {
  if (dispute.evidence_submitted_at) return 'text-emerald-700 font-semibold'
  if (isSoon(dispute.evidence_due_at)) return 'text-rose-600 font-semibold'
  return 'text-muted-foreground'
}

const deadlineLabel = (dispute: OrderDisputeCase): string => {
  if (dispute.evidence_submitted_at) return `已提交 ${props.formatDate(dispute.evidence_submitted_at)}`
  if (dispute.evidence_due_at) return props.formatDate(dispute.evidence_due_at)
  return '-'
}
</script>
