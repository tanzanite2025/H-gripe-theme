<template>
  <div class="h-full overflow-y-auto px-1 md:p-6">
    <div v-if="isLoadingOrders" class="text-center tz-text-secondary py-10 md:py-12 text-sm">
      {{ t('chatModal.orders.loading', 'Loading orders...') }}
    </div>
    <div v-else-if="ordersList.length > 0" class="space-y-2 md:space-y-0 md:grid md:grid-cols-2 md:gap-3">
      <div
        v-for="order in ordersList"
        :key="order.order_number"
        class="border tz-border-subtle md:tz-border-subtle rounded-2xl md:rounded-lg p-3 tz-surface-card transition-colors"
      >
        <div class="flex items-center justify-between mb-1 md:mb-2">
          <span class="tz-text-primary text-sm font-semibold md:font-medium">
            {{ order.order_number ? `Order #${order.order_number}` : 'Order' }}
          </span>
          <span class="tz-micro-label md:text-xs px-2 py-0.5 rounded-full tz-surface-subtle md:tz-surface-subtle tz-text-secondary">
            {{ order.status || 'Processing' }}
          </span>
        </div>
        <p class="tz-text-secondary text-xs">{{ order.total }} {{ order.currency || '' }}</p>
        <p v-if="order.item_count" class="tz-caption md:text-xs tz-text-secondary mt-1">
          {{ order.item_count }} {{ Number(order.item_count) > 1 ? t('chatModal.orders.items', 'items') : t('chatModal.orders.item', 'item') }}
        </p>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <span class="tz-micro-label rounded-full tz-surface-subtle px-2 py-1 tz-text-secondary">
            {{ fulfillmentModeLabel(order) }}
          </span>
          <span
            v-if="isMadeToOrderOrder(order)"
            class="tz-micro-label rounded-full border border-amber-300/30 bg-amber-300/10 px-2 py-1 text-amber-200"
          >
            {{ productionStatusLabel(order.production_status) }}
          </span>
          <span
            v-if="order.signature_required"
            class="tz-micro-label rounded-full border border-emerald-300/30 bg-emerald-300/10 px-2 py-1 text-emerald-200"
          >
            {{ t('chatModal.orders.signatureRequired', 'Signature required at delivery') }}
          </span>
        </div>
        <div v-if="isMadeToOrderOrder(order)" class="mt-2 space-y-1 tz-caption md:text-xs tz-text-secondary">
          <p v-if="order.production_started_at">
            {{ t('chatModal.orders.productionStartedAt', 'Production started') }}: {{ formatOrderDate(order.production_started_at) }}
          </p>
          <p v-if="order.production_completed_at">
            {{ t('chatModal.orders.productionCompletedAt', 'Production completed') }}: {{ formatOrderDate(order.production_completed_at) }}
          </p>
        </div>
        <div v-if="order.items?.length" class="mt-2 space-y-1">
          <div
            v-for="item in order.items.slice(0, 2)"
            :key="item.id || `${item.product_id}-${item.sku}`"
            class="flex items-center justify-between gap-2 rounded-xl tz-surface-subtle px-2 py-1 tz-caption tz-text-secondary"
          >
            <span class="truncate">{{ item.product_name || item.title || 'Product' }}</span>
            <span class="shrink-0">x{{ item.quantity || 1 }}</span>
          </div>
          <p v-if="order.items.length > 2" class="tz-micro-label tz-text-muted">
            +{{ order.items.length - 2 }} more
          </p>
        </div>
        <p class="tz-text-muted tz-caption md:text-xs mt-1">{{ order.date }}</p>
        <button
          type="button"
          class="mt-3 w-full rounded-full border tz-border-strong/40 bg-white px-3 py-2 tz-caption font-semibold text-slate-950 transition-colors hover:bg-white/90"
          @click="$emit('shareOrder', order)"
        >
          {{ t('chatModal.orders.confirm', 'Confirm order with support') }}
        </button>
      </div>
    </div>
    <div v-else class="text-center tz-text-secondary text-sm py-10 md:py-12">
      {{ t('chatModal.orders.empty', 'No orders yet') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import {
  isMadeToOrderFulfillment,
  normalizeFulfillmentMode,
  normalizeProductionStatus,
} from '~/utils/fulfillmentPresentation'

const { t, locale } = useI18n()

const orderFulfillmentMode = (order: Record<string, any>) => {
  if (order.fulfillment_mode) return normalizeFulfillmentMode(order.fulfillment_mode)
  return Array.isArray(order.items) && order.items.some((item: Record<string, any>) => (
    isMadeToOrderFulfillment(item.fulfillment_mode)
  ))
    ? 'made_to_order'
    : 'stock'
}

const isMadeToOrderOrder = (order: Record<string, any>) => (
  isMadeToOrderFulfillment(orderFulfillmentMode(order))
)

const fulfillmentModeLabel = (order: Record<string, any>) => {
  const mode = orderFulfillmentMode(order)
  if (mode === 'mixed') return t('chatModal.orders.mixed', 'Mixed fulfillment')
  if (mode === 'made_to_order') return t('chatModal.orders.madeToOrder', 'Made to order')
  return t('chatModal.orders.stock', 'In stock')
}

const productionStatusLabel = (value: unknown) => {
  const status = normalizeProductionStatus(value)
  const keyByStatus = {
    not_applicable: 'notApplicable',
    not_started: 'notStarted',
    started: 'started',
    completed: 'completed',
    cancelled: 'cancelled',
  } as const
  return t(
    `chatModal.orders.productionStatus.${keyByStatus[status]}`,
    status.replace(/_/g, ' '),
  )
}

const formatOrderDate = (value: unknown) => {
  const date = new Date(String(value || ''))
  if (Number.isNaN(date.getTime())) return String(value || '')
  try {
    return new Intl.DateTimeFormat(String(locale.value || 'en').replace('_', '-'), {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(date)
  } catch {
    return date.toLocaleString()
  }
}

defineProps<{
  ordersList: any[]
  isLoadingOrders: boolean
}>()

defineEmits<{
  'shareOrder': [order: any]
}>()
</script>
