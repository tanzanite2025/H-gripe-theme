<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>选择订单</DialogTitle>
        <DialogDescription>从当前客户的最近订单中选择，发送为订单卡片。</DialogDescription>
      </DialogHeader>

      <div class="flex min-h-[24rem] max-h-[72dvh] flex-col overflow-hidden">
        <div class="flex shrink-0 items-center justify-between gap-3 border-b px-4 py-3">
          <div class="min-w-0">
            <p class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">最近订单</p>
            <p class="mt-1 text-xs text-muted-foreground">只显示当前会话可见的订单。</p>
          </div>
          <AdminStatusBadge tone="gray">
            {{ orders.length }} 单
          </AdminStatusBadge>
        </div>

        <div v-if="!orders.length" class="flex min-h-0 flex-1 flex-col items-center justify-center gap-2 text-muted-foreground">
          <ShoppingCart class="size-8 opacity-50" />
          <span class="text-xs font-bold">暂无可发送订单</span>
        </div>

        <div v-else class="min-h-0 flex-1 overflow-y-auto p-4">
          <div class="grid gap-2">
            <button
              v-for="order in orders"
              :key="String(order.id)"
              type="button"
              class="flex min-w-0 items-start justify-between gap-3 rounded-xl border bg-card px-3 py-2.5 text-left transition hover:-translate-y-0.5 hover:border-admin-selected-border hover:shadow-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              @click="select(order)"
            >
              <span class="min-w-0 flex-1">
                <span class="block truncate text-xs font-black text-foreground">
                  订单 #{{ order.order_number || order.id }}
                </span>
                <span class="mt-1 block text-[11px] leading-5 text-muted-foreground">
                  {{ order.status || 'order' }} / {{ order.payment_status || 'payment' }} / {{ order.shipping_status || 'shipping' }}
                </span>
                <span class="mt-1 block text-[11px] text-muted-foreground">
                  {{ formatShortDate(order.created_at) }}
                </span>
              </span>
              <span class="shrink-0 text-right text-xs font-black text-foreground">
                {{ formatMoney(order.total_amount) }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { ShoppingCart } from '@lucide/vue'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { formatMoney, formatShortDate } from '@/lib/customerServicePresentation'
import type { CustomerOrderItem } from './customerServiceTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  orders?: CustomerOrderItem[]
}>(), {
  open: false,
  orders: () => [],
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', order: CustomerOrderItem): void
}>()

const select = (order: CustomerOrderItem): void => {
  emit('select', order)
}
</script>
