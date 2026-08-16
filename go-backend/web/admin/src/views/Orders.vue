<template>
  <div class="space-y-4">
    <AdminPageHeader
      :title="activeOrderTab === 'disputes' ? '拒付订单' : '订单管理'"
      :description="activeOrderTab === 'disputes' ? '按订单分析 Stripe / PayPal 拒付，并在必要时联系客户' : '查看订单履约、支付和物流状态'"
    >
      <template #actions>
        <Button v-if="activeOrderTab === 'list'" variant="outline" @click="exportOrders">
          <Download class="size-4" />
          导出订单
        </Button>
        <Button v-else variant="outline" :disabled="disputeLoading" @click="fetchOrderDisputes">
          <RefreshCw :class="['size-4', disputeLoading ? 'animate-spin' : '']" />
          刷新拒付
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="activeStatItems" />

    <OrderFilterPanel
      v-if="activeOrderTab === 'list'"
      :filters="filters"
      :order-status-options="orderStatusOptions"
      :payment-status-options="paymentStatusOptions"
      :shipping-status-options="shippingStatusOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <OrderTablePanel
      v-if="activeOrderTab === 'list'"
      :loading="loading"
      :orders="orders"
      :selected-orders="selectedOrders"
      :pagination="pagination"
      :selection-state="selectionState"
      :can-edit="hasPermission('order:edit')"
      :can-delete="hasPermission('order:delete')"
      :order-status-name="getOrderStatusName"
      :order-status-tone="orderStatusTone"
      :payment-status-name="getPaymentStatusName"
      :payment-status-tone="paymentStatusTone"
      :shipping-status-name="getShippingStatusName"
      :shipping-status-tone="shippingStatusTone"
      :shipping-name="shippingName"
      :format-money="formatMoney"
      :format-date="formatDate"
      @batch-status="requestBatchStatus"
      @toggle-all-orders="toggleAllOrders"
      @toggle-order="toggleOrder"
      @view-detail="showOrderDetail"
      @show-status="showStatusDialog"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <template v-else>
      <AdminFilterPanel>
        <form class="grid grid-cols-1 gap-3 md:grid-cols-[180px_220px_1fr_auto]" @submit.prevent="applyDisputeFilters">
          <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER / 渠道</span>
            <select v-model="disputeFilters.provider" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
              <option value="">全部渠道</option>
              <option value="stripe">Stripe</option>
              <option value="paypal">PayPal</option>
            </select>
          </label>
          <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
            <Input v-model="disputeFilters.status" placeholder="例如 needs_response / OPEN" />
          </label>
          <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
            <Input v-model="disputeFilters.search" placeholder="拒付号、订单号、邮箱、物流单号" />
          </label>
          <div class="flex items-end gap-2">
            <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="disputeLoading">
              查询
            </Button>
            <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetDisputeFilters">
              重置
            </Button>
          </div>
        </form>
      </AdminFilterPanel>

      <OrderDisputeTablePanel
        :loading="disputeLoading"
        :disputes="disputeOrders"
        :pagination="disputePagination"
        :format-money="formatMoney"
        :format-date="formatDate"
        @view-order="showDisputeOrderDetail"
        @contact-customer="openDisputeContactEmail"
        @open-payment-workbench="openPaymentWorkbench"
        @update-page="updateDisputePage"
        @update-page-size="updateDisputePageSize"
      />
    </template>

    <OrderDetailDialog
      v-model:open="detailDialogVisible"
      v-model:admin-note="adminNoteForm.admin_note"
      :current-order="currentOrder"
      :current-tracking-events="currentTrackingEvents"
      :current-tracking-shipment="currentTrackingShipment"
      :dispute-analysis="currentDisputeAnalysis"
      :dispute-analysis-loading="disputeAnalysisLoading"
      :syncing-tracking="syncingTracking"
      :saving-customs-item-id="savingCustomsItemID"
      :exporting-customs="exportingCustoms"
      :can-edit="hasPermission('order:edit')"
      :order-status-name="getOrderStatusName"
      :order-status-tone="orderStatusTone"
      :payment-status-name="getPaymentStatusName"
      :payment-status-tone="paymentStatusTone"
      :shipping-status-name="getShippingStatusName"
      :shipping-status-tone="shippingStatusTone"
      :tracking-sync-status-name="trackingSyncStatusName"
      :tracking-sync-status-tone="trackingSyncStatusTone"
      :tracking-registration-status-name="trackingRegistrationStatusName"
      :format-date="formatDate"
      :format-money="formatMoney"
      :shipping-name="shippingName"
      :shipping-address-line="shippingAddressLine"
      :order-carrier-label="orderCarrierLabel"
      :order-carrier-service-label="orderCarrierServiceLabel"
      @sync-tracking="syncCurrentOrderTracking"
      @update-note="updateAdminNote"
      @update-customs="updateOrderItemCustoms"
      @export-customs="exportOrderCustoms"
      @contact-dispute="openDisputeContactEmail"
      @open-payment-workbench="openPaymentWorkbench"
    />

    <OrderDisputeContactEmailDialog
      v-model:open="disputeEmailDialogVisible"
      :form="disputeEmailForm"
      :sending="disputeEmailSending"
      :mailto-url="disputeEmailMailtoUrl"
      @update:subject="disputeEmailForm.subject = $event"
      @update:body="disputeEmailForm.body = $event"
      @submit="submitDisputeContactEmail"
    />

    <OrderStatusDialog
      v-model:open="statusDialogVisible"
      :status-form="statusForm"
      :editable-order-status-options="editableOrderStatusOptions"
      :editable-shipping-status-options="editableShippingStatusOptions"
      :tracking-providers="trackingProviders"
      :carriers="carriers"
      :filtered-status-carrier-services="filteredStatusCarrierServices"
      :resolved-provider-carrier-code-label="resolvedProviderCarrierCodeLabel"
      :submitting="submitting"
      @submit="submitStatus"
      @carrier-change="handleStatusCarrierChange"
      @carrier-service-change="handleStatusCarrierServiceChange"
    />

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

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  Banknote,
  CalendarCheck2,
  Download,
  RefreshCw,
  ShieldAlert,
  ShoppingBag,
  TrendingUp
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import OrderDisputeContactEmailDialog from '@/components/admin/order/OrderDisputeContactEmailDialog.vue'
import OrderDisputeTablePanel from '@/components/admin/order/OrderDisputeTablePanel.vue'
import OrderDetailDialog from '@/components/admin/order/OrderDetailDialog.vue'
import OrderFilterPanel from '@/components/admin/order/OrderFilterPanel.vue'
import OrderStatusDialog from '@/components/admin/order/OrderStatusDialog.vue'
import OrderTablePanel from '@/components/admin/order/OrderTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { ordersApi } from '@/api/orders'
import { shippingApi } from '@/api/shipping'
import {
  editableOrderStatusOptions,
  editableShippingStatusOptions,
  formatDate,
  formatMoney,
  getOrderStatusName,
  getPaymentStatusName,
  getShippingStatusName,
  numericSelectID,
  orderStatusOptions,
  orderStatusTone,
  paymentStatusOptions,
  paymentStatusTone,
  selectValueFromID,
  shippingAddressLine,
  shippingName,
  shippingStatusOptions,
  shippingStatusTone,
  trackingRegistrationStatusName,
  trackingSyncStatusName,
  trackingSyncStatusTone
} from '@/lib/orderPresentation'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'
import type {
  OrderConfirmation,
  OrderDisputeAnalysis,
  OrderDisputeCase,
  OrderDisputeEmailForm,
  OrderFilters,
  OrderID,
  OrderPagination,
  OrderRecord,
  OrderStatItem,
  OrderStats,
  OrderStatusForm,
  OrderStatusTone,
  ShippingCarrier,
  ShippingCarrierService,
  TrackingCarrierMapping,
  TrackingEvent,
  TrackingProvider,
  TrackingShipment
} from '@/components/admin/order/orderTypes'

