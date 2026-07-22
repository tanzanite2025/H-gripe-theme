<template>
  <div class="space-y-4">
    <AdminPageHeader title="订单管理" description="查看订单履约、支付和物流状态">
      <template #actions>
        <Button variant="outline" @click="exportOrders">
          <Download class="size-4" />
          导出订单
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(200px,1.2fr)_repeat(3,minmax(130px,0.7fr))_repeat(2,minmax(130px,0.7fr))_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="订单号、客户或 Email" />
          </div>
        </label>

        <FilterSelect v-model="filters.status" label="订单状态" :options="orderStatusOptions" />
        <FilterSelect v-model="filters.payment_status" label="支付状态" :options="paymentStatusOptions" />
        <FilterSelect v-model="filters.shipping_status" label="物流状态" :options="shippingStatusOptions" />

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">START DATE / 开始日期</span>
          <Input v-model="filters.start_date" type="date" class="h-9" />
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">END DATE / 结束日期</span>
          <Input v-model="filters.end_date" type="date" class="h-9" />
        </label>

        <label class="space-y-1 block">
          <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION / 操作</span>
          <div class="flex items-center gap-2">
            <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
              <Search class="size-3.5" />
              搜索
            </Button>
            <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
              <RotateCcw class="size-3.5" />
              重置
            </Button>
          </div>
        </label>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" :batch-visible="selectedOrders.length > 0">
      <template #batch>
        <div class="flex flex-wrap items-center justify-between gap-2">
          <span class="text-xs font-medium">已选择 {{ selectedOrders.length }} 个订单</span>
          <div class="flex flex-wrap gap-2">
            <Button
              v-if="hasPermission('order:edit')"
              size="sm"
              @click="requestBatchStatus('completed')"
            >
              <CircleCheck class="size-3.5" />
              批量完成
            </Button>
            <Button
              v-if="hasPermission('order:edit')"
              variant="outline"
              size="sm"
              @click="requestBatchStatus('cancelled')"
            >
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
                @update:model-value="toggleAllOrders"
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
                :model-value="isSelected(order.id)"
                :aria-label="`选择订单 ${order.order_number}`"
                @update:model-value="toggleOrder(order, $event)"
              />
            </TableCell>
            <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ order.id }}</TableCell>
            <TableCell class="font-mono text-xs font-bold">{{ order.order_number }}</TableCell>
            <TableCell class="font-bold text-xs">{{ shippingName(order.shipping_address) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="orderStatusTone(order.status)">{{ getOrderStatusName(order.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="paymentStatusTone(order.payment_status)">{{ getPaymentStatusName(order.payment_status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="shippingStatusTone(order.shipping_status)">{{ getShippingStatusName(order.shipping_status) }}</AdminStatusBadge>
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
                  <DropdownMenuItem @select="showOrderDetail(order)">
                    <Eye class="size-4" />
                    查看详情
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('order:edit')" @select="showStatusDialog(order)">
                    <RefreshCw class="size-4" />
                    状态管理
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('order:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('order:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(order)"
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
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="detailDialogVisible">
      <DialogContent size="xl" class="max-h-[90dvh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>订单详情</DialogTitle>
          <DialogDescription>{{ currentOrder?.order_number || '订单信息' }}</DialogDescription>
        </DialogHeader>

        <div v-if="currentOrder" class="space-y-6">
          <OrderDetailSection title="订单信息">
            <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-2">
              <DetailItem label="订单号">{{ currentOrder.order_number }}</DetailItem>
              <DetailItem label="订单状态"><AdminStatusBadge :tone="orderStatusTone(currentOrder.status)">{{ getOrderStatusName(currentOrder.status) }}</AdminStatusBadge></DetailItem>
              <DetailItem label="支付状态"><AdminStatusBadge :tone="paymentStatusTone(currentOrder.payment_status)">{{ getPaymentStatusName(currentOrder.payment_status) }}</AdminStatusBadge></DetailItem>
              <DetailItem label="物流状态"><AdminStatusBadge :tone="shippingStatusTone(currentOrder.shipping_status)">{{ getShippingStatusName(currentOrder.shipping_status) }}</AdminStatusBadge></DetailItem>
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
            <div v-if="currentOrder.tracking_number && hasPermission('order:edit')" class="flex justify-end">
              <Button variant="outline" size="sm" class="rounded-full" :disabled="syncingTracking" @click="syncCurrentOrderTracking">
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
                <Textarea id="admin-note" v-model="adminNoteForm.admin_note" class="mt-2 min-h-24" placeholder="请输入管理员备注" />
                <Button
                  v-if="hasPermission('order:edit')"
                  size="sm"
                  class="mt-2 rounded-full"
                  @click="updateAdminNote"
                >
                  保存备注
                </Button>
              </div>
            </div>
          </OrderDetailSection>
        </div>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="statusDialogVisible">
      <DialogContent size="lg">
        <form @submit.prevent="submitStatus">
          <DialogHeader>
            <DialogTitle>状态管理</DialogTitle>
            <DialogDescription>更新订单 {{ statusForm.order_number }} 的履约状态。</DialogDescription>
          </DialogHeader>

          <div class="space-y-4 py-5">
            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 订单状态</span>
              <Select v-model="statusForm.status">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in editableOrderStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SHIPPING / 物流状态</span>
              <Select v-model="statusForm.shipping_status">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in editableShippingStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">TRACKING / 物流单号</span>
              <Input v-model="statusForm.tracking_number" />
            </label>

            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">PROVIDER / 追踪服务商</span>
              <Select v-model="statusForm.tracking_provider_id">
                <SelectTrigger class="w-full"><SelectValue placeholder="请选择追踪 Provider" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">请选择追踪 Provider</SelectItem>
                  <SelectItem v-for="provider in trackingProviders" :key="provider.id" :value="String(provider.id)">
                    {{ provider.provider_name }} / {{ provider.provider_code }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>

            <div class="grid gap-3 sm:grid-cols-2">
              <label class="block space-y-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">CARRIER / 本地承运商</span>
                <Select v-model="statusForm.carrier_id" @update:model-value="handleStatusCarrierChange">
                  <SelectTrigger class="w-full"><SelectValue placeholder="可选：选择承运商" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">不指定承运商</SelectItem>
                    <SelectItem v-for="carrier in carriers" :key="carrier.id" :value="String(carrier.id)">
                      {{ carrier.name }} / {{ carrier.code }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </label>

              <label class="block space-y-1">
                <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SERVICE / 线路服务</span>
                <Select v-model="statusForm.carrier_service_id" @update:model-value="handleStatusCarrierServiceChange">
                  <SelectTrigger class="w-full"><SelectValue placeholder="可选：选择线路服务" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">不指定线路服务</SelectItem>
                    <SelectItem v-for="service in filteredStatusCarrierServices" :key="service.id" :value="String(service.id)">
                      {{ service.service_name }} / {{ service.service_code }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </label>
            </div>

            <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
              保存时系统会按“线路服务映射优先、承运商映射其次”解析 Provider Carrier Code。
              当前可预览：<span class="font-mono font-bold text-foreground">{{ resolvedProviderCarrierCodeLabel }}</span>
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" @click="statusDialogVisible = false">取消</Button>
            <Button type="submit" :disabled="submitting">
              <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
              {{ submitting ? '正在保存' : '保存' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="confirmation.destructive"
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Banknote,
  CalendarCheck2,
  CircleCheck,
  CircleX,
  Download,
  Eye,
  LoaderCircle,
  MoreHorizontal,
  RefreshCw,
  RotateCcw,
  Search,
  ShoppingBag,
  Trash2,
  TrendingUp
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableEmpty,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { shippingApi } from '@/api/shipping'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const FilterSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    options: { type: Array, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => emit('update:modelValue', value)
      }, {
        default: () => [
          h(SelectTrigger, { class: 'h-9 w-full' }, { default: () => h(SelectValue) }),
          h(SelectContent, {}, {
            default: () => props.options.map((option) => h(SelectItem, { value: option.value }, { default: () => option.label }))
          })
        ]
      })
    ])
  }
})

const OrderDetailSection = defineComponent({
  props: { title: { type: String, required: true } },
  setup(props, { slots }) {
    return () => h('section', { class: 'space-y-3 border-t border-dashed pt-5 first:border-t-0 first:pt-0' }, [
      h('h3', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
      slots.default?.()
    ])
  }
})

const DetailItem = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots, attrs }) {
    return () => h('div', { ...attrs, class: ['border-b p-3 last:border-b-0 sm:border-r sm:last:border-r-0', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h('dd', { class: 'mt-1 text-xs font-bold' }, slots.default?.())
    ])
  }
})

const AmountRow = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: 0 }
  },
  setup(props) {
    return () => h('div', { class: 'flex items-center justify-between' }, [
      h('dt', { class: 'text-muted-foreground' }, props.label),
      h('dd', { class: 'tabular-nums' }, `¥${Number(props.value || 0).toFixed(2)}`)
    ])
  }
})

const authStore = useAuthStore()
const loading = ref(false)
const orders = ref([])
const selectedOrders = ref([])
const detailDialogVisible = ref(false)
const statusDialogVisible = ref(false)
const submitting = ref(false)
const syncingTracking = ref(false)
const currentOrder = ref(null)
const currentTrackingEvents = ref([])
const currentTrackingShipment = ref(null)
const stats = ref({})
const carriers = ref([])
const carrierServices = ref([])
const trackingProviders = ref([])
const trackingCarrierMappings = ref([])

const filters = reactive({
  search: '',
  status: 'all',
  payment_status: 'all',
  shipping_status: 'all',
  start_date: '',
  end_date: ''
})

const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const statusForm = reactive({
  id: null,
  order_number: '',
  status: 'pending',
  shipping_status: 'pending',
  tracking_number: '',
  tracking_provider_id: 'none',
  carrier_id: 'none',
  carrier_service_id: 'none'
})
const adminNoteForm = reactive({ admin_note: '' })
const confirmation = reactive({
  open: false,
  type: '',
  target: null,
  status: '',
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const orderStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '待支付', value: 'pending' },
  { label: '已支付', value: 'paid' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已完成', value: 'completed' },
  { label: '已取消', value: 'cancelled' },
  { label: '已退款', value: 'refunded' }
]
const paymentStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '未支付', value: 'unpaid' },
  { label: '已支付', value: 'paid' },
  { label: '已退款', value: 'refunded' }
]
const shippingStatusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '待处理', value: 'pending' },
  { label: '处理中', value: 'processing' },
  { label: '已发货', value: 'shipped' },
  { label: '已送达', value: 'delivered' }
]
const editableOrderStatusOptions = orderStatusOptions.filter((option) => !['all', 'paid', 'refunded'].includes(option.value))
const editableShippingStatusOptions = shippingStatusOptions.filter((option) => option.value !== 'all')

