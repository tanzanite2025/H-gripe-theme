const notifiedCustomerMessageIds = new Set<string>()
const notificationClaimChannelName = 'customer-service-notification-claims'
const notificationClaimWaitMs = 24

type CustomerServiceNotificationClaimMessage = {
  type: 'claim'
  eventId: string
  tabId: string
}

type CustomerServiceNotificationChannel = {
  postMessage: (message: CustomerServiceNotificationClaimMessage) => void
  close: () => void
  onmessage: ((event: MessageEvent<CustomerServiceNotificationClaimMessage>) => void) | null
}

const notificationTabId = (() => {
  if (typeof window === 'undefined') return ''
  try {
    const randomUUID = window.crypto?.randomUUID
    if (typeof randomUUID === 'function') return randomUUID.call(window.crypto)
  } catch {
    // Fall back to a timestamp/random identifier in older browsers.
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
})()
let notificationClaimChannel: CustomerServiceNotificationChannel | null = null
const notificationClaimWinners = new Map<string, string>()

export const rememberCustomerServiceMessageNotificationWithinTab = (eventId: unknown): boolean => {
  const normalized = String(eventId || '').trim()
  if (!normalized) return true
  if (notifiedCustomerMessageIds.has(normalized)) return false
  if (notifiedCustomerMessageIds.size >= 2048) notifiedCustomerMessageIds.clear()
  notifiedCustomerMessageIds.add(normalized)
  return true
}

const getCustomerServiceCrossTabNotificationClaimChannel = (): CustomerServiceNotificationChannel | null => {
  if (typeof window === 'undefined' || typeof BroadcastChannel === 'undefined' || !notificationTabId) return null
  if (notificationClaimChannel) return notificationClaimChannel

  const channel = new BroadcastChannel(notificationClaimChannelName) as unknown as CustomerServiceNotificationChannel
  channel.onmessage = (messageEvent) => {
    const message = messageEvent?.data
    if (!message || message.type !== 'claim' || !message.eventId || !message.tabId || message.tabId === notificationTabId) return
    const currentWinner = notificationClaimWinners.get(message.eventId)
    if (!currentWinner || message.tabId < currentWinner) {
      notificationClaimWinners.set(message.eventId, message.tabId)
    }
  }
  notificationClaimChannel = channel
  return channel
}

const waitForCustomerServiceCrossTabNotificationClaimPeers = (): Promise<void> => new Promise((resolve) => {
  if (typeof window === 'undefined') {
    resolve()
    return
  }
  window.setTimeout(resolve, notificationClaimWaitMs)
})

/**
 * Claim an event before producing a cross-tab notification. Web Locks gives
 * us an atomic browser-wide claim where available. BroadcastChannel provides
 * a deterministic short-wait fallback for browsers without Web Locks.
 */
const notificationClaimStoragePrefix = 'customer-service-notification-claim:'
const notificationClaimStorageTtlMs = 60_000

const readCustomerServiceSharedMessageNotificationClaim = (eventId: string): boolean => {
  if (typeof window === 'undefined') return false
  try {
    const claimedAt = Number(window.localStorage.getItem(`${notificationClaimStoragePrefix}${eventId}`) || 0)
    return Number.isFinite(claimedAt) && claimedAt > Date.now() - notificationClaimStorageTtlMs
  } catch {
    return false
  }
}

const writeCustomerServiceSharedMessageNotificationClaim = (eventId: string): void => {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(`${notificationClaimStoragePrefix}${eventId}`, String(Date.now()))
  } catch {
    // Storage can be unavailable in private/restricted browsing contexts.
  }
}

export const runWithCustomerServiceCrossTabMessageNotificationClaim = async (
  eventId: string,
  effect: () => void | Promise<void>,
): Promise<void> => {
  if (!eventId || typeof window === 'undefined' || !notificationTabId) {
    await effect()
    return
  }

  if (readCustomerServiceSharedMessageNotificationClaim(eventId)) return

  const browserLocks = (navigator as Navigator & {
    locks?: {
      request: (
        name: string,
        options: { ifAvailable: boolean },
        callback: (lock: unknown) => Promise<void> | void,
      ) => Promise<void>
    }
  }).locks

  if (browserLocks?.request) {
    await browserLocks.request(`customer-service-notification:${eventId}`, { ifAvailable: true }, async (lock) => {
      if (!lock) return
      if (readCustomerServiceSharedMessageNotificationClaim(eventId)) return
      writeCustomerServiceSharedMessageNotificationClaim(eventId)
      await effect()
    })
    return
  }

  const channel = getCustomerServiceCrossTabNotificationClaimChannel()
  if (!channel) {
    await effect()
    return
  }

  const currentWinner = notificationClaimWinners.get(eventId)
  if (currentWinner && currentWinner < notificationTabId) return
  notificationClaimWinners.set(eventId, notificationTabId)
  channel.postMessage({ type: 'claim', eventId, tabId: notificationTabId })
  await waitForCustomerServiceCrossTabNotificationClaimPeers()
  if (notificationClaimWinners.get(eventId) !== notificationTabId || readCustomerServiceSharedMessageNotificationClaim(eventId)) return
  writeCustomerServiceSharedMessageNotificationClaim(eventId)
  await effect()
  window.setTimeout(() => notificationClaimWinners.delete(eventId), notificationClaimStorageTtlMs)
}
