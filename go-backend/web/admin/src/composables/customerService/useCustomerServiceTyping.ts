import { computed, ref } from 'vue'

export const useCustomerServiceTyping = (options: Record<string, any> = {}) => {
  const selectedConversation = options.selectedConversation || ref(null)
  const replyMessage = options.replyMessage || ref('')
  const canSendTyping = options.canSendTyping || (() => true)
  const sendTyping = options.sendTyping || (() => false)
  const customerTypingByConversation = ref<Record<string, any>>({})
  let agentTypingIdleTimer: number | null = null
  let lastAgentTypingSignalAt = 0
  const customerTypingTimers = new Map<string, number>()

  const customerTypingFor = (conversationID: number | string) => {
    return customerTypingByConversation.value[String(conversationID)] || null
  }

  const selectedCustomerTyping = computed(() => {
    if (!selectedConversation.value?.id) return null
    return customerTypingFor(selectedConversation.value.id)
  })

  const clearCustomerTyping = (conversationID: number | string) => {
    const key = String(conversationID)
    const nextState = { ...customerTypingByConversation.value }
    delete nextState[key]
    customerTypingByConversation.value = nextState

    const timer = customerTypingTimers.get(key)
    if (timer) {
      window.clearTimeout(timer)
      customerTypingTimers.delete(key)
    }
  }

  const clearTypingTimers = () => {
    if (agentTypingIdleTimer) {
      window.clearTimeout(agentTypingIdleTimer)
      agentTypingIdleTimer = null
    }
    customerTypingTimers.forEach((timer) => window.clearTimeout(timer))
    customerTypingTimers.clear()
  }

  const resetAgentTypingState = () => {
    lastAgentTypingSignalAt = 0
  }

  const handleCustomerTypingEvent = (event: Record<string, any>) => {
    if (event?.actor?.kind !== 'customer' || !event.ticket_id) return

    const key = String(event.ticket_id)
    const payload = event.payload || {}
    if (payload.is_typing === false) {
      clearCustomerTyping(key)
      return
    }

    customerTypingByConversation.value = {
      ...customerTypingByConversation.value,
      [key]: {
        active: true,
        displayName: payload.display_name || '客户',
      },
    }

    const existingTimer = customerTypingTimers.get(key)
    if (existingTimer) {
      window.clearTimeout(existingTimer)
    }
    const expiresAt = Date.parse(payload.expires_at || '')
    const timeout = Number.isFinite(expiresAt)
      ? Math.max(1200, Math.min(5000, expiresAt - Date.now()))
      : 3500
    customerTypingTimers.set(key, window.setTimeout(() => {
      clearCustomerTyping(key)
    }, timeout))
  }

  const notifyAgentTyping = async (isTyping = true) => {
    if (!selectedConversation.value?.id || !canSendTyping()) return

    if (isTyping) {
      const now = Date.now()
      if (now - lastAgentTypingSignalAt < 2500) return
      lastAgentTypingSignalAt = now
    } else {
      lastAgentTypingSignalAt = 0
    }

    const sent = await sendTyping(isTyping)
    if (!sent) {
      console.warn('Customer-service conversation WebSocket is not connected for typing')
    }
  }

  const handleReplyTypingInput = () => {
    if (!replyMessage.value.trim()) {
      if (agentTypingIdleTimer) {
        window.clearTimeout(agentTypingIdleTimer)
        agentTypingIdleTimer = null
      }
      notifyAgentTyping(false)
      return
    }

    notifyAgentTyping(true)
    if (agentTypingIdleTimer) {
      window.clearTimeout(agentTypingIdleTimer)
    }
    agentTypingIdleTimer = window.setTimeout(() => {
      agentTypingIdleTimer = null
      notifyAgentTyping(false)
    }, 3500)
  }

  return {
    customerTypingByConversation,
    selectedCustomerTyping,
    clearTypingTimers,
    handleCustomerTypingEvent,
    handleReplyTypingInput,
    notifyAgentTyping,
    resetAgentTypingState,
  }
}

export default useCustomerServiceTyping
