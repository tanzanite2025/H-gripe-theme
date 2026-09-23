import { ref } from 'vue'

type CustomerServiceRealtimeControl = {
  type: 'ping' | 'typing'
  is_typing?: boolean
}

const notifiedCustomerMessageIds = new Set<string>()
let notificationPermissionRequestStarted = false
let notificationAudioContext: AudioContext | null = null
const notificationSoundEnabled = ref(true)
const notificationSoundVolume = ref(0.12)
const desktopNotificationEnabled = ref(true)
const suppressDesktopNotificationWhenFocused = ref(true)
const quietHoursEnabled = ref(false)
const quietHoursStart = ref('22:00')
const quietHoursEnd = ref('08:00')
const notificationPermission = ref<NotificationPermission | null>(null)
const notificationSettingsStorageKey = 'customer-service-notification-settings'
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

const loadNotificationSettings = (): void => {
  if (typeof window === 'undefined') return
  try {
    const stored = JSON.parse(window.localStorage.getItem(notificationSettingsStorageKey) || '{}') as Record<string, unknown>
    if (typeof stored.soundEnabled === 'boolean') notificationSoundEnabled.value = stored.soundEnabled
    if (typeof stored.soundVolume === 'number' && Number.isFinite(stored.soundVolume)) {
      notificationSoundVolume.value = Math.min(1, Math.max(0, stored.soundVolume))
    }
    if (typeof stored.desktopEnabled === 'boolean') desktopNotificationEnabled.value = stored.desktopEnabled
    if (typeof stored.suppressDesktopWhenFocused === 'boolean') {
      suppressDesktopNotificationWhenFocused.value = stored.suppressDesktopWhenFocused
    }
    if (typeof stored.quietHoursEnabled === 'boolean') quietHoursEnabled.value = stored.quietHoursEnabled
    if (typeof stored.quietHoursStart === 'string' && /^\d{2}:\d{2}$/.test(stored.quietHoursStart)) quietHoursStart.value = stored.quietHoursStart
    if (typeof stored.quietHoursEnd === 'string' && /^\d{2}:\d{2}$/.test(stored.quietHoursEnd)) quietHoursEnd.value = stored.quietHoursEnd
  } catch {
    // Ignore malformed local preferences and use defaults.
  }
  if ('Notification' in window) notificationPermission.value = Notification.permission
}

const persistNotificationSettings = (): void => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(notificationSettingsStorageKey, JSON.stringify({
    soundEnabled: notificationSoundEnabled.value,
    soundVolume: notificationSoundVolume.value,
    desktopEnabled: desktopNotificationEnabled.value,
    suppressDesktopWhenFocused: suppressDesktopNotificationWhenFocused.value,
    quietHoursEnabled: quietHoursEnabled.value,
    quietHoursStart: quietHoursStart.value,
    quietHoursEnd: quietHoursEnd.value,
  }))
}

loadNotificationSettings()

const rememberCustomerMessageNotification = (eventId: unknown): boolean => {
  const normalized = String(eventId || '').trim()
  if (!normalized) return true
  if (notifiedCustomerMessageIds.has(normalized)) return false
  if (notifiedCustomerMessageIds.size >= 2048) notifiedCustomerMessageIds.clear()
  notifiedCustomerMessageIds.add(normalized)
  return true
}

const getNotificationAudioContext = (): AudioContext | null => {
  if (typeof window === 'undefined') return null
  const AudioContextConstructor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!AudioContextConstructor) return null
  if (!notificationAudioContext) notificationAudioContext = new AudioContextConstructor()
  return notificationAudioContext
}

const unlockNotificationAudio = (): void => {
  const context = getNotificationAudioContext()
  if (context?.state === 'suspended') void context.resume().catch(() => undefined)
}

if (typeof window !== 'undefined') {
  window.addEventListener('pointerdown', unlockNotificationAudio, { passive: true })
  window.addEventListener('keydown', unlockNotificationAudio, { passive: true })
}

const playCustomerMessageNotificationSound = (): void => {
  const context = getNotificationAudioContext()
  if (!context || context.state === 'suspended' || notificationSoundVolume.value <= 0) return

  const now = context.currentTime
  const gain = context.createGain()
  const oscillator = context.createOscillator()
  oscillator.type = 'sine'
  oscillator.frequency.setValueAtTime(880, now)
  oscillator.frequency.setValueAtTime(660, now + 0.12)
  gain.gain.setValueAtTime(0.0001, now)
  gain.gain.exponentialRampToValueAtTime(notificationSoundVolume.value, now + 0.012)
  gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.3)
  oscillator.connect(gain)
  gain.connect(context.destination)
  oscillator.start(now)
  oscillator.stop(now + 0.3)
}

