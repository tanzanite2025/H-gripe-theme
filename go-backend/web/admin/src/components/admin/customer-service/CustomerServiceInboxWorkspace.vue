<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
    <div class="shrink-0">
      <CustomerServiceFilters
        :filters="filters"
        :assignable-agents="assignableAgents"
        :assignable-groups="assignableGroups"
        :loading="loading"
        @apply="emit('apply')"
        @reset="emit('reset')"
      />
    </div>

    <section class="grid min-h-0 flex-1 grid-rows-[minmax(0,0.85fr)_minmax(0,1.2fr)_minmax(0,0.95fr)] gap-4 overflow-hidden xl:grid-cols-[320px_minmax(0,1fr)_320px] xl:grid-rows-[minmax(0,1fr)] 2xl:grid-cols-[360px_minmax(0,1fr)_360px]">
      <CustomerConversationListPanel
        :conversations="conversations"
        :selected-conversation="selectedConversation"
        :assignable-agents="assignableAgents"
        :customer-typing-by-conversation="customerTypingByConversation"
        :pagination="pagination"
        :total-pages="totalPages"
        :loading="loading"
        @select="emit('select', $event)"
        @change-page="emit('change-page', $event)"
      />

      <CustomerConversationDetailPanel
        :transfer-to="transferTo"
        :reply-message="replyMessage"
        :selected-conversation="selectedConversation"
        :messages="messages"
        :messages-loading="messagesLoading"
        :selected-customer-typing="selectedCustomerTyping"
        :assignable-agents="assignableAgents"
        :transferring="transferring"
        :replying="replying"
        :can-edit="canEdit"
        @update:transfer-to="emit('update:transferTo', $event)"
        @update:reply-message="emit('update:replyMessage', $event)"
        @transfer="emit('transfer')"
        @send-reply="emit('send-reply')"
        @typing-input="emit('typing-input')"
      />

      <CustomerContextPanel
        :selected-conversation="selectedConversation"
        :customer-context="customerContext"
        :loading="contextLoading"
      />
    </section>
  </div>
</template>

<script setup>
import CustomerContextPanel from '@/components/admin/customer-service/CustomerContextPanel.vue'
import CustomerConversationDetailPanel from '@/components/admin/customer-service/CustomerConversationDetailPanel.vue'
import CustomerConversationListPanel from '@/components/admin/customer-service/CustomerConversationListPanel.vue'
import CustomerServiceFilters from '@/components/admin/customer-service/CustomerServiceFilters.vue'

defineProps({
  filters: { type: Object, required: true },
  conversations: { type: Array, default: () => [] },
  selectedConversation: { type: Object, default: null },
  customerTypingByConversation: { type: Object, default: () => ({}) },
  pagination: { type: Object, required: true },
  totalPages: { type: Number, default: 1 },
  loading: { type: Boolean, default: false },
  transferTo: { type: String, default: '' },
  replyMessage: { type: String, default: '' },
  messages: { type: Array, default: () => [] },
  messagesLoading: { type: Boolean, default: false },
  selectedCustomerTyping: { type: Object, default: null },
  assignableAgents: { type: Array, default: () => [] },
  assignableGroups: { type: Array, default: () => [] },
  transferring: { type: Boolean, default: false },
  replying: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  customerContext: { type: Object, default: null },
  contextLoading: { type: Boolean, default: false },
})

const emit = defineEmits([
  'apply',
  'reset',
  'select',
  'change-page',
  'update:transferTo',
  'update:replyMessage',
  'transfer',
  'send-reply',
  'typing-input',
])
</script>
