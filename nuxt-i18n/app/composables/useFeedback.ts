import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

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

const feedbackMessage = (err: unknown, fallback: string) => {
  if (err instanceof Error && err.message) return err.message
  return fallback
}

export const useFeedback = (threadKey: string) => {
  const auth = useAuth()

  const items = ref<FeedbackItem[]>([])
  const pagination = ref<FeedbackPaginationMeta | null>(null)
  const loadingList = ref(false)
  const loadingSubmit = ref(false)
  const error = ref<string | null>(null)
  const search = ref('')
  const eligibility = ref<FeedbackEligibility | null>(null)

  const fetchList = async (page = 1) => {
    loadingList.value = true
    error.value = null

    try {
      const params = new URLSearchParams({
        thread: threadKey,
        page: String(page),
        per_page: '20',
        status: 'approved',
      })

      const response = await auth.request<FeedbackListResponse>(
        `/feedback?${params.toString()}`,
        {
          headers: { accept: 'application/json' },
        },
        'Failed to load feedback'
      )

      items.value = response.data || []
      pagination.value = response.pagination
    } catch (err: unknown) {
      // eslint-disable-next-line no-console
      console.error('Failed to load feedback:', err)
      error.value = feedbackMessage(err, 'Failed to load feedback')
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

      const response = await auth.request<{ id: number; status: string; message?: string }>(
        '/feedback',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(body),
        },
        'Please log in to leave feedback.'
      )

      return {
        success: true as const,
        id: response.id,
        status: response.status,
        message: response.message || 'Feedback submitted, pending review.',
      }
    } catch (err: unknown) {
      // eslint-disable-next-line no-console
      console.error('Failed to submit feedback:', err)
      error.value = feedbackMessage(err, 'Failed to submit feedback')
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
      const params = new URLSearchParams({ thread: threadKey })
      const response = await auth.request<FeedbackEligibility>(
        `/feedback/eligibility?${params.toString()}`,
        {
          headers: { accept: 'application/json' },
        },
        'Unable to determine eligibility'
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
