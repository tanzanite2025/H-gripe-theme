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

<script setup>
import OrderDetailPanel from '@/components/admin/order/OrderDetailPanel.vue'
import { Dialog } from '@/components/ui/dialog'

defineProps({
  open: { type: Boolean, default: false },
  adminNote: { type: String, default: '' },
  currentOrder: { type: Object, default: null },
  currentTrackingEvents: { type: Array, default: () => [] },
  currentTrackingShipment: { type: Object, default: null },
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

const emit = defineEmits(['update:open', 'update:adminNote', 'sync-tracking', 'update-note'])
</script>
