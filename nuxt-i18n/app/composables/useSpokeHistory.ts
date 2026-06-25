import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

export interface SpokeHistoryItem {
  id: number
  wheel_type: string | null
  source_type: string | null
  rim_brand: string | null
  rim_model: string | null
  hub_brand: string | null
  hub_model: string | null
  erd_mm: number | null
  left_flange_pcd_mm: number | null
  right_flange_pcd_mm: number | null
  left_flange_to_center_mm: number | null
  right_flange_to_center_mm: number | null
  spoke_count: number | null
  lacing_pattern: string | null
  nipple_type: string | null
  left_length_mm: number | null
  right_length_mm: number | null
  created_at: string
  updated_at: string
}

export interface SpokeHistoryMeta {
  total: number
  total_pages: number
  page: number
  per_page: number
}

export interface SpokeHistoryResponse {
  items: SpokeHistoryItem[]
  meta: SpokeHistoryMeta
}

export const useSpokeHistory = () => {
  const auth = useAuth()

  const items = ref<SpokeHistoryItem[]>([])
  const meta = ref<SpokeHistoryMeta | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const searchText = ref('')

  const fetchHistory = async (options?: { search?: string; page?: number; perPage?: number; append?: boolean }) => {
    loading.value = true
    error.value = null

    const page = options?.page ?? 1
    const perPage = options?.perPage ?? 5
    const search = (options?.search ?? searchText.value).trim()
    const append = options?.append === true

    try {
      const query = new URLSearchParams({
        page: String(page),
        per_page: String(perPage),
      })
      if (search) {
        query.set('search', search)
      }

      const response = await auth.request<SpokeHistoryResponse>(
        `/spoke/history?${query.toString()}`,
        {
          headers: { accept: 'application/json' },
        },
        'Failed to load spoke length history',
      )

      if (!response || !Array.isArray(response.items)) {
        throw new Error('[CRITICAL] Spoke history API returned invalid data structure (missing items array)')
      }

      const nextItems = response.items
      const nextMeta = response.meta || null

      if (append && page > 1) {
        const existing = items.value || []
        const existingIds = new Set(existing.map((item) => item.id))
        const merged = [...existing, ...nextItems.filter((item) => !existingIds.has(item.id))]
        items.value = merged
      } else {
        items.value = nextItems
      }
      meta.value = nextMeta
    } catch (err: unknown) {
      // eslint-disable-next-line no-console
      console.error('Failed to load spoke history:', err)
      error.value = err instanceof Error ? err.message : 'Failed to load spoke length history'
      // 保持现有 items，不抛出错误，避免打断页面
    } finally {
      loading.value = false
    }
  }

  return {
    items,
    meta,
    loading,
    error,
    searchText,
    fetchHistory,
  }
}