const sendCustomerMessageDesktopNotification = (event: Record<string, any>): void => {
  if (typeof window === 'undefined' || !('Notification' in window)) return

  const show = () => {
    if (Notification.permission !== 'granted') return
    try {
      const notification = new Notification('收到新的客户消息', {
        body: '客服会话中有一条新消息，请及时处理。',
        tag: `customer-service-${event.event_id || event.ticket_id || 'message'}`,
      })
      notification.onclick = () => {
        window.focus()
        notification.close()
      }
    } catch {
      // Notification can still be unavailable in restricted browser contexts.
    }
  }

  if (Notification.permission === 'granted') {
    show()
    return
  }
  // Permission prompts must follow a deliberate user gesture. The first
  // realtime event only plays the sound; the UI can call requestPermission
  // from its notification-settings control when the agent opts in.
}

export const isCustomerServiceWithinQuietHours = (
  now: Date = new Date(),
  enabled = quietHoursEnabled.value,
  startValue = quietHoursStart.value,
  endValue = quietHoursEnd.value,
): boolean => {
  if (!enabled) return false
  if (!/^\d{2}:\d{2}$/.test(startValue) || !/^\d{2}:\d{2}$/.test(endValue)) return false
  const [startHour, startMinute] = startValue.split(':').map(Number)
  const [endHour, endMinute] = endValue.split(':').map(Number)
  if (startHour > 23 || endHour > 23 || startMinute > 59 || endMinute > 59) return false
  const start = startHour * 60 + startMinute
  const end = endHour * 60 + endMinute
  const current = now.getHours() * 60 + now.getMinutes()
  if (start === end) return true
  return start < end ? current >= start && current < end : current >= start || current < end
}

const isWithinQuietHours = (): boolean => {
  return isCustomerServiceWithinQuietHours()
}

const getNotificationClaimChannel = (): CustomerServiceNotificationChannel | null => {
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

const waitForNotificationClaimPeers = (): Promise<void> => new Promise((resolve) => {
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

const readSharedNotificationClaim = (eventId: string): boolean => {
  if (typeof window === 'undefined') return false
  try {
    const claimedAt = Number(window.localStorage.getItem(`${notificationClaimStoragePrefix}${eventId}`) || 0)
    return Number.isFinite(claimedAt) && claimedAt > Date.now() - notificationClaimStorageTtlMs
  } catch {
    return false
  }
}

const writeSharedNotificationClaim = (eventId: string): void => {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(`${notificationClaimStoragePrefix}${eventId}`, String(Date.now()))
  } catch {
    // Storage can be unavailable in private/restricted browsing contexts.
  }
}

const runWithCustomerMessageNotificationClaim = async (
  eventId: string,
  effect: () => void | Promise<void>,
): Promise<void> => {
  if (!eventId || typeof window === 'undefined' || !notificationTabId) {
    await effect()
    return
  }

  if (readSharedNotificationClaim(eventId)) return

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
      if (readSharedNotificationClaim(eventId)) return
      writeSharedNotificationClaim(eventId)
      await effect()
    })
    return
  }

  const channel = getNotificationClaimChannel()
  if (!channel) {
    await effect()
    return
  }

  const currentWinner = notificationClaimWinners.get(eventId)
  if (currentWinner && currentWinner < notificationTabId) return
  notificationClaimWinners.set(eventId, notificationTabId)
  channel.postMessage({ type: 'claim', eventId, tabId: notificationTabId })
  await waitForNotificationClaimPeers()
  if (notificationClaimWinners.get(eventId) !== notificationTabId || readSharedNotificationClaim(eventId)) return
  writeSharedNotificationClaim(eventId)
  await effect()
  window.setTimeout(() => notificationClaimWinners.delete(eventId), notificationClaimStorageTtlMs)
}

export const requestCustomerServiceNotificationPermission = async (): Promise<NotificationPermission | null> => {
  if (typeof window === 'undefined' || !('Notification' in window)) return null
  notificationPermission.value = Notification.permission
  if (Notification.permission !== 'default' || notificationPermissionRequestStarted) return notificationPermission.value
  notificationPermissionRequestStarted = true
  try {
    notificationPermission.value = await Notification.requestPermission()
    return notificationPermission.value
  } catch {
    notificationPermission.value = Notification.permission
    return notificationPermission.value
  }
}