const statItems = computed(() => [
  { key: 'total', label: '总订单数', value: stats.value.total || 0, icon: ShoppingBag, tone: 'gray' },
  { key: 'today', label: '今日订单', value: stats.value.today || 0, icon: CalendarCheck2, tone: 'blue' },
  { key: 'revenue', label: '总销售额', value: `¥${formatMoney(stats.value.total_revenue)}`, icon: Banknote, tone: 'green' },
  { key: 'today-revenue', label: '今日销售额', value: `¥${formatMoney(stats.value.today_revenue)}`, icon: TrendingUp, tone: 'amber' }
])
const selectionState = computed(() => {
  if (orders.value.length === 0 || selectedOrders.value.length === 0) return false
  return selectedOrders.value.length === orders.value.length ? true : 'indeterminate'
})
const filteredStatusCarrierServices = computed(() => {
  const carrierID = numericSelectID(statusForm.carrier_id)
  if (!carrierID) return carrierServices.value
  return carrierServices.value.filter((service) => Number(service.carrier_id) === carrierID)
})
const selectedStatusCarrierService = computed(() => {
  const serviceID = numericSelectID(statusForm.carrier_service_id)
  if (!serviceID) return null
  return carrierServices.value.find((service) => Number(service.id) === serviceID) || null
})
const selectedTrackingCarrierMapping = computed(() => {
  const providerID = numericSelectID(statusForm.tracking_provider_id)
  if (!providerID) return null

  const serviceID = numericSelectID(statusForm.carrier_service_id)
  if (serviceID) {
    const serviceMapping = trackingCarrierMappings.value.find((mapping) =>
      Number(mapping.provider_id) === providerID &&
      mapping.scope === 'carrier_service' &&
      Number(mapping.carrier_service_id) === serviceID
    )
    if (serviceMapping) return serviceMapping
  }

  const serviceCarrierID = selectedStatusCarrierService.value?.carrier_id
  const carrierID = numericSelectID(statusForm.carrier_id) || (serviceCarrierID ? Number(serviceCarrierID) : null)
  if (!carrierID) return null

  return trackingCarrierMappings.value.find((mapping) =>
    Number(mapping.provider_id) === providerID &&
    mapping.scope === 'carrier' &&
    Number(mapping.carrier_id) === carrierID
  ) || null
})
const resolvedProviderCarrierCodeLabel = computed(() => {
  const mapping = selectedTrackingCarrierMapping.value
  if (!mapping) return '未匹配映射'
  return `${mapping.provider_carrier_code}${mapping.provider_carrier_name ? ` / ${mapping.provider_carrier_name}` : ''}`
})

