<template>
  <DialogContent size="xl" class="max-h-[90dvh] overflow-y-auto">
    <DialogHeader>
      <DialogTitle>订单详情</DialogTitle>
      <DialogDescription>{{ currentOrder?.order_number || '订单信息' }}</DialogDescription>
    </DialogHeader>

    <Tabs v-if="currentOrder" default-value="overview" class="space-y-4">
      <TabsList class="h-auto flex-wrap justify-start rounded-xl">
        <TabsTrigger value="overview">概览</TabsTrigger>
        <TabsTrigger value="items">商品金额</TabsTrigger>
        <TabsTrigger value="tracking">物流轨迹</TabsTrigger>
        <TabsTrigger value="disputes">
          拒付分析
          <span v-if="disputeAnalysis?.summary?.total" class="ml-1 rounded-full bg-rose-500/10 px-1.5 py-0.5 text-[9px] text-rose-600">
            {{ disputeAnalysis.summary.total }}
          </span>
        </TabsTrigger>
        <TabsTrigger value="notes">备注</TabsTrigger>
      </TabsList>

      <TabsContent value="overview" class="space-y-6">
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
      </TabsContent>

      <TabsContent value="items" class="space-y-6">
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

      <OrderDetailSection title="清关资料">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <p class="text-xs text-muted-foreground">按订单商品逐行导出，申报价值以当前人工确认结果为准。</p>
          <Button
            variant="outline"
            size="sm"
            :disabled="exportingCustoms || !currentOrder.items?.length"
            @click="emit('export-customs')"
          >
            <Download :class="['size-3.5', exportingCustoms ? 'animate-pulse' : '']" />
            {{ exportingCustoms ? '生成中' : '下载清关资料' }}
          </Button>
        </div>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>商品</TableHead>
              <TableHead>HS Code</TableHead>
              <TableHead>CN Code</TableHead>
              <TableHead>原产国</TableHead>
              <TableHead>英文报关品名</TableHead>
              <TableHead class="text-right">最终申报价值</TableHead>
              <TableHead>状态</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="!currentOrder.items?.length" :colspan="7">暂无清关资料</TableEmpty>
            <TableRow v-for="item in currentOrder.items || []" :key="`customs-${item.id || item.sku}`">
              <TableCell class="max-w-40 truncate font-medium" :title="item.product_name || undefined">{{ item.product_name || '-' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ item.hs_code || '-' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ item.cn_code || '-' }}</TableCell>
              <TableCell class="font-mono text-xs">{{ item.country_of_origin || '-' }}</TableCell>
              <TableCell class="max-w-52 truncate text-xs" :title="item.customs_description || undefined">{{ item.customs_description || '-' }}</TableCell>
              <TableCell class="min-w-52">
                <div v-if="canEdit" class="flex items-center gap-2">
                  <Input
                    v-model="declaredValueDrafts[customsKey(item)]"
                    type="number"
                    min="0"
                    step="0.01"
                    class="w-32"
                    placeholder="未填写"
                    :disabled="isSavingCustomsItem(item)"
                  />
                  <Button
                    variant="outline"
                    size="icon"
                    class="size-8 shrink-0"
                    :disabled="!canSaveCustomsItem(item)"
                    title="保存申报价值"
                    :aria-label="`保存 ${item.product_name || '商品'} 的申报价值`"
                    @click="saveCustomsItem(item)"
                  >
                    <Save :class="['size-3.5', isSavingCustomsItem(item) ? 'animate-pulse' : '']" />
                  </Button>
                </div>
                <span v-else class="font-mono text-xs">{{ formatDeclaredValue(item.declared_value) }}</span>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-2">
                  <Checkbox
                    :model-value="declaredValueConfirmed(item)"
                    :disabled="!canEdit || isSavingCustomsItem(item)"
                    :aria-label="`确认 ${item.product_name || '商品'} 的申报价值`"
                    @update:model-value="setDeclaredValueConfirmed(item, $event)"
                  />
                  <AdminStatusBadge :tone="declaredValueConfirmed(item) ? 'green' : 'amber'">
                    {{ declaredValueConfirmed(item) ? '已确认' : '待确认' }}
                  </AdminStatusBadge>
                </div>
              </TableCell>
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
      </TabsContent>

      <TabsContent value="tracking" class="space-y-6">
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
      </TabsContent>

      <TabsContent value="disputes" class="space-y-4">
        <div v-if="disputeAnalysisLoading" class="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
          拒付分析加载中
        </div>
        <div v-else-if="!disputeAnalysis?.disputes?.length" class="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
          当前订单没有关联 Stripe / PayPal 拒付记录
        </div>
        <div v-else class="space-y-4">
          <div class="grid gap-2 sm:grid-cols-4">
            <div class="rounded-lg border bg-muted/30 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TOTAL / 拒付</div>
              <div class="mt-1 font-mono text-lg font-black">{{ disputeAnalysis.summary?.total || 0 }}</div>
            </div>
            <div class="rounded-lg border bg-muted/30 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">RESPONSE / 需响应</div>
              <div class="mt-1 font-mono text-lg font-black text-rose-600">{{ disputeAnalysis.summary?.needs_response || 0 }}</div>
            </div>
            <div class="rounded-lg border bg-muted/30 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">MISTAKE / 疑似误操作</div>
              <div class="mt-1 font-mono text-lg font-black text-amber-600">{{ disputeAnalysis.summary?.likely_mistake || 0 }}</div>
            </div>
            <div class="rounded-lg border bg-muted/30 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">BLOCKER / 证据阻断</div>
              <div class="mt-1 font-mono text-lg font-black">{{ disputeAnalysis.summary?.evidence_blocked || 0 }}</div>
            </div>
          </div>

          <div
            v-for="dispute in disputeAnalysis.disputes"
            :key="`${dispute.provider}-${dispute.dispute_id}`"
            class="rounded-lg border border-dashed p-4"
          >
            <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <AdminStatusBadge :tone="dispute.provider === 'paypal' ? 'blue' : 'gray'">{{ providerLabel(dispute.provider) }}</AdminStatusBadge>
                  <AdminStatusBadge :tone="disputeStatusTone(dispute)">{{ dispute.status || '-' }}</AdminStatusBadge>
                  <AdminStatusBadge :tone="assessmentTone(dispute.mistake_assessment?.level)">
                    {{ dispute.mistake_assessment?.label || '待判断' }}
                  </AdminStatusBadge>
                </div>
                <div class="mt-2 break-all font-mono text-xs font-bold">{{ dispute.provider_dispute_id || '-' }}</div>
                <p class="mt-1 text-xs text-muted-foreground">{{ dispute.suggested_action || dispute.mistake_assessment?.reason || '-' }}</p>
              </div>
              <div class="flex shrink-0 gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  class="rounded-full"
                  @click="emit('open-payment-workbench', dispute)"
                >
                  <CreditCard class="size-3.5" />
                  证据工作台
                </Button>
                <Button
                  size="sm"
                  class="rounded-full"
                  :disabled="!dispute.contact_draft?.can_send"
                  @click="emit('contact-dispute', dispute)"
                >
                  <Mail class="size-3.5" />
                  写邮件
                </Button>
              </div>
            </div>

            <dl class="mt-4 grid overflow-hidden rounded-lg border text-xs sm:grid-cols-3">
              <DetailItem label="金额">{{ disputeMoney(dispute.amount, dispute.currency) }}</DetailItem>
              <DetailItem label="原因">{{ dispute.reason || '-' }}</DetailItem>
              <DetailItem label="截止/提交">{{ disputeDeadline(dispute) }}</DetailItem>
              <DetailItem label="客户邮箱">{{ dispute.customer_email || '-' }}</DetailItem>
              <DetailItem label="物流单号">{{ dispute.tracking_number || '-' }}</DetailItem>
              <DetailItem label="证据完整度">
                {{ dispute.evidence_summary?.ready_count || 0 }}/{{ dispute.evidence_summary?.total_count || 0 }}
                <span v-if="dispute.evidence_summary?.blocker_count" class="text-rose-600">，阻断 {{ dispute.evidence_summary.blocker_count }}</span>
              </DetailItem>
            </dl>

            <div v-if="dispute.submission_blockers?.length" class="mt-3 rounded-lg border border-rose-500/20 bg-rose-500/5 p-3">
              <div class="text-[10px] font-black uppercase tracking-widest text-rose-700">BLOCKERS / 提交阻断</div>
              <ul class="mt-2 space-y-1 text-xs text-rose-700">
                <li v-for="blocker in dispute.submission_blockers" :key="blocker">{{ blocker }}</li>
              </ul>
            </div>
          </div>
        </div>
      </TabsContent>

      <TabsContent value="notes" class="space-y-6">
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
      </TabsContent>
    </Tabs>
  </DialogContent>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, watch } from 'vue'
