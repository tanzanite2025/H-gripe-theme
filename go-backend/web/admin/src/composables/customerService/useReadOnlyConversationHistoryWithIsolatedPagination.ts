import { computed, onScopeDispose, ref, watch, type Ref } from 'vue'
import { isAxiosError } from 'axios'
import customerServiceApi, {
  customerServiceConversationHistoryPageSize,
  type CustomerServiceConversationHistoryMessage,
} from '@/api/customerService'
import type { CustomerConversation } from '@/modules/customer-service/customerServiceTypes'

export function useReadOnlyConversationHistoryWithIsolatedPagination(conversation: Readonly<Ref<CustomerConversation | null>>) {
  const messages = ref<CustomerServiceConversationHistoryMessage[]>([])
  const loading = ref(false)
  const loadError = ref('')
  const pageNumber = ref(1)
  const totalMessages = ref(0)
  const totalPages = computed(() => Math.max(1, Math.ceil(totalMessages.value / customerServiceConversationHistoryPageSize)))
  let requestGeneration = 0
  let requestController: AbortController | null = null
  let disposed = false

  function cancelPendingConversationHistoryRequest() {
    requestGeneration++
    requestController?.abort()
    requestController = null
  }

  async function loadConversationHistoryPage(requestedPageNumber: number) {
    if (disposed || !conversation.value || !Number.isInteger(requestedPageNumber) || requestedPageNumber < 1 || requestedPageNumber > totalPages.value) return
    cancelPendingConversationHistoryRequest()
    const generation = requestGeneration
    const requestedConversationId = String(conversation.value.id)
    requestController = new AbortController()
    messages.value = []
    loadError.value = ''
    pageNumber.value = requestedPageNumber
    loading.value = true
    try {
      const result = await customerServiceApi.listConversationMessageHistoryPage(requestedConversationId, requestedPageNumber, requestController.signal)
      if (disposed || generation !== requestGeneration || String(conversation.value?.id) !== requestedConversationId) return
      totalMessages.value = result.total
      if (requestedPageNumber > totalPages.value) {
        await loadConversationHistoryPage(totalPages.value)
        return
      }
      messages.value = result.messages
    } catch (error) {
      if (disposed || generation !== requestGeneration) return
      const status = isAxiosError(error) ? error.response?.status : undefined
      loadError.value = status === 403 ? '没有权限查看该会话的聊天记录。'
        : status === 404 ? '该会话已不存在或无法查看。' : '聊天记录加载失败，请重试。'
    } finally {
      if (!disposed && generation === requestGeneration) {
        loading.value = false
        requestController = null
      }
    }
  }

  watch(conversation, () => {
    cancelPendingConversationHistoryRequest()
    messages.value = []
    loadError.value = ''
    totalMessages.value = 0
    pageNumber.value = 1
    loading.value = false
    if (conversation.value) void loadConversationHistoryPage(1)
  }, { immediate: true, flush: 'sync' })

  onScopeDispose(() => {
    disposed = true
    cancelPendingConversationHistoryRequest()
    messages.value = []
    loading.value = false
    loadError.value = ''
    totalMessages.value = 0
  })

  return { messages, loading, loadError, pageNumber, totalMessages, totalPages, loadConversationHistoryPage }
}
