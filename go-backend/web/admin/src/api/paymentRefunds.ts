import axios from '@/utils/axios'
import {
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export const paymentRefundApi = {
  async executePendingRefund(id: number | string, payload: Record<string, any>) {
    const endpoint = `/api/admin/payment/refunds/${id}/execute`
    const result = requireApiObject(unwrapApiPayload(await axios.post(endpoint, payload), endpoint), endpoint)
    const refund = requireApiObjectField(result, 'refund', endpoint)
    const execution = requireApiObjectField(result, 'execution', endpoint)
    requireApiNumberField(refund, 'id', endpoint)
    requireApiStringField(refund, 'status', endpoint)
    requireApiStringField(refund, 'refund_id', endpoint)
    requireApiNumberField(execution, 'id', endpoint)
    requireApiNumberField(execution, 'refund_id', endpoint)
    requireApiStringField(execution, 'status', endpoint)
    requireApiStringField(execution, 'provider_refund_id', endpoint)
    return result
  },
}

export default paymentRefundApi