import { CreditCard, Download, Mail, RefreshCw, Save } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import type {
  OrderCarrierLabelResolver,
  OrderDateFormatter,
  OrderDisputeAnalysis,
  OrderDisputeCase,
  OrderID,
  OrderItem,
  OrderMoneyFormatter,
  OrderRecord,
  OrderShippingAddressLineResolver,
  OrderShippingNameResolver,
  OrderStatusNameResolver,
  OrderStatusTone,
  OrderStatusToneResolver,
  TrackingEvent,
  TrackingShipment
} from './orderTypes'

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

const props = withDefaults(defineProps<{
  currentOrder?: OrderRecord | null
  currentTrackingEvents?: TrackingEvent[]
  currentTrackingShipment?: TrackingShipment | null
  disputeAnalysis?: OrderDisputeAnalysis | null
  disputeAnalysisLoading?: boolean
  adminNote?: string
  syncingTracking?: boolean
  savingCustomsItemId?: OrderID | null
  exportingCustoms?: boolean
  canEdit?: boolean
  orderStatusName: OrderStatusNameResolver
  orderStatusTone: OrderStatusToneResolver
  paymentStatusName: OrderStatusNameResolver
  paymentStatusTone: OrderStatusToneResolver
  shippingStatusName: OrderStatusNameResolver
  shippingStatusTone: OrderStatusToneResolver
  trackingSyncStatusName: OrderStatusNameResolver
  trackingSyncStatusTone: OrderStatusToneResolver
  trackingRegistrationStatusName: OrderStatusNameResolver
  formatDate: OrderDateFormatter
  formatMoney: OrderMoneyFormatter
  shippingName: OrderShippingNameResolver
  shippingAddressLine: OrderShippingAddressLineResolver
  orderCarrierLabel: OrderCarrierLabelResolver
  orderCarrierServiceLabel: OrderCarrierLabelResolver
}>(), {
  currentOrder: null,
  currentTrackingEvents: () => [],
  currentTrackingShipment: null,
  disputeAnalysis: null,
  disputeAnalysisLoading: false,
  adminNote: '',
  syncingTracking: false,
  savingCustomsItemId: null,
  exportingCustoms: false,
  canEdit: false
})

