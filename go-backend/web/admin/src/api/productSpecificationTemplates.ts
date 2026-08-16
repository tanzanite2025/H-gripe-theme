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

const readProductSpecTemplate = (response: unknown, endpoint: string): any => {
  const productSpecTemplate = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)
  requireApiNumberField(productSpecTemplate, 'id', endpoint)
  requireApiStringField(productSpecTemplate, 'name', endpoint)
  requireApiStringField(productSpecTemplate, 'slug', endpoint)
  requireApiBooleanField(productSpecTemplate, 'is_enabled', endpoint)
  return productSpecTemplate
}

export const productSpecTemplateApi = {
  async list(params: Record<string, any> = {}) {
    const endpoint = '/api/admin/product-specification-templates'
    return requireApiArray(
      unwrapApiPayload(await axios.get(endpoint, { params }), endpoint),
      endpoint,
      'data',
    )
  },

  async create(payload: Record<string, any>) {
    const endpoint = '/api/admin/product-specification-templates'
    return readProductSpecTemplate(await axios.post(endpoint, payload), endpoint)
  },

  async update(id: number | string, payload: Record<string, any>) {
    const endpoint = `/api/admin/product-specification-templates/${id}`
    return readProductSpecTemplate(await axios.put(endpoint, payload), endpoint)
  },

  async uploadImage(id: number | string, file: File) {
    const formData = new FormData()
    formData.append('file', file)
    const endpoint = `/api/admin/product-specification-templates/${id}/image`
    return readProductSpecTemplate(await axios.post(endpoint, formData), endpoint)
  },

  async deleteImage(id: number | string) {
    const endpoint = `/api/admin/product-specification-templates/${id}/image`
    return readProductSpecTemplate(await axios.delete(endpoint), endpoint)
  },

  async deleteProductSpecTemplate(id: number | string) {
    const endpoint = `/api/admin/product-specification-templates/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  }
}

export default productSpecTemplateApi
