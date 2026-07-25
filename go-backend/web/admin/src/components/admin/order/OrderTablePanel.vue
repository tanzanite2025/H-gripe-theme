<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedOrders.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedOrders.length }} 个订单</span>
        <div class="flex flex-wrap gap-2">
          <Button v-if="canEdit" size="sm" @click="emit('batch-status', 'completed')">
            <CircleCheck class="size-3.5" />
            批量完成
          </Button>
          <Button v-if="canEdit" variant="outline" size="sm" @click="emit('batch-status', 'cancelled')">
            <CircleX class="size-3.5" />
            批量取消
          </Button>
        </div>
      </div>
    </template>

    <Table class="min-w-[1180px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页订单"
              @update:model-value="emit('toggle-all-orders', $event)"
            />
          </TableHead>
          <TableHead class="w-16">ID</TableHead>
          <TableHead class="w-44">订单号</TableHead>
          <TableHead>客户</TableHead>
          <TableHead class="w-24">订单状态</TableHead>
          <TableHead class="w-24">支付状态</TableHead>
          <TableHead class="w-24">物流状态</TableHead>
          <TableHead class="w-28 text-right">总金额</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="orders.length === 0" :colspan="10">
          <div class="flex flex-col items-center text-muted-foreground">
            <ShoppingBag class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无订单</span>
          </div>
        </TableEmpty>

        <TableRow v-for="order in orders" :key="order.id">
          <TableCell>
            <Checkbox
              :model-value="isOrderSelected(order.id)"
              :aria-label="`选择订单 ${order.order_number}`"
              @update:model-value="emit('toggle-order', order, $event)"
            />
          </TableCell>
          <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ order.id }}</TableCell>
          <TableCell class="font-mono text-xs font-bold">{{ order.order_number }}</TableCell>
          <TableCell class="font-bold text-xs">{{ shippingName(order.shipping_address) }}</TableCell>
          <TableCell>
            <AdminStatusBadge :tone="orderStatusTone(order.status)">{{ orderStatusName(order.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="paymentStatusTone(order.payment_status)">{{ paymentStatusName(order.payment_status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="shippingStatusTone(order.shipping_status)">{{ shippingStatusName(order.shipping_status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell class="text-right font-mono text-xs font-bold tabular-nums">¥{{ formatMoney(order.total_amount) }}</TableCell>
          <TableCell class="font-mono text-[10px] text-muted-foreground/80">{{ formatDate(order.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理订单 ${order.order_number}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem @select="emit('view-detail', order)">
                  <Eye class="size-4" />
                  查看详情
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('show-status', order)">
                  <RefreshCw class="size-4" />
                  状态管理
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', order)"
                >
                  <Trash2 class="size-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
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

<script setup>
import { CircleCheck, CircleX, Eye, MoreHorizontal, RefreshCw, ShoppingBag, Trash2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const props = defineProps({
  loading: { type: Boolean, default: false },
  orders: { type: Array, default: () => [] },
  selectedOrders: { type: Array, default: () => [] },
  pagination: { type: Object, required: true },
  selectionState: { type: [Boolean, String], default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
  orderStatusName: { type: Function, required: true },
  orderStatusTone: { type: Function, required: true },
  paymentStatusName: { type: Function, required: true },
  paymentStatusTone: { type: Function, required: true },
  shippingStatusName: { type: Function, required: true },
  shippingStatusTone: { type: Function, required: true },
  shippingName: { type: Function, required: true },
  formatMoney: { type: Function, required: true },
  formatDate: { type: Function, required: true },
})

const emit = defineEmits([
  'batch-status',
  'toggle-all-orders',
  'toggle-order',
  'view-detail',
  'show-status',
  'delete',
  'update-page',
  'update-page-size',
])

const isOrderSelected = (orderId) => props.selectedOrders.some((order) => order.id === orderId)
</script>
