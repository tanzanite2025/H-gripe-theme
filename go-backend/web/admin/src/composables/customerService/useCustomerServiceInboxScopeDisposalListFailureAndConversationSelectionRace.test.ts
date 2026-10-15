import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import customerServiceApi from '@/api/customerService'
import { useCustomerServiceInbox } from './useCustomerServiceInbox'
import type { CustomerConversation, CustomerConversationMessage } from '@/modules/customer-service/customerServiceTypes'

vi.mock('@/api/customerService', () => ({
  default: {
    listConversations: vi.fn(), listAgents: vi.fn(), listMessages: vi.fn(),
    getConversationContext: vi.fn(), markMessagesRead: vi.fn(),
  },
}))

const customerConversationFixtures: CustomerConversation[] = [
  { id: 12, ticket_number: 'CONVERSATION-12', assigned_to: 1 },
  { id: 13, ticket_number: 'CONVERSATION-13', assigned_to: 1 },
]
const testScopes: EffectScope[] = []

function createCustomerServiceInboxWithDisposableScope() {
  const scope = effectScope()
  testScopes.push(scope)
  const inbox = scope.run(() => useCustomerServiceInbox())!
  return { scope, inbox }
}

function createDeferredCustomerServiceResponse<T>() {
  let resolve!: (value: T) => void
  let reject!: (reason: Error) => void
  const promise = new Promise<T>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise
    reject = rejectPromise
  })
  return { promise, resolve, reject }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  vi.mocked(customerServiceApi.listConversations).mockResolvedValue({
    conversations: customerConversationFixtures, pagination: { total: 2 },
  })
  vi.mocked(customerServiceApi.listAgents).mockResolvedValue([])
  vi.mocked(customerServiceApi.listMessages).mockResolvedValue([])
  vi.mocked(customerServiceApi.getConversationContext).mockResolvedValue({})
  vi.mocked(customerServiceApi.markMessagesRead).mockResolvedValue(undefined)
})

afterEach(() => {
  testScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
})