export const useCustomerServiceNotificationSettings = () => {
  const setSoundEnabled = (enabled: boolean): void => {
    notificationSoundEnabled.value = enabled
    persistNotificationSettings()
  }

  const setSoundVolume = (volume: number): void => {
    if (!Number.isFinite(volume)) return
    notificationSoundVolume.value = Math.min(1, Math.max(0, volume))
    persistNotificationSettings()
  }

  const setDesktopEnabled = (enabled: boolean): void => {
    desktopNotificationEnabled.value = enabled
    persistNotificationSettings()
  }

  const setSuppressDesktopWhenFocused = (enabled: boolean): void => {
    suppressDesktopNotificationWhenFocused.value = enabled
    persistNotificationSettings()
  }

  const setQuietHoursEnabled = (enabled: boolean): void => {
    quietHoursEnabled.value = enabled
    persistNotificationSettings()
  }

  const setQuietHoursStart = (value: string): void => {
    if (!/^\d{2}:\d{2}$/.test(value)) return
    quietHoursStart.value = value
    persistNotificationSettings()
  }

  const setQuietHoursEnd = (value: string): void => {
    if (!/^\d{2}:\d{2}$/.test(value)) return
    quietHoursEnd.value = value
    persistNotificationSettings()
  }

  const testNotificationSound = (): void => {
    playCustomerMessageNotificationSound()
  }

  return {
    soundEnabled: notificationSoundEnabled,
    soundVolume: notificationSoundVolume,
    desktopEnabled: desktopNotificationEnabled,
    suppressDesktopWhenFocused: suppressDesktopNotificationWhenFocused,
    quietHoursEnabled,
    quietHoursStart,
    quietHoursEnd,
    notificationPermission,
    setSoundEnabled,
    setSoundVolume,
    setDesktopEnabled,
    setSuppressDesktopWhenFocused,
    setQuietHoursEnabled,
    setQuietHoursStart,
    setQuietHoursEnd,
    testNotificationSound,
    requestPermission: requestCustomerServiceNotificationPermission,
  }
}

const notifyCustomerMessage = async (
  event: Record<string, any>,
  options: { shouldSilenceDesktopNotification?: (event: Record<string, any>) => boolean } = {},
): Promise<void> => {
  if (event.type !== 'conversation.message.created' || event.actor?.kind !== 'customer') return
  const eventId = String(event.event_id || '').trim()
  if (!rememberCustomerMessageNotification(eventId)) return
  await runWithCustomerMessageNotificationClaim(eventId, () => {
    if (!isWithinQuietHours() && notificationSoundEnabled.value) playCustomerMessageNotificationSound()
    if (
      desktopNotificationEnabled.value
      && !isWithinQuietHours()
      && !(options.shouldSilenceDesktopNotification?.(event) ?? false)
    ) {
      sendCustomerMessageDesktopNotification(event)
    }
  })
}

