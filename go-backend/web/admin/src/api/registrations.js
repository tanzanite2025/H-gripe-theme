import axios from '@/utils/axios'

const unwrapPayload = (response) => response.data?.data ?? response.data ?? {}

const unwrapPaged = (response) => ({
  data: Array.isArray(response.data?.data) ? response.data.data : [],
  pagination: response.data?.pagination ?? {
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0,
  },
})

const unwrapList = (response, key) => {
  const payload = unwrapPayload(response)
  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (key && Array.isArray(payload[key])) return payload[key]
  return []
}

export const registrationApi = {
  async getStats() {
    const response = await axios.get('/api/admin/registrations/stats')
    return unwrapPayload(response)
  },

  async listRegistrations(params = {}) {
    const response = await axios.get('/api/admin/registrations', { params })
    return unwrapPaged(response)
  },

  async updateRegistrationStatus(id, status) {
    const response = await axios.put(`/api/admin/registrations/${id}/status`, { status })
    return unwrapPayload(response)
  },

  async listExpiringWarranties(limit = 30) {
    const response = await axios.get('/api/admin/registrations/expiring', { params: { limit } })
    return unwrapList(response, 'data')
  },

  async listWarrantyClaims(params = {}) {
    const response = await axios.get('/api/admin/registrations/warranty-claims', { params })
    return unwrapPaged(response)
  },

  async getWarrantyClaim(id) {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}`)
    return unwrapPayload(response)
  },

  async updateWarrantyClaimStatus(id, status) {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/status`, { status })
    return unwrapPayload(response)
  },

  async updateWarrantyClaimResolution(id, resolution) {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/resolution`, { resolution })
    return unwrapPayload(response)
  },

  async listWarrantyClaimOrderItems(id) {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}/order-items`)
    return unwrapList(response, 'items')
  },

  async bindWarrantyClaimOrderItem(id, orderItemId) {
    const response = await axios.put(`/api/admin/registrations/warranty-claims/${id}/order-item`, {
      order_item_id: orderItemId || null
    })
    return unwrapPayload(response)
  },

  async listWarrantyServiceRecords(id) {
    const response = await axios.get(`/api/admin/registrations/warranty-claims/${id}/service-records`)
    return unwrapList(response, 'records')
  },

  async createWarrantyServiceRecord(id, payload) {
    const response = await axios.post(`/api/admin/registrations/warranty-claims/${id}/service-records`, payload)
    return unwrapPayload(response)
  },
}

export default registrationApi
