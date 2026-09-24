<template>
  <section class="account-tab-panel orders-tab" aria-live="polite">
    <div class="tab-head">
      <div>
        <p>{{ t('accountSidebar.orders.eyebrow', 'Purchase history') }}</p>
        <h3>{{ t('accountSidebar.orders.title', 'My orders') }}</h3>
      </div>
      <button
        type="button"
        class="orders-icon-button"
        :aria-label="t('accountSidebar.actions.refresh', 'Refresh')"
        :title="t('accountSidebar.actions.refresh', 'Refresh')"
        :disabled="loading"
        @click="loadOrders"
      >
        <Icon name="lucide:refresh-cw" :class="{ 'orders-spin': loading }" />
      </button>
    </div>

    <div v-if="loading" class="account-loading" role="status">
      <Icon name="lucide:loader-circle" class="orders-spin" />
      {{ t('accountSidebar.orders.loading', 'Loading orders...') }}
    </div>

    <div v-else-if="error" class="orders-state orders-state--error" role="alert">
      <Icon name="lucide:circle-alert" />
      <span>{{ error }}</span>
      <button type="button" class="orders-text-button" @click="loadOrders">
        {{ t('accountSidebar.actions.retry', 'Retry') }}
      </button>
    </div>

    <div v-else-if="!orders.length" class="account-empty">
      <Icon name="lucide:receipt-text" />
      <strong>{{ t('accountSidebar.orders.emptyTitle', 'No orders yet') }}</strong>
      <span>{{ t('accountSidebar.orders.emptyDescription', 'Completed purchases will appear here with their delivery details.') }}</span>
      <NuxtLink :to="localePath('/shop')" @click="emit('close')">
        {{ t('accountSidebar.actions.goShop', 'Go to shop') }}
      </NuxtLink>
    </div>

    <template v-else>
      <div class="orders-list">
        <article v-for="order in orders" :key="order.order_number" class="order-card">
          <button
            type="button"
            class="order-card__toggle"
            :aria-expanded="expandedOrderNumber === order.order_number"
            @click="toggleOrder(order.order_number)"
          >
            <span class="order-card__main">
              <strong>{{ order.order_number }}</strong>
              <span>{{ formatDate(order.created_at) }} · {{ itemCount(order) }} {{ itemCount(order) === 1 ? t('accountSidebar.orders.item', 'item') : t('accountSidebar.orders.items', 'items') }}</span>
            </span>
            <span class="order-card__aside">
              <span class="order-status" :class="`order-status--${statusClass(order)}`">{{ statusLabel(order) }}</span>
              <strong>{{ formatMinorMoney(order.total_minor, order.currency) }}</strong>
              <Icon :name="expandedOrderNumber === order.order_number ? 'lucide:chevron-up' : 'lucide:chevron-down'" />
            </span>
          </button>

          <div v-if="expandedOrderNumber === order.order_number" class="order-card__detail">
            <div v-if="detailLoading === order.order_number" class="account-loading" role="status">
              <Icon name="lucide:loader-circle" class="orders-spin" />
              {{ t('accountSidebar.orders.loadingDetail', 'Loading order details...') }}
            </div>

            <div v-else-if="detailError" class="orders-state orders-state--error" role="alert">
              <span>{{ detailError }}</span>
              <button type="button" class="orders-text-button" @click="loadDetail(order.order_number)">{{ t('accountSidebar.actions.retry', 'Retry') }}</button>
            </div>

            <template v-else-if="selectedOrder">
              <div class="order-detail-grid">
                <div>
                  <span class="order-detail-label">{{ t('accountSidebar.orders.shippingAddress', 'Shipping address') }}</span>
                  <address>{{ formatAddress(selectedOrder.shipping_address) }}</address>
                </div>
                <div>
                  <span class="order-detail-label">{{ t('accountSidebar.orders.delivery', 'Delivery') }}</span>
                  <p>{{ selectedOrder.shipping_method || '—' }}<br>{{ selectedOrder.shipping_status || selectedOrder.status || '—' }}</p>
                </div>
              </div>

              <ul class="order-items">
                <li v-for="(item, itemIndex) in selectedOrder.items || []" :key="`${item.product_id}-${item.variant_id || 'base'}-${item.product_name}-${itemIndex}`">
                  <span>{{ item.product_name || item.sku || t('accountSidebar.orders.product', 'Product') }} x {{ item.quantity }}</span>
                  <strong>{{ formatMinorMoney(item.total_minor, item.currency || selectedOrder.currency) }}</strong>
                </li>
              </ul>

              <div v-if="selectedOrder.tracking_shipments?.length" class="tracking-list">
                <span class="order-detail-label">{{ t('accountSidebar.orders.tracking', 'Tracking') }}</span>
                <div v-for="shipment in selectedOrder.tracking_shipments" :key="shipment.tracking_number" class="tracking-card">
                  <div class="tracking-card__head">
                    <strong>{{ shipment.tracking_number }}</strong>
                    <button type="button" class="orders-text-button" :disabled="trackingLoading === shipment.tracking_number" @click="loadTracking(shipment.tracking_number)">
                      {{ trackingLoading === shipment.tracking_number ? t('accountSidebar.orders.loadingTracking', 'Loading...') : t('accountSidebar.orders.viewTracking', 'View updates') }}
                    </button>
                  </div>
                  <ol v-if="trackingEvents[shipment.tracking_number]?.length" class="tracking-events">
                    <li v-for="event in trackingEvents[shipment.tracking_number]" :key="`${event.event_time}-${event.status}`">
                      <strong>{{ event.status || t('accountSidebar.orders.trackingUpdate', 'Tracking update') }}</strong>
                      <span>{{ event.description || event.location || '—' }}</span>
                      <time :datetime="event.event_time">{{ formatDate(event.event_time) }}</time>
                    </li>
                  </ol>
                  <p v-else-if="trackingErrors[shipment.tracking_number]" class="tracking-muted">{{ trackingErrors[shipment.tracking_number] }}</p>
                </div>
              </div>

              <section class="after-sales">
                <div class="after-sales__head">
                  <span class="order-detail-label">{{ t('accountSidebar.orders.afterSales', 'After-sales support') }}</span>
                  <button type="button" class="orders-text-button" :disabled="afterSalesLoading" @click="loadAfterSales(selectedOrder.order_number)">
                    {{ afterSalesLoading ? t('accountSidebar.orders.loadingAfterSales', 'Loading...') : t('accountSidebar.orders.refreshAfterSales', 'Refresh') }}
                  </button>
                </div>
                <p v-if="afterSalesError" class="orders-state orders-state--error">{{ afterSalesError }}</p>
                <div v-for="(record, recordIndex) in afterSalesCases" :key="`${record.created_at || 'case'}-${record.status || 'unknown'}-${recordIndex}`" class="after-sales__case">
                  <div class="after-sales__case-head"><strong>{{ record.reason }}</strong><span>{{ record.status }}</span></div>
                  <p>{{ record.description }}</p>
                  <p v-if="record.resolution"><strong>{{ t('accountSidebar.orders.resolution', 'Resolution') }}:</strong> {{ record.resolution }}</p>
                  <div v-for="shipment in record.return_shipments || []" :key="shipment.tracking_number" class="after-sales__shipment">
                    <strong>{{ t('accountSidebar.orders.returnTracking', 'Return tracking') }}</strong>
                    <a v-if="shipment.tracking_url" :href="shipment.tracking_url" target="_blank" rel="noopener">{{ shipment.tracking_number }}</a>
                    <span v-else>{{ shipment.tracking_number }}</span>
                  </div>
                  <ol v-if="record.events?.length" class="after-sales__events">
                    <li v-for="event in record.events" :key="`${event.created_at}-${event.to_status}`"><strong>{{ event.to_status }}</strong><span>{{ event.resolution }}</span><time :datetime="event.created_at">{{ formatDate(event.created_at) }}</time></li>
                  </ol>
                  <time v-if="record.created_at" :datetime="record.created_at">{{ formatDate(record.created_at) }}</time>
                </div>
                <form v-if="!afterSalesCases.length" class="after-sales__form" @submit.prevent="submitAfterSales(selectedOrder)">
                  <label>{{ t('accountSidebar.orders.afterSalesItem', 'Item') }}<select v-model.number="afterSalesItemIndex" required><option v-for="(item, index) in selectedOrder.items || []" :key="index" :value="index">{{ item.product_name || item.sku || `Item ${index + 1}` }}</option></select></label>
                  <label>{{ t('accountSidebar.orders.quantity', 'Quantity') }}<input v-model.number="afterSalesQuantity" type="number" min="1" :max="selectedOrder.items?.[afterSalesItemIndex]?.quantity || 1" required></label>
                  <label>{{ t('accountSidebar.orders.reason', 'Reason') }}<input v-model.trim="afterSalesReason" type="text" maxlength="120" required></label>
                  <label>{{ t('accountSidebar.orders.description', 'Description') }}<textarea v-model.trim="afterSalesDescription" rows="4" required /></label>
                  <button type="submit" class="after-sales__submit" :disabled="afterSalesSubmitting">{{ afterSalesSubmitting ? t('accountSidebar.orders.submitting', 'Submitting...') : t('accountSidebar.orders.submitRequest', 'Submit request') }}</button>
                </form>
              </section>

              <div class="order-detail-actions">
                <button v-if="canCancel(selectedOrder)" type="button" class="orders-danger-button" :disabled="cancelling" @click="cancelOrder(selectedOrder.order_number)">
                  <Icon name="lucide:x-circle" />
                  {{ cancelling ? t('accountSidebar.orders.cancelling', 'Cancelling...') : t('accountSidebar.orders.cancel', 'Cancel order') }}
                </button>
              </div>
            </template>
          </div>
        </article>
      </div>

      <div v-if="totalPages > 1" class="orders-pagination">
        <button type="button" class="orders-icon-button" :disabled="page <= 1 || loading" :aria-label="t('accountSidebar.orders.previous', 'Previous page')" @click="changePage(page - 1)">
          <Icon name="lucide:chevron-left" />
        </button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button type="button" class="orders-icon-button" :disabled="page >= totalPages || loading" :aria-label="t('accountSidebar.orders.next', 'Next page')" @click="changePage(page + 1)">
          <Icon name="lucide:chevron-right" />
        </button>
      </div>
    </template>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { ApiRequestError } from '~/composables/useApiRequest'
