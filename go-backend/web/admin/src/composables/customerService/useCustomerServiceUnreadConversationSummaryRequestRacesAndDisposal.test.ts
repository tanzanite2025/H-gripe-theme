import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import customerServiceApi from '@/api/customerService'
import { useCustomerServiceUnreadConversationSummary } from './useCustomerServiceUnreadConversationSummary'

vi.mock('@/api/customerService', () => ({ default: { listConversations: vi.fn() } }))
const unreadSummaryTestScopes: EffectScope[] = []

function createDisposableCustomerServiceUnreadSummary() {
  const scope = effectScope()
  unreadSummaryTestScopes.push(scope)
  return { scope, state: scope.run(() => useCustomerServiceUnreadConversationSummary())! }
}

function createDeferredCustomerServiceUnreadSummaryResponse() {
  let resolve!: (value: Record<string, any>) => void
  let reject!: (reason: Error) => void
  const promise = new Promise<Record<string, any>>((resolvePromise, rejectPromise) => { resolve = resolvePromise; reject = rejectPromise })
  return { promise, resolve, reject }
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.useFakeTimers()
  vi.spyOn(console, 'error').mockImplementation(() => {})
  vi.mocked(customerServiceApi.listConversations).mockResolvedValue({ pagination: { total: 128 } })
})
afterEach(() => {
  unreadSummaryTestScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
  vi.useRealTimers()
})

describe('customer-service unread conversation summary', () => {
  it('coalesces bursts of events into one background request and cancels pending debounce on direct refresh', async () => {
    const { state } = createDisposableCustomerServiceUnreadSummary()
    state.scheduleCustomerServiceUnreadConversationCountRefresh()
    await vi.advanceTimersByTimeAsync(150)
    state.scheduleCustomerServiceUnreadConversationCountRefresh()
    await vi.advanceTimersByTimeAsync(249)
    expect(customerServiceApi.listConversations).not.toHaveBeenCalled()
    await vi.advanceTimersByTimeAsync(1)
    expect(customerServiceApi.listConversations).toHaveBeenCalledTimes(1)
    expect(state.unreadConversationCount.value).toBe(128)
    state.scheduleCustomerServiceUnreadConversationCountRefresh()
    await state.refreshCustomerServiceUnreadConversationCount()
    await vi.advanceTimersByTimeAsync(500)
    expect(customerServiceApi.listConversations).toHaveBeenCalledTimes(2)
    expect(customerServiceApi.listConversations).toHaveBeenCalledWith({ unread: 'true', page: 1, page_size: 1 }, true, expect.any(AbortSignal))
  })

  it('aborts an older request and ignores its late success and failure', async () => {
    const first = createDeferredCustomerServiceUnreadSummaryResponse()
    const second = createDeferredCustomerServiceUnreadSummaryResponse()
    vi.mocked(customerServiceApi.listConversations).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise)
    const { state } = createDisposableCustomerServiceUnreadSummary()
    const firstRequest = state.refreshCustomerServiceUnreadConversationCount()
    const firstSignal = vi.mocked(customerServiceApi.listConversations).mock.calls[0][2]!
    const secondRequest = state.refreshCustomerServiceUnreadConversationCount()
    expect(firstSignal.aborted).toBe(true)
    second.resolve({ pagination: { total: 3 } })
    await secondRequest
    first.resolve({ pagination: { total: 90 } })
    await firstRequest
    expect(state.unreadConversationCount.value).toBe(3)

    const lateFailure = createDeferredCustomerServiceUnreadSummaryResponse()
    vi.mocked(customerServiceApi.listConversations).mockReturnValueOnce(lateFailure.promise)
    const oldRequest = state.refreshCustomerServiceUnreadConversationCount()
    await state.refreshCustomerServiceUnreadConversationCount()
    lateFailure.reject(new Error('old request failed'))
    await oldRequest
    expect(state.unreadConversationCount.value).toBe(128)
    expect(console.error).not.toHaveBeenCalled()
  })

  it('retains the last confirmed count after a current request fails', async () => {
    const { state } = createDisposableCustomerServiceUnreadSummary()
    await state.refreshCustomerServiceUnreadConversationCount()
    vi.mocked(customerServiceApi.listConversations).mockRejectedValueOnce(new Error('offline'))
    await state.refreshCustomerServiceUnreadConversationCount()
    expect(state.unreadConversationCount.value).toBe(128)
    expect(console.error).toHaveBeenCalledTimes(1)
  })

  it('cancels requests and timers on disposal and prevents subsequent reads or updates', async () => {
    const pending = createDeferredCustomerServiceUnreadSummaryResponse()
    vi.mocked(customerServiceApi.listConversations).mockReturnValueOnce(pending.promise)
    const { state, scope } = createDisposableCustomerServiceUnreadSummary()
    const request = state.refreshCustomerServiceUnreadConversationCount()
    const signal = vi.mocked(customerServiceApi.listConversations).mock.calls[0][2]!
    state.scheduleCustomerServiceUnreadConversationCountRefresh()
    scope.stop()
    expect(signal.aborted).toBe(true)
    pending.resolve({ pagination: { total: 67 } })
    await request
    await vi.advanceTimersByTimeAsync(500)
    await state.refreshCustomerServiceUnreadConversationCount()
    state.scheduleCustomerServiceUnreadConversationCountRefresh()
    await vi.advanceTimersByTimeAsync(500)
    expect(state.unreadConversationCount.value).toBe(0)
    expect(customerServiceApi.listConversations).toHaveBeenCalledTimes(1)
  })
})
