import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

export const productApi = {
  async list(params: Record<string, any> = {}) {
    const response = await axios.get('/api/admin/products', { params })
    return unwrapPayload(response)
  },

  async stats() {
    const response = await axios.get('/api/admin/products/stats')
    return unwrapPayload(response)
  },

  async get(id: number | string) {
    const payload = unwrapPayload(await axios.get(`/api/admin/products/${id}`))
    return payload.product ?? payload
  },

  async create(payload: Record<string, any>) {
    return unwrapPayload(await axios.post('/api/admin/products', payload))
  },

  async update(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.put(`/api/admin/products/${id}`, payload))
  },

  async updateStatus(id: number | string, status: string) {
    return unwrapPayload(await axios.patch(`/api/admin/products/${id}/status`, { status }))
  },

  async deleteProduct(id: number | string) {
    return unwrapPayload(await axios.delete(`/api/admin/products/${id}`))
  },

  async batchUpdateStatus(productIds: Array<number | string>, status: string) {
    return unwrapPayload(await axios.post('/api/admin/products/batch-status', { product_ids: productIds, status }))
  },

  async batchDelete(productIds: Array<number | string>) {
    return unwrapPayload(await axios.post('/api/admin/products/batch-delete', { product_ids: productIds }))
  }
}

export const productInformationTemplateApi = {
  async list(params: Record<string, any> = {}) {
    const response = await axios.get('/api/admin/product-information-templates', { params })
    const payload = response.data?.data ?? response.data ?? []
    return Array.isArray(payload) ? payload : []
  },

  async get(id: number | string) {
    const response = await axios.get(`/api/admin/product-information-templates/${id}`)
    return response.data?.data ?? response.data ?? {}
  },

  async create(payload: Record<string, any>) {
    const response = await axios.post('/api/admin/product-information-templates', payload)
    return response.data?.data ?? response.data ?? {}
  },

  async update(id: number | string, payload: Record<string, any>) {
    const response = await axios.put(`/api/admin/product-information-templates/${id}`, payload)
    return response.data?.data ?? response.data ?? {}
  },

  async remove(id: number | string) {
    const response = await axios.delete(`/api/admin/product-information-templates/${id}`)
    return response.data?.data ?? response.data ?? {}
  }
}

export default productApi
