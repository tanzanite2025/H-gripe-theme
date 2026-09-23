import { computed, reactive, ref } from 'vue'
import customerServiceApi from '@/api/customerService'
import type {
  AssignableAgent,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerServiceFiltersState,
} from '@/modules/customer-service/customerServiceTypes'

export const useCustomerServiceInbox = () => {
  const loading = ref(false)
  const messagesLoading = ref(false)
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
    const key = conversationKey(conversationID)
    if (!key) return
    delete draftMap[key]
    if (conversationKey(selectedConversation.value?.id) === key) replyMessage.value = ''
  }

  const fetchConversations = async () => {
    loading.value = true
    try {
      const data = await customerServiceApi.listConversations({
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search.trim() || undefined,
        view: filters.view || 'inbox',
        status: filters.status !== 'all' ? filters.status : undefined,
        identity: filters.identity !== 'all' ? filters.identity : undefined,
        unread: filters.unread === 'unread' ? 'true' : undefined,
      })
      conversations.value = data.conversations || []
      pagination.total = data.pagination?.total ?? conversations.value.length

      if (selectedConversation.value) {
        const refreshed = conversations.value.find((item) => Number(item.id) === Number(selectedConversation.value?.id))
        if (refreshed) {
          selectedConversation.value = refreshed
        } else {
          selectedConversation.value = null
          messages.value = []
          customerContext.value = null
          contextError.value = null
          contextLastUpdatedAt.value = null
        }
      }
    } catch (error) {
      console.error('Failed to fetch customer-service conversations:', error)
    } finally {
      loading.value = false
    }
  }

  const fetchContext = async (conversationID: number | string | null | undefined) => {
    if (!conversationID) {
      contextRequestSequence += 1
      customerContext.value = null
      contextError.value = null
      contextLastUpdatedAt.value = null
      return
    }

    const requestSequence = ++contextRequestSequence
    const requestedConversationKey = conversationKey(conversationID)
    contextLoading.value = true
    contextError.value = null
    try {
      const context = await customerServiceApi.getConversationContext(conversationID)
      if (requestSequence !== contextRequestSequence || conversationKey(selectedConversation.value?.id) !== requestedConversationKey) return
      customerContext.value = context
      contextLastUpdatedAt.value = new Date()
    } catch (error) {
      console.error('Failed to fetch customer-service context:', error)
      if (requestSequence !== contextRequestSequence || conversationKey(selectedConversation.value?.id) !== requestedConversationKey) return
      contextError.value = '请求失败，请稍后重试'
    } finally {
      if (requestSequence === contextRequestSequence) contextLoading.value = false
    }
  }

  const fetchAgents = async () => {
    try {
      assignableAgents.value = await customerServiceApi.listAgents()
    } catch (error) {
      console.error('Failed to fetch public chat agents:', error)
      assignableAgents.value = []
    }
  }

  const fetchMessages = async (conversationID: number | string | null | undefined) => {
    if (!conversationID) return
    messagesLoading.value = true
    try {
      messages.value = await customerServiceApi.listMessages(conversationID)
      await customerServiceApi.markMessagesRead(conversationID)
    } catch (error) {
      console.error('Failed to fetch customer-service messages:', error)
      messages.value = []
    } finally {
      messagesLoading.value = false
    }
  }

  const refreshInbox = async () => {
    await Promise.all([fetchConversations(), fetchAgents()])
    if (selectedConversation.value) {
      await Promise.all([
        fetchMessages(selectedConversation.value.id),
        fetchContext(selectedConversation.value.id),
      ])
    }
  }

  const selectConversation = async (conversation: CustomerConversation) => {
    saveCurrentDraft()
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
    await fetchConversations()
  }

  const changePage = async (page: number) => {
    pagination.page = Math.max(1, Math.min(page, totalPages.value))
    await fetchConversations()
  }

  const applyFilters = async () => {
    pagination.page = 1
    await fetchConversations()
  }

  const resetFilters = async () => {
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
    messagesLoading,
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
    refreshInbox,
    selectConversation,
    changePage,
    applyFilters,
    resetFilters,
    clearCurrentDraft,
  }
}

export default useCustomerServiceInbox

