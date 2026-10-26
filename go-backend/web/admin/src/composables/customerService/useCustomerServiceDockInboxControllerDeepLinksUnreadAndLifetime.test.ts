import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { effectScope, nextTick, reactive, type EffectScope } from 'vue'
import { toast } from 'vue-sonner'
import customerServiceApi from '@/api/customerService'
import { requestCustomerServiceNotificationPermission } from './useCustomerServiceNotificationSettingsAndQuietHours'
import { useCustomerServiceRealtime } from './useCustomerServiceRealtime'
import { useCustomerServiceDockInboxController } from './useCustomerServiceDockInboxController'

const routeFixture = reactive({ query: {} as Record<string, string> })
const routerFixture = { replace: vi.fn(async () => undefined) }
const connectCustomerServiceRealtime = vi.fn()
const closeCustomerServiceRealtime = vi.fn()
const customerServiceDockTestScopes: EffectScope[] = []

vi.mock('vue-router', () => ({ useRoute: () => routeFixture, useRouter: () => routerFixture }))
vi.mock('vue-sonner', () => ({ toast: { info: vi.fn() } }))
vi.mock('@/api/customerService', () => ({ default: { listConversations: vi.fn(), buildWebSocketUrl: vi.fn() } }))
vi.mock('./useCustomerServiceNotificationSettingsAndQuietHours', () => ({ requestCustomerServiceNotificationPermission: vi.fn() }))
vi.mock('./useCustomerServiceRealtime', () => ({ useCustomerServiceRealtime: vi.fn() }))

function createDisposableCustomerServiceDockInboxController() {
  const scope = effectScope()
  customerServiceDockTestScopes.push(scope)
  const state = scope.run(() => useCustomerServiceDockInboxController())!
  const realtimeCalls = vi.mocked(useCustomerServiceRealtime).mock.calls
  const realtimeOptions = realtimeCalls[realtimeCalls.length - 1][0]
  return { scope, state, realtimeOptions }
}

beforeEach(() => {
  vi.resetAllMocks()
  routeFixture.query = {}
  vi.mocked(customerServiceApi.listConversations).mockResolvedValue({ pagination: { total: 128 } })
  vi.mocked(useCustomerServiceRealtime).mockReturnValue({
    realtimeSocket: { value: null } as any,
    connectCustomerServiceRealtime,
    closeCustomerServiceRealtime,
    sendCustomerServiceRealtimeControl: vi.fn(),
  })
})
afterEach(() => { customerServiceDockTestScopes.splice(0).forEach((scope) => scope.stop()) })

describe('customer-service dock inbox controller', () => {
  it('refreshes its badge and connects in the background, then hands realtime ownership to the open inbox', async () => {
    const { state } = createDisposableCustomerServiceDockInboxController()
    await nextTick()
    await Promise.resolve()
    expect(connectCustomerServiceRealtime).toHaveBeenCalledTimes(1)
    expect(state.unreadBadgeLabel.value).toBe('99+')
    expect(state.inboxAriaLabel.value).toContain('128')
    state.openCustomerServiceInboxFromDock()
    expect(requestCustomerServiceNotificationPermission).toHaveBeenCalledTimes(1)
    expect(state.inboxOpen.value).toBe(true)
    expect(closeCustomerServiceRealtime).toHaveBeenCalledTimes(1)
    state.inboxOpen.value = false
    await nextTick()
    expect(connectCustomerServiceRealtime).toHaveBeenCalledTimes(2)
    expect(routerFixture.replace).not.toHaveBeenCalled()
  })

  it('opens route deep links without connecting the background socket and clears only the inbox query on close', async () => {
    routeFixture.query = { inbox: 'customer-service', filter: 'retained' }
    const { state } = createDisposableCustomerServiceDockInboxController()
    await nextTick()
    expect(state.inboxOpen.value).toBe(true)
    expect(connectCustomerServiceRealtime).not.toHaveBeenCalled()
    expect(requestCustomerServiceNotificationPermission).not.toHaveBeenCalled()
    state.inboxOpen.value = false
    await nextTick()
    expect(routerFixture.replace).toHaveBeenCalledWith({ query: { filter: 'retained' } })
    expect(connectCustomerServiceRealtime).toHaveBeenCalledTimes(1)
  })

  it('does not reconnect or remove the deep link after a close immediately followed by reopening', async () => {
    routeFixture.query = { inbox: 'customer-service' }
    const { state } = createDisposableCustomerServiceDockInboxController()
    state.inboxOpen.value = false
    state.inboxOpen.value = true
    await nextTick()
    expect(connectCustomerServiceRealtime).not.toHaveBeenCalled()
    expect(routerFixture.replace).not.toHaveBeenCalled()
  })

  it('shows customer message toasts while closed and ignores staff messages and messages while open', async () => {
    const { state, realtimeOptions } = createDisposableCustomerServiceDockInboxController()
    await nextTick()
    const customerMessage = { type: 'conversation.message.created', actor: { kind: 'customer' }, event_id: 'event-1' }
    await realtimeOptions.onRefresh!({ ...customerMessage, actor: { kind: 'agent' } })
    expect(toast.info).not.toHaveBeenCalled()
    await realtimeOptions.onRefresh!(customerMessage)
    expect(toast.info).toHaveBeenCalledTimes(1)
    const toastAction = vi.mocked(toast.info).mock.calls[0][1]!.action as { onClick: () => void }
    toastAction.onClick()
    expect(state.inboxOpen.value).toBe(true)
    await realtimeOptions.onRefresh!(customerMessage)
    expect(toast.info).toHaveBeenCalledTimes(1)
  })

  it('prevents deferred startup, toast actions and realtime callbacks from reopening a disposed launcher', async () => {
    const { state, scope, realtimeOptions } = createDisposableCustomerServiceDockInboxController()
    scope.stop()
    await nextTick()
    state.openCustomerServiceInboxFromDock()
    await realtimeOptions.onConnected!()
    await realtimeOptions.onRefresh!({ type: 'conversation.message.created', actor: { kind: 'customer' } })
    expect(connectCustomerServiceRealtime).not.toHaveBeenCalled()
    expect(customerServiceApi.listConversations).not.toHaveBeenCalled()
    expect(toast.info).not.toHaveBeenCalled()
    expect(state.inboxOpen.value).toBe(false)
    expect(closeCustomerServiceRealtime).toHaveBeenCalledTimes(1)
  })
})
