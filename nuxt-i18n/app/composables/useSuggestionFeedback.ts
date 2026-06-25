import { ref, computed, unref } from 'vue'
import type { MaybeRef } from 'vue'
import { useAuth } from '~/composables/useAuth'

interface SuggestionEligibility {
  loggedIn: boolean
  canAttach: boolean
  requiredLevel: string
  userLevel: string
  reason?: string | null
}

interface SuggestionResponse {
  id: number
  status: string
  message?: string
}

export interface SuggestionPayload {
  fullName: string
  email: string
  country: string
  orderNumber: string
  productCategory: string
  requestType: string
  message: string
  attachments: Array<{ name: string; url: string; size: number }>
  threadKey?: string
}

const suggestionMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message) return error.message
  return fallback
}

export const useSuggestionFeedback = (threadKey: MaybeRef<string> = 'product_service_suggestion') => {
  const auth = useAuth()
  const normalizedThread = computed(() => {
    const value = unref(threadKey)
    return value || 'product_service_suggestion'
  })

  const isSubmitting = ref(false)
  const errorMessage = ref<string | null>(null)
  const successMessage = ref<string | null>(null)
  const eligibility = ref<SuggestionEligibility | null>(null)

  const loadEligibility = async () => {
    try {
      const params = new URLSearchParams({ threadKey: normalizedThread.value })
      const response = await auth.request<SuggestionEligibility>(
        `/suggestion-feedback/eligibility?${params.toString()}`,
        {
          headers: { accept: 'application/json' },
        },
        'Unable to determine eligibility'
      )
      eligibility.value = response
    } catch (error) {
      // eslint-disable-next-line no-console
      console.warn('Failed to load suggestion eligibility', error)
      eligibility.value = {
        loggedIn: false,
        canAttach: false,
        requiredLevel: 'silver',
        userLevel: '',
        reason: 'Unable to determine eligibility',
      }
    }
  }

  const submitSuggestion = async (payload: SuggestionPayload) => {
    isSubmitting.value = true
    errorMessage.value = null
    successMessage.value = null

    try {
      const response = await auth.request<SuggestionResponse>(
        '/suggestion-feedback',
        {
          method: 'POST',
          headers: {
            accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            ...payload,
            threadKey: payload.threadKey || normalizedThread.value,
          }),
        },
        'Please log in before submitting feedback.'
      )

      successMessage.value = response.message || '反馈已提交，客服会尽快审核。'
      return response
    } catch (error: unknown) {
      // eslint-disable-next-line no-console
      console.error('submitSuggestion failed', error)
      const message = suggestionMessage(error, '提交失败，请稍后再试。')
      errorMessage.value = message
      throw new Error(message)
    } finally {
      isSubmitting.value = false
    }
  }

  const uploadAttachment = async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)

    const response = await auth.request<{ url: string; name: string; size: number }>(
      '/suggestion-feedback/upload',
      {
        method: 'POST',
        body: formData,
        // FormData should not have Content-Type manually set so the browser can add the boundary
        headers: { accept: 'application/json' },
      },
      'Failed to upload attachment'
    )
    return response
  }

  return {
    eligibility,
    isSubmitting,
    errorMessage,
    successMessage,
    loadEligibility,
    uploadAttachment,
    submitSuggestion,
  }
}
