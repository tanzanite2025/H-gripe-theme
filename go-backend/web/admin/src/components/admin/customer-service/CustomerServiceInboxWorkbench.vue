<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      v-if="showHeader"
      class="shrink-0"
      :title="title"
      :description="description"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="refreshInbox">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid v-if="showStats" class="shrink-0" :items="statItems" />

    <CustomerServiceInboxWorkspace
      v-model:transfer-to="transferTo"
      v-model:reply-message="replyMessage"
      :filters="filters"
      :conversations="filteredConversations"
      :selected-conversation="selectedConversation"
      :customer-typing-by-conversation="customerTypingByConversation"
      :pagination="pagination"
      :total-pages="totalPages"
      :loading="loading"
      :messages="messages"
      :messages-loading="messagesLoading"
      :selected-customer-typing="selectedCustomerTyping"
      :assignable-agents="assignableAgents"
      :transferring="transferring"
      :replying="replying"
      :can-edit="hasPermission('ticket:edit')"
      :customer-context="customerContext"
      :context-loading="contextLoading"
      :context-error="contextError"
      :context-last-updated-at="contextLastUpdatedAt"
      :selected-conversation-ids="selectedConversationIds"
      :batch-archiving="batchArchiving"
      @apply="applyFilters"
      @reset="resetFilters"
      @select="selectConversation"
      @archive="archiveConversation"
      @restore="restoreConversation"
      @close-and-archive="closeAndArchiveConversation"
      @resolve="resolveConversation"
      @reopen="reopenConversation"
      @toggle-selection="toggleConversationSelection"
      @toggle-page-selection="togglePageSelection"
      @bulk-archive="bulkArchiveConversations"
      @change-page="changePage"
      @transfer="transferConversation"
      @send-reply="sendReply"
      @send-message="sendCustomerServiceMessage"
      @refresh-context="refreshSelectedContext"
      @typing-input="handleReplyTypingInput"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { isAxiosError } from 'axios'
import { toast } from 'vue-sonner'
import {
  Clock3,
  Headset,
  MessagesSquare,
  RefreshCw,
  UserCheck,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import CustomerServiceInboxWorkspace from '@/components/admin/customer-service/CustomerServiceInboxWorkspace.vue'
import { Button } from '@/components/ui/button'
import customerServiceApi from '@/api/customerService'
import { useCustomerServiceInbox } from '@/composables/customerService/useCustomerServiceInbox'
import {
  useCustomerServiceNotificationSettings,
  useCustomerServiceRealtime,
} from '@/composables/customerService/useCustomerServiceRealtime'
import { useCustomerServiceTyping } from '@/composables/customerService/useCustomerServiceTyping'
import { useAuthStore } from '@/stores/auth'
import { statusDisplayValue } from '@/lib/customerServicePresentation'
import type { CustomerConversation, CustomerServiceSendMessagePayload } from '@/modules/customer-service/customerServiceTypes'

withDefaults(defineProps<{
  showHeader?: boolean
  showStats?: boolean
  title?: string
  description?: string
}>(), {
  showHeader: false,
  showStats: false,
  title: '客服对话',
  description: '处理网页 Public Chat 会话；客户侧只创建和读取自己的对话，客服侧在这里回复',
})

const authStore = useAuthStore()
const replying = ref(false)
const transferring = ref(false)
const batchArchiving = ref(false)
const selectedConversationIds = ref<string[]>([])

const hasPermission = (permission) => authStore.hasPermission(permission)
const { suppressDesktopWhenFocused } = useCustomerServiceNotificationSettings()

const {
  loading,
  messagesLoading,
  contextLoading,
  contextError,
  contextLastUpdatedAt,
  conversations,
  messages,
  customerContext,
  selectedConversation,
  replyMessage,
  transferTo,
  assignableAgents,
  pagination,
  filters,
  totalPages,
  filteredConversations,
  fetchConversations,
  fetchContext,
  fetchMessages,
  refreshInbox,
  selectConversation: selectInboxConversation,
  changePage,
  applyFilters,
  resetFilters,
  clearCurrentDraft,
} = useCustomerServiceInbox()

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
  closeCustomerServiceRealtime,
  sendCustomerServiceRealtimeControl,
} = useCustomerServiceRealtime({
  buildWebSocketUrl: (lastEventId: string) => {
    const conversationId = selectedConversation.value?.id
    if (!conversationId) return ''
    return customerServiceApi.buildWebSocketUrl('conversation', conversationId, lastEventId)
  },
  connectionKey: () => `conversation:${selectedConversation.value?.id || ''}`,
  shouldSilenceDesktopNotification: (event: Record<string, any>) => {
    if (!suppressDesktopWhenFocused.value || typeof document === 'undefined') return false
    if (document.visibilityState !== 'visible' || !document.hasFocus()) return false
    const selectedID = String(selectedConversation.value?.id || '').trim()
    const eventConversationID = String(event.ticket_id || event.conversation_id || '').trim()
    return Boolean(selectedID && eventConversationID && selectedID === eventConversationID)
  },
  onConnected: async () => {
    const conversationID = selectedConversation.value?.id
    if (!conversationID) return
    await Promise.all([
      fetchMessages(conversationID),
      fetchContext(conversationID),
    ])
  },
})

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
  sendTyping: (isTyping: boolean) => sendCustomerServiceRealtimeControl({
    type: 'typing',
    is_typing: isTyping,
  }),
})

