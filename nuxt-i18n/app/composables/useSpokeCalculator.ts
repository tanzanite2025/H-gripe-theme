import type { SpokeCalcResult, SpokeTensionRatio } from '~~/types/spoke'
import { useApiRequest } from '~/composables/useApiRequest'
import type { SpokeWheelBuildConfig, SpokeWheelSide } from '~/types/spokeCalculator'
import { hasSpokeCalculationGeometry, toSpokeCalcInput } from '~/utils/spokeCalculatorPayload'

export interface SpokeWheelCalculationResult {
  leftLengthMm: number | null
  rightLengthMm: number | null
  tensionRatio: SpokeTensionRatio | null
}

/**
 * Single API boundary for spoke calculations.
 *
 * The calculator panel owns editing state and result presentation. This hook
 * owns only request execution and response normalization, so future wizard
 * steps can reuse the same calculation path.
 */
export const useSpokeCalculator = () => {
  const { request } = useApiRequest()

  const calculateWheel = async (
    config: SpokeWheelBuildConfig,
    wheel: SpokeWheelSide,
    fallbackMessage: string,
  ): Promise<SpokeWheelCalculationResult | null> => {
    if (!hasSpokeCalculationGeometry(config)) {
      return null
    }

    const payload = await request<SpokeCalcResult>(
      '/spoke/calc',
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(toSpokeCalcInput(config, wheel)),
      },
      fallbackMessage,
    )

    if (!payload || !Number.isFinite(payload.leftLengthMm) || !Number.isFinite(payload.rightLengthMm)) {
      throw new Error('Invalid response format from server')
    }

    return {
      leftLengthMm: payload.leftLengthMm,
      rightLengthMm: payload.rightLengthMm,
      tensionRatio: payload.tensionRatio ?? null,
    }
  }

  return {
    calculateWheel,
  }
}
