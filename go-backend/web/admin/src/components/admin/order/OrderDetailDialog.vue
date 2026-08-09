<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <OrderDetailPanel
      :admin-note="adminNote"
      :current-order="currentOrder"
      :current-tracking-events="currentTrackingEvents"
      :current-tracking-shipment="currentTrackingShipment"
      :syncing-tracking="syncingTracking"
      :can-edit="canEdit"
      :order-status-name="orderStatusName"
      :order-status-tone="orderStatusTone"
      :payment-status-name="paymentStatusName"
      :payment-status-tone="paymentStatusTone"
      :shipping-status-name="shippingStatusName"
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
      @update:admin-note="emit('update:adminNote', $event)"
      @sync-tracking="emit('sync-tracking')"
      @update-note="emit('update-note')"
    />
  </Dialog>
</template>

<script setup lang="ts">
import OrderDetailPanel from '@/components/admin/order/OrderDetailPanel.vue'
import { Dialog } from '@/components/ui/dialog'
import type {
  OrderCarrierLabelResolver,
  OrderDateFormatter,
  OrderMoneyFormatter,
  OrderRecord,
  OrderShippingAddressLineResolver,
  OrderShippingNameResolver,
  OrderStatusNameResolver,
  OrderStatusToneResolver,
  TrackingEvent,
  TrackingShipment
} from './orderTypes'

withDefaults(defineProps<{
  open?: boolean
  adminNote?: string
  currentOrder?: OrderRecord | null
  currentTrackingEvents?: TrackingEvent[]
  currentTrackingShipment?: TrackingShipment | null
  syncingTracking?: boolean
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
  open: false,
  adminNote: '',
  currentOrder: null,
  currentTrackingEvents: () => [],
  currentTrackingShipment: null,
  syncingTracking: false,
  canEdit: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:adminNote', value: string): void
  (event: 'sync-tracking'): void
  (event: 'update-note'): void
}>()
</script>