const {
  connectCustomerServiceRealtime: connectInboxRealtime,
  closeCustomerServiceRealtime: closeInboxRealtime,
} = useCustomerServiceRealtime({
  buildWebSocketUrl: (lastEventId: string) => customerServiceApi.buildWebSocketUrl('inbox', undefined, lastEventId),
  connectionKey: () => 'inbox',
  shouldSilenceDesktopNotification: (event: Record<string, any>) => {
    if (!suppressDesktopWhenFocused.value || typeof document === 'undefined') return false
    if (document.visibilityState !== 'visible' || !document.hasFocus()) return false
    const selectedID = String(selectedConversation.value?.id || '').trim()
    const eventConversationID = String(event.ticket_id || event.conversation_id || '').trim()
    return Boolean(selectedID && eventConversationID && selectedID === eventConversationID)
  },
  onTyping: handleCustomerTypingEvent,
  onConnected: refreshInbox,
  onRefresh: async (event) => {
    await fetchConversations()

    if (!selectedConversation.value || Number(event.ticket_id) !== Number(selectedConversation.value.id)) {
      return
    }

    if (event.type === 'conversation.message.created') {
      await fetchMessages(selectedConversation.value.id)
    }
    if (event.type === 'conversation.assigned') {
      await fetchContext(selectedConversation.value.id)
    }
  }
})

const getCustomerServiceMessageToastLabel = (messageType?: string): string => {
  switch (String(messageType || '').trim().toLowerCase()) {
    case 'image':
      return '图片已发送'
    case 'order':
      return '订单已发送'
    case 'product':
      return '产品链接已发送'
    case 'video':
      return '视频已发送'
    default:
      return '回复已发送'
  }
}

const sendCustomerServiceMessage = async (payload: CustomerServiceSendMessagePayload) => {
  if (!selectedConversation.value || replying.value) return

  const message = String(payload.message || '').trim()
  if (!message) return
  const conversationID = selectedConversation.value.id

  replying.value = true
  try {
    await notifyAgentTyping(false)
    await customerServiceApi.sendMessage(conversationID, message, {
      messageType: payload.messageType,
      metadata: payload.metadata,
      attachmentUrl: payload.attachmentUrl,
      attachments: payload.attachments,
    })

    if (payload.clearReplyMessage) {
      clearCurrentDraft(conversationID)
    }

    toast.success(payload.toastLabel || getCustomerServiceMessageToastLabel(payload.messageType))
    await Promise.all([
      fetchMessages(conversationID),
      fetchConversations(),
      fetchContext(conversationID),
    ])
  } catch (error) {
    console.error('Failed to send customer-service message:', error)
  } finally {
    replying.value = false
  }
}

const selectConversation = async (conversation: CustomerConversation): Promise<void> => {
  resetAgentTypingState()
  await selectInboxConversation(conversation)
}

const refreshSelectedContext = () => {
  if (!selectedConversation.value?.id) return
  void fetchContext(selectedConversation.value.id)
}

watch(
  () => selectedConversation.value?.id,
  (conversationId) => {
    closeCustomerServiceRealtime()
    if (conversationId) {
      connectCustomerServiceRealtime()
    }
  },
)

