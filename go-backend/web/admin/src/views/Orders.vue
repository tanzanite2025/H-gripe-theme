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

    <OrderFilterPanel
      :filters="filters"
      :order-status-options="orderStatusOptions"
      :payment-status-options="paymentStatusOptions"
      :shipping-status-options="shippingStatusOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <OrderTablePanel
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

    <OrderDetailDialog
      v-model:open="detailDialogVisible"
      v-model:admin-note="adminNoteForm.admin_note"
      :current-order="currentOrder"
      :current-tracking-events="currentTrackingEvents"
      :current-tracking-shipment="currentTrackingShipment"
      :syncing-tracking="syncingTracking"
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

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Banknote,
  CalendarCheck2,
  Download,
  ShoppingBag,
  TrendingUp
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import OrderDetailDialog from '@/components/admin/order/OrderDetailDialog.vue'
import OrderFilterPanel from '@/components/admin/order/OrderFilterPanel.vue'
import OrderStatusDialog from '@/components/admin/order/OrderStatusDialog.vue'
import OrderTablePanel from '@/components/admin/order/OrderTablePanel.vue'
import { Button } from '@/components/ui/button'
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
