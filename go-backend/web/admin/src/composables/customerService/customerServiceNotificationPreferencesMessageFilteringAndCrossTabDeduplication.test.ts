import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { playCustomerServiceBrowserMessageNotificationSound, showCustomerServiceBrowserMessageDesktopNotification } from '@/modules/customer-service/customerServiceBrowserMessageNotificationEffects'

vi.mock('@/modules/customer-service/customerServiceBrowserMessageNotificationEffects', () => ({
  playCustomerServiceBrowserMessageNotificationSound: vi.fn(),
  showCustomerServiceBrowserMessageDesktopNotification: vi.fn(),
}))

const notificationStorageFixture = new Map<string, string>()
const notificationLocalStorageFixture = {
  getItem: vi.fn((key: string) => notificationStorageFixture.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { notificationStorageFixture.set(key, value) }),
}
const requestBrowserNotificationPermission = vi.fn(async () => 'granted' as NotificationPermission)
const browserNotificationFixture = { permission: 'default', requestPermission: requestBrowserNotificationPermission }
const requestBrowserNotificationLock = vi.fn(async (_name: string, _options: unknown, effect: (lock: object) => Promise<void>) => effect({}))

function createIncomingCustomerMessageNotificationEvent(eventId: string) {
  return { type: 'conversation.message.created', actor: { kind: 'customer' }, event_id: eventId, ticket_id: 12 }
}

beforeEach(() => {
  vi.resetModules()
  vi.clearAllMocks()
  vi.useFakeTimers()
  vi.setSystemTime(new Date(2026, 9, 23, 12, 0))
  notificationStorageFixture.clear()
  notificationLocalStorageFixture.getItem.mockImplementation((key) => notificationStorageFixture.get(key) ?? null)
  notificationLocalStorageFixture.setItem.mockImplementation((key, value) => { notificationStorageFixture.set(key, value) })
  vi.stubGlobal('window', { localStorage: notificationLocalStorageFixture, Notification: browserNotificationFixture, crypto: { randomUUID: () => 'test-tab' }, setTimeout })
  vi.stubGlobal('navigator', { locks: { request: requestBrowserNotificationLock } })
  vi.stubGlobal('Notification', browserNotificationFixture)
})
afterEach(() => { vi.useRealTimers(); vi.unstubAllGlobals() })

