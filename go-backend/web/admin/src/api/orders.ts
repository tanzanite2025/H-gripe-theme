import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiObject,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type {
  OrderDisputeAnalysis,
  OrderDisputeCase,
  OrderDisputeEmailForm,
} from '@/modules/order/orderTypes'

const readObjectPayload = <T = Record<string, any>>(response: unknown, path: string): T => (
  requireApiObject(unwrapApiPayload(response, path), path) as T
)

const readPaged = <T = any>(response: unknown, path: string) => {
  const responseBody = requireApiObject((response as { data?: unknown }).data, path, 'response body')
  const payload = unwrapApiPayload(response, path)
  const data = requireApiArray<T>(payload, path, 'data')
  return {
    data,
    pagination: requireApiPagination(responseBody, payload, path),
  }
}

export const ordersApi = {
  async listDisputes(params: Record<string, any> = {}) {
    const path = '/api/admin/orders/disputes'
    return readPaged<OrderDisputeCase>(await axios.get(path, { params }), path)
  },

  async getDisputeAnalysis(orderID: number | string) {
    const path = `/api/admin/orders/${orderID}/dispute-analysis`
    return readObjectPayload<OrderDisputeAnalysis>(await axios.get(path), path)
  },

  async sendDisputeContactEmail(orderID: number | string, form: OrderDisputeEmailForm) {
    const path = `/api/admin/orders/${orderID}/dispute-contact-email`
    return readObjectPayload(await axios.post(path, {
      provider: form.provider,
      dispute_id: form.dispute_id,
      subject: form.subject,
      body: form.body,
      confirm: true,
    }), path)
  },
}

export default ordersApi

