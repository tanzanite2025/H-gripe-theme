import axios from '@/utils/axios'

const unwrapPayload = (response) => response.data?.data ?? response.data ?? {}

const unwrapList = (response, key) => {
  const payload = unwrapPayload(response)

  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (key && Array.isArray(payload[key])) return payload[key]

  return []
}

export const shippingApi = {
  async listTemplates() {
    const response = await axios.get('/api/admin/shipping/templates')
    return unwrapList(response, 'templates')
  },

  async getTemplate(id) {
    const response = await axios.get(`/api/admin/shipping/templates/${id}`)
    return unwrapPayload(response)
  },

  async createTemplate(payload) {
    const response = await axios.post('/api/admin/shipping/templates', payload)
    return unwrapPayload(response)
  },

  async updateTemplate(id, payload) {
    const response = await axios.put(`/api/admin/shipping/templates/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteTemplate(id) {
    const response = await axios.delete(`/api/admin/shipping/templates/${id}`)
    return unwrapPayload(response)
  },

  async listTemplateBindings() {
    const response = await axios.get('/api/admin/shipping/template-bindings')
    return unwrapList(response, 'template_bindings')
  },

  async getTemplateBinding(id) {
    const response = await axios.get(`/api/admin/shipping/template-bindings/${id}`)
    return unwrapPayload(response)
  },

  async createTemplateBinding(payload) {
    const response = await axios.post('/api/admin/shipping/template-bindings', payload)
    return unwrapPayload(response)
  },

  async updateTemplateBinding(id, payload) {
    const response = await axios.put(`/api/admin/shipping/template-bindings/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteTemplateBinding(id) {
    const response = await axios.delete(`/api/admin/shipping/template-bindings/${id}`)
    return unwrapPayload(response)
  },

  async listZones() {
    const response = await axios.get('/api/admin/shipping/zones')
    return unwrapList(response, 'zones')
  },

  async getZone(id) {
    const response = await axios.get(`/api/admin/shipping/zones/${id}`)
    return unwrapPayload(response)
  },

  async createZone(payload) {
    const response = await axios.post('/api/admin/shipping/zones', payload)
    return unwrapPayload(response)
  },

  async updateZone(id, payload) {
    const response = await axios.put(`/api/admin/shipping/zones/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteZone(id) {
    const response = await axios.delete(`/api/admin/shipping/zones/${id}`)
    return unwrapPayload(response)
  },

  async listCarriers(params = {}) {
    const response = await axios.get('/api/admin/shipping/carriers', { params })
    return unwrapList(response, 'carriers')
  },

  async getCarrier(id) {
    const response = await axios.get(`/api/admin/shipping/carriers/${id}`)
    return unwrapPayload(response)
  },

  async createCarrier(payload) {
    const response = await axios.post('/api/admin/shipping/carriers', payload)
    return unwrapPayload(response)
  },

  async updateCarrier(id, payload) {
    const response = await axios.put(`/api/admin/shipping/carriers/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteCarrier(id) {
    const response = await axios.delete(`/api/admin/shipping/carriers/${id}`)
    return unwrapPayload(response)
  },

  async listPackagingRules() {
    const response = await axios.get('/api/admin/shipping/packaging-rules')
    return unwrapList(response, 'packaging_rules')
  },

  async getPackagingRule(id) {
    const response = await axios.get(`/api/admin/shipping/packaging-rules/${id}`)
    return unwrapPayload(response)
  },

  async createPackagingRule(payload) {
    const response = await axios.post('/api/admin/shipping/packaging-rules', payload)
    return unwrapPayload(response)
  },

  async updatePackagingRule(id, payload) {
    const response = await axios.put(`/api/admin/shipping/packaging-rules/${id}`, payload)
    return unwrapPayload(response)
  },

  async deletePackagingRule(id) {
    const response = await axios.delete(`/api/admin/shipping/packaging-rules/${id}`)
    return unwrapPayload(response)
  },
}

export default shippingApi
