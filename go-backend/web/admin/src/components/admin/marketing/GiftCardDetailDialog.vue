<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <GiftCardDetailPanel
      :status-update="statusUpdate"
      :current-gift-card="currentGiftCard"
      :loading="loading"
      :transactions="transactions"
      :status-submitting="statusSubmitting"
      :can-edit="canEdit"
      :format-currency="formatCurrency"
      :format-date="formatDate"
      :gift-card-status-name="giftCardStatusName"
      :gift-card-status-tone="giftCardStatusTone"
      :gift-card-status-options="giftCardStatusOptions"
      :transaction-type-name="transactionTypeName"
      @update:status-update="emit('update:statusUpdate', $event)"
      @update-status="emit('update-status')"
    />
  </Dialog>
</template>

<script setup>
import GiftCardDetailPanel from '@/components/admin/marketing/GiftCardDetailPanel.vue'
import { Dialog } from '@/components/ui/dialog'

defineProps({
  open: { type: Boolean, default: false },
  statusUpdate: { type: String, default: 'active' },
  currentGiftCard: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  transactions: { type: Array, default: () => [] },
  statusSubmitting: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  formatCurrency: { type: Function, required: true },
  formatDate: { type: Function, required: true },
  giftCardStatusName: { type: Function, required: true },
  giftCardStatusTone: { type: Function, required: true },
  giftCardStatusOptions: { type: Function, required: true },
  transactionTypeName: { type: Function, required: true },
})

const emit = defineEmits(['update:open', 'update:statusUpdate', 'update-status'])
</script>