interface OrderListResponse {
  orders?: OrderRecord[]
  pagination?: Partial<OrderPagination>
}

interface OrderDetailResponse {
  order?: OrderRecord | null
  tracking_shipment?: TrackingShipment | null
}

interface TrackingEventsResponse {
  data?: TrackingEvent[] | { data?: TrackingEvent[]; events?: TrackingEvent[] }
  events?: TrackingEvent[]
}

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const orders = ref<OrderRecord[]>([])
const disputeOrders = ref<OrderDisputeCase[]>([])
const selectedOrders = ref<OrderRecord[]>([])
const detailDialogVisible = ref(false)
const disputeEmailDialogVisible = ref(false)
const statusDialogVisible = ref(false)
const submitting = ref(false)
const disputeLoading = ref(false)
const disputeAnalysisLoading = ref(false)
const disputeEmailSending = ref(false)
const syncingTracking = ref(false)
const savingCustomsItemID = ref<OrderID | null>(null)
const exportingCustoms = ref(false)
const currentOrder = ref<OrderRecord | null>(null)
const disputeEmailOrderID = ref<OrderID | null>(null)
const currentDisputeAnalysis = ref<OrderDisputeAnalysis | null>(null)
const currentTrackingEvents = ref<TrackingEvent[]>([])
const currentTrackingShipment = ref<TrackingShipment | null>(null)
const stats = ref<OrderStats>({})
const carriers = ref<ShippingCarrier[]>([])
const carrierServices = ref<ShippingCarrierService[]>([])
const trackingProviders = ref<TrackingProvider[]>([])
const trackingCarrierMappings = ref<TrackingCarrierMapping[]>([])

