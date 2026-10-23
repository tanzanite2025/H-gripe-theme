import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, ref, type EffectScope } from 'vue'
import customerServiceApi, { type CustomerServiceConversationHistoryMessage } from '@/api/customerService'
import type { CustomerConversation } from '@/modules/customer-service/customerServiceTypes'
import { useReadOnlyConversationHistoryWithIsolatedPagination } from './useReadOnlyConversationHistoryWithIsolatedPagination'

vi.mock('@/api/customerService', () => ({
  default: { listConversationMessageHistoryPage: vi.fn() }, customerServiceConversationHistoryPageSize: 50,
}))

type ConversationHistoryPage = { messages: CustomerServiceConversationHistoryMessage[]; total: number }
const scopes: EffectScope[] = []

function createDeferredConversationHistoryPage() {
  let resolve!: (value: ConversationHistoryPage) => void
  let reject!: (reason: unknown) => void
  const promise = new Promise<ConversationHistoryPage>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

function createIsolatedConversationHistoryState() {
  const conversation = ref<CustomerConversation | null>(null)
  const scope = effectScope()
  scopes.push(scope)
  const state = scope.run(() => useReadOnlyConversationHistoryWithIsolatedPagination(conversation))!
  return { conversation, scope, state }
}

beforeEach(() => vi.resetAllMocks())
afterEach(() => scopes.splice(0).forEach((scope) => scope.stop()))

describe('isolated read-only conversation history', () => {
  it('waits for selection, replaces each page, and resets pagination for another customer', async () => {
    const { conversation, state } = createIsolatedConversationHistoryState()
    expect(customerServiceApi.listConversationMessageHistoryPage).not.toHaveBeenCalled()
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockResolvedValue({ messages: [{ id: 1, conversation_id: 43 }], total: 101 })
    conversation.value = { id: 43 }
    await vi.waitFor(() => expect(state.messages.value).toHaveLength(1))
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockResolvedValue({ messages: [{ id: 51, conversation_id: 43 }], total: 101 })
    const pageRequest = state.loadConversationHistoryPage(2)
    expect(state.messages.value).toEqual([])
    await pageRequest
    expect(state.messages.value.map((message) => message.id)).toEqual([51])
    expect(state.pageNumber.value).toBe(2)
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockResolvedValue({ messages: [{ id: 201, conversation_id: 10 }], total: 1 })
    conversation.value = { id: 10 }
    expect(state.messages.value).toEqual([])
    expect(state.pageNumber.value).toBe(1)
    await vi.waitFor(() => expect(state.messages.value[0]?.conversation_id).toBe(10))
    expect(state.totalPages.value).toBe(1)
    expect(vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mock.calls.map(([id, page]) => [id, page])).toEqual([['43', 1], ['43', 2], ['10', 1]])
  })

  it.each(['switch', 'close', 'dispose'])('aborts and ignores a late response after %s', async (cause) => {
    const pending = createDeferredConversationHistoryPage()
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockReturnValueOnce(pending.promise).mockResolvedValue({ messages: [], total: 0 })
    const { conversation, scope, state } = createIsolatedConversationHistoryState()
    conversation.value = { id: 43 }
    const signal = vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mock.calls[0]![2]!
    if (cause === 'switch') conversation.value = { id: 10 }
    if (cause === 'close') conversation.value = null
    if (cause === 'dispose') scope.stop()
    expect(signal.aborted).toBe(true)
    pending.resolve({ messages: [{ id: 1, conversation_id: 43 }], total: 1 })
    await Promise.resolve()
    await Promise.resolve()
    expect(state.messages.value).toEqual([])
    expect(state.totalMessages.value).toBe(0)
    expect(state.loadError.value).toBe('')
    expect(state.loading.value).toBe(false)
  })

  it('ignores a failed old request after reopening the same conversation', async () => {
    const first = createDeferredConversationHistoryPage()
    const second = createDeferredConversationHistoryPage()
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)
    const { conversation, state } = createIsolatedConversationHistoryState()
    conversation.value = { id: 43 }
    conversation.value = null
    conversation.value = { id: 43 }
    first.reject(new Error('Old request failed'))
    await Promise.resolve()
    expect(state.loadError.value).toBe('')
    expect(state.loading.value).toBe(true)
    second.resolve({ messages: [{ id: 2, conversation_id: 43 }], total: 1 })
    await vi.waitFor(() => expect(state.messages.value[0]?.id).toBe(2))
  })

  it('clears the previous page on failure and retries the requested page', async () => {
    const { conversation, state } = createIsolatedConversationHistoryState()
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockResolvedValueOnce({ messages: [{ id: 1, conversation_id: 43 }], total: 51 }).mockRejectedValueOnce(new Error('Offline'))
    conversation.value = { id: 43 }
    await vi.waitFor(() => expect(state.messages.value).toHaveLength(1))
    await state.loadConversationHistoryPage(2)
    expect(state.messages.value).toEqual([])
    expect(state.loadError.value).toContain('加载失败')
    vi.mocked(customerServiceApi.listConversationMessageHistoryPage).mockResolvedValueOnce({ messages: [{ id: 51, conversation_id: 43 }], total: 51 })
    await state.loadConversationHistoryPage(state.pageNumber.value)
    expect(state.loadError.value).toBe('')
    expect(state.messages.value[0]?.id).toBe(51)
  })
})