const hasPermission = (permission) => authStore.hasPermission(permission)

const getOrderStatusName = (status) => orderStatusOptions.find((option) => option.value === status)?.label || status
const orderStatusTone = (status) => ({
  pending: 'gray', paid: 'green', processing: 'amber', shipped: 'blue', completed: 'green', cancelled: 'coral', refunded: 'amber'
})[status] || 'gray'
const getPaymentStatusName = (status) => paymentStatusOptions.find((option) => option.value === status)?.label || status
const paymentStatusTone = (status) => ({ unpaid: 'gray', paid: 'green', refunded: 'amber' })[status] || 'gray'
const getShippingStatusName = (status) => shippingStatusOptions.find((option) => option.value === status)?.label || status
const shippingStatusTone = (status) => ({ pending: 'gray', processing: 'amber', shipped: 'blue', delivered: 'green' })[status] || 'gray'
const trackingSyncStatusName = (status) => ({
  pending: '待同步',
  syncing: '同步中',
  synced: '已同步',
  failed: '同步失败'
})[status] || '未建立'
const trackingSyncStatusTone = (status) => ({
  pending: 'gray',
  syncing: 'blue',
  synced: 'green',
  failed: 'coral'
})[status] || 'gray'
const trackingRegistrationStatusName = (status) => ({
  pending: '待登记',
  registered: '已登记',
  failed: '登记失败'
})[status] || '未建立'