import { useAuth } from '~/composables/useAuth'
import { formatMinorMoney } from '~/utils/money'

interface OrderItem {
  product_id?: number
  variant_id?: number | null
  product_name?: string
  sku?: string
  quantity?: number
  total_minor?: number
  currency?: string
}

interface TrackingShipment {
  tracking_number: string
  provider_carrier_code?: string
}

interface Order {
  order_number: string
  status?: string
  payment_status?: string
  shipping_status?: string
  shipping_method?: string
  currency?: string
  total_minor?: number
  created_at?: string
  shipping_address?: Record<string, string>
  items?: OrderItem[]
  tracking_shipments?: TrackingShipment[]
}

interface TrackingEvent {
  status?: string
  location?: string
  description?: string
  event_time: string
}

interface AfterSalesCase {
  status?: string
  reason?: string
  description?: string
  resolution?: string
  events?: Array<{ from_status?: string; to_status?: string; resolution?: string; created_at?: string }>
  return_shipments?: Array<{ tracking_number?: string; tracking_url?: string; carrier?: string }>
  created_at?: string
}

const props = defineProps<{ active?: boolean }>()
const emit = defineEmits<{ (event: 'close'): void }>()
const { t } = useI18n()
const localePath = useLocalePath()
const auth = useAuth()

