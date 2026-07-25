import axios from '@/utils/axios'

const unwrapPayload = (response: any) => response.data?.data ?? response.data ?? {}

const unwrapList = (response: any, key: string) => {
  const payload = unwrapPayload(response)

  if (Array.isArray(payload)) return payload
  if (Array.isArray(payload.data)) return payload.data
  if (key && Array.isArray(payload[key])) return payload[key]

  return []
}

export const productTypeApi = {
  async list(params: Record<string, any> = {}) {
    const response = await axios.get('/api/admin/product-types', { params })
    return unwrapList(response, 'product_types')
  },

  async create(payload: Record<string, any>) {
    return unwrapPayload(await axios.post('/api/admin/product-types', payload))
  },

  async update(id: number | string, payload: Record<string, any>) {
    return unwrapPayload(await axios.put(`/api/admin/product-types/${id}`, payload))
  },

  async deleteProductType(id: number | string) {
    return unwrapPayload(await axios.delete(`/api/admin/product-types/${id}`))
  }
}

export default productTypeApi