const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatMoney = (amount) => Number(amount || 0).toFixed(2)
const shippingName = (address) => [address?.first_name, address?.last_name].filter(Boolean).join(' ') || '-'
const shippingAddressLine = (address) => [address?.address_1, address?.address_2].filter(Boolean).join(' ') || '-'
const numericSelectID = (value) => {
  const parsed = Number(value)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}
const selectValueFromID = (value) => numericSelectID(value) ? String(value) : 'none'
const defaultTrackingProviderValue = () => trackingProviders.value[0]?.id ? String(trackingProviders.value[0].id) : 'none'
const providerValueForLocalShippingSource = (carrierIDValue, carrierServiceIDValue) => {
  const carrierServiceID = numericSelectID(carrierServiceIDValue)
  if (carrierServiceID) {
    const serviceMapping = trackingCarrierMappings.value.find((mapping) =>
      mapping.scope === 'carrier_service' &&
      Number(mapping.carrier_service_id) === carrierServiceID
    )
    if (serviceMapping?.provider_id) return String(serviceMapping.provider_id)
  }

  const service = carrierServiceID
    ? carrierServices.value.find((item) => Number(item.id) === carrierServiceID)
    : null
  const carrierID = numericSelectID(carrierIDValue) || numericSelectID(service?.carrier_id)
  if (carrierID) {
    const carrierMapping = trackingCarrierMappings.value.find((mapping) =>
      mapping.scope === 'carrier' &&
      Number(mapping.carrier_id) === carrierID
    )
    if (carrierMapping?.provider_id) return String(carrierMapping.provider_id)
  }

  return defaultTrackingProviderValue()
}
const defaultTrackingProviderForOrder = (order) => {
  const storedProvider = selectValueFromID(order?.tracking_provider_id)
  if (storedProvider !== 'none') return storedProvider
  return providerValueForLocalShippingSource(order?.carrier_id, order?.carrier_service_id)
}
const orderCarrierLabel = (order) => {
  const carrierID = Number(order?.carrier_id)
  if (!Number.isFinite(carrierID) || carrierID <= 0) return '-'
  const carrier = carriers.value.find((item) => Number(item.id) === carrierID)
  return carrier ? `${carrier.name} / ${carrier.code}` : `Carrier #${carrierID}`
}
const orderCarrierServiceLabel = (order) => {
  const serviceID = Number(order?.carrier_service_id)
  if (!Number.isFinite(serviceID) || serviceID <= 0) return '-'
  const service = carrierServices.value.find((item) => Number(item.id) === serviceID)
  return service ? `${service.service_name} / ${service.service_code}` : `Carrier service #${serviceID}`
}

const buildFilterParams = () => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.payment_status !== 'all' ? { payment_status: filters.payment_status } : {}),
  ...(filters.shipping_status !== 'all' ? { shipping_status: filters.shipping_status } : {}),
  ...(filters.start_date ? { start_date: filters.start_date } : {}),
  ...(filters.end_date ? { end_date: filters.end_date } : {})
})

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/orders/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch order stats:', error)
  }
}

