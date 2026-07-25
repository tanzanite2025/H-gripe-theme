import { ref } from 'vue'

const REALTIME_EVENTS = [
  'conversation.message.created',
  'conversation.messages.read',
  'conversation.assigned',
  'conversation.status.changed',
  'conversation.context.updated',
  'conversation.typing'
]

export const useCustomerServiceRealtime = (options: Record<string, any>) => {
  const realtimeSource = ref<EventSource | null>(null)
  let realtimeReconnectTimer: number | null = null
  let realtimeRefreshTimer: number | null = null

  const buildEventUrl = options.buildEventUrl || (() => '')
  const onTyping = options.onTyping || (() => {})
  const onRefresh = options.onRefresh || (() => Promise.resolve())

  const closeSource = () => {
    if (realtimeSource.value) {
      realtimeSource.value.close()
      realtimeSource.value = null
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

  const handleEvent = (event: MessageEvent) => {
    try {
      const payload = JSON.parse(event.data || '{}')
      if (payload.type === 'conversation.typing') {
        onTyping(payload)
        return
      }
      scheduleRefresh(payload)
    } catch (error) {
      console.warn('Invalid customer-service realtime event:', error)
    }
  }

  const connect = () => {
    if (typeof window === 'undefined' || !('EventSource' in window)) return

    closeSource()
    const source = new EventSource(buildEventUrl(), { withCredentials: true })
    realtimeSource.value = source

    REALTIME_EVENTS.forEach((eventType) => {
      source.addEventListener(eventType, handleEvent)
    })

    source.onerror = () => {
      if (realtimeSource.value === source) {
        realtimeSource.value = null
      }
      source.close()
      if (!realtimeReconnectTimer) {
        realtimeReconnectTimer = window.setTimeout(() => {
          realtimeReconnectTimer = null
          connect()
        }, 5000)
      }
    }
  }

  const close = () => {
    closeSource()
    if (realtimeReconnectTimer) {
      window.clearTimeout(realtimeReconnectTimer)
      realtimeReconnectTimer = null
    }
    if (realtimeRefreshTimer) {
      window.clearTimeout(realtimeRefreshTimer)
      realtimeRefreshTimer = null
    }
  }

  return {
    realtimeSource,
    connectCustomerServiceRealtime: connect,
    closeCustomerServiceRealtime: close
  }
}

export default useCustomerServiceRealtime
