import { computed, reactive, ref } from 'vue'
import customerServiceApi from '@/api/customerService'

const todayLocalDate = () => {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const adminTimezoneOffsetMinutes = () => -new Date().getTimezoneOffset()

export const useCustomerServiceInbox = () => {
  const loading = ref(false)
  const messagesLoading = ref(false)
  const contextLoading = ref(false)
  const regionAnalyticsLoading = ref(false)
  const conversations = ref<any[]>([])
  const messages = ref<any[]>([])
  const customerContext = ref<Record<string, any> | null>(null)
  const regionAnalytics = ref<Record<string, any> | null>(null)
  const selectedConversation = ref<Record<string, any> | null>(null)
  const replyMessage = ref('')
  const transferTo = ref('')
  const assignableAgents = ref<any[]>([])
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
  const filters = reactive({ search: '', status: 'all', identity: 'all', assignedTo: 'all', unread: 'all' })

  const totalPages = computed(() => Math.max(1, Math.ceil((pagination.total || 0) / pagination.pageSize)))

  const filteredConversations = computed(() => conversations.value)

  const fetchRegionAnalytics = async () => {
    regionAnalyticsLoading.value = true
    try {
      regionAnalytics.value = await customerServiceApi.getRegionAnalytics({
        date: todayLocalDate(),
        tz_offset_minutes: adminTimezoneOffsetMinutes(),
      })
    } catch (error) {
      console.error('Failed to fetch customer-service region analytics:', error)
      regionAnalytics.value = null
    } finally {
      regionAnalyticsLoading.value = false
    }
  }

  const fetchConversations = async () => {
    loading.value = true
    try {
      const data = await customerServiceApi.listConversations({
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search.trim() || undefined,
        status: filters.status !== 'all' ? filters.status : undefined,
        identity: filters.identity !== 'all' ? filters.identity : undefined,
        assigned_to: filters.assignedTo !== 'all' ? filters.assignedTo : undefined,
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
      customerContext.value = null
      return
    }

    contextLoading.value = true
    try {
      customerContext.value = await customerServiceApi.getConversationContext(conversationID)
    } catch (error) {
      console.error('Failed to fetch customer-service context:', error)
      customerContext.value = null
    } finally {
      contextLoading.value = false
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
    await Promise.all([fetchConversations(), fetchAgents(), fetchRegionAnalytics()])
    if (selectedConversation.value) {
      await Promise.all([
        fetchMessages(selectedConversation.value.id),
        fetchContext(selectedConversation.value.id),
      ])
    }
  }

  const selectConversation = async (conversation: Record<string, any>) => {
    selectedConversation.value = conversation
    replyMessage.value = ''
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
    filters.status = 'all'
    filters.identity = 'all'
    filters.assignedTo = 'all'
    filters.unread = 'all'
    pagination.page = 1
    await fetchConversations()
  }

  return {
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
    fetchAgents,
    fetchMessages,
    refreshInbox,
    selectConversation,
    changePage,
    applyFilters,
    resetFilters,
  }
}

export default useCustomerServiceInbox
