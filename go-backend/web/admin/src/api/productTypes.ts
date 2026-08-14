import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArray,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const readProductType = (response: unknown, endpoint: string): any => {
  const productType = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
  requireApiNumberField(productType, 'id', endpoint)
  requireApiStringField(productType, 'name', endpoint)
  requireApiStringField(productType, 'slug', endpoint)
  requireApiBooleanField(productType, 'is_enabled', endpoint)
  return productType
}

export const productTypeApi = {
  async list(params: Record<string, any> = {}) {
    const endpoint = '/api/admin/product-types'
    return requireApiArray(
      unwrapApiPayload(await axios.get(endpoint, { params }), endpoint),
      endpoint,
      'data',
    )
  },

  async create(payload: Record<string, any>) {
    const endpoint = '/api/admin/product-types'
    return readProductType(await axios.post(endpoint, payload), endpoint)
  },

  async update(id: number | string, payload: Record<string, any>) {
    const endpoint = `/api/admin/product-types/${id}`
    return readProductType(await axios.put(endpoint, payload), endpoint)
  },

  async uploadImage(id: number | string, file: File) {
    const formData = new FormData()
    formData.append('file', file)
    const endpoint = `/api/admin/product-types/${id}/image`
    return readProductType(await axios.post(endpoint, formData), endpoint)
  },

  async deleteImage(id: number | string) {
    const endpoint = `/api/admin/product-types/${id}/image`
    return readProductType(await axios.delete(endpoint), endpoint)
  },

  async deleteProductType(id: number | string) {
    const endpoint = `/api/admin/product-types/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  }
}

export default productTypeApi
