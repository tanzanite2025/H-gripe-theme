import { computed } from 'vue'
import { useAsyncData } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import { createServerRenderedInnerTubeValveFitmentMatrixFallback } from '~/data/tireguides/innerTubeValveFitmentServerRenderedFallback'

import type {
  InnerTubeValveFitmentMatrixMetadata,
  InnerTubeValveFitmentResult,
  InnerTubeValveFitmentSolveInput,
} from '~/types/innerTubeValveFitmentCalculationAndMatrixTypes'

export type {
  InnerTubeValveFitmentMode,
  InnerTubeValveFitmentStatus,
  InnerTubeValveStructure,
  InnerTubeRecommendationReasonKey,
  InnerTubeRecommendationKey,
  InnerTubeValveFitmentModelSource,
  InnerTubeValveFitmentRecommendation,
  InnerTubeValveFitmentResult,
  InnerTubeValveFitmentUncertaintyBoundary,
  InnerTubeValveFitmentUncertaintyReview,
  InnerTubeValveFitmentClearanceCell,
  InnerTubeValveFitmentAlternative,
  InnerTubeValveFitmentMatrixRow,
  InnerTubeValveFitmentMatrixMetadata,
  InnerTubeValveFitmentSolveInput,
} from '~/types/innerTubeValveFitmentCalculationAndMatrixTypes'

interface InnerTubeValveFitmentAPIEnvelope<T> {
  data?: T
}

const innerTubeValveFitmentMatrixEndpoint = '/engineering/inner-tube-fitment/matrix'
const innerTubeValveFitmentSolveEndpoint = '/engineering/inner-tube-fitment/solve'
const innerTubeValveFitmentMatrixServerTimeoutMs = 2500

export const useInnerTubeValveFitmentCalculator = async () => {
  const { request } = useApiRequest()
  const {
    data: matrixEnvelope,
    pending: matrixPending,
    error: matrixRequestError,
    refresh: refreshMatrix,
  } = await useAsyncData<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentMatrixMetadata>>(
    'inner-tube-valve-fitment-engineering-matrix-v1',
    async () => {
      const abortController = import.meta.server ? new AbortController() : null
      const timeoutHandle = abortController
        ? setTimeout(() => abortController.abort(), innerTubeValveFitmentMatrixServerTimeoutMs)
        : null

      try {
        return await request<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentMatrixMetadata>>(
          innerTubeValveFitmentMatrixEndpoint,
          { ...(abortController ? { signal: abortController.signal } : {}) },
          'Inner-tube valve fitment matrix is temporarily unavailable',
        )
      } finally {
        if (timeoutHandle) clearTimeout(timeoutHandle)
      }
    },
    { default: () => ({}) },
  )

  const metadata = computed(() => (
    matrixEnvelope.value?.data ?? createServerRenderedInnerTubeValveFitmentMatrixFallback()
  ))
  const defaultResult = computed(() => {
    return matrixEnvelope.value?.data?.rows
      .find(row => row.rim_depth_mm === 50)
      ?.recommended_result
      ?? null
  })

  const solveInnerTubeValveFitmentWithBackend = async (input: InnerTubeValveFitmentSolveInput) => {
    const response = await request<InnerTubeValveFitmentAPIEnvelope<InnerTubeValveFitmentResult>>(
      innerTubeValveFitmentSolveEndpoint,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(input),
      },
      'Inner-tube valve fitment calculation is temporarily unavailable',
    )
    if (!response.data) {
      throw new Error('Inner-tube valve fitment response is missing data')
    }
    return response.data
  }

  return {
    metadata,
    defaultResult,
    matrixPending,
    matrixError: matrixRequestError,
    refreshMatrix,
    solveInnerTubeValveFitmentWithBackend,
  }
}
