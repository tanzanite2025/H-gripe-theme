import { onScopeDispose, ref } from 'vue'
import customerServiceApi from '@/api/customerService'

export function useCustomerServiceUnreadConversationSummary() {
  const unreadConversationCount = ref(0)
  let unreadRefreshTimer: ReturnType<typeof setTimeout> | undefined
  let activeUnreadRequest: AbortController | undefined
  let unreadRequestGeneration = 0
  let disposed = false

  function cancelScheduledCustomerServiceUnreadRefresh(): void {
    if (unreadRefreshTimer === undefined) return
    clearTimeout(unreadRefreshTimer)
    unreadRefreshTimer = undefined
  }

  async function refreshCustomerServiceUnreadConversationCount(): Promise<void> {
    if (disposed) return
    cancelScheduledCustomerServiceUnreadRefresh()
    activeUnreadRequest?.abort()
    const controller = new AbortController()
    activeUnreadRequest = controller
    const generation = ++unreadRequestGeneration
    try {
      const data = await customerServiceApi.listConversations({ page: 1, page_size: 1, unread: 'true' }, true, controller.signal)
      if (disposed || controller.signal.aborted || generation !== unreadRequestGeneration) return
      const total = Number(data.pagination?.total || 0)
      unreadConversationCount.value = Number.isFinite(total) ? Math.max(0, Math.floor(total)) : 0
    } catch (error) {
      if (disposed || controller.signal.aborted || generation !== unreadRequestGeneration) return
      console.error('Failed to refresh customer-service unread summary:', error)
    }
  }

  function scheduleCustomerServiceUnreadConversationCountRefresh(): void {
    if (disposed) return
    cancelScheduledCustomerServiceUnreadRefresh()
    unreadRefreshTimer = setTimeout(() => {
      unreadRefreshTimer = undefined
      void refreshCustomerServiceUnreadConversationCount()
    }, 250)
  }

  onScopeDispose(() => {
    disposed = true
    unreadRequestGeneration++
    activeUnreadRequest?.abort()
    cancelScheduledCustomerServiceUnreadRefresh()
    unreadConversationCount.value = 0
  })

  return {
    unreadConversationCount,
    refreshCustomerServiceUnreadConversationCount,
    scheduleCustomerServiceUnreadConversationCountRefresh,
  }
}
