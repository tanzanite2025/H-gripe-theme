import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const customerServiceApi = {
  async getRegionAnalytics(params: Record<string, any> = {}) {
    const payload = unwrapPayload(await axios.get('/api/admin/customer-service/analytics/regions', { params }))
    return payload.analytics ?? null
  },

  async listConversations(params: Record<string, any> = {}) {
    return unwrapPayload(await axios.get('/api/admin/customer-service/conversations', { params }))
  },

  async getConversationContext(conversationId: number | string) {
    const payload = unwrapPayload(await axios.get(`/api/admin/customer-service/conversations/${conversationId}/context`))
    return payload.context ?? null
  },

  async listAgents() {
    const payload = unwrapPayload(await axios.get('/api/admin/customer-service/agents'))
    return payload.agents ?? []
  },

  buildEventsUrl(scope = 'inbox') {
    const query = new URLSearchParams({ scope })
    const baseURL = String(axios.defaults?.baseURL || '').replace(/\/$/, '')
    const path = `/api/admin/customer-service/events?${query.toString()}`
    return baseURL ? `${baseURL}${path}` : path
  },

  async listMessages(conversationId: number | string) {
    const payload = unwrapPayload(await axios.get(`/api/admin/customer-service/conversations/${conversationId}/messages`))
    return payload.messages ?? []
  },

  async markMessagesRead(conversationId: number | string) {
    return unwrapPayload(await axios.post(`/api/admin/customer-service/conversations/${conversationId}/messages/mark-read`))
  },

  async sendTyping(conversationId: number | string, isTyping: boolean) {
    return unwrapPayload(await axios.post(`/api/admin/customer-service/conversations/${conversationId}/typing`, {
      is_typing: isTyping
    }))
  },

  async sendMessage(conversationId: number | string, message: string) {
    return unwrapPayload(await axios.post(`/api/admin/customer-service/conversations/${conversationId}/messages`, { message }))
  },

  async transferConversation(conversationId: number | string, assignedTo: number) {
    return unwrapPayload(await axios.patch(`/api/admin/customer-service/conversations/${conversationId}/transfer`, {
      assigned_to: assignedTo
    }))
  }
}

export default customerServiceApi