const sendReply = async () => {
  await sendCustomerServiceMessage({
    message: replyMessage.value,
    messageType: 'text',
    clearReplyMessage: true,
    toastLabel: '回复已发送',
  })
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

const archiveConversation = async (conversation: CustomerConversation) => {
  try {
    await customerServiceApi.archiveConversation(conversation.id)
    await fetchConversations()
    toast.success('会话已归档，可在“已归档”中恢复', {
      duration: 8000,
      action: {
        label: '撤销',
        onClick: () => {
          void restoreConversation(conversation)
        },
      },
    })
  } catch (error) {
    console.error('Failed to archive customer-service conversation:', error)
  }
}

const restoreConversation = async (conversation: CustomerConversation) => {
  try {
    await customerServiceApi.restoreConversation(conversation.id)
    toast.success('会话已恢复到收件箱')
    await fetchConversations()
  } catch (error) {
    console.error('Failed to restore customer-service conversation:', error)
  }
}

const conversationStatusVersion = (conversation: CustomerConversation): number => {
  const version = Number(conversation.status_version || 0)
  return Number.isInteger(version) && version > 0 ? version : 0
}

const handleStatusMutationError = async (error: unknown) => {
  console.error('Failed to update customer-service conversation status:', error)
  if (isAxiosError(error) && error.response?.status === 409) {
    await fetchConversations()
    toast.warning('会话状态已被其他客服更新，列表已刷新，请重试')
    return
  }
  toast.error('会话状态更新失败，请稍后重试')
}

const reopenConversation = async (conversation: CustomerConversation) => {
  const expectedStatusVersion = conversationStatusVersion(conversation)
  if (!expectedStatusVersion) {
    await fetchConversations()
    toast.warning('会话版本已刷新，请重试')
    return
  }
  try {
    await customerServiceApi.updateConversationStatus(conversation.id, 'open', expectedStatusVersion, {
      archive: false,
      reasonCode: 'operator_reopen',
    })
    toast.success('会话已重新打开并恢复到收件箱')
    await fetchConversations()
  } catch (error) {
    await handleStatusMutationError(error)
  }
}

const closeAndArchiveConversation = async (conversation: CustomerConversation) => {
  const expectedStatusVersion = conversationStatusVersion(conversation)
  if (!expectedStatusVersion) {
    await fetchConversations()
    toast.warning('会话版本已刷新，请重试')
    return
  }
  try {
    const updated = await customerServiceApi.updateConversationStatus(conversation.id, 'closed', expectedStatusVersion, {
      archive: true,
      reasonCode: 'operator_close_and_archive',
    })
    const closedVersion = Number(updated.status_version || expectedStatusVersion + 1)
    await fetchConversations()
    toast.success('会话已关闭并归档', {
      duration: 8000,
      action: {
        label: '撤销',
        onClick: () => {
          void reopenConversation({
            ...conversation,
            status: 'closed',
            status_version: closedVersion,
            inbox_archived: true,
          })
        },
      },
    })
  } catch (error) {
    await handleStatusMutationError(error)
  }
}

const resolveConversation = async (conversation: CustomerConversation) => {
  const expectedStatusVersion = conversationStatusVersion(conversation)
  if (!expectedStatusVersion) {
    await fetchConversations()
    toast.warning('会话版本已刷新，请重试')
    return
  }
  try {
    const updated = await customerServiceApi.updateConversationStatus(conversation.id, 'resolved', expectedStatusVersion, {
      reasonCode: 'operator_resolve',
    })
    const resolvedVersion = Number(updated.status_version || expectedStatusVersion + 1)
    await fetchConversations()
    toast.success('会话已标记为已解决', {
      duration: 8000,
      action: {
        label: '撤销',
        onClick: () => {
          void reopenConversation({
            ...conversation,
            status: 'resolved',
            status_version: resolvedVersion,
          })
        },
      },
    })
  } catch (error) {
    await handleStatusMutationError(error)
  }
}

const normalizeConversationID = (id: string | number): string => String(id)

const toggleConversationSelection = (conversation: CustomerConversation, selected: boolean) => {
  const id = normalizeConversationID(conversation.id)
  const next = new Set(selectedConversationIds.value)
  if (selected) next.add(id)
  else next.delete(id)
  selectedConversationIds.value = Array.from(next)
}

const togglePageSelection = (items: CustomerConversation[], selected: boolean) => {
  const next = new Set(selectedConversationIds.value)
  items.forEach((conversation) => {
    const id = normalizeConversationID(conversation.id)
    if (selected) next.add(id)
    else next.delete(id)
  })
  selectedConversationIds.value = Array.from(next)
}

const bulkArchiveConversations = async () => {
  if (selectedConversationIds.value.length === 0 || batchArchiving.value) return
  const ids = [...selectedConversationIds.value]
  batchArchiving.value = true
  try {
    const result = await customerServiceApi.bulkArchiveConversations(ids)
    const archivedCount = Number(result.archived_count || ids.length)
    selectedConversationIds.value = []
    await fetchConversations()
    toast.success(`已归档 ${archivedCount} 个会话`, {
      duration: 8000,
      action: {
        label: '撤销',
        onClick: () => {
          void Promise.all(ids.map((id) => customerServiceApi.restoreConversation(id)))
            .then(async () => {
              await fetchConversations()
              toast.success('批量归档已撤销')
            })
            .catch((error) => {
              console.error('Failed to undo bulk archive:', error)
              toast.error('批量归档撤销失败，请到“已归档”中逐条恢复')
            })
        },
      },
    })
  } catch (error) {
    console.error('Failed to bulk archive customer-service conversations:', error)
  } finally {
    batchArchiving.value = false
  }
}

watch(
  () => conversations.value.map((conversation) => `${normalizeConversationID(conversation.id)}:${conversation.inbox_archived ? 1 : 0}`).join('|'),
  () => {
    const selectableIDs = new Set(
      conversations.value
        .filter((conversation) => !conversation.inbox_archived)
        .map((conversation) => normalizeConversationID(conversation.id)),
    )
    selectedConversationIds.value = selectedConversationIds.value.filter((id) => selectableIDs.has(id))
  },
)

onMounted(async () => {
  await refreshInbox()
  connectInboxRealtime()
  if (selectedConversation.value?.id) {
    connectCustomerServiceRealtime()
  }
})

onBeforeUnmount(() => {
  notifyAgentTyping(false)
  closeCustomerServiceRealtime()
  closeInboxRealtime()
  clearTypingTimers()
})
</script>

