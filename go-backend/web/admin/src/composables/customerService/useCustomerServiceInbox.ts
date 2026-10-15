import { computed, onScopeDispose, reactive, ref } from 'vue'
import customerServiceApi from '@/api/customerService'
import type {
  AssignableAgent,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerServiceFiltersState,
} from '@/modules/customer-service/customerServiceTypes'

interface CustomerServiceConversationFetchOptions {
  showLoadingIndicator?: boolean
}

interface CustomerServiceMessageFetchOptions {
  showLoadingIndicator?: boolean
  clearMessagesWhenRequestFails?: boolean
  markMessagesAsRead?: boolean
}

export const useCustomerServiceInbox = () => {
  const loading = ref(false)
  const conversationListError = ref<string | null>(null)
  const messagesLoading = ref(false)
  const messagesError = ref<string | null>(null)
  const messageReadReceiptError = ref<string | null>(null)
  const contextLoading = ref(false)
  const conversations = ref<CustomerConversation[]>([])
  const messages = ref<CustomerConversationMessage[]>([])
  const customerContext = ref<CustomerContext | null>(null)
  const contextError = ref<string | null>(null)
  const contextLastUpdatedAt = ref<Date | null>(null)
  const selectedConversation = ref<CustomerConversation | null>(null)
  const replyMessage = ref('')
  const draftMap = reactive<Record<string, string>>({})
  let contextRequestSequence = 0
  const transferTo = ref('')
  const assignableAgents = ref<AssignableAgent[]>([])
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
  const filters = reactive<CustomerServiceFiltersState>({
    search: '',
    view: 'inbox',
    status: 'all',
    identity: 'all',
    unread: 'all',
  })

  const totalPages = computed(() => Math.max(1, Math.ceil((pagination.total || 0) / pagination.pageSize)))

  const filteredConversations = computed(() => conversations.value)
  let conversationRequestSequence = 0
  let messagesRequestSequence = 0
  let agentsRequestSequence = 0
  let conversationSelectionSequence = 0
  let customerServiceInboxScopeDisposed = false

  onScopeDispose(() => {
    customerServiceInboxScopeDisposed = true
    conversationRequestSequence += 1
    messagesRequestSequence += 1
    contextRequestSequence += 1
    agentsRequestSequence += 1
    conversationSelectionSequence += 1
  })

  const conversationKey = (conversationID: number | string | null | undefined): string => {
    const normalized = String(conversationID ?? '').trim()
    return normalized ? normalized : ''
  }

  const saveCurrentDraft = (): void => {
    const key = conversationKey(selectedConversation.value?.id)
    if (!key) return
    const draft = replyMessage.value
    if (draft) draftMap[key] = draft
    else delete draftMap[key]
  }

  const clearCurrentDraft = (conversationID: number | string | null | undefined): void => {
    if (customerServiceInboxScopeDisposed) return
    const key = conversationKey(conversationID)
    if (!key) return
    delete draftMap[key]
    if (conversationKey(selectedConversation.value?.id) === key) replyMessage.value = ''
  }

  const clearSelectedCustomerConversationAndInvalidatePendingReads = (): void => {
    if (customerServiceInboxScopeDisposed) return
    saveCurrentDraft()
    conversationSelectionSequence += 1
    messagesRequestSequence += 1
    contextRequestSequence += 1
    selectedConversation.value = null
    messages.value = []
    messagesLoading.value = false
    messagesError.value = null
    messageReadReceiptError.value = null
    customerContext.value = null
    contextLoading.value = false
    contextError.value = null
    contextLastUpdatedAt.value = null
    replyMessage.value = ''
    transferTo.value = ''
  }

  const fetchConversations = async (
    options: CustomerServiceConversationFetchOptions = {},
  ): Promise<boolean> => {
    if (customerServiceInboxScopeDisposed) return false
    const showLoadingIndicator = options.showLoadingIndicator !== false
    if (showLoadingIndicator) loading.value = true
    const requestSequence = ++conversationRequestSequence
    try {
      const data = await customerServiceApi.listConversations({
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search.trim() || undefined,
        view: filters.view || 'inbox',
        status: filters.status !== 'all' ? filters.status : undefined,
        identity: filters.identity !== 'all' ? filters.identity : undefined,
        unread: filters.unread === 'unread' ? 'true' : undefined,
      }, true)
      if (customerServiceInboxScopeDisposed || requestSequence !== conversationRequestSequence) return false
      conversationListError.value = null
      conversations.value = data.conversations || []
      pagination.total = data.pagination?.total ?? conversations.value.length

      if (selectedConversation.value) {
        const refreshed = conversations.value.find((item) => Number(item.id) === Number(selectedConversation.value?.id))
        if (refreshed) {
          selectedConversation.value = refreshed
        } else {
          clearSelectedCustomerConversationAndInvalidatePendingReads()
        }
      }
      return true
    } catch (error) {
      if (customerServiceInboxScopeDisposed || requestSequence !== conversationRequestSequence) return false
      console.error('Failed to fetch customer-service conversations:', error)
      conversationListError.value = '会话列表暂时无法读取，当前无法确认客户会话和未读消息数量。请重新加载；如果仍然失败，请联系技术负责人。'
      conversations.value = []
      pagination.total = 0
      clearSelectedCustomerConversationAndInvalidatePendingReads()
      return false
    } finally {
      if (!customerServiceInboxScopeDisposed && requestSequence === conversationRequestSequence) loading.value = false
    }
  }

  const fetchContext = async (conversationID: number | string | null | undefined) => {
    if (customerServiceInboxScopeDisposed) return
    if (!conversationID) {
      contextRequestSequence += 1
      customerContext.value = null
      contextError.value = null
      contextLastUpdatedAt.value = null
      contextLoading.value = false
      return
    }
    if (conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)) return

    const requestSequence = ++contextRequestSequence
    const requestedConversationKey = conversationKey(conversationID)
    contextLoading.value = true
    contextError.value = null
    try {
      const context = await customerServiceApi.getConversationContext(conversationID)
      if (customerServiceInboxScopeDisposed || requestSequence !== contextRequestSequence || conversationKey(selectedConversation.value?.id) !== requestedConversationKey) return
      customerContext.value = context
      contextLastUpdatedAt.value = new Date()
    } catch (error) {
      if (customerServiceInboxScopeDisposed || requestSequence !== contextRequestSequence || conversationKey(selectedConversation.value?.id) !== requestedConversationKey) return
      console.error('Failed to fetch customer-service context:', error)
      contextError.value = '请求失败，请稍后重试'
    } finally {
      if (!customerServiceInboxScopeDisposed && requestSequence === contextRequestSequence) contextLoading.value = false
    }
  }

  const fetchAgents = async () => {
    if (customerServiceInboxScopeDisposed) return
    const requestSequence = ++agentsRequestSequence
    try {
      const agents = await customerServiceApi.listAgents()
      if (customerServiceInboxScopeDisposed || requestSequence !== agentsRequestSequence) return
      assignableAgents.value = agents
    } catch (error) {
      if (customerServiceInboxScopeDisposed || requestSequence !== agentsRequestSequence) return
      console.error('Failed to fetch public chat agents:', error)
      assignableAgents.value = []
    }
  }

  const fetchMessages = async (
    conversationID: number | string | null | undefined,
    options: CustomerServiceMessageFetchOptions = {},
  ) => {
    if (customerServiceInboxScopeDisposed || !conversationID || conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)) return
    const showLoadingIndicator = options.showLoadingIndicator !== false
    const clearMessagesWhenRequestFails = options.clearMessagesWhenRequestFails !== false
    const markMessagesAsRead = options.markMessagesAsRead !== false
    const requestSequence = ++messagesRequestSequence
    messagesLoading.value = showLoadingIndicator
    messagesError.value = null
    try {
      const nextMessages = await customerServiceApi.listMessages(conversationID, true)
      if (
        customerServiceInboxScopeDisposed
        || requestSequence !== messagesRequestSequence
        || conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)
      ) return
      messages.value = nextMessages
      messagesLoading.value = false
      if (markMessagesAsRead) {
        // Read receipts are a separate side effect. A failed receipt must not
        // discard successfully loaded customer messages or imply an empty history.
        try {
          await customerServiceApi.markMessagesRead(conversationID, true)
          if (!customerServiceInboxScopeDisposed && requestSequence === messagesRequestSequence && conversationKey(selectedConversation.value?.id) === conversationKey(conversationID)) messageReadReceiptError.value = null
        } catch (error) {
          if (customerServiceInboxScopeDisposed || requestSequence !== messagesRequestSequence || conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)) return
          console.error('Failed to update customer-service message read receipt:', error)
          messageReadReceiptError.value = '消息已显示，但已读状态暂时无法更新。请重新读取消息后确认未读提示，无需因此重复回复客户。'
        }
      }
    } catch (error) {
      if (customerServiceInboxScopeDisposed || requestSequence !== messagesRequestSequence || conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)) return
      console.error('Failed to fetch customer-service messages:', error)
      messagesError.value = clearMessagesWhenRequestFails
        ? '消息记录暂时无法读取，不能据此判断客户没有新消息。请重新读取；如果仍然失败，请联系技术负责人。'
        : '消息记录暂时无法更新，当前显示的内容可能不是最新记录。请重新读取；不能据此判断客户没有新消息。'
      if (clearMessagesWhenRequestFails) messages.value = []
    } finally {
      if (!customerServiceInboxScopeDisposed && requestSequence === messagesRequestSequence) messagesLoading.value = false
    }
  }

  const appendCustomerServiceMessageToSelectedConversation = (
    conversationID: number | string,
    message: CustomerConversationMessage,
  ): void => {
    if (customerServiceInboxScopeDisposed || conversationKey(selectedConversation.value?.id) !== conversationKey(conversationID)) return

    messagesRequestSequence += 1
    messagesLoading.value = false

    const messageID = String(message?.id ?? '').trim()
    const existingMessageIndex = messageID
      ? messages.value.findIndex((item) => String(item?.id ?? '').trim() === messageID)
      : -1
    if (existingMessageIndex >= 0) {
      const nextMessages = [...messages.value]
      nextMessages[existingMessageIndex] = message
      messages.value = nextMessages
      return
    }

    messages.value = [...messages.value, message]
  }

  const updateSelectedConversationPreviewAfterSendingMessage = (
    conversationID: number | string,
    message: CustomerConversationMessage,
  ): void => {
    if (customerServiceInboxScopeDisposed) return
    const normalizedConversationID = conversationKey(conversationID)
    const lastMessage = String(message.content || message.message || '').trim()
    if (!normalizedConversationID || !lastMessage) return

    conversations.value = conversations.value.map((conversation) => (
      conversationKey(conversation.id) === normalizedConversationID
        ? { ...conversation, last_message: lastMessage }
        : conversation
    ))

    const currentSelectedConversation = selectedConversation.value
    if (conversationKey(currentSelectedConversation?.id) === normalizedConversationID && currentSelectedConversation) {
      selectedConversation.value = {
        ...currentSelectedConversation,
        last_message: lastMessage,
      }
    }
  }

  const refreshInbox = async (): Promise<boolean> => {
    if (customerServiceInboxScopeDisposed) return false
    const currentSelectionSequence = conversationSelectionSequence
    const conversationListRequest = fetchConversations()
    const currentListSequence = conversationRequestSequence
    const [listAvailable] = await Promise.all([conversationListRequest, fetchAgents()])
    if (!listAvailable || customerServiceInboxScopeDisposed || currentListSequence !== conversationRequestSequence) return false
    if (currentSelectionSequence === conversationSelectionSequence && selectedConversation.value) {
      await Promise.all([
        fetchMessages(selectedConversation.value.id),
        fetchContext(selectedConversation.value.id),
      ])
    }
    return !customerServiceInboxScopeDisposed && currentListSequence === conversationRequestSequence
  }

  const selectConversation = async (conversation: CustomerConversation) => {
    if (customerServiceInboxScopeDisposed || conversationListError.value) return
    clearSelectedCustomerConversationAndInvalidatePendingReads()
    const currentSelectionSequence = conversationSelectionSequence
    selectedConversation.value = conversation
    replyMessage.value = draftMap[conversationKey(conversation.id)] || ''
    customerContext.value = null
    contextError.value = null
    contextLastUpdatedAt.value = null
    transferTo.value = conversation.assigned_to ? String(conversation.assigned_to) : ''
    await Promise.all([
      fetchMessages(conversation.id),
      fetchContext(conversation.id),
    ])
    if (customerServiceInboxScopeDisposed || currentSelectionSequence !== conversationSelectionSequence) return
    await fetchConversations()
  }

  const changePage = async (page: number) => {
    if (customerServiceInboxScopeDisposed) return
    pagination.page = Math.max(1, Math.min(page, totalPages.value))
    await fetchConversations()
  }

  const applyFilters = async () => {
    if (customerServiceInboxScopeDisposed) return
    pagination.page = 1
    await fetchConversations()
  }

  const resetFilters = async () => {
    if (customerServiceInboxScopeDisposed) return
    filters.search = ''
    filters.view = 'inbox'
    filters.status = 'all'
    filters.identity = 'all'
    filters.unread = 'all'
    pagination.page = 1
    await fetchConversations()
  }

  return {
    loading,
    conversationListError,
    messagesLoading,
    messagesError,
    messageReadReceiptError,
    contextLoading,
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
    draftMap,
    contextError,
    contextLastUpdatedAt,
    fetchConversations,
    fetchContext,
    fetchAgents,
    fetchMessages,
    appendCustomerServiceMessageToSelectedConversation,
    updateSelectedConversationPreviewAfterSendingMessage,
    refreshInbox,
    selectConversation,
    changePage,
    applyFilters,
    resetFilters,
    clearCurrentDraft,
    clearSelectedCustomerConversationAndInvalidatePendingReads,
  }
}

export default useCustomerServiceInbox

