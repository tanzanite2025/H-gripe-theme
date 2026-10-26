import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, type EffectScope } from 'vue'
import { notifyCustomerServiceIncomingCustomerMessage } from './notifyCustomerServiceIncomingCustomerMessage'
import { useCustomerServiceRealtime } from './useCustomerServiceRealtime'

vi.mock('./notifyCustomerServiceIncomingCustomerMessage', () => ({ notifyCustomerServiceIncomingCustomerMessage: vi.fn(async () => undefined) }))

class CustomerServiceTestWebSocket {
  static OPEN = 1
  static CONNECTING = 0
  static CLOSING = 2
  static CLOSED = 3
  static instances: CustomerServiceTestWebSocket[] = []
  readyState = CustomerServiceTestWebSocket.CONNECTING
  onopen?: () => void
  onclose?: () => void
  onmessage?: (event: { data: string }) => void
  onerror?: () => void
  send = vi.fn()
  constructor(public url: string) { CustomerServiceTestWebSocket.instances.push(this) }
  close = vi.fn(() => { this.readyState = CustomerServiceTestWebSocket.CLOSED; this.onclose?.() })
}

const realtimeTransportTestScopes: EffectScope[] = []
const browserEventTargetFixture = new EventTarget()
const documentEventTargetFixture = new EventTarget()

function createDisposableCustomerServiceRealtimeTransport(overrides: Parameters<typeof useCustomerServiceRealtime>[0] = {}) {
  const onRefresh = vi.fn()
  const onConnected = vi.fn()
  const onTyping = vi.fn()
  const scope = effectScope()
  realtimeTransportTestScopes.push(scope)
  const state = scope.run(() => useCustomerServiceRealtime({
    buildWebSocketUrl: (cursor) => `ws://test/inbox?last_event_id=${cursor}`,
    connectionKey: () => 'inbox',
    onRefresh, onConnected, onTyping, ...overrides,
  }))!
  state.connectCustomerServiceRealtime()
  return { state, scope, onRefresh, onConnected, onTyping, socket: CustomerServiceTestWebSocket.instances[CustomerServiceTestWebSocket.instances.length - 1] }
}

function deliverCustomerServiceTestRealtimeEvent(socket: CustomerServiceTestWebSocket, event: Record<string, unknown>, cursor = 'cursor-1') {
  socket.onmessage?.({ data: JSON.stringify({ type: 'event', cursor, event }) })
}

beforeEach(() => {
  vi.resetAllMocks()
  vi.useFakeTimers()
  CustomerServiceTestWebSocket.instances = []
  vi.spyOn(Math, 'random').mockReturnValue(0.5)
  vi.stubGlobal('WebSocket', CustomerServiceTestWebSocket)
  vi.stubGlobal('window', {
    WebSocket: CustomerServiceTestWebSocket,
    setTimeout,
    clearTimeout,
    addEventListener: browserEventTargetFixture.addEventListener.bind(browserEventTargetFixture),
    removeEventListener: browserEventTargetFixture.removeEventListener.bind(browserEventTargetFixture),
  })
  vi.stubGlobal('document', {
    visibilityState: 'visible',
    addEventListener: documentEventTargetFixture.addEventListener.bind(documentEventTargetFixture),
    removeEventListener: documentEventTargetFixture.removeEventListener.bind(documentEventTargetFixture),
  })
  vi.stubGlobal('navigator', { onLine: true })
})
afterEach(() => {
  realtimeTransportTestScopes.splice(0).forEach((scope) => scope.stop())
  vi.restoreAllMocks()
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('customer-service realtime transport', () => {
  it('keeps one socket, reconciles after connect, separates typing and deduplicates message frames', async () => {
    const { state, socket, onConnected, onTyping, onRefresh } = createDisposableCustomerServiceRealtimeTransport()
    state.connectCustomerServiceRealtime()
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(1)
    socket.readyState = CustomerServiceTestWebSocket.OPEN
    socket.onopen?.()
    expect(onConnected).toHaveBeenCalledTimes(1)
    deliverCustomerServiceTestRealtimeEvent(socket, { type: 'conversation.typing', event_id: 'typing-1' })
    expect(onTyping).toHaveBeenCalledTimes(1)
    expect(notifyCustomerServiceIncomingCustomerMessage).not.toHaveBeenCalled()
    const messageEvent = { type: 'conversation.message.created', actor: { kind: 'customer' }, event_id: 'message-1' }
    deliverCustomerServiceTestRealtimeEvent(socket, messageEvent)
    deliverCustomerServiceTestRealtimeEvent(socket, messageEvent)
    expect(notifyCustomerServiceIncomingCustomerMessage).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(350)
    expect(onRefresh).toHaveBeenCalledTimes(1)
    expect(state.sendCustomerServiceRealtimeControl({ type: 'typing', is_typing: true })).toBe(true)
    expect(socket.send).toHaveBeenCalledWith(JSON.stringify({ type: 'typing', is_typing: true }))
  })

  it('resumes from the last cursor on reconnect and resets it when switching conversation scope', async () => {
    let conversationScope = 'conversation:1'
    const { state, socket } = createDisposableCustomerServiceRealtimeTransport({ connectionKey: () => conversationScope })
    deliverCustomerServiceTestRealtimeEvent(socket, { type: 'conversation.updated', event_id: 'update-1' }, 'resume-cursor')
    socket.close()
    await vi.advanceTimersByTimeAsync(1000)
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(2)
    expect(CustomerServiceTestWebSocket.instances[1].url).toContain('last_event_id=resume-cursor')
    conversationScope = 'conversation:2'
    state.connectCustomerServiceRealtime()
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(3)
    expect(CustomerServiceTestWebSocket.instances[2].url).toBe('ws://test/inbox?last_event_id=')
  })

  it('cancels timers and browser recovery listeners on scope disposal and ignores late frames and reconnect attempts', async () => {
    const { state, socket, scope, onRefresh, onConnected } = createDisposableCustomerServiceRealtimeTransport()
    deliverCustomerServiceTestRealtimeEvent(socket, { type: 'conversation.updated', event_id: 'pending-update' })
    scope.stop()
    deliverCustomerServiceTestRealtimeEvent(socket, { type: 'conversation.updated', event_id: 'late-update' })
    socket.onopen?.()
    browserEventTargetFixture.dispatchEvent(new Event('online'))
    documentEventTargetFixture.dispatchEvent(new Event('visibilitychange'))
    state.connectCustomerServiceRealtime()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(1)
    expect(socket.close).toHaveBeenCalledTimes(1)
    expect(onRefresh).not.toHaveBeenCalled()
    expect(onConnected).not.toHaveBeenCalled()
    expect(state.realtimeSocket.value).toBeNull()
    expect(state.sendCustomerServiceRealtimeControl({ type: 'ping' })).toBe(false)
  })

  it('allows explicit pause and resume without retaining a reconnect timer', async () => {
    const { state, socket } = createDisposableCustomerServiceRealtimeTransport()
    socket.close()
    state.closeCustomerServiceRealtime()
    await vi.advanceTimersByTimeAsync(60_000)
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(1)
    state.connectCustomerServiceRealtime()
    expect(CustomerServiceTestWebSocket.instances).toHaveLength(2)
  })
})
