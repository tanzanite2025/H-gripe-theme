import { ref, computed } from 'vue'

export interface FeedbackItem {
  id: number
  thread_key: string
  user_id: number
  name: string | null
  content: string
  status: string
  locale: string | null
  created_at: string
}

export interface FeedbackPaginationMeta {
  total: number
  total_pages: number
  page: number
  per_page: number
}

export interface FeedbackListResponse {
  data: FeedbackItem[]
  pagination: FeedbackPaginationMeta
}

export interface FeedbackEligibility {
  can_post: boolean
  logged_in: boolean
  reason?: string | null
}

export interface CreateFeedbackPayload {
  content: string
  name?: string
  email?: string
  locale?: string
}

export const useFeedback = (threadKey: string) => {
  const config = useRuntimeConfig()

  const apiBase = computed(() => {
    const base = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    return base.replace(/\/$/, '')
  })

  const items = ref<FeedbackItem[]>([])
  const pagination = ref<FeedbackPaginationMeta | null>(null)
  const loadingList = ref(false)
  const loadingSubmit = ref(false)
  const error = ref<string | null>(null)
  const search = ref('')
  const eligibility = ref<FeedbackEligibility | null>(null)

  const buildHeaders = (needsJson = false) => {
    const headers: Record<string, string> = { accept: 'application/json' }
    if (needsJson) {
      headers['Content-Type'] = 'application/json'
    }
    const wpNonce = (config.public as { wpNonce?: string }).wpNonce
    if (wpNonce) {
      headers['X-WP-Nonce'] = String(wpNonce)
    }
    return headers
  }

  const fetchList = async (page = 1) => {
    loadingList.value = true
    error.value = null

    try {
      const response = await $fetch<FeedbackListResponse>(
        `${apiBase.value}/tanzanite/v1/feedback`,
        {
          method: 'GET',
          credentials: 'include',
          headers: buildHeaders(false),
          query: {
            thread: threadKey,
            page,
            per_page: 20,
            status: 'approved',
          },
        }
      )

      items.value = response.data || []
      pagination.value = response.pagination
    } catch (err: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load feedback:', err)
      error.value = err?.message || 'Failed to load feedback'
    } finally {
      loadingList.value = false
    }
  }

  const submitFeedback = async (payload: CreateFeedbackPayload) => {
    loadingSubmit.value = true
    error.value = null

    try {
      const body = {
        thread: threadKey,
        content: payload.content,
        name: payload.name,
        email: payload.email,
        locale: payload.locale,
      }

      const response = await $fetch<{ id: number; status: string; message?: string }>(
        `${apiBase.value}/tanzanite/v1/feedback`,
        {
          method: 'POST',
          credentials: 'include',
          headers: buildHeaders(true),
          body,
        }
      )

      return {
        success: true as const,
        id: response.id,
        status: response.status,
        message: response.message || 'Feedback submitted, pending review.',
      }
    } catch (err: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to submit feedback:', err)
      error.value = err?.message || 'Failed to submit feedback'
      return {
        success: false as const,
        error: err,
      }
    } finally {
      loadingSubmit.value = false
    }
  }

  const loadEligibility = async () => {
    try {
      const response = await $fetch<FeedbackEligibility>(
        `${apiBase.value}/tanzanite/v1/feedback/eligibility`,
        {
          method: 'GET',
          credentials: 'include',
          headers: buildHeaders(false),
          query: {
            thread: threadKey,
          },
        }
      )

      eligibility.value = response
    } catch (err) {
      // 资格接口失败时，不阻塞整体，只在控制台记录
      // eslint-disable-next-line no-console
      console.warn('Failed to load feedback eligibility:', err)
      eligibility.value = {
        can_post: false,
        logged_in: false,
        reason: 'Unable to determine eligibility',
      }
    }
  }

  return {
    // state
    items,
    pagination,
    loadingList,
    loadingSubmit,
    error,
    search,
    eligibility,

    // actions
    fetchList,
    submitFeedback,
    loadEligibility,
  }
}
