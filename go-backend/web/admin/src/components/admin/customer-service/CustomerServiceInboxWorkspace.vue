<template>
  <div class="flex min-h-0 flex-1 flex-col gap-4 overflow-hidden">
    <div class="flex shrink-0 items-center justify-between gap-2 lg:hidden">
      <div class="inline-flex min-w-0 flex-1 items-center rounded-lg border bg-muted/35 p-1">
        <Button
          variant="ghost"
          size="sm"
          class="min-w-0 flex-1 rounded-md text-xs"
          :class="{ 'bg-background shadow-sm': mobileView === 'conversations' }"
          @click="mobileView = 'conversations'"
        >
          <List class="size-3.5" />
          会话
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="min-w-0 flex-1 rounded-md text-xs"
 :class="{ 'bg-background shadow-sm': mobileView === 'messages'}"
          :disabled="!selectedConversation"
          @click="mobileView = 'messages'"
        >
          <MessageSquare class="size-3.5" />
          对话
        </Button>
      </div>
      <Button
        variant="outline"
        size="icon"
        class="shrink-0 rounded-full"
        aria-label="筛选客服会话"
        @click="mobileFiltersOpen = true"
      >
        <SlidersHorizontal class="size-4" />
      </Button>
    </div>

    <div class="hidden shrink-0 lg:block">
      <CustomerServiceFilters
        :filters="filters"
        :loading="loading"
        @apply="emit('apply')"
        @reset="emit('reset')"
      />
    </div>

    <section class="hidden min-h-0 flex-1 gap-4 overflow-hidden lg:grid lg:grid-cols-[minmax(300px,320px)_minmax(0,1fr)] lg:grid-rows-[minmax(0,1fr)] xl:grid-cols-[320px_minmax(0,1fr)] 2xl:grid-cols-[360px_minmax(0,1fr)]">
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
        :customer-context="customerContext"
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
        @send-message="emit('send-message', $event)"
        @open-context="openContextDialog"
        @typing-input="emit('typing-input')"
      />
    </section>

    <section class="min-h-0 flex-1 lg:hidden">
      <CustomerConversationListPanel
        v-if="mobileView === 'conversations'"
        :conversations="conversations"
        :selected-conversation="selectedConversation"
        :assignable-agents="assignableAgents"
        :customer-typing-by-conversation="customerTypingByConversation"
        :pagination="pagination"
        :total-pages="totalPages"
        :loading="loading"
        @select="selectMobileConversation"
        @change-page="emit('change-page', $event)"
      />

      <CustomerConversationDetailPanel
        v-else-if="mobileView === 'messages'"
        :transfer-to="transferTo"
        :reply-message="replyMessage"
        :selected-conversation="selectedConversation"
        :customer-context="customerContext"
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
        @send-message="emit('send-message', $event)"
        @open-context="openContextDialog"
        @typing-input="emit('typing-input')"
      />
    </section>
  </div>

  <Sheet v-model:open="mobileFiltersOpen">
    <SheetContent side="bottom" class="max-h-[86dvh] rounded-t-2xl p-0" @open-auto-focus.prevent>
      <SheetHeader class="border-b pr-12">
        <SheetTitle>筛选客服会话</SheetTitle>
        <SheetDescription>按客户、状态和未读消息筛选。</SheetDescription>
      </SheetHeader>
      <div class="overflow-y-auto p-4">
        <CustomerServiceFilters
          :filters="filters"
          :loading="loading"
          @apply="applyMobileFilters"
          @reset="resetMobileFilters"
        />
      </div>
    </SheetContent>
  </Sheet>

  <CustomerServiceContextDialog
    v-model:open="contextDialogOpen"
    :selected-conversation="selectedConversation"
    :customer-context="customerContext"
    :loading="contextLoading"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { List, MessageSquare, SlidersHorizontal } from '@lucide/vue'
import CustomerServiceContextDialog from '@/components/admin/customer-service/CustomerServiceContextDialog.vue'
import CustomerConversationDetailPanel from '@/components/admin/customer-service/CustomerConversationDetailPanel.vue'
import CustomerConversationListPanel from '@/components/admin/customer-service/CustomerConversationListPanel.vue'
import CustomerServiceFilters from '@/components/admin/customer-service/CustomerServiceFilters.vue'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet'
import type {
  AssignableAgent,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerPagination,
  CustomerServiceFiltersState,
  CustomerServiceSendMessagePayload,
  CustomerTypingByConversation,
  CustomerTypingState,
} from './customerServiceTypes'

const props = withDefaults(defineProps<{
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
  (event: 'send-message', payload: CustomerServiceSendMessagePayload): void
  (event: 'typing-input'): void
}>()

type MobileInboxView = 'conversations' | 'messages'

const mobileView = ref<MobileInboxView>('conversations')
const mobileFiltersOpen = ref(false)
const contextDialogOpen = ref(false)

const selectMobileConversation = (conversation: CustomerConversation) => {
  mobileView.value = 'messages'
  emit('select', conversation)
}

const openContextDialog = () => {
  if (!props.selectedConversation) return
  contextDialogOpen.value = true
}

const applyMobileFilters = () => {
  mobileFiltersOpen.value = false
  emit('apply')
}

const resetMobileFilters = () => {
  mobileFiltersOpen.value = false
  emit('reset')
}
</script>
