import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const readObjectPayload = (response: unknown, path: string) => (
  requireApiObject(unwrapApiPayload(response, path), path)
)

const readField = <T = unknown>(response: unknown, path: string, field: string): T => {
  const payload = readObjectPayload(response, path)
  return requireApiField<T>(payload, field, path)
}

export const customerServiceApi = {
  async getRegionAnalytics(params: Record<string, any> = {}) {
    const path = '/api/admin/customer-service/analytics/regions'
    return requireApiObject(readField(await axios.get(path, { params }), path, 'analytics'), path, 'field "analytics"')
  },

  async listConversations(params: Record<string, any> = {}) {
    const path = '/api/admin/customer-service/conversations'
    const response = await axios.get(path, { params })
    const body = requireApiObject(response.data, path, 'response body')
    const payload = requireApiObject(unwrapApiPayload(response, path), path)
    requireApiArrayField(payload, 'conversations', path)
    requireApiPagination(body, payload, path)
    return payload
  },

  async getConversationContext(conversationId: number | string) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/context`
    return requireApiObject(readField(await axios.get(path), path, 'context'), path, 'field "context"')
  },

  async listAgents() {
    const path = '/api/admin/customer-service/agents'
    return requireApiArrayField(readObjectPayload(await axios.get(path), path), 'agents', path)
  },

  async listAgentDirectory() {
    const path = '/api/admin/customer-service/agents'
    const payload = readObjectPayload(await axios.get(path), path)
    requireApiArrayField(payload, 'agents', path)
    requireApiArrayField(payload, 'groups', path)
    return payload
  },

  async listGroups() {
    const path = '/api/admin/customer-service/groups'
    return requireApiArrayField(readObjectPayload(await axios.get(path), path), 'groups', path)
  },

  async listAutoReplyRules() {
    const path = '/api/admin/customer-service/auto-reply/rules'
    return requireApiArrayField(readObjectPayload(await axios.get(path), path), 'rules', path)
  },

  async listAutoReplyFAQGroups(params: Record<string, any> = {}) {
    const path = '/api/admin/customer-service/auto-reply/faqs'
    return requireApiArrayField(readObjectPayload(await axios.get(path, { params }), path), 'pages', path)
  },

  async getAutoReplyRule(ruleId: number | string) {
    const path = `/api/admin/customer-service/auto-reply/rules/${ruleId}`
    return requireApiObject(readField(await axios.get(path), path, 'rule'), path, 'field "rule"')
  },

  async createAutoReplyRule(rule: Record<string, any>) {
    const path = '/api/admin/customer-service/auto-reply/rules'
    return requireApiObject(readField(await axios.post(path, rule), path, 'rule'), path, 'field "rule"')
  },

  async updateAutoReplyRule(ruleId: number | string, rule: Record<string, any>) {
    const path = `/api/admin/customer-service/auto-reply/rules/${ruleId}`
    return requireApiObject(readField(await axios.put(path, rule), path, 'rule'), path, 'field "rule"')
  },

  async deleteAutoReplyRule(ruleId: number | string) {
    const path = `/api/admin/customer-service/auto-reply/rules/${ruleId}`
    return requireApiAcknowledgement(await axios.delete(path), path)
  },

  buildEventsUrl(scope = 'inbox') {
    const query = new URLSearchParams({ scope })
    const baseURL = String(axios.defaults?.baseURL || '').replace(/\/$/, '')
    const path = `/api/admin/customer-service/events?${query.toString()}`
    return baseURL ? `${baseURL}${path}` : path
  },

  async listMessages(conversationId: number | string) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/messages`
    return requireApiArrayField(readObjectPayload(await axios.get(path), path), 'messages', path)
  },

  async markMessagesRead(conversationId: number | string) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/messages/mark-read`
    return requireApiAcknowledgement(await axios.post(path), path)
  },

  async sendTyping(conversationId: number | string, isTyping: boolean) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/typing`
    const payload = readObjectPayload(await axios.post(path, { is_typing: isTyping }), path)
    requireApiField(payload, 'typing', path)
    return payload
  },

  async sendMessage(conversationId: number | string, message: string, attachments: string[] = []) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/messages`
    const payload = readObjectPayload(await axios.post(path, { message, attachments }), path)
    return requireApiObjectField(payload, 'message', path)
  },

  async transferConversation(conversationId: number | string, assignedTo: number) {
    const path = `/api/admin/customer-service/conversations/${conversationId}/transfer`
    return requireApiAcknowledgement(await axios.patch(path, { assigned_to: assignedTo }), path)
  }
}

export default customerServiceApi
