import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

const unwrapPaged = (response: any) => {
  const payload = response.data || {}
  return {
    data: payload.data || [],
    pagination: payload.pagination || { page: 1, page_size: 20, total: 0, total_pages: 0 },
  }
}

export const paymentRiskApi = {
  async getSummary(provider = '') {
    return unwrapPayload(await axios.get('/api/admin/payment/risk/summary', {
      params: provider ? { provider } : undefined,
    }))
  },

  async recomputeSummary(provider = '') {
    return unwrapPayload(await axios.post('/api/admin/payment/risk/recompute', {
      provider: provider || undefined,
    }))
  },

  async listProtectionControls(includeExpired = true) {
    return unwrapPayload(await axios.get('/api/admin/payment/risk/controls', {
      params: { include_expired: includeExpired ? 'true' : 'false' },
    }))
  },

  async createProtectionControl(payload: Record<string, any>) {
    return unwrapPayload(await axios.post('/api/admin/payment/risk/controls', payload))
  },

  async revokeProtectionControl(id: number | string) {
    return unwrapPayload(await axios.post(`/api/admin/payment/risk/controls/${id}/revoke`, { confirm: true }))
  },

  async listRefundRecommendations(params: Record<string, any> = {}) {
    return unwrapPaged(await axios.get('/api/admin/payment/risk/refund-recommendations', { params }))
  },

  async updateRefundRecommendation(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.patch(`/api/admin/payment/risk/refund-recommendations/${id}`, payload))
  },

  async createPendingRefundFromRecommendation(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.post(`/api/admin/payment/risk/refund-recommendations/${id}/pending-refund`, payload))
  },

  async listProtectionControlAudit(id: number | string, params: Record<string, any> = {}) {
    return unwrapPayload(await axios.get(`/api/admin/payment/risk/controls/${id}/audit`, { params }))
  },

  async listDisputes(params: Record<string, any> = {}) {
    return unwrapPaged(await axios.get('/api/admin/payment/disputes', { params }))
  },

  async getDispute(id: number | string) {
    return unwrapPayload(await axios.get(`/api/admin/payment/disputes/${id}`))
  },

  async getDisputeEvidence(id: number | string) {
    return unwrapPayload(await axios.get(`/api/admin/payment/disputes/${id}/evidence`))
  },

  async submitDisputeEvidence(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.post(`/api/admin/payment/disputes/${id}/evidence/submit`, payload))
  },

  async listReviews(params: Record<string, any> = {}) {
    return unwrapPaged(await axios.get('/api/admin/payment/reviews', { params }))
  },

  async getReview(id: number | string) {
    return unwrapPayload(await axios.get(`/api/admin/payment/reviews/${id}`))
  },

  async createReview(payload: Record<string, any>) {
    return unwrapPayload(await axios.post('/api/admin/payment/reviews', payload))
  },

  async updateReview(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.patch(`/api/admin/payment/reviews/${id}`, payload))
  },
}

export default paymentRiskApi