const orders = ref<Order[]>([])
const page = ref(1)
const pageSize = 10
const total = ref(0)
const loading = ref(false)
const error = ref('')
const expandedOrderNumber = ref('')
const selectedOrder = ref<Order | null>(null)
const detailLoading = ref('')
const detailError = ref('')
let detailRequestId = 0
const cancelling = ref(false)
const afterSalesCases = ref<AfterSalesCase[]>([])
const afterSalesLoading = ref(false)
const afterSalesSubmitting = ref(false)
const afterSalesError = ref('')
const afterSalesItemIndex = ref(0)
const afterSalesQuantity = ref(1)
const afterSalesReason = ref('')
const afterSalesDescription = ref('')
const trackingLoading = ref('')
const trackingEvents = ref<Record<string, TrackingEvent[]>>({})
const trackingErrors = ref<Record<string, string>>({})

const totalPages = computed(() => Math.max(1, Math.ceil(total.value / pageSize)))

const unwrap = <T>(payload: T | { data?: T } | null | undefined): T | null => {
  if (payload && typeof payload === 'object' && 'data' in payload && (payload as { data?: T }).data !== undefined) {
    return (payload as { data: T }).data
  }
  return (payload as T) || null
}

const loadOrders = async () => {
  if (!auth.isAuthenticated.value) return
  loading.value = true
  error.value = ''
  try {
    const response = await auth.request<Order[] | { data?: Order[]; pagination?: { total?: number } }>(`/orders?page=${page.value}&page_size=${pageSize}`)
    const payload = unwrap<Order[]>(response)
    orders.value = Array.isArray(payload) ? payload : []
    const envelope = response && typeof response === 'object' && !Array.isArray(response) ? response : null
    total.value = Number(envelope?.pagination?.total || orders.value.length)
  } catch (err) {
    error.value = err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.loadFailed', 'Orders could not be loaded.')
  } finally {
    loading.value = false
  }
}

