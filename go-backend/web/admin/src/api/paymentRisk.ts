import axios from '@/utils/axios'
import {
  requireApiArray,
  requireApiArrayField,
  requireApiBooleanField,
  requireApiField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const readObjectPayload = (response: unknown, path: string) => (
  requireApiObject(unwrapApiPayload(response, path), path)
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

export const paymentRiskApi = {
  async getSummary(provider = '') {
    const path = '/api/admin/payment/risk/summary'
    const payload = readObjectPayload(await axios.get(path, {
      params: provider ? { provider } : undefined,
    }), path)
    requireApiBooleanField(payload, 'enabled', path)
    requireApiObjectField(payload, 'reports', path)
    return payload
  },

  async recomputeSummary(provider = '') {
    const path = '/api/admin/payment/risk/recompute'
    const payload = readObjectPayload(await axios.post(path, {
      provider: provider || undefined,
    }), path)
    requireApiBooleanField(payload, 'enabled', path)
    requireApiObjectField(payload, 'reports', path)
    return payload
  },

  async listProtectionControls(includeExpired = true) {
    const path = '/api/admin/payment/risk/controls'
    const payload = readObjectPayload(await axios.get(path, {
      params: { include_expired: includeExpired ? 'true' : 'false' },
    }), path)
    requireApiBooleanField(payload, 'enabled', path)
    requireApiArrayField(payload, 'controls', path)
    requireApiObjectField(payload, 'policy', path)
    return payload
  },

  async createProtectionControl(payload: Record<string, any>) {
    const path = '/api/admin/payment/risk/controls'
    return readObjectPayload(await axios.post(path, payload), path)
  },

  async revokeProtectionControl(id: number | string) {
    const path = `/api/admin/payment/risk/controls/${id}/revoke`
    return readObjectPayload(await axios.post(path, { confirm: true }), path)
  },

  async listRefundRecommendations(params: Record<string, any> = {}) {
    const path = '/api/admin/payment/risk/refund-recommendations'
    return readPaged(await axios.get(path, { params }), path)
  },

  async updateRefundRecommendation(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/payment/risk/refund-recommendations/${id}`
    return readObjectPayload(await axios.patch(path, payload), path)
  },

  async createPendingRefundFromRecommendation(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/payment/risk/refund-recommendations/${id}/pending-refund`
    const result = readObjectPayload(await axios.post(path, payload), path)
    requireApiObjectField(result, 'recommendation', path)
    requireApiObjectField(result, 'refund', path)
    return result
  },

  async listProtectionControlAudit(id: number | string, params: Record<string, any> = {}) {
    const path = `/api/admin/payment/risk/controls/${id}/audit`
    const payload = readObjectPayload(await axios.get(path, { params }), path)
    requireApiArrayField(payload, 'logs', path)
    requireApiObjectField(payload, 'pagination', path)
    const pagination = requireApiObjectField(payload, 'pagination', path)
    requireApiNumberField(pagination, 'page', path)
    requireApiNumberField(pagination, 'page_size', path)
    requireApiNumberField(pagination, 'total', path)
    requireApiNumberField(pagination, 'total_pages', path)
    return payload
  },

  async listDisputes(params: Record<string, any> = {}) {
    const path = '/api/admin/payment/disputes'
    return readPaged(await axios.get(path, { params }), path)
  },

  async getDispute(id: number | string) {
    const path = `/api/admin/payment/disputes/${id}`
    return readObjectPayload(await axios.get(path), path)
  },

  async getDisputeEvidence(id: number | string) {
    const path = `/api/admin/payment/disputes/${id}/evidence`
    return readObjectPayload(await axios.get(path), path)
  },

  async submitDisputeEvidence(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/payment/disputes/${id}/evidence/submit`
    return readObjectPayload(await axios.post(path, payload), path)
  },

  async listPayPalDisputes(params: Record<string, any> = {}) {
    const path = '/api/admin/payment/paypal-disputes'
    return readPaged(await axios.get(path, { params }), path)
  },

  async getPayPalDisputeEvidence(id: number | string) {
    const path = `/api/admin/payment/paypal-disputes/${id}/evidence`
    return readObjectPayload(await axios.get(path), path)
  },

  async submitPayPalDisputeEvidence(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/payment/paypal-disputes/${id}/evidence/submit`
    return readObjectPayload(await axios.post(path, payload), path)
  },

  paypalDisputeInvoicePDFUrl(id: number | string) {
    return `/api/admin/payment/paypal-disputes/${id}/evidence/invoice.pdf`
  },

  async previewPayPalInvoicePDF(payload: Record<string, any>) {
    return axios.post('/api/admin/payment/paypal-invoice-preview.pdf', payload, {
      responseType: 'blob',
    })
  },

  async getPayPalInvoiceSellerProfile() {
    const path = '/api/admin/settings/paypal-invoice-seller-profile'
    return readObjectPayload(await axios.get(path), path)
  },

  async updatePayPalInvoiceSellerProfile(payload: Record<string, any>) {
    const path = '/api/admin/settings/paypal-invoice-seller-profile'
    return readObjectPayload(await axios.put(path, payload), path)
  },

  async listReviews(params: Record<string, any> = {}) {
    const path = '/api/admin/payment/reviews'
    return readPaged(await axios.get(path, { params }), path)
  },

  async getReview(id: number | string) {
    const path = `/api/admin/payment/reviews/${id}`
    return readObjectPayload(await axios.get(path), path)
  },

  async createReview(payload: Record<string, any>) {
    const path = '/api/admin/payment/reviews'
    return readObjectPayload(await axios.post(path, payload), path)
  },

  async updateReview(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/payment/reviews/${id}`
    return readObjectPayload(await axios.patch(path, payload), path)
  },
}

export default paymentRiskApi