describe('customer-service notification settings and delivery boundaries', () => {
  it('loads existing preferences, persists edits, clamps volume and requests permission only after an explicit action', async () => {
    notificationStorageFixture.set('customer-service-notification-settings', JSON.stringify({ soundEnabled: false, soundVolume: 0.7, desktopEnabled: false }))
    const { useCustomerServiceNotificationSettings } = await import('./useCustomerServiceNotificationSettingsAndQuietHours')
    const settings = useCustomerServiceNotificationSettings()
    expect(settings.soundEnabled.value).toBe(false)
    expect(settings.soundVolume.value).toBe(0.7)
    expect(settings.desktopEnabled.value).toBe(false)
    expect(requestBrowserNotificationPermission).not.toHaveBeenCalled()
    settings.setSoundVolume(5)
    settings.setSoundVolume(Infinity)
    expect(settings.soundVolume.value).toBe(1)
    settings.setSoundEnabled(true)
    expect(JSON.parse(notificationStorageFixture.get('customer-service-notification-settings')!)).toMatchObject({ soundEnabled: true, soundVolume: 1 })
    settings.testNotificationSound()
    expect(playCustomerServiceBrowserMessageNotificationSound).toHaveBeenLastCalledWith(1)
    await settings.requestPermission()
    await settings.requestPermission()
    expect(requestBrowserNotificationPermission).toHaveBeenCalledTimes(1)
  })

  it('retains working session preferences when local storage is malformed or inaccessible', async () => {
    notificationStorageFixture.set('customer-service-notification-settings', '{invalid')
    const { useCustomerServiceNotificationSettings } = await import('./useCustomerServiceNotificationSettingsAndQuietHours')
    const settings = useCustomerServiceNotificationSettings()
    expect(settings.soundVolume.value).toBe(0.12)
    notificationLocalStorageFixture.setItem.mockImplementation(() => { throw new Error('storage unavailable') })
    expect(() => settings.setDesktopEnabled(false)).not.toThrow()
    expect(settings.desktopEnabled.value).toBe(false)
  })

  it('ignores noncustomer events and deduplicates the same customer message across local connections', async () => {
    const { notifyCustomerServiceIncomingCustomerMessage } = await import('./notifyCustomerServiceIncomingCustomerMessage')
    const event = createIncomingCustomerMessageNotificationEvent('same-event')
    await notifyCustomerServiceIncomingCustomerMessage({ ...event, actor: { kind: 'agent' } })
    await notifyCustomerServiceIncomingCustomerMessage({ ...event, type: 'conversation.updated' })
    expect(showCustomerServiceBrowserMessageDesktopNotification).not.toHaveBeenCalled()
    await Promise.all([notifyCustomerServiceIncomingCustomerMessage(event), notifyCustomerServiceIncomingCustomerMessage(event)])
    expect(playCustomerServiceBrowserMessageNotificationSound).toHaveBeenCalledTimes(1)
    expect(showCustomerServiceBrowserMessageDesktopNotification).toHaveBeenCalledTimes(1)
    expect(requestBrowserNotificationPermission).not.toHaveBeenCalled()
  })

  it('honors quiet hours and independently applies sound, desktop and focused-conversation preferences', async () => {
    const { useCustomerServiceNotificationSettings } = await import('./useCustomerServiceNotificationSettingsAndQuietHours')
    const { notifyCustomerServiceIncomingCustomerMessage } = await import('./notifyCustomerServiceIncomingCustomerMessage')
    const settings = useCustomerServiceNotificationSettings()
    settings.setQuietHoursEnabled(true)
    vi.setSystemTime(new Date(2026, 9, 23, 23, 30))
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('quiet-event'))
    expect(playCustomerServiceBrowserMessageNotificationSound).not.toHaveBeenCalled()
    expect(showCustomerServiceBrowserMessageDesktopNotification).not.toHaveBeenCalled()
    settings.setQuietHoursEnabled(false)
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('focused-event'), { shouldSilenceDesktopNotification: () => true })
    expect(playCustomerServiceBrowserMessageNotificationSound).toHaveBeenCalledTimes(1)
    expect(showCustomerServiceBrowserMessageDesktopNotification).not.toHaveBeenCalled()
    settings.setSoundEnabled(false)
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('desktop-only-event'))
    expect(playCustomerServiceBrowserMessageNotificationSound).toHaveBeenCalledTimes(1)
    expect(showCustomerServiceBrowserMessageDesktopNotification).toHaveBeenCalledTimes(1)
    settings.setDesktopEnabled(false)
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('disabled-event'))
    expect(showCustomerServiceBrowserMessageDesktopNotification).toHaveBeenCalledTimes(1)
  })

  it('uses a browser lock and shared claim to avoid repeating a message already handled by another tab', async () => {
    const { notifyCustomerServiceIncomingCustomerMessage } = await import('./notifyCustomerServiceIncomingCustomerMessage')
    notificationStorageFixture.set('customer-service-notification-claim:another-tab-event', String(Date.now()))
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('another-tab-event'))
    expect(showCustomerServiceBrowserMessageDesktopNotification).not.toHaveBeenCalled()
    await notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('unclaimed-event'))
    expect(requestBrowserNotificationLock).toHaveBeenCalledTimes(1)
    expect(notificationStorageFixture.has('customer-service-notification-claim:unclaimed-event')).toBe(true)
    expect(showCustomerServiceBrowserMessageDesktopNotification).toHaveBeenCalledTimes(1)
  })

  it('uses BroadcastChannel arbitration when browser locks are unavailable and lets the winning peer deliver the notification', async () => {
    vi.stubGlobal('navigator', {})
    class CustomerServicePeerNotificationClaimChannel {
      onmessage?: (event: { data: { type: string; eventId: string; tabId: string } }) => void
      close() {}
      postMessage(message: { eventId: string }) {
        this.onmessage?.({ data: { type: 'claim', eventId: message.eventId, tabId: 'a-peer-tab' } })
      }
    }
    vi.stubGlobal('BroadcastChannel', CustomerServicePeerNotificationClaimChannel)
    const { notifyCustomerServiceIncomingCustomerMessage } = await import('./notifyCustomerServiceIncomingCustomerMessage')
    const delivery = notifyCustomerServiceIncomingCustomerMessage(createIncomingCustomerMessageNotificationEvent('peer-wins-event'))
    await vi.advanceTimersByTimeAsync(24)
    await delivery
    expect(showCustomerServiceBrowserMessageDesktopNotification).not.toHaveBeenCalled()
    expect(playCustomerServiceBrowserMessageNotificationSound).not.toHaveBeenCalled()
  })
})
