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

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
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
              <DetailItem label="物流公司">{{ currentOrder.carrier_code || '-' }}</DetailItem>
              <DetailItem label="创建时间">{{ formatDate(currentOrder.created_at) }}</DetailItem>
              <DetailItem label="支付时间">{{ currentOrder.paid_at ? formatDate(currentOrder.paid_at) : '-' }}</DetailItem>
            </dl>
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
      <DialogContent size="sm">
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
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">CARRIER / 物流公司代码</span>
              <Input v-model="statusForm.carrier_code" />
            </label>
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
const currentOrder = ref(null)
const stats = ref({})

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
  carrier_code: ''
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

const hasPermission = (permission) => authStore.hasPermission(permission)

const getOrderStatusName = (status) => orderStatusOptions.find((option) => option.value === status)?.label || status
const orderStatusTone = (status) => ({
  pending: 'gray', paid: 'green', processing: 'amber', shipped: 'blue', completed: 'green', cancelled: 'coral', refunded: 'amber'
})[status] || 'gray'
const getPaymentStatusName = (status) => paymentStatusOptions.find((option) => option.value === status)?.label || status
const paymentStatusTone = (status) => ({ unpaid: 'gray', paid: 'green', refunded: 'amber' })[status] || 'gray'
const getShippingStatusName = (status) => shippingStatusOptions.find((option) => option.value === status)?.label || status
const shippingStatusTone = (status) => ({ pending: 'gray', processing: 'amber', shipped: 'blue', delivered: 'green' })[status] || 'gray'

const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatMoney = (amount) => Number(amount || 0).toFixed(2)
const shippingName = (address) => [address?.first_name, address?.last_name].filter(Boolean).join(' ') || '-'
const shippingAddressLine = (address) => [address?.address_1, address?.address_2].filter(Boolean).join(' ') || '-'

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
    const response = await axios.get(`/api/admin/orders/${order.id}`)
    currentOrder.value = response.data.order
    adminNoteForm.admin_note = currentOrder.value.admin_note || ''
    detailDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch order detail:', error)
  }
}

const showStatusDialog = (order) => {
  Object.assign(statusForm, {
    id: order.id,
    order_number: order.order_number,
    status: order.status,
    shipping_status: order.shipping_status,
    tracking_number: order.tracking_number || '',
    carrier_code: order.carrier_code || ''
  })
  statusDialogVisible.value = true
}

const submitStatus = async () => {
  submitting.value = true
  try {
    await axios.patch(`/api/admin/orders/${statusForm.id}/status`, { status: statusForm.status })
    await axios.patch(`/api/admin/orders/${statusForm.id}/shipping-status`, { shipping_status: statusForm.shipping_status })
    if (statusForm.tracking_number) {
      await axios.patch(`/api/admin/orders/${statusForm.id}/tracking`, {
        tracking_number: statusForm.tracking_number,
        carrier_code: statusForm.carrier_code
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

onMounted(refreshOrders)
</script>