const loadDetail = async (orderNumber: string) => {
  const requestId = ++detailRequestId
  detailLoading.value = orderNumber
  detailError.value = ''
  try {
    const response = await auth.request<Order | { data?: Order }>(`/orders/${encodeURIComponent(orderNumber)}`)
    if (requestId === detailRequestId) {
      selectedOrder.value = unwrap<Order>(response)
      await loadAfterSales(orderNumber)
    }
  } catch (err) {
    if (requestId === detailRequestId) {
      detailError.value = err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.detailFailed', 'Order details could not be loaded.')
    }
  } finally {
    if (requestId === detailRequestId) {
      detailLoading.value = ''
    }
  }
}

const toggleOrder = async (orderNumber: string) => {
  if (expandedOrderNumber.value === orderNumber) {
    detailRequestId += 1
    expandedOrderNumber.value = ''
    selectedOrder.value = null
    detailLoading.value = ''
    detailError.value = ''
    return
  }
  expandedOrderNumber.value = orderNumber
  selectedOrder.value = null
  await loadDetail(orderNumber)
}

const loadAfterSales = async (orderNumber: string) => {
  afterSalesLoading.value = true
  afterSalesError.value = ''
  try {
    const response = await auth.request<{ cases?: AfterSalesCase[] } | { data?: { cases?: AfterSalesCase[] } }>(`/orders/${encodeURIComponent(orderNumber)}/after-sales`)
    const payload = unwrap<{ cases?: AfterSalesCase[] }>(response)
    afterSalesCases.value = payload?.cases || []
  } catch (err) {
    afterSalesCases.value = []
    afterSalesError.value = err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.afterSalesLoadFailed', 'After-sales requests could not be loaded.')
  } finally {
    afterSalesLoading.value = false
  }
}