const filters = reactive<OrderFilters>({
  search: '',
  status: 'all',
  payment_status: 'all',
  shipping_status: 'all',
  start_date: '',
  end_date: ''
})

const pagination = reactive<OrderPagination>({ page: 1, pageSize: 20, total: 0 })
const disputePagination = reactive<OrderPagination>({ page: 1, pageSize: 20, total: 0 })
const disputeFilters = reactive({
  provider: '',
  status: '',
  search: ''
})
const disputeEmailForm = reactive<OrderDisputeEmailForm>({
  provider: '',
  dispute_id: null,
  to: '',
  subject: '',
  body: ''
})
const disputeEmailMailtoUrl = ref('')
const statusForm = reactive<OrderStatusForm>({
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
const confirmation = reactive<OrderConfirmation>({
  open: false,
  type: '',
  target: null,
  status: '',
  title: '',
  description: '',
  confirmLabel: '确定',
  destructive: false
})

const statItems = computed<OrderStatItem[]>(() => [
  { key: 'total', label: '总订单数', value: stats.value.total || 0, icon: ShoppingBag, tone: 'gray' },
  { key: 'today', label: '今日订单', value: stats.value.today || 0, icon: CalendarCheck2, tone: 'blue' },
  { key: 'revenue', label: '总销售额', value: `¥${formatMoney(stats.value.total_revenue)}`, icon: Banknote, tone: 'green' },
  { key: 'today-revenue', label: '今日销售额', value: `¥${formatMoney(stats.value.today_revenue)}`, icon: TrendingUp, tone: 'amber' }
])
const activeOrderTab = computed<'list' | 'disputes'>(() => route.name === 'OrdersDisputes' ? 'disputes' : 'list')
const activeStatItems = computed<OrderStatItem[]>(() => {
  if (activeOrderTab.value === 'list') return statItems.value
  const needsResponse = disputeOrders.value.filter((item) => item.needs_response).length
  const likelyMistakes = disputeOrders.value.filter((item) => item.mistake_assessment?.level === 'likely_mistake').length
  const evidenceBlocked = disputeOrders.value.filter((item) => Number(item.evidence_summary?.blocker_count || 0) > 0).length
  return [
    { key: 'disputes', label: '拒付订单', value: disputePagination.total || 0, icon: ShieldAlert, tone: 'gray' },
    { key: 'needs-response', label: '需要响应', value: needsResponse, icon: ShieldAlert, tone: needsResponse ? 'coral' : 'green' },
    { key: 'likely-mistakes', label: '疑似误操作', value: likelyMistakes, icon: CalendarCheck2, tone: likelyMistakes ? 'amber' : 'gray' },
    { key: 'evidence-blocked', label: '证据阻断', value: evidenceBlocked, icon: ShoppingBag, tone: evidenceBlocked ? 'coral' : 'green' }
  ]
})
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

const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
const defaultTrackingProviderValue = (): string => (trackingProviders.value[0]?.id ? String(trackingProviders.value[0].id) : 'none')
const providerValueForLocalShippingSource = (
  carrierIDValue: OrderID | null | undefined,
  carrierServiceIDValue: OrderID | null | undefined
): string => {
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
const defaultTrackingProviderForOrder = (order: OrderRecord | null): string => {
  const storedProvider = selectValueFromID(order?.tracking_provider_id)
  if (storedProvider !== 'none') return storedProvider
  return providerValueForLocalShippingSource(order?.carrier_id, order?.carrier_service_id)
}
const orderCarrierLabel = (order: OrderRecord | null): string => {
  const carrierID = Number(order?.carrier_id)
  if (!Number.isFinite(carrierID) || carrierID <= 0) return '-'
  const carrier = carriers.value.find((item) => Number(item.id) === carrierID)
  return carrier ? `${carrier.name} / ${carrier.code}` : `Carrier #${carrierID}`
}
const orderCarrierServiceLabel = (order: OrderRecord | null): string => {
  const serviceID = Number(order?.carrier_service_id)
  if (!Number.isFinite(serviceID) || serviceID <= 0) return '-'
  const service = carrierServices.value.find((item) => Number(item.id) === serviceID)
  return service ? `${service.service_name} / ${service.service_code}` : `Carrier service #${serviceID}`
}

const buildFilterParams = (): Record<string, string> => ({
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.status !== 'all' ? { status: filters.status } : {}),
  ...(filters.payment_status !== 'all' ? { payment_status: filters.payment_status } : {}),
  ...(filters.shipping_status !== 'all' ? { shipping_status: filters.shipping_status } : {}),
  ...(filters.start_date ? { start_date: filters.start_date } : {}),
  ...(filters.end_date ? { end_date: filters.end_date } : {})
})

const fetchStats = async (): Promise<void> => {
  try {
    const response = await axios.get<OrderStats>('/api/admin/orders/stats')
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch order stats:', error)
  }
}

const fetchOrders = async (): Promise<void> => {
  loading.value = true
  try {
    const response = await axios.get<OrderListResponse>('/api/admin/orders', {
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

const fetchOrderDisputes = async (): Promise<void> => {
  disputeLoading.value = true
  try {
    const payload = await ordersApi.listDisputes({
      page: disputePagination.page,
      page_size: disputePagination.pageSize,
      provider: disputeFilters.provider || undefined,
      status: disputeFilters.status.trim() || undefined,
      search: disputeFilters.search.trim() || undefined
    })
    disputeOrders.value = payload.data
    disputePagination.page = payload.pagination.page
    disputePagination.pageSize = payload.pagination.page_size
    disputePagination.total = payload.pagination.total
  } catch (error) {
    console.error('Failed to fetch order disputes:', error)
  } finally {
    disputeLoading.value = false
  }
}

const fetchShippingLookups = async (): Promise<void> => {
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

const unwrapTrackingEvents = (response: { data?: TrackingEventsResponse | TrackingEvent[] }): TrackingEvent[] => {
  const responseData = response.data
  const payload = Array.isArray(responseData) ? responseData : responseData?.data ?? responseData ?? {}
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (Array.isArray(payload.events)) return payload.events
  return []
}

const fetchOrderTrackingEvents = async (orderID: OrderID | null | undefined): Promise<void> => {
  if (!orderID) {
    currentTrackingEvents.value = []
    return
  }

  try {
    const response = await axios.get<TrackingEventsResponse>(`/api/v1/shipping/orders/${orderID}/tracking`)
    currentTrackingEvents.value = unwrapTrackingEvents(response)
  } catch (error) {
    currentTrackingEvents.value = []
    console.error('Failed to fetch order tracking events:', error)
  }
}

const refreshOrders = (): Promise<[void, void]> => Promise.all([fetchOrders(), fetchStats()])
const applyFilters = (): void => { pagination.page = 1; void fetchOrders() }
const resetFilters = (): void => {
  Object.assign(filters, { search: '', status: 'all', payment_status: 'all', shipping_status: 'all', start_date: '', end_date: '' })
  pagination.page = 1
  fetchOrders()
}
const updatePage = (page: number): void => { pagination.page = page; void fetchOrders() }
const updatePageSize = (pageSize: number): void => { pagination.pageSize = pageSize; pagination.page = 1; void fetchOrders() }
const applyDisputeFilters = (): void => { disputePagination.page = 1; void fetchOrderDisputes() }
const resetDisputeFilters = (): void => {
  Object.assign(disputeFilters, { provider: '', status: '', search: '' })
  disputePagination.page = 1
  void fetchOrderDisputes()
}
const updateDisputePage = (page: number): void => { disputePagination.page = page; void fetchOrderDisputes() }
const updateDisputePageSize = (pageSize: number): void => {
  disputePagination.pageSize = pageSize
  disputePagination.page = 1
  void fetchOrderDisputes()
}

const fetchOrderDisputeAnalysis = async (orderID: OrderID): Promise<void> => {
  disputeAnalysisLoading.value = true
  currentDisputeAnalysis.value = null
  try {
    currentDisputeAnalysis.value = await ordersApi.getDisputeAnalysis(orderID)
  } catch (error) {
    console.error('Failed to fetch order dispute analysis:', error)
  } finally {
    disputeAnalysisLoading.value = false
  }
}

const showOrderDetail = async (order: OrderRecord): Promise<void> => {
  try {
    currentTrackingShipment.value = null
    currentDisputeAnalysis.value = null
    const [response] = await Promise.all([
      axios.get<OrderDetailResponse>(`/api/admin/orders/${order.id}`),
      fetchShippingLookups(),
      fetchOrderTrackingEvents(order.id),
      fetchOrderDisputeAnalysis(order.id)
    ])
    currentOrder.value = response.data.order || null
    currentTrackingShipment.value = response.data.tracking_shipment || null
    adminNoteForm.admin_note = currentOrder.value.admin_note || ''
    detailDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch order detail:', error)
  }
}

const showDisputeOrderDetail = (dispute: OrderDisputeCase): void => {
  if (!dispute.order_id) return
  void showOrderDetail({
    id: dispute.order_id,
    order_number: dispute.order_number || undefined
  })
}

const openDisputeContactEmail = (dispute: OrderDisputeCase): void => {
  if (!dispute.order_id || !dispute.contact_draft?.can_send) return
  disputeEmailOrderID.value = dispute.order_id
  Object.assign(disputeEmailForm, {
    provider: dispute.provider,
    dispute_id: dispute.dispute_id,
    to: dispute.contact_draft.to || dispute.customer_email || '',
    subject: dispute.contact_draft.subject || '',
    body: dispute.contact_draft.body || ''
  })
  disputeEmailMailtoUrl.value = dispute.contact_draft.mailto_url || ''
  disputeEmailDialogVisible.value = true
}

const openPaymentWorkbench = (dispute: OrderDisputeCase): void => {
  void router.push({
    name: 'PaymentRiskDisputes',
    query: {
      provider: dispute.provider,
      dispute_id: String(dispute.dispute_id)
    }
  })
}

const submitDisputeContactEmail = async (): Promise<void> => {
  const orderID = disputeEmailOrderID.value
  if (!orderID || !disputeEmailForm.dispute_id) return

  disputeEmailSending.value = true
  try {
    await ordersApi.sendDisputeContactEmail(orderID, disputeEmailForm)
    toast.success('客户联系邮件已发送')
    disputeEmailDialogVisible.value = false
    if (currentOrder.value?.id === orderID) {
      await fetchOrderDisputeAnalysis(orderID)
    }
    if (activeOrderTab.value === 'disputes') {
      await fetchOrderDisputes()
    }
  } catch (error) {
    console.error('Failed to send dispute contact email:', error)
    toast.error(error?.response?.data?.error || '客户联系邮件发送失败')
  } finally {
    disputeEmailSending.value = false
  }
}

const showStatusDialog = async (order: OrderRecord): Promise<void> => {
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

const handleStatusCarrierChange = (value: string): void => {
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

const handleStatusCarrierServiceChange = (value: string): void => {
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

const submitStatus = async (): Promise<void> => {
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

const syncCurrentOrderTracking = async (): Promise<void> => {
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

const updateAdminNote = async (): Promise<void> => {
  try {
    if (!currentOrder.value?.id) return
    await axios.patch(`/api/admin/orders/${currentOrder.value.id}/admin-note`, { admin_note: adminNoteForm.admin_note })
    if (currentOrder.value) currentOrder.value.admin_note = adminNoteForm.admin_note
    toast.success('管理员备注已保存')
  } catch (error) {
    console.error('Failed to update admin note:', error)
  }
}

const updateOrderItemCustoms = async (
  orderItemID: OrderID,
  declaredValue: number | null,
  declaredValueConfirmed: boolean,
): Promise<void> => {
  if (!currentOrder.value?.id) return

  savingCustomsItemID.value = orderItemID
  try {
    await axios.patch(
      `/api/admin/orders/${currentOrder.value.id}/items/${orderItemID}/customs`,
      {
        declared_value: declaredValue,
        declared_value_confirmed: declaredValueConfirmed,
      },
    )

    const item = currentOrder.value.items?.find((candidate) => String(candidate.id) === String(orderItemID))
    if (item) {
      item.declared_value = declaredValue
      item.declared_value_confirmed = declaredValueConfirmed && declaredValue != null
    }
    toast.success('申报价值已保存')
  } catch (error) {
    console.error('Failed to update order item customs:', error)
    toast.error(error?.response?.data?.error || '申报价值保存失败')
  } finally {
    savingCustomsItemID.value = null
  }
}

const exportOrderCustoms = async (): Promise<void> => {
  if (!currentOrder.value?.id) return

  exportingCustoms.value = true
  try {
    const response = await axios.get(`/api/admin/orders/${currentOrder.value.id}/customs-export`, {
      responseType: 'blob',
    })
    const filename = `customs_${String(currentOrder.value.order_number || currentOrder.value.id).replace(/[^\w.-]+/g, '_')}.csv`
    const url = window.URL.createObjectURL(new Blob([response.data], { type: 'text/csv;charset=utf-8' }))
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.click()
    window.URL.revokeObjectURL(url)
    toast.success('清关资料已下载')
  } catch (error) {
    console.error('Failed to export order customs:', error)
    toast.error(error?.response?.data?.error || '清关资料下载失败')
  } finally {
    exportingCustoms.value = false
  }
}

const isSelected = (orderId: OrderID): boolean => selectedOrders.value.some((order) => order.id === orderId)
const toggleAllOrders = (checked: boolean | 'indeterminate'): void => { selectedOrders.value = checked === true ? [...orders.value] : [] }
const toggleOrder = (order: OrderRecord, checked: boolean | 'indeterminate'): void => {
  if (checked === true && !isSelected(order.id)) selectedOrders.value = [...selectedOrders.value, order]
  else if (checked !== true) selectedOrders.value = selectedOrders.value.filter((selected) => selected.id !== order.id)
}

const setConfirmation = (values: Partial<OrderConfirmation>): void => {
  Object.assign(confirmation, { open: true, destructive: false, confirmLabel: '确定', ...values })
}
const requestDelete = (order: OrderRecord): void => setConfirmation({
  type: 'delete', target: order, title: '删除订单？', description: `订单 ${order.order_number} 将被永久删除，此操作不可恢复。`, confirmLabel: '删除', destructive: true
})
const requestBatchStatus = (status: string): void => {
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

const executeConfirmedAction = async (): Promise<void> => {
  const { type, target, status } = confirmation
  confirmation.open = false
  try {
    if (type === 'delete') {
      if (!target || Array.isArray(target)) return
      await axios.delete(`/api/admin/orders/${target.id}`)
      toast.success('订单已删除')
    } else if (type === 'batch-status') {
      if (!target || !Array.isArray(target)) return
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

const exportOrders = async (): Promise<void> => {
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

const refreshActiveOrderTab = (): void => {
  if (activeOrderTab.value === 'disputes') {
    void fetchOrderDisputes()
    return
  }
  void refreshOrders()
}

onMounted(() => {
  refreshActiveOrderTab()
  void fetchShippingLookups()
})

watch(activeOrderTab, () => {
  refreshActiveOrderTab()
})
</script>