const emit = defineEmits<{
  (event: 'update:adminNote', value: string): void
  (event: 'sync-tracking'): void
  (event: 'update-note'): void
  (event: 'update-customs', orderItemId: OrderID, declaredValue: number | null, declaredValueConfirmed: boolean): void
  (event: 'export-customs'): void
  (event: 'contact-dispute', dispute: OrderDisputeCase): void
  (event: 'open-payment-workbench', dispute: OrderDisputeCase): void
}>()

const adminNoteModel = computed<string>({
  get: () => props.adminNote,
  set: (value: string) => emit('update:adminNote', value),
})

const declaredValueDrafts = reactive<Record<string, string>>({})
const declaredValueConfirmedDrafts = reactive<Record<string, boolean>>({})

const customsKey = (item: OrderItem): string => String(item.id ?? item.sku ?? item.product_name ?? '')

const initializeCustomsDrafts = (order?: OrderRecord | null): void => {
  Object.keys(declaredValueDrafts).forEach((key) => delete declaredValueDrafts[key])
  Object.keys(declaredValueConfirmedDrafts).forEach((key) => delete declaredValueConfirmedDrafts[key])
  for (const item of order?.items || []) {
    const key = customsKey(item)
    declaredValueDrafts[key] = item.declared_value == null ? '' : String(item.declared_value)
    declaredValueConfirmedDrafts[key] = Boolean(item.declared_value_confirmed)
  }
}

watch(() => props.currentOrder, initializeCustomsDrafts, { immediate: true })

const isSavingCustomsItem = (item: OrderItem): boolean => (
  props.savingCustomsItemId != null && String(props.savingCustomsItemId) === String(item.id)
)

const declaredValueConfirmed = (item: OrderItem): boolean => (
  declaredValueConfirmedDrafts[customsKey(item)] ?? Boolean(item.declared_value_confirmed)
)

const setDeclaredValueConfirmed = (item: OrderItem, value: boolean | 'indeterminate'): void => {
  declaredValueConfirmedDrafts[customsKey(item)] = value === true
}

const canSaveCustomsItem = (item: OrderItem): boolean => {
  if (!item.id || isSavingCustomsItem(item)) return false
  const raw = (declaredValueDrafts[customsKey(item)] || '').trim()
  if (raw === '') return !declaredValueConfirmed(item)
  const value = Number(raw)
  return Number.isFinite(value) && value >= 0
}

const saveCustomsItem = (item: OrderItem): void => {
  if (!canSaveCustomsItem(item)) return
  const raw = (declaredValueDrafts[customsKey(item)] || '').trim()
  const declaredValue = raw === '' ? null : Number(raw)
  emit('update-customs', item.id as OrderID, declaredValue, declaredValueConfirmed(item))
}

const providerLabel = (provider?: string | null): string => provider === 'paypal' ? 'PayPal' : 'Stripe'

const disputeStatusTone = (dispute: OrderDisputeCase): OrderStatusTone => {
  const status = String(dispute.status || '').toLowerCase()
  if (dispute.needs_response) return 'coral'
  if (['won', 'resolved', 'closed'].includes(status)) return 'green'
  if (status.includes('review')) return 'blue'
  if (status === 'lost') return 'amber'
  return 'gray'
}

const assessmentTone = (level?: string | null): OrderStatusTone => {
  if (level === 'likely_mistake') return 'amber'
  if (level === 'resolved') return 'green'
  if (level === 'evidence_gap' || level === 'no_email' || level === 'unlinked_order') return 'coral'
  return 'gray'
}

const disputeMoney = (amount?: number | string | null, currency?: string | null): string => props.formatMoney(amount, currency)

const formatDeclaredValue = (value?: number | string | null): string => {
  if (value == null || value === '') return '未填写'
  const amount = Number(value)
  return Number.isFinite(amount) ? amount.toFixed(2) : '未填写'
}

const disputeDeadline = (dispute: OrderDisputeCase): string => {
  if (dispute.evidence_submitted_at) return `已提交 ${props.formatDate(dispute.evidence_submitted_at)}`
  if (dispute.evidence_due_at) return props.formatDate(dispute.evidence_due_at)
  return '-'
}
</script>