const fetchOrders = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/orders', {
      params: { page: pagination.page, page_size: pagination.pageSize, ...buildFilterParams() }
    })
    orders.value = response.data.orders || []
    pagination.total = response.data.pagination?.total || 0
    selectedOrders.value = []
  } catch (error) {
    console.error('Failed to fetch orders:', error)
  } finally {
    loading.value = false
  }
}

const fetchShippingLookups = async () => {
  try {
    const [providerList, carrierList, serviceList, mappingList] = await Promise.all([
      shippingApi.listTrackingProviders({ enabled: 'true' }),
      shippingApi.listCarriers({ enabled: 'true' }),
      shippingApi.listCarrierServices({ enabled: 'true' }),
      shippingApi.listTrackingCarrierMappings({ enabled: 'true' })
    ])
    trackingProviders.value = providerList
    carriers.value = carrierList
    carrierServices.value = serviceList
    trackingCarrierMappings.value = mappingList
  } catch (error) {
    console.error('Failed to fetch shipping lookups:', error)
  }
}

const unwrapTrackingEvents = (response) => {
  const payload = response.data?.data ?? response.data ?? {}
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (Array.isArray(payload.events)) return payload.events
  return []
}

const fetchOrderTrackingEvents = async (orderID) => {
  if (!orderID) {
    currentTrackingEvents.value = []
    return
  }

  try {
    const response = await axios.get(`/api/v1/shipping/orders/${orderID}/tracking`)
    currentTrackingEvents.value = unwrapTrackingEvents(response)
  } catch (error) {
    currentTrackingEvents.value = []
    console.error('Failed to fetch order tracking events:', error)
  }
}

const refreshOrders = () => Promise.all([fetchOrders(), fetchStats()])
const applyFilters = () => { pagination.page = 1; fetchOrders() }
const resetFilters = () => {
  Object.assign(filters, { search: '', status: 'all', payment_status: 'all', shipping_status: 'all', start_date: '', end_date: '' })
  pagination.page = 1
  fetchOrders()
}
const updatePage = (page) => { pagination.page = page; fetchOrders() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchOrders() }

const showOrderDetail = async (order) => {
  try {
    currentTrackingShipment.value = null
    const [response] = await Promise.all([
      axios.get(`/api/admin/orders/${order.id}`),
      fetchShippingLookups(),
      fetchOrderTrackingEvents(order.id)
    ])
    currentOrder.value = response.data.order
    currentTrackingShipment.value = response.data.tracking_shipment || null
    adminNoteForm.admin_note = currentOrder.value.admin_note || ''
    detailDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch order detail:', error)
  }
}

const showStatusDialog = async (order) => {
  await fetchShippingLookups()
  Object.assign(statusForm, {
    id: order.id,
    order_number: order.order_number,
    status: order.status,
    shipping_status: order.shipping_status,
    tracking_number: order.tracking_number || '',
    tracking_provider_id: defaultTrackingProviderForOrder(order),
    carrier_id: selectValueFromID(order.carrier_id),
    carrier_service_id: selectValueFromID(order.carrier_service_id)
  })
  statusDialogVisible.value = true
}

const handleStatusCarrierChange = (value) => {
  const carrierID = numericSelectID(value)
  if (!carrierID) {
    statusForm.carrier_id = 'none'
    return
  }

  const service = selectedStatusCarrierService.value
  if (service && Number(service.carrier_id) !== carrierID) {
    statusForm.carrier_service_id = 'none'
  }
  statusForm.tracking_provider_id = providerValueForLocalShippingSource(carrierID, statusForm.carrier_service_id)
}

const handleStatusCarrierServiceChange = (value) => {
  const serviceID = numericSelectID(value)
  if (!serviceID) {
    statusForm.carrier_service_id = 'none'
    return
  }

  const service = carrierServices.value.find((item) => Number(item.id) === serviceID)
  if (service?.carrier_id) {
    statusForm.carrier_id = String(service.carrier_id)
  }
  statusForm.tracking_provider_id = providerValueForLocalShippingSource(statusForm.carrier_id, serviceID)
}

