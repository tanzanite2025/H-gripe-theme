import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

type APIID = string | number
type APIParams = Record<string, any>
type APIPayload = Record<string, any>

const readPayload = (response: unknown, endpoint: string) => (
  unwrapApiPayload(response, endpoint)
)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readDataArray = (response: unknown, endpoint: string) => (
  requireApiArrayField(readObjectPayload(response, endpoint), 'data', endpoint)
)

const readEntity = (response: unknown, endpoint: string) => {
  const entity = readObjectPayload(response, endpoint)
  requireApiNumberField(entity, 'id', endpoint)
  return entity
}

const readNamedObject = (response: unknown, endpoint: string, field: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiObjectField(payload, field, endpoint)
  return payload
}

export const shippingApi = {
  async quote(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/quote'
    return readObjectPayload(await axios.post(endpoint, payload), endpoint)
  },

  async listTemplates() {
    const endpoint = '/api/admin/shipping/templates'
    return readDataArray(await axios.get(endpoint), endpoint)
  },

  async getTemplate(id: APIID) {
    const endpoint = `/api/admin/shipping/templates/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createTemplate(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/templates'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateTemplate(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/templates/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteTemplate(id: APIID) {
    const endpoint = `/api/admin/shipping/templates/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listZones() {
    const endpoint = '/api/admin/shipping/zones'
    return readDataArray(await axios.get(endpoint), endpoint)
  },

  async getZone(id: APIID) {
    const endpoint = `/api/admin/shipping/zones/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createZone(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/zones'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateZone(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/zones/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteZone(id: APIID) {
    const endpoint = `/api/admin/shipping/zones/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listCarriers(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/carriers'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async getCarrier(id: APIID) {
    const endpoint = `/api/admin/shipping/carriers/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createCarrier(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/carriers'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateCarrier(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/carriers/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteCarrier(id: APIID) {
    const endpoint = `/api/admin/shipping/carriers/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listTrackingProviders(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/tracking-providers'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async getTrackingProvider(id: APIID) {
    const endpoint = `/api/admin/shipping/tracking-providers/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createTrackingProvider(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/tracking-providers'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateTrackingProvider(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/tracking-providers/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteTrackingProvider(id: APIID) {
    const endpoint = `/api/admin/shipping/tracking-providers/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listTrackingCarrierMappings(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/tracking-carrier-mappings'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async getTrackingCarrierMapping(id: APIID) {
    const endpoint = `/api/admin/shipping/tracking-carrier-mappings/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createTrackingCarrierMapping(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/tracking-carrier-mappings'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateTrackingCarrierMapping(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/tracking-carrier-mappings/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteTrackingCarrierMapping(id: APIID) {
    const endpoint = `/api/admin/shipping/tracking-carrier-mappings/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listTrackingShipments(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/tracking-shipments'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async listTrackingEvents(orderId: APIID) {
    const endpoint = `/api/admin/shipping/tracking-shipments/${orderId}/events`
    return readDataArray(await axios.get(endpoint), endpoint)
  },

  async getTrackingPollingState() {
    const endpoint = '/api/admin/shipping/tracking-polling'
    return readObjectPayload(await axios.get(endpoint), endpoint)
  },

  async getTrackingWebhookState() {
    const endpoint = '/api/admin/shipping/tracking-webhook'
    return readObjectPayload(await axios.get(endpoint), endpoint)
  },

  async syncDueTrackingShipments(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/tracking-shipments/sync-due'
    return readObjectPayload(await axios.post(endpoint, null, { params }), endpoint)
  },

  async registerTrackingShipment(orderId: APIID) {
    const endpoint = `/api/admin/shipping/tracking-shipments/${orderId}/register`
    return readNamedObject(await axios.post(endpoint), endpoint, 'shipment')
  },

  async syncTrackingShipment(orderId: APIID) {
    const endpoint = `/api/admin/shipping/tracking-shipments/${orderId}/sync`
    return readNamedObject(await axios.post(endpoint), endpoint, 'tracking')
  },

  async listCarrierServices(params: APIParams = {}) {
    const endpoint = '/api/admin/shipping/carrier-services'
    return readDataArray(await axios.get(endpoint, { params }), endpoint)
  },

  async getCarrierService(id: APIID) {
    const endpoint = `/api/admin/shipping/carrier-services/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createCarrierService(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/carrier-services'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updateCarrierService(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/carrier-services/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deleteCarrierService(id: APIID) {
    const endpoint = `/api/admin/shipping/carrier-services/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async listPackagingRules() {
    const endpoint = '/api/admin/shipping/packaging-rules'
    return readDataArray(await axios.get(endpoint), endpoint)
  },

  async getPackagingRule(id: APIID) {
    const endpoint = `/api/admin/shipping/packaging-rules/${id}`
    return readEntity(await axios.get(endpoint), endpoint)
  },

  async createPackagingRule(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/packaging-rules'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async updatePackagingRule(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/shipping/packaging-rules/${id}`
    return readEntity(await axios.put(endpoint, payload), endpoint)
  },

  async deletePackagingRule(id: APIID) {
    const endpoint = `/api/admin/shipping/packaging-rules/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async createPackagingRuleApply(payload: APIPayload) {
    const endpoint = '/api/admin/shipping/packaging-rules/apply'
    return readEntity(await axios.post(endpoint, payload), endpoint)
  },

  async deletePackagingRuleApply(applyId: APIID) {
    const endpoint = `/api/admin/shipping/packaging-rules/apply/${applyId}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default shippingApi
