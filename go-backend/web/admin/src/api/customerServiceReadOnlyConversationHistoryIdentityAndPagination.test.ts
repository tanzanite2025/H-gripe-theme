import { beforeEach, describe, expect, it, vi } from 'vitest'
import axios from '@/utils/axios'
import customerServiceApi from './customerService'

vi.mock('@/utils/axios', () => ({ default: { get: vi.fn() } }))

beforeEach(() => vi.resetAllMocks())

describe('read-only conversation history requests', () => {
  it('requests only one selected conversation and one bounded page, with cancellation', async () => {
    const controller = new AbortController()
    const messages = [{ id: 51, conversation_id: 43, content: 'Selected customer only' }]
    vi.mocked(axios.get).mockResolvedValue({ data: { code: 0, data: { messages, total: 51, has_more: false } } })
    await expect(customerServiceApi.listConversationMessageHistoryPage(43, 2, controller.signal)).resolves.toEqual({ messages, total: 51 })
    expect(axios.get).toHaveBeenCalledExactlyOnceWith('/api/admin/customer-service/conversations/43/messages', {
      params: { limit: 50, offset: 50 }, signal: controller.signal, suppressGlobalErrorToast: true,
    })
  })

  it.each([10, undefined, null, '43/../10'])('rejects messages with a mismatched or absent conversation ID: %s', async (conversationId) => {
    vi.mocked(axios.get).mockResolvedValue({ data: { messages: [{ id: 1, conversation_id: conversationId }], total: 1, has_more: false } })
    await expect(customerServiceApi.listConversationMessageHistoryPage(43, 1)).rejects.toThrow('another conversation')
  })

  it.each([
    { messages: [], total: 100, has_more: true },
    { messages: [], total: -1, has_more: false },
    { messages: [{ id: 1, conversation_id: 43 }], total: 0, has_more: false },
    { messages: [{ id: 1, conversation_id: 43 }], total: 100, has_more: false },
  ])('rejects inconsistent pagination instead of silently hiding history', async (payload) => {
    vi.mocked(axios.get).mockResolvedValue({ data: payload })
    await expect(customerServiceApi.listConversationMessageHistoryPage(43, 1)).rejects.toThrow('pagination')
  })

  it.each(['', 0, -1, '43/../10', 4294967296])('rejects an invalid conversation identity before issuing a request: %s', async (conversationId) => {
    await expect(customerServiceApi.listConversationMessageHistoryPage(conversationId, 1)).rejects.toThrow('Invalid conversation')
    expect(axios.get).not.toHaveBeenCalled()
  })
})
