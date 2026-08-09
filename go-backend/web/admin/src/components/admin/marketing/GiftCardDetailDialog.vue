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

<script setup lang="ts">
import GiftCardDetailPanel, {
  type GiftCardRecord,
  type GiftCardStatusOption,
  type GiftCardTransaction,
} from '@/components/admin/marketing/GiftCardDetailPanel.vue'
import type { AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Dialog } from '@/components/ui/dialog'

const props = withDefaults(defineProps<{
  open?: boolean
  statusUpdate?: string
  currentGiftCard?: GiftCardRecord | null
  loading?: boolean
  transactions?: GiftCardTransaction[]
  statusSubmitting?: boolean
  canEdit?: boolean
  formatCurrency: (value: unknown, currency?: string) => string
  formatDate: (value: unknown) => string
  giftCardStatusName: (status?: string) => string
  giftCardStatusTone: (status?: string) => AdminStatusTone
  giftCardStatusOptions: (giftCard: GiftCardRecord) => GiftCardStatusOption[]
  transactionTypeName: (type?: string) => string
}>(), {
  open: false,
  statusUpdate: 'active',
  currentGiftCard: null,
  loading: false,
  transactions: () => [],
  statusSubmitting: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:statusUpdate', value: string): void
  (event: 'update-status'): void
}>()
</script>
