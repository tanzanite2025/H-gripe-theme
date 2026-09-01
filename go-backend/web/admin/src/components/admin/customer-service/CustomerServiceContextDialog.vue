<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogSecondaryContent @open-auto-focus.prevent>
      <DialogHeader class="shrink-0 border-b px-4 py-3 pr-12 sm:px-5 sm:py-4">
        <DialogTitle class="truncate">{{ conversationTitle }}</DialogTitle>
        <DialogDescription class="break-all text-xs">
          {{ conversationDescription }}
        </DialogDescription>
      </DialogHeader>

      <div class="min-h-0 overflow-hidden p-3 sm:p-4">
        <CustomerContextPanel
          class="h-full"
          :selected-conversation="selectedConversation"
          :customer-context="customerContext"
          :loading="loading"
        />
      </div>
    </DialogSecondaryContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import CustomerContextPanel from '@/components/admin/customer-service/CustomerContextPanel.vue'
import { Dialog, DialogDescription, DialogHeader, DialogSecondaryContent, DialogTitle } from '@/components/ui/dialog'
import type { CustomerContext, CustomerConversation } from '@/modules/customer-service/customerServiceTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  selectedConversation?: CustomerConversation | null
  customerContext?: CustomerContext | null
  loading?: boolean
}>(), {
  open: false,
  selectedConversation: null,
  customerContext: null,
  loading: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
}>()

const conversationTitle = computed(() => props.selectedConversation?.customer_name || '匿名客户')
const conversationDescription = computed(() => {
  const value = props.selectedConversation?.conversation_id || props.selectedConversation?.ticket_number || props.selectedConversation?.id
  return value ? String(value) : '客户上下文'
})
</script>

