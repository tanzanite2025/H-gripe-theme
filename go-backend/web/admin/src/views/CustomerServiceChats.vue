<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="客服对话"
      description="处理网页 Public Chat 会话；客户侧只创建和读取自己的对话，客服侧在这里回复"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="refreshInbox">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <CustomerServiceRegionAnalytics :analytics="regionAnalytics" :loading="regionAnalyticsLoading" />

    <CustomerServiceFilters
      :filters="filters"
      :assignable-agents="assignableAgents"
      :loading="loading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <section class="grid min-h-[calc(100dvh-280px)] gap-4 xl:grid-cols-[320px_minmax(0,1fr)_320px] 2xl:grid-cols-[360px_minmax(0,1fr)_360px]">
      <CustomerConversationListPanel
        :conversations="filteredConversations"
        :selected-conversation="selectedConversation"
        :assignable-agents="assignableAgents"
        :customer-typing-by-conversation="customerTypingByConversation"
        :pagination="pagination"
        :total-pages="totalPages"
        :loading="loading"
        @select="selectConversation"
        @change-page="changePage"
      />

      <CustomerConversationDetailPanel
        v-model:transfer-to="transferTo"
        v-model:reply-message="replyMessage"
        :selected-conversation="selectedConversation"
        :messages="messages"
        :messages-loading="messagesLoading"
        :selected-customer-typing="selectedCustomerTyping"
        :assignable-agents="assignableAgents"
        :transferring="transferring"
        :replying="replying"
        :can-edit="hasPermission('ticket:edit')"
        @transfer="transferConversation"
        @send-reply="sendReply"
        @typing-input="handleReplyTypingInput"
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
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  Clock3,
  Headset,
  MessagesSquare,
  RefreshCw,
  UserCheck,
} from '@lucide/vue'
import CustomerContextPanel from '@/components/admin/customer-service/CustomerContextPanel.vue'
import CustomerConversationDetailPanel from '@/components/admin/customer-service/CustomerConversationDetailPanel.vue'
import CustomerConversationListPanel from '@/components/admin/customer-service/CustomerConversationListPanel.vue'
import CustomerServiceFilters from '@/components/admin/customer-service/CustomerServiceFilters.vue'
import CustomerServiceRegionAnalytics from '@/components/admin/customer-service/CustomerServiceRegionAnalytics.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import { useAuthStore } from '@/stores/auth'
import customerServiceApi from '@/api/customerService'
import { useCustomerServiceInbox } from '@/composables/customerService/useCustomerServiceInbox'
import { useCustomerServiceRealtime } from '@/composables/customerService/useCustomerServiceRealtime'
import { useCustomerServiceTyping } from '@/composables/customerService/useCustomerServiceTyping'
import { statusDisplayValue } from '@/lib/customerServicePresentation'

const authStore = useAuthStore()
const replying = ref(false)
const transferring = ref(false)

const hasPermission = (permission) => authStore.hasPermission(permission)

const {
  loading,
  messagesLoading,
  contextLoading,
  regionAnalyticsLoading,
  conversations,
  messages,
  customerContext,
  regionAnalytics,
  selectedConversation,
  replyMessage,
  transferTo,
  assignableAgents,
  pagination,
  filters,
  totalPages,
  filteredConversations,
  fetchRegionAnalytics,
  fetchConversations,
  fetchContext,
  fetchMessages,
  refreshInbox,
  selectConversation: selectInboxConversation,
  changePage,
  applyFilters,
  resetFilters,
} = useCustomerServiceInbox()

const {
  customerTypingByConversation,
  selectedCustomerTyping,
  clearTypingTimers,
  handleCustomerTypingEvent,
  handleReplyTypingInput,
  notifyAgentTyping,
  resetAgentTypingState,
} = useCustomerServiceTyping({
  selectedConversation,
  replyMessage,
  canSendTyping: () => hasPermission('ticket:edit'),
})

const statItems = computed(() => {
  const total = conversations.value.length
  const unread = conversations.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
  const active = conversations.value.filter((item) => statusDisplayValue(item.display_status || item.status) === 'active').length
  const closed = conversations.value.filter((item) => statusDisplayValue(item.display_status || item.status) === 'closed').length

  return [
    { key: 'total', label: '当前页会话', value: total, icon: MessagesSquare, tone: 'gray' },
    { key: 'unread', label: '未读消息', value: unread, icon: Clock3, tone: unread > 0 ? 'coral' : 'gray' },
    { key: 'active', label: '进行中', value: active, icon: Headset, tone: 'blue' },
    { key: 'closed', label: '已关闭', value: closed, icon: UserCheck, tone: 'green' }
  ]
})

const {
  connectCustomerServiceRealtime,
  closeCustomerServiceRealtime
} = useCustomerServiceRealtime({
  buildEventUrl: () => customerServiceApi.buildEventsUrl('inbox'),
  onTyping: handleCustomerTypingEvent,
  onRefresh: async (event) => {
    await Promise.all([fetchConversations(), fetchRegionAnalytics()])

    if (!selectedConversation.value || Number(event.ticket_id) !== Number(selectedConversation.value.id)) {
      return
    }

    if (event.type === 'conversation.message.created') {
      await fetchMessages(selectedConversation.value.id)
    }
    if (['conversation.context.updated', 'conversation.assigned'].includes(event.type)) {
      await fetchContext(selectedConversation.value.id)
    }
  }
})

const selectConversation = async (conversation) => {
  resetAgentTypingState()
  await selectInboxConversation(conversation)
}

const sendReply = async () => {
  if (!selectedConversation.value || !replyMessage.value.trim()) return
  const message = replyMessage.value.trim()
  replying.value = true
  try {
    await notifyAgentTyping(false)
    await customerServiceApi.sendMessage(selectedConversation.value.id, message)
    replyMessage.value = ''
    toast.success('回复已发送')
    await Promise.all([
      fetchMessages(selectedConversation.value.id),
      fetchConversations(),
      fetchContext(selectedConversation.value.id)
    ])
  } catch (error) {
    console.error('Failed to send customer-service reply:', error)
  } finally {
    replying.value = false
  }
}

const transferConversation = async () => {
  if (!selectedConversation.value || !transferTo.value) return
  transferring.value = true
  try {
    await customerServiceApi.transferConversation(selectedConversation.value.id, Number(transferTo.value))
    toast.success('会话已转接')
    await Promise.all([
      fetchConversations(),
      fetchContext(selectedConversation.value.id)
    ])
  } catch (error) {
    console.error('Failed to transfer customer-service conversation:', error)
  } finally {
    transferring.value = false
  }
}

onMounted(async () => {
  await refreshInbox()
  connectCustomerServiceRealtime()
})

onBeforeUnmount(() => {
  notifyAgentTyping(false)
  closeCustomerServiceRealtime()
  clearTypingTimers()
})
</script>
