import { ref, computed, unref } from 'vue'
import type { MaybeRef } from 'vue'

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

interface SuggestionPaginationMeta {
  total: number
  total_pages: number
  page: number
  per_page: number
}

interface SuggestionListResponse {
  data: any[]
  pagination: SuggestionPaginationMeta
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

export const useSuggestionFeedback = (threadKey: MaybeRef<string> = 'product_service_suggestion') => {
  const config = useRuntimeConfig()
  const normalizedThread = computed(() => {
    const value = unref(threadKey)
    return value || 'product_service_suggestion'
  })
  const base = computed(() => {
    const wpBase = (config.public as { wpApiBase?: string }).wpApiBase || '/wp-json'
    return wpBase.replace(/\/$/, '')
  })

  const isSubmitting = ref(false)
  const errorMessage = ref<string | null>(null)
  const successMessage = ref<string | null>(null)
  const eligibility = ref<SuggestionEligibility | null>(null)

  const buildHeaders = (needsJson = false) => {
    const headers: Record<string, string> = { accept: 'application/json' }
    if (needsJson) headers['Content-Type'] = 'application/json'
    const wpNonce = (config.public as { wpNonce?: string }).wpNonce
    if (wpNonce) headers['X-WP-Nonce'] = String(wpNonce)
    return headers
  }

  const loadEligibility = async () => {
    try {
      const response = await $fetch<SuggestionEligibility>(
        `${base.value}/tanzanite/v1/suggestion-feedback/eligibility`,
        {
          method: 'GET',
          credentials: 'include',
          headers: buildHeaders(false),
          query: {
            threadKey: normalizedThread.value,
          },
        }
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
      const response = await $fetch<SuggestionResponse>(
        `${base.value}/tanzanite/v1/suggestion-feedback`,
        {
          method: 'POST',
          credentials: 'include',
          headers: buildHeaders(true),
          body: {
            ...payload,
            threadKey: payload.threadKey || normalizedThread.value,
          },
        }
      )

      successMessage.value = response.message || '反馈已提交，客服会尽快审核。'
      return response
    } catch (error: any) {
      // eslint-disable-next-line no-console
      console.error('submitSuggestion failed', error)
      const message = error?.data?.message || error?.message || '提交失败，请稍后再试。'
      errorMessage.value = message
      throw new Error(message)
    } finally {
      isSubmitting.value = false
    }
  }

  return {
    eligibility,
    isSubmitting,
    errorMessage,
    successMessage,
    loadEligibility,
    submitSuggestion,
  }
}
