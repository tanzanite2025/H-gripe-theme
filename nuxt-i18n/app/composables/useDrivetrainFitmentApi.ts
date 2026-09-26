import { useApiRequest } from '~/composables/useApiRequest'
import type {
  DrivetrainApiEnvelope,
  DrivetrainCalculateRequest,
  DrivetrainCalculationResponse,
  DrivetrainMatrixResponse,
} from '~/types/drivetrainFitment'

const drivetrainPath = '/fitment/drivetrain'

export function useDrivetrainFitmentApi() {
  const { request } = useApiRequest()

  const fetchMatrix = () => request<DrivetrainApiEnvelope<DrivetrainMatrixResponse>>(
    `${drivetrainPath}/matrix`,
    {},
    'Failed to load drivetrain fitment matrix',
  )

  const calculate = (payload: DrivetrainCalculateRequest) => request<DrivetrainApiEnvelope<DrivetrainCalculationResponse>>(
    `${drivetrainPath}/calculate`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    },
    'Failed to calculate drivetrain fitment',
  )

  return {
    fetchMatrix,
    calculate,
  }
}