const submitAfterSales = async (order: Order) => {
  afterSalesSubmitting.value = true
  afterSalesError.value = ''
  try {
    const body = new FormData()
    body.append('reason', afterSalesReason.value)
    body.append('description', afterSalesDescription.value)
    body.append('items', JSON.stringify([{ item_index: afterSalesItemIndex.value, quantity: afterSalesQuantity.value }]))
    await auth.request(`/orders/${encodeURIComponent(order.order_number)}/after-sales`, { method: 'POST', body })
    afterSalesReason.value = ''
    afterSalesDescription.value = ''
    afterSalesQuantity.value = 1
    await loadAfterSales(order.order_number)
  } catch (err) {
    afterSalesError.value = err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.afterSalesSubmitFailed', 'The after-sales request could not be submitted.')
  } finally {
    afterSalesSubmitting.value = false
  }
}

const loadTracking = async (trackingNumber: string) => {
  trackingLoading.value = trackingNumber
  trackingErrors.value = { ...trackingErrors.value, [trackingNumber]: '' }
  try {
    const response = await auth.request<TrackingEvent[] | { data?: TrackingEvent[] | { data?: TrackingEvent[] } }>(`/shipping/track/${encodeURIComponent(trackingNumber)}`)
    const firstPayload = unwrap<TrackingEvent[] | { data?: TrackingEvent[] }>(response)
    const events = Array.isArray(firstPayload) ? firstPayload : (firstPayload?.data || [])
    trackingEvents.value = { ...trackingEvents.value, [trackingNumber]: Array.isArray(events) ? events : [] }
  } catch (err) {
    trackingErrors.value = {
      ...trackingErrors.value,
      [trackingNumber]: err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.trackingFailed', 'Tracking updates are not available yet.'),
    }
  } finally {
    trackingLoading.value = ''
  }
}

const cancelOrder = async (orderNumber: string) => {
  if (!window.confirm(t('accountSidebar.orders.cancelConfirm', 'Cancel this order?'))) return
  cancelling.value = true
  try {
    await auth.request(`/orders/${encodeURIComponent(orderNumber)}/cancel`, { method: 'POST' })
    await loadOrders()
    await loadDetail(orderNumber)
  } catch (err) {
    detailError.value = err instanceof ApiRequestError ? err.message : t('accountSidebar.orders.cancelFailed', 'This order could not be cancelled.')
  } finally {
    cancelling.value = false
  }
}

const changePage = async (nextPage: number) => {
  if (nextPage < 1 || nextPage > totalPages.value) return
  page.value = nextPage
  detailRequestId += 1
  expandedOrderNumber.value = ''
  selectedOrder.value = null
  detailLoading.value = ''
  detailError.value = ''
  await loadOrders()
}

const itemCount = (order: Order) => (order.items || []).reduce((count, item) => count + Number(item.quantity || 0), 0)
const canCancel = (order: Order) => String(order.status || '').toLowerCase() === 'pending' && String(order.payment_status || '').toLowerCase() === 'unpaid'
const statusClass = (order: Order) => {
  const orderStatus = String(order.status || '').toLowerCase()
  const terminalStatuses = new Set(['cancelled', 'payment_expired', 'refunded', 'disputed'])
  const status = terminalStatuses.has(orderStatus) ? orderStatus : (order.shipping_status || orderStatus || 'pending')
  return String(status).toLowerCase().replace(/[^a-z0-9_-]/g, '-')
}
const statusLabel = (order: Order) => t(`accountSidebar.orders.status.${statusClass(order)}`, order.shipping_status || order.status || t('accountSidebar.orders.status.pending', 'Pending'))
const formatDate = (value?: string) => {
  if (!value) return '—'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat(undefined, { year: 'numeric', month: 'short', day: 'numeric' }).format(date)
}
const formatAddress = (address?: Record<string, string>) => {
  if (!address) return '—'
  return [
    [address.first_name, address.last_name].filter(Boolean).join(' '),
    address.address1 || address.address_1,
    address.address2 || address.address_2,
    [address.city, address.state, address.postal_code].filter(Boolean).join(', '),
    address.country,
  ].filter(Boolean).join(', ') || '—'
}

