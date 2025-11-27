import { ref } from 'vue'
import type { SpokeCalcInput, SpokeCalcResult } from '~/types/spoke'
import type { SpokeCalcApiResponse } from '~/server/api/spoke-calc.post'

export const useSpokeCalculator = () => {
  const loading = ref(false)
  const error = ref<string | null>(null)
  const result = ref<SpokeCalcResult | null>(null)

  const calculate = async (input: SpokeCalcInput) => {
    loading.value = true
    error.value = null

    try {
      const { data, error: fetchError } = await useFetch<SpokeCalcApiResponse>('/api/spoke-calc', {
        method: 'POST',
        body: input,
      })

      if (fetchError.value) {
        throw fetchError.value
      }

      if (data.value) {
        result.value = {
          leftLengthMm: data.value.leftLengthMm,
          rightLengthMm: data.value.rightLengthMm,
        }
      } else {
        result.value = null
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
