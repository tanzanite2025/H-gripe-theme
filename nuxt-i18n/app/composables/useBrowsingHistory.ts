import { computed, ref } from 'vue'

interface BrowsingHistoryItem {
  id: number
  title: string
  thumbnail: string
  price: string
  url: string
  viewedAt: string
}

interface BrowsingHistoryBackend {
  id: number
  user_id: number
  product_id: number
  view_count: number
  last_viewed_at: string
  created_at: string
  updated_at: string
}

interface BrowsingHistoryResponse {
  history?: BrowsingHistoryBackend[]
  count?: number
}

const STORAGE_KEY = 'tz_browsing_history'
const MAX_ITEMS = 20
const BACKEND_PATH = '/user/browsing-history'

const unwrapResponseData = <T>(response: T | { data?: T } | null | undefined): T | null => {
  if (!response || typeof response !== 'object') {
    return (response as T) || null
  }
  if ('data' in response && response.data !== undefined) {
    return response.data as T
  }
  return response as T
}

export const useBrowsingHistory = () => {
  const history = ref<BrowsingHistoryItem[]>([])
  const auth = useAuth()
  const { isAuthenticated } = auth

  let syncTimeout: ReturnType<typeof setTimeout> | null = null
  const syncQueue = new Set<number>()

  const loadHistory = () => {
    if (!import.meta.client) return

    try {
      const stored = localStorage.getItem(STORAGE_KEY)
      if (stored) {
        const parsed = JSON.parse(stored)
        history.value = Array.isArray(parsed) ? parsed : []
      }
    } catch (error) {
      console.error('[BrowsingHistory] Failed to load local history:', error)
      history.value = []
    }
  }

  const saveHistory = () => {
    if (!import.meta.client) return

    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(history.value))
    } catch (error) {
      console.error('[BrowsingHistory] Failed to save local history:', error)
    }
  }

  const loadFromBackend = async () => {
    if (!isAuthenticated.value) return

    try {
      const response = await auth.request<BrowsingHistoryResponse | { data?: BrowsingHistoryResponse }>(
        `${BACKEND_PATH}?limit=${MAX_ITEMS}`,
        {
          headers: { accept: 'application/json' },
        },
        'Failed to load browsing history'
      )
      const data = unwrapResponseData<BrowsingHistoryResponse>(response)

      if (!data?.history || !Array.isArray(data.history)) return
    } catch (error) {
      console.error('[BrowsingHistory] Failed to load backend history:', error)
    }
  }

  const syncToBackend = async (productID: number) => {
    if (!isAuthenticated.value) return

    syncQueue.add(productID)

    if (syncTimeout) {
      clearTimeout(syncTimeout)
    }

    syncTimeout = setTimeout(async () => {
      const idsToSync = Array.from(syncQueue)
      syncQueue.clear()

      for (const id of idsToSync) {
        try {
          await auth.request(
            BACKEND_PATH,
            {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                accept: 'application/json',
              },
              body: JSON.stringify({ product_id: id }),
            },
            'Failed to sync browsing history'
          )
        } catch (error) {
          console.error(`[BrowsingHistory] Failed to sync product ${id}:`, error)
        }
      }
    }, 500)
  }

  const addToHistory = (item: Omit<BrowsingHistoryItem, 'viewedAt'>) => {
    try {
      history.value = history.value.filter((entry) => entry.id !== item.id)
      history.value.unshift({
        ...item,
        viewedAt: new Date().toISOString(),
      })

      if (history.value.length > MAX_ITEMS) {
        history.value = history.value.slice(0, MAX_ITEMS)
      }

      saveHistory()

      if (isAuthenticated.value) {
        syncToBackend(item.id)
      }
    } catch (error) {
      console.error('[BrowsingHistory] Failed to add item:', error)
    }
  }

  const clearHistory = async () => {
    try {
      history.value = []
      if (import.meta.client) {
        localStorage.removeItem(STORAGE_KEY)
      }

      if (isAuthenticated.value) {
        await auth.request(
          BACKEND_PATH,
          {
            method: 'DELETE',
            headers: { accept: 'application/json' },
          },
          'Failed to clear browsing history'
        )
      }
    } catch (error) {
      console.error('[BrowsingHistory] Failed to clear history:', error)
    }
  }

  const removeItem = async (id: number) => {
    try {
      history.value = history.value.filter((entry) => entry.id !== id)
      saveHistory()

      if (isAuthenticated.value) {
        await auth.request(
          `${BACKEND_PATH}/${id}`,
          {
            method: 'DELETE',
            headers: { accept: 'application/json' },
          },
          'Failed to remove browsing history item'
        )
      }
    } catch (error) {
      console.error('[BrowsingHistory] Failed to remove item:', error)
    }
  }

  const historyCount = computed(() => history.value.length)
  const hasHistory = computed(() => history.value.length > 0)

  if (import.meta.client) {
    loadHistory()
    if (isAuthenticated.value) {
      loadFromBackend()
    }
  }

  return {
    history,
    historyCount,
    hasHistory,
    addToHistory,
    clearHistory,
    removeItem,
    loadHistory,
    loadFromBackend,
  }
}
