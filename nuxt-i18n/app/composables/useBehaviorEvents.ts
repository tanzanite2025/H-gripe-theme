import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute } from 'vue-router'
import { useApiRequest } from '~/composables/useApiRequest'
import type {
  BehaviorEventMetadata,
  BehaviorEventType,
  TrackBehaviorEventInput,
} from '~/types/behavior'

interface QueuedBehaviorEvent {
  event_id: string
  event_type: BehaviorEventType
  anonymous_id: string
  session_id: string
  product_id?: number
  category_id?: number
  locale: string
  path: string
  referrer?: string
  metadata?: BehaviorEventMetadata
  occurred_at: string
}

const ANONYMOUS_ID_KEY = 'tz_recommendation_anonymous_id'
const SESSION_ID_KEY = 'tz_recommendation_session_id'
const MAX_QUEUE_SIZE = 100
const BATCH_SIZE = 50
const FLUSH_DELAY_MS = 800

let flushTimer: ReturnType<typeof setTimeout> | null = null
let activeFlush: Promise<void> | null = null
let lifecycleRegistered = false
let attributionCaptured = false

const createClientId = (prefix: string) => {
  if (typeof globalThis.crypto?.randomUUID === 'function') {
    return `${prefix}_${globalThis.crypto.randomUUID()}`
  }
  return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 12)}`
}

const readStorageValue = (storage: Storage | undefined, key: string) => {
  if (!storage) return ''
  try {
    return storage.getItem(key) || ''
  } catch {
    return ''
  }
}

const writeStorageValue = (storage: Storage | undefined, key: string, value: string) => {
  if (!storage) return
  try {
    storage.setItem(key, value)
  } catch {
    // Storage can be unavailable in private browsing or restricted contexts.
  }
}

const normalizeMetadata = (metadata?: BehaviorEventMetadata) => {
  if (!metadata) return undefined

  return Object.entries(metadata).reduce<BehaviorEventMetadata>((result, [key, value]) => {
    const normalizedKey = key.trim().slice(0, 64)
    if (normalizedKey) {
      result[normalizedKey] = typeof value === 'string' ? value.slice(0, 512) : value
    }
    return result
  }, {})
}

export const useBehaviorEvents = () => {
  const { request } = useApiRequest()
  const route = useRoute()
  const { locale } = useI18n()
  const anonymousId = useState<string>('tz-recommendation-anonymous-id', () => '')
  const sessionId = useState<string>('tz-recommendation-session-id', () => '')
  const queue = useState<QueuedBehaviorEvent[]>('tz-behavior-event-queue', () => [])

  const ensureIdentity = () => {
    if (!import.meta.client) return

    if (!anonymousId.value) {
      anonymousId.value = readStorageValue(window.localStorage, ANONYMOUS_ID_KEY)
      if (!anonymousId.value) {
        anonymousId.value = createClientId('anon')
        writeStorageValue(window.localStorage, ANONYMOUS_ID_KEY, anonymousId.value)
      }
    }

    if (!sessionId.value) {
      sessionId.value = readStorageValue(window.sessionStorage, SESSION_ID_KEY)
      if (!sessionId.value) {
        sessionId.value = createClientId('session')
        writeStorageValue(window.sessionStorage, SESSION_ID_KEY, sessionId.value)
      }
    }
  }

  const flush = async () => {
    if (!import.meta.client || activeFlush || queue.value.length === 0) return

    ensureIdentity()
    const batch = queue.value.splice(0, BATCH_SIZE)
    if (!batch.length) return

    activeFlush = (async () => {
      try {
        await request(
          '/behavior-events/batch',
          {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              Accept: 'application/json',
            },
            body: JSON.stringify({ events: batch }),
            keepalive: true,
          },
          'Failed to submit behavior events'
        )
      } catch (error) {
        queue.value = [...batch, ...queue.value].slice(-MAX_QUEUE_SIZE)
        console.warn('[BehaviorEvents] Failed to submit event batch:', error)
      } finally {
        activeFlush = null
        if (queue.value.length >= BATCH_SIZE) {
          void flush()
        }
      }
    })()

    await activeFlush
  }

  const scheduleFlush = () => {
    if (!import.meta.client || flushTimer) return

    flushTimer = setTimeout(() => {
      flushTimer = null
      void flush()
    }, FLUSH_DELAY_MS)
  }

  const track = (input: TrackBehaviorEventInput) => {
    if (!import.meta.client) return ''

    ensureIdentity()
    if (!anonymousId.value && !sessionId.value) return ''

    const item: QueuedBehaviorEvent = {
      event_id: createClientId('event'),
      event_type: input.eventType,
      anonymous_id: anonymousId.value,
      session_id: sessionId.value,
      locale: String(input.locale || locale.value || 'en').slice(0, 20),
      path: String(input.path || route.fullPath || window.location.pathname).slice(0, 1024),
      occurred_at: input.occurredAt || new Date().toISOString(),
    }

    if (Number.isInteger(input.productId) && Number(input.productId) > 0) {
      item.product_id = Number(input.productId)
    }
    if (Number.isInteger(input.categoryId) && Number(input.categoryId) > 0) {
      item.category_id = Number(input.categoryId)
    }

    const referrer = input.referrer || document.referrer
    if (referrer) item.referrer = referrer.slice(0, 1024)

    const metadata = normalizeMetadata(input.metadata)
    if (metadata && Object.keys(metadata).length) {
      item.metadata = metadata
    }

    queue.value.push(item)
    if (queue.value.length > MAX_QUEUE_SIZE) {
      queue.value.splice(0, queue.value.length - MAX_QUEUE_SIZE)
    }

    if (queue.value.length >= BATCH_SIZE) {
      void flush()
    } else {
      scheduleFlush()
    }

    return item.event_id
  }

  const registerLifecycle = () => {
    if (!import.meta.client || lifecycleRegistered) return
    lifecycleRegistered = true

    window.addEventListener('pagehide', () => {
      void flush()
    })
    document.addEventListener('visibilitychange', () => {
      if (document.visibilityState === 'hidden') {
        void flush()
      }
    })
  }

  const captureAttribution = () => {
    if (!import.meta.client || attributionCaptured) return

    const query = new URLSearchParams(window.location.search)
    const metadata: BehaviorEventMetadata = {}
    for (const key of [
      'utm_source',
      'utm_medium',
      'utm_campaign',
      'utm_term',
      'utm_content',
      'gclid',
      'fbclid',
      'msclkid',
      'ttclid',
    ]) {
      const value = query.get(key)?.trim()
      if (value) metadata[key] = value.slice(0, 256)
    }
    if (!Object.keys(metadata).length) return

    attributionCaptured = true
    track({
      eventType: 'ad_landing',
      metadata,
    })
  }

  onMounted(() => {
    ensureIdentity()
    captureAttribution()
    registerLifecycle()
  })

  return {
    anonymousId,
    sessionId,
    queue,
    ensureIdentity,
    track,
    flush,
    captureAttribution,
  }
}
