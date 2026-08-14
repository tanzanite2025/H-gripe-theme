import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const readPayload = (response: unknown, endpoint: string) => (
  unwrapApiPayload(response, endpoint)
)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readOfferPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiObjectField(payload, 'offer', endpoint)
  return payload
}

export default {
  async getConnection() {
    const endpoint = '/api/admin/google-merchant/connection'
    const payload = readObjectPayload(await axios.get(endpoint), endpoint)
    requireApiObjectField(payload, 'connection', endpoint)
    return payload
  },
  async updateConnection(payload: Record<string, any>) {
    const endpoint = '/api/admin/google-merchant/connection'
    const responsePayload = readObjectPayload(await axios.patch(endpoint, payload), endpoint)
    requireApiObjectField(responsePayload, 'connection', endpoint)
    return responsePayload
  },
  async startOAuth() {
    const endpoint = '/api/admin/google-merchant/oauth/start'
    const payload = readObjectPayload(await axios.post(endpoint), endpoint)
    requireApiStringField(payload, 'authorization_url', endpoint)
    return payload
  },
  async disconnect() {
    const endpoint = '/api/admin/google-merchant/disconnect'
    return requireApiAcknowledgement(await axios.post(endpoint), endpoint)
  },
  async listRemoteProducts(params: Record<string, any> = {}) {
    const endpoint = '/api/admin/google-merchant/remote-products'
    const payload = readObjectPayload(await axios.get(endpoint, { params }), endpoint)
    requireApiArrayField(payload, 'products', endpoint)
    if (payload.next_page_token !== undefined) {
      requireApiStringField(payload, 'next_page_token', endpoint)
    }
    return payload
  },
  async listOffers() {
    const endpoint = '/api/admin/google-merchant/offers'
    const payload = readObjectPayload(await axios.get(endpoint), endpoint)
    requireApiArrayField(payload, 'offers', endpoint)
    return payload
  },
  async reconcile() {
    const endpoint = '/api/admin/google-merchant/reconcile'
    const payload = readObjectPayload(await axios.post(endpoint), endpoint)
    const result = requireApiObjectField(payload, 'result', endpoint)
    requireApiNumberField(result, 'considered', endpoint)
    requireApiNumberField(result, 'synced', endpoint)
    requireApiNumberField(result, 'withdrawn', endpoint)
    requireApiNumberField(result, 'skipped', endpoint)
    requireApiNumberField(result, 'failed', endpoint)
    if (result.errors !== undefined) {
      requireApiArrayField(result, 'errors', endpoint)
    }
    return payload
  },
  async createOffer(payload: Record<string, any>) {
    const endpoint = '/api/admin/google-merchant/offers'
    return readOfferPayload(await axios.post(endpoint, payload), endpoint)
  },
  async updateOffer(id: number, payload: Record<string, any>) {
    const endpoint = `/api/admin/google-merchant/offers/${id}`
    return readOfferPayload(await axios.put(endpoint, payload), endpoint)
  },
  async validateOffer(id: number) {
    const endpoint = `/api/admin/google-merchant/offers/${id}/validate`
    return readOfferPayload(await axios.post(endpoint), endpoint)
  },
  async syncOffer(id: number) {
    const endpoint = `/api/admin/google-merchant/offers/${id}/sync`
    return readOfferPayload(await axios.post(endpoint), endpoint)
  },
  async removeRemoteOffer(id: number) {
    const endpoint = `/api/admin/google-merchant/offers/${id}/remove-remote`
    return readOfferPayload(await axios.post(endpoint), endpoint)
  },
  async deleteOffer(id: number) {
    const endpoint = `/api/admin/google-merchant/offers/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  }
}
