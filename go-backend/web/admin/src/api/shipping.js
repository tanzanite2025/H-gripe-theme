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
  async quote(payload) {
    const response = await axios.post('/api/admin/shipping/quote', payload)
    return unwrapPayload(response)
  },

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

  async listTrackingProviders(params = {}) {
    const response = await axios.get('/api/admin/shipping/tracking-providers', { params })
    return unwrapList(response, 'tracking_providers')
  },

  async getTrackingProvider(id) {
    const response = await axios.get(`/api/admin/shipping/tracking-providers/${id}`)
    return unwrapPayload(response)
  },

  async createTrackingProvider(payload) {
    const response = await axios.post('/api/admin/shipping/tracking-providers', payload)
    return unwrapPayload(response)
  },

  async updateTrackingProvider(id, payload) {
    const response = await axios.put(`/api/admin/shipping/tracking-providers/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteTrackingProvider(id) {
    const response = await axios.delete(`/api/admin/shipping/tracking-providers/${id}`)
    return unwrapPayload(response)
  },

  async listTrackingCarrierMappings(params = {}) {
    const response = await axios.get('/api/admin/shipping/tracking-carrier-mappings', { params })
    return unwrapList(response, 'tracking_carrier_mappings')
  },

  async getTrackingCarrierMapping(id) {
    const response = await axios.get(`/api/admin/shipping/tracking-carrier-mappings/${id}`)
    return unwrapPayload(response)
  },

  async createTrackingCarrierMapping(payload) {
    const response = await axios.post('/api/admin/shipping/tracking-carrier-mappings', payload)
    return unwrapPayload(response)
  },

  async updateTrackingCarrierMapping(id, payload) {
    const response = await axios.put(`/api/admin/shipping/tracking-carrier-mappings/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteTrackingCarrierMapping(id) {
    const response = await axios.delete(`/api/admin/shipping/tracking-carrier-mappings/${id}`)
    return unwrapPayload(response)
  },

  async listTrackingShipments(params = {}) {
    const response = await axios.get('/api/admin/shipping/tracking-shipments', { params })
    return unwrapList(response, 'tracking_shipments')
  },

  async listTrackingEvents(orderId) {
    const response = await axios.get(`/api/admin/shipping/tracking-shipments/${orderId}/events`)
    return unwrapList(response, 'tracking_events')
  },

  async getTrackingPollingState() {
    const response = await axios.get('/api/admin/shipping/tracking-polling')
    return unwrapPayload(response)
  },

  async getTrackingWebhookState() {
    const response = await axios.get('/api/admin/shipping/tracking-webhook')
    return unwrapPayload(response)
  },

  async syncDueTrackingShipments(params = {}) {
    const response = await axios.post('/api/admin/shipping/tracking-shipments/sync-due', null, { params })
    return unwrapPayload(response)
  },

  async registerTrackingShipment(orderId) {
    const response = await axios.post(`/api/admin/shipping/tracking-shipments/${orderId}/register`)
    return unwrapPayload(response)
  },

  async syncTrackingShipment(orderId) {
    const response = await axios.post(`/api/admin/shipping/tracking-shipments/${orderId}/sync`)
    return unwrapPayload(response)
  },

  async listCarrierServices(params = {}) {
    const response = await axios.get('/api/admin/shipping/carrier-services', { params })
    return unwrapList(response, 'carrier_services')
  },

  async getCarrierService(id) {
    const response = await axios.get(`/api/admin/shipping/carrier-services/${id}`)
    return unwrapPayload(response)
  },

  async createCarrierService(payload) {
    const response = await axios.post('/api/admin/shipping/carrier-services', payload)
    return unwrapPayload(response)
  },

  async updateCarrierService(id, payload) {
    const response = await axios.put(`/api/admin/shipping/carrier-services/${id}`, payload)
    return unwrapPayload(response)
  },

  async deleteCarrierService(id) {
    const response = await axios.delete(`/api/admin/shipping/carrier-services/${id}`)
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

  async createPackagingRuleApply(payload) {
    const response = await axios.post('/api/admin/shipping/packaging-rules/apply', payload)
    return unwrapPayload(response)
  },

  async deletePackagingRuleApply(applyId) {
    const response = await axios.delete(`/api/admin/shipping/packaging-rules/apply/${applyId}`)
    return unwrapPayload(response)
  },
}

export default shippingApi