watch(() => props.active, (active) => {
  if (active && !orders.value.length && !loading.value) loadOrders()
}, { immediate: true })
</script>

<style scoped>
.account-tab-panel { display: flex; min-height: 0; flex-direction: column; gap: 0.85rem; }
.orders-tab { min-width: 0; }
.tab-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 0.75rem; }
.tab-head p, .tab-head h3 { margin: 0; }
.tab-head p { color: rgba(5, 150, 105, 0.9); font-size: var(--tz-type-micro-label); font-weight: 800; letter-spacing: 0.15em; text-transform: uppercase; }
.tab-head h3 { margin-top: 0.18rem; color: var(--tz-text-primary); font-size: 1.05rem; font-weight: 850; }
.account-loading, .account-empty { display: flex; min-height: 12rem; flex-direction: column; align-items: center; justify-content: center; gap: 0.55rem; border-radius: 1.25rem; background: var(--tz-surface-card); padding: 1rem; text-align: center; }
.account-loading, .account-empty span { color: var(--tz-text-secondary); font-size: 0.82rem; }
.account-loading :deep(svg) { width: 1.4rem; height: 1.4rem; }
.account-empty :deep(svg) { width: 2.2rem; height: 2.2rem; color: rgba(5, 150, 105, 0.72); }
.account-empty strong { color: var(--tz-text-primary); font-size: 0.95rem; }
.account-empty a { margin-top: 0.3rem; border-radius: 999px; background: var(--tz-action-primary); color: #ffffff; padding: 0.65rem 1rem; font-size: 0.8rem; font-weight: 850; text-decoration: none; }
.orders-icon-button { display: inline-flex; width: 2.25rem; height: 2.25rem; align-items: center; justify-content: center; border: 1px solid var(--tz-border-subtle); border-radius: 0.7rem; background: var(--tz-input-surface); color: var(--tz-text-secondary); }
.orders-icon-button:disabled { cursor: not-allowed; opacity: 0.5; }
.orders-icon-button :deep(svg) { width: 1rem; height: 1rem; }
.orders-spin { animation: orders-spin 0.85s linear infinite; }
.orders-list { display: grid; gap: 0.6rem; }
.order-card { overflow: hidden; border: 1px solid var(--tz-border-subtle); border-radius: 0.95rem; background: var(--tz-card-surface); }
.order-card__toggle { display: flex; width: 100%; align-items: center; justify-content: space-between; gap: 0.7rem; border: 0; background: transparent; padding: 0.8rem; text-align: left; color: inherit; }
.order-card__main, .order-card__aside { display: flex; min-width: 0; flex-direction: column; gap: 0.25rem; }
.order-card__main strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--tz-text-primary); font-size: 0.82rem; }
.order-card__main span { color: var(--tz-text-muted); font-size: 0.7rem; }
.order-card__aside { align-items: flex-end; flex: 0 0 auto; }
.order-card__aside > strong { color: var(--tz-text-primary); font-size: 0.78rem; }
.order-card__aside > :deep(svg) { width: 0.85rem; height: 0.85rem; color: var(--tz-text-muted); }
.order-status { border-radius: 999px; padding: 0.18rem 0.42rem; background: var(--tz-input-surface); color: var(--tz-text-secondary); font-size: 0.62rem; font-weight: 800; text-transform: capitalize; }
.order-status--shipped, .order-status--delivered, .order-status--completed { background: #dcfce7; color: #166534; }
.order-status--cancelled, .order-status--payment_expired { background: #fee2e2; color: #991b1b; }
.order-card__detail { display: grid; gap: 0.75rem; border-top: 1px solid var(--tz-border-subtle); padding: 0.8rem; }
.order-detail-grid { display: grid; gap: 0.7rem; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.order-detail-grid address, .order-detail-grid p { margin: 0.25rem 0 0; color: var(--tz-text-secondary); font-size: 0.74rem; font-style: normal; line-height: 1.45; }
.order-detail-label { display: block; color: var(--tz-text-muted); font-size: 0.64rem; font-weight: 800; text-transform: uppercase; letter-spacing: 0.08em; }
.order-items, .tracking-events { display: grid; gap: 0.4rem; margin: 0; padding: 0; list-style: none; }
.order-items li { display: flex; justify-content: space-between; gap: 0.6rem; color: var(--tz-text-secondary); font-size: 0.73rem; }
.order-items strong { color: var(--tz-text-primary); white-space: nowrap; }
.tracking-list { display: grid; gap: 0.45rem; }
.tracking-card { display: grid; gap: 0.45rem; border-radius: 0.7rem; background: var(--tz-input-surface); padding: 0.65rem; }
.tracking-card__head { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; }
.tracking-card__head strong { color: var(--tz-text-primary); font-size: 0.74rem; }
.tracking-events li { display: grid; grid-template-columns: auto 1fr auto; gap: 0.4rem; align-items: baseline; color: var(--tz-text-secondary); font-size: 0.68rem; }
.tracking-events strong { color: var(--tz-text-primary); }
.tracking-events time { color: var(--tz-text-muted); white-space: nowrap; }
.tracking-muted { margin: 0; color: var(--tz-text-muted); font-size: 0.7rem; }
.after-sales { display: grid; gap: 0.6rem; border-top: 1px solid var(--tz-border-subtle); padding-top: 0.75rem; }
.after-sales__head, .after-sales__case-head { display: flex; align-items: center; justify-content: space-between; gap: 0.6rem; }
.after-sales__case { display: grid; gap: 0.3rem; border-radius: 0.7rem; background: var(--tz-input-surface); padding: 0.65rem; color: var(--tz-text-secondary); font-size: 0.72rem; }
.after-sales__case p { margin: 0; }
.after-sales__case time { color: var(--tz-text-muted); font-size: 0.66rem; }
.after-sales__form { display: grid; gap: 0.55rem; }
.after-sales__form label { display: grid; gap: 0.25rem; color: var(--tz-text-secondary); font-size: 0.7rem; font-weight: 700; }
.after-sales__form input, .after-sales__form select, .after-sales__form textarea { width: 100%; border: 1px solid var(--tz-border-subtle); border-radius: 0.55rem; background: var(--tz-input-surface); color: var(--tz-text-primary); padding: 0.5rem; font: inherit; }
.after-sales__submit { justify-self: end; border: 0; border-radius: 0.6rem; background: var(--tz-action-primary); color: #fff; padding: 0.5rem 0.75rem; font-size: 0.72rem; font-weight: 800; }
.after-sales__submit:disabled { opacity: 0.5; }
.orders-text-button { border: 0; background: transparent; color: #047857; padding: 0; font-size: 0.7rem; font-weight: 800; }
.orders-text-button:disabled { opacity: 0.5; }
.order-detail-actions { display: flex; justify-content: flex-end; }
.orders-danger-button { display: inline-flex; align-items: center; gap: 0.35rem; border: 1px solid #fecaca; border-radius: 0.65rem; background: #fff1f2; color: #b91c1c; padding: 0.45rem 0.65rem; font-size: 0.7rem; font-weight: 800; }
.orders-danger-button :deep(svg) { width: 0.85rem; height: 0.85rem; }
.orders-state { display: flex; align-items: center; gap: 0.5rem; color: var(--tz-text-secondary); font-size: 0.75rem; }
.orders-state--error { color: #b91c1c; }
.orders-pagination { display: flex; align-items: center; justify-content: center; gap: 0.65rem; color: var(--tz-text-muted); font-size: 0.72rem; }
@keyframes orders-spin { to { transform: rotate(360deg); } }
@media (max-width: 520px) { .order-detail-grid { grid-template-columns: 1fr; } .tracking-events li { grid-template-columns: 1fr; gap: 0.15rem; } }
</style>
