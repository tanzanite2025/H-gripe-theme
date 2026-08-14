import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArray,
  requireApiField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

const readObjectPayload = (response: unknown, path: string) => (
  requireApiObject(unwrapApiPayload(response, path), path)
)

export const productApi = {
  async list(params: Record<string, any> = {}) {
    const path = '/api/admin/products'
    const response = await axios.get(path, { params })
    const payload = readObjectPayload(response, path)
    requireApiArray(payload.products, path, 'field "products"')
    requireApiObjectField(payload, 'pagination', path)
    const pagination = requireApiObjectField(payload, 'pagination', path)
    requireApiNumberField(pagination, 'page', path)
    requireApiNumberField(pagination, 'page_size', path)
    requireApiNumberField(pagination, 'total', path)
    requireApiNumberField(pagination, 'total_pages', path)
    return payload
  },

  async stats() {
    const path = '/api/admin/products/stats'
    const payload = readObjectPayload(await axios.get(path), path)
    requireApiNumberField(payload, 'total', path)
    requireApiNumberField(payload, 'featured', path)
    requireApiNumberField(payload, 'low_stock', path)
    requireApiNumberField(payload, 'out_of_stock', path)
    return payload
  },

  async get(id: number | string) {
    const path = `/api/admin/products/${id}`
    return requireApiObjectField(readObjectPayload(await axios.get(path), path), 'product', path)
  },

  async translations(id: number | string) {
    const path = `/api/admin/products/${id}/translations`
    return requireApiObjectField(readObjectPayload(await axios.get(path), path), 'translation_group', path)
  },

  async copyTranslation(id: number | string, targetLocale: string) {
    const path = `/api/admin/products/${id}/translations/copy`
    const payload = readObjectPayload(await axios.post(path, {
      target_locale: targetLocale
    }), path)
    requireApiObjectField(payload, 'product', path)
    requireApiObjectField(payload, 'translation_group', path)
    return payload
  },

  async create(payload: Record<string, any>) {
    const path = '/api/admin/products'
    return requireApiObjectField(readObjectPayload(await axios.post(path, payload), path), 'product', path)
  },

  async update(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/products/${id}`
    return requireApiObjectField(readObjectPayload(await axios.put(path, payload), path), 'product', path)
  },

  async updateStatus(id: number | string, status: string) {
    const path = `/api/admin/products/${id}/status`
    return requireApiAcknowledgement(await axios.patch(path, { status }), path)
  },

  async deleteProduct(id: number | string) {
    const path = `/api/admin/products/${id}`
    return requireApiAcknowledgement(await axios.delete(path), path)
  },

  async batchUpdateStatus(productIds: Array<number | string>, status: string) {
    const path = '/api/admin/products/batch-status'
    return requireApiAcknowledgement(await axios.post(path, { product_ids: productIds, status }), path)
  },

  async batchDelete(productIds: Array<number | string>) {
    const path = '/api/admin/products/batch-delete'
    return requireApiAcknowledgement(await axios.post(path, { product_ids: productIds }), path)
  }
}

export const productInformationTemplateApi = {
  async list(params: Record<string, any> = {}) {
    const path = '/api/admin/product-information-templates'
    const payload = unwrapApiPayload(await axios.get(path, { params }), path)
    return requireApiArray(payload, path, 'data')
  },

  async get(id: number | string) {
    const path = `/api/admin/product-information-templates/${id}`
    return requireApiObject(unwrapApiPayload(await axios.get(path), path), path, 'data')
  },

  async create(payload: Record<string, any>) {
    const path = '/api/admin/product-information-templates'
    return requireApiObject(unwrapApiPayload(await axios.post(path, payload), path), path, 'data')
  },

  async update(id: number | string, payload: Record<string, any>) {
    const path = `/api/admin/product-information-templates/${id}`
    return requireApiObject(unwrapApiPayload(await axios.put(path, payload), path), path, 'data')
  },

  async remove(id: number | string) {
    const path = `/api/admin/product-information-templates/${id}`
    return requireApiAcknowledgement(await axios.delete(path), path)
  }
}

export default productApi
