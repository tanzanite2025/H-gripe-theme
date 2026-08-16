import { ref } from 'vue'

type CustomerServiceRealtimeControl = {
  type: 'ping' | 'typing'
  is_typing?: boolean
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
