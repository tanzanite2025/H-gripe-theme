import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const paymentRefundApi = {
  async executePendingRefund(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.post(`/api/admin/payment/refunds/${id}/execute`, payload))
  },
}

export default paymentRefundApi