describe('customer-service inbox lifetime, list failures and selection races', () => {
  it('preserves successfully loaded history when updating the read receipt fails', async () => {
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    const customerMessages: CustomerConversationMessage[] = [{ id: 91, content: 'Unread customer question' }]
    vi.mocked(customerServiceApi.listMessages).mockResolvedValueOnce(customerMessages)
    vi.mocked(customerServiceApi.markMessagesRead).mockRejectedValueOnce(new Error('Read receipt unavailable'))
    await inbox.fetchMessages(12)
    expect(inbox.messages.value).toEqual(customerMessages)
    expect(inbox.messagesError.value).toBeNull()
    expect(inbox.messageReadReceiptError.value).toContain('无需因此重复回复')
    vi.mocked(customerServiceApi.listMessages).mockResolvedValueOnce(customerMessages)
    await inbox.fetchMessages(12)
    expect(inbox.messageReadReceiptError.value).toBeNull()
  })

  it('distinguishes unavailable message history from successfully loaded empty history', async () => {
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    vi.mocked(customerServiceApi.listMessages).mockRejectedValueOnce(new Error('Message history unavailable'))
    await inbox.fetchMessages(12)
    expect(inbox.messagesError.value).toContain('无法读取')
    expect(customerServiceApi.markMessagesRead).not.toHaveBeenCalled()
    await inbox.fetchMessages(12)
    expect(inbox.messages.value).toEqual([])
    expect(inbox.messagesError.value).toBeNull()
  })

  it('does not update messages or mark them read when a pending read finishes after disposal', async () => {
    const pendingMessages = createDeferredCustomerServiceResponse<CustomerConversationMessage[]>()
    vi.mocked(customerServiceApi.listMessages).mockReturnValueOnce(pendingMessages.promise)
    const { scope, inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    const request = inbox.fetchMessages(12)
    scope.stop()
    pendingMessages.resolve([{ id: 91, content: 'Late customer message' }])
    await request
    expect(inbox.messages.value).toEqual([])
    expect(customerServiceApi.markMessagesRead).not.toHaveBeenCalled()
    await inbox.fetchMessages(12)
    expect(customerServiceApi.listMessages).toHaveBeenCalledTimes(1)
  })

  it('does not start detail reads or apply late list and agent responses after disposal', async () => {
    const pendingList = createDeferredCustomerServiceResponse<Record<string, unknown>>()
    const pendingAgents = createDeferredCustomerServiceResponse<unknown[]>()
    vi.mocked(customerServiceApi.listConversations).mockReturnValueOnce(pendingList.promise)
    vi.mocked(customerServiceApi.listAgents).mockReturnValueOnce(pendingAgents.promise)
    const { scope, inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    const refresh = inbox.refreshInbox()
    scope.stop()
    pendingList.resolve({ conversations: customerConversationFixtures, pagination: { total: 2 } })
    pendingAgents.resolve([{ id: 22 }])
    expect(await refresh).toBe(false)
    expect(inbox.conversations.value).toEqual([])
    expect(inbox.assignableAgents.value).toEqual([])
    expect(customerServiceApi.listMessages).not.toHaveBeenCalled()
    expect(customerServiceApi.getConversationContext).not.toHaveBeenCalled()
  })

  it('represents a current list failure explicitly, clears obsolete data and preserves the reply draft', async () => {
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    await inbox.fetchConversations()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    inbox.messages.value = [{ id: 91, content: 'Previously loaded message' }]
    inbox.replyMessage.value = 'Unsent customer reply'
    vi.mocked(customerServiceApi.listConversations).mockRejectedValueOnce(new Error('List unavailable'))
    expect(await inbox.refreshInbox()).toBe(false)
    expect(inbox.conversationListError.value).toContain('无法确认')
    expect(inbox.conversations.value).toEqual([])
    expect(inbox.selectedConversation.value).toBeNull()
    expect(inbox.messages.value).toEqual([])
    expect(inbox.customerContext.value).toBeNull()
    expect(inbox.replyMessage.value).toBe('')
    expect(inbox.draftMap['12']).toBe('Unsent customer reply')
    expect(inbox.loading.value).toBe(false)
    expect(customerServiceApi.listMessages).not.toHaveBeenCalled()
    expect(customerServiceApi.getConversationContext).not.toHaveBeenCalled()
    await inbox.selectConversation(customerConversationFixtures[0])
    expect(inbox.selectedConversation.value).toBeNull()
    expect(await inbox.fetchConversations()).toBe(true)
    expect(inbox.conversationListError.value).toBeNull()
    await inbox.selectConversation(customerConversationFixtures[0])
    expect(inbox.replyMessage.value).toBe('Unsent customer reply')
  })

  it('ignores an old list failure after a newer list has succeeded', async () => {
    const pendingOldList = createDeferredCustomerServiceResponse<Record<string, unknown>>()
    vi.mocked(customerServiceApi.listConversations).mockReturnValueOnce(pendingOldList.promise)
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    const oldRequest = inbox.fetchConversations()
    expect(await inbox.fetchConversations({ showLoadingIndicator: false })).toBe(true)
    expect(inbox.loading.value).toBe(false)
    pendingOldList.reject(new Error('Stale failure'))
    expect(await oldRequest).toBe(false)
    expect(inbox.conversationListError.value).toBeNull()
    expect(inbox.conversations.value).toEqual(customerConversationFixtures)
  })

  it('prevents old selection completion from marking another conversation read or refreshing the list', async () => {
    const pendingOldMessages = createDeferredCustomerServiceResponse<CustomerConversationMessage[]>()
    vi.mocked(customerServiceApi.listMessages).mockReturnValueOnce(pendingOldMessages.promise)
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    const oldSelection = inbox.selectConversation(customerConversationFixtures[0])
    inbox.replyMessage.value = 'Conversation 12 draft'
    await inbox.selectConversation(customerConversationFixtures[1])
    expect(customerServiceApi.listConversations).toHaveBeenCalledTimes(1)
    pendingOldMessages.resolve([{ id: 91, content: 'Old conversation message' }])
    await oldSelection
    expect(inbox.selectedConversation.value?.id).toBe(13)
    expect(inbox.messages.value).toEqual([])
    expect(inbox.draftMap['12']).toBe('Conversation 12 draft')
    expect(customerServiceApi.markMessagesRead).toHaveBeenCalledTimes(1)
    expect(customerServiceApi.markMessagesRead).toHaveBeenCalledWith(13, true)
    expect(customerServiceApi.listConversations).toHaveBeenCalledTimes(1)
  })

  it('invalidates pending message and context reads when the list loses the selected conversation', async () => {
    const pendingMessages = createDeferredCustomerServiceResponse<CustomerConversationMessage[]>()
    const pendingContext = createDeferredCustomerServiceResponse<Record<string, unknown>>()
    vi.mocked(customerServiceApi.listMessages).mockReturnValueOnce(pendingMessages.promise)
    vi.mocked(customerServiceApi.getConversationContext).mockReturnValueOnce(pendingContext.promise)
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    const messageRead = inbox.fetchMessages(12)
    const contextRead = inbox.fetchContext(12)
    vi.mocked(customerServiceApi.listConversations).mockResolvedValueOnce({ conversations: [], pagination: { total: 0 } })
    await inbox.fetchConversations()
    pendingMessages.resolve([{ id: 91 }])
    pendingContext.resolve({ customer: { id: 99 } })
    await Promise.all([messageRead, contextRead])
    expect(inbox.selectedConversation.value).toBeNull()
    expect(inbox.messages.value).toEqual([])
    expect(inbox.customerContext.value).toBeNull()
    expect(inbox.messagesLoading.value).toBe(false)
    expect(inbox.contextLoading.value).toBe(false)
    expect(customerServiceApi.markMessagesRead).not.toHaveBeenCalled()
  })

  it('keeps the latest agent response when overlapping requests finish out of order', async () => {
    const pendingOldAgents = createDeferredCustomerServiceResponse<unknown[]>()
    vi.mocked(customerServiceApi.listAgents).mockReturnValueOnce(pendingOldAgents.promise)
    vi.mocked(customerServiceApi.listAgents).mockResolvedValueOnce([{ id: 2 }])
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    const oldRequest = inbox.fetchAgents()
    await inbox.fetchAgents()
    pendingOldAgents.resolve([{ id: 1 }])
    await oldRequest
    expect(inbox.assignableAgents.value).toEqual([{ id: 2 }])
  })

  it('keeps previous history on a failed background refresh and explicitly identifies it as potentially stale', async () => {
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    inbox.messages.value = [{ id: 91, content: 'Previously loaded history' }]
    vi.mocked(customerServiceApi.listMessages).mockRejectedValueOnce(new Error('Background history refresh failed'))
    await inbox.fetchMessages(12, { clearMessagesWhenRequestFails: false, showLoadingIndicator: false })
    expect(inbox.messages.value).toEqual([{ id: 91, content: 'Previously loaded history' }])
    expect(inbox.messagesError.value).toContain('可能不是最新记录')
    expect(customerServiceApi.markMessagesRead).not.toHaveBeenCalled()
  })

  it('does not apply a late read-receipt failure to a newer conversation', async () => {
    const pendingReceipt = createDeferredCustomerServiceResponse<Record<string, unknown>>()
    vi.mocked(customerServiceApi.markMessagesRead).mockReturnValueOnce(pendingReceipt.promise)
    const { inbox } = createCustomerServiceInboxWithDisposableScope()
    inbox.selectedConversation.value = customerConversationFixtures[0]
    const oldRead = inbox.fetchMessages(12)
    await vi.waitFor(() => expect(customerServiceApi.markMessagesRead).toHaveBeenCalledWith(12, true))
    expect(inbox.messagesLoading.value).toBe(false)
    await inbox.selectConversation(customerConversationFixtures[1])
    pendingReceipt.reject(new Error('Old read receipt failed'))
    await oldRead
    expect(inbox.selectedConversation.value?.id).toBe(13)
    expect(inbox.messagesError.value).toBeNull()
    expect(inbox.messageReadReceiptError.value).toBeNull()
  })
})
