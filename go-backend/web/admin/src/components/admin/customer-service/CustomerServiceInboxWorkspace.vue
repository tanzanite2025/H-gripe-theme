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

<script setup lang="ts">
import CustomerContextPanel from '@/components/admin/customer-service/CustomerContextPanel.vue'
import CustomerConversationDetailPanel from '@/components/admin/customer-service/CustomerConversationDetailPanel.vue'
import CustomerConversationListPanel from '@/components/admin/customer-service/CustomerConversationListPanel.vue'
import CustomerServiceFilters from '@/components/admin/customer-service/CustomerServiceFilters.vue'
import type {
  AssignableAgent,
  AssignableGroup,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerPagination,
  CustomerServiceFiltersState,
  CustomerTypingByConversation,
  CustomerTypingState,
} from './customerServiceTypes'

withDefaults(defineProps<{
  filters: CustomerServiceFiltersState
  conversations?: CustomerConversation[]
  selectedConversation?: CustomerConversation | null
  customerTypingByConversation?: CustomerTypingByConversation
  pagination: CustomerPagination
  totalPages?: number
  loading?: boolean
  transferTo?: string
  replyMessage?: string
  messages?: CustomerConversationMessage[]
  messagesLoading?: boolean
  selectedCustomerTyping?: CustomerTypingState | null
  assignableAgents?: AssignableAgent[]
  assignableGroups?: AssignableGroup[]
  transferring?: boolean
  replying?: boolean
  canEdit?: boolean
  customerContext?: CustomerContext | null
  contextLoading?: boolean
}>(), {
  conversations: () => [],
  selectedConversation: null,
  customerTypingByConversation: () => ({}),
  totalPages: 1,
  loading: false,
  transferTo: '',
  replyMessage: '',
  messages: () => [],
  messagesLoading: false,
  selectedCustomerTyping: null,
  assignableAgents: () => [],
  assignableGroups: () => [],
  transferring: false,
  replying: false,
  canEdit: false,
  customerContext: null,
  contextLoading: false,
})

const emit = defineEmits<{
  (event: 'apply'): void
  (event: 'reset'): void
  (event: 'select', conversation: CustomerConversation): void
  (event: 'change-page', page: number): void
  (event: 'update:transferTo', value: string): void
  (event: 'update:replyMessage', value: string): void
  (event: 'transfer'): void
  (event: 'send-reply'): void
  (event: 'typing-input'): void
}>()
</script>