export const useCustomerServiceRealtime = (options: Record<string, any>) => {
  const realtimeSocket = ref<WebSocket | null>(null)
  let realtimeReconnectTimer: number | null = null
  let realtimeRefreshTimer: number | null = null
  let lastEventId = ''
  let cursorScope = ''
  let activeSocketScope = ''
  let reconnectEnabled = false
  let reconnectAttempt = 0
  let browserRecoveryListenersAttached = false
  const seenEventIds = new Set<string>()

  const buildWebSocketUrl = options.buildWebSocketUrl || (() => '')
  const connectionKey = options.connectionKey || (() => '')
  const onTyping = options.onTyping || (() => {})
  const onRefresh = options.onRefresh || (() => Promise.resolve())
  const onConnected = options.onConnected || (() => Promise.resolve())
  const shouldSilenceDesktopNotification = options.shouldSilenceDesktopNotification || (() => false)

  const runReconciliation = async () => {
    try {
      await onConnected()
    } catch (error) {
      console.warn('Customer-service HTTP reconciliation failed:', error)
    }
  }

  const closeSocket = () => {
    const socket = realtimeSocket.value
    realtimeSocket.value = null
    activeSocketScope = ''
    if (socket && socket.readyState < WebSocket.CLOSING) {
      socket.close()
    }
  }

  const scheduleRefresh = (event: Record<string, any>) => {
    if (realtimeRefreshTimer) {
      window.clearTimeout(realtimeRefreshTimer)
    }

    realtimeRefreshTimer = window.setTimeout(async () => {
      realtimeRefreshTimer = null
      await onRefresh(event)
    }, 350)
  }

  const clearReconnectTimer = () => {
    if (realtimeReconnectTimer) {
      window.clearTimeout(realtimeReconnectTimer)
      realtimeReconnectTimer = null
    }
  }

  const rememberEvent = (eventId: unknown) => {
    if (typeof eventId !== 'string' || !eventId) return true
    if (seenEventIds.has(eventId)) return false
    if (seenEventIds.size >= 2048) {
      seenEventIds.clear()
    }
    seenEventIds.add(eventId)
    return true
  }

  const updateCursor = (cursor: unknown) => {
    if (typeof cursor === 'string' && cursor) {
      lastEventId = cursor
    }
  }

  const resetCursorForScope = () => {
    const nextScope = String(connectionKey() || '')
    if (nextScope === cursorScope) return

    cursorScope = nextScope
    lastEventId = ''
    seenEventIds.clear()
  }

  const socketCanStayOpen = (socket: WebSocket | null, scope: string) => {
    if (!socket || activeSocketScope !== scope) return false
    return socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING
  }

  const recoverBrowserRealtime = () => {
    if (!reconnectEnabled || !String(connectionKey() || '')) return
    void runReconciliation()
    connect()
  }

  const handleVisibilityChange = () => {
    if (document.visibilityState === 'visible') {
      recoverBrowserRealtime()
    }
  }

  const attachBrowserRecoveryListeners = () => {
    if (browserRecoveryListenersAttached || typeof window === 'undefined') return
    window.addEventListener('online', recoverBrowserRealtime)
    document.addEventListener('visibilitychange', handleVisibilityChange)
    browserRecoveryListenersAttached = true
  }

  const removeBrowserRecoveryListeners = () => {
    if (!browserRecoveryListenersAttached || typeof window === 'undefined') return
    window.removeEventListener('online', recoverBrowserRealtime)
    document.removeEventListener('visibilitychange', handleVisibilityChange)
    browserRecoveryListenersAttached = false
  }

  const handleFrame = (raw: unknown) => {
    if (typeof raw !== 'string') return

    try {
      const frame = JSON.parse(raw || '{}')
      updateCursor(frame?.cursor)
      if (frame?.type !== 'event' || !frame.event || typeof frame.event !== 'object') return

      const event = frame.event as Record<string, any>
      if (!rememberEvent(event.event_id)) return
      if (event.type === 'conversation.typing') {
        onTyping(event)
        return
      }
      void notifyCustomerMessage(event, { shouldSilenceDesktopNotification })
      scheduleRefresh(event)
    } catch (error) {
      console.warn('Invalid customer-service WebSocket frame:', error)
    }
  }

  const scheduleReconnect = () => {
    if (!reconnectEnabled || realtimeReconnectTimer || typeof window === 'undefined') return
    if (navigator.onLine === false) return

    const exponentialDelay = Math.min(30_000, 1_000 * (2 ** reconnectAttempt))
    const jitteredDelay = Math.round(exponentialDelay * (0.8 + Math.random() * 0.4))
    reconnectAttempt = Math.min(reconnectAttempt + 1, 5)
    realtimeReconnectTimer = window.setTimeout(() => {
      realtimeReconnectTimer = null
      connect()
    }, jitteredDelay)
  }

  const connect = () => {
    if (typeof window === 'undefined' || !('WebSocket' in window)) return

    resetCursorForScope()
    const scope = cursorScope
    const socketURL = String(buildWebSocketUrl(lastEventId) || '')
    if (!socketURL) return

    reconnectEnabled = true
    attachBrowserRecoveryListeners()
    if (socketCanStayOpen(realtimeSocket.value, scope)) return

    clearReconnectTimer()
    closeSocket()
    const socket = new WebSocket(socketURL)
    realtimeSocket.value = socket
    activeSocketScope = scope

    socket.onopen = () => {
      if (realtimeSocket.value !== socket) return
      reconnectAttempt = 0
      void runReconciliation()
    }
    socket.onmessage = (event) => {
      if (realtimeSocket.value !== socket) return
      handleFrame(event.data)
    }
    socket.onclose = () => {
      if (realtimeSocket.value === socket) {
        realtimeSocket.value = null
        activeSocketScope = ''
        scheduleReconnect()
      }
    }
    socket.onerror = () => {
      // Browsers follow an error with close. Reconnect from onclose so only one
      // timer is ever scheduled for a connection failure.
    }
  }

  const sendCustomerServiceRealtimeControl = (control: CustomerServiceRealtimeControl) => {
    const socket = realtimeSocket.value
    if (!socket || socket.readyState !== WebSocket.OPEN) return false

    try {
      socket.send(JSON.stringify(control))
      return true
    } catch (error) {
      console.warn('Failed to send customer-service WebSocket control:', error)
      return false
    }
  }

  const close = () => {
    reconnectEnabled = false
    closeSocket()
    clearReconnectTimer()
    if (realtimeRefreshTimer) {
      window.clearTimeout(realtimeRefreshTimer)
      realtimeRefreshTimer = null
    }
    removeBrowserRecoveryListeners()
  }

  return {
    realtimeSocket,
    connectCustomerServiceRealtime: connect,
    closeCustomerServiceRealtime: close,
    sendCustomerServiceRealtimeControl,
  }
}

export default useCustomerServiceRealtime