const submitStatus = async () => {
  submitting.value = true
  try {
    await axios.patch(`/api/admin/orders/${statusForm.id}/status`, { status: statusForm.status })
    await axios.patch(`/api/admin/orders/${statusForm.id}/shipping-status`, { shipping_status: statusForm.shipping_status })
    const trackingNumber = statusForm.tracking_number?.trim()
    if (trackingNumber) {
      const trackingProviderID = numericSelectID(statusForm.tracking_provider_id)
      const carrierID = numericSelectID(statusForm.carrier_id)
      const carrierServiceID = numericSelectID(statusForm.carrier_service_id)

      if (!trackingProviderID) {
        toast.error('请选择追踪 Provider')
        return
      }
      if (!carrierID && !carrierServiceID) {
        toast.error('请选择本地承运商或线路服务')
        return
      }

      await axios.patch(`/api/admin/orders/${statusForm.id}/tracking`, {
        tracking_number: trackingNumber,
        tracking_provider_id: trackingProviderID,
        carrier_id: carrierID,
        carrier_service_id: carrierServiceID
      })
    }
    toast.success('订单状态已更新')
    statusDialogVisible.value = false
    await refreshOrders()
  } catch (error) {
    console.error('Failed to update order status:', error)
  } finally {
    submitting.value = false
  }
}

const syncCurrentOrderTracking = async () => {
  if (!currentOrder.value?.id) return

  syncingTracking.value = true
  try {
    const response = await axios.post(`/api/admin/orders/${currentOrder.value.id}/tracking/sync`)
    currentTrackingEvents.value = response.data?.tracking?.events || []
    currentTrackingShipment.value = response.data?.tracking?.shipment || currentTrackingShipment.value
    toast.success(`物流轨迹已同步：${currentTrackingEvents.value.length} 条`)
  } catch (error) {
    console.error('Failed to sync tracking info:', error)
    toast.error(error.response?.data?.error || '物流轨迹同步失败')
  } finally {
    syncingTracking.value = false
  }
}

const updateAdminNote = async () => {
  try {
    await axios.patch(`/api/admin/orders/${currentOrder.value.id}/admin-note`, { admin_note: adminNoteForm.admin_note })
    currentOrder.value.admin_note = adminNoteForm.admin_note
    toast.success('管理员备注已保存')
  } catch (error) {
    console.error('Failed to update admin note:', error)
  }
}

const isSelected = (orderId) => selectedOrders.value.some((order) => order.id === orderId)
const toggleAllOrders = (checked) => { selectedOrders.value = checked === true ? [...orders.value] : [] }
const toggleOrder = (order, checked) => {
  if (checked === true && !isSelected(order.id)) selectedOrders.value = [...selectedOrders.value, order]
  else if (checked !== true) selectedOrders.value = selectedOrders.value.filter((selected) => selected.id !== order.id)
}

const setConfirmation = (values) => Object.assign(confirmation, { open: true, destructive: false, confirmLabel: '确定', ...values })
const requestDelete = (order) => setConfirmation({
  type: 'delete', target: order, title: '删除订单？', description: `订单 ${order.order_number} 将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status) => {
  const completing = status === 'completed'
  setConfirmation({
    type: 'batch-status',
    target: [...selectedOrders.value],
    status,
    title: completing ? '批量完成订单？' : '批量取消订单？',
    description: `将 ${selectedOrders.value.length} 个订单标记为${completing ? '已完成' : '已取消'}。`,
    confirmLabel: completing ? '批量完成' : '批量取消',
    destructive: !completing
  })
}

const executeConfirmedAction = async () => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete') {
      await axios.delete(`/api/admin/orders/${target.id}`)
      toast.success('订单已删除')
    } else if (type === 'batch-status') {
      const response = await axios.post('/api/admin/orders/batch-status', {
        order_ids: target.map((order) => order.id),
        status
      })
      toast.success(`批量更新成功：${response.data.updated}/${response.data.total}`)
    }
    await refreshOrders()
  } catch (error) {
    console.error('Failed to update orders:', error)
  }
}

const exportOrders = async () => {
  try {
    const query = new URLSearchParams(buildFilterParams()).toString()
    const response = await axios.get(`/api/admin/orders/export${query ? `?${query}` : ''}`, { responseType: 'blob' })
    const url = window.URL.createObjectURL(new Blob([response.data], { type: 'text/csv' }))
    const link = document.createElement('a')
    link.href = url
    link.download = `orders_${Date.now()}.csv`
    link.click()
    window.URL.revokeObjectURL(url)
    toast.success('订单已导出')
  } catch (error) {
    console.error('Failed to export orders:', error)
  }
}

onMounted(() => {
  refreshOrders()
  fetchShippingLookups()
})
</script>
