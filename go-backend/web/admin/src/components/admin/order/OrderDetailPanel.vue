<template>
  <DialogContent size="xl" class="max-h-[90dvh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle>订单详情</DialogTitle>
      <DialogDescription>{{ currentOrder?.order_number || '订单信息' }}</DialogDescription>
    </DialogHeader>

    <div v-if="currentOrder" class="space-y-6">
      <OrderDetailSection title="订单信息">
        <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-2">
          <DetailItem label="订单号">{{ currentOrder.order_number }}</DetailItem>
          <DetailItem label="订单状态"><AdminStatusBadge :tone="orderStatusTone(currentOrder.status)">{{ orderStatusName(currentOrder.status) }}</AdminStatusBadge></DetailItem>
          <DetailItem label="支付状态"><AdminStatusBadge :tone="paymentStatusTone(currentOrder.payment_status)">{{ paymentStatusName(currentOrder.payment_status) }}</AdminStatusBadge></DetailItem>
          <DetailItem label="物流状态"><AdminStatusBadge :tone="shippingStatusTone(currentOrder.shipping_status)">{{ shippingStatusName(currentOrder.shipping_status) }}</AdminStatusBadge></DetailItem>
          <DetailItem label="支付方式">{{ currentOrder.payment_method || '-' }}</DetailItem>
          <DetailItem label="物流方式">{{ currentOrder.shipping_method || '-' }}</DetailItem>
          <DetailItem label="物流单号">{{ currentOrder.tracking_number || '-' }}</DetailItem>
          <DetailItem label="本地承运商">{{ orderCarrierLabel(currentOrder) }}</DetailItem>
          <DetailItem label="线路服务">{{ orderCarrierServiceLabel(currentOrder) }}</DetailItem>
          <DetailItem label="Provider Code">{{ currentOrder.provider_carrier_code || '-' }}</DetailItem>
          <DetailItem label="创建时间">{{ formatDate(currentOrder.created_at) }}</DetailItem>
          <DetailItem label="支付时间">{{ currentOrder.paid_at ? formatDate(currentOrder.paid_at) : '-' }}</DetailItem>
        </dl>
        <div v-if="currentOrder.tracking_number" class="rounded-xl border bg-muted/30 p-3">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TRACKING SYNC / 轨迹同步</p>
              <p class="mt-1 text-xs text-muted-foreground">来自订单发货信息的追踪状态记录，后续自动轮询和 webhook 都会围绕这里更新。</p>
            </div>
            <AdminStatusBadge :tone="trackingSyncStatusTone(currentTrackingShipment?.sync_status)">
              {{ trackingSyncStatusName(currentTrackingShipment?.sync_status) }}
            </AdminStatusBadge>
          </div>
          <dl class="mt-3 grid gap-2 text-xs sm:grid-cols-4">
            <div class="rounded-lg bg-background/80 p-2">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">登记状态</dt>
              <dd class="mt-1 font-bold">{{ trackingRegistrationStatusName(currentTrackingShipment?.registration_status) }}</dd>
            </div>
            <div class="rounded-lg bg-background/80 p-2">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">事件数量</dt>
              <dd class="mt-1 font-mono font-bold">{{ currentTrackingShipment?.event_count ?? currentTrackingEvents.length }}</dd>
            </div>
            <div class="rounded-lg bg-background/80 p-2">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">最后同步</dt>
              <dd class="mt-1 font-mono text-[10px] font-bold">{{ formatDate(currentTrackingShipment?.last_synced_at) }}</dd>
            </div>
            <div class="rounded-lg bg-background/80 p-2">
              <dt class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">下次自动同步</dt>
              <dd class="mt-1 font-mono text-[10px] font-bold">{{ formatDate(currentTrackingShipment?.next_sync_at) }}</dd>
            </div>
          </dl>
          <p v-if="currentTrackingShipment?.last_error" class="mt-2 rounded-lg border border-destructive/20 bg-destructive/10 px-3 py-2 text-xs font-medium text-destructive">
            {{ currentTrackingShipment.last_error }}
          </p>
        </div>
        <div v-if="currentOrder.tracking_number && canEdit" class="flex justify-end">
          <Button variant="outline" size="sm" class="rounded-full" :disabled="syncingTracking" @click="emit('sync-tracking')">
            <RefreshCw :class="['size-3.5', syncingTracking ? 'animate-spin' : '']" />
            {{ syncingTracking ? '同步中' : '同步轨迹' }}
          </Button>
        </div>
      </OrderDetailSection>

      <OrderDetailSection title="收货地址">
        <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-2">
          <DetailItem label="姓名">{{ shippingName(currentOrder.shipping_address) }}</DetailItem>
          <DetailItem label="电话">{{ currentOrder.shipping_address?.phone || '-' }}</DetailItem>
          <DetailItem label="邮箱" class="sm:col-span-2">{{ currentOrder.shipping_address?.email || '-' }}</DetailItem>
          <DetailItem label="地址" class="sm:col-span-2">{{ shippingAddressLine(currentOrder.shipping_address) }}</DetailItem>
          <DetailItem label="城市">{{ currentOrder.shipping_address?.city || '-' }}</DetailItem>
          <DetailItem label="省/州">{{ currentOrder.shipping_address?.state || '-' }}</DetailItem>
          <DetailItem label="邮编">{{ currentOrder.shipping_address?.postal_code || '-' }}</DetailItem>
          <DetailItem label="国家">{{ currentOrder.shipping_address?.country || '-' }}</DetailItem>
        </dl>
      </OrderDetailSection>

      <OrderDetailSection title="订单商品">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>商品名称</TableHead>
              <TableHead>SKU</TableHead>
              <TableHead class="text-right">单价</TableHead>
              <TableHead class="text-right">数量</TableHead>
              <TableHead class="text-right">小计</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="!currentOrder.items?.length" :colspan="5">暂无商品明细</TableEmpty>
            <TableRow v-for="item in currentOrder.items || []" :key="item.id || item.sku">
              <TableCell class="font-medium">{{ item.product_name }}</TableCell>
              <TableCell class="font-mono text-xs">{{ item.sku }}</TableCell>
              <TableCell class="text-right tabular-nums">¥{{ formatMoney(item.price) }}</TableCell>
              <TableCell class="text-right tabular-nums">{{ item.quantity }}</TableCell>
              <TableCell class="text-right font-medium tabular-nums">¥{{ formatMoney(item.total) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </OrderDetailSection>

      <OrderDetailSection title="金额明细">
        <dl class="ml-auto max-w-md space-y-2 text-sm">
          <AmountRow label="商品小计" :value="currentOrder.subtotal_amount" />
          <AmountRow label="运费" :value="currentOrder.shipping_fee" />
          <AmountRow label="税费" :value="currentOrder.tax_amount" />
          <AmountRow label="优惠" :value="-Number(currentOrder.discount_amount || 0)" />
          <div class="flex items-center justify-between border-t border-dashed pt-3 text-base font-black italic uppercase">
            <dt>订单总额</dt>
            <dd class="tabular-nums text-primary">¥{{ formatMoney(currentOrder.total_amount) }}</dd>
          </div>
        </dl>
      </OrderDetailSection>

      <OrderDetailSection title="物流轨迹">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead class="w-44">时间</TableHead>
              <TableHead class="w-32">状态</TableHead>
              <TableHead class="w-40">位置</TableHead>
              <TableHead>描述</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="currentTrackingEvents.length === 0" :colspan="4">暂无物流轨迹</TableEmpty>
            <TableRow v-for="event in currentTrackingEvents" :key="event.id || `${event.tracking_number}-${event.event_time}-${event.status}`">
              <TableCell class="font-mono text-[10px] text-muted-foreground">{{ formatDate(event.event_time) }}</TableCell>
              <TableCell><AdminStatusBadge tone="blue">{{ event.status || '-' }}</AdminStatusBadge></TableCell>
              <TableCell class="text-xs">{{ event.location || '-' }}</TableCell>
              <TableCell class="text-xs">{{ event.description || '-' }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </OrderDetailSection>

      <OrderDetailSection title="备注">
        <div class="space-y-4">
          <div>
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">NOTE / 客户备注</span>
            <p class="mt-1 text-sm">{{ currentOrder.customer_note || '-' }}</p>
          </div>
          <div>
            <Label for="admin-note">管理员备注</Label>
            <Textarea id="admin-note" v-model="adminNoteModel" class="mt-2 min-h-24" placeholder="请输入管理员备注" />
            <Button
              v-if="canEdit"
              size="sm"
              class="mt-2 rounded-full"
              @click="emit('update-note')"
            >
              保存备注
            </Button>
          </div>
        </div>
      </OrderDetailSection>
    </div>
  </DialogContent>
</template>

<script setup>
import { computed, defineComponent, h } from 'vue'
import { RefreshCw } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import {
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'

const OrderDetailSection = defineComponent({
  props: { title: { type: String, required: true } },
  setup(props, { slots }) {
    return () => h('section', { class: 'space-y-3 border-t border-dashed pt-5 first:border-t-0 first:pt-0' }, [
      h('h3', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
      slots.default?.(),
    ])
  },
})

const DetailItem = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots, attrs }) {
    return () => h('div', { ...attrs, class: ['border-b p-3 last:border-b-0 sm:border-r sm:last:border-r-0', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h('dd', { class: 'mt-1 text-xs font-bold' }, slots.default?.()),
    ])
  },
})

const AmountRow = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: 0 },
  },
  setup(props) {
    return () => h('div', { class: 'flex items-center justify-between' }, [
      h('dt', { class: 'text-muted-foreground' }, props.label),
      h('dd', { class: 'tabular-nums' }, `¥${Number(props.value || 0).toFixed(2)}`),
    ])
  },
})

const props = defineProps({
  currentOrder: { type: Object, default: null },
  currentTrackingEvents: { type: Array, default: () => [] },
  currentTrackingShipment: { type: Object, default: null },
  adminNote: { type: String, default: '' },
  syncingTracking: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  orderStatusName: { type: Function, required: true },
  orderStatusTone: { type: Function, required: true },
  paymentStatusName: { type: Function, required: true },
  paymentStatusTone: { type: Function, required: true },
  shippingStatusName: { type: Function, required: true },
  shippingStatusTone: { type: Function, required: true },
  trackingSyncStatusName: { type: Function, required: true },
  trackingSyncStatusTone: { type: Function, required: true },
  trackingRegistrationStatusName: { type: Function, required: true },
  formatDate: { type: Function, required: true },
  formatMoney: { type: Function, required: true },
  shippingName: { type: Function, required: true },
  shippingAddressLine: { type: Function, required: true },
  orderCarrierLabel: { type: Function, required: true },
  orderCarrierServiceLabel: { type: Function, required: true },
})

const emit = defineEmits(['update:adminNote', 'sync-tracking', 'update-note'])

const adminNoteModel = computed({
  get: () => props.adminNote,
  set: (value) => emit('update:adminNote', value),
})
</script>
