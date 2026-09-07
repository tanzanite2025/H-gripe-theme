<template>
  <AdminTablePanel :loading="loading">
    <Table class="min-w-[1120px]">
      <TableHeader>
        <TableRow>
          <TableHead>订单 / 客户</TableHead>
          <TableHead>订单金额</TableHead>
          <TableHead>证据包</TableHead>
          <TableHead>触发条件</TableHead>
          <TableHead>完整度</TableHead>
          <TableHead class="text-right">查看</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="orders.length === 0" :colspan="6">
          <div class="flex flex-col items-center text-muted-foreground">
            <FileCheck2 class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无订单</span>
          </div>
        </TableEmpty>

        <TableRow
          v-for="order in orders"
          :key="String(order.order_id)"
          class="cursor-pointer"
          :class="selectedOrderId === order.order_id ? 'bg-admin-selected-soft' : ''"
          @click="$emit('select-order', order)"
        >
          <TableCell>
            <span class="block font-mono text-xs font-black">{{ order.order_number || `订单 #${order.order_id}` }}</span>
            <span class="block max-w-64 truncate text-xs font-bold">{{ customerName(order) }}</span>
            <span class="block max-w-64 truncate text-[10px] text-muted-foreground/70">{{ order.customer_email || '-' }}</span>
          </TableCell>
          <TableCell>
            <span class="block text-xs font-black">{{ formatMoney(order.total_amount, order.currency) }}</span>
            <span class="block text-[10px] text-muted-foreground/70">
              USD 快照 {{ formatMoney(order.order_total_usd_snapshot, 'USD') }}
            </span>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="orderEvidencePackageStatusTone(order.package_status)">
              {{ orderEvidencePackageStatusName(order.package_status) }}
            </AdminStatusBadge>
            <span v-if="order.package_version" class="mt-1 block text-[10px] text-muted-foreground/70">
              v{{ order.package_version }}
            </span>
          </TableCell>
          <TableCell>
            <div class="flex flex-wrap gap-1">
              <span v-if="order.is_high_value" class="trigger-mark trigger-mark-amber">订单 ≥ 750 USD</span>
              <span v-if="order.has_spoke_tension_qc" class="trigger-mark trigger-mark-blue">需要张力表</span>
              <span v-if="!order.is_high_value && !order.has_spoke_tension_qc" class="text-[10px] text-muted-foreground/70">
                基础订单证据
              </span>
            </div>
          </TableCell>
          <TableCell>
            <div class="min-w-36">
              <div class="flex items-center justify-between gap-2 text-[10px] font-mono font-bold">
                <span>{{ evidencePercent(order) }}%</span>
                <span class="text-muted-foreground/70">
                  {{ satisfiedCount(order) }}/{{ order.total_evidence_items || 0 }}
                </span>
              </div>
              <div class="mt-1 h-1.5 overflow-hidden rounded-full bg-muted">
                <div
                  class="h-full rounded-full bg-emerald-500 transition-[width]"
                  :class="order.package_id && evidencePercent(order) < 100 ? 'bg-amber-500' : ''"
                  :style="{ width: `${evidencePercent(order)}%` }"
                />
              </div>
              <span v-if="order.pending_evidence_items" class="mt-1 block text-[10px] text-rose-600">
                待处理 {{ order.pending_evidence_items }} 项
              </span>
            </div>
          </TableCell>
          <TableCell class="text-right">
            <Button
              variant="outline"
              size="icon"
              class="size-8 rounded-full"
              :aria-label="`查看订单 ${order.order_number || order.order_id} 的证据`"
              title="查看履约证据"
              @click.stop="$emit('select-order', order)"
            >
              <Eye class="size-3.5" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="$emit('update-page', $event)"
        @update:page-size="$emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { Eye, FileCheck2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { formatMoney } from '@/lib/orderPresentation'
import {
  orderEvidenceCustomerName,
  orderEvidencePackageStatusName,
  orderEvidencePackageStatusTone,
  orderEvidencePercent,
} from '@/lib/orderEvidencePresentation'
import type { OrderEvidenceOrderSummary } from '@/modules/order/orderEvidenceTypes'

withDefaults(defineProps<{
  orders?: OrderEvidenceOrderSummary[]
  loading?: boolean
  selectedOrderId?: number | string | null
  pagination: {
    page: number
    pageSize: number
    total: number
  }
}>(), {
  orders: () => [],
  loading: false,
  selectedOrderId: null,
})

defineEmits<{
  (event: 'select-order', order: OrderEvidenceOrderSummary): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const customerName = orderEvidenceCustomerName
const evidencePercent = orderEvidencePercent
const satisfiedCount = (order: OrderEvidenceOrderSummary): number => (
  Number(order.complete_evidence_items || 0) + Number(order.waived_evidence_items || 0)
)
</script>

<style scoped>
.trigger-mark {
  display: inline-flex;
  align-items: center;
  min-height: 1.25rem;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 10px;
  font-weight: 700;
  line-height: 1;
}

.trigger-mark-amber {
  background: rgb(245 158 11 / 0.12);
  color: rgb(180 83 9);
}

.trigger-mark-blue {
  background: rgb(59 130 246 / 0.12);
  color: rgb(37 99 235);
}
</style>
