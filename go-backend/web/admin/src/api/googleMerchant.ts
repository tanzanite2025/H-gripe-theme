import axios from '@/utils/axios'

const unwrap = (response: any) => response.data?.data ?? response.data ?? {}

export default {
  async getConnection() {
    return unwrap(await axios.get('/api/admin/google-merchant/connection'))
  },
  async updateConnection(payload: Record<string, any>) {
    return unwrap(await axios.patch('/api/admin/google-merchant/connection', payload))
  },
  async startOAuth() {
    return unwrap(await axios.post('/api/admin/google-merchant/oauth/start'))
  },
  async disconnect() {
    return unwrap(await axios.post('/api/admin/google-merchant/disconnect'))
  },
  async listRemoteProducts(params: Record<string, any> = {}) {
    return unwrap(await axios.get('/api/admin/google-merchant/remote-products', { params }))
  },
  async listOffers() {
    return unwrap(await axios.get('/api/admin/google-merchant/offers'))
  },
  async reconcile() {
    return unwrap(await axios.post('/api/admin/google-merchant/reconcile'))
  },
  async createOffer(payload: Record<string, any>) {
    return unwrap(await axios.post('/api/admin/google-merchant/offers', payload))
  },
  async updateOffer(id: number, payload: Record<string, any>) {
    return unwrap(await axios.put(`/api/admin/google-merchant/offers/${id}`, payload))
  },
  async validateOffer(id: number) {
    return unwrap(await axios.post(`/api/admin/google-merchant/offers/${id}/validate`))
  },
  async syncOffer(id: number) {
    return unwrap(await axios.post(`/api/admin/google-merchant/offers/${id}/sync`))
  },
  async removeRemoteOffer(id: number) {
    return unwrap(await axios.post(`/api/admin/google-merchant/offers/${id}/remove-remote`))
  },
  async deleteOffer(id: number) {
    return unwrap(await axios.delete(`/api/admin/google-merchant/offers/${id}`))
  }
}
