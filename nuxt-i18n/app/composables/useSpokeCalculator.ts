import { ref } from 'vue'
import type { SpokeCalcInput, SpokeCalcResult } from '~/types/spoke'

// Define the response shape from Go backend
interface SpokeCalcApiResponse {
  leftLengthMm: number
  rightLengthMm: number
  debug: any
}
export const useSpokeCalculator = () => {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const result = ref<SpokeCalcResult | null>(null)

  const calculate = async (input: SpokeCalcInput) => {
    loading.value = true
    error.value = null

    try {
      // Use the actual Go backend endpoint instead of Nuxt mock
      const auth = useAuth()
      const data = await auth.request<SpokeCalcApiResponse>('/spoke/calc', {
        method: 'POST',
        body: JSON.stringify(input)
      })

      if (data && data.leftLengthMm && data.rightLengthMm) {
        result.value = {
          leftLengthMm: data.leftLengthMm,
          rightLengthMm: data.rightLengthMm,
        }
      } else {
        result.value = null
        throw new Error('Invalid response format from server')
      }
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Spoke calc failed', e)
      error.value = e?.message || 'Failed to calculate spoke lengths'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    result,
    calculate,
  }
}
