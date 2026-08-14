import axios from '@/utils/axios'
import {
  requireApiAcknowledgement,
  requireApiArrayField,
  requireApiField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
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

export const faqAdminApi = {
  async listFAQs(params: APIParams = {}) {
    const endpoint = '/api/admin/faqs'
    const response = await axios.get(endpoint, { params })
    const body = requireApiObject(response.data, endpoint, 'response body')
    const payload = readObjectPayload(response, endpoint)
    requireApiArrayField(payload, 'faqs', endpoint)
    requireApiPagination(body, payload, endpoint)
    return payload
  },

  async listFAQGroups(params: APIParams = {}) {
    const endpoint = '/api/admin/faqs/grouped'
    const payload = readObjectPayload(await axios.get(endpoint, { params }), endpoint)
    requireApiArrayField(payload, 'pages', endpoint)
    requireApiNumberField(payload, 'total', endpoint)
    return payload
  },

  async getFAQ(id: APIID) {
    const endpoint = `/api/admin/faqs/${id}`
    return requireApiObjectField(readObjectPayload(await axios.get(endpoint), endpoint), 'faq', endpoint)
  },

  async createFAQ(payload: APIPayload) {
    const endpoint = '/api/admin/faqs'
    return requireApiObjectField(readObjectPayload(await axios.post(endpoint, payload), endpoint), 'faq', endpoint)
  },

  async updateFAQ(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/faqs/${id}`
    return requireApiObjectField(readObjectPayload(await axios.put(endpoint, payload), endpoint), 'faq', endpoint)
  },

  async deleteFAQ(id: APIID) {
    const endpoint = `/api/admin/faqs/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },

  async deleteFAQs(ids: APIID[]) {
    const endpoint = '/api/admin/faqs/batch-delete'
    const result = readObjectPayload(await axios.post(endpoint, { faq_ids: ids }), endpoint)
    requireApiNumberField(result, 'deleted', endpoint)
    requireApiNumberField(result, 'total', endpoint)
    return result
  },

  async listStructure(locale: string) {
    const endpoint = '/api/admin/faqs/structure'
    const payload = readObjectPayload(await axios.get(endpoint, { params: { locale } }), endpoint)
    requireApiArrayField(payload, 'pages', endpoint)
    return payload
  },

  async updatePage(pageID: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/faqs/pages/${pageID}`
    return requireApiObjectField(readObjectPayload(await axios.put(endpoint, payload), endpoint), 'page', endpoint)
  },

  async createCategory(payload: APIPayload) {
    const endpoint = '/api/admin/faqs/categories'
    return requireApiObjectField(readObjectPayload(await axios.post(endpoint, payload), endpoint), 'category', endpoint)
  },

  async updateCategory(id: APIID, payload: APIPayload) {
    const endpoint = `/api/admin/faqs/categories/${id}`
    return requireApiObjectField(readObjectPayload(await axios.put(endpoint, payload), endpoint), 'category', endpoint)
  },

  async deleteCategory(id: APIID) {
    const endpoint = `/api/admin/faqs/categories/${id}`
    return requireApiAcknowledgement(await axios.delete(endpoint), endpoint)
  },
}

export default faqAdminApi
